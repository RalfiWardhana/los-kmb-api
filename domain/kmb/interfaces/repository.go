package interfaces

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Repository interface {
	GetTrxReject(idNumber string, config response.LockSystemConfig) (data []entity.TrxLockSystem, err error)
	GetTrxCancel(idNumber string, config response.LockSystemConfig) (data []entity.TrxLockSystem, err error)
	SaveTrxLockSystem(trxLockSystem entity.TrxLockSystem) (err error)
	GetTrxLockSystem(idNumber string, chassisNumber string, engineNumber string) (data entity.TrxLockSystem, bannedType string, err error)
	ScanTrxMaster(prospectID string) (countMaster int, err error)
	ScanTrxPrescreening(prospectID string) (count int, err error)
	GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error)
	GetFilteringForJourney(prospectID string) (filtering entity.FilteringKMB, err error)
	GetMappingDukcapilVD(statusVD, customerStatus, customerSegment string, isValid bool) (resultDukcapilVD entity.MappingResultDukcapilVD, err error)
	GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment string) (resultDukcapil entity.MappingResultDukcapil, err error)
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
	GetMappingDeviasi(prospectID string) (confirmDeviasi entity.ConfirmDeviasi, err error)
	GetMappingNegativeCustomer(req response.NegativeCustomer) (data entity.MappingNegativeCustomer, err error)
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

	MasterMappingCluster(req entity.MasterMappingCluster) (data entity.MasterMappingCluster, err error)
	MasterMappingMaxDSR(req entity.MasterMappingMaxDSR) (data entity.MasterMappingMaxDSR, err error)
	GetEncB64(myString string) (encryptedString entity.EncryptedString, err error)
	GetCurrentTrxWithRejectDSR(idNumber string) (data entity.TrxStatus, err error)
	GetBannedPMKDSR(idNumber string) (data entity.TrxBannedPMKDSR, err error)
	GetCurrentTrxWithReject(idNumber string) (data entity.TrxReject, err error)
	GetBannedChassisNumber(chassisNumber string) (data entity.TrxBannedChassisNumber, err error)
	GetCurrentTrxWithRejectChassisNumber(chassisNumber string) (data []entity.RejectChassisNumber, err error)

	GetRecalculate(prospectID string) (getRecalculate entity.GetRecalculate, err error)
	SaveRecalculate(beforeRecalculate entity.TrxRecalculate, afterRecalculate entity.TrxRecalculate) (err error)
	SaveToStaging(prospectID string) (err error)

	GetMappingVehicleAge(vehicleAge int, cluster string, bpkbNameType, tenor int, resultPefindo string, af float64) (data entity.MappingVehicleAge, err error)

	MasterMappingIncomeMaxDSR(totalIncome float64) (data entity.MasterMappingIncomeMaxDSR, err error)

	MasterMappingDeviasiDSR(totalIncome float64) (data entity.MasterMappingDeviasiDSR, err error)
	GetBranchDeviasi(BranchID string, customerStatus string, NTF float64) (data entity.MappingBranchDeviasi, err error)

	ScanTrxPrinciple(prospectID string) (count int, err error)
	GetPrincipleStepOne(prospectID string) (data entity.TrxPrincipleStepOne, err error)
	GetPrincipleStepTwo(prospectID string) (data entity.TrxPrincipleStepTwo, err error)
	GetPrincipleStepThree(prospectID string) (data entity.TrxPrincipleStepThree, err error)
	GetPrincipleEmergencyContact(prospectID string) (data entity.TrxPrincipleEmergencyContact, err error)

	ScanTrxKPM(prospectID string) (count int, err error)
	GetTrxKPM(prospectID string) (data entity.TrxKPM, err error)
	GetTrxKPMStatus(prospectID string) (data entity.TrxKPMStatus, err error)
	UpdateTrxKPMDecision(id string, prospectID string, decision string) (err error)
}
