package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/query"
	"los-kmb-api/shared/utils"
	"os"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

func (u usecase) Prescreening(ctx context.Context, req request.Metrics, filtering entity.FilteringKMB, accessToken string) (trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, trxDetail entity.TrxDetail, err error) {

	var (
		dataKmob, dataWgOff, dataKmbOff, dataWgOnl entity.ScanInstallmentAmount
		instOther                                  response.InstallmentOther
		ntfOther                                   response.NTFOther
		dsrDetails                                 response.DsrDetails
		installmentOther                           float64
		installmentOtherSpouse                     float64
		ntfDetails                                 response.NTFDetails
		ntfOtherAmount                             float64
		ntfOtherAmountSpouse                       float64
		ntfConfinsAmount                           response.OutstandingConfins
		confins                                    response.OutstandingConfins
		topup                                      response.OutstandingConfins
		ntfAmount                                  float64
	)

	// cek ntf gantung all lob
	var customerData []request.CustomerData
	customerData = append(customerData, request.CustomerData{
		IDNumber:   req.CustomerPersonal.IDNumber,
		LegalName:  req.CustomerPersonal.LegalName,
		BirthDate:  req.CustomerPersonal.BirthDate,
		MotherName: req.CustomerPersonal.SurgateMotherName,
	})

	if req.CustomerPersonal.MaritalStatus == constant.MARRIED && req.CustomerSpouse != nil {
		spouse := *req.CustomerSpouse
		customerData = append(customerData, request.CustomerData{
			IDNumber:   spouse.IDNumber,
			LegalName:  spouse.LegalName,
			BirthDate:  spouse.BirthDate,
			MotherName: spouse.SurgateMotherName,
		})
	}

	for i, customer := range customerData {

		kmobOff := query.ScanInstallmentAmountKmobOFF(customer.IDNumber, customer.LegalName, customer.BirthDate, customer.MotherName)

		dataKmob, err = u.repository.ScanKmobOff(kmobOff)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Scan DB KMOB")
			return
		}

		instOther.InstallmentAmountKmobOff = dataKmob.InstallmentAmount
		ntfOther.NTFAmountKmobOff = dataKmob.NTF

		wgOnl := query.ScanInstallmentAmountWgONL(customer.IDNumber, customer.LegalName, customer.BirthDate, customer.MotherName)

		dataWgOnl, err = u.repository.ScanWgOnl(wgOnl)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Scan DB WG ONL")
			return
		}

		instOther.InstallmentAmountWgOnl = dataWgOnl.InstallmentAmount
		ntfOther.NTFAmountWgOnl = dataWgOnl.NTF

		wgOFF := query.ScanInstallmentAmountWgOff(customer.IDNumber, customer.LegalName, customer.BirthDate, customer.MotherName)

		dataWgOff, err = u.repository.ScanWgOff(wgOFF)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Scan DB WG Offline")
			return
		}

		instOther.InstallmentAmountWgOff = dataWgOff.InstallmentAmount
		ntfOther.NTFAmountWgOff = dataWgOff.NTF

		kmbOFF := query.ScanInstallmentAmountKmbOff(customer.IDNumber, customer.LegalName, customer.BirthDate, customer.MotherName)

		dataKmbOff, err = u.repository.ScanKmbOff(kmbOFF)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Scan DB KMB Offline")
			return
		}

		instOther.InstallmentAmountKmbOff = dataKmbOff.InstallmentAmount
		ntfOther.NTFAmountKmbOff = dataKmbOff.NTF

		if i == 0 {
			installmentOther = dataKmob.InstallmentAmount + dataKmbOff.InstallmentAmount + dataWgOff.InstallmentAmount + dataWgOnl.InstallmentAmount
			dsrDetails.Customer = instOther

			ntfOtherAmount = dataKmob.NTF + dataKmbOff.NTF + dataWgOff.NTF + dataWgOnl.NTF
			ntfDetails.Customer = ntfOther
		} else if i == 1 {
			installmentOtherSpouse = dataKmob.InstallmentAmount + dataKmbOff.InstallmentAmount + dataWgOff.InstallmentAmount + dataWgOnl.InstallmentAmount
			dsrDetails.Spouse = instOther

			ntfOtherAmountSpouse = dataKmob.NTF + dataKmbOff.NTF + dataWgOff.NTF + dataWgOnl.NTF
			ntfDetails.Spouse = ntfOther
		}
	}

	if filtering.CustomerID != nil {

		reqNTFConfins, _ := json.Marshal(map[string]interface{}{
			"prospect_id": req.Transaction.ProspectID,
			"customer_id": filtering.CustomerID,
		})

		header := map[string]string{
			"Authorization": accessToken,
		}

		var ntfResp *resty.Response
		ntfResp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_CONFINS_URL"), reqNTFConfins, header, constant.METHOD_POST, true, 3, 60, req.Transaction.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + "Call NTF Confins API")
		}

		if ntfResp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + "Call NTF Confins API")
		}

		json.Unmarshal([]byte(jsoniter.Get(ntfResp.Body(), "data").ToString()), &confins)

		reqTopup, _ := json.Marshal(map[string]interface{}{
			"customer_id": filtering.CustomerID,
			"engine_no":   req.Item.NoEngine,
		})

		var ntfTopup *resty.Response
		ntfResp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_TOPUP_URL"), reqTopup, header, constant.METHOD_POST, true, 3, 60, req.Transaction.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + "Call NTF TOPUP API")
			return
		}

		if ntfResp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + "Call NTF TOPUP API")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(ntfTopup.Body(), "data").ToString()), &topup)

		ntfConfinsAmount.TotalOutstanding = confins.TotalOutstanding - topup.TotalOutstanding

	}

	ntfAmount = req.Apk.NTF + (ntfOtherAmount + ntfOtherAmountSpouse) + ntfConfinsAmount.TotalOutstanding

	// auto approve <= 20jt
	if ntfAmount <= constant.LIMIT_PRESCREENING {
		trxPrescreening = entity.TrxPrescreening{
			ProspectID: req.Transaction.ProspectID,
			Decision:   constant.DB_DECISION_APR,
			Reason:     "Dokumen Sesuai",
			CreatedBy:  constant.SYSTEM_CREATED,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			SourceDecision: constant.SOURCE_DECISION_PRESCREENING,
			NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			CreatedBy:      constant.SYSTEM_CREATED,
		}
	} else {
		trxDetail = entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_UNPROCESS,
			Decision:       constant.DB_DECISION_CREDIT_PROCESS,
			SourceDecision: constant.SOURCE_DECISION_PRESCREENING,
			NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			CreatedBy:      constant.SYSTEM_CREATED,
		}
	}

	sntfDetails, _ := json.Marshal(ntfDetails)
	sdsrDetails, _ := json.Marshal(dsrDetails)

	trxFMF = response.TrxFMF{
		NTFAkumulasi:           ntfAmount,
		NTFOtherAmount:         ntfOtherAmount,
		NTFOtherAmountSpouse:   ntfOtherAmountSpouse,
		NTFOtherAmountDetail:   string(utils.SafeEncoding(sntfDetails)),
		NTFConfinsAmount:       ntfConfinsAmount.TotalOutstanding,
		NTFConfins:             confins.TotalOutstanding,
		NTFTopup:               topup.TotalOutstanding,
		InstallmentOther:       installmentOther,
		InstallmentOtherSpouse: installmentOtherSpouse,
		InstallmentOtherDetail: string(utils.SafeEncoding(sdsrDetails)),
	}

	return trxPrescreening, trxFMF, trxDetail, err
}
