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

func NewMultiUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, usecase interfaces.Usecase) interfaces.MultiUsecase {
	return &multiUsecase{
		usecase:    usecase,
		repository: repository,
		httpclient: httpclient,
	}
}

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
	}
}

func (u multiUsecase) Filtering(ctx context.Context, req request.Filtering, married bool, accessToken string, hrisAccessToken string) (respFiltering response.Filtering, err error) {

	var (
		customer             []request.SpouseDupcheck
		dataCustomer         []response.SpDupCekCustomerByID
		blackList            response.UsecaseApi
		sp                   response.SpDupCekCustomerByID
		isBlacklist          bool
		resPefindo           response.PefindoResult
		reqPefindo           request.Pefindo
		trxDetailBiro        []entity.TrxDetailBiro
		respFilteringPefindo response.Filtering
		resCMO               response.EmployeeCMOResponse
		resFPD               response.FpdCMOResponse
		bpkbName             bool
		clusterCmo           string
		savedCluster         string
		useDefaultCluster    bool
	)

	requestID := ctx.Value(echo.HeaderXRequestID).(string)

	entityFiltering := entity.FilteringKMB{ProspectID: req.ProspectID, RequestID: requestID, BranchID: req.BranchID, BpkbName: req.BPKBName}

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

			respFiltering = response.Filtering{ProspectID: req.ProspectID, Code: blackList.Code, Decision: blackList.Result, Reason: blackList.Reason, IsBlacklist: isBlacklist}

			entityFiltering.Decision = blackList.Result
			entityFiltering.Reason = blackList.Reason
			entityFiltering.IsBlacklist = 1

			err = u.usecase.SaveFiltering(entityFiltering, trxDetailBiro)

			return
		}
	}

	reqPefindo = request.Pefindo{
		ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
		IDMember:          constant.USER_PBK_KMB_FILTEERING,
		User:              constant.USER_PBK_KMB_FILTEERING,
		ProspectID:        req.ProspectID,
		BranchID:          req.BranchID,
		IDNumber:          req.IDNumber,
		LegalName:         req.LegalName,
		BirthDate:         req.BirthDate,
		SurgateMotherName: req.MotherName,
		Gender:            req.Gender,
		BPKBName:          req.BPKBName,
	}

	if req.Spouse != nil {
		reqPefindo.MaritalStatus = constant.MARRIED
		reqPefindo.SpouseIDNumber = req.Spouse.IDNumber
		reqPefindo.SpouseLegalName = req.Spouse.LegalName
		reqPefindo.SpouseBirthDate = req.Spouse.BirthDate
		reqPefindo.SpouseSurgateMotherName = req.Spouse.MotherName
		reqPefindo.SpouseGender = req.Spouse.Gender
	}

	mainCustomer := dataCustomer[0]
	if mainCustomer.CustomerStatus == "" || mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_NEW {
		mainCustomer.CustomerStatus = constant.STATUS_KONSUMEN_NEW
		mainCustomer.CustomerSegment = constant.RO_AO_REGULAR
	}

	/* Process Get Cluster based on CMO_ID starts here */

	resCMO, err = u.usecase.GetEmployeeData(ctx, req.CMOID, accessToken, hrisAccessToken)
	if err != nil {
		return
	}

	if resCMO.CMOCategory == "" {
		respFiltering.NextProcess = false
	} else {
		bpkbName = strings.Contains(os.Getenv("NAMA_SAMA"), req.BPKBName)
		bpkbString := "NAMA BEDA"
		defaultCluster := constant.CLUSTER_C
		if bpkbName {
			bpkbString = "NAMA SAMA"
			defaultCluster = constant.CLUSTER_B
		}

		if resCMO.CMOCategory == constant.CMO_BARU {
			clusterCmo = defaultCluster
			// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak bulan join_date nya
			useDefaultCluster = true
		} else {
			// Mendapatkan value FPD dari masing-masing jenis BPKB
			resFPD, err = u.usecase.GetFpdCMO(ctx, req.CMOID, bpkbString, accessToken)
			if err != nil {
				return
			}

			if !resFPD.FpdExist {
				clusterCmo = defaultCluster
				// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak tanggal hit filtering nya (assume: today)
				useDefaultCluster = true
			} else {
				// Check Cluster
				var mappingFpdCluster entity.MasterMappingFpdCluster
				mappingFpdCluster, err = u.repository.MasterMappingFpdCluster(resFPD.CmoFpd)
				if err != nil {
					return
				}

				if mappingFpdCluster.Cluster == "" {
					clusterCmo = defaultCluster
					// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak tanggal hit filtering nya (assume: today)
					useDefaultCluster = true
				} else {
					clusterCmo = mappingFpdCluster.Cluster
				}
			}
		}

		if useDefaultCluster == true {
			savedCluster, err = u.usecase.SaveCmoNoFPD(req.ProspectID, req.CMOID, resCMO.CMOCategory, resCMO.JoinDate, clusterCmo)
			if err != nil {
				return
			}
			if savedCluster != "" {
				clusterCmo = savedCluster
			}
		}
	}

	respFiltering.ClusterCMO = clusterCmo

	entityFiltering.CMOID = req.CMOID
	entityFiltering.CMOJoinDate = resCMO.JoinDate
	entityFiltering.CMOCategory = resCMO.CMOCategory
	entityFiltering.CMOFPD = resFPD.CmoFpd
	entityFiltering.CMOAccSales = resFPD.CmoAccSales
	entityFiltering.CMOCluster = respFiltering.ClusterCMO

	/* Process Get Cluster based on CMO_ID ends here */

	// hit ke pefindo
	respFilteringPefindo, resPefindo, trxDetailBiro, err = u.usecase.FilteringPefindo(ctx, reqPefindo, mainCustomer.CustomerStatus, clusterCmo, accessToken)
	if err != nil {
		return
	}

	respFiltering = respFilteringPefindo

	respFiltering.ProspectID = req.ProspectID
	respFiltering.CustomerSegment = mainCustomer.CustomerSegment

	entityFiltering.Cluster = respFiltering.Cluster

	primePriority, _ := utils.ItemExists(mainCustomer.CustomerSegment, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

	if primePriority && (mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_AO || mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_RO) {
		respFiltering.Code = blackList.Code
		respFiltering.Decision = blackList.Result
		respFiltering.Reason = mainCustomer.CustomerStatus + " " + mainCustomer.CustomerSegment
		respFiltering.NextProcess = true

		entityFiltering.Cluster = constant.CLUSTER_PRIME_PRIORITY
	}

	// save transaction
	entityFiltering.Decision = respFiltering.Decision
	entityFiltering.CustomerStatus = mainCustomer.CustomerStatus
	entityFiltering.CustomerSegment = mainCustomer.CustomerSegment
	entityFiltering.CustomerID = mainCustomer.CustomerID

	if respFiltering.NextProcess {
		entityFiltering.NextProcess = 1
	}

	// ada data pefindo
	if resPefindo.Score != "" {
		entityFiltering.MaxOverdueBiro = resPefindo.MaxOverdue
		entityFiltering.MaxOverdueLast12monthsBiro = resPefindo.MaxOverdueLast12Months
		entityFiltering.ScoreBiro = resPefindo.Score

		var isWoContractBiro, isWoWithCollateralBiro int
		if resPefindo.WoContract {
			isWoContractBiro = 1
		}
		if resPefindo.WoAdaAgunan {
			isWoWithCollateralBiro = 1
		}
		entityFiltering.IsWoContractBiro = isWoContractBiro
		entityFiltering.IsWoWithCollateralBiro = isWoWithCollateralBiro

		entityFiltering.TotalInstallmentAmountBiro = resPefindo.AngsuranAktifPbk
		entityFiltering.TotalBakiDebetNonCollateralBiro = resPefindo.TotalBakiDebetNonAgunan

		if resPefindo.Category != nil {
			entityFiltering.Category = resPefindo.Category
		}
		if resPefindo.MaxOverdueKORules != nil {
			entityFiltering.MaxOverdueKORules = resPefindo.MaxOverdueKORules
		}
		if resPefindo.MaxOverdueLast12MonthsKORules != nil {
			entityFiltering.MaxOverdueLast12MonthsKORules = resPefindo.MaxOverdueLast12MonthsKORules
		}
	}

	entityFiltering.Reason = respFiltering.Reason

	err = u.usecase.SaveFiltering(entityFiltering, trxDetailBiro)

	return
}

