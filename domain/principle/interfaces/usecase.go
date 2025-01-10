package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	CheckNokaNosin(ctx context.Context, r request.PrincipleAsset) (data response.UsecaseApi, err error)
	CheckPMK(branchID, customerKMB string, income float64, homeStatus, professionID, birthDate string, tenor int, maritalStatus string, empYear, empMonth, stayYear, stayMonth int) (data response.UsecaseApi, err error)
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	VehicleCheck(manufactureYear, cmoCluster, bkpbName string, tenor int, configValue response.DupcheckConfig, filtering entity.FilteringKMB, af float64) (data response.UsecaseApi, err error)
	GetEmployeeData(ctx context.Context, employeeID string) (data response.EmployeeCMOResponse, err error)
	GetFpdCMO(ctx context.Context, CmoID string, BPKBNameType string) (data response.FpdCMOResponse, err error)
	CheckCmoNoFPD(prospectID string, cmoID string, cmoCategory string, cmoJoinDate string, defaultCluster string, bpkbName string) (clusterCMOSaved string, entitySaveTrxNoFPd entity.TrxCmoNoFPD, err error)
	Pefindo(ctx context.Context, r request.Pefindo, customerStatus, clusterCMO string, bpkbName string) (data response.Filtering, responsePefindo response.PefindoResult, trxDetailBiro []entity.TrxDetailBiro, err error)
	CheckLatestPaidInstallment(ctx context.Context, prospectID string, customerID string, accessToken string) (respRrdDate string, monthsDiff int, err error)
	Save(transaction entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, transactionCMOnoFPD entity.TrxCmoNoFPD) (err error)
	BannedPMKOrDSR(idNumber string) (data response.UsecaseApi, err error)
	Rejection(prospectID string, encrypted string, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedPMKDSR entity.TrxBannedPMKDSR, err error)
	CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error)
	Dukcapil(ctx context.Context, r request.PrinciplePemohon, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error)
	Asliri(ctx context.Context, r request.PrinciplePemohon, accessToken string) (data response.Ekyc, err error)
	Ktp(ctx context.Context, r request.PrinciplePemohon, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error)
	PrincipleStep(idNumber string) (step response.StepPrinciple, err error)
	PrincipleElaborateLTV(ctx context.Context, r request.PrincipleElaborateLTV, accessToken string) (data response.PrincipleElaborateLTV, err error)
	AgreementChassisNumberIntegrator(ctx context.Context, prospectID, chassisNumber string, accessToken string) (data response.AgreementChassisNumber, err error)
	RejectTenor36(cluster string) (data response.UsecaseApi, err error)
	Scorepro(ctx context.Context, req request.PrinciplePembiayaan, principleStepOne entity.TrxPrincipleStepOne, principleStepTwo entity.TrxPrincipleStepTwo, pefindoScore, customerStatus, customerSegment string, installmentTopUp float64, spDupcheck response.SpDupCekCustomerByID, filtering entity.FilteringKMB, accessToken string) (responseScs response.IntegratorScorePro, data response.ScorePro, respPefindoIDX response.PefindoIDX, err error)
	DsrCheck(ctx context.Context, req request.PrinciplePembiayaan, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, agreementChasisNumber response.AgreementChassisNumber, accessToken string, configValue response.DupcheckConfig) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error)
	TotalDsrFmfPbk(ctx context.Context, totalIncome, newInstallment, totalInstallmentPBK float64, prospectID, customerSegment, accessToken string, SpDupcheckMap response.SpDupcheckMap, configValue response.DupcheckConfig, filtering entity.FilteringKMB) (data response.UsecaseApi, trxFMF response.TrxFMF, err error)
	PrincipleCoreCustomer(ctx context.Context, prospectID string, accessToken string) (err error)
	PrincipleMarketingProgram(ctx context.Context, prospectID string, accessToken string) (err error)
	MDMGetMasterMappingBranchEmployee(ctx context.Context, prospectID, branchID, accessToken string) (data response.MDMMasterMappingBranchEmployeeResponse, err error)
	GetDataPrinciple(ctx context.Context, req request.PrincipleGetData, accessToken string) (data map[string]interface{}, err error)
	CheckOrderPendingPrinciple(ctx context.Context) (err error)
	PrinciplePublish(ctx context.Context, req request.PrinciplePublish, accessToken string) (err error)

	Step2Wilen(idNumber string) (step response.Step2Wilen, err error)

	GetLTV(mappingElaborateLTV []entity.MappingElaborateLTV, resultPefindo, bpkbName, manufactureYear string, tenor int, bakiDebet float64) (ltv int, err error)
	MarsevGetLoanAmount(ctx context.Context, req request.ReqMarsevLoanAmount, prospectID string, accessToken string) (marsevLoanAmountRes response.MarsevLoanAmountResponse, err error)
	MarsevGetMarketingProgram(ctx context.Context, req request.ReqMarsevFilterProgram, prospectID string, accessToken string) (marsevFilterProgramRes response.MarsevFilterProgramResponse, err error)
	MarsevCalculateInstallment(ctx context.Context, req request.ReqMarsevCalculateInstallment, prospectID string, accessToken string) (marsevCalculateInstallmentRes response.MarsevCalculateInstallmentResponse, err error)
	MDMGetMasterAsset(ctx context.Context, branchID string, search string, prospectID string, accessToken string) (assetList response.AssetList, err error)
	MDMGetAssetYear(ctx context.Context, branchID string, assetCode string, search string, prospectID string, accessToken string) (assetMP response.AssetYearList, err error)
	MDMGetMappingLicensePlate(ctx context.Context, licensePlate string, prospectID string, accessToken string) (mdmMasterMappingLicensePlateRes response.MDMMasterMappingLicensePlateResponse, err error)
	CheckBannedChassisNumber(chassisNumber string) (data response.UsecaseApi, err error)
	CheckAgreementChassisNumber(ctx context.Context, prospectID, chassisNumber, idNumber, spouseIDNumber string, accessToken string) (responseAgreementChassisNumber response.AgreementChassisNumber, data response.UsecaseApi, err error)
}

type MultiUsecase interface {
	PrinciplePemohon(ctx context.Context, r request.PrinciplePemohon) (data response.UsecaseApi, err error)
	PrinciplePembiayaan(ctx context.Context, r request.PrinciplePembiayaan, accessToken string) (data response.UsecaseApi, err error)
	PrincipleEmergencyContact(ctx context.Context, req request.PrincipleEmergencyContact, accessToken string) (data response.UsecaseApi, err error)
	GetMaxLoanAmout(ctx context.Context, req request.GetMaxLoanAmount, accessToken string) (data response.GetMaxLoanAmountData, err error)
	GetAvailableTenor(ctx context.Context, req request.GetAvailableTenor, accessToken string) (data []response.GetAvailableTenorData, err error)
	Submission2Wilen(ctx context.Context, req request.Submission2Wilen, accessToken string) (resp response.Submission2Wilen, err error)
}
