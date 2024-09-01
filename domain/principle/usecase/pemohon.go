package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

func (u usecase) GetEmployeeData(ctx context.Context, employeeID string) (data response.EmployeeCMOResponse, err error) {

	var (
		dataEmployee        response.EmployeeCareerHistory
		respGetEmployeeData response.GetEmployeeByID
		today               string
		parsedTime          time.Time
		todayDate           time.Time
		givenDate           time.Time
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	header := map[string]string{
		"Authorization": "Bearer " + middlewares.HrisApiData.Token,
	}

	payload := request.ReqHrisCareerHistory{
		Limit:     "100",
		Page:      1,
		Column:    "real_career_date",
		Ascending: false,
		Query:     "employee_id==" + employeeID,
	}

	param, _ := json.Marshal(payload)
	getDataEmployee, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("HRIS_GET_EMPLOYEE_DATA_URL"), param, header, constant.METHOD_POST, false, 0, timeout, "", middlewares.UserInfoData.AccessToken)

	if getDataEmployee.StatusCode() == 504 || getDataEmployee.StatusCode() == 502 {
		// err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Employee Data Timeout")
		return
	}

	if getDataEmployee.StatusCode() != 200 && getDataEmployee.StatusCode() != 504 && getDataEmployee.StatusCode() != 502 {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Get Employee Data Error")
		return
	}

	if err == nil && getDataEmployee.StatusCode() == 200 {
		json.Unmarshal([]byte(jsoniter.Get(getDataEmployee.Body()).ToString()), &respGetEmployeeData)

		isCmoActive := false
		if len(respGetEmployeeData.Data) > 0 && respGetEmployeeData.Data[0].PositionGroupCode == "AO" {
			isCmoActive = true
		}

		var lastIndex int = -1
		// Cek dulu apakah saat ini employee tersebut adalah berposisi sebagai "CMO"
		if isCmoActive {
			// Mencari index terakhir yang mengandung position_group_code "AO"
			for i, emp := range respGetEmployeeData.Data {
				if emp.PositionGroupCode == "AO" {
					lastIndex = i
				}
			}
		}

		if lastIndex == -1 {
			// Jika tidak ada data dengan position_group_code "AO"
			data = response.EmployeeCMOResponse{}
		} else {
			dataEmployee = respGetEmployeeData.Data[lastIndex]
			if dataEmployee.RealCareerDate == "" {
				// err = errors.New(constant.ERROR_UPSTREAM + " - RealCareerDate Empty")
				return
			}

			parsedTime, err = time.Parse("2006-01-02T15:04:05", dataEmployee.RealCareerDate)
			if err != nil {
				// err = errors.New(constant.ERROR_UPSTREAM + " - Error Parse RealCareerDate")
				return
			}

			layout := constant.FORMAT_DATE

			dataEmployee.RealCareerDate = parsedTime.Format(layout)

			today = time.Now().Format(layout)
			// memvalidasi bulan+tahun yang diberikan tidak lebih besar dari bulan+tahun hari ini
			err = utils.ValidateDiffMonthYear(dataEmployee.RealCareerDate, today)
			if err != nil {
				// err = errors.New(constant.ERROR_UPSTREAM + " - Error Validate MonthYear of RealCareerDate")
				return
			}

			todayDate, err = time.Parse(layout, today)
			if err != nil {
				// err = errors.New(constant.ERROR_UPSTREAM + " - Error Parse todayDate")
				return
			}

			givenDate, err = time.Parse(layout, dataEmployee.RealCareerDate)
			if err != nil {
				// err = errors.New(constant.ERROR_UPSTREAM + " - Error Parse givenDate")
				return
			}

			diffOfMonths := utils.DiffInMonths(todayDate, givenDate)

			cmoJoinedAge, _ := strconv.Atoi(os.Getenv("CMO_JOINED_AGE"))
			var cmoCategory string = constant.CMO_BARU
			if diffOfMonths > cmoJoinedAge {
				cmoCategory = constant.CMO_LAMA
			}

			data = response.EmployeeCMOResponse{
				EmployeeID:         dataEmployee.EmployeeID,
				EmployeeName:       dataEmployee.EmployeeName,
				EmployeeIDWithName: dataEmployee.EmployeeID + " - " + dataEmployee.EmployeeName,
				JoinDate:           dataEmployee.RealCareerDate,
				PositionGroupCode:  dataEmployee.PositionGroupCode,
				PositionGroupName:  dataEmployee.PositionGroupName,
				CMOCategory:        cmoCategory,
			}
		}

	} else {
		// err = errors.New(constant.ERROR_BAD_REQUEST + " - Get Employee Data Error")
		return
	}

	return
}

