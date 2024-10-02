package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/common/platformevent"
	mockplatformevent "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrincipleEmergencyContact(t *testing.T) {
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name                             string
		request                          request.PrincipleEmergencyContact
		principleStepThree               entity.TrxPrincipleStepThree
		errGetPrincipleStepThree         error
		errSavePrincipleEmergencyContact error
		errSaveToWorker                  error
		expectSaveToWorker               bool
		err                              error
	}{
		{
			name: "success",
			request: request.PrincipleEmergencyContact{
				ProspectID:   "PROS-001",
				Name:         "John Doe",
				Relationship: "Friend",
				MobilePhone:  "1234567890",
			},
			principleStepThree: entity.TrxPrincipleStepThree{IDNumber: "123456"},
			expectSaveToWorker: true,
		},
		{
			name: "error get principle step three",
			request: request.PrincipleEmergencyContact{
				ProspectID: "PROS-002",
			},
			errGetPrincipleStepThree: errors.New("failed to get principle step three"),
			err:                      errors.New("failed to get principle step three"),
		},
		{
			name: "error save principle emergency contact",
			request: request.PrincipleEmergencyContact{
				ProspectID: "PROS-003",
			},
			errSavePrincipleEmergencyContact: errors.New("failed to save principle emergency contact"),
			err:                              errors.New("failed to save principle emergency contact"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			mockRepository.On("GetPrincipleStepThree", tc.request.ProspectID).Return(tc.principleStepThree, tc.errGetPrincipleStepThree)
			mockRepository.On("SavePrincipleEmergencyContact", mock.Anything, mock.Anything).Return(tc.errSavePrincipleEmergencyContact)

			if tc.expectSaveToWorker {
				mockRepository.On("SaveToWorker", mock.Anything).Return(tc.errSaveToWorker).Maybe()
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, platformEvent)

			err := usecase.PrincipleEmergencyContact(ctx, tc.request, accessToken)

			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			time.Sleep(100 * time.Millisecond)

			mockRepository.AssertExpectations(t)
			mockHttpClient.AssertExpectations(t)
		})
	}
}
