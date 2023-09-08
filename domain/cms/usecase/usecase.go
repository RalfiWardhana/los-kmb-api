package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"strings"

	"github.com/allegro/bigcache/v3"
)

type (
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		cache      *bigcache.BigCache
	}
)

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, cache *bigcache.BigCache) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
		cache:      cache,
	}
}

func (u usecase) GetInquiryPrescreening(ctx context.Context, req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryData, rowTotal int, err error) {

	var (
		industry []entity.SpIndustryTypeMaster
		photos   []entity.TrxCustomerPhoto
		action   bool
	)

	if u.cache != nil {

		getValue, _ := u.cache.Get("GetSpIndustryTypeMaster")

		if getValue == nil {
			industry, err = u.repository.GetSpIndustryTypeMaster()

			if err != nil {
				if err.Error() == constant.RECORD_NOT_FOUND {
					err = errors.New(constant.ERROR_BAD_REQUEST + " - Industry Type Master Kosong")
				} else {
					err = errors.New(constant.ERROR_UPSTREAM + " - Get Industry Type Error")
				}
				return
			}

			u.cache.Set("GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

			for _, description := range industry {
				u.cache.Set(strings.ReplaceAll(description.IndustryTypeID, " ", ""), []byte(description.Description))
			}
		}
	} else {
		return
	}

	// get inquiry pre screening
	result, rowTotal, err := u.repository.GetInquiryPrescreening(req, pagination)

	if err != nil {
		if err.Error() == constant.RECORD_NOT_FOUND {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Pre Screening Inquiry Kosong")
		} else {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Pre Screening Inquiry Error")
		}
		return
	}

	for _, inq := range result {

		industry_type, _ := u.cache.Get(inq.IndustryTypeID)

		// get trx_customer_photo
		photos, err = u.repository.GetCustomerPhoto(inq.ProspectID)

		if err != nil {
			if err.Error() == constant.RECORD_NOT_FOUND {
				err = errors.New(constant.ERROR_BAD_REQUEST + " - Customer Photo kosong")
			} else {
				err = errors.New(constant.ERROR_UPSTREAM + " - Get Customer Photo Error")
			}
			return
		}

		var photoData []entity.DataCustomerPhoto

		for _, photo := range photos {
			photoEntry := entity.DataCustomerPhoto{
				PhotoID:  photo.PhotoID,
				PhotoURL: photo.PhotoURL,
			}
			photoData = append(photoData, photoEntry)
		}

		if inq.Activity == constant.ACTIVITY_UNPROCESS && inq.SourceDecision == constant.PRESCREENING {
			action = true
		}

		row := entity.InquiryData{
			Prescreening: entity.DataPrescreening{
				ShowAction: action,
				Decision:   inq.Decision,
				Reason:     inq.Reason,
				DecisionBy: inq.DecisionBy,
				DecisionAt: inq.DecisionAt,
			},
			General: entity.DataGeneral{
				ProspectID:     inq.ProspectID,
				BranchName:     inq.BranchName,
				IncomingSource: inq.IncomingSource,
				CreatedAt:      inq.CreatedAt,
			},
			Personal: entity.DataPersonal{
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
			Spouse: entity.DataSpouse{
				SpouseIDNumber:     inq.SpouseIDNumber,
				SpouseLegalName:    inq.SpouseLegalName,
				SpouseCompanyName:  inq.SpouseCompanyName,
				SpouseCompanyPhone: inq.SpouseCompanyPhone,
				SpouseMobilePhone:  inq.SpouseMobilePhone,
				SpouseProfession:   inq.SpouseProfession,
			},
			Employment: entity.DataEmployment{
				EmploymentSinceMonth:  inq.EmploymentSinceMonth,
				EmploymentSinceYear:   inq.EmploymentSinceYear,
				CompanyName:           inq.CompanyName,
				MonthlyFixedIncome:    inq.MonthlyFixedIncome,
				MonthlyVariableIncome: inq.MonthlyVariableIncome,
				SpouseIncome:          inq.SpouseIncome,
				ProfessionID:          inq.ProfessionID,
				JobTypeID:             inq.JobTypeID,
				IndustryTypeID:        string(industry_type),
			},
			ItemApk: entity.DataItemApk{
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
			Surveyor: entity.DataSurveyor{
				Destination:  inq.Destination,
				RegDate:      inq.RegDate,
				AssignDate:   inq.AssignDate,
				SurveyorName: inq.SurveyorName,
				ResultDate:   inq.ResultDate,
				Status:       inq.Status,
			},
			Emcon: entity.DataEmcon{
				EmconName:        inq.EmconName,
				Relationship:     inq.Relationship,
				EmconMobilePhone: inq.EmconMobilePhone,
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