func (u usecase) FilteringPefindo(ctx context.Context, reqs request.Pefindo, customerStatus, clusterCMO string, accessToken string) (data response.Filtering, responsePefindo response.PefindoResult, trxDetailBiro []entity.TrxDetailBiro, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	var (
		bpkbName bool
	)

	bpkbName = strings.Contains(os.Getenv("NAMA_SAMA"), reqs.BPKBName)

	active, _ := strconv.ParseBool(os.Getenv("ACTIVE_PBK"))
	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_PBK"))

	data.ProspectID = reqs.ProspectID
	data.CustomerStatus = customerStatus

	if active {
		var (
			checkPefindo  response.ResponsePefindo
			pefindoResult response.PefindoResult
		)

		if dummy {

			getData, errDummy := u.repository.DummyDataPbk(reqs.IDNumber)

			if errDummy != nil || getData == (entity.DummyPBK{}) {
				checkPefindo.Code = "201"
				checkPefindo.Result = constant.RESPONSE_PEFINDO_DUMMY_NOT_FOUND
			} else {
				if err = json.Unmarshal([]byte(getData.Response), &checkPefindo); err != nil {
					err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data pefindo dummy")
					return
				}
			}

		} else {

			var resp *resty.Response

			param, _ := json.Marshal(reqs)

			resp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("NEW_KMB_PBK_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, reqs.ProspectID, accessToken)

			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - failed get data pefindo")
				return
			}

			if err = json.Unmarshal(resp.Body(), &checkPefindo); err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data pefindo")
				return
			}
		}

		// Check Cluster
		var mappingCluster entity.MasterMappingCluster
		mappingCluster.BranchID = reqs.BranchID
		mappingCluster.CustomerStatus = constant.STATUS_KONSUMEN_NEW
		bpkbString := "NAMA BEDA"

		if bpkbName {
			bpkbString = "NAMA SAMA"
			mappingCluster.BpkbNameType = 1
		}
		if strings.Contains(constant.STATUS_KONSUMEN_RO_AO, customerStatus) {
			mappingCluster.CustomerStatus = "AO/RO"
		}

		mappingCluster, err = u.repository.MasterMappingCluster(mappingCluster)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Mapping cluster error")
			return
		}

		if mappingCluster.Cluster == "" {
			data.Cluster = constant.CLUSTER_C
		} else {
			data.Cluster = mappingCluster.Cluster
		}

		// handling response pefindo
		if checkPefindo.Code == "200" || checkPefindo.Code == "201" {
			if reflect.TypeOf(checkPefindo.Result).String() != "string" {
				setPefindo, _ := json.Marshal(checkPefindo.Result)

				if errs := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(setPefindo, &pefindoResult); errs != nil {
					err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data pefindo")
					return
				}
			}
		}

		if checkPefindo.Code == "200" && pefindoResult.Score != constant.PEFINDO_UNSCORE {
			if pefindoResult.Category != nil {
				if !bpkbName {
					if pefindoResult.MaxOverdueLast12MonthsKORules != nil {
						if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12MonthsKORules) <= constant.PBK_OVD_LAST_12 {
							if pefindoResult.MaxOverdueKORules == nil {
								data.Code = constant.NAMA_BEDA_CURRENT_OVD_NULL_CODE
								data.CustomerStatus = customerStatus
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) <= constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_BEDA_CURRENT_OVD_UNDER_LIMIT_CODE
								data.CustomerStatus = customerStatus
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) > constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_BEDA_CURRENT_OVD_OVER_LIMIT_CODE
								data.CustomerStatus = customerStatus
								data.Decision = constant.DECISION_REJECT
								data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
							}
						} else {
							data.Code = constant.NAMA_BEDA_12_OVD_OVER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_REJECT
							data.Reason = fmt.Sprintf("NAMA BEDA %s & OVD 12 Bulan Terakhir > %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
						}
					} else {
						data.Code = constant.NAMA_BEDA_12_OVD_NULL_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_PASS
						data.Reason = fmt.Sprintf("NAMA BEDA %s & OVD 12 Bulan Terakhir Null", getReasonCategoryRoman(pefindoResult.Category))
					}
				} else if bpkbName {
					if pefindoResult.MaxOverdueLast12MonthsKORules != nil {
						if checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12MonthsKORules) <= constant.PBK_OVD_LAST_12 {
							if pefindoResult.MaxOverdueKORules == nil {
								data.Code = constant.NAMA_SAMA_CURRENT_OVD_NULL_CODE
								data.CustomerStatus = customerStatus
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) <= constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_SAMA_CURRENT_OVD_UNDER_LIMIT_CODE
								data.CustomerStatus = customerStatus
								data.Decision = constant.DECISION_PASS
								data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", getReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
							} else if checkNullMaxOverdue(pefindoResult.MaxOverdueKORules) > constant.PBK_OVD_CURRENT {
								data.Code = constant.NAMA_SAMA_CURRENT_OVD_OVER_LIMIT_CODE
								data.CustomerStatus = customerStatus
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
							data.CustomerStatus = customerStatus
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
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_PASS
						data.Reason = fmt.Sprintf("NAMA SAMA %s & OVD 12 Bulan Terakhir Null", getReasonCategoryRoman(pefindoResult.Category))
					}
				}

				if data.Decision == constant.DECISION_PASS {
					data.NextProcess = true
				}

				if data.Decision == constant.DECISION_REJECT {
					data.Code = constant.WO_AGUNAN_REJECT_CODE
				}

				var isReasonBakiDebet bool

				// BPKB Nama Sama
				if bpkbName {
					if pefindoResult.WoContract { //Wo Contract Yes

						if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = false
									data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									if data.Decision == constant.DECISION_PASS {
										data.NextProcess = false
										data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = true
										data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
										isReasonBakiDebet = true
									}
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
							}

						} else { //Wo Agunan No
							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									isReasonBakiDebet = true
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									isReasonBakiDebet = true
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							}
						}
					} else { //Wo Contract No
						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								isReasonBakiDebet = true
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
							}

						} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								isReasonBakiDebet = true
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
							}
						} else {
							data.NextProcess = true
							data.Code = constant.WO_AGUNAN_PASS_CODE
							data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
						}
					}
				}

				// BPKB Nama Beda
				if !bpkbName {
					if pefindoResult.WoContract { //Wo Contract Yes

						if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = false
									data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									if data.Decision == constant.DECISION_PASS {
										data.NextProcess = false
										data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
									} else {
										data.NextProcess = true
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
										isReasonBakiDebet = true
									}
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
							}

						} else { //Wo Agunan No
							if customerStatus == constant.STATUS_KONSUMEN_NEW {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									if data.Decision == constant.DECISION_PASS {
										data.NextProcess = true
										data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
										isReasonBakiDebet = true
									} else {
										data.NextProcess = false
										data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
									}
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
								if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
									data.NextProcess = true
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									isReasonBakiDebet = true
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								}
							}
						}
					} else { //Wo Contract No
						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								if data.Decision == constant.DECISION_PASS {
									data.NextProcess = true
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
									isReasonBakiDebet = true
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
								}
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
							}

						} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
								isReasonBakiDebet = true
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, getReasonCategoryRoman(pefindoResult.Category))
							}
						} else {
							data.NextProcess = true
							data.Code = constant.WO_AGUNAN_PASS_CODE
							data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, getReasonCategoryRoman(pefindoResult.Category))
						}
					}
				}

				// Reason Baki Debet
				if (data.Decision == constant.DECISION_REJECT && data.NextProcess) || isReasonBakiDebet {
					if pefindoResult.TotalBakiDebetNonAgunan <= 3000000 {
						data.Reason = fmt.Sprintf("%s %s & Baki Debet <= 3 Juta", bpkbString, getReasonCategoryRoman(pefindoResult.Category))
					}
					if pefindoResult.TotalBakiDebetNonAgunan > 3000000 && pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
						data.Reason = fmt.Sprintf("%s %s & Baki Debet > 3 - 20 Juta", bpkbString, getReasonCategoryRoman(pefindoResult.Category))
					}
				}

				// Reject ovd include all Cluster E, F and bpkbname false
				if (pefindoResult.MaxOverdueLast12Months != nil && checkNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12Months) > constant.PBK_OVD_LAST_12) ||
					(pefindoResult.MaxOverdue != nil && checkNullMaxOverdue(pefindoResult.MaxOverdue) > constant.PBK_OVD_CURRENT) {

					if pefindoResult.TotalBakiDebetNonAgunan > 3000000 && pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET && strings.Contains("Cluster E Cluster F", clusterCMO) {
						data.Reason = fmt.Sprintf("%s %s & Baki Debet > 3 - 20 Juta & Tidak dapat dibiayai", bpkbString, getReasonCategoryRoman(pefindoResult.Category))
						data.NextProcess = false
						data.Decision = constant.DECISION_REJECT
					}

					if !bpkbName {
						data.Reason = fmt.Sprintf("%s & Baki Debet > Threshold", bpkbString)
						data.NextProcess = false
						data.Decision = constant.DECISION_REJECT
					}
				}

				if checkPefindo.Konsumen != (response.PefindoResultKonsumen{}) {
					trxDetailBiroC := entity.TrxDetailBiro{
						ProspectID:                             reqs.ProspectID,
						Subject:                                "CUSTOMER",
						Source:                                 "PBK",
						BiroID:                                 checkPefindo.Konsumen.PefindoID,
						Score:                                  checkPefindo.Konsumen.Score,
						MaxOverdue:                             checkPefindo.Konsumen.MaxOverdue,
						MaxOverdueLast12months:                 checkPefindo.Konsumen.MaxOverdueLast12Months,
						InstallmentAmount:                      checkPefindo.Konsumen.AngsuranAktifPbk,
						WoContract:                             checkPefindo.Konsumen.WoContract,
						WoWithCollateral:                       checkPefindo.Konsumen.WoAdaAgunan,
						BakiDebetNonCollateral:                 checkPefindo.Konsumen.BakiDebetNonAgunan,
						UrlPdfReport:                           checkPefindo.Konsumen.DetailReport,
						Plafon:                                 checkPefindo.Konsumen.Plafon,
						FasilitasAktif:                         checkPefindo.Konsumen.FasilitasAktif,
						KualitasKreditTerburuk:                 checkPefindo.Konsumen.KualitasKreditTerburuk,
						BulanKualitasTerburuk:                  checkPefindo.Konsumen.BulanKualitasTerburuk,
						BakiDebetKualitasTerburuk:              checkPefindo.Konsumen.BakiDebetKualitasTerburuk,
						KualitasKreditTerakhir:                 checkPefindo.Konsumen.KualitasKreditTerakhir,
						BulanKualitasKreditTerakhir:            checkPefindo.Konsumen.BulanKualitasKreditTerakhir,
						OverdueLastKORules:                     checkPefindo.Konsumen.OverdueLastKORules,
						OverdueLast12MonthsKORules:             checkPefindo.Konsumen.OverdueLast12MonthsKORules,
						Category:                               checkPefindo.Konsumen.Category,
						MaxOverdueAgunanKORules:                checkPefindo.Konsumen.MaxOverdueAgunanKORules,
						MaxOverdueAgunanLast12MonthsKORules:    checkPefindo.Konsumen.MaxOverdueAgunanLast12MonthsKORules,
						MaxOverdueNonAgunanKORules:             checkPefindo.Konsumen.MaxOverdueNonAgunanKORules,
						MaxOverdueNonAgunanLast12MonthsKORules: checkPefindo.Konsumen.MaxOverdueNonAgunanLast12MonthsKORules,
					}
					trxDetailBiro = append(trxDetailBiro, trxDetailBiroC)
					data.PbkReportCustomer = &checkPefindo.Konsumen.DetailReport
				}
				if checkPefindo.Pasangan != (response.PefindoResultPasangan{}) {
					trxDetailBiroC := entity.TrxDetailBiro{
						ProspectID:                             reqs.ProspectID,
						Subject:                                "SPOUSE",
						Source:                                 "PBK",
						BiroID:                                 checkPefindo.Pasangan.PefindoID,
						Score:                                  checkPefindo.Pasangan.Score,
						MaxOverdue:                             checkPefindo.Pasangan.MaxOverdue,
						MaxOverdueLast12months:                 checkPefindo.Pasangan.MaxOverdueLast12Months,
						InstallmentAmount:                      checkPefindo.Pasangan.AngsuranAktifPbk,
						WoContract:                             checkPefindo.Pasangan.WoContract,
						WoWithCollateral:                       checkPefindo.Pasangan.WoAdaAgunan,
						BakiDebetNonCollateral:                 checkPefindo.Pasangan.BakiDebetNonAgunan,
						UrlPdfReport:                           checkPefindo.Pasangan.DetailReport,
						Plafon:                                 checkPefindo.Pasangan.Plafon,
						FasilitasAktif:                         checkPefindo.Pasangan.FasilitasAktif,
						KualitasKreditTerburuk:                 checkPefindo.Pasangan.KualitasKreditTerburuk,
						BulanKualitasTerburuk:                  checkPefindo.Pasangan.BulanKualitasTerburuk,
						BakiDebetKualitasTerburuk:              checkPefindo.Pasangan.BakiDebetKualitasTerburuk,
						KualitasKreditTerakhir:                 checkPefindo.Pasangan.KualitasKreditTerakhir,
						BulanKualitasKreditTerakhir:            checkPefindo.Pasangan.BulanKualitasKreditTerakhir,
						OverdueLastKORules:                     checkPefindo.Pasangan.OverdueLastKORules,
						OverdueLast12MonthsKORules:             checkPefindo.Pasangan.OverdueLast12MonthsKORules,
						Category:                               checkPefindo.Pasangan.Category,
						MaxOverdueAgunanKORules:                checkPefindo.Pasangan.MaxOverdueAgunanKORules,
						MaxOverdueAgunanLast12MonthsKORules:    checkPefindo.Pasangan.MaxOverdueAgunanLast12MonthsKORules,
						MaxOverdueNonAgunanKORules:             checkPefindo.Pasangan.MaxOverdueNonAgunanKORules,
						MaxOverdueNonAgunanLast12MonthsKORules: checkPefindo.Pasangan.MaxOverdueNonAgunanLast12MonthsKORules,
					}
					trxDetailBiro = append(trxDetailBiro, trxDetailBiroC)
					data.PbkReportSpouse = &checkPefindo.Pasangan.DetailReport
				}

				data.TotalBakiDebet = pefindoResult.TotalBakiDebetNonAgunan

			} else {
				data.Code = constant.PBK_NO_HIT
				data.CustomerStatus = customerStatus
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK No Hit - Kategori Konsumen Null"
				data.NextProcess = true
			}

		} else if checkPefindo.Code == "201" || pefindoResult.Score == constant.PEFINDO_UNSCORE {

			if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
				data.Code = constant.NAMA_SAMA_UNSCORE_RO_AO_CODE
				data.CustomerStatus = customerStatus
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + customerStatus
				data.NextProcess = true
			} else {
				data.Code = constant.NAMA_SAMA_UNSCORE_NEW_CODE
				data.CustomerStatus = customerStatus
				data.Decision = constant.DECISION_PASS
				data.Reason = "PBK Tidak Ditemukan - " + customerStatus
				data.NextProcess = true
			}

		} else if checkPefindo.Code == "202" {
			data.Code = constant.PBK_NO_HIT
			data.CustomerStatus = customerStatus
			data.Decision = constant.DECISION_PASS
			data.Reason = "No Hit PBK"
			data.NextProcess = true
		}

		responsePefindo = pefindoResult

	} else {
		data.Code = constant.PBK_NO_HIT
		data.CustomerStatus = customerStatus
		data.Decision = constant.DECISION_PASS
		data.Reason = "No Hit PBK"
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
		"lob_id":              2,
	})

	custDupcheck, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("NEW_KMB_DUPCHECK_URL"), req, map[string]string{}, constant.METHOD_POST, false, 0, timeout, prospectID, accessToken)

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

