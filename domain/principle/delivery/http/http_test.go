package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/request"
	responses "los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/KB-FMF/los-common-library/response"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVerifyAsset(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.PrincipleAsset{
		ProspectID:         "SAL-1140024080800017",
		IDNumber:           "3505151204000001",
		SpouseIDNumber:     "3506126712000001",
		ManufactureYear:    2020,
		NoChassis:          "MHKV1AA2JBK107322",
		NoEngine:           "73218JAJK",
		BranchID:           "426",
		CC:                 1500,
		TaxDate:            "2022-03-02",
		STNKExpiredDate:    "2025-03-20",
		OwnerAsset:         "JONATHAN",
		LicensePlate:       "B3006TBJ",
		Color:              "HITAM",
		Brand:              "HONDA",
		ResidenceAddress:   "Jl. PATIMURA",
		ResidenceRT:        "001",
		ResidenceRW:        "002",
		ResidenceProvince:  "JAWA TIMUR",
		ResidenceCity:      "KOTA MALANG",
		ResidenceKecamatan: "LOWOKWARU",
		ResidenceKelurahan: "LOWOKWARU",
		ResidenceZipCode:   "65111",
		ResidenceAreaPhone: "021",
		ResidencePhone:     "86605224",
		HomeStatus:         "SD",
		StaySinceYear:      2024,
		StaySinceMonth:     4,
		AssetCode:          "K-KWS.MOTOR.SMASH MUFLER",
		STNKPhoto:          "http://www.example.com",
		KPMID:              123456,
	}

	// t.Run("success", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	data, _ := json.Marshal(body)

	// 	reqID := utils.GenerateUUID()

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-asset", strings.NewReader(string(data)))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	req.Header.Set(echo.HeaderXRequestID, reqID)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	c.Set(constant.HeaderXRequestID, reqID)

	// 	mockResponse := responses.UsecaseApi{
	// 		Code:   constant.CODE_AGREEMENT_NOT_FOUND,
	// 		Result: constant.DECISION_PASS,
	// 		Reason: constant.REASON_AGREEMENT_NOT_FOUND,
	// 	}
	// 	mockUsecase.On("CheckNokaNosin", mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

	// 	_ = handler.VerifyAsset(c)

	// 	assert.Equal(t, http.StatusOK, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")

	// 	mockUsecase.AssertExpectations(t)
	// })

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		_, _ = json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-asset", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.VerifyAsset(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	// t.Run("error validate", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	invalidBody := body
	// 	invalidBody.ProspectID = ""

	// 	data, _ := json.Marshal(invalidBody)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-asset", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	_ = handler.VerifyAsset(c)

	// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	// })

	// t.Run("error usecase", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	data, _ := json.Marshal(body)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-asset", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	mockUsecase.On("CheckNokaNosin", mock.Anything, mock.Anything).Return(responses.UsecaseApi{}, errors.New("some error")).Once()

	// 	_ = handler.VerifyAsset(c)

	// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	// })
}

