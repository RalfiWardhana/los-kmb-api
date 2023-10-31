package usecase

import (
	"context"
	"errors"
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

func TestElaborateScheme(t *testing.T) {
	testcases := []struct {
		name               string
		trxElaborateLtv    entity.MappingElaborateLTV
		errTrxElaborateLtv error
		req                request.Metrics
		errResult          error
		result             response.UsecaseApi
	}{
		{
			name: "ElaborateScheme MappingElaborateLTV error",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
				},
			},
			errTrxElaborateLtv: errors.New("Get MappingElaborateLTV Error"),
			errResult:          errors.New(constant.ERROR_UPSTREAM + " - GetElaborateLtv Error"),
		},
		{
			name: "ElaborateScheme reject",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
				},
				Apk: request.Apk{
					AF:  17000000,
					OTR: 17000000,
				},
			},
			trxElaborateLtv: entity.MappingElaborateLTV{
				LTV: 70,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.STRING_CODE_REJECT_NTF_ELABORATE,
				Reason:         constant.REASON_REJECT_NTF_ELABORATE,
				SourceDecision: constant.SOURCE_DECISION_ELABORATE_LTV,
			},
		},
		{
			name: "ElaborateScheme pass",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
				},
				Apk: request.Apk{
					AF:  8000000,
					OTR: 17000000,
				},
			},
			trxElaborateLtv: entity.MappingElaborateLTV{
				LTV: 70,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.STRING_CODE_PASS_ELABORATE,
				Reason:         constant.REASON_PASS_ELABORATE,
				SourceDecision: constant.SOURCE_DECISION_ELABORATE_LTV,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetElaborateLtv", tc.req.Transaction.ProspectID).Return(tc.trxElaborateLtv, tc.errTrxElaborateLtv)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, err := usecase.ElaborateScheme(tc.req)
			require.Equal(t, tc.result, data)
			require.Equal(t, tc.errResult, err)
		})
	}
}

func TestElaborateIncome(t *testing.T) {
	testcases := []struct {
		name                      string
		req                       request.Metrics
		filtering                 entity.FilteringKMB
		pefindoIDX                response.PefindoIDX
		spDupcheckMap             response.SpDupcheckMap
		responseScs               response.IntegratorScorePro
		accessToken               string
		result                    response.UsecaseApi
		errResult                 error
		MasterBranch              entity.MasterBranch
		errMasterBranch           error
		mappingElaborateIncome    entity.MappingElaborateIncome
		errMappingElaborateIncome error
		code                      int
		body                      string
		errResp                   error
	}{
		{
			name: "ElaborateIncome MasterBranch error",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
					BranchID:   "426",
				},
			},
			errMasterBranch: errors.New("Get GetMasterBranch Error"),
			errResult:       errors.New(constant.ERROR_UPSTREAM + " - GetMasterBranch Error"),
		},
		{
			name: "ElaborateIncome error Call Low Income Timeout",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
					BranchID:   "426",
				},
			},
			errResp:   errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Low Income Timeout"),
			errResult: errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Low Income Timeout"),
		},
		{
			name: "ElaborateIncome error Call Low Income Error",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
					BranchID:   "426",
				},
			},
			code:      400,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Call Low Income Error"),
		},
		{
			name: "ElaborateIncome reject",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
					BranchID:   "426",
				},
			},
			code: 200,
			body: `{"messages":"OK","data":{"no_application":"EFM04486202102200011","income":387300,"range":"< 2.5 Juta"},"errors":null,"server_time":"2023-10-31T08:52:26+07:00"}`,
			mappingElaborateIncome: entity.MappingElaborateIncome{
				Result: constant.DECISION_REJECT,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_ELABORATE_INCOME,
				Reason:         constant.REASON_REJECT_ELABORATE_INCOME,
				SourceDecision: constant.SOURCE_DECISION_ELABORATE_INCOME,
			},
		},
		{
			name: "ElaborateIncome pass",
			req: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST001",
					BranchID:   "426",
				},
			},
			code: 200,
			body: `{"messages":"OK","data":{"no_application":"EFM04486202102200011","income":387300,"range":"< 2.5 Juta"},"errors":null,"server_time":"2023-10-31T08:52:26+07:00"}`,
			mappingElaborateIncome: entity.MappingElaborateIncome{
				Result: constant.DECISION_PASS,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PASS_ELABORATE_INCOME,
				Reason:         constant.REASON_PASS_ELABORATE_INCOME,
				SourceDecision: constant.SOURCE_DECISION_ELABORATE_INCOME,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetMasterBranch", tc.req.Transaction.BranchID).Return(tc.MasterBranch, tc.errMasterBranch)
			mockRepository.On("GetMappingElaborateIncome", mock.Anything).Return(tc.mappingElaborateIncome, tc.errMappingElaborateIncome)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			os.Setenv("LOW_INCOME_API", "http://localhost/")
			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("LOW_INCOME_API"), httpmock.NewStringResponder(tc.code, tc.body))
			resp, _ := rst.R().Post(os.Getenv("LOW_INCOME_API"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("LOW_INCOME_API"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 60, tc.req.Transaction.ProspectID, tc.accessToken).Return(resp, tc.errResp).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, err := usecase.ElaborateIncome(ctx, tc.req, tc.filtering, tc.pefindoIDX, tc.spDupcheckMap, tc.responseScs, tc.accessToken)
			require.Equal(t, tc.result, data)
			require.Equal(t, tc.errResult, err)
		})
	}
}
