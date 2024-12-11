package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

func (u usecase) DsrCheck(ctx context.Context, req request.DupcheckApi, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, accessToken string, configValue response.DupcheckConfig) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error) {

	var (
		dsr                  float64
		instOther            response.InstallmentOther
		dsrDetails           response.DsrDetails
		reasonCustomerStatus string
	)

	reasonMaxDsr := "Threshold"

	konsumen := customerData[0]

	if konsumen.CustomerSegment == constant.RO_AO_PRIME || konsumen.CustomerSegment == constant.RO_AO_PRIORITY {
		reasonCustomerStatus = konsumen.StatusKonsumen + " " + konsumen.CustomerSegment
	} else {
		reasonCustomerStatus = konsumen.StatusKonsumen
	}

	header := map[string]string{}

	for i, customer := range customerData {

		jsonCustomer, _ := json.Marshal(customer)
		var installmentLOS *resty.Response

		installmentLOS, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("INSTALLMENT_PENDING_URL"), jsonCustomer, header, constant.METHOD_POST, true, 2, 60, req.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Installment Pending API Error")
			return
		}

		if installmentLOS.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Installment Pending API Error")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(installmentLOS.Body(), "data").ToString()), &instOther)

		if i == 0 {
			installmentOther = instOther.InstallmentAmountKmbOff + instOther.InstallmentAmountKmobOff + instOther.InstallmentAmountNewKmb + instOther.InstallmentAmountUC + instOther.InstallmentAmountWgOff + instOther.InstallmentAmountWgOnl
			if instOther != (response.InstallmentOther{}) {
				dsrDetails.Customer = instOther
			}
		} else if i == 1 {
			installmentOtherSpouse = instOther.InstallmentAmountKmbOff + instOther.InstallmentAmountKmobOff + instOther.InstallmentAmountNewKmb + instOther.InstallmentAmountUC + instOther.InstallmentAmountWgOff + instOther.InstallmentAmountWgOnl
			if instOther != (response.InstallmentOther{}) {
				dsrDetails.Spouse = instOther
			}
		}

	}

	if dsrDetails != (response.DsrDetails{}) {
		result.Details = dsrDetails
	}

	if konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_NEW {
		dsr = ((installmentAmount + (installmentOther + installmentOtherSpouse) + (installmentConfins + installmentConfinsSpouse)) / income) * 100
		data.Dsr = dsr

		if dsr > configValue.Data.MaxDsr {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_DSRGT35
			data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
			data.SourceDecision = constant.SOURCE_DECISION_DSR

			_ = mapstructure.Decode(data, &result)
			return
		}

	} else {

		var installment = installmentConfins

		if installmentConfins > 0 {

			var hitChassisNumber *resty.Response

			hitChassisNumber, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+req.RangkaNo, nil, map[string]string{}, constant.METHOD_GET, true, 2, 60, req.ProspectID, accessToken)

			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - DsrCheck Call Get Agreement of Chassis Number Timeout")
				return
			}

			if hitChassisNumber.StatusCode() != 200 {
				err = errors.New(constant.ERROR_UPSTREAM + " - DsrCheck Call Get Agreement of Chassis Number Error")
				return
			}

			var responseAgreementChassisNumber response.AgreementChassisNumber
			err = json.Unmarshal([]byte(jsoniter.Get(hitChassisNumber.Body(), "data").ToString()), &responseAgreementChassisNumber)

			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - DsrCheck Unmarshal Get Agreement of Chassis Number Error")
				return
			}

			if responseAgreementChassisNumber != (response.AgreementChassisNumber{}) && responseAgreementChassisNumber.InstallmentAmount > 0 {
				reasonCustomerStatus = reasonCustomerStatus + " " + constant.TOP_UP

				installmentTopup = responseAgreementChassisNumber.InstallmentAmount
				installment = installmentConfins - installmentTopup

				var pencairan float64
				pencairan = req.OTRPrice - req.DPAmount
				if req.Dealer == constant.DEALER_PSA && req.AdminFee != nil {
					pencairan -= *req.AdminFee
				}

				if pencairan <= 0 {
					err = errors.New(constant.ERROR_UPSTREAM + " - Perhitungan OTR - DP harus lebih dari 0")
					return
				}

				totalOutstanding := responseAgreementChassisNumber.OutstandingPrincipal + responseAgreementChassisNumber.OutstandingInterest + responseAgreementChassisNumber.LcInstallment
				minimumPencairan := ((pencairan - totalOutstanding) / pencairan) * 100

				dsrDetails.DetailTopUP = response.DetailTopUP{
					Pencairan:              pencairan,
					AgreementChassisNumber: responseAgreementChassisNumber,
					MinimumPencairan:       minimumPencairan,
					TotalOutstanding:       totalOutstanding,
				}

				result.Details = dsrDetails

				var configMinimumPencairanROTopUp float64

				if konsumen.CustomerSegment == constant.RO_AO_PRIME {
					configMinimumPencairanROTopUp = configValue.Data.MinimumPencairanROTopUp.Prime
				} else if konsumen.CustomerSegment == constant.RO_AO_PRIORITY {
					configMinimumPencairanROTopUp = configValue.Data.MinimumPencairanROTopUp.Priority
				} else {
					configMinimumPencairanROTopUp = configValue.Data.MinimumPencairanROTopUp.Regular
				}

				if minimumPencairan < configMinimumPencairanROTopUp {

					dsr = ((installmentAmount + (installment + installmentConfinsSpouse) + (installmentOther + installmentOtherSpouse)) / income) * 100
					data.Dsr = dsr
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_PENCAIRAN_TOPUP
					data.Reason = fmt.Sprintf("%s %s", reasonCustomerStatus, constant.REASON_PENCAIRAN_TOPUP)

					// set sebagai dupcheck
					data.SourceDecision = constant.SOURCE_DECISION_DUPCHECK

					_ = mapstructure.Decode(data, &result)

					return
				}

			}
		}

		dsr = ((installmentAmount + (installment + installmentConfinsSpouse) + (installmentOther + installmentOtherSpouse)) / income) * 100

		data.Dsr = dsr

		if konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_RO {
			if konsumen.CustomerSegment == constant.RO_AO_PRIME {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_DSRLTE35
				data.SourceDecision = constant.SOURCE_DECISION_DSR
				data.Reason = fmt.Sprintf("%s", reasonCustomerStatus)

				_ = mapstructure.Decode(data, &result)
				return
			} else if dsr > configValue.Data.MaxDsr {
				data.Result = constant.DECISION_REJECT
				data.Code = constant.CODE_DSRGT35
				data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
				data.SourceDecision = constant.SOURCE_DECISION_DSR

				_ = mapstructure.Decode(data, &result)
				return
			}
		} else if konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
			if konsumen.CustomerSegment == constant.RO_AO_PRIME && installmentTopup > 0 {
				// go next
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_DSRLTE35
				data.SourceDecision = constant.SOURCE_DECISION_DSR
				data.Reason = fmt.Sprintf("%s", reasonCustomerStatus)

				_ = mapstructure.Decode(data, &result)
				return
			} else if dsr > configValue.Data.MaxDsr {
				data.Result = constant.DECISION_REJECT
				data.Code = constant.CODE_DSRGT35
				data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
				data.SourceDecision = constant.SOURCE_DECISION_DSR

				_ = mapstructure.Decode(data, &result)
				return
			}
		}

	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_DSRLTE35
	data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRLTE35, reasonMaxDsr)
	data.SourceDecision = constant.SOURCE_DECISION_DSR

	_ = mapstructure.Decode(data, &result)
	return
}