func (u usecase) BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string) {

	customerType = constant.MESSAGE_BERSIH

	if spDupcheck != (response.SpDupCekCustomerByID{}) {

		if spDupcheck.CustomerStatus == "" {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
		} else {
			data.StatusKonsumen = spDupcheck.CustomerStatus
		}

		if spDupcheck.BadType == constant.BADTYPE_B {
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

		}

	} else {
		data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
	}

	data = response.UsecaseApi{StatusKonsumen: data.StatusKonsumen, Code: constant.CODE_NON_BLACKLIST_ALL, Reason: constant.REASON_NON_BLACKLIST, Result: constant.DECISION_PASS}

	return
}

func (u usecase) SaveFiltering(transaction entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro) (err error) {

	err = u.repository.SaveFiltering(transaction, trxDetailBiro)

	if err != nil {

		if strings.Contains(err.Error(), "deadline") {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Save Filtering Timeout")
			return
		}

		err = errors.New(constant.ERROR_BAD_REQUEST + " - Save Filtering Error ProspectID Already Exist")
	}

	return
}

func (u usecase) FilteringProspectID(prospectID string) (data request.OrderIDCheck, err error) {

	row, err := u.repository.GetFilteringByID(prospectID)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Order ID")
	}

	data.ProspectID = prospectID + " - true"

	if row > 0 {
		data.ProspectID = prospectID + " - false"
	}

	return
}

