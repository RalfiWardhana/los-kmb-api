package interfaces

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	FilteringBlackList(reqs request.BodyRequest, request_id string) (result response.DupcheckResult, err error)
	FilteringKreditmu(reqs request.BodyRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error)
	FilteringPefindo(reqs request.BodyRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error)
	HitPefindoPrimePriority(reqs request.BodyRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error)
	GetDummyData(noktp string) (config entity.DummyColumn, err error)
	GetDummyPBK(noktp string) (data entity.DummyPBK, err error)
	GetDataProfessionGroup(prefix string) (data entity.ProfessionGroup, err error)
	GetDataGetMappingDp(BranchID, StatusKonsumen string) (data []entity.RangeBranchDp, err error)
	GetBranchDpTest(BranchID, StatusKonsumen, ProfessionGroup string, totalBakiDebetNonAgunan int) (data entity.BranchDp, err error)
	CheckStatusCategory(reqs request.BodyRequest, status_konsumen, request_id string, accessToken string) (data response.DupcheckResult, err error)
}

type MultiUsecase interface {
	Filtering(reqs request.BodyRequest, accessToken string) (data interface{}, err error)
}
