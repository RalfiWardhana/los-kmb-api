package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	FilteringPefindo(ctx context.Context, reqs request.FilteringRequest, status_konsumen, request_id string) (data response.Filtering, responsePefindo interface{}, err error)
	CheckStatusCategory(ctx context.Context, reqs request.FilteringRequest, status_konsumen, accessToken string) (data response.DupcheckResult, err error)
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	SaveFilteringLogs(transaction entity.FilteringKMB) (err error)
	FilteringProspectID(prospectID string) (data request.OrderIDCheck, err error)
}

type MultiUsecase interface {
	Filtering(ctx context.Context, reqs request.Filtering, married bool, accessToken string) (data response.Filtering, err error)
}
