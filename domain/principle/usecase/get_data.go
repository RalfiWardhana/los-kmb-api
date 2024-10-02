package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) GetDataPrinciple(ctx context.Context, req request.PrincipleGetData, accessToken string) (data map[string]interface{}, err error) {

	switch req.Context {

	case "Domisili":

		principleStepOne, err := u.repository.GetPrincipleStepOne(req.ProspectID)

		if err != nil {
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

		principleStepTwo, err := u.repository.GetPrincipleStepTwo(req.ProspectID)

		if err != nil {
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

		principleStepThree, err := u.repository.GetPrincipleStepThree(req.ProspectID)

		if err != nil {
			return data, err
		}

		principleStepOne, err := u.repository.GetPrincipleStepOne(req.ProspectID)
		if err != nil {
			return data, err
		}

		var (
			marsevLoanAmountRes response.MarsevLoanAmountResponse
			assetList           response.AssetList
		)

		timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
		header := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
		}

		// get loan amount
		payload := request.ReqMarsevLoanAmount{
			BranchID:      principleStepOne.BranchID,
			OTR:           5000000,
			MaxLTV:        50,
			IsRecalculate: false,
			LoanAmount:    2000000,
			DPAmount:      0,
		}

		param, _ := json.Marshal(payload)

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

		}

		timeOut, _ = strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

		request, _ := json.Marshal(map[string]interface{}{
			"branch_id": principleStepOne.BranchID,
			"lob_id":    11,
			"page_size": 10,
			"search":    principleStepOne.AssetCode,
		})

		resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_ASSET_URL"), request, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, req.ProspectID, accessToken)

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

		data = map[string]interface{}{
			"is_psa":           marsevLoanAmountRes.Data.IsPsa,
			"manufacture_year": principleStepOne.ManufactureYear,
			"finance_purpose":  principleStepThree.FinancePurpose,
			"model":            assetList.Records[0].AssetDisplay,
			"brand":            assetList.Records[0].Brand,
			"type":             assetList.Records[0].Model,
		}

		return data, err

	case "Emergency":

		principleStepFour, err := u.repository.GetPrincipleEmergencyContact(req.ProspectID)

		if err != nil {
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

	}

	return
}
