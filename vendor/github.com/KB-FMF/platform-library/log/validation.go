package log

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/url"
	"regexp"
	"time"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	_ = validate.RegisterValidation("Method", ValidationMethod)
	_ = validate.RegisterValidation("UrlPath", ValidationPath)
	_ = validate.RegisterValidation("UrlHost", ValidationHost)
	_ = validate.RegisterValidation("IPAddress", ValidationIPAddress)
	_ = validate.RegisterValidation("Level", ValidationLogLevel)
	_ = validate.RegisterValidation("DateLog", ValidationDateLog)
	_ = validate.RegisterValidation("FloatNumber", ValidationFloatNumber)
	_ = validate.RegisterValidation("Map", ValidationMap)
}

const (
	dateTimeLayout = "2006/01/02 15:04:05.000"
	regIPv4        = `^(((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4})`
	regIPv6        = `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`
)

var (
	regexUrlPath  = regexp.MustCompile(`^(?:[^/.\s]+\.)*(?:/[^/\s]+)*/?$`)
	regexLogLevel = regexp.MustCompile(`(^INFO$|^ERROR$|^CRITICAL$|^WARNING$)`)
	regexMethod   = regexp.MustCompile(`(^GET$|^PUT$|^POST$|^DELETE$|^OPTION$|^HEAD$|^PATCH$|^TRACE$)`)
	regexIP       = regexp.MustCompile(fmt.Sprintf("%s|%s", regIPv4, regIPv6))
	regexUrlHost  = regexp.MustCompile(`^(https?:\/\/)([A-Za-z0-9]{1})([A-Za-z0-9.\-_:#])+([A-Za-z0-9.\-\/_#:]+)[A-Za-z0-9]{2,5}(:[0-9]{1,5})?(.)?$`)
)

type Request struct {
	Timestamp string                 `json:"timestamp" validate:"DateLog"`
	Level     string                 `json:"level" validate:"Level"`
	IpAddress string                 `json:"ip_address" validate:"IPAddress"`
	Method    string                 `json:"method" validate:"Method"`
	Path      string                 `json:"path" validate:"UrlPath"`
	Host      string                 `json:"host" validate:"UrlHost"`
	Duration  float64                `json:"duration" validate:"FloatNumber"`
	Header    map[string]interface{} `json:"header" validate:"Map"`
	Request   map[string]interface{} `json:"request" validate:"Map"`
	Response  map[string]interface{} `json:"response" validate:"Map"`
}

type PayloadLogGeneral struct {
	Timestamp string                 `json:"timestamp" validate:"DateLog"`
	Level     string                 `json:"level" validate:"Level"`
	IpAddress string                 `json:"ip_address" validate:"IPAddress"`
	Method    string                 `json:"method"`
	Path      string                 `json:"path"`
	Host      string                 `json:"host"`
	Duration  float64                `json:"duration" validate:"FloatNumber"`
	Header    map[string]interface{} `json:"header" validate:"Map"`
	Request   map[string]interface{} `json:"request" validate:"Map"`
	Response  map[string]interface{} `json:"response" validate:"Map"`
}

func getTag(fe validator.FieldError) (bool, string) {
	switch fe.Tag() {
	case "DateLog":
		return true, fe.Field() + " Must be valid timestamp " + dateTimeLayout
	case "Level":
		return true, fe.Field() + " Must be valid log level " + "INFO, ERROR, CRITICAL, WARNING"
	case "IPAddress":
		return true, fe.Field() + " Must be valid IP address"
	case "Method":
		return true, fe.Field() + " Only accept one of followings: GET, PUT, POST, DELETE, OPTION, HEAD, PATCH, TRACE"
	case "UrlPath":
		return true, fe.Field() + " Invalid value"
	case "UrlHost":
		return true, fe.Field() + " Invalid value"
	case "FloatNumber":
		return true, fe.Field() + " Must be valid float number"
	case "Map":
		return true, fe.Field() + " Cannot be null"
	default:
		return false, fe.Error() // default error
	}
}

func ValidationMap(fl validator.FieldLevel) bool {
	val := fl.Field().Interface()
	if v, ok := val.(map[string]interface{}); ok {
		if v == nil {
			return false
		}
		return true
	}

	return false
}

// ValidationMethod request
func ValidationMethod(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return false
	} else {
		if !regexMethod.MatchString(val) {
			return false
		}
	}
	return true
}

// ValidationPath accept url
func ValidationPath(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return false
	} else {
		if !regexUrlPath.MatchString(val) {
			return false
		}
	}
	return true
}

// ValidationHost accept url
func ValidationHost(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return false
	} else {
		if !regexUrlHost.MatchString(val) {
			return false
		}

		raw, err := url.Parse(val)
		if err != nil {
			return false
		}

		host := raw.Hostname()
		if host == "" {
			return false
		}
	}
	return true
}

// ValidationIPAddress request
func ValidationIPAddress(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return false
	} else {
		if !regexIP.MatchString(val) {
			return false
		}
	}

	return true
}

// ValidationLogLevel log flag request
func ValidationLogLevel(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return false
	} else {
		if !regexLogLevel.MatchString(val) {
			return false
		}
	}

	return true
}

func ValidationDateLog(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	if len(val) == 0 {
		return false
	}
	_, err := time.Parse(dateTimeLayout, val)
	if err != nil {
		return false
	}

	return true
}

func ValidationFloatNumber(fl validator.FieldLevel) bool {
	obj := fl.Field().Interface()
	if _, ok := obj.(float64); !ok {
		return false
	}
	return true
}
