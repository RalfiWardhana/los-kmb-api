package usecase

import (
	"context"
	"encoding/json"
	"errors"
	mocksCache "los-kmb-api/domain/cache/mocks"
	"los-kmb-api/domain/cms/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"

	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetReasonPrescreening(t *testing.T) {
	testcases := []struct {
		name     string
		row      int
		req      request.ReqReasonPrescreening
		data     []entity.ReasonMessage
		errGet   error
		errFinal error
	}{
		{
			name: "test error get reason",
			req: request.ReqReasonPrescreening{
				ReasonID: "11",
			},
			errGet:   errors.New("upstream_service_error - Get Reason Prescreening ID"),
			errFinal: errors.New("upstream_service_error - Get Reason Prescreening ID"),
		},
		{
			name: "test success get reason",
			req: request.ReqReasonPrescreening{
				ReasonID: "99",
			},
			data: []entity.ReasonMessage{
				{
					ReasonID:      "11",
					Code:          "12",
					ReasonMessage: "Akte Jual Beli Tidak Sesuai",
				},
			},
			row: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			var cache *bigcache.BigCache

			mockRepository.On("GetReasonPrescreening", mock.Anything, mock.Anything).Return(tc.data, tc.row, tc.errGet)

			usecase := NewUsecase(mockRepository, mockHttpClient, cache)

			result, _, err := usecase.GetReasonPrescreening(context.Background(), tc.req, mock.Anything)
			require.Equal(t, tc.data, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestGetCancelReason(t *testing.T) {
	testcases := []struct {
		name     string
		row      int
		data     []entity.CancelReason
		errGet   error
		errFinal error
	}{
		{
			name:     "test error get reason",
			errGet:   errors.New("upstream_service_error - Get Cancel Reason"),
			errFinal: errors.New("upstream_service_error - Get Cancel Reason"),
		},
		{
			name: "test success get reason",
			data: []entity.CancelReason{
				{
					ReasonID: "1",
					Show:     "1",
					Reason:   "Ganti Program Marketing",
				},
			},
			row: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			var cache *bigcache.BigCache

			mockRepository.On("GetCancelReason", mock.Anything, mock.Anything).Return(tc.data, tc.row, tc.errGet)

			usecase := NewUsecase(mockRepository, mockHttpClient, cache)

			result, _, err := usecase.GetCancelReason(context.Background(), mock.Anything)
			require.Equal(t, tc.data, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestReviewPrescreening(t *testing.T) {
	// Persiapkan objek usecase dengan mock repository
	// mockRepository := new(mocks.Repository)
	// mockHttpClient := new(httpclient.MockHttpClient)
	// var cache *bigcache.BigCache
	// usecase := NewUsecase(mockRepository, mockHttpClient, cache)

	var (
		errSave      error
		reason       string
		prescreening entity.TrxPrescreening
		trxDetail    entity.TrxDetail
		trxStatus    entity.TrxStatus
		data         response.ReviewPrescreening
	)

	// Kasus uji 1: Status UNPROCESS dan SourceDecision PRESCREENING, Decision APPROVE
	t.Run("ValidReviewCaseApprove", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			Reason:     "Valid reason",
			DecisionBy: "User123",
		}

		reason = string(req.Reason)

		decisionMapping := map[string]struct {
			Code           int
			StatusProcess  string
			Activity       string
			Decision       string
			DecisionDetail string
			DecisionStatus string
			ActivityStatus string
			NextStep       interface{}
			SourceDecision string
		}{
			constant.DECISION_REJECT: {
				Code:           constant.CODE_REJECT_PRESCREENING,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				DecisionStatus: constant.DB_DECISION_REJECT,
				DecisionDetail: constant.DB_DECISION_REJECT,
				ActivityStatus: constant.ACTIVITY_STOP,
				SourceDecision: constant.PRESCREENING,
			},
			constant.DECISION_APPROVE: {
				Code:           constant.CODE_PASS_PRESCREENING,
				StatusProcess:  constant.STATUS_ONPROCESS,
				Activity:       constant.ACTIVITY_PROCESS,
				Decision:       constant.DB_DECISION_APR,
				DecisionStatus: constant.DB_DECISION_CREDIT_PROCESS,
				DecisionDetail: constant.DB_DECISION_PASS,
				ActivityStatus: constant.ACTIVITY_UNPROCESS,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			},
		}

		decisionInfo, ok := decisionMapping[req.Decision]
		if !ok {
			err := errors.New(constant.ERROR_UPSTREAM + " - Decision tidak valid")
			t.Errorf("Expected error 'Decision tidak valid', but got: %v", err)

		}

		data.ProspectID = "TST-DEV"
		data.Code = constant.CODE_PASS_PRESCREENING
		data.Decision = constant.DB_DECISION_APR
		data.Reason = "Valid reason"

		info, _ := json.Marshal(data)

		prescreening = entity.TrxPrescreening{
			ProspectID: req.ProspectID,
			Decision:   decisionInfo.Decision,
			Reason:     reason,
			CreatedBy:  req.DecisionBy,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  decisionInfo.StatusProcess,
			Activity:       decisionInfo.Activity,
			Decision:       decisionInfo.DecisionDetail,
			RuleCode:       decisionInfo.Code,
			SourceDecision: constant.PRESCREENING,
			NextStep:       decisionInfo.NextStep,
			Info:           string(info),
			CreatedBy:      req.DecisionBy,
		}
		if req.Decision == constant.DECISION_REJECT {
			trxStatus.RuleCode = decisionInfo.Code
			trxStatus.Reason = reason
		}

		trxStatus.ProspectID = req.ProspectID
		trxStatus.StatusProcess = decisionInfo.StatusProcess
		trxStatus.Activity = decisionInfo.ActivityStatus
		trxStatus.Decision = decisionInfo.DecisionStatus
		trxStatus.SourceDecision = decisionInfo.SourceDecision

		status := entity.TrxStatus{
			ProspectID:     "TST-DEV",
			Activity:       constant.ACTIVITY_UNPROCESS,
			SourceDecision: constant.PRESCREENING,
		}
		mockRepository.On("GetTrxStatus", req.ProspectID).Return(status, errSave).Once()
		mockRepository.On("SavePrescreening", prescreening, trxDetail, trxStatus).Return(errSave).Once()

		result, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.DB_DECISION_APR, result.Decision)
	})

	t.Run("ValidReviewCaseReject", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_REJECT,
			Reason:     "Reject reason",
			DecisionBy: "User123",
		}

		reason = string(req.Reason)

		decisionMapping := map[string]struct {
			Code           int
			StatusProcess  string
			Activity       string
			Decision       string
			DecisionDetail string
			DecisionStatus string
			ActivityStatus string
			NextStep       interface{}
			SourceDecision string
		}{
			constant.DECISION_REJECT: {
				Code:           constant.CODE_REJECT_PRESCREENING,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				DecisionStatus: constant.DB_DECISION_REJECT,
				DecisionDetail: constant.DB_DECISION_REJECT,
				ActivityStatus: constant.ACTIVITY_STOP,
				SourceDecision: constant.PRESCREENING,
			},
			constant.DECISION_APPROVE: {
				Code:           constant.CODE_PASS_PRESCREENING,
				StatusProcess:  constant.STATUS_ONPROCESS,
				Activity:       constant.ACTIVITY_PROCESS,
				Decision:       constant.DB_DECISION_APR,
				DecisionStatus: constant.DB_DECISION_CREDIT_PROCESS,
				DecisionDetail: constant.DB_DECISION_PASS,
				ActivityStatus: constant.ACTIVITY_UNPROCESS,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			},
		}

		decisionInfo, ok := decisionMapping[req.Decision]
		if !ok {
			err := errors.New(constant.ERROR_UPSTREAM + " - Decision tidak valid")
			t.Errorf("Expected error 'Decision tidak valid', but got: %v", err)

		}

		data.ProspectID = "TST-DEV"
		data.Code = constant.CODE_REJECT_PRESCREENING
		data.Decision = constant.DB_DECISION_REJECT
		data.Reason = "Reject reason"

		info, _ := json.Marshal(data)

		prescreening = entity.TrxPrescreening{
			ProspectID: req.ProspectID,
			Decision:   decisionInfo.Decision,
			Reason:     reason,
			CreatedBy:  req.DecisionBy,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  decisionInfo.StatusProcess,
			Activity:       decisionInfo.Activity,
			Decision:       decisionInfo.DecisionDetail,
			RuleCode:       decisionInfo.Code,
			SourceDecision: constant.PRESCREENING,
			NextStep:       decisionInfo.NextStep,
			Info:           string(info),
			CreatedBy:      req.DecisionBy,
		}
		if req.Decision == constant.DECISION_REJECT {
			trxStatus.RuleCode = decisionInfo.Code
			trxStatus.Reason = reason
		}

		trxStatus.ProspectID = req.ProspectID
		trxStatus.StatusProcess = decisionInfo.StatusProcess
		trxStatus.Activity = decisionInfo.ActivityStatus
		trxStatus.Decision = decisionInfo.DecisionStatus
		trxStatus.SourceDecision = decisionInfo.SourceDecision

		status := entity.TrxStatus{
			ProspectID:     "TST-DEV",
			Activity:       constant.ACTIVITY_UNPROCESS,
			SourceDecision: constant.PRESCREENING,
		}
		mockRepository.On("GetTrxStatus", req.ProspectID).Return(status, errSave).Once()
		mockRepository.On("SavePrescreening", prescreening, trxDetail, trxStatus).Return(errSave).Once()

		result, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.DB_DECISION_REJECT, result.Decision)
	})

	// Kasus uji 2: Status UNPROCESS dan SourceDecision PRESCREENING, Decision tidak valid
	t.Run("InvalidDecisionCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   "InvalidDecision",
			Reason:     "Invalid reason",
			DecisionBy: "User123",
		}
		status := entity.TrxStatus{
			ProspectID:     "TST-DEV",
			Activity:       constant.ACTIVITY_UNPROCESS,
			SourceDecision: constant.PRESCREENING,
		}
		mockRepository.On("GetTrxStatus", req.ProspectID).Return(status, errSave).Once()
		mockRepository.On("SavePrescreening", prescreening, trxDetail, trxStatus).Return(errSave).Once()

		_, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		assert.Error(t, err)
	})

	t.Run("InvalidStatusCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			Reason:     "Valid reason",
			DecisionBy: "User123",
		}

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Status order tidak dalam prescreening")

		mockRepository.On("GetTrxStatus", req.ProspectID).Return(entity.TrxStatus{
			Activity:       constant.ACTIVITY_PROCESS,
			SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
		}, errSave).Once()

		_, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

	t.Run("ErrorStatusCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			Reason:     "Valid reason",
			DecisionBy: "User123",
		}
		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Get status order error")

		mockRepository.On("GetTrxStatus", req.ProspectID).Return(entity.TrxStatus{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

	t.Run("ErrorSavePrescreening", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			Reason:     "Valid reason",
			DecisionBy: "User123",
		}
		mockRepository.On("GetTrxStatus", mock.Anything).Return(entity.TrxStatus{
			Activity:       constant.ACTIVITY_UNPROCESS,
			SourceDecision: constant.PRESCREENING,
		}, nil).Once()

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Save prescreening error")

		mockRepository.On("SavePrescreening", mock.Anything, mock.Anything, mock.Anything).Return(errors.New(constant.ERROR_UPSTREAM + " - Save prescreening error")).Once()

		_, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

}

func TestReviewPrescreeningInvalidStatus(t *testing.T) {
	// Persiapkan objek usecase dengan mock repository
	mockRepository := new(mocks.Repository)
	mockHttpClient := new(httpclient.MockHttpClient)
	var cache *bigcache.BigCache
	usecase := NewUsecase(mockRepository, mockHttpClient, cache)

	var (
		errSave error
	)

	// Kasus uji 3: Status tidak UNPROCESS atau SourceDecision bukan PRESCREENING
	t.Run("InvalidStatusCase", func(t *testing.T) {
		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			Reason:     "Valid reason",
			DecisionBy: "User123",
		}
		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Status order tidak dalam prescreening")

		mockRepository.On("GetTrxStatus", req.ProspectID).Return(entity.TrxStatus{
			Activity:       constant.ACTIVITY_PROCESS,
			SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
		}, errSave)

		_, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

}

func Test_usecase_GetInquiryPrescreening(t *testing.T) {
	ctx := context.Background()
	// var errSave error
	t.Run("EmptyGetSpIndustryTypeMaster", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		mockRepository.On("GetInquiryPrescreening", mock.Anything, mock.Anything).Return([]entity.InquiryPrescreening{}, 1, errors.New(constant.RECORD_NOT_FOUND)).Once()

		// Create a test request and pagination data.
		req := request.ReqInquiryPrescreening{}
		pagination := interface{}(nil)

		mocksCache.On("Get", mock.Anything).Return(nil, errors.New(constant.RECORD_NOT_FOUND))

		mockRepository.On("GetSpIndustryTypeMaster").Return([]entity.SpIndustryTypeMaster{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

		mocksCache.On("Set", mock.Anything, mock.Anything).Return(mock.Anything, mock.Anything)

		_, _, err := usecase.GetInquiryPrescreening(ctx, req, pagination)
		assert.Error(t, err)
	})

	t.Run("EmptyGetCustomerPhoto", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqInquiryPrescreening{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquiryPrescreening", req, 1).Return([]entity.InquiryPrescreening{
			{
				CmoRecommendation: 1,
			},
		}, 1, nil).Once()

		errPhoto := errors.New(constant.RECORD_NOT_FOUND)

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetInquiryPrescreening(ctx, req, 1)

		assert.Error(t, errPhoto, err)
	})

	t.Run("EmptyGetSurveyorData", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqInquiryPrescreening{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquiryPrescreening", req, 1).Return([]entity.InquiryPrescreening{
			{
				CmoRecommendation: 1,
			},
		}, 1, nil).Once()

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, nil).Once()

		surveyor := []entity.TrxSurveyor{}
		errSurveyor := errors.New(constant.RECORD_NOT_FOUND)

		mockRepository.On("GetSurveyorData", mock.Anything).Return(surveyor, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetInquiryPrescreening(ctx, req, 1)

		assert.Error(t, errSurveyor, err)
	})

	t.Run("EmptyGetInquiryPrescreening", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		req := request.ReqInquiryPrescreening{}

		mocksCache.On("Get", "GetSpIndustryTypeMaster").Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquiryPrescreening", req, 1).Return([]entity.InquiryPrescreening{}, 1, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetInquiryPrescreening(ctx, req, 1)

		// Verifikasi bahwa error yang diharapkan terjadi
		assert.Error(t, err)
	})

}

