package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	CheckNokaNosin(ctx context.Context, r request.PrincipleAsset) (data response.UsecaseApi, err error)
	CheckPMK(branchID, customerKMB string, income float64, homeStatus, professionID, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi, err error)
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	GetEmployeeData(ctx context.Context, employeeID string) (data response.EmployeeCMOResponse, err error)
	GetFpdCMO(ctx context.Context, CmoID string, BPKBNameType string) (data response.FpdCMOResponse, err error)
	CheckCmoNoFPD(prospectID string, cmoID string, cmoCategory string, cmoJoinDate string, defaultCluster string, bpkbName string) (clusterCMOSaved string, entitySaveTrxNoFPd entity.TrxCmoNoFPD, err error)
	Pefindo(ctx context.Context, r request.Pefindo, customerStatus, clusterCMO string, bpkbName string) (data response.Filtering, responsePefindo response.PefindoResult, trxDetailBiro []entity.TrxDetailBiro, err error)
	Save(transaction entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, transactionCMOnoFPD entity.TrxCmoNoFPD) (err error)
	BannedPMKOrDSR(idNumber string) (data response.UsecaseApi, err error)
	Rejection(prospectID string, encrypted string, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedPMKDSR entity.TrxBannedPMKDSR, err error)
	CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error)
	Dukcapil(ctx context.Context, r request.PrinciplePemohon, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error)
	Asliri(ctx context.Context, r request.PrinciplePemohon, accessToken string) (data response.Ekyc, err error)
	Ktp(ctx context.Context, r request.PrinciplePemohon, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error)
	PrincipleStep(idNumber string) (step response.StepPrinciple, err error)
}

type MultiUsecase interface {
	PrinciplePemohon(ctx context.Context, r request.PrinciplePemohon) (data response.UsecaseApi, err error)
}
