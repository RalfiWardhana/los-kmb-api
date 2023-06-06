package utils

import (
	"fmt"
	"los-kmb-api/models/entity"

	"github.com/allegro/bigcache/v3"
	"github.com/jinzhu/gorm"
)

var (
	handler handlerCache
)

type handlerCache struct {
	db            *gorm.DB
	cache         *bigcache.BigCache
	isDevelopment bool
}

func NewCache(cache *bigcache.BigCache, db *gorm.DB, env bool) {

	handler = handlerCache{
		cache:         cache,
		db:            db,
		isDevelopment: env,
	}

}

func GetCache() *bigcache.BigCache {
	return handler.cache
}

func ValidatorFromCache(keyName string) (model entity.AppConfig, err error) {

	if handler.cache != nil {

		getValue, _ := handler.cache.Get(keyName)

		if getValue == nil {

			if err = handler.db.Raw(fmt.Sprintf("SELECT * FROM app_config WITH (nolock) WHERE group_name = '%s' AND is_active = 1", keyName)).Scan(&model).Error; err != nil {
				return
			}
			handler.cache.Set(keyName, []byte(model.Value))
			if handler.isDevelopment {
				fmt.Println("From DB: ", model.Value)
			}

		} else {
			model.Value = string(getValue)
			if handler.isDevelopment {
				fmt.Println("From Cache: ", model.Value)
			}
		}

	} else {

		if err = handler.db.Raw(fmt.Sprintf("SELECT * FROM app_config WITH (nolock) WHERE group_name = '%s' AND is_active = 1", keyName)).Scan(&model).Error; err != nil {
			if handler.isDevelopment {
				fmt.Println("handler.cache: ", handler.cache)
			}
			return
		}
	}

	return
}
