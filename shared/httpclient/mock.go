package httpclient

import (
	"github.com/stretchr/testify/mock"
	"gopkg.in/resty.v1"
)

type MockHttpClient struct {
	mock.Mock
}

func (m MockHttpClient) CallWebSocket(url string, param interface{}, header map[string]string, timeOut int) (resp *resty.Response, err error) {

	args := m.Called(url, param, header, timeOut)
	return args.Get(0).(*resty.Response), args.Error(1)

}