func (u usecase) GetResultFiltering(prospectID string) (respFiltering response.Filtering, err error) {

	getResultFiltering, err := u.repository.GetResultFiltering(prospectID)
	if err != nil || len(getResultFiltering) == 0 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Result Filtering Error")
		return
	}

	respFiltering = response.Filtering{
		ProspectID:      getResultFiltering[0].ProspectID,
		Decision:        getResultFiltering[0].Decision,
		Reason:          getResultFiltering[0].Reason,
		CustomerStatus:  getResultFiltering[0].CustomerStatus,
		CustomerSegment: getResultFiltering[0].CustomerSegment,
		IsBlacklist:     getResultFiltering[0].IsBlacklist,
		NextProcess:     getResultFiltering[0].NextProcess,
	}

	if getResultFiltering[0].TotalBakiDebetNonCollateralBiro != nil {
		var totalBakiDebet float64
		totalBakiDebet, err = utils.GetFloat(getResultFiltering[0].TotalBakiDebetNonCollateralBiro)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetResultFiltering GetFloat Error")
			return
		}
		respFiltering.TotalBakiDebet = totalBakiDebet
	}

	for _, v := range getResultFiltering {
		if v.Subject == "CUSTOMER" {
			respFiltering.PbkReportCustomer = v.UrlPdfReport
		}
		if v.Subject == "SPOUSE" {
			respFiltering.PbkReportSpouse = v.UrlPdfReport
		}
	}

	return
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

