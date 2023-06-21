package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	entity "los-kmb-api/models/dupcheck"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/query"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

func (u usecase) DsrCheck(ctx context.Context, prospectID, engineNo string, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, newDupcheck entity.NewDupcheck, accessToken string) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error) {

	var (
		dsr                                        float64
		dataKmob, dataWgOff, dataKmbOff, dataWgOnl entity.ScanInstallmentAmount
		instOther                                  response.InstallmentOther
		dsrDetails                                 response.DsrDetails
		customerStatus                             string
	)

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	config, err := u.repository.GetDupcheckConfig()

	if err != nil {
		err = errors.New("upstream_service_error - Error Get Parameterize Config")
		return
	}

	var configValue response.DupcheckConfig

	json.Unmarshal([]byte(config.Value), &configValue)

	reasonMaxDsr := int(configValue.Data.MaxDsr)

	konsumen := customerData[0]

	customerStatus = konsumen.StatusKonsumen

	for i, customer := range customerData {

		encrypted, _ := u.repository.GetEncryptedValue(customer.IDNumber, customer.LegalName, customer.MotherName)

		kmobWG := query.ScanInstallmentAmountKmobOFF(encrypted.IDNumber, encrypted.LegalName, customer.BirthDate, encrypted.SurgateMotherName)

		dataKmob, err = u.repository.ScanKmobOff(kmobWG)

		if err != nil {
			err = errors.New("upstream_service_error - Scan DB KMOB")
			return
		}

		instOther.InstallmentAmountKmobOff = dataKmob.InstallmentAmount

		wgOnl := query.ScanInstallmentAmountWgONL(encrypted.IDNumber, encrypted.LegalName, customer.BirthDate, encrypted.SurgateMotherName)

		dataWgOnl, err = u.repository.ScanWgOnl(wgOnl)

		if err != nil {
			err = errors.New("upstream_service_error - Scan DB WG ONL")
			return
		}

		instOther.InstallmentAmountWgOnl = dataWgOnl.InstallmentAmount

		wgOFF := query.ScanInstallmentAmountWgOff(customer.IDNumber, encrypted.LegalName, customer.BirthDate, encrypted.SurgateMotherName)

		dataWgOff, err = u.repository.ScanWgOff(wgOFF)

		if err != nil {
			err = errors.New("upstream_service_error - Scan DB WG Offline")
			return
		}

		instOther.InstallmentAmountWgOff = dataWgOff.InstallmentAmount

		kmbOFF := query.ScanInstallmentAmountKmbOff(customer.IDNumber, customer.LegalName, customer.BirthDate, customer.MotherName)

		dataKmbOff, err = u.repository.ScanKmbOff(kmbOFF)

		if err != nil {
			err = errors.New("upstream_service_error - Scan DB KMB Offline")
			return
		}

		instOther.InstallmentAmountKmbOff = dataKmbOff.InstallmentAmount
		if i == 0 {
			installmentOther = dataKmob.InstallmentAmount + dataKmbOff.InstallmentAmount + dataWgOff.InstallmentAmount + dataWgOnl.InstallmentAmount
			dsrDetails.Customer = instOther
		} else if i == 1 {
			installmentOtherSpouse = dataKmob.InstallmentAmount + dataKmbOff.InstallmentAmount + dataWgOff.InstallmentAmount + dataWgOnl.InstallmentAmount
			dsrDetails.Spouse = instOther
		}

	}

	if dsrDetails != (response.DsrDetails{}) {
		result.Details = dsrDetails
	}

	if konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_NEW {

		dsr = ((installmentAmount + (installmentOther + installmentOtherSpouse) + (installmentConfins + installmentConfinsSpouse)) / income) * 100

		data.Dsr = dsr

		dsrBypass, _ := u.repository.GetDSRBypass()

		if dsrBypass.Value == constant.FLAG_ON {
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_DSRLTE35
			data.Reason = fmt.Sprintf("%s %s %d", customerStatus, constant.REASON_DSRLTE, reasonMaxDsr)

			return
		}

		if dsr > configValue.Data.MaxDsr {

			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_DSRGT35
			data.Reason = fmt.Sprintf("%s %s %d", customerStatus, constant.REASON_DSRGT35, reasonMaxDsr)

			_ = mapstructure.Decode(data, &result)

			return

		}

	} else {

		var (
			chassisResp response.SpDupcekChasisNo
			installment float64
			chassis     *resty.Response
		)

		body, _ := json.Marshal(map[string]interface{}{
			"transaction_id": prospectID,
			"engine_no":      engineNo,
			"id_number":      konsumen.IDNumber,
		})

		if newDupcheck.CustomerType == constant.RO_AO_PRIME || newDupcheck.CustomerType == constant.RO_AO_PRIORITY {
			customerStatus = fmt.Sprintf("%s %s", konsumen.StatusKonsumen, newDupcheck.CustomerType)
		}

		if newDupcheck.CustomerType == constant.RO_AO_PRIME && installmentConfins > 0 {

			installment = installmentConfins

		} else if installmentConfins > 0 {

			chassis, err = u.httpclient.EngineAPI(ctx, constant.FILTERING_LOG, os.Getenv("KMOB_CHASSIS_URL"), body, map[string]string{}, constant.METHOD_POST, true, 2, timeOut, prospectID, accessToken)

			if err != nil {
				err = errors.New("upstream_service_timeout - Call Dupcheck Chassis Number")
				return
			}

			if chassis.StatusCode() != 200 {
				err = errors.New("upstream_service_error - Call Dupcheck Chassis Number")
				return
			}

			json.Unmarshal([]byte(jsoniter.Get(chassis.Body(), "data").ToString()), &chassisResp)

			if chassisResp.InstallmentAmount == nil {
				installment = installmentConfins
			} else {
				installmentTopup = chassisResp.InstallmentAmount.(float64)
				installment = installmentConfins - installmentTopup
			}

		}

		dsr = ((installmentAmount + (installment + installmentConfinsSpouse) + (installmentOther + installmentOtherSpouse)) / income) * 100

		data.Dsr = dsr

		dsrBypass, _ := u.repository.GetDSRBypass()

		if dsrBypass.Value == constant.FLAG_ON {
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_DSRLTE35
			data.Reason = fmt.Sprintf("%s %s %d", customerStatus, constant.REASON_DSRLTE, reasonMaxDsr)

			return
		}

		if dsr > configValue.Data.MaxDsr && newDupcheck.CustomerType != constant.RO_AO_PRIME {

			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_DSRGT35
			data.Reason = fmt.Sprintf("%s %s %d", customerStatus, constant.REASON_DSRGT35, reasonMaxDsr)

			_ = mapstructure.Decode(data, &result)

			return

		}

	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_DSRLTE35
	data.Reason = fmt.Sprintf("%s %s %d", customerStatus, constant.REASON_DSRLTE35, reasonMaxDsr)

	_ = mapstructure.Decode(data, &result)

	return
}
