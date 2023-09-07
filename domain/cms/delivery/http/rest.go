package http

import (
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/shared/common"

	"github.com/labstack/echo/v4"
)

type handlerCMS struct {
	usecase    interfaces.Usecase
	repository interfaces.Repository
	Json       common.JSON
}

func CMSHandler(cmsroute *echo.Group, usecase interfaces.Usecase, repository interfaces.Repository, json common.JSON, middlewares *middlewares.AccessMiddleware) {
	handler := handlerCMS{
		usecase:    usecase,
		repository: repository,
		Json:       json,
	}

	cmsroute.POST("/cms/prescreening/inquiry", handler.PrescreeningInquiry, middlewares.AccessMiddleware())
}

// CMS NEW KMB Tools godoc
// @Description Api Prescreening
// @Tags Prescreening
// @Produce json
// @Param body body request. true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v2/kmb/cms/prescreening/inquiry [post]
func (c *handlerCMS) PrescreeningInquiry(ctx echo.Context) (err error) {

	return
}
