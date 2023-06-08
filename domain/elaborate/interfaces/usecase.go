package interfaces

import (
	"context"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	ResultElaborate(ctx context.Context, reqs request.BodyRequestElaborate) (data response.ElaborateResult, err error)
}

type MultiUsecase interface {
	Elaborate(ctx context.Context, reqs request.BodyRequestElaborate, accessToken string) (data response.ElaborateResult, err error)
}
