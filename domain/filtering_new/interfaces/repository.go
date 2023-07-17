package interfaces

import "los-kmb-api/models/entity"

type Repository interface {
	DummyDataPbk(query string) (data entity.DummyPBK, err error)
	DataGetMappingDp(branchID, statusKonsumen string) (data []entity.RangeBranchDp, err error)
	BranchDpData(query string) (data entity.BranchDp, err error)
	SaveDupcheckResult(data entity.FilteringKMB) (err error)
	GetFilteringByID(prospectID string) (row int, err error)
}