func (u usecase) GetEmployeeData(ctx context.Context, employeeID string, accessToken string, hrisAccessToken string) (data response.EmployeeCMOResponse, err error) {

	var (
		respGetEmployeeData response.GetEmployeeByID
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	header := map[string]string{
		"Authorization": "Bearer " + hrisAccessToken,
	}

	payload := request.ReqHrisCareerHistory{
		Limit:     "100",
		Page:      1,
		Column:    "real_career_date",
		Ascending: false,
		Query:     "employee_id==" + employeeID,
	}

	param, _ := json.Marshal(payload)
	getDataEmployee, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("HRIS_GET_EMPLOYEE_DATA_URL"), param, header, constant.METHOD_POST, false, 0, timeout, "", accessToken)

	if getDataEmployee.StatusCode() == 504 || getDataEmployee.StatusCode() == 502 {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Employee Data Timeout")
		return
	}

	if getDataEmployee.StatusCode() != 200 && getDataEmployee.StatusCode() != 504 && getDataEmployee.StatusCode() != 502 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Employee Data Error")
		return
	}

	if err == nil && getDataEmployee.StatusCode() == 200 {
		json.Unmarshal([]byte(jsoniter.Get(getDataEmployee.Body()).ToString()), &respGetEmployeeData)

		// Mencari index terakhir yang mengandung position_group_code "AO"
		var lastIndex int = -1
		for i, emp := range respGetEmployeeData.Data {
			if emp.PositionGroupCode == "AO" {
				lastIndex = i
			}
		}

		if lastIndex == -1 {
			// Jika tidak ada data dengan position_group_code "AO"
			data = response.EmployeeCMOResponse{}
		} else {
			dataEmployee := respGetEmployeeData.Data[lastIndex]
			if dataEmployee.RealCareerDate == "" {
				err = errors.New(constant.ERROR_UPSTREAM + " - RealCareerDate Empty")
			}

			parsedTime, err := time.Parse("2006-01-02T15:04:05", dataEmployee.RealCareerDate)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Parse RealCareerDate")
			}

			dataEmployee.RealCareerDate = parsedTime.Format("2006-01-02")

			today := time.Now().Format("2006-01-02")
			// memvalidasi bulan+tahun yang diberikan tidak lebih besar dari bulan+tahun hari ini
			err = utils.ValidateDiffMonthYear(dataEmployee.RealCareerDate, today)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Validate MonthYear of RealCareerDate")
			}

			todayDate, err := time.Parse("2006-01-02", today)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Parse todayDate")
			}

			givenDate, err := time.Parse("2006-01-02", dataEmployee.RealCareerDate)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Parse givenDate")
			}

			diffOfMonths := utils.DiffInMonths(todayDate, givenDate)

			cmoJoinedAge, _ := strconv.Atoi(os.Getenv("CMO_JOINED_AGE"))
			var cmoCategory string = constant.CMO_BARU
			if diffOfMonths > cmoJoinedAge {
				cmoCategory = constant.CMO_LAMA
			}

			data = response.EmployeeCMOResponse{
				EmployeeID:         dataEmployee.EmployeeID,
				EmployeeName:       dataEmployee.EmployeeName,
				EmployeeIDWithName: dataEmployee.EmployeeID + " - " + dataEmployee.EmployeeName,
				JoinDate:           dataEmployee.RealCareerDate,
				PositionGroupCode:  dataEmployee.PositionGroupCode,
				PositionGroupName:  dataEmployee.PositionGroupName,
				CMOCategory:        cmoCategory,
			}
		}
	} else {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Get Employee Data Error")
		return
	}

	return
}

