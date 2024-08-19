package usecase

import (
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
)

func (u usecase) PrincipleStep(idNumber string) (step response.StepPrinciple, err error) {

	data, err := u.repository.GetTrxPrincipleStatus(idNumber)

	if err != nil {
		return
	}

	step.ProspectID = data.ProspectID
	step.UpdatedAt = data.UpdatedAt.Format(constant.FORMAT_DATE_TIME)

	switch data.Decision {

	case constant.DECISION_REJECT:

		step.ColorCode = "#FF0000"

		switch data.Step {

		case 1:
			step.Status = constant.REASON_ASSET_REJECT

		case 2:
			step.Status = constant.REASON_PROFIL_REJECT

		case 3:
			step.Status = constant.REASON_PEMBIAYAAN_REJECT
		}

	case constant.DECISION_PASS:

		step.ColorCode = "#00FF00"

		switch data.Step {

		case 1:
			step.Status = constant.REASON_ASSET_APPOVE

		case 2:
			step.Status = constant.REASON_PROFIL_APPROVE

		case 3:
			step.Status = constant.REASON_PEMBIAYAAN_APPROVE
		}

	case constant.DECISION_CANCEL:

		step.ColorCode = "#FFFF00"
		step.Status = constant.REASON_CANCEL
	}

	return
}
