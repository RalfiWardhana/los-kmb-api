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

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

func (u multiUsecase) Submission2Wilen(ctx context.Context, req request.Submission2Wilen, accessToken string) (resp response.Submission2Wilen, err error) {

	var trxKPM entity.TrxKPM

	// Check Banned Chassis Number
	bannedChassisNumber, err := u.usecase.CheckBannedChassisNumber(req.NoChassis)
	if err != nil {
		return
	}

	trxKPM.CheckNokaNosinCode = bannedChassisNumber.Code
	trxKPM.CheckNokaNosinResult = bannedChassisNumber.Result
	trxKPM.CheckNokaNosinReason = bannedChassisNumber.Reason

	if bannedChassisNumber.Result == constant.DECISION_REJECT {
		resp.Code = bannedChassisNumber.Code
		resp.Result = constant.DECISION_KPM_REJECT
		resp.Reason = bannedChassisNumber.Reason
		return
	}

	// Check Chassis Number with Active Aggrement
	agereementChassisNumber, checkChassisNumber, err := u.usecase.CheckAgreementChassisNumber(ctx, req.ProspectID, req.NoChassis, req.IDNumber, req.SpouseIDNumber, accessToken)
	if err != nil {
		return
	}

	trxKPM.CheckNokaNosinCode = checkChassisNumber.Code
	trxKPM.CheckNokaNosinResult = checkChassisNumber.Result
	trxKPM.CheckNokaNosinReason = checkChassisNumber.Reason

	if checkChassisNumber.Result == constant.DECISION_REJECT {
		resp.Code = checkChassisNumber.Code
		resp.Result = constant.DECISION_KPM_REJECT
		resp.Reason = checkChassisNumber.Reason
		return
	}

	// Check Blacklist
	var (
		customer                  []request.SpouseDupcheck
		sp                        response.SpDupCekCustomerByID
		dataCustomer              []response.SpDupCekCustomerByID
		spMap                     response.SpDupcheckMap
		dupcheckData              response.SpDupcheckMap
		trxDetailBiro             []entity.TrxDetailBiro
		entityTransactionCMOnoFPD entity.TrxCmoNoFPD
		married                   bool
		blackList                 response.UsecaseApi
		customerType              string
	)

	income := req.MonthlyFixedIncome + req.SpouseIncome
	save := entity.FilteringKMB{ProspectID: req.ProspectID, RequestID: ctx.Value(echo.HeaderXRequestID).(string), BranchID: req.BranchID, BpkbName: req.BPKBNameType}
	customer = append(customer, request.SpouseDupcheck{IDNumber: req.IDNumber, LegalName: req.LegalName, BirthDate: req.BirthDate, MotherName: req.SurgateMotherName})

	if req.MaritalStatus == constant.MARRIED {
		married = true
	}

	if married {
		customer = append(customer, request.SpouseDupcheck{IDNumber: req.SpouseIDNumber, LegalName: req.SpouseLegalName, BirthDate: req.SpouseBirthDate, MotherName: req.SpouseSurgateMotherName})
	}

	for i := 0; i < len(customer); i++ {

		sp, err = u.usecase.DupcheckIntegrator(ctx, req.ProspectID, customer[i].IDNumber, customer[i].LegalName, customer[i].BirthDate, customer[i].MotherName, middlewares.UserInfoData.AccessToken)

		dataCustomer = append(dataCustomer, sp)

		if err != nil {
			return
		}

		blackList, customerType = u.usecase.BlacklistCheck(i, sp)

		if i == 0 {
			spMap.CustomerType = customerType
		} else if i == 1 {
			spMap.SpouseType = customerType
		}

		trxKPM.CheckBlacklistResult = blackList.Result
		trxKPM.CheckBlacklistCode = blackList.Code
		trxKPM.CheckBlacklistReason = blackList.Reason

		if blackList.Result == constant.DECISION_REJECT {

			save.Decision = blackList.Result
			save.Reason = blackList.Reason
			save.IsBlacklist = 1

			dupcheckData.CustomerType = spMap.CustomerType
			dupcheckData.SpouseType = spMap.SpouseType

			err = u.usecase.Save(save, trxDetailBiro, entityTransactionCMOnoFPD)
			if err != nil {
				return
			}

			resp.Code = blackList.Code
			resp.Result = constant.DECISION_KPM_REJECT
			resp.Reason = blackList.Reason

			return
		}
	}

	// Check APU PPT

	// Check PMK
	dupcheckData.CustomerType = spMap.CustomerType
	dupcheckData.SpouseType = spMap.SpouseType

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

	dupcheckData.StatusKonsumen = mainCustomer.CustomerStatusKMB

	if mainCustomer.MaxOverdueDaysROAO != nil {
		dupcheckData.MaxOverdueDaysROAO = *mainCustomer.MaxOverdueDaysROAO
	}

	if mainCustomer.NumberOfPaidInstallment != nil {
		dupcheckData.NumberOfPaidInstallment = *mainCustomer.NumberOfPaidInstallment
	}

	if mainCustomer.MaxOverdueDaysforActiveAgreement != nil {
		dupcheckData.MaxOverdueDaysforActiveAgreement = *mainCustomer.MaxOverdueDaysforActiveAgreement
	}

	dupcheckData.CustomerID = mainCustomer.CustomerID

	dupcheckData.OSInstallmentDue = mainCustomer.OSInstallmentDue
	dupcheckData.NumberofAgreement = mainCustomer.NumberofAgreement

	if mainCustomer.CustomerStatusKMB == constant.STATUS_KONSUMEN_AO || mainCustomer.CustomerStatusKMB == constant.STATUS_KONSUMEN_RO {
		if dupcheckData.NumberofAgreement == 0 {
			dupcheckData.AgreementStatus = constant.AGREEMENT_LUNAS
		} else {
			dupcheckData.AgreementStatus = constant.AGREEMENT_AKTIF
		}
	}

	mainCustomer.CustomerStatus = mainCustomer.CustomerStatusKMB

	pmk, err := u.usecase.CheckPMK(req.BranchID, mainCustomer.CustomerStatusKMB, income, req.HomeStatus, req.ProfessionID, req.BirthDate, req.Tenor, req.MaritalStatus, req.EmploymentSinceYear, req.EmploymentSinceMonth, req.StaySinceYear, req.StaySinceMonth)
	if err != nil {
		return
	}

	trxKPM.CheckPMKResult = pmk.Result
	trxKPM.CheckPMKCode = pmk.Code
	trxKPM.CheckPMKReason = pmk.Reason

	if pmk.Result == constant.DECISION_REJECT {
		resp.Code = pmk.Code
		resp.Result = constant.DECISION_KPM_REJECT
		resp.Reason = pmk.Reason

		return
	}

	// Chekc Pefindo
	var (
		reqPefindo            request.Pefindo
		spouseGender          string
		respCmoBranch         response.MDMMasterMappingBranchEmployeeResponse
		cmoID                 string
		hrisCMO               response.EmployeeCMOResponse
		mdmFPD                response.FpdCMOResponse
		clusterCMO            string
		savedCluster          string
		useDefaultCluster     bool
		respRrdDate           string
		monthsDiff            int
		expiredContractConfig entity.AppConfig
		filtering             response.Filtering
		pefindo               response.PefindoResult
		data                  response.Filtering

		cluster = constant.CLUSTER_C
		bpkb    = constant.BPKB_NAMA_BEDA
	)

	if req.SpouseIDNumber != "" {
		if req.Gender == "M" {
			spouseGender = "F"
		} else {
			spouseGender = "M"
		}
	}

	dupcheckData.InstallmentAmountFMF = dataCustomer[0].TotalInstallment
	if married {
		dupcheckData.InstallmentAmountSpouseFMF = dataCustomer[1].TotalInstallment
	}

	reqPefindo = request.Pefindo{
		ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
		IDMember:          constant.USER_PBK_KMB_FILTEERING,
		User:              constant.USER_PBK_KMB_FILTEERING,
		ProspectID:        req.ProspectID,
		BranchID:          req.BranchID,
		IDNumber:          req.IDNumber,
		LegalName:         req.LegalName,
		BirthDate:         req.BirthDate,
		SurgateMotherName: req.SurgateMotherName,
		Gender:            req.Gender,
		BPKBName:          req.BPKBNameType,
	}

	if married && req.SpouseIDNumber != "" && req.SpouseLegalName != "" && req.SpouseBirthDate != "" && req.SpouseSurgateMotherName != "" {
		reqPefindo.MaritalStatus = constant.MARRIED
		reqPefindo.SpouseIDNumber = req.SpouseIDNumber
		reqPefindo.SpouseLegalName = req.SpouseLegalName
		reqPefindo.SpouseBirthDate = req.SpouseBirthDate
		reqPefindo.SpouseSurgateMotherName = req.SpouseSurgateMotherName
		reqPefindo.SpouseGender = spouseGender
	}

	respCmoBranch, err = u.usecase.MDMGetMasterMappingBranchEmployee(ctx, req.ProspectID, req.BranchID, accessToken)
	if err != nil {
		return
	}

	if len(respCmoBranch.Data) > 0 {
		cmoID = respCmoBranch.Data[0].CMOID
	} else {
		err = errors.New(constant.ERROR_UPSTREAM + " - CMO Dedicated Not Found")
		return
	}

	/* Process Get Cluster based on CMO_ID starts here */
	hrisCMO, err = u.usecase.GetEmployeeData(ctx, cmoID)
	if err != nil {
		return
	}

	if hrisCMO.CMOCategory == "" {
		err = errors.New(constant.ERROR_UPSTREAM + " - CMO Not Found")
		return
	}

	bpkbName := strings.Contains(os.Getenv("NAMA_SAMA"), req.BPKBNameType)

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
		mdmFPD, err = u.usecase.GetFpdCMO(ctx, cmoID, bpkb)
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
		savedCluster, entityTransactionCMOnoFPD, err = u.usecase.CheckCmoNoFPD(req.ProspectID, cmoID, hrisCMO.CMOCategory, hrisCMO.JoinDate, clusterCMO, bpkb)
		if err != nil {
			return
		}
		if savedCluster != "" {
			clusterCMO = savedCluster
		}
	}

	save.CMOID = cmoID
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
	data.ProspectID = req.ProspectID
	data.CustomerSegment = mainCustomer.CustomerSegment
	data.CustomerStatusKMB = mainCustomer.CustomerStatusKMB

	save.CustomerStatusKMB = mainCustomer.CustomerStatusKMB
	save.Cluster = filtering.Cluster

	dupcheckData.Cluster = filtering.Cluster

	primePriority, _ := utils.ItemExists(mainCustomer.CustomerSegment, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

	if primePriority && (mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_AO || mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_RO) {
		data.Code = blackList.Code
		data.Decision = constant.DECISION_PASS
		data.Reason = mainCustomer.CustomerStatus + " " + mainCustomer.CustomerSegment
		data.NextProcess = true

		save.Cluster = constant.CLUSTER_PRIME_PRIORITY

		if mainCustomer.CustomerStatus == constant.STATUS_KONSUMEN_RO {
			if (mainCustomer.CustomerID != nil) && (mainCustomer.CustomerID.(string) != "") {
				respRrdDate, monthsDiff, err = u.usecase.CheckLatestPaidInstallment(ctx, req.ProspectID, mainCustomer.CustomerID.(string), middlewares.UserInfoData.AccessToken)
				if err != nil {
					return
				}

				// Get config expired_contract
				expiredContractConfig, err = u.repository.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
				if err != nil {
					err = errors.New(constant.ERROR_UPSTREAM + " - Get Expired Contract Config Error")
					return
				}

				var configValueExpContract response.ExpiredContractConfig
				json.Unmarshal([]byte(expiredContractConfig.Value), &configValueExpContract)

				if configValueExpContract.Data.ExpiredContractCheckEnabled && !(monthsDiff <= configValueExpContract.Data.ExpiredContractMaxMonths) {
					// Jalur mirip seperti customer segment "REGULAR"
					data.Code = filtering.Code
					data.Decision = filtering.Decision
					data.Reason = constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + filtering.Reason
					data.NextProcess = filtering.NextProcess
				}
			} else {
				err = errors.New(constant.ERROR_BAD_REQUEST + " - Customer RO then CustomerID should not be empty")
				return
			}
		}
	}

	// save transaction
	save.Decision = data.Decision
	save.CustomerStatus = mainCustomer.CustomerStatus
	save.CustomerSegment = mainCustomer.CustomerSegment
	save.CustomerID = mainCustomer.CustomerID

	if respRrdDate == "" {
		save.RrdDate = nil
	} else {
		save.RrdDate = respRrdDate
	}

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

	reqDukcapil := request.PrinciplePemohon{
		ProspectID:        req.ProspectID,
		LegalAddress:      req.LegalAddress,
		BirthDate:         req.BirthDate,
		BirthPlace:        req.BirthPlace,
		Gender:            req.Gender,
		MaritalStatus:     req.MaritalStatus,
		IDNumber:          req.IDNumber,
		LegalCity:         req.LegalCity,
		LegalKecamatan:    req.LegalKecamatan,
		LegalKelurahan:    req.LegalKelurahan,
		LegalName:         req.LegalName,
		ProfessionID:      req.ProfessionID,
		LegalProvince:     req.LegalProvince,
		LegalRT:           req.LegalRT,
		LegalRW:           req.LegalRW,
		SurgateMotherName: req.SurgateMotherName,
		BpkbName:          req.BPKBNameType,
		SelfiePhoto:       req.SelfiePhoto,
		KtpPhoto:          req.KtpPhoto,
	}

	dukcapil, err := u.usecase.Dukcapil(ctx, reqDukcapil, reqMetricsEkyc, middlewares.UserInfoData.AccessToken)

	if err != nil && err.Error() != fmt.Sprintf("%s - Dukcapil", constant.TYPE_CONTINGENCY) {
		return
	}

	trxKPM.CheckEkycResult = dukcapil.Result
	trxKPM.CheckEkycCode = dukcapil.Code
	trxKPM.CheckEkycReason = dukcapil.Reason
	trxKPM.CheckEkycSource = dukcapil.Source
	trxKPM.CheckEkycInfo = dukcapil.Info
	trxKPM.CheckEkycSimiliarity = dukcapil.Similiarity

	if err != nil && err.Error() == fmt.Sprintf("%s - Dukcapil", constant.TYPE_CONTINGENCY) {

		asliri, errAsliri := u.usecase.Asliri(ctx, reqDukcapil, middlewares.UserInfoData.AccessToken)
		err = errAsliri

		if err != nil {

			ktp, errKtp := u.usecase.Ktp(ctx, reqDukcapil, reqMetricsEkyc, middlewares.UserInfoData.AccessToken)
			err = errKtp

			if err != nil {
				return response.Submission2Wilen{}, err
			}

			trxKPM.CheckEkycResult = ktp.Result
			trxKPM.CheckEkycCode = ktp.Code
			trxKPM.CheckEkycReason = ktp.Reason
			trxKPM.CheckEkycSource = ktp.Source
			trxKPM.CheckEkycInfo = ktp.Info
			trxKPM.CheckEkycSimiliarity = ktp.Similiarity

		} else {

			trxKPM.CheckEkycResult = asliri.Result
			trxKPM.CheckEkycCode = asliri.Code
			trxKPM.CheckEkycReason = asliri.Reason
			trxKPM.CheckEkycSource = asliri.Source
			trxKPM.CheckEkycInfo = asliri.Info
			trxKPM.CheckEkycSimiliarity = asliri.Similiarity

		}
	}

	if trxKPM.CheckEkycResult != nil && trxKPM.CheckEkycResult == constant.DECISION_REJECT {
		resp.Result = constant.DECISION_KPM_REJECT
		resp.Code = trxKPM.CheckEkycCode.(string)

		if trxKPM.CheckEkycReason != nil {
			resp.Reason = trxKPM.CheckEkycReason.(string)
		}

		return
	}

	err = u.usecase.Save(save, trxDetailBiro, entityTransactionCMOnoFPD)
	if err != nil {
		return
	}

	trxKPM.FilteringResult = filtering.Decision
	trxKPM.FilteringCode = filtering.Code
	trxKPM.FilteringReason = filtering.Reason

	if !data.NextProcess {
		trxKPM.FilteringResult = constant.DECISION_REJECT

		resp.Result = constant.DECISION_KPM_REJECT
		resp.Code = filtering.Code.(string)
		resp.Reason = filtering.Reason

		return
	}

	var OverrideFlowLikeRegular bool

	resultPefindo := save.Decision

	if save.CustomerSegment != nil && strings.Contains("PRIME PRIORITY", save.CustomerSegment.(string)) {
		cluster = constant.CLUSTER_PRIME_PRIORITY

		// Cek apakah customer RO PRIME/PRIORITY ini termasuk jalur `expired_contract tidak <= 6 bulan`
		if save.CustomerStatus == constant.STATUS_KONSUMEN_RO {
			if mainCustomer.RRDDate == nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty")
				return
			}

			RrdDateString := mainCustomer.RRDDate.(string)
			CreatedAtString := time.Now().Format(time.RFC3339)

			var RrdDate time.Time
			RrdDate, err = time.Parse(time.RFC3339, RrdDateString)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of RrdDate (" + RrdDateString + ")")
				return
			}

			var CreatedAt time.Time
			CreatedAt, err = time.Parse(time.RFC3339, CreatedAtString)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of CreatedAt (" + CreatedAtString + ")")
				return
			}

			var MonthsOfExpiredContract int
			MonthsOfExpiredContract, err = utils.PreciseMonthsDifference(RrdDate, CreatedAt)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Difference of months RrdDate and CreatedAt is negative (-)")
				return
			}

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
	}

	if (save.CustomerSegment != nil && !strings.Contains("PRIME PRIORITY", save.CustomerSegment.(string))) || (OverrideFlowLikeRegular) {
		if save.ScoreBiro == nil || save.ScoreBiro == "" || save.ScoreBiro == constant.UNSCORE_PBK {
			resultPefindo = constant.DECISION_PBK_NO_HIT
		} else if save.MaxOverdueBiro != nil || save.MaxOverdueLast12monthsBiro != nil {
			// KO Rules Include All
			ovdCurrent, _ := save.MaxOverdueBiro.(int64)
			ovd12, _ := save.MaxOverdueLast12monthsBiro.(int64)

			if ovdCurrent > constant.PBK_OVD_CURRENT || ovd12 > constant.PBK_OVD_LAST_12 {
				resultPefindo = constant.DECISION_REJECT
			}
		}
	}

	// Check Hasil Pefindo
	var (
		assetMP response.AssetYearList
		otr     float64
	)

	assetMP, err = u.usecase.MDMGetAssetYear(ctx, req.BranchID, req.AssetCode, req.ManufactureYear, req.ProspectID, accessToken)
	if err != nil {
		return
	}

	if len(assetMP.Records) > 0 {
		otr = float64(assetMP.Records[0].MarketPriceValue)
	}

	// get loan amount
	var mappingElaborateLTV []entity.MappingElaborateLTV
	mappingElaborateLTV, err = u.repository.GetMappingElaborateLTV(resultPefindo, clusterCMO)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error")
		return
	}

	ltv, err := u.usecase.GetLTV(mappingElaborateLTV, resultPefindo, req.BPKBNameType, req.ManufactureYear, req.Tenor, pefindo.TotalBakiDebetNonAgunan)
	if err != nil {
		return
	}

	payloadMaxLoan := request.ReqMarsevLoanAmount{
		BranchID:      req.BranchID,
		OTR:           otr,
		MaxLTV:        ltv,
		IsRecalculate: false,
	}

	marsevLoanAmountRes, err := u.usecase.MarsevGetLoanAmount(ctx, payloadMaxLoan, req.ProspectID, accessToken)
	if err != nil {
		return
	}

	if req.LoanAmount > marsevLoanAmountRes.Data.LoanAmountMaximum {
		resp.Result = constant.DECISION_KPM_READJUST
		resp.Code = constant.READJUST_LOAN_AMOUNT_CODE_2WILEN
		resp.Reason = constant.READJUST_LOAN_AMOUNT_REASON_2WILEN

		context := constant.READJUST_LOAN_AMOUNT_CONTEXT_2WILEN
		resp.ReadjustContext = &context

		return
	}

	// Check Scorepro
	var customerStatus string
	if save.CustomerStatus == nil {
		customerStatus = constant.STATUS_KONSUMEN_NEW
	} else {
		customerStatus = save.CustomerStatus.(string)
	}

	var customerSegment string
	if save.CustomerSegment == nil {
		customerSegment = constant.RO_AO_REGULAR
	} else {
		customerSegment = save.CustomerSegment.(string)
	}

	// get asset
	assetList, err := u.usecase.MDMGetMasterAsset(ctx, req.BranchID, req.AssetCode, req.ProspectID, accessToken)
	if err != nil {
		return
	}

	var categoryId, brand string
	if len(assetList.Records) > 0 {
		categoryId = assetList.Records[0].CategoryID
		brand = assetList.Records[0].Brand
	}

	// get marketing program
	bpkbStatusCode := "DN"
	if strings.Contains(os.Getenv("NAMA_SAMA"), req.BPKBNameType) {
		bpkbStatusCode = "SN"
	}

	customerTypeMarsev := utils.CapitalizeEachWord(customerStatus)
	if customerStatus != constant.STATUS_KONSUMEN_NEW {
		customerTypeMarsev = constant.STATUS_KONSUMEN_RO_AO + " " + utils.CapitalizeEachWord(customerSegment)
		if customerSegment == constant.RO_AO_REGULAR {
			customerTypeMarsev = constant.STATUS_KONSUMEN_RO_AO + " Standard"
		}
	}

	manufactureYear, _ := strconv.Atoi(req.ManufactureYear)

	financeType := "PM"

	payloadFilterProgram := request.ReqMarsevFilterProgram{
		Page:                   1,
		Limit:                  10,
		BranchID:               req.BranchID,
		FinancingTypeCode:      financeType,
		CustomerOccupationCode: req.ProfessionID,
		BpkbStatusCode:         bpkbStatusCode,
		SourceApplication:      constant.MARSEV_SOURCE_APPLICATION_KPM,
		CustomerType:           customerTypeMarsev,
		AssetUsageTypeCode:     req.AssetUsageTypeCode,
		AssetCategory:          categoryId,
		AssetBrand:             brand,
		AssetYear:              manufactureYear,
		LoanAmount:             req.LoanAmount,
		SalesMethodID:          5,
	}

	marsevFilterProgramRes, err := u.usecase.MarsevGetMarketingProgram(ctx, payloadFilterProgram, req.ProspectID, accessToken)
	if err != nil {
		return
	}

	if len(marsevFilterProgramRes.Data) == 0 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Marsev Filter Program Not Found")
		return
	}

	var mdmMasterMappingLicensePlateRes response.MDMMasterMappingLicensePlateResponse
	mdmMasterMappingLicensePlateRes, err = u.usecase.MDMGetMappingLicensePlate(ctx, req.LicensePlate, req.ProspectID, accessToken)
	if err != nil {
		return
	}

	mappingLicensePlate := mdmMasterMappingLicensePlateRes.Data.Records[0]

	payloadCalculate := request.ReqMarsevCalculateInstallment{
		ProgramID:              marsevFilterProgramRes.Data[0].ID,
		BranchID:               req.BranchID,
		CustomerOccupationCode: req.ProfessionID,
		AssetUsageTypeCode:     req.AssetUsageTypeCode,
		AssetYear:              manufactureYear,
		BpkbStatusCode:         bpkbStatusCode,
		LoanAmount:             req.LoanAmount,
		Otr:                    otr,
		RegionCode:             mappingLicensePlate.AreaID,
		AssetCategory:          categoryId,
		CustomerBirthDate:      req.BirthDate,
		Tenor:                  req.Tenor,
	}

	var marsevCalculateInstallmentRes response.MarsevCalculateInstallmentResponse
	marsevCalculateInstallmentRes, err = u.usecase.MarsevCalculateInstallment(ctx, payloadCalculate, req.ProspectID, accessToken)
	if err != nil {
		return
	}

	if len(marsevCalculateInstallmentRes.Data) == 0 {
		resp.Result = constant.DECISION_KPM_READJUST
		resp.Code = constant.READJUST_TENOR_CODE_2WILEN
		resp.Reason = constant.READJUST_TENOR_REASON_2WILEN

		context := constant.READJUST_TENOR_CONTEXT_2WILEN
		resp.ReadjustContext = &context

		return
	}

	reqScp := request.PrinciplePembiayaan{
		ProspectID: req.ProspectID,
		NTF:        marsevCalculateInstallmentRes.Data[0].NTF,
		OTR:        otr,
		Tenor:      req.Tenor,
	}

	reqOneScp := entity.TrxPrincipleStepOne{
		ResidenceZipCode: req.ResidenceZipCode,
		BPKBName:         req.BPKBNameType,
		ManufactureYear:  req.ManufactureYear,
		HomeStatus:       req.HomeStatus,
	}

	birthDate, _ := time.Parse("2006-01-02", req.BirthDate)
	reqTwoScp := entity.TrxPrincipleStepTwo{
		MobilePhone:         req.MobilePhone,
		Gender:              req.Gender,
		MaritalStatus:       req.MaritalStatus,
		ProfessionID:        req.ProfessionID,
		EmploymentSinceYear: req.EmploymentSinceYear,
		BirthDate:           birthDate,
	}

	var scoreBiro string
	if save.ScoreBiro != nil {
		scoreBiro = save.ScoreBiro.(string)
	}

	var installmentTopup float64
	if agereementChassisNumber != (response.AgreementChassisNumber{}) && agereementChassisNumber.InstallmentAmount > 0 {
		installmentTopup = agereementChassisNumber.InstallmentAmount
	}

	responseScs, metricsScs, _, err := u.usecase.Scorepro(ctx, reqScp, reqOneScp, reqTwoScp, scoreBiro, customerStatus, customerSegment, installmentTopup, mainCustomer, save, accessToken)
	if err != nil {
		return
	}

	trxKPM.ScoreProResult = metricsScs.Result
	trxKPM.ScoreProCode = metricsScs.Code
	trxKPM.ScoreProReason = metricsScs.Reason
	trxKPM.ScoreProInfo = metricsScs.Info
	trxKPM.ScoreProScoreResult = responseScs.ScoreResult

	if metricsScs.Result == constant.DECISION_REJECT {
		resp.Result = constant.DECISION_KPM_REJECT
		resp.Code = metricsScs.Code
		resp.Reason = metricsScs.Reason

		return
	}

	// Check DSR
	var (
		customerData               []request.CustomerData
		installmentAmountFMF       float64
		installmentAmountSpouseFMF float64
		configValue                response.DupcheckConfig
	)

	config, err := u.repository.GetConfig("dupcheck", "KMB-OFF", "dupcheck_kmb_config")
	if err != nil {
		return
	}

	json.Unmarshal([]byte(config.Value), &configValue)

	customerData = append(customerData, request.CustomerData{
		TransactionID:   req.ProspectID,
		StatusKonsumen:  customerStatus,
		CustomerSegment: customerSegment,
		IDNumber:        req.IDNumber,
		LegalName:       req.LegalName,
		BirthDate:       req.BirthDate,
		MotherName:      req.SurgateMotherName,
	})

	installmentAmountFMF = dataCustomer[0].TotalInstallment

	if married {
		customerData = append(customerData, request.CustomerData{
			TransactionID: req.ProspectID,
			IDNumber:      req.SpouseIDNumber,
			LegalName:     req.SpouseLegalName,
			BirthDate:     req.SpouseBirthDate,
			MotherName:    req.SurgateMotherName,
		})

		installmentAmountSpouseFMF = dataCustomer[1].TotalInstallment
	}

	dealer := "NON PSA"
	if marsevLoanAmountRes.Data.IsPsa {
		dealer = "PSA"
	}

	income = req.MonthlyFixedIncome + 0 + req.SpouseIncome

	reqScp.DPAmount = marsevCalculateInstallmentRes.Data[0].DPAmount
	reqScp.Dealer = dealer
	reqScp.AdminFee = marsevCalculateInstallmentRes.Data[0].AdminFee

	dsr, mappingDSR, instOther, instOtherSpouse, instTopup, err := u.usecase.DsrCheck(ctx, reqScp, customerData, req.InstallmentAmount, installmentAmountFMF, installmentAmountSpouseFMF, income, agereementChassisNumber, accessToken, configValue)
	if err != nil {
		return
	}

	decodedData := utils.SafeDecoding(trxKPM.DupcheckData)
	err = json.Unmarshal([]byte(decodedData), &dupcheckData)
	if err != nil {
		return
	}

	dupcheckData.InstallmentAmountFMF = installmentAmountFMF
	dupcheckData.InstallmentAmountSpouseFMF = installmentAmountSpouseFMF
	dupcheckData.InstallmentAmountOther = instOther
	dupcheckData.InstallmentAmountOtherSpouse = instOtherSpouse
	dupcheckData.InstallmentTopup = instTopup
	dupcheckData.Dsr = dsr.Dsr
	dupcheckData.DetailsDSR = mappingDSR.Details
	dupcheckData.ConfigMaxDSR = configValue.Data.MaxDsr

	trxKPM.CheckDSRResult = dsr.Result
	trxKPM.CheckDSRCode = dsr.Code
	trxKPM.CheckDSRReason = dsr.Reason

	if dsr.Result == constant.DECISION_REJECT {
		resp.Result = constant.DECISION_KPM_REJECT
		resp.Code = dsr.Code
		resp.Reason = dsr.Reason

		return
	}

	var totalInstallmentPBK float64
	if save.TotalInstallmentAmountBiro != nil {
		totalInstallmentPBK, err = utils.GetFloat(save.TotalInstallmentAmountBiro)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetFloat TotalInstallmentAmountBiro Error")
			return
		}
	}

	metricsTotalDsrFmfPbk, trxFMFTotalDsrFmfPbk, err := u.usecase.TotalDsrFmfPbk(ctx, income, req.InstallmentAmount, totalInstallmentPBK, req.ProspectID, customerSegment, accessToken, dupcheckData, configValue, save)
	if err != nil {
		return
	}

	infoTotalDSR, _ := json.Marshal(map[string]interface{}{
		"dsr_fmf":                   dupcheckData.Dsr,
		"dsr_pbk":                   trxFMFTotalDsrFmfPbk.DSRPBK,
		"total_dsr":                 trxFMFTotalDsrFmfPbk.TotalDSR,
		"installment_threshold":     trxFMFTotalDsrFmfPbk.InstallmentThreshold,
		"latest_installment_amount": trxFMFTotalDsrFmfPbk.LatestInstallmentAmount,
	})

	trxKPM.CheckDSRFMFPBKResult = metricsTotalDsrFmfPbk.Result
	trxKPM.CheckDSRFMFPBKCode = metricsTotalDsrFmfPbk.Code
	trxKPM.CheckDSRFMFPBKReason = metricsTotalDsrFmfPbk.Reason
	trxKPM.CheckDSRFMFPBKInfo = string(utils.SafeEncoding(infoTotalDSR))

	resp.Result = metricsTotalDsrFmfPbk.Result
	resp.Code = metricsTotalDsrFmfPbk.Code
	resp.Reason = metricsTotalDsrFmfPbk.Reason

	if dsr.Result == constant.DECISION_REJECT {
		resp.Result = constant.DECISION_KPM_REJECT
		resp.Code = metricsTotalDsrFmfPbk.Code
		resp.Reason = metricsTotalDsrFmfPbk.Reason

		return
	}

	// Cek Usia Kendaraan
	mappingCluster := entity.MasterMappingCluster{
		BranchID:       req.BranchID,
		CustomerStatus: save.CustomerStatus.(string),
	}
	if strings.Contains(os.Getenv("NAMA_SAMA"), req.BPKBNameType) {
		mappingCluster.BpkbNameType = 1
	}
	if strings.Contains(constant.STATUS_KONSUMEN_RO_AO, save.CustomerStatus.(string)) {
		mappingCluster.CustomerStatus = "AO/RO"
	}

	mappingCluster, err = u.repository.MasterMappingCluster(mappingCluster)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error")
		return
	}

	var cmoCluster string
	if clusterName, ok := save.CMOCluster.(string); ok {
		cmoCluster = clusterName
	} else {
		cmoCluster = mappingCluster.Cluster
	}

	ageVehicle, err := u.usecase.VehicleCheck(req.ManufactureYear, cmoCluster, req.BPKBNameType, req.Tenor, configValue, save, marsevLoanAmountRes.Data.AmountOfFinance)
	if err != nil {
		return
	}

	trxKPM.CheckVehicleResult = ageVehicle.Result
	trxKPM.CheckVehicleCode = ageVehicle.Code
	trxKPM.CheckVehicleReason = ageVehicle.Reason
	trxKPM.CheckVehicleInfo = ageVehicle.Info

	resp.Result = constant.DECISION_KPM_APPROVE
	resp.Code = ageVehicle.Code
	resp.Reason = ageVehicle.Reason
	if ageVehicle.Result == constant.DECISION_REJECT {
		resp.Result = constant.DECISION_KPM_REJECT
	}

	// Submit to Sally

	return
}

func (u usecase) CheckAgreementChassisNumber(ctx context.Context, prospectID, chassisNumber, idNumber, spouseIDNumber string, accessToken string) (responseAgreementChassisNumber response.AgreementChassisNumber, data response.UsecaseApi, err error) {

	var (
		hitChassisNumber *resty.Response
	)

	hitChassisNumber, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+chassisNumber, nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, prospectID, accessToken)
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
		return responseAgreementChassisNumber, data, err
	}

	if responseAgreementChassisNumber.IsRegistered && responseAgreementChassisNumber.IsActive && len(responseAgreementChassisNumber.IDNumber) > 0 {
		listNikKonsumenDanPasangan := make([]string, 0)

		listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, idNumber)
		if spouseIDNumber != "" {
			listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, spouseIDNumber)
		}

		if !utils.Contains(listNikKonsumenDanPasangan, responseAgreementChassisNumber.IDNumber) {
			data.Code = constant.CODE_REJECT_CHASSIS_NUMBER
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_CHASSIS_NUMBER
		} else {
			if responseAgreementChassisNumber.IDNumber == idNumber {
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
