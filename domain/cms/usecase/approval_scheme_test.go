package usecase

import (
	"los-kmb-api/domain/cms/mocks"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/httpclient"
	"testing"

	"github.com/allegro/bigcache/v3"
	"github.com/stretchr/testify/require"
)

func TestApprovalScheme(t *testing.T) {
	testcases := []struct {
		name string
		req  request.ReqSubmitApproval
		resp response.RespApprovalScheme
		err  error
	}{
		{
			name: "test app DRM",
			req: request.ReqSubmitApproval{
				FinalApproval: "COM",
				Alias:         "CBM",
			},
			resp: response.RespApprovalScheme{
				NextStep: "DRM",
				IsFinal:  false,
			},
		},
		{
			name: "test app GMO",
			req: request.ReqSubmitApproval{
				FinalApproval: "COM",
				Alias:         "DRM",
			},
			resp: response.RespApprovalScheme{
				NextStep: "GMO",
				IsFinal:  false,
			},
		},
		{
			name: "test app final",
			req: request.ReqSubmitApproval{
				FinalApproval: "COM",
				Alias:         "COM",
			},
			resp: response.RespApprovalScheme{
				NextStep: "",
				IsFinal:  true,
			},
		},
		{
			name: "test app final need escalation",
			req: request.ReqSubmitApproval{
				FinalApproval:  "COM",
				Alias:          "COM",
				NeedEscalation: true,
			},
			resp: response.RespApprovalScheme{
				NextStep:     "GMC",
				IsFinal:      false,
				IsEscalation: true,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			mockRepository := new(mocks.Repository)
			var cache *bigcache.BigCache
			mockHttpClient := new(httpclient.MockHttpClient)

			usecase := NewUsecase(mockRepository, mockHttpClient, cache)

			data, err := usecase.ApprovalScheme(tc.req)
			require.Equal(t, tc.resp, data)
			require.Equal(t, tc.err, err)
		})
	}
}