func TestGetAkkk(t *testing.T) {
	testcases := []struct {
		name       string
		getAkkk    entity.Akkk
		getAkkkErr error
		industry   []entity.SpIndustryTypeMaster
		result     entity.Akkk
		err        error
	}{
		{
			name: "get akkk",
			getAkkk: entity.Akkk{
				IndustryTypeID:          "SE_e8521",
				MonthlyFixedIncome:      float64(5000000),
				MonthlyVariableIncome:   float64(600000),
				SpouseIncome:            float64(4000000),
				Plafond:                 float64(21000000),
				BakiDebet:               float64(300000),
				BakiDebetTerburuk:       float64(600000),
				SpousePlafond:           float64(200000),
				SpouseBakiDebet:         float64(100000),
				SpouseBakiDebetTerburuk: float64(333000),
				TotalAgreementAktif:     float64(2),
				MaxOVDAgreementAktif:    float64(4),
				LastMaxOVDAgreement:     float64(3),
				LatestInstallment:       float64(304000),
				NTFAkumulasi:            float64(23000000),
				TotalInstallment:        float64(4000000),
				TotalIncome:             float64(7899800),
				TotalDSR:                float64(7.8),
				EkycSimiliarity:         float64(9.7),
			},
			industry: []entity.SpIndustryTypeMaster{
				{
					IndustryTypeID: "SE_e8521",
					Description:    "Pendidikan Menengah Umum/Madrasah Aliyah Pemerintah",
				},
			},
			result: entity.Akkk{
				IndustryType:            "Pendidikan Menengah Umum/Madrasah Aliyah Pemerintah",
				IndustryTypeID:          "SE_e8521",
				MonthlyFixedIncome:      float64(5000000),
				MonthlyVariableIncome:   float64(600000),
				SpouseIncome:            float64(4000000),
				Plafond:                 float64(21000000),
				BakiDebet:               float64(300000),
				BakiDebetTerburuk:       float64(600000),
				SpousePlafond:           float64(200000),
				SpouseBakiDebet:         float64(100000),
				SpouseBakiDebetTerburuk: float64(333000),
				TotalAgreementAktif:     float64(2),
				MaxOVDAgreementAktif:    float64(4),
				LastMaxOVDAgreement:     float64(3),
				LatestInstallment:       float64(304000),
				NTFAkumulasi:            float64(23000000),
				TotalInstallment:        float64(4000000),
				TotalIncome:             float64(7899800),
				TotalDSR:                float64(7.8),
				EkycSimiliarity:         float64(9.7),
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mocksCache := &mocksCache.Repository{}

			mocksCache.On("Get", mock.Anything).Return(nil, errors.New("data not found")).Once()
			mockRepository.On("GetAkkk", "TEST001").Return(tc.getAkkk, tc.getAkkkErr).Once()
			mockRepository.On("GetSpIndustryTypeMaster").Return(tc.industry, nil).Once()
			mocksCache.On("Set", mock.Anything, mock.Anything).Return(nil)
			mocksCache.On("Get", mock.Anything).Return([]byte("Pendidikan Menengah Umum/Madrasah Aliyah Pemerintah"), nil).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

			akkk, err := usecase.GetAkkk("TEST001")
			require.Equal(t, tc.result, akkk)
			require.Equal(t, tc.err, err)
		})
	}
	return
}

