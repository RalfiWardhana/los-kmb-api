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
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

// func ini digunakan ketika Submit to LOS
// dan ketika Approve prescreening di menu prescreening oleh CA
func (u metrics) MetricsLos(ctx context.Context, reqMetrics request.Metrics, accessToken, hrisAccessToken string) (resultMetrics interface{}, err error) {

	var (
		married           bool
		details           []entity.TrxDetail
		reqDupcheck       request.DupcheckApi
		dupcheckData      response.SpDupcheckMap
		customerStatus    string
		customerSegment   string
		metricsDupcheck   response.UsecaseApi
		filtering         entity.FilteringKMB
		trxPrescreening   entity.TrxPrescreening
		trxFMF            response.TrxFMF
		trxFMFDupcheck    response.TrxFMF
		trxDetailDupcheck []entity.TrxDetail
		cbFound           bool
		cmoCluster        string
		mappingMaxDSR     entity.MasterMappingIncomeMaxDSR
	)

	// cek principle order
	var countTrxPrinciple int
	countTrxPrinciple, err = u.repository.ScanTrxPrinciple(reqMetrics.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Check Principle Order Error")
		return
	}

	if countTrxPrinciple > 0 {
		resultMetrics, err = u.PrincipleSubmission(ctx, reqMetrics, accessToken)
		return
	}

	// cek 2wilen order
	var countTrxKPM int
	countTrxKPM, err = u.repository.ScanTrxKPM(reqMetrics.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Check 2Wilen Order Error")
		return
	}

	if countTrxKPM > 0 {
		trxKPMStatus, errGetTrxKPMStatus := u.repository.GetTrxKPMStatus(reqMetrics.Transaction.ProspectID)
		if errGetTrxKPMStatus != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Check 2Wilen Order Status Error")
			return
		}

		if trxKPMStatus.Decision == constant.DECISION_KPM_APPROVE {
			resultMetrics, err = u.Submission2Wilen(ctx, reqMetrics, accessToken)
			return
		}
	}

	// cek trx_master
	var trxMaster int
	trxMaster, err = u.repository.ScanTrxMaster(reqMetrics.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error")
		return
	}

	// ProspectID Already Exist
	if trxMaster > 0 {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID Already Exist")
		return
	}

	// cek prescreening
	var countTrx int
	countTrx, err = u.repository.ScanTrxPrescreening(reqMetrics.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Prescreening Error")
		return
	}

	// cek filtering
	if countTrx == 0 {
		filtering, err = u.repository.GetFilteringResult(reqMetrics.Transaction.ProspectID)

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
	} else {
		filtering, err = u.repository.GetFilteringForJourney(reqMetrics.Transaction.ProspectID)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error")
			return
		}
	}

	// cek elaborate ltv
	_, err = u.repository.GetElaborateLtv(reqMetrics.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Belum melakukan pengecekan LTV")
		return
	}

	if filtering.ScoreBiro != nil {
		if filtering.ScoreBiro.(string) != "" && filtering.ScoreBiro.(string) != constant.DECISION_PBK_NO_HIT && filtering.ScoreBiro.(string) != constant.PEFINDO_UNSCORE {
			cbFound = true
		}
	}

	// belum prescreening
	if countTrx == 0 {
		// STEP 1 CMO not recommend
		if reqMetrics.Agent.CmoRecom == constant.CMO_NOT_RECOMMEDED {
			details = append(details, entity.TrxDetail{
				ProspectID:     reqMetrics.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_ONPROCESS,
				Activity:       constant.ACTIVITY_PROCESS,
				Decision:       constant.DB_DECISION_REJECT,
				RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
				Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
				SourceDecision: constant.CMO_AGENT,
				NextStep:       constant.PRESCREENING,
			})

			details = append(details, entity.TrxDetail{
				ProspectID:     reqMetrics.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				SourceDecision: constant.PRESCREENING,
				RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
				Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:      constant.SYSTEM_CREATED,
			})

			trxPrescreening = entity.TrxPrescreening{
				ProspectID: reqMetrics.Transaction.ProspectID,
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			}

			resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Save Transaction Error")
				return
			}
			return
		}

		// STEP 1 CMO recommend
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
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
		trxPrescreening, trxFMF, trxPrescreeningDetail, err = u.usecase.Prescreening(ctx, reqMetrics, filtering, accessToken)
		if err != nil {
			return
		}

		details = append(details, trxPrescreeningDetail)

		// prescreening ke CA
		if trxPrescreening.Decision != constant.DB_DECISION_APR {
			resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
			if err != nil {
				return
			}

			return
		}
	}

	// get mapping cluster
	mappingCluster := entity.MasterMappingCluster{
		BranchID:       reqMetrics.Transaction.BranchID,
		CustomerStatus: filtering.CustomerStatus.(string),
	}
	if strings.Contains(os.Getenv("NAMA_SAMA"), reqMetrics.Item.BPKBName) {
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

	//  STEP 3 tenor 36
	if reqMetrics.Apk.Tenor >= 36 {
		var trxTenor response.UsecaseApi
		if reqMetrics.Apk.Tenor == 36 {
			trxTenor, err = u.usecase.RejectTenor36(cmoCluster)
			if err != nil {
				return
			}
		} else if reqMetrics.Apk.Tenor > 36 {
			trxTenor = response.UsecaseApi{
				Code:   constant.CODE_REJECT_TENOR,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_TENOR,
			}
		}

		if trxTenor.Result == constant.DECISION_REJECT {
			details = append(details, entity.TrxDetail{
				ProspectID:     reqMetrics.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				RuleCode:       trxTenor.Code,
				SourceDecision: constant.SOURCE_DECISION_TENOR,
				CreatedBy:      constant.SYSTEM_CREATED,
				Reason:         trxTenor.Reason,
				Info:           fmt.Sprintf("Cluster : %s", cmoCluster),
			})

			resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, trxTenor.Reason)
			if err != nil {
				return
			}
			return
		}

		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       trxTenor.Code,
			SourceDecision: constant.SOURCE_DECISION_TENOR,
			NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			CreatedBy:      constant.SYSTEM_CREATED,
			Reason:         trxTenor.Reason,
			Info:           fmt.Sprintf("Cluster : %s", cmoCluster),
		})
	}

	var selfieImage, ktpImage, legalZipCode, companyZipCode string

	for i := 0; i < len(reqMetrics.CustomerPhoto); i++ {

		if reqMetrics.CustomerPhoto[i].ID == constant.TAG_KTP_PHOTO {
			ktpImage = reqMetrics.CustomerPhoto[i].Url

		} else if reqMetrics.CustomerPhoto[i].ID == constant.TAG_SELFIE_PHOTO {
			selfieImage = reqMetrics.CustomerPhoto[i].Url
		}
	}

	for i := 0; i < len(reqMetrics.Address); i++ {

		if reqMetrics.Address[i].Type == constant.ADDRESS_TYPE_LEGAL {
			legalZipCode = reqMetrics.Address[i].ZipCode

		} else if reqMetrics.Address[i].Type == constant.ADDRESS_TYPE_COMPANY {
			companyZipCode = reqMetrics.Address[i].ZipCode
		}
	}

	if companyZipCode == "" {
		companyZipCode = legalZipCode
	}

	if filtering.CustomerSegment == nil {
		customerSegment = constant.RO_AO_REGULAR
	} else {
		customerSegment = filtering.CustomerSegment.(string)
	}

	var numOfDependence int
	if reqMetrics.CustomerPersonal.NumOfDependence != nil {
		numOfDependence = *reqMetrics.CustomerPersonal.NumOfDependence
	}

	reqDupcheck = request.DupcheckApi{
		ProspectID:            reqMetrics.Transaction.ProspectID,
		BranchID:              reqMetrics.Transaction.BranchID,
		ImageKtp:              ktpImage,
		ImageSelfie:           selfieImage,
		MonthlyFixedIncome:    reqMetrics.CustomerEmployment.MonthlyFixedIncome,
		HomeStatus:            reqMetrics.CustomerPersonal.HomeStatus,
		MonthlyVariableIncome: *reqMetrics.CustomerEmployment.MonthlyVariableIncome,
		SpouseIncome:          *reqMetrics.CustomerEmployment.SpouseIncome,
		JobType:               reqMetrics.CustomerEmployment.JobType,
		JobPosition:           reqMetrics.CustomerEmployment.JobPosition,
		ProfessionID:          reqMetrics.CustomerEmployment.ProfessionID,
		EmploymentSinceYear:   reqMetrics.CustomerEmployment.EmploymentSinceYear,
		EmploymentSinceMonth:  reqMetrics.CustomerEmployment.EmploymentSinceMonth,
		StaySinceYear:         reqMetrics.CustomerPersonal.StaySinceYear,
		StaySinceMonth:        reqMetrics.CustomerPersonal.StaySinceMonth,
		BirthDate:             reqMetrics.CustomerPersonal.BirthDate,
		BirthPlace:            reqMetrics.CustomerPersonal.BirthPlace,
		Tenor:                 reqMetrics.Apk.Tenor,
		IDNumber:              reqMetrics.CustomerPersonal.IDNumber,
		LegalName:             reqMetrics.CustomerPersonal.LegalName,
		MotherName:            reqMetrics.CustomerPersonal.SurgateMotherName,
		MobilePhone:           reqMetrics.CustomerPersonal.MobilePhone,
		EngineNo:              reqMetrics.Item.NoEngine,
		RangkaNo:              reqMetrics.Item.NoChassis,
		ManufactureYear:       reqMetrics.Item.ManufactureYear,
		BPKBName:              reqMetrics.Item.BPKBName,
		NumOfDependence:       numOfDependence,
		DPAmount:              reqMetrics.Apk.DPAmount,
		OTRPrice:              reqMetrics.Apk.OTR,
		NTF:                   reqMetrics.Apk.NTF,
		LegalZipCode:          legalZipCode,
		CompanyZipCode:        companyZipCode,
		Gender:                reqMetrics.CustomerPersonal.Gender,
		InstallmentAmount:     reqMetrics.Apk.InstallmentAmount,
		MaritalStatus:         reqMetrics.CustomerPersonal.MaritalStatus,
		CustomerSegment:       customerSegment,
		Dealer:                reqMetrics.Apk.Dealer,
		AdminFee:              reqMetrics.Apk.AdminFee,
		CMOCluster:            cmoCluster,
		AF:                    reqMetrics.Apk.AF,
		Filtering:             filtering,
	}

	if reqMetrics.CustomerSpouse != nil {
		var spouse = request.DupcheckApiSpouse{
			BirthDate:  reqMetrics.CustomerSpouse.BirthDate,
			Gender:     reqMetrics.CustomerSpouse.Gender,
			IDNumber:   reqMetrics.CustomerSpouse.IDNumber,
			LegalName:  reqMetrics.CustomerSpouse.LegalName,
			MotherName: reqMetrics.CustomerSpouse.SurgateMotherName,
		}

		reqDupcheck.Spouse = &spouse
		married = true
	}

	//Get parameterize config
	config, err := u.repository.GetConfig("dupcheck", "KMB-OFF", "dupcheck_kmb_config")

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Config Error")
		return
	}

	var configValue response.DupcheckConfig

	json.Unmarshal([]byte(config.Value), &configValue)

	// get config max dsr
	if mappingCluster.Cluster == "" {
		mappingCluster.Cluster = "Cluster C"
	}

	reqDupcheck.Cluster = mappingCluster.Cluster

	totalIncome := reqDupcheck.MonthlyFixedIncome + reqDupcheck.MonthlyVariableIncome + reqDupcheck.SpouseIncome
	mappingMaxDSR, err = u.repository.MasterMappingIncomeMaxDSR(totalIncome)
	if err != nil {
		if err.Error() != constant.DATA_NOT_FOUND {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Max DSR error")
			return
		}
	} else {
		configValue.Data.MaxDsr = mappingMaxDSR.DSRThreshold
	}

	dupcheckData, customerStatus, metricsDupcheck, trxFMFDupcheck, trxDetailDupcheck, err = u.multiUsecase.Dupcheck(ctx, reqDupcheck, married, accessToken, hrisAccessToken, configValue)
	if err != nil {
		return
	}

	trxFMF.DupcheckData = dupcheckData
	trxFMF.CustomerStatus = customerStatus
	trxFMF.DSRFMF = dupcheckData.Dsr
	trxFMF.TrxBannedPMKDSR = trxFMFDupcheck.TrxBannedPMKDSR
	trxFMF.TrxBannedChassisNumber = trxFMFDupcheck.TrxBannedChassisNumber
	trxFMF.TrxEDD = trxFMFDupcheck.TrxEDD

	// cek tread AO as regular customer
	if dupcheckData.IsFlowAsAORegular {
		customerSegment = constant.RO_AO_REGULAR
		filtering.CustomerSegment = constant.RO_AO_REGULAR
	}

	// internal record
	if dupcheckData.CustomerID != nil {
		if dupcheckData.CustomerID.(string) != "" {
			var resInternalRecord *resty.Response
			resInternalRecord, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("INTERNAL_RECORD_URL")+dupcheckData.CustomerID.(string), nil, map[string]string{}, constant.METHOD_GET, true, 3, 60, reqMetrics.Transaction.ProspectID, accessToken)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Interal Record Error")
				return
			}
			jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resInternalRecord.Body(), "data").ToString()), &trxFMF.AgreementCONFINS)
		}
	}

	if metricsDupcheck.Result == constant.DECISION_REJECT {
		details = append(details, trxDetailDupcheck...)

		info, _ := json.Marshal(dupcheckData)

		addDetail := entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       metricsDupcheck.Code,
			SourceDecision: metricsDupcheck.SourceDecision,
			Reason:         metricsDupcheck.Reason,
			Info:           string(utils.SafeEncoding(info)),
		}

		details = append(details, addDetail)

		resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsDupcheck.Reason)
		if err != nil {
			return
		}

		return
	}

	details = append(details, trxDetailDupcheck...)

	reqMetricsEkyc := request.MetricsEkyc{
		CBFound:         cbFound,
		CustomerStatus:  customerStatus,
		CustomerSegment: customerSegment,
	}

	decisionEkyc, trxDetailEkyc, trxFMFEkyc, err := u.multiUsecase.Ekyc(ctx, reqMetrics, reqMetricsEkyc, accessToken)
	if err != nil {
		return
	}

	if len(trxDetailEkyc) > 0 {
		details = append(details, trxDetailEkyc...)
	}

	trxFMF.EkycSource = trxFMFEkyc.EkycSource
	trxFMF.EkycSimiliarity = trxFMFEkyc.EkycSimiliarity
	trxFMF.EkycReason = trxFMFEkyc.EkycReason

	if decisionEkyc.Result == constant.DECISION_REJECT {

		addDetail := entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       decisionEkyc.Code,
			SourceDecision: decisionEkyc.Source,
			Reason:         decisionEkyc.Reason,
			Info:           decisionEkyc.Info,
		}

		details = append(details, addDetail)

		resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, decisionEkyc.Reason)
		if err != nil {
			return
		}

		return
	}

	details = append(details, entity.TrxDetail{
		ProspectID:     reqMetrics.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       decisionEkyc.Code,
		SourceDecision: decisionEkyc.Source,
		Info:           decisionEkyc.Info,
		Reason:         decisionEkyc.Reason,
		NextStep:       constant.SOURCE_DECISION_BIRO,
	})

	metricsPefindo, err := u.usecase.Pefindo(cbFound, reqMetrics.Item.BPKBName, filtering, dupcheckData)
	if err != nil {
		return
	}

	if metricsPefindo.Result == constant.DECISION_REJECT {

		addDetail := entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       metricsPefindo.Code,
			SourceDecision: metricsPefindo.SourceDecision,
			Reason:         metricsPefindo.Reason,
		}

		details = append(details, addDetail)

		resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsPefindo.Reason)
		if err != nil {
			return
		}

		return
	}

	details = append(details, entity.TrxDetail{
		ProspectID:     reqMetrics.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       metricsPefindo.Code,
		SourceDecision: metricsPefindo.SourceDecision,
		Reason:         metricsPefindo.Reason,
		NextStep:       constant.SOURCE_DECISION_SCOREPRO,
	})

	var scoreBiro string
	if filtering.ScoreBiro != nil {
		scoreBiro = filtering.ScoreBiro.(string)
	}

	responseScs, metricsScs, pefindoIDX, err := u.usecase.Scorepro(ctx, reqMetrics, scoreBiro, customerSegment, dupcheckData, accessToken, filtering)
	if err != nil {
		return
	}

	trxFMF.ScsDecision = response.ScsDecision{
		ScsDate:   time.Now().Format("2006-01-02"),
		ScsStatus: metricsScs.Result,
		ScsScore:  responseScs.ScoreResult,
	}

	// handling flow deviasi scorepro
	if metricsScs.IsDeviasi {

		// insert deviasi
		trxFMF.TrxDeviasi = entity.TrxDeviasi{
			ProspectID: reqMetrics.Transaction.ProspectID,
			DeviasiID:  constant.CODE_DEVIASI_SCOREPRO,
			Reason:     metricsScs.Reason,
		}

		// detail scorepro
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       metricsScs.Code,
			SourceDecision: metricsScs.Source,
			Info:           metricsScs.Info,
			Reason:         responseScs.ScoreResult,
			NextStep:       constant.SOURCE_DECISION_DEVIASI,
		})

		// detail deviasi
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       constant.RULE_CODE_DEVIASI_SCOREPRO,
			SourceDecision: constant.SOURCE_DECISION_DEVIASI,
			Reason:         metricsScs.Reason,
			NextStep:       constant.SOURCE_DECISION_CA,
		})

		// detail ca
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_UNPROCESS,
			Decision:       constant.DB_DECISION_CREDIT_PROCESS,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.SOURCE_DECISION_CA,
		})

		resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsScs.Reason)
		if err != nil {
			return
		}

		return

	} else {
		if metricsScs.Result == constant.DECISION_REJECT {

			addDetail := entity.TrxDetail{
				ProspectID:     reqMetrics.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				RuleCode:       metricsScs.Code,
				SourceDecision: metricsScs.Source,
				Info:           metricsScs.Info,
				Reason:         responseScs.ScoreResult,
			}

			details = append(details, addDetail)

			resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsScs.Reason)
			if err != nil {
				return
			}

			return
		}

		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       metricsScs.Code,
			SourceDecision: metricsScs.Source,
			Info:           metricsScs.Info,
			Reason:         responseScs.ScoreResult,
			NextStep:       constant.SOURCE_DECISION_DSR,
		})
	}

	var totalInstallmentPBK float64
	if filtering.TotalInstallmentAmountBiro != nil {
		totalInstallmentPBK, err = utils.GetFloat(filtering.TotalInstallmentAmountBiro)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat TotalInstallmentAmountBiro Error")
			return
		}
	}

	income := reqDupcheck.MonthlyFixedIncome + reqDupcheck.MonthlyVariableIncome + reqDupcheck.SpouseIncome
	metricsTotalDsrFmfPbk, trxFMFTotalDsrFmfPbk, err := u.usecase.TotalDsrFmfPbk(ctx, income, reqMetrics.Apk.InstallmentAmount, totalInstallmentPBK, reqMetrics.Transaction.ProspectID, customerSegment, accessToken, dupcheckData, configValue, filtering, reqMetrics.Apk.NTF)
	if err != nil {
		return
	}

	trxFMF.DSRPBK = trxFMFTotalDsrFmfPbk.DSRPBK
	trxFMF.TotalDSR = trxFMFTotalDsrFmfPbk.TotalDSR

	infoTotalDSR, _ := json.Marshal(map[string]interface{}{
		"dsr_fmf":                   trxFMF.DSRFMF,
		"dsr_pbk":                   trxFMF.DSRPBK,
		"total_dsr":                 trxFMF.TotalDSR,
		"installment_threshold":     trxFMFTotalDsrFmfPbk.InstallmentThreshold,
		"latest_installment_amount": trxFMFTotalDsrFmfPbk.LatestInstallmentAmount,
	})

	if trxFMFTotalDsrFmfPbk.LatestInstallmentAmount > 0 {
		trxFMF.LatestInstallmentAmount = trxFMFTotalDsrFmfPbk.LatestInstallmentAmount
	}

	if metricsTotalDsrFmfPbk.IsDeviasi {
		// insert deviasi
		trxFMF.TrxDeviasi = entity.TrxDeviasi{
			ProspectID: reqMetrics.Transaction.ProspectID,
			DeviasiID:  constant.CODE_DEVIASI_DSR,
			Reason:     metricsTotalDsrFmfPbk.Reason,
		}

		// detail TotalDsrFmfPbk
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       metricsTotalDsrFmfPbk.Code,
			SourceDecision: metricsTotalDsrFmfPbk.SourceDecision,
			Info:           string(utils.SafeEncoding(infoTotalDSR)),
			Reason:         metricsTotalDsrFmfPbk.Reason,
			NextStep:       constant.SOURCE_DECISION_DEVIASI,
		})

		// detail deviasi
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       constant.RULE_CODE_DEVIASI_DSR,
			SourceDecision: constant.SOURCE_DECISION_DEVIASI,
			Reason:         metricsTotalDsrFmfPbk.Reason,
			NextStep:       constant.SOURCE_DECISION_CA,
		})

		// detail ca
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_UNPROCESS,
			Decision:       constant.DB_DECISION_CREDIT_PROCESS,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.SOURCE_DECISION_CA,
		})

		resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsTotalDsrFmfPbk.Reason)
		if err != nil {
			return
		}

		return
	} else {
		if metricsTotalDsrFmfPbk.Result == constant.DECISION_REJECT {

			addDetail := entity.TrxDetail{
				ProspectID:     reqMetrics.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				RuleCode:       metricsTotalDsrFmfPbk.Code,
				SourceDecision: metricsTotalDsrFmfPbk.SourceDecision,
				Info:           string(utils.SafeEncoding(infoTotalDSR)),
				Reason:         metricsTotalDsrFmfPbk.Reason,
			}

			details = append(details, addDetail)

			resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsTotalDsrFmfPbk.Reason)
			if err != nil {
				return
			}

			return
		}

		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       metricsTotalDsrFmfPbk.Code,
			SourceDecision: metricsTotalDsrFmfPbk.SourceDecision,
			Info:           string(utils.SafeEncoding(infoTotalDSR)),
			Reason:         metricsTotalDsrFmfPbk.Reason,
			NextStep:       constant.SOURCE_DECISION_ELABORATE_LTV,
		})
	}

	metricsElaborateScheme, err := u.usecase.ElaborateScheme(reqMetrics)
	if err != nil {
		return
	}

	if metricsElaborateScheme.Result == constant.DECISION_REJECT {

		addDetail := entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       metricsElaborateScheme.Code,
			SourceDecision: metricsElaborateScheme.SourceDecision,
			Reason:         metricsElaborateScheme.Reason,
		}

		details = append(details, addDetail)

		resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsElaborateScheme.Reason)
		if err != nil {
			return
		}

		return
	}

	var metricsElaborateIncome response.UsecaseApi
	if cbFound {
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       metricsElaborateScheme.Code,
			SourceDecision: metricsElaborateScheme.SourceDecision,
			Reason:         metricsElaborateScheme.Reason,
			NextStep:       constant.SOURCE_DECISION_ELABORATE_INCOME,
		})

		metricsElaborateIncome, err = u.usecase.ElaborateIncome(ctx, reqMetrics, filtering, pefindoIDX, dupcheckData, responseScs, accessToken)
		if err != nil {
			return
		}

		if metricsElaborateIncome.Result == constant.DECISION_REJECT {

			addDetail := entity.TrxDetail{
				ProspectID:     reqMetrics.Transaction.ProspectID,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				RuleCode:       metricsElaborateIncome.Code,
				SourceDecision: metricsElaborateIncome.SourceDecision,
				Reason:         metricsElaborateIncome.Reason,
			}

			details = append(details, addDetail)

			resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, metricsElaborateIncome.Reason)
			if err != nil {
				return
			}

			return
		}

		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       metricsElaborateIncome.Code,
			SourceDecision: metricsElaborateIncome.SourceDecision,
			Reason:         metricsElaborateIncome.Reason,
			NextStep:       constant.SOURCE_DECISION_CA,
		})
	} else {
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			RuleCode:       metricsElaborateScheme.Code,
			SourceDecision: metricsElaborateScheme.SourceDecision,
			Reason:         metricsElaborateScheme.Reason,
			NextStep:       constant.SOURCE_DECISION_CA,
		})
	}

	finalReasonMetrics := metricsDupcheck.Reason

	if metricsElaborateIncome.Reason != "" {
		finalReasonMetrics = fmt.Sprintf("%s - %s", finalReasonMetrics, metricsElaborateIncome.Reason)
	} else {
		finalReasonMetrics = fmt.Sprintf("%s - %s", finalReasonMetrics, metricsElaborateScheme.Reason)
	}

	if countTrx == 0 && trxFMF.NTFAkumulasi <= 20000000 {
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_APR,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.SOURCE_DECISION_CA,
		})

		trxFMF.TrxCaDecision = entity.TrxCaDecision{
			ProspectID: reqMetrics.Transaction.ProspectID,
			Decision:   constant.DB_DECISION_APR,
			CreatedBy:  constant.SYSTEM_CREATED,
			DecisionBy: constant.SYSTEM_CREATED,
		}

		finalReasonMetrics = fmt.Sprintf("%s - PBK PASS - NTF <= Quick App", metricsDupcheck.Reason)
	} else {
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_UNPROCESS,
			Decision:       constant.DB_DECISION_CREDIT_PROCESS,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.SOURCE_DECISION_CA,
		})
	}

	resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, finalReasonMetrics)
	if err != nil {
		return
	}

	return
}

func (u usecase) SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (resp response.Metrics, err error) {

	var decision string

	err = u.repository.SaveTransaction(countTrx, data, trxPrescreening, trxFMF, details, reason)

	detail := details[len(details)-1]

	switch detail.Decision {

	case constant.DB_DECISION_APR:
		decision = constant.DECISION_APPROVE

	case constant.DB_DECISION_PASS:
		decision = constant.JSON_DECISION_PASS

	case constant.DB_DECISION_REJECT:
		decision = constant.JSON_DECISION_REJECT

	case constant.DB_DECISION_CREDIT_PROCESS:
		decision = constant.JSON_DECISION_CREDIT_PROCESS
	}

	resp = response.Metrics{ProspectID: detail.ProspectID, Code: detail.RuleCode, Decision: decision, DecisionReason: reason}
	return
}
