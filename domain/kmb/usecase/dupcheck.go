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
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

func (u multiUsecase) Dupcheck(ctx context.Context, req request.DupcheckApi, married bool, accessToken string, configValue response.DupcheckConfig) (mapping response.SpDupcheckMap, status string, data response.UsecaseApi, trxFMF response.TrxFMF, trxDetail []entity.TrxDetail, err error) {

	var (
		customer     []request.SpouseDupcheck
		blackList    response.UsecaseApi
		sp           response.SpDupCekCustomerByID
		dataCustomer []response.SpDupCekCustomerByID
		spMap        response.SpDupcheckMap
		customerType string
	)

	// Check Banned Chassis Number
	bannedChassisNumber, err := u.usecase.CheckBannedChassisNumber(req.RangkaNo)
	if err != nil {
		return
	}

	if bannedChassisNumber.Result == constant.DECISION_REJECT {
		data = bannedChassisNumber
		mapping.Reason = data.Reason
		return
	}

	// Pernah Reject Chassis Number
	rejectChassisNumber, trxBannedChassisNumber, err := u.usecase.CheckRejectChassisNumber(req, configValue)
	if err != nil {
		return
	}

	if rejectChassisNumber.Result == constant.DECISION_REJECT {
		data = rejectChassisNumber
		mapping.Reason = data.Reason

		trxFMF.TrxBannedChassisNumber = trxBannedChassisNumber
		return
	}

	// Check Banned PMK atau DSR
	bannedPMKDSR, err := u.usecase.CheckBannedPMKDSR(req.IDNumber)
	if err != nil {
		return
	}

	if bannedPMKDSR.Result == constant.DECISION_REJECT {
		data = bannedPMKDSR
		mapping.Reason = data.Reason
		return
	}

	// Pernah Reject PMK atau DSR atau NIK
	trxReject, trxBannedPMKDSR, err := u.usecase.CheckRejection(req.IDNumber, req.ProspectID, configValue)
	if err != nil {
		return
	}

	if trxReject.Result == constant.DECISION_REJECT {
		data = trxReject
		mapping.Reason = data.Reason

		trxFMF.TrxBannedPMKDSR = trxBannedPMKDSR
		return
	}

	trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_PASS, RuleCode: trxReject.Code, SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR, Reason: trxReject.Reason, NextStep: constant.SOURCE_DECISION_BLACKLIST})

	prospectID := req.ProspectID
	income := req.MonthlyFixedIncome + req.MonthlyVariableIncome + req.SpouseIncome
	customer = append(customer, request.SpouseDupcheck{IDNumber: req.IDNumber, LegalName: req.LegalName, BirthDate: req.BirthDate, MotherName: req.MotherName})

	if married {
		customer = append(customer, request.SpouseDupcheck{IDNumber: req.Spouse.IDNumber, LegalName: req.Spouse.LegalName, BirthDate: req.Spouse.BirthDate, MotherName: req.Spouse.MotherName})
	}

	// Loop check Blacklist customer and spouse
	for i := 0; i < len(customer); i++ {

		// Call API Dupcheck V2
		sp, err = u.usecase.DupcheckIntegrator(ctx, prospectID, customer[i].IDNumber, customer[i].LegalName, customer[i].BirthDate, customer[i].MotherName, accessToken)

		if err != nil {
			return
		}

		dataCustomer = append(dataCustomer, sp)

		blackList, customerType = u.usecase.BlacklistCheck(i, sp)

		if i == 0 {
			spMap.CustomerType = customerType
		} else if i == 1 {
			spMap.SpouseType = customerType
		}

		if blackList.Result == constant.DECISION_REJECT {
			data = blackList
			mapping.CustomerType = spMap.CustomerType
			mapping.SpouseType = spMap.SpouseType
			mapping.Reason = data.Reason
			return
		}
	}

	trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_PASS, RuleCode: blackList.Code, SourceDecision: constant.SOURCE_DECISION_BLACKLIST, Reason: blackList.Reason, NextStep: constant.SOURCE_DECISION_PMK})

	//Set Data customerType and spouseType -- Blacklist. Warning, Or Clean --
	mapping.CustomerType = spMap.CustomerType
	mapping.SpouseType = spMap.SpouseType

	//Check vehicle age
	ageVehicle, err := u.usecase.VehicleCheck(req.ManufactureYear, req.CMOCluster, req.BPKBName, req.Tenor, configValue)

	if err != nil {
		return
	}

	if ageVehicle.Result == constant.DECISION_REJECT {
		data = ageVehicle
		mapping.Reason = data.Reason
		return
	}

	trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_PASS, RuleCode: ageVehicle.Code, SourceDecision: constant.SOURCE_DECISION_PMK, Reason: ageVehicle.Reason, NextStep: constant.SOURCE_DECISION_NOKANOSIN})

	// Check Chassis Number with Active Aggrement
	checkChassisNumber, err := u.usecase.CheckAgreementChassisNumber(ctx, req, accessToken)
	if err != nil {
		return
	}

	if checkChassisNumber.Result == constant.DECISION_REJECT {
		data = checkChassisNumber
		mapping.Reason = data.Reason
		return
	}

	trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_PASS, RuleCode: checkChassisNumber.Code, SourceDecision: constant.SOURCE_DECISION_NOKANOSIN, Reason: checkChassisNumber.Reason, NextStep: constant.SOURCE_DECISION_PMK})

	//dataCustomer[0] is result main dupcheck customer
	mainCustomer := dataCustomer[0]

	customerKMB := constant.STATUS_KONSUMEN_NEW

	//Call API Dupcheck V2 - Scan Agreement KMOB
	if mainCustomer != (response.SpDupCekCustomerByID{}) {
		customerKMB, err = u.usecase.CustomerKMB(mainCustomer)

		if err != nil {
			return
		}
	}

	status = customerKMB
	mapping.StatusKonsumen = customerKMB

	if mainCustomer.MaxOverdueDaysROAO != nil {
		mapping.MaxOverdueDaysROAO = *mainCustomer.MaxOverdueDaysROAO
	}

	if mainCustomer.NumberOfPaidInstallment != nil {
		mapping.NumberOfPaidInstallment = *mainCustomer.NumberOfPaidInstallment
	}

	if mainCustomer.MaxOverdueDaysforActiveAgreement != nil {
		mapping.MaxOverdueDaysforActiveAgreement = *mainCustomer.MaxOverdueDaysforActiveAgreement
	}

	mapping.CustomerID = mainCustomer.CustomerID

	mapping.OSInstallmentDue = mainCustomer.OSInstallmentDue
	mapping.NumberofAgreement = mainCustomer.NumberofAgreement

	if customerKMB == constant.STATUS_KONSUMEN_AO || customerKMB == constant.STATUS_KONSUMEN_RO {
		if mapping.NumberofAgreement == 0 {
			mapping.AgreementStatus = constant.AGREEMENT_LUNAS
		} else {
			mapping.AgreementStatus = constant.AGREEMENT_AKTIF
		}
	}

	// Check PMK
	pmk, err := u.usecase.PMK(req.BranchID, customerKMB, req.MonthlyFixedIncome, req.HomeStatus, req.ProfessionID, req.EmploymentSinceYear, req.EmploymentSinceMonth, req.StaySinceYear, req.StaySinceMonth, req.BirthDate, req.Tenor, req.MaritalStatus)
	if err != nil {
		return
	}

	if pmk.Result == constant.DECISION_REJECT {
		data = pmk
		mapping.Reason = fmt.Sprintf("%s -%s", customerKMB, data.Reason)
		return
	}

	trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_PASS, RuleCode: pmk.Code, SourceDecision: constant.SOURCE_DECISION_PMK, Reason: pmk.Reason, NextStep: constant.SOURCE_DECISION_DSR})

	var customerData []request.CustomerData
	customerData = append(customerData, request.CustomerData{
		TransactionID:   req.ProspectID,
		StatusKonsumen:  customerKMB,
		CustomerSegment: req.CustomerSegment,
		IDNumber:        req.IDNumber,
		LegalName:       req.LegalName,
		BirthDate:       req.BirthDate,
		MotherName:      req.MotherName,
	})

	mapping.InstallmentAmountFMF = dataCustomer[0].TotalInstallment

	if married {
		spouse := *req.Spouse
		customerData = append(customerData, request.CustomerData{
			TransactionID: req.ProspectID,
			IDNumber:      spouse.IDNumber,
			LegalName:     spouse.LegalName,
			BirthDate:     spouse.BirthDate,
			MotherName:    spouse.MotherName,
		})

		mapping.InstallmentAmountSpouseFMF = dataCustomer[1].TotalInstallment
	}

	if req.Cluster != "" {
		mapping.Cluster = req.Cluster
	}

	// Check DSR
	dsr, mappingDSR, instOther, instOtherSpouse, instTopup, err := u.usecase.DsrCheck(ctx, req, customerData, req.InstallmentAmount, mapping.InstallmentAmountFMF, mapping.InstallmentAmountSpouseFMF, income, accessToken, configValue)
	if err != nil {
		return
	}

	data = dsr
	mapping.InstallmentAmountOther = instOther
	mapping.InstallmentAmountOtherSpouse = instOtherSpouse
	mapping.InstallmentTopup = instTopup
	mapping.Dsr = dsr.Dsr
	mapping.Reason = data.Reason
	mapping.DetailsDSR = mappingDSR.Details
	mapping.ConfigMaxDSR = configValue.Data.MaxDsr

	if dsr.Result == constant.DECISION_REJECT {
		return
	}

	trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_PASS, RuleCode: dsr.Code, SourceDecision: dsr.SourceDecision, Reason: dsr.Reason, NextStep: constant.SOURCE_DECISION_DUPCHECK})

	// Check Confins
	reasonCustomer := customerKMB
	if strings.Contains("PRIME PRIORITY", req.CustomerSegment) {
		reasonCustomer = fmt.Sprintf("%s %s", customerKMB, req.CustomerSegment)
	}

	if customerKMB == constant.STATUS_KONSUMEN_RO {
		if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {
			checkConfins := response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_MAX_OVD_CONFINS,
				Reason:         fmt.Sprintf("%s %s %d", reasonCustomer, constant.REASON_REJECT_CONFINS_MAXOVD, configValue.Data.MaxOvd),
				StatusKonsumen: customerKMB,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
			}

			data = checkConfins
			mapping.Reason = data.Reason
			return
		} else {
			checkConfins := response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PASS_MAX_OVD_CONFINS,
				Reason:         fmt.Sprintf("%s %s %d", reasonCustomer, constant.REASON_PASS_CONFINS_MAXOVD, configValue.Data.MaxOvd),
				StatusKonsumen: customerKMB,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
			}

			data = checkConfins
			mapping.Reason = data.Reason
		}
	} else if customerKMB == constant.STATUS_KONSUMEN_AO {

		if mapping.InstallmentTopup > 0 {
			reasonCustomer = fmt.Sprintf("%s Top Up", reasonCustomer)
		}

		if (req.CustomerSegment == constant.RO_AO_REGULAR && mapping.MaxOverdueDaysforActiveAgreement > configValue.Data.MaxOvdAORegular) ||
			(strings.Contains("PRIME PRIORITY", req.CustomerSegment) && mapping.InstallmentTopup <= 0 && mapping.MaxOverdueDaysforActiveAgreement > configValue.Data.MaxOvdAOPrimePriority) {
			checkConfins := response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_MENUNGGAK,
				Reason:         fmt.Sprintf("%s %s", reasonCustomer, constant.REASON_MENUNGGAK),
				StatusKonsumen: customerKMB,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
			}

			data = checkConfins
			mapping.Reason = data.Reason
			return
		} else {
			if mapping.NumberOfPaidInstallment >= configValue.Data.AngsuranBerjalan {
				if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {
					checkConfins := response.UsecaseApi{
						Result:         constant.DECISION_REJECT,
						Code:           constant.CODE_MAX_OVD_CONFINS,
						Reason:         fmt.Sprintf("%s - Current >= 6 Bulan Angsuran %s %d", reasonCustomer, constant.REASON_REJECT_CONFINS_MAXOVD, configValue.Data.MaxOvd),
						StatusKonsumen: customerKMB,
						SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
					}

					data = checkConfins
					mapping.Reason = data.Reason
					return

				} else {
					checkConfins := response.UsecaseApi{
						Result:         constant.DECISION_PASS,
						Code:           constant.CODE_PASS_MAX_OVD_CONFINS,
						Reason:         fmt.Sprintf("%s - Current >= 6 Bulan Angsuran %s %d", reasonCustomer, constant.REASON_PASS_CONFINS_MAXOVD, configValue.Data.MaxOvd),
						StatusKonsumen: customerKMB,
						SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
					}

					data = checkConfins
					mapping.Reason = data.Reason

				}
			} else if mapping.NumberOfPaidInstallment > 1 && mapping.NumberOfPaidInstallment < configValue.Data.AngsuranBerjalan {
				if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {
					checkConfins := response.UsecaseApi{
						Result:         constant.DECISION_REJECT,
						Code:           constant.CODE_MAX_OVD_CONFINS,
						Reason:         fmt.Sprintf("%s - Current < 6 Bulan Angsuran %s %d", reasonCustomer, constant.REASON_REJECT_CONFINS_MAXOVD, configValue.Data.MaxOvd),
						StatusKonsumen: customerKMB,
						SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
					}

					data = checkConfins
					mapping.Reason = data.Reason
					return

				} else {
					checkConfins := response.UsecaseApi{
						Result:         constant.DECISION_PASS,
						Code:           constant.CODE_PASS_MAX_OVD_CONFINS,
						Reason:         fmt.Sprintf("%s - Current < 6 Bulan Angsuran %s %d", reasonCustomer, constant.REASON_PASS_CONFINS_MAXOVD, configValue.Data.MaxOvd),
						StatusKonsumen: customerKMB,
						SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
					}

					data = checkConfins
					mapping.Reason = data.Reason

				}
			} else if mapping.NumberOfPaidInstallment <= 1 {
				if mapping.MaxOverdueDaysforActiveAgreement == 0 {
					checkConfins := response.UsecaseApi{
						Result:         constant.DECISION_REJECT,
						Code:           constant.CODE_REJECT_JATUH_TEMPO_PERTAMA,
						Reason:         fmt.Sprintf("%s - Current < 6 Bulan Angsuran - Belum Jatuh Tempo Pertama", reasonCustomer),
						StatusKonsumen: customerKMB,
						SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
					}

					data = checkConfins
					mapping.Reason = data.Reason
					return
				}
			}
		}
	} else {
		checkConfins := response.UsecaseApi{
			Result:         constant.DECISION_PASS,
			Code:           constant.CODE_PASS_MAX_OVD_CONFINS,
			Reason:         fmt.Sprintf("%s", reasonCustomer),
			StatusKonsumen: customerKMB,
			SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
		}

		data = checkConfins
		mapping.Reason = data.Reason
	}

	if data.Result == constant.DECISION_PASS {
		info, _ := json.Marshal(mapping)
		trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_PASS, RuleCode: data.Code, SourceDecision: data.SourceDecision, Reason: data.Reason, Info: string(utils.SafeEncoding(info)), NextStep: constant.SOURCE_DECISION_DUKCAPIL})
	}

	return

}