func TestGetInquiryPrescreening(t *testing.T) {

	testcases := []struct {
		name     string
		row      int
		req      request.ReqInquiryPrescreening
		data     []entity.InquiryData
		inquiry  []entity.InquiryPrescreening
		photos   []entity.DataPhoto
		surveyor []entity.TrxSurveyor
		errGet   error
		errFinal error
	}{
		{
			name: "test success get reason",
			req:  request.ReqInquiryPrescreening{},
			// data: responseData,
			photos:   []entity.DataPhoto{},
			surveyor: []entity.TrxSurveyor{},
			inquiry: []entity.InquiryPrescreening{
				{
					CmoRecommendation: 0,
					Activity:          constant.ACTIVITY_UNPROCESS,
					SourceDecision:    constant.PRESCREENING,
					Decision:          constant.DB_DECISION_REJECT,
				},
			},
			row: 1,
		},
		{
			name: "test success get reason",
			req:  request.ReqInquiryPrescreening{},
			// data: responseData,
			photos: []entity.DataPhoto{
				{
					PhotoID: "KTP",
					Url:     "jsdhfigshhjgh",
				},
			},
			surveyor: []entity.TrxSurveyor{
				{
					SurveyorName: "ujang",
				},
			},
			inquiry: []entity.InquiryPrescreening{
				{
					CmoRecommendation: 1,
					Decision:          constant.DB_DECISION_APR,
				},
			},
			row: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mocksCache := &mocksCache.Repository{}
			usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)
			mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)
			mockRepository.On("GetInquiryPrescreening", tc.req, 1).Return(tc.inquiry, 1, tc.errGet).Once()
			mockRepository.On("GetSpIndustryTypeMaster").Return(tc.inquiry, 1, tc.errGet).Once()
			mockRepository.On("GetCustomerPhoto", mock.Anything).Return(tc.photos, nil).Once()
			mockRepository.On("GetSurveyorData", mock.Anything).Return(tc.surveyor, nil).Once()

			result, _, _ := usecase.GetInquiryPrescreening(context.Background(), tc.req, mock.Anything)
			assert.Equal(t, 1, len(result))
		})
	}
}

