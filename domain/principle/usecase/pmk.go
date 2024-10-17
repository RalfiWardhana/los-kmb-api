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

func (u usecase) CheckPMK(branchID, customerKMB string, income float64, homeStatus, professionID, birthDate string, tenor int, maritalStatus string, empYear, empMonth, stayYear, stayMonth int) (data response.UsecaseApi, err error) {

	location, _ := time.LoadLocation("Asia/Jakarta")

	layout := constant.FORMAT_DATE

	data = response.UsecaseApi{Result: constant.DECISION_PASS, Code: constant.CODE_PMK_SESUAI, Reason: constant.REASON_PMK_SESUAI, SourceDecision: constant.SOURCE_DECISION_PMK}

	config, err := u.repository.GetConfig(constant.GROUP_PMK, constant.LOB_KMB_OFF, constant.KEY_PMK)
	if err != nil {
		return
	}

	var configData entity.ConfigPMK

	json.Unmarshal([]byte(config.Value), &configData)

	minimalIncome, err := u.repository.GetMinimalIncomePMK(branchID, customerKMB)
	if err != nil {
		return
	}

	if int(income) < minimalIncome.Income {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_INCOME
		data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_INCOME)
		return
	}

	if empYear != 0 && empMonth != 0 {
		var timeNow int64
		if professionID == constant.PROFESSION_ID_WRST || professionID == constant.PROFESSION_ID_PRO {
			timeNow = time.Now().AddDate(-configData.Data.LengthOfBusiness, 0, 0).Unix()
		} else {
			timeNow = time.Now().AddDate(-configData.Data.LengthOfWork, 0, 0).Unix()
		}

		convTime, _ := time.ParseInLocation(layout, setTimetoParse(empMonth, empYear), location)

		unixTime := convTime.Unix()

		if unixTime > timeNow {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_REJECT_WORK_EXPERIENCE
			data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_WORK_EXPERIENCE)
			return
		}

	}

	if stayYear != 0 && stayMonth != 0 {

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

		convTime, _ := time.ParseInLocation(layout, setTimetoParse(stayMonth, stayYear), location)

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

		convTime, _ := time.ParseInLocation(layout, birthDate, location)

		currentTime := time.Now()
		dateToday := currentTime.Format(layout)
		todayTime, _ := time.ParseInLocation(layout, dateToday, location)

		age = utils.HumanAgeCalculator(convTime, todayTime)

		if maritalStatus == constant.MARITAL_SINGLE {
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

		unix57th := convTime.AddDate(57, 0, 0).Unix()
		unixTenor := todayTime.AddDate(0, tenor, 0).Unix()

		if unixTenor > unix57th {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_REJECT_MAX_AGE
			data.Reason = fmt.Sprintf(" %s", constant.REASON_REJECT_MAX_AGE)
			return
		}
	}
	return
}

func setTimetoParse(month int, year int) string {

	if month >= 10 {
		return fmt.Sprintf("%d-%d-01", year, month)
	}

	return fmt.Sprintf("%d-0%d-01", year, month)
}