func (u usecase) GetFpdCMO(ctx context.Context, CmoID string, BPKBNameType string) (data response.FpdCMOResponse, err error) {
	var (
		respGetFPD response.GetFPDCmoByID
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	header := map[string]string{
		"Authorization": middlewares.UserInfoData.AccessToken,
	}

	lobID := constant.LOBID_KMB
	cmoID := CmoID
	endpointURL := fmt.Sprintf(os.Getenv("AGREEMENT_LTV_FPD")+"?lob_id=%d&cmo_id=%s", lobID, cmoID)

	getDataFpd, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, endpointURL, nil, header, constant.METHOD_GET, false, 0, timeout, "", middlewares.UserInfoData.AccessToken)

	if getDataFpd.StatusCode() == 504 || getDataFpd.StatusCode() == 502 {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get FPD Data Timeout")
		return
	}

	if getDataFpd.StatusCode() != 200 && getDataFpd.StatusCode() != 504 && getDataFpd.StatusCode() != 502 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get FPD Data Error")
		return
	}

	if err == nil && getDataFpd.StatusCode() == 200 {
		json.Unmarshal([]byte(jsoniter.Get(getDataFpd.Body()).ToString()), &respGetFPD)

		// Mencari nilai fpd untuk bpkb_name_type "NAMA BEDA"
		var fpdNamaBeda float64 = 0
		var accSalesNamaBeda int = 0
		for _, item := range respGetFPD.Data {
			if item.BpkbNameType == constant.BPKB_NAMA_BEDA {
				fpdNamaBeda = item.Fpd
				accSalesNamaBeda = item.AccSales
				break
			}
		}

		// Mencari nilai fpd untuk bpkb_name_type "NAMA SAMA"
		var fpdNamaSama float64 = 0
		var accSalesNamaSama int = 0
		for _, item := range respGetFPD.Data {
			if item.BpkbNameType == constant.BPKB_NAMA_SAMA {
				fpdNamaSama = item.Fpd
				accSalesNamaSama = item.AccSales
				break
			}
		}

		if fpdNamaBeda <= 0 && accSalesNamaBeda <= 0 && fpdNamaSama <= 0 && accSalesNamaSama <= 0 {
			// Ini pertanda CMO tidak punya SALES SAMA SEKALI,
			// maka nantinya di usecase akan diarahkan ke Cluster Default sesuai jenis BPKBNameType nya,
			// setelah itu dilakukan proses penyimpanan ke table "trx_cmo_no_fpd" sebagai data penampung rentang tanggal kapan hingga kapan CMO_ID tersebut diarahkan sebagai Default Cluster
			data = response.FpdCMOResponse{}
		} else {
			if BPKBNameType == constant.BPKB_NAMA_BEDA {
				data = response.FpdCMOResponse{
					FpdExist:    true,
					CmoFpd:      fpdNamaBeda,
					CmoAccSales: accSalesNamaBeda,
				}
			}

			if BPKBNameType == constant.BPKB_NAMA_SAMA {
				data = response.FpdCMOResponse{
					FpdExist:    true,
					CmoFpd:      fpdNamaSama,
					CmoAccSales: accSalesNamaSama,
				}
			}

		}
	} else {
		// err = errors.New(constant.ERROR_BAD_REQUEST + " - Get FPD Data Error")
		return
	}

	return
}

