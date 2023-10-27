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
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) Scorepro(ctx context.Context, req request.Metrics, pefindoScore, customerSegment string, spDupcheck response.SpDupcheckMap, accessToken string) (responseScs response.IntegratorScorePro, data response.ScorePro, respPefindoIDX response.PefindoIDX, err error) {

	var (
		residenceZipCode              string
		firstDigitsOfResidenceZipCode string
		scoreGenerator                entity.ScoreGenerator
		trxDetailBiro                 []entity.TrxDetailBiro
		pefindoIDX                    request.PefindoIDX
		reqScoreproIntegrator         request.ScoreProIntegrator
	)

	// DEFAULT
	scoreGenerator = entity.ScoreGenerator{
		Key:               os.Getenv("SCOREPRO_DEFAULT_KEY"),
		ScoreGeneratorsID: os.Getenv("SCOREPRO_DEFAULT_SCORE_GENERATOR_ID"),
	}

	if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_NEW {
		for _, v := range req.Address {
			if v.Type == constant.ADDRESS_TYPE_RESIDENCE {
				residenceZipCode = v.ZipCode
				firstDigitsOfResidenceZipCode = string(v.ZipCode[0])
				break
			}
		}
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

	trxDetailBiro, err = u.repository.GetTrxDetailBIro(req.Transaction.ProspectID)

	// Get Pefindo IDX
	pefindoIDX = request.PefindoIDX{
		ProspectID: req.Transaction.ProspectID,
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
	resp, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("PEFINDO_IDX_URL"), paramPefindoIDX, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.Transaction.ProspectID, accessToken)

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
		ProspectID:       req.Transaction.ProspectID,
		CBFound:          cbFound,
		StatusKonsumen:   spDupcheck.StatusKonsumen,
		RequestorID:      os.Getenv("SCOREPRO_REQUESTID"),
		Journey:          constant.JOURNEY_SCOREPRO,
		PhoneNumber:      req.CustomerPersonal.MobilePhone,
		ScoreGeneratorID: scoreGenerator.ScoreGeneratorsID,
	}

	intZipcode, _ := strconv.Atoi(residenceZipCode[0:2])
	ntfOtr := math.Floor(req.Apk.NTF/req.Apk.OTR*100) / 100

	currentYear, _ := strconv.Atoi(time.Now().Format("2006-01-02")[0:4])

	bpkbKey := req.Item.BPKBName
	if bpkbKey == "KK" {
		bpkbKey = "O"
	}

	if scoreGenerator.Key == os.Getenv("KEY_SCOREPRO_IDX_2W_JABOJABAR") {

		BPKBYear, _ := strconv.Atoi(req.Item.ManufactureYear)
		ageVehicle := currentYear - BPKBYear

		reqScoreproIntegrator.Data = map[string]interface{}{
			"bpkb_name":      bpkbKey,
			"worst_24mth":    respPefindoIDX.Worst24Mth,
			"gender":         req.CustomerPersonal.Gender,
			"marital_status": req.CustomerPersonal.MaritalStatus,
			"ntf_otr":        ntfOtr,
			"zip_code":       intZipcode,
			"tenor":          req.Apk.Tenor,
			"vehicle_age":    ageVehicle,
			"profession_id":  req.CustomerEmployment.ProfessionID,
		}

	} else if scoreGenerator.Key == os.Getenv("KEY_SCOREPRO_IDX_2W_OTHERS") {

		employmentSinceYear, _ := strconv.Atoi(req.CustomerEmployment.EmploymentSinceYear)
		employmentSinceYear = currentYear - employmentSinceYear

		reqScoreproIntegrator.Data = map[string]interface{}{
			"bpkb_name":        bpkbKey,
			"ntf_otr":          ntfOtr,
			"zip_code":         intZipcode,
			"worst_12mth_auto": respPefindoIDX.Worst12MthAuto,
			"gender":           req.CustomerPersonal.Gender,
			"marital_status":   req.CustomerPersonal.MaritalStatus,
			"tenor":            req.Apk.Tenor,
			"length_of_empl":   employmentSinceYear,
			"home_status":      req.CustomerPersonal.HomeStatus,
		}

	} else if scoreGenerator.Key == os.Getenv("KEY_SCOREPRO_IDX_2W_AORO") {

		location, _ := time.LoadLocation("Asia/Jakarta")
		layout := "2006-01-02"
		convTime, _ := time.ParseInLocation(layout, req.CustomerPersonal.BirthDate, location)

		currentTime := time.Now()
		dateToday := currentTime.Format(layout)
		todayTime, _ := time.ParseInLocation(layout, dateToday, location)

		age := utils.HumanAgeCalculator(convTime, todayTime)

		var activeLoanTypeLast6M string
		getActiveLoanTypeLast6M, err := u.repository.GetActiveLoanTypeLast6M(spDupcheck.CustomerID.(string))
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetActiveLoanTypeLast6M Scorepro Error")
			return responseScs, data, respPefindoIDX, err
		}

		if strings.Replace(getActiveLoanTypeLast6M.ActiveLoanTypeLast6M, " ", "", -1) == ";;" {
			getActiveLoanTypeLast24M, err := u.repository.GetActiveLoanTypeLast24M(spDupcheck.CustomerID.(string))
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - GetActiveLoanTypeLast24M Scorepro Error")
				return responseScs, data, respPefindoIDX, err
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

		var moblast string
		getMoblast, err := u.repository.GetMoblast(spDupcheck.CustomerID.(string))
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetMoblast Scorepro Error")
			return responseScs, data, respPefindoIDX, err
		}

		if getMoblast.Moblast == "" {
			moblast = "9999"
		} else {
			intMob, _ := strconv.Atoi(getMoblast.Moblast)
			if intMob > 24 {
				moblast = "9999"
			} else {
				moblast = getMoblast.Moblast
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
			"gender":                  req.CustomerPersonal.Gender,
			"marital_status":          req.CustomerPersonal.MaritalStatus,
		}

	}

	paramScorepro, _ := json.Marshal(reqScoreproIntegrator)

	resp, _ = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("SCOREPRO_IDX_URL"), paramScorepro, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.Transaction.ProspectID, accessToken)

	if resp.StatusCode() != 200 {

		responseScs = response.IntegratorScorePro{
			ProspectID:  req.Transaction.ProspectID,
			Result:      constant.DECISION_PASS,
			ScoreResult: constant.SCOREPRO_RESULT_MEDIUM_2ND,
		}

	} else {

		err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &responseScs)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal IntegratorScorePro Error")
			return
		}

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
		if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_AO && spDupcheck.InstallmentTopup == 0 && spDupcheck.NumberOfPaidInstallment >= 6 {
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
			data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
			data.Source = constant.SOURCE_DECISION_SCOREPRO
			data.Info = string(info)
			return
		}

		if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_RO || (spDupcheck.InstallmentTopup > 0 && spDupcheck.MaxOverdueDaysforActiveAgreement <= 30) {
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
			data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
			data.Source = constant.SOURCE_DECISION_SCOREPRO
			data.Info = string(info)
			return
		}
	}

	if cbFound {
		if spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_RO || spDupcheck.StatusKonsumen == constant.STATUS_KONSUMEN_AO {
			if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(info)
			} else {
				if responseScs.ScoreResult == "LOW" {
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(info)
				} else {
					data.Result = constant.DECISION_PASS
					data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(info)
				}
			}
		} else {
			if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(info)
			} else {
				if strings.ToUpper(pefindoScore) == "VERY HIGH RISK" {
					if strings.Contains("2,3,4,5", responseScs.Segmen) {
						data.Result = constant.DECISION_REJECT
						data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(info)
					} else if strings.Contains("6,7,8,9,10,11,12", responseScs.Segmen) {
						data.Result = constant.DECISION_PASS
						data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(info)
					} else {
						if responseScs.ScoreResult == "LOW" {
							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
							data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
							data.Source = constant.SOURCE_DECISION_SCOREPRO
							data.Info = string(info)
						} else {
							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
							data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
							data.Source = constant.SOURCE_DECISION_SCOREPRO
							data.Info = string(info)
						}
					}
				} else {
					if responseScs.ScoreResult == "LOW" {
						data.Result = constant.DECISION_REJECT
						data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(info)
					} else {
						data.Result = constant.DECISION_PASS
						data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(info)
					}
				}
			}
		}
	} else {
		if responseScs.IsTsi {
			if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(info)
			} else {
				if responseScs.Result == constant.DECISION_PASS {
					data.Result = constant.DECISION_PASS
					data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(info)
				} else {
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(info)
				}
			}
		} else {
			if responseScs.Result == constant.DECISION_PASS {
				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
				data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
				data.Source = constant.SOURCE_DECISION_SCOREPRO
				data.Info = string(info)
			} else {
				if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
					// Handling ASS-SCORE
					segmen, _ := strconv.Atoi(responseScs.Segmen)
					segmenAssScore, _ := strconv.Atoi(os.Getenv("SCOREPRO_SEGMEN_ASS_SCORE"))
					if strings.Contains(responseScs.Status, "ASS-") && segmen > 0 && segmen <= segmenAssScore {
						data.Result = constant.DECISION_REJECT
						data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(info)
					} else {
						data.Result = constant.DECISION_PASS
						data.Code = constant.CODE_SCOREPRO_GTEMIN_THRESHOLD
						data.Reason = constant.REASON_SCOREPRO_GTEMIN_THRESHOLD
						data.Source = constant.SOURCE_DECISION_SCOREPRO
						data.Info = string(info)
					}
				} else {
					data.Result = constant.DECISION_REJECT
					data.Code = constant.CODE_SCOREPRO_LTMIN_THRESHOLD
					data.Reason = constant.REASON_SCOREPRO_LTMIN_THRESHOLD
					data.Source = constant.SOURCE_DECISION_SCOREPRO
					data.Info = string(info)
				}
			}
		}
	}
	return
}
