package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	elaborateInterfaces "los-kmb-api/domain/elaborate/interfaces"
	"los-kmb-api/domain/filtering/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

type (
	multiUsecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		usecase    interfaces.Usecase
	}
	usecase struct {
		repository          interfaces.Repository
		elaborateRepository elaborateInterfaces.Repository
		httpclient          httpclient.HttpClient
	}
)

func NewMultiUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, usecase interfaces.Usecase) (interfaces.MultiUsecase, interfaces.Usecase) {
	return &multiUsecase{
		usecase:    usecase,
		repository: repository,
		httpclient: httpclient,
	}, usecase
}

func NewUsecase(repository interfaces.Repository, elaborateRepository elaborateInterfaces.Repository, httpclient httpclient.HttpClient) interfaces.Usecase {
	return &usecase{
		repository:          repository,
		elaborateRepository: elaborateRepository,
		httpclient:          httpclient,
	}
}

func (u multiUsecase) Filtering(ctx context.Context, reqs request.FilteringRequest, accessToken string) (data response.DupcheckResult, err error) {

	var savedata entity.ApiDupcheckKmb

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	savedata.RequestID = requestID

	request, _ := json.Marshal(reqs)
	savedata.Request = string(request)
	savedata.ProspectID = reqs.Data.ProspectID

	if err = u.repository.SaveData(savedata); err != nil {
		err = fmt.Errorf("failed process dupcheck")
		return
	}

	checkBlacklist, err := u.usecase.FilteringBlackList(ctx, reqs, accessToken)
	if err != nil {
		err = fmt.Errorf("failed process check blacklist")
		return
	}

	arrNonBlackList := strings.Split(constant.CODE_NON_BLACKLIST, ",")

	var (
		isBlacklist     bool = true
		updateFiltering entity.ApiDupcheckKmbUpdate
	)

	for _, v := range arrNonBlackList {
		if v == checkBlacklist.Code {
			isBlacklist = false
			break
		}
	}

	if isBlacklist {
		updateFiltering.Code = checkBlacklist.Code
		updateFiltering.Reason = checkBlacklist.Reason
		updateFiltering.Decision = checkBlacklist.Decision
		data = checkBlacklist

	} else {

		konsumen, err := u.usecase.CheckStatusCategory(ctx, reqs, checkBlacklist.StatusKonsumen, accessToken)
		if err != nil {
			err = fmt.Errorf("failed fetching data customer domain")
			return konsumen, err
		}

		check_prime_priority, _ := utils.ItemExists(konsumen.KategoriStatusKonsumen, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

		// hit ke pefindo
		pefindo, isRejectClusterEF, err := u.usecase.FilteringPefindo(ctx, reqs, konsumen.StatusKonsumen, konsumen.KategoriStatusKonsumen, accessToken)
		if err != nil {
			err = fmt.Errorf("failed fetching data pefindo")
			return pefindo, err
		}

		if isRejectClusterEF {
			updateFiltering.Code = pefindo.Code
			updateFiltering.Reason = pefindo.Reason
			updateFiltering.Decision = pefindo.Decision
			data = pefindo
		}

		if check_prime_priority {
			updateFiltering.Code = konsumen.Code
			updateFiltering.Reason = konsumen.Reason
			updateFiltering.Decision = konsumen.Decision
			data = konsumen

		} else {
			updateFiltering.Code = pefindo.Code
			updateFiltering.Reason = pefindo.Reason
			updateFiltering.Decision = pefindo.Decision
			data = pefindo
		}
		data.StatusKonsumen = checkBlacklist.StatusKonsumen
		if data.StatusKonsumen != constant.STATUS_KONSUMEN_NEW {
			data.KategoriStatusKonsumen = konsumen.KategoriStatusKonsumen
		}
	}

	resp, _ := json.Marshal(data)

	updateFiltering.RequestID = savedata.RequestID
	updateFiltering.Response = string(resp)

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("failed update data filtering")
		return
	}

	return
}

