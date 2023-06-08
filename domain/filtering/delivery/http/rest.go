package http

import (
	"los-kmb-api/domain/filtering/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"strconv"

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
// @Param body body request.FilteringRequest true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.DupcheckResult}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /filtering [post]
func (c *handlerKmbFiltering) Filtering(ctx echo.Context) (err error) {

	var r request.FilteringRequest

	if err := ctx.Bind(&r); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", err)
	}

	if err := ctx.Validate(&r); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", r, err)
	}

	if r.Data.MaritalStatus == constant.MARRIED {

		var genderSpouse request.GenderCompare

		if r.Data.Gender != r.Data.Spouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := ctx.Validate(&genderSpouse); err != nil {
			return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", r, err)
		}

	}

	data, err := c.multiusecase.Filtering(ctx.Request().Context(), r, middlewares.UserInfoData.AccessToken)

	if err != nil {
		return c.Json.ServiceUnavailableV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", r)
	}

	data.Code, _ = strconv.Atoi(data.Code.(string))

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.FILTERING_LOG, "LOS - KMB FILTERING", r, data)
}
