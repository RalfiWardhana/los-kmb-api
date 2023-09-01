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
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
)

func (u multiUsecase) Dupcheck(ctx context.Context, req request.DupcheckApi, married bool, accessToken string) (mapping response.SpDupcheckMap, status string, data response.UsecaseApi, err error) {

	var (
		customer       []request.SpouseDupcheck
		blackList      response.UsecaseApi
		sp             response.SpDupCekCustomerByID
		dataCustomer   []response.SpDupCekCustomerByID
		spMap          response.SpDupcheckMap
		customerType   string
		newDupcheck    entity.NewDupcheck
		faceCompareReq request.FaceCompareRequest
	)

	prospectID := req.ProspectID
	income := req.MonthlyFixedIncome + req.MonthlyVariableIncome + req.SpouseIncome
	customer = append(customer, request.SpouseDupcheck{IDNumber: req.IDNumber, LegalName: req.LegalName, BirthDate: req.BirthDate, MotherName: req.MotherName})

	if married {
		customer = append(customer, request.SpouseDupcheck{IDNumber: req.Spouse.IDNumber, LegalName: req.Spouse.LegalName, BirthDate: req.Spouse.BirthDate, MotherName: req.Spouse.MotherName})
	}

	// Face Compare with Faceplus
	faceCompareReq.ProspectID = req.ProspectID
	faceCompareReq.ImageKtp = req.ImageKtp
	faceCompareReq.ImageSelfie = req.ImageSelfie
	faceCompareReq.IDNumber = req.IDNumber
	faceCompareReq.BirthDate = req.BirthDate
	faceCompareReq.BirthPlace = req.BirthPlace
	faceCompareReq.LegalName = req.LegalName
	faceCompareReq.Lob = constant.LOB_KMB

	imageKtp, err := u.usecase.GetBase64Media(ctx, req.ImageKtp, 0, accessToken)
	if err != nil {
		return
	}

	imageSelfie, err := u.usecase.GetBase64Media(ctx, req.ImageSelfie, 0, accessToken)
	if err != nil {
		return
	}

	faceCompare, err := u.usecase.FacePlus(ctx, imageKtp, imageSelfie, faceCompareReq, accessToken)

	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		return
	}

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
		return
	}

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
		return
	}

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

	//Set Data customerType and spouseType -- Blacklist. Warning, Or Clean --
	mapping.CustomerType = spMap.CustomerType
	mapping.SpouseType = spMap.SpouseType

	//Check vehicle age
	ageVehicle, err := u.usecase.VehicleCheck(req.ManufactureYear)

	if err != nil {
		return
	}

	if ageVehicle.Result == constant.DECISION_REJECT {
		data = ageVehicle
		mapping.Reason = data.Reason
		return
	}

	// Check Reject No. Rangka
	checkNoka, err := u.usecase.CheckNoka(ctx, req, checkRejectionNoka, accessToken)

	if err != nil {
		return
	}

	if checkNoka.Result == constant.DECISION_REJECT {
		data = checkNoka
		mapping.Reason = data.Reason
		return
	}

	// Check Chassis Number with Active Aggrement
	checkChassisNumber, err := u.usecase.CheckChassisNumber(ctx, req, checkRejectionNoka, accessToken)

	if err != nil {
		return
	}

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

			var customerDomainData response.DataCustomer
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
		return
	}

	mapping.InstallmentAmountOther = instOther
	mapping.InstallmentAmountOtherSpouse = instOtherSpouse
	mapping.InstallmentTopup = instTopup
	mapping.Dsr = dsr.Dsr

	data = dsr
	mapping.Reason = data.Reason

	return

}

func (u usecase) BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string) {

	// index 0 = consument
	// index 1 = spouse

	customerType = constant.MESSAGE_BERSIH

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

	data = response.UsecaseApi{StatusKonsumen: data.StatusKonsumen, Code: constant.CODE_NON_BLACKLIST, Reason: constant.REASON_NON_BLACKLIST, Result: constant.DECISION_PASS}

	return
}

