package eventhandlers

import (
	"context"
	"los-kmb-api/domain/principle/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"

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
	app.Handle(constant.KEY_PREFIX_UPDATE_STATUS_NEW_KMB, handler.PrincipleUpdateStatus)
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

	principleData, _ = h.repository.GetPrincipleStepOne(req.ProspectID)
	if principleData != (entity.TrxPrincipleStepOne{}) {
		if req.OrderStatus == constant.PRINCIPLE_STATUS_CANCEL_SALLY {
			_ = h.repository.UpdateToCancel(req.ProspectID)
		}
	}

	return nil
}
