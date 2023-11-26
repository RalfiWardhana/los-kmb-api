package usecase

import (
	"context"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	ctx := context.Background()
	testcases := []struct {
		name          string
		reqMetrics    request.Metrics
		resultMetrics interface{}
		err           error
	}{}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockUsecase := new(mocks.Usecase)
			mockMultiUsecase := new(mocks.MultiUsecase)

			metrics := NewMetrics(mockRepository, mockHttpClient, mockUsecase, mockMultiUsecase)
			_, err := metrics.MetricsLos(ctx, tc.reqMetrics, "token")
			// require.Equal(t, tc.resultMetrics, result)
			require.Equal(t, tc.err, err)
		})
	}
}

func TestSaveTransaction(t *testing.T) {
	testcases := []struct {
		name            string
		countTrx        int
		data            request.Metrics
		trxPrescreening entity.TrxPrescreening
		trxFMF          response.TrxFMF
		details         []entity.TrxDetail
		reason          string
		resp            response.Metrics
		err, errsave    error
	}{
		{
			name: "test save transaction apr",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_APR,
				},
			},
			resp: response.Metrics{
				Decision: constant.DECISION_APPROVE,
			},
		},
		{
			name: "test save transaction rej",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_REJECT,
				},
			},
			resp: response.Metrics{
				Decision: constant.JSON_DECISION_REJECT,
			},
		},
		{
			name: "test save transaction pas",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_PASS,
				},
			},
			resp: response.Metrics{
				Decision: constant.JSON_DECISION_PASS,
			},
		},
		{
			name: "test save transaction cpr",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_CREDIT_PROCESS,
				},
			},
			resp: response.Metrics{
				Decision: constant.JSON_DECISION_CREDIT_PROCESS,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("SaveTransaction", tc.countTrx, tc.data, tc.trxPrescreening, tc.trxFMF, tc.details, tc.reason).Return(tc.errsave)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, err := usecase.SaveTransaction(tc.countTrx, tc.data, tc.trxPrescreening, tc.trxFMF, tc.details, tc.reason)
			require.Equal(t, tc.resp, data)
			require.Equal(t, tc.err, err)
		})
	}
}
