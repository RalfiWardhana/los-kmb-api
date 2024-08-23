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
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/mitchellh/mapstructure"
)

func (u multiUsecase) PrinciplePembiayaan(ctx context.Context, r request.PrinciplePembiayaan, accessToken string) (resp response.UsecaseApi, err error) {

	var (
		filteringKMB               entity.FilteringKMB
		principleStepOne           entity.TrxPrincipleStepOne
		principleStepTwo           entity.TrxPrincipleStepTwo
		config                     entity.AppConfig
		cmoCluster                 string
		configValue                response.DupcheckConfig
		married                    bool
		customer                   []request.SpouseDupcheck
		sp                         response.SpDupCekCustomerByID
		dataCustomer               []response.SpDupCekCustomerByID
		agereementChassisNumber    response.AgreementChassisNumber
		installmentTopup           float64
		installmentAmountFMF       float64
		installmentAmountSpouseFMF float64
		customerData               []request.CustomerData
		monthlyVariableIncome      float64
		spouseIncome               float64
		income                     float64
		wg                         sync.WaitGroup
		errChan                    = make(chan error, 4)
		trxPrincipleStepThree      entity.TrxPrincipleStepThree
	)

	defer func() {
		if err == nil {
			if r.MonthlyVariableIncome != nil {
				monthlyVariableIncome = *r.MonthlyVariableIncome
			}

			trxPrincipleStepThree.ProspectID = r.ProspectID
			trxPrincipleStepThree.IDNumber = principleStepTwo.IDNumber
			trxPrincipleStepThree.Tenor = r.Tenor
			trxPrincipleStepThree.AF = r.AF
			trxPrincipleStepThree.NTF = r.NTF
			trxPrincipleStepThree.OTR = r.OTR
			trxPrincipleStepThree.DPAmount = r.DPAmount
			trxPrincipleStepThree.AdminFee = r.AdminFee
			trxPrincipleStepThree.InstallmentAmount = r.InstallmentAmount
			trxPrincipleStepThree.Dealer = r.Dealer
			trxPrincipleStepThree.MonthlyVariableIncome = monthlyVariableIncome
			trxPrincipleStepThree.Decision = resp.Result
			trxPrincipleStepThree.Reason = resp.Reason

			_ = u.repository.SavePrincipleStepThree(trxPrincipleStepThree)
		}
	}()

	wg.Add(4)
	go func() {
		defer wg.Done()
		principleStepOne, err = u.repository.GetPrincipleStepOne(r.ProspectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		principleStepTwo, err = u.repository.GetPrincipleStepTwo(r.ProspectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		filteringKMB, err = u.repository.GetFilteringResult(r.ProspectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		config, err = u.repository.GetConfig("dupcheck", "KMB-OFF", "dupcheck_kmb_config")

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return response.UsecaseApi{}, err
	}

	json.Unmarshal([]byte(config.Value), &configValue)

	// get mapping cluster
	mappingCluster := entity.MasterMappingCluster{
		BranchID:       principleStepOne.BranchID,
		CustomerStatus: filteringKMB.CustomerStatus.(string),
	}
	if strings.Contains(os.Getenv("NAMA_SAMA"), principleStepOne.BPKBName) {
		mappingCluster.BpkbNameType = 1
	}
	if strings.Contains(constant.STATUS_KONSUMEN_RO_AO, filteringKMB.CustomerStatus.(string)) {
		mappingCluster.CustomerStatus = "AO/RO"
	}

	mappingCluster, err = u.repository.MasterMappingCluster(mappingCluster)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error")
		return
	}

	if clusterName, ok := filteringKMB.CMOCluster.(string); ok {
		cmoCluster = clusterName
	} else {
		cmoCluster = mappingCluster.Cluster
	}

	// Cek Tahun Kendaraan
	ageVehicle, err := u.usecase.VehicleCheck(principleStepOne.ManufactureYear, cmoCluster, principleStepOne.BPKBName, r.Tenor, configValue, filteringKMB, r.AF)
	if err != nil {
		return
	}

	trxPrincipleStepThree.CheckVehicleResult = ageVehicle.Result
	trxPrincipleStepThree.CheckVehicleCode = ageVehicle.Code
	trxPrincipleStepThree.CheckVehicleReason = ageVehicle.Reason

	if ageVehicle.Result == constant.DECISION_REJECT {
		resp = ageVehicle
		return
	}

	// Pengecekan Tenor 36
	if r.Tenor >= 36 {
		var trxTenor response.UsecaseApi
		if r.Tenor == 36 {
			trxTenor, err = u.usecase.RejectTenor36(cmoCluster)
			if err != nil {
				return
			}
		} else if r.Tenor > 36 {
			trxTenor = response.UsecaseApi{
				Code:   constant.CODE_REJECT_TENOR,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_TENOR,
			}
		}

		trxPrincipleStepThree.CheckRejectTenor36Result = trxTenor.Result
		trxPrincipleStepThree.CheckRejectTenor36Code = trxTenor.Code
		trxPrincipleStepThree.CheckRejectTenor36Reason = trxTenor.Reason

		if trxTenor.Result == constant.DECISION_REJECT {
			resp = trxTenor
			return
		}
	}

	// Scorepro
	if principleStepTwo.MaritalStatus == constant.MARRIED {
		married = true
	}

	birthDate := principleStepTwo.BirthDate.Format(constant.FORMAT_DATE)
	var spouseBirthDate string
	if date, ok := principleStepTwo.SpouseBirthDate.(time.Time); ok {
		spouseBirthDate = date.Format(constant.FORMAT_DATE)
	}

	customer = append(customer, request.SpouseDupcheck{IDNumber: principleStepTwo.IDNumber, LegalName: principleStepTwo.LegalName, BirthDate: birthDate, MotherName: principleStepTwo.SurgateMotherName})
	if married {
		customer = append(customer, request.SpouseDupcheck{IDNumber: principleStepTwo.SpouseIDNumber.(string), LegalName: principleStepTwo.SpouseLegalName.(string), BirthDate: spouseBirthDate, MotherName: principleStepTwo.SpouseSurgateMotherName.(string)})
	}

	for i := 0; i < len(customer); i++ {
		sp, err = u.usecase.DupcheckIntegrator(ctx, r.ProspectID, customer[i].IDNumber, customer[i].LegalName, customer[i].BirthDate, customer[i].MotherName, accessToken)

		if err != nil {
			return
		}

		dataCustomer = append(dataCustomer, sp)
	}
	mainCustomer := dataCustomer[0]

	agereementChassisNumber, err = u.usecase.AgreementChassisNumberIntegrator(ctx, principleStepOne.NoChassis, r.ProspectID, accessToken)
	if err != nil {
		return
	}

	if agereementChassisNumber != (response.AgreementChassisNumber{}) && agereementChassisNumber.InstallmentAmount > 0 {
		installmentTopup = agereementChassisNumber.InstallmentAmount
	}

	var scoreBiro string
	if filteringKMB.ScoreBiro != nil {
		scoreBiro = filteringKMB.ScoreBiro.(string)
	}

	var customerStatus string
	if filteringKMB.CustomerStatus == nil {
		customerStatus = constant.STATUS_KONSUMEN_NEW
	} else {
		customerStatus = filteringKMB.CustomerStatus.(string)
	}

	var customerSegment string
	if filteringKMB.CustomerSegment == nil {
		customerSegment = constant.RO_AO_REGULAR
	} else {
		customerSegment = filteringKMB.CustomerSegment.(string)
	}

	metricsScs, err := u.usecase.Scorepro(ctx, r, principleStepOne, principleStepTwo, scoreBiro, customerStatus, customerSegment, installmentTopup, mainCustomer, accessToken)
	if err != nil {
		return
	}

	trxPrincipleStepThree.ScoreProResult = metricsScs.Result
	trxPrincipleStepThree.ScoreProCode = metricsScs.Code
	trxPrincipleStepThree.ScoreProReason = metricsScs.Reason

	if metricsScs.Result == constant.DECISION_REJECT {
		resp.Result = metricsScs.Result
		resp.Code = metricsScs.Code
		resp.Reason = metricsScs.Reason
		resp.SourceDecision = metricsScs.Source
		resp.Info = metricsScs.Info
		return
	}

	// Check DSR
	customerData = append(customerData, request.CustomerData{
		TransactionID:   r.ProspectID,
		StatusKonsumen:  customerStatus,
		CustomerSegment: customerSegment,
		IDNumber:        principleStepTwo.IDNumber,
		LegalName:       principleStepTwo.LegalName,
		BirthDate:       birthDate,
		MotherName:      principleStepTwo.SurgateMotherName,
	})

	installmentAmountFMF = dataCustomer[0].TotalInstallment

	if married {
		customerData = append(customerData, request.CustomerData{
			TransactionID: r.ProspectID,
			IDNumber:      principleStepTwo.SpouseIDNumber.(string),
			LegalName:     principleStepTwo.SpouseLegalName.(string),
			BirthDate:     spouseBirthDate,
			MotherName:    principleStepTwo.SurgateMotherName,
		})

		installmentAmountSpouseFMF = dataCustomer[1].TotalInstallment
	}

	if r.MonthlyVariableIncome != nil {
		monthlyVariableIncome = *r.MonthlyVariableIncome
	}
	if principleStepTwo.SpouseIncome != nil {
		spouseIncome = principleStepTwo.SpouseIncome.(float64)
	}

	income = principleStepTwo.MonthlyFixedIncome + monthlyVariableIncome + spouseIncome

	dsr, err := u.usecase.DsrCheck(ctx, r, customerData, r.InstallmentAmount, installmentAmountFMF, installmentAmountSpouseFMF, income, agereementChassisNumber, accessToken, configValue)
	if err != nil {
		return
	}

	trxPrincipleStepThree.CheckDSRResult = dsr.Result
	trxPrincipleStepThree.CheckDSRCode = dsr.Code
	trxPrincipleStepThree.CheckDSRReason = dsr.Reason

	resp = dsr

	return
}

func (u usecase) VehicleCheck(manufactureYear, cmoCluster, bkpbName string, tenor int, configValue response.DupcheckConfig, filtering entity.FilteringKMB, af float64) (data response.UsecaseApi, err error) {

	data.SourceDecision = constant.SOURCE_DECISION_PMK

	currentYear, _ := strconv.Atoi(time.Now().Format("2006-01-02")[0:4])
	BPKBYear, _ := strconv.Atoi(manufactureYear)

	ageVehicle := currentYear - BPKBYear

	ageVehicle += int(tenor / 12)

	bpkbNameType := 0
	if strings.Contains(os.Getenv("NAMA_SAMA"), bkpbName) {
		bpkbNameType = 1
	}

	resultPefindo := checkResultPefindo(filtering)

	detailInfo := map[string]interface{}{
		"vehicle_age":    ageVehicle,
		"cluster":        cmoCluster,
		"bpkb_name_type": bpkbNameType,
		"tenor":          tenor,
		"af":             af,
		"result_pbk":     resultPefindo,
	}

	if ageVehicle <= configValue.Data.VehicleAge {

		mapping, err := u.repository.GetMappingVehicleAge(ageVehicle, cmoCluster, bpkbNameType, tenor, resultPefindo, af)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Vehicle Age Error")
			return data, err
		}

		detailInfo["info"] = mapping.Info
		info, _ := json.Marshal(detailInfo)

		if mapping.Decision == constant.DECISION_REJECT {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_VEHICLE_AGE_MAX
			data.Reason = fmt.Sprintf("%s Ketentuan", constant.REASON_VEHICLE_AGE_MAX)
			data.Info = string(info)
			return data, nil
		}

		data.Result = constant.DECISION_PASS
		data.Code = constant.CODE_VEHICLE_SESUAI
		data.Reason = constant.REASON_VEHICLE_SESUAI
		data.Info = string(info)
		return data, nil

	} else {

		detailInfo["info"] = constant.INFO_VEHICLE_AGE
		info, _ := json.Marshal(detailInfo)

		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_VEHICLE_AGE_MAX
		data.Reason = fmt.Sprintf("%s %d Tahun", constant.REASON_VEHICLE_AGE_MAX, configValue.Data.VehicleAge)
		data.Info = string(info)
		return
	}

}

func checkResultPefindo(filtering entity.FilteringKMB) (resultPefindo string) {
	// check hit pefindo
	if filtering.ScoreBiro != nil {
		if filtering.ScoreBiro.(string) != "" && filtering.ScoreBiro.(string) != constant.DECISION_PBK_NO_HIT && filtering.ScoreBiro.(string) != constant.PEFINDO_UNSCORE {
			// use ovd pefindo all
			maxOverdueLast12Months, _ := utils.GetFloat(filtering.MaxOverdueLast12monthsBiro)
			maxOverdueDays, _ := utils.GetFloat(filtering.MaxOverdueBiro)

			// pass or reject
			if maxOverdueLast12Months > constant.PBK_OVD_LAST_12 {
				resultPefindo = constant.DECISION_REJECT
			} else if maxOverdueDays > constant.PBK_OVD_CURRENT {
				resultPefindo = constant.DECISION_REJECT
			} else {
				resultPefindo = constant.DECISION_PASS
			}
		}
	} else {
		resultPefindo = constant.NO_HIT_PBK
	}

	return resultPefindo
}

func (u usecase) RejectTenor36(cluster string) (result response.UsecaseApi, err error) {

	var (
		pass        bool
		configValue map[string][]string
	)

	config, err := u.repository.GetConfig("tenor36", "KMB-OFF", "exclusion_tenor36")
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetConfig exclusion_tenor36 Error")
		return
	}

	json.Unmarshal([]byte(config.Value), &configValue)

	for _, v := range configValue["data"] {
		if v == cluster {
			pass = true
			break
		}
	}

	if pass {
		result.Code = constant.CODE_PASS_TENOR
		result.Result = constant.DECISION_PASS
		result.Reason = constant.REASON_PASS_TENOR
	} else {
		result.Code = constant.CODE_REJECT_TENOR
		result.Result = constant.DECISION_REJECT
		result.Reason = constant.REASON_REJECT_TENOR
	}

	return
}

