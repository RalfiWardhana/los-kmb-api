package httpclient

import (
	"errors"
	"os"
	"time"

	"los-kmb-api/shared/constant"

	"gopkg.in/resty.v1"
)

type httpClientHandler struct{}

func NewHttpClient() HttpClient {
	return &httpClientHandler{}
}

type HttpClient interface {
	CallWebSocket(url string, param interface{}, header map[string]string, timeOut int) (resp *resty.Response, err error)
	CustomerDomain(endpoint string, param interface{}, header map[string]string, method string, timeOut int, accessToken string) (resp *resty.Response, err error)
}

func (h httpClientHandler) CallWebSocket(url string, param interface{}, header map[string]string, timeOut int) (resp *resty.Response, err error) {

	header["Content-Type"] = "application/json"

	client := resty.New()

	client.SetTimeout(time.Second * time.Duration(timeOut))
	resp, err = client.R().SetHeaders(header).SetBody(param).Post(url)

	if err != nil {
		err = errors.New(constant.CONNECTION_ERROR)
		return
	}

	return

}

func (h httpClientHandler) CustomerDomain(endpoint string, param interface{}, header map[string]string, method string, timeOut int, accessToken string) (resp *resty.Response, err error) {

	header["Content-Type"] = "application/json"
	header["Authorization"] = accessToken

	client := resty.New()

	if os.Getenv("APP_ENV") != "production" {
		client.SetDebug(true)
	}

	url := os.Getenv("CUSTOMER_V3_BASE_URL") + endpoint

	switch method {
	case constant.METHOD_POST:
		resp, err = client.R().SetHeaders(header).SetBody(param).Post(url)
	case constant.METHOD_GET:
		resp, err = client.R().SetHeaders(header).SetBody(param).Get(url)
	case constant.METHOD_PUT:
		resp, err = client.R().SetHeaders(header).SetBody(param).Put(url)
	case constant.METHOD_DELETE:
		resp, err = client.R().SetHeaders(header).SetBody(param).Delete(url)
	}

	if err != nil {
		err = errors.New(constant.CONNECTION_ERROR)
		return
	}
	return

}
