package usecase

import (
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/shared/httpclient"
)

type (
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
	}
)

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
	}
}

func (u usecase) GetInquiryPrescreening() {
}