func TestGetInquiryCa(t *testing.T) {
	ctx := context.Background()
	// var errSave error
	t.Run("EmptyGetSpIndustryTypeMaster", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		mockRepository.On("GetInquiryCa", mock.Anything, mock.Anything).Return([]entity.InquiryCa{}, 1, errors.New(constant.RECORD_NOT_FOUND)).Once()

		// Create a test request and pagination data.
		req := request.ReqInquiryCa{}
		pagination := interface{}(nil)

		mocksCache.On("Get", mock.Anything).Return(nil, errors.New(constant.RECORD_NOT_FOUND))

		mockRepository.On("GetSpIndustryTypeMaster").Return([]entity.SpIndustryTypeMaster{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

		mocksCache.On("Set", mock.Anything, mock.Anything).Return(mock.Anything, mock.Anything)

		_, _, err := usecase.GetInquiryCa(ctx, req, pagination)
		assert.Error(t, err)
	})

	t.Run("EmptyGetCustomerPhoto", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqInquiryCa{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquiryCa", req, 1).Return([]entity.InquiryCa{
			{
				CaDecision: constant.DB_DECISION_APR,
			},
		}, 1, nil).Once()

		errPhoto := errors.New(constant.RECORD_NOT_FOUND)

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetInquiryCa(ctx, req, 1)

		assert.Error(t, errPhoto, err)
	})

	t.Run("EmptyGetSurveyorData", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqInquiryCa{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquiryCa", req, 1).Return([]entity.InquiryCa{
			{
				CaDecision: constant.DB_DECISION_APR,
			},
		}, 1, nil).Once()

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, nil).Once()

		surveyor := []entity.TrxSurveyor{}
		errSurveyor := errors.New(constant.RECORD_NOT_FOUND)

		mockRepository.On("GetSurveyorData", mock.Anything).Return(surveyor, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetInquiryCa(ctx, req, 1)

		assert.Error(t, errSurveyor, err)
	})

	t.Run("EmptyGetHistoryApproval", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqInquiryCa{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquiryCa", req, 1).Return([]entity.InquiryCa{
			{
				CaDecision: constant.DB_DECISION_APR,
			},
		}, 1, nil).Once()

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, nil).Once()

		mockRepository.On("GetSurveyorData", mock.Anything).Return([]entity.TrxSurveyor{}, nil).Once()

		history := []entity.HistoryApproval{}
		errData := errors.New(constant.RECORD_NOT_FOUND)

		mockRepository.On("GetHistoryApproval", mock.Anything).Return(history, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetInquiryCa(ctx, req, 1)

		assert.Error(t, errData, err)
	})

	t.Run("EmptyGetInquiryCa", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		req := request.ReqInquiryCa{}

		mocksCache.On("Get", "GetSpIndustryTypeMaster").Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquiryCa", req, 1).Return([]entity.InquiryCa{}, 1, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetInquiryCa(ctx, req, 1)

		// Verifikasi bahwa error yang diharapkan terjadi
		assert.Error(t, err)
	})

}

func TestUsecaseGetInquiryCa(t *testing.T) {

	testcases := []struct {
		name           string
		row            int
		req            request.ReqInquiryCa
		data           []entity.InquiryDataCa
		inquiry        []entity.InquiryCa
		photos         []entity.DataPhoto
		surveyor       []entity.TrxSurveyor
		histories      []entity.HistoryApproval
		internalRecord []entity.TrxInternalRecord
		errGet         error
		errFinal       error
	}{
		{
			name:           "test success get ",
			req:            request.ReqInquiryCa{},
			photos:         []entity.DataPhoto{},
			surveyor:       []entity.TrxSurveyor{},
			histories:      []entity.HistoryApproval{},
			internalRecord: []entity.TrxInternalRecord{},
			inquiry: []entity.InquiryCa{
				{
					Activity:       constant.ACTIVITY_UNPROCESS,
					SourceDecision: constant.PRESCREENING,
				},
			},
			row: 1,
		},
		{
			name: "test success get data approve",
			req:  request.ReqInquiryCa{},
			photos: []entity.DataPhoto{
				{
					PhotoID: "KTP",
					Url:     "jsdhfigshhjgh",
				},
			},
			surveyor: []entity.TrxSurveyor{
				{
					SurveyorName: "ujang",
				},
			},
			histories: []entity.HistoryApproval{
				{
					Decision:       "REJ",
					NeedEscalation: 1,
				},
			},
			internalRecord: []entity.TrxInternalRecord{
				{
					ProspectID: "PPID",
				},
			},
			inquiry: []entity.InquiryCa{
				{
					ShowAction:     true,
					StatusDecision: constant.DB_DECISION_APR,
				},
			},
			row: 1,
		},
		{
			name: "test success get data reject",
			req:  request.ReqInquiryCa{},
			photos: []entity.DataPhoto{
				{
					PhotoID: "KTP",
					Url:     "jsdhfigshhjgh",
				},
			},
			surveyor: []entity.TrxSurveyor{
				{
					SurveyorName: "ujang",
				},
			},
			histories: []entity.HistoryApproval{
				{
					Decision:       "REJ",
					NeedEscalation: 1,
				},
			},
			internalRecord: []entity.TrxInternalRecord{
				{
					ProspectID: "PPID",
				},
			},
			inquiry: []entity.InquiryCa{
				{
					ShowAction:     true,
					StatusDecision: constant.DB_DECISION_REJECT,
				},
			},
			row: 1,
		},
		{
			name: "test success get data cancel",
			req:  request.ReqInquiryCa{},
			photos: []entity.DataPhoto{
				{
					PhotoID: "KTP",
					Url:     "jsdhfigshhjgh",
				},
			},
			surveyor: []entity.TrxSurveyor{
				{
					SurveyorName: "ujang",
				},
			},
			histories: []entity.HistoryApproval{
				{
					Decision:       "REJ",
					NeedEscalation: 1,
				},
			},
			internalRecord: []entity.TrxInternalRecord{
				{
					ProspectID: "PPID",
				},
			},
			inquiry: []entity.InquiryCa{
				{
					ShowAction:     true,
					StatusDecision: constant.DB_DECISION_CANCEL,
				},
			},
			row: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mocksCache := &mocksCache.Repository{}
			usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)
			mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)
			mockRepository.On("GetInquiryCa", tc.req, 1).Return(tc.inquiry, 1, tc.errGet).Once()
			mockRepository.On("GetSpIndustryTypeMaster").Return(tc.inquiry, 1, tc.errGet).Once()
			mockRepository.On("GetCustomerPhoto", mock.Anything).Return(tc.photos, nil).Once()
			mockRepository.On("GetSurveyorData", mock.Anything).Return(tc.surveyor, nil).Once()
			mockRepository.On("GetHistoryApproval", mock.Anything).Return(tc.histories, nil).Once()
			mockRepository.On("GetInternalRecord", mock.Anything).Return(tc.internalRecord, nil).Once()

			result, _, _ := usecase.GetInquiryCa(context.Background(), tc.req, mock.Anything)
			assert.Equal(t, 1, len(result))
		})
	}
}

