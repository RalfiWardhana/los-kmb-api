package http

import (
	"los-kmb-api/domain/elaborate_ltv/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/dto"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/authorization"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"time"

	"github.com/labstack/echo/v4"
)

type handlerKmbElaborate struct {
	usecase       interfaces.Usecase
	repository    interfaces.Repository
	authorization authorization.Authorization
	Json          common.JSON
}

func ElaborateHandler(kmbroute *echo.Group, usecase interfaces.Usecase, repository interfaces.Repository, authorization authorization.Authorization, json common.JSON, middlewares *middlewares.AccessMiddleware) {
	handler := handlerKmbElaborate{
		usecase:       usecase,
		repository:    repository,
		authorization: authorization,
		Json:          json,
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
		go c.repository.SaveLogOrchestrator(ctx.Request().Header, req, resp, "/api/v3/kmb/elaborate", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	err = c.authorization.Authorization(dto.AuthModel{
		ClientID:   ctx.Request().Header.Get(constant.HEADER_CLIENT_ID),
		Credential: ctx.Request().Header.Get(constant.HEADER_AUTHORIZATION),
	}, time.Now().Local())

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
		return ctxJson
	}

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.InternalServerErrorCustomV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
		return ctxJson
	}

	accessToken := middlewares.UserInfoData.AccessToken

	data, err := c.usecase.Elaborate(ctx.Request().Context(), req, accessToken)

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
		return ctxJson
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, data)
	return ctxJson
}