func TestVerifyPemohon(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.PrinciplePemohon{
		ProspectID:              "SAL-1140024080800017",
		IDNumber:                "3505151204000001",
		SpouseIDNumber:          "3506126712000001",
		MobilePhone:             "085880529100",
		Email:                   "test@test.com",
		LegalName:               "Test",
		FullName:                "Test",
		BirthDate:               "1993-11-12",
		BirthPlace:              "JEMBER",
		SurgateMotherName:       "IBU",
		Gender:                  "M",
		Religion:                "1",
		LegalAddress:            "Jl. PATIMURA",
		LegalRT:                 "001",
		LegalRW:                 "003",
		LegalProvince:           "JAWA TIMUR",
		LegalCity:               "KOTA MALANG",
		LegalKecamatan:          "LOWOKWARU",
		LegalKelurahan:          "LOWOKWARU",
		LegalZipCode:            "65111",
		LegalPhoneArea:          "104",
		LegalPhone:              "86605224",
		Education:               "S1",
		ProfessionID:            "KRYSW",
		JobType:                 "0012",
		JobPosition:             "M",
		EmploymentSinceMonth:    2,
		EmploymentSinceYear:     2020,
		CompanyName:             "PT KB Finansia",
		EconomySectorID:         "06",
		IndustryTypeID:          "1000",
		CompanyAddress:          "Dermaga Lama",
		CompanyRT:               "001",
		CompanyRW:               "003",
		CompanyProvince:         "Jawa Barat",
		CompanyCity:             "Bandung",
		CompanyKecamatan:        "Lembang",
		CompanyKelurahan:        "Lembang",
		CompanyZipCode:          "13470",
		CompanyPhoneArea:        "021",
		CompanyPhone:            "86605224",
		MonthlyFixedIncome:      5000000,
		MaritalStatus:           "M",
		SpouseLegalName:         "YULINAR NIATI",
		SpouseFullName:          "YULINAR NIATI",
		SpouseBirthDate:         "1992-09-11",
		SpouseBirthPlace:        "Jakarta",
		SpouseSurgateMotherName: "MAMA",
		SpouseMobilePhone:       "085880529111",
		SpouseIncome:            5000000,
		SelfiePhoto:             "https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/selfie_SAL-1140024081400003.jpg",
		KtpPhoto:                "https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/ktp_SAL-1140024081400003.jpg",
	}

	// t.Run("success", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	data, _ := json.Marshal(body)

	// 	reqID := utils.GenerateUUID()

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pemohon", strings.NewReader(string(data)))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	req.Header.Set(echo.HeaderXRequestID, reqID)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	c.Set(constant.HeaderXRequestID, reqID)

	// 	mockResponse := responses.UsecaseApi{
	// 		Code:   constant.CODE_PERNAH_REJECT_PMK_DSR,
	// 		Result: constant.DECISION_REJECT,
	// 		Reason: constant.REASON_PERNAH_REJECT_PMK_DSR,
	// 	}
	// 	mockMultiUsecase.On("PrinciplePemohon", mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

	// 	_ = handler.VerifyPemohon(c)

	// 	assert.Equal(t, http.StatusOK, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")

	// 	mockUsecase.AssertExpectations(t)
	// })

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		_, _ = json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pemohon", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.VerifyPemohon(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	// t.Run("error validate", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	invalidBody := body
	// 	invalidBody.ProspectID = ""

	// 	data, _ := json.Marshal(invalidBody)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pemohon", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	_ = handler.VerifyPemohon(c)

	// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	// })

	// t.Run("error usecase", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	data, _ := json.Marshal(body)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pemohon", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	mockMultiUsecase.On("PrinciplePemohon", mock.Anything, mock.Anything).Return(responses.UsecaseApi{}, errors.New("some error")).Once()

	// 	_ = handler.VerifyPemohon(c)

	// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	// })
}

