package interfaces

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

type Usecase interface {
	FilteringBlackList(ctx context.Context, reqs request.FilteringRequest, request_id string) (result response.DupcheckResult, err error)
	FilteringKreditmu(ctx context.Context, reqs request.FilteringRequest, status_konsumen, accessToken string) (data response.DupcheckResult, err error)
	FilteringPefindo(ctx context.Context, reqs request.FilteringRequest, status_konsumen, request_id string) (data response.DupcheckResult, err error)
	HitPefindoPrimePriority(ctx context.Context, reqs request.FilteringRequest, status_konsumen, accessToken string) (data response.DupcheckResult, err error)
	GetDummyData(noktp string) (config entity.DummyColumn, err error)
	GetDummyPBK(noktp string) (data entity.DummyPBK, err error)
	GetDataProfessionGroup(prefix string) (data entity.ProfessionGroup, err error)
	GetDataGetMappingDp(BranchID, StatusKonsumen string) (data []entity.RangeBranchDp, err error)
	GetBranchDpTest(BranchID, StatusKonsumen, ProfessionGroup string, totalBakiDebetNonAgunan int) (data entity.BranchDp, err error)
	CheckStatusCategory(ctx context.Context, reqs request.FilteringRequest, status_konsumen, accessToken string) (data response.DupcheckResult, err error)
}

type MultiUsecase interface {
	Filtering(ctx context.Context, reqs request.FilteringRequest, accessToken string) (data interface{}, err error)
}
