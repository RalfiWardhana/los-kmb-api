package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

func (u usecase) Prescreening(ctx context.Context, req request.Metrics, filtering entity.FilteringKMB, accessToken string) (trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, trxDetail entity.TrxDetail, err error) {

	var (
		ntfOther             response.NTFOther
		ntfDetails           response.NTFDetails
		ntfOtherAmount       float64
		ntfOtherAmountSpouse float64
		ntfConfinsAmount     response.OutstandingConfins
		confins              response.OutstandingConfins
		topup                response.IntegratorAgreementChassisNumber
		ntfAmount            float64
		customerID           string
	)

	if filtering.CustomerID != nil {
		customerID = filtering.CustomerID.(string)
	}

	// cek ntf gantung all lob
	var customerData []request.CustomerData
	customerData = append(customerData, request.CustomerData{
		TransactionID: req.Transaction.ProspectID,
		IDNumber:      req.CustomerPersonal.IDNumber,
		LegalName:     req.CustomerPersonal.LegalName,
		BirthDate:     req.CustomerPersonal.BirthDate,
		MotherName:    req.CustomerPersonal.SurgateMotherName,
		CustomerID:    customerID,
	})

	if req.CustomerPersonal.MaritalStatus == constant.MARRIED && req.CustomerSpouse != nil {
		spouse := *req.CustomerSpouse
		customerData = append(customerData, request.CustomerData{
			TransactionID: req.Transaction.ProspectID,
			IDNumber:      spouse.IDNumber,
			LegalName:     spouse.LegalName,
			BirthDate:     spouse.BirthDate,
			MotherName:    spouse.SurgateMotherName,
		})
	}

	header := map[string]string{}

	for i, customer := range customerData {

		jsonCustomer, _ := json.Marshal(customer)
		var ntfLOS *resty.Response

		ntfLOS, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_PENDING_URL"), jsonCustomer, header, constant.METHOD_POST, true, 3, 60, req.Transaction.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error")
			return
		}

		if ntfLOS.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(ntfLOS.Body(), "data").ToString()), &ntfOther)

		if i == 0 {
			// customer
			ntfOtherAmount = ntfOther.NTFAmountKmbOff + ntfOther.NTFAmountWgOff + ntfOther.NTFAmountKmobOff + ntfOther.NTFAmountUC + ntfOther.NTFAmountWgOnl + ntfOther.NTFAmountNewKmb
			ntfDetails.Customer = ntfOther
			confins.TotalOutstanding = ntfOther.TotalOutstanding

		} else if i == 1 {
			// spouse
			ntfOtherAmountSpouse = ntfOther.NTFAmountKmbOff + ntfOther.NTFAmountWgOff + ntfOther.NTFAmountKmobOff + ntfOther.NTFAmountUC + ntfOther.NTFAmountWgOnl + ntfOther.NTFAmountNewKmb
			ntfDetails.Spouse = ntfOther
		}
	}

	var ntfTopup *resty.Response
	ntfTopup, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+req.Item.NoChassis, nil, header, constant.METHOD_GET, true, 3, 60, req.Transaction.ProspectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error")
		return
	}

	if ntfTopup.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error")
		return
	}

	err = json.Unmarshal([]byte(jsoniter.Get(ntfTopup.Body(), "data").ToString()), &topup)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error")
		return
	}

	ntfConfinsAmount.TotalOutstanding = confins.TotalOutstanding - topup.OutstandingPrincipal

	ntfAmount = req.Apk.NTF + (ntfOtherAmount + ntfOtherAmountSpouse) + ntfConfinsAmount.TotalOutstanding

	// auto approve <= 20jt
	if ntfAmount <= constant.LIMIT_PRESCREENING {
		trxPrescreening = entity.TrxPrescreening{
			ProspectID: req.Transaction.ProspectID,
			Decision:   constant.DB_DECISION_APR,
			Reason:     "Dokumen Sesuai",
			CreatedBy:  constant.SYSTEM_CREATED,
			DecisionBy: constant.SYSTEM_CREATED,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_PROCESS,
			Decision:       constant.DB_DECISION_PASS,
			SourceDecision: constant.PRESCREENING,
			RuleCode:       constant.CODE_PASS_PRESCREENING,
			Info:           trxPrescreening.Reason,
			NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			CreatedBy:      constant.SYSTEM_CREATED,
		}
	} else {
		trxDetail = entity.TrxDetail{
			ProspectID:     req.Transaction.ProspectID,
			StatusProcess:  constant.STATUS_ONPROCESS,
			Activity:       constant.ACTIVITY_UNPROCESS,
			Decision:       constant.DB_DECISION_CREDIT_PROCESS,
			SourceDecision: constant.PRESCREENING,
			CreatedBy:      constant.SYSTEM_CREATED,
		}
	}

	sntfDetails, _ := json.Marshal(ntfDetails)

	trxFMF = response.TrxFMF{
		NTFAkumulasi:         ntfAmount,
		NTFOtherAmount:       ntfOtherAmount,
		NTFOtherAmountSpouse: ntfOtherAmountSpouse,
		NTFOtherAmountDetail: string(utils.SafeEncoding(sntfDetails)),
		NTFConfinsAmount:     ntfConfinsAmount.TotalOutstanding,
		NTFConfins:           confins.TotalOutstanding,
		NTFTopup:             topup.OutstandingPrincipal,
	}

	return trxPrescreening, trxFMF, trxDetail, err
}
