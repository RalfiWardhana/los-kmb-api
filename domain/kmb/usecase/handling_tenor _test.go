package usecase

import (
	"errors"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectTenor36(t *testing.T) {
	responseAppConfig := entity.AppConfig{
		Value: `{"data":["Cluster A", "Cluster B"]}`,
	}
	testcases := []struct {
		name              string
		cluster           string
		result            response.UsecaseApi
		errResult         error
		responseAppConfig error
		trxStatus         entity.TrxStatus
	}{
		{
			name:              "handling tenor error config",
			cluster:           "Cluster A",
			responseAppConfig: errors.New(constant.ERROR_UPSTREAM + " - GetConfig exclusion_tenor36 Error"),
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - GetConfig exclusion_tenor36 Error"),
		},
		{
			name:    "handling tenor reject",
			cluster: "Cluster C",
			result: response.UsecaseApi{
				Code:   constant.CODE_REJECT_TENOR,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_TENOR},
		},
		{
			name:    "handling tenor pass",
			cluster: "Cluster A",
			result: response.UsecaseApi{
				Code:   constant.CODE_PASS_TENOR,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_PASS_TENOR},
			trxStatus: entity.TrxStatus{
				ProspectID:     "SLY-123457678",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Decision:       constant.DB_DECISION_REJECT,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetConfig", "tenor36", "KMB-OFF", "exclusion_tenor36").Return(responseAppConfig, tc.responseAppConfig)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.RejectTenor36(tc.cluster)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errResult, err)
		})
	}

}
