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
	VehicleCheck(manufactureYear, cmoCluster, bkpbName string, tenor int, configValue response.DupcheckConfig, filtering entity.FilteringKMB, af float64) (data response.UsecaseApi, err error)
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
	PrincipleElaborateLTV(ctx context.Context, r request.PrincipleElaborateLTV) (data response.ElaborateLTV, err error)
	AgreementChassisNumberIntegrator(ctx context.Context, prospectID, chassisNumber string, accessToken string) (data response.AgreementChassisNumber, err error)
	RejectTenor36(cluster string) (data response.UsecaseApi, err error)
	Scorepro(ctx context.Context, req request.PrinciplePembiayaan, principleStepOne entity.TrxPrincipleStepOne, principleStepTwo entity.TrxPrincipleStepTwo, pefindoScore, customerStatus, customerSegment string, installmentTopUp float64, spDupcheck response.SpDupCekCustomerByID, accessToken string) (data response.ScorePro, err error)
	DsrCheck(ctx context.Context, req request.PrinciplePembiayaan, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, agreementChasisNumber response.AgreementChassisNumber, accessToken string, configValue response.DupcheckConfig) (data response.UsecaseApi, err error)
}

type MultiUsecase interface {
	PrinciplePemohon(ctx context.Context, r request.PrinciplePemohon) (data response.UsecaseApi, err error)
	PrinciplePembiayaan(ctx context.Context, r request.PrinciplePembiayaan, accessToken string) (data response.UsecaseApi, err error)
}
