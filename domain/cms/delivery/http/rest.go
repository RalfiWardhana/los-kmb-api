package http

import (
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"strconv"

	"github.com/labstack/echo/v4"
)

type handlerCMS struct {
	usecase    interfaces.Usecase
	repository interfaces.Repository
	Json       common.JSON
	producer   platformevent.PlatformEventInterface
}

func CMSHandler(cmsroute *echo.Group, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, producer platformevent.PlatformEventInterface, middlewares *middlewares.AccessMiddleware) {
	handler := handlerCMS{
		usecase:    usecase,
		repository: repository,
		Json:       json,
		producer:   producer,
	}

	cmsroute.GET("/cms/prescreening/list-reason", handler.ListReason, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/prescreening/inquiry", handler.PrescreeningInquiry, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/prescreening/review", handler.ReviewPrescreening, middlewares.AccessMiddleware())
}

// CMS NEW KMB Tools godoc
// @Description Api Prescreening
// @Tags Prescreening
// @Produce json
// @Param body body request. true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/prescreening/inquiry [get]
func (c *handlerCMS) PrescreeningInquiry(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqInquiryPrescreening{
		Search: ctx.QueryParam("search"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	data, rowTotal, err := c.usecase.GetInquiryPrescreening(ctx.Request().Context(), req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Inquiry", req, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Inquiry", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Inquiry", req, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api Prescreening
// @Tags Prescreening
// @Produce json
// @Param body body request. true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/prescreening/review [post]
func (c *handlerCMS) ReviewPrescreening(ctx echo.Context) (err error) {

	var (
		resp        interface{}
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqReviewPrescreening
		ctxJson     error
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Get(constant.HeaderXRequestID).(string)}
		go c.repository.SaveLogOrchestrator(headers, req, resp, "/api/v3/kmb/cms/prescreening/review", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.InternalServerErrorCustomV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, err)
		return ctxJson
	}

	data, err := c.usecase.ReviewPrescreening(ctx.Request().Context(), req)

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, err)
		return ctxJson
	}

	if err != nil {
		resp = c.Json.EventServiceError(ctx.Request().Context(), accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, err)
	} else {
		resp = c.Json.EventSuccess(ctx.Request().Context(), accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, data)
	}

	if data.Decision == constant.DECISION_REJECT {
		c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_UPDATE_STATUS_FILTERING, req.ProspectID, utils.StructToMap(resp), 0)
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, data)
	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api Prescreening
// @Tags Prescreening
// @Produce json
// @Param body body request. true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/prescreening/list-reason [get]
func (c *handlerCMS) ListReason(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
	)

	req := request.ReqReasonPrescreening{
		ReasonID: ctx.QueryParam("reason_id"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 50,
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening", req, err)
	}

	data, rowTotal, err := c.usecase.GetReasonPrescreening(ctx.Request().Context(), req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening", req, response.ReasonMessageRow{Reason: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening", req, response.ReasonMessageRow{
		Reason:         data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}
