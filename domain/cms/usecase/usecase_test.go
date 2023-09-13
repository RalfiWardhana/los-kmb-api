package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/domain/cms/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func useCaseUnitTest(t *testing.T) (interfaces.Usecase, *mocks.Repository) {

	var cache *bigcache.BigCache

	mockRepository := new(mocks.Repository)
	mockHttpClient := new(httpclient.MockHttpClient)

	cache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(time.Duration(1000) * time.Second))

	return NewUsecase(mockRepository, mockHttpClient, cache), mockRepository
}

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
	mockRepository := new(mocks.Repository)
	mockHttpClient := new(httpclient.MockHttpClient)
	var cache *bigcache.BigCache
	usecase := NewUsecase(mockRepository, mockHttpClient, cache)

	var (
		errSave      error
		reason       string
		prescreening entity.TrxPrescreening
		trxDetail    entity.TrxDetail
		trxStatus    entity.TrxStatus
		data         response.ReviewPrescreening
	)

	// Kasus uji 1: Status UNPROCESS dan SourceDecision PRESCREENING, Decision APPROVE
	t.Run("ValidReviewCase", func(t *testing.T) {
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
		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(status, errSave)
		mockRepository.On("SavePrescreening", prescreening, trxDetail, trxStatus).Return(errSave)

		result, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa tidak ada error yang terjadi
		if err != nil {
			t.Errorf("Expected no error, but got: %v", err)
		}

		// Verifikasi bahwa data yang dikembalikan sesuai dengan ekspektasi
		// Anda dapat menambahkan lebih banyak asserstion sesuai kebutuhan
		assert.Equal(t, constant.DB_DECISION_APR, result.Decision)
	})

	// Kasus uji 2: Status UNPROCESS dan SourceDecision PRESCREENING, Decision tidak valid
	t.Run("InvalidDecisionCase", func(t *testing.T) {
		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   "InvalidDecision",
			Reason:     "Invalid reason",
			DecisionBy: "User123",
		}

		_, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		assert.Error(t, err)
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

func TestReviewPrescreeningStatusError(t *testing.T) {
	// Persiapkan objek usecase dengan mock repository
	mockRepository := new(mocks.Repository)
	mockHttpClient := new(httpclient.MockHttpClient)
	var cache *bigcache.BigCache
	usecase := NewUsecase(mockRepository, mockHttpClient, cache)

	// Kasus uji 4: Status tidak UNPROCESS atau SourceDecision bukan PRESCREENING
	t.Run("ErrorStatusCase", func(t *testing.T) {
		req := request.ReqReviewPrescreening{
			ProspectID: "TST-DEV", // Ganti dengan ID yang sesuai
			Decision:   constant.DECISION_APPROVE,
			Reason:     "Valid reason",
			DecisionBy: "User123",
		}
		errFinal := errors.New(constant.ERROR_UPSTREAM + " - Get status prescreening error")

		mockRepository.On("GetStatusPrescreening", req.ProspectID).Return(entity.TrxStatus{}, errors.New(constant.RECORD_NOT_FOUND))

		_, err := usecase.ReviewPrescreening(context.Background(), req)

		// Verifikasi bahwa error yang diharapkan terjadi
		require.Equal(t, errFinal, err)
	})

}

func Test_usecase_GetInquiryPrescreening(t *testing.T) {
	// Create an instance of your usecase with the mock repository and cache.
	var cache *bigcache.BigCache
	u := usecase{
		repository: &mocks.Repository{},
		cache:      cache,
	}

	// Create a test request and pagination data.
	req := request.ReqInquiryPrescreening{}
	pagination := interface{}(nil)
	ctx := context.Background()

	u.cache = nil

	data, _, err := u.GetInquiryPrescreening(ctx, req, pagination)

	assert.Nil(t, err)
	require.Equal(t, []entity.InquiryData(nil), data)
}
