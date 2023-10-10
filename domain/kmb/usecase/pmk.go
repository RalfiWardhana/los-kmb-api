package usecase

import (
	"encoding/json"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"time"
)

func (u usecase) PMK(income float64, homeStatus, jobPos, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi) {

	location, _ := time.LoadLocation("Asia/Jakarta")

	data = response.UsecaseApi{Result: constant.DECISION_PASS, Code: constant.CODE_PMK_SESUAI, Reason: constant.REASON_PMK_SESUAI}

	config, _ := u.repository.GetKMOBOff()

	var configData entity.ConfigPMK

	json.Unmarshal([]byte(config.Value), &configData)

	if income < configData.Data.MinimalIncome {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_INCOME
		data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_INCOME)
		return
	}

	if empYear != "" && empMonth != "" {
		timeNow := time.Now().AddDate(-configData.Data.LengthOfWork, 0, 0).Unix()

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