func (u usecase) ConsumentCheck(req response.SpDupCekCustomerByID) (data response.UsecaseApi, mapping response.SpDupcheckMap, err error) {

	if req != (response.SpDupCekCustomerByID{}) {

		var config entity.AppConfig

		config, err = u.repository.GetDupcheckConfig()

		if err != nil {
			err = errors.New("upstream_service_error - Error Get Parameterize Config")
			return
		}

		var configValue response.DupcheckConfig

		json.Unmarshal([]byte(config.Value), &configValue)

		if req.MaxOverdueDaysROAO != nil {
			mapping.MaxOverdueDaysROAO = *req.MaxOverdueDaysROAO
		}

		if req.NumberOfPaidInstallment != nil {
			mapping.NumberOfPaidInstallment = *req.NumberOfPaidInstallment
		}

		if req.MaxOverdueDaysforActiveAgreement != nil {
			mapping.MaxOverdueDaysforActiveAgreement = *req.MaxOverdueDaysforActiveAgreement
		}

		mapping.CustomerID = req.CustomerID

		if (req.TotalInstallment <= 0 && req.RRDDate != nil) || (req.TotalInstallment > 0 && req.RRDDate != nil && req.NumberOfPaidInstallment == nil) {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_RO

			if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_RO_OVDLTE30
				data.Reason = fmt.Sprintf("RO - OVD Maks <= %d days", configValue.Data.MinOvd)

			} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MaxOvd {

				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_RO_OVDGT30_LTE90
				data.Reason = fmt.Sprintf("RO - OVD Maks > %d-%d days", configValue.Data.MinOvd, configValue.Data.MaxOvd)

			} else {

				data.Result = constant.DECISION_REJECT
				data.Code = constant.CODE_RO_OVDGT90
				data.Reason = fmt.Sprintf("RO - OVD Maks > %d days", configValue.Data.MaxOvd)

			}

			return

		} else if req.TotalInstallment > 0 {

			if req.NumberOfPaidInstallment != nil && mapping.NumberOfPaidInstallment >= 0 {

				data.StatusKonsumen = constant.STATUS_KONSUMEN_AO

				if req.OSInstallmentDue <= 0 {

					if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment >= configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSGT6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Lancar >= %d bulan Angsuran - OVD Maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MaxOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSGT6_OVDGT30_LTE90
							data.Reason = fmt.Sprintf("AO - Lancar >= %d bulan Angsuran - OVD maks > %d-%d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd, configValue.Data.MaxOvd)

						} else {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_LANCAR_ANGSGT6_OVDGT90
							data.Reason = fmt.Sprintf("AO - Lancar >= %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MaxOvd)

						}

					} else if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment < configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_OVDGT90
							data.Reason = fmt.Sprintf("AO - OVD Maks > %d days", configValue.Data.MaxOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSLTE6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Lancar < %d bulan Angsuran - OVD maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)
						} else {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSLTE6_OVDGT30
							data.Reason = fmt.Sprintf("AO - Lancar < %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)
						}
					}

				} else {

					if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment >= configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSGT6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Menunggak >= %d bulan Angsuran - OVD Maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MaxOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSGT6_OVDGT30_LTE90
							data.Reason = fmt.Sprintf("AO - Menunggak >= %d bulan Angsuran - OVD maks > %d - %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd, configValue.Data.MaxOvd)
						} else {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSGT6_OVDGT90
							data.Reason = fmt.Sprintf("AO - Menunggak >= %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MaxOvd)

						}

					} else if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment < configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_OVDGT90
							data.Reason = fmt.Sprintf("AO - OVD Maks > %d days", configValue.Data.MaxOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSLTE6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Menunggak < %d bulan Angsuran - OVD maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)

						} else {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSLTE6_OVDGT30
							data.Reason = fmt.Sprintf("AO - Menunggak < %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)
						}
					}
				}

			}

		} else {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_KONSUMEN_UNIDENTIFIED
			data.Reason = constant.REASON_KONSUMEN_UNIDENTIFIED

		}

	} else {
		data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
		data.Result = constant.DECISION_PASS
		data.Code = constant.CODE_KONSUMEN_NEW
		data.Reason = constant.REASON_KONSUMEN_NEW
	}

	return
}

