package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"los-kmb-api/domain/kmb/interfaces"
	entity "los-kmb-api/models/dupcheck"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
	"los-kmb-api/models/other"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

type (
	multiUsecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		usecase    interfaces.Usecase
	}
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
	}
)

func NewMultiUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) (interfaces.MultiUsecase, interfaces.Usecase) {
	usecase := NewUsecase(repository, httpclient)

	return &multiUsecase{
		usecase:    usecase,
		repository: repository,
		httpclient: httpclient,
	}, usecase
}

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
	}
}

func (u usecase) DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	req, _ := json.Marshal(map[string]interface{}{
		"transaction_id":      prospectID,
		"id_number":           idNumber,
		"legal_name":          legalName,
		"birth_date":          birthDate,
		"surgate_mother_name": surgateName,
	})

	custDupcheck, err := u.httpclient.EngineAPI(ctx, constant.FILTERING_LOG, os.Getenv("DUPCHECK_URL"), req, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Dupcheck Timeout")
		return
	}

	if custDupcheck.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Dupcheck Error")
		return
	}

	json.Unmarshal([]byte(jsoniter.Get(custDupcheck.Body(), "data").ToString()), &spDupcheck)

	return

}

func (u usecase) CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error) {

	if spDupcheck == (response.SpDupCekCustomerByID{}) {
		statusKonsumen = constant.STATUS_KONSUMEN_NEW
		return
	}

	if (spDupcheck.TotalInstallment <= 0 && spDupcheck.RRDDate != nil) || (spDupcheck.TotalInstallment > 0 && spDupcheck.RRDDate != nil && spDupcheck.NumberOfPaidInstallment == nil) {
		statusKonsumen = constant.STATUS_KONSUMEN_RO
		return

	} else if spDupcheck.TotalInstallment > 0 {
		statusKonsumen = constant.STATUS_KONSUMEN_AO
		return

	} else {
		statusKonsumen = constant.STATUS_KONSUMEN_NEW
		return
	}

}

func (u usecase) VehicleCheck(manufactureYear string) (data response.UsecaseApi, err error) {

	config, err := u.repository.GetDupcheckConfig()

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Parameterize Config")
		return
	}

	var configValue response.DupcheckConfig

	json.Unmarshal([]byte(config.Value), &configValue)

	currentYear, _ := strconv.Atoi(time.Now().Format("2006-01-02")[0:4])
	BPKBYear, _ := strconv.Atoi(manufactureYear)

	ageVehicle := currentYear - BPKBYear

	if ageVehicle <= configValue.Data.VehicleAge {
		data.Result = constant.DECISION_PASS
		data.Code = constant.CODE_VEHICLE_SESUAI
		data.Reason = constant.REASON_VEHICLE_SESUAI
		return

	} else {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_VEHICLE_AGE_MAX
		data.Reason = fmt.Sprintf("%s %d Tahun", constant.REASON_VEHICLE_AGE_MAX, configValue.Data.VehicleAge)
		return
	}

}

func (u usecase) GetLatestPaidInstallment(ctx context.Context, req request.ReqLatestPaidInstallment, prospectID string, accessToken string) (data response.LatestPaidInstallmentData, err error) {

	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_LATEST_PAID_INSTALLMENT"))

	if dummy {
		dummyLatestPaidInstallment, _ := u.repository.GetDummyLatestPaidInstallment(req.IDNumber)

		var latestPaidInstallment response.LatestPaidInstallment

		json.Unmarshal([]byte(dummyLatestPaidInstallment.Response), &latestPaidInstallment)

		data = latestPaidInstallment.Data

	} else {

		var dupcheckMDM *resty.Response
		dupcheckMDM, err = u.httpclient.EngineAPI(ctx, constant.DUPCHECK_LOG, fmt.Sprintf("%s/%s/3", os.Getenv("DUPCHECK_GET_LATEST_PAID_INSTALLMENT"), req.CustomerID), nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, prospectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Dupcheck MDM Latest Paid Installment Timeout")
			return
		}

		if dupcheckMDM.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Dupcheck MDM Latest Paid Installment Error")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(dupcheckMDM.Body(), "data").ToString()), &data)
	}

	return
}

func CentralizeLog(logFile, message, status, context string, isError bool, logger other.CustomLog) {

	config.SetCustomLog(logFile, isError, map[string]interface{}{
		"status":  status,
		"message": message,
		"data":    logger,
	}, context)

}

