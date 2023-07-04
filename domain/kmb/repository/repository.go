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

func (r repoHandler) GetLatestBannedRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error) {

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT * FROM dupcheck_rejection_nokanosin WHERE NoRangka = '%s' AND IsBanned = 1 ORDER BY created_at DESC LIMIT 1", noRangka)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetLatestRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT * FROM dupcheck_rejection_nokanosin WHERE NoRangka = '%s' AND CAST(created_at as DATE) = '%s' ORDER BY created_at DESC LIMIT 1", noRangka, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetAllReject(idNumber string) (data []entity.DupcheckRejectionPMK, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT di.BranchID, di.ProspectID, di.IDNumber, di.DtmUpd, dr.reject_pmk, dr.reject_dsr, fi.final_approval, fi.dtm_final_approval FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) = '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) ORDER BY di.DtmUpd ASC", idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetHistoryRejectAttempt(idNumber string) (data []entity.DupcheckRejectionPMK, err error) {

	searchRange := time.Now().AddDate(0, 0, -30)
	searchRangeString := searchRange.Format("2006-01-02")

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT x.* FROM (SELECT COUNT(*) AS reject_attempt, DATE(fi.dtm_final_approval) as date FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND (dr.reject_pmk IS NOT NULL OR dr.reject_dsr IS NOT NULL) AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) >= '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) GROUP BY date) x ORDER BY x.date DESC LIMIT 1", idNumber, searchRangeString)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetCheckingRejectAttempt(idNumber, blackListDate string) (data entity.DupcheckRejectionPMK, err error) {

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT x.* FROM (SELECT COUNT(*) AS reject_attempt, DATE(fi.dtm_final_approval) as date FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) = '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) GROUP BY date) x ORDER BY x.date DESC", idNumber, blackListDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SaveDataNoka(data entity.DupcheckRejectionNokaNosin) (err error) {
	data.CreatedAt = time.Now()
	if err = r.kmbOffDB.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveDataApiLog(data entity.TrxApiLog) (err error) {
	data.Timestamps = time.Now()
	if err = r.kmbOffDB.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetDummyAgreementChassisNumber(idNumber string) (data entity.DummyAgreementChassisNumber, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT * FROM dummy_agreement_chassis_number WITH (nolock) WHERE id_number = '%s'", idNumber)).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r *repoHandler) GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig) {
	if lob == "" || key == "" {
		if err := r.losDB.
			Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s'", groupName)).
			Scan(&appConfig).Error; err != nil {
			return appConfig
		}

		return
	}
	if err := r.losDB.
		Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s' AND lob = '%s' AND [key]= '%s' AND is_active = 1", groupName, lob, key)).
		Scan(&appConfig).Error; err != nil {
		return appConfig
	}

	return appConfig
}

func (r *repoHandler) SaveVerificationFaceCompare(data entity.VerificationFaceCompare) error {
	if err := r.losDB.Create(&data).Error; err != nil {
		return fmt.Errorf("save verification face compare error: %w", err)
	}
	return nil
}

func (r repoHandler) GetDataInquiry(idNumber string) (data []entity.DataInquiry, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT di.ProspectID, di.IDNumber, di.LegalName, fi.final_approval, CAST(di.DtmUpd as DATE) AS DtmUpd, drp.reject_dsr FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk drp ON (di.ProspectID = drp.ProspectID AND drp.reject_dsr = 1) WHERE di.IDNumber = '%s' AND fi.final_approval IS NOT NULL AND CAST(di.DtmUpd as DATE) = '%s' ORDER BY di.tst DESC", idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}