func (u usecase) CustomerDomainGetData(ctx context.Context, req request.ReqCustomerDomain, prospectID string, accessToken string) (customerDomainData response.DataCustomer, err error) {

	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_CUSTOMER_DOMAIN_GET_DATA"))

	if dummy {

		dummyCustomerDomain, _ := u.repository.GetDummyCustomerDomain(req.IDNumber)

		var customerDomain response.CustomerDomain

		json.Unmarshal([]byte(dummyCustomerDomain.Response), &customerDomain)

		customerDomainData = customerDomain.Data

	} else {

		param, _ := json.Marshal(req)

		header := map[string]string{
			"Authorization": accessToken,
		}

		url := os.Getenv("CUSTOMER_DOMAIN_GET_DATA")

		resp, err := u.httpclient.EngineAPI(ctx, constant.DUPCHECK_LOG, url, param, header, constant.METHOD_POST, false, 0, 60, prospectID, accessToken)

		if err != nil && resp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Customer Domain")
			return customerDomainData, err
		}

		var customerDomain response.CustomerDomain

		json.Unmarshal(resp.Body(), &customerDomain)

		customerDomainData = customerDomain.Data

	}

	return
}

func (u usecase) NokaBanned30D(req request.DupcheckApi) (data response.RejectionNoka, err error) {

	data.Code = constant.CODE_REJECTION_OK
	data.Result = constant.RESULT_REJECTION_OK
	data.Reason = constant.REASON_REJECTION_OK

	nokaBanned30D, err := u.repository.GetLatestBannedRejectionNoka(req.RangkaNo)
	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Latest Banned Rejection Noka")
		return
	}

	data.IsBannedActive = false
	if nokaBanned30D != (entity.DupcheckRejectionNokaNosin{}) && nokaBanned30D.IsBanned == 1 {
		bannedDate := nokaBanned30D.CreatedAt
		dueDate := bannedDate.AddDate(0, 0, constant.DAY_RANGE_BANNED_REJECT_NOKA)
		dueDateString := dueDate.Format("2006-01-02")

		if time.Now().Format(constant.FORMAT_DATE) <= dueDateString {
			data.IsBannedActive = true
		}

		if data.IsBannedActive {
			data.Code = constant.CODE_REJECT_NOKA_NOSIN
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_NOKA_NOSIN
			return
		}
	}
	return
}

