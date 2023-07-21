package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/filtering_new/interfaces"
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
	"time"

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

func (u multiUsecase) Filtering(ctx context.Context, req request.Filtering, married bool, accessToken string) (data response.Filtering, err error) {

	var (
		customer     []request.SpouseDupcheck
		dataCustomer []response.SpDupCekCustomerByID
		blackList    response.UsecaseApi
		sp           response.SpDupCekCustomerByID
		isBlacklist  bool
		resPefindo   interface{}
		reqs         request.FilteringRequest
	)

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	reqJson, _ := json.Marshal(req)

	requestTime := time.Now()

	filtering := entity.FilteringKMB{ProspectID: req.ProspectID, RequestID: requestID, DtmRequest: requestTime, Request: string(reqJson)}

	customer = append(customer, request.SpouseDupcheck{IDNumber: req.IDNumber, LegalName: req.LegalName, BirthDate: req.BirthDate, MotherName: req.MotherName})

	if married {
		customer = append(customer, request.SpouseDupcheck{IDNumber: req.Spouse.IDNumber, LegalName: req.Spouse.LegalName, BirthDate: req.Spouse.BirthDate, MotherName: req.Spouse.MotherName})
	}

	for i := 0; i < len(customer); i++ {

		sp, err = u.usecase.DupcheckIntegrator(ctx, req.ProspectID, customer[i].IDNumber, customer[i].LegalName, customer[i].BirthDate, customer[i].MotherName, accessToken)

		dataCustomer = append(dataCustomer, sp)

		if dataCustomer[i] != (response.SpDupCekCustomerByID{}) {
			dupcheckData, _ := json.Marshal(dataCustomer[i])

			if i > 0 {
				filtering.ResultDupcheckPasangan = string(dupcheckData)
			} else {
				filtering.ResultDupcheckKonsumen = string(dupcheckData)
			}
		}

		if err != nil {
			return
		}

		blackList, _ = u.usecase.BlacklistCheck(i, sp)

		if blackList.Result == constant.DECISION_REJECT {

			isBlacklist = true

			data = response.Filtering{ProspectID: req.ProspectID, Code: blackList.Code, Decision: blackList.Result, Reason: blackList.Reason, IsBlacklist: isBlacklist}

			response, _ := json.Marshal(data)
			filtering.Response = string(response)
			filtering.DtmResponse = time.Now()
			filtering.Decision = blackList.Result

			err = u.usecase.SaveFilteringLogs(filtering)

			return
		}
	}

	reqs = request.FilteringRequest{
		Data: request.Data{
			ProspectID:        req.ProspectID,
			BranchID:          req.BranchID,
			IDNumber:          req.IDNumber,
			LegalName:         req.LegalName,
			BirthDate:         req.BirthDate,
			SurgateMotherName: req.MotherName,
			Gender:            req.Gender,
			BPKBName:          req.BPKBName,
		},
	}

	if req.Spouse != nil {
		reqs.Data.MaritalStatus = constant.MARRIED
		dataSpouse := request.Spouse{
			IDNumber:          req.Spouse.IDNumber,
			LegalName:         req.Spouse.LegalName,
			BirthDate:         req.Spouse.BirthDate,
			SurgateMotherName: req.Spouse.MotherName,
			Gender:            req.Spouse.Gender,
		}
		reqs.Data.Spouse = &dataSpouse
	}

	mainCustomer := dataCustomer[0]
	if mainCustomer.CustomerStatus == "" {
		mainCustomer.CustomerStatus = constant.STATUS_KONSUMEN_NEW
		mainCustomer.CustomerSegment = constant.RO_AO_REGULAR
	}

	// hit ke pefindo
	data, resPefindo, err = u.usecase.FilteringPefindo(ctx, reqs, mainCustomer.CustomerStatus, accessToken)
	if err != nil {
		return
	}

	data.ProspectID = req.ProspectID

	primePriority, _ := utils.ItemExists(mainCustomer.CustomerSegment, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

	if primePriority && (mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_AO || mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_RO) {
		data.Code = blackList.Code
		data.Decision = blackList.Result
		data.Reason = blackList.Reason
	}

	response, _ := json.Marshal(data)
	filtering.ResultPefindo = resPefindo
	filtering.Response = string(response)
	filtering.DtmResponse = time.Now()
	filtering.Decision = data.Decision
	filtering.SourceCreditBiro = "PBK"
	filtering.StatusKonsumen = mainCustomer.CustomerStatus

	err = u.usecase.SaveFilteringLogs(filtering)

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

	return
}

func (u usecase) FilteringPefindo(ctx context.Context, reqs request.FilteringRequest, customerStatus, accessToken string) (data response.Filtering, responsePefindo interface{}, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	var (
		bpkbName string
	)

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
	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_PBK"))

	if active {
		var (
			checkPefindo  response.ResponsePefindo
			pefindoResult response.PefindoResult
		)

		if dummy {

			var getData entity.DummyPBK

			getData, err = u.repository.DummyDataPbk(reqs.Data.IDNumber)

			if getData == (entity.DummyPBK{}) {
				checkPefindo.Code = "201"
				checkPefindo.Result = constant.RESPONSE_PEFINDO_DUMMY_NOT_FOUND

				resp := map[string]string{
					"code":   checkPefindo.Code,
					"result": constant.RESPONSE_PEFINDO_DUMMY_NOT_FOUND,
				}
				responsePefindo, _ = json.Marshal(resp)

			} else {
				if err != nil {
					err = fmt.Errorf("failed get data dummy pefindo")
					return
				}

				if err = json.Unmarshal([]byte(getData.Response), &checkPefindo); err != nil {
					err = fmt.Errorf("error unmarshal data pefindo dummy")
					return
				}

				responsePefindo, _ = json.Marshal(checkPefindo)

			}

		} else {

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

		}

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
			if bpkbName == constant.NAMA_BEDA {
				if pefindoResult.MaxOverdueLast12Months != nil {
					if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12Months) <= constant.PBK_OVD_LAST_12 {
						if pefindoResult.MaxOverdue == nil {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_NULL_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA BEDA & PBK OVD 12 Bulan Terakhir <= %d", constant.PBK_OVD_LAST_12)
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) <= constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_UNDER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA BEDA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) > constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_OVER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_REJECT
							data.Reason = fmt.Sprintf("NAMA BEDA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						}
					} else {
						data.Code = constant.NAMA_BEDA_12_OVD_OVER_LIMIT_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_REJECT
						data.Reason = fmt.Sprintf("NAMA BEDA & OVD 12 Bulan Terakhir > %d", constant.PBK_OVD_LAST_12)
					}
				} else {
					data.Code = constant.NAMA_BEDA_12_OVD_NULL_CODE
					data.CustomerStatus = customerStatus
					data.Decision = constant.DECISION_PASS
					data.Reason = "NAMA BEDA & OVD 12 Bulan Terakhir Null"
				}
			} else if bpkbName == constant.NAMA_SAMA {
				if pefindoResult.MaxOverdueLast12Months != nil {
					if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12Months) <= constant.PBK_OVD_LAST_12 {
						if pefindoResult.MaxOverdue == nil {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_NULL_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA SAMA & PBK OVD 12 Bulan Terakhir <= %d", constant.PBK_OVD_LAST_12)
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) <= constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_UNDER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA SAMA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						} else if checkNullMaxOverdue(pefindoResult.MaxOverdue) > constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_OVER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_REJECT
							data.NextProcess = false
							data.Reason = fmt.Sprintf("NAMA SAMA & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						}
					} else {
						data.Code = constant.NAMA_SAMA_12_OVD_OVER_LIMIT_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_REJECT
						data.NextProcess = false
						data.Reason = fmt.Sprintf("NAMA SAMA & OVD 12 Bulan Terakhir > %d", constant.PBK_OVD_LAST_12)
					}
				} else {
					data.Code = constant.NAMA_SAMA_12_OVD_NULL_CODE
					data.CustomerStatus = customerStatus
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
				if pefindoResult.WoContract {

					if !pefindoResult.WoAdaAgunan { //wo_agunan No
						if pefindoResult.TotalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data.Code = constant.WO_AGUNAN_REJECT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_REJECT
							data.NextProcess = false
							data.Reason = "NAMA " + nama + " & Baki Debet > 20 Juta"
						} else if bpkbName == constant.NAMA_SAMA && pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
							data.Code = constant.WO_AGUNAN_PASS_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_REJECT
							data.NextProcess = true
							data.Reason = "NAMA " + nama + " & Baki Debet Sesuai Ketentuan"
						} else {
							data.Code = constant.WO_AGUNAN_PASS_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_REJECT
							data.NextProcess = false
							data.Reason = "NAMA " + nama + " & Baki Debet Sesuai Ketentuan"
						}

					} else { //wo_agunan Yes
						data.Code = constant.WO_AGUNAN_REJECT_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_REJECT
						data.NextProcess = false
						data.Reason = "NAMA " + nama + " & Ada Fasilitas WO Agunan"
					}
				} else { //wo_contract No
					if pefindoResult.TotalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data.Code = constant.WO_AGUNAN_REJECT_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_REJECT
						data.NextProcess = false
						data.Reason = "NAMA " + nama + " & Baki Debet > 20 Juta"
					} else if bpkbName == constant.NAMA_SAMA && pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
						data.Code = constant.WO_AGUNAN_PASS_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_REJECT
						data.NextProcess = true
						data.Reason = "NAMA " + nama + " & Baki Debet Sesuai Ketentuan"
					} else {
						data.Code = constant.WO_AGUNAN_PASS_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_REJECT
						data.NextProcess = false
						data.Reason = "NAMA " + nama + " & Baki Debet Sesuai Ketentuan"
					}

				}
			}

			if checkPefindo.Konsumen != (response.PefindoResultKonsumen{}) {
				data.PbkReportCustomer = &checkPefindo.Konsumen.DetailReport
			}
			if checkPefindo.Pasangan != (response.PefindoResultPasangan{}) {
				data.PbkReportSpouse = &checkPefindo.Pasangan.DetailReport
			}

		} else if checkPefindo.Code == "201" || pefindoResult.Score == constant.PEFINDO_UNSCORE {

			if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
				data.Code = constant.NAMA_SAMA_UNSCORE_RO_AO_CODE
				data.CustomerStatus = customerStatus
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + customerStatus
			} else {
				data.Code = constant.NAMA_SAMA_UNSCORE_NEW_CODE
				data.CustomerStatus = customerStatus
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + customerStatus
			}

		} else if checkPefindo.Code == "202" {
			data.Code = constant.SERVICE_PBK_UNAVAILABLE_CODE
			data.CustomerStatus = customerStatus
			data.Reason = "Service PBK tidak tersedia"
		}

	} else {
		data.Code = constant.PBK_NO_HIT
		data.CustomerStatus = customerStatus
		data.Reason = "Akses ke PBK ditutup"

	}

	if data.Decision == constant.DECISION_PASS {
		data.NextProcess = true
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

func (u usecase) DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error) {

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	req, _ := json.Marshal(map[string]interface{}{
		"transaction_id":      prospectID,
		"id_number":           idNumber,
		"legal_name":          legalName,
		"birth_date":          birthDate,
		"surgate_mother_name": surgateName,
	})

	custDupcheck, err := u.httpclient.EngineAPI(ctx, constant.FILTERING_LOG, os.Getenv("DUPCHECK_URL"), req, map[string]string{}, constant.METHOD_POST, false, 0, timeout, prospectID, accessToken)

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

func (u usecase) BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string) {

	customerType = constant.MESSAGE_BERSIH

	if spDupcheck != (response.SpDupCekCustomerByID{}) {

		data.StatusKonsumen = constant.STATUS_KONSUMEN_AO

		if (spDupcheck.TotalInstallment <= 0 && spDupcheck.RRDDate != nil) || (spDupcheck.TotalInstallment > 0 && spDupcheck.RRDDate != nil && spDupcheck.NumberOfPaidInstallment == nil) {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_RO
		}

		if spDupcheck.IsSimiliar == 1 && index == 0 {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_KONSUMEN_SIMILIAR
			data.Reason = constant.REASON_KONSUMEN_SIMILIAR
			customerType = constant.MESSAGE_BLACKLIST

		} else if spDupcheck.BadType == constant.BADTYPE_B {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST
			}
			return

		} else if spDupcheck.MaxOverdueDays > 90 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_OVD_90DAYS

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_OVD_90DAYS
			}
			return

		} else if spDupcheck.NumOfAssetInventoried > 0 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_ASSET_INVENTORY

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_ASSET_INVENTORY
			}
			return

		} else if spDupcheck.IsRestructure == 1 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_RESTRUCTURE

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_RESTRUCTURE
			}
			return

		} else if spDupcheck.BadType == constant.BADTYPE_W {
			customerType = constant.MESSAGE_WARNING
		}

	} else {
		data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
	}

	data = response.UsecaseApi{StatusKonsumen: data.StatusKonsumen, Code: constant.CODE_NON_BLACKLIST, Reason: constant.REASON_NON_BLACKLIST, Result: constant.DECISION_PASS}

	return
}

func (u usecase) SaveFilteringLogs(transaction entity.FilteringKMB) (err error) {

	err = u.repository.SaveDupcheckResult(transaction)

	if err != nil {

		if strings.Contains(err.Error(), "deadline") {
			err = errors.New("upstream_service_timeout - Save Filtering Timeout")
			return
		}

		err = errors.New("upstream_service_error - Save Filtering Error")
	}

	return
}

func (u usecase) FilteringProspectID(prospectID string) (data request.OrderIDCheck, err error) {

	row, err := u.repository.GetFilteringByID(prospectID)

	if err != nil {
		err = errors.New("upstream_service_error - Get Filtering Order ID")
	}

	data.ProspectID = prospectID + " - true"

	if row > 0 {
		data.ProspectID = prospectID + " - false"
	}

	return
}
