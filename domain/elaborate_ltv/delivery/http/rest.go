package http

import (
	"los-kmb-api/domain/elaborate_ltv/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"

	"github.com/labstack/echo/v4"
)

type handlerKmbElaborate struct {
	usecase    interfaces.Usecase
	repository interfaces.Repository
	Json       common.JSON
}

func ElaborateHandler(kmbroute *echo.Group, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware) {
	handler := handlerKmbElaborate{
		usecase:    usecase,
		repository: repository,
		Json:       json,
	}
	kmbroute.POST("/elaborate", handler.Elaborate, middlewares.AccessMiddleware())
}

// ElaborateLTV Tools godoc
// @Description ElaborateLTV
// @Tags Tools
// @Produce json
// @Param body body request.ElaborateLTV true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ElaborateLTV}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/elaborate [post]
func (c *handlerKmbElaborate) Elaborate(ctx echo.Context) (err error) {

	var (
		req     request.ElaborateLTV
		resp    interface{}
		ctxJson error
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Get(constant.HeaderXRequestID).(string)}
		go c.repository.SaveLogOrchestrator(headers, req, resp, "/api/v3/kmb/elaborate", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.InternalServerErrorCustomV3(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB ELABORATE", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB ELABORATE", req, err)
		return ctxJson
	}

	accessToken := middlewares.UserInfoData.AccessToken

	data, err := c.usecase.Elaborate(ctx.Request().Context(), req, accessToken)

	if err != nil {
		ctxJson, resp = c.Json.ServiceUnavailableV3(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB ELABORATE", req)
		return ctxJson
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB ELABORATE", req, data)
	return ctxJson
}
