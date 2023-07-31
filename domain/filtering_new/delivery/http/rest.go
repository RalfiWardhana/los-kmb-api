package http

import (
	"errors"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/labstack/echo/v4"
)

type handlerKmbFiltering struct {
	multiusecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	Json         common.JSON
	producer     platformevent.PlatformEvent
}

func FilteringHandler(kmbroute *echo.Group, multiUsecase interfaces.MultiUsecase, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware,
	producer platformevent.PlatformEvent) {
	handler := handlerKmbFiltering{
		multiusecase: multiUsecase,
		usecase:      usecase,
		repository:   repository,
		Json:         json,
		producer:     producer,
	}
	kmbroute.POST("/filtering", handler.Filtering, middlewares.AccessMiddleware())
	kmbroute.POST("/produce/filtering", handler.ProduceFiltering, middlewares.AccessMiddleware())
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
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", err)
	}

	// decrypt request
	req.IDNumber, err = utils.PlatformDecryptText(req.IDNumber)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Decrypt Error")
		return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
	}
	req.LegalName, err = utils.PlatformDecryptText(req.LegalName)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Decrypt Error")
		return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
	}
	req.MotherName, err = utils.PlatformDecryptText(req.MotherName)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Decrypt Error")
		return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
	}

	if req.Spouse == nil {
		if err := ctx.Validate(&req); err != nil {
			return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
		}
	} else {
		req.Spouse.IDNumber, err = utils.PlatformDecryptText(req.Spouse.IDNumber)
		if err != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Decrypt Error")
			return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
		}
		req.Spouse.LegalName, err = utils.PlatformDecryptText(req.Spouse.LegalName)
		if err != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Decrypt Error")
			return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
		}
		req.Spouse.MotherName, err = utils.PlatformDecryptText(req.Spouse.MotherName)
		if err != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Decrypt Error")
			return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
		}

		if err := ctx.Validate(&req); err != nil {
			return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
		}

		var genderSpouse request.GenderCompare

		if req.Gender != req.Spouse.Gender {
			genderSpouse.Gender = true
		} else {
			genderSpouse.Gender = false
		}

		if err := ctx.Validate(&genderSpouse); err != nil {
			return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
		}

		married = true
	}

	check, err := c.usecase.FilteringProspectID(req.ProspectID)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
	}

	if err := ctx.Validate(&check); err != nil {
		return c.Json.BadRequestErrorValidationV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
	}

	data, err := c.multiusecase.Filtering(ctx.Request().Context(), req, married, middlewares.UserInfoData.AccessToken)

	if err != nil {
		return c.Json.ServerSideErrorV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, err)
	}

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", req, data)
}

func (c *handlerKmbFiltering) ProduceFiltering(ctx echo.Context) (err error) {

	var (
		req request.Filtering
	)

	if err := ctx.Bind(&req); err != nil {
		return c.Json.InternalServerErrorCustomV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING", err)
	}

	c.producer.PublishEvent(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION, constant.KEY_PREFIX_FILTERING, req.ProspectID, utils.StructToMap(req), 0)

	return c.Json.SuccessV2(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB FILTERING - Please wait, your request is being processed", req, nil)
}
