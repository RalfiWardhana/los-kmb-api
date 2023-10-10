package httpclient

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/go-resty/resty/v2"
)

type MockHttpClient struct {
	mock.Mock
}

func (m MockHttpClient) EngineAPI(ctx context.Context, logFile, link string, param []byte, header map[string]string, method string, retry bool, retryNumber int, timeOut int, prospectID string, accessToken string) (resp *resty.Response, err error) {

	args := m.Called(ctx, logFile, link, param, header, method, retry, retryNumber, timeOut, prospectID, accessToken)
	return args.Get(0).(*resty.Response), args.Error(1)
}

func (m MockHttpClient) CustomerAPI(ctx context.Context, logFile, endpoint string, param []byte, method string, accessToken string, prospectID string, keyTimeout string) (resp *resty.Response, err error) {

	args := m.Called(ctx, logFile, endpoint, param, method, accessToken, prospectID, keyTimeout)
	return args.Get(0).(*resty.Response), args.Error(1)
}

// MediaClient implements HttpClient
func (m MockHttpClient) MediaClient(ctx context.Context, logFile, url string, method string, param interface{}, header map[string]string, timeOut int, customerID int, accessToken string) (resp *resty.Response, err error) {
	args := m.Called(ctx, logFile, url, method, param, header, timeOut, customerID)
	return args.Get(0).(*resty.Response), args.Error(1)
}
