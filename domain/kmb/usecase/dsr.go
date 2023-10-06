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
	"os"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

func (u usecase) DsrCheck(ctx context.Context, prospectID, chassisNumber string, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, accessToken string) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error) {

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

		installmentLOS, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("INSTALLMENT_PENDING_URL"), jsonCustomer, header, constant.METHOD_POST, true, 3, 60, prospectID, accessToken)

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
			dsrDetails.Customer = instOther
		} else if i == 1 {
			installmentOtherSpouse = instOther.InstallmentAmountKmbOff + instOther.InstallmentAmountKmobOff + instOther.InstallmentAmountNewKmb + instOther.InstallmentAmountUC + instOther.InstallmentAmountWgOff + instOther.InstallmentAmountWgOnl
			dsrDetails.Spouse = instOther
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

			_ = mapstructure.Decode(data, &result)

			return

		}

	} else {

		var (
			chassisResp entity.SpDupcekChasisNo
			installment float64
		)

		if installmentConfins > 0 {

			chassisResp, err = u.repository.GetInstallmentAmountChassisNumber(chassisNumber)

			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Chassis Number")
				return
			}
			installment = installmentConfins - chassisResp.InstallmentAmount

		}

		dsr = ((installmentAmount + (installment + installmentConfinsSpouse) + (installmentOther + installmentOtherSpouse)) / income) * 100

		data.Dsr = dsr

		if konsumen.StatusKonsumen != constant.STATUS_KONSUMEN_RO && konsumen.CustomerSegment != constant.RO_AO_PRIME {
			if dsr > configValue.Data.MaxDsr {

				data.Result = constant.DECISION_REJECT
				data.Code = constant.CODE_DSRGT35
				data.Reason = fmt.Sprintf("%s %s %d", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)

				_ = mapstructure.Decode(data, &result)

				return
			}
		}

	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_DSRLTE35
	data.Reason = fmt.Sprintf("%s %s %d", reasonCustomerStatus, constant.REASON_DSRLTE35, reasonMaxDsr)

	_ = mapstructure.Decode(data, &result)

	return
}
