package repository

import (
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces"
	entity "los-kmb-api/models/dupcheck"
	"los-kmb-api/shared/constant"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	losDB     *gorm.DB
	logsDB    *gorm.DB
	confinsDB *gorm.DB
	stagingDB *gorm.DB
	wgOffDB   *gorm.DB
	kmbOffDB  *gorm.DB
}

func NewRepository(los, logs, confins, staging, wgOff, kmbOff *gorm.DB) interfaces.Repository {
	return &repoHandler{
		losDB:     los,
		logsDB:    logs,
		confinsDB: confins,
		stagingDB: staging,
		wgOffDB:   wgOff,
		kmbOffDB:  kmbOff,
	}
}

func (r repoHandler) GetDupcheckConfig() (config entity.AppConfig, err error) {

	if err = r.losDB.Raw("SELECT [key], [value] FROM app_config WHERE lob = 'KMOB-OFF' AND [key] = 'dupcheck_kmob_config' AND group_name = 'dupcheck'").Scan(&config).Error; err != nil {
		return
	}

	if config == (entity.AppConfig{}) {
		err = errors.New(constant.ERROR_NOT_FOUND)
		return
	}
	return
}

func (r repoHandler) GetNewDupcheck(prospectID string) (data entity.NewDupcheck, err error) {

	if err = r.losDB.Raw(fmt.Sprintf("SELECT TOP 1 ProspectID, customer_status, customer_type FROM new_dupcheck WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC", prospectID)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SaveNewDupcheck(newDupcheck entity.NewDupcheck) (err error) {

	if err = r.losDB.Create(&newDupcheck).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetDummyCustomerDomain(idNumber string) (data entity.DummyCustomerDomain, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT * FROM dummy_customer_domain WITH (nolock) WHERE id_number = '%s'", idNumber)).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetDummyLatestPaidInstallment(idNumber string) (data entity.DummyLatestPaidInstallment, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT * FROM dummy_latest_paid_installment WITH (nolock) WHERE id_number = '%s'", idNumber)).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetEncryptedValue(idNumber string, legalName string, motherName string) (encrypted entity.Encrypted, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName, SCP.dbo.ENC_B64('SEC','%s') AS SurgateMotherName, SCP.dbo.ENC_B64('SEC','%s') AS IDNumber`,
		legalName, motherName, idNumber)).Scan(&encrypted).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetDSRBypass() (config entity.AppConfig, err error) {

	if err = r.losDB.Raw("SELECT [key], [value] FROM app_config WHERE lob = 'KMOB-OFF' AND [key] = 'dsr-bypass' AND group_name = 'dsr_setting'").Scan(&config).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) ScanWgOff(query string) (data entity.ScanInstallmentAmount, err error) {
	if err = r.wgOffDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanKmbOff(query string) (data entity.ScanInstallmentAmount, err error) {
	if err = r.kmbOffDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanKmobOff(query string) (data entity.ScanInstallmentAmount, err error) {

	if err = r.losDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanWgOnl(query string) (data entity.ScanInstallmentAmount, err error) {

	if err = r.losDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) GetKMOBOff() (config entity.AppConfig, err error) {

	if err = r.losDB.Raw("SELECT [key], [value] FROM app_config WHERE lob = 'KMOB-OFF' AND [key] = 'pmk_kmob_off' AND group_name = 'pmk_config'").Scan(&config).Error; err != nil {
		return
	}

	return
}
