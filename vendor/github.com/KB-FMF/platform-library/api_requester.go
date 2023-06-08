package platform

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
)

var apiRequesterWrapper = APIRequesterWrapper{
	mu:           new(sync.RWMutex),
	apiRequester: &APIRequesterImplementation{HTTPClient: &http.Client{}},
}

// APIRequesterWrapper is the APIRequester with locker for setting the APIRequester.
type APIRequesterWrapper struct {
	apiRequester APIRequester
	mu           *sync.RWMutex
}

// APIRequester abstraction of HTTP Client that will make API calls to ocr backend.
// `body` is POST-requests' bodies if applicable.
// `target` pointer to value which response string will be unmarshalled to.
type APIRequester interface {
	Call(ctx context.Context, method string, url string, header http.Header, body interface{}, target interface{}) *Error
	CallMultipart(request *http.Request, target interface{}) *Error
}

// APIRequesterImplementation is the default implementation of APIRequester.
type APIRequesterImplementation struct {
	HTTPClient *http.Client
}

// Call makes HTTP requests with JSON-format body.
// `body` is POST-requests' bodies if applicable.
// `result` pointer to value which response string will be unmarshalled to.
func (a *APIRequesterImplementation) Call(ctx context.Context, method string, url string, header http.Header, body interface{}, target interface{}) *Error {
	var reqBody io.Reader
	switch method {
	case http.MethodGet:
		reqBody = nil
	case http.MethodPost:
		jsonMarshal, err := json.Marshal(body)
		if err != nil {
			return FromGoErr(err)
		}
		reqBody = bytes.NewBuffer(jsonMarshal)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return FromGoErr(err)
	}

	if header != nil {
		req.Header = header
	}

	return a.doRequest(req, target)
}

// doRequest get JSON from API and decodes it.
func (a *APIRequesterImplementation) doRequest(req *http.Request, target interface{}) *Error {
	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return FromGoErr(err)
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return FromGoErr(err)
	}

	// if unavailable service return forbidden, but also data
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return FromHTTPErr(resp.StatusCode, respBody)
	}

	if err := json.Unmarshal(respBody, &target); err != nil {
		return FromGoErr(err)
	}

	return nil
}

// SetAPIRequester sets the APIRequester for API call.
func SetAPIRequester(apiRequester APIRequester) {
	apiRequesterWrapper.mu.Lock()
	defer apiRequesterWrapper.mu.Unlock()

	apiRequesterWrapper.apiRequester = apiRequester
}

func (a *APIRequesterImplementation) CallMultipart(request *http.Request, target interface{}) *Error {
	return a.doRequest(request, target)
}
