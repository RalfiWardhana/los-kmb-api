package platformlog

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/url"
	"os"
	"strings"

	platformLog "github.com/KB-FMF/platform-library/log"
)

type PlatformLog struct{}

var Log *PlatformLog
var Logger *platformLog.Logger

type PlatformLogInterface interface {
	CreateLogger()
	WriteLog(accessToken, level, link, method string, duration float64, header map[string]string, request, response map[string]interface{}) error
}

func NewPlatformLog() *PlatformLog {
	return &PlatformLog{}
}

func GetPlatformEnv() string {

	env := os.Getenv("APP_ENV")

	if strings.Contains(strings.ToLower(env), "production") {
		return platformLog.ENV_PRODUCTION
	} else if strings.Contains(strings.ToLower(env), "staging") {
		return platformLog.ENV_STAGING
	} else {
		return platformLog.ENV_DEVELOPMENT
	}
}

func (pl PlatformLog) CreateLogger() {

	platformEnv := GetPlatformEnv()

	Logger = platformLog.New(platformEnv)
}

func (pl PlatformLog) WriteLog(accessToken, level, link, method string, duration float64, header map[string]string, request, response map[string]interface{}) (string, error) {

	if !strings.Contains(link, "http") {
		link = "http://" + link
	}

	parsedURL, errParseUrl := url.Parse(link)
	if errParseUrl != nil {
		return "", errParseUrl
	}

	timestamp := utils.GenerateTimeWithFormat(constant.FORMAT_DATE_TIME_MS)

	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}

	body := map[string]interface{}{
		"duration":   duration,
		"header":     header,
		"host":       parsedURL.Scheme + "://" + parsedURL.Host,
		"ip_address": "127.0.0.1",
		"level":      level,
		"method":     method,
		"path":       parsedURL.Path,
		"request":    request,
		"response":   response,
		"timestamp":  timestamp,
	}

	payloadByte, _ := json.Marshal(body)
	payloadBase64 := base64.RawStdEncoding.EncodeToString(payloadByte)

	errLogger := Logger.Log(accessToken, body)
	if errLogger != nil {
		// log.Info("Payload body platform:", body) // gommon log
		errPlatform, ok := errLogger.Errors.(error)
		if !ok {
			errPlatform = fmt.Errorf("unspecified error platform log: %s", errLogger.Messages)
		}
		return payloadBase64, errPlatform
	}
	Logger.Flush(10)

	return payloadBase64, nil
}
