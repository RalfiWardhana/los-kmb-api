package http

import (
	"errors"
	"fmt"
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

type handlerCMS struct {
	usecase    interfaces.Usecase
	repository interfaces.Repository
	Json       common.JSON
	producer   platformevent.PlatformEvent
}

func CMSHandler(cmsroute *echo.Group, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, producer platformevent.PlatformEvent, middlewares *middlewares.AccessMiddleware) {
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
	cmsroute.POST("/cms/ca/recalculate", handler.RecalculateOrder, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/search", handler.SearchInquiry, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/approval/inquiry", handler.ApprovalInquiry, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/approval/reason", handler.ApprovalReason, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/approval/submit-approval", handler.SubmitApproval, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/ne/submit", handler.SubmitNE, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/ne/inquiry", handler.NEInquiry, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/ne/inquiry/:prospect_id", handler.NEInquiryDetail, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/mapping-cluster/inquiry", handler.MappingClusterInquiry, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/mapping-cluster/download", handler.DownloadMappingCluster, middlewares.AccessMiddleware())
	cmsroute.POST("/cms/mapping-cluster/upload", handler.UploadMappingCluster, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/mapping-cluster/branch", handler.MappingClusterBranch, middlewares.AccessMiddleware())
	cmsroute.GET("/cms/mapping-cluster/change-log", handler.MappingClusterChangeLog, middlewares.AccessMiddleware())
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
		Search:      ctx.QueryParam("search"),
		UserID:      ctx.QueryParam("user_id"),
		BranchID:    ctx.QueryParam("branch_id"),
		MultiBranch: ctx.QueryParam("multi_branch"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Inquiry", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Inquiry", req, err)
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

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, data)

	if data.Decision == constant.DB_DECISION_REJECT {
		response := response.Metrics{
			ProspectID:     data.ProspectID,
			Decision:       req.Decision,
			Code:           data.Code,
			DecisionReason: string(data.Reason),
		}
		responseEvent := c.Json.EventSuccess(ctx.Request().Context(), accessToken, constant.NEW_KMB_LOG, "LOS - Pre Screening Review", req, response)
		go c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.ProspectID, utils.StructToMap(responseEvent), 0)

	} else if data.Decision == constant.DB_DECISION_APR {
		reqAfterPrescreening := request.AfterPrescreening{
			ProspectID: req.ProspectID,
		}
		go c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_AFTER_PRESCREENING, req.ProspectID, utils.StructToMap(reqAfterPrescreening), 0)
	}

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
	limit, _ := strconv.Atoi(os.Getenv("LIMIT_PAGE_REASON"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: limit,
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
		Search:      ctx.QueryParam("search"),
		BranchID:    ctx.QueryParam("branch_id"),
		MultiBranch: ctx.QueryParam("multi_branch"),
		Filter:      ctx.QueryParam("filter"),
		UserID:      ctx.QueryParam("user_id"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Inquiry", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Inquiry", req, err)
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
		UserID:      ctx.QueryParam("user_id"),
		BranchID:    ctx.QueryParam("branch_id"),
		MultiBranch: ctx.QueryParam("multi_branch"),
		Search:      ctx.QueryParam("search"),
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
// @Router /api/v3/kmb/cms/akkk/view/{prospect_id} [get]
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
// @Description Api Submit NE
// @Tags Submit NE
// @Produce json
// @Param body body request.MetricsNE true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ApiResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/ne/submit [post]
func (c *handlerCMS) SubmitNE(ctx echo.Context) (err error) {

	var (
		resp        interface{}
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.MetricsNE
		ctxJson     error
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Get(constant.HeaderXRequestID).(string)}
		c.repository.SaveLogOrchestrator(headers, req, resp, "/api/v3/kmb/cms/ne/submit", constant.METHOD_POST, req.Transaction.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.InternalServerErrorCustomV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Submit NE Error", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Submit NE Error", req, err)
		return ctxJson
	}

	payloadFiltering, err := c.usecase.SubmitNE(ctx.Request().Context(), req)

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Submit NE Error", req, err)
		return ctxJson
	}

	//produce filtering for NE
	go c.producer.PublishEvent(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_FILTERING, req.Transaction.ProspectID, utils.StructToMap(payloadFiltering), 0)

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Submit NE Success", req, nil)

	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param body body request.ReqInquiryCa true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/ne/inquiry [get]
func (c *handlerCMS) NEInquiry(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqInquiryNE{
		Search:      ctx.QueryParam("search"),
		BranchID:    ctx.QueryParam("branch_id"),
		MultiBranch: ctx.QueryParam("multi_branch"),
		Filter:      ctx.QueryParam("filter"),
		UserID:      ctx.QueryParam("user_id"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry", req, err)
	}

	data, rowTotal, err := c.usecase.GetInquiryNE(ctx.Request().Context(), req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry", req, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry", req, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param prospect_id path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{data=request.MetricsNE}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/ne/inquiry/{prospect_id} [get]
func (c *handlerCMS) NEInquiryDetail(ctx echo.Context) (err error) {

	var (
		ctxJson error
	)

	prospectID := ctx.Param("prospect_id")

	if prospectID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID does not exist")
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry Detail", prospectID, err)
		return ctxJson
	}

	data, err := c.usecase.GetInquiryNEDetail(ctx.Request().Context(), prospectID)

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry Detail", prospectID, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - NE Inquiry Detail", prospectID, data)
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
		Limit: 50,
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

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Cancel Order", req, data)

	if data.Status == constant.CANCEL_STATUS_SUCCESS {
		response := response.Metrics{
			ProspectID:     data.ProspectID,
			Code:           constant.CODE_CREDIT_COMMITTEE,
			Decision:       constant.DECISION_CANCEL,
			DecisionReason: string(data.Reason),
		}
		responseEvent := c.Json.EventSuccess(ctx.Request().Context(), accessToken, constant.NEW_KMB_LOG, "LOS - CA Cancel Order", req, response)

		go c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.ProspectID, utils.StructToMap(responseEvent), 0)
	}

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
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqReturnOrder
		ctxJson     error
	)

	if err := ctx.Bind(&req); err != nil {
		ctxJson, _ = c.Json.InternalServerErrorCustomV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, _ = c.Json.BadRequestErrorValidationV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", req, err)
		return ctxJson
	}

	data, err := c.usecase.ReturnOrder(ctx.Request().Context(), req)

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", req, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Return Order", req, data)
	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api CA
// @Tags CA
// @Produce json
// @Param body body request.ReqRecalculateOrder true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ApiResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/ca/recalculate [post]
func (c *handlerCMS) RecalculateOrder(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqRecalculateOrder
		ctxJson     error
	)

	if err := ctx.Bind(&req); err != nil {
		ctxJson, _ = c.Json.InternalServerErrorCustomV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Recalculate Order", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, _ = c.Json.BadRequestErrorValidationV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Recalculate Order", req, err)
		return ctxJson
	}

	data, err := c.usecase.RecalculateOrder(ctx.Request().Context(), req, accessToken)

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Recalculate Order", req, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - CA Recalculate Order", req, data)
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
		Search:      ctx.QueryParam("search"),
		BranchID:    ctx.QueryParam("branch_id"),
		MultiBranch: ctx.QueryParam("multi_branch"),
		Filter:      ctx.QueryParam("filter"),
		UserID:      ctx.QueryParam("user_id"),
		Alias:       ctx.QueryParam("alias"),
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
	limit, _ := strconv.Atoi(os.Getenv("LIMIT_PAGE_REASON"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: limit,
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
		resp        interface{}
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqSubmitApproval
		ctxJson     error
	)

	// Save Log Orchestrator
	defer func() {
		headers := map[string]string{constant.HeaderXRequestID: ctx.Get(constant.HeaderXRequestID).(string)}
		go c.repository.SaveLogOrchestrator(headers, req, resp, "/api/v3/kmb/cms/approval/submit-approval", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
	}()

	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.InternalServerErrorCustomV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", req, err)
		return ctxJson
	}

	data, err := c.usecase.SubmitApproval(ctx.Request().Context(), req)

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", req, err)
		return ctxJson
	}

	ctxJson, resp = c.Json.SuccessV3(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", req, data)

	if data.IsFinal && !data.NeedEscalation && data.Decision != constant.DECISION_RETURN {
		response := response.Metrics{
			ProspectID:     data.ProspectID,
			Code:           req.RuleCode,
			Decision:       req.Decision,
			DecisionReason: string(data.Reason),
		}

		responseEvent := c.Json.EventSuccess(ctx.Request().Context(), accessToken, constant.NEW_KMB_LOG, "LOS - Approval Submit Decision", req, response)

		go c.producer.PublishEvent(ctx.Request().Context(), accessToken, constant.TOPIC_SUBMISSION_LOS, constant.KEY_PREFIX_CALLBACK, req.ProspectID, utils.StructToMap(responseEvent), 0)
	}

	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api Mapping Cluster
// @Tags Mapping Cluster
// @Produce json
// @Param search query string false "search"
// @Param branch_id query string false "branch_id"
// @Param customer_status query string false "customer_status"
// @Param bpkb_name_type query string false "bpkb_name_type"
// @Param cluster query string false "cluster"
// @Param page query string false "page"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/mapping-cluster/inquiry [get]
func (c *handlerCMS) MappingClusterInquiry(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqListMappingCluster{
		Search:         ctx.QueryParam("search"),
		BranchID:       ctx.QueryParam("branch_id"),
		CustomerStatus: ctx.QueryParam("customer_status"),
		BPKBNameType:   ctx.QueryParam("bpkb_name_type"),
		Cluster:        ctx.QueryParam("cluster"),
	}

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", req, err)
	}

	data, rowTotal, err := c.usecase.GetInquiryMappingCluster(req, pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", req, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", req, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", req, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}

// CMS NEW KMB Tools godoc
// @Description Api Mapping Cluster
// @Tags Mapping Cluster
// @Produce octet-stream
// @Success 200 {file} file "application/octet-stream"
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/mapping-cluster/download [get]
func (c *handlerCMS) DownloadMappingCluster(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
		genName     string
	)

	defer func() {
		if genName != "" {
			os.Remove(fmt.Sprintf("./%s.xlsx", genName))
		}
	}()

	genName, filename, err := c.usecase.GenerateExcelMappingCluster()

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster", nil, err)
	}

	return ctx.Attachment(fmt.Sprintf("./%s.xlsx", genName), filename)
}

