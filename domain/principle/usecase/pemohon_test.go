package usecase

import (
	"context"
	"errors"
	"fmt"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
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
	"time"

	errorLib "github.com/KB-FMF/los-common-library/errors"
	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetEmployeeData(t *testing.T) {
	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	testCases := []struct {
		name           string
		employeeID     string
		mockResponse   response.GetEmployeeByID
		mockStatusCode int
		mockError      error
		expectedError  error
		expectedData   response.EmployeeCMOResponse
	}{
		{
			name:       "success",
			employeeID: "123",
			mockResponse: response.GetEmployeeByID{
				Data: []response.EmployeeCareerHistory{
					{
						EmployeeID:        "123",
						EmployeeName:      "John Doe",
						PositionGroupCode: "AO",
						PositionGroupName: "Credit Marketing Officer",
						RealCareerDate:    "2022-01-01T00:00:00",
					},
				},
			},
			mockStatusCode: 200,
			expectedError:  nil,
			expectedData: response.EmployeeCMOResponse{
				EmployeeID:         "123",
				EmployeeName:       "John Doe",
				EmployeeIDWithName: "123 - John Doe",
				JoinDate:           "2022-01-01",
				PositionGroupCode:  "AO",
				PositionGroupName:  "Credit Marketing Officer",
				CMOCategory:        constant.CMO_LAMA,
			},
		},
		{
			name:           "error employee data timeout",
			employeeID:     "123",
			mockResponse:   response.GetEmployeeByID{},
			mockStatusCode: 504,
			expectedError:  fmt.Errorf(errorLib.ErrGatewayTimeout + " - Get employee data"),
			expectedData:   response.EmployeeCMOResponse{},
			mockError:      fmt.Errorf(errorLib.ErrGatewayTimeout + " - Get employee data"),
		},
		{
			name:           "error employee data not found",
			employeeID:     "123",
			mockResponse:   response.GetEmployeeByID{},
			mockStatusCode: 404,
			expectedError:  fmt.Errorf(errorLib.ErrBadRequest + " - Get employee data"),
			expectedData:   response.EmployeeCMOResponse{},
			mockError:      fmt.Errorf(errorLib.ErrBadRequest + " - Get employee data"),
		},
		{
			name:       "error real career date empty",
			employeeID: "123",
			mockResponse: response.GetEmployeeByID{
				Data: []response.EmployeeCareerHistory{
					{
						EmployeeID:        "123",
						EmployeeName:      "John Doe",
						PositionGroupCode: "AO",
						RealCareerDate:    "",
					},
				},
			},
			mockStatusCode: 200,
			expectedError:  fmt.Errorf(errorLib.ErrServiceUnavailable + " - Get employee data"),
			expectedData:   response.EmployeeCMOResponse{},
			mockError:      fmt.Errorf(errorLib.ErrServiceUnavailable + " - Get employee data"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockResponseBody, err := jsoniter.MarshalToString(tc.mockResponse)
			if err != nil {
				t.Fatalf("failed to marshal mock response: %v", err)
			}

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("HRIS_GET_EMPLOYEE_DATA_URL"), httpmock.NewStringResponder(tc.mockStatusCode, mockResponseBody))
			resp, _ := rst.R().Post(os.Getenv("HRIS_GET_EMPLOYEE_DATA_URL"))

			mockHttpClient.On("EngineAPI", mock.Anything, constant.DILEN_KMB_LOG, os.Getenv("HRIS_GET_EMPLOYEE_DATA_URL"), mock.Anything, map[string]string{"Authorization": "Bearer "}, constant.METHOD_POST, false, 0, timeout, "", "").Return(resp, tc.mockError).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			data, err := usecase.GetEmployeeData(context.Background(), tc.employeeID)

			require.Equal(t, tc.expectedError, err)
			require.Equal(t, tc.expectedData, data)
		})
	}
}

