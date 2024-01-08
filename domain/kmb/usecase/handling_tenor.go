package usecase

import (
	"encoding/json"
	"errors"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
)

func jsonToMap(jsonStr string) map[string]interface{} {
	result := make(map[string]interface{})
	json.Unmarshal([]byte(jsonStr), &result)
	return result
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
