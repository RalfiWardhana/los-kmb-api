package usecase

import (
	"errors"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
)

func (u usecase) RejectTenor36(idNumber string) (result response.UsecaseApi, err error) {

	currentTrxWithRejectDSR, err := u.repository.GetCurrentTrxWithRejectDSR(idNumber)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetCurrentTrxWithRejectDSR Error")
		return
	}

	if currentTrxWithRejectDSR.ProspectID != "" {
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
