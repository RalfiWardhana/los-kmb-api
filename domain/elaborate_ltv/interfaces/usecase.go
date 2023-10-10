package interfaces

import (
	"context"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	Elaborate(ctx context.Context, reqs request.ElaborateLTV, accessToken string) (data response.ElaborateLTV, err error)
}
