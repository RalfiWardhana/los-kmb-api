package usecase

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

func (u usecase) ApprovalScheme(req request.ReqApprovalScheme) (app response.RespApprovalScheme, err error) {

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

	final := "COM"

	var nextStep string
	isFinal := false

	for i, v := range limit {
		if req.DecisionAlias == v.Alias {
			// add next
			if req.DecisionAlias != final {
				nextStep = limit[i+1].Alias
				break
			} else {
				isFinal = true
				break
			}
		}
	}

	app = response.RespApprovalScheme{
		NextStep: nextStep,
		IsFinal:  isFinal,
	}

	return
}