func (u usecase) CheckBannedPMKDSR(idNumber string) (data response.UsecaseApi, err error) {

	var encryptedIDNumber entity.EncryptedString
	encryptedIDNumber, err = u.repository.GetEncB64(idNumber)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetEncB64 ID Number Error")
		return
	}

	var trxReject entity.TrxBannedPMKDSR
	trxReject, err = u.repository.GetBannedPMKDSR(encryptedIDNumber.MyString)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Banned PMK DSR Error")
		return
	}

	if trxReject != (entity.TrxBannedPMKDSR{}) {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_PERNAH_REJECT_PMK_DSR
		data.Reason = constant.REASON_PERNAH_REJECT_PMK_DSR
		data.SourceDecision = constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR
		return
	}

	return
}

func (u usecase) CheckRejection(idNumber, prospectID string, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedPMKDSR entity.TrxBannedPMKDSR, err error) {

	var encryptedIDNumber entity.EncryptedString
	encryptedIDNumber, err = u.repository.GetEncB64(idNumber)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetEncB64 ID Number Error")
		return
	}

	var trxReject entity.TrxReject
	trxReject, err = u.repository.GetCurrentTrxWithReject(encryptedIDNumber.MyString)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Trx Reject Error")
		return
	}

	if trxReject.RejectPMKDSR > 0 {
		if (trxReject.RejectPMKDSR + trxReject.RejectNIK) >= configValue.Data.AttemptPMKDSR {
			//banned 30 hari
			trxBannedPMKDSR = entity.TrxBannedPMKDSR{
				ProspectID: prospectID,
				IDNumber:   encryptedIDNumber.MyString,
			}
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_PERNAH_REJECT_PMK_DSR
			data.Reason = constant.REASON_PERNAH_REJECT_PMK_DSR
			data.SourceDecision = constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR
			return
		}
	}

	if trxReject.RejectNIK >= configValue.Data.AttemptPMKDSR {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_PERNAH_REJECT_NIK
		data.Reason = constant.REASON_PERNAH_REJECT_NIK
		data.SourceDecision = constant.SOURCE_DECISION_NIK
		return
	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_BELUM_PERNAH_REJECT
	data.Reason = constant.REASON_BELUM_PERNAH_REJECT
	data.SourceDecision = constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR

	return
}

