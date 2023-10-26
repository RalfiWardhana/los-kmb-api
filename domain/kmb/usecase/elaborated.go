package usecase

import (
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
)

func (u usecase) ElaboratedScheme(prospectID string, req request.Metrics) (data response.UsecaseApi, err error) {

	var (
		trxElaborateLtv entity.MappingElaborateLTV
	)

	trxElaborateLtv, err = u.repository.GetElaborateLtv(prospectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetElaborateLtv Error")
		return
	}

	ltv := (req.Apk.AF / req.Apk.OTR) * 100

	if ltv > float64(trxElaborateLtv.LTV) {
		data = response.UsecaseApi{
			Result:         constant.DECISION_PASS,
			Code:           constant.STRING_CODE_PASS_ELABORATE,
			Reason:         constant.REASON_PASS_ELABORATE,
			SourceDecision: constant.SOURCE_DECISION_TENOR,
		}
	} else {
		data = response.UsecaseApi{
			Result:         constant.DECISION_PASS,
			Code:           constant.STRING_CODE_PASS_ELABORATE,
			Reason:         constant.REASON_PASS_ELABORATE,
			SourceDecision: constant.SOURCE_DECISION_TENOR,
		}
	}

	return
}
