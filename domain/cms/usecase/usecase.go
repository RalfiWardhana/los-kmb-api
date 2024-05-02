package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	cache "los-kmb-api/domain/cache/interfaces"
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"mime/multipart"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	jsoniter "github.com/json-iterator/go"
	"github.com/xuri/excelize/v2"
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

func (u usecase) GetAkkk(prospectID string) (data entity.Akkk, err error) {
	data, err = u.repository.GetAkkk(prospectID)

	if err != nil {
		if err.Error() == constant.RECORD_NOT_FOUND {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - " + err.Error())
			return
		}
		err = errors.New(constant.ERROR_UPSTREAM + " - " + err.Error())
		return
	}

	industryType, _ := u.cache.Get(data.IndustryTypeID.(string))

	var industry []entity.SpIndustryTypeMaster
	if industryType == nil {
		industry, err = u.repository.GetSpIndustryTypeMaster()

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - " + err.Error())
			return
		}

		for _, description := range industry {
			u.cache.Set(strings.ReplaceAll(description.IndustryTypeID, " ", ""), []byte(description.Description))
		}
		industryType, _ = u.cache.Get(data.IndustryTypeID.(string))
	}

	if industryType != nil {
		data.IndustryType = strings.TrimSpace(string(industryType))
	}

	if data.MonthlyFixedIncome != nil {
		data.MonthlyFixedIncome, _ = utils.GetFloat(data.MonthlyFixedIncome)
	}

	if data.MonthlyVariableIncome != nil {
		data.MonthlyVariableIncome, _ = utils.GetFloat(data.MonthlyVariableIncome)
	}

	if data.SpouseIncome != nil {
		data.SpouseIncome, _ = utils.GetFloat(data.SpouseIncome)
	}

	if data.Plafond != nil {
		data.Plafond, _ = utils.GetFloat(data.Plafond)
	}

	if data.BakiDebet != nil {
		data.BakiDebet, _ = utils.GetFloat(data.BakiDebet)
	}

	if data.BakiDebetTerburuk != nil {
		data.BakiDebetTerburuk, _ = utils.GetFloat(data.BakiDebetTerburuk)
	}

	if data.SpousePlafond != nil {
		data.SpousePlafond, _ = utils.GetFloat(data.SpousePlafond)
	}

	if data.SpouseBakiDebet != nil {
		data.SpouseBakiDebet, _ = utils.GetFloat(data.SpouseBakiDebet)
	}

	if data.SpouseBakiDebetTerburuk != nil {
		data.SpouseBakiDebetTerburuk, _ = utils.GetFloat(data.SpouseBakiDebetTerburuk)
	}

	if data.TotalAgreementAktif != nil {
		data.TotalAgreementAktif, _ = utils.GetFloat(data.TotalAgreementAktif)
	}

	if data.MaxOVDAgreementAktif != nil {
		data.MaxOVDAgreementAktif, _ = utils.GetFloat(data.MaxOVDAgreementAktif)
	}

	if data.LastMaxOVDAgreement != nil {
		data.LastMaxOVDAgreement, _ = utils.GetFloat(data.LastMaxOVDAgreement)
	}

	if data.LatestInstallment != nil {
		data.LatestInstallment, _ = utils.GetFloat(data.LatestInstallment)
	}

	if data.NTFAkumulasi != nil {
		data.NTFAkumulasi, _ = utils.GetFloat(data.NTFAkumulasi)
	}

	if data.TotalInstallment != nil {
		data.TotalInstallment, _ = utils.GetFloat(data.TotalInstallment)
	}

	if data.TotalIncome != nil {
		data.TotalIncome, _ = utils.GetFloat(data.TotalIncome)
	}

	if data.TotalDSR != nil {
		data.TotalDSR, _ = utils.GetFloat(data.TotalDSR)
	}

	if data.EkycSimiliarity != nil {
		data.EkycSimiliarity, _ = utils.GetFloat(data.EkycSimiliarity)
	}

	return
}