func (u usecase) CheckBannedChassisNumber(chassisNumber string) (data response.UsecaseApi, err error) {

	var trxReject entity.TrxBannedChassisNumber
	trxReject, err = u.repository.GetBannedChassisNumber(chassisNumber)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Banned Chassis Number Error")
		return
	}

	if trxReject != (entity.TrxBannedChassisNumber{}) {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_NOKA_NOSIN
		data.Reason = constant.REASON_REJECT_NOKA_NOSIN
		data.SourceDecision = constant.SOURCE_DECISION_NOKANOSIN
	}

	return
}

func (u usecase) CheckRejectChassisNumber(req request.DupcheckApi, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedChassisNumber entity.TrxBannedChassisNumber, err error) {

	var rejectChassisNumber []entity.RejectChassisNumber
	rejectChassisNumber, err = u.repository.GetCurrentTrxWithRejectChassisNumber(req.RangkaNo)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Reject Chassis Number Error")
		return
	}

	if len(rejectChassisNumber) > 0 {
		trxReject := rejectChassisNumber[len(rejectChassisNumber)-1]
		if (len(rejectChassisNumber) >= configValue.Data.AttemptChassisNumber) || (len(rejectChassisNumber) == 2 && (req.IDNumber != trxReject.IDNumber ||
			req.LegalName != trxReject.LegalName ||
			req.BirthDate != trxReject.BirthDate ||
			req.BirthPlace != trxReject.BirthPlace ||
			req.Gender != trxReject.Gender ||
			req.MaritalStatus != trxReject.MaritalStatus ||
			req.NumOfDependence != trxReject.NumOfDependence ||
			req.StaySinceYear != trxReject.StaySinceYear ||
			req.StaySinceMonth != trxReject.StaySinceMonth ||
			req.HomeStatus != trxReject.HomeStatus ||
			req.LegalZipCode != trxReject.LegalZipCode ||
			req.CompanyZipCode != trxReject.CompanyZipCode ||
			req.ProfessionID != trxReject.ProfessionID ||
			req.MonthlyFixedIncome != trxReject.MonthlyFixedIncome ||
			req.EmploymentSinceYear != trxReject.EmploymentSinceYear ||
			req.EmploymentSinceMonth != trxReject.EmploymentSinceMonth ||
			req.EngineNo != trxReject.EngineNo ||
			req.RangkaNo != trxReject.ChassisNo ||
			req.BPKBName != trxReject.BPKBName ||
			req.ManufactureYear != trxReject.ManufactureYear ||
			req.OTRPrice != trxReject.OTR ||
			req.Tenor != trxReject.Tenor)) {
			//banned 30 hari
			trxBannedChassisNumber = entity.TrxBannedChassisNumber{
				ProspectID: req.ProspectID,
				ChassisNo:  req.RangkaNo,
			}
		}

		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_NOKA_NOSIN
		data.Reason = constant.REASON_REJECT_NOKA_NOSIN
		data.SourceDecision = constant.SOURCE_DECISION_NOKANOSIN
		return
	}

	return
}

