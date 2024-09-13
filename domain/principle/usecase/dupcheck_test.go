package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
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

func TestBlacklistCheck(t *testing.T) {
	testcases := []struct {
		name         string
		spCustomer   response.SpDupCekCustomerByID
		index        int
		customerType string
		res          response.UsecaseApi
	}{
		{
			name:         "test new",
			customerType: constant.MESSAGE_BERSIH,
			res: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_NON_BLACKLIST_ALL,
				Reason:         constant.REASON_NON_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
		},
		{
			name: "test ro is similar",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				BadType:               constant.BADTYPE_B,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
				TotalInstallment:      0,
				RRDDate:               "2024-06-02",
				IsSimiliar:            1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_SIMILIAR,
				Reason:         constant.REASON_KONSUMEN_SIMILIAR,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name: "test ro BadType",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				BadType:               constant.BADTYPE_B,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
				TotalInstallment:      0,
				RRDDate:               "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name:  "test ro pasangan BadType",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				BadType:               constant.BADTYPE_B,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
				TotalInstallment:      0,
				RRDDate:               "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name: "test ro MaxOverdueDays",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
				TotalInstallment:      0,
				RRDDate:               "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST_OVD_90DAYS,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name:  "test ro pasangan MaxOverdueDays",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
				TotalInstallment:      0,
				RRDDate:               "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST_OVD_90DAYS,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name: "test ro NumOfAssetInventoried",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
				TotalInstallment:      0,
				RRDDate:               "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST_ASSET_INVENTORY,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name:  "test ro pasangan NumOfAssetInventoried",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
				TotalInstallment:      0,
				RRDDate:               "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST_ASSET_INVENTORY,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name: "test ro IsRestructure",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:   constant.STATUS_KONSUMEN_RO,
				IsRestructure:    1,
				TotalInstallment: 0,
				RRDDate:          "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST_RESTRUCTURE,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name:  "test ro pasangan IsRestructure",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:   constant.STATUS_KONSUMEN_RO,
				IsRestructure:    1,
				TotalInstallment: 0,
				RRDDate:          "2024-06-02",
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST_RESTRUCTURE,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
			},
		},
		{
			name: "test BadType W",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_AO,
				BadType:        constant.BADTYPE_W,
			},
			customerType: constant.MESSAGE_WARNING,
			res: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_NON_BLACKLIST_ALL,
				Reason:         constant.REASON_NON_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_AO,
			},
		},
		{
			name:         "test new sp",
			spCustomer:   response.SpDupCekCustomerByID{},
			customerType: constant.MESSAGE_BERSIH,
			res: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_NON_BLACKLIST_ALL,
				Reason:         constant.REASON_NON_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, ct := usecase.BlacklistCheck(tc.index, tc.spCustomer)
			require.Equal(t, tc.res, result)
			require.Equal(t, tc.customerType, ct)
		})
	}
}

