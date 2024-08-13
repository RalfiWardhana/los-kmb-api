package usecase

import (
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
			category                string
		)

		maxOverdueDays, _ = utils.GetFloat(filtering.MaxOverdueKORules)
		maxOverdueLast12Months, _ = utils.GetFloat(filtering.MaxOverdueLast12MonthsKORules)
		category = getReasonCategoryRoman(filtering.Category)

		if maxOverdueLast12Months > constant.PBK_OVD_LAST_12 {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_OVD12GT60,
				Reason:         constant.REJECT_REASON_OVD_PEFINDO,
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) && category != "(III)" {
				data.Result = constant.DECISION_PASS
			}
		} else if maxOverdueDays > constant.PBK_OVD_CURRENT {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_CURRENT_GT30,
				Reason:         constant.REJECT_REASON_OVD_PEFINDO,
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) && category != "(III)" {
				data.Result = constant.DECISION_PASS
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

				isWoContractBiro, _ = utils.GetFloat(filtering.IsWoContractBiro)
				isWoWithCollateralBiro, _ = utils.GetFloat(filtering.IsWoWithCollateralBiro)
				totalBakiDebetNonAgunan, _ = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)

				if isWoContractBiro > 0 {
					if isWoWithCollateralBiro > 0 {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
							Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
								Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category),
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
								Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, category),
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
							Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
							Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_PASS,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					}
				}
			} else {
				data = response.UsecaseApi{
					Code:           constant.CODE_PEFINDO_BPKB_BEDA,
					Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, category, data.Reason),
					Result:         constant.DECISION_REJECT,
					SourceDecision: constant.SOURCE_DECISION_BIRO,
				}
			}
		} else if data.Result == constant.DECISION_PASS {
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) {

				isWoContractBiro, _ = utils.GetFloat(filtering.IsWoContractBiro)
				isWoWithCollateralBiro, _ = utils.GetFloat(filtering.IsWoWithCollateralBiro)
				totalBakiDebetNonAgunan, _ = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)

				if isWoContractBiro > 0 {
					if isWoWithCollateralBiro > 0 {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
							Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
								Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category),
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
								Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, category),
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
							Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
							Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_PASS,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					}
				}
			} else {

				isWoContractBiro, _ = utils.GetFloat(filtering.IsWoContractBiro)
				isWoWithCollateralBiro, _ = utils.GetFloat(filtering.IsWoWithCollateralBiro)
				totalBakiDebetNonAgunan, _ = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)

				if isWoContractBiro > 0 {
					if isWoWithCollateralBiro > 0 {
						data = response.UsecaseApi{
							Code:           constant.NAMA_BEDA_WO_AGUNAN_REJECT_CODE,
							Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, category, constant.ADA_FASILITAS_WO_AGUNAN),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_GT20J,
								Reason:         fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category),
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_LTE20J,
								Reason:         fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, category),
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_GT20J,
							Reason:         fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category),
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						data = response.UsecaseApi{
							Code:           constant.NAMA_BEDA_NO_FACILITY_WO_CODE,
							Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, category, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
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

	return
}

// function to map reason category values to Roman numerals
func getReasonCategoryRoman(category interface{}) (str string) {
	num, _ := utils.GetFloat(category)
	switch num {
	case 1:
		str = "(I)"
	case 2:
		str = "(II)"
	case 3:
		str = "(III)"
	}
	return str
}