func (u usecase) CheckAgreementChassisNumber(ctx context.Context, reqs request.DupcheckApi, accessToken string) (data response.UsecaseApi, err error) {

	var (
		responseAgreementChassisNumber response.AgreementChassisNumber
		hitChassisNumber               *resty.Response
	)

	hitChassisNumber, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+reqs.RangkaNo, nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, reqs.ProspectID, accessToken)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Get Agreement of Chassis Number Timeout")
		return
	}

	if hitChassisNumber.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Get Agreement of Chassis Number Error")
		return
	}

	err = json.Unmarshal([]byte(jsoniter.Get(hitChassisNumber.Body(), "data").ToString()), &responseAgreementChassisNumber)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal Get Agreement of Chassis Number Error")
		return data, err
	}

	if responseAgreementChassisNumber.IsRegistered && responseAgreementChassisNumber.IsActive && len(responseAgreementChassisNumber.IDNumber) > 0 {
		listNikKonsumenDanPasangan := make([]string, 0)

		listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, reqs.IDNumber)
		if reqs.Spouse != nil && reqs.Spouse.IDNumber != "" {
			listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, reqs.Spouse.IDNumber)
		}

		if !utils.Contains(listNikKonsumenDanPasangan, responseAgreementChassisNumber.IDNumber) {
			data.Code = constant.CODE_REJECT_CHASSIS_NUMBER
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_CHASSIS_NUMBER
		} else {
			if responseAgreementChassisNumber.IDNumber == reqs.IDNumber {
				data.Code = constant.CODE_OK_CONSUMEN_MATCH
				data.Result = constant.DECISION_PASS
				data.Reason = constant.REASON_OK_CONSUMEN_MATCH
			} else {
				data.Code = constant.CODE_REJECTION_FRAUD_POTENTIAL
				data.Result = constant.DECISION_REJECT
				data.Reason = constant.REASON_REJECTION_FRAUD_POTENTIAL
			}
		}
	} else {
		data.Code = constant.CODE_AGREEMENT_NOT_FOUND
		data.Result = constant.DECISION_PASS
		data.Reason = constant.REASON_AGREEMENT_NOT_FOUND
	}

	data.SourceDecision = constant.SOURCE_DECISION_NOKANOSIN
	return
}

