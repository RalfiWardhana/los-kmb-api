package http

import (
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"

	"github.com/labstack/echo/v4"
)

type handlerKmbFiltering struct {
	multiusecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	Json         common.JSON
}

func FilteringHandler(kmbroute *echo.Group, multiUsecase interfaces.MultiUsecase, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware) {
	handler := handlerKmbFiltering{
		multiusecase: multiUsecase,
		usecase:      usecase,
		repository:   repository,
		Json:         json,
	}
	kmbroute.POST("/filtering", handler.Filtering, middlewares.AccessMiddleware())
}

// KmbFiltering Tools godoc
// @Description KmbFiltering
// @Tags Tools
// @Produce json
// @Param body body request.Filtering true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.DupcheckResult}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/filtering [post]
func (c *handlerKmbFiltering) Filtering(ctx echo.Context) (err error) {

	var (
		req     request.Filtering
		married bool
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
	}

	if req.Spouse != nil {

		var genderSpouse request.GenderCompare

		if req.Gender != req.Spouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := ctx.Validate(&genderSpouse); err != nil {
			return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
		}

		married = true
	}

	check, err := c.usecase.FilteringProspectID(req.ProspectID)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
	}

	if err := ctx.Validate(&check); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
	}

	data, err := c.multiusecase.Filtering(ctx.Request().Context(), req, married, middlewares.UserInfoData.AccessToken)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, err)
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", req, data)
}
