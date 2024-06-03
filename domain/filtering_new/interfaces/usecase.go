package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	FilteringPefindo(ctx context.Context, reqPefindo request.Pefindo, customerStatus, clusterCMO string, accessToken string) (data response.Filtering, responsePefindo response.PefindoResult, trxDetailBiro []entity.TrxDetailBiro, err error)
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	SaveFiltering(transaction entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, transactionCMOnoFPD entity.TrxCmoNoFPD) (err error)
	FilteringProspectID(prospectID string) (data request.OrderIDCheck, err error)
	GetResultFiltering(prospectID string) (respFiltering response.Filtering, err error)
	GetEmployeeData(ctx context.Context, employeeID string, accessToken string, hrisAccessToken string) (data response.EmployeeCMOResponse, err error)
	GetFpdCMO(ctx context.Context, CmoID string, BPKBNameType string, accessToken string) (data response.FpdCMOResponse, err error)
	CheckCmoNoFPD(prospectID string, cmoID string, cmoCategory string, cmoJoinDate string, defaultCluster string, bpkbName string) (clusterCMOSaved string, entitySaveTrxNoFPd entity.TrxCmoNoFPD, err error)
	CheckLatestPaidInstallment(ctx context.Context, prospectID string, customerID string, accessToken string) (respRrdDate string, monthsDiff int, err error)
}

type MultiUsecase interface {
	Filtering(ctx context.Context, reqFiltering request.Filtering, married bool, accessToken string, hrisAccessToken string) (data response.Filtering, err error)
}
