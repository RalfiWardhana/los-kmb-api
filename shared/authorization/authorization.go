package authorization

import (
	"fmt"
	"los-kmb-api/models/dto"
	"los-kmb-api/shared/authorization/interfaces"
	"los-kmb-api/shared/constant"
	"time"
)

type authorization struct {
	repository interfaces.Repository
}

type Authorization interface {
	Authorization(authRequest dto.AuthModel, now time.Time) (err error)
}

func NewAuth(repository interfaces.Repository) Authorization {
	return &authorization{
		repository: repository,
	}
}

func (u authorization) Authorization(authRequest dto.AuthModel, now time.Time) (err error) {

	auth, err := u.repository.GetAuth(authRequest)

	if err != nil {
		err = fmt.Errorf(constant.ERROR_UNAUTHORIZED)
		return
	}

	if auth.ClientActive != 1 || auth.TokenActive != 1 {
		err = fmt.Errorf(constant.ERROR_INACTIVE_CREDENTIAL)
		return
	}

	timeNow := now.Unix()

	if auth.Expired != nil {

		expired := auth.Expired.(time.Time).Unix()

		if timeNow > expired {
			err = fmt.Errorf(constant.ERROR_INACTIVE_CREDENTIAL)
			return
		}
	}

	if auth.AccessToken != authRequest.Credential {
		err = fmt.Errorf(constant.ERROR_INACTIVE_CREDENTIAL)
		return
	}

	return

}
