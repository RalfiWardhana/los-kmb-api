package http

import (
	"fmt"
	"los-kmb-api/domain/principle/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/utils"

	_ "github.com/KB-FMF/los-common-library/errors"
	"github.com/KB-FMF/los-common-library/response"

	"github.com/labstack/echo/v4"
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
	principleRoute.POST("/verify-asset", handler.VerifyAsset, middlewares.AccessMiddleware())
	principleRoute.POST("/verify-pemohon", handler.VerifyPemohon, middlewares.AccessMiddleware())
	principleRoute.GET("/step-principle/:id_number", handler.StepPrinciple, middlewares.AccessMiddleware())
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

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.usecase.CheckNokaNosin(ctx.Request().Context(), r)

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
// @Param body body request.PrinciplePemohon true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.UsecaseApi}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/verify-pemohon [post]
func (c *handler) VerifyPemohon(ctx echo.Context) (err error) {

	var r request.PrinciplePemohon

	if err = ctx.Bind(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "799"), err)
	}
	if err = ctx.Validate(&r); err != nil {
		return c.responses.BadRequest(ctx, fmt.Sprintf("PRINCIPLE-%s", "800"), err)
	}

	data, err := c.multiusecase.PrinciplePemohon(ctx.Request().Context(), r)

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

		return c.responses.Error(ctx, fmt.Sprintf("PRINCIPLE-%s", code), err)
	}

	return c.responses.Result(ctx, fmt.Sprintf("PRINCIPLE-%s", "001"), data)

}
