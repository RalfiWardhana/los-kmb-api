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
		customer      []request.SpouseDupcheck
		dataCustomer  []response.SpDupCekCustomerByID
		blackList     response.UsecaseApi
		sp            response.SpDupCekCustomerByID
		isBlacklist   bool
		resPefindo    response.PefindoResult
		reqs          request.FilteringRequest
		trxDetailBiro []entity.TrxDetailBiro
	)

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	filtering := entity.FilteringKMB{ProspectID: req.ProspectID, RequestID: requestID, BranchID: req.BranchID, BpkbName: req.BPKBName}

	customer = append(customer, request.SpouseDupcheck{IDNumber: req.IDNumber, LegalName: req.LegalName, BirthDate: req.BirthDate, MotherName: req.MotherName})

	if married {
		customer = append(customer, request.SpouseDupcheck{IDNumber: req.Spouse.IDNumber, LegalName: req.Spouse.LegalName, BirthDate: req.Spouse.BirthDate, MotherName: req.Spouse.MotherName})
	}

	for i := 0; i < len(customer); i++ {

		sp, err = u.usecase.DupcheckIntegrator(ctx, req.ProspectID, customer[i].IDNumber, customer[i].LegalName, customer[i].BirthDate, customer[i].MotherName, accessToken)

		dataCustomer = append(dataCustomer, sp)

		if err != nil {
			return
		}

		blackList, _ = u.usecase.BlacklistCheck(i, sp)

		if blackList.Result == constant.DECISION_REJECT {

			isBlacklist = true

			data = response.Filtering{ProspectID: req.ProspectID, Code: blackList.Code, Decision: blackList.Result, Reason: blackList.Reason, IsBlacklist: isBlacklist}

			filtering.Decision = blackList.Result
			filtering.IsBlacklist = 1

			err = u.usecase.SaveFilteringLogs(filtering, trxDetailBiro)

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
	data, resPefindo, trxDetailBiro, err = u.usecase.FilteringPefindo(ctx, reqs, mainCustomer.CustomerStatus, accessToken)
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

	// save transaction
	filtering.Decision = data.Decision
	filtering.CustomerStatus = mainCustomer.CustomerStatus
	filtering.CustomerSegment = mainCustomer.CustomerSegment

	if data.NextProcess {
		filtering.NextProcess = 1
	}

	filtering.MaxOverdueBiro = resPefindo.MaxOverdue
	filtering.MaxOverdueLast12monthsBiro = resPefindo.MaxOverdueLast12Months

	var isWoContractBiro, isWoWithCollateralBiro int
	if resPefindo.WoContract {
		isWoContractBiro = 1
	}
	if resPefindo.WoAdaAgunan {
		isWoWithCollateralBiro = 1
	}
	filtering.IsWoContractBiro = isWoContractBiro
	filtering.IsWoWithCollateralBiro = isWoWithCollateralBiro

	filtering.TotalInstallmentAmountBiro = resPefindo.AngsuranAktifPbk
	filtering.TotalBakiDebetNonCollateralBiro = resPefindo.TotalBakiDebetNonAgunan

	err = u.usecase.SaveFilteringLogs(filtering, trxDetailBiro)

	return
}

func (u usecase) FilteringPefindo(ctx context.Context, reqs request.FilteringRequest, customerStatus, accessToken string) (data response.Filtering, responsePefindo response.PefindoResult, trxDetailBiro []entity.TrxDetailBiro, err error) {

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

			getData, errDummy := u.repository.DummyDataPbk(reqs.Data.IDNumber)

			if errDummy != nil || getData == (entity.DummyPBK{}) {
				checkPefindo.Code = "201"
				checkPefindo.Result = constant.RESPONSE_PEFINDO_DUMMY_NOT_FOUND
			} else {
				if err = json.Unmarshal([]byte(getData.Response), &checkPefindo); err != nil {
					err = fmt.Errorf("error unmarshal data pefindo dummy")
					return
				}
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
		}

		// handling response pefindo
		if checkPefindo.Code == "200" || checkPefindo.Code == "201" {
			if reflect.TypeOf(checkPefindo.Result).String() != "string" {
				setPefindo, _ := json.Marshal(checkPefindo.Result)

				if errs := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(setPefindo, &pefindoResult); errs != nil {
					err = fmt.Errorf("error unmarshal data pefindo")
					return
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

				data.CustomerStatus = customerStatus
				data.Code = constant.WO_AGUNAN_REJECT_CODE

				// BPKB Nama Sama
				if bpkbName == constant.NAMA_SAMA {
					if pefindoResult.WoContract { //Wo Contract Yes

						if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = false
									data.Reason = "Nama Sama & " + constant.ADA_FASILITAS_WO_AGUNAN
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_RO_AO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = constant.NAMA_SAMA_BAKI_DEBET_SESUAI
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI
								}
							} else {
								data.NextProcess = false
								data.Reason = constant.ADA_FASILITAS_WO_AGUNAN
							}

						} else { //Wo Agunan No
							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = constant.NAMA_SAMA_BAKI_DEBET_SESUAI
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_RO_AO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = constant.NAMA_SAMA_BAKI_DEBET_SESUAI
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI
								}
							}
						}
					} else { //Wo Contract No
						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = constant.NAMA_SAMA_BAKI_DEBET_SESUAI
							} else {
								data.NextProcess = false
								data.Reason = constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI
							}

						} else if customerStatus == constant.STATUS_KONSUMEN_RO_AO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = constant.NAMA_SAMA_BAKI_DEBET_SESUAI
							} else {
								data.NextProcess = false
								data.Reason = constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI
							}
						} else {
							data.NextProcess = true
							data.Code = constant.WO_AGUNAN_PASS_CODE
							data.Reason = constant.TIDAK_ADA_FASILITAS_WO_AGUNAN
						}
					}
				}

				// BPKB Nama Beda
				if bpkbName == constant.NAMA_BEDA {
					if pefindoResult.WoContract { //Wo Contract Yes

						if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = false
									data.Reason = "Nama Beda & " + constant.ADA_FASILITAS_WO_AGUNAN
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_RO_AO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = constant.NAMA_BEDA_BAKI_DEBET_SESUAI
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI
								}
							} else {
								data.NextProcess = false
								data.Reason = constant.ADA_FASILITAS_WO_AGUNAN
							}

						} else { //Wo Agunan No
							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = constant.NAMA_BEDA_BAKI_DEBET_SESUAI
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_RO_AO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = constant.NAMA_BEDA_BAKI_DEBET_SESUAI
								} else {
									data.NextProcess = false
									data.Reason = constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI
								}
							}
						}
					} else { //Wo Contract No
						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = constant.NAMA_BEDA_BAKI_DEBET_SESUAI
							} else {
								data.NextProcess = false
								data.Reason = constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI
							}

						} else if customerStatus == constant.STATUS_KONSUMEN_RO_AO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = constant.NAMA_BEDA_BAKI_DEBET_SESUAI
							} else {
								data.NextProcess = false
								data.Reason = constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI
							}
						} else {
							data.NextProcess = true
							data.Code = constant.WO_AGUNAN_PASS_CODE
							data.Reason = constant.TIDAK_ADA_FASILITAS_WO_AGUNAN
						}
					}
				}
			}

			if checkPefindo.Konsumen != (response.PefindoResultKonsumen{}) {
				trxDetailBiroC := entity.TrxDetailBiro{
					ProspectID:             reqs.Data.ProspectID,
					Subject:                "CUSTOMER",
					Source:                 "PBK",
					BiroID:                 checkPefindo.Konsumen.PefindoID,
					Score:                  checkPefindo.Konsumen.Score,
					MaxOverdue:             checkPefindo.Konsumen.MaxOverdue,
					MaxOverdueLast12months: checkPefindo.Konsumen.MaxOverdueLast12Months,
					InstallmentAmount:      checkPefindo.Konsumen.AngsuranAktifPbk,
					WoContract:             checkPefindo.Konsumen.WoContract,
					WoWithCollateral:       checkPefindo.Konsumen.WoAdaAgunan,
					BakiDebetNonCollateral: checkPefindo.Konsumen.BakiDebetNonAgunan,
					UrlPdfReport:           checkPefindo.Konsumen.DetailReport,
				}
				trxDetailBiro = append(trxDetailBiro, trxDetailBiroC)
				data.PbkReportCustomer = &checkPefindo.Konsumen.DetailReport
			}
			if checkPefindo.Pasangan != (response.PefindoResultPasangan{}) {
				trxDetailBiroC := entity.TrxDetailBiro{
					ProspectID:             reqs.Data.ProspectID,
					Subject:                "SPOUSE",
					Source:                 "PBK",
					BiroID:                 checkPefindo.Pasangan.PefindoID,
					Score:                  checkPefindo.Pasangan.Score,
					MaxOverdue:             checkPefindo.Pasangan.MaxOverdue,
					MaxOverdueLast12months: checkPefindo.Pasangan.MaxOverdueLast12Months,
					InstallmentAmount:      checkPefindo.Pasangan.AngsuranAktifPbk,
					WoContract:             checkPefindo.Pasangan.WoContract,
					WoWithCollateral:       checkPefindo.Pasangan.WoAdaAgunan,
					BakiDebetNonCollateral: checkPefindo.Pasangan.BakiDebetNonAgunan,
					UrlPdfReport:           checkPefindo.Pasangan.DetailReport,
				}
				trxDetailBiro = append(trxDetailBiro, trxDetailBiroC)
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

		responsePefindo = pefindoResult

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

func (u usecase) SaveFilteringLogs(transaction entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro) (err error) {

	err = u.repository.SaveDupcheckResult(transaction, trxDetailBiro)

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
