package http

import (
	"errors"
	"los-kmb-api/domain/filtering_new/interfaces/mocks"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	mocksJson "los-kmb-api/shared/common/json/mocks"
	mockplatformauth "los-kmb-api/shared/common/platformauth/adapter/mocks"
	mockplatformcache "los-kmb-api/shared/common/platformcache/mocks"
	"los-kmb-api/shared/common/platformevent"
	mockplatformevent "los-kmb-api/shared/common/platformevent/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

func TestProduceFiltering(t *testing.T) {

	testcases := []struct {
		name    string
		reqbody string
	}{
		{
			name: "test err bind",
			reqbody: `{
				"prospect_id": "SAL-TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"cmo_id": "86244INH",
				"chassis_number": "KAN23112",
				"engine_number": "SIN3124",
				"spouse": nil
			}`,
		},
		{
			name: "test customer spouse",
			reqbody: `{
				"prospect_id": "SAL-TST0020230809013",
				"branch_id": "426",
				"id_number": "2+YBWfT+vxaYsm5074O4l+5yxnxd6nq/Nbhv+TMbpA0=",
				"legal_name": "4PpfG0+nSmIZEeJ+a7dTgo7ON6yaH1Cnk32fxnNQbw==",
				"birth_date": "1978-09-03",
				"gender": "F",
				"surgate_mother_name": "OXWbbWkFNYjsj0fT3ITJYi7T0yZ6SZY=",
				"bpkb_name": "K",
				"cmo_id": "86244INH",
				"chassis_number": "KAN23112",
				"engine_number": "SIN3124",
				"spouse": {
					"spouse_id_number": "2+YBWfT+vxaYsm5074O4l+5yxnxd6nq/Nbhv+TMbpA0=",
					"spouse_legal_name": "4PpfG0+nSmIZEeJ+a7dTgo7ON6yaH1Cnk32fxnNQbw==",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "M",
					"spouse_surgate_mother_name": "OXWbbWkFNYjsj0fT3ITJYi7T0yZ6SZY="
				}
			}`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockMultiUsecase := new(mocks.MultiUsecase)
			mockUsecase := new(mocks.Usecase)
			mockRepository := new(mocks.Repository)
			mockJson := new(mocksJson.JSON)
			mockPlatformCache := new(mockplatformcache.PlatformCacheInterface)
			mockAuth := new(mockplatformauth.PlatformAuthInterface)
			e := echo.New()
			e.Validator = common.NewValidator()

			// Create a request and recorder for testing
			req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb//produce/filtering", strings.NewReader(string(tc.reqbody)))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)

			// Create a test HTTP server
			srv := httptest.NewServer(e)

			// Define the mock middleware
			mockMiddleware := middlewares.NewAccessMiddleware()

			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			if tc.name == "test PublishEvent" {
				mockPlatformEvent.On("PublishEvent", ctx.Request().Context(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil)
			}
			mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockJson.On("ServerSideErrorV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{})
			mockAuth.On("Validation", mock.Anything, "").Return(nil, nil)
			mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

			route := e.Group("/kmb")
			FilteringHandler(route, mockMultiUsecase, mockUsecase, mockRepository, mockJson, mockMiddleware, platformEvent, mockPlatformCache, mockAuth)

			handler := &handlerKmbFiltering{
				multiusecase: mockMultiUsecase,
				usecase:      mockUsecase,
				repository:   mockRepository,
				Json:         mockJson,
				producer:     platformEvent,
				authadapter:  mockAuth,
			}
			err := handler.ProduceFiltering(ctx)
			if err != nil {
				t.Errorf("error '%s' was not expected, but got: ", err)
			}
			srv.Close()

		})
	}
}

func TestRemoveCacheFiltering(t *testing.T) {
	testcases := []struct {
		name       string
		prospectID string
		errCache   error
	}{
		{
			name:       "test remove cache success",
			prospectID: "SAL-123456789",
		},
		{
			name:       "test remove cache errppid",
			prospectID: "",
		},
		{
			name:       "test remove cache err",
			prospectID: "SAL-123456789",
			errCache:   errors.New("error"),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockMultiUsecase := new(mocks.MultiUsecase)
			mockUsecase := new(mocks.Usecase)
			mockRepository := new(mocks.Repository)
			mockJson := new(mocksJson.JSON)
			mockPlatformCache := new(mockplatformcache.PlatformCacheInterface)

			handler := &handlerKmbFiltering{
				multiusecase: mockMultiUsecase,
				usecase:      mockUsecase,
				repository:   mockRepository,
				Json:         mockJson,
				cache:        mockPlatformCache,
			}
			e := echo.New()
			e.Validator = common.NewValidator()

			mockPlatformCache.On("SetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, tc.errCache)
			mockJson.On("BadRequestErrorBindV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()
			mockJson.On("ServerSideErrorV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()
			mockJson.On("SuccessV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

			// Create a request and recorder for testing
			req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cache/filtering/:prospect_id", nil)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			c.SetParamNames("prospect_id")
			c.SetParamValues(tc.prospectID)

			// Call the handler
			err := handler.RemoveCacheFiltering(c)
			if err != nil {
				t.Errorf("error '%s' was not expected, but got: ", err)
			}
		})
	}
}
