package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/cms/mocks"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	mocksJson "los-kmb-api/shared/common/json/mocks"
	"los-kmb-api/shared/common/platformevent"
	platformEventMockery "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
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

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

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
		mockUsecase.On("GetReasonPrescreening", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.ListReason(ctx)
		assert.Nil(t, err)
	})
}

func TestApprovalReason(t *testing.T) {
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
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/approval/reason?type=REJ&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetApprovalReason", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]entity.ApprovalReason{}, 0, nil).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Call the handler
		err := handler.ApprovalReason(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/approval/reason?type=APR", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.ApprovalReason{}
		statusCode := http.StatusNotFound

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockUsecase.On("GetApprovalReason", mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()
		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.ApprovalReason(ctx)
		assert.Nil(t, err)
	})

	t.Run("Error parameter approval reason", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/approval/reason", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.ApprovalReason(ctx)
		assert.Nil(t, err)
	})

	t.Run("bad request", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/approval/reason", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.ApprovalReason{}
		statusCode := http.StatusBadRequest
		mockUsecase.On("GetApprovalReason", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.ApprovalReason(ctx)
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
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/inquiry?search=aa&user_id=abc&branch_id=426&multi_branch=0&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything).Return([]entity.InquiryData{}, 0, nil).Once()

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		// Call the handler
		err := handler.PrescreeningInquiry(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/inquiry?search=aa&user_id=abc&branch_id=426&multi_branch=0&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockResponse := []entity.InquiryData{}
		statusCode := http.StatusOK

		mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := handler.PrescreeningInquiry(c)
		assert.Nil(t, err)
	})

	t.Run("internal server", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/inquiry?search=aa&user_id=abc&branch_id=426&page=1", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryData{}
		statusCode := http.StatusOK

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()

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

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockUsecase.On("GetInquiryPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.PrescreeningInquiry(ctx)
		assert.Nil(t, err)
	})

}

func TestReviewPrescreening(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	body := request.ReqReviewPrescreening{
		ProspectID: "EFM03406412522151348",
		Decision:   "APPROVE",
		Reason:     "sesuai",
		DecisionBy: "SYSTEM",
	}

	t.Run("success review", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()
		// var errData error

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockResponse := response.ReviewPrescreening{
			ProspectID: body.ProspectID,
			Code:       constant.CODE_PASS_PRESCREENING,
			Decision:   constant.DECISION_PASS,
			Reason:     "OK",
		}
		mockUsecase.On("ReviewPrescreening", mock.Anything, mock.Anything).Return(mockResponse, nil)

		mockJson.On("SuccessV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockJson.On("EventSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ApiResponse{})

		// Call the handler
		err := handler.ReviewPrescreening(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		reqID := utils.GenerateUUID()
		ctx.Set(constant.HeaderXRequestID, reqID)

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		err := handler.ReviewPrescreening(ctx)
		assert.Nil(t, err)
	})

	t.Run("error bad request", func(t *testing.T) {
		body.ProspectID = "EFM0340641252215134812345"
		data, _ := json.Marshal(body)

		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/prescreening/review", bytes.NewBuffer(data))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		reqID := utils.GenerateUUID()
		ctx.Set(constant.HeaderXRequestID, reqID)
		ctx.Request().Header.Add("content-type", "application/json")

		mockResponse := response.ReviewPrescreening{}
		statusCode := http.StatusBadRequest

		mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockUsecase.On("ReviewPrescreening", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.ReviewPrescreening(ctx)
		assert.Nil(t, err)

	})
}

func TestCMSHandler(t *testing.T) {
	e := echo.New()
	e.Validator = common.NewValidator()

	// Create a request and recorder for testing
	req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/prescreening/list-reason?reason_id=xxx&page=2", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	reqID := utils.GenerateUUID()
	c := e.NewContext(req, rec)
	c.Set(constant.HeaderXRequestID, reqID)

	// Create a test HTTP server
	srv := httptest.NewServer(e)

	// Define the mock middleware
	mockMiddleware := middlewares.NewAccessMiddleware()

	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)
	mockPlatformEvent := platformEventMockery.NewPlatformEventInterface(t)
	var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent
	// Initialize the handler with mocks or stubs
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}

	// Create a new Echo group and register the routes with the mock middleware
	cmsRoute := e.Group("/cms")
	CMSHandler(cmsRoute, mockUsecase, mockRepository, mockJson, platformEvent, mockMiddleware)

	mockResponse := []entity.ReasonMessage{}
	statusCode := http.StatusOK
	mockUsecase.On("GetReasonPrescreening", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()
	mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := handler.ListReason(c)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
	assert.Nil(t, err)

	// Cleanup resources if needed
	srv.Close()
}