func (u usecase) GetFpdCMO(ctx context.Context, CmoID string, BPKBNameType string, accessToken string) (data response.FpdCMOResponse, err error) {
	var (
		respGetFPD response.GetFPDCmoByID
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	header := map[string]string{
		"Authorization": accessToken,
	}

	lobID := constant.LOBID_KMB
	cmoID := CmoID
	endpointURL := fmt.Sprintf(os.Getenv("AGREEMENT_LTV_FPD")+"?lob_id=%d&cmo_id=%s", lobID, cmoID)
	fmt.Printf("isi dari endpointURL getFPD : %v\n", endpointURL)

	getDataFpd, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, endpointURL, nil, header, constant.METHOD_GET, false, 0, timeout, "", accessToken)

	if getDataFpd.StatusCode() == 504 || getDataFpd.StatusCode() == 502 {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get FPD Data Timeout")
		return
	}

	if getDataFpd.StatusCode() != 200 && getDataFpd.StatusCode() != 504 && getDataFpd.StatusCode() != 502 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get FPD Data Error")
		return
	}

	if err == nil && getDataFpd.StatusCode() == 200 {
		json.Unmarshal([]byte(jsoniter.Get(getDataFpd.Body()).ToString()), &respGetFPD)

		// Mencari nilai fpd untuk bpkb_name_type "NAMA BEDA"
		var fpdNamaBeda float64 = 0
		var accSalesNamaBeda int = 0
		for _, item := range respGetFPD.Data {
			if item.BpkbNameType == "NAMA BEDA" {
				fpdNamaBeda = item.Fpd
				accSalesNamaBeda = item.AccSales
				break
			}
		}

		// Mencari nilai fpd untuk bpkb_name_type "NAMA SAMA"
		var fpdNamaSama float64 = 0
		var accSalesNamaSama int = 0
		for _, item := range respGetFPD.Data {
			if item.BpkbNameType == "NAMA SAMA" {
				fpdNamaSama = item.Fpd
				accSalesNamaSama = item.AccSales
				break
			}
		}

		if fpdNamaBeda <= 0 && accSalesNamaBeda <= 0 && fpdNamaSama <= 0 && accSalesNamaSama <= 0 {
			// Ini pertanda CMO tidak punya SALES SAMA SEKALI,
			// maka nantinya di usecase akan diarahkan ke Cluster Default sesuai jenis BPKBNameType nya,
			// setelah itu dilakukan proses penyimpanan ke table "trx_cmo_no_fpd" sebagai data penampung rentang tanggal kapan hingga kapan CMO_ID tersebut diarahkan sebagai Default Cluster
			data = response.FpdCMOResponse{}
		} else {
			if BPKBNameType == "NAMA BEDA" {
				data = response.FpdCMOResponse{
					FpdExist:    true,
					CmoFpd:      fpdNamaBeda,
					CmoAccSales: accSalesNamaBeda,
				}
			}

			if BPKBNameType == "NAMA SAMA" {
				data = response.FpdCMOResponse{
					FpdExist:    true,
					CmoFpd:      fpdNamaSama,
					CmoAccSales: accSalesNamaSama,
				}
			}

			if BPKBNameType != "NAMA BEDA" && BPKBNameType != "NAMA SAMA" {
				data = response.FpdCMOResponse{}
			}
		}
	} else {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Get FPD Data Error")
		return
	}

	return
}

