package interfaces

import (
	"context"
	entity "los-kmb-api/models/dupcheck"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
)

// check face compare => los-api : domain face compare
// bypass reject no. rangka
// check rejection history
// check rejection no. rangka

type Usecase interface {
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	VehicleCheck(manufactureYear string) (data response.UsecaseApi, err error)
	CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error)
	PMK(income float64, homeStatus, jobPos, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi)
	DsrCheck(ctx context.Context, prospectID, engineNo string, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, newDupcheck entity.NewDupcheck, accessToken string) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error)
	CustomerDomainGetData(ctx context.Context, req request.ReqCustomerDomain, prospectID, accessToken string) (customerDomainData response.CustomerDomainData, err error)
	GetLatestPaidInstallment(ctx context.Context, req request.ReqLatestPaidInstallment, prospectID, accessToken string) (data response.LatestPaidInstallmentData, err error)
	RejectionNoka(engineNo, idNumber string) (data response.UsecaseApi, err error)
}

type MultiUsecase interface {
	Dupcheck(ctx context.Context, reqs request.DupcheckApi, married bool, accessToken string) (mapping response.SpDupcheckMap, status string, data response.UsecaseApi, err error)
}
