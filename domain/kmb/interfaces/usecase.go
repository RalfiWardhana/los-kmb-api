package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	Prescreening(ctx context.Context, reqs request.Metrics, filtering entity.FilteringKMB, accessToken string) (trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, trxDetail entity.TrxDetail, err error)
	RejectTenor36(idNumber string) (result response.UsecaseApi, err error)
	CheckBannedChassisNumber(chassisNumber string) (data response.UsecaseApi, err error)
	CheckBannedPMKDSR(idNumber string) (data response.UsecaseApi, err error)
	CheckRejection(idNumber, prospectID string, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedPMKDSR entity.TrxBannedPMKDSR, err error)
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	VehicleCheck(manufactureYear string, tenor int) (data response.UsecaseApi, err error)
	CheckRejectChassisNumber(req request.DupcheckApi, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedChassisNumber entity.TrxBannedChassisNumber, err error)
	CheckAgreementChassisNumber(ctx context.Context, reqs request.DupcheckApi, accessToken string) (data response.UsecaseApi, err error)
	CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error)
	PMK(branchID string, statusKonsumen string, income float64, homeStatus, professionID, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi, err error)
	DsrCheck(ctx context.Context, req request.DupcheckApi, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, accessToken string) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error)
	GetBase64Media(ctx context.Context, url string, customerID int, accessToken string) (base64Image string, err error)
	SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (resp response.Metrics, err error)
}

type MultiUsecase interface {
	Dupcheck(ctx context.Context, reqs request.DupcheckApi, married bool, accessToken string) (mapping response.SpDupcheckMap, status string, data response.UsecaseApi, trxFMF response.TrxFMF, trxDetail []entity.TrxDetail, err error)
}

type Metrics interface {
	MetricsLos(ctx context.Context, req request.Metrics, accessToken string) (data interface{}, err error)
}
