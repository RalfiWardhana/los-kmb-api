package common

import (
	"context"
	"fmt"
	"io"
	"los-kmb-api/shared/common/platformlog"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/utils"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"los-kmb-api/shared/constant"

	"github.com/labstack/echo/v4"
	gommonLog "github.com/labstack/gommon/log"
	"github.com/sirupsen/logrus"
)

type CentralizeLogParameter struct {
	Link       string
	Method     string
	Action     string
	Type       string
	LogFile    string
	MsgLogFile string
	LevelLog   string
	Request    interface{}
	Response   interface{}
}

func CentralizeLog(ctx context.Context, accessToken string, logParam CentralizeLogParameter) error {
	var err error
	var isOkGetLink, isOkGetMethod bool
	var duration float64
	var link, method string

	useLogPlatform, _ := strconv.ParseBool(os.Getenv("USE_LOG_PLATFORM"))
	useLogFile, _ := strconv.ParseBool(os.Getenv("USE_LOG_FILE"))
	isLogFileEnabled, _ := strconv.ParseBool(os.Getenv(logParam.LogFile))

	if (!useLogPlatform && !useLogFile) || !isLogFileEnabled {
		return nil
	}

	// get values
	requestID, _ := ctx.Value(echo.HeaderXRequestID).(string)
	if logParam.Link != "" {
		link = logParam.Link
	} else {
		link, isOkGetLink = ctx.Value(constant.CTX_KEY_INCOMING_REQUEST_URL).(string)
		if !isOkGetLink {
			link = os.Getenv("DUMMY_URL_LOGS")
		}
	}
	if logParam.Method != "" {
		method = logParam.Method
	} else {
		method, isOkGetMethod = ctx.Value(constant.CTX_KEY_INCOMING_REQUEST_METHOD).(string)
		if !isOkGetMethod {
			method = constant.METHOD_GET
		}
	}

	// create headers
	header := map[string]string{
		echo.HeaderXRequestID: requestID,
	}
	if logParam.Action != "" {
		header["action"] = logParam.Action
	}
	if logParam.Type != "" {
		header["type"] = logParam.Type
	}
	if logParam.MsgLogFile == "" {
		logParam.MsgLogFile = logParam.Type
	}

	// calculate request time (duration) in ms
	requestTime, ok := ctx.Value(constant.CTX_KEY_REQUEST_TIME).(int64)
	if ok {
		duration = float64(utils.GenerateTimeInMilisecond() - requestTime)
	}

	// format request and response
	mapRequest := map[string]interface{}{}
	if logParam.Request != nil {
		mapRequest = utils.StructToMap(logParam.Request)
	}
	mapResponse := map[string]interface{}{}
	if logParam.Response != nil {
		mapResponse = utils.StructToMap(logParam.Response)
	}

	if useLogPlatform {
		payloadBase64, err := platformlog.Log.WriteLog(accessToken, logParam.LevelLog, link, method, duration, header, mapRequest, mapResponse)
		if err != nil {
			gommonLog.Error("[Error] Write Platform LOG ", err)
			errWriteLog := WriteFileLog(logParam.LogFile, logParam.MsgLogFile, constant.PLATFORM_LOG_LEVEL_ERROR, link, method, duration, header, mapRequest, map[string]interface{}{
				"payload_log": payloadBase64,
				"errors":      err.Error(),
			})
			if errWriteLog != nil {
				gommonLog.Error("[Error] Write File LOG ", errWriteLog)
			}
		}
	}

	if useLogFile {
		err = WriteFileLog(logParam.LogFile, logParam.MsgLogFile, logParam.LevelLog, link, method, duration, header, mapRequest, mapResponse)
		if err != nil {
			gommonLog.Error("[Error] Write File LOG ", err)
		}
	}

	return err

}

func WriteFileLog(keyConfig, msg, levelLog, link, method string, duration float64, header map[string]string, request map[string]interface{}, response map[string]interface{}) error {

	dateNow := time.Now().Local().Format(constant.FORMAT_DATE)

	if config.DateLogFile[keyConfig] != dateNow {
		config.CreateCustomLogFile(keyConfig)
	}

	logPath := os.Getenv("LOG_FILE")
	active, _ := strconv.ParseBool(os.Getenv(keyConfig))
	if logPath != "" && active {

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

		var isError bool
		switch levelLog {
		case constant.PLATFORM_LOG_LEVEL_INFO:
			isError = false
		default:
			isError = true
		}

		body := map[string]interface{}{
			"environment": os.Getenv("APP_ENV"),
			"duration":    fmt.Sprintf("%fms", duration),
			"header":      header,
			"host":        parsedURL.Scheme + "://" + parsedURL.Host,
			"ip_address":  "127.0.0.1",
			"level":       levelLog,
			"method":      method,
			"path":        parsedURL.Path,
			"request":     request,
			"response":    response,
			"timestamp":   timestamp,
		}

		logger := logrus.New()
		logFile := config.GetLogFile[keyConfig]

		logger.SetOutput(io.MultiWriter(logFile, os.Stdout))
		logger.SetFormatter(&logrus.JSONFormatter{})
		if isError {
			logger.WithFields(body).Error(msg)
			return nil
		}
		logger.WithFields(body).Info(msg)
		return nil
	}
	return nil
}
