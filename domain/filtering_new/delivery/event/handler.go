package eventhandlers

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
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
	multiusecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	validator    *common.Validator
	producer     platformevent.PlatformEvent
	Json         common.JSON
}

func NewServiceApplication(app *platformevent.ConsumerRouter, repository interfaces.Repository, usecase interfaces.Usecase, multiUsecase interfaces.MultiUsecase, validator *common.Validator, producer platformevent.PlatformEvent, json common.JSON) {
	handler := handlers{
		multiusecase: multiUsecase,
		usecase:      usecase,
		repository:   repository,
		validator:    validator,
		producer:     producer,
		Json:         json,
	}
	app.Handle(constant.KEY_PREFIX_FILTERING, handler.Filtering)
}

func (h handlers) Filtering(ctx context.Context, event event.Event) (err error) {
	middlewares.GetPlatformAuth()

	body := event.GetBody()
	var (
		married         bool
		req             request.Filtering
		resultFiltering response.Filtering
		resp            interface{}
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Value(constant.HeaderXRequestID).(string)}
		go h.repository.SaveLogOrchestrator(headers, req, resp, "/api/v3/kmb/filtering", constant.METHOD_POST, req.ProspectID, ctx.Value(constant.HeaderXRequestID).(string))
	}()

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &req)

	if err != nil {
		log.Println(err.Error())
		log.Println("success consume data stream")
		log.Println("error Unmarshal")
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Unmarshal body error")
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	// Write Success Log
	requestLog := utils.StructToMap(req)
	requestLog["topic_key"] = string(event.GetKey())
	requestLog["topic_name"] = constant.TOPIC_SUBMISSION
	requestLog["rawBody"] = base64.RawStdEncoding.EncodeToString(body)
	common.CentralizeLog(ctx, middlewares.UserInfoData.AccessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Method:     http.MethodPost,
		Action:     "CONSUME_EVENT",
		Type:       "EVENT_PLATFORM_LIBRARY",
		LogFile:    constant.LOG_EVENT,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    requestLog,
		Response: map[string]interface{}{
			"messages": "success consume data stream",
		},
	})

	log.Println("consumed")

	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_URL, fmt.Sprintf("%s/api/v3/kmb/filtering", constant.LOS_KMB_BASE_URL))
	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_METHOD, constant.METHOD_POST)

	err = h.validator.Validate(req)
	if err != nil {
		log.Println(err.Error())
		log.Println("success consume data stream")
		log.Println("error validation")
		resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	if req.Spouse != nil {

		var genderSpouse request.GenderCompare

		if req.Gender != req.Spouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := h.validator.Validate(&genderSpouse); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}

		married = true
	}

	check, errCheck := h.usecase.FilteringProspectID(req.ProspectID)
	if errCheck != nil {
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}
	if err := h.validator.Validate(&check); err != nil {
		resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	resultFiltering, err = h.multiusecase.Filtering(ctx, req, married, middlewares.UserInfoData.AccessToken)
	if err != nil {
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
	} else {
		resp = h.Json.EventSuccess(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, resultFiltering)
	}

	h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
	log.Println("produced")

	return nil
}
