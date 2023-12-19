package interfaces

import "los-kmb-api/models/entity"

type Repository interface {
	SaveDataElaborate(data entity.ApiElaborateKmb) (err error)
	UpdateDataElaborate(data entity.ApiElaborateKmbUpdate) (err error)
	GetClusterBranchElaborate(branchId string, customerStatus string, bpkb int) (cluster entity.ClusterBranch, err error)
	GetFilteringResult(prospect_id string) (filtering entity.ApiDupcheckKmbUpdate, err error)
	GetResultElaborate(branchId string, customerStatus string, bpkb int, resultPefindo string, tenor int, ageVehicle string, ltv float64, bakiDebet float64) (data entity.ResultElaborate, err error)
	GetMappingLtvOvd(cluster, resultPefindo string, tenor int, ltv float64) (data entity.ResultElaborate, err error)
}
