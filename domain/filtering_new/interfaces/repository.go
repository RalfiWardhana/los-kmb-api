package interfaces

import "los-kmb-api/models/entity"

type Repository interface {
	DummyDataPbk(query string) (data entity.DummyPBK, err error)
	SaveFiltering(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro) (err error)
	GetFilteringByID(prospectID string) (row int, err error)
	MasterMappingCluster(req entity.MasterMappingCluster) (data entity.MasterMappingCluster, err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
	GetResultFiltering(prospectID string) (data []entity.ResultFiltering, err error)
}
