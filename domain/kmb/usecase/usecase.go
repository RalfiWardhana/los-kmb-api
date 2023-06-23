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
		err = errors.New("upstream_service_timeout - Call Dupcheck Timeout")
		return
	}

	if custDupcheck.StatusCode() != 200 {
		err = errors.New("upstream_service_error - Call Dupcheck Error")
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
		err = errors.New("upstream_service_error - Error Get Parameterize Config")
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
			err = errors.New("upstream_service_timeout - Call Dupcheck MDM Latest Paid Installment Timeout")
			return
		}

		if dupcheckMDM.StatusCode() != 200 {
			err = errors.New("upstream_service_error - Call Dupcheck MDM Latest Paid Installment Error")
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
			err = errors.New("upstream_service_error - Call Customer Domain")
			CentralizeLog(constant.DUPCHECK_LOG, "Customer Domain", constant.MESSAGE_SUCCESS, "GET_DATA", true, other.CustomLog{Info: req, Error: err.Error()})
			return customerDomainData, err
		}

		var customerDomain response.CustomerDomain

		json.Unmarshal(resp.Body(), &customerDomain)

		customerDomainData = customerDomain.Data

		CentralizeLog(constant.DUPCHECK_LOG, "Customer Domain", constant.MESSAGE_SUCCESS, "GET_DATA", false, other.CustomLog{Info: other.ResultLog{Request: req, Response: customerDomain}})
	}

	return
}

func (u usecase) RejectionNoka(noRangka, idNumber string) (data response.UsecaseApi, err error) {

	var (
		inRejectionNoka  bool
		getHistoryReject []response.DupcheckRejectionPMK
		rejection        []response.DupcheckRejectionPMK
	)

	nokaBanned30D, err := u.repository.GetLatestBannedRejectionNoka(noRangka)
	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = errors.New("upstream_service_error - Error Get Latest Banned Rejection Noka")
		return
	}

	nokaBannedCurrentDate, err := u.repository.GetLatestRejectionNoka(noRangka)
	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = errors.New("upstream_service_error - Error Get Latest Rejection Noka")
		return
	}

	inRejectionNoka = false
	if nokaBanned30D != (entity.DupcheckRejectionNokaNosin{}) || nokaBannedCurrentDate != (entity.DupcheckRejectionNokaNosin{}) {

		var IsBannedActive bool

		if nokaBanned30D != (entity.DupcheckRejectionNokaNosin{}) && nokaBanned30D.IsBanned == 1 {
			bannedDate := nokaBanned30D.CreatedAt
			dueDate := bannedDate.AddDate(0, 0, constant.DAY_RANGE_BANNED_REJECT_NOKA)
			dueDateString := dueDate.Format("2006-01-02")

			IsBannedActive = false
			if time.Now().Format(constant.FORMAT_DATE) <= dueDateString {
				IsBannedActive = true
			}
		}

		if nokaBannedCurrentDate != (entity.DupcheckRejectionNokaNosin{}) || IsBannedActive {
			inRejectionNoka = true
			data.Code = constant.CODE_REJECTION_OK
			data.Result = constant.RESULT_REJECTION_OK
			data.Reason = constant.REASON_REJECTION_OK
			return
		}
	}

	if !inRejectionNoka {

		rejection, err = u.repository.GetAllReject(idNumber)
		if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
			err = errors.New("upstream_service_error - Error Get All Rejection")
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
			getHistoryReject, err = u.repository.GetHistoryRejectAttempt(idNumber)
			if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
				err = errors.New("upstream_service_error - Error Get All Rejection")
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

					checkHistoryReject, _ := u.repository.GetCheckingRejectAttempt(idNumber, dateString)

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
