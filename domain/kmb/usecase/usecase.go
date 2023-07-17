package usecase

import (
	"los-kmb-api/domain/kmb/interfaces"
	"los-kmb-api/shared/httpclient"
)

type (
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
	}
	multiUsecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		usecase    interfaces.Usecase
	}
	metrics struct {
		repository   interfaces.Repository
		httpclient   httpclient.HttpClient
		usecase      interfaces.Usecase
		multiUsecase interfaces.MultiUsecase
	}
)

func NewMultiUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, usecase interfaces.Usecase) interfaces.MultiUsecase {
	return &multiUsecase{
		repository: repository,
		httpclient: httpclient,
		usecase:    usecase,
	}
}

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
	}
}

func NewMetrics(repository interfaces.Repository, httpclient httpclient.HttpClient, usecase interfaces.Usecase, multiUsecase interfaces.MultiUsecase) interfaces.Metrics {
	return &metrics{
		repository:   repository,
		httpclient:   httpclient,
		usecase:      usecase,
		multiUsecase: multiUsecase,
	}
}
