package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"los-kmb-api/domain/principle/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/http"

	_ "github.com/KB-FMF/los-common-library/errors"
	"github.com/KB-FMF/los-common-library/response"
	"golang.org/x/time/rate"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type handler struct {
	multiusecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	responses    response.Response
}

func Handler(principleRoute *echo.Group, multiusecase interfaces.MultiUsecase, usecase interfaces.Usecase, repository interfaces.Repository, responses response.Response, middlewares *middlewares.AccessMiddleware) {
	handler := handler{
		multiusecase: multiusecase,
		usecase:      usecase,
		repository:   repository,
		responses:    responses,
	}

	// Rate limiter configuration with a limit of 20 requests per second
	limiter := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStore(rate.Limit(3)), // Limit to 1 requests per second
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"message":     "Too many requests. Please try again after a few seconds.",
				"errors":      "Too many requests",
				"code":        "LOS-PRINCIPLE-429",
				"data":        nil,
				"server_time": utils.GenerateTimeNow(),
			})
		},
	})

	principleRoute.POST("/verify-asset", handler.VerifyAsset, middlewares.AccessMiddleware(), limiter)
	principleRoute.POST("/verify-pemohon", handler.VerifyPemohon, middlewares.AccessMiddleware(), limiter)
	principleRoute.GET("/step-principle/:id_number", handler.StepPrinciple, middlewares.AccessMiddleware())
	principleRoute.POST("/elaborate-ltv", handler.ElaborateLTV, middlewares.AccessMiddleware())
	principleRoute.POST("/verify-pembiayaan", handler.VerifyPembiayaan, middlewares.AccessMiddleware(), limiter)
	principleRoute.POST("/emergency-contact", handler.EmergencyContact, middlewares.AccessMiddleware(), limiter)
	principleRoute.POST("/core-customer/:prospectID", handler.CoreCustomer, middlewares.AccessMiddleware())
	principleRoute.POST("/marketing-program/:prospectID", handler.MarketingProgram, middlewares.AccessMiddleware())
	principleRoute.POST("/principle-data", handler.GetPrincipleData, middlewares.AccessMiddleware())
	principleRoute.GET("/auto-cancel", handler.AutoCancel, middlewares.AccessMiddleware())
	principleRoute.POST("/principle-publish", handler.PrinciplePublish, middlewares.AccessMiddleware())
}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param body body request.PrincipleAsset true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.UsecaseApi}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/verify-asset [post]
func (c *handler) VerifyAsset(ctx echo.Context) (err error) {

	var r request.PrincipleAsset

	defer func() {
		body, _ := json.Marshal(r)
		ctx.Request().Body = io.NopCloser(bytes.NewBuffer(body))
	}()

	ctx, err = utils.Sanitized(ctx)

	if err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.usecase.CheckNokaNosin(ctx.Request().Context(), r)

	if err != nil {

		if err.Error() == constant.ERROR_MAX_EXCEED {
			return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", "429"), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(constant.PRINCIPLE_ERROR_EXCEED_RESPONSE_MESSAGE))
		}

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(constant.PRINCIPLE_ERROR_RESPONSE_MESSAGE))
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data, response.WithMessage(data.Reason))

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param body body request.PrinciplePemohon true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.UsecaseApi}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/verify-pemohon [post]
func (c *handler) VerifyPemohon(ctx echo.Context) (err error) {

	var r request.PrinciplePemohon

	defer func() {
		body, _ := json.Marshal(r)
		ctx.Request().Body = io.NopCloser(bytes.NewBuffer(body))
	}()

	ctx, err = utils.Sanitized(ctx)

	if err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.multiusecase.PrinciplePemohon(ctx.Request().Context(), r)

	if err != nil {

		if err.Error() == constant.ERROR_MAX_EXCEED {
			return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", "429"), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(constant.PRINCIPLE_ERROR_EXCEED_RESPONSE_MESSAGE))
		}

		errorMessage := constant.PRINCIPLE_ERROR_RESPONSE_MESSAGE
		if err.Error() == constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE {
			errorMessage = constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE
		}

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(errorMessage))
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data, response.WithMessage(data.Reason))

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param id_number path string true "ID Number"
// @Success 200 {object} response.ApiResponse{data=response.StepPrinciple}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/step-principle/{id_number} [get]
func (c *handler) StepPrinciple(ctx echo.Context) (err error) {

	var validate = request.ValidateNik{
		IDNumber: ctx.Param("id_number"),
	}

	if err = ctx.Bind(&validate); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&validate); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.usecase.PrincipleStep(validate.IDNumber)

	if err != nil {

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(constant.PRINCIPLE_ERROR_RESPONSE_MESSAGE))
	}

	if data.Status == "" {
		return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "003"), nil)
	}

	if data.Status == constant.REASON_PROSES_SURVEY {
		return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "002"), data, response.WithMessage("Kamu masih memiliki pengajuan lain yang sedang diproses"))

	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data)

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param body body request.PrincipleElaborateLTV true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.UsecaseApi}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/elaborate-ltv [post]
func (c *handler) ElaborateLTV(ctx echo.Context) (err error) {

	var r request.PrincipleElaborateLTV

	defer func() {
		body, _ := json.Marshal(r)
		ctx.Request().Body = io.NopCloser(bytes.NewBuffer(body))
	}()

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.usecase.PrincipleElaborateLTV(ctx.Request().Context(), r, middlewares.UserInfoData.AccessToken)

	if err != nil {

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(constant.PRINCIPLE_ERROR_RESPONSE_MESSAGE))
	}

	if data.AdjustTenor && data.LTV == 0 {
		return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "002"), data, response.WithMessage(data.Reason))
	}

	if !data.AdjustTenor && data.LTV == 0 {
		return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "003"), data, response.WithMessage(data.Reason))
	}

	if r.LoanAmount > 0 {
		return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "004"), data)
	}
	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data)

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param body body request.PrinciplePembiayaan true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.UsecaseApi}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/verify-pembiayaan [post]
func (c *handler) VerifyPembiayaan(ctx echo.Context) (err error) {

	var r request.PrinciplePembiayaan

	defer func() {
		body, _ := json.Marshal(r)
		ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}()

	ctx, err = utils.Sanitized(ctx)

	if err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.multiusecase.PrinciplePembiayaan(ctx.Request().Context(), r, middlewares.UserInfoData.AccessToken)

	if err != nil {

		if err.Error() == constant.ERROR_MAX_EXCEED {
			return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", "429"), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(constant.PRINCIPLE_ERROR_EXCEED_RESPONSE_MESSAGE))
		}

		errorMessage := constant.PRINCIPLE_ERROR_RESPONSE_MESSAGE
		if err.Error() == constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE {
			errorMessage = constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE
		}

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(errorMessage))
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data, response.WithMessage(data.Reason))

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param body body request.PrincipleEmergencyContact true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.UsecaseApi}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/emergency-contact [post]
func (c *handler) EmergencyContact(ctx echo.Context) (err error) {

	var r request.PrincipleEmergencyContact

	defer func() {
		body, _ := json.Marshal(r)
		ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}()

	ctx, err = utils.Sanitized(ctx)

	if err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.multiusecase.PrincipleEmergencyContact(ctx.Request().Context(), r, middlewares.UserInfoData.AccessToken)

	if err != nil {
		errorMessage := constant.PRINCIPLE_ERROR_RESPONSE_MESSAGE
		if err.Error() == constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE {
			errorMessage = constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE
		}

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err, response.WithHttpCode(http.StatusInternalServerError), response.WithMessage(errorMessage))
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data)

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param prospectID path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/core-customer/{prospectID} [post]
func (c *handler) CoreCustomer(ctx echo.Context) (err error) {

	prospectID := ctx.Param("prospectID")

	if prospectID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID does not exist")
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}

	err = c.usecase.PrincipleCoreCustomer(ctx.Request().Context(), prospectID, middlewares.UserInfoData.AccessToken)

	if err != nil {

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err)
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), nil)

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param prospectID path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/marketing-program/{prospectID} [post]
func (c *handler) MarketingProgram(ctx echo.Context) (err error) {

	prospectID := ctx.Param("prospectID")

	if prospectID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID does not exist")
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}

	err = c.usecase.PrincipleMarketingProgram(ctx.Request().Context(), prospectID, middlewares.UserInfoData.AccessToken)

	if err != nil {

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err)
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), nil)

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param prospectID path string true "Prospect ID"
// @Param body body request.PrincipleGetData true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/principle-data [post]
func (c *handler) GetPrincipleData(ctx echo.Context) (err error) {

	var r request.PrincipleGetData

	defer func() {
		body, _ := json.Marshal(r)
		ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}()

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.usecase.GetDataPrinciple(ctx.Request().Context(), r, middlewares.UserInfoData.AccessToken)

	if err != nil {

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err)
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data)

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param prospectID path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/auto-cancel [get]
func (c *handler) AutoCancel(ctx echo.Context) (err error) {

	err = c.usecase.CheckOrderPendingPrinciple(ctx.Request().Context())

	if err != nil {

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err)
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), "sukses auto cancel")

}

// KmbPrinciple Tools godoc
// @Description KmbPrinciple
// @Tags KmbPrinciple
// @Produce json
// @Param body body request.PrinciplePublish true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/principle-publish [post]
func (c *handler) PrinciplePublish(ctx echo.Context) (err error) {

	var r request.PrinciplePublish

	defer func() {
		body, _ := json.Marshal(r)
		ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}()

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	err = c.usecase.PrinciplePublish(ctx.Request().Context(), r, middlewares.UserInfoData.AccessToken)

	if err != nil {

		code, err := utils.WrapError(err)

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err)
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), "success publish event principle")

}
