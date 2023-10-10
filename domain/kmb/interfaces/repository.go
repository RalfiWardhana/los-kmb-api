package interfaces

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Repository interface {
	ScanTrxMaster(prospectID string) (countMaster int, err error)
	ScanTrxPrescreening(prospectID string) (count int, err error)
	GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error)
	GetFilteringForJourney(prospectID string) (filtering entity.FilteringKMB, err error)
	SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (newErr error)
	GetLogOrchestrator(prospectID string) (logOrchestrator entity.LogOrchestrator, err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
	SaveTrxJourney(prospectID string, request interface{}) (err error)
	GetTrxJourney(prospectID string) (trxJourney entity.TrxJourney, err error)
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
	GetKMBOff() (config entity.AppConfig, err error)
	GetMinimalIncomePMK(branchID string, statusKonsumen string) (responseIncomePMK entity.MappingIncomePMK, err error)
	GetInstallmentAmountChassisNumber(chassisNumber string) (data entity.SpDupcekChasisNo, err error)

	GetLatestBannedRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error)
	GetLatestRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error)
	GetAllReject(idNumber string) (data []entity.DupcheckRejectionPMK, err error)
	GetHistoryRejectAttempt(idNumber string) (data []entity.DupcheckRejectionPMK, err error)
	GetCheckingRejectAttempt(idNumber, blackListDate string) (data entity.DupcheckRejectionPMK, err error)

	SaveDataNoka(data entity.DupcheckRejectionNokaNosin) (err error)
	SaveDataApiLog(data entity.TrxApiLog) (err error)

	GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig)
	SaveVerificationFaceCompare(data entity.VerificationFaceCompare) error

	GetEncB64(myString string) (encryptedString entity.EncryptedString, err error)
	GetCurrentTrxWithRejectDSR(idNumber string) (data entity.TrxStatus, err error)
	GetCurrentTrxWithReject(idNumber string) (data entity.TrxReject, err error)
	ScanPreTrxJourney(prospectID string) (countMaster, countFiltering int, err error)
	GetBiroData(prospectID string) (data entity.FilteringKMB, err error)
}
