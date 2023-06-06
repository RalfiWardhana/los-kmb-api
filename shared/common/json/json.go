package json

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	models "los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/iancoleman/strcase"
	"github.com/labstack/echo/v4"
	"gopkg.in/go-playground/validator.v9"
)

type (
	response struct {
	}
)

func NewResponse() common.JSON {
	return &response{}
}

func (c *response) Ok(ctx echo.Context, data interface{}) error {
	return ctx.JSON(http.StatusOK, models.ApiResponse{
		Message:    constant.MESSAGE_KMB_FILTERING,
		Errors:     nil,
		Data:       data,
		ServerTime: utils.GenerateTimeNow(),
	})
}

func (c *response) ServiceUnavailable(ctx echo.Context) error {
	return ctx.JSON(http.StatusServiceUnavailable, models.ApiResponse{
		Message:    constant.MESSAGE_KMB_FILTERING,
		Errors:     "service_unavailable",
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	})
}

func (c *response) InternalServerError(ctx echo.Context, err error) error {
	apiError := handleInternalError(err)
	return ctx.JSON(http.StatusInternalServerError, models.ApiResponse{
		Message:    constant.MESSAGE_KMB_FILTERING + " - " + apiError,
		Errors:     "internal_server_error",
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	})
}

func (c *response) BadRequestErrorValidation(ctx echo.Context, err error) error {
	var errors = make([]models.ErrorValidation, len(err.(validator.ValidationErrors)))

	for k, v := range err.(validator.ValidationErrors) {
		errors[k] = models.ErrorValidation{
			Field:   strcase.ToSnake(v.Field()),
			Message: formatMessage(v),
		}
	}
	return ctx.JSON(http.StatusBadRequest, models.ApiResponse{
		Message:    constant.MESSAGE_KMB_FILTERING,
		Errors:     errors,
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	})
}

func (c *response) BadGateway(ctx echo.Context, message string) error {
	return ctx.JSON(http.StatusBadGateway, models.ApiResponse{
		Message:    message,
		Errors:     "upstream_service_error",
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	})
}

func handleUnmarshalError(err error) []models.ErrorValidation {
	var apiErrors []models.ErrorValidation
	if he, ok := err.(*echo.HTTPError); ok {
		if ute, ok := he.Internal.(*json.UnmarshalTypeError); ok {
			valError := models.ErrorValidation{
				Field:   ute.Field,
				Message: ute.Error(),
			}
			apiErrors = append(apiErrors, valError)
		}
		if se, ok := he.Internal.(*json.SyntaxError); ok {
			valError := models.ErrorValidation{
				Field:   "Syntax Error",
				Message: se.Error(),
			}
			apiErrors = append(apiErrors, valError)
		}
		if iue, ok := he.Internal.(*json.InvalidUnmarshalError); ok {
			valError := models.ErrorValidation{
				Field:   iue.Type.String(),
				Message: iue.Error(),
			}
			apiErrors = append(apiErrors, valError)
		}
	}
	return apiErrors
}

func handleInternalError(err error) (apiErrors string) {

	if he, ok := err.(*echo.HTTPError); ok {
		if _, ok := he.Internal.(*json.UnmarshalTypeError); ok {
			apiErrors = "Unmarshal Type Error"
			return
		}
		if _, ok := he.Internal.(*json.SyntaxError); ok {
			apiErrors = "Syntax Error"
			return
		}
		if _, ok := he.Internal.(*json.InvalidUnmarshalError); ok {
			apiErrors = "Invalid Unmarshal Error"
			return
		}

		if strings.Contains(err.Error(), "unexpected EOF") {
			apiErrors = "Unexpected EOF"
			return
		}

		if strings.Contains(err.Error(), "unexpected end") {
			apiErrors = "Unexpected end Of JSON Input"
			return
		}

	}

	apiErrors = "Other"
	return
}

func (c *response) BadRequestWithSpecificFieldsV2(ctx echo.Context, accessToken, logFile, msg string, fields ...[]string) error {
	errors := make([]models.ErrorValidation, len(fields))

	for key, val := range fields {
		errors[key] = formatMessage2(val[0], val[1])
	}

	apiResponse := models.ApiResponse{
		Message:    msg,
		Errors:     errors,
		ServerTime: utils.GenerateTimeNow(),
	}

	_ = common.CentralizeLog(ctx.Request().Context(), accessToken, common.CentralizeLogParameter{
		LogFile:    logFile,
		MsgLogFile: constant.MSG_INCOMING_REQUEST,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Response:   apiResponse,
	})

	return ctx.JSON(http.StatusBadRequest, apiResponse)
}

func formatMessage(err validator.FieldError) string {

	param := err.Param()

	message := fmt.Sprintf("Field validation for '%s' failed on the '%s'", strcase.ToSnake(err.Field()), err.Tag())

	switch err.Tag() {
	case constant.TAG_REQUIRED:
		message = "required"
	case constant.TAG_DATE_FORMAT:
		message = "accepted:format=YYYY-MM-DD"
	case constant.TAG_LEN:
		message = fmt.Sprintf("accepted:len=%s", param)
	case constant.TAG_ALLOW_CHARS_NAME:
		message = "accepted:value=A-Z,a-z.'` "
	case constant.TAG_MIN:
		message = fmt.Sprintf("accepted:min=%s", param)
	}
	return message
}

func formatMessage2(field string, tag string) models.ErrorValidation {
	errors := models.ErrorValidation{
		Field:   strcase.ToSnake(field),
		Message: "required",
	}

	switch tag {
	case "required":
		errors.Message = "required"
	case "mobilephone":
		errors.Message = "accepted:start=08XXXXXXXXXX"
	case "minphone":
		errors.Message = "accepted:min=10"
	case "maxphone":
		errors.Message = "accepted:max=14"
	case "number":
		errors.Message = "accepted:number"
	case "encryption":
		errors.Message = "invalid encryption"
	}
	return errors
}
