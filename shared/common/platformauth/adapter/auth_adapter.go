package authadapter

import (
	"los-kmb-api/shared/common/platformlog"

	"github.com/KB-FMF/platform-library"
	"github.com/KB-FMF/platform-library/auth"
)

type AuthAdapter struct {
	RealAuth *auth.Auth
}

var _ PlatformAuthInterface = (*AuthAdapter)(nil) // Compile-time check

func NewAuthAdapter(authInstance *auth.Auth) *AuthAdapter {
	return &AuthAdapter{
		RealAuth: authInstance,
	}
}

func NewPlatformAuth() PlatformAuthInterface {
	authInstance := auth.New(platformlog.GetPlatformEnv())
	return NewAuthAdapter(authInstance)
}

func (a *AuthAdapter) Validation(token string, appName string) (*platform.Response, *platform.Error) {
	return a.RealAuth.Validation(token, appName)
}

func (a *AuthAdapter) Login(data map[string]interface{}) (*platform.Response, *platform.Error) {
	return a.RealAuth.Login(data)
}
