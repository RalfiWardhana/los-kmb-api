package usecase

import (
	"context"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
)

func (u usecase) Recalculate(ctx context.Context, req request.Recalculate) (data response.Recalculate, err error) {

	var (
		afterRec  entity.TrxRecalculate
		beforeRec entity.GetRecalculate
	)

	beforeRec, err = u.repository.GetRecalculate(req.ProspectID)
	if err != nil {
		err = errors.New(fmt.Sprintf("%s - %s", constant.ERROR_UPSTREAM, "GetRecalculate Error"))
		return
	}

	saveBeforeRecalculate := entity.TrxRecalculate{
		ProspectID:          req.ProspectID,
		ProductOfferingID:   beforeRec.ProductOfferingID,
		ProductOfferingDesc: beforeRec.ProductOfferingDesc,
		Tenor:               beforeRec.Tenor,
		LoanAmount:          beforeRec.LoanAmount,
		AF:                  beforeRec.AF,
		InstallmentAmount:   beforeRec.InstallmentAmount,
		DPAmount:            beforeRec.DPAmount,
		PercentDP:           beforeRec.PercentDP,
		AdminFee:            beforeRec.AdminFee,
		ProvisionFee:        beforeRec.ProvisionFee,
		FidusiaFee:          beforeRec.FidusiaFee,
		AssetInsuranceFee:   beforeRec.AssetInsuranceFee,
		LifeInsuranceFee:    beforeRec.LifeInsuranceFee,
		NTF:                 beforeRec.NTF,
		NTFAkumulasi:        beforeRec.NTFAkumulasi,
		InterestRate:        beforeRec.InterestRate,
		InterestAmount:      beforeRec.InterestAmount,
		DSRFMF:              beforeRec.DSRFMF,
		TotalDSR:            beforeRec.TotalDSR,
	}

	afterRec = entity.TrxRecalculate{
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

	totalInstallment := afterRec.InstallmentAmount + beforeRec.TotalInstallmentFMF
	afterRec.DSRFMF = totalInstallment / beforeRec.TotalIncome

	if beforeRec.DSRFMF == beforeRec.TotalDSR {
		afterRec.TotalDSR = afterRec.DSRFMF
	} else {
		afterRec.TotalDSR = afterRec.DSRFMF + beforeRec.DSRPBK
	}

	afterRec.NTFAkumulasi = afterRec.NTF + (beforeRec.NTFAkumulasi - beforeRec.NTF)

	err = u.repository.SaveRecalculate(saveBeforeRecalculate, afterRec, req)
	if err == nil {
		data.ProspectID = req.ProspectID
	}

	return
}
