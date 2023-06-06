package usecase

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"los-kmb-api/domain/filtering/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"

	"github.com/google/uuid"
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

func (u multiUsecase) Filtering(reqs request.BodyRequest, accessToken string) (data interface{}, err error) {

	var savedata entity.ApiDupcheckKmb

	id := uuid.New()

	savedata.RequestID = id.String()

	request, _ := json.Marshal(reqs)
	savedata.Request = string(request)
	savedata.ProspectID = reqs.Data.ProspectID
	if err = u.repository.SaveData(savedata); err != nil {
		err = fmt.Errorf("failed process dupcheck")
		return
	}

	check_blacklist, err := u.usecase.FilteringBlackList(reqs, savedata.RequestID)
	if err != nil {
		err = fmt.Errorf("failed process check blacklist")
		return
	}

	ints := utils.AizuArrayInt(constant.CODE_BLACKLIST, "17")

	check_blacklist_code, _ := utils.ItemExists(check_blacklist.Code, ints)

	var updateFiltering entity.ApiDupcheckKmbUpdate

	// check status & kategori konsumen
	if check_blacklist_code {
		check_konsumen, errs := u.usecase.CheckStatusCategory(reqs, check_blacklist.StatusKonsumen, savedata.RequestID, accessToken)
		if errs != nil {
			err = fmt.Errorf("failed fetching data customer domain")
			return
		}

		check_prime_priority, _ := utils.ItemExists(check_konsumen.KategoriStatusKonsumen, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

		if !check_prime_priority { //bukan prime/priority
			// hit ke pefindo
			check_pefindo, errs := u.usecase.FilteringPefindo(reqs, check_konsumen.StatusKonsumen, savedata.RequestID)

			if errs != nil {
				err = fmt.Errorf("failed process pefindo")
				return
			}

			check_pefindo.StatusKonsumen = check_blacklist.StatusKonsumen
			check_pefindo.KategoriStatusKonsumen = check_konsumen.KategoriStatusKonsumen

			check_pefindo.NextProcess = 1
			if check_pefindo.Decision == constant.DECISION_REJECT {
				check_pefindo.NextProcess = 0
			}

			updateFiltering.Code = utils.IntConvStr(check_pefindo.Code)
			updateFiltering.Reason = check_pefindo.Reason
			updateFiltering.Decision = check_pefindo.Decision
			data = check_pefindo

		} else {
			updateFiltering.Code = utils.IntConvStr(check_konsumen.Code)
			updateFiltering.Reason = check_konsumen.Reason
			updateFiltering.Decision = check_konsumen.Decision
			data = check_konsumen
		}

	} else {
		updateFiltering.Code = utils.IntConvStr(check_blacklist.Code)
		updateFiltering.Reason = check_blacklist.Reason
		updateFiltering.Decision = check_blacklist.Decision
		data = check_blacklist
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

func (u usecase) FilteringBlackList(reqs request.BodyRequest, request_id string) (result response.DupcheckResult, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	var updateFiltering entity.ApiDupcheckKmbUpdate

	updateFiltering.RequestID = request_id

	param_konsumen := map[string]string{
		"birth_date":          reqs.Data.BirthDate,
		"id_number":           reqs.Data.IDNumber,
		"legal_name":          reqs.Data.LegalName,
		"surgate_mother_name": reqs.Data.SurgateMotherName,
		"transaction_id":      reqs.Data.ProspectID,
	}

	var dupcheck_data response.DupCheckData
	var getdupcheck response.DataDupcheck
	var dupcheck_data_spouse response.DupCheckData
	var getdupcheckspouse response.DataDupcheck

	resp, errs := u.httpclient.CallWebSocket(os.Getenv("DUPCHECK_URL"), param_konsumen, map[string]string{}, timeOut)

	if errs != nil || resp.StatusCode() != 200 {
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

		updateFiltering.RequestID = request_id

		param_pasangan := map[string]string{
			"birth_date":          reqs.Data.Spouse.BirthDate,
			"id_number":           reqs.Data.Spouse.IDNumber,
			"legal_name":          reqs.Data.Spouse.LegalName,
			"surgate_mother_name": reqs.Data.Spouse.SurgateMotherName,
			"transaction_id":      reqs.Data.ProspectID,
		}

		resp, errs := u.httpclient.CallWebSocket(os.Getenv("DUPCHECK_URL"), param_pasangan, map[string]string{}, timeOut)

		if errs != nil || resp.StatusCode() != 200 {
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
				} else if spouse_result.Code == constant.CODE_SPOSE_BADTYPE_W {
					result.Code = constant.CODE_MAX_OVER_DUE_DAYS_SPOSE_BADTYPE_W
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
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE_BADTYPE_B + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_MAX_OVER_DUE_DAYS {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_MAX_OVER_DUE_DAYS
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE_MAX_OVER_DUE_DAYS + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_NUM_OF_ASSET_INVENTORIED {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_NUM_OF_ASSET_INVENTORIED
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE_NUM_OF_ASSET_INVENTORIED + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_BADTYPE_W {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_BADTYPE_W
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE_BADTYPE_W + " & " + spouse_result.Reason
				} else if spouse_result.Code == constant.CODE_IS_RESTRUCTURE_SPOSE_BERSIH {
					result.Code = constant.CODE_IS_RESTRUCTURE_SPOSE_BERSIH
					result.Decision = constant.DECISION_REJECT
					result.Reason = constant.REASON_SPOSE_IS_RESTRUCTURE_BERSIH + " & " + spouse_result.Reason
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
					result.Reason = constant.REASON_IS_RESTRUCTURE + " & " + spouse_result.Reason
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

func (u usecase) CheckStatusCategory(reqs request.BodyRequest, status_konsumen, request_id string, accessToken string) (data response.DupcheckResult, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	var updateFiltering entity.ApiDupcheckKmbUpdate

	updateFiltering.RequestID = request_id

	param, _ := json.Marshal(map[string]interface{}{
		"id_number":           reqs.Data.IDNumber,
		"legal_name":          reqs.Data.LegalName,
		"birth_date":          reqs.Data.BirthDate,
		"surgate_mother_name": reqs.Data.SurgateMotherName,
		"lob_id":              constant.LOBID_KMB,
	})

	resp, err := u.httpclient.CustomerDomain("/api/v3/customer/personal-data", param, map[string]string{}, constant.METHOD_POST, timeOut, accessToken)

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

				ResposePefindo, errs := u.HitPefindoPrimePriority(reqs, status_konsumen, request_id)

				if errs != nil {
					err = fmt.Errorf("failed process hit pefindo prime priority")
					return
				}

				data.PbkReport = ResposePefindo.PbkReport

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
		} else if status_konsumen == constant.STATUS_KONSUMEN_NEW {
			data.Code = constant.CUSTOMER_STATUS_CODE_NEW
			data.Decision = constant.DECISION_PASS
			data.Reason = constant.REASON_CUSTOMER_STATUS_CODE_NEW
			data.StatusKonsumen = status_konsumen
		}
		data.KategoriStatusKonsumen = segmentName
	}

	updateFiltering.CustomerType = string(data.KategoriStatusKonsumen)

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("failed update data on check status category")
		return
	}

	return
}

func (u usecase) FilteringKreditmu(reqs request.BodyRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	var updateFiltering entity.ApiDupcheckKmbUpdate

	updateFiltering.RequestID = request_id

	paramKreditmu := map[string]string{
		"birth_date":          reqs.Data.BirthDate,
		"id_number":           reqs.Data.IDNumber,
		"legal_name":          reqs.Data.LegalName,
		"surgate_mother_name": reqs.Data.SurgateMotherName,
	}

	resp, err := u.httpclient.CallWebSocket(os.Getenv("KREDITMU_URL"), paramKreditmu, map[string]string{}, timeOut)

	var check_kreditmu_konsumen response.KreditMuResponse

	if err != nil || resp.StatusCode() != 200 && resp.StatusCode() != 400 {
		err = fmt.Errorf("FAILED FETCHING DATA KREDITMU")
		return
	}

	if err = json.Unmarshal(resp.Body(), &check_kreditmu_konsumen); err != nil {
		err = fmt.Errorf("KMB FILTERING SERVICE UNAVAILABLE")
		return
	}

	kreditmu_check_code, _ := utils.ItemExists(check_kreditmu_konsumen.Code, []string{constant.CORE_API_005, constant.CORE_API_019})

	limit_inactive := strings.Split(check_kreditmu_konsumen.Data.LimitStatus, "_")

	if kreditmu_check_code || resp.StatusCode() == 400 {
		data.Code = constant.KREDITMU_NEW
		data.StatusKonsumen = status_konsumen
	} else {
		if check_kreditmu_konsumen.Data.CustomerStatus == constant.KREDITMU_VERIFY {
			if check_kreditmu_konsumen.Data.LimitStatus == constant.KREDITMU_ACTIVE || limit_inactive[0] == constant.KREDITMU_INACTIVE {
				data.StatusKonsumen = "REGISTERED " + status_konsumen

				if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
					data.Code = constant.KREDITMU_STATUS_CODE_RO_AO
					data.Decision = constant.DECISION_PASS
					data.Reason = constant.REASON_KREDITMU_STATUS_CODE_RO_AO
					data.StatusKonsumen = status_konsumen
				} else if status_konsumen == constant.STATUS_KONSUMEN_NEW {
					data.Code = constant.KREDITMU_STATUS_CODE_NEW
					data.Decision = constant.DECISION_PASS
					data.Reason = constant.REASON_KREDITMU_STATUS_CODE_NEW
					data.StatusKonsumen = status_konsumen
				}

			} else {
				data.Code = constant.KREDITMU_NEW
				data.StatusKonsumen = status_konsumen
			}

		} else {
			data.Code = constant.KREDITMU_NEW
			data.StatusKonsumen = status_konsumen
		}
	}

	updateFiltering.ResultKreditmu = string(resp.Body())

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("FAILED FETCHING DATA KREDITMU")
		return
	}

	return
}

func (u usecase) FilteringPefindo(reqs request.BodyRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	var bpkbName string

	var updateFiltering entity.ApiDupcheckKmbUpdate

	updateFiltering.RequestID = request_id

	namaSama := utils.AizuArrayString(os.Getenv("NAMA_SAMA"))
	namaBeda := utils.AizuArrayString(os.Getenv("NAMA_BEDA"))

	bpkb_nama_sama, _ := utils.ItemExists(reqs.Data.BPKBName, namaSama)
	bpkb_nama_beda, _ := utils.ItemExists(reqs.Data.BPKBName, namaBeda)

	if bpkb_nama_sama {
		bpkbName = constant.NAMA_SAMA
	} else if bpkb_nama_beda {
		bpkbName = constant.NAMA_BEDA
	}

	paramPefindo := map[string]string{
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
	}

	if reqs.Data.MaritalStatus == constant.MARRIED {

		paramPefindo = map[string]string{
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
		}
	}

	active, _ := strconv.ParseBool(os.Getenv("ACTIVE_PBK"))
	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_DUPCHECK"))

	if active {
		var check_pefindo response.ResposePefindo

		if dummy {
			getdata, errs := u.GetDummyPBK(reqs.Data.IDNumber)

			if getdata == (entity.DummyPBK{}) {
				check_pefindo.Code = "201"
				check_pefindo.Result = "Pefindo Dummy Data Not Found"

				resp := map[string]string{
					"code":   check_pefindo.Code,
					"result": "Pefindo Dummy Data Not Found",
				}
				ResponsePefindo, _ := json.Marshal(resp)
				updateFiltering.ResultPefindo = ResponsePefindo

			} else {
				if errs != nil {
					err = fmt.Errorf("FAILED FETCHING DATA PEFINDO")
					return
				}

				if err = json.Unmarshal([]byte(getdata.Response), &check_pefindo); err != nil {
					err = fmt.Errorf("KMB FILTERING SERVICE UNAVAILABLE")
					return
				}
				ResponsePefindo, _ := json.Marshal(check_pefindo)
				updateFiltering.ResultPefindo = ResponsePefindo
			}

		} else {

			resp, errs := u.httpclient.CallWebSocket(os.Getenv("PBK_URL"), paramPefindo, map[string]string{}, timeOut)

			if errs != nil || resp.StatusCode() != 200 && resp.StatusCode() != 400 {
				err = fmt.Errorf("FAILED FETCHING DATA PEFINDO")
				return
			}

			if err = json.Unmarshal(resp.Body(), &check_pefindo); err != nil {
				err = fmt.Errorf("KMB FILTERING SERVICE UNAVAILABLE")
				return
			}

			updateFiltering.ResultPefindo = string(resp.Body())

		}

		if check_pefindo.Code == "200" && check_pefindo.Result != "UNSCORE" {

			c, _ := json.Marshal(check_pefindo.Result)
			var pefindo_result response.PefindoResult

			if errs := json.Unmarshal(c, &pefindo_result); errs != nil {
				err = fmt.Errorf("KMB FILTERING SERVICE UNAVAILABLE")
				return
			}

			if bpkbName == constant.NAMA_BEDA {
				if pefindo_result.MaxOverdueLast12Months != nil {
					if checkNullMaxOverdueLast12Months(pefindo_result.MaxOverdueLast12Months) <= constant.PBK_OVD_LAST_12 {
						if pefindo_result.MaxOverdue == nil {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_NULL_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA BEDA & PBK OVD 12 Bulan Terakhir <= %d", constant.PBK_OVD_LAST_12)
						} else if checkNullMaxOverdue(pefindo_result.MaxOverdue) <= constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_UNDER_LIMIT_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA BEDA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						} else if checkNullMaxOverdue(pefindo_result.MaxOverdue) > constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_OVER_LIMIT_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_REJECT
							data.Reason = fmt.Sprintf("NAMA BEDA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						}
					} else {
						data.Code = constant.NAMA_BEDA_12_OVD_OVER_LIMIT_CODE
						data.StatusKonsumen = status_konsumen
						data.Decision = constant.DECISION_REJECT
						data.Reason = fmt.Sprintf("NAMA BEDA & OVD 12 Bulan Terakhir > %d", constant.PBK_OVD_LAST_12)
					}
				} else {
					data.Code = constant.NAMA_BEDA_12_OVD_NULL_CODE
					data.StatusKonsumen = status_konsumen
					data.Decision = constant.DECISION_PASS
					data.Reason = "NAMA BEDA & OVD 12 Bulan Terakhir Null"
				}
			} else if bpkbName == constant.NAMA_SAMA {
				if pefindo_result.MaxOverdueLast12Months != nil {
					if checkNullMaxOverdueLast12Months(pefindo_result.MaxOverdueLast12Months) <= constant.PBK_OVD_LAST_12 {
						if pefindo_result.MaxOverdue == nil {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_NULL_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA SAMA & PBK OVD 12 Bulan Terakhir <= %d", constant.PBK_OVD_LAST_12)
						} else if checkNullMaxOverdue(pefindo_result.MaxOverdue) <= constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_UNDER_LIMIT_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA SAMA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						} else if checkNullMaxOverdue(pefindo_result.MaxOverdue) > constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_OVER_LIMIT_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_REJECT
							data.Reason = fmt.Sprintf("NAMA SAMA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						}
					} else {
						data.Code = constant.NAMA_SAMA_12_OVD_OVER_LIMIT_CODE
						data.StatusKonsumen = status_konsumen
						data.Decision = constant.DECISION_REJECT
						data.Reason = fmt.Sprintf("NAMA SAMA & OVD 12 Bulan Terakhir > %d", constant.PBK_OVD_LAST_12)
					}
				} else {
					data.Code = constant.NAMA_SAMA_12_OVD_NULL_CODE
					data.StatusKonsumen = status_konsumen
					data.Decision = constant.DECISION_PASS
					data.Reason = "NAMA SAMA & OVD 12 Bulan Terakhir Null"
				}

			}

			if data.Decision == constant.DECISION_REJECT {
				var nama string
				if bpkbName == constant.NAMA_SAMA {
					nama = "SAMA"
				} else {
					nama = "BEDA"
				}
				if pefindo_result.WoContract {

					if !pefindo_result.WoAdaAgunan { //wo_agunan No
						if pefindo_result.TotalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data.Code = constant.WO_AGUNAN_REJECT_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_REJECT
							data.Reason = "NAMA " + nama + " & Baki Debet > 20 Juta"
						} else {
							data.Code = constant.WO_AGUNAN_PASS_CODE
							data.StatusKonsumen = status_konsumen
							data.Decision = constant.DECISION_PASS
							data.Reason = "NAMA " + nama + " & Baki Debet Sesuai Ketentuan"
						}

					} else { //wo_agunan Yes
						data.Code = constant.WO_AGUNAN_REJECT_CODE
						data.StatusKonsumen = status_konsumen
						data.Decision = constant.DECISION_REJECT
						data.Reason = "NAMA " + nama + " & Ada Fasilitas WO Agunan"
					}
				} else { //wo_contract No
					if pefindo_result.TotalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data.Code = constant.WO_AGUNAN_REJECT_CODE
						data.StatusKonsumen = status_konsumen
						data.Decision = constant.DECISION_REJECT
						data.Reason = "NAMA " + nama + " & Baki Debet > 20 Juta"
					} else {
						data.Code = constant.WO_AGUNAN_PASS_CODE
						data.StatusKonsumen = status_konsumen
						data.Decision = constant.DECISION_PASS
						data.Reason = "NAMA " + nama + " & Baki Debet Sesuai Ketentuan"
					}

				}
			}

			data.PbkReport = pefindo_result.DetailReport
			data.TotalBakiDebet = pefindo_result.TotalBakiDebetNonAgunan

		} else if check_pefindo.Code == "201" || check_pefindo.Result != "UNSCORE" {

			if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
				data.Code = constant.NAMA_SAMA_UNSCORE_RO_AO_CODE
				data.StatusKonsumen = status_konsumen
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + status_konsumen
			} else if status_konsumen == constant.STATUS_KONSUMEN_NEW {
				data.Code = constant.NAMA_SAMA_UNSCORE_NEW_CODE
				data.StatusKonsumen = status_konsumen
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + status_konsumen
			}

		} else if check_pefindo.Code == "202" {
			data.Code = constant.SERVICE_PBK_UNAVAILABLE_CODE
			data.StatusKonsumen = status_konsumen
			data.Reason = "Service PBK tidak tersedia"
		}

	} else {
		data.Code = constant.PBK_NO_HIT
		data.Decision = constant.DECISION_PBK_NO_HIT
		data.Reason = "Akses ke PBK ditutup"

	}

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("FAILED FETCHING DATA KREDITMU")
		return
	}

	return
}

