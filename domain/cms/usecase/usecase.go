package usecase

import (
	"context"
	"encoding/json"
	"errors"
	cache "los-kmb-api/domain/cache/interfaces"
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"strings"
)

type (
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		cache      cache.Repository
	}
)

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, cache cache.Repository) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
		cache:      cache,
	}
}

func (u usecase) GetReasonPrescreening(ctx context.Context, req request.ReqReasonPrescreening, pagination interface{}) (data []entity.ReasonMessage, rowTotal int, err error) {

	data, rowTotal, err = u.repository.GetReasonPrescreening(req.ReasonID, pagination)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetInquiryPrescreening(ctx context.Context, req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryData, rowTotal int, err error) {

	var (
		industry           []entity.SpIndustryTypeMaster
		photos             []entity.CustomerPhoto
		surveyor           []entity.TrxSurveyor
		action             bool
		cmo_recommendation string
		decision           string
	)

	getValue, _ := u.cache.Get("GetSpIndustryTypeMaster")

	if getValue == nil {
		industry, err = u.repository.GetSpIndustryTypeMaster()

		if err != nil {
			return
		}

		u.cache.Set("GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		for _, description := range industry {
			u.cache.Set(strings.ReplaceAll(description.IndustryTypeID, " ", ""), []byte(description.Description))
		}
	}

	// get inquiry pre screening
	result, rowTotal, err := u.repository.GetInquiryPrescreening(req, pagination)

	if err != nil {
		return []entity.InquiryData{}, 0, err
	}

	for _, inq := range result {

		industry_type, _ := u.cache.Get(inq.IndustryTypeID)

		// get trx_customer_photo
		photos, err = u.repository.GetCustomerPhoto(inq.ProspectID)

		if err != nil {
			return
		}

		var photoData []entity.CustomerPhoto

		if len(photos) > 0 {
			for _, photo := range photos {
				photoEntry := entity.CustomerPhoto{
					PhotoID: photo.PhotoID,
					Url:     photo.Url,
				}
				photoData = append(photoData, photoEntry)
			}
		}

		if len(photoData) < 1 {
			photoData = []entity.CustomerPhoto{}
		}

		// get trx_surveyor
		surveyor, err = u.repository.GetSurveyorData(inq.ProspectID)

		if err != nil {
			return
		}

		var surveyorData []entity.TrxSurveyor

		if len(surveyor) > 0 {
			for _, survey := range surveyor {
				surveyorEntry := entity.TrxSurveyor{
					Destination:  survey.Destination,
					RequestDate:  survey.RequestDate,
					AssignDate:   survey.AssignDate,
					SurveyorName: survey.SurveyorName,
					SurveyorNote: survey.SurveyorNote,
					ResultDate:   survey.ResultDate,
					Status:       survey.Status,
				}
				surveyorData = append(surveyorData, surveyorEntry)
			}
		}

		if len(surveyorData) < 1 {
			surveyorData = []entity.TrxSurveyor{}
		}

		action = false
		if inq.Activity == constant.ACTIVITY_UNPROCESS && inq.SourceDecision == constant.PRESCREENING {
			action = true
		}
		if inq.CmoRecommendation == 1 {
			cmo_recommendation = "Recommended"
		} else {
			cmo_recommendation = "Not Recommended"
		}

		decision = ""
		if inq.Decision == constant.DB_DECISION_APR {
			decision = "Sesuai"
		} else if inq.Decision == constant.DB_DECISION_REJECT {
			decision = "Tidak Sesuai"
		}

		row := entity.InquiryData{
			Prescreening: entity.DataPrescreening{
				CmoRecommendation: cmo_recommendation,
				ShowAction:        action,
				Decision:          decision,
				Reason:            inq.Reason,
				DecisionBy:        inq.DecisionBy,
				DecisionName:      inq.DecisionName,
				DecisionAt:        inq.DecisionAt,
			},
			General: entity.DataGeneral{
				ProspectID:     inq.ProspectID,
				BranchName:     inq.BranchName,
				IncomingSource: inq.IncomingSource,
				CreatedAt:      inq.CreatedAt,
			},
			Personal: entity.CustomerPersonal{
				IDNumber:          inq.IDNumber,
				LegalName:         inq.LegalName,
				CustomerStatus:    inq.CustomerStatus,
				BirthPlace:        inq.BirthPlace,
				BirthDate:         inq.BirthDate,
				SurgateMotherName: inq.SurgateMotherName,
				Gender:            inq.Gender,
				MobilePhone:       inq.MobilePhone,
				Email:             inq.Email,
				NumOfDependence:   inq.NumOfDependence,
				StaySinceYear:     inq.StaySinceYear,
				StaySinceMonth:    inq.StaySinceMonth,
				ExtCompanyPhone:   inq.ExtCompanyPhone,
				SourceOtherIncome: inq.SourceOtherIncome,
				Education:         inq.Education,
				MaritalStatus:     inq.MaritalStatus,
				HomeStatus:        inq.HomeStatus,
			},
			Spouse: entity.CustomerSpouse{
				IDNumber:     inq.SpouseIDNumber,
				LegalName:    inq.SpouseLegalName,
				CompanyName:  inq.SpouseCompanyName,
				CompanyPhone: inq.SpouseCompanyPhone,
				MobilePhone:  inq.SpouseMobilePhone,
				ProfessionID: inq.SpouseProfession,
			},
			Employment: entity.CustomerEmployment{
				EmploymentSinceMonth:  inq.EmploymentSinceMonth,
				EmploymentSinceYear:   inq.EmploymentSinceYear,
				CompanyName:           inq.CompanyName,
				MonthlyFixedIncome:    inq.MonthlyFixedIncome,
				MonthlyVariableIncome: inq.MonthlyVariableIncome,
				SpouseIncome:          inq.SpouseIncome,
				ProfessionID:          inq.ProfessionID,
				JobType:               inq.JobTypeID,
				JobPosition:           inq.JobPosition,
				IndustryTypeID:        strings.TrimSpace(string(industry_type)),
			},
			ItemApk: entity.DataItemApk{
				Supplier:              inq.Supplier,
				ProductOfferingID:     inq.ProductOfferingID,
				AssetDescription:      inq.AssetDescription,
				AssetType:             inq.AssetType,
				ManufacturingYear:     inq.ManufacturingYear,
				Color:                 inq.Color,
				ChassisNumber:         inq.ChassisNumber,
				EngineNumber:          inq.EngineNumber,
				InterestRate:          inq.InterestRate,
				InsuranceRate:         inq.InsuranceRate,
				Tenor:                 inq.InstallmentPeriod,
				OTR:                   inq.OTR,
				DPAmount:              inq.DPAmount,
				AF:                    inq.FinanceAmount,
				NTF:                   inq.NTF,
				NTFPlusInterestAmount: inq.Total,
				InterestAmount:        inq.InterestAmount,
				InsuranceAmount:       inq.InsuranceAmount,
				InstallmentAmount:     inq.MonthlyInstallment,
				AdminFee:              inq.AdminFee,
				ProvisionFee:          inq.ProvisionFee,
				FirstPayment:          inq.FirstPayment,
				FirstInstallment:      inq.FirstInstallment,
				FirstPaymentDate:      inq.FirstPaymentDate,
			},
			Surveyor: surveyorData,
			Emcon: entity.CustomerEmcon{
				Name:         inq.EmconName,
				Relationship: inq.Relationship,
				MobilePhone:  inq.EmconMobilePhone,
			},
			Address: entity.DataAddress{
				LegalAddress:       inq.LegalAddress,
				LegalRTRW:          inq.LegalRTRW,
				LegalKelurahan:     inq.LegalKelurahan,
				LegalKecamatan:     inq.LegalKecamatan,
				LegalZipCode:       inq.LegalZipCode,
				LegalCity:          inq.LegalCity,
				ResidenceAddress:   inq.ResidenceAddress,
				ResidenceRTRW:      inq.ResidenceRTRW,
				ResidenceKelurahan: inq.ResidenceKelurahan,
				ResidenceKecamatan: inq.ResidenceKecamatan,
				ResidenceZipCode:   inq.ResidenceZipCode,
				ResidenceCity:      inq.ResidenceCity,
				CompanyAddress:     inq.CompanyAddress,
				CompanyRTRW:        inq.CompanyRTRW,
				CompanyKelurahan:   inq.CompanyKelurahan,
				CompanyKecamatan:   inq.CompanyKecamatan,
				CompanyZipCode:     inq.CompanyZipCode,
				CompanyCity:        inq.CompanyCity,
				CompanyAreaPhone:   inq.CompanyAreaPhone,
				CompanyPhone:       inq.CompanyPhone,
				EmergencyAddress:   inq.EmergencyAddress,
				EmergencyRTRW:      inq.EmergencyRTRW,
				EmergencyKelurahan: inq.EmergencyKelurahan,
				EmergencyKecamatan: inq.EmergencyKecamatan,
				EmergencyZipcode:   inq.EmergencyZipcode,
				EmergencyCity:      inq.EmergencyCity,
				EmergencyAreaPhone: inq.EmergencyAreaPhone,
				EmergencyPhone:     inq.EmergencyPhone,
			},
			Photo: photoData,
		}

		data = append(data, row)

	}

	return
}

func (u usecase) ReviewPrescreening(ctx context.Context, req request.ReqReviewPrescreening) (data response.ReviewPrescreening, err error) {

	var (
		trxStatus entity.TrxStatus
		reason    = string(req.Reason)
	)

	status, err := u.repository.GetStatusPrescreening(req.ProspectID)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get status prescreening error")
		return
	}

	// Bisa melakukan review jika status UNPR dan source_decision PRESCREENING
	if status.Activity == constant.ACTIVITY_UNPROCESS && status.SourceDecision == constant.PRESCREENING {

		decisionMapping := map[string]struct {
			Code           int
			StatusProcess  string
			Activity       string
			Decision       string
			DecisionDetail string
			DecisionStatus string
			ActivityStatus string
			NextStep       interface{}
			SourceDecision string
		}{
			constant.DECISION_REJECT: {
				Code:           constant.CODE_REJECT_PRESCREENING,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				DecisionStatus: constant.DB_DECISION_REJECT,
				DecisionDetail: constant.DB_DECISION_REJECT,
				ActivityStatus: constant.ACTIVITY_STOP,
				SourceDecision: constant.PRESCREENING,
			},
			constant.DECISION_APPROVE: {
				Code:           constant.CODE_PASS_PRESCREENING,
				StatusProcess:  constant.STATUS_ONPROCESS,
				Activity:       constant.ACTIVITY_PROCESS,
				Decision:       constant.DB_DECISION_APR,
				DecisionStatus: constant.DB_DECISION_CREDIT_PROCESS,
				DecisionDetail: constant.DB_DECISION_PASS,
				ActivityStatus: constant.ACTIVITY_UNPROCESS,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			},
		}

		decisionInfo, ok := decisionMapping[req.Decision]
		if !ok {
			err = errors.New(constant.ERROR_UPSTREAM + " - Decision tidak valid")
			return
		}

		data.ProspectID = req.ProspectID
		data.Code = decisionInfo.Code
		data.Decision = decisionInfo.Decision
		data.Reason = reason

		info, _ := json.Marshal(data)

		trxPrescreening := entity.TrxPrescreening{
			ProspectID:   req.ProspectID,
			Decision:     decisionInfo.Decision,
			Reason:       reason,
			CreatedBy:    req.DecisionBy,
			DecisionByBy: req.DecisionByName,
		}

		trxDetail := entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  decisionInfo.StatusProcess,
			Activity:       decisionInfo.Activity,
			Decision:       decisionInfo.DecisionDetail,
			RuleCode:       decisionInfo.Code,
			SourceDecision: constant.PRESCREENING,
			NextStep:       decisionInfo.NextStep,
			Info:           string(info),
			CreatedBy:      req.DecisionBy,
		}

		if req.Decision == constant.DECISION_REJECT {
			trxStatus.RuleCode = decisionInfo.Code
			trxStatus.Reason = reason
		}

		trxStatus.ProspectID = req.ProspectID
		trxStatus.StatusProcess = decisionInfo.StatusProcess
		trxStatus.Activity = decisionInfo.ActivityStatus
		trxStatus.Decision = decisionInfo.DecisionStatus
		trxStatus.SourceDecision = decisionInfo.SourceDecision

		err = u.repository.SavePrescreening(trxPrescreening, trxDetail, trxStatus)
		if err != nil {
			return
		}
	} else {
		err = errors.New(constant.ERROR_UPSTREAM + " - Status order tidak dalam prescreening")
		return
	}

	return
}