func TestStepPrinciple(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockUsecase := new(mocks.Usecase)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		usecase:   mockUsecase,
		responses: libResponse,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-principle/:id_number")
		c.SetParamNames("id_number")
		c.SetParamValues("3505151204000001")

		mockData := responses.StepPrinciple{
			ProspectID: "SAL-1140024080800017",
			ColorCode:  "#00FF00",
			Status:     constant.REASON_PROFIL_APPROVE,
		}
		mockUsecase.On("PrincipleStep", "3505151204000001").Return(mockData, nil).Once()

		_ = handler.StepPrinciple(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")
		mockUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/step-principle/3505151204000001", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.StepPrinciple(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/step-principle/invalid_id", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.StepPrinciple(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-principle/:id_number")
		c.SetParamNames("id_number")
		c.SetParamValues("3505151204000001")

		mockUsecase.On("PrincipleStep", "3505151204000001").Return(responses.StepPrinciple{}, errors.New("some error")).Once()

		_ = handler.StepPrinciple(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
		mockUsecase.AssertExpectations(t)
	})
}

func TestElaborateLTV(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.PrincipleElaborateLTV{
		ProspectID:     "SAL-1140024080800017",
		Tenor:          12,
		FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
		LoanAmount:     1000000,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/elaborate-ltv", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockResponse := responses.PrincipleElaborateLTV{
			LTV: 80,
		}
		mockUsecase.On("PrincipleElaborateLTV", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

		_ = handler.ElaborateLTV(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-004")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/elaborate-ltv", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.ElaborateLTV(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		invalidBody := body
		invalidBody.ProspectID = ""

		data, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/elaborate-ltv", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.ElaborateLTV(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/elaborate-ltv", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("PrincipleElaborateLTV", mock.Anything, mock.Anything, mock.Anything).Return(responses.PrincipleElaborateLTV{}, errors.New("some error")).Once()

		_ = handler.ElaborateLTV(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	})
}

func TestVerifyPembiayaan(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	// body := request.PrinciplePembiayaan{
	// 	ProspectID:        "SAL-1140024080800017",
	// 	Tenor:             12,
	// 	AF:                106000000,
	// 	NTF:               23500000,
	// 	OTR:               5650000,
	// 	DPAmount:          1900000,
	// 	AdminFee:          2000000,
	// 	InstallmentAmount: 4935000,
	// 	Dealer:            "NON PSA",
	// 	AssetCategoryID:   "BEBEK",
	// 	FinancePurpose:    "Modal Kerja Fasilitas Modal Usaha",
	// 	TipeUsaha:         "Jasa Kesehatan",
	// }

	// t.Run("success", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	data, _ := json.Marshal(body)

	// 	reqID := utils.GenerateUUID()

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pembiayaan", strings.NewReader(string(data)))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	req.Header.Set(echo.HeaderXRequestID, reqID)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	c.Set(constant.HeaderXRequestID, reqID)

	// 	mockResponse := responses.UsecaseApi{
	// 		Code:   constant.CODE_AGREEMENT_NOT_FOUND,
	// 		Result: constant.DECISION_PASS,
	// 		Reason: constant.REASON_AGREEMENT_NOT_FOUND,
	// 	}
	// 	mockMultiUsecase.On("PrinciplePembiayaan", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

	// 	_ = handler.VerifyPembiayaan(c)

	// 	assert.Equal(t, http.StatusOK, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")

	// 	mockUsecase.AssertExpectations(t)
	// })

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pembiayaan", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.VerifyPembiayaan(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	// t.Run("error validate", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	invalidBody := body
	// 	invalidBody.ProspectID = ""

	// 	data, _ := json.Marshal(invalidBody)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pembiayaan", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	_ = handler.VerifyPembiayaan(c)

	// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	// })

	// t.Run("error usecase", func(t *testing.T) {
	// 	e := echo.New()
	// 	e.Validator = common.NewValidator()

	// 	data, _ := json.Marshal(body)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/verify-pembiayaan", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	mockMultiUsecase.On("PrinciplePembiayaan", mock.Anything, mock.Anything, mock.Anything).Return(responses.UsecaseApi{}, errors.New("some error")).Once()

	// 	_ = handler.VerifyPembiayaan(c)

	// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	// })
}

type MockValidator struct {
	mock.Mock
}

func (m *MockValidator) Validate(i interface{}) error {
	args := m.Called(i)
	return args.Error(0)
}

func TestEmergencyContact(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	// body := request.PrincipleEmergencyContact{
	// 	ProspectID:   "SAL-1140024080800016",
	// 	Name:         "MULYADI",
	// 	Relationship: "FM",
	// 	MobilePhone:  "08567891231",
	// 	Address:      "JL.PEGANGSAAN 1",
	// 	Rt:           "008",
	// 	Rw:           "017",
	// 	Kelurahan:    "TEGAL PARANG",
	// 	Kecamatan:    "MAMPANG PRAPATAN",
	// 	City:         "JAKARTA SELATAN",
	// 	Province:     "DKI JAKARTA",
	// 	ZipCode:      "12790",
	// 	AreaPhone:    "021",
	// 	Phone:        "567892",
	// }

	// t.Run("success", func(t *testing.T) {
	// 	e := echo.New()
	// 	mockValidator := new(MockValidator)
	// 	mockValidator.On("Validate", mock.Anything).Return(nil)

	// 	e.Validator = mockValidator

	// 	data, _ := json.Marshal(body)

	// 	reqID := utils.GenerateUUID()

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/emergency-contact", strings.NewReader(string(data)))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	req.Header.Set(echo.HeaderXRequestID, reqID)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	c.Set(constant.HeaderXRequestID, reqID)

	// 	mockResponse := responses.UsecaseApi{
	// 		Code:   constant.EMERGENCY_PASS_CODE,
	// 		Result: constant.DECISION_PASS,
	// 		Reason: constant.EMERGENCY_PASS_REASON,
	// 	}
	// 	mockMultiUsecase.On("PrincipleEmergencyContact", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

	// 	_ = handler.EmergencyContact(c)

	// 	assert.Equal(t, http.StatusOK, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")

	// 	mockUsecase.AssertExpectations(t)
	// 	mockValidator.AssertExpectations(t)
	// })

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/emergency-contact", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.EmergencyContact(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	// t.Run("error validate", func(t *testing.T) {
	// 	e := echo.New()
	// 	mockValidator := new(MockValidator)
	// 	mockValidator.On("Validate", mock.Anything).Return(errors.New("validation error"))

	// 	e.Validator = mockValidator

	// 	invalidBody := body
	// 	invalidBody.ProspectID = ""

	// 	data, _ := json.Marshal(invalidBody)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/emergency-contact", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	_ = handler.EmergencyContact(c)

	// 	assert.Equal(t, http.StatusBadRequest, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	// })

	// t.Run("error usecase", func(t *testing.T) {
	// 	e := echo.New()
	// 	mockValidator := new(MockValidator)
	// 	mockValidator.On("Validate", mock.Anything).Return(nil)

	// 	e.Validator = mockValidator

	// 	data, _ := json.Marshal(body)

	// 	req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/emergency-contact", bytes.NewReader(data))
	// 	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	// 	rec := httptest.NewRecorder()
	// 	c := e.NewContext(req, rec)

	// 	mockMultiUsecase.On("PrincipleEmergencyContact", mock.Anything, mock.Anything, mock.Anything).Return(responses.UsecaseApi{}, errors.New("some error")).Once()

	// 	_ = handler.EmergencyContact(c)

	// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	// 	assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	// })
}

func TestCoreCustomer(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/core-customer/:prospectID")
		c.SetParamNames("prospectID")
		c.SetParamValues("SAL-1140024080800017")

		c.Set(constant.HeaderXRequestID, reqID)

		mockUsecase.On("PrincipleCoreCustomer", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_ = handler.CoreCustomer(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error prospect id empty", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/core-customer/:prospectID")
		c.SetParamNames("prospectID")
		c.SetParamValues("")

		_ = handler.CoreCustomer(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/core-customer/:prospectID")
		c.SetParamNames("prospectID")
		c.SetParamValues("SAL-1140024080800017")

		mockUsecase.On("PrincipleCoreCustomer", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("some error")).Once()

		_ = handler.CoreCustomer(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	})
}

func TestMarketingProgram(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/marketing-program/:prospectID")
		c.SetParamNames("prospectID")
		c.SetParamValues("SAL-1140024080800017")

		c.Set(constant.HeaderXRequestID, reqID)

		mockUsecase.On("PrincipleMarketingProgram", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_ = handler.MarketingProgram(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error prospect id empty", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/marketing-program/:prospectID")
		c.SetParamNames("prospectID")
		c.SetParamValues("")

		_ = handler.MarketingProgram(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/marketing-program/:prospectID")
		c.SetParamNames("prospectID")
		c.SetParamValues("SAL-1140024080800017")

		mockUsecase.On("PrincipleMarketingProgram", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("some error")).Once()

		_ = handler.MarketingProgram(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	})
}

func TestGetPrincipleData(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.PrincipleGetData{
		Context:        "PRINCIPLE",
		ProspectID:     "SAL-1140024080800017",
		FinancePurpose: "Modal Kerja",
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-data", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockResponse := map[string]interface{}{
			"prospect_id": "SAL-1140024080800017",
			"status":      "ACTIVE",
			"data": map[string]interface{}{
				"customer_info": map[string]interface{}{
					"id_number":    "3505151204000001",
					"full_name":    "Test Customer",
					"birth_date":   "1993-11-12",
					"birth_place":  "JEMBER",
					"mobile_phone": "085880529100",
				},
			},
		}

		mockUsecase.On("GetDataPrinciple", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

		_ = handler.GetPrincipleData(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-data", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.GetPrincipleData(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		invalidBody := body
		invalidBody.Context = ""
		invalidBody.ProspectID = ""

		data, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-data", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.GetPrincipleData(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-data", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("GetDataPrinciple", mock.Anything, mock.Anything, mock.Anything).Return(map[string]interface{}{}, errors.New("some error")).Once()

		_ = handler.GetPrincipleData(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-")
	})
}

func TestAutoCancel(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/auto-cancel", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockUsecase.On("CheckOrderPendingPrinciple", mock.Anything).Return(nil).Once()

		_ = handler.AutoCancel(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")
		assert.Contains(t, rec.Body.String(), "sukses auto cancel")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		req := httptest.NewRequest(http.MethodGet, "/api/v3/kmb/auto-cancel", nil)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("CheckOrderPendingPrinciple", mock.Anything).Return(errors.New("some error")).Once()

		_ = handler.AutoCancel(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-")

		mockUsecase.AssertExpectations(t)
	})
}

func TestPrinciplePublish(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.PrinciplePublish{
		StatusCode: "APPROVE",
		ProspectID: "SAL-1140024080800017",
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-publish", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockUsecase.On("PrinciplePublish", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_ = handler.PrinciplePublish(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")
		assert.Contains(t, rec.Body.String(), "success publish event principle")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-publish", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.PrinciplePublish(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		invalidBody := body
		invalidBody.StatusCode = ""
		invalidBody.ProspectID = ""

		data, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-publish", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.PrinciplePublish(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-publish", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("PrinciplePublish", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("some error")).Once()

		_ = handler.PrinciplePublish(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with different status code", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		differentBody := request.PrinciplePublish{
			StatusCode: "REJECT",
			ProspectID: "SAL-1140024080800017",
		}

		data, _ := json.Marshal(differentBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/principle-publish", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("PrinciplePublish", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_ = handler.PrinciplePublish(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "PRINCIPLE-001")
		assert.Contains(t, rec.Body.String(), "success publish event principle")

		mockUsecase.AssertExpectations(t)
	})
}

func TestGetMaxLoanAmount(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")
	os.Setenv("NAMA_SAMA", "K,P")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.GetMaxLoanAmount{
		ProspectID:         "SAL-1140024080800004",
		BranchID:           "426",
		IDNumber:           "3506126712000001",
		BirthDate:          "1992-09-11",
		SurgateMotherName:  "IBU",
		LegalName:          "Arya Danu",
		MobilePhone:        "085880529100",
		BPKBNameType:       "K",
		ManufactureYear:    "2020",
		AssetCode:          "SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT",
		AssetUsageTypeCode: "C",
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/max-loan-amount", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockResponse := responses.GetMaxLoanAmountData{
			MaxLoanAmount: 50000000,
		}

		mockMultiUsecase.On("GetMaxLoanAmout", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

		_ = handler.GetMaxLoanAmount(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")

		mockMultiUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/max-loan-amount", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.GetMaxLoanAmount(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		invalidBody := body
		invalidBody.ProspectID = ""
		invalidBody.BranchID = ""

		data, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/max-loan-amount", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.GetMaxLoanAmount(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/max-loan-amount", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockMultiUsecase.On("GetMaxLoanAmout", mock.Anything, mock.Anything, mock.Anything).Return(responses.GetMaxLoanAmountData{}, errors.New("some error")).Once()

		_ = handler.GetMaxLoanAmount(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-")

		mockMultiUsecase.AssertExpectations(t)
	})

	t.Run("success with different amount", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/max-loan-amount", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockResponse := responses.GetMaxLoanAmountData{
			MaxLoanAmount: 75000000,
		}

		mockMultiUsecase.On("GetMaxLoanAmout", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

		_ = handler.GetMaxLoanAmount(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")

		mockMultiUsecase.AssertExpectations(t)
	})
}

func TestGetAvailableTenor(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")
	os.Setenv("NAMA_SAMA", "K,P")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.GetAvailableTenor{
		ProspectID:         "SAL-1140024080800004",
		BranchID:           "426",
		IDNumber:           "3506126712000001",
		BirthDate:          "1992-09-11",
		SurgateMotherName:  "IBU",
		LegalName:          "Arya Danu",
		MobilePhone:        "085880529100",
		BPKBNameType:       "K",
		ManufactureYear:    "2020",
		AssetCode:          "SUZUKI,KMOBIL,GRAND VITARA.JLX 2,0 AT",
		AssetUsageTypeCode: "C",
		LicensePlate:       "B3006TBJ",
		LoanAmount:         105000000,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/available-tenor", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockResponse := []responses.GetAvailableTenorData{
			{
				Tenor:             12,
				IsPsa:             false,
				Dealer:            "NON PSA",
				InstallmentAmount: 4935000,
				AF:                106000000,
				AdminFee:          2000000,
				DPAmount:          1900000,
				NTF:               23500000,
				AssetCategoryID:   "BEBEK",
				OTR:               5650000,
			},
		}

		mockMultiUsecase.On("GetAvailableTenor", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

		_ = handler.GetAvailableTenor(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")

		mockMultiUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/available-tenor", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.GetAvailableTenor(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		invalidBody := body
		invalidBody.ProspectID = ""
		invalidBody.BranchID = ""

		data, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/available-tenor", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.GetAvailableTenor(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/available-tenor", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockMultiUsecase.On("GetAvailableTenor", mock.Anything, mock.Anything, mock.Anything).Return([]responses.GetAvailableTenorData{}, errors.New("some error")).Once()

		_ = handler.GetAvailableTenor(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-")

		mockMultiUsecase.AssertExpectations(t)
	})

	t.Run("success with multiple tenors", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/available-tenor", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockResponse := []responses.GetAvailableTenorData{
			{
				Tenor:             12,
				IsPsa:             false,
				Dealer:            "NON PSA",
				InstallmentAmount: 4935000,
				AF:                106000000,
				AdminFee:          2000000,
				DPAmount:          1900000,
				NTF:               23500000,
				AssetCategoryID:   "BEBEK",
				OTR:               5650000,
			},
			{
				Tenor:             24,
				IsPsa:             false,
				Dealer:            "NON PSA",
				InstallmentAmount: 2935000,
				AF:                106000000,
				AdminFee:          2000000,
				DPAmount:          1900000,
				NTF:               23500000,
				AssetCategoryID:   "BEBEK",
				OTR:               5650000,
			},
		}

		mockMultiUsecase.On("GetAvailableTenor", mock.Anything, mock.Anything, mock.Anything).Return(mockResponse, nil).Once()

		_ = handler.GetAvailableTenor(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")

		mockMultiUsecase.AssertExpectations(t)
	})
}

func TestHistory2Wilen(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")
	os.Setenv("NAMA_SAMA", "K,P")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	prospectID := "SAL-1140024080800004"
	body := request.History2Wilen{
		ProspectID: &prospectID,
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/2wilen/history", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockResponse := []responses.History2Wilen{
			{
				ID:              "1",
				ProspectID:      "SAL-1140024080800004",
				OrderStatusName: "PENDING",
				CreatedAt:       "2024-01-31 10:00:00",
			},
		}

		mockUsecase.On("History2Wilen", body).Return(mockResponse, nil).Once()

		_ = handler.History2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/2wilen/history", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.History2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		prospectID2 := prospectID + "SAL-1140024080800004SAL-1140024080800004"
		invalidBody := request.History2Wilen{
			ProspectID: &prospectID2,
		}

		data, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/2wilen/history", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.History2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/2wilen/history", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("History2Wilen", body).Return([]responses.History2Wilen{}, errors.New("some error")).Once()

		_ = handler.History2Wilen(c)

		fmt.Println(rec.Body.String())
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with multiple history records", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/2wilen/history", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockResponse := []responses.History2Wilen{
			{
				ID:              "1",
				ProspectID:      "SAL-1140024080800004",
				OrderStatusName: "PENDING",
				CreatedAt:       "2024-01-31 10:00:00",
			},
			{
				ID:              "2",
				ProspectID:      "SAL-1140024080800004",
				OrderStatusName: "APPROVED",
				CreatedAt:       "2024-01-31 11:00:00",
			},
		}

		mockUsecase.On("History2Wilen", body).Return(mockResponse, nil).Once()

		_ = handler.History2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")

		mockUsecase.AssertExpectations(t)
	})
}

func TestPublish2Wilen(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")
	os.Setenv("NAMA_SAMA", "K,P")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	body := request.Publish2Wilen{
		StatusCode: "APPROVE",
		ProspectID: "SAL-1140024080800017",
	}

	t.Run("success", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		reqID := utils.GenerateUUID()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/publish-2wilen", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderXRequestID, reqID)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set(constant.HeaderXRequestID, reqID)

		mockUsecase.On("Publish2Wilen", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_ = handler.Publish2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")
		assert.Contains(t, rec.Body.String(), "success publish event 2wilen")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/publish-2wilen", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Publish2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		invalidBody := request.Publish2Wilen{
			StatusCode: "",
			ProspectID: "",
		}

		data, _ := json.Marshal(invalidBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/publish-2wilen", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Publish2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		data, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/publish-2wilen", bytes.NewReader(data))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("Publish2Wilen", mock.Anything, mock.Anything, mock.Anything).
			Return(errors.New("some error")).Once()

		_ = handler.Publish2Wilen(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-")

		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with different status code", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		differentBody := request.Publish2Wilen{
			StatusCode: "REJECT",
			ProspectID: "SAL-1140024080800017",
		}

		data, _ := json.Marshal(differentBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/publish-2wilen", strings.NewReader(string(data)))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mockUsecase.On("Publish2Wilen", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		_ = handler.Publish2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")
		assert.Contains(t, rec.Body.String(), "success publish event 2wilen")

		mockUsecase.AssertExpectations(t)
	})
}

func TestStep2Wilen(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")
	os.Setenv("PLATFORM_LIBRARY_KEY", "PLATFORMS-APIToEncryptDecryptAPI")

	mockUsecase := new(mocks.Usecase)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		usecase:   mockUsecase,
		responses: libResponse,
	}

	t.Run("success with data", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		encryptedID, _ := utils.PlatformEncryptText("3505151204000001")
		reqBody := request.CheckStep2Wilen{
			IDNumber: encryptedID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-2wilen")

		mockData := responses.Step2Wilen{
			ProspectID: "SAL-1140024080800017",
			ColorCode:  "#00FF00",
			Status:     "APPROVED",
			UpdatedAt:  "2025-04-09T10:15:30Z",
		}
		mockUsecase.On("Step2Wilen", "3505151204000001").Return(mockData, nil).Once()

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")
		assert.Contains(t, rec.Body.String(), "SAL-1140024080800017")
		assert.Contains(t, rec.Body.String(), "#00FF00")
		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with empty status", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		encryptedID, _ := utils.PlatformEncryptText("3505151204000001")
		reqBody := request.CheckStep2Wilen{
			IDNumber: encryptedID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-2wilen")

		mockData := responses.Step2Wilen{
			ProspectID: "SAL-1140024080800017",
			ColorCode:  "#CCCCCC",
			Status:     "",
			UpdatedAt:  "2025-04-09T10:15:30Z",
		}
		mockUsecase.On("Step2Wilen", "3505151204000001").Return(mockData, nil).Once()

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-001")
		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with ongoing application - DECISION_KPM_READJUST", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		encryptedID, _ := utils.PlatformEncryptText("3505151204000001")
		reqBody := request.CheckStep2Wilen{
			IDNumber: encryptedID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-2wilen")

		mockData := responses.Step2Wilen{
			ProspectID: "SAL-1140024080800017",
			ColorCode:  "#FFCC00",
			Status:     constant.DECISION_KPM_READJUST,
			UpdatedAt:  "2025-04-09T10:15:30Z",
		}
		mockUsecase.On("Step2Wilen", "3505151204000001").Return(mockData, nil).Once()

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-002")
		assert.Contains(t, rec.Body.String(), "Kamu masih memiliki pengajuan lain yang sedang diproses")
		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with ongoing application - STATUS_KPM_WAIT_2WILEN", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		encryptedID, _ := utils.PlatformEncryptText("3505151204000001")
		reqBody := request.CheckStep2Wilen{
			IDNumber: encryptedID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-2wilen")

		mockData := responses.Step2Wilen{
			ProspectID: "SAL-1140024080800017",
			ColorCode:  "#FFCC00",
			Status:     constant.STATUS_KPM_WAIT_2WILEN,
			UpdatedAt:  "2025-04-09T10:15:30Z",
		}
		mockUsecase.On("Step2Wilen", "3505151204000001").Return(mockData, nil).Once()

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-002")
		assert.Contains(t, rec.Body.String(), "Kamu masih memiliki pengajuan lain yang sedang diproses")
		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with ongoing application - DECISION_KPM_APPROVE", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		encryptedID, _ := utils.PlatformEncryptText("3505151204000001")
		reqBody := request.CheckStep2Wilen{
			IDNumber: encryptedID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-2wilen")

		mockData := responses.Step2Wilen{
			ProspectID: "SAL-1140024080800017",
			ColorCode:  "#00FF00",
			Status:     constant.DECISION_KPM_APPROVE,
			UpdatedAt:  "2025-04-09T10:15:30Z",
		}
		mockUsecase.On("Step2Wilen", "3505151204000001").Return(mockData, nil).Once()

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-002")
		assert.Contains(t, rec.Body.String(), "Kamu masih memiliki pengajuan lain yang sedang diproses")
		mockUsecase.AssertExpectations(t)
	})

	t.Run("success with ongoing application - STATUS_LOS_PROCESS_2WILEN", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		encryptedID, _ := utils.PlatformEncryptText("3505151204000001")
		reqBody := request.CheckStep2Wilen{
			IDNumber: encryptedID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-2wilen")

		mockData := responses.Step2Wilen{
			ProspectID: "SAL-1140024080800017",
			ColorCode:  "#FFCC00",
			Status:     constant.STATUS_LOS_PROCESS_2WILEN,
			UpdatedAt:  "2025-04-09T10:15:30Z",
		}
		mockUsecase.On("Step2Wilen", "3505151204000001").Return(mockData, nil).Once()

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-002")
		assert.Contains(t, rec.Body.String(), "Kamu masih memiliki pengajuan lain yang sedang diproses")
		mockUsecase.AssertExpectations(t)
	})

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/step-2wilen", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		reqBody := struct {
			InvalidField string `json:"invalid_field"`
		}{
			InvalidField: "test",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/step-2wilen", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-800")
	})

	t.Run("error usecase", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		encryptedID, _ := utils.PlatformEncryptText("3505151204000001")
		reqBody := request.CheckStep2Wilen{
			IDNumber: encryptedID,
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v3/kmb/step-2wilen")

		mockUsecase.On("Step2Wilen", "3505151204000001").Return(responses.Step2Wilen{}, errors.New("some error")).Once()

		_ = handler.Step2Wilen(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-")
		assert.Contains(t, rec.Body.String(), constant.PRINCIPLE_ERROR_RESPONSE_MESSAGE)
		mockUsecase.AssertExpectations(t)
	})
}

func TestSubmission2Wilen(t *testing.T) {
	os.Setenv("APP_PREFIX_NAME", "LOS")

	mockMultiUsecase := new(mocks.MultiUsecase)
	mockUsecase := new(mocks.Usecase)
	mockRepository := new(mocks.Repository)
	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))

	handler := &handler{
		multiusecase: mockMultiUsecase,
		usecase:      mockUsecase,
		repository:   mockRepository,
		responses:    libResponse,
	}

	t.Run("error bind", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/submission-2wilen", strings.NewReader("invalid json"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Submission2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-799")
	})

	t.Run("error validate", func(t *testing.T) {
		e := echo.New()
		e.Validator = common.NewValidator()

		reqBody := struct {
			InvalidField string `json:"invalid_field"`
		}{
			InvalidField: "test",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/submission-2wilen", bytes.NewBuffer(jsonBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = handler.Submission2Wilen(c)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Contains(t, rec.Body.String(), "WLN-800")
	})
}
