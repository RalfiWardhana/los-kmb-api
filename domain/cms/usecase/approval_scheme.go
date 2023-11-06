package usecase

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

func (u usecase) ApprovalScheme(req request.ReqSubmitApproval) (result response.RespApprovalScheme, err error) {

	// get master limit
	limit := []entity.MappingLimitApprovalScheme{
		{
			Alias: "CBM",
		},
		{
			Alias: "DRM",
		},
		{
			Alias: "GMO",
		},
		{
			Alias: "COM",
		},
		{
			Alias: "GMC",
		},
		{
			Alias: "UCC",
		},
	}

	for i, v := range limit {
		if req.Alias == v.Alias {
			// add next
			if req.Alias != req.FinalApproval {
				result.NextStep = limit[i+1].Alias
				break
			} else {
				if req.NeedEscalation {
					result.NextStep = limit[i+1].Alias
					result.IsEscalation = true
				} else {
					result.IsFinal = true
				}
				break
			}
		}
	}
	return
}