func (u usecase) SubmitNE(ctx context.Context, req request.MetricsNE) (data interface{}, err error) {

	filtering := request.Filtering{
		ProspectID: req.Transaction.ProspectID,
		BranchID:   req.Transaction.BranchID,
		BirthDate:  req.CustomerPersonal.BirthDate,
		Gender:     req.CustomerPersonal.Gender,
		BPKBName:   req.Item.BPKBName,
	}

	filtering.IDNumber, _ = utils.PlatformEncryptText(req.CustomerPersonal.IDNumber)
	filtering.LegalName, _ = utils.PlatformEncryptText(req.CustomerPersonal.LegalName)
	filtering.MotherName, _ = utils.PlatformEncryptText(req.CustomerPersonal.SurgateMotherName)

	if req.CustomerSpouse != nil {
		IDNumber, _ := utils.PlatformEncryptText(req.CustomerSpouse.IDNumber)
		LegalName, _ := utils.PlatformEncryptText(req.CustomerSpouse.LegalName)
		MotherName, _ := utils.PlatformEncryptText(req.CustomerSpouse.SurgateMotherName)

		filtering.Spouse = &request.FilteringSpouse{
			IDNumber:   IDNumber,
			LegalName:  LegalName,
			MotherName: MotherName,
			BirthDate:  req.CustomerSpouse.BirthDate,
			Gender:     req.CustomerSpouse.Gender,
		}

	}

	elaborateLTV := request.ElaborateLTV{
		ProspectID:        req.Transaction.ProspectID,
		Tenor:             req.Apk.Tenor,
		ManufacturingYear: req.Item.ManufactureYear,
	}

	var journey request.Metrics
	err = copier.Copy(&journey, &req)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Submit NE error")
		return
	}

	journey.CustomerPersonal.IDNumber = filtering.IDNumber
	journey.CustomerPersonal.LegalName = filtering.LegalName
	journey.CustomerPersonal.FullName = filtering.LegalName
	journey.CustomerPersonal.SurgateMotherName = filtering.MotherName

	if req.CustomerSpouse != nil {
		IDNumber, _ := utils.PlatformEncryptText(req.CustomerSpouse.IDNumber)
		LegalName, _ := utils.PlatformEncryptText(req.CustomerSpouse.LegalName)
		MotherName, _ := utils.PlatformEncryptText(req.CustomerSpouse.SurgateMotherName)

		spouse := &request.CustomerSpouse{
			IDNumber:          IDNumber,
			LegalName:         LegalName,
			SurgateMotherName: MotherName,
		}

		journey.CustomerSpouse.IDNumber = spouse.IDNumber
		journey.CustomerSpouse.LegalName = spouse.LegalName
		journey.CustomerSpouse.FullName = spouse.LegalName
		journey.CustomerSpouse.SurgateMotherName = spouse.SurgateMotherName
	}

	err = u.repository.SubmitNE(req, filtering, elaborateLTV, journey)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - " + err.Error())
		return
	}

	// the func will return the payload of filtering and produce it later
	data = filtering

	return
}

func (u usecase) GetInquiryNE(ctx context.Context, req request.ReqInquiryNE, pagination interface{}) (data []entity.InquiryDataNE, rowTotal int, err error) {

	result, rowTotal, err := u.repository.GetInquiryNE(req, pagination)

	if err != nil {
		return
	}

	data = result

	return
}

func (u usecase) GetInquiryNEDetail(ctx context.Context, prospectID string) (data request.MetricsNE, err error) {

	var (
		trxNewEntry entity.NewEntry
	)

	trxNewEntry, err = u.repository.GetInquiryNEDetail(prospectID)
	if err != nil {
		if err.Error() == constant.RECORD_NOT_FOUND {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - " + err.Error())
		} else {
			err = errors.New(constant.ERROR_UPSTREAM + " - " + err.Error())
		}
		return
	}

	json.Unmarshal([]byte(trxNewEntry.PayloadNE), &data)

	return
}