func (u usecase) BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string) {

	// index 0 = consument
	// index 1 = spouse

	customerType = constant.MESSAGE_BERSIH
	data.SourceDecision = constant.SOURCE_DECISION_BLACKLIST

	if spDupcheck != (response.SpDupCekCustomerByID{}) {

		data.StatusKonsumen = constant.STATUS_KONSUMEN_AO

		if (spDupcheck.TotalInstallment <= 0 && spDupcheck.RRDDate != nil) || (spDupcheck.TotalInstallment > 0 && spDupcheck.RRDDate != nil && spDupcheck.NumberOfPaidInstallment == nil) {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_RO
		}

		if spDupcheck.IsSimiliar == 1 && index == 0 {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_KONSUMEN_SIMILIAR
			data.Reason = constant.REASON_KONSUMEN_SIMILIAR
			customerType = constant.MESSAGE_BLACKLIST
			return

		} else if spDupcheck.BadType == constant.BADTYPE_B {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST
			}
			return

		} else if spDupcheck.MaxOverdueDays > 90 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_OVD_90DAYS

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_OVD_90DAYS
			}
			return

		} else if spDupcheck.NumOfAssetInventoried > 0 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_ASSET_INVENTORY

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_ASSET_INVENTORY
			}
			return

		} else if spDupcheck.IsRestructure == 1 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_RESTRUCTURE

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_RESTRUCTURE
			}
			return

		} else if spDupcheck.BadType == constant.BADTYPE_W {
			customerType = constant.MESSAGE_WARNING
		}

	} else {
		data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
	}

	data = response.UsecaseApi{StatusKonsumen: data.StatusKonsumen, Code: constant.CODE_NON_BLACKLIST_ALL, Reason: constant.REASON_NON_BLACKLIST, Result: constant.DECISION_PASS}

	return
}

