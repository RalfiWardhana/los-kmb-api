package interfaces

import "los-kmb-api/models/entity"

type Repository interface {
	SaveDataElaborate(data entity.ApiElaborateKmb) (err error)
	UpdateDataElaborate(data entity.ApiElaborateKmbUpdate) (err error)
	GetClusterBranchElaborate(branch_id string, cust_status string, bpkb int) (cluster entity.ClusterBranch, err error)
	GetFilteringResult(prospect_id string) (filtering entity.ApiDupcheckKmbUpdate, err error)
	GetResultElaborate(branch_id string, cust_status string, bpkb int, result_pefindo string, tenor int, age_vehicle string, ltv float64, baki_debet float64) (data entity.ResultElaborate, err error)
}