// CMS NEW KMB Tools godoc
// @Description Api Mapping Cluster
// @Tags Mapping Cluster
// @Produce json
// @Param excel_file formData file true "upload file"
// @Param user_id formData string true "user id"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/mapping-cluster/upload [post]
func (c *handlerCMS) UploadMappingCluster(ctx echo.Context) (err error) {

	var (
		accessToken = middlewares.UserInfoData.AccessToken
		req         request.ReqUploadMappingCluster
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Update Mapping Cluster", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Update Mapping Cluster", req, err)
	}

	file, err := ctx.FormFile("excel_file")
	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Update Mapping Cluster", nil, errors.New(constant.ERROR_BAD_REQUEST+" - Silakan unggah file excel yang valid"))
	}

	src, err := file.Open()
	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Update Mapping Cluster", nil, errors.New(constant.ERROR_BAD_REQUEST+" - Silakan unggah file excel yang valid"))
	}
	defer src.Close()

	mime := file.Header.Get("Content-Type")
	if mime != "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Update Mapping Cluster", nil, errors.New(constant.ERROR_BAD_REQUEST+" - Silakan unggah file berformat .xlsx"))
	}

	err = c.usecase.UpdateMappingCluster(req, src)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Update Mapping Cluster", nil, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Upload Success", nil, nil)
}

