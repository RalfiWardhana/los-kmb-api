package interfaces

import (
	"los-kmb-api/models/entity"
)

type Repository interface {
	SaveTrxElaborateLTV(data entity.TrxElaborateLTV) (err error)
	GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error)
	GetFilteringDetail(prospectID string) (filtering []entity.TrxDetailBiro, err error)
	GetMappingPBKScoreGrade() (mappingPBKScoreGrade []entity.MappingPBKScoreGrade, err error)
	GetMappingBranchPBK(branchID string, gradePBK string) (mappingBranchByPBKScore entity.MappingBranchByPBKScore, err error)
	GetMappingElaborateLTV(resultPefindo, cluster string, bpkb_name_type int, customerStatus, gradePBK, gradeBranch string) (data []entity.MappingElaborateLTV, err error)
	GetMappingElaborateLTVOvd(resultPefindo, cluster string) (data []entity.MappingElaborateLTV, err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
	GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error)
}
