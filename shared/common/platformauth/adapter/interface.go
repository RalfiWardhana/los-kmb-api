package authadapter

import "github.com/KB-FMF/platform-library"

type PlatformAuthInterface interface {
	Validation(token string, appName string) (resp *platform.Response, err *platform.Error)
	Login(data map[string]interface{}) (resp *platform.Response, err *platform.Error)
}