func (u usecase) CheckRejectionNoka(req request.DupcheckApi) (data response.RejectionNoka, err error) {

	var (
		inRejectionNoka    bool
		getHistoryReject   []entity.DupcheckRejectionPMK
		rejection          []entity.DupcheckRejectionPMK
		checkHistoryReject entity.DupcheckRejectionPMK
	)

	nokaBannedCurrentDate, err := u.repository.GetLatestRejectionNoka(req.RangkaNo)

	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Latest Rejection Noka")
		return
	}

	inRejectionNoka = false
	data.CurrentBannedNotEmpty = false
	if nokaBannedCurrentDate != (entity.DupcheckRejectionNokaNosin{}) {

		// Must be simplified
		inRejectionNoka = true
		data.CurrentBannedNotEmpty = true

		data.IDNumber = nokaBannedCurrentDate.IDNumber
		data.LegalName = nokaBannedCurrentDate.LegalName
		data.BirthPlace = nokaBannedCurrentDate.BirthPlace
		data.BirthDate = nokaBannedCurrentDate.BirthDate
		data.MonthlyFixedIncome = nokaBannedCurrentDate.MonthlyFixedIncome
		data.EmploymentSinceYear = nokaBannedCurrentDate.EmploymentSinceYear
		data.EmploymentSinceMonth = nokaBannedCurrentDate.EmploymentSinceMonth
		data.StaySinceYear = nokaBannedCurrentDate.StaySinceYear
		data.StaySinceMonth = nokaBannedCurrentDate.StaySinceMonth
		data.BPKBName = nokaBannedCurrentDate.BPKBName
		data.Gender = nokaBannedCurrentDate.Gender
		data.MaritalStatus = nokaBannedCurrentDate.MaritalStatus
		data.NumOfDependence = nokaBannedCurrentDate.NumOfDependence
		data.NTF = nokaBannedCurrentDate.NTF
		data.OTRPrice = nokaBannedCurrentDate.OTRPrice
		data.LegalZipCode = nokaBannedCurrentDate.LegalZipCode
		data.Tenor = nokaBannedCurrentDate.Tenor
		data.ManufacturingYear = nokaBannedCurrentDate.ManufacturingYear
		data.ProfessionID = nokaBannedCurrentDate.ProfessionID
		data.CompanyZipCode = nokaBannedCurrentDate.CompanyZipCode
		data.HomeStatus = nokaBannedCurrentDate.HomeStatus

		data.Code = constant.CODE_REJECTION_OK
		data.Result = constant.RESULT_REJECTION_OK
		data.Reason = constant.REASON_REJECTION_OK
		return
	}

	if !inRejectionNoka {

		rejection, err = u.repository.GetAllReject(req.IDNumber)

		if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
			err = errors.New(constant.ERROR_UPSTREAM + " - Error Get All Rejection")
			return
		}

		if len(rejection) >= constant.ATTEMPT_REJECT {
			for i := 0; i < len(rejection); i++ {
				if rejection[i].RejectPMK == 1 || rejection[i].RejectDSR == 1 {
					data.Code = constant.CODE_REJECT_PMK_DSR
					data.Result = constant.DECISION_REJECT
					data.Reason = constant.REASON_REJECT_PMK_DSR
					return
				}
			}

			data.Code = constant.CODE_REJECT_NIK_KTP
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_NIK_KTP
			return

		} else {
			getHistoryReject, err = u.repository.GetHistoryRejectAttempt(req.IDNumber)

			if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Get All Rejection")
				return
			}

			if len(getHistoryReject) > 0 {
				historyResult := getHistoryReject[0]
				blackListDate := historyResult.Date

				if historyResult.RejectAttempt >= constant.ATTEMPT_REJECT_PMK_DSR {
					parsedDate, _ := time.Parse("2006-01-02", blackListDate)
					dueDate := parsedDate.AddDate(0, 0, 30)
					dueDateString := dueDate.Format("2006-01-02")

					if time.Now().Format(constant.FORMAT_DATE) < dueDateString {
						data.Code = constant.CODE_REJECT_PMK_DSR
						data.Result = constant.DECISION_REJECT
						data.Reason = constant.REASON_REJECT_PMK_DSR
						return
					}
				} else {
					date, _ := time.Parse(time.RFC3339, blackListDate)
					dateString := date.Format("2006-01-02")

					checkHistoryReject, err = u.repository.GetCheckingRejectAttempt(req.IDNumber, dateString)
					if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
						err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Checking Reject Attempt")
						return
					}

					if checkHistoryReject.RejectAttempt >= constant.ATTEMPT_REJECT_PMK_DSR {
						data.Code = constant.CODE_REJECT_PMK_DSR
						data.Result = constant.DECISION_REJECT
						data.Reason = constant.REASON_REJECT_PMK_DSR
						return
					}
				}
			}

			data.Code = constant.CODE_REJECTION_OK
			data.Result = constant.RESULT_REJECTION_OK
			return
		}
	}
	return
}