func (u usecase) CustomerDomainGetData(ctx context.Context, req request.ReqCustomerDomain, prospectID string, accessToken string) (customerDomainData response.CustomerDomainData, err error) {

	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_CUSTOMER_DOMAIN_GET_DATA"))

	if dummy {
		dummyCustomerDomain, _ := u.repository.GetDummyCustomerDomain(req.IDNumber)

		var customerDomain response.CustomerDomain

		json.Unmarshal([]byte(dummyCustomerDomain.Response), &customerDomain)

		customerDomainData = customerDomain.Data

	} else {

		param, _ := json.Marshal(req)

		header := map[string]string{
			"Authorization": accessToken,
		}

		url := os.Getenv("CUSTOMER_DOMAIN_GET_DATA")

		resp, err := u.httpclient.EngineAPI(ctx, constant.DUPCHECK_LOG, url, param, header, constant.METHOD_POST, false, 0, 60, prospectID, accessToken)

		if err != nil && resp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Customer Domain")
			return customerDomainData, err
		}

		var customerDomain response.CustomerDomain

		json.Unmarshal(resp.Body(), &customerDomain)

		customerDomainData = customerDomain.Data

		CentralizeLog(constant.DUPCHECK_LOG, "Customer Domain", constant.MESSAGE_SUCCESS, "GET_DATA", false, other.CustomLog{Info: other.ResultLog{Request: req, Response: customerDomain}})
	}

	return
}

func (u usecase) NokaBanned30D(req request.DupcheckApi) (data response.RejectionNoka, err error) {

	data.Code = constant.CODE_REJECTION_OK
	data.Result = constant.RESULT_REJECTION_OK
	data.Reason = constant.REASON_REJECTION_OK

	nokaBanned30D, err := u.repository.GetLatestBannedRejectionNoka(req.RangkaNo)
	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Latest Banned Rejection Noka")
		return
	}

	data.IsBannedActive = false
	if nokaBanned30D != (response.DupcheckRejectionNokaNosin{}) && nokaBanned30D.IsBanned == 1 {
		bannedDate := nokaBanned30D.CreatedAt
		dueDate := bannedDate.AddDate(0, 0, constant.DAY_RANGE_BANNED_REJECT_NOKA)
		dueDateString := dueDate.Format("2006-01-02")

		if time.Now().Format(constant.FORMAT_DATE) <= dueDateString {
			data.IsBannedActive = true
		}

		if data.IsBannedActive {
			data.Code = constant.CODE_REJECT_NOKA_NOSIN
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_NOKA_NOSIN
			return
		}
	}
	return
}

