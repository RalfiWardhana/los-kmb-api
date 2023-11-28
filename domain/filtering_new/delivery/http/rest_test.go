package http

import (
	"los-kmb-api/domain/filtering_new/interfaces/mocks"
	"los-kmb-api/middlewares"
	"los-kmb-api/shared/common"
	mocksJson "los-kmb-api/shared/common/json/mocks"
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
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": nil
			}`,
		},
		{
			name: "test customer spouse",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
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

			if tc.name == "test customer spouse" {
				mockPlatformEvent.On("PublishEvent", ctx.Request().Context(), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}
			mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			route := e.Group("/kmb")
			FilteringHandler(route, mockMultiUsecase, mockUsecase, mockRepository, mockJson, mockMiddleware, platformEvent)

			handler := &handlerKmbFiltering{
				multiusecase: mockMultiUsecase,
				usecase:      mockUsecase,
				repository:   mockRepository,
				Json:         mockJson,
				producer:     platformEvent,
			}
			err := handler.ProduceFiltering(ctx)
			if err != nil {
				t.Errorf("error '%s' was not expected, but got: ", err)
			}
			srv.Close()

		})
	}
}
