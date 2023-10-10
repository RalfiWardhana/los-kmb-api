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
		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(status, errSave).Once()
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
		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(status, errSave).Once()
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
		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(status, errSave).Once()
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

		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(entity.TrxStatus{
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
		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Get status prescreening error")

		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(entity.TrxStatus{}, errors.New(constant.RECORD_NOT_FOUND)).Once()

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
		mockRepository.On("GetStatusPrescreening", mock.Anything).Return(entity.TrxStatus{
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

		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(entity.TrxStatus{
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
			mockRepository.On("GetCustomerPhoto", mock.Anything).Return(tc.photos, nil).Once()
			mockRepository.On("GetSurveyorData", mock.Anything).Return(tc.surveyor, nil).Once()

			result, _, _ := usecase.GetInquiryPrescreening(context.Background(), tc.req, mock.Anything)
			assert.Equal(t, 1, len(result))
		})
	}
}
