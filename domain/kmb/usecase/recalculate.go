package usecase

import (
	"context"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

func (u usecase) Recalculate(ctx context.Context, req request.Recalculate) (data response.Recalculate, err error) {

	data.ProspectID = req.ProspectID
	trxRecalculate := entity.TrxRecalculate{
		ProspectID:          req.ProspectID,
		ProductOfferingID:   req.ProductOfferingID,
		ProductOfferingDesc: req.ProductOfferingDesc,
		Tenor:               &req.Tenor,
		LoanAmount:          req.LoanAmount,
		AF:                  req.AF,
		InstallmentAmount:   req.InstallmentAmount,
		DPAmount:            req.DPAmount,
		PercentDP:           req.PercentDP,
		AdminFee:            req.AdminFee,
		ProvisionFee:        req.ProvisionFee,
		FidusiaFee:          req.FidusiaFee,
		AssetInsuranceFee:   req.AssetPremiumAmountToCustomer,
		LifeInsuranceFee:    req.LifePremiumAmountToCustomer,
		NTF:                 req.NTF,
		InterestRate:        req.InterestRate,
		InterestAmount:      req.InterestAmount,
	}

	err = u.repository.SaveRecalculate(trxRecalculate)
	return
}
