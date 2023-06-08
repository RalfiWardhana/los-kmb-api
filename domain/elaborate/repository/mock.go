package repository

import (
	"los-kmb-api/models/entity"

	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (mock *MockRepository) SaveDataElaborate(data entity.ApiElaborateKmb) (err error) {
	args := mock.Called(data)
	return args.Error(0)
}

func (mock *MockRepository) UpdateDataElaborate(data entity.ApiElaborateKmbUpdate) (err error) {
	args := mock.Called(data)
	return args.Error(0)
}

func (mock *MockRepository) GetClusterBranchElaborate(query string) (cluster entity.ClusterBranch, err error) {
	args := mock.Called(query)
	return args.Get(0).(entity.ClusterBranch), args.Error(1)
}

func (mock *MockRepository) GetResultElaborate(query string) (data entity.ResultElaborate, err error) {
	args := mock.Called(query)
	return args.Get(0).(entity.ResultElaborate), args.Error(1)
}