func (u usecase) CheckRejectionNoka(req request.DupcheckApi) (data response.RejectionNoka, err error) {

	var (
		inRejectionNoka    bool
		getHistoryReject   []entity.DupcheckRejectionPMK
		rejection          []entity.DupcheckRejectionPMK
		checkHistoryReject entity.DupcheckRejectionPMK
	)

	nokaBannedCurrentDate, err := u.repository.GetLatestRejectionNoka(req.RangkaNo)

	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Latest Rejection Noka")
		return
	}

	inRejectionNoka = false
	data.CurrentBannedNotEmpty = false
	if nokaBannedCurrentDate != (response.DupcheckRejectionNokaNosin{}) {

		// Must be simplified
		inRejectionNoka = true
		data.CurrentBannedNotEmpty = true

		data.IDNumber = nokaBannedCurrentDate.IDNumber
		data.LegalName = nokaBannedCurrentDate.LegalName
		data.BirthPlace = nokaBannedCurrentDate.BirthPlace
		data.BirthDate = nokaBannedCurrentDate.BirthDate
		data.MonthlyFixedIncome = nokaBannedCurrentDate.MonthlyFixedIncome
		data.EmploymentSinceYear = nokaBannedCurrentDate.EmploymentSinceYear
		data.EmploymentSinceMonth = nokaBannedCurrentDate.EmploymentSinceMonth
		data.StaySinceYear = nokaBannedCurrentDate.StaySinceYear
		data.StaySinceMonth = nokaBannedCurrentDate.StaySinceMonth
		data.BPKBName = nokaBannedCurrentDate.BPKBName
		data.Gender = nokaBannedCurrentDate.Gender
		data.MaritalStatus = nokaBannedCurrentDate.MaritalStatus
		data.NumOfDependence = nokaBannedCurrentDate.NumOfDependence
		data.NTF = nokaBannedCurrentDate.NTF
		data.OTRPrice = nokaBannedCurrentDate.OTRPrice
		data.LegalZipCode = nokaBannedCurrentDate.LegalZipCode
		data.Tenor = nokaBannedCurrentDate.Tenor
		data.ManufacturingYear = nokaBannedCurrentDate.ManufacturingYear
		data.ProfessionID = nokaBannedCurrentDate.ProfessionID
		data.CompanyZipCode = nokaBannedCurrentDate.CompanyZipCode
		data.HomeStatus = nokaBannedCurrentDate.HomeStatus

		data.Code = constant.CODE_REJECTION_OK
		data.Result = constant.RESULT_REJECTION_OK
		data.Reason = constant.REASON_REJECTION_OK
		return
	}

	if !inRejectionNoka {

		rejection, err = u.repository.GetAllReject(req.IDNumber)

		if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
			err = errors.New(constant.ERROR_UPSTREAM + " - Error Get All Rejection")
			return
		}

		if len(rejection) >= constant.ATTEMPT_REJECT {
			for i := 0; i < len(rejection); i++ {
				if rejection[i].RejectPMK == 1 || rejection[i].RejectDSR == 1 {
					data.Code = constant.CODE_REJECT_PMK_DSR
					data.Result = constant.DECISION_REJECT
					data.Reason = constant.REASON_REJECT_PMK_DSR
					return
				}
			}

			data.Code = constant.CODE_REJECT_NIK_KTP
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_NIK_KTP
			return

		} else {
			getHistoryReject, err = u.repository.GetHistoryRejectAttempt(req.IDNumber)

			if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Get All Rejection")
				return
			}

			if len(getHistoryReject) > 0 {
				historyResult := getHistoryReject[0]
				blackListDate := historyResult.Date

				if historyResult.RejectAttempt >= constant.ATTEMPT_REJECT_PMK_DSR {
					parsedDate, _ := time.Parse("2006-01-02", blackListDate)
					dueDate := parsedDate.AddDate(0, 0, 30)
					dueDateString := dueDate.Format("2006-01-02")

					if time.Now().Format(constant.FORMAT_DATE) < dueDateString {
						data.Code = constant.CODE_REJECT_PMK_DSR
						data.Result = constant.DECISION_REJECT
						data.Reason = constant.REASON_REJECT_PMK_DSR
						return
					}
				} else {
					date, _ := time.Parse(time.RFC3339, blackListDate)
					dateString := date.Format("2006-01-02")

					checkHistoryReject, err = u.repository.GetCheckingRejectAttempt(req.IDNumber, dateString)
					if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
						err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Checking Reject Attempt")
						return
					}

					if checkHistoryReject.RejectAttempt >= constant.ATTEMPT_REJECT_PMK_DSR {
						data.Code = constant.CODE_REJECT_PMK_DSR
						data.Result = constant.DECISION_REJECT
						data.Reason = constant.REASON_REJECT_PMK_DSR
						return
					}
				}
			}

			data.Code = constant.CODE_REJECTION_OK
			data.Result = constant.RESULT_REJECTION_OK
			return
		}
	}
	return
}

