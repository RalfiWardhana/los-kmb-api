package authadapter

import "github.com/KB-FMF/platform-library"

type PlatformAuthInterface interface {
	Validation(token, appName string) (*platform.Response, *platform.Error)
	Login(data map[string]interface{}) (*platform.Response, *platform.Error)
}