func (u usecase) DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	req, _ := json.Marshal(map[string]interface{}{
		"transaction_id":      prospectID,
		"id_number":           idNumber,
		"legal_name":          legalName,
		"birth_date":          birthDate,
		"surgate_mother_name": surgateName,
	})

	custDupcheck, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("DUPCHECK_URL"), req, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Dupcheck Timeout")
		return
	}

	if custDupcheck.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Dupcheck Error")
		return
	}

	json.Unmarshal([]byte(jsoniter.Get(custDupcheck.Body(), "data").ToString()), &spDupcheck)

	return

}

func (u usecase) CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error) {

	if spDupcheck == (response.SpDupCekCustomerByID{}) {
		statusKonsumen = constant.STATUS_KONSUMEN_NEW
		return
	}

	if (spDupcheck.TotalInstallment <= 0 && spDupcheck.RRDDate != nil) || (spDupcheck.TotalInstallment > 0 && spDupcheck.RRDDate != nil && spDupcheck.NumberOfPaidInstallment == nil) {
		statusKonsumen = constant.STATUS_KONSUMEN_RO
		return

	} else if spDupcheck.TotalInstallment > 0 {
		statusKonsumen = constant.STATUS_KONSUMEN_AO
		return

	} else {
		statusKonsumen = constant.STATUS_KONSUMEN_NEW
		return
	}

}