func (u usecase) SaveCmoNoFPD(prospectID string, cmoID string, cmoCategory string, cmoJoinDate string, defaultCluster string) (clusterCMOSaved string, err error) {
	var today string

	currentDate := time.Now().Format("2006-01-02")

	if cmoCategory == constant.CMO_LAMA {
		today = currentDate
	} else {
		today = cmoJoinDate
	}

	// Cek apakah CMO_ID sudah pernah tersimpan di dalam table `trx_cmo_no_fpd`
	var TrxCmoNoFpd entity.TrxCmoNoFPD
	TrxCmoNoFpd, err = u.repository.CheckCMONoFPD(cmoID, currentDate)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Check CMO No FPD error")
		return
	}

	clusterCMOSaved = "" // init data for default response
	// Jika CMO_ID sudah ada
	if TrxCmoNoFpd.CMOID != "" {
		if currentDate >= TrxCmoNoFpd.DefaultClusterStartDate && currentDate <= TrxCmoNoFpd.DefaultClusterEndDate {
			// CMO_ID sudah ada dan masih di dalam rentang tanggal `DefaultClusterStartDate` dan `DefaultClusterEndDate`
			defaultCluster = TrxCmoNoFpd.DefaultCluster
			clusterCMOSaved = defaultCluster
		} else {
			today = currentDate
		}
	}

	if clusterCMOSaved == "" {
		// Parsing tanggal hari ini ke dalam format time.Time
		todayTime, err := time.Parse("2006-01-02", today)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - todayTime parse error")
		}

		// Menambahkan 3 bulan
		defaultClusterMonthsDuration, _ := strconv.Atoi(os.Getenv("DEFAULT_CLUSTER_MONTHS_DURATION"))
		threeMonthsLater := todayTime.AddDate(0, defaultClusterMonthsDuration, 0)
		// Mengambil tanggal terakhir dari bulan tersebut
		threeMonthsLater = time.Date(threeMonthsLater.Year(), threeMonthsLater.Month(), 0, 0, 0, 0, 0, threeMonthsLater.Location())
		// Parsing threeMonthsLater ke dalam format "yyyy-mm-dd" sebagai string
		threeMonthsLaterString := threeMonthsLater.Format("2006-01-02")

		entitySaveTrxNoFPd := entity.TrxCmoNoFPD{
			ProspectID:              prospectID,
			CMOID:                   cmoID,
			CmoCategory:             cmoCategory,
			CmoJoinDate:             cmoJoinDate,
			DefaultCluster:          defaultCluster,
			DefaultClusterStartDate: today,
			DefaultClusterEndDate:   threeMonthsLaterString,
			CreatedAt:               time.Time{},
		}

		err = u.repository.SaveTrxCmoNoFPD(entitySaveTrxNoFPd)

		if err != nil {
			if strings.Contains(err.Error(), "deadline") {
				err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Save Trx CMO No FPD Timeout")
			}

			err = errors.New(constant.ERROR_BAD_REQUEST + " - Save Trx CMO No FPD Error ProspectID Already Exist")
		}
	}

	return
}
