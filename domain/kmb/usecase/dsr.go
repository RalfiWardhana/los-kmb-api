package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

func (u usecase) DsrCheck(ctx context.Context, req request.DupcheckApi, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, accessToken string) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error) {

	var (
		dsr                  float64
		instOther            response.InstallmentOther
		dsrDetails           response.DsrDetails
		reasonCustomerStatus string
	)

	config, err := u.repository.GetDupcheckConfig()

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Config Error")
		return
	}

	var configValue response.DupcheckConfig

	json.Unmarshal([]byte(config.Value), &configValue)

	reasonMaxDsr := int(configValue.Data.MaxDsr)

	konsumen := customerData[0]

	if konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_RO || konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
		reasonCustomerStatus = konsumen.StatusKonsumen + " " + konsumen.CustomerSegment
	} else {
		reasonCustomerStatus = konsumen.StatusKonsumen
	}

	header := map[string]string{}

	for i, customer := range customerData {

		jsonCustomer, _ := json.Marshal(customer)
		var installmentLOS *resty.Response

		installmentLOS, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("INSTALLMENT_PENDING_URL"), jsonCustomer, header, constant.METHOD_POST, true, 3, 60, req.ProspectID, accessToken)

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
			data.Reason = fmt.Sprintf("%s %s %d", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
			data.SourceDecision = constant.SOURCE_DECISION_DSR

			_ = mapstructure.Decode(data, &result)
			return
		}

	} else {

		var (
			installment float64
		)

		if installmentConfins > 0 {

			var hitChassisNumber *resty.Response

			hitChassisNumber, err = u.httpclient.EngineAPI(ctx, constant.DUPCHECK_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+req.RangkaNo, nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, req.ProspectID, accessToken)

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

			if responseAgreementChassisNumber != (response.AgreementChassisNumber{}) {

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

				if minimumPencairan < configValue.Data.MinimumPencairanROTopUp {

					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_TOPUP_MENUNGGAK
					data.Reason = fmt.Sprintf("%s %s", reasonCustomerStatus, constant.REASON_TOPUP_MENUNGGAK)

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
				data.Reason = fmt.Sprintf("%s %s %d", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
				data.SourceDecision = constant.SOURCE_DECISION_DSR

				_ = mapstructure.Decode(data, &result)
				return
			}
		} else if konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
			if konsumen.CustomerSegment == constant.RO_AO_PRIME && installmentTopup > 0 {
				// go next
				reasonCustomerStatus = reasonCustomerStatus + " " + constant.TOP_UP

				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_DSRLTE35
				data.SourceDecision = constant.SOURCE_DECISION_DSR
				data.Reason = fmt.Sprintf("%s", reasonCustomerStatus)

				_ = mapstructure.Decode(data, &result)
				return
			} else if dsr > configValue.Data.MaxDsr {
				data.Result = constant.DECISION_REJECT
				data.Code = constant.CODE_DSRGT35
				data.Reason = fmt.Sprintf("%s %s %d", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
				data.SourceDecision = constant.SOURCE_DECISION_DSR

				_ = mapstructure.Decode(data, &result)
				return
			}
		} else if dsr > configValue.Data.MaxDsr {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_DSRGT35
			data.Reason = fmt.Sprintf("%s %s %d", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
			data.SourceDecision = constant.SOURCE_DECISION_DSR

			_ = mapstructure.Decode(data, &result)
			return
		}

	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_DSRLTE35
	data.Reason = fmt.Sprintf("%s %s %d", reasonCustomerStatus, constant.REASON_DSRLTE35, reasonMaxDsr)
	data.SourceDecision = constant.SOURCE_DECISION_DSR

	_ = mapstructure.Decode(data, &result)
	return
}
