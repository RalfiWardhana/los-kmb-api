package middlewares

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"los-kmb-api/shared/common"
	"los-kmb-api/shared/utils"

	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"gopkg.in/resty.v1"
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

func (m *AccessMiddleware) AccessMiddleware() echo.MiddlewareFunc {
	return func(handlerFunc echo.HandlerFunc) echo.HandlerFunc {
		return func(context echo.Context) error {

			var err error
			_, err = GetPlatformAuth()
			if err != nil {
				return m.BadGateway(context, err.Error())
			}
			return handlerFunc(context)

		}
	}

}