func (u usecase) HitPefindoPrimePriority(reqs request.BodyRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	var updateFiltering entity.ApiDupcheckKmbUpdate

	updateFiltering.RequestID = request_id

	paramPefindo := map[string]string{
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
	}

	if reqs.Data.MaritalStatus == constant.MARRIED {

		paramPefindo = map[string]string{
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
		}
	}

	active, _ := strconv.ParseBool(os.Getenv("ACTIVE_PBK"))
	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_DUPCHECK"))

	if active {
		var check_pefindo response.ResposePefindo

		if dummy {
			getdata, errs := u.GetDummyPBK(reqs.Data.IDNumber)

			if getdata == (entity.DummyPBK{}) {
				check_pefindo.Code = "201"
				check_pefindo.Result = "Pefindo Dummy Data Not Found"

				resp := map[string]string{
					"code":   check_pefindo.Code,
					"result": "Pefindo Dummy Data Not Found",
				}
				ResponsePefindo, _ := json.Marshal(resp)
				updateFiltering.ResultPefindo = ResponsePefindo

			} else {
				if errs != nil {
					err = fmt.Errorf("FAILED FETCHING DATA PEFINDO")
					return
				}

				if err = json.Unmarshal([]byte(getdata.Response), &check_pefindo); err != nil {
					err = fmt.Errorf("KMB FILTERING SERVICE UNAVAILABLE")
					return
				}
				ResponsePefindo, _ := json.Marshal(check_pefindo)
				updateFiltering.ResultPefindo = ResponsePefindo
			}

		} else {

			resp, errs := u.httpclient.CallWebSocket(os.Getenv("PBK_URL"), paramPefindo, map[string]string{}, timeOut)

			if errs != nil || resp.StatusCode() != 200 && resp.StatusCode() != 400 {
				err = fmt.Errorf("failed fetching data pefindo")
				return
			}

			if err = json.Unmarshal(resp.Body(), &check_pefindo); err != nil {
				err = fmt.Errorf("error unsmarshal data pefindo")
				return
			}

			updateFiltering.ResultPefindo = string(resp.Body())

		}

		if check_pefindo.Code == "200" && check_pefindo.Result != "UNSCORE" {

			c, _ := json.Marshal(check_pefindo.Result)
			var pefindo_result response.PefindoResult

			if errs := json.Unmarshal(c, &pefindo_result); errs != nil {
				err = fmt.Errorf("KMB FILTERING SERVICE UNAVAILABLE")
				return
			}
			data.PbkReport = pefindo_result.DetailReport

		} else if check_pefindo.Code == "201" || check_pefindo.Result != "UNSCORE" {

			if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
				data.Code = constant.NAMA_SAMA_UNSCORE_RO_AO_CODE
				data.StatusKonsumen = status_konsumen
				data.Reason = "PBK Tidak Ditemukan - " + status_konsumen
			} else if status_konsumen == constant.STATUS_KONSUMEN_NEW {
				data.Code = constant.NAMA_SAMA_UNSCORE_NEW_CODE
				data.StatusKonsumen = status_konsumen
				data.Reason = "PBK Tidak Ditemukan - " + status_konsumen
			}
		} else if check_pefindo.Code == "202" {
			data.Code = constant.SERVICE_PBK_UNAVAILABLE_CODE
			data.StatusKonsumen = status_konsumen
			data.Reason = "Service PBK tidak tersedia"
		}

	} else {
		data.Code = constant.PBK_NO_HIT
		data.Decision = constant.DECISION_PBK_NO_HIT
		data.Reason = "Akses ke PBK ditutup"

	}

	if err = u.repository.UpdateData(updateFiltering); err != nil {
		err = fmt.Errorf("FAILED FETCHING DATA PEFINDO")
		return
	}

	return
}

