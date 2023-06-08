package http

import (
	"los-kmb-api/domain/elaborate/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"

	"github.com/labstack/echo/v4"
)

type handlerKmbElaborate struct {
	multiusecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	Json         common.JSON
}

func ElaborateHandler(kmbroute *echo.Group, multiUsecase interfaces.MultiUsecase, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware) {
	handler := handlerKmbElaborate{
		multiusecase: multiUsecase,
		usecase:      usecase,
		repository:   repository,
		Json:         json,
	}
	kmbroute.POST("/elaborate", handler.Elaborate, middlewares.AccessMiddleware())
}

// KmbElaborate Tools godoc
// @Description KmbElaborate
// @Tags Tools
// @Produce json
// @Param body body request.BodyRequestElaborate true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ElaborateResult}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /elaborate [post]
func (c *handlerKmbElaborate) Elaborate(ctx echo.Context) (err error) {

	var req request.BodyRequestElaborate

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB ELABORATE", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB ELABORATE", req, err)
	}

	accessToken := middlewares.UserInfoData.AccessToken

	data, err := c.multiusecase.Elaborate(ctx.Request().Context(), req, accessToken)

	if err != nil {
		return c.Json.ServiceUnavailableV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req)
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, data)
}
