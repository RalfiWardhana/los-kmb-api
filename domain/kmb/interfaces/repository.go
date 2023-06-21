package interfaces

import entity "los-kmb-api/models/dupcheck"

type Repository interface {
	GetDupcheckConfig() (config entity.AppConfig, err error)
	GetNewDupcheck(ProspectID string) (data entity.NewDupcheck, err error)
	SaveNewDupcheck(newDupcheck entity.NewDupcheck) (err error)
	GetDummyCustomerDomain(idNumber string) (data entity.DummyCustomerDomain, err error)
	GetDummyLatestPaidInstallment(idNumber string) (data entity.DummyLatestPaidInstallment, err error)
	GetEncryptedValue(idNumber string, legalName string, motherName string) (encrypted entity.Encrypted, err error)

	ScanKmbOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanKmobOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanWgOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanWgOnl(query string) (data entity.ScanInstallmentAmount, err error)
	GetDSRBypass() (config entity.AppConfig, err error)
	GetKMOBOff() (config entity.AppConfig, err error)
}
