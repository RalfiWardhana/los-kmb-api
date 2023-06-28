package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/go-resty/resty/v2"
)

type httpClientHandler struct{}

func NewHttpClient() HttpClient {
	return &httpClientHandler{}
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . HttpClient
type HttpClient interface {
	EngineAPI(ctx context.Context, logFile, link string, param []byte, header map[string]string, method string, retry bool, retryNumber int, timeOut int, prospectID string, accessToken string) (resp *resty.Response, err error)
	CustomerAPI(ctx context.Context, logFile, endpoint string, param []byte, method string, accessToken string, prospectID string, keyTimeout string) (resp *resty.Response, err error)
	MediaClient(ctx context.Context, logFile, link, method string, param interface{}, header map[string]string, timeOut int, customerID int, accessToken string) (resp *resty.Response, err error)
}

func (h httpClientHandler) EngineAPI(ctx context.Context, logFile, link string, param []byte, header map[string]string, method string, retry bool, retryNumber int, timeOut int, prospectID string, accessToken string) (resp *resty.Response, err error) { // nambahin accessToken buat logging
	var levelLog string
	var currentRetryAttempt int
	mapRequest := map[string]interface{}{}

	if param != nil {
		err = json.Unmarshal(param, &mapRequest)
		if err != nil {
			err = fmt.Errorf(constant.ERROR_UNMARSHAL)
			return
		}
	}

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if ok {
		requestID = ""
	}

	header["Content-Type"] = "application/json"
	header[echo.HeaderXRequestID] = requestID

	client := resty.New()
	if os.Getenv("APP_ENV") != "production" {
		client.SetDebug(true)
	}
	client.SetTimeout(time.Second * time.Duration(timeOut))
	if retry {
		client.SetRetryCount(retryNumber)
		client.AddRetryCondition(
			func(r *resty.Response, err error) bool {
				retry := r.StatusCode() >= 500 || r.StatusCode() == 0 || r.StatusCode() == 404
				if retry {
					var respBody interface{}
					_ = json.Unmarshal(r.Body(), &respBody)
					currentRetryAttempt = currentRetryAttempt + 1
					common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
						Link:       link,
						Method:     method,
						LogFile:    logFile,
						MsgLogFile: constant.MSG_INT_API,
						LevelLog:   levelLog,
						Request:    mapRequest,
						Response:   respBody,
					})
				}
				return retry
			})

	}

	switch method {
	case constant.METHOD_POST:
		resp, err = client.R().SetHeaders(header).SetBody(param).Post(link)
	case constant.METHOD_GET:
		resp, err = client.R().SetHeaders(header).SetBody(param).Get(link)
	case constant.METHOD_PUT:
		resp, err = client.R().SetHeaders(header).SetBody(param).Put(link)
	case constant.METHOD_DELETE:
		resp, err = client.R().SetHeaders(header).SetBody(param).Delete(link)
	}

	if err != nil {
		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       link,
			Method:     method,
			LogFile:    logFile,
			MsgLogFile: constant.MSG_INT_API,
			LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
			Request:    mapRequest,
			Response:   map[string]interface{}{"errors": err.Error()},
		})
		err = fmt.Errorf(constant.ERROR_CONNECTION)
		if retryNumber > 0 && currentRetryAttempt >= retryNumber {
			err = fmt.Errorf(constant.RESTY_MAX_RETRY_ERROR)
		}
		return
	}

	var respBody interface{}
	err = json.Unmarshal(resp.Body(), &respBody)
	if err != nil {
		return
	}

	if resp.StatusCode() == 200 || resp.StatusCode() == 201 {
		levelLog = constant.PLATFORM_LOG_LEVEL_INFO
	} else {
		levelLog = constant.PLATFORM_LOG_LEVEL_ERROR
		if retryNumber > 0 && currentRetryAttempt >= retryNumber {
			common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
				Link:       link,
				Method:     method,
				LogFile:    logFile,
				MsgLogFile: constant.MSG_INT_API,
				LevelLog:   levelLog,
				Request:    mapRequest,
				Response:   respBody,
			})
			err = errors.New(constant.RESTY_MAX_RETRY_ERROR)
			return
		}
	}
	common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       link,
		Method:     method,
		LogFile:    logFile,
		MsgLogFile: constant.MSG_INT_API,
		LevelLog:   levelLog,
		Request:    mapRequest,
		Response:   respBody,
	})

	return

}

