package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"time"
)

func (u usecase) PMK(branchID, customerKMB string, income float64, homeStatus, professionID, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi, err error) {

	location, _ := time.LoadLocation("Asia/Jakarta")

	data = response.UsecaseApi{Result: constant.DECISION_PASS, Code: constant.CODE_PMK_SESUAI, Reason: constant.REASON_PMK_SESUAI, SourceDecision: constant.SOURCE_DECISION_PMK}

	config, _ := u.repository.GetAppConfig()

	var configData entity.ConfigPMK

	json.Unmarshal([]byte(config.Value), &configData)

	minimalIncome, err := u.repository.GetMinimalIncomePMK(branchID, customerKMB)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Minimal Income PMK Error")
		return
	}

	if int(income) < minimalIncome.Income {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_INCOME
		data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_INCOME)
		return
	}

	if empYear != "" && empMonth != "" {
		var timeNow int64
		if professionID == constant.PROFESSION_ID_WRST || professionID == constant.PROFESSION_ID_PRO {
			timeNow = time.Now().AddDate(-configData.Data.LengthOfBusiness, 0, 0).Unix()
		} else {
			timeNow = time.Now().AddDate(-configData.Data.LengthOfWork, 0, 0).Unix()
		}

		convTime, _ := time.ParseInLocation("2006-01-02", fmt.Sprintf("%s-%s-01", empYear, empMonth), location)

		unixTime := convTime.Unix()

		if unixTime > timeNow {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_REJECT_WORK_EXPERIENCE
			data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_WORK_EXPERIENCE)
			return
		}

	}

	if stayYear != "" && stayMonth != "" {

		var lengthOfStay int

		switch homeStatus {
		case "SD", "KL", "KP":
			lengthOfStay = configData.Data.LengthOfStay.RumahSendiri
		case "PE":
			lengthOfStay = configData.Data.LengthOfStay.RumahDinas
		case "KR", "KS":
			lengthOfStay = configData.Data.LengthOfStay.RumahKontrak
		default:
			lengthOfStay = 2
		}

		timeNow := time.Now().AddDate(-lengthOfStay, 0, 0).Unix()
		convTime, _ := time.ParseInLocation("2006-01-02", fmt.Sprintf("%s-%s-01", stayYear, stayMonth), location)

		unixTime := convTime.Unix()

		if unixTime > timeNow {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_REJECT_HOME_SINCE
			data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_HOME_SINCE)
			return
		}

	}

	if birthDate != "" {
		var (
			age    int
			ageMin int
		)

		layout := "2006-01-02"
		convTime, _ := time.ParseInLocation(layout, birthDate, location)

		currentTime := time.Now()
		dateToday := currentTime.Format(layout)
		todayTime, _ := time.ParseInLocation(layout, dateToday, location)

		age = utils.HumanAgeCalculator(convTime, todayTime)

		if maritalStatus == "S" {
			ageMin = configData.Data.MinAgeSingle
		} else {
			ageMin = configData.Data.MinAgeMarried
		}

		if age < ageMin {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_REJECT_MIN_AGE
			data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_MIN_AGE_THRESHOLD)
			return
		}

		unix60th := convTime.AddDate(configData.Data.MaxAgeLimit, 0, 0).Unix()
		unixTenor := todayTime.AddDate(0, tenor, 0).Unix()

		if unixTenor > unix60th {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_REJECT_MAX_AGE
			data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_MAX_AGE)
			return
		}
	}
	return
}
