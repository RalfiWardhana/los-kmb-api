package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func stringPtr(s string) *string {
	return &s
}

func TestCompareNumOfDependence(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        *int
		expected bool
	}{
		{
			name:     "both values equal",
			a:        5,
			b:        intPtr(5),
			expected: true,
		},
		{
			name:     "b is nil and a is 0",
			a:        0,
			b:        nil,
			expected: true,
		},
		{
			name:     "b is nil and a is not 0",
			a:        5,
			b:        nil,
			expected: false,
		},
		{
			name:     "values not equal",
			a:        5,
			b:        intPtr(6),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareNumOfDependence(tt.a, tt.b)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareMonthlyVariableIncome(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        *float64
		expected bool
	}{
		{
			name:     "both values equal",
			a:        1000.0,
			b:        float64Ptr(1000.0),
			expected: true,
		},
		{
			name:     "b is nil and a is 0",
			a:        0,
			b:        nil,
			expected: true,
		},
		{
			name:     "b is nil and a is not 0",
			a:        1000.0,
			b:        nil,
			expected: false,
		},
		{
			name:     "values not equal",
			a:        1000.0,
			b:        float64Ptr(2000.0),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareMonthlyVariableIncome(tt.a, tt.b)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareSpouseIncome(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        *float64
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "a nil and b is 0",
			a:        nil,
			b:        float64Ptr(0),
			expected: true,
		},
		{
			name:     "float64 values equal",
			a:        float64(1000.0),
			b:        float64Ptr(1000.0),
			expected: true,
		},
		{
			name:     "int to float64 comparison",
			a:        int(1000),
			b:        float64Ptr(1000.0),
			expected: true,
		},
		{
			name:     "int64 to float64 comparison",
			a:        int64(1000),
			b:        float64Ptr(1000.0),
			expected: true,
		},
		{
			name:     "float32 to float64 comparison",
			a:        float32(1000.0),
			b:        float64Ptr(1000.0),
			expected: true,
		},
		{
			name:     "invalid type comparison",
			a:        "1000",
			b:        float64Ptr(1000.0),
			expected: false,
		},
		{
			name:     "value mismatch",
			a:        float64(1000.0),
			b:        float64Ptr(2000.0),
			expected: false,
		},
		{
			name:     "a has value but b is nil",
			a:        float64(1000.0),
			b:        nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareSpouseIncome(tt.a, tt.b)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestSubmission2Wilen(t *testing.T) {
	os.Setenv("BIRO_VALID_DAYS", "4")
	os.Setenv("NTF_PENDING_URL", "http://localhost/")
	os.Setenv("AGREEMENT_OF_CHASSIS_NUMBER_URL", "http://localhost/")
	os.Setenv("INTERNAL_RECORD_URL", "http://localhost/")
	os.Setenv("NAMA_SAMA", "SAMA")
	ctx := context.Background()

	testcases := []struct {
		name                      string
		reqMetrics                request.Metrics
		trxMaster                 int
		errScanTrxMaster          error
		filtering                 entity.FilteringKMB
		errGetFilteringResult     error
		errGetElaborateLtv        error
		trxKPM                    entity.TrxKPM
		errGetTrxKPM              error
		mappingCluster            entity.MasterMappingCluster
		errMappingCluster         error
		codeGetNTFPending         int
		bodyGetNTFPending         string
		errGetNTFPending          error
		codeGetNTFTopUp           int
		bodyGetNTFTopUp           string
		errGetNTFTopUp            error
		codeInternalRecord        int
		bodyInternalRecord        string
		errInternalRecord         error
		errSaveTransaction        error
		resultMetrics             interface{}
		err                       error
		expectGetNTFPending       bool
		expectGetNTFTopUp         bool
		expectGetInternalRecord   bool
		trxPrescreening           entity.TrxPrescreening
		trxFMF                    response.TrxFMF
		prescreeningDetail        entity.TrxDetail
		errPrescreening           error
		shouldCallSaveTransaction bool
		shouldCallPrescreening    bool
	}{
		{
			name: "error scan trx master",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			errScanTrxMaster:          errors.New("db error"),
			err:                       errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error"),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    false,
		},
		{
			name: "prospect id already exists",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster:                 1,
			err:                       errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID Already Exist"),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    false,
		},
		{
			name: "error get filtering result - not found",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			errGetFilteringResult:     errors.New(constant.RECORD_NOT_FOUND),
			err:                       errors.New(fmt.Sprintf("%s - Belum melakukan filtering atau hasil filtering sudah lebih dari %s hari", constant.ERROR_BAD_REQUEST, os.Getenv("BIRO_VALID_DAYS"))),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    false,
		},
		{
			name: "error get filtering result - other error",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			errGetFilteringResult:     errors.New("db error"),
			err:                       errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    false,
		},
		{
			name: "cannot proceed - next process not 1",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			filtering: entity.FilteringKMB{
				NextProcess: 0,
			},
			err:                       errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses"),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    false,
		},
		{
			name: "error get elaborate ltv",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			filtering: entity.FilteringKMB{
				NextProcess: 1,
			},
			errGetElaborateLtv:        errors.New("db error"),
			err:                       errors.New(constant.ERROR_BAD_REQUEST + " - Belum melakukan pengecekan LTV"),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    false,
		},
		{
			name: "error get trx kpm",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			filtering: entity.FilteringKMB{
				NextProcess: 1,
			},
			errGetTrxKPM:              errors.New("db error"),
			err:                       errors.New(constant.ERROR_BAD_REQUEST + " - Get Trx KPM 2Wilen Error"),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    false,
		},
		{
			name: "error unmarshall DSR FMF PBK info",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				CustomerPersonal: request.CustomerPersonal{
					Education:     "SMA",
					MaritalStatus: "M",
				},
				Agent: request.Agent{
					CmoRecom: constant.CMO_AGENT,
				},
			},
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			trxKPM: entity.TrxKPM{
				Education:          "SMA",
				DupcheckData:       `{"status_konsumen":"NEW"}`,
				CheckDSRFMFPBKInfo: `invalid json`,
			},
			err:                       errors.New(constant.ERROR_UPSTREAM + " - Unmarshal DSR FMF PBK Info Error"),
			shouldCallSaveTransaction: false,
			shouldCallPrescreening:    true,
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_APR,
				Reason:     "Valid",
			},
		},
		{
			name: "success flow - modified data with cmo not recommended",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				CustomerPersonal: request.CustomerPersonal{
					Education: "S1",
				},
				Agent: request.Agent{
					CmoRecom: constant.CMO_NOT_RECOMMEDED,
				},
			},
			filtering: entity.FilteringKMB{
				NextProcess: 1,
			},
			trxKPM: entity.TrxKPM{
				Education: "SMA",
			},
			resultMetrics:             response.Metrics{},
			shouldCallSaveTransaction: true,
			shouldCallPrescreening:    false,
		},
		{
			name: "success flow - not modified data (all fields match)",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
					BranchID:   "B001",
				},
				CustomerPersonal: request.CustomerPersonal{
					Education:       "SMA",
					NumOfDependence: intPtr(2),
					HomeStatus:      "OWN",
					StaySinceYear:   "2020",
					StaySinceMonth:  "1",
					MaritalStatus:   "M",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:          "123",
					LegalName:         "TEST SPOUSE",
					BirthDate:         "2015-01-01",
					SurgateMotherName: "MAMA SPOUSE",
				},
				CustomerEmployment: request.CustomerEmployment{
					ProfessionID:          "PRO-1",
					JobType:               "PERMANENT",
					JobPosition:           "STAFF",
					MonthlyFixedIncome:    5000000,
					MonthlyVariableIncome: float64Ptr(0),
					SpouseIncome:          float64Ptr(1000000),
				},
				Item: request.Item{
					AssetUsage:      "PR",
					CategoryID:      "CAT-1",
					AssetCode:       "ASSET-1",
					ManufactureYear: "2020",
					NoChassis:       "CHASSIS-1",
					BPKBName:        "OWNER-1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
			},
			filtering: entity.FilteringKMB{
				NextProcess:     1,
				CustomerStatus:  "RO",
				CustomerID:      "123456",
				CustomerSegment: "PRIME",
			},
			trxKPM: entity.TrxKPM{
				Education:            "SMA",
				NumOfDependence:      2,
				HomeStatus:           "OWN",
				StaySinceYear:        2020,
				StaySinceMonth:       1,
				ProfessionID:         "PRO-1",
				JobType:              "PERMANENT",
				JobPosition:          "STAFF",
				MonthlyFixedIncome:   5000000,
				SpouseIncome:         float64(1000000),
				AssetUsageTypeCode:   "PR",
				AssetCategoryID:      "CAT-1",
				AssetCode:            "ASSET-1",
				ManufactureYear:      "2020",
				NoChassis:            "CHASSIS-1",
				BPKBName:             "OWNER-1",
				DupcheckData:         `{"status_konsumen":"NEW", "reason":"TEST", "dsr":80, "customer_id":"123456"}`,
				ScoreProResult:       "PASS",
				ScoreProScoreResult:  700,
				CheckEkycCode:        "PASS",
				CheckEkycReason:      "Valid",
				CheckEkycInfo:        "Success",
				CheckEkycSource:      "KTP",
				CheckDSRFMFPBKCode:   "PASS",
				CheckDSRFMFPBKInfo:   `{"dsr_pbk":80,"total_dsr":90,"latest_installment_amount":1000000}`,
				CheckDSRFMFPBKReason: "Valid DSR",
				CreatedAt:            time.Now(),
				CheckBlacklistCode:   "PASS",
				CheckBlacklistReason: "Not blacklisted",
				CheckVehicleCode:     "PASS",
				CheckVehicleReason:   "Valid vehicle",
				CheckVehicleInfo:     "Vehicle info ok",
				CheckPMKCode:         "PASS",
				CheckPMKReason:       "Valid PMK",
				CheckDSRCode:         "PASS",
				CheckDSRReason:       "Valid DSR",
				RuleCode:             "PASS",
				Reason:               "All checks passed",
				FilteringCode:        "PASS",
				FilteringReason:      "Valid filtering",
				ScoreProCode:         "PASS",
				ScoreProReason:       "Good score",
				ScoreProInfo:         "Score details",
			},
			codeGetNTFPending:         200,
			bodyGetNTFPending:         `{"data":{"ntf_amount_kmb_off":1000,"ntf_amount_wg_off":1000,"ntf_amount_kmob_off":1000,"ntf_amount_uc":1000,"ntf_amount_wg_onl":1000,"ntf_amount_new_kmb":1000,"total_outstanding":5000000}}`,
			codeGetNTFTopUp:           200,
			bodyGetNTFTopUp:           `{"data":{"outstanding_principal":500000}}`,
			codeInternalRecord:        200,
			bodyInternalRecord:        `{"data":{}}`,
			expectGetNTFPending:       true,
			expectGetNTFTopUp:         true,
			expectGetInternalRecord:   true,
			resultMetrics:             response.Metrics{},
			shouldCallSaveTransaction: true,
			shouldCallPrescreening:    false,
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_APR,
				Reason:     "sesuai",
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			mappingCluster: entity.MasterMappingCluster{
				BranchID:       "B001",
				CustomerStatus: "NEW",
				Cluster:        "A",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockUsecase := new(mocks.Usecase)
			mockMultiUsecase := new(mocks.MultiUsecase)

			mockRepository.On("ScanTrxMaster", tc.reqMetrics.Transaction.ProspectID).Return(tc.trxMaster, tc.errScanTrxMaster)
			mockRepository.On("GetFilteringResult", tc.reqMetrics.Transaction.ProspectID).Return(tc.filtering, tc.errGetFilteringResult)
			mockRepository.On("GetElaborateLtv", tc.reqMetrics.Transaction.ProspectID).Return(entity.MappingElaborateLTV{}, tc.errGetElaborateLtv)
			mockRepository.On("GetTrxKPM", tc.reqMetrics.Transaction.ProspectID).Return(tc.trxKPM, tc.errGetTrxKPM)
			mockRepository.On("MasterMappingCluster", mock.Anything).Return(tc.mappingCluster, tc.errMappingCluster)

			if tc.shouldCallSaveTransaction {
				mockUsecase.On("SaveTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resultMetrics, tc.errSaveTransaction)
			}

			if tc.shouldCallPrescreening {
				mockUsecase.On("Prescreening", ctx, tc.reqMetrics, tc.filtering, "token").Return(tc.trxPrescreening, tc.trxFMF, tc.prescreeningDetail, tc.errPrescreening)
			}

			if tc.expectGetNTFPending {
				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("NTF_PENDING_URL"), httpmock.NewStringResponder(tc.codeGetNTFPending, tc.bodyGetNTFPending))
				resp, _ := rst.R().Post(os.Getenv("NTF_PENDING_URL"))

				mockHttpClient.On("EngineAPI",
					ctx,
					constant.NEW_KMB_LOG,
					os.Getenv("NTF_PENDING_URL"),
					mock.AnythingOfType("[]uint8"),
					mock.AnythingOfType("map[string]string"),
					constant.METHOD_POST,
					true,
					3,
					60,
					tc.reqMetrics.Transaction.ProspectID,
					"token",
				).Return(resp, tc.errGetNTFPending)
			}

			if tc.expectGetNTFTopUp {
				rst2 := resty.New()
				httpmock.ActivateNonDefault(rst2.GetClient())
				defer httpmock.DeactivateAndReset()

				chassisURL := os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL") + tc.reqMetrics.Item.NoChassis
				httpmock.RegisterResponder(constant.METHOD_GET, chassisURL, httpmock.NewStringResponder(tc.codeGetNTFTopUp, tc.bodyGetNTFTopUp))
				resp, _ := rst2.R().Get(chassisURL)

				mockHttpClient.On("EngineAPI",
					ctx,
					constant.NEW_KMB_LOG,
					chassisURL,
					mock.AnythingOfType("[]uint8"),
					mock.AnythingOfType("map[string]string"),
					constant.METHOD_GET,
					true,
					3,
					60,
					tc.reqMetrics.Transaction.ProspectID,
					"token",
				).Return(resp, tc.errGetNTFTopUp)
			}

			if tc.expectGetInternalRecord {
				rst3 := resty.New()
				httpmock.ActivateNonDefault(rst3.GetClient())
				defer httpmock.DeactivateAndReset()

				if tc.filtering.CustomerID != nil {
					customerID := tc.filtering.CustomerID.(string)
					internalRecordURL := os.Getenv("INTERNAL_RECORD_URL") + customerID

					httpmock.RegisterResponder(constant.METHOD_GET, internalRecordURL, httpmock.NewStringResponder(tc.codeInternalRecord, tc.bodyInternalRecord))
					resp, _ := rst3.R().Get(internalRecordURL)

					mockHttpClient.On("EngineAPI",
						ctx,
						constant.NEW_KMB_LOG,
						internalRecordURL,
						mock.AnythingOfType("[]uint8"),
						mock.AnythingOfType("map[string]string"),
						constant.METHOD_GET,
						true,
						3,
						60,
						tc.reqMetrics.Transaction.ProspectID,
						"token",
					).Return(resp, tc.errInternalRecord)
				}

				var dupcheckData response.SpDupcheckMap
				err := json.Unmarshal([]byte(tc.trxKPM.DupcheckData), &dupcheckData)
				if err == nil && dupcheckData.CustomerID != nil {
					customerID := dupcheckData.CustomerID.(string)
					internalRecordURL := os.Getenv("INTERNAL_RECORD_URL") + customerID

					httpmock.RegisterResponder(constant.METHOD_GET, internalRecordURL, httpmock.NewStringResponder(tc.codeInternalRecord, tc.bodyInternalRecord))
					resp, _ := rst3.R().Get(internalRecordURL)

					mockHttpClient.On("EngineAPI",
						ctx,
						constant.NEW_KMB_LOG,
						internalRecordURL,
						mock.AnythingOfType("[]uint8"),
						mock.AnythingOfType("map[string]string"),
						constant.METHOD_GET,
						true,
						3,
						60,
						tc.reqMetrics.Transaction.ProspectID,
						"token",
					).Return(resp, tc.errInternalRecord)
				}
			}

			if tc.err == nil {
				mockRepository.On("UpdateTrxKPMStatus", mock.Anything, mock.Anything).Return(nil)
			}

			metrics := NewMetrics(mockRepository, mockHttpClient, mockUsecase, mockMultiUsecase)
			result, err := metrics.Submission2Wilen(ctx, tc.reqMetrics, "token")
			require.Equal(t, tc.resultMetrics, result)
			require.Equal(t, tc.err, err)
		})
	}
}
