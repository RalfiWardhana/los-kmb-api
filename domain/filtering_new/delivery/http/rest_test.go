package http

import (
	"github.com/stretchr/testify/mock"
)

type MockEchoContext struct {
	mock.Mock
}

func (m *MockEchoContext) Validate(obj interface{}) error {
	args := m.Called(obj)
	return args.Error(0)
}
