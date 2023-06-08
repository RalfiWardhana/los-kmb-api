package platform

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// GoErrCode error code for errors happen inside Go code.
var GoErrCode = "GO_ERROR"

// Error is the conventional OCR error.
type Error struct {
	Messages string      `json:"messages"`
	Data     interface{} `json:"data,omitempty"`
	Errors   interface{} `json:"errors"`
	Code     string      `json:"code"`
}

// Error returns error message.
// This enables ocr.Error to comply with Go error interface.
func (e *Error) Error() interface{} {
	return e.Errors
}

// ErrorMessage returns error message.
// This enables ocr.Error to comply with Go error interface.
func (e *Error) ErrorMessage() string {
	return e.Messages
}

// GetErrorCode returns error code coming from ocr backend.
func (e *Error) GetErrorCode() string {
	return e.Code
}

// FromGoErr generates ocr.Error from generic go errors.
func FromGoErr(err error) *Error {
	return &Error{
		Code:     strconv.Itoa(http.StatusTeapot),
		Messages: GoErrCode,
		Errors:   err,
	}
}

// FromHTTPErr generates ocr.Error from http errors with non 2xx status.
func FromHTTPErr(status int, respBody []byte) *Error {
	var httpError *Error
	if err := json.Unmarshal(respBody, &httpError); err != nil {
		return FromGoErr(err)
	}
	httpError.Code = strconv.Itoa(status)

	return httpError
}