func TestCaInquiry(t *testing.T) {
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
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/ca/inquiry?search=aa&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetInquiryCa", mock.Anything, mock.Anything, mock.Anything).Return([]entity.InquiryDataCa{}, 0, nil).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Call the handler
		err := handler.CaInquiry(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	// Create an
	t.Run("error not found", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/ca/inquiry", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryDataCa{}
		statusCode := http.StatusOK
		mockUsecase.On("GetInquiryCa", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.CaInquiry(ctx)
		assert.Nil(t, err)
	})

	t.Run("bad request", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/ca/inquiry", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryDataCa{}
		statusCode := http.StatusBadRequest
		mockUsecase.On("GetInquiryCa", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.CaInquiry(ctx)
		assert.Nil(t, err)
	})

}

func TestSearchInquiry(t *testing.T) {
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
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/search?user_id=1212&search=aa&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetSearchInquiry", mock.Anything, mock.Anything, mock.Anything).Return([]entity.InquiryDataSearch{}, 0, nil).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Call the handler
		err := handler.SearchInquiry(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error not found", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/search?user_id=1212&search=aa&page=2", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryDataSearch{}
		statusCode := http.StatusOK
		mockUsecase.On("GetSearchInquiry", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.SearchInquiry(ctx)
		assert.Nil(t, err)
	})

	t.Run("bad request", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/search?user_id=1212&search=aa&page=1", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryDataSearch{}
		statusCode := http.StatusBadRequest
		mockUsecase.On("GetSearchInquiry", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.SearchInquiry(ctx)
		assert.Nil(t, err)
	})

}

func TestCancelOrder(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	body := request.ReqCancelOrder{
		ProspectID:   "EFM03406412522151348",
		CancelReason: "belum sesuai",
		CreatedBy:    "5XeZs9PCeiPcZGS6azt",
		DecisionBy:   "User CA - KMB",
	}

	t.Run("success cancel", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()
		var errData error

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/cancel", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("CancelOrder", mock.Anything, mock.Anything).Return(response.CancelResponse{}, errData).Once()

		mockJson.On("SuccessV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		// Call the handler
		err := handler.CancelOrder(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/cancel", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		reqID := utils.GenerateUUID()
		ctx.Set(constant.HeaderXRequestID, reqID)

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		err := handler.CancelOrder(ctx)
		assert.Nil(t, err)
	})

	t.Run("error bad request", func(t *testing.T) {
		body.ProspectID = "EFM0340641252215134812345"
		data, _ := json.Marshal(body)

		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/cancel", bytes.NewBuffer(data))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		reqID := utils.GenerateUUID()
		ctx.Set(constant.HeaderXRequestID, reqID)
		ctx.Request().Header.Add("content-type", "application/json")

		mockResponse := response.CancelResponse{}
		statusCode := http.StatusBadRequest

		mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockUsecase.On("CancelOrder", mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.CancelOrder(ctx)
		assert.Nil(t, err)

	})
}

func TestReturnOrder(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	body := request.ReqReturnOrder{
		ProspectID: "EFM03406412522151348",
		CreatedBy:  "5XeZs9PCeiPcZGS6azt",
		DecisionBy: "User CA - KMB",
	}

	t.Run("success return", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()
		var errData error

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/return", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("ReturnOrder", mock.Anything, mock.Anything).Return(response.ReturnResponse{}, errData).Once()

		mockJson.On("SuccessV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		// Call the handler
		err := handler.ReturnOrder(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/return", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		reqID := utils.GenerateUUID()
		ctx.Set(constant.HeaderXRequestID, reqID)

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("InternalServerErrorCustomV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		err := handler.ReturnOrder(ctx)
		assert.Nil(t, err)
	})

	t.Run("error bad request", func(t *testing.T) {
		body.ProspectID = "EFM0340641252215134812345"
		data, _ := json.Marshal(body)

		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/return", bytes.NewBuffer(data))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		reqID := utils.GenerateUUID()
		ctx.Set(constant.HeaderXRequestID, reqID)
		ctx.Request().Header.Add("content-type", "application/json")

		mockResponse := response.ReturnResponse{}
		statusCode := http.StatusBadRequest

		mockJson.On("BadRequestErrorValidationV3", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		mockUsecase.On("ReturnOrder", mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.ReturnOrder(ctx)
		assert.Nil(t, err)

	})
}

func TestSaveAsDraft(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	body := request.ReqSaveAsDraft{
		ProspectID: "EFM03406412522151348",
		Decision:   "REJECT",
		SlikResult: "Dalam Perhatian Khusus",
		Note:       "Bahaya Nih",
		CreatedBy:  "5XeZs9PCeiPcZGS6azt",
		DecisionBy: "User CA - KMB",
	}

	t.Run("success save as draft", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()
		var errData error

		data, _ := json.Marshal(body)

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/save-as-draft", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("SaveAsDraft", mock.Anything, mock.Anything).Return(response.CAResponse{}, errData).Once()

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		// Call the handler
		err := handler.SaveAsDraft(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/save-as-draft", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		err := handler.SaveAsDraft(ctx)
		assert.Nil(t, err)
	})

	t.Run("error bad request", func(t *testing.T) {
		body.ProspectID = "EFM0340641252215134812345"
		data, _ := json.Marshal(body)

		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/save-as-draft", bytes.NewBuffer(data))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		ctx.Request().Header.Add("content-type", "application/json")

		mockResponse := response.CAResponse{}
		statusCode := http.StatusBadRequest

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("SaveAsDraft", mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.SaveAsDraft(ctx)
		assert.Nil(t, err)

	})
}

func TestSubmitDecision(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	body := request.ReqSubmitDecision{
		ProspectID:   "EFM03406412522151348",
		NTFAkumulasi: 20000019,
		Decision:     "REJECT",
		SlikResult:   "Dalam Perhatian Khusus",
		Note:         "Bahaya Nih",
		CreatedBy:    "5XeZs9PCeiPcZGS6azt",
		DecisionBy:   "User CA - KMB",
	}

	t.Run("success submit decision", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()
		var errData error

		data, _ := json.Marshal(body)

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/submit-decision", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("SubmitDecision", mock.Anything, mock.Anything).Return(response.CAResponse{}, errData).Once()

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		// Call the handler
		err := handler.SubmitDecision(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/submit-decision", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		err := handler.SubmitDecision(ctx)
		assert.Nil(t, err)
	})

	t.Run("error bad request", func(t *testing.T) {
		body.ProspectID = "EFM0340641252215134812345"
		data, _ := json.Marshal(body)

		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/ca/submit-decision", bytes.NewBuffer(data))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		ctx.Request().Header.Add("content-type", "application/json")

		mockResponse := response.CAResponse{}
		statusCode := http.StatusBadRequest

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("SubmitDecision", mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.SubmitDecision(ctx)
		assert.Nil(t, err)

	})
}

func TestCancelReason(t *testing.T) {
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
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/ca/cancel-reason?page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetCancelReason", mock.Anything, mock.Anything).Return([]entity.CancelReason{}, 0, nil).Once()

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Call the handler
		err := handler.CancelReason(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/ca/cancel-reason", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.CancelReason{}
		statusCode := http.StatusNotFound
		mockUsecase.On("GetCancelReason", mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()
		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.CancelReason(ctx)
		assert.Nil(t, err)
	})
}

func TestApprovalInquiry(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	t.Run("success_approval_inquiry", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/approval/inquiry?user_id=1212&search=aa&page=1", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetInquiryApproval", mock.Anything, mock.Anything, mock.Anything).Return([]entity.InquiryDataApproval{}, 0, nil).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Call the handler
		err := handler.ApprovalInquiry(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error not found", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/approval/inquiry?user_id=1212&search=aa&page=2", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryDataApproval{}
		statusCode := http.StatusOK
		mockUsecase.On("GetInquiryApproval", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New(constant.RECORD_NOT_FOUND)).Once()

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.ApprovalInquiry(ctx)
		assert.Nil(t, err)
	})

	t.Run("bad request approval", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/cms/approval/inquiry?user_id=1212&search=aa&page=1", nil)
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockResponse := []entity.InquiryDataApproval{}
		statusCode := http.StatusBadRequest
		mockUsecase.On("GetInquiryApproval", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		err := handler.ApprovalInquiry(ctx)
		assert.Nil(t, err)
	})
}

func TestSubmitApproval(t *testing.T) {
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	mockJson := new(mocksJson.JSON)

	// Create an instance of the handler
	handler := &handlerCMS{
		usecase:    mockUsecase,
		repository: mockRepository,
		Json:       mockJson,
	}
	body := request.ReqSubmitApproval{
		ProspectID:     "EFM03406412522151348",
		FinalApproval:  "UCC",
		Decision:       "REJECT",
		RuleCode:       "3741",
		Alias:          "CBM",
		Reason:         "Oke",
		NeedEscalation: false,
		Note:           "Bahaya Nih",
		CreatedBy:      "5XeZs9PCeiPcZGS6azt",
		DecisionBy:     "User CA - KMB",
	}

	t.Run("success submit approval", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()
		var errData error

		data, _ := json.Marshal(body)

		// Create a request and recorder for testing
		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/approval/submit-approval", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("SubmitApproval", mock.Anything, mock.Anything).Return(response.ApprovalResponse{}, errData).Once()

		mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		// Call the handler
		err := handler.SubmitApproval(c)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/approval/submit-approval", strings.NewReader("error"))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)

		mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		err := handler.SubmitApproval(ctx)
		assert.Nil(t, err)
	})

	t.Run("error bad request", func(t *testing.T) {
		body.ProspectID = "EFM0340641252215134812345"
		data, _ := json.Marshal(body)

		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/cms/approval/submit-approval", bytes.NewBuffer(data))
		rec := httptest.NewRecorder()

		ctx := e.NewContext(req, rec)
		ctx.Request().Header.Add("content-type", "application/json")

		mockResponse := response.ApprovalResponse{}
		statusCode := http.StatusBadRequest

		mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, response.ApiResponse{}).Once()

		mockUsecase.On("SubmitApproval", mock.Anything, mock.Anything).Return(mockResponse, statusCode, errors.New("failed")).Once()

		err := handler.SubmitApproval(ctx)
		assert.Nil(t, err)

	})
}