func (u usecase) CheckNoka(ctx context.Context, reqs request.DupcheckApi, nokaBanned response.RejectionNoka, accessToken string) (data response.UsecaseApi, err error) {

	var (
		nokaData entity.DupcheckRejectionNokaNosin
	)

	data.Code = constant.CODE_REJECTION_NOTIF
	data.Result = constant.RESULT_NOTIF
	data.Reason = constant.REASON_NOTIF

	// Noka Data to Save
	nokaData.Id = utils.UniqueID(15)
	nokaData.NoRangka = reqs.RangkaNo
	nokaData.NoMesin = reqs.EngineNo
	nokaData.ProspectID = reqs.ProspectID
	nokaData.IDNumber = reqs.IDNumber
	nokaData.LegalName = reqs.LegalName
	nokaData.BirthPlace = reqs.BirthPlace
	nokaData.BirthDate = reqs.BirthDate
	nokaData.MonthlyFixedIncome = reqs.MonthlyFixedIncome
	nokaData.EmploymentSinceYear = reqs.EmploymentSinceYear
	nokaData.EmploymentSinceMonth = reqs.EmploymentSinceMonth
	nokaData.StaySinceYear = reqs.StaySinceYear
	nokaData.StaySinceMonth = reqs.StaySinceMonth
	nokaData.BPKBName = reqs.BPKBName
	nokaData.Gender = reqs.Gender
	nokaData.MaritalStatus = reqs.MaritalStatus
	nokaData.NumOfDependence = reqs.NumOfDependence
	nokaData.NTF = reqs.NTF
	nokaData.OTRPrice = reqs.OTRPrice
	nokaData.LegalZipCode = reqs.LegalZipCode
	nokaData.Tenor = reqs.Tenor
	nokaData.ManufacturingYear = reqs.ManufactureYear
	nokaData.ProfessionID = reqs.ProfessionID
	nokaData.CompanyZipCode = reqs.CompanyZipCode
	nokaData.HomeStatus = reqs.HomeStatus

	// Check Current Date Banned Status
	if nokaBanned.CurrentBannedNotEmpty {
		numberOfRetry := nokaBanned.NumberOfRetry + 1

		if numberOfRetry == 1 {
			nokaData.NumberOfRetry = numberOfRetry

		} else if numberOfRetry > 1 && numberOfRetry < constant.ATTEMPT_BANNED_REJECTION_NOKA {
			nokaData.NumberOfRetry = numberOfRetry

			// Maybe could be simplified
			if nokaBanned.IDNumber != reqs.IDNumber && nokaBanned.LegalName != reqs.LegalName && nokaBanned.BirthPlace != reqs.BirthPlace && nokaBanned.BirthDate != reqs.BirthDate && nokaBanned.MonthlyFixedIncome != reqs.MonthlyFixedIncome && nokaBanned.EmploymentSinceYear != reqs.EmploymentSinceYear && nokaBanned.EmploymentSinceMonth != reqs.EmploymentSinceMonth && nokaBanned.StaySinceYear != reqs.StaySinceYear && nokaBanned.StaySinceMonth != reqs.StaySinceMonth && nokaBanned.BPKBName != reqs.BPKBName && nokaBanned.Gender != reqs.Gender && nokaBanned.MaritalStatus != reqs.MaritalStatus && nokaBanned.NumOfDependence != reqs.NumOfDependence && nokaBanned.NTF != reqs.NTF && nokaBanned.OTRPrice != reqs.OTRPrice && nokaBanned.LegalZipCode != reqs.LegalZipCode && nokaBanned.Tenor != reqs.Tenor && nokaBanned.ManufacturingYear != reqs.ManufactureYear && nokaBanned.ProfessionID != reqs.ProfessionID && nokaBanned.CompanyZipCode != reqs.CompanyZipCode && nokaBanned.HomeStatus != reqs.HomeStatus {
				// MAKSUDNYA -> KETIKA ADA DATA DALAM KETENTUAN PARAMETER KREDIT BERBEDA DENGAN INPUTAN SEBELUMNYA, MAKA NO. RANGKA TSB DI BANNED 30 HARI
				nokaBanned.IsBanned = 1
			}
		} else if numberOfRetry >= constant.ATTEMPT_BANNED_REJECTION_NOKA {
			nokaBanned.NumberOfRetry = numberOfRetry
			nokaBanned.IsBanned = 1
		}

		if err = u.repository.SaveDataNoka(nokaData); err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Error Save Data Noka Nosin")
			return
		}

		data.Code = constant.CODE_REJECT_NOKA_NOSIN
		data.Result = constant.DECISION_REJECT
		data.Reason = constant.REASON_REJECT_NOKA_NOSIN
		return
	}

	return
}

