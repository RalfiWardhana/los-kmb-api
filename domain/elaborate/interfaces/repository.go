package interfaces

import "los-kmb-api/models/entity"

type Repository interface {
	SaveDataElaborate(data entity.ApiElaborateKmb) (err error)
	UpdateDataElaborate(data entity.ApiElaborateKmbUpdate) (err error)
	GetClusterBranchElaborate(query string) (cluster entity.ClusterBranch, err error)
	GetResultElaborate(query string) (data entity.ResultElaborate, err error)
}
