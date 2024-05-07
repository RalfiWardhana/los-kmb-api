package middlewares

import (
	"context"
	"fmt"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"reflect"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

type AccessMiddleware struct {
	common.JSON
}

var UserInfoData UserInfo

type UserInfo struct {
	AccessToken string `json:"access_token"`
	ExpiredAt   string `json:"expired_at"`
}

func (m *UserInfo) IsStructureEmpty() bool {
	return reflect.DeepEqual(m, UserInfo{})
}

var HrisApiData HrisApiInfo

type HrisApiInfo struct {
	Token       string `json:"token"`
	ExpiredTime int    `json:"expired_time"`
}

func (m *HrisApiInfo) IsStructureEmpty() bool {
	return reflect.DeepEqual(m, HrisApiInfo{})
}

func NewAccessMiddleware() *AccessMiddleware {
	return &AccessMiddleware{}
}

func GetPlatformAuth() (userInfo UserInfo, err error) {

	client := resty.New()
	if os.Getenv("APP_ENV") != "production" {
		client.SetDebug(true)
	}
	if UserInfoData.AccessToken == "" {
		body := map[string]interface{}{
			"secret_key":         os.Getenv("PLATFORM_SECRET_KEY"),
			"source_application": "LOS",
		}

		resp, err := client.R().SetBody(body).Post(os.Getenv("PLATFORM_AUTH_BASE_URL") + "/v1/auth/login")
		if err != nil || resp.StatusCode() != 200 {
			err = fmt.Errorf("error get access token")
			return userInfo, err
		}

		userInfo := new(UserInfo)
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), userInfo)

		UserInfoData = *userInfo
		return UserInfoData, err

	} else {

		expired, _ := time.Parse("2006-01-02T15:04:05Z07:00", UserInfoData.ExpiredAt)
		expired5Minute := expired.Add(-5 * time.Minute)

		if utils.DiffTwoDate(expired5Minute).Seconds() > 0 {

			body := map[string]interface{}{
				"secret_key":         os.Getenv("PLATFORM_SECRET_KEY"),
				"source_application": "LOS",
			}

			resp, err := client.R().SetBody(body).Post(os.Getenv("PLATFORM_AUTH_BASE_URL") + "/v1/auth/login")

			if err != nil || resp.StatusCode() != 200 {
				err = fmt.Errorf("error get access token")
				return userInfo, err
			}

			userInfo := new(UserInfo)
			jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), userInfo)

			UserInfoData = *userInfo
			return UserInfoData, err

		}
	}

	return
}

func GetTokenHris() (hrisApiInfo HrisApiInfo, err error) {

	client := resty.New()
	if os.Getenv("APP_ENV") != "production" {
		client.SetDebug(true)
	}
	if HrisApiData.Token == "" {
		body := map[string]interface{}{
			"api_key": os.Getenv("HRIS_API_KEY"),
		}

		resp, err := client.R().SetBody(body).Post(os.Getenv("HRIS_GET_TOKEN_URL"))
		if err != nil || resp.StatusCode() != 200 {
			err = fmt.Errorf("error get access token hris")
			return hrisApiInfo, err
		}

		hrisApiInfo := new(HrisApiInfo)
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), hrisApiInfo)

		HrisApiData = *hrisApiInfo
		return HrisApiData, err

	} else {

		expired := time.Now().Add(time.Second * time.Duration(HrisApiData.ExpiredTime))
		expired5Minute := expired.Add(-5 * time.Minute)

		if utils.DiffTwoDate(expired5Minute).Seconds() > 0 {

			body := map[string]interface{}{
				"api_key": os.Getenv("HRIS_API_KEY"),
			}

			resp, err := client.R().SetBody(body).Post(os.Getenv("HRIS_GET_TOKEN_URL"))

			if err != nil || resp.StatusCode() != 200 {
				err = fmt.Errorf("error get access token hris")
				return HrisApiData, err
			}

			hrisApiInfo := new(HrisApiInfo)
			jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), hrisApiInfo)

			HrisApiData = *hrisApiInfo
			return HrisApiData, err

		}
	}

	return
}

func (m *AccessMiddleware) AccessMiddleware() echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {

			var err error

			// platform token
			_, err = GetPlatformAuth()
			if err != nil {
				return m.BadGateway(context, err.Error())
			}

			// hris token
			_, err = GetTokenHris()
			if err != nil {
				return m.BadGateway(context, err.Error())
			}
			return handlerFunc(context)

		}
	}

}

func (m *AccessMiddleware) SetupHeadersAndContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			// Get Incoming Request Time and Request ID
			startTime := time.Now().Local().UnixNano() / int64(time.Millisecond)
			reqId := ctx.Request().Header.Get(echo.HeaderXRequestID)
			if reqId == "" {
				reqId = utils.GenerateUUID()
			}

			// Set Response Headers
			ctx.Response().Header().Set(echo.HeaderXRequestID, reqId)
			ctx.Response().Header().Set(constant.CTX_KEY_TAG_VERSION, os.Getenv("APP_VERSION"))
			ctx.Response().Header().Set(constant.CTX_KEY_LOS_VERSION, os.Getenv("LOS_VERSION"))

			// Set Echo Request Headers
			incoming_request_url := constant.LOS_KMB_BASE_URL + ctx.Request().URL.Path

			ctx.Set(constant.CTX_KEY_REQUEST_TIME, startTime)
			ctx.Set(echo.HeaderXRequestID, reqId)
			ctx.Set(constant.CTX_KEY_INCOMING_REQUEST_URL, incoming_request_url)
			ctx.Set(constant.CTX_KEY_INCOMING_REQUEST_METHOD, ctx.Request().Method)

			// Set Golang Context
			reqCtx := ctx.Request().Context()
			reqCtx = context.WithValue(reqCtx, constant.CTX_KEY_REQUEST_TIME, startTime)
			reqCtx = context.WithValue(reqCtx, echo.HeaderXRequestID, reqId)
			reqCtx = context.WithValue(reqCtx, constant.CTX_KEY_INCOMING_REQUEST_URL, incoming_request_url)
			reqCtx = context.WithValue(reqCtx, constant.CTX_KEY_INCOMING_REQUEST_METHOD, ctx.Request().Method)
			ctx.SetRequest(ctx.Request().WithContext(reqCtx))

			return next(ctx)
		}
	}
}
