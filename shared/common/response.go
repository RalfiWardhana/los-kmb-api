package common

import (
	"context"
	models "los-kmb-api/models/response"

	"github.com/labstack/echo/v4"
)

type JSON interface {
	SuccessV2(ctx echo.Context, accessToken, logFile, message string, req, data interface{}) error
	ServiceUnavailableV2(ctx echo.Context, accessToken, logFile, message string, req interface{}) error
	InternalServerErrorCustomV2(ctx echo.Context, accessToken, logFile, message string, err error) error
	BadRequestErrorValidationV2(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) error
	BadGateway(ctx echo.Context, message string) error
	ServerSideErrorV2(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) error
	EventSuccess(ctx context.Context, accessToken, logFile, message string, req, data interface{}) (apiResponse models.ApiResponse)
	EventServiceError(ctx context.Context, accessToken, logFile, message string, req interface{}, err error) (apiResponse models.ApiResponse)
	EventBadRequestErrorValidation(ctx context.Context, accessToken, logFile, message string, req interface{}, err error) (apiResponse models.ApiResponse)
	EventRequestErrorBindV3(ctx context.Context, accessToken, logFile, message string, req interface{}, err error) (apiResponse models.ApiResponse)

	SuccessV3(ctx echo.Context, accessToken, logFile, message string, req, data interface{}) (ctxJson error, apiResponse models.ApiResponse)
	ServiceUnavailableV3(ctx echo.Context, accessToken, logFile, message string, req interface{}) (ctxJson error, apiResponse models.ApiResponse)
	InternalServerErrorCustomV3(ctx echo.Context, accessToken, logFile, message string, err error) (ctxJson error, apiResponse models.ApiResponse)
	BadRequestErrorBindV3(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponse)
	BadRequestErrorValidationV3(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponse)
	ServerSideErrorV3(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponse)

	ErrorStandard(ctx echo.Context, accessToken, logFile, code string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponseV2)
	ErrorBindStandard(ctx echo.Context, accessToken, logFile, code string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponseV2)
	ErrorValidationStandard(ctx echo.Context, accessToken, logFile, code string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponseV2)
	SuccessStandard(ctx echo.Context, accessToken, logFile, code string, req, data interface{}) (ctxJson error, apiResponse models.ApiResponseV2)
}
