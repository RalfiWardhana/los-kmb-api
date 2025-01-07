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
	"os"
	"strconv"
)

func (u multiUsecase) PrincipleEmergencyContact(ctx context.Context, req request.PrincipleEmergencyContact, accessToken string) (data response.UsecaseApi, err error) {
	var (
		principleStepOne             entity.TrxPrincipleStepOne
		principleStepThree           entity.TrxPrincipleStepThree
		trxPrincipleEmergencyContact entity.TrxPrincipleEmergencyContact
	)

	principleStepOne, err = u.repository.GetPrincipleStepOne(req.ProspectID)
	if err != nil {
		return
	}

	principleStepThree, err = u.repository.GetPrincipleStepThree(req.ProspectID)
	if err != nil {
		return
	}

	if principleStepThree.Decision == constant.DECISION_REJECT {
		err = errors.New(constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE)
		return
	}

	trxWorker, err := u.repository.GetTrxWorker(req.ProspectID, constant.WORKER_CATEGORY_PRINCIPLE_KMB)
	if err != nil {
		return
	}

	if len(trxWorker) == 0 {

		phone := req.Phone
		if len(req.Phone) > 10 {
			phone = req.Phone[:10]
		}

		trxPrincipleEmergencyContact = entity.TrxPrincipleEmergencyContact{
			ProspectID:   req.ProspectID,
			Name:         req.Name,
			Relationship: req.Relationship,
			MobilePhone:  req.MobilePhone,
			Address:      req.Address,
			Rt:           req.Rt,
			Rw:           req.Rw,
			Kelurahan:    req.Kelurahan,
			Kecamatan:    req.Kecamatan,
			City:         req.City,
			Province:     req.Province,
			ZipCode:      req.ZipCode,
			AreaPhone:    req.AreaPhone,
			Phone:        phone,
		}

		err = u.repository.SavePrincipleEmergencyContact(trxPrincipleEmergencyContact, principleStepThree.IDNumber)
		if err != nil {
			return
		}

		timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

		var worker []entity.TrxWorker

		headerParamLos, _ := json.Marshal(
			map[string]string{
				"X-Client-ID":   os.Getenv("CLIENT_LOS"),
				"Authorization": os.Getenv("AUTH_LOS"),
			})

		var sequence int

		err = u.usecase.PrincipleCoreCustomer(ctx, req.ProspectID, accessToken)

		if err != nil {
			//insert customer
			sequence += 1

			worker = append(worker, entity.TrxWorker{ProspectID: req.ProspectID, Activity: constant.WORKER_UNPROCESS, EndPointTarget: os.Getenv("PRINCIPLE_CORE_CUSTOMER_URL") + req.ProspectID,
				EndPointMethod: constant.METHOD_POST, Header: string(headerParamLos), Payload: "",
				ResponseTimeout: timeOut, APIType: constant.WORKER_TYPE_RAW, MaxRetry: 6, CountRetry: 0,
				Category: constant.WORKER_CATEGORY_PRINCIPLE_KMB, Action: constant.WORKER_ACTION_UPDATE_CORE_CUSTOMER, Sequence: sequence,
			})

			//get marketing program
			sequence += 1

			worker = append(worker, entity.TrxWorker{ProspectID: req.ProspectID, Activity: constant.WORKER_IDLE, EndPointTarget: os.Getenv("PRINCIPLE_MARKETING_PROGRAM_URL") + req.ProspectID,
				EndPointMethod: constant.METHOD_POST, Header: string(headerParamLos), Payload: "",
				ResponseTimeout: timeOut, APIType: constant.WORKER_TYPE_RAW, MaxRetry: 6, CountRetry: 0,
				Category: constant.WORKER_CATEGORY_PRINCIPLE_KMB, Action: constant.WORKER_ACTION_GET_MARKETING_PROGRAM, Sequence: sequence,
			})

		} else {

			err = u.usecase.PrincipleMarketingProgram(ctx, req.ProspectID, accessToken)

			if err != nil {

				//get marketing program
				sequence += 1

				worker = append(worker, entity.TrxWorker{ProspectID: req.ProspectID, Activity: constant.WORKER_UNPROCESS, EndPointTarget: os.Getenv("PRINCIPLE_MARKETING_PROGRAM_URL") + req.ProspectID,
					EndPointMethod: constant.METHOD_POST, Header: string(headerParamLos), Payload: "",
					ResponseTimeout: timeOut, APIType: constant.WORKER_TYPE_RAW, MaxRetry: 6, CountRetry: 0,
					Category: constant.WORKER_CATEGORY_PRINCIPLE_KMB, Action: constant.WORKER_ACTION_GET_MARKETING_PROGRAM, Sequence: sequence,
				})

			} else {

				statusCode := constant.PRINCIPLE_STATUS_SUBMIT_SALLY
				u.producer.PublishEvent(ctx, accessToken, constant.TOPIC_SUBMISSION_PRINCIPLE, constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE, req.ProspectID, utils.StructToMap(request.Update2wPrincipleTransaction{
					OrderID:       req.ProspectID,
					KpmID:         principleStepOne.KPMID,
					Source:        3,
					StatusCode:    statusCode,
					ProductName:   principleStepOne.AssetCode,
					BranchCode:    principleStepOne.BranchID,
					AssetTypeCode: constant.KPM_ASSET_TYPE_CODE_MOTOR,
				}), 0)

			}

		}

		u.repository.SaveToWorker(worker)
	}

	data.Code = constant.EMERGENCY_PASS_CODE
	data.Result = constant.DECISION_PASS
	data.Reason = constant.EMERGENCY_PASS_REASON

	return
}
