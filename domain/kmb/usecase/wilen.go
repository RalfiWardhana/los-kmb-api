package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

func (u metrics) Submission2Wilen(ctx context.Context, req request.Metrics, accessToken string) (resultMetrics interface{}, err error) {

	var (
		details              []entity.TrxDetail
		filtering            entity.FilteringKMB
		trxPrescreening      entity.TrxPrescreening
		trxFMF               response.TrxFMF
		cmoCluster           string
		ntfOther             response.NTFOther
		ntfDetails           response.NTFDetails
		ntfOtherAmount       float64
		ntfOtherAmountSpouse float64
		ntfConfinsAmount     response.OutstandingConfins
		confins              response.OutstandingConfins
		topup                response.IntegratorAgreementChassisNumber
		ntfAmount            float64
		customerID           string
		dupcheckData         response.SpDupcheckMap
		trxKPM               entity.TrxKPM
		isModified           bool
	)

	defer func() {
		if err == nil {
			_ = u.repository.UpdateTrxKPMStatus(trxKPM.ID, constant.STATUS_LOS_PROCESS_2WILEN)
		}
	}()

	// cek trx_master
	var trxMaster int
	trxMaster, err = u.repository.ScanTrxMaster(req.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error")
		return
	}

	// ProspectID Already Exist
	if trxMaster > 0 {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID Already Exist")
		return
	}

	filtering, err = u.repository.GetFilteringResult(req.Transaction.ProspectID)
	if err != nil {
		if err.Error() == constant.RECORD_NOT_FOUND {
			err = errors.New(fmt.Sprintf("%s - Belum melakukan filtering atau hasil filtering sudah lebih dari %s hari", constant.ERROR_BAD_REQUEST, os.Getenv("BIRO_VALID_DAYS")))
		} else {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error")
		}
		return
	}

	if filtering.NextProcess != 1 {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses")
		return
	}

	// cek elaborate ltv
	_, err = u.repository.GetElaborateLtv(req.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Belum melakukan pengecekan LTV")
		return
	}

	trxKPM, err = u.repository.GetTrxKPM(req.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Get Trx KPM 2Wilen Error")
		return
	}

	staySinceYear := strconv.Itoa(trxKPM.StaySinceYear)
	staySinceMonth := strconv.Itoa(trxKPM.StaySinceMonth)

	bpkbNameTypeKPM := 0
	if strings.Contains(os.Getenv("NAMA_SAMA"), trxKPM.BPKBName) {
		bpkbNameTypeKPM = 1
	}

	bpkbNameTypeReq := 0
	if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
		bpkbNameTypeReq = 1
	}

	if trxKPM.Education != req.CustomerPersonal.Education ||
		!compareNumOfDependence(trxKPM.NumOfDependence, req.CustomerPersonal.NumOfDependence) ||
		trxKPM.HomeStatus != req.CustomerPersonal.HomeStatus ||
		staySinceYear != req.CustomerPersonal.StaySinceYear ||
		staySinceMonth != req.CustomerPersonal.StaySinceMonth ||
		trxKPM.ProfessionID != req.CustomerEmployment.ProfessionID ||
		trxKPM.JobType != req.CustomerEmployment.JobType ||
		trxKPM.JobPosition != req.CustomerEmployment.JobPosition ||
		trxKPM.MonthlyFixedIncome != req.CustomerEmployment.MonthlyFixedIncome ||
		!compareMonthlyVariableIncome(0, req.CustomerEmployment.MonthlyVariableIncome) ||
		!compareSpouseIncome(trxKPM.SpouseIncome, req.CustomerEmployment.SpouseIncome) ||
		trxKPM.AssetUsageTypeCode != req.Item.AssetUsage ||
		trxKPM.AssetCategoryID != req.Item.CategoryID ||
		trxKPM.AssetCode != req.Item.AssetCode ||
		trxKPM.ManufactureYear != req.Item.ManufactureYear ||
		trxKPM.NoChassis != req.Item.NoChassis ||
		bpkbNameTypeKPM != bpkbNameTypeReq {
		isModified = true
	}

	if isModified {
		// STEP 1 CMO not recommend
		if req.Agent.CmoRecom == constant.CMO_NOT_RECOMMEDED {
			details = append(details, entity.TrxDetail{
				ProspectID:     req.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_ONPROCESS,
				Activity:       constant.ACTIVITY_PROCESS,
				Decision:       constant.DB_DECISION_REJECT,
				RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
				Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
				SourceDecision: constant.CMO_AGENT,
				NextStep:       constant.PRESCREENING,
			})

			details = append(details, entity.TrxDetail{
				ProspectID:     req.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				SourceDecision: constant.PRESCREENING,
				RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
				Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:      constant.SYSTEM_CREATED,
			})

			trxPrescreening = entity.TrxPrescreening{
				ProspectID: req.Transaction.ProspectID,
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			}

			resultMetrics, err = u.usecase.SaveTransaction(0, req, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Save Transaction Error")
				return
			}
			return
		}

		// STEP 1 CMO recommend
		details = append(details, entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       constant.CODE_CMO_RECOMMENDED,
			Reason:         constant.REASON_CMO_RECOMMENDED,
			SourceDecision: constant.CMO_AGENT,
			NextStep:       constant.PRESCREENING,
		})

		// STEP 2 prescreening
		var trxPrescreeningDetail entity.TrxDetail
		trxPrescreening, trxFMF, trxPrescreeningDetail, err = u.usecase.Prescreening(ctx, req, filtering, accessToken)
		if err != nil {
			return
		}

		details = append(details, trxPrescreeningDetail)

		// prescreening ke CA
		if trxPrescreening.Decision != constant.DB_DECISION_APR {
			resultMetrics, err = u.usecase.SaveTransaction(0, req, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
			if err != nil {
				return
			}

			return
		}
	} else {
		trxPrescreening = entity.TrxPrescreening{
			ProspectID: req.Transaction.ProspectID,
			Decision:   constant.DB_DECISION_APR,
			Reason:     "sesuai",
			CreatedBy:  constant.SYSTEM_CREATED,
			DecisionBy: constant.SYSTEM_CREATED,
		}

		// trx detail cmo
		trxDetailCMO := entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       constant.CODE_CMO_RECOMMENDED,
			SourceDecision: constant.CMO_AGENT,
			NextStep:       constant.PRESCREENING,
			CreatedBy:      constant.SYSTEM_CREATED,
			Reason:         constant.REASON_CMO_RECOMMENDED,
		}

		details = append(details, trxDetailCMO)

		// trx detail prescreening
		trxDetailPrescreening := entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       constant.CODE_PASS_PRESCREENING,
			SourceDecision: constant.PRESCREENING,
			NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			Info:           "Sesuai",
			CreatedBy:      constant.SYSTEM_CREATED,
			Reason:         trxPrescreening.Reason,
		}

		details = append(details, trxDetailPrescreening)

		if filtering.CustomerID != nil {
			customerID = filtering.CustomerID.(string)
		}

		// cek ntf gantung all lob
		var customerData []request.CustomerData
		customerData = append(customerData, request.CustomerData{
			TransactionID: req.Transaction.ProspectID,
			IDNumber:      req.CustomerPersonal.IDNumber,
			LegalName:     req.CustomerPersonal.LegalName,
			BirthDate:     req.CustomerPersonal.BirthDate,
			MotherName:    req.CustomerPersonal.SurgateMotherName,
			CustomerID:    customerID,
		})

		if req.CustomerPersonal.MaritalStatus == constant.MARRIED && req.CustomerSpouse != nil {
			spouse := *req.CustomerSpouse
			customerData = append(customerData, request.CustomerData{
				TransactionID: req.Transaction.ProspectID,
				IDNumber:      spouse.IDNumber,
				LegalName:     spouse.LegalName,
				BirthDate:     spouse.BirthDate,
				MotherName:    spouse.SurgateMotherName,
			})
		}

		header := map[string]string{}

		for i, customer := range customerData {

			jsonCustomer, _ := json.Marshal(customer)
			var ntfLOS *resty.Response

			ntfLOS, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_PENDING_URL"), jsonCustomer, header, constant.METHOD_POST, true, 3, 60, req.Transaction.ProspectID, accessToken)

			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error")
				return
			}

			if ntfLOS.StatusCode() != 200 {
				err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error")
				return
			}

			json.Unmarshal([]byte(jsoniter.Get(ntfLOS.Body(), "data").ToString()), &ntfOther)

			if i == 0 {
				// customer
				ntfOtherAmount = ntfOther.NTFAmountKmbOff + ntfOther.NTFAmountWgOff + ntfOther.NTFAmountKmobOff + ntfOther.NTFAmountUC + ntfOther.NTFAmountWgOnl + ntfOther.NTFAmountNewKmb
				ntfDetails.Customer = ntfOther
				confins.TotalOutstanding = ntfOther.TotalOutstanding

			} else if i == 1 {
				// spouse
				ntfOtherAmountSpouse = ntfOther.NTFAmountKmbOff + ntfOther.NTFAmountWgOff + ntfOther.NTFAmountKmobOff + ntfOther.NTFAmountUC + ntfOther.NTFAmountWgOnl + ntfOther.NTFAmountNewKmb
				ntfDetails.Spouse = ntfOther
			}
		}

		var ntfTopup *resty.Response
		ntfTopup, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+req.Item.NoChassis, nil, header, constant.METHOD_GET, true, 3, 60, req.Transaction.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error")
			return
		}

		if ntfTopup.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error")
			return
		}

		err = json.Unmarshal([]byte(jsoniter.Get(ntfTopup.Body(), "data").ToString()), &topup)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error")
			return
		}

		ntfConfinsAmount.TotalOutstanding = confins.TotalOutstanding - topup.OutstandingPrincipal

		ntfAmount = req.Apk.NTF + (ntfOtherAmount + ntfOtherAmountSpouse) + ntfConfinsAmount.TotalOutstanding

		sntfDetails, _ := json.Marshal(ntfDetails)

		trxFMF = response.TrxFMF{
			NTFAkumulasi:         ntfAmount,
			NTFOtherAmount:       ntfOtherAmount,
			NTFOtherAmountSpouse: ntfOtherAmountSpouse,
			NTFOtherAmountDetail: string(utils.SafeEncoding(sntfDetails)),
			NTFConfinsAmount:     ntfConfinsAmount.TotalOutstanding,
			NTFConfins:           confins.TotalOutstanding,
			NTFTopup:             topup.OutstandingPrincipal,
		}

		req.Transaction.ApplicationSource = "K"

		resultMetrics, err = u.usecase.SaveTransaction(0, req, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
		if err != nil {
			return
		}
	}

	// get mapping cluster
	mappingCluster := entity.MasterMappingCluster{
		BranchID:       req.Transaction.BranchID,
		CustomerStatus: filtering.CustomerStatus.(string),
	}
	if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
		mappingCluster.BpkbNameType = 1
	}
	if strings.Contains(constant.STATUS_KONSUMEN_RO_AO, filtering.CustomerStatus.(string)) {
		mappingCluster.CustomerStatus = "AO/RO"
	}

	mappingCluster, err = u.repository.MasterMappingCluster(mappingCluster)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error")
		return
	}

	if clusterName, ok := filtering.CMOCluster.(string); ok {
		cmoCluster = clusterName
	} else {
		cmoCluster = mappingCluster.Cluster
	}

	if req.Apk.Tenor >= 36 {
		// trx detail reject tenor 36
		trxDetailTenor := entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       constant.CODE_PASS_TENOR,
			SourceDecision: constant.SOURCE_DECISION_TENOR,
			NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			CreatedBy:      constant.SYSTEM_CREATED,
			Reason:         constant.REASON_PASS_TENOR,
			Info:           fmt.Sprintf("Cluster : %s", cmoCluster),
		}

		details = append(details, trxDetailTenor)
	}

	decodedData := utils.SafeDecoding(trxKPM.DupcheckData)
	err = json.Unmarshal([]byte(decodedData), &dupcheckData)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal Dupcheck Data Error")
		return
	}

	trxFMF.DupcheckData = dupcheckData
	trxFMF.CustomerStatus = dupcheckData.StatusKonsumen
	trxFMF.DSRFMF = dupcheckData.Dsr

	// internal record
	if dupcheckData.CustomerID != nil {
		if dupcheckData.CustomerID.(string) != "" {
			var resInternalRecord *resty.Response
			resInternalRecord, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("INTERNAL_RECORD_URL")+dupcheckData.CustomerID.(string), nil, map[string]string{}, constant.METHOD_GET, true, 3, 60, req.Transaction.ProspectID, accessToken)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Interal Record Error")
				return
			}
			jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resInternalRecord.Body(), "data").ToString()), &trxFMF.AgreementCONFINS)
		}
	}

	// trx detail rejection
	trxDetailRejection := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       constant.CODE_BELUM_PERNAH_REJECT,
		SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
		NextStep:       constant.SOURCE_DECISION_BLACKLIST,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         constant.REASON_BELUM_PERNAH_REJECT,
	}

	details = append(details, trxDetailRejection)

	// trx detail blacklist
	trxDetailBlacklist := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.CheckBlacklistCode,
		SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
		NextStep:       constant.SOURCE_DECISION_PMK,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.CheckBlacklistReason,
	}

	details = append(details, trxDetailBlacklist)

	// trx detail vehicle check
	trxDetailVehicleCheck := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.CheckVehicleCode,
		SourceDecision: constant.SOURCE_DECISION_PMK,
		NextStep:       constant.SOURCE_DECISION_NOKANOSIN,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.CheckVehicleReason,
		Info:           trxKPM.CheckVehicleInfo,
	}

	details = append(details, trxDetailVehicleCheck)

	// trx detail noka nosin
	trxDetailChassisNumber := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.RuleCode,
		SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
		NextStep:       constant.SOURCE_DECISION_PMK,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.Reason,
	}

	details = append(details, trxDetailChassisNumber)

	// trx detail pmk
	trxDetailPMK := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.CheckPMKCode,
		SourceDecision: constant.SOURCE_DECISION_PMK,
		NextStep:       constant.SOURCE_DECISION_DSR,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.CheckPMKReason,
	}

	details = append(details, trxDetailPMK)

	// trx detail dsr
	trxDetailDSR := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.CheckDSRCode,
		SourceDecision: constant.SOURCE_DECISION_DSR,
		NextStep:       constant.SOURCE_DECISION_DUPCHECK,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.CheckDSRReason,
	}

	details = append(details, trxDetailDSR)

	// trx detail dupcheck
	info, _ := json.Marshal(dupcheckData)

	reasonCustomer := dupcheckData.StatusKonsumen

	var customerSegment string
	if filtering.CustomerSegment == nil {
		customerSegment = constant.RO_AO_REGULAR
	} else {
		customerSegment = filtering.CustomerSegment.(string)
	}

	if strings.Contains("PRIME PRIORITY", customerSegment) {
		reasonCustomer = fmt.Sprintf("%s %s", dupcheckData.StatusKonsumen, customerSegment)
	}
	trxDetailDupcheck := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       constant.CODE_PASS_MAX_OVD_CONFINS,
		SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
		NextStep:       constant.SOURCE_DECISION_DUKCAPIL,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         fmt.Sprintf("%s", reasonCustomer),
		Info:           string(utils.SafeEncoding(info)),
	}

	details = append(details, trxDetailDupcheck)

	// trx detail ekyc
	var ekycSource string
	if trxKPM.CheckEkycSource != nil {
		ekycSource = trxKPM.CheckEkycSource.(string)
	}

	var ekycSourceDetail string
	switch ekycSource {
	case constant.SOURCE_DECISION_KTP_VALIDATOR:
		ekycSourceDetail = "KTP VALIDATOR"
	case constant.ASLIRI:
		ekycSourceDetail = "ASLI RI"
	default:
		ekycSourceDetail = "DUKCAPIL"
	}

	trxDetailEkyc := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.CheckEkycCode,
		SourceDecision: ekycSource,
		NextStep:       constant.SOURCE_DECISION_BIRO,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.CheckEkycReason,
		Info:           trxKPM.CheckEkycInfo,
	}

	details = append(details, trxDetailEkyc)

	trxFMF.EkycSource = ekycSourceDetail
	trxFMF.EkycSimiliarity = trxKPM.CheckEkycSimiliarity
	trxFMF.EkycReason = trxKPM.CheckEkycReason

	// trx detail pefindo
	trxDetailPefindo := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.FilteringCode,
		SourceDecision: constant.SOURCE_DECISION_BIRO,
		NextStep:       constant.SOURCE_DECISION_SCOREPRO,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.FilteringReason,
	}

	details = append(details, trxDetailPefindo)

	// trx detail scorepro
	trxFMF.ScsDecision = response.ScsDecision{
		ScsDate:   trxKPM.CreatedAt.Format("2006-01-02"),
		ScsStatus: trxKPM.ScoreProResult,
		ScsScore:  trxKPM.ScoreProScoreResult,
	}

	trxDetailScorepro := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.ScoreProCode,
		SourceDecision: constant.SOURCE_DECISION_SCOREPRO,
		NextStep:       constant.SOURCE_DECISION_DSR,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxKPM.ScoreProReason,
		Info:           trxKPM.ScoreProInfo,
	}

	details = append(details, trxDetailScorepro)

	// trx detail dsr fmf pbk
	trxDetailDSRPBK := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       trxKPM.CheckDSRFMFPBKCode,
		SourceDecision: constant.SOURCE_DECISION_DSR,
		Info:           trxKPM.CheckDSRFMFPBKInfo,
		Reason:         trxKPM.CheckDSRFMFPBKReason,
		NextStep:       constant.SOURCE_DECISION_ELABORATE_LTV,
	}

	details = append(details, trxDetailDSRPBK)

	var infoTotalDSR map[string]interface{}
	decodedData = utils.SafeDecoding(trxKPM.CheckDSRFMFPBKInfo.(string))
	err = json.Unmarshal([]byte(decodedData), &infoTotalDSR)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal DSR FMF PBK Info Error")
		return
	}

	trxFMF.DSRPBK = infoTotalDSR["dsr_pbk"]
	trxFMF.TotalDSR = infoTotalDSR["total_dsr"]

	var latestInstallmentAmount float64
	if infoTotalDSR["latest_installment_amount"] != nil {
		latestInstallmentAmount, err = utils.GetFloat(infoTotalDSR["latest_installment_amount"])
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat latest_installment_amount Error")
			return
		}
	}

	if latestInstallmentAmount > 0 {
		trxFMF.LatestInstallmentAmount = latestInstallmentAmount
	}

	// trx detail ltv
	trxDetailLTV := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       constant.STRING_CODE_PASS_ELABORATE,
		SourceDecision: constant.SOURCE_DECISION_ELABORATE_LTV,
		Reason:         constant.REASON_PASS_ELABORATE,
		NextStep:       constant.SOURCE_DECISION_CA,
	}

	details = append(details, trxDetailLTV)

	details = append(details, entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.SOURCE_DECISION_CA,
	})

	finalReasonMetrics := dupcheckData.Reason

	resultMetrics, err = u.usecase.SaveTransaction(1, req, trxPrescreening, trxFMF, details, finalReasonMetrics)
	if err != nil {
		return
	}

	return
}

func compareNumOfDependence(a int, b *int) bool {
	if b == nil {
		return a == 0
	}

	return a == *b
}

func compareMonthlyVariableIncome(a float64, b *float64) bool {
	if b == nil {
		return a == 0
	}

	return a == *b
}

func compareSpouseIncome(a interface{}, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && b != nil && *b == 0 {
		return true
	}

	var aFloat float64
	switch v := a.(type) {
	case float64:
		aFloat = v
	case int:
		aFloat = float64(v)
	case int64:
		aFloat = float64(v)
	case float32:
		aFloat = float64(v)
	default:
		return false
	}

	if b == nil {
		return aFloat == 0
	}

	return aFloat == *b
}
