package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	FilteringPefindo(ctx context.Context, reqs request.FilteringRequest, status_konsumen, request_id string) (data response.Filtering, responsePefindo response.PefindoResult, trxDetailBiro []entity.TrxDetailBiro, err error)
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	SaveFiltering(transaction entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro) (err error)
	FilteringProspectID(prospectID string) (data request.OrderIDCheck, err error)
}

type MultiUsecase interface {
	Filtering(ctx context.Context, reqs request.Filtering, married bool, accessToken string) (data response.Filtering, err error)
}
