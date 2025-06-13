package usecase

import (
	"context"
	"errors"
	mocksCache "los-kmb-api/domain/cache/mocks"
	"los-kmb-api/domain/cms/mocks"
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

func TestGenerateFormAKKK(t *testing.T) {
	os.Setenv("GENERATOR_FORM_AKKK_URL", "http://localhost/generator-form-akkk")
	ctx := context.Background()
	accessToken := "token"
	testcases := []struct {
		name               string
		reqs               request.RequestGenerateFormAKKK
		trxStatus          entity.TrxStatus
		errtrxStatus       error
		httpcode           int
		httpbody           string
		httperr            error
		errSaveUrlFormAKKK error
		errSaveWorker      error
		data               interface{}
		errResult          error
	}{
		{
			name: "TestGenerateFormAKKK GetTrxStatus error",
			reqs: request.RequestGenerateFormAKKK{
				ProspectID: "SAL-1234567",
			},
			errtrxStatus: errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK GetTrxStatus Error"),
			errResult:    errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK GetTrxStatus Error"),
		},
		{
			name: "TestGenerateFormAKKK GetTrxStatus error Pengajuan Belum Selesai",
			reqs: request.RequestGenerateFormAKKK{
				ProspectID: "SAL-1234567",
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_UNPROCESS,
			},
			errResult: errors.New(constant.ERROR_BAD_REQUEST + " - GenerateFormAKKK GetTrxStatus Status Pengajuan Belum Selesai"),
		},
		{
			name: "TestGenerateFormAKKK EngineAPI not 200",
			reqs: request.RequestGenerateFormAKKK{
				ProspectID: "SAL-1234567",
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
			},
			httpcode:  500,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Failed Generate Form AKKK"),
		},
		{
			name: "TestGenerateFormAKKK EngineAPI 200 MediaUrl empty",
			reqs: request.RequestGenerateFormAKKK{
				ProspectID: "SAL-1234567",
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
			},
			httpcode:  200,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Unmarshal MediaUrl Form AKKK Error"),
		},
		{
			name: "TestGenerateFormAKKK EngineAPI SaveUrlFormAKKK error",
			reqs: request.RequestGenerateFormAKKK{
				ProspectID: "SAL-1234567",
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
			},
			httpbody:           `{"data":{"media_url":"https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1161124061800051/formAKKK_SAL-1161124061800051.pdf","path":"formAKKK_SAL-1161124061800051.pdf"}}`,
			httpcode:           200,
			errSaveUrlFormAKKK: errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK SaveUrlFormAKKK Error"),
			errResult:          errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK SaveUrlFormAKKK Error"),
		},
		{
			name: "TestGenerateFormAKKK EngineAPI SaveUrlFormAKKK oke",
			reqs: request.RequestGenerateFormAKKK{
				ProspectID: "SAL-1234567",
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
			},
			httpbody: `{"data":{"media_url":"https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1161124061800051/formAKKK_SAL-1161124061800051.pdf","path":"formAKKK_SAL-1161124061800051.pdf"}}`,
			httpcode: 200,
			data: response.ResponseGenerateFormAKKK{
				MediaUrl: "https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1161124061800051/formAKKK_SAL-1161124061800051.pdf",
				Path:     "formAKKK_SAL-1161124061800051.pdf",
			},
		},
		{
			name: "TestGenerateFormAKKK SaveWorker error",
			reqs: request.RequestGenerateFormAKKK{
				ProspectID: "SAL-1234567",
				Source:     constant.SYSTEM,
			},
			trxStatus: entity.TrxStatus{
				Activity: constant.ACTIVITY_STOP,
			},
			httpcode:      500,
			errSaveWorker: errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK SaveWorker Error"),
			errResult:     errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK SaveWorker Error"),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockCache := new(mocksCache.Repository)

			mockRepository.On("GetTrxStatus", tc.reqs.ProspectID).Return(tc.trxStatus, tc.errtrxStatus)
			mockRepository.On("SaveUrlFormAKKK", tc.reqs.ProspectID, mock.Anything).Return(tc.errSaveUrlFormAKKK)
			mockRepository.On("SaveWorker", mock.Anything).Return(tc.errSaveWorker)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("GENERATOR_FORM_AKKK_URL"), httpmock.NewStringResponder(tc.httpcode, tc.httpbody))
			resp, _ := rst.R().Post(os.Getenv("GENERATOR_FORM_AKKK_URL"))
			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("GENERATOR_FORM_AKKK_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 60, tc.reqs.ProspectID, accessToken).Return(resp, tc.httperr).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, mockCache)

			result, err := usecase.GenerateFormAKKK(ctx, tc.reqs, accessToken)

			require.Equal(t, tc.data, result)
			require.Equal(t, tc.errResult, err)
		})
	}

}
