package usecase

import (
	"errors"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
)

func (u usecase) InsertStaging(prospectID string) (data response.InsertStaging, err error) {
	if err = u.repository.SaveToStaging(prospectID); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - " + err.Error())
		return
	}
	data.ProspectID = prospectID
	return
}
