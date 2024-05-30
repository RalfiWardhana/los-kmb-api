package eventhandlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformcache"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/http"
	"os"

	"github.com/KB-FMF/platform-library/event"
	jsoniter "github.com/json-iterator/go"
)

type handlers struct {
	multiusecase  interfaces.MultiUsecase
	usecase       interfaces.Usecase
	repository    interfaces.Repository
	validator     *common.Validator
	producer      platformevent.PlatformEvent
	Json          common.JSON
	platformCache platformcache.PlatformCacheInterface
}

func NewServiceFiltering(app *platformevent.ConsumerRouter, repository interfaces.Repository, usecase interfaces.Usecase, multiUsecase interfaces.MultiUsecase, validator *common.Validator, producer platformevent.PlatformEvent, json common.JSON,
	platformCache platformcache.PlatformCacheInterface) {
	handler := handlers{
		multiusecase:  multiUsecase,
		usecase:       usecase,
		repository:    repository,
		validator:     validator,
		producer:      producer,
		Json:          json,
		platformCache: platformCache,
	}
	app.Handle(constant.KEY_PREFIX_FILTERING, handler.Filtering)
}

func (h handlers) Filtering(ctx context.Context, event event.Event) (err error) {
	middlewares.GetPlatformAuth()
	body := event.GetBody()

	var (
		married         bool
		req             request.Filtering
		reqEncrypted    request.Filtering
		resultFiltering response.Filtering
		resp            interface{}
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Value(constant.HeaderXRequestID).(string)}
		h.repository.SaveLogOrchestrator(headers, reqEncrypted, resp, "/api/v3/kmb/consume/filtering", constant.METHOD_POST, req.ProspectID, ctx.Value(constant.HeaderXRequestID).(string))
	}()

	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &reqEncrypted)
	err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &req)

	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Unmarshal body error")
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
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
		LogFile:    constant.NEW_KMB_LOG,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    requestLog,
		Response: map[string]interface{}{
			"messages": "success consume data stream",
		},
	})

	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_URL, fmt.Sprintf("%s/api/v3/kmb/consume/filtering", constant.LOS_KMB_BASE_URL))
	ctx = context.WithValue(ctx, constant.CTX_KEY_INCOMING_REQUEST_METHOD, constant.METHOD_POST)

	err = h.validator.Validate(req)
	if err != nil {
		resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
		return nil
	}

	// decrypt request
	req.IDNumber, _ = utils.PlatformDecryptText(req.IDNumber)
	req.LegalName, _ = utils.PlatformDecryptText(req.LegalName)
	req.MotherName, _ = utils.PlatformDecryptText(req.MotherName)

	if req.Spouse != nil {
		req.Spouse.IDNumber, _ = utils.PlatformDecryptText(req.Spouse.IDNumber)
		req.Spouse.LegalName, _ = utils.PlatformDecryptText(req.Spouse.LegalName)
		req.Spouse.MotherName, _ = utils.PlatformDecryptText(req.Spouse.MotherName)

		var genderSpouse request.GenderCompare

		if req.Gender != req.Spouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := h.validator.Validate(&genderSpouse); err != nil {
			resp = h.Json.EventBadRequestErrorValidation(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
			return nil
		}

		married = true
	}

	// filtering already exist
	check, errCheck := h.usecase.FilteringProspectID(req.ProspectID)
	if errCheck != nil {
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, errCheck)

		// if the order is not from NE then produce event
		if req.ProspectID[0:2] != "NE" {
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
		}

		return nil
	}
	if err := h.validator.Validate(&check); err != nil {
		resp, err = h.platformCache.GetCache(ctx, middlewares.UserInfoData.AccessToken, os.Getenv("CACHE_COLLECTION_NAME"), fmt.Sprintf(constant.DOC_FILTERING, req.ProspectID))
		if err != nil {
			resultFiltering, err = h.usecase.GetResultFiltering(req.ProspectID)
			if err != nil {
				resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
			} else {
				resp = h.Json.EventSuccess(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, resultFiltering)
				h.platformCache.SetCache(ctx, middlewares.UserInfoData.AccessToken, os.Getenv("CACHE_COLLECTION_NAME"), fmt.Sprintf(constant.DOC_FILTERING, req.ProspectID), resp, os.Getenv("CACHE_FILTERING_EXPIRED"))
			}
		}

		var rs response.ApiResponse
		resp, _ := json.Marshal(resp)
		json.Unmarshal(resp, &rs)

		// if the order is not from NE then produce event
		if req.ProspectID[0:2] != "NE" {
			h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(rs), 0)
		}

		return nil
	}

	resultFiltering, err = h.multiusecase.Filtering(ctx, req, married, middlewares.UserInfoData.AccessToken)
	if err != nil {
		resp = h.Json.EventServiceError(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
	} else {
		resp = h.Json.EventSuccess(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, resultFiltering)
		h.platformCache.SetCache(ctx, middlewares.UserInfoData.AccessToken, os.Getenv("CACHE_COLLECTION_NAME"), fmt.Sprintf(constant.DOC_FILTERING, req.ProspectID), resp, os.Getenv("CACHE_FILTERING_EXPIRED"))
	}

	// if the order is not from NE then produce event
	if req.ProspectID[0:2] != "NE" {
		h.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
	}

	return nil
}
