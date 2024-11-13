package http

import (
	"errors"
	"fmt"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/platformcache"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"

	"github.com/labstack/echo/v4"
)

type handlerKmbFiltering struct {
	multiusecase interfaces.MultiUsecase
	usecase      interfaces.Usecase
	repository   interfaces.Repository
	Json         common.JSON
	producer     platformevent.PlatformEventInterface
	cache        platformcache.PlatformCacheInterface
}

func FilteringHandler(kmbroute *echo.Group, multiUsecase interfaces.MultiUsecase, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware,
	producer platformevent.PlatformEventInterface, cache platformcache.PlatformCacheInterface) {
	handler := handlerKmbFiltering{
		multiusecase: multiUsecase,
		usecase:      usecase,
		repository:   repository,
		Json:         json,
		producer:     producer,
		cache:        cache,
	}
	kmbroute.POST("/produce/filtering", handler.ProduceFiltering, middlewares.AccessMiddleware())
	kmbroute.DELETE("/cache/filtering/:prospect_id", handler.RemoveCacheFiltering, middlewares.AccessMiddleware())
	kmbroute.GET("/employee/employee-data/:employee_id", handler.GetEmployeeData, middlewares.AccessMiddleware())
}

// Produce Filtering Tools godoc
// @Description Produce Filtering via REST API
// @Tags Filtering
// @Produce json
// @Param body body request.Filtering true "Body payload"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/produce/filtering [post]
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

// Remove Cache Filtering Tools godoc
// @Description Remove Cache Filtering via REST API
// @Tags Filtering
// @Produce json
// @Param prospect_id path string true "Prospect ID"
// @Success 200 {object} response.ApiResponse{}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/cache/filtering/{prospect_id} [delete]
func (c *handlerKmbFiltering) RemoveCacheFiltering(ctx echo.Context) (err error) {

	var (
		ctxJson    error
		prospectID string
	)

	prospectID = ctx.Param("prospect_id")

	if prospectID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID does not exist")
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB REMOVE CACHE FILTERING", prospectID, err)
		return ctxJson
	}

	_, err = c.cache.SetCache(ctx.Request().Context(), middlewares.UserInfoData.AccessToken, os.Getenv("CACHE_COLLECTION_NAME"), fmt.Sprintf(constant.DOC_FILTERING, prospectID), map[string]string{"prospect_id": prospectID}, "1")
	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB REMOVE CACHE FILTERING", prospectID, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB REMOVE CACHE FILTERING - SUCCESS", prospectID, nil)
	return ctxJson
}

// CMS NEW KMB Tools godoc
// @Description Api Get Employee Data
// @Tags Employee
// @Produce json
// @Param employee_id path string true "Employee ID"
// @Success 200 {object} response.ApiResponse{data=response.EmployeeCMOResponse}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/employee/employee-data/{employee_id} [get]
func (c *handlerKmbFiltering) GetEmployeeData(ctx echo.Context) (err error) {

	var (
		accessToken     = middlewares.UserInfoData.AccessToken
		hrisAccessToken = middlewares.HrisApiData.Token
		ctxJson         error
	)

	employeeID := ctx.Param("employee_id")

	if employeeID == "" {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - EmployeeID does not exist")
		ctxJson, _ = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - GET EMPLOYEE DATA", employeeID, err)
		return ctxJson
	}

	data, err := c.usecase.GetEmployeeData(ctx.Request().Context(), employeeID, accessToken, hrisAccessToken)

	if err != nil {
		ctxJson, _ = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - GET EMPLOYEE DATA", employeeID, err)
		return ctxJson
	}

	ctxJson, _ = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - GET EMPLOYEE DATA", employeeID, data)
	return ctxJson
}