func (u usecase) FilteringBlackList(ctx context.Context, reqs request.FilteringRequest, accessToken string) (result response.DupcheckResult, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	var updateFiltering entity.ApiDupcheckKmbUpdate

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	updateFiltering.RequestID = requestID

	param, _ := json.Marshal(map[string]string{
		"birth_date":          reqs.Data.BirthDate,
		"id_number":           reqs.Data.IDNumber,
		"legal_name":          reqs.Data.LegalName,
		"surgate_mother_name": reqs.Data.SurgateMotherName,
		"transaction_id":      reqs.Data.ProspectID,
	})

	var (
		dupcheck_data        response.DupCheckData
		getdupcheck          response.DataDupcheck
		dupcheck_data_spouse response.DupCheckData
		getdupcheckspouse    response.DataDupcheck
	)

	resp, err := u.httpclient.EngineAPI(ctx, constant.FILTERING_LOG, os.Getenv("DUPCHECK_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, reqs.Data.ProspectID, accessToken)

	if err != nil || resp.StatusCode() != 200 {
		err = fmt.Errorf("failed process dupcheck")
		return
	}

	if err = json.Unmarshal(resp.Body(), &dupcheck_data); err != nil {
		err = fmt.Errorf("error unmarshal dupcheck response")
		return
	}

	getdupcheck = dupcheck_data.Data
	responseskonsumen, _ := json.Marshal(getdupcheck)
	updateFiltering.ResultDupcheckKonsumen = string(responseskonsumen)

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("failed process update data filtering blacklist")
		return
	}

	var spouse_result response.DupcheckResult

	spouse_flag := 0

	// Pasangan
	if reqs.Data.MaritalStatus == constant.MARRIED {
		spouse_flag = 1

		var updateFiltering entity.ApiDupcheckKmbUpdate

		updateFiltering.RequestID = requestID

		param, _ := json.Marshal(map[string]string{
			"birth_date":          reqs.Data.Spouse.BirthDate,
			"id_number":           reqs.Data.Spouse.IDNumber,
			"legal_name":          reqs.Data.Spouse.LegalName,
			"surgate_mother_name": reqs.Data.Spouse.SurgateMotherName,
			"transaction_id":      reqs.Data.ProspectID,
		})

		resp, err = u.httpclient.EngineAPI(ctx, constant.FILTERING_LOG, os.Getenv("DUPCHECK_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, reqs.Data.ProspectID, accessToken)

		if err != nil || resp.StatusCode() != 200 {
			err = fmt.Errorf("failed process dupcheck spouse")
			return
		}

		if err = json.Unmarshal(resp.Body(), &dupcheck_data_spouse); err != nil {
			err = fmt.Errorf("error unmarshal dupcheck spouse response")
			return
		}
		getdupcheckspouse = dupcheck_data_spouse.Data

		responsespasangan, _ := json.Marshal(getdupcheckspouse)
		updateFiltering.ResultDupcheckPasangan = string(responsespasangan)

		if err = u.repository.UpdateData(updateFiltering); err != nil {
			err = fmt.Errorf("failed process update data filtering dupcheck spouse")
			return
		}

		if getdupcheckspouse != (response.DataDupcheck{}) {
			if getdupcheckspouse.BadType == constant.BADTYPE_B {
				spouse_result.Code = constant.CODE_SPOSE_BADTYPE_B
				spouse_result.Decision = constant.DECISION_REJECT
				spouse_result.Reason = constant.REASON_SPOSE_BADTYPE_B
			} else if getdupcheckspouse.MaxOverduedays > constant.MAX_OVER_DUE_DAYS {
				spouse_result.Code = constant.CODE_SPOSE_MAX_OVER_DUE_DAYS
				spouse_result.Decision = constant.DECISION_REJECT
				spouse_result.Reason = constant.REASON_SPOSE_MAX_OVER_DUE_DAYS
			} else if getdupcheckspouse.NumOfAssetInventoried > constant.NUM_OF_ASSET_INVENTORIED {
				spouse_result.Code = constant.CODE_SPOSE_NUM_OF_ASSET_INVENTORIED
				spouse_result.Decision = constant.DECISION_REJECT
				spouse_result.Reason = constant.REASON_SPOSE_NUM_OF_ASSET_INVENTORIED
			} else if getdupcheckspouse.IsRestructure == constant.IS_RESTRUCTURE {
				spouse_result.Code = constant.CODE_SPOSE_IS_RESTRUCTURE
				spouse_result.Decision = constant.DECISION_REJECT
				spouse_result.Reason = constant.REASON_IS_RESTRUCTURE
			} else {
				spouse_result.Code = constant.CODE_SPOSE_BERSIH
				spouse_result.Decision = constant.DECISION_PASS
				spouse_result.Reason = constant.REASON_SPOSE_BERSIH
			}
		} else {
			spouse_result.Code = constant.CODE_PASANGAN_BARU
			spouse_result.Decision = constant.DECISION_PASS
			spouse_result.Reason = constant.REASON_PASANGAN_BARU
		}

		if err != nil {
			err = fmt.Errorf("failed fetching data confins pasangan")
			return
		}

	}

	// Konsumen
	if getdupcheck != (response.DataDupcheck{}) {
		result.StatusKonsumen = "RO/AO"

		if getdupcheck.BadType == constant.BADTYPE_B {
			if spouse_flag == 1 {
				if spouse_result.Code == constant.CODE_SPOSE_BADTYPE_B {
					result.Code = constant.CODE_BADTYPE_B_SPOSE_BADTYPE_B
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BADTYPE_B + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_MAX_OVER_DUE_DAYS {
					result.Code = constant.CODE_BADTYPE_B_SPOSE_MAX_OVER_DUE_DAYS
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BADTYPE_B + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_NUM_OF_ASSET_INVENTORIED {
					result.Code = constant.CODE_BADTYPE_B_SPOSE_NUM_OF_ASSET_INVENTORIED
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BADTYPE_B + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_BERSIH {
					result.Code = constant.CODE_BADTYPE_B_SPOSE_BERSIH
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BADTYPE_B + " & " + spouse_result.Reason
				} else {
					result.Code = constant.CODE_BADTYPE_B_SPOSE
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BADTYPE_B + " & " + spouse_result.Reason
				}
			} else {
				result.Code = constant.CODE_BADTYPE_B
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_BADTYPE_B
			}
		} else if getdupcheck.MaxOverduedays > constant.MAX_OVER_DUE_DAYS {
			if spouse_flag == 1 {
				if spouse_result.Code == constant.CODE_SPOSE_BADTYPE_B {
					result.Code = constant.CODE_MAX_OVER_DUE_DAYS_SPOSE_BADTYPE_B
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_MAX_OVER_DUE_DAYS + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_MAX_OVER_DUE_DAYS {
					result.Code = constant.CODE_MAX_OVER_DUE_DAYS_SPOSE_MAX_OVER_DUE_DAYS
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_MAX_OVER_DUE_DAYS + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_NUM_OF_ASSET_INVENTORIED {
					result.Code = constant.CODE_MAX_OVER_DUE_DAYS_SPOSE_NUM_OF_ASSET_INVENTORIED
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_MAX_OVER_DUE_DAYS + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_IS_RESTRUCTURE {
					result.Code = constant.CODE_MAX_OVER_DUE_DAYS_SPOSE_IS_RESTRUCTURE
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_MAX_OVER_DUE_DAYS + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_BERSIH {
					result.Code = constant.CODE_MAX_OVER_DUE_DAYS_SPOSE_BERSIH
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_MAX_OVER_DUE_DAYS + " & " + spouse_result.Reason
				} else {
					result.Code = constant.CODE_MAX_OVER_DUE_DAYS_SPOSE
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_MAX_OVER_DUE_DAYS + " & " + spouse_result.Reason
				}
			} else {
				result.Code = constant.CODE_MAX_OVER_DUE_DAYS
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_MAX_OVER_DUE_DAYS
			}
		} else if getdupcheck.NumOfAssetInventoried > constant.NUM_OF_ASSET_INVENTORIED {
			if spouse_flag == 1 {
				if spouse_result.Code == constant.CODE_SPOSE_BADTYPE_B {
					result.Code = constant.CODE_NUM_OF_ASSET_INVENTORIED_SPOSE_BADTYPE_B
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_NUM_OF_ASSET_INVENTORIED + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_MAX_OVER_DUE_DAYS {
					result.Code = constant.CODE_NUM_OF_ASSET_INVENTORIED_SPOSE_MAX_OVER_DUE_DAYS
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_NUM_OF_ASSET_INVENTORIED + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_NUM_OF_ASSET_INVENTORIED {
					result.Code = constant.CODE_NUM_OF_ASSET_INVENTORIED_SPOSE_NUM_OF_ASSET_INVENTORIED
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_NUM_OF_ASSET_INVENTORIED + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_BADTYPE_W {
					result.Code = constant.CODE_NUM_OF_ASSET_INVENTORIED_SPOSE_BADTYPE_W
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_NUM_OF_ASSET_INVENTORIED + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_BERSIH {
					result.Code = constant.CODE_NUM_OF_ASSET_INVENTORIED_SPOSE_BERSIH
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_NUM_OF_ASSET_INVENTORIED + " & " + spouse_result.Reason
				} else {
					result.Code = constant.CODE_NUM_OF_ASSET_INVENTORIED_SPOSE
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_NUM_OF_ASSET_INVENTORIED + " & " + spouse_result.Reason
				}
			} else {
				result.Code = constant.CODE_NUM_OF_ASSET_INVENTORIED
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_NUM_OF_ASSET_INVENTORIED
			}
		} else if getdupcheck.IsRestructure == constant.IS_RESTRUCTURE {
			if spouse_flag == 1 {
				if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_BADTYPE_B {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_BADTYPE_B
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_MAX_OVER_DUE_DAYS {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_MAX_OVER_DUE_DAYS
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_NUM_OF_ASSET_INVENTORIED {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_NUM_OF_ASSET_INVENTORIED
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_BADTYPE_W {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_BADTYPE_W
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_BERSIH {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_BERSIH
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE + " & " + spouse_result.Reason
				} else {
					result.Code = constant.CODE_IS_RESTRUCTURE
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE + " & " + spouse_result.Reason
				}
			} else {
				result.Code = constant.CODE_IS_RESTRUCTURE
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE
			}
		} else {
			if spouse_flag == 1 {
				if spouse_result.Code == constant.CODE_SPOSE_BADTYPE_B {
					result.Code = constant.CODE_BERSIH_SPOSE_BADTYPE_B
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BERSIH + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_MAX_OVER_DUE_DAYS {
					result.Code = constant.CODE_BERSIH_SPOSE_MAX_OVER_DUE_DAYS
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BERSIH + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_NUM_OF_ASSET_INVENTORIED {
					result.Code = constant.CODE_BERSIH_SPOSE_NUM_OF_ASSET_INVENTORIED
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BERSIH + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_IS_RESTRUCTURE {
					result.Code = constant.CODE_SPOSE_IS_RESTRUCTURE
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_BERSIH + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_SPOSE_BERSIH {
					result.Code = constant.CODE_BERSIH_SPOSE_BERSIH
					result.Decision = constant.DECISION_PASS
					result.Reason = constant.REASON_BERSIH + " & " + spouse_result.Reason
				} else {
					result.Code = constant.CODE_BERSIH_SPOSE
					result.Decision = constant.DECISION_PASS
					result.Reason = constant.REASON_BERSIH + " & " + spouse_result.Reason
				}
			} else {
				result.Code = constant.CODE_BERSIH
				result.Decision = constant.DECISION_PASS
				result.Reason = constant.REASON_BERSIH
			}
		}

	} else {
		result.StatusKonsumen = "NEW"

		if spouse_flag == 1 {
			if spouse_result.Code == constant.CODE_SPOSE_BADTYPE_B {
				result.Code = constant.CODE_NEW_CUSTOMER_SPOSE_BADTYPE_B
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_NEW_CUSTOMER + " & " + spouse_result.Reason
			} else if spouse_result.Code == constant.CODE_SPOSE_MAX_OVER_DUE_DAYS {
				result.Code = constant.CODE_NEW_CUSTOMER_SPOSE_MAX_OVER_DUE_DAYS
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_NEW_CUSTOMER + " & " + spouse_result.Reason
			} else if spouse_result.Code == constant.CODE_SPOSE_NUM_OF_ASSET_INVENTORIED {
				result.Code = constant.CODE_NEW_CUSTOMER_SPOSE_NUM_OF_ASSET_INVENTORIED
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_NEW_CUSTOMER + " & " + spouse_result.Reason
			} else if spouse_result.Code == constant.CODE_NEW_CUSTOMER_IS_RESTRUCTURE {
				result.Code = constant.CODE_NEW_CUSTOMER_IS_RESTRUCTURE
				result.Decision = constant.DECISION_REJECT
				result.Reason = constant.REASON_NEW_CUSTOMER_IS_RESTRUCTURE + " & " + spouse_result.Reason
			} else if spouse_result.Code == constant.CODE_SPOSE_BERSIH {
				result.Code = constant.CODE_NEW_CUSTOMER_SPOSE_BERSIH
				result.Decision = constant.DECISION_PASS
				result.Reason = constant.REASON_NEW_CUSTOMER + " & " + spouse_result.Reason
			} else {
				result.Code = constant.CODE_NEW_CUSTOMER_SPOSE
				result.Decision = constant.DECISION_PASS
				result.Reason = constant.REASON_NEW_CUSTOMER + " & " + spouse_result.Reason
			}
		} else {
			result.Code = constant.CODE_NEW_CUSTOMER
			result.Decision = constant.DECISION_PASS
			result.Reason = constant.REASON_NEW_CUSTOMER
		}

	}
	result.IsBlacklist = 1
	result.NextProcess = 0
	if result.Decision == constant.DECISION_PASS {
		result.IsBlacklist = 0
		result.NextProcess = 1
	}
	return
}

func (u usecase) CheckStatusCategory(ctx context.Context, reqs request.FilteringRequest, status_konsumen, accessToken string) (data response.DupcheckResult, err error) {

	var updateFiltering entity.ApiDupcheckKmbUpdate

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	updateFiltering.RequestID = requestID

	param, _ := json.Marshal(map[string]interface{}{
		"id_number":           reqs.Data.IDNumber,
		"legal_name":          reqs.Data.LegalName,
		"birth_date":          reqs.Data.BirthDate,
		"surgate_mother_name": reqs.Data.SurgateMotherName,
		"lob_id":              constant.LOBID_KMB,
	})

	resp, err := u.httpclient.CustomerAPI(ctx, constant.FILTERING_LOG, "/api/v3/customer/personal-data", param, constant.METHOD_POST, accessToken, reqs.Data.ProspectID, "DEFAULT_TIMEOUT_30S")

	var check_customer_kategori response.CustomerDomain

	if err != nil || resp.StatusCode() != 200 && resp.StatusCode() != 400 {
		err = fmt.Errorf("failed get status category consument")
		return
	}

	if err = json.Unmarshal(resp.Body(), &check_customer_kategori); err != nil {
		err = fmt.Errorf("error unmarshal status category consument")
		return
	}

	customer_check_code, _ := utils.ItemExists(check_customer_kategori.Code, []string{constant.CORE_API_005, constant.CORE_API_019})

	if customer_check_code || resp.StatusCode() == 400 {
		data.Code = constant.CUSTOMER_NEW
		data.StatusKonsumen = status_konsumen
		if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
			data.KategoriStatusKonsumen = constant.RO_AO_REGULAR
		}
	} else {
		// Get the First Segmentation
		segmentation := check_customer_kategori.Data.CustomerSegmentation
		var segmentName string
		if len(segmentation) > 0 {
			segmentName = segmentation[0].SegmentName
		}

		// dummy category consument
		active, _ := strconv.ParseBool(os.Getenv("DUMMY_CUSTOMER_CATEGORY"))
		if active && segmentName == "" {
			segmentName = os.Getenv("CUSTOMER_CATEGORY")
		}

		customer_prime_priority, _ := utils.ItemExists(segmentName, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

		if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
			if customer_prime_priority { //PRIME/PRIORITY
				data.Code = constant.CUSTOMER_STATUS_CODE_RO_AO_PRIME_PRIORITY
				data.Decision = constant.DECISION_PASS
				data.Reason = constant.REASON_CUSTOMER_STATUS_CODE_RO_AO_PRIME_PRIORITY
				data.StatusKonsumen = status_konsumen
				data.NextProcess = 1
			} else { //REGULER
				data.Code = constant.CUSTOMER_STATUS_CODE_RO_AO
				data.Decision = constant.DECISION_PASS
				data.Reason = constant.REASON_CUSTOMER_STATUS_CODE_RO_AO
				data.StatusKonsumen = status_konsumen
			}
			data.KategoriStatusKonsumen = segmentName
		} else if status_konsumen == constant.STATUS_KONSUMEN_NEW {
			data.Code = constant.CUSTOMER_STATUS_CODE_NEW
			data.Decision = constant.DECISION_PASS
			data.Reason = constant.REASON_CUSTOMER_STATUS_CODE_NEW
			data.StatusKonsumen = status_konsumen
		}
	}

	updateFiltering.CustomerType = string(data.KategoriStatusKonsumen)

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("failed update data on check status category")
		return
	}

	return
}

func (u usecase) FilteringPefindo(ctx context.Context, reqs request.FilteringRequest, status_konsumen, kategoriStatusKonsumen, accessToken string) (data response.DupcheckResult, isRejectClusterEF bool, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	var (
		bpkbName                 string
		updateFiltering          entity.ApiDupcheckKmbUpdate
		resultPefindoExcludeBNPL string
		resultPefindoIncludeAll  string
	)

	updateFiltering.RequestID = requestID

	namaSama := utils.AizuArrayString(os.Getenv("NAMA_SAMA"))
	namaBeda := utils.AizuArrayString(os.Getenv("NAMA_BEDA"))

	bpkbNamaSama, _ := utils.ItemExists(reqs.Data.BPKBName, namaSama)
	bpkbNamaBeda, _ := utils.ItemExists(reqs.Data.BPKBName, namaBeda)

	if bpkbNamaSama {
		bpkbName = constant.NAMA_SAMA
	} else if bpkbNamaBeda {
		bpkbName = constant.NAMA_BEDA
	}

	active, _ := strconv.ParseBool(os.Getenv("ACTIVE_PBK"))

	if active {
		var (
			checkPefindo    response.ResponsePefindo
			pefindoResult   response.PefindoResult
			responsePefindo interface{}
		)

		var resp *resty.Response

		param, _ := json.Marshal(map[string]string{
			"ClientKey":         os.Getenv("CLIENTKEY_CORE_PBK"),
			"IDMember":          constant.USER_PBK_KMB_FILTEERING,
			"user":              constant.USER_PBK_KMB_FILTEERING,
			"IDNumber":          reqs.Data.IDNumber,
			"ProspectID":        reqs.Data.ProspectID,
			"LegalName":         reqs.Data.LegalName,
			"BirthDate":         reqs.Data.BirthDate,
			"SurgateMotherName": reqs.Data.SurgateMotherName,
			"Gender":            reqs.Data.Gender,
			"MaritalStatus":     reqs.Data.MaritalStatus,
		})

		if reqs.Data.MaritalStatus == constant.MARRIED {

			param, _ = json.Marshal(map[string]string{
				"ClientKey":                os.Getenv("CLIENTKEY_CORE_PBK"),
				"IDMember":                 constant.USER_PBK_KMB_FILTEERING,
				"user":                     constant.USER_PBK_KMB_FILTEERING,
				"IDNumber":                 reqs.Data.IDNumber,
				"ProspectID":               reqs.Data.ProspectID,
				"LegalName":                reqs.Data.LegalName,
				"BirthDate":                reqs.Data.BirthDate,
				"SurgateMotherName":        reqs.Data.SurgateMotherName,
				"Gender":                   reqs.Data.Gender,
				"MaritalStatus":            reqs.Data.MaritalStatus,
				"Spouse_IDNumber":          reqs.Data.Spouse.IDNumber,
				"Spouse_LegalName":         reqs.Data.Spouse.LegalName,
				"Spouse_BirthDate":         reqs.Data.Spouse.BirthDate,
				"Spouse_SurgateMotherName": reqs.Data.Spouse.SurgateMotherName,
				"Spouse_Gender":            reqs.Data.Spouse.Gender,
			})
		}

		resp, err = u.httpclient.EngineAPI(ctx, constant.FILTERING_LOG, os.Getenv("PBK_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, reqs.Data.ProspectID, accessToken)

		if err != nil || resp.StatusCode() != 200 && resp.StatusCode() != 400 {
			err = fmt.Errorf("failed get data pefindo")
			return
		}

		if err = json.Unmarshal(resp.Body(), &checkPefindo); err != nil {
			err = fmt.Errorf("error unmarshal data pefindo")
			return
		}

		responsePefindo = string(resp.Body())

		updateFiltering.ResultPefindo = responsePefindo

		// handling response pefindo
		if checkPefindo.Code == "200" || checkPefindo.Code == "201" {

			if reflect.TypeOf(checkPefindo.Result).String() != "string" {
				if checkPefindo.Result != constant.NOT_MATCH_PBK {
					setPefindo, _ := json.Marshal(checkPefindo.Result)

					if errs := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(setPefindo, &pefindoResult); errs != nil {
						err = fmt.Errorf("error unmarshal data pefindo")
						return
					}
				}
			}
		}

		if checkPefindo.Code == "200" && pefindoResult.Score != constant.PEFINDO_UNSCORE {
			// KO Rules Exclude BNPL
			if pefindoResult.Category != nil {
				if bpkbName == constant.NAMA_BEDA {
					if pefindoResult.MaxOverdueLast12MonthsKORules != nil {
						if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12MonthsKORules) <= constant.PBK_OVD_LAST_12 {
							if pefindoResult.MaxOverdueKORules == nil {
								data.Code = constant.NAMA_BEDA_CURRENT_OVD_NULL_CODE
								data.StatusKonsumen = status_konsumen
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) <= constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_BEDA_CURRENT_OVD_UNDER_LIMIT_CODE
								data.StatusKonsumen = status_konsumen
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) > constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_BEDA_CURRENT_OVD_OVER_LIMIT_CODE
								data.StatusKonsumen = status_konsumen
								data.Decision = constant.DECISION_REJECT
								data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
							}
						} else {
							data.Code = constant.NAMA_BEDA_12_OVD_OVER_LIMIT_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_REJECT
							data.Reason = fmt.Sprintf("NAMA BEDA %s & OVD 12 Bulan Terakhir > %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
						}
					} else {
						data.Code = constant.NAMA_BEDA_12_OVD_NULL_CODE
						data.StatusKonsumen = status_konsumen
						data.Decision = constant.DECISION_PASS
						data.Reason = fmt.Sprintf("NAMA BEDA %s & OVD 12 Bulan Terakhir Null", getReasonCategoryRoman(pefindoResult.Category))
					}
				} else if bpkbName == constant.NAMA_SAMA {
					if pefindoResult.MaxOverdueLast12MonthsKORules != nil {
						if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12MonthsKORules) <= constant.PBK_OVD_LAST_12 {
							if pefindoResult.MaxOverdueKORules == nil {
								data.Code = constant.NAMA_SAMA_CURRENT_OVD_NULL_CODE
								data.StatusKonsumen = status_konsumen
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) <= constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_SAMA_CURRENT_OVD_UNDER_LIMIT_CODE
								data.StatusKonsumen = status_konsumen
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) > constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_SAMA_CURRENT_OVD_OVER_LIMIT_CODE
								data.StatusKonsumen = status_konsumen
								data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)

								data.Decision = func() string {
									if checkNullCategory(pefindoResult.Category) == 3 {
										return constant.DECISION_REJECT
									}
									return constant.DECISION_PASS
								}()
							}
						} else {
							data.Code = constant.NAMA_SAMA_12_OVD_OVER_LIMIT_CODE
							data.StatusKonsumen = status_konsumen
							data.Reason = fmt.Sprintf("NAMA SAMA %s & OVD 12 Bulan Terakhir > %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)

							data.Decision = func() string {
								if checkNullCategory(pefindoResult.Category) == 3 {
									return constant.DECISION_REJECT
								}
								return constant.DECISION_PASS
							}()
						}
					} else {
						data.Code = constant.NAMA_SAMA_12_OVD_NULL_CODE
						data.StatusKonsumen = status_konsumen
						data.Decision = constant.DECISION_PASS
						data.Reason = fmt.Sprintf("NAMA SAMA %s & OVD 12 Bulan Terakhir Null", getReasonCategoryRoman(pefindoResult.Category))
					}
				}

				resultPefindoExcludeBNPL = data.Decision
				updateFiltering.ResultPefindoExcludeBNPL = resultPefindoExcludeBNPL
				if pefindoResult.Category != nil {
					updateFiltering.CategoryExcludeBNPL = checkNullCategory(pefindoResult.Category)
				}
				if pefindoResult.MaxOverdueKORules != nil {
					updateFiltering.OverdueCurrentExcludeBNPL = checkNullMaxOverdue(pefindoResult.MaxOverdueKORules)
				}
				if pefindoResult.MaxOverdueLast12MonthsKORules != nil {
					updateFiltering.OverdueLast12MonthsExcludeBNPL = checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12MonthsKORules)
				}

				resultPefindo := data.Decision

				if resultPefindo == constant.DECISION_PASS {
					data.NextProcess = 1
				}

				check_prime_priority, _ := utils.ItemExists(kategoriStatusKonsumen, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

				if !check_prime_priority {
					// BPKB Nama Sama
					if bpkbName == constant.NAMA_SAMA {
						if pefindoResult.WoContract { //Wo Contract Yes

							if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

								if status_konsumen == constant.STATUS_KONSUMEN_NEW {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										if resultPefindo == constant.DECISION_PASS {
											data.NextProcess = 0
											data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
										} else {
											data.NextProcess = 1
											data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
										}
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else {
									data.NextProcess = 0
									data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
								}

							} else { //Wo Agunan No
								if status_konsumen == constant.STATUS_KONSUMEN_NEW {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										data.NextProcess = 1
										data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										data.NextProcess = 1
										data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								}
							}
						} else { //Wo Contract No
							if status_konsumen == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = 1
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								} else {
									data.NextProcess = 0
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}

							} else if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = 1
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								} else {
									data.NextProcess = 0
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else {
								data.NextProcess = 1
								data.Code = constant.WO_AGUNAN_PASS_CODE
								data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
							}
						}
					}

					// BPKB Nama Beda
					if bpkbName == constant.NAMA_BEDA {
						if pefindoResult.WoContract { //Wo Contract Yes

							if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

								if status_konsumen == constant.STATUS_KONSUMEN_NEW {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										if resultPefindo == constant.DECISION_PASS {
											data.NextProcess = 0
											data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
										} else {
											data.NextProcess = 1
											data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
										}
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else {
									data.NextProcess = 0
									data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
								}

							} else { //Wo Agunan No
								if status_konsumen == constant.STATUS_KONSUMEN_NEW {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										if resultPefindo == constant.DECISION_PASS {
											data.NextProcess = 1
											data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
										} else {
											data.NextProcess = 0
											data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
										}
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
									if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
										data.NextProcess = 1
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									}
								}
							}
						} else { //Wo Contract No
							if status_konsumen == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									if resultPefindo == constant.DECISION_PASS {
										data.NextProcess = 1
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = 0
										data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else {
									data.NextProcess = 0
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}

							} else if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = 1
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								} else {
									data.NextProcess = 0
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else {
								data.NextProcess = 1
								data.Code = constant.WO_AGUNAN_PASS_CODE
								data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
							}
						}
					}

					data.StatusKonsumen = status_konsumen
					if data.Decision == constant.DECISION_REJECT {
						data.Code = constant.WO_AGUNAN_REJECT_CODE
					} else {
						data.Code = constant.WO_AGUNAN_PASS_CODE
					}
				}

				data.PbkReport = pefindoResult.DetailReport
				data.TotalBakiDebet = pefindoResult.TotalBakiDebetNonAgunan

				updateFiltering.PefindoScore = &pefindoResult.Score
			} else {
				data.Code = constant.PBK_NO_HIT
				data.StatusKonsumen = status_konsumen
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK No Hit - Kategori Konsumen Null"
				data.NextProcess = 1

				updateFiltering.PefindoScore = new(string)
				*updateFiltering.PefindoScore = constant.UNSCORE_PBK

				updateFiltering.ResultPefindoExcludeBNPL = constant.DECISION_PASS
			}

			updateFiltering.PefindoID = &checkPefindo.Konsumen.PefindoID
			if checkPefindo.Pasangan != (response.PefindoResultPasangan{}) {
				updateFiltering.PefindoIDSpouse = &checkPefindo.Pasangan.PefindoID
			}

			// KO Rules Include All
			if bpkbName == constant.NAMA_BEDA {
				if pefindoResult.MaxOverdueLast12Months != nil {
					if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12Months) <= constant.PBK_OVD_LAST_12 {
						if pefindoResult.MaxOverdue == nil {
							resultPefindoIncludeAll = constant.DECISION_PASS
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) <= constant.PBK_OVD_CURRENT {
							resultPefindoIncludeAll = constant.DECISION_PASS
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) > constant.PBK_OVD_CURRENT {
							resultPefindoIncludeAll = constant.DECISION_REJECT
						}
					} else {
						resultPefindoIncludeAll = constant.DECISION_REJECT
					}
				} else {
					resultPefindoIncludeAll = constant.DECISION_PASS
				}
			} else if bpkbName == constant.NAMA_SAMA {
				if pefindoResult.MaxOverdueLast12Months != nil {
					if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12Months) <= constant.PBK_OVD_LAST_12 {
						if pefindoResult.MaxOverdue == nil {
							resultPefindoIncludeAll = constant.DECISION_PASS
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) <= constant.PBK_OVD_CURRENT {
							resultPefindoIncludeAll = constant.DECISION_PASS
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) > constant.PBK_OVD_CURRENT {
							resultPefindoIncludeAll = constant.DECISION_REJECT
						}
					} else {
						resultPefindoIncludeAll = constant.DECISION_REJECT
					}
				} else {
					resultPefindoIncludeAll = constant.DECISION_PASS
				}
			}

			updateFiltering.ResultPefindoIncludeAll = resultPefindoIncludeAll
			if pefindoResult.MaxOverdue != nil {
				updateFiltering.OverdueCurrentIncludeAll = checkNullMaxOverdue(pefindoResult.MaxOverdue)
			}
			if pefindoResult.MaxOverdueLast12Months != nil {
				updateFiltering.OverdueLast12MonthsIncludeAll = checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12Months)
			}

			// Check Reject Cluster E & F
			namaSama := utils.AizuArrayString(os.Getenv("NAMA_SAMA"))
			namaBeda := utils.AizuArrayString(os.Getenv("NAMA_BEDA"))

			bpkbNamaSama, _ := utils.ItemExists(reqs.Data.BPKBName, namaSama)
			bpkbNamaBeda, _ := utils.ItemExists(reqs.Data.BPKBName, namaBeda)

			var (
				bpkbNameType  int
				clusterBranch entity.ClusterBranch
			)

			if bpkbNamaSama {
				bpkbNameType = 1
			} else if bpkbNamaBeda {
				bpkbNameType = 0
			}

			// Get Cluster Branch
			clusterBranch, err = u.elaborateRepository.GetClusterBranchElaborate(reqs.Data.BranchID, status_konsumen, bpkbNameType)
			if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
				err = fmt.Errorf("failed get cluster branch")
				return
			}

			if clusterBranch != (entity.ClusterBranch{}) {

				if pefindoResult.TotalBakiDebetNonAgunan > constant.RANGE_CLUSTER_BAKI_DEBET_REJECT && pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
					if clusterBranch.Cluster == constant.CLUSTER_E || clusterBranch.Cluster == constant.CLUSTER_F {
						if resultPefindoIncludeAll == constant.DECISION_REJECT {
							isRejectClusterEF = true

							data.Code = constant.CODE_REJECT_CLUSTER_E_F
							data.Decision = constant.DECISION_REJECT
							data.NextProcess = 0

							bpkbNamePrefix := "NAMA SAMA"
							if bpkbName == constant.NAMA_BEDA {
								bpkbNamePrefix = "NAMA BEDA"
							}
							data.Reason = fmt.Sprintf("%s "+constant.REASON_REJECT_CLUSTER_E_F, bpkbNamePrefix, getReasonCategoryRoman(pefindoResult.Category))
						}
					}
				}
			}
		} else if checkPefindo.Code == "201" || pefindoResult.Score == constant.PEFINDO_UNSCORE {

			if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
				data.Code = constant.NAMA_SAMA_UNSCORE_RO_AO_CODE
				data.StatusKonsumen = status_konsumen
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + status_konsumen
				data.NextProcess = 1
			} else if status_konsumen == constant.STATUS_KONSUMEN_NEW {
				data.Code = constant.NAMA_SAMA_UNSCORE_NEW_CODE
				data.StatusKonsumen = status_konsumen
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + status_konsumen
				data.NextProcess = 1
			}

			updateFiltering.PefindoID = &checkPefindo.Konsumen.PefindoID
			updateFiltering.PefindoScore = &pefindoResult.Score

			unscore := constant.PEFINDO_UNSCORE

			if pefindoResult.Score == "" {
				updateFiltering.PefindoScore = &unscore
			}
			if checkPefindo.Pasangan != (response.PefindoResultPasangan{}) {
				updateFiltering.PefindoIDSpouse = &checkPefindo.Pasangan.PefindoID
			}

		} else if checkPefindo.Code == "202" {
			// data.Code = constant.SERVICE_PBK_UNAVAILABLE_CODE
			// data.StatusKonsumen = status_konsumen
			// data.Reason = "Service PBK tidak tersedia"

			if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
				data.Code = constant.NAMA_SAMA_UNSCORE_RO_AO_CODE
			} else if status_konsumen == constant.STATUS_KONSUMEN_NEW {
				data.Code = constant.NAMA_SAMA_UNSCORE_NEW_CODE
			}

			data.Reason = constant.REASON_FILTERING_PBK_DOWN
			data.StatusKonsumen = status_konsumen
			data.Decision = constant.DECISION_PASS
			data.NextProcess = 1

			unscore := constant.PEFINDO_UNSCORE
			updateFiltering.PefindoScore = &unscore
		}

	} else {
		data.Code = constant.PBK_NO_HIT
		data.StatusKonsumen = status_konsumen
		data.Reason = "Akses ke PBK ditutup"

	}

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("failed update data filtering")
		return
	}

	return
}

func checkNullMaxOverdueLast12Months(MaxOverdueLast12Months interface{}) float64 {
	var max_overdue_last12_months float64

	if utils.CheckVriable(MaxOverdueLast12Months) == reflect.String.String() {
		max_overdue_last12_months = utils.StrConvFloat64(MaxOverdueLast12Months.(string))
	} else {
		max_overdue_last12_months = MaxOverdueLast12Months.(float64)
	}

	return max_overdue_last12_months
}

func checkNullMaxOverdue(MaxOverdueLast interface{}) float64 {
	var max_overdue_months float64

	if utils.CheckVriable(MaxOverdueLast) == reflect.String.String() {
		max_overdue_months = utils.StrConvFloat64(MaxOverdueLast.(string))
	} else {
		max_overdue_months = MaxOverdueLast.(float64)
	}

	return max_overdue_months
}

func checkNullCategory(Category interface{}) float64 {
	var category float64

	if utils.CheckVriable(Category) == reflect.String.String() {
		category = utils.StrConvFloat64(Category.(string))
	} else {
		category = Category.(float64)
	}

	return category
}

// function to map reason category values to Roman numerals
func getReasonCategoryRoman(category interface{}) string {
	switch category.(float64) {
	case 1:
		return "(I)"
	case 2:
		return "(II)"
	case 3:
		return "(III)"
	default:
		return ""
	}
}
