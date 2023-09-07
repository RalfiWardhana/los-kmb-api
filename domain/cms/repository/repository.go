package repository

import (
	"fmt"
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"

	"github.com/jinzhu/gorm"
)

type repoHandler struct {
	los  *gorm.DB
	core *gorm.DB
}

func NewRepository(core *gorm.DB, los *gorm.DB) interfaces.Repository {
	return &repoHandler{
		los:  los,
		core: core,
	}
}

func (r repoHandler) GetSpIndustryTypeMaster() (data []entity.SpIndustryTypeMaster, err error) {

	if err = r.core.Raw("exec[spIndustryTypeMaster] '01/01/2007'").Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		err = fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
}