func (u usecase) CheckNoka(ctx context.Context, reqs request.DupcheckApi, nokaBanned response.RejectionNoka, accessToken string) (data response.UsecaseApi, err error) {

	var (
		nokaData entity.DupcheckRejectionNokaNosin
	)

	data.Code = constant.CODE_REJECTION_NOTIF
	data.Result = constant.RESULT_NOTIF
	data.Reason = constant.REASON_NOTIF

	// Noka Data to Save
	nokaData.Id = utils.UniqueID(15)
	nokaData.NoRangka = reqs.RangkaNo
	nokaData.NoMesin = reqs.EngineNo
	nokaData.ProspectID = reqs.ProspectID
	nokaData.IDNumber = reqs.IDNumber
	nokaData.LegalName = reqs.LegalName
	nokaData.BirthPlace = reqs.BirthPlace
	nokaData.BirthDate = reqs.BirthDate
	nokaData.MonthlyFixedIncome = reqs.MonthlyFixedIncome
	nokaData.EmploymentSinceYear = reqs.EmploymentSinceYear
	nokaData.EmploymentSinceMonth = reqs.EmploymentSinceMonth
	nokaData.StaySinceYear = reqs.StaySinceYear
	nokaData.StaySinceMonth = reqs.StaySinceMonth
	nokaData.BPKBName = reqs.BPKBName
	nokaData.Gender = reqs.Gender
	nokaData.MaritalStatus = reqs.MaritalStatus
	nokaData.NumOfDependence = reqs.NumOfDependence
	nokaData.NTF = reqs.NTF
	nokaData.OTRPrice = reqs.OTRPrice
	nokaData.LegalZipCode = reqs.LegalZipCode
	nokaData.Tenor = reqs.Tenor
	nokaData.ManufacturingYear = reqs.ManufactureYear
	nokaData.ProfessionID = reqs.ProfessionID
	nokaData.CompanyZipCode = reqs.CompanyZipCode
	nokaData.HomeStatus = reqs.HomeStatus

	// Check Current Date Banned Status
	if nokaBanned.CurrentBannedNotEmpty {
		numberOfRetry := nokaBanned.NumberOfRetry + 1

		if numberOfRetry == 1 {
			nokaData.NumberOfRetry = numberOfRetry

		} else if numberOfRetry > 1 && numberOfRetry < constant.ATTEMPT_BANNED_REJECTION_NOKA {
			nokaData.NumberOfRetry = numberOfRetry

			// Maybe could be simplified
			if nokaBanned.IDNumber != reqs.IDNumber && nokaBanned.LegalName != reqs.LegalName && nokaBanned.BirthPlace != reqs.BirthPlace && nokaBanned.BirthDate != reqs.BirthDate && nokaBanned.MonthlyFixedIncome != reqs.MonthlyFixedIncome && nokaBanned.EmploymentSinceYear != reqs.EmploymentSinceYear && nokaBanned.EmploymentSinceMonth != reqs.EmploymentSinceMonth && nokaBanned.StaySinceYear != reqs.StaySinceYear && nokaBanned.StaySinceMonth != reqs.StaySinceMonth && nokaBanned.BPKBName != reqs.BPKBName && nokaBanned.Gender != reqs.Gender && nokaBanned.MaritalStatus != reqs.MaritalStatus && nokaBanned.NumOfDependence != reqs.NumOfDependence && nokaBanned.NTF != reqs.NTF && nokaBanned.OTRPrice != reqs.OTRPrice && nokaBanned.LegalZipCode != reqs.LegalZipCode && nokaBanned.Tenor != reqs.Tenor && nokaBanned.ManufacturingYear != reqs.ManufactureYear && nokaBanned.ProfessionID != reqs.ProfessionID && nokaBanned.CompanyZipCode != reqs.CompanyZipCode && nokaBanned.HomeStatus != reqs.HomeStatus {
				// MAKSUDNYA -> KETIKA ADA DATA DALAM KETENTUAN PARAMETER KREDIT BERBEDA DENGAN INPUTAN SEBELUMNYA, MAKA NO. RANGKA TSB DI BANNED 30 HARI
				nokaBanned.IsBanned = 1
			}
		} else if numberOfRetry >= constant.ATTEMPT_BANNED_REJECTION_NOKA {
			nokaBanned.NumberOfRetry = numberOfRetry
			nokaBanned.IsBanned = 1
		}

		if err = u.repository.SaveDataNoka(nokaData); err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Error Save Data Noka Nosin")
			return
		}

		data.Code = constant.CODE_REJECT_NOKA_NOSIN
		data.Result = constant.DECISION_REJECT
		data.Reason = constant.REASON_REJECT_NOKA_NOSIN
		return
	}

	return
}

