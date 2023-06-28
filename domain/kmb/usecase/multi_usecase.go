package usecase

import (
	"context"
	"los-kmb-api/domain/kmb/interfaces"
	request "los-kmb-api/models/dupcheck"
	"los-kmb-api/shared/utils"
)

type multiUsecase struct {
	usecase interfaces.Usecase
}

func NewMultiUsecase(usecase interfaces.Usecase) interfaces.MultiUsecase {
	return &multiUsecase{usecase: usecase}
}

func (u multiUsecase) GetPhoto(ctx context.Context, req request.FaceCompareRequest, accessToken string) (selfie1 string, selfie2 string, isBlur bool, err error) {

	selfie1Media := utils.GetIsMedia(req.ImageSelfie1)

	if selfie1Media {
		selfie1, err = u.usecase.DecodeMedia(ctx, req.ImageSelfie1, req.CustomerID, accessToken)
		if err != nil {
			return
		}

	} else {
		selfie1, err = utils.DecodeNonMedia(req.ImageSelfie1)
		if err != nil {
			return
		}
	}

	isBlur, err = u.usecase.DetectBlurness(ctx, selfie1)

	if err != nil {
		return
	}

	selfie2Media := utils.GetIsMedia(req.ImageSelfie2)

	if selfie2Media {
		selfie2, err = u.usecase.DecodeMedia(ctx, req.ImageSelfie2, req.CustomerID, accessToken)
		if err != nil {
			return
		}

	} else {
		selfie2, err = utils.DecodeNonMedia(req.ImageSelfie2)
		if err != nil {
			return
		}
	}
	return
}
