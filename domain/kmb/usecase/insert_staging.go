package usecase

import (
	"errors"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"strings"
)

func (u usecase) InsertStaging(prospectID string) (data response.InsertStaging, err error) {
	if err = u.repository.SaveToStaging(prospectID); err != nil {
		if !strings.Contains(err.Error(), "duplicate") {
			err = errors.New(constant.ERROR_UPSTREAM + " - " + err.Error())
			return
		}
	}
	data.ProspectID = prospectID
	return
}
