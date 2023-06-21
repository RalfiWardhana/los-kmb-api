package usecase

import (
	entity "los-kmb-api/models/dupcheck"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"

	"github.com/stretchr/testify/mock"
)

type MockUsecase struct {
	mock.Mock
}

type MockMultiUsecase struct {
	mock.Mock
}

func (m MockUsecase) DupcheckIntegrator(prospectID, idNumber, legalName, birthDate, surgateName string) (spDupcheck response.SpDupCekCustomerByID, err error) {
	args := m.Called(prospectID, idNumber, legalName, birthDate, surgateName)
	return args.Get(0).(response.SpDupCekCustomerByID), args.Error(1)
}

func (m MockUsecase) BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string) {
	args := m.Called(index, spDupcheck)
	return args.Get(0).(response.UsecaseApi), args.String(1)
}

func (m MockUsecase) VehicleCheck(manufactureYear string) (data response.UsecaseApi, err error) {
	args := m.Called(manufactureYear)
	return args.Get(0).(response.UsecaseApi), args.Error(1)
}

func (m MockUsecase) CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error) {
	args := m.Called(spDupcheck)
	return args.String(0), args.Error(1)
}

func (m MockUsecase) PMK(income float64, homeStatus, jobPos, empYear, empMonth, stayYear, stayMonth, birthDate string, tenor int, maritalStatus string) (data response.UsecaseApi) {
	args := m.Called(income, homeStatus, jobPos, empYear, empMonth, stayYear, stayMonth, birthDate, tenor, maritalStatus)
	return args.Get(0).(response.UsecaseApi)
}

func (m MockUsecase) DsrCheck(prospectID, engineNo string, customerData []request.CustomerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income float64, newDupcheck entity.NewDupcheck) (data response.UsecaseApi, result response.Dsr, installmentOther, installmentOtherSpouse, installmentTopup float64, err error) {
	args := m.Called(prospectID, engineNo, customerData, installmentAmount, installmentConfins, installmentConfinsSpouse, income, newDupcheck)
	return args.Get(0).(response.UsecaseApi), args.Get(1).(response.Dsr), args.Get(2).(float64), args.Get(3).(float64), args.Get(4).(float64), args.Error(5)
}

func (mock MockUsecase) CustomerDomainGetData(req request.CustomerDomain, accessToken string) (customerDomainData response.CustomerDomainData, err error) {
	args := mock.Called(req, accessToken)
	return args.Get(0).(response.CustomerDomainData), args.Error(1)
}

func (mock MockUsecase) GetLatestPaidInstallment(req request.LatestPaidInstallment) (data response.LatestPaidInstallmentData, err error) {
	args := mock.Called(req)
	return args.Get(0).(response.LatestPaidInstallmentData), args.Error(1)
}
