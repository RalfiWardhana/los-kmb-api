package usecase

import (
	"context"
	"database/sql"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	platformEventMockery "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
)

func TestPrincipleStep(t *testing.T) {
	testCases := []struct {
		name           string
		idNumber       string
		trxPrinciple   entity.TrxPrincipleStatus
		errPrinciple   error
		trxStatus      entity.TrxStatus
		errStatus      error
		errUpdate      error
		expectedResult response.StepPrinciple
		expectedError  error
	}{
		{
			name:     "success decision pass step 1",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_PASS,
				Step:       1,
				UpdatedAt:  time.Now(),
			},
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				Status:     constant.REASON_ASSET_APPOVE,
				ColorCode:  "#00FF00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success decision pass step 2",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_PASS,
				Step:       2,
				UpdatedAt:  time.Now(),
			},
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				Status:     constant.REASON_PROFIL_APPROVE,
				ColorCode:  "#00FF00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success decision pass step 3",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_PASS,
				Step:       3,
				UpdatedAt:  time.Now(),
			},
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				Status:     constant.REASON_PEMBIAYAAN_APPROVE,
				ColorCode:  "#00FF00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success credit process no trx status",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_CREDIT_PROCESS,
				UpdatedAt:  time.Now(),
			},
			errStatus: gorm.ErrRecordNotFound,
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				Status:     constant.REASON_PROSES_SURVEY,
				ColorCode:  "#FFCC00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success credit process with cancel status",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_CREDIT_PROCESS,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
				Decision: constant.DB_DECISION_CANCEL,
			},
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success credit process with reject status",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_CREDIT_PROCESS,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
				Decision: constant.DB_DECISION_REJECT,
			},
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success credit process with approve status",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_CREDIT_PROCESS,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
				Decision: constant.DB_DECISION_APR,
			},
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success credit process with ongoing status",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_CREDIT_PROCESS,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: "ONGOING",
			},
			expectedResult: response.StepPrinciple{
				ProspectID: "PROS-123",
				Status:     constant.REASON_PROSES_SURVEY,
				ColorCode:  "#FFCC00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:          "error get principle status",
			idNumber:      "1234567890",
			errPrinciple:  errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:     "error get trx status",
			idNumber: "1234567890",
			trxPrinciple: entity.TrxPrincipleStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_CREDIT_PROCESS,
				UpdatedAt:  time.Now(),
			},
			errStatus:     errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:           "record not found principle status",
			idNumber:       "1234567890",
			errPrinciple:   gorm.ErrRecordNotFound,
			expectedResult: response.StepPrinciple{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)

			mockRepository.On("GetTrxPrincipleStatus", tc.idNumber).Return(tc.trxPrinciple, tc.errPrinciple)

			if tc.errPrinciple == nil && tc.trxPrinciple.Decision == constant.DECISION_CREDIT_PROCESS {
				mockRepository.On("GetTrxStatus", tc.trxPrinciple.ProspectID).Return(tc.trxStatus, tc.errStatus)

				if tc.errStatus == nil && tc.trxStatus.Activity == constant.ACTIVITY_STOP {
					switch tc.trxStatus.Decision {
					case constant.DB_DECISION_CANCEL:
						mockRepository.On("UpdateToCancel", tc.trxPrinciple.ProspectID).Return(tc.errUpdate)
					case constant.DB_DECISION_REJECT:
						mockRepository.On("UpdateTrxPrincipleStatus", tc.trxPrinciple.ProspectID, constant.DECISION_REJECT, 4).Return(tc.errUpdate)
					case constant.DB_DECISION_APR:
						mockRepository.On("UpdateTrxPrincipleStatus", tc.trxPrinciple.ProspectID, constant.DECISION_APPROVE, 4).Return(tc.errUpdate)
					}
				}
			}

			usecase := NewUsecase(mockRepository, nil, nil)

			result, err := usecase.PrincipleStep(tc.idNumber)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ProspectID, result.ProspectID)
				require.Equal(t, tc.expectedResult.Status, result.Status)
				require.Equal(t, tc.expectedResult.ColorCode, result.ColorCode)
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestStep2Wilen(t *testing.T) {
	testCases := []struct {
		name           string
		idNumber       string
		trxKPMStatus   entity.TrxKPMStatus
		errKPMStatus   error
		trxStatus      entity.TrxStatus
		errStatus      error
		errUpdate      error
		expectedResult response.Step2Wilen
		expectedError  error
	}{
		{
			name:     "success readjust",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_KPM_READJUST,
				UpdatedAt:  time.Now(),
			},
			expectedResult: response.Step2Wilen{
				ProspectID: "PROS-123",
				Status:     constant.DECISION_KPM_READJUST,
				ColorCode:  "#FFCC00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success wait 2wilen",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ProspectID: "PROS-123",
				Decision:   constant.STATUS_KPM_WAIT_2WILEN,
				UpdatedAt:  time.Now(),
			},
			expectedResult: response.Step2Wilen{
				ProspectID: "PROS-123",
				Status:     constant.STATUS_KPM_WAIT_2WILEN,
				ColorCode:  "#FFCC00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success kpm approve",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ProspectID: "PROS-123",
				Decision:   constant.DECISION_KPM_APPROVE,
				UpdatedAt:  time.Now(),
			},
			expectedResult: response.Step2Wilen{
				ProspectID: "PROS-123",
				Status:     constant.DECISION_KPM_APPROVE,
				ColorCode:  "#00FF00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success los process no trx status",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ProspectID: "PROS-123",
				Decision:   constant.STATUS_LOS_PROCESS_2WILEN,
				UpdatedAt:  time.Now(),
			},
			errStatus: gorm.ErrRecordNotFound,
			expectedResult: response.Step2Wilen{
				ProspectID: "PROS-123",
				Status:     constant.STATUS_LOS_PROCESS_2WILEN,
				ColorCode:  "#FFCC00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:     "success los process with cancel status",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ID:         "1",
				ProspectID: "PROS-123",
				Decision:   constant.STATUS_LOS_PROCESS_2WILEN,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
				Decision: constant.DB_DECISION_CANCEL,
			},
			expectedResult: response.Step2Wilen{},
		},
		{
			name:     "success los process with reject status",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ID:         "1",
				ProspectID: "PROS-123",
				Decision:   constant.STATUS_LOS_PROCESS_2WILEN,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
				Decision: constant.DB_DECISION_REJECT,
			},
			expectedResult: response.Step2Wilen{},
		},
		{
			name:     "success los process with approve status",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ID:         "1",
				ProspectID: "PROS-123",
				Decision:   constant.STATUS_LOS_PROCESS_2WILEN,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
				Decision: constant.DB_DECISION_APR,
			},
			expectedResult: response.Step2Wilen{},
		},
		{
			name:     "success los process with ongoing status",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ProspectID: "PROS-123",
				Decision:   constant.STATUS_LOS_PROCESS_2WILEN,
				UpdatedAt:  time.Now(),
			},
			trxStatus: entity.TrxStatus{
				Activity: "ONGOING",
			},
			expectedResult: response.Step2Wilen{
				ProspectID: "PROS-123",
				Status:     constant.STATUS_LOS_PROCESS_2WILEN,
				ColorCode:  "#FFCC00",
				UpdatedAt:  time.Now().Format(constant.FORMAT_DATE_TIME),
			},
		},
		{
			name:          "error get kpm status",
			idNumber:      "1234567890",
			errKPMStatus:  errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:     "error get trx status",
			idNumber: "1234567890",
			trxKPMStatus: entity.TrxKPMStatus{
				ProspectID: "PROS-123",
				Decision:   constant.STATUS_LOS_PROCESS_2WILEN,
				UpdatedAt:  time.Now(),
			},
			errStatus:     errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:           "record not found kpm status",
			idNumber:       "1234567890",
			errKPMStatus:   gorm.ErrRecordNotFound,
			expectedResult: response.Step2Wilen{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)

			mockRepository.On("GetTrxKPMStatus", tc.idNumber).Return(tc.trxKPMStatus, tc.errKPMStatus)

			if tc.errKPMStatus == nil && tc.trxKPMStatus.Decision == constant.STATUS_LOS_PROCESS_2WILEN {
				mockRepository.On("GetTrxStatus", tc.trxKPMStatus.ProspectID).Return(tc.trxStatus, tc.errStatus)

				if tc.errStatus == nil && tc.trxStatus.Activity == constant.ACTIVITY_STOP {
					switch tc.trxStatus.Decision {
					case constant.DB_DECISION_CANCEL:
						mockRepository.On("UpdateTrxKPMDecision", tc.trxKPMStatus.ID, tc.trxKPMStatus.ProspectID, constant.STATUS_LOS_CANCEL_2WILEN).Return(tc.errUpdate)
					case constant.DB_DECISION_REJECT:
						mockRepository.On("UpdateTrxKPMDecision", tc.trxKPMStatus.ID, tc.trxKPMStatus.ProspectID, constant.STATUS_LOS_REJECTED_2WILEN).Return(tc.errUpdate)
					case constant.DB_DECISION_APR:
						mockRepository.On("UpdateTrxKPMDecision", tc.trxKPMStatus.ID, tc.trxKPMStatus.ProspectID, constant.STATUS_LOS_APPROVED_2WILEN).Return(tc.errUpdate)
					}
				}
			}

			usecase := NewUsecase(mockRepository, nil, nil)

			result, err := usecase.Step2Wilen(tc.idNumber)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult.ProspectID, result.ProspectID)
				require.Equal(t, tc.expectedResult.Status, result.Status)
				require.Equal(t, tc.expectedResult.ColorCode, result.ColorCode)
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestPrinciplePublish(t *testing.T) {
	ctx := context.Background()
	accessToken := "test-token"

	testCases := []struct {
		name             string
		request          request.PrinciplePublish
		principleStepOne entity.TrxPrincipleStepOne
		errGetPrinciple  error
		errPublish       error
		expectedError    error
	}{
		{
			name: "success",
			request: request.PrinciplePublish{
				ProspectID: "PROS-123",
				StatusCode: "APPROVE",
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				KPMID:     1,
				AssetCode: "MOT",
				BranchID:  "BR001",
			},
		},
		{
			name: "error get principle step one",
			request: request.PrinciplePublish{
				ProspectID: "PROS-123",
				StatusCode: "APPROVE",
			},
			errGetPrinciple: errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
		{
			name: "error publish event",
			request: request.PrinciplePublish{
				ProspectID: "PROS-123",
				StatusCode: "APPROVE",
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				KPMID:     1,
				AssetCode: "MOT",
				BranchID:  "BR001",
			},
			errPublish:    errors.New("publish error"),
			expectedError: errors.New("publish error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockProducer := new(platformEventMockery.PlatformEventInterface)

			mockRepository.On("GetPrincipleStepOne", tc.request.ProspectID).Return(tc.principleStepOne, tc.errGetPrinciple)

			if tc.errGetPrinciple == nil {
				expectedPayload := map[string]interface{}{
					"order_id":                       tc.request.ProspectID,
					"kpm_id":                         tc.principleStepOne.KPMID,
					"source":                         3,
					"status_code":                    tc.request.StatusCode,
					"product_name":                   tc.principleStepOne.AssetCode,
					"branch_code":                    tc.principleStepOne.BranchID,
					"asset_type_code":                "11",
					"amount":                         float64(0),
					"referral_code":                  "",
					"is_2w_principle_approval_order": false,
				}

				mockProducer.On("PublishEvent",
					ctx,
					accessToken,
					"",
					"",
					tc.request.ProspectID,
					expectedPayload,
					0,
				).Return(tc.errPublish)
			}

			usecase := NewUsecase(mockRepository, nil, mockProducer)

			err := usecase.PrinciplePublish(ctx, tc.request, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			mockRepository.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}

func TestPublish2Wilen(t *testing.T) {
	ctx := context.Background()
	accessToken := "test-token"

	testCases := []struct {
		name          string
		request       request.Publish2Wilen
		trxKPM        entity.TrxKPM
		errGetTrxKPM  error
		errPublish    error
		expectedError error
	}{
		{
			name: "success",
			request: request.Publish2Wilen{
				ProspectID: "PROS-123",
				StatusCode: constant.DECISION_KPM_APPROVE,
			},
			trxKPM: entity.TrxKPM{
				KPMID:     1,
				AssetCode: "MOT",
				BranchID:  "BR001",
				ReferralCode: sql.NullString{
					String: "TQ72AJ",
					Valid:  true,
				},
			},
		},
		{
			name: "error get trx kpm",
			request: request.Publish2Wilen{
				ProspectID: "PROS-123",
				StatusCode: "APPROVE",
			},
			errGetTrxKPM:  errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name: "error publish event",
			request: request.Publish2Wilen{
				ProspectID: "PROS-123",
				StatusCode: "APPROVE",
			},
			trxKPM: entity.TrxKPM{
				KPMID:     1,
				AssetCode: "MOT",
				BranchID:  "BR001",
				ReferralCode: sql.NullString{
					String: "TQ72AJ",
					Valid:  true,
				},
			},
			errPublish:    errors.New("publish error"),
			expectedError: errors.New("publish error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockProducer := new(platformEventMockery.PlatformEventInterface)

			mockRepository.On("GetTrxKPM", tc.request.ProspectID).Return(tc.trxKPM, tc.errGetTrxKPM)

			if tc.errGetTrxKPM == nil {
				expectedPayload := map[string]interface{}{
					"order_id":                       tc.request.ProspectID,
					"kpm_id":                         tc.trxKPM.KPMID,
					"source":                         3,
					"status_code":                    tc.request.StatusCode,
					"product_name":                   tc.trxKPM.AssetCode,
					"branch_code":                    tc.trxKPM.BranchID,
					"asset_type_code":                "11",
					"amount":                         float64(0),
					"referral_code":                  tc.trxKPM.ReferralCode.String,
					"is_2w_principle_approval_order": true,
				}

				mockProducer.On("PublishEvent",
					ctx,
					accessToken,
					"",
					"",
					tc.request.ProspectID,
					expectedPayload,
					0,
				).Return(tc.errPublish)
			}

			usecase := NewUsecase(mockRepository, nil, mockProducer)

			err := usecase.Publish2Wilen(ctx, tc.request, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			mockRepository.AssertExpectations(t)
			mockProducer.AssertExpectations(t)
		})
	}
}
