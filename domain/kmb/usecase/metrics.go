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

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

// func ini digunakan ketika Submit to LOS
// dan ketika Approve prescreening di menu prescreening oleh CA
func (u metrics) MetricsLos(ctx context.Context, reqMetrics request.Metrics, accessToken string) (resultMetrics interface{}, err error) {

	var (
		married           bool
		details           []entity.TrxDetail
		reqDupcheck       request.DupcheckApi
		dupcheckData      response.SpDupcheckMap
		customerStatus    string
		customerSegment   string
		decisionMetrics   response.UsecaseApi
		filtering         entity.FilteringKMB
		trxPrescreening   entity.TrxPrescreening
		trxFMF            response.TrxFMF
		trxFMFDupcheck    response.TrxFMF
		trxDetailDupcheck []entity.TrxDetail
		cbFound           bool
	)

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
				Info:           constant.REASON_CMO_NOT_RECOMMENDED,
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
				Info:           constant.REASON_CMO_NOT_RECOMMENDED,
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
			Info:           constant.REASON_CMO_RECOMMENDED,
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

	//  STEP 3 tenor 36
	if reqMetrics.Apk.Tenor >= 36 {
		var trxTenor response.UsecaseApi
		trxTenor, err = u.usecase.RejectTenor36(reqMetrics.CustomerPersonal.IDNumber)
		if err != nil {
			return
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
				Info:           trxTenor.Reason,
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

	reqDupcheck = request.DupcheckApi{
		ProspectID:            reqMetrics.Transaction.ProspectID,
		BranchID:              reqMetrics.Transaction.BranchID,
		ImageKtp:              ktpImage,
		ImageSelfie:           selfieImage,
		MonthlyFixedIncome:    reqMetrics.CustomerEmployment.MonthlyFixedIncome,
		HomeStatus:            reqMetrics.CustomerPersonal.HomeStatus,
		MonthlyVariableIncome: *reqMetrics.CustomerEmployment.MonthlyVariableIncome,
		SpouseIncome:          *reqMetrics.CustomerEmployment.SpouseIncome,
		JobPosition:           reqMetrics.CustomerEmployment.JobPosition,
		ProfessionID:          reqMetrics.CustomerEmployment.ProfessionID,
		EmploymentSinceYear:   reqMetrics.CustomerEmployment.EmploymentSinceYear,
		EmploymentSinceMonth:  reqMetrics.CustomerEmployment.EmploymentSinceMonth,
		StaySinceYear:         reqMetrics.CustomerPersonal.StaySinceYear,
		StaySinceMonth:        reqMetrics.CustomerPersonal.StaySinceMonth,
		BirthDate:             reqMetrics.CustomerPersonal.BirthDate,
		BirthPlace:            reqDupcheck.BirthPlace,
		Tenor:                 reqMetrics.Apk.Tenor,
		IDNumber:              reqMetrics.CustomerPersonal.IDNumber,
		LegalName:             reqMetrics.CustomerPersonal.LegalName,
		MotherName:            reqMetrics.CustomerPersonal.SurgateMotherName,
		EngineNo:              reqMetrics.Item.NoEngine,
		RangkaNo:              reqMetrics.Item.NoChassis,
		ManufactureYear:       reqMetrics.Item.ManufactureYear,
		BPKBName:              reqMetrics.Item.BPKBName,
		NumOfDependence:       reqDupcheck.NumOfDependence,
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

	dupcheckData, customerStatus, decisionMetrics, trxFMFDupcheck, trxDetailDupcheck, err = u.multiUsecase.Dupcheck(ctx, reqDupcheck, married, accessToken)
	if err != nil {
		return
	}

	trxFMF.DupcheckData = dupcheckData
	trxFMF.CustomerStatus = customerStatus
	trxFMF.DSRFMF = dupcheckData.Dsr
	trxFMF.TrxBannedPMKDSR = trxFMFDupcheck.TrxBannedPMKDSR
	trxFMF.TrxBannedChassisNumber = trxFMFDupcheck.TrxBannedChassisNumber

	// internal record
	if dupcheckData.CustomerID != nil {
		var resInternalRecord *resty.Response
		resInternalRecord, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("INTERNAL_RECORD_URL")+dupcheckData.CustomerID.(string), nil, map[string]string{}, constant.METHOD_GET, true, 3, 60, reqMetrics.Transaction.ProspectID, accessToken)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Interal Record Error")
			return
		}

		err = json.Unmarshal([]byte(jsoniter.Get(resInternalRecord.Body(), "data").ToString()), &trxFMF.AgreementCONFINS)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Unmarshal Interal Record Error")
			return
		}
	}

	if decisionMetrics.Result == constant.DECISION_REJECT {
		details = append(details, trxDetailDupcheck...)

		addDetail := entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       decisionMetrics.Code,
			SourceDecision: decisionMetrics.SourceDecision,
			Info:           decisionMetrics.Reason,
		}

		if decisionMetrics.SourceDecision == constant.SOURCE_DECISION_DSR || decisionMetrics.SourceDecision == constant.SOURCE_DECISION_DUPCHECK {
			info, _ := json.Marshal(dupcheckData)
			addDetail.Info = string(utils.SafeEncoding(info))
		}

		details = append(details, addDetail)

		resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, decisionMetrics.Reason)
		if err != nil {
			return
		}

		return
	}

	details = append(details, trxDetailDupcheck...)

	details = append(details, entity.TrxDetail{
		ProspectID:     reqMetrics.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       decisionMetrics.Code,
		SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
		Info:           decisionMetrics.Reason,
		NextStep:       constant.SOURCE_DECISION_DUKCAPIL,
	})

	decisionEkyc, trxDetailEkyc, err := u.multiUsecase.Ekyc(ctx, reqMetrics, cbFound, accessToken)
	if err != nil {
		return
	}

	if len(trxDetailEkyc) > 0 {
		details = append(details, trxDetailEkyc...)
	}

	if decisionEkyc.Result == constant.DECISION_REJECT {

		addDetail := entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       decisionEkyc.Code,
			SourceDecision: decisionEkyc.Source,
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
		NextStep:       constant.SOURCE_DECISION_CA,
	})

	details = append(details, entity.TrxDetail{
		ProspectID:     reqMetrics.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.SOURCE_DECISION_CA,
	})

	resultMetrics, err = u.usecase.SaveTransaction(countTrx, reqMetrics, trxPrescreening, trxFMF, details, decisionMetrics.Reason)
	if err != nil {
		return
	}

	// log.Println(dupcheckData)
	// log.Println(customerStatus)
	// log.Println(decisionMetrics)
	// log.Println(filtering)

	return
}

func (u usecase) SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (resp response.Metrics, err error) {

	var decision string

	err = u.repository.SaveTransaction(countTrx, data, trxPrescreening, trxFMF, details, reason)

	detail := details[len(details)-1]

	switch detail.Decision {

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
