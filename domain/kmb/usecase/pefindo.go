package usecase

import (
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strings"
)

func (u usecase) Pefindo(cbFound bool, bpkbName string, filtering entity.FilteringKMB, spDupcheck response.SpDupcheckMap) (data response.UsecaseApi, err error) {

	var customerSegment string
	if filtering.CustomerSegment != nil {
		customerSegment = filtering.CustomerSegment.(string)
	}

	if customerSegment == constant.RO_AO_PRIME || customerSegment == constant.RO_AO_PRIORITY {
		if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_AO && spDupcheck.InstallmentTopup == 0 && spDupcheck.NumberOfPaidInstallment >= 6 {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_PRIME_PRIORITY,
				Reason:         fmt.Sprintf("%s %s >= 6 bulan - PBK Pass", spDupcheck.StatusKonsumen, customerSegment),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
			return
		}

		if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_RO || (spDupcheck.InstallmentTopup > 0 && spDupcheck.MaxOverdueDaysforActiveAgreement <= 30) {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_PRIME_PRIORITY,
				Reason:         fmt.Sprintf("%s %s - PBK Pass", spDupcheck.StatusKonsumen, customerSegment),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
			return
		}
	}

	if cbFound {
		var (
			maxOverdueDays          float64
			maxOverdueLast12Months  float64
			isWoContractBiro        float64
			isWoWithCollateralBiro  float64
			totalBakiDebetNonAgunan float64
		)

		if filtering.MaxOverdueKORules != nil {
			maxOverdueDays, err = utils.GetFloat(filtering.MaxOverdueKORules)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat MaxOverdueBiro Pefindo Error")
				return
			}
		} else {
			if filtering.MaxOverdueBiro != nil {
				maxOverdueDays, err = utils.GetFloat(filtering.MaxOverdueBiro)
				if err != nil {
					err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat MaxOverdueBiro Pefindo Error")
					return
				}
			}
		}

		if filtering.MaxOverdueLast12MonthsKORules != nil {
			maxOverdueLast12Months, err = utils.GetFloat(filtering.MaxOverdueLast12MonthsKORules)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat MaxOverdueLast12monthsBiro Pefindo Error")
				return
			}
		} else {
			if filtering.MaxOverdueLast12monthsBiro != nil {
				maxOverdueLast12Months, err = utils.GetFloat(filtering.MaxOverdueLast12monthsBiro)
				if err != nil {
					err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat MaxOverdueLast12monthsBiro Pefindo Error")
					return
				}
			}
		}

		if maxOverdueLast12Months > constant.PBK_OVD_LAST_12 {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_OVD12GT60,
				Reason:         fmt.Sprintf(constant.REASON_PEFINDO_OVD12GT60, constant.PBK_OVD_LAST_12),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
		} else if maxOverdueDays > constant.PBK_OVD_CURRENT {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_CURRENT_GT30,
				Reason:         fmt.Sprintf(constant.REASON_PEFINDO_CURRENT_GT30, constant.PBK_OVD_CURRENT),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
		} else {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_OVD12LTE60_CURRENT_LTE30,
				Reason:         fmt.Sprintf(constant.REASON_PEFINDO_OVD12LTE60_CURRENT_LTE30, constant.PBK_OVD_CURRENT),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
		}

		if data.Result == constant.DECISION_REJECT {
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) {
				if filtering.IsWoContractBiro != nil {
					isWoContractBiro, err = utils.GetFloat(filtering.IsWoContractBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat IsWoContractBiro Pefindo Error")
						return
					}
				}

				if filtering.IsWoWithCollateralBiro != nil {
					isWoWithCollateralBiro, err = utils.GetFloat(filtering.IsWoWithCollateralBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat IsWoWithCollateralBiro Pefindo Error")
						return
					}
				}

				if filtering.TotalBakiDebetNonCollateralBiro != nil {
					totalBakiDebetNonAgunan, err = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat TotalBakiDebetNonCollateralBiro Pefindo Error")
						return
					}
				}

				if isWoContractBiro > 0 {
					if isWoWithCollateralBiro > 0 {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
							Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_SAMA, constant.ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
								Reason:         constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI,
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
								Reason:         constant.NAMA_SAMA_BAKI_DEBET_SESUAI,
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
							Reason:         constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
							Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_SAMA, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_PASS,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					}
				}
			} else {
				data = response.UsecaseApi{
					Code:           constant.CODE_PEFINDO_BPKB_BEDA,
					Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_BEDA, data.Reason),
					Result:         constant.DECISION_REJECT,
					SourceDecision: constant.SOURCE_DECISION_BIRO,
				}
			}
		}

		if data.Result == constant.DECISION_PASS {
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) {
				if filtering.IsWoContractBiro != nil {
					isWoContractBiro, err = utils.GetFloat(filtering.IsWoContractBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat IsWoContractBiro Pefindo Error")
						return
					}
				}

				if filtering.IsWoWithCollateralBiro != nil {
					isWoWithCollateralBiro, err = utils.GetFloat(filtering.IsWoWithCollateralBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat IsWoWithCollateralBiro Pefindo Error")
						return
					}
				}

				if filtering.TotalBakiDebetNonCollateralBiro != nil {
					totalBakiDebetNonAgunan, err = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat TotalBakiDebetNonCollateralBiro Pefindo Error")
						return
					}
				}

				if isWoContractBiro > 0 {
					if isWoWithCollateralBiro > 0 {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
							Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_SAMA, constant.ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
								Reason:         constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI,
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
								Reason:         constant.NAMA_SAMA_BAKI_DEBET_SESUAI,
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
							Reason:         constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
							Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_SAMA, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_PASS,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					}
				}
			} else {
				if filtering.IsWoContractBiro != nil {
					isWoContractBiro, err = utils.GetFloat(filtering.IsWoContractBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat IsWoContractBiro Pefindo Error")
						return
					}
				}

				if filtering.IsWoWithCollateralBiro != nil {
					isWoWithCollateralBiro, err = utils.GetFloat(filtering.IsWoWithCollateralBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat IsWoWithCollateralBiro Pefindo Error")
						return
					}
				}

				if filtering.TotalBakiDebetNonCollateralBiro != nil {
					totalBakiDebetNonAgunan, err = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)
					if err != nil {
						err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat TotalBakiDebetNonCollateralBiro Pefindo Error")
						return
					}
				}

				if isWoContractBiro > 0 {
					if isWoWithCollateralBiro > 0 {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
							Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_BEDA, constant.ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
								Reason:         constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI,
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
								Reason:         constant.NAMA_BEDA_BAKI_DEBET_SESUAI,
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
							Reason:         constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
							Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_BEDA, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_PASS,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					}
				}
			}
		}

	} else {
		data = response.UsecaseApi{
			Code:           constant.CODE_PEFINDO_NO,
			Reason:         constant.REASON_PEFINDO_NOTFOUND,
			Result:         constant.DECISION_PASS,
			SourceDecision: constant.SOURCE_DECISION_BIRO,
		}
	}

	if filtering.Reason != nil {
		data.Reason = filtering.Reason.(string)
	}

	return
}