func TestDupcheckIntegrator(t *testing.T) {
	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	accessToken := "test_token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	ProspectID := "SAL-123456"
	IDNumber := "1234567890"
	LegalName := "John Doe"
	BirthDate := "1990-01-01"
	MotherName := "Jane Doe"

	testcases := []struct {
		name          string
		rDupcheckCode int
		rDupcheckBody string
		errHttp       error
		resDupcheck   response.SpDupCekCustomerByID
		errFinal      error
	}{
		{
			name:     "test HTTP error",
			errHttp:  errors.New("upstream_service_timeout - Call Dupcheck Timeout"),
			errFinal: errors.New("upstream_service_timeout - Call Dupcheck Timeout"),
		},
		{
			name:          "test non-200 response",
			rDupcheckCode: 400,
			errFinal:      errors.New("upstream_service_error - Call Dupcheck Error"),
		},
		{
			name:          "test successful response",
			rDupcheckCode: 200,
			rDupcheckBody: `{
				"data": {
					"customer_id": "123",
					"id_number": "1234567890",
					"full_name": "John Doe",
					"birth_date": "1990-01-01",
					"surgate_mother_name": "Jane Doe",
					"customer_status": "NEW",
					"bad_type": "A",
					"max_overduedays": 0,
					"num_of_asset_inventoried": 0,
					"is_restructure": 0,
					"is_similiar": 0
				}
			}`,
			resDupcheck: response.SpDupCekCustomerByID{
				CustomerID:            "123",
				IDNumber:              "1234567890",
				FullName:              "John Doe",
				BirthDate:             "1990-01-01",
				SurgateMotherName:     "Jane Doe",
				CustomerStatus:        "NEW",
				BadType:               "A",
				MaxOverdueDays:        0,
				NumOfAssetInventoried: 0,
				IsRestructure:         0,
				IsSimiliar:            0,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("DUPCHECK_URL"),
				httpmock.NewStringResponder(tc.rDupcheckCode, tc.rDupcheckBody))
			resp, _ := rst.R().Post(os.Getenv("DUPCHECK_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("DUPCHECK_URL"),
				mock.Anything, map[string]string{}, constant.METHOD_POST, false, 0, timeout,
				ProspectID, accessToken).Return(resp, tc.errHttp).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.DupcheckIntegrator(ctx, ProspectID, IDNumber, LegalName, BirthDate, MotherName, accessToken)
			require.Equal(t, tc.resDupcheck, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestBannedPMKOrDSR(t *testing.T) {
	testcases := []struct {
		name           string
		encrypted      string
		mockTrxBanned  entity.TrxBannedPMKDSR
		mockError      error
		expectedResult response.UsecaseApi
		expectedError  error
	}{
		{
			name:          "test error from repository",
			encrypted:     "encrypted_data",
			mockError:     errors.New("repository error"),
			expectedError: errors.New("repository error"),
		},
		{
			name:           "test no banned record",
			encrypted:      "encrypted_data",
			mockTrxBanned:  entity.TrxBannedPMKDSR{},
			expectedResult: response.UsecaseApi{},
		},
		{
			name:      "test banned record found",
			encrypted: "encrypted_data",
			mockTrxBanned: entity.TrxBannedPMKDSR{
				ProspectID: "SAL-123",
				IDNumber:   "123",
			},
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PERNAH_REJECT_PMK_DSR,
				Reason:         constant.REASON_PERNAH_REJECT_PMK_DSR,
				SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockRepository.On("GetBannedPMKDSR", tc.encrypted).Return(tc.mockTrxBanned, tc.mockError)

			usecase := NewUsecase(mockRepository, nil)

			result, err := usecase.BannedPMKOrDSR(tc.encrypted)

			require.Equal(t, tc.expectedResult, result)
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestCustomerKMB(t *testing.T) {
	testcases := []struct {
		name           string
		spDupcheck     response.SpDupCekCustomerByID
		expectedStatus string
		expectedError  error
	}{
		{
			name:           "test new customer",
			spDupcheck:     response.SpDupCekCustomerByID{},
			expectedStatus: constant.STATUS_KONSUMEN_NEW,
		},
		{
			name: "test RO customer case 1",
			spDupcheck: response.SpDupCekCustomerByID{
				TotalInstallment: 0,
				RRDDate:          &[]string{"2023-01-01"}[0],
			},
			expectedStatus: constant.STATUS_KONSUMEN_RO,
		},
		{
			name: "test RO customer case 2",
			spDupcheck: response.SpDupCekCustomerByID{
				TotalInstallment: 1,
				RRDDate:          &[]string{"2023-01-01"}[0],
			},
			expectedStatus: constant.STATUS_KONSUMEN_RO,
		},
		{
			name: "test AO customer",
			spDupcheck: response.SpDupCekCustomerByID{
				TotalInstallment: 1,
			},
			expectedStatus: constant.STATUS_KONSUMEN_AO,
		},
		{
			name: "test new customer case 2",
			spDupcheck: response.SpDupCekCustomerByID{
				TotalInstallment: -1,
				RRDDate:          nil,
			},
			expectedStatus: constant.STATUS_KONSUMEN_NEW,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			usecase := NewUsecase(nil, nil)

			status, err := usecase.CustomerKMB(tc.spDupcheck)

			require.Equal(t, tc.expectedStatus, status)
			require.Equal(t, tc.expectedError, err)
		})
	}
}

func TestRejection(t *testing.T) {
	testcases := []struct {
		name           string
		prospectID     string
		encrypted      string
		configValue    response.DupcheckConfig
		mockTrxReject  entity.TrxReject
		mockError      error
		expectedResult response.UsecaseApi
		expectedBanned entity.TrxBannedPMKDSR
		expectedError  error
	}{
		{
			name:          "test error from repository",
			prospectID:    "SAL-123",
			encrypted:     "encrypted_data",
			mockError:     errors.New("repository error"),
			expectedError: errors.New("repository error"),
		},
		{
			name:       "test PMK DSR rejection",
			prospectID: "SAL-123",
			encrypted:  "encrypted_data",
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					AttemptPMKDSR: 3,
				},
			},
			mockTrxReject: entity.TrxReject{RejectPMKDSR: 2, RejectNIK: 1},
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PERNAH_REJECT_PMK_DSR,
				Reason:         constant.REASON_PERNAH_REJECT_PMK_DSR,
				SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
			},
			expectedBanned: entity.TrxBannedPMKDSR{ProspectID: "SAL-123", IDNumber: "encrypted_data"},
		},
		{
			name:       "test NIK rejection",
			prospectID: "SAL-123",
			encrypted:  "encrypted_data",
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					AttemptPMKDSR: 3,
					AttemptNIK:    3,
				},
			},
			mockTrxReject: entity.TrxReject{RejectNIK: 3},
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PERNAH_REJECT_NIK,
				Reason:         constant.REASON_PERNAH_REJECT_NIK,
				SourceDecision: constant.SOURCE_DECISION_NIK,
			},
		},
		{
			name:       "test no rejection",
			prospectID: "SAL-123",
			encrypted:  "encrypted_data",
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					AttemptPMKDSR: 3,
					AttemptNIK:    3,
				},
			},
			mockTrxReject: entity.TrxReject{RejectPMKDSR: 1, RejectNIK: 1},
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_BELUM_PERNAH_REJECT,
				Reason:         constant.REASON_BELUM_PERNAH_REJECT,
				SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockRepository.On("GetRejection", tc.encrypted).Return(tc.mockTrxReject, tc.mockError)

			usecase := NewUsecase(mockRepository, nil)

			result, banned, err := usecase.Rejection(tc.prospectID, tc.encrypted, tc.configValue)

			require.Equal(t, tc.expectedResult, result)
			require.Equal(t, tc.expectedBanned, banned)
			require.Equal(t, tc.expectedError, err)
		})
	}
}
