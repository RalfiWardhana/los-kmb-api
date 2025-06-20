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
		principleStepOne                 entity.TrxPrincipleStepOne
		errGetPrincipleStepOne           error
		principleStepThree               entity.TrxPrincipleStepThree
		errGetPrincipleStepThree         error
		resGetTrxWorker                  []entity.TrxWorker
		errGetTrxWorker                  error
		errSavePrincipleEmergencyContact error
		errSaveToWorker                  error
		errPrincipleCustomer             error
		errPrincipleMarketingProgram     error
		expectGetTrxWorker               bool
		expectSaveEmergencyContact       bool
		expectSaveToWorker               bool
		expectedError                    error
		expectPublishEvent               bool
	}{
		{
			name: "success",
			request: request.PrincipleEmergencyContact{
				ProspectID:   "PROS-001",
				Name:         "John Doe",
				Relationship: "Friend",
				MobilePhone:  "12345678901111",
				Phone:        "12345678901111",
			},
			principleStepThree:         entity.TrxPrincipleStepThree{IDNumber: "123456"},
			expectGetTrxWorker:         true,
			expectSaveEmergencyContact: true,
			expectSaveToWorker:         true,
			expectPublishEvent:         true,
		},
		{
			name: "error get principle step one",
			request: request.PrincipleEmergencyContact{
				ProspectID: "PROS-002",
			},
			errGetPrincipleStepOne:     errors.New("failed to get principle step one"),
			expectSaveEmergencyContact: false,
			expectedError:              errors.New("failed to get principle step one"),
		},
		{
			name: "error get principle step three",
			request: request.PrincipleEmergencyContact{
				ProspectID: "PROS-002",
			},
			errGetPrincipleStepThree:   errors.New("failed to get principle step three"),
			expectSaveEmergencyContact: false,
			expectedError:              errors.New("failed to get principle step three"),
		},
		{
			name: "error reject principle step three",
			request: request.PrincipleEmergencyContact{
				ProspectID: "PROS-002",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				Decision: constant.DECISION_REJECT,
			},
			expectSaveEmergencyContact: false,
			expectedError:              errors.New(constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE),
		},
		{
			name: "error get trx worker",
			request: request.PrincipleEmergencyContact{
				ProspectID: "PROS-003",
			},
			principleStepThree:         entity.TrxPrincipleStepThree{IDNumber: "123456"},
			errGetTrxWorker:            errors.New("failed to get trx worker"),
			expectSaveEmergencyContact: true,
			expectGetTrxWorker:         true,
			expectedError:              errors.New("failed to get trx worker"),
		},
		{
			name: "error save principle emergency contact",
			request: request.PrincipleEmergencyContact{
				ProspectID: "PROS-003",
			},
			principleStepThree:               entity.TrxPrincipleStepThree{IDNumber: "123456"},
			errSavePrincipleEmergencyContact: errors.New("failed to save principle emergency contact"),
			expectSaveEmergencyContact:       true,
			expectGetTrxWorker:               true,
			expectedError:                    errors.New("failed to save principle emergency contact"),
		},
		{
			name: "error principle customer",
			request: request.PrincipleEmergencyContact{
				ProspectID:   "PROS-001",
				Name:         "John Doe",
				Relationship: "Friend",
				MobilePhone:  "12345678901111",
				Phone:        "12345678901111",
			},
			principleStepThree:         entity.TrxPrincipleStepThree{IDNumber: "123456"},
			errPrincipleCustomer:       errors.New("failed"),
			expectedError:              errors.New("failed"),
			expectGetTrxWorker:         true,
			expectSaveEmergencyContact: true,
			expectSaveToWorker:         true,
			expectPublishEvent:         false,
		},
		{
			name: "error principle marketing program",
			request: request.PrincipleEmergencyContact{
				ProspectID:   "PROS-001",
				Name:         "John Doe",
				Relationship: "Friend",
				MobilePhone:  "12345678901111",
				Phone:        "12345678901111",
			},
			principleStepThree:           entity.TrxPrincipleStepThree{IDNumber: "123456"},
			errPrincipleMarketingProgram: errors.New("failed"),
			expectedError:                errors.New("failed"),
			expectGetTrxWorker:           true,
			expectSaveEmergencyContact:   true,
			expectSaveToWorker:           true,
			expectPublishEvent:           false,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockUsecase := new(mocks.Usecase)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			mockRepository.On("GetPrincipleStepOne", tc.request.ProspectID).Return(tc.principleStepOne, tc.errGetPrincipleStepOne)
			mockRepository.On("GetPrincipleStepThree", tc.request.ProspectID).Return(tc.principleStepThree, tc.errGetPrincipleStepThree)

			if tc.expectGetTrxWorker {
				mockRepository.On("GetTrxWorker", tc.request.ProspectID, mock.Anything).Return(tc.resGetTrxWorker, tc.errGetTrxWorker)
			}

			if tc.expectSaveEmergencyContact {
				mockRepository.On("SavePrincipleEmergencyContact", mock.Anything, mock.Anything).Return(tc.errSavePrincipleEmergencyContact)
			}

			mockUsecase.On("PrincipleCoreCustomer", ctx, tc.request.ProspectID, accessToken).Return(tc.errPrincipleCustomer)

			if tc.errPrincipleCustomer == nil {
				mockUsecase.On("PrincipleMarketingProgram", ctx, tc.request.ProspectID, accessToken).Return(tc.errPrincipleMarketingProgram)
			}

			if tc.expectSaveToWorker {
				mockRepository.On("SaveToWorker", mock.Anything).Return(tc.errSaveToWorker)
			}

			if tc.expectPublishEvent {
				mockPlatformEvent.On("PublishEvent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}

			multiUsecase := NewMultiUsecase(mockRepository, mockHttpClient, platformEvent, mockUsecase)

			_, err := multiUsecase.PrincipleEmergencyContact(ctx, tc.request, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
