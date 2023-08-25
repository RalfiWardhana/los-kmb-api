package usecase

import (
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"

	"github.com/stretchr/testify/mock"
)

type MockUsecase struct {
	mock.Mock
}

func (m MockUsecase) ResultElaborate(reqs request.BodyRequestElaborate) (data response.ElaborateResult, err error) {
	args := m.Called(reqs)
	return args.Get(0).(response.ElaborateResult), args.Error(1)
}
