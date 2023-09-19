package http

import (
	"errors"
	"los-kmb-api/domain/cms/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/common"
	mocksJson "los-kmb-api/shared/common/json/mocks"

	"los-kmb-api/shared/constant"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListReason(t *testing.T) {
	// Create a new Echo instance
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/list-reason?reason_id=1&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetReasonPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.ReasonMessage{}, 0, nil).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Call the handler
		err := handler.ListReason(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("Error parameter", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/list-reason", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		err := handler.ListReason(ctx)
		assert.Nil(t, err)
	})

	t.Run("bad request", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/list-reason", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.ReasonMessage{}
		statusCode := http.StatusBadRequest
		mockUsecase.On("GetReasonPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.ListReason(ctx)
		assert.Nil(t, err)
	})
}

func TestPrescreeningInquiry(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/inquiry?search=aa&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.InquiryData{}, 0, nil).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Call the handler
		err := handler.PrescreeningInquiry(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	// Create an
	t.Run("error not found", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/inquiry", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryData{}
		statusCode := http.StatusOK
		mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()

		err := handler.PrescreeningInquiry(ctx)
		assert.Nil(t, err)
	})

	t.Run("bad request", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/inquiry", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryData{}
		statusCode := http.StatusBadRequest
		mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.PrescreeningInquiry(ctx)
		assert.Nil(t, err)
	})

}

// func TestReviewPrescreening(t *testing.T) {
// 	mockUsecase := new(mocks.Usecase)
// 	mockRepository := new(mocks.Repository)
// 	mockJson := new(mocksJson.JSON)

// 	// Create an instance of the handler
// 	handler := &handlerCMS{
// 		usecase:    mockUsecase,
// 		repository: mockRepository,
// 		Json:       mockJson,
// 	}
// 	body := request.ReqReviewPrescreening{
// 		ProspectID: "EFM03406412522151348",
// 		Decision:   "APPROVE",
// 		Reason:     "sesuai",
// 		DecisionBy: "SYSTEM",
// 	}

// 	t.Run("success review", func(t *testing.T) {
// 		e := echo.New()
// 		e.Validator = common.NewValidator()
// 		var errData error

// 		data, _ := json.Marshal(body)

// 		reqID := utils.GenerateUUID()

// 		// Create a request and recorder for testing
// 		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", strings.NewReader(string(data)))
// 		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 		req.Header.Set(echo.HeaderXRequestID, reqID)
// 		rec := httptest.NewRecorder()
// 		c := e.NewContext(req, rec)

// 		mockUsecase.On("ReviewPrescreening", mock.Anything, mock.Anything).Return(response.ReviewPrescreening{}, errData).Once()

// 		mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

// 		// mockJson.On("ServerSideErrorV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 		mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

// 		mockJson.On("SuccessV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

// 		mockJson.On("EventSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ApiResponse{})

// 		// Call the handler
// 		err := handler.ReviewPrescreening(c)
// 		if err != nil {
// 			t.Errorf("error '%s' was not expected, but got: ", err)
// 		}
// 	})

// 	t.Run("error bind", func(t *testing.T) {
// 		e := echo.New()

// 		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", strings.NewReader("error"))
// 		rec := httptest.NewRecorder()

// 		ctx := e.NewContext(req, rec)

// 		err := handler.ReviewPrescreening(ctx)
// 		assert.Nil(t, err)
// 	})

// 	t.Run("error bad request", func(t *testing.T) {
// 		body.ProspectID = "EFM0340641252215134812345"
// 		data, _ := json.Marshal(body)

// 		e := echo.New()

// 		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", bytes.NewBuffer(data))
// 		rec := httptest.NewRecorder()

// 		ctx := e.NewContext(req, rec)
// 		ctx.Request().Header.Add("content-type", "application/json")
// 		mockResponse := response.ReviewPrescreening{}
// 		statusCode := http.StatusBadRequest
// 		mockUsecase.On("ReviewPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

// 		err := handler.ReviewPrescreening(ctx)
// 		assert.Nil(t, err)
// 	})
// }

// func TestCMSHandler(t *testing.T) {
// 	// Create a new Echo instance
// 	e := echo.New()
// 	e.Validator = common.NewValidator()

// 	body := request.ReqReviewPrescreening{
// 		ProspectID: "EFM03406412522151348",
// 		Decision:   "APPROVE",
// 		Reason:     "sesuai",
// 		DecisionBy: "SYSTEM",
// 	}

// 	// Create a request and recorder for testing
// 	data, _ := json.Marshal(body)
// 	// Create a request and recorder for testing
// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", strings.NewReader(string(data)))
// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
// 	rec := httptest.NewRecorder()
// 	c := e.NewContext(req, rec)

// 	// Create a test HTTP server
// 	srv := httptest.NewServer(e)

// 	// Define the mock middleware
// 	mockMiddleware := middlewares.NewAccessMiddleware()

// 	mockUsecase := new(mocks.Usecase)
// 	mockRepository := new(mocks.Repository)
// 	mockJson := new(mocksJson.JSON)
// 	mockPlatformEvent := platformEventMockery.NewPlatformEventInterface(t)
// 	var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

// 	mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	mockUsecase.On("ReviewPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ReviewPrescreening{}, 0, errors.New(constant.RECORD_NOT_FOUND)).Once()

// 	mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	mockJson.On("ServerSideErrorV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	mockJson.On("EventServiceError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ApiResponse{})

// 	mockJson.On("EventSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ApiResponse{})

// 	// mockPlatformEvent.On("PublishEvent", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
// 	// 	Return(nil).Once()

// 	mockJson.On("SuccessV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

// 	// Initialize the handler with mocks or stubs
// 	handler := &handlerCMS{
// 		usecase:    mockUsecase,
// 		repository: mockRepository,
// 		Json:       mockJson,
// 	}

// 	// Create a new Echo group and register the routes with the mock middleware
// 	cmsRoute := e.Group("/cms")
// 	CMSHandler(cmsRoute, mockUsecase, mockRepository, mockJson, platformEvent, mockMiddleware)

// 	// Test the ListReason route
// 	// Add more assertions here to verify the response body and middleware behavior if needed

// 	err := handler.ReviewPrescreening(c)
// 	if err != nil {
// 		t.Errorf("error '%s' was not expected, but got: ", err)
// 	}

// 	// Cleanup resources if needed
// 	srv.Close()
// }
