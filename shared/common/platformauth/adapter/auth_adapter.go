package authadapter

import (
	"los-kmb-api/shared/common/platformlog"

	"github.com/KB-FMF/platform-library"
	"github.com/KB-FMF/platform-library/auth"
)

type AuthAdapter struct {
	RealAuth *auth.Auth
}

func NewPlatformAuth() *AuthAdapter {
	return &AuthAdapter{
		RealAuth: auth.New(platformlog.GetPlatformEnv()),
	}
}

func (a *AuthAdapter) Validation(token string, appName string) (resp *platform.Response, err *platform.Error) {
	resp, err = a.RealAuth.Validation(token, appName)
	return resp, err
}

func (a *AuthAdapter) Login(data map[string]interface{}) (resp *platform.Response, err *platform.Error) {
	resp, err = a.RealAuth.Login(data)
	return resp, err
}
