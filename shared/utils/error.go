package utils

import (
	"strings"

	"github.com/KB-FMF/los-common-library/errors"
)

func WrapError(errWord error) (code string, err error) {

	splitErrMessage := strings.Split(errWord.Error(), " - ")
	switch splitErrMessage[0] {
	case errors.ErrBadGateway:
		code = "995"
		err = errWord
	case errors.ErrGatewayTimeout:
		code = "996"
		err = errWord
	case errors.ErrServiceUnavailable:
		code = "998"
		err = errWord
	case errors.ErrTooManyRequests:
		code = "429"
		err = errWord
	default:
		code = "999"
		err = errWord
	}

	return
}