func (u usecase) Scorepro(ctx context.Context, req request.PrinciplePembiayaan, principleStepOne entity.TrxPrincipleStepOne, principleStepTwo entity.TrxPrincipleStepTwo, pefindoScore, customerStatus, customerSegment string, installmentTopUp float64, spDupcheck response.SpDupCekCustomerByID, accessToken string) (data response.ScorePro, err error) {

	var (
		residenceZipCode              string
		firstDigitsOfResidenceZipCode string
		scoreGenerator                entity.ScoreGenerator
		trxDetailBiro                 []entity.TrxDetailBiro
		pefindoIDX                    request.PefindoIDX
		reqScoreproIntegrator         request.ScoreProIntegrator
		responseScs                   response.IntegratorScorePro
		respPefindoIDX                response.PefindoIDX
	)

	// DEFAULT
	scoreGenerator = entity.ScoreGenerator{
		Key:               os.Getenv("SCOREPRO_DEFAULT_KEY"),
		ScoreGeneratorsID: os.Getenv("SCOREPRO_DEFAULT_SCORE_GENERATOR_ID"),
	}

	residenceZipCode = principleStepOne.ResidenceZipCode
	firstDigitsOfResidenceZipCode = string(residenceZipCode[0])

	if customerStatus == constant.STATUS_KONSUMEN_NEW {
		scoreGenerator, err = u.repository.GetScoreGenerator(firstDigitsOfResidenceZipCode)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetScoreGenerator Scorepro Error")
			return
		}
	} else {
		scoreGenerator, err = u.repository.GetScoreGeneratorROAO()
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetScoreGeneratorROAO Scorepro Error")
			return
		}
	}

	trxDetailBiro, err = u.repository.GetTrxDetailBIro(req.ProspectID)
	if err != nil {
		return
	}

	// Get Pefindo IDX
	pefindoIDX = request.PefindoIDX{
		ProspectID: req.ProspectID,
		ModelType:  scoreGenerator.Key,
	}

	for _, v := range trxDetailBiro {
		if v.Score != "" && v.Score != constant.DECISION_PBK_NO_HIT && v.Score != constant.PEFINDO_UNSCORE {
			if v.Subject == constant.CUSTOMER {
				pefindoIDX.CBFoundCustomer = true
				pefindoIDX.PefindoIDCustomer = v.BiroID
			}
			if v.Subject == constant.SPOUSE {
				pefindoIDX.CBFoundSpouse = true
				pefindoIDX.PefindoIDSpouse = v.BiroID
			}
		}
	}

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	paramPefindoIDX, _ := json.Marshal(pefindoIDX)
	resp, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("PEFINDO_IDX_URL"), paramPefindoIDX, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.ProspectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Pefindo IDX Error")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Pefindo IDX Error")
		return
	}

	err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &respPefindoIDX)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal Pefindo IDX Error")
		return
	}

	var cbFound bool
	if pefindoIDX.CBFoundCustomer || pefindoIDX.CBFoundSpouse {
		cbFound = true
	}

	reqScoreproIntegrator = request.ScoreProIntegrator{
		ProspectID:       req.ProspectID,
		CBFound:          cbFound,
		StatusKonsumen:   customerStatus,
		RequestorID:      os.Getenv("SCOREPRO_REQUESTID"),
		Journey:          constant.JOURNEY_SCOREPRO,
		PhoneNumber:      principleStepTwo.MobilePhone,
		ScoreGeneratorID: scoreGenerator.ScoreGeneratorsID,
	}

	intZipcode, _ := strconv.Atoi(residenceZipCode[0:2])
	ntfOtr := math.Floor(req.NTF/req.OTR*100) / 100

	currentYear, _ := strconv.Atoi(time.Now().Format("2006-01-02")[0:4])

	bpkbKey := principleStepOne.BPKBName
	if bpkbKey == "KK" {
		bpkbKey = "O"
	}

	if scoreGenerator.Key == os.Getenv("KEY_SCOREPRO_IDX_2W_JABOJABAR") {

		BPKBYear, _ := strconv.Atoi(principleStepOne.ManufactureYear)
		ageVehicle := currentYear - BPKBYear

		reqScoreproIntegrator.Data = map[string]interface{}{
			"bpkb_name":      bpkbKey,
			"worst_24mth":    respPefindoIDX.Worst24Mth,
			"gender":         principleStepTwo.Gender,
			"marital_status": principleStepTwo.MaritalStatus,
			"ntf_otr":        ntfOtr,
			"zip_code":       intZipcode,
			"tenor":          req.Tenor,
			"vehicle_age":    ageVehicle,
			"profession_id":  principleStepTwo.ProfessionID,
		}

	} else if scoreGenerator.Key == os.Getenv("KEY_SCOREPRO_IDX_2W_OTHERS") {

		employmentSinceYear, _ := strconv.Atoi(principleStepTwo.EmploymentSinceYear)
		employmentSinceYear = currentYear - employmentSinceYear

		reqScoreproIntegrator.Data = map[string]interface{}{
			"bpkb_name":        bpkbKey,
			"ntf_otr":          ntfOtr,
			"zip_code":         intZipcode,
			"worst_12mth_auto": respPefindoIDX.Worst12MthAuto,
			"gender":           principleStepTwo.Gender,
			"marital_status":   principleStepTwo.MaritalStatus,
			"tenor":            req.Tenor,
			"length_of_empl":   employmentSinceYear,
			"home_status":      principleStepOne.HomeStatus,
		}

	} else if scoreGenerator.Key == os.Getenv("KEY_SCOREPRO_IDX_2W_AORO") {

		birthDate := principleStepTwo.BirthDate.Format(constant.FORMAT_DATE)

		location, _ := time.LoadLocation("Asia/Jakarta")
		layout := "2006-01-02"
		convTime, _ := time.ParseInLocation(layout, birthDate, location)

		currentTime := time.Now()
		dateToday := currentTime.Format(layout)
		todayTime, _ := time.ParseInLocation(layout, dateToday, location)

		age := utils.HumanAgeCalculator(convTime, todayTime)

		var activeLoanTypeLast6M string
		getActiveLoanTypeLast6M, err := u.repository.GetActiveLoanTypeLast6M(spDupcheck.CustomerID.(string))
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetActiveLoanTypeLast6M Scorepro Error")
			return data, err
		}

		if strings.Replace(getActiveLoanTypeLast6M.ActiveLoanTypeLast6M, " ", "", -1) == ";;" {
			getActiveLoanTypeLast24M, err := u.repository.GetActiveLoanTypeLast24M(spDupcheck.CustomerID.(string))
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - GetActiveLoanTypeLast24M Scorepro Error")
				return data, err
			}

			if getActiveLoanTypeLast24M.AgreementNo != "" {
				activeLoanTypeLast6M = "999"
			} else {
				activeLoanTypeLast6M = "9999"
			}
		} else {
			if getActiveLoanTypeLast6M.ActiveLoanTypeLast6M == "" {
				activeLoanTypeLast6M = "9999"
			} else {
				activeLoanTypeLast6M = getActiveLoanTypeLast6M.ActiveLoanTypeLast6M
			}
		}

		var moblast int
		getMoblast, err := u.repository.GetMoblast(spDupcheck.CustomerID.(string))
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetMoblast Scorepro Error")
			return data, err
		}

		if getMoblast.Moblast == "" {
			moblast = 9999
		} else {
			intMob, _ := strconv.Atoi(getMoblast.Moblast)
			if intMob > 24 {
				moblast = 9999
			} else {
				moblast = intMob
			}
		}

		reqScoreproIntegrator.Data = map[string]interface{}{
			"zip_code":                intZipcode,
			"ntf_otr":                 ntfOtr,
			"bpkb_name":               bpkbKey,
			"worst_24mth_auto":        respPefindoIDX.Worst24MthAuto,
			"age":                     age,
			"active_loan_type_last6m": activeLoanTypeLast6M,
			"nom03_12mth_all":         respPefindoIDX.Nom0312MntAll,
			"moblast":                 moblast,
			"gender":                  principleStepTwo.Gender,
			"marital_status":          principleStepTwo.MaritalStatus,
		}

	}

	paramScorepro, _ := json.Marshal(reqScoreproIntegrator)

	resp, _ = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("SCOREPRO_IDX_URL"), paramScorepro, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.ProspectID, accessToken)

	if resp.StatusCode() != 200 {

		responseScs = response.IntegratorScorePro{
			ProspectID:  req.ProspectID,
			Result:      constant.DECISION_PASS,
			ScoreResult: constant.SCOREPRO_RESULT_MEDIUM_2ND,
		}

	} else {

		json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &responseScs)

	}

	info, _ := json.Marshal(responseScs)

	// HANDLING RESPONSE TSI
	if responseScs.IsTsi && responseScs.Status != "" {
		if strings.Contains(responseScs.Status, "TSH") {
			responseScs.ScoreResult = "HIGH"
		} else if strings.Contains(responseScs.Status, "TSL") {
			responseScs.ScoreResult = "LOW"
		} else {
			responseScs.ScoreResult = "MEDIUM"
		}

		if responseScs.Score != nil {
			score := responseScs.Score.(string)
			responseScs.Score = score[:strings.IndexByte(score, '-')]
		}

	}

	// Handling ASS-SCORE
	if strings.Contains(responseScs.Status, "ASS-") {
		segmen, _ := strconv.Atoi(responseScs.Segmen)
		segmenAssScore, _ := strconv.Atoi(os.Getenv("SCOREPRO_SEGMEN_ASS_SCORE"))

		if segmen > 0 && segmen <= segmenAssScore {
			responseScs.Result = constant.DECISION_REJECT
		}
	}

	// PRIME PRIORITY
	if customerSegment == constant.RO_AO_PRIME || customerSegment == constant.RO_AO_PRIORITY {
		if customerStatus == constant.STATUS_KONSUMEN_AO && installmentTopUp == 0 && spDupcheck.NumberOfPaidInstallment != nil && *spDupcheck.NumberOfPaidInstallment >= 6 {
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
			data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
			data.Source = constant.SOURCE_DECISION_SCOREPRO
			data.Info = string(utils.SafeEncoding(info))
			return
		}

		if customerStatus == constant.STATUS_KONSUMEN_RO || (installmentTopUp > 0 && spDupcheck.MaxOverdueDaysforActiveAgreement != nil && *spDupcheck.MaxOverdueDaysforActiveAgreement <= 30) {
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
			data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
			data.Source = constant.SOURCE_DECISION_SCOREPRO
			data.Info = string(utils.SafeEncoding(info))
			return
		}
	}

	if cbFound {
		if customerStatus == constant.STATUS_KONSUMEN_RO || customerStatus == constant.STATUS_KONSUMEN_AO {
			if strings.Contains(os.Getenv("NAMA_SAMA"), principleStepOne.BPKBName) {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(utils.SafeEncoding(info))
			} else {
				if responseScs.ScoreResult == "LOW" {
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(utils.SafeEncoding(info))
				} else {
					data.Result = constant.DECISION_PASS
					data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(utils.SafeEncoding(info))
				}
			}
		} else {
			if strings.Contains(os.Getenv("NAMA_SAMA"), principleStepOne.BPKBName) {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(utils.SafeEncoding(info))
			} else {
				if strings.ToUpper(pefindoScore) == "VERY HIGH RISK" {
					segmenReject := map[string]bool{"2": true, "3": true, "4": true, "5": true}
					segmenPass := map[string]bool{"6": true, "7": true, "8": true, "9": true, "10": true, "11": true, "12": true}

					if _, ok := segmenReject[responseScs.Segmen]; ok {
						data.Result = constant.DECISION_REJECT
						data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(utils.SafeEncoding(info))
					} else if _, ok := segmenPass[responseScs.Segmen]; ok {
						data.Result = constant.DECISION_PASS
						data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(utils.SafeEncoding(info))
					} else {
						if responseScs.ScoreResult == "LOW" {
							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
							data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
							data.Source = constant.SOURCE_DECISION_SCOREPRO
							data.Info = string(utils.SafeEncoding(info))
						} else {
							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
							data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
							data.Source = constant.SOURCE_DECISION_SCOREPRO
							data.Info = string(utils.SafeEncoding(info))
						}
					}
				} else {
					if responseScs.ScoreResult == "LOW" {
						data.Result = constant.DECISION_REJECT
						data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(utils.SafeEncoding(info))
					} else {
						data.Result = constant.DECISION_PASS
						data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(utils.SafeEncoding(info))
					}
				}
			}
		}
	} else {
		if responseScs.IsTsi {
			if strings.Contains(os.Getenv("NAMA_SAMA"), principleStepOne.BPKBName) {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(utils.SafeEncoding(info))
			} else {
				if responseScs.Result == constant.DECISION_PASS {
					data.Result = constant.DECISION_PASS
					data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(utils.SafeEncoding(info))
				} else {
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(utils.SafeEncoding(info))
				}
			}
		} else {
			if responseScs.Result == constant.DECISION_PASS {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(utils.SafeEncoding(info))
			} else {
				if strings.Contains(os.Getenv("NAMA_SAMA"), principleStepOne.BPKBName) {
					// Handling ASS-SCORE
					segmen, _ := strconv.Atoi(responseScs.Segmen)
					segmenAssScore, _ := strconv.Atoi(os.Getenv("SCOREPRO_SEGMEN_ASS_SCORE"))
					if strings.Contains(responseScs.Status, "ASS-") && segmen > 0 && segmen <= segmenAssScore {
						data.Result = constant.DECISION_REJECT
						data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(utils.SafeEncoding(info))
					} else {
						data.Result = constant.DECISION_PASS
						data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(utils.SafeEncoding(info))
					}
				} else {
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(utils.SafeEncoding(info))
				}
			}
		}
	}
	return
}

