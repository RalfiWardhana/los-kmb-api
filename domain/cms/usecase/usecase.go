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

	data, rowTotal, err = u.repository.GetReasonPrescreening(req, pagination)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetInquiryPrescreening(ctx context.Context, req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryData, rowTotal int, err error) {

	var (
		industry          []entity.SpIndustryTypeMaster
		photos            []entity.DataPhoto
		surveyor          []entity.TrxSurveyor
		action            bool
		cmoRecommendation string
		decision          string
	)

	// get inquiry pre screening
	result, rowTotal, err := u.repository.GetInquiryPrescreening(req, pagination)

	if err != nil {
		return []entity.InquiryData{}, 0, err
	}

	for _, inq := range result {

		industryType, _ := u.cache.Get(inq.IndustryTypeID)

		if industryType == nil {
			industry, err = u.repository.GetSpIndustryTypeMaster()

			if err != nil {
				return
			}

			for _, description := range industry {
				u.cache.Set(strings.ReplaceAll(description.IndustryTypeID, " ", ""), []byte(description.Description))
			}
		}

		// get trx_customer_photo
		photos, err = u.repository.GetCustomerPhoto(inq.ProspectID)

		if err != nil {
			return
		}

		var photoData []entity.DataPhoto

		if len(photos) > 0 {
			for _, photo := range photos {
				photoEntry := entity.DataPhoto{
					PhotoID: photo.PhotoID,
					Label:   photo.Label,
					Url:     photo.Url,
				}
				photoData = append(photoData, photoEntry)
			}
		}

		if len(photoData) < 1 {
			photoData = []entity.DataPhoto{}
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
			cmoRecommendation = "Recommended"
		} else {
			cmoRecommendation = "Not Recommended"
		}

		decision = ""
		if inq.Decision == constant.DB_DECISION_APR {
			decision = "Sesuai"
		} else if inq.Decision == constant.DB_DECISION_REJECT {
			decision = "Tidak Sesuai"
		}

		row := entity.InquiryData{
			Prescreening: entity.DataPrescreening{
				CmoRecommendation: cmoRecommendation,
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
				OrderAt:        inq.OrderAt,
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
				IndustryTypeID:        strings.TrimSpace(string(industryType)),
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
				Tenor:                 inq.InstallmentPeriod,
				OTR:                   inq.OTR,
				DPAmount:              inq.DPAmount,
				AF:                    inq.FinanceAmount,
				NTF:                   inq.NTF,
				NTFAkumulasi:          inq.NTFAkumulasi,
				NTFPlusInterestAmount: inq.Total,
				InterestAmount:        inq.InterestAmount,
				LifeInsuranceFee:      inq.LifeInsuranceFee,
				AssetInsuranceFee:     inq.AssetInsuranceFee,
				InsuranceAmount:       inq.InsuranceAmount,
				InstallmentAmount:     inq.MonthlyInstallment,
				AdminFee:              inq.AdminFee,
				ProvisionFee:          inq.ProvisionFee,
				FirstInstallment:      inq.FirstInstallment,
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
			ProspectID: req.ProspectID,
			Decision:   decisionInfo.Decision,
			Reason:     reason,
			CreatedBy:  req.DecisionBy,
			DecisionBy: req.DecisionByName,
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

func (u usecase) GetInquiryCa(ctx context.Context, req request.ReqInquiryCa, pagination interface{}) (data []entity.InquiryDataCa, rowTotal int, err error) {

	var (
		industry       []entity.SpIndustryTypeMaster
		photos         []entity.DataPhoto
		surveyor       []entity.TrxSurveyor
		histories      []entity.HistoryApproval
		internalRecord []entity.TrxInternalRecord
	)

	// get inquiry pre screening
	result, rowTotal, err := u.repository.GetInquiryCa(req, pagination)

	if err != nil {
		return []entity.InquiryDataCa{}, 0, err
	}

	for _, inq := range result {

		industryType, _ := u.cache.Get(inq.IndustryTypeID)

		if industryType == nil {
			industry, err = u.repository.GetSpIndustryTypeMaster()

			if err != nil {
				return
			}

			for _, description := range industry {
				u.cache.Set(strings.ReplaceAll(description.IndustryTypeID, " ", ""), []byte(description.Description))
			}
		}

		// get trx_customer_photo
		photos, err = u.repository.GetCustomerPhoto(inq.ProspectID)

		if err != nil {
			return
		}

		var photoData []entity.DataPhoto

		if len(photos) > 0 {
			for _, photo := range photos {
				photoEntry := entity.DataPhoto{
					PhotoID: photo.PhotoID,
					Label:   photo.Label,
					Url:     photo.Url,
				}
				photoData = append(photoData, photoEntry)
			}
		}

		if len(photoData) < 1 {
			photoData = []entity.DataPhoto{}
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

		// get trx_history_approval
		histories, err = u.repository.GetHistoryApproval(inq.ProspectID)

		if err != nil {
			return
		}

		var (
			historyData []entity.HistoryApproval
			escalation  string
		)

		if len(histories) > 0 {
			for _, history := range histories {
				escalation = "No"
				if history.NeedEscalation.(int64) == 1 {
					escalation = "Yes"
				}
				historyEntry := entity.HistoryApproval{
					DecisionBy:            history.DecisionBy,
					Decision:              history.Decision,
					CreatedAt:             history.CreatedAt,
					NextFinalApprovalFlag: history.NextFinalApprovalFlag,
					NeedEscalation:        escalation,
					SourceDecision:        history.SourceDecision,
					NextStep:              history.NextStep,
					Note:                  history.Note,
					SlikResult:            history.SlikResult,
				}
				historyData = append(historyData, historyEntry)
			}
		}

		if len(historyData) < 1 {
			historyData = []entity.HistoryApproval{}
		}

		// get trx_internal_record
		internalRecord, err = u.repository.GetInternalRecord(inq.ProspectID)

		if err != nil {
			return
		}

		var (
			internalData []entity.TrxInternalRecord
		)

		if len(internalRecord) > 0 {
			for _, record := range internalRecord {
				recordEntry := entity.TrxInternalRecord{
					ApplicationID:        strings.Trim(record.ApplicationID, " "),
					ProductType:          record.ProductType,
					AgreementDate:        record.AgreementDate,
					AssetCode:            strings.Trim(record.AssetCode, " "),
					Tenor:                record.Tenor,
					OutstandingPrincipal: record.OutstandingPrincipal,
					InstallmentAmount:    record.InstallmentAmount,
					ContractStatus:       strings.Trim(record.ContractStatus, " "),
					CurrentCondition:     record.CurrentCondition,
				}
				internalData = append(internalData, recordEntry)
			}
		}

		if len(internalData) < 1 {
			internalData = []entity.TrxInternalRecord{}
		}

		row := entity.InquiryDataCa{
			CA: entity.DataCa{
				ShowAction:         inq.ShowAction,
				CaDecision:         inq.CaDecision,
				CaNote:             inq.CANote,
				ActionDate:         inq.ActionDate,
				ScsDate:            inq.ScsDate,
				ScsScore:           inq.ScsScore,
				ScsStatus:          inq.ScsStatus,
				BiroCustomerResult: inq.BiroCustomerResult,
				BiroSpouseResult:   inq.BiroSpouseResult,
			},
			InternalRecord: internalData,
			Approval:       historyData,
			Draft: entity.TrxDraftCaDecision{
				Decision:   inq.DraftDecision,
				SlikResult: inq.DraftSlikResult,
				Note:       inq.DraftNote,
			},
			General: entity.DataGeneral{
				ProspectID:     inq.ProspectID,
				BranchName:     inq.BranchName,
				IncomingSource: inq.IncomingSource,
				CreatedAt:      inq.CreatedAt,
				OrderAt:        inq.OrderAt,
			},
			Personal: entity.CustomerPersonal{
				IDNumber:          inq.IDNumber,
				LegalName:         inq.LegalName,
				CustomerID:        inq.CustomerID,
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
				SurveyResult:      inq.SurveyResult,
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
				IndustryTypeID:        strings.TrimSpace(string(industryType)),
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
				Tenor:                 inq.InstallmentPeriod,
				OTR:                   inq.OTR,
				DPAmount:              inq.DPAmount,
				AF:                    inq.FinanceAmount,
				NTF:                   inq.NTF,
				NTFAkumulasi:          inq.NTFAkumulasi,
				NTFPlusInterestAmount: inq.Total,
				InterestAmount:        inq.InterestAmount,
				LifeInsuranceFee:      inq.LifeInsuranceFee,
				AssetInsuranceFee:     inq.AssetInsuranceFee,
				InsuranceAmount:       inq.InsuranceAmount,
				InstallmentAmount:     inq.MonthlyInstallment,
				AdminFee:              inq.AdminFee,
				ProvisionFee:          inq.ProvisionFee,
				FirstInstallment:      inq.FirstInstallment,
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

func (u usecase) SaveAsDraft(ctx context.Context, req request.ReqSaveAsDraft) (data response.CAResponse, err error) {

	var (
		trxDraft entity.TrxDraftCaDecision
		decision string
	)

	switch req.Decision {
	case constant.DECISION_REJECT:
		decision = constant.DB_DECISION_REJECT
	case constant.DECISION_APPROVE:
		decision = constant.DB_DECISION_APR
	}

	trxDraft = entity.TrxDraftCaDecision{
		ProspectID: req.ProspectID,
		Decision:   decision,
		SlikResult: req.SlikResult,
		Note:       req.Note,
		CreatedBy:  req.CreatedBy,
		DecisionBy: req.DecisionBy,
	}

	data = response.CAResponse{
		ProspectID: req.ProspectID,
		Decision:   req.Decision,
		SlikResult: req.SlikResult,
		Note:       req.Note,
	}

	err = u.repository.SaveDraftData(trxDraft)
	if err != nil {
		return
	}

	return
}

func (u usecase) SubmitDecision(ctx context.Context, req request.ReqSubmitDecision) (data response.CAResponse, err error) {

	var (
		trxCaDecision entity.TrxCaDecision
		trxDetail     entity.TrxDetail
		trxStatus     entity.TrxStatus
		decision      string
	)

	// get limit approval for final_approval
	limit, err := u.repository.GetLimitApproval(req.NTFAkumulasi)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get limit approval error")
		return
	}

	switch req.Decision {
	case constant.DECISION_REJECT:
		decision = constant.DB_DECISION_REJECT
	case constant.DECISION_APPROVE:
		decision = constant.DB_DECISION_APR
	}

	trxCaDecision = entity.TrxCaDecision{
		ProspectID:    req.ProspectID,
		Decision:      decision,
		SlikResult:    req.SlikResult,
		Note:          req.Note,
		CreatedBy:     req.CreatedBy,
		DecisionBy:    req.DecisionBy,
		FinalApproval: limit.Alias,
	}

	trxStatus = entity.TrxStatus{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
	}

	trxDetail = entity.TrxDetail{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		NextStep:       constant.DB_DECISION_BRANCH_MANAGER,
		Info:           req.SlikResult,
		CreatedBy:      req.CreatedBy,
	}

	err = u.repository.ProcessTransaction(trxCaDecision, trxStatus, trxDetail)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Submit Decision error")
		return
	}

	data = response.CAResponse{
		ProspectID: req.ProspectID,
		Decision:   req.Decision,
		SlikResult: req.SlikResult,
		Note:       req.Note,
	}

	return
}

func (u usecase) GetSearchInquiry(ctx context.Context, req request.ReqSearchInquiry, pagination interface{}) (data []entity.InquiryDataSearch, rowTotal int, err error) {

	var (
		industry  []entity.SpIndustryTypeMaster
		photos    []entity.DataPhoto
		surveyor  []entity.TrxSurveyor
		trxDetail []entity.TrxDetail
	)

	// get inquiry pre screening
	result, rowTotal, err := u.repository.GetInquirySearch(req, pagination)

	if err != nil {
		return []entity.InquiryDataSearch{}, 0, err
	}

	for _, inq := range result {

		industryType, _ := u.cache.Get(inq.IndustryTypeID)

		if industryType == nil {
			industry, err = u.repository.GetSpIndustryTypeMaster()

			if err != nil {
				return
			}

			for _, description := range industry {
				u.cache.Set(strings.ReplaceAll(description.IndustryTypeID, " ", ""), []byte(description.Description))
			}
		}

		// get trx_customer_photo
		photos, err = u.repository.GetCustomerPhoto(inq.ProspectID)

		if err != nil {
			return
		}

		var photoData []entity.DataPhoto

		if len(photos) > 0 {
			for _, photo := range photos {
				photoEntry := entity.DataPhoto{
					PhotoID: photo.PhotoID,
					Label:   photo.Label,
					Url:     photo.Url,
				}
				photoData = append(photoData, photoEntry)
			}
		}

		if len(photoData) < 1 {
			photoData = []entity.DataPhoto{}
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

		// get data history process
		trxDetail, err = u.repository.GetHistoryProcess(inq.ProspectID)

		if err != nil {
			return
		}

		var historyData []entity.TrxDetail

		if len(trxDetail) > 0 {
			for _, detail := range trxDetail {
				detailEntry := entity.TrxDetail{
					SourceDecision: detail.SourceDecision,
					Decision:       detail.Decision,
					Info:           detail.Info,
					CreatedAt:      detail.CreatedAt,
				}
				historyData = append(historyData, detailEntry)
			}
		}

		if len(historyData) < 1 {
			historyData = []entity.TrxDetail{}
		}

		row := entity.InquiryDataSearch{
			Action: entity.ActionSearch{
				FinalStatus:   inq.FinalStatus,
				ActionReturn:  inq.ActionReturn,
				ActionCancel:  inq.ActionCancel,
				ActionFormAkk: inq.ActionFormAkk,
			},
			HistoryProcess: historyData,
			General: entity.DataGeneral{
				ProspectID:     inq.ProspectID,
				BranchName:     inq.BranchName,
				IncomingSource: inq.IncomingSource,
				CreatedAt:      inq.CreatedAt,
				OrderAt:        inq.OrderAt,
			},
			Personal: entity.CustomerPersonal{
				IDNumber:          inq.IDNumber,
				LegalName:         inq.LegalName,
				CustomerID:        inq.CustomerID,
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
				IndustryTypeID:        strings.TrimSpace(string(industryType)),
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
				Tenor:                 inq.InstallmentPeriod,
				OTR:                   inq.OTR,
				DPAmount:              inq.DPAmount,
				AF:                    inq.FinanceAmount,
				NTF:                   inq.NTF,
				NTFAkumulasi:          inq.NTFAkumulasi,
				NTFPlusInterestAmount: inq.Total,
				InterestAmount:        inq.InterestAmount,
				LifeInsuranceFee:      inq.LifeInsuranceFee,
				AssetInsuranceFee:     inq.AssetInsuranceFee,
				InsuranceAmount:       inq.InsuranceAmount,
				InstallmentAmount:     inq.MonthlyInstallment,
				AdminFee:              inq.AdminFee,
				ProvisionFee:          inq.ProvisionFee,
				FirstInstallment:      inq.FirstInstallment,
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

func (u usecase) CancelOrder(ctx context.Context, req request.ReqCancelOrder) (data response.CancelResponse, err error) {

	var (
		trxStatus     entity.TrxStatus
		trxDetail     entity.TrxDetail
		trxCaDecision entity.TrxCaDecision
	)

	trxCaDecision = entity.TrxCaDecision{
		ProspectID: req.ProspectID,
		Decision:   constant.DB_DECISION_CANCEL,
		Note:       req.CancelReason,
		CreatedBy:  req.CreatedBy,
		DecisionBy: req.DecisionBy,
	}

	trxStatus = entity.TrxStatus{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_FINAL,
		Activity:       constant.ACTIVITY_STOP,
		Decision:       constant.DB_DECISION_CANCEL,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		Reason:         req.CancelReason,
	}

	trxDetail = entity.TrxDetail{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_FINAL,
		Activity:       constant.ACTIVITY_STOP,
		Decision:       constant.DB_DECISION_CANCEL,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		Info:           req.CancelReason,
		CreatedBy:      req.CreatedBy,
	}

	err = u.repository.ProcessTransaction(trxCaDecision, trxStatus, trxDetail)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Process Cancel Order error")
		return
	}

	data = response.CancelResponse{
		ProspectID: req.ProspectID,
		Reason:     req.CancelReason,
		Status:     constant.CANCEL_STATUS_SUCCESS,
	}

	return
}

func (u usecase) ReturnOrder(ctx context.Context, req request.ReqReturnOrder) (data response.ReturnResponse, err error) {

	var (
		trxStatus entity.TrxStatus
		trxDetail entity.TrxDetail
	)

	trxStatus = entity.TrxStatus{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		SourceDecision: constant.PRESCREENING,
		Reason:         constant.REASON_RETURN_ORDER,
	}

	trxDetail = entity.TrxDetail{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		SourceDecision: constant.CMO_AGENT,
		NextStep:       constant.PRESCREENING,
		Info:           constant.REASON_RETURN_ORDER,
		CreatedBy:      constant.SYSTEM_CREATED,
	}

	err = u.repository.ProcessReturnOrder(req.ProspectID, trxStatus, trxDetail)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Process Return Order error")
		return
	}

	data = response.ReturnResponse{
		ProspectID: req.ProspectID,
		Status:     constant.RETURN_STATUS_SUCCESS,
	}

	return
}