func (h httpClientHandler) CustomerAPI(ctx context.Context, logFile, endpoint string, param []byte, method string, accessToken string, prospectID string, keyTimeout string) (resp *resty.Response, err error) {
	var levelLog string
	mapRequest := map[string]interface{}{}

	if param != nil {
		err = json.Unmarshal(param, &mapRequest)
		if err != nil {
			err = fmt.Errorf(constant.ERROR_UNMARSHAL)
			return
		}
	}

	header := map[string]string{
		"Authorization": accessToken,
	}
	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if ok {
		requestID = ""
	}

	header[echo.HeaderXRequestID] = requestID

	client := resty.New()
	if os.Getenv("APP_ENV") != "production" {
		client.SetDebug(true)
	}
	timeout, _ := strconv.Atoi(os.Getenv(keyTimeout))
	client.SetTimeout(time.Second * time.Duration(timeout))

	url := os.Getenv("CUSTOMER_V3_BASE_URL") + endpoint

	switch method {
	case constant.METHOD_POST:
		resp, err = client.R().SetHeaders(header).SetBody(param).Post(url)
	case constant.METHOD_GET:
		resp, err = client.R().SetHeaders(header).SetBody(param).Get(url)
	case constant.METHOD_PUT:
		resp, err = client.R().SetHeaders(header).SetBody(param).Put(url)
	case constant.METHOD_DELETE:
		resp, err = client.R().SetHeaders(header).SetBody(param).Delete(url)
	}

	if err != nil {
		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       url,
			Method:     method,
			LogFile:    logFile,
			MsgLogFile: constant.MSG_CORE_API,
			LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
			Request:    mapRequest,
			Response:   map[string]string{"errors": err.Error()},
		})
		err = fmt.Errorf(constant.ERROR_CONNECTION)
		return
	}

	var kreditmuResponse response.CustomerDomain

	json.Unmarshal(resp.Body(), &kreditmuResponse)

	if resp.StatusCode() == 400 {
		if kreditmuResponse.Code == constant.CORE_TOKEN_EXPIRED || kreditmuResponse.Message == constant.TOKEN_INVALID {
			middlewares.UserInfoData = middlewares.UserInfo{}
		}
	}

	if resp.StatusCode() >= 200 && resp.StatusCode() < 300 {
		levelLog = constant.PLATFORM_LOG_LEVEL_INFO
	} else {
		levelLog = constant.PLATFORM_LOG_LEVEL_ERROR
	}
	common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       url,
		Method:     method,
		LogFile:    logFile,
		MsgLogFile: constant.MSG_CORE_API,
		LevelLog:   levelLog,
		Request:    mapRequest,
		Response:   kreditmuResponse,
	})

	return

}

func (h httpClientHandler) MediaClient(ctx context.Context, logFile, link, method string, param interface{}, header map[string]string, timeOut int, customerID int, accessToken string) (resp *resty.Response, err error) {
	var levelLog string
	mapRequest := map[string]interface{}{}

	client := resty.New()

	client.SetTimeout(time.Second * time.Duration(timeOut))

	switch method {
	case constant.METHOD_GET:
		resp, err = client.R().SetHeaders(header).Get(link)
	}

	if err != nil {
		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       link,
			Method:     method,
			LogFile:    logFile,
			MsgLogFile: constant.MSG_PLATFORM_API,
			LevelLog:   levelLog,
			Request:    mapRequest,
			Response:   map[string]interface{}{"errors": err.Error()},
		})
		err = errors.New(constant.CONNECTION_ERROR)
		return
	}

	var mediaResponse interface{}
	json.Unmarshal(resp.Body(), &mediaResponse)

	levelLog = constant.PLATFORM_LOG_LEVEL_INFO

	if resp.StatusCode() != 200 {
		levelLog = constant.PLATFORM_LOG_LEVEL_ERROR
	}

	common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       link,
		Method:     method,
		LogFile:    logFile,
		MsgLogFile: constant.MSG_PLATFORM_API,
		LevelLog:   levelLog,
		Request:    mapRequest,
		Response:   mediaResponse,
	})

	return

}
