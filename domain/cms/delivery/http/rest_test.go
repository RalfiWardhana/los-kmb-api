package http

import (
	"errors"
	"los-kmb-api/domain/cms/interfaces/mocks"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	mocksJson "los-kmb-api/shared/common/json/mocks"
	"los-kmb-api/shared/constant"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

type MockEchoContext struct {
	mock.Mock
}

func (m *MockEchoContext) Validate(obj interface{}) error {
	args := m.Called(obj)
	return args.Error(0)
}

func TestListReason(t *testing.T) {
	// Create a new Echo instance

	e := echo.New()
	e.Validator = common.NewValidator()

	// Create a request and recorder for testing
	req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/list-reason?reason_id=1&page=1", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	mockUsecase.On("GetReasonPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.ReasonMessage{}, 0, nil).Once()

	mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}

	// Call the handler
	err := handler.ListReason(c)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestPrescreeningInquiry(t *testing.T) {
	// Create a new Echo instance

	e := echo.New()
	e.Validator = common.NewValidator()

	// Create a request and recorder for testing
	req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/inquiry?search=aa&page=1", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.InquiryData{}, 0, nil).Once()

	mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}

	// Call the handler
	err := handler.PrescreeningInquiry(c)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestReviewPrescreening(t *testing.T) {
	// Create a new Echo instance

	e := echo.New()
	e.Validator = common.NewValidator()
	var errData error

	body := `{
		"prospect_id": "EFM03406412522151348",
		"decision": "APPROVE",
		"reason": "sesuai",
		"decision_by": "SYSTEM"
	}`

	// Create a request and recorder for testing
	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	mockUsecase.On("ReviewPrescreening", mock.Anything, mock.Anything).Return(response.ReviewPrescreening{}, errData).Once()

	mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}

	// Call the handler
	err := handler.ReviewPrescreening(c)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestCMSHandler(t *testing.T) {
	// Create a new Echo instance
	e := echo.New()
	e.Validator = common.NewValidator()

	// Create a request and recorder for testing
	req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/list-reason?reason_id=1&page=1", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Create a test HTTP server
	srv := httptest.NewServer(e)

	// Define the mock middleware
	mockMiddleware := middlewares.NewAccessMiddleware()

	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	mockUsecase.On("GetReasonPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.ReasonMessage{}, 0, errors.New(constant.RECORD_NOT_FOUND)).Once()

	mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Initialize the handler with mocks or stubs
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}

	// Create a new Echo group and register the routes with the mock middleware
	cmsRoute := e.Group("/cms")
	CMSHandler(cmsRoute, mockUsecase, mockRepository, mockJson, mockMiddleware)

	// Test the ListReason route
	// Add more assertions here to verify the response body and middleware behavior if needed

	err := handler.ListReason(c)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}

	// Cleanup resources if needed
	srv.Close()
}