func TestSaveAsDraft(t *testing.T) {

	var (
		errSave  error
		decision string
	)

	t.Run("ValidSaveApprove", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqSaveAsDraft{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			SlikResult: "Lancar",
			Note:       "oke",
			CreatedBy:  "247653786",
			DecisionBy: "User123",
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = constant.DB_DECISION_REJECT
		case constant.DECISION_APPROVE:
			decision = constant.DB_DECISION_APR
		}

		trxDraft := entity.TrxDraftCaDecision{
			ProspectID: req.ProspectID,
			Decision:   decision,
			SlikResult: req.SlikResult,
			Note:       req.Note,
			CreatedBy:  req.CreatedBy,
			DecisionBy: req.DecisionBy,
		}

		mockRepository.On("SaveDraftData", trxDraft).Return(errSave).Once()

		result, err := usecase.SaveAsDraft(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.DECISION_APPROVE, result.Decision)
	})

	t.Run("ValidSaveReject", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqSaveAsDraft{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_REJECT,
			SlikResult: "Lancar",
			Note:       "oke",
			CreatedBy:  "247653786",
			DecisionBy: "User123",
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = constant.DB_DECISION_REJECT
		case constant.DECISION_APPROVE:
			decision = constant.DB_DECISION_APR
		}

		trxDraft := entity.TrxDraftCaDecision{
			ProspectID: req.ProspectID,
			Decision:   decision,
			SlikResult: req.SlikResult,
			Note:       req.Note,
			CreatedBy:  req.CreatedBy,
			DecisionBy: req.DecisionBy,
		}

		mockRepository.On("SaveDraftData", trxDraft).Return(errSave).Once()

		result, err := usecase.SaveAsDraft(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.DECISION_REJECT, result.Decision)
	})

	t.Run("ErrorSaveDraft", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		req := request.ReqSaveAsDraft{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			SlikResult: "Lancar",
			Note:       "oke",
			CreatedBy:  "247653786",
			DecisionBy: "User123",
		}

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Save prescreening error")

		mockRepository.On("SaveDraftData", mock.Anything).Return(errors.New(constant.ERROR_UPSTREAM + " - Save prescreening error")).Once()

		_, err := usecase.SaveAsDraft(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})
}

func TestReturnOrder(t *testing.T) {

	var (
		errSave error
	)

	req := request.ReqReturnOrder{
		ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
		CreatedBy:  "247653786",
		DecisionBy: "User123",
	}

	trxStatus := entity.TrxStatus{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		SourceDecision: constant.PRESCREENING,
		Reason:         constant.REASON_RETURN_ORDER,
	}

	trxDetail := entity.TrxDetail{
		ProspectID:     req.ProspectID,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		SourceDecision: constant.CMO_AGENT,
		NextStep:       constant.PRESCREENING,
		Info:           constant.REASON_RETURN_ORDER,
		CreatedBy:      constant.SYSTEM_CREATED,
	}

	t.Run("ValidSaveApprove", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		mockRepository.On("ProcessReturnOrder", mock.Anything, trxStatus, trxDetail).Return(errSave).Once()

		result, err := usecase.ReturnOrder(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.RETURN_STATUS_SUCCESS, result.Status)
	})

	t.Run("ErrorProcessReturnOrder", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Process Return Order error")

		mockRepository.On("ProcessReturnOrder", mock.Anything, trxStatus, trxDetail).Return(errors.New(constant.ERROR_UPSTREAM + " - Process Return Order error")).Once()

		_, err := usecase.ReturnOrder(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})
}