func (u usecase) CheckChassisNumber(ctx context.Context, reqs request.DupcheckApi, nokaBanned response.RejectionNoka, accessToken string) (data response.UsecaseApi, err error) {

	var (
		nokaData                       entity.DupcheckRejectionNokaNosin
		trxApiLog                      entity.TrxApiLog
		responseAgreementChassisNumber entity.AgreementChassisNumber
		response_api_log               []byte
	)

	trxApiLog.ProspectID = reqs.ProspectID
	trxApiLog.Request = os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL") + reqs.RangkaNo
	trxApiLog.DtmRequest = time.Now()

	// Hit Integrator Get Chasis Number
	dummyChassis, _ := strconv.ParseBool(os.Getenv("DUMMY_AGREEMENT_OF_CHASSIS_NUMBER"))

	if dummyChassis {
		dummyChassisAgreementNumber, _ := u.repository.GetDummyAgreementChassisNumber(reqs.IDNumber)

		var resAgreementChassisNumber response.ResAgreementChassisNumber

		json.Unmarshal([]byte(dummyChassisAgreementNumber.Response), &resAgreementChassisNumber)

		responseAgreementChassisNumber = resAgreementChassisNumber.Data

	} else {

		var hitChassisNumber *resty.Response

		hitChassisNumber, err = u.httpclient.EngineAPI(ctx, constant.DUPCHECK_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+reqs.RangkaNo, nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, reqs.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Get Agreement of Chassis Number Timeout")
			return
		}

		if hitChassisNumber.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Get Agreement of Chassis Number Error")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(hitChassisNumber.Body(), "data").ToString()), &responseAgreementChassisNumber)
	}

	response_api_log, _ = json.Marshal(responseAgreementChassisNumber)
	trxApiLog.Response = string(response_api_log)
	trxApiLog.DtmResponse = time.Now()
	trxApiLog.Type = constant.TYPE_API_LOGS_NOKA

	if err = u.repository.SaveDataApiLog(trxApiLog); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Save Data to API Logs")
		return
	}

	if responseAgreementChassisNumber == (response.AgreementChassisNumber{}) {
		return
	}

	agreement := responseAgreementChassisNumber
	if agreement.IsRegistered && agreement.IsActive && len(agreement.IDNumber) > 0 {
		listNikKonsumenDanPasangan := make([]string, 0)

		listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, reqs.IDNumber)
		if reqs.Spouse != nil && reqs.Spouse.IDNumber != "" {
			listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, reqs.Spouse.IDNumber)
		}

		if !utils.Contains(listNikKonsumenDanPasangan, agreement.IDNumber) {
			// Noka Data to Save
			nokaData.Id = utils.UniqueID(15)
			nokaData.NoRangka = reqs.RangkaNo
			nokaData.NoMesin = reqs.EngineNo
			nokaData.ProspectID = reqs.ProspectID
			nokaData.IDNumber = reqs.IDNumber
			nokaData.LegalName = reqs.LegalName
			nokaData.BirthPlace = reqs.BirthPlace
			nokaData.BirthDate = reqs.BirthDate
			nokaData.MonthlyFixedIncome = reqs.MonthlyFixedIncome
			nokaData.EmploymentSinceYear = reqs.EmploymentSinceYear
			nokaData.EmploymentSinceMonth = reqs.EmploymentSinceMonth
			nokaData.StaySinceYear = reqs.StaySinceYear
			nokaData.StaySinceMonth = reqs.StaySinceMonth
			nokaData.BPKBName = reqs.BPKBName
			nokaData.Gender = reqs.Gender
			nokaData.MaritalStatus = reqs.MaritalStatus
			nokaData.NumOfDependence = reqs.NumOfDependence
			nokaData.NTF = reqs.NTF
			nokaData.OTRPrice = reqs.OTRPrice
			nokaData.LegalZipCode = reqs.LegalZipCode
			nokaData.Tenor = reqs.Tenor
			nokaData.ManufacturingYear = reqs.ManufactureYear
			nokaData.ProfessionID = reqs.ProfessionID
			nokaData.CompanyZipCode = reqs.CompanyZipCode
			nokaData.HomeStatus = reqs.HomeStatus

			nokaData.NumberOfRetry = 0

			if err = u.repository.SaveDataNoka(nokaData); err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Save Data Noka Nosin")
				return
			}

			data.Code = constant.CODE_REJECT_CHASSIS_NUMBER
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_CHASSIS_NUMBER
		} else {
			if agreement.IDNumber == reqs.IDNumber {
				data.Code = constant.CODE_OK_CONSUMEN_MATCH
				data.Result = constant.RESULT_OK
				data.Reason = constant.REASON_OK_CONSUMEN_MATCH
			} else {
				data.Code = constant.CODE_REJECTION_FRAUD_POTENTIAL
				data.Result = constant.DECISION_REJECT
				data.Reason = constant.REASON_REJECTION_FRAUD_POTENTIAL
			}
		}
	} else {
		data.Code = constant.CODE_AGREEMENT_NOT_FOUND
		data.Result = constant.RESULT_OK
		data.Reason = constant.REASON_AGREEMENT_NOT_FOUND
	}
	return
}
