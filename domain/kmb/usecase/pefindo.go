package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strings"
	"time"
)

func (u usecase) Pefindo(cbFound bool, bpkbName string, filtering entity.FilteringKMB, spDupcheck response.SpDupcheckMap) (data response.UsecaseApi, err error) {

	var (
		customerSegment         string
		RrdDateString           string
		CreatedAtString         string
		RrdDate                 time.Time
		CreatedAt               time.Time
		MonthsOfExpiredContract int
		OverrideFlowLikeRegular bool
		expiredContractConfig   entity.AppConfig
	)

	if filtering.CustomerSegment != nil {
		customerSegment = filtering.CustomerSegment.(string)
	}

	// INTERCEPT PERBAIKAN FLOW RO PRIME/PRIORITY (NON-TOPUP) | CHECK EXPIRED_CONTRACT
	if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_RO && spDupcheck.InstallmentTopup <= 0 && (customerSegment == constant.RO_AO_PRIME || customerSegment == constant.RO_AO_PRIORITY) {
		if spDupcheck.RRDDate == nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty")
			return
		}

		var RrdDateTime time.Time
		if rrdDateStr, ok := spDupcheck.RRDDate.(string); ok {
			RrdDateTime, err = time.Parse(time.RFC3339, rrdDateStr)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Invalid RrdDate format")
				return
			}
		} else if RrdDateTime, ok = spDupcheck.RRDDate.(time.Time); !ok {
			err = errors.New(constant.ERROR_UPSTREAM + " - RrdDate must be string or time.Time")
			return
		}

		RrdDateString = RrdDateTime.Format(time.RFC3339)
		CreatedAtString = filtering.CreatedAt.Format(time.RFC3339)

		RrdDate, _ = time.Parse(time.RFC3339, RrdDateString)
		CreatedAt, _ = time.Parse(time.RFC3339, CreatedAtString)
		MonthsOfExpiredContract, _ = utils.PreciseMonthsDifference(RrdDate, CreatedAt)

		// Get config expired_contract
		expiredContractConfig, err = u.repository.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Expired Contract Config Error")
			return
		}

		var configValueExpContract response.ExpiredContractConfig
		json.Unmarshal([]byte(expiredContractConfig.Value), &configValueExpContract)

		if configValueExpContract.Data.ExpiredContractCheckEnabled && !(MonthsOfExpiredContract <= configValueExpContract.Data.ExpiredContractMaxMonths) {
			// Jalur mirip seperti customer segment "REGULAR"
			OverrideFlowLikeRegular = true
		}
	}

	if customerSegment == constant.RO_AO_PRIME || customerSegment == constant.RO_AO_PRIORITY {
		// handle priority
		if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_AO && spDupcheck.InstallmentTopup == 0 && customerSegment == constant.RO_AO_PRIORITY {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_PRIME_PRIORITY,
				Reason:         fmt.Sprintf("%s %s - PBK Pass", spDupcheck.StatusKonsumen, customerSegment),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}

			return
		}

		// handle prime
		if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_AO && spDupcheck.InstallmentTopup == 0 && (spDupcheck.NumberOfPaidInstallment >= 6 || spDupcheck.AgreementSettledExist) && customerSegment == constant.RO_AO_PRIME {
			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_PRIME_PRIORITY,
				Reason:         fmt.Sprintf("%s %s - PBK Pass", spDupcheck.StatusKonsumen, customerSegment),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}

			return
		}

		// ADD INTERCEPT CONDITIONAL RELATED TO `PERBAIKAN RO PRIME/PRIORITY (NON-TOPUP)`
		if ((spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_RO || (spDupcheck.InstallmentTopup > 0 && spDupcheck.MaxOverdueDaysforActiveAgreement <= 30)) && !OverrideFlowLikeRegular) || (OverrideFlowLikeRegular && !cbFound) {
			if OverrideFlowLikeRegular {
				data = response.UsecaseApi{
					Code:           constant.CODE_PEFINDO_PRIME_PRIORITY_EXP_CONTRACT_6MONTHS,
					Reason:         constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + fmt.Sprintf("%s %s - PBK Pass", spDupcheck.StatusKonsumen, customerSegment),
					Result:         constant.DECISION_PASS,
					SourceDecision: constant.SOURCE_DECISION_BIRO,
				}
			} else {
				data = response.UsecaseApi{
					Code:           constant.CODE_PEFINDO_PRIME_PRIORITY,
					Reason:         fmt.Sprintf("%s %s - PBK Pass", spDupcheck.StatusKonsumen, customerSegment),
					Result:         constant.DECISION_PASS,
					SourceDecision: constant.SOURCE_DECISION_BIRO,
				}
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

		// cr pbk inquiries
		if filtering.NumberOfInquiriesLast1Month != nil {
			var (
				mappingRiskLevel []entity.MappingRiskLevel
				rejectRiskLevel  bool
				inquiries        int
				reasonRiskLevel  string
			)

			inquiries = *filtering.NumberOfInquiriesLast1Month

			cacheRiskLevel, _ := u.repository.GetCache("GetMappingRiskLevel")
			json.Unmarshal(cacheRiskLevel, &mappingRiskLevel)

			if len(mappingRiskLevel) == 0 {
				mappingRiskLevel, err = u.repository.GetMappingRiskLevel()
				if err != nil {
					err = errors.New(constant.ERROR_UPSTREAM + " - erro GetMappingRiskLevel - " + err.Error())
				}
				cacheRiskLevel, _ = json.Marshal(mappingRiskLevel)
				u.repository.SetCache("GetMappingRiskLevel", cacheRiskLevel)
			}

			for _, v := range mappingRiskLevel {
				if inquiries >= v.InquiryStart && inquiries <= v.InquiryEnd {
					rejectRiskLevel = true
					reasonRiskLevel = fmt.Sprintf("Number of Inquiry PBK %s", v.RiskLevel)
				}
			}
			if rejectRiskLevel {
				if OverrideFlowLikeRegular {
					data.Reason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + reasonRiskLevel
				}

				data.Code = constant.CODE_REJECT_INQUIRIES
				data.StatusKonsumen = spDupcheck.StatusKonsumen
				data.Result = constant.DECISION_REJECT
				data.SourceDecision = constant.SOURCE_DECISION_BIRO
				return
			}
		}
		// enc cr pbk inquiries

		// START - NEW KO RULES | CR 2025-01-10
		if filtering.NewKoRules != nil {
			var newKoRules response.ResultNewKoRules
			if newKoRulesStr, ok := filtering.NewKoRules.(string); ok {
				json.Unmarshal([]byte(newKoRulesStr), &newKoRules)
			}

			if newKoRules.CategoryPBK != "" {
				isRejectNewKoRules := false
				if newKoRules.CategoryPBK == constant.REJECT_LUNAS_DISKON {
					data.Code = constant.CODE_REJECT_LUNAS_DISKON
					data.Reason = constant.REASON_LUNAS_DISKON
					isRejectNewKoRules = true
				} else if newKoRules.CategoryPBK == constant.REJECT_FASILITAS_DIALIHKAN_DIJUAL {
					data.Code = constant.CODE_REJECT_FASILITAS_DIALIHKAN_DIJUAL
					data.Reason = constant.REASON_FASILITAS_DIALIHKAN_DIJUAL
					isRejectNewKoRules = true
				} else if newKoRules.CategoryPBK == constant.REJECT_HAPUS_TAGIH {
					data.Code = constant.CODE_REJECT_HAPUS_TAGIH
					data.Reason = constant.REASON_HAPUS_TAGIH
					isRejectNewKoRules = true
				} else if newKoRules.CategoryPBK == constant.REJECT_REPOSSES {
					data.Code = constant.CODE_REJECT_REPOSSES
					data.Reason = constant.REASON_REPOSSES
					isRejectNewKoRules = true
				} else if newKoRules.CategoryPBK == constant.REJECT_RESTRUCTURE {
					data.Code = constant.CODE_REJECT_RESTRUCTURE
					data.Reason = constant.REASON_RESTRUCTURE
					isRejectNewKoRules = true
				}

				if isRejectNewKoRules {
					if OverrideFlowLikeRegular {
						data.Reason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + data.Reason
					}

					data.StatusKonsumen = spDupcheck.StatusKonsumen
					data.Result = constant.DECISION_REJECT
					data.SourceDecision = constant.SOURCE_DECISION_BIRO

					return
				}
			}
		}
		// END - NEW KO RULES | CR 2025-01-10

		if maxOverdueLast12Months > constant.PBK_OVD_LAST_12 {
			koRulesReason := constant.REJECT_REASON_OVD_PEFINDO
			if OverrideFlowLikeRegular {
				koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
			}

			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_OVD12GT60,
				Reason:         koRulesReason,
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) && category != "(III)" {
				data.Result = constant.DECISION_PASS
			}
		} else if maxOverdueDays > constant.PBK_OVD_CURRENT {
			koRulesReason := constant.REJECT_REASON_OVD_PEFINDO
			if OverrideFlowLikeRegular {
				koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
			}

			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_CURRENT_GT30,
				Reason:         koRulesReason,
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			}
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) && category != "(III)" {
				data.Result = constant.DECISION_PASS
			}
		} else {
			koRulesReason := fmt.Sprintf(constant.REASON_PEFINDO_OVD12LTE60_CURRENT_LTE30, constant.PBK_OVD_CURRENT)
			if OverrideFlowLikeRegular {
				koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
			}

			data = response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_OVD12LTE60_CURRENT_LTE30,
				Reason:         koRulesReason,
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
						koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.ADA_FASILITAS_WO_AGUNAN)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
							Reason:         koRulesReason,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							koRulesReason := fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category)
							if OverrideFlowLikeRegular {
								koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
							}

							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
								Reason:         koRulesReason,
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							koRulesReason := fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, category)
							if OverrideFlowLikeRegular {
								koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
							}

							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
								Reason:         koRulesReason,
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						koRulesReason := fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
							Reason:         koRulesReason,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
							Reason:         koRulesReason,
							Result:         constant.DECISION_PASS,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					}
				}
			} else {

				isWoContractBiro, _ = utils.GetFloat(filtering.IsWoContractBiro)
				isWoWithCollateralBiro, _ = utils.GetFloat(filtering.IsWoWithCollateralBiro)
				totalBakiDebetNonAgunan, _ = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)

				if isWoWithCollateralBiro > 0 {
					koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, category, constant.ADA_FASILITAS_WO_AGUNAN)
					if OverrideFlowLikeRegular {
						koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
					}

					data = response.UsecaseApi{
						Code:           constant.NAMA_BEDA_WO_AGUNAN_REJECT_CODE,
						Reason:         koRulesReason,
						Result:         constant.DECISION_REJECT,
						SourceDecision: constant.SOURCE_DECISION_BIRO,
					}
				} else if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
					koRulesReason := fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category)
					if OverrideFlowLikeRegular {
						koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
					}

					data = response.UsecaseApi{
						Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_GT20J,
						Reason:         koRulesReason,
						Result:         constant.DECISION_REJECT,
						SourceDecision: constant.SOURCE_DECISION_BIRO,
					}
				} else if isWoContractBiro > 0 {
					koRulesReason := fmt.Sprintf("%s %s & %s dan %s", constant.REASON_BPKB_BEDA, category, data.Reason, constant.ADA_FASILITAS_WO_NON_AGUNAN)
					if OverrideFlowLikeRegular {
						koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
					}

					data = response.UsecaseApi{
						Code:           constant.CODE_PEFINDO_BPKB_BEDA,
						Reason:         koRulesReason,
						Result:         constant.DECISION_REJECT,
						SourceDecision: constant.SOURCE_DECISION_BIRO,
					}
				} else {
					koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, category, data.Reason)
					if OverrideFlowLikeRegular {
						koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
					}

					data = response.UsecaseApi{
						Code:           constant.CODE_PEFINDO_BPKB_BEDA,
						Reason:         koRulesReason,
						Result:         constant.DECISION_REJECT,
						SourceDecision: constant.SOURCE_DECISION_BIRO,
					}
				}
			}
		} else if data.Result == constant.DECISION_PASS {
			if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) {

				isWoContractBiro, _ = utils.GetFloat(filtering.IsWoContractBiro)
				isWoWithCollateralBiro, _ = utils.GetFloat(filtering.IsWoWithCollateralBiro)
				totalBakiDebetNonAgunan, _ = utils.GetFloat(filtering.TotalBakiDebetNonCollateralBiro)

				if isWoContractBiro > 0 {
					if isWoWithCollateralBiro > 0 {
						koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.ADA_FASILITAS_WO_AGUNAN)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
							Reason:         koRulesReason,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							koRulesReason := fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category)
							if OverrideFlowLikeRegular {
								koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
							}

							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
								Reason:         koRulesReason,
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							koRulesReason := fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, category)
							if OverrideFlowLikeRegular {
								koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
							}

							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
								Reason:         koRulesReason,
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						koRulesReason := fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
							Reason:         koRulesReason,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, category, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
							Reason:         koRulesReason,
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
						koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, category, constant.ADA_FASILITAS_WO_AGUNAN)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.NAMA_BEDA_WO_AGUNAN_REJECT_CODE,
							Reason:         koRulesReason,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
							koRulesReason := fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category)
							if OverrideFlowLikeRegular {
								koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
							}

							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_GT20J,
								Reason:         koRulesReason,
								Result:         constant.DECISION_REJECT,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						} else {
							koRulesReason := fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, category)
							if OverrideFlowLikeRegular {
								koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
							}

							data = response.UsecaseApi{
								Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_LTE20J,
								Reason:         koRulesReason,
								Result:         constant.DECISION_PASS,
								SourceDecision: constant.SOURCE_DECISION_BIRO,
							}
						}
					}
				} else {
					if totalBakiDebetNonAgunan > constant.BAKI_DEBET {
						koRulesReason := fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, category)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_GT20J,
							Reason:         koRulesReason,
							Result:         constant.DECISION_REJECT,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					} else {
						koRulesReason := fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, category, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN)
						if OverrideFlowLikeRegular {
							koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
						}

						data = response.UsecaseApi{
							Code:           constant.NAMA_BEDA_NO_FACILITY_WO_CODE,
							Reason:         koRulesReason,
							Result:         constant.DECISION_PASS,
							SourceDecision: constant.SOURCE_DECISION_BIRO,
						}
					}
				}
			}
		}

	} else {
		koRulesReason := constant.REASON_PEFINDO_NOTFOUND
		if OverrideFlowLikeRegular {
			koRulesReason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + koRulesReason
		}

		data = response.UsecaseApi{
			Code:           constant.CODE_PEFINDO_NO,
			Reason:         koRulesReason,
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
