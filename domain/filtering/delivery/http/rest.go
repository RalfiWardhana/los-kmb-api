package http

import (
	"los-kmb-api/domain/filtering/interfaces"
	middlewares "los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	constants "los-kmb-api/shared/constant"

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
// @Param body body request.BodyRequest true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.DupcheckResult}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /kmb-filtering [post]
func (c *handlerKmbFiltering) Filtering(ctx echo.Context) (err error) {

	var req request.BodyRequest

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerError(ctx, err)
	}

	if err := ctx.Validate(&req); err != nil {
		return c.Json.BadRequestErrorValidation(ctx, err)
	}

	accessToken := middlewares.UserInfoData.AccessToken

	data, err := c.multiusecase.Filtering(req, accessToken)

	if req.Data.MaritalStatus == constants.MARRIED {

		var genderSpouse request.GenderCompare

		if req.Data.Gender != req.Data.Spouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := ctx.Validate(&genderSpouse); err != nil {
			return c.Json.BadRequestErrorValidation(ctx, err)
		}

	}

	if err != nil {
		return c.Json.ServiceUnavailable(ctx)
	}

	return c.Json.Ok(ctx, data)
}
