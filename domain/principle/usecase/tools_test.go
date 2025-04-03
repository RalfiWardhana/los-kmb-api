package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	platformEventMockery "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/mock"
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

func TestCheckOrderPendingPrinciple(t *testing.T) {
	ctx := context.Background()
	middlewares.UserInfoData.AccessToken = "test-token"

	testCases := []struct {
		name          string
		pendingOrders []entity.AutoCancel
		errScan       error
		errUpdate     error
		errPublish    error
		expectedError error
		updateCalled  bool
		publishCalled bool
	}{
		{
			name: "success with multiple orders",
			pendingOrders: []entity.AutoCancel{
				{
					ProspectID: "PROS-123",
					KPMID:      123,
					AssetCode:  "MOT",
					BranchID:   "BR001",
				},
				{
					ProspectID: "PROS-124",
					KPMID:      124,
					AssetCode:  "MOT",
					BranchID:   "BR002",
				},
			},
			updateCalled:  true,
			publishCalled: true,
		},
		{
			name: "success with update error skips publish",
			pendingOrders: []entity.AutoCancel{
				{
					ProspectID: "PROS-123",
					KPMID:      123,
					AssetCode:  "MOT",
					BranchID:   "BR001",
				},
			},
			errUpdate:     errors.New("update error"),
			updateCalled:  true,
			publishCalled: false,
		},
		{
			name:          "success with empty orders",
			pendingOrders: []entity.AutoCancel{},
			updateCalled:  false,
			publishCalled: false,
		},
		{
			name:          "error scanning orders",
			errScan:       errors.New("scan error"),
			expectedError: errors.New("scan error"),
			updateCalled:  false,
			publishCalled: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockProducer := new(platformEventMockery.PlatformEventInterface)

			mockRepository.On("ScanOrderPending").Return(tc.pendingOrders, tc.errScan)

			if tc.updateCalled {
				for _, order := range tc.pendingOrders {
					mockRepository.On("UpdateToCancel", order.ProspectID).Return(tc.errUpdate)

					if tc.publishCalled && tc.errUpdate == nil {
						expectedPayload := map[string]interface{}{
							"order_id":                       order.ProspectID,
							"kpm_id":                         order.KPMID,
							"source":                         3,
							"status_code":                    "LOS-C",
							"product_name":                   order.AssetCode,
							"branch_code":                    order.BranchID,
							"asset_type_code":                "11",
							"amount":                         0,
							"referral_code":                  "",
							"is_2w_principle_approval_order": false,
						}

						mockProducer.On("PublishEvent",
							ctx,
							middlewares.UserInfoData.AccessToken,
							"",
							"",
							order.ProspectID,
							expectedPayload,
							0,
						).Return(tc.errPublish)
					}
				}
			}

			usecase := NewUsecase(mockRepository, nil, mockProducer)

			err := usecase.CheckOrderPendingPrinciple(ctx)

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
					"amount":                         0,
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
				StatusCode: "APPROVE",
			},
			trxKPM: entity.TrxKPM{
				KPMID:        1,
				AssetCode:    "MOT",
				BranchID:     "BR001",
				ReferralCode: "TQ72AJ",
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
				KPMID:        1,
				AssetCode:    "MOT",
				BranchID:     "BR001",
				ReferralCode: "TQ72AJ",
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
					"amount":                         0,
					"referral_code":                  tc.trxKPM.ReferralCode,
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

func TestPrincipleMarketingProgram(t *testing.T) {
	ctx := context.Background()
	accessToken := "test-token"
	middlewares.UserInfoData.AccessToken = accessToken

	sampleTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("MDM_MASTER_BRANCH_URL", "http://test-mdm/branch/")
	os.Setenv("SALLY_SUBMISSION_2W_PRINCIPLE_URL", "http://test-sally/submit")

	defaultFilteringKMB := entity.FilteringKMB{
		Decision:        "PASS",
		Reason:          "OK",
		CustomerStatus:  interface{}(constant.STATUS_KONSUMEN_NEW),
		CustomerSegment: interface{}(constant.RO_AO_REGULAR),
		IsBlacklist:     0,
		NextProcess:     1,
	}

	testCases := []struct {
		name                string
		prospectID          string
		principleStepOne    entity.TrxPrincipleStepOne
		errPrincipleOne     error
		principleStepTwo    entity.TrxPrincipleStepTwo
		errPrincipleTwo     error
		principleStepThree  entity.TrxPrincipleStepThree
		errPrincipleThree   error
		emergencyContact    entity.TrxPrincipleEmergencyContact
		errEmergencyContact error
		marketingProgram    entity.TrxPrincipleMarketingProgram
		errMarketingProgram error
		filteringKMB        entity.FilteringKMB
		errFilteringKMB     error
		elaborateLTV        entity.MappingElaborateLTV
		errElaborateLTV     error
		detailBiro          []entity.TrxDetailBiro
		errDetailBiro       error
		branchResStatus     int
		branchResBody       string
		branchErr           error
		sallyResStatus      int
		sallyResBody        string
		sallyErr            error
		errPublishEvent     error
		expectedError       error
	}{
		{
			name:       "success",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				BranchID:        "BR001",
				CMOID:           "CMO123",
				CMOName:         "Test CMO",
				LicensePlate:    "B1234CD",
				BPKBName:        "K",
				OwnerAsset:      "Test Owner",
				STNKExpiredDate: sampleTime,
				TaxDate:         sampleTime,
				ManufactureYear: "2023",
				AssetCode:       "MOT",
				Color:           "BLACK",
				CC:              "150",
				NoChassis:       "CHS123",
				NoEngine:        "ENG123",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				KtpPhoto:    "http://test/ktp.jpg",
				SelfiePhoto: "http://test/selfie.jpg",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				InstallmentAmount: 1000000,
				AssetCategoryID:   "CAT1",
				Tenor:             24,
				OTR:               25000000,
				FinancePurpose:    "PRODUCTIVE",
				Dealer:            "NON-PSA",
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				CustomerID: 12345,
			},
			marketingProgram: entity.TrxPrincipleMarketingProgram{
				ProgramID:                  "PRG001",
				ProgramName:                "Test Program",
				ProductOfferingID:          "OFF001",
				ProductOfferingDescription: "Test Offering",
				LoanAmount:                 20000000,
				LoanAmountMaximum:          22000000,
				AdminFee:                   500000,
				ProvisionFee:               200000,
				DPAmount:                   5000000,
				FinanceAmount:              20000000,
			},
			filteringKMB: entity.FilteringKMB{
				Decision:        "PASS",
				Reason:          "OK", // Initialize with non-nil value
				CustomerStatus:  interface{}(constant.STATUS_KONSUMEN_NEW),
				CustomerSegment: interface{}(constant.RO_AO_REGULAR), // Changed from stringPtr
				IsBlacklist:     0,
				NextProcess:     1,
			},
			elaborateLTV: entity.MappingElaborateLTV{
				LTV: 80,
			},
			detailBiro: []entity.TrxDetailBiro{
				{
					Subject:      constant.CUSTOMER,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
			},
			branchResStatus: 200,
			branchResBody: `{
				"data": {
					"branch_name": "Test Branch"
				}
			}`,
			sallyResStatus: 200,
			sallyResBody: `{
				"message": "Success"
			}`,
		},
		{
			name:            "error get principle step one",
			prospectID:      "PROS-123",
			errPrincipleOne: errors.New("database error"),
			filteringKMB:    defaultFilteringKMB, // Add default value
			expectedError:   errors.New("database error"),
		},
		{
			name:            "error get principle step two",
			prospectID:      "PROS-123",
			errPrincipleTwo: errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
		{
			name:              "error get principle step three",
			prospectID:        "PROS-123",
			errPrincipleThree: errors.New("database error"),
			expectedError:     errors.New("database error"),
		},
		{
			name:                "error get emergency contact",
			prospectID:          "PROS-123",
			errEmergencyContact: errors.New("database error"),
			expectedError:       errors.New("database error"),
		},
		{
			name:                "error get marketing program",
			prospectID:          "PROS-123",
			errMarketingProgram: errors.New("database error"),
			expectedError:       errors.New("database error"),
		},
		{
			name:            "error get filtering result",
			prospectID:      "PROS-123",
			errFilteringKMB: errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
		{
			name:            "error get elaborate ltv",
			prospectID:      "PROS-123",
			errElaborateLTV: errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
		{
			name:          "error get detail biro",
			prospectID:    "PROS-123",
			errDetailBiro: errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:            "error branch api",
			prospectID:      "PROS-123",
			branchResStatus: 500,
			expectedError:   errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Branch Error"),
		},
		{
			name:            "error branch api unmarshal",
			prospectID:      "PROS-123",
			branchResStatus: 200,
			branchResBody:   "invalid json",
			expectedError:   errors.New("unexpected end of JSON input"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockProducer := new(platformEventMockery.PlatformEventInterface)

			mockRepository.On("GetPrincipleStepOne", tc.prospectID).Return(tc.principleStepOne, tc.errPrincipleOne)
			mockRepository.On("GetPrincipleStepTwo", tc.prospectID).Return(tc.principleStepTwo, tc.errPrincipleTwo)
			mockRepository.On("GetPrincipleStepThree", tc.prospectID).Return(tc.principleStepThree, tc.errPrincipleThree)
			mockRepository.On("GetPrincipleEmergencyContact", tc.prospectID).Return(tc.emergencyContact, tc.errEmergencyContact)
			mockRepository.On("GetPrincipleMarketingProgram", tc.prospectID).Return(tc.marketingProgram, tc.errMarketingProgram)
			mockRepository.On("GetFilteringResult", tc.prospectID).Return(tc.filteringKMB, tc.errFilteringKMB)
			mockRepository.On("GetElaborateLtv", tc.prospectID).Return(tc.elaborateLTV, tc.errElaborateLTV)
			mockRepository.On("GetTrxDetailBIro", tc.prospectID).Return(tc.detailBiro, tc.errDetailBiro)

			if tc.errPrincipleOne == nil && tc.errPrincipleTwo == nil && tc.errPrincipleThree == nil &&
				tc.errEmergencyContact == nil && tc.errMarketingProgram == nil && tc.errFilteringKMB == nil &&
				tc.errElaborateLTV == nil && tc.errDetailBiro == nil {

				branchURL := os.Getenv("MDM_MASTER_BRANCH_URL") + tc.principleStepOne.BranchID
				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_GET, branchURL,
					httpmock.NewStringResponder(tc.branchResStatus, tc.branchResBody))
				branchResp, _ := rst.R().Get(branchURL)

				mockHttpClient.On("EngineAPI",
					ctx,
					constant.DILEN_KMB_LOG,
					branchURL,
					[]byte(nil),
					map[string]string{
						"Content-Type":  "application/json",
						"Authorization": accessToken,
					},
					constant.METHOD_GET,
					false,
					0,
					mock.AnythingOfType("int"),
					tc.prospectID,
					accessToken,
				).Return(branchResp, tc.branchErr)

				if tc.branchResStatus == 200 {
					sallyURL := os.Getenv("SALLY_SUBMISSION_2W_PRINCIPLE_URL")
					httpmock.RegisterResponder(constant.METHOD_POST, sallyURL,
						httpmock.NewStringResponder(tc.sallyResStatus, tc.sallyResBody))
					sallyResp, _ := rst.R().Post(sallyURL)

					mockHttpClient.On("EngineAPI",
						ctx,
						constant.DILEN_KMB_LOG,
						sallyURL,
						mock.MatchedBy(func(param []byte) bool {
							var js map[string]interface{}
							return json.Unmarshal(param, &js) == nil
						}),
						map[string]string{
							"Content-Type":  "application/json",
							"Authorization": accessToken,
						},
						constant.METHOD_POST,
						false,
						0,
						mock.AnythingOfType("int"),
						tc.prospectID,
						accessToken,
					).Return(sallyResp, tc.sallyErr)

					if tc.sallyResStatus == 200 {
						mockProducer.On("PublishEvent",
							ctx,
							accessToken,
							constant.TOPIC_SUBMISSION_PRINCIPLE,
							constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE,
							tc.prospectID,
							mock.MatchedBy(func(data map[string]interface{}) bool {
								return data["order_id"] == tc.prospectID &&
									data["source"] == 3 &&
									data["status_code"] == constant.PRINCIPLE_STATUS_SUBMIT_SALLY &&
									data["product_name"] == tc.principleStepOne.AssetCode &&
									data["branch_code"] == tc.principleStepOne.BranchID &&
									data["asset_type_code"] == constant.KPM_ASSET_TYPE_CODE_MOTOR
							}),
							0,
						).Return(tc.errPublishEvent)
					}
				}
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, mockProducer)

			err := usecase.PrincipleMarketingProgram(ctx, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMapperBPKBOwnershipStatusID(t *testing.T) {
	testCases := []struct {
		name          string
		bpkbName      string
		expectedValue int
	}{
		{
			name:          "K returns 1",
			bpkbName:      "K",
			expectedValue: 1,
		},
		{
			name:          "P returns 2",
			bpkbName:      "P",
			expectedValue: 2,
		},
		{
			name:          "KK returns 3",
			bpkbName:      "KK",
			expectedValue: 3,
		},
		{
			name:          "O returns 4",
			bpkbName:      "O",
			expectedValue: 4,
		},
		{
			name:          "unknown returns 4",
			bpkbName:      "UNKNOWN",
			expectedValue: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MapperBPKBOwnershipStatusID(tc.bpkbName)
			require.Equal(t, tc.expectedValue, result)
		})
	}
}
