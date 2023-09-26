package eventhandlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/http"
	"os"

	"github.com/KB-FMF/platform-library/event"
	jsoniter "github.com/json-iterator/go"
)

type handlers struct {
	metrics    interfaces.Metrics
	usecase    interfaces.Usecase
	repository interfaces.Repository
	validator  *common.Validator
	producer   platformevent.PlatformEvent
	Json       common.JSON
}

func NewServiceKMB(app *platformevent.ConsumerRouter, repository interfaces.Repository, usecase interfaces.Usecase, metrics interfaces.Metrics, validator *common.Validator, producer platformevent.PlatformEvent, json common.JSON) {
	handler := handlers{
		metrics:    metrics,
		usecase:    usecase,
		repository: repository,
		validator:  validator,
		producer:   producer,
		Json:       json,
	}
	app.Handle(constant.KEY_PREFIX_SUBMIT_TO_LOS, handler.KMBIndex)
	app.Handle(constant.KEY_PREFIX_AFTER_PRESCREENING, handler.KMBAfterPrescreening)
}

// event submit to los
func (h handlers) KMBIndex(ctx context.Context, event event.Event) (err error) {
	middlewares.GetPlatformAuth()

	getBody := string(event.GetBody())
	getBody = utils.SafeJsonReplacer(getBody)

	body := []byte(getBody)

	var (
		req          request.Metrics
		reqEncrypted request.Metrics
		resp         interface{}
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Value(constant.HeaderXRequestID).(string)}
		go h.repository.SaveLogOrchestrator(headers, reqEncrypted, resp, "/api/v3/kmb/consume/journey", constant.METHOD_POST, req.Transaction.ProspectID, ctx.Value(constant.HeaderXRequestID).(string))
	}()

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &reqEncrypted)
	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &req)

	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Unmarshal body error")
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	// Write Success Log
	requestLog := utils.StructToMap(req)
	requestLog["topic_key"] = string(event.GetKey())
	requestLog["topic_name"] = constant.TOPIC_SUBMISSION_LOS
	requestLog["rawBody"] = base64.RawStdEncoding.EncodeToString(body)
	common.CentralizeLog(ctx, middlewares.UserInfoData.AccessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Method:     http.MethodPost,
		Action:     "CONSUME_EVENT",
		Type:       "EVENT_PLATFORM_LIBRARY",
		LogFile:    constant.NEW_KMB_LOG,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    requestLog,
		Response: map[string]interface{}{
			"messages": "success consume data stream",
		},
	})

	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_URL, fmt.Sprintf("%s/api/v3/kmb/consume/journey", constant.LOS_KMB_BASE_URL))
	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_METHOD, constant.METHOD_POST)

	err = h.validator.Validate(req)
	if err != nil {
		resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	// decrypt request
	req.CustomerPersonal.IDNumber, _ = utils.PlatformDecryptText(req.CustomerPersonal.IDNumber)
	req.CustomerPersonal.LegalName, _ = utils.PlatformDecryptText(req.CustomerPersonal.LegalName)
	req.CustomerPersonal.FullName, _ = utils.PlatformDecryptText(req.CustomerPersonal.FullName)
	req.CustomerPersonal.SurgateMotherName, _ = utils.PlatformDecryptText(req.CustomerPersonal.SurgateMotherName)

	if req.CustomerSpouse != nil {
		req.CustomerSpouse.IDNumber, _ = utils.PlatformDecryptText(req.CustomerSpouse.IDNumber)
		req.CustomerSpouse.LegalName, _ = utils.PlatformDecryptText(req.CustomerSpouse.LegalName)
		req.CustomerSpouse.FullName, _ = utils.PlatformDecryptText(req.CustomerSpouse.FullName)
		req.CustomerSpouse.SurgateMotherName, _ = utils.PlatformDecryptText(req.CustomerSpouse.SurgateMotherName)

		var genderSpouse request.GenderCompare

		if req.CustomerPersonal.Gender != req.CustomerSpouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := h.validator.Validate(&genderSpouse); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}
	}

	if req.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_WRST || req.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_PRO {
		var newOmset []request.CustomerOmset
		if req.CustomerOmset != nil {
			omset, _ := json.Marshal(*req.CustomerOmset)
			json.Unmarshal(omset, &newOmset)
		}

		var validateOmset request.ValidateOmset
		validateOmset.CustomerOmset = newOmset
		if err := h.validator.Validate(&validateOmset); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}

	}

	if req.CustomerPersonal.MaritalStatus == constant.MARRIED {
		var spouseVal request.MarriedValidator
		spouseVal.CustomerSpouse = true
		if req.CustomerSpouse == nil {
			spouseVal.CustomerSpouse = false
		}

		if err := h.validator.Validate(&spouseVal); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}
	} else {
		var spouseVal request.SingleValidator
		spouseVal.CustomerSpouse = true
		if req.CustomerSpouse != nil {
			spouseVal.CustomerSpouse = false
		}

		if err := h.validator.Validate(&spouseVal); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}
	}

	resp, err = h.metrics.MetricsLos(ctx, req, middlewares.UserInfoData.AccessToken)
	if err != nil {
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
	} else {
		_ = h.repository.SaveTrxJourney(req.Transaction.ProspectID, reqEncrypted)
		resp = h.Json.EventSuccess(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, resp)
	}

	h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)

	return nil
}

