package usecase

import (
	"context"
	"errors"
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
	)

	// // Scan Order ID must have data in filtering before request to Metrics
	// rowMaster, rowFtr, err := u.repository.ScanPreTrxJourney(reqMetrics.Transaction.ProspectID)

	// // Order ID not found at filtering kmob
	// if rowFtr == 0 {
	// 	err = errors.New(constant.ERROR_BAD_REQUEST + " - Filtering Data Not Found")
	// 	return
	// }

	// // Order ID already have final decision
	// if rowMaster > 0 {
	// 	err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID Already Exist")
	// 	return
	// }

	// if err != nil {
	// 	err = errors.New(constant.ERROR_UPSTREAM + " - PreTrxJourney Error")
	// 	return
	// }

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