func (u usecase) VehicleCheck(manufactureYear, cmoCluster, bkpbName string, tenor int, configValue response.DupcheckConfig) (data response.UsecaseApi, err error) {

	data.SourceDecision = constant.SOURCE_DECISION_PMK

	currentYear, _ := strconv.Atoi(time.Now().Format("2006-01-02")[0:4])
	BPKBYear, _ := strconv.Atoi(manufactureYear)

	ageVehicle := currentYear - BPKBYear

	ageVehicle += int(tenor / 12)

	if ageVehicle <= configValue.Data.VehicleAge {
		bpkbNameType := 0
		if strings.Contains(os.Getenv("NAMA_SAMA"), bkpbName) {
			bpkbNameType = 1
		}

		mapping, err := u.repository.GetMappingVehicleAge(ageVehicle, cmoCluster, bpkbNameType, tenor)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Vehicle Age Error")
			return data, err
		}

		if mapping.Decision == constant.DECISION_REJECT {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_VEHICLE_AGE_MAX
			data.Reason = fmt.Sprintf("%s Ketentuan", constant.REASON_VEHICLE_AGE_MAX)
			return data, nil
		}

		data.Result = constant.DECISION_PASS
		data.Code = constant.CODE_VEHICLE_SESUAI
		data.Reason = constant.REASON_VEHICLE_SESUAI
		return data, nil

	} else {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_VEHICLE_AGE_MAX
		data.Reason = fmt.Sprintf("%s %d Tahun", constant.REASON_VEHICLE_AGE_MAX, configValue.Data.VehicleAge)
		return
	}

}
