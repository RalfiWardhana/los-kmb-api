package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckNokaNosin(t *testing.T) {
	accessToken := ""
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name                         string
		request                      request.PrincipleAsset
		errAgreementChasisNumber     error
		resAgreementChasisNumberCode int
		resAgreementChasisNumberBody string
		result                       response.UsecaseApi
		err                          error
		expectSavePrincipleStepOne   bool
	}{
		{
			name: "error get agreement chassis number",
			request: request.PrincipleAsset{
				ProspectID: "SAL-123",
				NoChassis:  "123",
			},
			errAgreementChasisNumber: errors.New("something wrong"),
			err:                      errors.New("something wrong"),
		},
		{
			name: "error status code get agreement chassis number",
			request: request.PrincipleAsset{
				ProspectID: "SAL-123",
				NoChassis:  "123",
			},
			resAgreementChasisNumberCode: 400,
		},
		{
			name: "error unmarshal",
			request: request.PrincipleAsset{
				ProspectID: "SAL-123",
				NoChassis:  "123",
			},
			resAgreementChasisNumberCode: 200,
			resAgreementChasisNumberBody: "invalid json",
			err:                          &json.SyntaxError{},
		},
		{
			name: "reject consumen not match",
			request: request.PrincipleAsset{
				ProspectID: "SAL-123",
				IDNumber:   "7612379",
				NoChassis:  "123",
			},
			resAgreementChasisNumberCode: 200,
			resAgreementChasisNumberBody: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"161234339","installment_amount":0,"is_active":true,"is_registered":true,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_REJECT_CHASSIS_NUMBER,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_CHASSIS_NUMBER,
			},
			expectSavePrincipleStepOne: true,
		},
		{
			name: "pass consumen match",
			request: request.PrincipleAsset{
				ProspectID: "SAL-123",
				IDNumber:   "7612379",
				NoChassis:  "123",
			},
			resAgreementChasisNumberCode: 200,
			resAgreementChasisNumberBody: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"7612379","installment_amount":0,"is_active":true,"is_registered":true,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_OK_CONSUMEN_MATCH,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_OK_CONSUMEN_MATCH,
			},
			expectSavePrincipleStepOne: true,
		},
		{
			name: "reject fraud potential",
			request: request.PrincipleAsset{
				ProspectID:     "SAL-123",
				IDNumber:       "7612379",
				NoChassis:      "123",
				SpouseIDNumber: "161234339",
			},
			resAgreementChasisNumberCode: 200,
			resAgreementChasisNumberBody: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"161234339","installment_amount":0,"is_active":true,"is_registered":true,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_REJECTION_FRAUD_POTENTIAL,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECTION_FRAUD_POTENTIAL,
			},
			expectSavePrincipleStepOne: true,
		},
		{
			name: "pass agreement not found",
			request: request.PrincipleAsset{
				ProspectID:     "SAL-123",
				IDNumber:       "7612379",
				NoChassis:      "123",
				SpouseIDNumber: "161234339",
			},
			resAgreementChasisNumberCode: 200,
			resAgreementChasisNumberBody: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"","installment_amount":0,"is_active":false,"is_registered":false,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_AGREEMENT_NOT_FOUND,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_AGREEMENT_NOT_FOUND,
			},
			expectSavePrincipleStepOne: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"), httpmock.NewStringResponder(tc.resAgreementChasisNumberCode, tc.resAgreementChasisNumberBody))
			resp, _ := rst.R().Post(os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+tc.request.NoChassis, []byte(nil), map[string]string{}, constant.METHOD_GET, true, 6, 60, tc.request.ProspectID, accessToken).Return(resp, tc.errAgreementChasisNumber).Once()
			if tc.expectSavePrincipleStepOne {
				mockRepository.On("SavePrincipleStepOne", mock.AnythingOfType("entity.TrxPrincipleStepOne")).Return(nil).Once()
			}

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.CheckNokaNosin(ctx, tc.request)
			require.Equal(t, tc.result, result)
			if tc.err != nil {
				require.IsType(t, tc.err, err)
				if tc.name == "error unmarshal" {
					require.IsType(t, &json.SyntaxError{}, err)
				}
			} else {
				require.NoError(t, err)
			}

			mockRepository.AssertExpectations(t)
			mockHttpClient.AssertExpectations(t)
		})
	}
}
