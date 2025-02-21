package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) GetDataPrinciple(ctx context.Context, req request.PrincipleGetData, accessToken string) (data map[string]interface{}, err error) {

	switch req.Context {

	case "Domisili":

		principleStepOne, _ := u.repository.GetPrincipleStepOne(req.ProspectID)

		if req.KPMID > 0 && req.KPMID != principleStepOne.KPMID {
			err = errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match")
			return data, err
		}

		data = map[string]interface{}{
			"id_number":            principleStepOne.IDNumber,
			"id_number_spouse":     principleStepOne.SpouseIDNumber,
			"residence_address":    principleStepOne.ResidenceAddress,
			"residence_rt":         principleStepOne.ResidenceRT,
			"residence_rw":         principleStepOne.ResidenceRW,
			"residence_province":   principleStepOne.ResidenceProvince,
			"residence_city":       principleStepOne.ResidenceCity,
			"residence_kecamatan":  principleStepOne.ResidenceKecamatan,
			"residence_kelurahan":  principleStepOne.ResidenceKelurahan,
			"residence_zipcode":    principleStepOne.ResidenceZipCode,
			"residence_area_phone": principleStepOne.ResidenceAreaPhone,
			"residence_phone":      principleStepOne.ResidencePhone,
			"home_status":          principleStepOne.HomeStatus,
			"stay_since_year":      principleStepOne.StaySinceYear,
			"stay_since_month":     principleStepOne.StaySinceMonth,
		}

		return data, err

	case "Pemohon":

		principleStepOne, _ := u.repository.GetPrincipleStepOne(req.ProspectID)
		principleStepTwo, _ := u.repository.GetPrincipleStepTwo(req.ProspectID)

		if req.KPMID > 0 && req.KPMID != principleStepOne.KPMID {
			err = errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match")
			return data, err
		}

		data = map[string]interface{}{
			"id_number":                  principleStepTwo.IDNumber,
			"legal_name":                 principleStepTwo.LegalName,
			"mobile_phone":               principleStepTwo.MobilePhone,
			"full_name":                  principleStepTwo.FullName,
			"birth_date":                 principleStepTwo.BirthDate,
			"birth_place":                principleStepTwo.BirthPlace,
			"surgate_mother_name":        principleStepTwo.SurgateMotherName,
			"gender":                     principleStepTwo.Gender,
			"spouse_id_number":           principleStepTwo.SpouseIDNumber,
			"legal_address":              principleStepTwo.LegalAddress,
			"legal_rt":                   principleStepTwo.LegalRT,
			"legal_rw":                   principleStepTwo.LegalRW,
			"legal_province":             principleStepTwo.LegalProvince,
			"legal_city":                 principleStepTwo.LegalCity,
			"legal_kecamatan":            principleStepTwo.LegalKecamatan,
			"legal_kelurahan":            principleStepTwo.LegalKelurahan,
			"legal_zipcode":              principleStepTwo.LegalZipCode,
			"legal_area_phone":           principleStepTwo.LegalAreaPhone,
			"legal_phone":                principleStepTwo.LegalPhone,
			"company_address":            principleStepTwo.CompanyAddress,
			"company_rt":                 principleStepTwo.CompanyRT,
			"company_rw":                 principleStepTwo.CompanyRW,
			"company_province":           principleStepTwo.CompanyProvince,
			"company_city":               principleStepTwo.CompanyCity,
			"company_kecamatan":          principleStepTwo.CompanyKecamatan,
			"company_kelurahan":          principleStepTwo.CompanyKelurahan,
			"company_zipcode":            principleStepTwo.CompanyZipCode,
			"company_area_phone":         principleStepTwo.CompanyAreaPhone,
			"company_phone":              principleStepTwo.CompanyPhone,
			"monthly_fixed_income":       principleStepTwo.MonthlyFixedIncome,
			"marital_status":             principleStepTwo.MaritalStatus,
			"spouse_income":              principleStepTwo.SpouseIncome,
			"selfie_photo":               principleStepTwo.SelfiePhoto,
			"ktp_photo":                  principleStepTwo.KtpPhoto,
			"spouse_full_name":           principleStepTwo.SpouseFullName,
			"spouse_birth_date":          principleStepTwo.SpouseBirthDate,
			"spouse_birth_place":         principleStepTwo.SpouseBirthPlace,
			"spouse_gender":              principleStepTwo.SpouseGender,
			"spouse_legal_name":          principleStepTwo.SpouseLegalName,
			"spouse_mobile_phone":        principleStepTwo.SpouseMobilePhone,
			"spouse_surgate_mother_name": principleStepTwo.SpouseSurgateMotherName,
			"employment_since_month":     principleStepTwo.EmploymentSinceMonth,
			"employment_since_year":      principleStepTwo.EmploymentSinceYear,
			"email":                      principleStepTwo.Email,
			"religion":                   principleStepTwo.Religion,
		}

		return data, err

	case "Biaya":

		var (
			wg                 sync.WaitGroup
			errChan            = make(chan error, 4)
			principleStepOne   entity.TrxPrincipleStepOne
			principleStepTwo   entity.TrxPrincipleStepTwo
			principleStepThree entity.TrxPrincipleStepThree
			filteringKMB       entity.FilteringKMB
			dealer             string
		)

		wg.Add(4)

		go func() {
			defer wg.Done()
			principleStepOne, _ = u.repository.GetPrincipleStepOne(req.ProspectID)

			if err != nil {
				errChan <- err
			}
		}()

		go func() {
			defer wg.Done()
			principleStepTwo, _ = u.repository.GetPrincipleStepTwo(req.ProspectID)
			if err != nil {
				errChan <- err
			}
		}()

		go func() {
			defer wg.Done()
			principleStepThree, _ = u.repository.GetPrincipleStepThree(req.ProspectID)

			if err != nil {
				errChan <- err
			}
		}()

		go func() {
			defer wg.Done()
			filteringKMB, err = u.repository.GetFilteringResult(req.ProspectID)

			if err != nil {
				errChan <- err
			}
		}()

		go func() {
			wg.Wait()
			close(errChan)
		}()

		if err := <-errChan; err != nil {
			return data, err
		}

		if req.KPMID > 0 && req.KPMID != principleStepOne.KPMID {
			err = errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match")
			return data, err
		}

		var (
			marsevLoanAmountRes    response.MarsevLoanAmountResponse
			assetList              response.AssetList
			marsevFilterProgramRes response.MarsevFilterProgramResponse
		)

		timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
		header := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
		}

		// get loan amount
		payloadMaxLoan := request.ReqMarsevLoanAmount{
			BranchID:      principleStepOne.BranchID,
			OTR:           5000000,
			MaxLTV:        50,
			IsRecalculate: false,
			LoanAmount:    2000000,
			DPAmount:      0,
		}

		param, _ := json.Marshal(payloadMaxLoan)

		resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_LOAN_AMOUNT_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, req.ProspectID, accessToken)
		if err != nil {
			return data, err
		}

		if resp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Loan Amount Error")
			return data, err
		}

		if resp.StatusCode() == 200 {
			if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &marsevLoanAmountRes); err != nil {
				return data, err
			}

			dealer = "NON PSA"
			if marsevLoanAmountRes.Data.IsPsa {
				dealer = "PSA"
			}

		}

		timeOut, _ = strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

		payloadAsset, _ := json.Marshal(map[string]interface{}{
			"branch_id": principleStepOne.BranchID,
			"lob_id":    11,
			"page_size": 10,
			"search":    principleStepOne.AssetCode,
		})

		resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_ASSET_URL"), payloadAsset, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, req.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout")
			return data, err
		}

		if resp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
			return data, err
		}

		json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &assetList)

		if len(assetList.Records) == 0 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
			return data, err
		}

		var categoryId string

		if len(assetList.Records) > 0 {
			categoryId = assetList.Records[0].CategoryID
		}

		if req.FinancePurpose != "" {
			// get marketing program
			bpkbStatusCode := "DN"
			if strings.Contains(os.Getenv("NAMA_SAMA"), principleStepOne.BPKBName) {
				bpkbStatusCode = "SN"
			}

			var customerStatus string
			if filteringKMB.CustomerStatus == nil {
				customerStatus = constant.STATUS_KONSUMEN_NEW
			} else {
				customerStatus = filteringKMB.CustomerStatus.(string)
			}

			var customerSegment string
			if filteringKMB.CustomerSegment == nil {
				customerSegment = constant.RO_AO_REGULAR
			} else {
				customerSegment = filteringKMB.CustomerSegment.(string)
			}

			customerType := utils.CapitalizeEachWord(customerStatus)
			if customerStatus != constant.STATUS_KONSUMEN_NEW {
				customerType = constant.STATUS_KONSUMEN_RO_AO + " " + utils.CapitalizeEachWord(customerSegment)
				if customerSegment == constant.RO_AO_REGULAR {
					customerType = constant.STATUS_KONSUMEN_RO_AO + " Standard"
				}
			}

			manufactureYear, _ := strconv.Atoi(principleStepOne.ManufactureYear)

			financeType := "PM"
			if req.FinancePurpose == constant.FINANCE_PURPOSE_MODAL_KERJA {
				financeType = "PMK"
			}

			payloadFilterProgram := request.ReqMarsevFilterProgram{
				Page:                   1,
				Limit:                  10,
				BranchID:               principleStepOne.BranchID,
				FinancingTypeCode:      financeType,
				CustomerOccupationCode: principleStepTwo.ProfessionID,
				BpkbStatusCode:         bpkbStatusCode,
				SourceApplication:      constant.MARSEV_SOURCE_APPLICATION_KPM,
				CustomerType:           customerType,
				AssetUsageTypeCode:     "C",
				AssetCategory:          categoryId,
				AssetBrand:             principleStepOne.Brand,
				AssetYear:              manufactureYear,
				LoanAmount:             2000000,
				SalesMethodID:          5,
			}

			param, _ = json.Marshal(payloadFilterProgram)

			resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_FILTER_PROGRAM_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, req.ProspectID, accessToken)
			if err != nil {
				return data, err
			}

			if resp.StatusCode() != 200 {
				err = errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Filter Program Error")
				return data, err
			}

			if resp.StatusCode() == 200 {
				if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &marsevFilterProgramRes); err != nil {
					return data, err
				}
			}
		}

		data = map[string]interface{}{
			"dealer":           dealer,
			"is_psa":           marsevLoanAmountRes.Data.IsPsa,
			"manufacture_year": principleStepOne.ManufactureYear,
			"finance_purpose":  principleStepThree.FinancePurpose,
			"model":            assetList.Records[0].AssetDisplay,
			"brand":            assetList.Records[0].Brand,
			"type":             assetList.Records[0].Model,
		}

		if req.FinancePurpose != "" {
			data["tenors"] = make([]interface{}, 0)
			if len(marsevFilterProgramRes.Data) > 0 {
				data["tenors"] = marsevFilterProgramRes.Data[0].Tenors
			}
		}

		return data, err

	case "Emergency":

		principleStepOne, _ := u.repository.GetPrincipleStepOne(req.ProspectID)
		principleStepFour, _ := u.repository.GetPrincipleEmergencyContact(req.ProspectID)

		if req.KPMID > 0 && req.KPMID != principleStepOne.KPMID {
			err = errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match")
			return data, err
		}

		data = map[string]interface{}{
			"name":         principleStepFour.Name,
			"relationship": principleStepFour.Relationship,
			"mobile_phone": principleStepFour.MobilePhone,
			"address":      principleStepFour.Address,
			"rt":           principleStepFour.Rt,
			"rw":           principleStepFour.Rw,
			"kelurahan":    principleStepFour.Kelurahan,
			"kecamatan":    principleStepFour.Kecamatan,
			"city":         principleStepFour.City,
			"province":     principleStepFour.Province,
			"zip_code":     principleStepFour.ZipCode,
			"area_phone":   principleStepFour.AreaPhone,
			"phone":        principleStepFour.Phone,
		}

		return data, err

	case "Readjust":

		trxKPM, err := u.repository.GetTrxKPM(req.ProspectID)
		if err != nil {
			return data, err
		}

		if req.KPMID > 0 && req.KPMID != trxKPM.KPMID {
			err = errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match")
			return data, err
		}

		birthDate := trxKPM.BirthDate.Format(constant.FORMAT_DATE)
		var spouseBirthDate string
		if date, ok := trxKPM.SpouseBirthDate.(time.Time); ok {
			spouseBirthDate = date.Format(constant.FORMAT_DATE)
		}

		var assetList response.AssetList
		timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

		payloadAsset, _ := json.Marshal(map[string]interface{}{
			"branch_id": trxKPM.BranchID,
			"lob_id":    11,
			"page_size": 10,
			"search":    trxKPM.AssetCode,
		})

		resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_ASSET_URL"), payloadAsset, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, req.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout")
			return data, err
		}

		if resp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
			return data, err
		}

		json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &assetList)

		if len(assetList.Records) == 0 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
			return data, err
		}

		var assetDisplay string
		if len(assetList.Records) > 0 {
			assetDisplay = assetList.Records[0].AssetDisplay
		}

		data = map[string]interface{}{
			"id_number":                  trxKPM.IDNumber,
			"legal_name":                 trxKPM.LegalName,
			"mobile_phone":               trxKPM.MobilePhone,
			"email":                      trxKPM.Email,
			"birth_place":                trxKPM.BirthPlace,
			"birth_date":                 birthDate,
			"surgate_mother_name":        trxKPM.SurgateMotherName,
			"gender":                     trxKPM.Gender,
			"residence_address":          trxKPM.ResidenceAddress,
			"residence_rt":               trxKPM.ResidenceRT,
			"residence_rw":               trxKPM.ResidenceRW,
			"residence_province":         trxKPM.ResidenceProvince,
			"residence_city":             trxKPM.ResidenceCity,
			"residence_kecamatan":        trxKPM.ResidenceKecamatan,
			"residence_kelurahan":        trxKPM.ResidenceKelurahan,
			"residence_zipcode":          trxKPM.ResidenceZipCode,
			"branch_id":                  trxKPM.BranchID,
			"asset_code":                 trxKPM.AssetCode,
			"asset_display":              assetDisplay,
			"manufacture_year":           trxKPM.ManufactureYear,
			"license_plate":              trxKPM.LicensePlate,
			"asset_usage_type_code":      trxKPM.AssetUsageTypeCode,
			"bpkb_name_type":             trxKPM.BPKBName,
			"owner_asset":                trxKPM.OwnerAsset,
			"loan_amount":                trxKPM.LoanAmount,
			"max_loan_amount":            trxKPM.MaxLoanAmount,
			"tenor":                      trxKPM.Tenor,
			"installment_amount":         trxKPM.InstallmentAmount,
			"num_of_dependence":          trxKPM.NumOfDependence,
			"marital_status":             trxKPM.MaritalStatus,
			"spouse_id_number":           trxKPM.SpouseIDNumber,
			"spouse_legal_name":          trxKPM.SpouseLegalName,
			"spouse_birth_date":          spouseBirthDate,
			"spouse_birth_place":         trxKPM.SpouseBirthPlace,
			"spouse_surgate_mother_name": trxKPM.SpouseSurgateMotherName,
			"spouse_mobile_phone":        trxKPM.SpouseMobilePhone,
			"education":                  trxKPM.Education,
			"profession_id":              trxKPM.ProfessionID,
			"job_type":                   trxKPM.JobType,
			"job_position":               trxKPM.JobPosition,
			"employment_since_month":     trxKPM.EmploymentSinceMonth,
			"employment_since_year":      trxKPM.EmploymentSinceYear,
			"monthly_fixed_income":       trxKPM.MonthlyFixedIncome,
			"spouse_income":              trxKPM.SpouseIncome,
			"chassis_number":             trxKPM.NoChassis,
			"home_status":                trxKPM.HomeStatus,
			"stay_since_year":            trxKPM.StaySinceYear,
			"stay_since_month":           trxKPM.StaySinceMonth,
			"ktp_photo":                  trxKPM.KtpPhoto,
			"selfie_photo":               trxKPM.SelfiePhoto,
			"af":                         trxKPM.AF,
			"ntf":                        trxKPM.NTF,
			"otr":                        trxKPM.OTR,
			"down_payment_amount":        trxKPM.DPAmount,
			"admin_fee":                  trxKPM.AdminFee,
			"dealer":                     trxKPM.Dealer,
			"asset_category_id":          trxKPM.AssetCategoryID,
			"kpm_id":                     trxKPM.KPMID,
			"result_pefindo":             trxKPM.ResultPefindo,
			"baki_debet":                 trxKPM.BakiDebet,
			"readjust_context":           trxKPM.ReadjustContext,
		}

		return data, err

	}

	return
}
