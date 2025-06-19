package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	platformEventMockery "los-kmb-api/shared/common/platformevent/mocks"
	"testing"

	"github.com/stretchr/testify/require"
)

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
							"amount":                         float64(0),
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
		})
	}
}