func (u usecase) TotalDsrFmfPbk(ctx context.Context, totalIncome, newInstallment, totalInstallmentPBK float64, prospectID, customerSegment, accessToken string, SpDupcheckMap response.SpDupcheckMap, configValue response.DupcheckConfig, filtering entity.FilteringKMB, NTF float64) (data response.UsecaseApi, trxFMF response.TrxFMF, err error) {

	var (
		RrdDateString           string
		CreatedAtString         string
		RrdDate                 time.Time
		CreatedAt               time.Time
		MonthsOfExpiredContract int
		OverrideFlowLikeRegular bool
		expiredContractConfig   entity.AppConfig
	)

	dsrPBK := totalInstallmentPBK / totalIncome * 100

	totalDSR := SpDupcheckMap.Dsr + dsrPBK

	trxFMF = response.TrxFMF{
		DSRPBK:   dsrPBK,
		TotalDSR: totalDSR,
	}

	reasonMaxDsr := "Threshold"

	var reasonCustomerStatus string
	if customerSegment == constant.RO_AO_PRIME || customerSegment == constant.RO_AO_PRIORITY {
		reasonCustomerStatus = SpDupcheckMap.StatusKonsumen + " " + customerSegment
	} else {
		reasonCustomerStatus = SpDupcheckMap.StatusKonsumen
	}

	if SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
		if SpDupcheckMap.InstallmentTopup > 0 {
			reasonCustomerStatus = reasonCustomerStatus + " " + constant.TOP_UP
		}
	}

	// INTERCEPT PERBAIKAN FLOW RO PRIME/PRIORITY (NON-TOPUP) | CHECK EXPIRED_CONTRACT
	if (customerSegment == constant.RO_AO_PRIME || customerSegment == constant.RO_AO_PRIORITY) && (SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_RO && SpDupcheckMap.InstallmentTopup <= 0) {
		if SpDupcheckMap.RRDDate == nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty")
			return
		}

		var RrdDateTime time.Time
		if rrdDateStr, ok := SpDupcheckMap.RRDDate.(string); ok {
			RrdDateTime, err = time.Parse(time.RFC3339, rrdDateStr)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Invalid RrdDate format")
				return
			}
		} else if RrdDateTime, ok = SpDupcheckMap.RRDDate.(time.Time); !ok {
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

	RuleCodeForDSRLTE35 := constant.CODE_TOTAL_DSRLTE35
	if OverrideFlowLikeRegular {
		totalDSR = SpDupcheckMap.Dsr
		trxFMF.TotalDSR = SpDupcheckMap.Dsr

		reasonCustomerStatus = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + reasonCustomerStatus
		RuleCodeForDSRLTE35 = constant.CODE_TOTAL_DSRLTE35_EXP_CONTRACT_6MONTHS
	}

	if !OverrideFlowLikeRegular {
		if (SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_RO || SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_AO) && customerSegment == constant.RO_AO_PRIME {
			var (
				resp                      *resty.Response
				respLatestPaidInstallment response.LatestPaidInstallmentData
				latestInstallmentAmount   float64
			)

			if SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_RO {
				resp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("LASTEST_PAID_INSTALLMENT_URL")+SpDupcheckMap.CustomerID.(string)+"/2", nil, map[string]string{}, constant.METHOD_GET, false, 0, 30, prospectID, accessToken)

				if err != nil {
					err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call LatestPaidInstallmentData Timeout")
					return
				}

				if resp.StatusCode() != 200 {
					err = errors.New(constant.ERROR_UPSTREAM + " - Call LatestPaidInstallmentData Error")
					return
				}

				err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &respLatestPaidInstallment)
				if err != nil {
					err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal LatestPaidInstallmentData Error")
					return
				}

				latestInstallmentAmount = respLatestPaidInstallment.InstallmentAmount // RO PRIME

			} else if SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
				if SpDupcheckMap.InstallmentTopup > 0 {
					latestInstallmentAmount = SpDupcheckMap.InstallmentTopup // RO TOP UP PRIME (secara teknis datanya adalah AO TOP UP PRIME)
				}
			}

			installmentThreshold := latestInstallmentAmount * 1.5

			trxFMF.LatestInstallmentAmount = latestInstallmentAmount
			trxFMF.InstallmentThreshold = installmentThreshold

			if SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_RO {
				if newInstallment < installmentThreshold {
					trxFMF.TotalDSR = SpDupcheckMap.Dsr
					data = response.UsecaseApi{
						Result:         constant.DECISION_PASS,
						Code:           RuleCodeForDSRLTE35,
						Reason:         fmt.Sprintf("%s", reasonCustomerStatus),
						SourceDecision: constant.SOURCE_DECISION_DSR,
					}
					return
				} else {
					totalDSR = SpDupcheckMap.Dsr
					trxFMF.TotalDSR = SpDupcheckMap.Dsr
				}
			} else if SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
				if SpDupcheckMap.InstallmentTopup > 0 {
					if SpDupcheckMap.MaxOverdueDaysforActiveAgreement <= 30 {
						if newInstallment < installmentThreshold {
							trxFMF.TotalDSR = SpDupcheckMap.Dsr
							data = response.UsecaseApi{
								Result:         constant.DECISION_PASS,
								Code:           RuleCodeForDSRLTE35,
								Reason:         fmt.Sprintf("%s", reasonCustomerStatus),
								SourceDecision: constant.SOURCE_DECISION_DSR,
							}
							return
						} else {
							totalDSR = SpDupcheckMap.Dsr
							trxFMF.TotalDSR = SpDupcheckMap.Dsr
						}
					}
				} else {
					if SpDupcheckMap.NumberOfPaidInstallment >= 6 || SpDupcheckMap.AgreementSettledExist {
						totalDSR = SpDupcheckMap.Dsr
						trxFMF.TotalDSR = SpDupcheckMap.Dsr
					}
				}
			}
		} else if (SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_RO || SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_AO) && customerSegment == constant.RO_AO_PRIORITY {
			if SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_RO {
				totalDSR = SpDupcheckMap.Dsr
				trxFMF.TotalDSR = SpDupcheckMap.Dsr
			} else if SpDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
				if SpDupcheckMap.InstallmentTopup > 0 {
					if SpDupcheckMap.MaxOverdueDaysforActiveAgreement <= 30 {
						totalDSR = SpDupcheckMap.Dsr
						trxFMF.TotalDSR = SpDupcheckMap.Dsr
					}
				} else {
					if SpDupcheckMap.NumberOfPaidInstallment >= 6 || SpDupcheckMap.AgreementSettledExist {
						totalDSR = SpDupcheckMap.Dsr
						trxFMF.TotalDSR = SpDupcheckMap.Dsr
					}
				}
			}
		}
	}

	if totalDSR <= configValue.Data.MaxDsr {
		data = response.UsecaseApi{
			Result:         constant.DECISION_PASS,
			Code:           RuleCodeForDSRLTE35,
			Reason:         fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_TOTAL_DSRLTE, reasonMaxDsr),
			SourceDecision: constant.SOURCE_DECISION_DSR,
		}
	} else {
		data = response.UsecaseApi{
			Result:         constant.DECISION_REJECT,
			Code:           constant.CODE_TOTAL_DSRGT35,
			Reason:         fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_TOTAL_DSRGT, reasonMaxDsr),
			SourceDecision: constant.SOURCE_DECISION_DSR,
		}

		var (
			branchDeviasi     entity.MappingBranchDeviasi
			mappingDeviasiDSR entity.MasterMappingDeviasiDSR
			maxDsrDeviasi     float64
		)
		branchDeviasi, err = u.repository.GetBranchDeviasi(filtering.BranchID, SpDupcheckMap.StatusKonsumen, NTF)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetBranchDeviasi on usecase TotalDsrFmfPbk Error")
			return
		}

		if branchDeviasi.BranchID != "" {
			mappingDeviasiDSR, err = u.repository.MasterMappingDeviasiDSR(totalIncome)
			maxDsrDeviasi = mappingDeviasiDSR.DSRThreshold
			if mappingDeviasiDSR.DSRThreshold > 0 && totalDSR <= maxDsrDeviasi {
				data.IsDeviasi = true
			}
		}
	}

	return
}
