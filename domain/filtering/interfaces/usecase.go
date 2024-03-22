package interfaces

import (
	"context"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	FilteringBlackList(ctx context.Context, reqs request.FilteringRequest, request_id string) (result response.DupcheckResult, err error)
	FilteringPefindo(ctx context.Context, reqs request.FilteringRequest, status_konsumen, request_id string) (data response.DupcheckResult, isRejectClusterEF bool, err error)
	CheckStatusCategory(ctx context.Context, reqs request.FilteringRequest, status_konsumen, accessToken string) (data response.DupcheckResult, err error)
}

type MultiUsecase interface {
	Filtering(ctx context.Context, reqs request.FilteringRequest, accessToken string) (data response.DupcheckResult, err error)
}