// event after prescreening
func (h handlers) KMBAfterPrescreening(ctx context.Context, event event.Event) (err error) {
	middlewares.GetPlatformAuth()

	getBody := string(event.GetBody())
	getBody = utils.SafeJsonReplacer(getBody)

	body := []byte(getBody)

	var (
		trxJourney           entity.TrxJourney
		logOrhcerstrators    entity.LogOrchestrator
		reqAfterPrescreening request.AfterPrescreening
		req                  request.Metrics
		reqEncrypted         request.Metrics
		resp                 interface{}
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Value(constant.HeaderXRequestID).(string)}
		go h.repository.SaveLogOrchestrator(headers, reqEncrypted, resp, "/api/v3/kmb/consume/journey-after-prescreening", constant.METHOD_POST, reqAfterPrescreening.ProspectID, ctx.Value(constant.HeaderXRequestID).(string))
	}()

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &reqAfterPrescreening)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Unmarshal body error")
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	trxJourney, err = h.repository.GetTrxJourney(reqAfterPrescreening.ProspectID)
	if err != nil {
		logOrhcerstrators, err = h.repository.GetLogOrchestrator(reqAfterPrescreening.ProspectID)
		if err != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Request not exist")
			resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}

		err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(logOrhcerstrators.RequestData), &reqEncrypted)
		err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(logOrhcerstrators.RequestData), &req)

	} else {
		err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(trxJourney.Request), &reqEncrypted)
		err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(trxJourney.Request), &req)
	}

	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Unmarshal body error")
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	// Write Success Log
	requestLog := utils.StructToMap(req)
	requestLog["topic_key"] = string(event.GetKey())
	requestLog["topic_name"] = constant.TOPIC_SUBMISSION_LOS
	requestLog["rawBody"] = base64.RawStdEncoding.EncodeToString(body)
	common.CentralizeLog(ctx, middlewares.UserInfoData.AccessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Method:     http.MethodPost,
		Action:     "CONSUME_EVENT",
		Type:       "EVENT_PLATFORM_LIBRARY",
		LogFile:    constant.NEW_KMB_LOG,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    requestLog,
		Response: map[string]interface{}{
			"messages": "success consume data stream",
		},
	})

	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_URL, fmt.Sprintf("%s/api/v3/kmb/consume/journey-after-prescreening", constant.LOS_KMB_BASE_URL))
	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_METHOD, constant.METHOD_POST)

	err = h.validator.Validate(req)
	if err != nil {
		resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	// decrypt request
	req.CustomerPersonal.IDNumber, _ = utils.PlatformDecryptText(req.CustomerPersonal.IDNumber)
	req.CustomerPersonal.LegalName, _ = utils.PlatformDecryptText(req.CustomerPersonal.LegalName)
	req.CustomerPersonal.FullName, _ = utils.PlatformDecryptText(req.CustomerPersonal.FullName)
	req.CustomerPersonal.SurgateMotherName, _ = utils.PlatformDecryptText(req.CustomerPersonal.SurgateMotherName)

	if req.CustomerSpouse != nil {
		req.CustomerSpouse.IDNumber, _ = utils.PlatformDecryptText(req.CustomerSpouse.IDNumber)
		req.CustomerSpouse.LegalName, _ = utils.PlatformDecryptText(req.CustomerSpouse.LegalName)
		req.CustomerSpouse.FullName, _ = utils.PlatformDecryptText(req.CustomerSpouse.FullName)
		req.CustomerSpouse.SurgateMotherName, _ = utils.PlatformDecryptText(req.CustomerSpouse.SurgateMotherName)

		var genderSpouse request.GenderCompare

		if req.CustomerPersonal.Gender != req.CustomerSpouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := h.validator.Validate(&genderSpouse); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}
	}

	if req.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_WRST || req.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_PRO {
		var newOmset []request.CustomerOmset
		if req.CustomerOmset != nil {
			omset, _ := json.Marshal(*req.CustomerOmset)
			json.Unmarshal(omset, &newOmset)
		}

		var validateOmset request.ValidateOmset
		validateOmset.CustomerOmset = newOmset
		if err := h.validator.Validate(&validateOmset); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}

	}

	if req.CustomerPersonal.MaritalStatus == constant.MARRIED {
		var spouseVal request.MarriedValidator
		spouseVal.CustomerSpouse = true
		if req.CustomerSpouse == nil {
			spouseVal.CustomerSpouse = false
		}

		if err := h.validator.Validate(&spouseVal); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}
	} else {
		var spouseVal request.SingleValidator
		spouseVal.CustomerSpouse = true
		if req.CustomerSpouse != nil {
			spouseVal.CustomerSpouse = false
		}

		if err := h.validator.Validate(&spouseVal); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}
	}

	resp, err = h.metrics.MetricsLos(ctx, req, middlewares.UserInfoData.AccessToken)
	if err != nil {
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
	} else {
		resp = h.Json.EventSuccess(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, resp)
	}

	h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.Transaction.ProspectID, utils.StructToMap(resp), 0)

	return nil
}