func (u usecase) GetReasonPrescreening(ctx context.Context, req request.ReqReasonPrescreening, pagination interface{}) (data []entity.ReasonMessage, rowTotal int, err error) {

	data, rowTotal, err = u.repository.GetReasonPrescreening(req, pagination)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetCancelReason(ctx context.Context, pagination interface{}) (data []entity.CancelReason, rowTotal int, err error) {

	data, rowTotal, err = u.repository.GetCancelReason(pagination)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetApprovalReason(ctx context.Context, req request.ReqApprovalReason, pagination interface{}) (data []entity.ApprovalReason, rowTotal int, err error) {

	data, rowTotal, err = u.repository.GetApprovalReason(req, pagination)

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

		industryType, _ = u.cache.Get(inq.IndustryTypeID)

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
		if req.BranchID != constant.BRANCHID_HO && inq.Activity == constant.ACTIVITY_UNPROCESS && inq.SourceDecision == constant.PRESCREENING {
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

		birthDate := inq.BirthDate.Format("02-01-2006")

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
			Personal: entity.DataPersonal{
				IDNumber:          inq.IDNumber,
				LegalName:         inq.LegalName,
				CustomerStatus:    inq.CustomerStatus,
				BirthPlace:        inq.BirthPlace,
				BirthDate:         birthDate,
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

	status, err := u.repository.GetTrxStatus(req.ProspectID)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get status order error")
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
			Info           string
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
				Info:           constant.REASON_TIDAK_SESUAI,
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
				Info:           constant.REASON_SESUAI,
			},
		}

		decisionInfo, ok := decisionMapping[req.Decision]
		if !ok {
			err = errors.New(constant.ERROR_UPSTREAM + " - Decision tidak valid")
			return
		}

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
			Info:           decisionInfo.Info,
			CreatedBy:      req.DecisionBy,
			Reason:         reason,
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

		data.Code = decisionInfo.Code
		data.ProspectID = req.ProspectID
		data.Decision = decisionInfo.Decision
		data.Reason = reason

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
		action         bool
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

		industryType, _ = u.cache.Get(inq.IndustryTypeID)

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
				if history.NeedEscalation == 1 {
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

		var statusDecision string
		if inq.StatusDecision == constant.DB_DECISION_APR {
			statusDecision = constant.DECISION_APPROVE
		} else if inq.StatusDecision == constant.DB_DECISION_REJECT {
			statusDecision = constant.DECISION_REJECT
		} else if inq.StatusDecision == constant.DB_DECISION_CANCEL {
			statusDecision = constant.DECISION_CANCEL
		}

		action = inq.ShowAction
		if req.BranchID == constant.BRANCHID_HO {
			action = false
		}

		birthDate := inq.BirthDate.Format("02-01-2006")

		row := entity.InquiryDataCa{
			CA: entity.DataCa{
				ShowAction:         action,
				ActionEditData:     inq.ActionEditData,
				AdditionalDP:       inq.AdditionalDP,
				StatusDecision:     statusDecision,
				StatusReason:       inq.StatusReason,
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
			Personal: entity.DataPersonal{
				IDNumber:          inq.IDNumber,
				LegalName:         inq.LegalName,
				CustomerID:        inq.CustomerID,
				CustomerStatus:    inq.CustomerStatus,
				BirthPlace:        inq.BirthPlace,
				BirthDate:         birthDate,
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
		trxCaDecision      entity.TrxCaDecision
		trxDetail          entity.TrxDetail
		trxStatus          entity.TrxStatus
		limit              entity.MappingLimitApprovalScheme
		trxHistoryApproval entity.TrxHistoryApprovalScheme
		nextFinal          int
		decision           string
		decision_detail    string
	)

	status, err := u.repository.GetTrxStatus(req.ProspectID)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get status order error")
		return
	}

	// Bisa melakukan submit jika status UNPR dan decision CPR
	if status.Activity == constant.ACTIVITY_UNPROCESS && status.Decision == constant.DB_DECISION_CREDIT_PROCESS {

		// get limit approval for final_approval
		limit, err = u.repository.GetLimitApproval(req.NTFAkumulasi)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get limit approval error")
			return
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = constant.DB_DECISION_REJECT
			decision_detail = constant.DB_DECISION_REJECT
		case constant.DECISION_APPROVE:
			decision = constant.DB_DECISION_APR
			decision_detail = constant.DB_DECISION_PASS
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
			RuleCode:       constant.CODE_CBM,
			SourceDecision: constant.DB_DECISION_BRANCH_MANAGER,
			Reason:         req.SlikResult,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       decision_detail,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
			NextStep:       constant.DB_DECISION_BRANCH_MANAGER,
			Info:           req.SlikResult,
			CreatedBy:      req.CreatedBy,
			Reason:         req.SlikResult,
		}

		nextFinal = 0
		if trxCaDecision.FinalApproval == constant.DB_DECISION_BRANCH_MANAGER {
			nextFinal = 1
		}

		trxHistoryApproval = entity.TrxHistoryApprovalScheme{
			ProspectID:            trxCaDecision.ProspectID,
			Decision:              trxCaDecision.Decision,
			Reason:                trxCaDecision.SlikResult.(string),
			Note:                  trxCaDecision.Note,
			CreatedBy:             trxCaDecision.CreatedBy,
			DecisionBy:            trxCaDecision.DecisionBy,
			NeedEscalation:        0,
			NextFinalApprovalFlag: nextFinal,
			SourceDecision:        trxDetail.SourceDecision,
			NextStep:              trxDetail.NextStep.(string),
		}

		err = u.repository.ProcessTransaction(trxCaDecision, trxHistoryApproval, trxStatus, trxDetail)
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
	} else {
		err = errors.New(constant.ERROR_UPSTREAM + " - Status order tidak sedang dalam credit process")
		return
	}

	return
}

func (u usecase) GetSearchInquiry(ctx context.Context, req request.ReqSearchInquiry, pagination interface{}) (data []entity.InquiryDataSearch, rowTotal int, err error) {

	var (
		industry       []entity.SpIndustryTypeMaster
		photos         []entity.DataPhoto
		surveyor       []entity.TrxSurveyor
		historyProcess []entity.HistoryProcess
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

		industryType, _ = u.cache.Get(inq.IndustryTypeID)

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
		historyProcess, err = u.repository.GetHistoryProcess(inq.ProspectID)

		if err != nil {
			return
		}

		var (
			historyData []entity.HistoryProcess
			reason      string
		)

		if len(historyProcess) > 0 {
			for _, detail := range historyProcess {
				if detail.SourceDecision != constant.SOURCE_DECISION_DSR || detail.NextStep != constant.SOURCE_DECISION_DUPCHECK {
					reason = detail.Reason
					if detail.SourceDecision == constant.CREDIT_ANALYSIS || detail.SourceDecision == constant.CREDIT_COMMITEE {
						reason = fmt.Sprintf("%s: %s", detail.Alias, detail.Reason)
					}
					detailEntry := entity.HistoryProcess{
						SourceDecision: detail.SourceDecision,
						Decision:       detail.Decision,
						Reason:         reason,
						CreatedAt:      detail.CreatedAt,
					}
					historyData = append(historyData, detailEntry)
				}
			}
		}

		if len(historyData) < 1 {
			historyData = []entity.HistoryProcess{}
		}

		birthDate := inq.BirthDate.Format("02-01-2006")

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
			Personal: entity.DataPersonal{
				IDNumber:          inq.IDNumber,
				LegalName:         inq.LegalName,
				CustomerID:        inq.CustomerID,
				CustomerStatus:    inq.CustomerStatus,
				BirthPlace:        inq.BirthPlace,
				BirthDate:         birthDate,
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
		trxStatus          entity.TrxStatus
		trxDetail          entity.TrxDetail
		trxCaDecision      entity.TrxCaDecision
		trxHistoryApproval entity.TrxHistoryApprovalScheme
	)

	status, err := u.repository.GetTrxStatus(req.ProspectID)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get status order error")
		return
	}

	// Bisa melakukan CANCEL jika decision tidak sama dengan REJECT
	if status.Decision != constant.DB_DECISION_REJECT {

		trxCaDecision = entity.TrxCaDecision{
			ProspectID: req.ProspectID,
			Decision:   constant.DB_DECISION_CANCEL,
			Note:       req.CancelReason,
			SlikResult: "",
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
			Reason:         req.CancelReason,
		}

		trxHistoryApproval = entity.TrxHistoryApprovalScheme{
			ID:                    uuid.New().String(),
			ProspectID:            trxCaDecision.ProspectID,
			Decision:              trxCaDecision.Decision,
			Reason:                req.CancelReason,
			CreatedAt:             time.Now(),
			CreatedBy:             trxCaDecision.CreatedBy,
			DecisionBy:            trxCaDecision.DecisionBy,
			NeedEscalation:        0,
			NextFinalApprovalFlag: 0,
			SourceDecision:        trxDetail.SourceDecision,
		}

		err = u.repository.ProcessTransaction(trxCaDecision, trxHistoryApproval, trxStatus, trxDetail)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Process Cancel Order error")
			return
		}

		data = response.CancelResponse{
			ProspectID: req.ProspectID,
			Reason:     req.CancelReason,
			Status:     constant.CANCEL_STATUS_SUCCESS,
		}

	} else {
		err = errors.New(constant.ERROR_UPSTREAM + " - Status order tidak dapat dicancel")
		return
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
		Reason:         constant.REASON_RETURN_ORDER,
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

func (u usecase) RecalculateOrder(ctx context.Context, req request.ReqRecalculateOrder, accessToken string) (data response.RecalculateResponse, err error) {

	var (
		trxStatus             entity.TrxStatus
		trxDetail             entity.TrxDetail
		trxHistoryApproval    entity.TrxHistoryApprovalScheme
		respSubmitRecalculate response.SubmitRecalculateResponse
	)

	trxStatus = entity.TrxStatus{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		SourceDecision: constant.NEED_RECALCULATE,
		Reason:         constant.REASON_NEED_RECALCULATE,
	}

	infoMap := map[string]float64{
		"dp_amount": req.DPAmount,
	}
	info, _ := json.Marshal(infoMap)

	trxDetail = entity.TrxDetail{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		NextStep:       constant.NEED_RECALCULATE,
		Info:           string(info),
		CreatedBy:      req.CreatedBy,
		Reason:         constant.REASON_NEED_RECALCULATE,
	}

	trxHistoryApproval = entity.TrxHistoryApprovalScheme{
		ProspectID:            req.ProspectID,
		Decision:              constant.DB_DECISION_SDP,
		Reason:                trxStatus.Reason,
		Note:                  fmt.Sprintf("Nilai DP: %.0f", req.DPAmount),
		CreatedBy:             req.CreatedBy,
		DecisionBy:            req.DecisionBy,
		NeedEscalation:        0,
		NextFinalApprovalFlag: 1,
		SourceDecision:        trxDetail.SourceDecision,
	}

	// hit sally recalculate
	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	header := map[string]string{
		"Authorization": accessToken,
	}

	param, _ := json.Marshal(map[string]interface{}{
		"prospect_id":   req.ProspectID,
		"dp_amount_los": req.DPAmount,
	})

	submitRecalculate, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("SUBMIT_RECALCULATE_SALLY"), param, header, constant.METHOD_POST, false, 0, timeout, req.ProspectID, accessToken)

	if submitRecalculate.StatusCode() == 504 || submitRecalculate.StatusCode() == 502 {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Submit Recalculate to Sally Timeout")
		return
	}

	if submitRecalculate.StatusCode() != 200 && submitRecalculate.StatusCode() != 504 && submitRecalculate.StatusCode() != 502 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Submit Recalculate to Sally Error")
		return
	}

	if err == nil && submitRecalculate.StatusCode() == 200 {
		json.Unmarshal([]byte(jsoniter.Get(submitRecalculate.Body()).ToString()), &respSubmitRecalculate)
	}

	if err == nil && respSubmitRecalculate.Code == 200 {
		err = u.repository.ProcessRecalculateOrder(req.ProspectID, trxStatus, trxDetail, trxHistoryApproval)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Process Recalculate Order error")
			return
		}

		data = response.RecalculateResponse{
			ProspectID: req.ProspectID,
			DPAmount:   req.DPAmount,
			Status:     constant.RECALCULATE_STATUS_SUCCESS,
		}
	} else {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Submit Recalculate to Sally Error")
		return
	}

	return
}

func (u usecase) GetInquiryApproval(ctx context.Context, req request.ReqInquiryApproval, pagination interface{}) (data []entity.InquiryDataApproval, rowTotal int, err error) {

	var (
		industry       []entity.SpIndustryTypeMaster
		photos         []entity.DataPhoto
		surveyor       []entity.TrxSurveyor
		histories      []entity.HistoryApproval
		internalRecord []entity.TrxInternalRecord
	)

	// get inquiry pre screening
	result, rowTotal, err := u.repository.GetInquiryApproval(req, pagination)

	if err != nil {
		return []entity.InquiryDataApproval{}, 0, err
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

		industryType, _ = u.cache.Get(inq.IndustryTypeID)

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
		)

		if len(histories) > 0 {
			for _, history := range histories {
				historyEntry := entity.HistoryApproval{
					DecisionBy:            history.DecisionBy,
					Decision:              history.Decision,
					CreatedAt:             history.CreatedAt,
					NextFinalApprovalFlag: history.NextFinalApprovalFlag,
					NeedEscalation:        history.NeedEscalation,
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

		var statusDecision string
		if inq.StatusDecision == constant.DB_DECISION_APR {
			statusDecision = constant.DECISION_APPROVE
		} else if inq.StatusDecision == constant.DB_DECISION_REJECT {
			statusDecision = constant.DECISION_REJECT
		} else if inq.StatusDecision == constant.DB_DECISION_CANCEL {
			statusDecision = constant.DECISION_CANCEL
		}

		birthDate := inq.BirthDate.Format("02-01-2006")

		row := entity.InquiryDataApproval{
			CA: entity.DataApproval{
				ShowAction:         inq.ShowAction,
				ActionFormAkk:      inq.ActionFormAkk,
				IsLastApproval:     inq.IsLastApproval,
				HasReturn:          inq.HasReturn,
				StatusDecision:     statusDecision,
				StatusReason:       inq.StatusReason,
				FinalApproval:      inq.FinalApproval,
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
			General: entity.DataGeneral{
				ProspectID:     inq.ProspectID,
				BranchName:     inq.BranchName,
				IncomingSource: inq.IncomingSource,
				CreatedAt:      inq.CreatedAt,
				OrderAt:        inq.OrderAt,
			},
			Personal: entity.DataPersonal{
				IDNumber:          inq.IDNumber,
				LegalName:         inq.LegalName,
				CustomerID:        inq.CustomerID,
				CustomerStatus:    inq.CustomerStatus,
				BirthPlace:        inq.BirthPlace,
				BirthDate:         birthDate,
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

func (u usecase) SubmitApproval(ctx context.Context, req request.ReqSubmitApproval) (data response.ApprovalResponse, err error) {

	var (
		trxDetail                                     entity.TrxDetail
		trxStatus                                     entity.TrxStatus
		recentDP                                      entity.AFMobilePhone
		trxRecalculate                                entity.TrxRecalculate
		approvalScheme                                response.RespApprovalScheme
		decision                                      string
		decision_detail                               string
		rej, apr, pas, rtn, cpr, cra, onp, unpr, prcd string
	)

	rej = constant.DB_DECISION_REJECT
	apr = constant.DB_DECISION_APR
	pas = constant.DB_DECISION_PASS
	rtn = constant.DB_DECISION_RTN
	cpr = constant.DB_DECISION_CREDIT_PROCESS
	cra = constant.DB_DECISION_CREDIT_ANALYST
	onp = constant.STATUS_ONPROCESS
	unpr = constant.ACTIVITY_UNPROCESS
	prcd = constant.ACTIVITY_PROCESS

	switch req.Decision {
	case constant.DECISION_REJECT:
		decision = rej
		decision_detail = rej

	case constant.DECISION_APPROVE:
		decision = apr
		decision_detail = pas

	case constant.DECISION_RETURN:
		decision = rtn
		decision_detail = rtn
	}

	if req.Decision == constant.DECISION_RETURN {

		recentDP, err = u.repository.GetAFMobilePhone(req.ProspectID)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Recently DP Amount error")
			return
		}

		// validate DP Amount should greather than recent DP
		if req.DPAmount <= recentDP.DPAmount {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Nilai DP harus lebih besar dari DP Awal")
			return
		}

		// validate DP Amount should lower than recent OTR
		if req.DPAmount > recentDP.OTR {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Nilai DP tidak boleh lebih dari OTR")
			return
		}
	}

	approvalScheme, err = utils.ApprovalScheme(req)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Approval Scheme error")
		return
	}

	trxStatus = entity.TrxStatus{
		ProspectID:     req.ProspectID,
		StatusProcess:  onp,
		Activity:       unpr,
		Decision:       cpr,
		RuleCode:       req.RuleCode,
		SourceDecision: approvalScheme.NextStep,
		Reason:         req.Reason,
	}

	trxDetail = entity.TrxDetail{
		ProspectID:     req.ProspectID,
		StatusProcess:  onp,
		Activity:       prcd,
		Decision:       decision_detail,
		RuleCode:       req.RuleCode,
		SourceDecision: req.Alias,
		Info:           req.Reason,
		CreatedBy:      req.CreatedBy,
		Reason:         req.Reason,
	}

	if approvalScheme.NextStep != "" {
		trxDetail.NextStep = approvalScheme.NextStep
	}

	if approvalScheme.IsFinal && !req.NeedEscalation && req.Decision != constant.DECISION_RETURN {
		trxStatus.Decision = decision
		trxStatus.StatusProcess = constant.STATUS_FINAL
		trxStatus.Activity = constant.ACTIVITY_STOP

		trxDetail.StatusProcess = constant.STATUS_FINAL
		trxDetail.Activity = constant.ACTIVITY_STOP
	}

	if req.Decision == constant.DECISION_RETURN {
		trxStatus.SourceDecision = cra
		trxStatus.RuleCode = constant.CODE_CREDIT_COMMITTEE
		trxStatus.Reason = "Returning Order"

		trxDetail.NextStep = cra
		trxDetail.Info = constant.REASON_RETURN_APPROVAL
		trxDetail.Reason = "Returning Order"

		trxRecalculate = entity.TrxRecalculate{
			ProspectID:   req.ProspectID,
			AdditionalDP: req.DPAmount,
		}
	}

	err = u.repository.SubmitApproval(req, trxStatus, trxDetail, trxRecalculate, approvalScheme)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Submit Approval error")
		return
	}

	data = response.ApprovalResponse{
		ProspectID:     req.ProspectID,
		Decision:       req.Decision,
		Reason:         req.Reason,
		Note:           req.Note,
		IsFinal:        approvalScheme.IsFinal,
		NeedEscalation: approvalScheme.IsEscalation,
	}

	return
}

func (u usecase) GetInquiryMappingCluster(req request.ReqListMappingCluster, pagination interface{}) (data []entity.InquiryMappingCluster, rowTotal int, err error) {

	data, rowTotal, err = u.repository.GetInquiryMappingCluster(req, pagination)

	if err != nil {
		return
	}

	return
}

func (u usecase) GenerateExcelMappingCluster() (genName, fileName string, err error) {

	var (
		mappingClusterBranchs []entity.InquiryMappingCluster
	)

	mappingClusterBranchs, _, err = u.repository.GetInquiryMappingCluster(request.ReqListMappingCluster{}, nil)
	if err != nil && err.Error() != constant.RECORD_NOT_FOUND {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get mapping cluster branch error")
		return
	}

	xlsx := excelize.NewFile()
	defer func() {
		if err := xlsx.Close(); err != nil {
			return
		}
	}()

	sheetName := "Mapping Cluster Branch"

	index := xlsx.NewSheet("Sheet1")
	xlsx.SetActiveSheet(index)
	xlsx.SetSheetName("Sheet1", sheetName)

	rowHeader := []string{"branch_id", "branch_name", "customer_status", "bpkb_name_type", "cluster"}

	colSize := []float64{13, 34, 18, 20, 14}

	centerAlignment := &excelize.Alignment{
		Horizontal: "center",
	}

	boldFont := &excelize.Font{
		Bold: true, Family: "Calibri", Size: 11, Color: "000000",
	}

	border := []excelize.Border{
		{Type: "left", Color: "000000", Style: 1}, {Type: "top", Color: "000000", Style: 1}, {Type: "bottom", Color: "000000", Style: 1}, {Type: "right", Color: "000000", Style: 1},
	}

	colorHeader := excelize.Fill{
		Type: "pattern", Color: []string{"#BCBCBC"}, Pattern: 1,
	}

	styleHeader, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: centerAlignment,
		Font:      boldFont,
		Border:    border,
		Fill:      colorHeader,
	})

	styleBody, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: centerAlignment,
		Border:    border,
	})

	streamWriter, err := xlsx.NewStreamWriter(sheetName)
	if err != nil {
		return
	}

	for rowID := 1; rowID <= len(mappingClusterBranchs)+1; rowID++ {
		row := make([]interface{}, 5)
		if rowID == 1 {
			for idx, val := range rowHeader {
				row[idx] = excelize.Cell{StyleID: styleHeader, Value: val}
				streamWriter.SetColWidth(idx+1, idx+2, colSize[idx])
			}
		} else {
			row[0] = excelize.Cell{StyleID: styleBody, Value: mappingClusterBranchs[rowID-2].BranchID}
			row[1] = excelize.Cell{StyleID: styleBody, Value: mappingClusterBranchs[rowID-2].BranchName}
			row[2] = excelize.Cell{StyleID: styleBody, Value: mappingClusterBranchs[rowID-2].CustomerStatus}
			row[3] = excelize.Cell{StyleID: styleBody, Value: mappingClusterBranchs[rowID-2].BpkbNameType}
			row[4] = excelize.Cell{StyleID: styleBody, Value: mappingClusterBranchs[rowID-2].Cluster}
		}

		cell, _ := excelize.CoordinatesToCellName(1, rowID)
		if err = streamWriter.SetRow(cell, row); err != nil {
			return
		}
	}

	if err = streamWriter.Flush(); err != nil {
		return
	}

	now := time.Now()
	fileName = "MappingCluster_" + now.Format("20060102150405") + ".xlsx"
	genName = utils.GenerateUUID()

	if err = xlsx.SaveAs(fmt.Sprintf("./%s.xlsx", genName)); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Save excel mapping cluster error")
		return
	}

	return
}