func TestGetFpdCMO(t *testing.T) {
	os.Setenv("AGREEMENT_LTV_FPD", "http://10.9.100.122:8181/api/v1/agreement/ltv-fpd")

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testCases := []struct {
		name           string
		cmoID          string
		bpkbNameType   string
		mockResponse   response.GetFPDCmoByID
		mockStatusCode int
		mockError      error
		expectedError  error
		expectedData   response.FpdCMOResponse
	}{
		{
			name:         "success fpd data for nama beda",
			cmoID:        "CMO01",
			bpkbNameType: "NAMA BEDA",
			mockResponse: response.GetFPDCmoByID{
				Data: []response.FpdData{
					{
						BpkbNameType: "NAMA BEDA",
						Fpd:          1.5,
						AccSales:     10,
					},
				},
			},
			mockStatusCode: 200,
			expectedError:  nil,
			expectedData: response.FpdCMOResponse{
				FpdExist:    true,
				CmoFpd:      1.5,
				CmoAccSales: 10,
			},
		},
		{
			name:         "success fpd data for nama sama",
			cmoID:        "CMO02",
			bpkbNameType: "NAMA SAMA",
			mockResponse: response.GetFPDCmoByID{
				Data: []response.FpdData{
					{
						BpkbNameType: "NAMA SAMA",
						Fpd:          2.0,
						AccSales:     15,
					},
				},
			},
			mockStatusCode: 200,
			expectedError:  nil,
			expectedData: response.FpdCMOResponse{
				FpdExist:    true,
				CmoFpd:      2.0,
				CmoAccSales: 15,
			},
		},
		{
			name:         "success fpd data empty nama beda nama sama",
			cmoID:        "CMO11",
			bpkbNameType: "NAMA BEDA",
			mockResponse: response.GetFPDCmoByID{
				Data: []response.FpdData{},
			},
			mockStatusCode: 200,
			expectedError:  nil,
			expectedData: response.FpdCMOResponse{
				FpdExist:    false,
				CmoFpd:      0,
				CmoAccSales: 0,
			},
		},
		{
			name:         "success fpd data bpkb name other",
			cmoID:        "CMO12",
			bpkbNameType: "NAMA AGAK LAEN",
			mockResponse: response.GetFPDCmoByID{
				Data: []response.FpdData{
					{
						BpkbNameType: "NAMA BEDA",
						Fpd:          1.5,
						AccSales:     10,
					},
					{
						BpkbNameType: "NAMA SAMA",
						Fpd:          2.0,
						AccSales:     15,
					},
				},
			},
			mockStatusCode: 200,
			expectedError:  nil,
			expectedData: response.FpdCMOResponse{
				FpdExist:    false,
				CmoFpd:      0,
				CmoAccSales: 0,
			},
		},
		{
			name:           "error fpd data timeout",
			cmoID:          "CMO03",
			bpkbNameType:   "NAMA BEDA",
			mockResponse:   response.GetFPDCmoByID{},
			mockStatusCode: 504,
			expectedError:  fmt.Errorf(errorLib.ErrGatewayTimeout + " - Get fpd data"),
			expectedData:   response.FpdCMOResponse{},
			mockError:      fmt.Errorf(errorLib.ErrGatewayTimeout + " - Get fpd data"),
		},
		{
			name:           "error fpd data not found",
			cmoID:          "CMO04",
			bpkbNameType:   "NAMA SAMA",
			mockResponse:   response.GetFPDCmoByID{},
			mockStatusCode: 404,
			expectedError:  fmt.Errorf(errorLib.ErrBadRequest + " - Get fpd data"),
			expectedData:   response.FpdCMOResponse{},
			mockError:      fmt.Errorf(errorLib.ErrBadRequest + " - Get fpd data"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockResponseBody, err := jsoniter.MarshalToString(tc.mockResponse)
			if err != nil {
				t.Fatalf("failed to marshal mock response: %v", err)
			}

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("AGREEMENT_LTV_FPD")+"?lob_id=2&cmo_id="+tc.cmoID, httpmock.NewStringResponder(tc.mockStatusCode, mockResponseBody))
			resp, _ := rst.R().SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": ""}).Get(os.Getenv("AGREEMENT_LTV_FPD") + "?lob_id=2&cmo_id=" + tc.cmoID)

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("AGREEMENT_LTV_FPD")+"?lob_id=2&cmo_id="+tc.cmoID, []byte(nil), map[string]string{"Authorization": ""}, constant.METHOD_GET, false, 0, timeout, "", "").Return(resp, tc.mockError).Once()
			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			data, err := usecase.GetFpdCMO(ctx, tc.cmoID, tc.bpkbNameType)

			require.Equal(t, tc.expectedError, err)
			require.Equal(t, tc.expectedData, data)
		})
	}
}

