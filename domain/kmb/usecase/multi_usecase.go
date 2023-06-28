package usecase

import (
	"los-kmb-api/domain/kmb/interfaces"
)

type multiUsecase struct {
	usecase interfaces.Usecase
}

func NewMultiUsecase(usecase interfaces.Usecase) interfaces.MultiUsecase {
	return &multiUsecase{usecase: usecase}
}
