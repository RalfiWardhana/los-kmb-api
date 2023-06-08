package auth

import (
	"net"
	"net/http"
	"time"

	"github.com/KB-FMF/platform-library"
)

const (
	ENV_DEVELOPMENT = "DEVELOPMENT"
	ENV_STAGING     = "STAGING"
	ENV_PRODUCTION  = "PRODUCTION"
	TagVersion      = "v1"
	ServiceName     = "auth"
	appName         = "platform-auth-api"
)

type Auth struct {
	config       Config
	apiRequester platform.APIRequester
}

func New(environment string) *Auth {
	var config Config
	if environment == ENV_DEVELOPMENT {
		config = newDevelopmentConfig()
	} else if environment == ENV_STAGING {
		config = newStagingConfig()
	} else if environment == ENV_PRODUCTION {
		config = newProductionConfig()
	} else {
		panic("no selected environment (DEVELOPMENT | STAGING | PRODUCTION)")
	}

	var apiRequester *platform.APIRequesterImplementation

	if config.simpleHttp {
		apiRequester = &platform.APIRequesterImplementation{HTTPClient: &http.Client{
			Timeout: time.Duration(config.httpTimeout) * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        config.maxIdleConns,
				MaxIdleConnsPerHost: config.maxIdleConnsPerHost,
				IdleConnTimeout:     time.Duration(config.idleConnTimeout) * time.Second,
				DisableKeepAlives:   config.disableKeepAlive,
			},
		}}
	} else {
		apiRequester = &platform.APIRequesterImplementation{HTTPClient: &http.Client{
			Timeout: time.Duration(config.httpTimeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   time.Duration(config.dialContextTimeout) * time.Second,
					KeepAlive: time.Duration(config.dialContextKeepAlive) * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     config.forceAttemptHttp2,
				MaxIdleConns:          config.maxIdleConns,
				MaxIdleConnsPerHost:   config.maxIdleConnsPerHost,
				IdleConnTimeout:       time.Duration(config.idleConnTimeout) * time.Second,
				TLSHandshakeTimeout:   time.Duration(config.tlsHandshakeTimeout) * time.Second,
				ExpectContinueTimeout: time.Duration(config.expectContinueTimeout) * time.Second,
			},
		}}
	}

	return &Auth{
		config:       config,
		apiRequester: apiRequester,
	}
}