func TestCheckCmoNoFPD(t *testing.T) {
	testcases := []struct {
		name               string
		prospectID         string
		cmoID              string
		cmoCategory        string
		cmoJoinDate        string
		defaultCluster     string
		bpkbName           string
		mockReturnData     entity.TrxCmoNoFPD
		mockReturnError    error
		expectedCluster    string
		expectedEntitySave entity.TrxCmoNoFPD
		expectedError      error
	}{
		{
			name:            "test new cmo",
			prospectID:      "SAL0002",
			cmoID:           "CMO02",
			cmoCategory:     constant.CMO_BARU,
			cmoJoinDate:     "2024-05-28",
			defaultCluster:  "Cluster C",
			bpkbName:        "NAMA BEDA",
			mockReturnData:  entity.TrxCmoNoFPD{},
			expectedCluster: "",
			expectedEntitySave: entity.TrxCmoNoFPD{
				ProspectID:              "SAL0002",
				BPKBName:                "NAMA BEDA",
				CMOID:                   "CMO02",
				CmoCategory:             constant.CMO_BARU,
				CmoJoinDate:             "2024-05-28",
				DefaultCluster:          "Cluster C",
				DefaultClusterStartDate: "2024-05-28",
				DefaultClusterEndDate:   "2024-04-30",
			},
			expectedError: nil,
		},
		{
			name:               "test error in repository",
			prospectID:         "SAL0003",
			cmoID:              "CMO03",
			cmoCategory:        constant.CMO_LAMA,
			cmoJoinDate:        "2023-01-01",
			defaultCluster:     "Cluster B",
			bpkbName:           "NAMA SAMA",
			mockReturnData:     entity.TrxCmoNoFPD{},
			mockReturnError:    errors.New("repository error"),
			expectedCluster:    "",
			expectedEntitySave: entity.TrxCmoNoFPD{},
			expectedError:      errors.New("repository error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("CheckCMONoFPD", tc.cmoID, tc.bpkbName).Return(tc.mockReturnData, tc.mockReturnError)

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			clusterCMOSaved, entitySaveTrxNoFPd, err := usecase.CheckCmoNoFPD(tc.prospectID, tc.cmoID, tc.cmoCategory, tc.cmoJoinDate, tc.defaultCluster, tc.bpkbName)
			require.Equal(t, tc.expectedCluster, clusterCMOSaved)
			require.Equal(t, tc.expectedEntitySave, entitySaveTrxNoFPd)
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestPrinciplePemohon(t *testing.T) {
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name                       string
		request                    request.PrinciplePemohon
		resGetPrincipleStepOne     entity.TrxPrincipleStepOne
		errGetPrincipleStepOne     error
		resGetConfig               entity.AppConfig
		errGetConfig               error
		resBannedPMKOrDSR          response.UsecaseApi
		errBannedPMKOrDSR          error
		resRejection               response.UsecaseApi
		errRejection               error
		resDupcheckIntegrator      response.SpDupCekCustomerByID
		errDupcheckIntegrator      error
		resBlacklistCheck          response.UsecaseApi
		resCustomerKMB             string
		errCustomerKMB             error
		resCheckPMK                response.UsecaseApi
		errCheckPMK                error
		resGetEmployeeData         response.EmployeeCMOResponse
		errGetEmployeeData         error
		resGetFpdCMO               response.FpdCMOResponse
		errGetFpdCMO               error
		resMasterMappingFpdCluster entity.MasterMappingFpdCluster
		errMasterMappingFpdCluster error
		resClusterCheckCmoNoFPD    string
		resEntityCheckCmoNoFPD     entity.TrxCmoNoFPD
		errCheckCmoNoFPD           error
		resFilteringPefindo        response.Filtering
		resPefindo                 response.PefindoResult
		resDetailBiroPefindo       []entity.TrxDetailBiro
		errPefindo                 error
		resDukcapil                response.Ekyc
		errDukcapil                error
		resAsliri                  response.Ekyc
		errAsliri                  error
		resKtp                     response.Ekyc
		errKtp                     error
		errSave                    error
		errSavePrincipleStepTwo    error
		errUpdatePrincipleStepOne  error
		result                     response.UsecaseApi
		err                        error
		expectPublishEvent         bool
	}{
		{
			name: "error get principle step one",
			request: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				IDNumber:   "1234567890",
			},
			errGetPrincipleStepOne: errors.New("something wrong"),
			err:                    errors.New("something wrong"),
		},
		{
			name: "error get config",
			request: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				IDNumber:   "1234567890",
			},
			errGetConfig: errors.New("something wrong"),
			err:          errors.New("something wrong"),
		},
		{
			name: "error banned pmk dsr",
			request: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				IDNumber:   "1234567890",
			},
			errBannedPMKOrDSR: errors.New("something wrong"),
			err:               errors.New("something wrong"),
		},
		{
			name: "reject banned pmk dsr",
			request: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				IDNumber:   "1234567890",
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_PERNAH_REJECT_PMK_DSR,
				Reason: "Data diri tidak lolos verifikasi",
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_PERNAH_REJECT_PMK_DSR,
				Reason: "Data diri tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error rejection",
			request: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				IDNumber:   "1234567890",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errRejection: errors.New("something wrong"),
			err:          errors.New("something wrong"),
		},
		{
			name: "reject rejection",
			request: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				IDNumber:   "1234567890",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_PERNAH_REJECT_PMK_DSR,
				Reason: "Data diri tidak lolos verifikasi",
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_PERNAH_REJECT_PMK_DSR,
				Reason: "Data diri tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error dupcheck integrator",
			request: request.PrinciplePemohon{
				ProspectID:     "SAL-123",
				IDNumber:       "1234567890",
				MaritalStatus:  constant.MARRIED,
				SpouseIDNumber: "987654321",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errDupcheckIntegrator: errors.New("something wrong"),
			err:                   errors.New("something wrong"),
		},
		{
			name: "reject check blacklist",
			request: request.PrinciplePemohon{
				ProspectID:     "SAL-123",
				IDNumber:       "1234567890",
				MaritalStatus:  constant.MARRIED,
				SpouseIDNumber: "987654321",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_KONSUMEN_SIMILIAR,
				Reason: "Data diri tidak lolos verifikasi",
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_KONSUMEN_SIMILIAR,
				Reason: "Data diri tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error customer kmb",
			request: request.PrinciplePemohon{
				ProspectID:     "SAL-123",
				IDNumber:       "1234567890",
				MaritalStatus:  constant.MARRIED,
				SpouseIDNumber: "987654321",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errCustomerKMB: errors.New("something wrong"),
			err:            errors.New("something wrong"),
		},
		{
			name: "error get employee data",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_REJECT_INCOME,
				Reason: fmt.Sprintf(" %s", constant.REASON_REJECT_INCOME),
			},
			errGetEmployeeData: errors.New("something wrong"),
			err:                errors.New("something wrong"),
		},
		{
			name: "error cmo category",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: "",
			},
			expectPublishEvent: true,
		},
		{
			name: "error get fpd cmo",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			errGetFpdCMO: errors.New("something wrong"),
			err:          errors.New("something wrong"),
		},
		{
			name: "error master mapping fpd cluster",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: true,
			},
			errMasterMappingFpdCluster: errors.New("something wrong"),
			err:                        errors.New("something wrong"),
		},
		{
			name: "error check cmo no fpd",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			errCheckCmoNoFPD: errors.New("something wrong"),
			err:              errors.New("something wrong"),
		},
		{
			name: "error pefindo",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID: "123",
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			resClusterCheckCmoNoFPD: "Cluster A",
			resEntityCheckCmoNoFPD: entity.TrxCmoNoFPD{
				CmoCategory: constant.CMO_LAMA,
			},
			errPefindo: errors.New("something wrong"),
			err:        errors.New("something wrong"),
		},
		{
			name: "error dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID:      "123",
				CustomerStatus:  constant.STATUS_KONSUMEN_AO,
				CustomerSegment: constant.RO_AO_PRIME,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			resClusterCheckCmoNoFPD: "Cluster A",
			resEntityCheckCmoNoFPD: entity.TrxCmoNoFPD{
				CmoCategory: constant.CMO_LAMA,
			},
			errDukcapil: errors.New("something wrong"),
			err:         errors.New("something wrong"),
		},
		{
			name: "error asliri ktp",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID:      "123",
				CustomerStatus:  constant.STATUS_KONSUMEN_AO,
				CustomerSegment: constant.RO_AO_PRIME,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			resClusterCheckCmoNoFPD: "Cluster A",
			resEntityCheckCmoNoFPD: entity.TrxCmoNoFPD{
				CmoCategory: constant.CMO_LAMA,
			},
			resPefindo: response.PefindoResult{
				Score:                         "HIGH RISK",
				Category:                      3,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             10,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: errors.New(fmt.Sprintf("%s - Dukcapil", constant.TYPE_CONTINGENCY)),
			errAsliri:   errors.New("something wrong"),
			errKtp:      errors.New("something wrong"),
			err:         errors.New("something wrong"),
		},
		{
			name: "error save",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:   "P",
				OwnerAsset: "Test Legal Name",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID:      "123",
				CustomerStatus:  constant.STATUS_KONSUMEN_AO,
				CustomerSegment: constant.RO_AO_PRIME,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			resClusterCheckCmoNoFPD: "Cluster A",
			resEntityCheckCmoNoFPD: entity.TrxCmoNoFPD{
				CmoCategory: constant.CMO_LAMA,
			},
			resFilteringPefindo: response.Filtering{
				NextProcess: true,
			},
			resDukcapil: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			errSave: errors.New("something wrong"),
			err:     errors.New("something wrong"),
		},
		{
			name: "error save principle step two",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:   "P",
				OwnerAsset: "Test Legal Name",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID:      "123",
				CustomerStatus:  constant.STATUS_KONSUMEN_AO,
				CustomerSegment: constant.RO_AO_PRIME,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			resClusterCheckCmoNoFPD: "Cluster A",
			resEntityCheckCmoNoFPD: entity.TrxCmoNoFPD{
				CmoCategory: constant.CMO_LAMA,
			},
			resFilteringPefindo: response.Filtering{
				NextProcess: true,
			},
			resDukcapil: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			errSavePrincipleStepTwo: errors.New("something wrong"),
			err:                     errors.New("something wrong"),
		},
		{
			name: "error update principle step one",
			request: request.PrinciplePemohon{
				ProspectID:              "SAL-123",
				IDNumber:                "1234567890",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "987654321",
				SpouseLegalName:         "Test Legal Name",
				SpouseBirthDate:         "2000-11-11",
				SpouseSurgateMotherName: "Test Mother Name",
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:   "P",
				OwnerAsset: "Test Legal Name",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resBannedPMKOrDSR: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resRejection: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resDupcheckIntegrator: response.SpDupCekCustomerByID{
				CustomerID:      "123",
				CustomerStatus:  constant.STATUS_KONSUMEN_AO,
				CustomerSegment: constant.RO_AO_PRIME,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_NEW,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			resClusterCheckCmoNoFPD: "Cluster A",
			resEntityCheckCmoNoFPD: entity.TrxCmoNoFPD{
				CmoCategory: constant.CMO_LAMA,
			},
			resFilteringPefindo: response.Filtering{
				NextProcess: true,
			},
			resDukcapil: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			errUpdatePrincipleStepOne: errors.New("something wrong"),
			err:                       errors.New("something wrong"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			mockRepository.On("GetPrincipleStepOne", tc.request.ProspectID).Return(tc.resGetPrincipleStepOne, tc.errGetPrincipleStepOne)
			mockRepository.On("GetConfig", "dupcheck", constant.LOB_KMB_OFF, "dupcheck_kmb_config").Return(tc.resGetConfig, tc.errGetConfig)
			mockRepository.On("MasterMappingFpdCluster", mock.Anything).Return(tc.resMasterMappingFpdCluster, tc.errMasterMappingFpdCluster)
			mockRepository.On("SavePrincipleStepTwo", mock.AnythingOfType("entity.TrxPrincipleStepTwo")).Return(tc.errSavePrincipleStepTwo)
			mockRepository.On("UpdatePrincipleStepOne", tc.request.ProspectID, mock.AnythingOfType("entity.TrxPrincipleStepOne")).Return(tc.errUpdatePrincipleStepOne)

			mockUsecase := new(mocks.Usecase)
			mockRepository.On("GetEncB64", tc.request.IDNumber).Return(entity.EncryptedString{MyString: "encryted"}, nil)
			mockUsecase.On("BannedPMKOrDSR", mock.Anything).Return(tc.resBannedPMKOrDSR, tc.errBannedPMKOrDSR)
			mockUsecase.On("Rejection", tc.request.ProspectID, mock.Anything, mock.Anything).Return(tc.resRejection, entity.TrxBannedPMKDSR{}, tc.errRejection)
			mockUsecase.On("DupcheckIntegrator", ctx, tc.request.ProspectID, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "").Return(tc.resDupcheckIntegrator, tc.errDupcheckIntegrator)
			mockUsecase.On("BlacklistCheck", mock.Anything, tc.resDupcheckIntegrator).Return(tc.resBlacklistCheck, mock.Anything)
			if tc.resBlacklistCheck.Result == constant.DECISION_REJECT {
				mockUsecase.On("Save", mock.AnythingOfType("entity.FilteringKMB"), mock.AnythingOfType("[]entity.TrxDetailBiro"), mock.AnythingOfType("entity.TrxCmoNoFPD")).Return(nil)
			}
			mockUsecase.On("CustomerKMB", mock.AnythingOfType("response.SpDupCekCustomerByID")).Return(tc.resCustomerKMB, tc.errCustomerKMB)
			mockUsecase.On("CheckPMK", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resCheckPMK, tc.errCheckPMK)
			mockUsecase.On("GetEmployeeData", ctx, mock.Anything).Return(tc.resGetEmployeeData, tc.errGetEmployeeData)
			mockUsecase.On("GetFpdCMO", ctx, mock.Anything, mock.Anything).Return(tc.resGetFpdCMO, tc.errGetFpdCMO)
			mockUsecase.On("CheckCmoNoFPD", tc.request.ProspectID, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resClusterCheckCmoNoFPD, tc.resEntityCheckCmoNoFPD, tc.errCheckCmoNoFPD)
			mockUsecase.On("Pefindo", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resFilteringPefindo, tc.resPefindo, tc.resDetailBiroPefindo, tc.errPefindo)
			mockUsecase.On("Dukcapil", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resDukcapil, tc.errDukcapil)
			mockUsecase.On("Asliri", ctx, mock.Anything, mock.Anything).Return(tc.resAsliri, tc.errAsliri)
			mockUsecase.On("Ktp", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resKtp, tc.errKtp)
			mockUsecase.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(tc.errSave)

			if tc.expectPublishEvent {
				mockPlatformEvent.On("PublishEvent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}

			multiUsecase := NewMultiUsecase(mockRepository, mockHttpClient, platformEvent, mockUsecase)

			result, err := multiUsecase.PrinciplePemohon(ctx, tc.request)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.result, result)
			}

			time.Sleep(100 * time.Millisecond)
		})
	}
}

func TestSave(t *testing.T) {
	testcases := []struct {
		name                string
		transaction         entity.FilteringKMB
		trxDetailBiro       []entity.TrxDetailBiro
		transactionCMOnoFPD entity.TrxCmoNoFPD
		errSaveFiltering    error
		expectedError       error
	}{
		{
			name: "success",
			transaction: entity.FilteringKMB{
				ProspectID: "SAL-001",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{ProspectID: "SAL-001"},
			},
			transactionCMOnoFPD: entity.TrxCmoNoFPD{
				ProspectID: "SAL-001",
			},
			errSaveFiltering: nil,
			expectedError:    nil,
		},
		{
			name: "error deadline exceeded",
			transaction: entity.FilteringKMB{
				ProspectID: "SAL-002",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{ProspectID: "SAL-002"},
			},
			transactionCMOnoFPD: entity.TrxCmoNoFPD{
				ProspectID: "SAL-002",
			},
			errSaveFiltering: errors.New("context deadline exceeded"),
			expectedError:    errors.New("context deadline exceeded"),
		},
		{
			name: "error prospect id already exists",
			transaction: entity.FilteringKMB{
				ProspectID: "SAL-003",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{ProspectID: "SAL-003"},
			},
			transactionCMOnoFPD: entity.TrxCmoNoFPD{
				ProspectID: "SAL-003",
			},
			errSaveFiltering: errors.New("duplicate key value violates unique constraint"),
			expectedError:    errors.New("duplicate key value violates unique constraint"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("SaveFiltering", tc.transaction, tc.trxDetailBiro, tc.transactionCMOnoFPD).Return(tc.errSaveFiltering).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			err := usecase.Save(tc.transaction, tc.trxDetailBiro, tc.transactionCMOnoFPD)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestCheckLatestPaidInstallment(t *testing.T) {
	os.Setenv("LASTEST_PAID_INSTALLMENT_URL", "http://api.example.com/latest-paid-installment/")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
	accessToken := "access-token"

	testCases := []struct {
		name               string
		prospectID         string
		customerID         string
		mockResponse       response.LatestPaidInstallmentData
		mockStatusCode     int
		mockError          error
		expectedRrdDate    string
		expectedMonthsDiff int
		expectedError      error
	}{
		{
			name:               "success",
			prospectID:         "SAL-123",
			customerID:         "12345",
			mockResponse:       response.LatestPaidInstallmentData{RRDDate: time.Now().AddDate(0, -3, 0).Format(time.RFC3339)},
			mockStatusCode:     200,
			mockError:          nil,
			expectedRrdDate:    time.Now().AddDate(0, -3, 0).Format("2006-01-02"),
			expectedMonthsDiff: 3,
			expectedError:      nil,
		},
		{
			name:               "error empty rrd date",
			prospectID:         "SAL-123",
			customerID:         "12345",
			mockResponse:       response.LatestPaidInstallmentData{RRDDate: ""},
			mockStatusCode:     200,
			mockError:          nil,
			expectedRrdDate:    "",
			expectedMonthsDiff: 0,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - Result LatestPaidInstallmentData rrd_date Empty String"),
		},
		{
			name:               "error invalid rrd date format",
			prospectID:         "SAL-123",
			customerID:         "12345",
			mockResponse:       response.LatestPaidInstallmentData{RRDDate: "invalid-date"},
			mockStatusCode:     200,
			mockError:          nil,
			expectedRrdDate:    "",
			expectedMonthsDiff: 0,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of response rrd_date (invalid-date)"),
		},
		{
			name:               "error api timeout",
			prospectID:         "SAL-123",
			customerID:         "12345",
			mockResponse:       response.LatestPaidInstallmentData{},
			mockStatusCode:     504,
			mockError:          errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call LatestPaidInstallmentData Timeout"),
			expectedRrdDate:    "",
			expectedMonthsDiff: 0,
			expectedError:      errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call LatestPaidInstallmentData Timeout"),
		},
		{
			name:               "error api error",
			prospectID:         "SAL-123",
			customerID:         "12345",
			mockResponse:       response.LatestPaidInstallmentData{},
			mockStatusCode:     500,
			mockError:          nil,
			expectedRrdDate:    "",
			expectedMonthsDiff: 0,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - Call LatestPaidInstallmentData Error"),
		},
		{
			name:               "error negative months difference",
			prospectID:         "SAL-123",
			customerID:         "12345",
			mockResponse:       response.LatestPaidInstallmentData{RRDDate: time.Now().AddDate(0, 1, 0).Format(time.RFC3339)},
			mockStatusCode:     200,
			mockError:          nil,
			expectedRrdDate:    "",
			expectedMonthsDiff: 0,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - Difference of months rrd_date and current_date is negative (-)"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockResponseBody, err := jsoniter.MarshalToString(map[string]interface{}{"data": tc.mockResponse})
			if err != nil {
				t.Fatalf("failed to marshal mock response: %v", err)
			}

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("LASTEST_PAID_INSTALLMENT_URL")+tc.customerID+"/2", httpmock.NewStringResponder(tc.mockStatusCode, mockResponseBody))
			resp, _ := rst.R().SetHeaders(map[string]string{"Content-Type": "application/json", "Authorization": accessToken}).Get(os.Getenv("LASTEST_PAID_INSTALLMENT_URL") + tc.customerID + "/2")

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("LASTEST_PAID_INSTALLMENT_URL")+tc.customerID+"/2", []byte(nil), map[string]string{}, constant.METHOD_GET, false, 0, timeout, tc.prospectID, accessToken).Return(resp, tc.mockError).Once()
			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			rrdDate, monthsDiff, err := usecase.CheckLatestPaidInstallment(ctx, tc.prospectID, tc.customerID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedRrdDate, rrdDate)
				require.Equal(t, tc.expectedMonthsDiff, monthsDiff)
			}
		})
	}
}
