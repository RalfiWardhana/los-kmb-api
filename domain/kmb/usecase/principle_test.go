package usecase

import (
	"context"
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

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrincipleSubmission(t *testing.T) {
	os.Setenv("BIRO_VALID_DAYS", "4")
	os.Setenv("NTF_PENDING_URL", "http://localhost/")
	os.Setenv("AGREEMENT_OF_CHASSIS_NUMBER_URL", "http://localhost/")
	os.Setenv("INTERNAL_RECORD_URL", "http://localhost/")
	ctx := context.Background()

	testcases := []struct {
		name                     string
		reqMetrics               request.Metrics
		trxMaster                int
		errScanTrxMaster         error
		filtering                entity.FilteringKMB
		errGetFilteringResult    error
		errGetElaborateLtv       error
		resGetPrincipleStepOne   entity.TrxPrincipleStepOne
		errGetPrincipleStepOne   error
		resGetPrincipleStepTwo   entity.TrxPrincipleStepTwo
		errGetPrincipleStepTwo   error
		resGetPrincipleStepThree entity.TrxPrincipleStepThree
		errGetPrincipleStepThree error
		mappingCluster           entity.MasterMappingCluster
		errMappingCluster        error
		codeGetNTFPending        int
		bodyGetNTFPending        string
		errGetNTFPending         error
		codeGetNTFTopUp          int
		bodyGetNTFTopUp          string
		errGetNTFTopUp           error
		codeInternalRecord       int
		bodyInternalRecord       string
		errInternalRecord        error
		customerID               string
		errSaveTransaction       error
		resultMetrics            interface{}
		err                      error
		expectGetNTFPending      bool
		expectGetNTFTopUp        bool
		expectGetInternalRecord  bool
	}{
		{
			name: "test metrics errScanTrxMaster",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:              errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error"),
			errScanTrxMaster: errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error"),
		},
		{
			name: "test metrics trxMaster > 0",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:       errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID Already Exist"),
			trxMaster: 1,
		},
		{
			name: "test metrics errGetFilteringResult Belum melakukan filtering",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:                   errors.New(fmt.Sprintf("%s - Belum melakukan filtering atau hasil filtering sudah lebih dari %s hari", constant.ERROR_BAD_REQUEST, os.Getenv("BIRO_VALID_DAYS"))),
			trxMaster:             0,
			errGetFilteringResult: errors.New(constant.RECORD_NOT_FOUND),
		},
		{
			name: "test metrics errGetFilteringResult selain Belum melakukan filtering",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:                   errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
			trxMaster:             0,
			errGetFilteringResult: errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
		},
		{
			name: "test metrics errGetFilteringResult Tidak bisa lanjut proses",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:       errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses"),
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    0,
				CustomerStatus: "NEW",
			},
		},
		{
			name: "test metrics errGetElaborateLtv",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:       errors.New(constant.ERROR_BAD_REQUEST + " - Belum melakukan pengecekan LTV"),
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			errGetElaborateLtv: errors.New(constant.ERROR_BAD_REQUEST + " - Belum melakukan pengecekan LTV"),
		},
		{
			name: "error get principle step one",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			errGetPrincipleStepOne: errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step One Error"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step One Error"),
		},
		{
			name: "error get principle step two",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			errGetPrincipleStepTwo: errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step Two Error"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step Two Error"),
		},
		{
			name: "error get principle step three",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			errGetPrincipleStepThree: errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step Three Error"),
			err:                      errors.New(constant.ERROR_UPSTREAM + " - Get Principle Step Three Error"),
		},
		{
			name: "error first save transaction",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			resultMetrics:      response.Metrics{},
			errSaveTransaction: errors.New(constant.ERROR_UPSTREAM + " - Save Transaction Error"),
			err:                errors.New(constant.ERROR_UPSTREAM + " - Save Transaction Error"),
		},
		{
			name: "error get ntf pending",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			resultMetrics:       response.Metrics{},
			errGetNTFPending:    errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error"),
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error"),
			expectGetNTFPending: true,
		},
		{
			name: "error code get ntf pending",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			resultMetrics:       response.Metrics{},
			codeGetNTFPending:   400,
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error"),
			expectGetNTFPending: true,
		},
		{
			name: "error get ntf top up",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			resultMetrics:       response.Metrics{},
			codeGetNTFPending:   200,
			errGetNTFPending:    nil,
			errGetNTFTopUp:      errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error"),
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error"),
			expectGetNTFPending: true,
			expectGetNTFTopUp:   true,
		},
		{
			name: "error code get ntf top up",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			resultMetrics:       response.Metrics{},
			codeGetNTFPending:   200,
			errGetNTFPending:    nil,
			codeGetNTFTopUp:     400,
			errGetNTFTopUp:      nil,
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error"),
			expectGetNTFPending: true,
			expectGetNTFTopUp:   true,
		},
		{
			name: "error unmarshal get ntf top up",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			resultMetrics:       response.Metrics{},
			codeGetNTFPending:   200,
			errGetNTFPending:    nil,
			codeGetNTFTopUp:     200,
			errGetNTFTopUp:      nil,
			bodyGetNTFTopUp:     `invalid json`,
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error"),
			expectGetNTFPending: true,
			expectGetNTFTopUp:   true,
		},
		{
			name: "error get mapping cluster",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "RO",
				CustomerID:     "123",
			},
			resultMetrics:       response.Metrics{},
			codeGetNTFPending:   200,
			errGetNTFPending:    nil,
			codeGetNTFTopUp:     200,
			errGetNTFTopUp:      nil,
			bodyGetNTFTopUp:     `{"data":{"lc_installment":100000}}`,
			errMappingCluster:   errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error"),
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error"),
			expectGetNTFPending: true,
			expectGetNTFTopUp:   true,
		},
		{
			name: "error unmarshal dupcheck data",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				DupcheckData: `invalid json`,
			},
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "RO",
				CustomerID:     "123",
			},
			resultMetrics:       response.Metrics{},
			codeGetNTFPending:   200,
			errGetNTFPending:    nil,
			codeGetNTFTopUp:     200,
			errGetNTFTopUp:      nil,
			bodyGetNTFTopUp:     `{"data":{"lc_installment":100000}}`,
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Unmarshal Dupcheck Data Error"),
			expectGetNTFPending: true,
			expectGetNTFTopUp:   true,
		},
		{
			name: "error get internal record",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				DupcheckData: `{"max_overduedays_roao":60,"customer_id": "123456"}`,
			},
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "RO",
				CustomerID:     "123",
				CMOCluster:     "Cluster C",
			},
			resultMetrics:           response.Metrics{},
			codeGetNTFPending:       200,
			errGetNTFPending:        nil,
			codeGetNTFTopUp:         200,
			errGetNTFTopUp:          nil,
			bodyGetNTFTopUp:         `{"data":{"lc_installment":100000}}`,
			customerID:              "123456",
			errInternalRecord:       errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Interal Record Error"),
			err:                     errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Interal Record Error"),
			expectGetNTFPending:     true,
			expectGetNTFTopUp:       true,
			expectGetInternalRecord: true,
		},
		{
			name: "error unmarshal dsr fmf pbk info",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				DupcheckData: `{"max_overduedays_roao":60,"customer_id": "123456"}`,
			},
			resGetPrincipleStepThree: entity.TrxPrincipleStepThree{
				CheckDSRFMFPBKInfo: `invalid json`,
			},
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "RO",
				CustomerID:     "123",
				CMOCluster:     "Cluster C",
			},
			resultMetrics:           response.Metrics{},
			codeGetNTFPending:       200,
			errGetNTFPending:        nil,
			codeGetNTFTopUp:         200,
			errGetNTFTopUp:          nil,
			bodyGetNTFTopUp:         `{"data":{"lc_installment":100000}}`,
			customerID:              "123456",
			codeInternalRecord:      200,
			errInternalRecord:       nil,
			err:                     errors.New(constant.ERROR_UPSTREAM + " - Unmarshal DSR FMF PBK Info Error"),
			expectGetNTFPending:     true,
			expectGetNTFTopUp:       true,
			expectGetInternalRecord: true,
		},
		{
			name: "error parse float installment amount",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				DupcheckData: `{"max_overduedays_roao":60,"customer_id": "123456"}`,
			},
			resGetPrincipleStepThree: entity.TrxPrincipleStepThree{
				CheckDSRFMFPBKInfo: `{"latest_installment_amount":"invalid float","total_dsr": 20}`,
			},
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "RO",
				CustomerID:     "123",
				CMOCluster:     "Cluster C",
			},
			resultMetrics:           response.Metrics{},
			codeGetNTFPending:       200,
			errGetNTFPending:        nil,
			codeGetNTFTopUp:         200,
			errGetNTFTopUp:          nil,
			bodyGetNTFTopUp:         `{"data":{"lc_installment":100000}}`,
			customerID:              "123456",
			codeInternalRecord:      200,
			errInternalRecord:       nil,
			err:                     errors.New(constant.ERROR_UPSTREAM + " - GetFloat latest_installment_amount Error"),
			expectGetNTFPending:     true,
			expectGetNTFTopUp:       true,
			expectGetInternalRecord: true,
		},
		{
			name: "success",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Item: request.Item{
					NoChassis: "ASD123",
				},
				CustomerPersonal: request.CustomerPersonal{
					MaritalStatus: constant.MARRIED,
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				DupcheckData: `{"max_overduedays_roao":60,"customer_id": "123456"}`,
			},
			resGetPrincipleStepThree: entity.TrxPrincipleStepThree{
				CheckDSRFMFPBKInfo: `{"latest_installment_amount":6000000,"total_dsr": 20}`,
			},
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "RO",
				CustomerID:     "123",
				CMOCluster:     "Cluster C",
			},
			resultMetrics:           response.Metrics{},
			codeGetNTFPending:       200,
			errGetNTFPending:        nil,
			codeGetNTFTopUp:         200,
			errGetNTFTopUp:          nil,
			bodyGetNTFTopUp:         `{"data":{"lc_installment":100000}}`,
			customerID:              "123456",
			codeInternalRecord:      200,
			errInternalRecord:       nil,
			expectGetNTFPending:     true,
			expectGetNTFTopUp:       true,
			expectGetInternalRecord: true,
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
			mockRepository.On("GetPrincipleStepOne", tc.reqMetrics.Transaction.ProspectID).Return(tc.resGetPrincipleStepOne, tc.errGetPrincipleStepOne)
			mockRepository.On("GetPrincipleStepTwo", tc.reqMetrics.Transaction.ProspectID).Return(tc.resGetPrincipleStepTwo, tc.errGetPrincipleStepTwo)
			mockRepository.On("GetPrincipleStepThree", tc.reqMetrics.Transaction.ProspectID).Return(tc.resGetPrincipleStepThree, tc.errGetPrincipleStepThree)
			mockRepository.On("MasterMappingCluster", mock.Anything).Return(tc.mappingCluster, tc.errMappingCluster)
			mockUsecase.On("SaveTransaction", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resultMetrics, tc.errSaveTransaction)

			if tc.expectGetNTFPending {
				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("NTF_PENDING_URL"), httpmock.NewStringResponder(tc.codeGetNTFPending, tc.bodyGetNTFPending))
				resp, _ := rst.R().Post(os.Getenv("NTF_PENDING_URL"))

				mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_PENDING_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, true, 3, 60, tc.reqMetrics.Transaction.ProspectID, "token").Return(resp, tc.errGetNTFPending)
			}

			if tc.expectGetNTFTopUp {
				rst2 := resty.New()
				httpmock.ActivateNonDefault(rst2.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"), httpmock.NewStringResponder(tc.codeGetNTFTopUp, tc.bodyGetNTFTopUp))
				resp, _ := rst2.R().Get(os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"))

				mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+tc.reqMetrics.Item.NoChassis, mock.Anything, mock.Anything, constant.METHOD_GET, true, 3, 60, tc.reqMetrics.Transaction.ProspectID, "token").Return(resp, tc.errGetNTFTopUp)
			}

			if tc.expectGetInternalRecord {
				rst3 := resty.New()
				httpmock.ActivateNonDefault(rst3.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("INTERNAL_RECORD_URL"), httpmock.NewStringResponder(tc.codeInternalRecord, tc.bodyInternalRecord))
				resp, _ := rst3.R().Get(os.Getenv("INTERNAL_RECORD_URL"))

				mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("INTERNAL_RECORD_URL")+tc.customerID, mock.Anything, map[string]string{}, constant.METHOD_GET, true, 3, 60, tc.reqMetrics.Transaction.ProspectID, "token").Return(resp, tc.errInternalRecord)
			}

			metrics := NewMetrics(mockRepository, mockHttpClient, mockUsecase, mockMultiUsecase)
			result, err := metrics.PrincipleSubmission(ctx, tc.reqMetrics, "token")
			require.Equal(t, tc.resultMetrics, result)
			require.Equal(t, tc.err, err)
		})
	}
}
