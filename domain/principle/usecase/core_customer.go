package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) PrincipleCoreCustomer(ctx context.Context, prospectID string, req request.PrincipleCoreCustomer, accessToken string) (err error) {

	var (
		principleStepOne             entity.TrxPrincipleStepOne
		principleStepTwo             entity.TrxPrincipleStepTwo
		principleStepThree           entity.TrxPrincipleStepThree
		trxPrincipleEmergencyContact entity.TrxPrincipleEmergencyContact
		customerValidateRes          response.CustomerDomainValidate
		insertCustomerRes            response.CustomerDomainInsert
		updateCustomerTransactionRes response.CustomerDomainUpdateCustomerTransaction
		wg                           sync.WaitGroup
		errChan                      = make(chan error, 4)
	)

	wg.Add(4)
	go func() {
		defer wg.Done()
		principleStepOne, err = u.repository.GetPrincipleStepOne(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		principleStepTwo, err = u.repository.GetPrincipleStepTwo(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		principleStepThree, err = u.repository.GetPrincipleStepThree(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		trxPrincipleEmergencyContact, err = u.repository.GetPrincipleEmergencyContact(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return err
	}

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": accessToken,
	}

	// validate data customer
	birthDateStr := principleStepTwo.BirthDate.Format(constant.FORMAT_DATE)
	param, _ := json.Marshal(map[string]interface{}{
		"id_number":           principleStepTwo.IDNumber,
		"legal_name":          principleStepTwo.LegalName,
		"birth_date":          birthDateStr,
		"surgate_mother_name": principleStepTwo.SurgateMotherName,
		"mobile_phone":        principleStepTwo.MobilePhone,
	})

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("CUSTOMER_V3_BASE_URL")+"/api/v3/customer/validate-data", param, header, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 400 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Customer Validate Data Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &customerValidateRes); err != nil {
			return
		}
	}

	if customerValidateRes.Data.CustomerID > 0 || customerValidateRes.Data.KPMID > 0 {
		trxPrincipleEmergencyContact.CustomerID = customerValidateRes.Data.CustomerID
		trxPrincipleEmergencyContact.KPMID = customerValidateRes.Data.KPMID

		err = u.repository.SavePrincipleEmergencyContact(trxPrincipleEmergencyContact, principleStepThree.IDNumber)
		if err != nil {
			return
		}
	}

	// insert data customer
	isInsertCore := false
	if customerValidateRes.Data.CustomerID == 0 {
		isInsertCore = true
	}

	spouseBirthDateStr := ""
	if principleStepTwo.SpouseBirthDate != nil {
		spouseBirthDateStr = principleStepTwo.SpouseBirthDate.(time.Time).Format(constant.FORMAT_DATE)
	}

	paramInsertCust := map[string]interface{}{
		"is_insert_core":       isInsertCore,
		"prospect_id":          prospectID,
		"no_kk":                "",
		"lob_id":               constant.LOBID_KMB,
		"id_number":            principleStepTwo.IDNumber,
		"legal_name":           principleStepTwo.LegalName,
		"full_name":            principleStepTwo.FullName,
		"birth_date":           birthDateStr,
		"birth_place":          principleStepTwo.BirthPlace,
		"gender":               principleStepTwo.Gender,
		"profession_id":        principleStepTwo.ProfessionID,
		"mobile_phone":         principleStepTwo.MobilePhone,
		"marital_status_id":    principleStepTwo.MaritalStatus,
		"surgate_mother_name":  principleStepTwo.SurgateMotherName,
		"personal_npwp_number": "",
		"ktp_media_url":        principleStepTwo.KtpPhoto,
		"kk_media_url":         "",
		"selfie_media_url":     principleStepTwo.SelfiePhoto,
		"npwp_media_url":       "",
		"spouse":               nil,
	}

	if principleStepTwo.MaritalStatus == constant.MARRIED {
		paramInsertCust["spouse"] = map[string]interface{}{
			"id_number":            principleStepTwo.SpouseIDNumber,
			"full_name":            principleStepTwo.SpouseFullName,
			"mobile_phone":         principleStepTwo.SpouseMobilePhone,
			"birth_date":           spouseBirthDateStr,
			"birth_place":          principleStepTwo.SpouseBirthPlace,
			"gender":               principleStepTwo.SpouseGender,
			"surgate_mother_name":  principleStepTwo.SpouseSurgateMotherName,
			"personal_npwp_number": "",
			"ktp_media_url":        "",
			"npwp_media_url":       "",
		}
	}

	param, _ = json.Marshal(paramInsertCust)

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("CUSTOMER_V3_BASE_URL")+"/api/v3/customer/transaction", param, header, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 400 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Insert Customer Data Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &insertCustomerRes); err != nil {
			return
		}
	}

	if insertCustomerRes.Data.CustomerID > 0 && insertCustomerRes.Data.CustomerID != customerValidateRes.Data.CustomerID {
		trxPrincipleEmergencyContact.CustomerID = insertCustomerRes.Data.CustomerID

		err = u.repository.SavePrincipleEmergencyContact(trxPrincipleEmergencyContact, principleStepThree.IDNumber)
		if err != nil {
			return
		}
	}

	// update data customer transaction
	customerPersonal := map[string]interface{}{
		"birth_place":             principleStepTwo.BirthPlace,
		"full_name":               principleStepTwo.FullName,
		"gender":                  principleStepTwo.Gender,
		"mobile_phone":            principleStepTwo.MobilePhone,
		"education":               principleStepTwo.Education,
		"marital_status":          principleStepTwo.MaritalStatus,
		"home_status":             principleStepOne.HomeStatus,
		"stay_since_month":        principleStepOne.StaySinceMonth,
		"stay_since_year":         principleStepOne.StaySinceYear,
		"profession_id":           principleStepTwo.ProfessionID,
		"job_type":                principleStepTwo.JobType,
		"job_position":            principleStepTwo.JobPosition,
		"industry_type_id":        principleStepTwo.IndustryTypeID,
		"employment_since_month":  principleStepTwo.EmploymentSinceMonth,
		"employment_since_year":   principleStepTwo.EmploymentSinceYear,
		"monthly_fixed_income":    principleStepTwo.MonthlyFixedIncome,
		"spouse_income":           principleStepTwo.SpouseIncome,
		"monthly_variable_income": principleStepThree.MonthlyVariableIncome,
	}

	paramUpdateCustTransaction := map[string]interface{}{
		"transaction":       "APK_AKK",
		"prospect_id":       prospectID,
		"customer_personal": customerPersonal,
		"customer_emcon": map[string]interface{}{
			"emergency_contact_mobile_phone": trxPrincipleEmergencyContact.MobilePhone,
			"emergency_contact_name":         trxPrincipleEmergencyContact.Name,
			"emergency_contact_relationship": trxPrincipleEmergencyContact.Relationship,
		},
		"customer_address": map[string]interface{}{
			"company_address":                   principleStepTwo.CompanyAddress,
			"company_area_phone":                principleStepTwo.CompanyAreaPhone,
			"company_city":                      principleStepTwo.CompanyCity,
			"company_kecamatan":                 principleStepTwo.CompanyKecamatan,
			"company_kelurahan":                 principleStepTwo.CompanyKelurahan,
			"company_phone":                     principleStepTwo.CompanyPhone,
			"company_province":                  principleStepTwo.CompanyProvince,
			"company_rt":                        principleStepTwo.CompanyRT,
			"company_rw":                        principleStepTwo.CompanyRW,
			"company_zip_code":                  principleStepTwo.CompanyZipCode,
			"legal_address":                     principleStepTwo.LegalAddress,
			"legal_area_phone":                  principleStepTwo.LegalAreaPhone,
			"legal_city":                        principleStepTwo.LegalCity,
			"legal_kecamatan":                   principleStepTwo.LegalKecamatan,
			"legal_kelurahan":                   principleStepTwo.LegalKelurahan,
			"legal_province":                    principleStepTwo.LegalProvince,
			"legal_phone":                       principleStepTwo.LegalPhone,
			"legal_rt":                          principleStepTwo.LegalRT,
			"legal_rw":                          principleStepTwo.LegalRW,
			"legal_zip_code":                    principleStepTwo.LegalZipCode,
			"residence_address":                 principleStepOne.ResidenceAddress,
			"residence_area_phone":              principleStepOne.ResidenceAreaPhone,
			"residence_city":                    principleStepOne.ResidenceCity,
			"residence_kecamatan":               principleStepOne.ResidenceKecamatan,
			"residence_kelurahan":               principleStepOne.ResidenceKelurahan,
			"residence_phone":                   principleStepOne.ResidencePhone,
			"residence_rt":                      principleStepOne.ResidenceRT,
			"residence_rw":                      principleStepOne.ResidenceRW,
			"residence_zip_code":                principleStepOne.ResidenceZipCode,
			"emergency_contact_address":         trxPrincipleEmergencyContact.CompanyStreetName + " " + trxPrincipleEmergencyContact.HomeNumber,
			"emergency_contact_city":            trxPrincipleEmergencyContact.City,
			"emergency_contact_home_phone":      trxPrincipleEmergencyContact.Phone,
			"emergency_contact_home_phone_area": trxPrincipleEmergencyContact.AreaPhone,
			"emergency_contact_kecamatan":       trxPrincipleEmergencyContact.Kecamatan,
			"emergency_contact_kelurahan":       trxPrincipleEmergencyContact.Kelurahan,
			"emergency_contact_province":        trxPrincipleEmergencyContact.Province,
			"emergency_contact_rt":              trxPrincipleEmergencyContact.Rt,
			"emergency_contact_rw":              trxPrincipleEmergencyContact.Rw,
			"emergency_contact_zip_code":        trxPrincipleEmergencyContact.ZipCode,
		},
		"customer_photo": map[string]interface{}{
			"ktp_media_url":    principleStepTwo.KtpPhoto,
			"selfie_media_url": principleStepTwo.SelfiePhoto,
		},
		"user_information": map[string]interface{}{
			"user_id":    req.UserInformation.UserID,
			"user_title": req.UserInformation.UserTitle,
		},
	}

	if principleStepTwo.MaritalStatus == constant.MARRIED {
		paramUpdateCustTransaction["customer_spouse"] = map[string]interface{}{
			"id_number":           principleStepTwo.SpouseIDNumber,
			"birth_date":          spouseBirthDateStr,
			"birth_place":         principleStepTwo.SpouseBirthPlace,
			"full_name":           principleStepTwo.SpouseFullName,
			"gender":              principleStepTwo.SpouseGender,
			"mobile_phone":        principleStepTwo.SpouseMobilePhone,
			"surgate_mother_name": principleStepTwo.SpouseSurgateMotherName,
		}
	}

	param, _ = json.Marshal(paramUpdateCustTransaction)

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("CUSTOMER_V3_BASE_URL")+"/api/v3/customer/transaction/"+prospectID, param, header, constant.METHOD_PUT, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 400 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Update Customer Transaction Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &updateCustomerTransactionRes); err != nil {
			return
		}
	}

	return
}
