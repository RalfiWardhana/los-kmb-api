package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/dto"
	"los-kmb-api/shared/authorization/interfaces"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/jinzhu/gorm"
)

type repoHandler struct {
	NewKmb *gorm.DB
}

func NewRepository(NewKmb *gorm.DB) interfaces.Repository {
	return &repoHandler{
		NewKmb: NewKmb,
	}
}

func (r repoHandler) GetAuth(req dto.AuthModel) (data dto.AuthJoinTable, err error) {
	cache := utils.GetCache()
	var (
		keyName  string
		getValue []byte
	)

	keyName = fmt.Sprintf("%s_%s", req.ClientID, req.Credential)

	// get cache
	if cache != nil {
		getValue, _ = cache.Get(keyName)
		if getValue != nil {
			json.Unmarshal(getValue, &data)
			if data.ClientActive == 1 {
				return
			}
		}
	}

	// get from db
	if err = r.NewKmb.Raw(fmt.Sprintf("SELECT aac.is_active as client_active, ac.is_active as token_active, ac.access_token, ac.expiry as expired FROM app_auth_clients aac WITH (nolock) LEFT JOIN app_auth_credentials ac WITH (nolock) ON ac.client_id = aac.client_id WHERE aac.client_id = '%s' AND ac.resource_id ='%s' AND ac.access_token = '%s'", req.ClientID, constant.KMB_RESOURCE_ID, req.Credential)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	// set cache
	if data != (dto.AuthJoinTable{}) {
		value, _ := json.Marshal(data)
		cache.Set(keyName, value)
	}

	return
}
