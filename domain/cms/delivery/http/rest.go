package http

import (
	"errors"
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
	cmsroute.GET("/cms/ca/inquiry", handler.CaInquiry, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/ca/save-as-draft", handler.SaveAsDraft, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/ca/submit-decision", handler.SubmitDecision, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/akkk/view/:prospect_id", handler.GetAkkk, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/ca/cancel", handler.CancelOrder, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/ca/cancel-reason", handler.CancelReason, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/ca/return", handler.ReturnOrder, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/search", handler.SearchInquiry, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/approval/inquiry", handler.ApprovalInquiry, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/approval/reason", handler.ApprovalReason, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/approval/submit-approval", handler.SubmitApproval, middlewares.AccessMiddleware())
}

// CMS NEW KMB Tools godoc
// @Description Api Prescreening
// @Tags Prescreening
// @Produce json
// @Param search query string false "search"
// @Param branch_id query string false "branch_id"
// @Param page query string false "page"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/prescreening/inquiry [get]
func (c *handlerCMS) PrescreeningInquiry(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqInquiryPrescreening{
		Search:   ctx.QueryParam("search"),
		BranchID: ctx.QueryParam("branch_id"),
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
// @Param body body request.ReqReviewPrescreening true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ReviewPrescreening}
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

	if data.Decision == constant.DECISION_REJECT {
		c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.ProspectID, utils.StructToMap(resp), 0)
	} else if data.Decision == constant.DB_DECISION_APR {
		reqAfterPrescreening := request.AfterPrescreening{
			ProspectID: req.ProspectID,
		}
		c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_AFTER_PRESCREENING, req.ProspectID, utils.StructToMap(reqAfterPrescreening), 0)
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, data)
	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api Prescreening
// @Tags Prescreening
// @Produce json
// @Param reason_id query string false "reason_id"
// @Param page query string false "page"
// @Success 200 {object} response.ApiResponse{data=response.ReasonMessageRow}
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

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param body body request.ReqInquiryCa true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/ca/inquiry [get]
func (c *handlerCMS) CaInquiry(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqInquiryCa{
		Search:   ctx.QueryParam("search"),
		BranchID: ctx.QueryParam("branch_id"),
		Filter:   ctx.QueryParam("filter"),
		UserID:   ctx.QueryParam("user_id"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	data, rowTotal, err := c.usecase.GetInquiryCa(ctx.Request().Context(), req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Inquiry", req, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Inquiry", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Inquiry", req, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param body body request.ReqSaveAsDraft true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.CAResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/save-as-draft [post]
func (c *handlerCMS) SaveAsDraft(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqSaveAsDraft
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Save as Draft", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Save as Draft", req, err)
	}

	data, err := c.usecase.SaveAsDraft(ctx.Request().Context(), req)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Save as Draft", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Save as Draft", req, data)
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param body body request.ReqSubmitDecision true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.CAResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/ca/submit-decision [post]
func (c *handlerCMS) SubmitDecision(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqSubmitDecision
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Submit Decision", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Submit Decision", req, err)
	}

	data, err := c.usecase.SubmitDecision(ctx.Request().Context(), req)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Submit Decision", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Submit Decision", req, data)
}

