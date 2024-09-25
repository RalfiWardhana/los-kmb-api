package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common/platformevent"
	mockplatformevent "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
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
		errGetMasterBranchCMO        error
		resGetMasterBranchCMOCode    int
		resGetMasterBranchCMOBody    string
		result                       response.UsecaseApi
		err                          error
		expectGetMasterBranchCMO     bool
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
			resGetMasterBranchCMOCode: 200,
			resGetMasterBranchCMOBody: `{"data":[{"cmo_id":"63478","cmo_name":"NURHAKIKI"},{"cmo_id":"26985","cmo_name":"YANTO ERIYANTO"}]}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_REJECT_CHASSIS_NUMBER,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_CHASSIS_NUMBER,
			},
			expectGetMasterBranchCMO:   true,
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
			resGetMasterBranchCMOCode: 200,
			resGetMasterBranchCMOBody: `{"data":[{"cmo_id":"63478","cmo_name":"NURHAKIKI"},{"cmo_id":"26985","cmo_name":"YANTO ERIYANTO"}]}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_OK_CONSUMEN_MATCH,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_OK_CONSUMEN_MATCH,
			},
			expectGetMasterBranchCMO:   true,
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
			resGetMasterBranchCMOCode: 200,
			resGetMasterBranchCMOBody: `{"data":[{"cmo_id":"63478","cmo_name":"NURHAKIKI"},{"cmo_id":"26985","cmo_name":"YANTO ERIYANTO"}]}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_REJECTION_FRAUD_POTENTIAL,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECTION_FRAUD_POTENTIAL,
			},
			expectGetMasterBranchCMO:   true,
			expectSavePrincipleStepOne: true,
		},
		{
			name: "error get master branch cmo",
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
			resGetMasterBranchCMOCode: 500,
			errGetMasterBranchCMO:     errors.New("something wrong"),
			result: response.UsecaseApi{
				Code:   constant.CODE_AGREEMENT_NOT_FOUND,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_AGREEMENT_NOT_FOUND,
			},
			err:                      errors.New("something wrong"),
			expectGetMasterBranchCMO: true,
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
			resGetMasterBranchCMOCode: 200,
			resGetMasterBranchCMOBody: `{"data":[{"cmo_id":"63478","cmo_name":"NURHAKIKI"},{"cmo_id":"26985","cmo_name":"YANTO ERIYANTO"}]}`,
			result: response.UsecaseApi{
				Code:   constant.CODE_AGREEMENT_NOT_FOUND,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_AGREEMENT_NOT_FOUND,
			},
			expectGetMasterBranchCMO:   true,
			expectSavePrincipleStepOne: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"), httpmock.NewStringResponder(tc.resAgreementChasisNumberCode, tc.resAgreementChasisNumberBody))
			resp, _ := rst.R().Post(os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+tc.request.NoChassis, []byte(nil), map[string]string{}, constant.METHOD_GET, true, 6, 60, tc.request.ProspectID, accessToken).Return(resp, tc.errAgreementChasisNumber)

			if tc.expectGetMasterBranchCMO {
				rst2 := resty.New()
				httpmock.ActivateNonDefault(rst2.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("MDM_MASTER_MAPPING_BRANCH_EMPLOYEE_URL"), httpmock.NewStringResponder(tc.resGetMasterBranchCMOCode, tc.resGetMasterBranchCMOBody))
				resp2, _ := rst2.R().Get(os.Getenv("MDM_MASTER_MAPPING_BRANCH_EMPLOYEE_URL"))

				mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MASTER_MAPPING_BRANCH_EMPLOYEE_URL")+"?lob_id="+strconv.Itoa(constant.LOBID_KMB)+"&branch_id="+tc.request.BranchID, []byte(nil), map[string]string{"Authorization": "", "Content-Type": "application/json"}, constant.METHOD_GET, false, 0, 0, tc.request.ProspectID, accessToken).Return(resp2, tc.errGetMasterBranchCMO)
			}

			if tc.expectSavePrincipleStepOne {
				mockRepository.On("SavePrincipleStepOne", mock.AnythingOfType("entity.TrxPrincipleStepOne")).Return(nil).Once()
				mockPlatformEvent.On("PublishEvent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, platformEvent)

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

func TestMDMGetMasterMappingBranchEmployee(t *testing.T) {
	os.Setenv("MDM_MASTER_MAPPING_BRANCH_EMPLOYEE_URL", "https://dev-core-masterdata-area-api.kbfinansia.com/api/v1/master-data/area/mapping/branch-employee")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		prospectID       string
		branchID         string
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.MDMMasterMappingBranchEmployeeResponse
		expectedError    error
	}{
		{
			name:             "success",
			prospectID:       "SAL-123",
			branchID:         "426",
			resEngineAPICode: 200,
			resEngineAPIBody: `{"data":[{"cmo_id":"63478","cmo_name":"NURHAKIKI"},{"cmo_id":"26985","cmo_name":"YANTO ERIYANTO"}]}`,
			expectedResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "63478",
						CMOName: "NURHAKIKI",
					},
					{
						CMOID:   "26985",
						CMOName: "YANTO ERIYANTO",
					},
				},
			},
			expectedError: nil,
		},
		{
			name:          "error api response",
			prospectID:    "SAL-123",
			branchID:      "426",
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New("network error"),
		},
		{
			name:             "not 200 status code",
			prospectID:       "SAL-123",
			branchID:         "426",
			resEngineAPICode: 400,
			resEngineAPIBody: `{"error": "Bad Request"}`,
			expectedError:    errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping Branch Employee Error"),
		},
		{
			name:             "invalid json response",
			prospectID:       "SAL-123",
			branchID:         "426",
			resEngineAPICode: 200,
			resEngineAPIBody: `invalid json`,
			expectedError:    errors.New("unexpected end of JSON input"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("MDM_MASTER_MAPPING_BRANCH_EMPLOYEE_URL") + "?lob_id=" + strconv.Itoa(constant.LOBID_KMB) + "&branch_id=" + tc.branchID
			httpmock.RegisterResponder(constant.METHOD_GET, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Get(url)

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, []byte(nil), map[string]string{
				"Content-Type":  "application/json",
				"Authorization": accessToken,
			}, constant.METHOD_GET, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MDMGetMasterMappingBranchEmployee(ctx, tc.prospectID, tc.branchID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}
