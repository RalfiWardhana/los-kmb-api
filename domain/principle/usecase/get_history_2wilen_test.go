package usecase

import (
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHistory2Wilen(t *testing.T) {
	testCases := []struct {
		name             string
		request          request.History2Wilen
		historyData      []entity.TrxKPMStatusHistory
		errHistory       error
		expectedResponse []response.History2Wilen
		expectedError    error
	}{
		{
			name: "success with multiple records",
			request: request.History2Wilen{
				ProspectID: ptr("PROS-123"),
			},
			historyData: []entity.TrxKPMStatusHistory{
				{
					ID:           "1",
					ProspectID:   "PROS-123",
					Decision:     "WAITING",
					CreatedAt:    time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
					IDNumber:     ptr("ID-111"),
					KpmID:        ptr(111),
					ReferralCode: ptr("REF-111"),
				},
				{
					ID:           "2",
					ProspectID:   "PROS-123",
					Decision:     "APPROVED",
					CreatedAt:    time.Date(2024, 1, 2, 15, 45, 0, 0, time.UTC),
					IDNumber:     ptr("ID-222"),
					KpmID:        ptr(222),
					ReferralCode: ptr("REF-222"),
					LoanAmount:   ptr(10.000),
				},
			},
			expectedResponse: []response.History2Wilen{
				{
					ID:              "1",
					ProspectID:      "PROS-123",
					OrderStatusName: "WAITING",
					CreatedAt:       "2024-01-01 10:30:00",
					IDNumber:        ptr("ID-111"),
					KPMID:           ptr(111),
					ReferralCode:    ptr("REF-111"),
				},
				{
					ID:              "2",
					ProspectID:      "PROS-123",
					OrderStatusName: "APPROVED",
					CreatedAt:       "2024-01-02 15:45:00",
					IDNumber:        ptr("ID-222"),
					KPMID:           ptr(222),
					ReferralCode:    ptr("REF-222"),
					LoanAmount:      ptr(10.000),
				},
			},
		},
		{
			name: "success with single record",
			request: request.History2Wilen{
				ProspectID: ptr("PROS-123"),
			},
			historyData: []entity.TrxKPMStatusHistory{
				{
					ID:           "1",
					ProspectID:   "PROS-123",
					Decision:     "WAITING",
					CreatedAt:    time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
					IDNumber:     ptr("ID-111"),
					KpmID:        ptr(111),
					ReferralCode: ptr("REF-111"),
				},
			},
			expectedResponse: []response.History2Wilen{
				{
					ID:              "1",
					ProspectID:      "PROS-123",
					OrderStatusName: "WAITING",
					CreatedAt:       "2024-01-01 10:30:00",
					IDNumber:        ptr("ID-111"),
					KPMID:           ptr(111),
					ReferralCode:    ptr("REF-111"),
				},
			},
		},
		{
			name: "success with empty records",
			request: request.History2Wilen{
				ProspectID: ptr("PROS-123"),
			},
			historyData:      []entity.TrxKPMStatusHistory{},
			expectedResponse: nil,
		},
		{
			name: "error getting history",
			request: request.History2Wilen{
				ProspectID: ptr("PROS-123"),
			},
			errHistory:    errors.New("database error"),
			expectedError: errors.New("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockRepository.
				On("GetTrxKPMStatusHistory", tc.request).
				Return(tc.historyData, tc.errHistory)

			usecase := NewUsecase(mockRepository, nil, nil)

			result, err := usecase.History2Wilen(tc.request)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tc.expectedResponse), len(result))

				for i := range result {
					require.Equal(t, tc.expectedResponse[i], result[i])
				}
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}