func TestCancelOrder(t *testing.T) {
	var (
		errSave       error
		isCancel      bool
		trxStatus     entity.TrxStatus
		trxDetail     entity.TrxDetail
		trxCaDecision entity.TrxCaDecision
		// data          response.CancelResponse
	)

	status := entity.TrxStatus{
		ProspectID:     "TST-DEV",
		Activity:       constant.ACTIVITY_UNPROCESS,
		SourceDecision: constant.DB_DECISION_APR,
	}

	req := request.ReqCancelOrder{
		ProspectID:   "TST-DEV", // Ganti dengan ID yang sesuai
		CancelReason: "reason",
		CreatedBy:    "agsa6srt",
		DecisionBy:   "User123",
	}

	isCancel = true

	t.Run("ValidCancelCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		trxCaDecision = entity.TrxCaDecision{
			ProspectID: req.ProspectID,
			Decision:   constant.DB_DECISION_CANCEL,
			Note:       req.CancelReason,
			CreatedBy:  req.CreatedBy,
			DecisionBy: req.DecisionBy,
		}

		trxStatus = entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_CANCEL,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
			Reason:         req.CancelReason,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  constant.STATUS_FINAL,
			Activity:       constant.ACTIVITY_STOP,
			Decision:       constant.DB_DECISION_CANCEL,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
			Info:           req.CancelReason,
			CreatedBy:      req.CreatedBy,
		}

		mockRepository.On("GetTrxStatus", req.ProspectID).Return(status, errSave).Once()

		mockRepository.On("ProcessTransaction", isCancel, trxCaDecision, trxStatus, trxDetail).Return(errSave).Once()

		result, err := usecase.CancelOrder(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.CANCEL_STATUS_SUCCESS, result.Status)
	})

	t.Run("ErrorGetTrxStatusCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Get status order error")

		mockRepository.On("GetTrxStatus", req.ProspectID).Return(entity.TrxStatus{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, err := usecase.CancelOrder(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

	t.Run("ErrorProcessTransaction", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		mockRepository.On("GetTrxStatus", mock.Anything).Return(entity.TrxStatus{
			Activity:       constant.ACTIVITY_UNPROCESS,
			SourceDecision: constant.PRESCREENING,
		}, nil).Once()

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Process Cancel Order error")

		mockRepository.On("ProcessTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(constant.ERROR_UPSTREAM + " - Process Cancel Order error")).Once()

		_, err := usecase.CancelOrder(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

	t.Run("InvalidStatusCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Status order tidak dapat dicancel")

		mockRepository.On("GetTrxStatus", req.ProspectID).Return(entity.TrxStatus{
			Activity:       constant.ACTIVITY_STOP,
			SourceDecision: constant.DB_DECISION_REJECT,
		}, errSave).Once()

		_, err := usecase.CancelOrder(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})
}

func TestSubmitDecision(t *testing.T) {
	var (
		errSave       error
		isCancel      bool
		trxStatus     entity.TrxStatus
		trxDetail     entity.TrxDetail
		trxCaDecision entity.TrxCaDecision
		limit         entity.MappingLimitApprovalScheme
		// data          response.CAResponse
	)

	t.Run("ValidSubmitCaseApprove", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqSubmitDecision{
			ProspectID:   "TST-DEV",
			NTFAkumulasi: 123456.55,
			Decision:     constant.DECISION_APPROVE,
			SlikResult:   "lancar",
			Note:         "noted",
			CreatedBy:    "agsa6srt",
			DecisionBy:   "User123",
		}

		decisionMapping := map[string]struct {
			Code           int
			StatusProcess  string
			Activity       string
			Decision       string
			DecisionDetail string
			DecisionStatus string
			ActivityStatus string
			NextStep       interface{}
			SourceDecision string
		}{
			constant.DECISION_REJECT: {
				Code:           constant.CODE_REJECT_PRESCREENING,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				DecisionStatus: constant.DB_DECISION_REJECT,
				DecisionDetail: constant.DB_DECISION_REJECT,
				ActivityStatus: constant.ACTIVITY_STOP,
				SourceDecision: constant.PRESCREENING,
			},
			constant.DECISION_APPROVE: {
				Code:           constant.CODE_PASS_PRESCREENING,
				StatusProcess:  constant.STATUS_ONPROCESS,
				Activity:       constant.ACTIVITY_PROCESS,
				Decision:       constant.DB_DECISION_APR,
				DecisionStatus: constant.DB_DECISION_CREDIT_PROCESS,
				DecisionDetail: constant.DB_DECISION_PASS,
				ActivityStatus: constant.ACTIVITY_UNPROCESS,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			},
		}

		decisionInfo, ok := decisionMapping[req.Decision]
		if !ok {
			err := errors.New(constant.ERROR_UPSTREAM + " - Decision tidak valid")
			t.Errorf("Expected error 'Decision tidak valid', but got: %v", err)

		}

		trxCaDecision = entity.TrxCaDecision{
			ProspectID:    req.ProspectID,
			Decision:      decisionInfo.Decision,
			SlikResult:    req.SlikResult,
			Note:          req.Note,
			CreatedBy:     req.CreatedBy,
			DecisionBy:    req.DecisionBy,
			FinalApproval: limit.Alias,
		}

		trxStatus = entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  decisionInfo.StatusProcess,
			Activity:       decisionInfo.ActivityStatus,
			Decision:       decisionInfo.DecisionStatus,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
			Reason:         req.SlikResult,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  decisionInfo.StatusProcess,
			Activity:       decisionInfo.Activity,
			Decision:       decisionInfo.DecisionDetail,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
			NextStep:       constant.DB_DECISION_BRANCH_MANAGER,
			Info:           req.SlikResult,
			CreatedBy:      req.CreatedBy,
		}

		status := entity.TrxStatus{
			ProspectID: "TST-DEV",
			Activity:   constant.ACTIVITY_UNPROCESS,
			Decision:   constant.DB_DECISION_CREDIT_PROCESS,
		}
		mockRepository.On("GetTrxStatus", req.ProspectID).Return(status, errSave).Once()
		mockRepository.On("GetLimitApproval", req.NTFAkumulasi).Return(limit, errSave).Once()
		mockRepository.On("ProcessTransaction", isCancel, trxCaDecision, trxStatus, trxDetail).Return(errSave).Once()

		result, err := usecase.SubmitDecision(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.DECISION_APPROVE, result.Decision)
	})
	t.Run("ValidSubmitCaseReject", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqSubmitDecision{
			ProspectID:   "TST-DEV",
			NTFAkumulasi: 123456.55,
			Decision:     constant.DECISION_REJECT,
			SlikResult:   "tidak lancar",
			Note:         "noted",
			CreatedBy:    "agsa6srt",
			DecisionBy:   "User123",
		}

		decisionMapping := map[string]struct {
			Code           int
			StatusProcess  string
			Activity       string
			Decision       string
			DecisionDetail string
			DecisionStatus string
			ActivityStatus string
			NextStep       interface{}
			SourceDecision string
		}{
			constant.DECISION_REJECT: {
				Code:           constant.CODE_REJECT_PRESCREENING,
				StatusProcess:  constant.STATUS_FINAL,
				Activity:       constant.ACTIVITY_STOP,
				Decision:       constant.DB_DECISION_REJECT,
				DecisionStatus: constant.DB_DECISION_REJECT,
				DecisionDetail: constant.DB_DECISION_REJECT,
				ActivityStatus: constant.ACTIVITY_STOP,
				SourceDecision: constant.PRESCREENING,
			},
			constant.DECISION_APPROVE: {
				Code:           constant.CODE_PASS_PRESCREENING,
				StatusProcess:  constant.STATUS_ONPROCESS,
				Activity:       constant.ACTIVITY_PROCESS,
				Decision:       constant.DB_DECISION_APR,
				DecisionStatus: constant.DB_DECISION_CREDIT_PROCESS,
				DecisionDetail: constant.DB_DECISION_PASS,
				ActivityStatus: constant.ACTIVITY_UNPROCESS,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				NextStep:       constant.SOURCE_DECISION_DUPCHECK,
			},
		}

		decisionInfo, ok := decisionMapping[req.Decision]
		if !ok {
			err := errors.New(constant.ERROR_UPSTREAM + " - Decision tidak valid")
			t.Errorf("Expected error 'Decision tidak valid', but got: %v", err)

		}

		trxCaDecision = entity.TrxCaDecision{
			ProspectID:    req.ProspectID,
			Decision:      decisionInfo.Decision,
			SlikResult:    req.SlikResult,
			Note:          req.Note,
			CreatedBy:     req.CreatedBy,
			DecisionBy:    req.DecisionBy,
			FinalApproval: limit.Alias,
		}

		trxStatus = entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  decisionInfo.StatusProcess,
			Activity:       decisionInfo.ActivityStatus,
			Decision:       decisionInfo.DecisionStatus,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
			Reason:         req.SlikResult,
		}

		trxDetail = entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  decisionInfo.StatusProcess,
			Activity:       decisionInfo.Activity,
			Decision:       decisionInfo.DecisionDetail,
			RuleCode:       constant.CODE_CREDIT_COMMITTEE,
			SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
			NextStep:       constant.DB_DECISION_BRANCH_MANAGER,
			Info:           req.SlikResult,
			CreatedBy:      req.CreatedBy,
		}

		status := entity.TrxStatus{
			ProspectID: "TST-DEV",
			Activity:   constant.ACTIVITY_UNPROCESS,
			Decision:   constant.DB_DECISION_CREDIT_PROCESS,
		}
		mockRepository.On("GetTrxStatus", req.ProspectID).Return(status, errSave).Once()
		mockRepository.On("GetLimitApproval", req.NTFAkumulasi).Return(limit, errSave).Once()
		mockRepository.On("ProcessTransaction", isCancel, trxCaDecision, trxStatus, trxDetail).Return(errSave).Once()

		result, err := usecase.SubmitDecision(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.DECISION_REJECT, result.Decision)
	})

	t.Run("InvalidDecisionCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)
		req := request.ReqSubmitDecision{
			ProspectID:   "TST-DEV",
			NTFAkumulasi: 123456.55,
			Decision:     "InvalidDecision",
			SlikResult:   "lancar",
			Note:         "noted",
			CreatedBy:    "agsa6srt",
			DecisionBy:   "User123",
		}
		status := entity.TrxStatus{
			ProspectID: "TST-DEV",
			Activity:   constant.ACTIVITY_UNPROCESS,
			Decision:   constant.DB_DECISION_CREDIT_PROCESS,
		}
		mockRepository.On("GetTrxStatus", req.ProspectID).Return(status, errSave).Once()
		mockRepository.On("GetLimitApproval", req.NTFAkumulasi).Return(limit, errSave).Once()
		mockRepository.On("ProcessTransaction", isCancel, trxCaDecision, trxStatus, trxDetail).Return(errSave).Once()

		_, err := usecase.SubmitDecision(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		assert.Error(t, err)
	})

	t.Run("InvalidStatusCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Status order tidak sedang dalam credit process")

		mockRepository.On("GetTrxStatus", mock.Anything).Return(entity.TrxStatus{
			Activity: constant.ACTIVITY_PROCESS,
			Decision: constant.DB_DECISION_APR,
		}, errSave).Once()

		req := request.ReqSubmitDecision{
			ProspectID:   "TST-DEV",
			NTFAkumulasi: 123456.55,
			Decision:     constant.DECISION_APPROVE,
			SlikResult:   "lancar",
			Note:         "noted",
			CreatedBy:    "agsa6srt",
			DecisionBy:   "User123",
		}

		_, err := usecase.SubmitDecision(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

	t.Run("ErrorGetTrxStatusCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Get status order error")

		mockRepository.On("GetTrxStatus", mock.Anything).Return(entity.TrxStatus{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

		req := request.ReqSubmitDecision{
			ProspectID:   "TST-DEV",
			NTFAkumulasi: 123456.55,
			Decision:     constant.DECISION_APPROVE,
			SlikResult:   "lancar",
			Note:         "noted",
			CreatedBy:    "agsa6srt",
			DecisionBy:   "User123",
		}
		_, err := usecase.SubmitDecision(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

	t.Run("ErrorGetLimitApprovalCase", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Get limit approval error")

		status := entity.TrxStatus{
			ProspectID: "TST-DEV",
			Activity:   constant.ACTIVITY_UNPROCESS,
			Decision:   constant.DB_DECISION_CREDIT_PROCESS,
		}
		mockRepository.On("GetTrxStatus", mock.Anything).Return(status, nil).Once()

		mockRepository.On("GetLimitApproval", mock.Anything).Return(entity.MappingLimitApprovalScheme{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

		req := request.ReqSubmitDecision{
			ProspectID:   "TST-DEV",
			NTFAkumulasi: 123456.55,
			Decision:     constant.DECISION_APPROVE,
			SlikResult:   "lancar",
			Note:         "noted",
			CreatedBy:    "agsa6srt",
			DecisionBy:   "User123",
		}
		_, err := usecase.SubmitDecision(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

	t.Run("ErrorProcessTransaction", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		var cache *bigcache.BigCache
		usecase := NewUsecase(mockRepository, mockHttpClient, cache)

		req := request.ReqSubmitDecision{
			ProspectID:   "TST-DEV",
			NTFAkumulasi: 123456.55,
			Decision:     constant.DECISION_APPROVE,
			SlikResult:   "lancar",
			Note:         "noted",
			CreatedBy:    "agsa6srt",
			DecisionBy:   "User123",
		}

		mockRepository.On("GetTrxStatus", mock.Anything).Return(entity.TrxStatus{
			Activity: constant.ACTIVITY_UNPROCESS,
			Decision: constant.DB_DECISION_CREDIT_PROCESS,
		}, nil).Once()

		mockRepository.On("GetLimitApproval", mock.Anything).Return(entity.MappingLimitApprovalScheme{
			Alias: "CBM",
		}, nil).Once()

		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Submit Decision error")

		mockRepository.On("ProcessTransaction", isCancel, mock.Anything, mock.Anything, mock.Anything).Return(errors.New(constant.ERROR_UPSTREAM + " - Submit Decision error")).Once()

		_, err := usecase.SubmitDecision(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})
}

func TestGetSearchInquiry(t *testing.T) {
	ctx := context.Background()
	// var errSave error
	t.Run("EmptyGetSpIndustryTypeMaster", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		mockRepository.On("GetInquirySearch", mock.Anything, mock.Anything).Return([]entity.InquirySearch{}, 1, errors.New(constant.RECORD_NOT_FOUND)).Once()

		// Create a test request and pagination data.
		req := request.ReqSearchInquiry{}
		pagination := interface{}(nil)

		mocksCache.On("Get", mock.Anything).Return(nil, errors.New(constant.RECORD_NOT_FOUND))

		mockRepository.On("GetSpIndustryTypeMaster").Return([]entity.SpIndustryTypeMaster{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

		mocksCache.On("Set", mock.Anything, mock.Anything).Return(mock.Anything, mock.Anything)

		_, _, err := usecase.GetSearchInquiry(ctx, req, pagination)
		assert.Error(t, err)
	})

	t.Run("EmptyGetCustomerPhoto", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqSearchInquiry{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquirySearch", req, 1).Return([]entity.InquirySearch{
			{
				FinalStatus: constant.DB_DECISION_BRANCH_MANAGER,
			},
		}, 1, nil).Once()

		errPhoto := errors.New(constant.RECORD_NOT_FOUND)

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetSearchInquiry(ctx, req, 1)

		assert.Error(t, errPhoto, err)
	})

	t.Run("EmptyGetSurveyorData", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqSearchInquiry{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquirySearch", req, 1).Return([]entity.InquirySearch{
			{
				FinalStatus: constant.DB_DECISION_BRANCH_MANAGER,
			},
		}, 1, nil).Once()

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, nil).Once()

		surveyor := []entity.TrxSurveyor{}
		errSurveyor := errors.New(constant.RECORD_NOT_FOUND)

		mockRepository.On("GetSurveyorData", mock.Anything).Return(surveyor, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetSearchInquiry(ctx, req, 1)

		assert.Error(t, errSurveyor, err)
	})

	t.Run("EmptyGetHistoryProcessData", func(t *testing.T) {
		// Create an instance of your usecase with the mock repository and cache.
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		// Create a test request and pagination data.
		req := request.ReqSearchInquiry{}

		mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquirySearch", req, 1).Return([]entity.InquirySearch{
			{
				FinalStatus: constant.DB_DECISION_BRANCH_MANAGER,
			},
		}, 1, nil).Once()

		photos := []entity.DataPhoto{
			{
				PhotoID: "KTP",
				Url:     "jsdhfigshhjgh",
			},
		}

		mockRepository.On("GetCustomerPhoto", mock.Anything).Return(photos, nil).Once()

		mockRepository.On("GetSurveyorData", mock.Anything).Return([]entity.TrxSurveyor{}, nil).Once()

		detail := []entity.TrxDetail{}
		errSurveyor := errors.New(constant.RECORD_NOT_FOUND)

		mockRepository.On("GetHistoryProcess", mock.Anything).Return(detail, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetSearchInquiry(ctx, req, 1)

		assert.Error(t, errSurveyor, err)
	})

	t.Run("EmptyGetInquiryCa", func(t *testing.T) {
		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)
		mocksCache := &mocksCache.Repository{}
		usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)

		req := request.ReqSearchInquiry{}

		mocksCache.On("Get", "GetSpIndustryTypeMaster").Return([]byte("SuccessRetrieve"), nil)

		mockRepository.On("GetSpIndustryTypeMaster", mock.Anything).Return([]entity.SpIndustryTypeMaster{
			{
				IndustryTypeID: "ASa",
				Description:    "asas",
				IsActive:       true,
			},
		}, nil).Once()

		mocksCache.On("Set", "GetSpIndustryTypeMaster", []byte("SuccessRetrieve"))

		mockRepository.On("GetInquirySearch", req, 1).Return([]entity.InquirySearch{}, 1, errors.New(constant.RECORD_NOT_FOUND)).Once()

		_, _, err := usecase.GetSearchInquiry(ctx, req, 1)

		// Verifikasi bahwa error yang diharapkan terjadi
		assert.Error(t, err)
	})
}

func TestUsecaseGetSearchInquiry(t *testing.T) {

	testcases := []struct {
		name      string
		row       int
		req       request.ReqSearchInquiry
		data      []entity.InquiryDataSearch
		inquiry   []entity.InquirySearch
		photos    []entity.DataPhoto
		surveyor  []entity.TrxSurveyor
		trxDetail []entity.TrxDetail
		errGet    error
		errFinal  error
	}{
		{
			name:      "test success get ",
			req:       request.ReqSearchInquiry{},
			photos:    []entity.DataPhoto{},
			surveyor:  []entity.TrxSurveyor{},
			trxDetail: []entity.TrxDetail{},
			inquiry: []entity.InquirySearch{
				{
					FinalStatus: constant.DB_DECISION_BRANCH_MANAGER,
				},
			},
			row: 1,
		},
		{
			name: "test success get data approve",
			req:  request.ReqSearchInquiry{},
			photos: []entity.DataPhoto{
				{
					PhotoID: "KTP",
					Url:     "jsdhfigshhjgh",
				},
			},
			surveyor: []entity.TrxSurveyor{
				{
					SurveyorName: "ujang",
				},
			},
			trxDetail: []entity.TrxDetail{
				{
					ProspectID: "PPID",
				},
			},
			inquiry: []entity.InquirySearch{
				{
					FinalStatus: constant.DB_DECISION_BRANCH_MANAGER,
				},
			},
			row: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mocksCache := &mocksCache.Repository{}
			usecase := NewUsecase(mockRepository, mockHttpClient, mocksCache)
			mocksCache.On("Get", mock.Anything).Return([]byte("SuccessRetrieve"), nil)
			mockRepository.On("GetInquirySearch", tc.req, 1).Return(tc.inquiry, 1, tc.errGet).Once()
			mockRepository.On("GetSpIndustryTypeMaster").Return(tc.inquiry, 1, tc.errGet).Once()
			mockRepository.On("GetCustomerPhoto", mock.Anything).Return(tc.photos, nil).Once()
			mockRepository.On("GetSurveyorData", mock.Anything).Return(tc.surveyor, nil).Once()
			mockRepository.On("GetHistoryProcess", mock.Anything).Return(tc.trxDetail, nil).Once()

			result, _, _ := usecase.GetSearchInquiry(context.Background(), tc.req, mock.Anything)
			assert.Equal(t, 1, len(result))
		})
	}
}
