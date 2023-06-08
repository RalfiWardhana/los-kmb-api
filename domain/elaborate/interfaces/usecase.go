package interfaces

import (
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	ResultElaborate(reqs request.BodyRequestElaborate) (data response.ElaborateResult, err error)
}

type MultiUsecase interface {
	Elaborate(reqs request.BodyRequestElaborate, accessToken string) (data response.ElaborateResult, err error)
}