func (u usecase) DsrCheck(ctx context.Context, req request.PrinciplePembiayaan, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, agreementChasisNumber response.AgreementChassisNumber, accessToken string, configValue response.DupcheckConfig) (data response.UsecaseApi, err error) {

	var (
		dsr, installmentOther, installmentOtherSpouse, installmentTopup float64
		instOther                                                       response.InstallmentOther
		dsrDetails                                                      response.DsrDetails
		reasonCustomerStatus                                            string
		result                                                          response.Dsr
	)

	reasonMaxDsr := "Threshold"

	konsumen := customerData[0]

	if konsumen.CustomerSegment == constant.RO_AO_PRIME || konsumen.CustomerSegment == constant.RO_AO_PRIORITY {
		reasonCustomerStatus = konsumen.StatusKonsumen + " " + konsumen.CustomerSegment
	} else {
		reasonCustomerStatus = konsumen.StatusKonsumen
	}

	header := map[string]string{}

	for i, customer := range customerData {

		jsonCustomer, _ := json.Marshal(customer)
		var installmentLOS *resty.Response

		installmentLOS, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("INSTALLMENT_PENDING_URL"), jsonCustomer, header, constant.METHOD_POST, true, 2, 60, req.ProspectID, accessToken)

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
			data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
			data.SourceDecision = constant.SOURCE_DECISION_DSR

			_ = mapstructure.Decode(data, &result)
			return
		}

	} else {

		var installment = installmentConfins

		if installmentConfins > 0 {

			if agreementChasisNumber != (response.AgreementChassisNumber{}) && agreementChasisNumber.InstallmentAmount > 0 {
				reasonCustomerStatus = reasonCustomerStatus + " " + constant.TOP_UP

				installmentTopup = agreementChasisNumber.InstallmentAmount
				installment = installmentConfins - installmentTopup

				var pencairan float64
				pencairan = req.OTR - req.DPAmount
				if req.Dealer == constant.DEALER_PSA {
					pencairan -= req.AdminFee
				}

				if pencairan <= 0 {
					err = errors.New(constant.ERROR_UPSTREAM + " - Perhitungan OTR - DP harus lebih dari 0")
					return
				}

				totalOutstanding := agreementChasisNumber.OutstandingPrincipal + agreementChasisNumber.OutstandingInterest + agreementChasisNumber.LcInstallment
				minimumPencairan := ((pencairan - totalOutstanding) / pencairan) * 100

				dsrDetails.DetailTopUP = response.DetailTopUP{
					Pencairan:              pencairan,
					AgreementChassisNumber: agreementChasisNumber,
					MinimumPencairan:       minimumPencairan,
					TotalOutstanding:       totalOutstanding,
				}

				result.Details = dsrDetails

				var configMinimumPencairanROTopUp float64

				if konsumen.CustomerSegment == constant.RO_AO_PRIME {
					configMinimumPencairanROTopUp = configValue.Data.MinimumPencairanROTopUp.Prime
				} else if konsumen.CustomerSegment == constant.RO_AO_PRIORITY {
					configMinimumPencairanROTopUp = configValue.Data.MinimumPencairanROTopUp.Priority
				} else {
					configMinimumPencairanROTopUp = configValue.Data.MinimumPencairanROTopUp.Regular
				}

				if minimumPencairan < configMinimumPencairanROTopUp {

					dsr = ((installmentAmount + (installment + installmentConfinsSpouse) + (installmentOther + installmentOtherSpouse)) / income) * 100
					data.Dsr = dsr
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_PENCAIRAN_TOPUP
					data.Reason = fmt.Sprintf("%s %s", reasonCustomerStatus, constant.REASON_PENCAIRAN_TOPUP)

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
				data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
				data.SourceDecision = constant.SOURCE_DECISION_DSR

				_ = mapstructure.Decode(data, &result)
				return
			}
		} else if konsumen.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
			if konsumen.CustomerSegment == constant.RO_AO_PRIME && installmentTopup > 0 {
				// go next
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_DSRLTE35
				data.SourceDecision = constant.SOURCE_DECISION_DSR
				data.Reason = fmt.Sprintf("%s", reasonCustomerStatus)

				_ = mapstructure.Decode(data, &result)
				return
			} else if dsr > configValue.Data.MaxDsr {
				data.Result = constant.DECISION_REJECT
				data.Code = constant.CODE_DSRGT35
				data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRGT35, reasonMaxDsr)
				data.SourceDecision = constant.SOURCE_DECISION_DSR

				_ = mapstructure.Decode(data, &result)
				return
			}
		}

	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_DSRLTE35
	data.Reason = fmt.Sprintf("%s %s %s", reasonCustomerStatus, constant.REASON_DSRLTE35, reasonMaxDsr)
	data.SourceDecision = constant.SOURCE_DECISION_DSR

	_ = mapstructure.Decode(data, &result)
	return
}

func (u usecase) AgreementChassisNumberIntegrator(ctx context.Context, prospectID, chassisNumber string, accessToken string) (data response.AgreementChassisNumber, err error) {

	var hitChassisNumber *resty.Response

	hitChassisNumber, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+chassisNumber, nil, map[string]string{}, constant.METHOD_GET, true, 2, 60, prospectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - DsrCheck Call Get Agreement of Chassis Number Timeout")
		return
	}

	if hitChassisNumber.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - DsrCheck Call Get Agreement of Chassis Number Error")
		return
	}

	err = json.Unmarshal([]byte(jsoniter.Get(hitChassisNumber.Body(), "data").ToString()), &data)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - DsrCheck Unmarshal Get Agreement of Chassis Number Error")
		return
	}

	return
}