func (u usecase) CheckCmoNoFPD(prospectID string, cmoID string, cmoCategory string, cmoJoinDate string, defaultCluster string, bpkbName string) (clusterCMOSaved string, entitySaveTrxNoFPd entity.TrxCmoNoFPD, err error) {

	var (
		today     string
		todayTime time.Time
		layout    = constant.FORMAT_DATE
	)

	currentDate := time.Now().Format(layout)

	if cmoCategory == constant.CMO_LAMA {
		today = currentDate
	} else {
		today = cmoJoinDate
	}

	// Cek apakah CMO_ID sudah pernah tersimpan di dalam table `trx_cmo_no_fpd`
	var TrxCmoNoFpd entity.TrxCmoNoFPD

	TrxCmoNoFpd, err = u.repository.CheckCMONoFPD(cmoID, bpkbName)
	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Check CMO No FPD error")
		return
	}

	clusterCMOSaved = "" // init data for default response
	// Jika CMO_ID sudah ada
	if TrxCmoNoFpd.CMOID != "" {
		if currentDate >= TrxCmoNoFpd.DefaultClusterStartDate && currentDate <= TrxCmoNoFpd.DefaultClusterEndDate {
			// CMO_ID sudah ada dan masih di dalam rentang tanggal `DefaultClusterStartDate` dan `DefaultClusterEndDate`
			defaultCluster = TrxCmoNoFpd.DefaultCluster
			clusterCMOSaved = defaultCluster
		} else {
			today = currentDate
		}
	}

	if clusterCMOSaved == "" {
		// Parsing tanggal hari ini ke dalam format time.Time
		todayTime, _ = time.Parse(layout, today)

		// Menambahkan 3 bulan
		defaultClusterMonthsDuration, _ := strconv.Atoi(os.Getenv("DEFAULT_CLUSTER_MONTHS_DURATION"))
		threeMonthsLater := todayTime.AddDate(0, defaultClusterMonthsDuration, 0)
		// Mengambil tanggal terakhir dari bulan tersebut
		threeMonthsLater = time.Date(threeMonthsLater.Year(), threeMonthsLater.Month(), 0, 0, 0, 0, 0, threeMonthsLater.Location())
		// Parsing threeMonthsLater ke dalam format "yyyy-mm-dd" sebagai string
		threeMonthsLaterString := threeMonthsLater.Format(layout)

		entitySaveTrxNoFPd = entity.TrxCmoNoFPD{
			ProspectID:              prospectID,
			BPKBName:                bpkbName,
			CMOID:                   cmoID,
			CmoCategory:             cmoCategory,
			CmoJoinDate:             cmoJoinDate,
			DefaultCluster:          defaultCluster,
			DefaultClusterStartDate: today,
			DefaultClusterEndDate:   threeMonthsLaterString,
			CreatedAt:               time.Time{},
		}
	}

	return
}