func (u usecase) GetDummyPBK(noktp string) (data entity.DummyPBK, err error) {

	query := fmt.Sprintf("SELECT * FROM new_pefindo_kmb WHERE IDNumber = '%s'", noktp)

	data, err = u.repository.DummyDataPbk(query)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetDummyData(noktp string) (data entity.DummyColumn, err error) {

	query := fmt.Sprintf("SELECT * FROM dupcheck_confins_new WHERE NoKTP = '%s'", noktp)

	data, err = u.repository.DummyData(query)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetDataProfessionGroup(prefix string) (data entity.ProfessionGroup, err error) {

	query := fmt.Sprintf("SELECT * FROM profession_group WHERE prefix = '%s'", prefix)

	data, err = u.repository.DataProfessionGroup(query)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetDataGetMappingDp(BranchID, StatusKonsumen string) (data []entity.RangeBranchDp, err error) {

	query := fmt.Sprintf("SELECT mbd.* FROM dbo.mapping_branch_dp mdp LEFT JOIN dbo.mapping_baki_debet mbd ON mdp.baki_debet = mbd.id LEFT JOIN dbo.master_list_dp mld ON mdp.master_list_dp = mld.id WHERE mdp.branch = '%s' AND mdp.customer_status = '%s'", BranchID, StatusKonsumen)

	data, err = u.repository.DataGetMappingDp(query)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetBranchDp(BranchID, StatusKonsumen, ProfessionGroup string, totalBakiDebetNonAgunan int) (data entity.BranchDp, err error) {

	get_data, _ := u.GetDataGetMappingDp(BranchID, StatusKonsumen)

	var queryAdd string

	if totalBakiDebetNonAgunan <= get_data[0].RangeEnd {
		if StatusKonsumen == constant.STATUS_KONSUMEN_NEW {
			queryAdd = fmt.Sprintf("AND a.customer_status = '%s'AND a.profession_group = '%s'", StatusKonsumen, ProfessionGroup)
		} else {
			queryAdd = fmt.Sprintf("AND a.customer_status = '%s'AND a.profession_group IS NULL", StatusKonsumen)
		}
	} else {
		queryAdd = fmt.Sprintf("AND c.range_start <= %d AND c.range_end >= %d", totalBakiDebetNonAgunan, totalBakiDebetNonAgunan)
	}

	query := fmt.Sprintf("SELECT TOP 1 a.branch,a.customer_status,a.profession_group,b.minimal_dp_name,b.minimal_dp_value FROM dbo.mapping_branch_dp a WITH (NOLOCK) INNER JOIN dbo.master_list_dp b WITH (NOLOCK) ON a.master_list_dp = b.id LEFT JOIN dbo.mapping_baki_debet c WITH (NOLOCK) ON a.baki_debet = c.id WHERE a.branch = '%s' %s ORDER BY a.created_at ASC", BranchID, queryAdd)

	data, err = u.repository.BranchDpData(query)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetBranchDpTest(BranchID, StatusKonsumen, ProfessionGroup string, totalBakiDebetNonAgunan int) (data entity.BranchDp, err error) {

	var queryAdd string

	if StatusKonsumen == constant.STATUS_KONSUMEN_NEW {
		queryAdd = fmt.Sprintf("AND a.customer_status = '%s'AND a.profession_group = '%s'", StatusKonsumen, ProfessionGroup)
	} else {
		queryAdd = fmt.Sprintf("AND a.customer_status = '%s'AND a.profession_group IS NULL", StatusKonsumen)
	}

	query := fmt.Sprintf("SELECT TOP 1 a.branch,a.customer_status,a.profession_group,b.minimal_dp_name,b.minimal_dp_value FROM dbo.mapping_branch_dp a WITH (NOLOCK) INNER JOIN dbo.master_list_dp b WITH (NOLOCK) ON a.master_list_dp = b.id LEFT JOIN dbo.mapping_baki_debet c WITH (NOLOCK) ON a.baki_debet = c.id WHERE a.branch = '%s' %s ORDER BY a.created_at ASC", BranchID, queryAdd)

	data, err = u.repository.BranchDpData(query)

	if err != nil {
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