func (u usecase) CheckChassisNumber(ctx context.Context, reqs request.DupcheckApi, nokaBanned response.RejectionNoka, accessToken string) (data response.UsecaseApi, err error) {

	var (
		nokaData                       entity.DupcheckRejectionNokaNosin
		trxApiLog                      entity.TrxApiLog
		responseAgreementChassisNumber response.AgreementChassisNumber
		response_api_log               []byte
	)

	trxApiLog.ProspectID = reqs.ProspectID
	trxApiLog.Request = os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL") + reqs.RangkaNo
	trxApiLog.DtmRequest = time.Now()

	// Hit Integrator Get Chasis Number
	dummyChassis, _ := strconv.ParseBool(os.Getenv("DUMMY_AGREEMENT_OF_CHASSIS_NUMBER"))

	if dummyChassis {
		dummyChassisAgreementNumber, _ := u.repository.GetDummyAgreementChassisNumber(reqs.IDNumber)

		var resAgreementChassisNumber response.ResAgreementChassisNumber

		json.Unmarshal([]byte(dummyChassisAgreementNumber.Response), &resAgreementChassisNumber)

		responseAgreementChassisNumber = resAgreementChassisNumber.Data

	} else {

		var hitChassisNumber *resty.Response

		hitChassisNumber, err = u.httpclient.EngineAPI(ctx, constant.DUPCHECK_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+reqs.RangkaNo, nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, reqs.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Get Agreement of Chassis Number Timeout")
			return
		}

		if hitChassisNumber.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Get Agreement of Chassis Number Error")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(hitChassisNumber.Body(), "data").ToString()), &responseAgreementChassisNumber)
	}

	response_api_log, _ = json.Marshal(responseAgreementChassisNumber)
	trxApiLog.Response = string(response_api_log)
	trxApiLog.DtmResponse = time.Now()
	trxApiLog.Type = constant.TYPE_API_LOGS_NOKA

	if err = u.repository.SaveDataApiLog(trxApiLog); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Save Data to API Logs")
		return
	}

	if responseAgreementChassisNumber == (response.AgreementChassisNumber{}) {
		return
	}

	agreement := responseAgreementChassisNumber
	if agreement.IsRegistered && agreement.IsActive && len(agreement.IDNumber) > 0 {
		listNikKonsumenDanPasangan := make([]string, 0)

		listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, reqs.IDNumber)
		if reqs.Spouse != nil && reqs.Spouse.IDNumber != "" {
			listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, reqs.Spouse.IDNumber)
		}

		if !utils.Contains(listNikKonsumenDanPasangan, agreement.IDNumber) {
			// Noka Data to Save
			nokaData.Id = utils.UniqueID(15)
			nokaData.NoRangka = reqs.RangkaNo
			nokaData.NoMesin = reqs.EngineNo
			nokaData.ProspectID = reqs.ProspectID
			nokaData.IDNumber = reqs.IDNumber
			nokaData.LegalName = reqs.LegalName
			nokaData.BirthPlace = reqs.BirthPlace
			nokaData.BirthDate = reqs.BirthDate
			nokaData.MonthlyFixedIncome = reqs.MonthlyFixedIncome
			nokaData.EmploymentSinceYear = reqs.EmploymentSinceYear
			nokaData.EmploymentSinceMonth = reqs.EmploymentSinceMonth
			nokaData.StaySinceYear = reqs.StaySinceYear
			nokaData.StaySinceMonth = reqs.StaySinceMonth
			nokaData.BPKBName = reqs.BPKBName
			nokaData.Gender = reqs.Gender
			nokaData.MaritalStatus = reqs.MaritalStatus
			nokaData.NumOfDependence = reqs.NumOfDependence
			nokaData.NTF = reqs.NTF
			nokaData.OTRPrice = reqs.OTRPrice
			nokaData.LegalZipCode = reqs.LegalZipCode
			nokaData.Tenor = reqs.Tenor
			nokaData.ManufacturingYear = reqs.ManufactureYear
			nokaData.ProfessionID = reqs.ProfessionID
			nokaData.CompanyZipCode = reqs.CompanyZipCode
			nokaData.HomeStatus = reqs.HomeStatus

			nokaData.NumberOfRetry = 0

			if err = u.repository.SaveDataNoka(nokaData); err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error Save Data Noka Nosin")
				return
			}

			data.Code = constant.CODE_REJECT_CHASSIS_NUMBER
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_CHASSIS_NUMBER
		} else {
			if agreement.IDNumber == reqs.IDNumber {
				data.Code = constant.CODE_OK_CONSUMEN_MATCH
				data.Result = constant.RESULT_OK
				data.Reason = constant.REASON_OK_CONSUMEN_MATCH
			} else {
				data.Code = constant.CODE_REJECTION_FRAUD_POTENTIAL
				data.Result = constant.DECISION_REJECT
				data.Reason = constant.REASON_REJECTION_FRAUD_POTENTIAL
			}
		}
	} else {
		data.Code = constant.CODE_AGREEMENT_NOT_FOUND
		data.Result = constant.RESULT_OK
		data.Reason = constant.REASON_AGREEMENT_NOT_FOUND
	}
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

