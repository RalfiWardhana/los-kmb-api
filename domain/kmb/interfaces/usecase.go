package interfaces

import (
	"context"
	entity "los-kmb-api/models/dupcheck"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
)

type Usecase interface {
	DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error)
	BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string)
	VehicleCheck(manufactureYear string) (data response.UsecaseApi, err error)
	CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error)
	PMK(income float64, homeStatus, jobPos, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi)
	DsrCheck(ctx context.Context, prospectID, engineNo string, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, newDupcheck entity.NewDupcheck, accessToken string) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error)
	CustomerDomainGetData(ctx context.Context, req request.ReqCustomerDomain, prospectID, accessToken string) (customerDomainData response.CustomerDomainData, err error)
	GetLatestPaidInstallment(ctx context.Context, req request.ReqLatestPaidInstallment, prospectID, accessToken string) (data response.LatestPaidInstallmentData, err error)
	NokaBanned30D(reqs request.DupcheckApi) (data response.RejectionNoka, err error)
	CheckRejectionNoka(reqs request.DupcheckApi) (data response.RejectionNoka, err error)
	CheckNoka(ctx context.Context, reqs request.DupcheckApi, nokaBanned30D response.RejectionNoka, accessToken string) (data response.UsecaseApi, err error)
	CheckChassisNumber(ctx context.Context, reqs request.DupcheckApi, nokaBanned response.RejectionNoka, accessToken string) (data response.UsecaseApi, err error)
	DecodeMedia(ctx context.Context, url string, customerID int, accessToken string) (base64Image string, err error)
	DetectBlurness(ctx context.Context, selfie1 string) (decision bool, err error)
	FacePlus(ctx context.Context, selfie1 string, selfie2 string, req request.FaceCompareRequest, accessToken string, getPhotoInfo interface{}) (result response.FaceCompareResponse, err error)
}

type MultiUsecase interface {
	GetPhoto(ctx context.Context, req request.FaceCompareRequest, accessToken string) (selfie1 string, selfie2 string, isBlur bool, err error)
}

type Metrics interface {
	Dupcheck(ctx context.Context, reqs request.DupcheckApi, married bool, accessToken string) (mapping response.SpDupcheckMap, status string, data response.UsecaseApi, err error)
}
