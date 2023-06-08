package json

import (
	"encoding/json"
	"fmt"
	models "los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/http"
	"strings"

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

func (c *response) SuccessV2(ctx echo.Context, accessToken, logFile, message string, req, data interface{}) error {

	//create response
	apiResponse := models.ApiResponse{
		Message:    message,
		Errors:     nil,
		Data:       data,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Get(echo.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx.Request().Context(), accessToken, common.CentralizeLogParameter{
		LogFile:    logFile,
		MsgLogFile: constant.MSG_INCOMING_REQUEST,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    req,
		Response:   apiResponse,
	})

	return ctx.JSON(http.StatusOK, apiResponse)
}

func (c *response) ServiceUnavailableV2(ctx echo.Context, accessToken, logFile, message string, req interface{}) error {

	apiResponse := models.ApiResponse{
		Message:    message,
		Errors:     "service_unavailable",
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Get(echo.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx.Request().Context(), accessToken, common.CentralizeLogParameter{
		LogFile:    logFile,
		MsgLogFile: constant.MSG_INCOMING_REQUEST,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
		Request:    req,
		Response:   apiResponse,
	})

	return ctx.JSON(http.StatusServiceUnavailable, apiResponse)
}

func (c *response) InternalServerErrorCustomV2(ctx echo.Context, accessToken, logFile, message string, err error) error {
	apiError := handleInternalError(err)

	apiResponse := models.ApiResponse{
		Message:    message + " - " + apiError,
		Errors:     constant.INTERNAL_SERVER_ERROR,
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Get(echo.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx.Request().Context(), accessToken, common.CentralizeLogParameter{
		LogFile:    logFile,
		MsgLogFile: constant.MSG_INCOMING_REQUEST,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
		Response:   apiResponse,
	})
	return ctx.JSON(http.StatusInternalServerError, apiResponse)
}

func (c *response) BadRequestErrorValidationV2(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) error {
	var errors = make([]models.ErrorValidation, len(err.(validator.ValidationErrors)))

	for k, v := range err.(validator.ValidationErrors) {
		field := strcase.ToSnake(v.Field())

		errors[k] = models.ErrorValidation{
			Field:   field,
			Message: formatMessage(v),
		}

	}
	apiResponse := models.ApiResponse{
		Message:    message,
		Errors:     errors,
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Get(echo.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx.Request().Context(), accessToken, common.CentralizeLogParameter{
		LogFile:    logFile,
		MsgLogFile: constant.MSG_INCOMING_REQUEST,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
		Request:    req,
		Response:   apiResponse,
	})

	return ctx.JSON(http.StatusBadRequest, apiResponse)
}

func (c *response) BadGateway(ctx echo.Context, message string) error {

	apiResponse := models.ApiResponse{
		Message:    message,
		Errors:     "upstream_service_error",
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Get(echo.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	return ctx.JSON(http.StatusBadGateway, apiResponse)
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
