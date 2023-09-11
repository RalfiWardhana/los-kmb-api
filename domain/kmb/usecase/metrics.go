package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"
)

func (u metrics) MetricsLos(ctx context.Context, reqMetrics request.Metrics, accessToken string) (resultMetrics interface{}, err error) {

	var (
		married         bool
		details         []entity.TrxDetail
		reqDupcheck     request.DupcheckApi
		dupcheckData    response.SpDupcheckMap
		customerStatus  string
		decisionMetrics response.UsecaseApi
		additionalTrx   response.Additional
		filtering       entity.FilteringKMB
		trxPrescreening entity.TrxPrescreening
		trxFMF          response.TrxFMF
	)

	// cek filtering
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

	// cek trx_master
	var trxMaster int
	trxMaster, err = u.repository.ScanTrxMaster(reqMetrics.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error")
		return
	}

	// Order ID already have final decision
	if trxMaster > 0 {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID Already Exist")
		return
	}

	//cek prescreening
	var trxPrescreeningDetail entity.TrxDetail
	trxPrescreening, trxFMF, trxPrescreeningDetail, err = u.usecase.Prescreening(ctx, reqMetrics, filtering, accessToken)
	if err != nil {
		return
	}

	details = append(details, trxPrescreeningDetail)

	if trxPrescreening.Decision != constant.DB_DECISION_APR {
		resultMetrics, err = u.usecase.SaveTransaction(reqMetrics, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
		if err != nil {
			return
		}

		return
	} else {
		trxDetail := entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_UNPROCESS,
			Decision:       constant.DB_DECISION_CREDIT_PROCESS,
			SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
			NextStep:       constant.SOURCE_DECISION_BIRO,
			CreatedBy:      constant.SYSTEM_CREATED,
		}

		details = append(details, trxDetail)
		resultMetrics, err = u.usecase.SaveTransaction(reqMetrics, trxPrescreening, trxFMF, details, trxPrescreening.Reason)
		if err != nil {
			return
		}

		return
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

	reqDupcheck = request.DupcheckApi{
		ProspectID:            reqMetrics.Transaction.ProspectID,
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
		OTRPrice:              reqMetrics.Apk.OTR,
		NTF:                   reqMetrics.Apk.NTF,
		LegalZipCode:          legalZipCode,
		CompanyZipCode:        companyZipCode,
		Gender:                reqMetrics.CustomerPersonal.Gender,
		InstallmentAmount:     reqMetrics.Apk.InstallmentAmount,
		MaritalStatus:         reqMetrics.CustomerPersonal.MaritalStatus,
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

	dupcheckData, customerStatus, decisionMetrics, err = u.multiUsecase.Dupcheck(ctx, reqDupcheck, married, accessToken)

	if err != nil {
		return
	}

	additionalTrx = response.Additional{
		DupcheckData:   dupcheckData,
		CustomerStatus: customerStatus,
	}

	if decisionMetrics.Result == constant.DECISION_REJECT {
		details = append(details, entity.TrxDetail{
			ProspectID:     reqMetrics.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_REJECT,
			RuleCode:       decisionMetrics.Code,
			SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
			Info:           decisionMetrics.Reason,
		})

		// resultMetrics, err = u.usecase.SaveTransaction(details, reason, callback, req, additionalTrx)
		// if err != nil {
		// 	return
		// }

		// return
	}

	details = append(details, entity.TrxDetail{
		ProspectID:     reqMetrics.Transaction.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       decisionMetrics.Code,
		SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
		Info:           decisionMetrics.Reason,
		NextStep:       constant.SOURCE_DECISION_BIRO,
	})

	//Get data filtering where DtmResponse < BIRO_VALID_DAYS
	filtering, err = u.repository.GetBiroData(reqMetrics.Transaction.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Filtering > " + os.Getenv("BIRO_VALID_DAYS") + " Days")
		return
	}

	log.Println(dupcheckData)
	log.Println(customerStatus)
	log.Println(decisionMetrics)
	log.Println(additionalTrx)
	log.Println(filtering)

	resultMetrics = details

	return
}

func (u usecase) SaveTransaction(data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (resp response.Metrics, err error) {

	var decision string

	err = u.repository.SaveTransaction(data, trxPrescreening, trxFMF, details, reason)

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