func (u usecase) VehicleCheck(manufactureYear string) (data response.UsecaseApi, err error) {

	config, err := u.repository.GetDupcheckConfig()

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Parameterize Config")
		return
	}

	var configValue response.DupcheckConfig

	json.Unmarshal([]byte(config.Value), &configValue)

	currentYear, _ := strconv.Atoi(time.Now().Format("2006-01-02")[0:4])
	BPKBYear, _ := strconv.Atoi(manufactureYear)

	ageVehicle := currentYear - BPKBYear

	if ageVehicle <= configValue.Data.VehicleAge {
		data.Result = constant.DECISION_PASS
		data.Code = constant.CODE_VEHICLE_SESUAI
		data.Reason = constant.REASON_VEHICLE_SESUAI
		return

	} else {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_VEHICLE_AGE_MAX
		data.Reason = fmt.Sprintf("%s %d Tahun", constant.REASON_VEHICLE_AGE_MAX, configValue.Data.VehicleAge)
		return
	}

}

func (u usecase) GetLatestPaidInstallment(ctx context.Context, req request.ReqLatestPaidInstallment, prospectID string, accessToken string) (data response.LatestPaidInstallmentData, err error) {

	dummy, _ := strconv.ParseBool(os.Getenv("DUMMY_LATEST_PAID_INSTALLMENT"))

	if dummy {
		dummyLatestPaidInstallment, _ := u.repository.GetDummyLatestPaidInstallment(req.IDNumber)

		var latestPaidInstallment response.LatestPaidInstallment

		json.Unmarshal([]byte(dummyLatestPaidInstallment.Response), &latestPaidInstallment)

		data = latestPaidInstallment.Data

	} else {

		var dupcheckMDM *resty.Response
		dupcheckMDM, err = u.httpclient.EngineAPI(ctx, constant.DUPCHECK_LOG, fmt.Sprintf("%s/%s/3", os.Getenv("DUPCHECK_GET_LATEST_PAID_INSTALLMENT"), req.CustomerID), nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, prospectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Dupcheck MDM Latest Paid Installment Timeout")
			return
		}

		if dupcheckMDM.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Dupcheck MDM Latest Paid Installment Error")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(dupcheckMDM.Body(), "data").ToString()), &data)
	}

	return
}
