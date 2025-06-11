package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) Pefindo(ctx context.Context, r request.Pefindo, customerStatus, customerSegment, clusterCMO string, bpkbName string, isOverrideFlowLikeRegular bool) (data response.Filtering, responsePefindo response.PefindoResult, trxDetailBiro []entity.TrxDetailBiro, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("PEFINDO_TIMEOUT"))

	data.ProspectID = r.ProspectID
	data.CustomerStatus = customerStatus

	var (
		checkPefindo  response.ResponsePefindo
		pefindoResult response.PefindoResult
	)

	param, _ := json.Marshal(r)

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("NEW_KMB_PBK_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, r.ProspectID, middlewares.UserInfoData.AccessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - failed get data pefindo")
		return
	}

	if err = json.Unmarshal(resp.Body(), &checkPefindo); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data pefindo")
		return
	}

	// Check Cluster
	var mappingCluster entity.MasterMappingCluster

	mappingCluster.BranchID = r.BranchID
	mappingCluster.CustomerStatus = constant.STATUS_KONSUMEN_NEW

	var namaSama bool

	if bpkbName == constant.BPKB_NAMA_SAMA {
		mappingCluster.BpkbNameType = 1
		namaSama = true
	}
	if strings.Contains(constant.STATUS_KONSUMEN_RO_AO, customerStatus) {
		mappingCluster.CustomerStatus = "AO/RO"
	}

	mappingCluster, err = u.repository.MasterMappingCluster(mappingCluster)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - mapping cluster error")
		return
	}

	if mappingCluster.Cluster == "" {
		data.Cluster = constant.CLUSTER_C
	} else {
		data.Cluster = mappingCluster.Cluster
	}

	// handling response pefindo
	if checkPefindo.Code == strconv.Itoa(http.StatusOK) || checkPefindo.Code == strconv.Itoa(http.StatusCreated) {
		if reflect.TypeOf(checkPefindo.Result).String() != "string" {
			setPefindo, _ := json.Marshal(checkPefindo.Result)

			if errs := jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(setPefindo, &pefindoResult); errs != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data pefindo")
				return
			}
		}
	}

	if checkPefindo.Code == strconv.Itoa(http.StatusOK) && pefindoResult.Score != constant.PEFINDO_UNSCORE {
		isRejectPefindo := false
		// check inquiry pbk
		primePriority, _ := utils.ItemExists(customerSegment, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})
		if (!((customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO) && primePriority) || isOverrideFlowLikeRegular) && pefindoResult.InquiriesLast1Month > 10 {
			data.Code = constant.CODE_REJECT_INQUIRY_1MONTHS
			data.Reason = constant.REASON_INQUIRY_1MONTHS
			isRejectPefindo = true
		} else {
			// new ko rules pbk
			if pefindoResult.NewKoRules != (response.ResultNewKoRules{}) && pefindoResult.NewKoRules.CategoryPBK != "" {
				if pefindoResult.NewKoRules.CategoryPBK == constant.REJECT_LUNAS_DISKON {
					data.Code = constant.CODE_REJECT_LUNAS_DISKON
					data.Reason = constant.REASON_LUNAS_DISKON
					isRejectPefindo = true
				} else if pefindoResult.NewKoRules.CategoryPBK == constant.REJECT_FASILITAS_DIALIHKAN_DIJUAL {
					data.Code = constant.CODE_REJECT_FASILITAS_DIALIHKAN_DIJUAL
					data.Reason = constant.REASON_FASILITAS_DIALIHKAN_DIJUAL
					isRejectPefindo = true
				} else if pefindoResult.NewKoRules.CategoryPBK == constant.REJECT_HAPUS_TAGIH {
					data.Code = constant.CODE_REJECT_HAPUS_TAGIH
					data.Reason = constant.REASON_HAPUS_TAGIH
					isRejectPefindo = true
				} else if pefindoResult.NewKoRules.CategoryPBK == constant.REJECT_REPOSSES {
					data.Code = constant.CODE_REJECT_REPOSSES
					data.Reason = constant.REASON_REPOSSES
					isRejectPefindo = true
				} else if pefindoResult.NewKoRules.CategoryPBK == constant.REJECT_RESTRUCTURE {
					data.Code = constant.CODE_REJECT_RESTRUCTURE
					data.Reason = constant.REASON_RESTRUCTURE
					isRejectPefindo = true
				}
			}
		}

		if isRejectPefindo {
			data.CustomerStatus = customerStatus
			data.Decision = constant.DECISION_REJECT
			data.NextProcess = false

			if checkPefindo.Konsumen != (response.PefindoResultKonsumen{}) {
				trxDetailBiroC := entity.TrxDetailBiro{
					ProspectID:                             r.ProspectID,
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
					ProspectID:                             r.ProspectID,
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

			responsePefindo = pefindoResult

			return
		}

		if pefindoResult.Category != nil {
			if !namaSama {
				if pefindoResult.MaxOverdueLast12MonthsKORules != nil {
					if utils.CheckNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12MonthsKORules) <= constant.PBK_OVD_LAST_12 {
						if pefindoResult.MaxOverdueKORules == nil {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_NULL_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
						} else if utils.CheckNullMaxOverdue(pefindoResult.MaxOverdueKORules) <= constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_UNDER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						} else if utils.CheckNullMaxOverdue(pefindoResult.MaxOverdueKORules) > constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_BEDA_CURRENT_OVD_OVER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_REJECT
							data.Reason = fmt.Sprintf("NAMA BEDA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						}
					} else {
						data.Code = constant.NAMA_BEDA_12_OVD_OVER_LIMIT_CODE
						data.CustomerStatus = customerStatus
						data.Decision = constant.DECISION_REJECT
						data.Reason = fmt.Sprintf("NAMA BEDA %s & OVD 12 Bulan Terakhir > %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
					}
				} else {
					data.Code = constant.NAMA_BEDA_12_OVD_NULL_CODE
					data.CustomerStatus = customerStatus
					data.Decision = constant.DECISION_PASS
					data.Reason = fmt.Sprintf("NAMA BEDA %s & OVD 12 Bulan Terakhir Null", utils.GetReasonCategoryRoman(pefindoResult.Category))
				}
			} else {
				if pefindoResult.MaxOverdueLast12MonthsKORules != nil {
					if utils.CheckNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12MonthsKORules) <= constant.PBK_OVD_LAST_12 {
						if pefindoResult.MaxOverdueKORules == nil {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_NULL_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)
						} else if utils.CheckNullMaxOverdue(pefindoResult.MaxOverdueKORules) <= constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_UNDER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Decision = constant.DECISION_PASS
							data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current <= %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)
						} else if utils.CheckNullMaxOverdue(pefindoResult.MaxOverdueKORules) > constant.PBK_OVD_CURRENT {
							data.Code = constant.NAMA_SAMA_CURRENT_OVD_OVER_LIMIT_CODE
							data.CustomerStatus = customerStatus
							data.Reason = fmt.Sprintf("NAMA SAMA %s & PBK OVD 12 Bulan Terakhir <= %d & OVD Current > %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12, constant.PBK_OVD_CURRENT)

							data.Decision = func() string {
								if utils.CheckNullCategory(pefindoResult.Category) == 3 {
									return constant.DECISION_REJECT
								}
								return constant.DECISION_PASS
							}()
						}
					} else {
						data.Code = constant.NAMA_SAMA_12_OVD_OVER_LIMIT_CODE
						data.CustomerStatus = customerStatus
						data.Reason = fmt.Sprintf("NAMA SAMA %s & OVD 12 Bulan Terakhir > %d", utils.GetReasonCategoryRoman(pefindoResult.Category), constant.PBK_OVD_LAST_12)

						data.Decision = func() string {
							if utils.CheckNullCategory(pefindoResult.Category) == 3 {
								return constant.DECISION_REJECT
							}
							return constant.DECISION_PASS
						}()
					}
				} else {
					data.Code = constant.NAMA_SAMA_12_OVD_NULL_CODE
					data.CustomerStatus = customerStatus
					data.Decision = constant.DECISION_PASS
					data.Reason = fmt.Sprintf("NAMA SAMA %s & OVD 12 Bulan Terakhir Null", utils.GetReasonCategoryRoman(pefindoResult.Category))
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
			if namaSama {
				if pefindoResult.WoContract { //Wo Contract Yes

					if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = false
								data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								if data.Decision == constant.DECISION_PASS {
									data.NextProcess = false
									data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
								} else {
									data.NextProcess = true
									data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
									isReasonBakiDebet = true
								}
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						} else {
							data.NextProcess = false
							data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
						}

					} else { //Wo Agunan No
						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
								isReasonBakiDebet = true
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
								isReasonBakiDebet = true
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						}
					}
				} else { //Wo Contract No
					if customerStatus == constant.STATUS_KONSUMEN_NEW {
						if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
							data.NextProcess = true
							data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							isReasonBakiDebet = true
						} else {
							data.NextProcess = false
							data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
						}

					} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
						if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
							data.NextProcess = true
							data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							isReasonBakiDebet = true
						} else {
							data.NextProcess = false
							data.Reason = fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
						}
					} else {
						data.NextProcess = true
						data.Code = constant.WO_AGUNAN_PASS_CODE
						data.Reason = fmt.Sprintf("NAMA SAMA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
					}
				}
			}

			// BPKB Nama Beda
			if !namaSama {
				if pefindoResult.WoContract { //Wo Contract Yes

					if pefindoResult.WoAdaAgunan { //Wo Agunan Yes

						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = false
								data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								if data.Decision == constant.DECISION_PASS {
									data.NextProcess = false
									data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
								} else {
									data.NextProcess = true
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
									isReasonBakiDebet = true
								}
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						} else {
							data.NextProcess = false
							data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
						}

					} else { //Wo Agunan No
						if customerStatus == constant.STATUS_KONSUMEN_NEW {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								if data.Decision == constant.DECISION_PASS {
									data.NextProcess = true
									data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
									isReasonBakiDebet = true
								} else {
									data.NextProcess = false
									data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
								}
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
							if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
								data.NextProcess = true
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
								isReasonBakiDebet = true
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						}
					}
				} else { //Wo Contract No
					if customerStatus == constant.STATUS_KONSUMEN_NEW {
						if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
							if data.Decision == constant.DECISION_PASS {
								data.NextProcess = true
								data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
								isReasonBakiDebet = true
							} else {
								data.NextProcess = false
								data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
							}
						} else {
							data.NextProcess = false
							data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
						}

					} else if customerStatus == constant.STATUS_KONSUMEN_AO || customerStatus == constant.STATUS_KONSUMEN_RO {
						if pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
							data.NextProcess = true
							data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
							isReasonBakiDebet = true
						} else {
							data.NextProcess = false
							data.Reason = fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, utils.GetReasonCategoryRoman(pefindoResult.Category))
						}
					} else {
						data.NextProcess = true
						data.Code = constant.WO_AGUNAN_PASS_CODE
						data.Reason = fmt.Sprintf("NAMA BEDA %s & "+constant.TIDAK_ADA_FASILITAS_WO_AGUNAN, utils.GetReasonCategoryRoman(pefindoResult.Category))
					}
				}
			}

			// Reason Baki Debet
			if (data.Decision == constant.DECISION_REJECT && data.NextProcess) || isReasonBakiDebet {
				if pefindoResult.TotalBakiDebetNonAgunan <= constant.RANGE_CLUSTER_BAKI_DEBET_REJECT {
					data.Reason = fmt.Sprintf("%s %s %s", bpkbName, utils.GetReasonCategoryRoman(pefindoResult.Category), constant.WORDING_BAKIDEBET_LOWERTHAN_THRESHOLD)
				}
				if pefindoResult.TotalBakiDebetNonAgunan > constant.RANGE_CLUSTER_BAKI_DEBET_REJECT && pefindoResult.TotalBakiDebetNonAgunan <= constant.BAKI_DEBET {
					data.Reason = fmt.Sprintf("%s %s %s", bpkbName, utils.GetReasonCategoryRoman(pefindoResult.Category), constant.WORDING_BAKIDEBET_HIGHERTHAN_THRESHOLD)
				}
			}

			// Reject ovd include all Cluster E, F and bpkbname false
			if (pefindoResult.MaxOverdueLast12Months != nil && utils.CheckNullMaxOverdueLast12Months(pefindoResult.MaxOverdueLast12Months) > constant.PBK_OVD_LAST_12) ||
				(pefindoResult.MaxOverdue != nil && utils.CheckNullMaxOverdue(pefindoResult.MaxOverdue) > constant.PBK_OVD_CURRENT) {

				// -- Update CR 2024-09-12: change `Cluster E Cluster F`` from REJECT to PASS -- //

				// Reason ovd include all
				if data.NextProcess && !namaSama {
					data.Reason = fmt.Sprintf("%s & %s", bpkbName, constant.REJECT_REASON_OVD_PEFINDO)
					data.NextProcess = false
					data.Decision = constant.DECISION_REJECT
				}
			}

			if checkPefindo.Konsumen != (response.PefindoResultKonsumen{}) {
				trxDetailBiroC := entity.TrxDetailBiro{
					ProspectID:                             r.ProspectID,
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
					ProspectID:                             r.ProspectID,
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

	} else if checkPefindo.Code == strconv.Itoa(http.StatusCreated) || pefindoResult.Score == constant.PEFINDO_UNSCORE {

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

	return
}
