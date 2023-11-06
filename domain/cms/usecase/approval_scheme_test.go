package usecase

import (
	"los-kmb-api/domain/cms/interfaces/mocks"
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
		req  request.ReqApprovalScheme
		resp response.RespApprovalScheme
		err  error
	}{
		{
			name: "test app",
			req: request.ReqApprovalScheme{
				DecisionAlias: "CBM",
			},
			resp: response.RespApprovalScheme{
				NextStep: "DRM",
				IsFinal:  false,
			},
		},
		{
			name: "test app",
			req: request.ReqApprovalScheme{
				DecisionAlias: "DRM",
			},
			resp: response.RespApprovalScheme{
				NextStep: "GMO",
				IsFinal:  false,
			},
		},
		{
			name: "test app",
			req: request.ReqApprovalScheme{
				DecisionAlias: "COM",
			},
			resp: response.RespApprovalScheme{
				NextStep: "",
				IsFinal:  true,
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
