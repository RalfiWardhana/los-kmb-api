package http

import (
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/dto"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/authorization"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/common/platformlog"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"time"

	"github.com/KB-FMF/platform-library/auth"

	"github.com/labstack/echo/v4"
)

type handlerKMB struct {
	metrics       interfaces.Metrics
	usecase       interfaces.Usecase
	repository    interfaces.Repository
	authorization authorization.Authorization
	Json          common.JSON
	producer      platformevent.PlatformEventInterface
}

func KMBHandler(kmbroute *echo.Group, metrics interfaces.Metrics, usecase interfaces.Usecase, repository interfaces.Repository, authorization authorization.Authorization, json common.JSON, middlewares *middlewares.AccessMiddleware, producer platformevent.PlatformEventInterface) {
	handler := handlerKMB{
		metrics:       metrics,
		usecase:       usecase,
		repository:    repository,
		authorization: authorization,
		Json:          json,
		producer:      producer,
	}
	kmbroute.POST("/produce/journey", handler.ProduceJourney, middlewares.AccessMiddleware())
	kmbroute.POST("/produce/journey-after-prescreening", handler.ProduceJourneyAfterPrescreening, middlewares.AccessMiddleware())
	kmbroute.POST("/recalculate", handler.Recalculate, middlewares.AccessMiddleware())
	kmbroute.POST("/lock-system", handler.LockSystem, middlewares.AccessMiddleware())
	kmbroute.POST("/insert-staging/:prospectID", handler.InsertStagingIndex, middlewares.AccessMiddleware())
	kmbroute.POST("/go-live", handler.GoLive, middlewares.AccessMiddleware())
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

func (c *handlerKMB) LockSystem(ctx echo.Context) (err error) {
	var (
		req     request.LockSystem
		ctxJson error
	)

	auth := auth.New(platformlog.GetPlatformEnv())
	_, errAuth := auth.Validation(ctx.Request().Header.Get(constant.HEADER_AUTHORIZATION), "")
	if errAuth != nil {
		if errAuth.GetErrorCode() == "401" {
			err = fmt.Errorf("unauthorized - Invalid token")
			ctxJson, _ = c.Json.ErrorStandard(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS-LST", req, err)
			return ctxJson
		} else {
			err = fmt.Errorf("unauthorized - %v", errAuth.ErrorMessage())
			ctxJson, _ = c.Json.ErrorStandard(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS-LST", req, err)
			return ctxJson
		}
	}

	if err := ctx.Bind(&req); err != nil {
		ctxJson, _ = c.Json.ErrorBindStandard(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS-LST", req, err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, _ = c.Json.ErrorValidationStandard(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS-LST", req, err)
		return ctxJson
	}

	req.IDNumber, _ = utils.PlatformDecryptText(req.IDNumber)

	data, err := c.usecase.LockSystem(ctx.Request().Context(), req.IDNumber)

	if err != nil {
		ctxJson, _ = c.Json.ErrorStandard(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS-LST", req, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessStandard(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS-LST", req, data)
	return ctxJson
}

// Recalculate
// @Description Recalculate
// @Tags Recalculate
// @Produce json
// @Param body body request.Recalculate true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/recalculate [post]
func (c *handlerKMB) Recalculate(ctx echo.Context) (err error) {
	var (
		req     request.Recalculate
		resp    interface{}
		ctxJson error
	)

	// Save Log Orchestrator
	defer func() {
		go c.repository.SaveLogOrchestrator(ctx.Request().Header, req, resp, "/api/v3/kmb/recalculate", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	err = c.authorization.Authorization(dto.AuthModel{
		ClientID:   ctx.Request().Header.Get(constant.HEADER_CLIENT_ID),
		Credential: ctx.Request().Header.Get(constant.HEADER_AUTHORIZATION),
	}, time.Now().Local())

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB RECALCULATE", req, err)
		return ctxJson
	}

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB RECALCULATE", req, err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB RECALCULATE", req, err)
		return ctxJson
	}

	data, err := c.usecase.Recalculate(ctx.Request().Context(), req)

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB RECALCULATE", req, err)
		return ctxJson
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB RECALCULATE - Success", req, data)
	return ctxJson
}

// Insert Staging
// @Description Insert Staging
// @Tags Insert Staging
// @Produce json
// @Param prospectID path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/insert-staging/{prospectID} [post]
func (c *handlerKMB) InsertStagingIndex(ctx echo.Context) (err error) {
	var (
		ctxJson error
	)

	prospectID := ctx.Param("prospectID")

	err = c.authorization.Authorization(dto.AuthModel{
		ClientID:   ctx.Request().Header.Get(constant.HEADER_CLIENT_ID),
		Credential: ctx.Request().Header.Get(constant.HEADER_AUTHORIZATION),
	}, time.Now().Local())

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB Insert Staging", prospectID, err)
		return ctxJson
	}

	if prospectID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID does not exist")
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB Insert Staging", prospectID, err)
		return ctxJson
	}

	data, err := c.usecase.InsertStaging(prospectID)

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB Insert Staging", prospectID, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB Insert Staging Success", prospectID, data)
	return ctxJson
}

// Produce Sync Go-Live
// @Description Sync Go-Live
// @Tags Sync Go-Live
// @Produce json
// @Param body body request.SyncGoLive true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/go-live [post]
func (c *handlerKMB) GoLive(ctx echo.Context) (err error) {

	var (
		req     request.SyncGoLive
		resp    interface{}
		ctxJson error
	)

	// Save Log Orchestrator
	defer func() {
		go c.repository.SaveLogOrchestrator(ctx.Request().Header, req, resp, "/api/v3/kmb/go-live", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	err = c.authorization.Authorization(dto.AuthModel{
		ClientID:   ctx.Request().Header.Get(constant.HEADER_CLIENT_ID),
		Credential: ctx.Request().Header.Get(constant.HEADER_AUTHORIZATION),
	}, time.Now().Local())

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Sync Go-Live", req.ProspectID, err)
		return ctxJson
	}

	if err := ctx.Bind(&req); err != nil {
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Sync Go-Live", req, err)
		return ctxJson
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - Sync Go-Live", req, req)

	c.producer.PublishEvent(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK_GOLIVE, req.ProspectID, utils.StructToMap(resp), 0)

	return ctxJson
}
