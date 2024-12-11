package eventhandlers

import (
	"context"
	"encoding/base64"
	"los-kmb-api/domain/principle/interfaces"
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
	usecase    interfaces.Usecase
	repository interfaces.Repository
	validator  *common.Validator
	producer   platformevent.PlatformEventInterface
	Json       common.JSON
}

func NewServicePrinciple(app *platformevent.ConsumerRouter, repository interfaces.Repository, usecase interfaces.Usecase, validator *common.Validator, producer platformevent.PlatformEventInterface, json common.JSON) {
	handler := handlers{
		usecase:    usecase,
		repository: repository,
		validator:  validator,
		producer:   producer,
		Json:       json,
	}
	app.Handle("new_kmb_status_update", handler.PrincipleUpdateStatus)
}

// event update status principle order
func (h handlers) PrincipleUpdateStatus(ctx context.Context, event event.Event) (err error) {
	middlewares.GetPlatformAuth()
	body := event.GetBody()

	var (
		req           request.PrincipleUpdateStatus
		principleData entity.TrxPrincipleStepOne
	)

	_ = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(body, &req)

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

	principleData, _ = h.repository.GetPrincipleStepOne(req.ProspectID)
	if principleData != (entity.TrxPrincipleStepOne{}) {
		if req.OrderStatus == constant.PRINCIPLE_STATUS_CANCEL_SALLY {
			_ = h.repository.UpdateToCancel(req.ProspectID)
		}
	}

	return nil
}
