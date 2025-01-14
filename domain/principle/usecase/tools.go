package usecase

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

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

		trxStatus, err := u.repository.GetTrxStatus(data.ProspectID)
		if err != nil {
			if err.Error() != constant.RECORD_NOT_FOUND {
				return response.StepPrinciple{}, err
			} else {
				step.ColorCode = "#FFCC00"
				step.Status = constant.REASON_PROSES_SURVEY
				return step, nil
			}
		}

		if trxStatus != (entity.TrxStatus{}) {
			if trxStatus.Activity == constant.ACTIVITY_STOP {
				switch trxStatus.Decision {
				case constant.DB_DECISION_CANCEL:
					_ = u.repository.UpdateToCancel(data.ProspectID)
					return step, nil
				case constant.DB_DECISION_REJECT:
					_ = u.repository.UpdateTrxPrincipleStatus(data.ProspectID, constant.DECISION_REJECT, 4)
					return step, nil
				case constant.DB_DECISION_APR:
					_ = u.repository.UpdateTrxPrincipleStatus(data.ProspectID, constant.DECISION_APPROVE, 4)
					return step, nil
				}
			}
		}

		step.ColorCode = "#FFCC00"
		step.Status = constant.REASON_PROSES_SURVEY
	}

	return
}

func (u usecase) PrinciplePublish(ctx context.Context, req request.PrinciplePublish, accessToken string) (err error) {

	principleStepOne, err := u.repository.GetPrincipleStepOne(req.ProspectID)
	if err != nil {
		return
	}

	return u.producer.PublishEvent(ctx, accessToken, constant.TOPIC_SUBMISSION_PRINCIPLE, constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE, req.ProspectID, utils.StructToMap(request.Update2wPrincipleTransaction{
		OrderID:       req.ProspectID,
		KpmID:         principleStepOne.KPMID,
		Source:        3,
		StatusCode:    req.StatusCode,
		ProductName:   principleStepOne.AssetCode,
		BranchCode:    principleStepOne.BranchID,
		AssetTypeCode: constant.KPM_ASSET_TYPE_CODE_MOTOR,
	}), 0)
}

func (u usecase) Step2Wilen(idNumber string) (resp response.Step2Wilen, err error) {

	data, err := u.repository.GetTrxKPMStatus(idNumber)
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return
		}

		return response.Step2Wilen{}, nil
	}

	resp.ProspectID = data.ProspectID
	resp.UpdatedAt = data.UpdatedAt.Format(constant.FORMAT_DATE_TIME)

	switch data.Decision {

	case constant.DECISION_KPM_READJUST:

		resp.ColorCode = "#00FF00"
		resp.Status = constant.REASON_PROSES_READJUST

	case constant.DECISION_CREDIT_PROCESS:

		trxStatus, err := u.repository.GetTrxStatus(data.ProspectID)
		if err != nil {
			if err.Error() != constant.RECORD_NOT_FOUND {
				return response.Step2Wilen{}, err
			} else {
				resp.ColorCode = "#FFCC00"
				resp.Status = constant.REASON_PROSES_SURVEY
				return resp, nil
			}
		}

		if trxStatus != (entity.TrxStatus{}) {
			if trxStatus.Activity == constant.ACTIVITY_STOP {
				switch trxStatus.Decision {
				case constant.DB_DECISION_CANCEL:
					_ = u.repository.UpdateTrxKPMStatus(data.ID, constant.DECISION_CANCEL)
					return response.Step2Wilen{}, err
				case constant.DB_DECISION_REJECT:
					_ = u.repository.UpdateTrxKPMStatus(data.ID, constant.DECISION_REJECT)
					return response.Step2Wilen{}, err
				case constant.DB_DECISION_APR:
					_ = u.repository.UpdateTrxKPMStatus(data.ID, constant.DECISION_APPROVE)
					return response.Step2Wilen{}, err
				}
			}
		}

		resp.ColorCode = "#FFCC00"
		resp.Status = constant.REASON_PROSES_SURVEY
	}

	return
}
