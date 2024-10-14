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
	"strings"
	"sync"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

func (u metrics) PrincipleSubmission(ctx context.Context, req request.Metrics, accessToken string) (resultMetrics interface{}, err error) {

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
		principleStepOne     entity.TrxPrincipleStepOne
		principleStepTwo     entity.TrxPrincipleStepTwo
		principleStepThree   entity.TrxPrincipleStepThree
		wg                   sync.WaitGroup
		errChan              = make(chan error, 3)
	)

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

	wg.Add(3)
	go func() {
		defer wg.Done()
		principleStepOne, err = u.repository.GetPrincipleStepOne(req.Transaction.ProspectID)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step One Error")
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		principleStepTwo, err = u.repository.GetPrincipleStepTwo(req.Transaction.ProspectID)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step Two Error")
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		principleStepThree, err = u.repository.GetPrincipleStepThree(req.Transaction.ProspectID)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step Three Error")
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return nil, err
	}

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
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         trxPrescreening.Reason,
	}

	details = append(details, trxDetailPrescreening)

	resultMetrics, err = u.usecase.SaveTransaction(0, req, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
	if err != nil {
		return
	}

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
			RuleCode:       principleStepThree.CheckRejectTenor36Code,
			SourceDecision: constant.SOURCE_DECISION_TENOR,
			NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			CreatedBy:      constant.SYSTEM_CREATED,
			Reason:         principleStepThree.CheckRejectTenor36Reason,
			Info:           fmt.Sprintf("Cluster : %s", cmoCluster),
		}

		details = append(details, trxDetailTenor)
	}

	decodedData := utils.SafeDecoding(principleStepTwo.DupcheckData)
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
	rejectionCode := principleStepTwo.CheckBannedPMKDSRCode
	rejectionReason := principleStepTwo.CheckBannedPMKDSRReason
	if rejectionCode == "" {
		rejectionCode = principleStepTwo.CheckRejectionCode
		rejectionReason = principleStepTwo.CheckRejectionReason
	}

	trxDetailRejection := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       rejectionCode,
		SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
		NextStep:       constant.SOURCE_DECISION_BLACKLIST,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         rejectionReason,
	}

	details = append(details, trxDetailRejection)

	// trx detail blacklist
	trxDetailBlacklist := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepTwo.CheckBlacklistCode,
		SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
		NextStep:       constant.SOURCE_DECISION_PMK,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepTwo.CheckBlacklistReason,
	}

	details = append(details, trxDetailBlacklist)

	// trx detail vehicle check
	trxDetailVehicleCheck := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepThree.CheckVehicleCode,
		SourceDecision: constant.SOURCE_DECISION_PMK,
		NextStep:       constant.SOURCE_DECISION_NOKANOSIN,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepThree.CheckVehicleReason,
		Info:           principleStepThree.CheckVehicleInfo,
	}

	details = append(details, trxDetailVehicleCheck)

	// trx detail noka nosin
	trxDetailChassisNumber := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepOne.RuleCode,
		SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
		NextStep:       constant.SOURCE_DECISION_PMK,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepOne.Reason,
	}

	details = append(details, trxDetailChassisNumber)

	// trx detail pmk
	trxDetailPMK := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepTwo.CheckPMKCode,
		SourceDecision: constant.SOURCE_DECISION_PMK,
		NextStep:       constant.SOURCE_DECISION_DSR,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepTwo.CheckPMKReason,
	}

	details = append(details, trxDetailPMK)

	// trx detail dsr
	trxDetailDSR := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepThree.CheckDSRCode,
		SourceDecision: constant.SOURCE_DECISION_DSR,
		NextStep:       constant.SOURCE_DECISION_DUPCHECK,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepThree.CheckDSRReason,
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
	if principleStepTwo.CheckEkycSource != nil {
		ekycSource = principleStepTwo.CheckEkycSource.(string)
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
		RuleCode:       principleStepTwo.CheckEkycCode,
		SourceDecision: ekycSource,
		NextStep:       constant.SOURCE_DECISION_BIRO,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepTwo.CheckEkycReason,
		Info:           principleStepTwo.CheckEkycInfo,
	}

	details = append(details, trxDetailEkyc)

	trxFMF.EkycSource = ekycSourceDetail
	trxFMF.EkycSimiliarity = principleStepTwo.CheckEkycSimiliarity
	trxFMF.EkycReason = principleStepTwo.CheckEkycReason

	// trx detail pefindo
	trxDetailPefindo := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepTwo.FilteringCode,
		SourceDecision: constant.SOURCE_DECISION_BIRO,
		NextStep:       constant.SOURCE_DECISION_SCOREPRO,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepTwo.FilteringReason,
	}

	details = append(details, trxDetailPefindo)

	// trx detail scorepro
	trxFMF.ScsDecision = response.ScsDecision{
		ScsDate:   principleStepThree.CreatedAt.Format("2006-01-02"),
		ScsStatus: principleStepThree.ScoreProResult,
		ScsScore:  principleStepThree.ScoreProScoreResult,
	}

	trxDetailScorepro := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepThree.ScoreProCode,
		SourceDecision: constant.SOURCE_DECISION_SCOREPRO,
		NextStep:       constant.SOURCE_DECISION_DSR,
		CreatedBy:      constant.SYSTEM_CREATED,
		Reason:         principleStepThree.ScoreProReason,
		Info:           principleStepThree.ScoreProInfo,
	}

	details = append(details, trxDetailScorepro)

	// trx detail dsr fmf pbk
	trxDetailDSRPBK := entity.TrxDetail{
		ProspectID:     req.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       principleStepThree.CheckDSRFMFPBKCode,
		SourceDecision: constant.SOURCE_DECISION_DSR,
		Info:           principleStepThree.CheckDSRFMFPBKInfo,
		Reason:         principleStepThree.CheckDSRFMFPBKReason,
		NextStep:       constant.SOURCE_DECISION_ELABORATE_LTV,
	}

	details = append(details, trxDetailDSRPBK)

	var infoTotalDSR map[string]interface{}
	decodedData = utils.SafeDecoding(principleStepThree.CheckDSRFMFPBKInfo.(string))
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
		SourceDecision: constant.SOURCE_DECISION_DSR,
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
