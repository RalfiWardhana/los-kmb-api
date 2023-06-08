package platformlog

import (
	"errors"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/url"
	"os"
	"strings"

	platformLog "github.com/KB-FMF/platform-library/log"
)

type PlatformLog struct{}

type PlatformLogInterface interface {
	WriteLog(accessToken, level, link, method string, duration float64, header map[string]string, request, response map[string]interface{}) error
}

func NewPlatformLog() *PlatformLog {
	return &PlatformLog{}
}

func (pl PlatformLog) WriteLog(accessToken, level, link, method string, duration float64, header map[string]string, request, response map[string]interface{}) error {
	var logEnv string

	env := os.Getenv("APP_ENV")

	if strings.Contains(strings.ToLower(env), "production") {
		logEnv = platformLog.ENV_PRODUCTION
	} else if strings.Contains(strings.ToLower(env), "staging") {
		logEnv = platformLog.ENV_STAGING
	} else {
		logEnv = platformLog.ENV_DEVELOPMENT
	}

	if !strings.Contains(link, "http") {
		link = "http://" + link
	}

	parsedURL, errParseUrl := url.Parse(link)
	if errParseUrl != nil {
		return errParseUrl
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

	logger := platformLog.New(logEnv)
	errLogger := logger.Log(accessToken, body)
	if errLogger != nil {
		// log.Info("Payload body platform:", body) // gommon log
		return errors.New(errLogger.Messages)
	}
	logger.Flush(10)

	return nil
}
