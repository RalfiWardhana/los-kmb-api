package repository

import (
	"los-kmb-api/models/entity"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) DummyData(query string) (data entity.DummyColumn, err error) {
	args := m.Called(query)
	return args.Get(0).(entity.DummyColumn), args.Error(1)
}

func (m *MockRepository) DummyDataPbk(query string) (data entity.DummyPBK, err error) {
	args := m.Called(query)
	return args.Get(0).(entity.DummyPBK), args.Error(1)
}

func (m *MockRepository) DataProfessionGroup(query string) (data entity.ProfessionGroup, err error) {
	args := m.Called(query)
	return args.Get(0).(entity.ProfessionGroup), args.Error(1)
}

func (m *MockRepository) DataGetMappingDp(query string) (data []entity.RangeBranchDp, err error) {
	args := m.Called(query)
	return args.Get(0).([]entity.RangeBranchDp), args.Error(1)
}

func (m *MockRepository) BranchDpData(query string) (data entity.BranchDp, err error) {
	args := m.Called(query)
	return args.Get(0).(entity.BranchDp), args.Error(1)
}

func (mock *MockRepository) SaveData(data entity.ApiDupcheckKmb) (err error) {
	args := mock.Called(data)
	return args.Error(0)
}

func (mock *MockRepository) UpdateData(data entity.ApiDupcheckKmbUpdate) (err error) {
	args := mock.Called(data)
	return args.Error(0)
}