// CMS NEW KMB Tools godoc
// @Description Api Search Inquiry
// @Tags Search Inquiry
// @Produce json
// @Param body body request.ReqSearchInquiry true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/search [get]
func (c *handlerCMS) SearchInquiry(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqSearchInquiry{
		UserID: ctx.QueryParam("user_id"),
		Search: ctx.QueryParam("search"),
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Search Inquiry", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Search Inquiry", req, err)
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	data, rowTotal, err := c.usecase.GetSearchInquiry(ctx.Request().Context(), req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Search Inquiry", req, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Search Inquiry", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Search Inquiry", req, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api AKKK
// @Tags AKKK
// @Produce json
// @Param prospect_id path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{data=response.ReasonMessageRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/akkk/view/{prospect_id} [get]
func (c *handlerCMS) GetAkkk(ctx echo.Context) (err error) {

	var (
		ctxJson error
	)

	prospectID := ctx.Param("prospect_id")

	if prospectID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID does not exist")
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB AKKK", prospectID, err)
		return ctxJson
	}

	data, err := c.usecase.GetAkkk(prospectID)

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB AKKK", prospectID, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB AKKK", prospectID, data)
	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param page query string false "page"
// @Success 200 {object} response.ApiResponse{data=response.ReasonMessageRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/ca/cancel-reason [get]
func (c *handlerCMS) CancelReason(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
	)

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	data, rowTotal, err := c.usecase.GetCancelReason(ctx.Request().Context(), pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA", pagination, response.ReasonMessageRow{Reason: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA", pagination, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA", pagination, response.ReasonMessageRow{
		Reason:         data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param body body request.ReqCancelOrder true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ApiResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/ca/cancel [post]
func (c *handlerCMS) CancelOrder(ctx echo.Context) (err error) {

	var (
		resp        interface{}
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqCancelOrder
		ctxJson     error
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Get(constant.HeaderXRequestID).(string)}
		go c.repository.SaveLogOrchestrator(headers, req, resp, "/api/v3/kmb/cms/ca/cancel", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.InternalServerErrorCustomV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Cancel Order", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Cancel Order", req, err)
		return ctxJson
	}

	data, err := c.usecase.CancelOrder(ctx.Request().Context(), req)

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Cancel Order", req, err)
		return ctxJson
	}

	if data.Status == constant.CANCEL_STATUS_SUCCESS {
		c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.ProspectID, utils.StructToMap(resp), 0)
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Cancel Order", req, data)
	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param body body request.ReqReturnOrder true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ApiResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/ca/return [post]
func (c *handlerCMS) ReturnOrder(ctx echo.Context) (err error) {

	var (
		resp        interface{}
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqReturnOrder
		ctxJson     error
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Get(constant.HeaderXRequestID).(string)}
		go c.repository.SaveLogOrchestrator(headers, req, resp, "/api/v3/kmb/cms/ca/return", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.InternalServerErrorCustomV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", req, err)
		return ctxJson
	}

	data, err := c.usecase.ReturnOrder(ctx.Request().Context(), req)

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", req, err)
		return ctxJson
	}

	if data.Status == constant.RETURN_STATUS_SUCCESS {
		c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.ProspectID, utils.StructToMap(resp), 0)
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", req, data)
	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api Credit Approval
// @Tags Credit Approval
// @Produce json
// @Param body body request.ReqInquiryCa true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/approval/inquiry [get]
func (c *handlerCMS) ApprovalInquiry(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqInquiryApproval{
		Search:   ctx.QueryParam("search"),
		BranchID: ctx.QueryParam("branch_id"),
		Filter:   ctx.QueryParam("filter"),
		UserID:   ctx.QueryParam("user_id"),
		Alias:    ctx.QueryParam("alias"),
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Inquiry", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Inquiry", req, err)
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	data, rowTotal, err := c.usecase.GetInquiryApproval(ctx.Request().Context(), req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Inquiry", req, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Inquiry", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Inquiry", req, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api Credit Approval
// @Tags Credit Approval
// @Produce json
// @Param page query string false "page" and type
// @Success 200 {object} response.ApiResponse{data=response.ReasonMessageRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/approval/reason [get]
func (c *handlerCMS) ApprovalReason(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
	)

	req := request.ReqApprovalReason{
		Type: ctx.QueryParam("type"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 50,
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Reason", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Reason", req, err)
	}

	data, rowTotal, err := c.usecase.GetApprovalReason(ctx.Request().Context(), req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Reason", pagination, response.ReasonMessageRow{Reason: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Reason", pagination, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Reason", pagination, response.ReasonMessageRow{
		Reason:         data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api Credit Approval
// @Tags Credit Approval
// @Produce json
// @Param body body request.ReqSubmitApproval true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ApprovalResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/approval/submit-approval [post]
func (c *handlerCMS) SubmitApproval(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqSubmitApproval
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", req, err)
	}

	data, err := c.usecase.SubmitApproval(ctx.Request().Context(), req)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", req, data)
}
