package repository

import (
	"encoding/json"
	"fmt"
	"los-kmb-api/models/dto"
	"los-kmb-api/shared/authorization/interfaces"
	"los-kmb-api/shared/utils"

	"github.com/jinzhu/gorm"
)

type repoHandler struct {
	losDB *gorm.DB
}

func NewRepository(los *gorm.DB) interfaces.Repository {
	return &repoHandler{
		losDB: los,
	}
}

func (r repoHandler) GetAuth(trx dto.AuthModel) (auth dto.AuthJoinTable, err error) {

	cache := utils.GetCache()

	if cache != nil {

		keyName := fmt.Sprintf("%s_%s", trx.ClientID, trx.Credential)
		getValue, _ := cache.Get(keyName)

		if getValue == nil {

			if err = r.losDB.Raw(fmt.Sprintf("SELECT aac.is_active as client_active, ac.is_active as token_active, ac.access_token, ac.expiry as expired FROM app_auth_clients aac WITH (nolock) LEFT JOIN app_auth_credentials ac WITH (nolock) ON ac.client_id = aac.client_id WHERE aac.client_id = '%s'", trx.ClientID)).Scan(&auth).Error; err != nil {
				return
			}

			value, _ := json.Marshal(auth)

			cache.Set(keyName, value)

		} else {

			json.Unmarshal(getValue, &auth)
		}

	} else {

		if err = r.losDB.Raw(fmt.Sprintf("SELECT aac.is_active as client_active, ac.is_active as token_active, ac.access_token, ac.expiry as expired FROM app_auth_clients aac WITH (nolock) LEFT JOIN app_auth_credentials ac WITH (nolock) ON ac.client_id = aac.client_id WHERE aac.client_id = '%s'", trx.ClientID)).Scan(&auth).Error; err != nil {
			return
		}

	}

	return
}
