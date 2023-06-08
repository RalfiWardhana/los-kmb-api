package usecase

import (
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"

	"github.com/stretchr/testify/mock"
)

type MockUsecase struct {
	mock.Mock
}

func (m MockUsecase) FilteringBlackList(reqs request.FilteringRequest, request_id string) (result response.DupcheckResult, err error) {
	args := m.Called(reqs, request_id)
	return args.Get(0).(response.DupcheckResult), args.Error(1)
}

func (m MockUsecase) FilteringKreditmu(reqs request.FilteringRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error) {
	args := m.Called(reqs, status_konsumen, request_id)
	return args.Get(0).(response.DupcheckResult), args.Error(1)
}

func (m MockUsecase) FilteringPefindo(reqs request.FilteringRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error) {
	args := m.Called(reqs, status_konsumen, request_id)
	return args.Get(0).(response.DupcheckResult), args.Error(1)
}
