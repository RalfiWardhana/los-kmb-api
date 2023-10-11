package http

import (
	"los-kmb-api/domain/kmb/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/labstack/echo/v4"
)

type handlerKMB struct {
	metrics      interfaces.Metrics
	multiUsecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	Json         common.JSON
	producer     platformevent.PlatformEvent
}

func KMBHandler(kmbroute *echo.Group, metrics interfaces.Metrics, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware, producer platformevent.PlatformEvent) {
	handler := handlerKMB{
		metrics:    metrics,
		usecase:    usecase,
		repository: repository,
		Json:       json,
		producer:   producer,
	}
	kmbroute.POST("/produce/journey", handler.ProduceJourney, middlewares.AccessMiddleware())
	kmbroute.POST("/produce/journey-after-prescreening", handler.ProduceJourneyAfterPrescreening, middlewares.AccessMiddleware())
}

// Produce Journey
// @Description Submit to LOS
// @Tags Submit to LOS
// @Produce json
// @Param body body request.Metrics true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/produce/journey [post]
func (c *handlerKMB) ProduceJourney(ctx echo.Context) (err error) {

	var (
		req     request.Metrics
		ctxJson error
	)

	if err := ctx.Bind(&req); err != nil {
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
		return ctxJson
	}

	c.producer.PublishEvent(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_SUBMIT_TO_LOS, req.Transaction.ProspectID, utils.StructToMap(req), 0)

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB - Please wait, your request is being processed", req, nil)
}

// Produce Journey After Prescreening
// @Description Journey After Prescreening
// @Tags Submit to LOS
// @Produce json
// @Param body body request.AfterPrescreening true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/produce/journey-after-prescreening [post]
func (c *handlerKMB) ProduceJourneyAfterPrescreening(ctx echo.Context) (err error) {

	var (
		req     request.AfterPrescreening
		ctxJson error
	)

	if err := ctx.Bind(&req); err != nil {
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB", req, err)
		return ctxJson
	}

	c.producer.PublishEvent(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_AFTER_PRESCREENING, req.ProspectID, utils.StructToMap(req), 0)

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Journey KMB - Please wait, your request is being processed", req, nil)
}
