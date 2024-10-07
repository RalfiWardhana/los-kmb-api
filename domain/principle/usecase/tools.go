package usecase

import (
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"

	"github.com/jinzhu/gorm"
)

func (u usecase) PrincipleStep(idNumber string) (step response.StepPrinciple, err error) {

	data, err := u.repository.GetTrxPrincipleStatus(idNumber)

	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return
		}

		return response.StepPrinciple{}, nil
	}

	step.ProspectID = data.ProspectID
	step.UpdatedAt = data.UpdatedAt.Format(constant.FORMAT_DATE_TIME)

	switch data.Decision {

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

	case constant.DECISION_CREDIT_PROCESS:

		step.ColorCode = "#FFCC00"
		step.Status = constant.REASON_PROSES_SURVEY
	}

	return
}