// CMS NEW KMB Tools godoc
// @Description Api Mapping Cluster
// @Tags Mapping Cluster
// @Produce json
// @Param branch_id query string false "branch_id"
// @Param branch_name query string false "branch_name"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/mapping-cluster/branch [get]
func (c *handlerCMS) MappingClusterBranch(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	req := request.ReqListMappingClusterBranch{
		BranchID:   ctx.QueryParam("branch_id"),
		BranchName: ctx.QueryParam("branch_name"),
	}

	data, err := c.usecase.GetMappingClusterBranch(req)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Get Mapping Cluster Branch", nil, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Get Mapping Cluster Branch", nil, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Get Mapping Cluster Branch", nil, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    len(data),
	})
}

// CMS NEW KMB Tools godoc
// @Description Api Mapping Cluster
// @Tags Mapping Cluster
// @Produce json
// @Param page query string false "page"
// @Success 200 {object} response.ApiResponse{data=response.InquiryRow}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cms/mapping-cluster/change-log [get]
func (c *handlerCMS) MappingClusterChangeLog(ctx echo.Context) (err error) {

	var accessToken = middlewares.UserInfoData.AccessToken

	page, _ := strconv.Atoi(ctx.QueryParam("page"))
	pagination := request.RequestPagination{
		Page:  page,
		Limit: 10,
	}

	data, rowTotal, err := c.usecase.GetMappingClusterChangeLog(pagination)

	if err != nil && err.Error() == constant.RECORD_NOT_FOUND {
		return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", nil, response.InquiryRow{Inquiry: data})
	}

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", nil, err)
	}

	return c.Json.SuccessV2(ctx, accessToken, constant.NEW_KMB_LOG, "LOS - Mapping Cluster Inquiry", nil, response.InquiryRow{
		Inquiry:        data,
		RecordFiltered: len(data),
		RecordTotal:    rowTotal,
	})
}
