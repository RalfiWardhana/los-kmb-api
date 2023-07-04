package http

import (
	"los-kmb-api/domain/kmb/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"

	"github.com/labstack/echo/v4"
)

type handlerKMOB struct {
	metrics      interfaces.Metrics
	multiUsecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	Json         common.JSON
}

func KMBHandler(kmbroute *echo.Group, metrics interfaces.Metrics, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware) {
	handler := handlerKMOB{
		metrics:    metrics,
		usecase:    usecase,
		repository: repository,
		Json:       json,
	}
	kmbroute.POST("/dupcheck", handler.Dupcheck, middlewares.AccessMiddleware())
}

// KmbDupcheck Tools godoc
// @Description KmbDupcheck
// @Tags Tools
// @Produce json
// @Param body body request.DupcheckApi true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.DupcheckResult}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /dupcheck [post]
func (c *handlerKMOB) Dupcheck(ctx echo.Context) (err error) {

	var (
		req     request.DupcheckApi
		married bool
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB DUPCHECK", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB DUPCHECK", req, err)
	}

	if req.Spouse != nil {

		if err := ctx.Validate(req.Spouse); err != nil {
			return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB DUPCHECK", req, err)

		}

		var genderSpouse request.GenderCompare

		if req.Gender != req.Spouse.Gender {
			genderSpouse.Gender = true

		} else {
			genderSpouse.Gender = false
		}

		if err := ctx.Validate(&genderSpouse); err != nil {
			return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB DUPCHECK", req, err)
		}

		married = true
	}

	accessToken := middlewares.UserInfoData.AccessToken

	_, _, data, err := c.multiUsecase.Dupcheck(ctx.Request().Context(), req, married, accessToken)

	if err != nil {
		return c.Json.ServiceUnavailableV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB DUPCHECK", req)
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB DUPCHECK", req, data)
}