func (u multiUsecase) PrinciplePemohon(ctx context.Context, r request.PrinciplePemohon) (resp response.UsecaseApi, err error) {

	var (
		data                response.Filtering
		trxPrincipleStepTwo entity.TrxPrincipleStepTwo
	)

	defer func() {

		var code string

		if data.Code != nil {
			code = data.Code.(string)
		}

		resp = response.UsecaseApi{
			Result: data.Decision,
			Code:   code,
			Reason: data.Reason,
		}

		if err == nil {
			birthDate, _ := time.Parse(constant.FORMAT_DATE, r.BirthDate)
			var spouseBirthDate interface{}
			if r.SpouseBirthDate != "" {
				spouseBirthDate, _ = time.Parse(constant.FORMAT_DATE, r.SpouseBirthDate)
			}

			trxPrincipleStepTwo.ProspectID = r.ProspectID
			trxPrincipleStepTwo.IDNumber = r.IDNumber
			trxPrincipleStepTwo.LegalName = r.LegalName
			trxPrincipleStepTwo.MobilePhone = r.MobilePhone
			trxPrincipleStepTwo.FullName = r.FullName
			trxPrincipleStepTwo.BirthDate = birthDate
			trxPrincipleStepTwo.BirthPlace = r.BirthPlace
			trxPrincipleStepTwo.SurgateMotherName = r.SurgateMotherName
			trxPrincipleStepTwo.Gender = r.Gender
			trxPrincipleStepTwo.SpouseIDNumber = utils.CheckEmptyString(r.SpouseIDNumber)
			trxPrincipleStepTwo.LegalAddress = r.LegalAddress
			trxPrincipleStepTwo.LegalRT = r.LegalRT
			trxPrincipleStepTwo.LegalRW = r.LegalRW
			trxPrincipleStepTwo.LegalProvince = r.LegalProvince
			trxPrincipleStepTwo.LegalCity = r.LegalCity
			trxPrincipleStepTwo.LegalKecamatan = r.LegalKecamatan
			trxPrincipleStepTwo.LegalKelurahan = r.LegalKelurahan
			trxPrincipleStepTwo.LegalZipCode = r.LegalZipCode
			trxPrincipleStepTwo.LegalAreaPhone = r.LegalPhoneArea
			trxPrincipleStepTwo.LegalPhone = r.LegalPhone
			trxPrincipleStepTwo.CompanyAddress = r.CompanyAddress
			trxPrincipleStepTwo.CompanyRT = r.CompanyRT
			trxPrincipleStepTwo.CompanyRW = r.CompanyRW
			trxPrincipleStepTwo.CompanyProvince = r.CompanyProvince
			trxPrincipleStepTwo.CompanyCity = r.CompanyCity
			trxPrincipleStepTwo.CompanyKecamatan = r.CompanyKecamatan
			trxPrincipleStepTwo.CompanyKelurahan = r.CompanyKelurahan
			trxPrincipleStepTwo.CompanyZipCode = r.CompanyZipCode
			trxPrincipleStepTwo.CompanyAreaPhone = r.CompanyPhoneArea
			trxPrincipleStepTwo.CompanyPhone = r.CompanyPhone
			trxPrincipleStepTwo.MonthlyFixedIncome = r.MonthlyFixedIncome
			trxPrincipleStepTwo.MaritalStatus = r.MaritalStatus
			trxPrincipleStepTwo.SpouseIncome = r.SpouseIncome
			trxPrincipleStepTwo.SelfiePhoto = utils.CheckEmptyString(r.SelfiePhoto)
			trxPrincipleStepTwo.KtpPhoto = utils.CheckEmptyString(r.KtpPhoto)
			trxPrincipleStepTwo.SpouseFullName = utils.CheckEmptyString(r.SpouseFullName)
			trxPrincipleStepTwo.SpouseBirthDate = spouseBirthDate
			trxPrincipleStepTwo.SpouseBirthPlace = utils.CheckEmptyString(r.SpouseBirthPlace)
			trxPrincipleStepTwo.SpouseGender = utils.CheckEmptyString(r.SpouseGender)
			trxPrincipleStepTwo.SpouseLegalName = utils.CheckEmptyString(r.SpouseLegalName)
			trxPrincipleStepTwo.SpouseMobilePhone = utils.CheckEmptyString(r.SpouseMobilePhone)
			trxPrincipleStepTwo.SpouseSurgateMotherName = utils.CheckEmptyString(r.SpouseSurgateMotherName)
			trxPrincipleStepTwo.EconomySectorID = r.EconomySectorID
			trxPrincipleStepTwo.Education = r.Education
			trxPrincipleStepTwo.EmploymentSinceMonth = r.EmploymentSinceMonth
			trxPrincipleStepTwo.EmploymentSinceYear = r.EmploymentSinceYear
			trxPrincipleStepTwo.IndustryTypeID = r.IndustryTypeID
			trxPrincipleStepTwo.JobPosition = r.JobPosition
			trxPrincipleStepTwo.JobType = r.JobType
			trxPrincipleStepTwo.ProfessionID = r.ProfessionID
			trxPrincipleStepTwo.Decision = data.Decision
			trxPrincipleStepTwo.Reason = data.Reason

			_ = u.repository.SavePrincipleStepTwo(trxPrincipleStepTwo)
		}
	}()

	draft, _ := u.repository.GetDraftPrinciple(r.ProspectID)

	dupcheckConfig, err := u.repository.GetConfig("dupcheck", constant.LOB_KMB_OFF, "dupcheck_kmb_config")

	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Config Error")
		return
	}

	var configValue response.DupcheckConfig

	json.Unmarshal([]byte(dupcheckConfig.Value), &configValue)

	encrypted, _ := u.repository.GetEncB64(r.IDNumber)

	checkPmkOrDsr, err := u.usecase.BannedPMKOrDSR(encrypted.MyString)
	if err != nil {
		return
	}

	trxPrincipleStepTwo.CheckBannedPMKDSRResult = checkPmkOrDsr.Result
	trxPrincipleStepTwo.CheckBannedPMKDSRCode = checkPmkOrDsr.Code
	trxPrincipleStepTwo.CheckBannedPMKDSRReason = checkPmkOrDsr.Reason

	if checkPmkOrDsr.Result == constant.DECISION_REJECT {
		data.Decision = checkPmkOrDsr.Result
		data.Code = checkPmkOrDsr.Code
		data.Reason = checkPmkOrDsr.Reason
		return
	}

	// Pernah Reject PMK atau DSR atau NIK
	trxReject, _, err := u.usecase.Rejection(r.ProspectID, encrypted.MyString, configValue)
	if err != nil {
		return
	}

	// save banned
	trxPrincipleStepTwo.CheckRejectionResult = trxReject.Result
	trxPrincipleStepTwo.CheckRejectionCode = trxReject.Code
	trxPrincipleStepTwo.CheckRejectionReason = trxReject.Reason

	if trxReject.Result == constant.DECISION_REJECT {

		data.Decision = trxReject.Result
		data.Code = trxReject.Code
		data.Reason = trxReject.Reason

		// trxFMF.TrxBannedPMKDSR = trxBannedPMKDSR
		return
	}

	var (
		customer                  []request.SpouseDupcheck
		dataCustomer              []response.SpDupCekCustomerByID
		blackList                 response.UsecaseApi
		sp                        response.SpDupCekCustomerByID
		isBlacklist               bool
		pefindo                   response.PefindoResult
		reqPefindo                request.Pefindo
		trxDetailBiro             []entity.TrxDetailBiro
		filtering                 response.Filtering
		hrisCMO                   response.EmployeeCMOResponse
		mdmFPD                    response.FpdCMOResponse
		clusterCMO                string
		savedCluster              string
		useDefaultCluster         bool
		entityTransactionCMOnoFPD entity.TrxCmoNoFPD

		cluster = constant.CLUSTER_C
		bpkb    = constant.BPKB_NAMA_BEDA
	)

	income := r.MonthlyFixedIncome + r.SpouseIncome
	save := entity.FilteringKMB{ProspectID: r.ProspectID, RequestID: ctx.Value(echo.HeaderXRequestID).(string), BranchID: draft.BranchID, BpkbName: draft.BPKBName}
	customer = append(customer, request.SpouseDupcheck{IDNumber: r.IDNumber, LegalName: r.LegalName, BirthDate: r.BirthDate, MotherName: r.SurgateMotherName})

	if r.MaritalStatus == constant.MARRIED {
		customer = append(customer, request.SpouseDupcheck{IDNumber: r.SpouseIDNumber, LegalName: r.SpouseLegalName, BirthDate: r.SpouseBirthDate, MotherName: r.SpouseSurgateMotherName})
	}

	for i := 0; i < len(customer); i++ {

		sp, err = u.usecase.DupcheckIntegrator(ctx, r.ProspectID, customer[i].IDNumber, customer[i].LegalName, customer[i].BirthDate, customer[i].MotherName, middlewares.UserInfoData.AccessToken)

		dataCustomer = append(dataCustomer, sp)

		if err != nil {
			return
		}

		blackList, _ = u.usecase.BlacklistCheck(i, sp)

		trxPrincipleStepTwo.CheckBlacklistResult = blackList.Result
		trxPrincipleStepTwo.CheckBlacklistCode = blackList.Code
		trxPrincipleStepTwo.CheckBlacklistReason = blackList.Reason

		if blackList.Result == constant.DECISION_REJECT {

			isBlacklist = true

			data = response.Filtering{ProspectID: r.ProspectID, Code: blackList.Code, Decision: blackList.Result, Reason: blackList.Reason, IsBlacklist: isBlacklist}

			save.Decision = blackList.Result
			save.Reason = blackList.Reason
			save.IsBlacklist = 1

			err = u.usecase.Save(save, trxDetailBiro, entityTransactionCMOnoFPD)

			return
		}
	}

	mainCustomer := dataCustomer[0]

	if mainCustomer.CustomerStatus == "" || mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_NEW {
		mainCustomer.CustomerStatus = constant.STATUS_KONSUMEN_NEW
		mainCustomer.CustomerSegment = constant.RO_AO_REGULAR
	}

	if mainCustomer.CustomerStatusKMB == "" {
		mainCustomer.CustomerStatusKMB = constant.STATUS_KONSUMEN_NEW
	}

	if mainCustomer != (response.SpDupCekCustomerByID{}) {
		mainCustomer.CustomerStatusKMB, err = u.usecase.CustomerKMB(mainCustomer)

		if err != nil {
			return
		}
	}

	mainCustomer.CustomerStatus = mainCustomer.CustomerStatusKMB

	pmk, _ := u.usecase.CheckPMK(draft.BranchID, mainCustomer.CustomerStatusKMB, income, draft.HomeStatus, r.ProfessionID, r.BirthDate, 12, r.MaritalStatus, r.EmploymentSinceYear, r.EmploymentSinceMonth, draft.StaySinceYear, draft.StaySinceMonth)

	trxPrincipleStepTwo.CheckPMKResult = pmk.Result
	trxPrincipleStepTwo.CheckPMKCode = pmk.Code
	trxPrincipleStepTwo.CheckPMKReason = pmk.Reason

	if pmk.Result == constant.DECISION_REJECT {
		data.Decision = pmk.Result
		data.Reason = pmk.Reason

		// belum save
	}

	reqPefindo = request.Pefindo{
		ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
		IDMember:          constant.USER_PBK_KMB_FILTEERING,
		User:              constant.USER_PBK_KMB_FILTEERING,
		ProspectID:        r.ProspectID,
		BranchID:          draft.BranchID,
		IDNumber:          r.IDNumber,
		LegalName:         r.LegalName,
		BirthDate:         r.BirthDate,
		SurgateMotherName: r.SurgateMotherName,
		Gender:            r.Gender,
		BPKBName:          draft.BPKBName,
	}

	if r.MaritalStatus == constant.MARRIED && r.SpouseIDNumber != "" && r.SpouseLegalName != "" && r.SpouseBirthDate != "" && r.SpouseSurgateMotherName != "" {
		reqPefindo.MaritalStatus = constant.MARRIED
		reqPefindo.SpouseIDNumber = r.SpouseIDNumber
		reqPefindo.SpouseLegalName = r.SpouseLegalName
		reqPefindo.SpouseBirthDate = r.SpouseBirthDate
		reqPefindo.SpouseSurgateMotherName = r.SpouseSurgateMotherName
		reqPefindo.SpouseGender = r.SpouseGender
	}

	/* Process Get Cluster based on CMO_ID starts here */

	hrisCMO, err = u.usecase.GetEmployeeData(ctx, draft.CMOID)
	if err != nil {
		return
	}

	if hrisCMO.CMOCategory == "" {
		// err = errors.New(constant.ERROR_BAD_REQUEST + " - CMO_ID " + req.CMOID + " not found on HRIS API")
		return
	}

	bpkbName := strings.Contains(os.Getenv("NAMA_SAMA"), draft.BPKBName)

	if bpkbName {
		bpkb = constant.BPKB_NAMA_SAMA
		cluster = constant.CLUSTER_B
	}

	if hrisCMO.CMOCategory == constant.NEW {
		clusterCMO = cluster
		// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak bulan join_date nya
		useDefaultCluster = true
	} else {
		// Mendapatkan value FPD dari masing-masing jenis BPKB
		mdmFPD, err = u.usecase.GetFpdCMO(ctx, draft.CMOID, bpkb)
		if err != nil {
			return
		}

		if !mdmFPD.FpdExist {
			clusterCMO = cluster
			// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak tanggal hit filtering nya (assume: today)
			useDefaultCluster = true
		} else {
			// Check Cluster
			var mappingFpdCluster entity.MasterMappingFpdCluster
			mappingFpdCluster, err = u.repository.MasterMappingFpdCluster(mdmFPD.CmoFpd)
			if err != nil {
				return
			}

			if mappingFpdCluster.Cluster == "" {
				clusterCMO = cluster
				// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak tanggal hit filtering nya (assume: today)
				useDefaultCluster = true
			} else {
				clusterCMO = mappingFpdCluster.Cluster
			}
		}
	}

	if useDefaultCluster {
		savedCluster, entityTransactionCMOnoFPD, err = u.usecase.CheckCmoNoFPD(r.ProspectID, draft.CMOID, hrisCMO.CMOCategory, hrisCMO.JoinDate, clusterCMO, bpkb)
		if err != nil {
			return
		}
		if savedCluster != "" {
			clusterCMO = savedCluster
		}
	}

	save.CMOID = draft.CMOID
	save.CMOJoinDate = hrisCMO.JoinDate
	save.CMOCategory = hrisCMO.CMOCategory
	save.CMOFPD = mdmFPD.CmoFpd
	save.CMOAccSales = mdmFPD.CmoAccSales
	save.CMOCluster = clusterCMO

	/* Process Get Cluster based on CMO_ID ends here */

	// hit ke pefindo
	filtering, pefindo, trxDetailBiro, err = u.usecase.Pefindo(ctx, reqPefindo, mainCustomer.CustomerStatus, clusterCMO, bpkb)
	if err != nil {
		return
	}

	data = filtering
	data.ClusterCMO = clusterCMO
	data.ProspectID = r.ProspectID
	data.CustomerSegment = mainCustomer.CustomerSegment
	data.CustomerStatusKMB = mainCustomer.CustomerStatusKMB

	save.CustomerStatusKMB = mainCustomer.CustomerStatusKMB
	save.Cluster = filtering.Cluster

	trxPrincipleStepTwo.FilteringResult = filtering.Decision
	trxPrincipleStepTwo.FilteringCode = filtering.Code
	trxPrincipleStepTwo.FilteringReason = filtering.Reason

	primePriority, _ := utils.ItemExists(mainCustomer.CustomerSegment, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

	if primePriority && (mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_AO || mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_RO) {
		data.Code = blackList.Code
		data.Decision = blackList.Result
		data.Reason = mainCustomer.CustomerStatus + " " + mainCustomer.CustomerSegment
		data.NextProcess = true

		save.Cluster = constant.CLUSTER_PRIME_PRIORITY
	}

	// save transaction
	save.Decision = data.Decision
	save.CustomerStatus = mainCustomer.CustomerStatus
	save.CustomerSegment = mainCustomer.CustomerSegment
	save.CustomerID = mainCustomer.CustomerID

	if data.NextProcess {
		save.NextProcess = 1
	}

	// ada data pefindo
	if pefindo.Score != "" && pefindo.Category != nil {
		save.MaxOverdueBiro = pefindo.MaxOverdue
		save.MaxOverdueLast12monthsBiro = pefindo.MaxOverdueLast12Months
		save.ScoreBiro = pefindo.Score

		var isWoContractBiro, isWoWithCollateralBiro int
		if pefindo.WoContract {
			isWoContractBiro = 1
		}
		if pefindo.WoAdaAgunan {
			isWoWithCollateralBiro = 1
		}
		save.IsWoContractBiro = isWoContractBiro
		save.IsWoWithCollateralBiro = isWoWithCollateralBiro

		save.TotalInstallmentAmountBiro = pefindo.AngsuranAktifPbk
		save.TotalBakiDebetNonCollateralBiro = pefindo.TotalBakiDebetNonAgunan
		save.Category = pefindo.Category

		if pefindo.MaxOverdueKORules != nil {
			save.MaxOverdueKORules = pefindo.MaxOverdueKORules
		}
		if pefindo.MaxOverdueLast12MonthsKORules != nil {
			save.MaxOverdueLast12MonthsKORules = pefindo.MaxOverdueLast12MonthsKORules
		}
	}

	save.Reason = filtering.Reason

	var cbFound bool

	if pefindo.Score != "" && pefindo.Score != constant.DECISION_PBK_NO_HIT && pefindo.Score != constant.PEFINDO_UNSCORE {
		cbFound = true
	}

	reqMetricsEkyc := request.MetricsEkyc{
		CBFound:         cbFound,
		CustomerStatus:  mainCustomer.CustomerStatusKMB,
		CustomerSegment: mainCustomer.CustomerSegment,
	}

	r.BpkbName = draft.BPKBName

	dukcapil, err := u.usecase.Dukcapil(ctx, r, reqMetricsEkyc, middlewares.UserInfoData.AccessToken)

	if err != nil && err.Error() != fmt.Sprintf("%s - Dukcapil", constant.TYPE_CONTINGENCY) {
		return
	}

	if err != nil && err.Error() == fmt.Sprintf("%s - Dukcapil", constant.TYPE_CONTINGENCY) {

		asliri, err := u.usecase.Asliri(ctx, r, middlewares.UserInfoData.AccessToken)

		if err != nil {

			ktp, err := u.usecase.Ktp(ctx, r, reqMetricsEkyc, middlewares.UserInfoData.AccessToken)
			if err != nil {
				return response.UsecaseApi{}, err
			}

			// trxFMF.EkycSource = "KTP VALIDATOR"
			// trxFMF.EkycSimiliarity = data.Similiarity
			// trxFMF.EkycReason = data.Reason

			trxPrincipleStepTwo.CheckEkycResult = ktp.Result
			trxPrincipleStepTwo.CheckEkycCode = ktp.Code
			trxPrincipleStepTwo.CheckEkycReason = ktp.Reason
		}

		// trxFMF.EkycSource = "ASLI RI"
		// trxFMF.EkycSimiliarity = data.Similiarity
		// trxFMF.EkycReason = data.Reason

		trxPrincipleStepTwo.CheckEkycResult = asliri.Result
		trxPrincipleStepTwo.CheckEkycCode = asliri.Code
		trxPrincipleStepTwo.CheckEkycReason = asliri.Reason
	}

	// trxFMF.EkycSource = "DUKCAPIL"
	// trxFMF.EkycSimiliarity = data.Similiarity
	// trxFMF.EkycReason = data.Reason

	trxPrincipleStepTwo.CheckEkycResult = dukcapil.Result
	trxPrincipleStepTwo.CheckEkycCode = dukcapil.Code
	trxPrincipleStepTwo.CheckEkycReason = dukcapil.Reason

	err = u.usecase.Save(save, trxDetailBiro, entityTransactionCMOnoFPD)

	return

}

func (u usecase) Save(transaction entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, transactionCMOnoFPD entity.TrxCmoNoFPD) (err error) {

	err = u.repository.SaveFiltering(transaction, trxDetailBiro, transactionCMOnoFPD)

	if err != nil {

		if strings.Contains(err.Error(), "deadline") {
			// err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Save Filtering Timeout")
			return
		}

		// err = errors.New(constant.ERROR_BAD_REQUEST + " - Save Filtering Error ProspectID Already Exist")
	}

	return
}
