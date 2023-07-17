package repository

import (
	"los-kmb-api/models/entity"

	"github.com/stretchr/testify/mock"
)

//counterfeiter:generate . Repository
type MockRepository struct {
	mock.Mock
}

func (mock *MockRepository) GetDupcheckConfig() (config entity.AppConfig, err error) {
	args := mock.Called()
	return args.Get(0).(entity.AppConfig), args.Error(1)
}

func (mock *MockRepository) GetNewDupcheck(prospectID string) (data entity.NewDupcheck, err error) {
	args := mock.Called(prospectID)
	return args.Get(0).(entity.NewDupcheck), args.Error(1)
}

func (mock *MockRepository) SaveNewDupcheck(newDupcheck entity.NewDupcheck) (err error) {
	args := mock.Called(newDupcheck)
	return args.Error(0)
}

func (mock *MockRepository) GetDummyCustomerDomain(idNumber string) (data entity.DummyCustomerDomain, err error) {
	args := mock.Called(idNumber)
	return args.Get(0).(entity.DummyCustomerDomain), args.Error(1)
}

func (mock *MockRepository) GetDummyLatestPaidInstallment(idNumber string) (data entity.DummyLatestPaidInstallment, err error) {
	args := mock.Called(idNumber)
	return args.Get(0).(entity.DummyLatestPaidInstallment), args.Error(1)
}

func (mock *MockRepository) GetEncryptedValue(idNumber string, legalName string, motherName string) (encrypted entity.Encrypted, err error) {
	args := mock.Called()
	return args.Get(0).(entity.Encrypted), args.Error(1)
}