func (u usecase) UpdateMappingCluster(req request.ReqUploadMappingCluster, file multipart.File) (err error) {

	var (
		dataClusterBefore string
		dataClusterAfter  string
		history           entity.HistoryConfigChanges
		cluster           []entity.MasterMappingCluster
	)

	f, err := excelize.OpenReader(file)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Open file excel mapping cluster error")
		return
	}
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return
	}

	var clusterRegex = regexp.MustCompile(`^Cluster [A-Z]$`)
	var uniqueMappingData = make(map[string]int)
	for i, row := range rows {
		if i == 0 {
			if len(row) == 0 {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "format file excel tidak sesuai: kolom pertama harus berjudul 'branch_id'")
			} else if row[0] != "branch_id" {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "format file excel tidak sesuai: kolom pertama harus berjudul 'branch_id'")
			} else if row[2] != "customer_status" {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "format file excel tidak sesuai: kolom ketiga harus berjudul 'customer_status'")
			} else if row[3] != "bpkb_name_type" {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "format file excel tidak sesuai: kolom keempat harus berjudul 'bpkb_name_type'")
			} else if len(row) < 5 {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "format file excel tidak sesuai: kolom kelima harus berjudul 'cluster'")
			}
		} else {
			if len(row) == 0 {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + ", nilai branch_id tidak boleh kosong")
			}

			branchID := strings.TrimSpace(row[0])
			if branchID == "" {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + ", nilai branch_id tidak boleh kosong")
			} else if branchID == "0" {
				branchID = constant.BRANCH_ID_PRIME_PRIORITY
			}

			customerStatus := strings.ToUpper(strings.TrimSpace(row[2]))
			if customerStatus != constant.STATUS_KONSUMEN_NEW && customerStatus != "AO/RO" {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + ", nilai customer_status harus " + constant.STATUS_KONSUMEN_NEW + " atau AO/RO")
			}

			bpkbName, err := strconv.Atoi(row[3])
			if err != nil {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + ", nilai bpkb_name_type harus 0 atau 1")
			}

			if bpkbName != 0 && bpkbName != 1 {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + ", nilai bpkb_name_type harus 0 atau 1")
			}

			if len(row) < 5 {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + ", nilai cluster tidak boleh kosong")
			}

			clusterStr := strings.TrimSpace(row[4])
			if strings.EqualFold(clusterStr, constant.CLUSTER_PRIME_PRIORITY) {
				clusterStr = strings.ToUpper(clusterStr)
			} else {
				clusterStr = strings.Title(strings.ToLower(clusterStr))
			}

			if clusterStr != constant.CLUSTER_PRIME_PRIORITY && !clusterRegex.MatchString(clusterStr) {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + ", nilai cluster tidak sesuai ketentuan")
			}

			uniqueKey := fmt.Sprintf("%s-%s-%d", branchID, customerStatus, bpkbName)

			if rowIndex, exists := uniqueMappingData[uniqueKey]; exists {
				return errors.New(constant.ERROR_BAD_REQUEST + " - " + "row " + strconv.Itoa(i+1) + " dan row " + strconv.Itoa(rowIndex+1) + ", entri duplikat untuk nilai branch_id, customer_status, dan bpkb_name_type")
			}

			uniqueMappingData[uniqueKey] = i

			cluster = append(cluster, entity.MasterMappingCluster{
				BranchID:       branchID,
				CustomerStatus: customerStatus,
				BpkbNameType:   bpkbName,
				Cluster:        clusterStr,
			})
		}
	}

	if len(cluster) > 0 {
		existingCluster, err := u.repository.GetMappingCluster()
		if err != nil && err.Error() != constant.RECORD_NOT_FOUND {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get existing mapping cluster branch error")
			return err
		}

		jsonDataBefore, err := json.Marshal(existingCluster)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Error converting to JSON mapping cluster before")
			return err
		}

		dataClusterBefore = string(jsonDataBefore)

		jsonDataAfter, err := json.Marshal(cluster)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Error converting to JSON mapping cluster after")
			return err
		}

		dataClusterAfter = string(jsonDataAfter)

		history = entity.HistoryConfigChanges{
			ID:         utils.GenerateUUID(),
			ConfigID:   "kmb_mapping_cluster_branch",
			ObjectName: "kmb_mapping_cluster_branch",
			Action:     "UPDATE",
			DataBefore: dataClusterBefore,
			DataAfter:  dataClusterAfter,
			CreatedBy:  req.UserID,
			CreatedAt:  time.Now(),
		}

		err = u.repository.BatchUpdateMappingCluster(cluster, history)
		if err != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - " + err.Error())
			return err
		}
	} else {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Mapping cluster branch dalam file excel kosong")
		return err
	}

	return
}

func (u usecase) GetMappingClusterBranch(req request.ReqListMappingClusterBranch) (data []entity.ConfinsBranch, err error) {

	data, err = u.repository.GetMappingClusterBranch(req)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetMappingClusterChangeLog(pagination interface{}) (data []entity.MappingClusterChangeLog, rowTotal int, err error) {

	data, rowTotal, err = u.repository.GetMappingClusterChangeLog(pagination)

	if err != nil {
		return
	}

	return
}
