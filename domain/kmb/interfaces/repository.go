package interfaces

import entity "los-kmb-api/models/dupcheck"

type Repository interface {
	GetDupcheckConfig() (config entity.AppConfig, err error)
	GetNewDupcheck(ProspectID string) (data entity.NewDupcheck, err error)
	SaveNewDupcheck(newDupcheck entity.NewDupcheck) (err error)
	GetDummyCustomerDomain(idNumber string) (data entity.DummyCustomerDomain, err error)
	GetDummyLatestPaidInstallment(idNumber string) (data entity.DummyLatestPaidInstallment, err error)
	GetDummyAgreementChassisNumber(idNumber string) (data entity.DummyAgreementChassisNumber, err error)
	GetEncryptedValue(idNumber string, legalName string, motherName string) (encrypted entity.Encrypted, err error)

	ScanKmbOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanKmobOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanWgOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanWgOnl(query string) (data entity.ScanInstallmentAmount, err error)
	GetDSRBypass() (config entity.AppConfig, err error)
	GetKMOBOff() (config entity.AppConfig, err error)

	GetLatestBannedRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error)
	GetLatestRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error)
	GetAllReject(idNumber string) (data []entity.DupcheckRejectionPMK, err error)
	GetHistoryRejectAttempt(idNumber string) (data []entity.DupcheckRejectionPMK, err error)
	GetCheckingRejectAttempt(idNumber, blackListDate string) (data entity.DupcheckRejectionPMK, err error)

	SaveDataNoka(data entity.DupcheckRejectionNokaNosin) (err error)
	SaveDataApiLog(data entity.TrxApiLog) (err error)

	GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig)
	SaveVerificationFaceCompare(data entity.VerificationFaceCompare) error
}
