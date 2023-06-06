package common

import "github.com/labstack/echo/v4"

type JSON interface {
	Ok(ctx echo.Context, data interface{}) error
	ServiceUnavailable(ctx echo.Context) error
	InternalServerError(ctx echo.Context, err error) error
	BadRequestErrorValidation(ctx echo.Context, err error) error
	BadRequestWithSpecificFieldsV2(ctx echo.Context, accessToken, logFile, msg string, fields ...[]string) error
	BadGateway(ctx echo.Context, message string) error
}
