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
}
