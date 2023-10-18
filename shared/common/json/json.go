package json

import (
	"context"
	"encoding/json"
	"fmt"
	models "los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/http"
	"os"
	"strconv"
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

func (c *response) ServerSideErrorV2(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) error {
	var errors string
	var statusCode int

	handleError := strings.Split(err.Error(), " - ")

	if len(handleError) > 1 {
		message = fmt.Sprintf("%s - %s", message, handleError[1])
	} else {
		message = fmt.Sprintf("%s - %s", message, err.Error())
	}
	errors = handleError[0]

	switch handleError[0] {
	case constant.ERROR_UPSTREAM:
		statusCode = http.StatusBadGateway
	case constant.ERROR_UPSTREAM_TIMEOUT:
		statusCode = http.StatusGatewayTimeout
	case constant.ERROR_SERVICE_UNAVAILABLE:
		statusCode = http.StatusServiceUnavailable
	case constant.ERROR_BAD_REQUEST:
		statusCode = http.StatusBadRequest
	case constant.ERROR_DATA_CONFLICT:
		statusCode = http.StatusConflict
	default:
		statusCode = http.StatusServiceUnavailable
		errors = constant.ERROR_SERVICE_UNAVAILABLE
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

	return ctx.JSON(statusCode, apiResponse)
}

func handleUnmarshalError(err error) []models.ErrorValidation {
	var apiErrors []models.ErrorValidation
	if he, ok := err.(*echo.HTTPError); ok {
		if ute, ok := he.Internal.(*json.UnmarshalTypeError); ok {
			valError := models.ErrorValidation{
				Field:   ute.Field,
				Message: fmt.Sprintf("accepted:type=%s", ute.Type.String()),
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
	case constant.TAG_MAX:
		message = fmt.Sprintf("accepted:max=%s", param)
	case constant.TAG_GT:
		message = fmt.Sprintf("accepted:gt=%s", param)
	case constant.TAG_REQUIRED:
		message = "required"
	case constant.TAG_DATE_FORMAT:
		message = "accepted:format=YYYY-MM-DD"
	case constant.TAG_LEN:
		message = fmt.Sprintf("accepted:len=%s", param)
	case constant.TAG_URL:
		message = "accepted:format=https,http"
	case constant.TAG_ALLOW_CHARS_NAME:
		message = "accepted:value=A-Z,a-z.'` "
	case constant.TAG_MIN:
		message = fmt.Sprintf("accepted:min=%s", param)
	case constant.TAG_PROSPECTID:
		message = "accepted:value=A-Z, 0-9, min=10, max=20, First two characters must be alphabetic"
	case constant.TAG_NUMBER:
		message = "accepted:value=0-9"
	case constant.TAG_ID_NUMBER:
		message = "accepted:valid encryption, length=16, number only, non zero for first number"
	case constant.TAG_BRANCH_ID:
		message = "accepted:length=3, number only"
	case constant.TAG_ALLOW_NAME:
		message = "accepted:valid encryption, max=100, value=A-Z,a-z.'` "
	case constant.TAG_GENDER:
		message = fmt.Sprintf("accepted:value=%s", common.Gender)
	case constant.TAG_SPOUSE_GENDER:
		message = "gender must difference"
	case constant.TAG_STATUS_KONSUMEN:
		message = fmt.Sprintf("accepted:value=%s", common.StatusKonsumen)
	case constant.TAG_RECOM:
		message = "accepted:value=0,1"
	case constant.TAG_CHANNEL:
		message = fmt.Sprintf("accepted:value=%s", common.Channel)
	case constant.TAG_LOB:
		message = fmt.Sprintf("accepted:value=%s", common.Lob)
	case constant.TAG_INCOMING:
		message = fmt.Sprintf("accepted:value=%s", common.Incoming)
	case constant.TAG_HOME:
		message = fmt.Sprintf("accepted:value=%s", common.Home)
	case constant.TAG_BPKB_NAME:
		message = "accepted:value=K,P,O,KK"
	case constant.TAG_KTP:
		message = "accepted:value=KTP"
	case constant.TAG_ADDRESS:
		message = fmt.Sprintf("accepted:value=%s", common.Address)
	case constant.TAG_MARITAL:
		message = fmt.Sprintf("accepted:value=%s", common.Marital)
	case constant.TAG_EDUCATION:
		message = fmt.Sprintf("accepted:value=%s", common.Education)
	case constant.TAG_PROFESSION:
		message = fmt.Sprintf("accepted:value=%s", common.ProfID)
	case constant.TAG_PHOTO:
		message = fmt.Sprintf("accepted:value=%s", common.Photo)
	case constant.TAG_RELATIONSHIP:
		message = fmt.Sprintf("accepted:value=%s", common.Relationship)
	case constant.TAG_FTR_PROSPECTID:
		prospectID := strings.Split(err.Value().(string), " - ")[0]
		message = fmt.Sprintf("%s sebelumnya sudah masuk dan diproses", prospectID)
	case constant.TAG_TENOR:
		message = fmt.Sprintf("accepted:value=%s", common.Tenor)
	case constant.TAG_SPOUSE_NOT_NULL:
		message = "accepted:customer_spouse can't be null"
	case constant.TAG_SPOUSE_NULL:
		message = "accepted:customer_spouse must be null"
	case constant.TAG_RELATION:
		message = fmt.Sprintf("accepted:value=%s", common.Relation)
	}
	return message
}

func (c *response) EventSuccess(ctx context.Context, accessToken, logFile, message string, req, data interface{}) (apiResponse models.ApiResponse) {

	//create response
	apiResponse = models.ApiResponse{
		Message:    message,
		Errors:     nil,
		Data:       data,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Value(constant.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Method:     http.MethodPost,
		LogFile:    logFile,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    req,
		Response:   apiResponse,
	})

	return apiResponse
}

func (c *response) EventServiceError(ctx context.Context, accessToken, logFile, message string, req interface{}, err error) (apiResponse models.ApiResponse) {
	var (
		errors     string
		statusCode int
	)

	handleError := strings.Split(err.Error(), " - ")

	if len(handleError) > 1 {
		message = fmt.Sprintf("%s - %s", message, handleError[1])
	} else {
		message = fmt.Sprintf("%s - %s", message, err.Error())
	}
	errors = handleError[0]

	switch handleError[0] {
	case constant.ERROR_UPSTREAM:
		statusCode = http.StatusBadGateway
	case constant.ERROR_UPSTREAM_TIMEOUT:
		statusCode = http.StatusGatewayTimeout
	case constant.ERROR_SERVICE_UNAVAILABLE:
		statusCode = http.StatusServiceUnavailable
	case constant.ERROR_BAD_REQUEST:
		statusCode = http.StatusBadRequest
	case constant.ERROR_DATA_CONFLICT:
		statusCode = http.StatusConflict
	default:
		statusCode = http.StatusServiceUnavailable
		errors = constant.ERROR_SERVICE_UNAVAILABLE
	}

	apiResponse = models.ApiResponse{
		Message:    message,
		Errors:     strconv.Itoa(statusCode) + " " + errors,
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Value(constant.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Method:     http.MethodPost,
		LogFile:    logFile,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
		Request:    req,
		Response:   apiResponse,
	})

	return apiResponse
}

func (c *response) EventBadRequestErrorValidation(ctx context.Context, accessToken, logFile, message string, req interface{}, err error) (apiResponse models.ApiResponse) {
	var errors = make([]models.ErrorValidation, len(err.(validator.ValidationErrors)))

	for k, v := range err.(validator.ValidationErrors) {
		field := strcase.ToSnake(v.Field())

		errors[k] = models.ErrorValidation{
			Field:   field,
			Message: formatMessage(v),
		}

	}
	apiResponse = models.ApiResponse{
		Message:    message,
		Errors:     errors,
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Value(constant.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Method:     http.MethodPost,
		LogFile:    logFile,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
		Request:    req,
		Response:   apiResponse,
	})

	return apiResponse
}

func (c *response) EventRequestErrorBindV3(ctx context.Context, accessToken, logFile, message string, req interface{}, err error) (apiResponse models.ApiResponse) {
	errors := handleUnmarshalError(err)

	apiResponse = models.ApiResponse{
		Message:    message,
		Errors:     errors,
		Data:       nil,
		ServerTime: utils.GenerateTimeNow(),
	}
	requestID, ok := ctx.Value(constant.HeaderXRequestID).(string)
	if ok {
		apiResponse.RequestID = requestID
	}

	_ = common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Method:     http.MethodPost,
		LogFile:    logFile,
		MsgLogFile: constant.MSG_CONSUME_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
		Request:    req,
		Response:   apiResponse,
	})
	return apiResponse
}

func (c *response) SuccessV3(ctx echo.Context, accessToken, logFile, message string, req, data interface{}) (ctxJson error, apiResponse models.ApiResponse) {

	//create response
	apiResponse = models.ApiResponse{
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

	return ctx.JSON(http.StatusOK, apiResponse), apiResponse
}

func (c *response) ServiceUnavailableV3(ctx echo.Context, accessToken, logFile, message string, req interface{}) (ctxJson error, apiResponse models.ApiResponse) {

	apiResponse = models.ApiResponse{
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

	return ctx.JSON(http.StatusServiceUnavailable, apiResponse), apiResponse
}

func (c *response) InternalServerErrorCustomV3(ctx echo.Context, accessToken, logFile, message string, err error) (ctxJson error, apiResponse models.ApiResponse) {
	apiError := handleInternalError(err)

	apiResponse = models.ApiResponse{
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
	return ctx.JSON(http.StatusInternalServerError, apiResponse), apiResponse
}

func (c *response) BadRequestErrorBindV3(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponse) {
	errors := handleUnmarshalError(err)

	apiResponse = models.ApiResponse{
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
		Response:   apiResponse,
	})
	return ctx.JSON(http.StatusBadRequest, apiResponse), apiResponse
}

func (c *response) BadRequestErrorValidationV3(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponse) {
	var errors = make([]models.ErrorValidation, len(err.(validator.ValidationErrors)))

	for k, v := range err.(validator.ValidationErrors) {
		field := strcase.ToSnake(v.Field())

		errors[k] = models.ErrorValidation{
			Field:   field,
			Message: formatMessage(v),
		}

	}
	apiResponse = models.ApiResponse{
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

	return ctx.JSON(http.StatusBadRequest, apiResponse), apiResponse
}

func (c *response) ServerSideErrorV3(ctx echo.Context, accessToken, logFile, message string, req interface{}, err error) (ctxJson error, apiResponse models.ApiResponse) {
	var errors string
	var statusCode int

	handleError := strings.Split(err.Error(), " - ")

	if len(handleError) > 1 {
		message = fmt.Sprintf("%s - %s", message, handleError[1])
	} else {
		message = fmt.Sprintf("%s - %s", message, err.Error())
	}
	errors = handleError[0]

	switch handleError[0] {
	case constant.ERROR_UPSTREAM:
		statusCode = http.StatusBadGateway
	case constant.ERROR_UPSTREAM_TIMEOUT:
		statusCode = http.StatusGatewayTimeout
	case constant.ERROR_SERVICE_UNAVAILABLE:
		statusCode = http.StatusServiceUnavailable
	case constant.ERROR_BAD_REQUEST:
		statusCode = http.StatusBadRequest
	case constant.ERROR_DATA_CONFLICT:
		statusCode = http.StatusConflict
	case constant.ERROR_UNAUTHORIZED:
		statusCode = http.StatusUnauthorized
	case constant.ERROR_INACTIVE_CREDENTIAL:
		statusCode = http.StatusUnauthorized
	default:
		statusCode = http.StatusServiceUnavailable
		errors = constant.ERROR_SERVICE_UNAVAILABLE
	}

	apiResponse = models.ApiResponse{
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

	return ctx.JSON(statusCode, apiResponse), apiResponse
}
