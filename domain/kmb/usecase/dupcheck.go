package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces"
	entity "los-kmb-api/models/dupcheck"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
	"los-kmb-api/models/other"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"strings"
)

type metrics struct {
	repository   interfaces.Repository
	httpclient   httpclient.HttpClient
	usecase      interfaces.Usecase
	multiUsecase interfaces.MultiUsecase
}

func NewMetrics(repository interfaces.Repository, httpclient httpclient.HttpClient, usecase interfaces.Usecase, multiUsecase interfaces.MultiUsecase) interfaces.Metrics {

	return &metrics{
		repository:   repository,
		httpclient:   httpclient,
		usecase:      usecase,
		multiUsecase: multiUsecase,
	}
}

func (u metrics) Dupcheck(ctx context.Context, req request.DupcheckApi, married bool, accessToken string) (mapping response.SpDupcheckMap, status string, data response.UsecaseApi, err error) {

	var (
		customer       []request.SpouseDupcheck
		blackList      response.UsecaseApi
		sp             response.SpDupCekCustomerByID
		dataCustomer   []response.SpDupCekCustomerByID
		spMap          response.SpDupcheckMap
		customerType   string
		newDupcheck    entity.NewDupcheck
		getPhotoInfo   map[string]interface{}
		faceCompareReq request.FaceCompareRequest
	)

	prospectID := req.ProspectID
	income := req.MonthlyFixedIncome + req.MonthlyVariableIncome + req.SpouseIncome
	customer = append(customer, request.SpouseDupcheck{IDNumber: req.IDNumber, LegalName: req.LegalName, BirthDate: req.BirthDate, MotherName: req.MotherName})

	if married {
		customer = append(customer, request.SpouseDupcheck{IDNumber: req.Spouse.IDNumber, LegalName: req.Spouse.LegalName, BirthDate: req.Spouse.BirthDate, MotherName: req.Spouse.MotherName})
	}

	// Face Compare with Faceplus
	faceCompareReq.CustomerID = req.CustomerID
	faceCompareReq.ImageSelfie1 = req.ImageSelfie1
	faceCompareReq.ImageSelfie2 = req.ImageSelfie2
	faceCompareReq.ImageKtp = req.ImageKtp
	faceCompareReq.IDNumber = req.IDNumber
	faceCompareReq.BirthDate = req.BirthDate
	faceCompareReq.BirthPlace = req.BirthPlace
	faceCompareReq.LegalName = req.LegalName
	faceCompareReq.Lob = constant.LOB_KMB

	selfie1, selfie2, isBlur, err := u.multiUsecase.GetPhoto(ctx, faceCompareReq, accessToken)

	if err != nil {
		getPhotoInfo = map[string]interface{}{
			"error": err.Error(),
		}
	} else {
		getPhotoInfo = map[string]interface{}{
			"is_blur": isBlur,
		}
	}

	faceCompare, err := u.usecase.FacePlus(ctx, selfie1, selfie2, faceCompareReq, accessToken, getPhotoInfo)

	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		CentralizeLog(constant.DUPCHECK_LOG, "Check Face Compare", constant.MESSAGE_E, "FACE_COMPARE SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	CentralizeLog(constant.DUPCHECK_LOG, "Check Face Compare", constant.MESSAGE_SUCCESS, "FACE_COMPARE SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: faceCompare})

	if faceCompare.Result == constant.DECISION_REJECT {
		data.Code = constant.CODE_REJECT_FACE_COMPARE
		data.Result = faceCompare.Result
		data.Reason = faceCompare.Reason
		mapping.Reason = data.Reason
		return
	}

	// Exception Reject No. Rangka Banned 30 Days
	nokaBanned30D, err := u.usecase.NokaBanned30D(req)

	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		CentralizeLog(constant.DUPCHECK_LOG, "Check Banned 30D Noka Nosin", constant.MESSAGE_E, "NOKA_NOSIN SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	CentralizeLog(constant.DUPCHECK_LOG, "Check Banned 30D Noka Nosin", constant.MESSAGE_SUCCESS, "NOKA_NOSIN SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: nokaBanned30D})

	if nokaBanned30D.Result == constant.DECISION_REJECT {
		data.Code = nokaBanned30D.Code
		data.Result = nokaBanned30D.Result
		data.Reason = nokaBanned30D.Reason
		mapping.Reason = data.Reason
		return
	}

	// Exception Reject History No. Rangka
	checkRejectionNoka, err := u.usecase.CheckRejectionNoka(req)

	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		CentralizeLog(constant.DUPCHECK_LOG, "Check Histiory Rejection Noka Nosin", constant.MESSAGE_E, "NOKA_NOSIN SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	CentralizeLog(constant.DUPCHECK_LOG, "Check Histiory Rejection Noka Nosin", constant.MESSAGE_SUCCESS, "NOKA_NOSIN SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: checkRejectionNoka})

	fmt.Println(checkRejectionNoka)
	if checkRejectionNoka.Result == constant.DECISION_REJECT {
		data.Code = checkRejectionNoka.Code
		data.Result = checkRejectionNoka.Result
		data.Reason = checkRejectionNoka.Reason
		mapping.Reason = data.Reason
		return
	}

	// Loop check Blacklist customer and spouse
	for i := 0; i < len(customer); i++ {

		// Call API Dupcheck V2
		sp, err = u.usecase.DupcheckIntegrator(ctx, prospectID, customer[i].IDNumber, customer[i].LegalName, customer[i].BirthDate, customer[i].MotherName, accessToken)

		if err != nil {
			CentralizeLog(constant.DUPCHECK_LOG, "Check SP Dupcheck", constant.MESSAGE_E, "CALL_DUPCHECK", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
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
			CentralizeLog(constant.DUPCHECK_LOG, "Check BlackList", constant.MESSAGE_SUCCESS, "BLACKLIST_SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: data})
			return
		}
	}

	//Set Data customerType and spouseType -- Blacklist. Warning, Or Clean --
	mapping.CustomerType = spMap.CustomerType
	mapping.SpouseType = spMap.SpouseType

	//Check vehicle age
	ageVehicle, err := u.usecase.VehicleCheck(req.ManufactureYear)

	if err != nil {
		CentralizeLog(constant.DUPCHECK_LOG, "Check Vehicle", constant.MESSAGE_E, "VEHICLE_SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	CentralizeLog(constant.DUPCHECK_LOG, "Check Vehicle", constant.MESSAGE_SUCCESS, "VEHICLE_SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: ageVehicle})

	if ageVehicle.Result == constant.DECISION_REJECT {
		data = ageVehicle
		mapping.Reason = data.Reason
		return
	}

	// Check Reject No. Rangka
	checkNoka, err := u.usecase.CheckNoka(ctx, req, checkRejectionNoka, accessToken)

	if err != nil {
		CentralizeLog(constant.DUPCHECK_LOG, "Check Noka Nosin", constant.MESSAGE_E, "NOKA_NOSIN SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	CentralizeLog(constant.DUPCHECK_LOG, "Check Noka Nosin", constant.MESSAGE_SUCCESS, "NOKA_NOSIN SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: checkNoka})

	if checkNoka.Result == constant.DECISION_REJECT {
		data = checkNoka
		mapping.Reason = data.Reason
		return
	}

	// Check Chassis Number with Active Aggrement
	checkChassisNumber, err := u.usecase.CheckChassisNumber(ctx, req, checkRejectionNoka, accessToken)

	if err != nil {
		CentralizeLog(constant.DUPCHECK_LOG, "Check Chassis Number", constant.MESSAGE_E, "NOKA_NOSIN SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	CentralizeLog(constant.DUPCHECK_LOG, "Check Chassis Number", constant.MESSAGE_SUCCESS, "NOKA_NOSIN SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: checkChassisNumber})

	if checkChassisNumber.Result == constant.DECISION_REJECT {
		data = checkChassisNumber
		mapping.Reason = data.Reason
		return
	}

	//dataCustomer[0] is result main dupcheck customer
	mainCustomer := dataCustomer[0]

	customerKMB := constant.STATUS_KONSUMEN_NEW

	//Call API Dupcheck V2 - Scan Agreement KMOB
	if mainCustomer != (response.SpDupCekCustomerByID{}) {
		customerKMB, err = u.usecase.CustomerKMB(mainCustomer)

		if err != nil {
			CentralizeLog(constant.DUPCHECK_LOG, "Check Dupcheck KMOB", constant.MESSAGE_E, "STATUS_CUSTOMER_CHECK", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
			return
		}
	}

	status = customerKMB

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

	//Get parameterize config
	config, err := u.repository.GetDupcheckConfig()

	if err != nil {
		err = errors.New("upstream_service_error - Error Get Parameterize Config")
		CentralizeLog(constant.DUPCHECK_LOG, "Check Dupcheck Config", constant.MESSAGE_E, "GET_DUPCHECK_PARAMETERIZE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	var configValue response.DupcheckConfig

	json.Unmarshal([]byte(config.Value), &configValue)

	mapping.OSInstallmentDue = mainCustomer.OSInstallmentDue
	mapping.NumberofAgreement = mainCustomer.NumberofAgreement
	mapping.AgreementStatus = constant.AGREEMENT_AKTIF
	if mapping.NumberofAgreement == 0 {
		mapping.AgreementStatus = constant.AGREEMENT_LUNAS
	}

	if customerKMB == constant.STATUS_KONSUMEN_AO || customerKMB == constant.STATUS_KONSUMEN_RO {

		if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {
			checkOVD := response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_MAX_OVD_CONFINS,
				Reason:         fmt.Sprintf("%s %s %d", customerKMB, constant.REASON_REJECT_CONFINS_MAXOVD, configValue.Data.MaxOvd),
				StatusKonsumen: customerKMB,
			}

			data = checkOVD
			mapping.Reason = data.Reason
			return
		}
	}

	// Check PMK
	pmk := u.usecase.PMK(req.MonthlyFixedIncome, req.HomeStatus, req.JobPosition, req.EmploymentSinceYear, req.EmploymentSinceMonth, req.StaySinceYear, req.StaySinceMonth, req.BirthDate, req.Tenor, req.MaritalStatus)

	CentralizeLog(constant.DUPCHECK_LOG, "Check PMK", constant.MESSAGE_SUCCESS, "PMK_SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: pmk})

	if pmk.Result == constant.DECISION_REJECT {
		data = pmk
		mapping.Reason = fmt.Sprintf("%s -%s", customerKMB, data.Reason)
		return
	}

	if pmk.Result == constant.DECISION_REJECT {
		data = pmk
		mapping.Reason = data.Reason

		return
	}

	var customerData []request.CustomerData
	customerData = append(customerData, request.CustomerData{
		StatusKonsumen: customerKMB,
		IDNumber:       req.IDNumber,
		LegalName:      req.LegalName,
		BirthDate:      req.BirthDate,
		MotherName:     req.MotherName,
	})

	mapping.InstallmentAmountFMF = dataCustomer[0].TotalInstallment

	if married {
		spouse := *req.Spouse
		customerData = append(customerData, request.CustomerData{
			StatusKonsumen: "",
			IDNumber:       spouse.IDNumber,
			LegalName:      spouse.LegalName,
			BirthDate:      spouse.BirthDate,
			MotherName:     spouse.MotherName,
		})

		mapping.InstallmentAmountSpouseFMF = dataCustomer[1].TotalInstallment
	}

	if customerKMB == constant.STATUS_KONSUMEN_RO || customerKMB == constant.STATUS_KONSUMEN_AO {
		newDupcheck, _ = u.repository.GetNewDupcheck(req.ProspectID)
		if newDupcheck.CustomerType == "" {
			newDupcheck.ProspectID = req.ProspectID
			newDupcheck.CustomerStatus = customerKMB

			// LobID 1 WG, 2 KMB, 3 KMOB, 4 UC
			reqCustomerDomain := request.ReqCustomerDomain{
				IDNumber:   req.IDNumber,
				LegalName:  req.LegalName,
				BirthDate:  req.BirthDate,
				MotherName: req.MotherName,
				LobID:      constant.LOBID_KMB,
			}

			var customerDomainData response.CustomerDomainData
			customerDomainData, err = u.usecase.CustomerDomainGetData(ctx, reqCustomerDomain, req.ProspectID, accessToken)
			if err != nil {
				return
			}

			if len(customerDomainData.CustomerSegmentation) > 0 {
				newDupcheck.CustomerType = customerDomainData.CustomerSegmentation[0].SegmentName
			} else {
				newDupcheck.CustomerType = constant.RO_AO_REGULAR
			}

			err = u.repository.SaveNewDupcheck(newDupcheck)
			if err != nil {
				return
			}
		}
	}

	// Check DSR
	dsr, _, instOther, instOtherSpouse, instTopup, err := u.usecase.DsrCheck(ctx, req.ProspectID, req.EngineNo, customerData, req.InstallmentAmount, mapping.InstallmentAmountFMF, mapping.InstallmentAmountSpouseFMF, income, newDupcheck, accessToken)
	if err != nil {
		CentralizeLog(constant.DUPCHECK_LOG, "Check DSR", constant.MESSAGE_E, "DSR_SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: err.Error()})
		return
	}

	mapping.InstallmentAmountOther = instOther
	mapping.InstallmentAmountOtherSpouse = instOtherSpouse
	mapping.InstallmentTopup = instTopup
	mapping.Dsr = dsr.Dsr

	CentralizeLog(constant.DUPCHECK_LOG, "Check DSR", constant.MESSAGE_SUCCESS, "DSR_SERVICE", false, other.CustomLog{ProspectID: prospectID, Info: dsr})

	data = dsr
	mapping.Reason = data.Reason

	return

}
