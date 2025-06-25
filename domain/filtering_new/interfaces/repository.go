package interfaces

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
)

type Repository interface {
	GetCache(key string) ([]byte, error)
	SetCache(key string, entry []byte) error
	DummyDataPbk(query string) (data entity.DummyPBK, err error)
	SaveFiltering(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, dataCMOnoFPD entity.TrxCmoNoFPD, historyCheckAsset []entity.TrxHistoryCheckingAsset, lockingSystem entity.TrxLockSystem) (err error)
	GetFilteringByID(prospectID string) (row int, err error)
	GetMappingRiskLevel() (data []entity.MappingRiskLevel, err error)
	MasterMappingCluster(req entity.MasterMappingCluster) (data entity.MasterMappingCluster, err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
	GetResultFiltering(prospectID string) (data []entity.ResultFiltering, err error)
	MasterMappingFpdCluster(FpdValue float64) (data entity.MasterMappingFpdCluster, err error)
	CheckCMONoFPD(cmoID string, bpkbName string) (data entity.TrxCmoNoFPD, err error)
	GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error)
	GetAssetCancel(chassisNumber string, engineNumber string, lockSystemConfig response.LockSystemConfig) (historyData response.DataCheckLockAsset, found bool, err error)
	GetAssetReject(chassisNumber string, engineNumber string, lockSystemConfig response.LockSystemConfig) (historyData response.DataCheckLockAsset, found bool, err error)
}
