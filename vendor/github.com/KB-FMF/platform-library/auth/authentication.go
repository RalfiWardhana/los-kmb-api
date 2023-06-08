package auth

import (
	"context"
	"github.com/KB-FMF/platform-library"
	"github.com/KB-FMF/platform-library/utils"
	"net/http"
)

func (a *Auth) Login(data map[string]interface{}) (*platform.Response, *platform.Error) {
	urlScheme, err := utils.UrlParse(a.config.authServiceUrl, TagVersion, ServiceName, "login")
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	response := &platform.Response{}
	if err := a.apiRequester.Call(context.Background(), http.MethodPost, urlScheme, headers, data, &response); err != nil {
		return nil, err
	}

	return response, nil
}

func (a *Auth) Validation(token string, applicationName string) (*platform.Response, *platform.Error) {
	urlScheme, err := utils.UrlParse(a.config.authServiceUrl, TagVersion, ServiceName, "validation")
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", token)
	if applicationName == "" {
		applicationName = appName
	}
	headers.Set("application_name", applicationName)

	response := &platform.Response{}
	if err := a.apiRequester.Call(context.Background(), http.MethodGet, urlScheme, headers, []byte(""), &response); err != nil {
		return nil, err
	}

	return response, nil
}
