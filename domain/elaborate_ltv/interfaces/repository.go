package interfaces

import (
	"los-kmb-api/models/entity"
)

type Repository interface {
	SaveTrxElaborateLTV(data entity.TrxElaborateLTV) (err error)
	GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error)
	GetMappingElaborateLTV(resultPefindo, cluster string) (data []entity.MappingElaborateLTV, err error)
	GetMappingElaborateLTVOvd(resultPefindo, cluster string) (data []entity.MappingElaborateLTV, err error)
	SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error)
}
