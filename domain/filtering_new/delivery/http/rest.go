package http

import (
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/labstack/echo/v4"
)

type handlerKmbFiltering struct {
	multiusecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	Json         common.JSON
	producer     platformevent.PlatformEventInterface
}

func FilteringHandler(kmbroute *echo.Group, multiUsecase interfaces.MultiUsecase, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware,
	producer platformevent.PlatformEventInterface) {
	handler := handlerKmbFiltering{
		multiusecase: multiUsecase,
		usecase:      usecase,
		repository:   repository,
		Json:         json,
		producer:     producer,
	}
	kmbroute.POST("/produce/filtering", handler.ProduceFiltering, middlewares.AccessMiddleware())
}

// Produce Filtering Tools godoc
// @Description Produce Filtering via REST API
// @Tags Filtering
// @Produce json
// @Param body body request.Filtering true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/produce/filtering [post]
func (c *handlerKmbFiltering) ProduceFiltering(ctx echo.Context) (err error) {

	var (
		req request.Filtering
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", err)
	}

	c.producer.PublishEvent(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_FILTERING, req.ProspectID, utils.StructToMap(req), 0)

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING - Please wait, your request is being processed", req, nil)
}
