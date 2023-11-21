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
	GetMappingDukcapil(statusVD, statusFR string) (resultDukcapil entity.MappingResultDukcapil, err error)
	SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (newErr error)
	GetLogOrchestrator(prospectID string) (logOrchestrator entity.LogOrchestrator, err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
	SaveTrxJourney(prospectID string, request interface{}) (err error)
	GetTrxJourney(prospectID string) (trxJourney entity.TrxJourney, err error)
	GetEncryptedValue(idNumber string, legalName string, motherName string) (encrypted entity.Encrypted, err error)

	ScanKmbOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanKmobOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanWgOff(query string) (data entity.ScanInstallmentAmount, err error)
	ScanWgOnl(query string) (data entity.ScanInstallmentAmount, err error)
	GetMinimalIncomePMK(branchID string, statusKonsumen string) (responseIncomePMK entity.MappingIncomePMK, err error)

	GetScoreGenerator(zipCode string) (score entity.ScoreGenerator, err error)
	GetScoreGeneratorROAO() (score entity.ScoreGenerator, err error)
	GetTrxDetailBIro(prospectID string) (trxDetailBiro []entity.TrxDetailBiro, err error)
	GetActiveLoanTypeLast6M(customerID string) (score entity.GetActiveLoanTypeLast6M, err error)
	GetActiveLoanTypeLast24M(customerID string) (score entity.GetActiveLoanTypeLast24M, err error)
	GetMoblast(customerID string) (score entity.GetMoblast, err error)

	GetElaborateLtv(prospectID string) (elaborateLTV entity.MappingElaborateLTV, err error)
	GetMasterBranch(branchID string) (masterBranch entity.MasterBranch, err error)
	GetMappingElaborateIncome(mappingElaborateIncome entity.MappingElaborateIncome) (result entity.MappingElaborateIncome, err error)

	GetLatestBannedRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error)
	GetLatestRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error)
	GetAllReject(idNumber string) (data []entity.DupcheckRejectionPMK, err error)
	GetHistoryRejectAttempt(idNumber string) (data []entity.DupcheckRejectionPMK, err error)
	GetCheckingRejectAttempt(idNumber, blackListDate string) (data entity.DupcheckRejectionPMK, err error)

	SaveDataNoka(data entity.DupcheckRejectionNokaNosin) (err error)
	SaveDataApiLog(data entity.TrxApiLog) (err error)

	GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error)
	SaveVerificationFaceCompare(data entity.VerificationFaceCompare) error

	GetEncB64(myString string) (encryptedString entity.EncryptedString, err error)
	GetCurrentTrxWithRejectDSR(idNumber string) (data entity.TrxStatus, err error)
	GetBannedPMKDSR(idNumber string) (data entity.TrxBannedPMKDSR, err error)
	GetCurrentTrxWithReject(idNumber string) (data entity.TrxReject, err error)
	GetBannedChassisNumber(chassisNumber string) (data entity.TrxBannedChassisNumber, err error)
	GetCurrentTrxWithRejectChassisNumber(chassisNumber string) (data []entity.RejectChassisNumber, err error)

	SaveRecalculate(data entity.TrxRecalculate) (err error)
}
