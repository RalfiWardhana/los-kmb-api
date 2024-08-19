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
		err = errors.Wrap(splitErrMessage[0], errWord)
	case errors.ErrGatewayTimeout:
		code = "996"
		err = errors.Wrap(splitErrMessage[0], errWord)
	case errors.ErrServiceUnavailable:
		code = "998"
		err = errors.Wrap(splitErrMessage[0], errWord)
	default:
		code = "999"
		err = errors.Wrap(errors.ErrInternalServerError, errWord)
	}

	return
}
