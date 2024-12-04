package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	Prescreening(ctx context.Context, reqs request.Metrics, filtering entity.FilteringKMB, accessToken string) (trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, trxDetail entity.TrxDetail, err error)
	RejectTenor36(cluster string) (result response.UsecaseApi, err error)
	CheckBannedChassisNumber(chassisNumber string) (data response.UsecaseApi, err error)
	CheckBannedPMKDSR(idNumber string) (data response.UsecaseApi, err error)
	CheckRejection(idNumber, prospectID string, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedPMKDSR entity.TrxBannedPMKDSR, err error)
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	NegativeCustomerCheck(ctx context.Context, reqs request.DupcheckApi, accessToken string) (data response.UsecaseApi, negativeCustomer response.NegativeCustomer, err error)
	CheckMobilePhoneFMF(ctx context.Context, reqs request.DupcheckApi, accessToken, hrisAccessToken string) (data response.UsecaseApi, err error)
	VehicleCheck(manufactureYear, cmoCluster, bpkbName string, tenor int, configValue response.DupcheckConfig, filteing entity.FilteringKMB, af float64) (data response.UsecaseApi, err error)
	CheckRejectChassisNumber(req request.DupcheckApi, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedChassisNumber entity.TrxBannedChassisNumber, err error)
	CheckAgreementChassisNumber(ctx context.Context, reqs request.DupcheckApi, accessToken string) (data response.UsecaseApi, err error)
	CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error)
	PMK(branchID string, statusKonsumen string, income float64, homeStatus, professionID, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi, err error)
	DsrCheck(ctx context.Context, req request.DupcheckApi, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, accessToken string, configValue response.DupcheckConfig) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error)
	Dukcapil(ctx context.Context, req request.Metrics, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error)
	Asliri(ctx context.Context, req request.Metrics, accessToken string) (data response.Ekyc, err error)
	Ktp(ctx context.Context, req request.Metrics, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error)
	Pefindo(cbFound bool, bpkbName string, filtering entity.FilteringKMB, spDupcheck response.SpDupcheckMap) (data response.UsecaseApi, err error)
	Scorepro(ctx context.Context, req request.Metrics, pefindoScore, customerSegment string, spDupcheck response.SpDupcheckMap, accessToken string, filtering entity.FilteringKMB) (responseScs response.IntegratorScorePro, data response.ScorePro, pefindoIDX response.PefindoIDX, err error)
	ElaborateScheme(req request.Metrics) (data response.UsecaseApi, err error)
	ElaborateIncome(ctx context.Context, req request.Metrics, filtering entity.FilteringKMB, pefindoIDX response.PefindoIDX, spDupcheckMap response.SpDupcheckMap, responseScs response.IntegratorScorePro, accessToken string) (data response.UsecaseApi, err error)
	TotalDsrFmfPbk(ctx context.Context, totalIncome, newInstallment, totalInstallmentPBK float64, prospectID, customerSegment, accessToken string, SpDupcheckMap response.SpDupcheckMap, configValue response.DupcheckConfig, filtering entity.FilteringKMB, NTF float64) (data response.UsecaseApi, trxFMF response.TrxFMF, err error)
	SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (resp response.Metrics, err error)
	CheckAgreementLunas(ctx context.Context, prospectID string, customerId string, filterKMBOnly bool, accessToken string) (responseMDM response.ConfinsAgreementCustomer, isDataExist bool, err error)
	Recalculate(ctx context.Context, req request.Recalculate) (data response.Recalculate, err error)
	InsertStaging(prospectID string) (data response.InsertStaging, err error)
}

type MultiUsecase interface {
	Dupcheck(ctx context.Context, reqs request.DupcheckApi, married bool, accessToken, hrisAccessToken string, configValue response.DupcheckConfig) (mapping response.SpDupcheckMap, status string, data response.UsecaseApi, trxFMF response.TrxFMF, trxDetail []entity.TrxDetail, err error)
	Ekyc(ctx context.Context, req request.Metrics, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, trxDetail []entity.TrxDetail, trxFMF response.TrxFMF, err error)
}

type Metrics interface {
	MetricsLos(ctx context.Context, req request.Metrics, accessToken, hrisAccessToken string) (data interface{}, err error)
}
