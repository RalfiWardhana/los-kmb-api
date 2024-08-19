package interfaces

import "los-kmb-api/models/entity"

type Repository interface {
	GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error)
	GetMinimalIncomePMK(branchID string, statusKonsumen string) (responseIncomePMK entity.MappingIncomePMK, err error)
	GetDraftPrinciple(prospectID string) (data entity.DraftPrinciple, err error)
	MasterMappingFpdCluster(FpdValue float64) (data entity.MasterMappingFpdCluster, err error)
	MasterMappingCluster(req entity.MasterMappingCluster) (data entity.MasterMappingCluster, err error)
	SaveFiltering(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, dataCMOnoFPD entity.TrxCmoNoFPD) (err error)
	GetBannedPMKDSR(idNumber string) (data entity.TrxBannedPMKDSR, err error)
	GetEncB64(myString string) (encryptedString entity.EncryptedString, err error)
	GetRejection(idNumber string) (data entity.TrxReject, err error)
	GetMappingDukcapilVD(statusVD, customerStatus, customerSegment string, isValid bool) (resultDukcapil entity.MappingResultDukcapilVD, err error)
	GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment string) (resultDukcapil entity.MappingResultDukcapil, err error)
	CheckCMONoFPD(cmoID string, bpkbName string) (data entity.TrxCmoNoFPD, err error)
	SavePrincipleStepOne(data entity.TrxPrincipleStepOne) (err error)
	GetTrxPrincipleStatus(nik string) (data entity.TrxPrincipleStatus, err error)
}
