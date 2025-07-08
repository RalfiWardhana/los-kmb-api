package usecase

import (
	"context"
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecalculate(t *testing.T) {

	tenor := 12
	testcases := []struct {
		name              string
		req               request.Recalculate
		errGetRecalculate error
		result            response.Recalculate
		beforeRec         entity.GetRecalculate
		afterRec          entity.TrxRecalculate
		errFinal          error
	}{
		{
			name: "recalculate",
			req: request.Recalculate{
				ProspectID:                   "TEST1",
				Tenor:                        12,
				ProductOfferingID:            "12Ee12",
				ProductOfferingDesc:          "product offering",
				DPAmount:                     500000,
				NTF:                          12000000,
				AF:                           10000000,
				AdminFee:                     5000,
				InstallmentAmount:            100000,
				PercentDP:                    12.34,
				LifePremiumAmountToCustomer:  12000,
				AssetPremiumAmountToCustomer: 0,
				FidusiaFee:                   0,
				InterestRate:                 1.2,
				InterestAmount:               100000,
				ProvisionFee:                 0,
				LoanAmount:                   11000000,
			},
			beforeRec: entity.GetRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            200000,
				NTF:                 10000000,
				NTFAkumulasi:        15000000,
				AF:                  10000000,
				AdminFee:            5000,
				InstallmentAmount:   700000,
				PercentDP:           12.34,
				AssetInsuranceFee:   12000,
				LifeInsuranceFee:    0,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				TotalInstallmentFMF: 800000,
				TotalIncome:         15000000,
				DSRFMF:              21.1,
				TotalDSR:            21.1,
			},
			afterRec: entity.TrxRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            500000,
				InstallmentAmount:   100000,
				NTF:                 12000000,
				NTFAkumulasi:        17000000,
				AF:                  10000000,
				AdminFee:            5000,
				PercentDP:           12.34,
				AssetInsuranceFee:   0,
				LifeInsuranceFee:    12000,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				DSRFMF:              0.06,
				TotalDSR:            0.06,
			},
			result: response.Recalculate{
				ProspectID: "TEST1",
			},
		},
		{
			name:              "recalculate err get",
			errGetRecalculate: errors.New("error get"),
			errFinal:          errors.New(fmt.Sprintf("%s - %s", constant.ERROR_UPSTREAM, "GetRecalculate Error")),
			req: request.Recalculate{
				ProspectID:                   "TEST1",
				Tenor:                        12,
				ProductOfferingID:            "12Ee12",
				ProductOfferingDesc:          "product offering",
				DPAmount:                     500000,
				NTF:                          12000000,
				AF:                           10000000,
				AdminFee:                     5000,
				InstallmentAmount:            100000,
				PercentDP:                    12.34,
				LifePremiumAmountToCustomer:  12000,
				AssetPremiumAmountToCustomer: 0,
				FidusiaFee:                   0,
				InterestRate:                 1.2,
				InterestAmount:               100000,
				ProvisionFee:                 0,
				LoanAmount:                   11000000,
			},
			beforeRec: entity.GetRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            200000,
				NTF:                 10000000,
				NTFAkumulasi:        15000000,
				AF:                  10000000,
				AdminFee:            5000,
				InstallmentAmount:   700000,
				PercentDP:           12.34,
				AssetInsuranceFee:   12000,
				LifeInsuranceFee:    0,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				TotalInstallmentFMF: 800000,
				TotalIncome:         15000000,
				DSRFMF:              21.1,
				TotalDSR:            21.1,
			},
			afterRec: entity.TrxRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            500000,
				InstallmentAmount:   100000,
				NTF:                 12000000,
				NTFAkumulasi:        17000000,
				AF:                  10000000,
				AdminFee:            5000,
				PercentDP:           12.34,
				AssetInsuranceFee:   0,
				LifeInsuranceFee:    12000,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				DSRFMF:              0.06,
				TotalDSR:            0.06,
			},
		},
		{
			name: "recalculate pbk",
			req: request.Recalculate{
				ProspectID:                   "TEST1",
				Tenor:                        12,
				ProductOfferingID:            "12Ee12",
				ProductOfferingDesc:          "product offering",
				DPAmount:                     500000,
				NTF:                          12000000,
				AF:                           10000000,
				AdminFee:                     5000,
				InstallmentAmount:            100000,
				PercentDP:                    12.34,
				LifePremiumAmountToCustomer:  12000,
				AssetPremiumAmountToCustomer: 0,
				FidusiaFee:                   0,
				InterestRate:                 1.2,
				InterestAmount:               100000,
				ProvisionFee:                 0,
				LoanAmount:                   11000000,
			},
			beforeRec: entity.GetRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            200000,
				NTF:                 10000000,
				NTFAkumulasi:        15000000,
				AF:                  10000000,
				AdminFee:            5000,
				InstallmentAmount:   700000,
				PercentDP:           12.34,
				AssetInsuranceFee:   12000,
				LifeInsuranceFee:    0,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				TotalInstallmentFMF: 800000,
				TotalIncome:         15000000,
				DSRFMF:              11.1,
				DSRPBK:              10.1,
				TotalDSR:            21.1,
			},
			afterRec: entity.TrxRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            500000,
				InstallmentAmount:   100000,
				NTF:                 12000000,
				NTFAkumulasi:        17000000,
				AF:                  10000000,
				AdminFee:            5000,
				PercentDP:           12.34,
				AssetInsuranceFee:   0,
				LifeInsuranceFee:    12000,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				DSRFMF:              0.06,
				TotalDSR:            10.16,
			},
			result: response.Recalculate{
				ProspectID: "TEST1",
			},
		},
		{
			name:     "recalculate err get",
			errFinal: errors.New(fmt.Sprintf("%s - %s", constant.ERROR_UPSTREAM, "GetRecalculate Error")),
			req: request.Recalculate{
				ProspectID:                   "TEST1",
				Tenor:                        12,
				ProductOfferingID:            "12Ee12",
				ProductOfferingDesc:          "product offering",
				DPAmount:                     500000,
				NTF:                          12000000,
				AF:                           10000000,
				AdminFee:                     5000,
				InstallmentAmount:            100000,
				PercentDP:                    12.34,
				LifePremiumAmountToCustomer:  12000,
				AssetPremiumAmountToCustomer: 0,
				FidusiaFee:                   0,
				InterestRate:                 1.2,
				InterestAmount:               100000,
				ProvisionFee:                 0,
				LoanAmount:                   11000000,
			},
			beforeRec: entity.GetRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            200000,
				NTF:                 10000000,
				NTFAkumulasi:        15000000,
				AF:                  10000000,
				AdminFee:            5000,
				InstallmentAmount:   700000,
				PercentDP:           12.34,
				AssetInsuranceFee:   12000,
				LifeInsuranceFee:    0,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				TotalInstallmentFMF: 800000,
				TotalIncome:         15000000,
				DSRFMF:              21.1,
				TotalDSR:            21.1,
			},
			afterRec: entity.TrxRecalculate{
				ProspectID:          "TEST1",
				Tenor:               &tenor,
				ProductOfferingID:   "12Ee12",
				ProductOfferingDesc: "product offering",
				DPAmount:            500000,
				InstallmentAmount:   100000,
				NTF:                 12000000,
				NTFAkumulasi:        17000000,
				AF:                  10000000,
				AdminFee:            5000,
				PercentDP:           12.34,
				AssetInsuranceFee:   0,
				LifeInsuranceFee:    12000,
				FidusiaFee:          0,
				InterestRate:        1.2,
				InterestAmount:      100000,
				ProvisionFee:        0,
				LoanAmount:          11000000,
				DSRFMF:              0.06,
				TotalDSR:            0.06,
			},
		},
	}

	ctx := context.Background()

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			saveBeforeRecalculate := entity.TrxRecalculate{
				ProspectID:          tc.req.ProspectID,
				ProductOfferingID:   tc.beforeRec.ProductOfferingID,
				ProductOfferingDesc: tc.beforeRec.ProductOfferingDesc,
				Tenor:               tc.beforeRec.Tenor,
				LoanAmount:          tc.beforeRec.LoanAmount,
				AF:                  tc.beforeRec.AF,
				InstallmentAmount:   tc.beforeRec.InstallmentAmount,
				DPAmount:            tc.beforeRec.DPAmount,
				PercentDP:           tc.beforeRec.PercentDP,
				AdminFee:            tc.beforeRec.AdminFee,
				ProvisionFee:        tc.beforeRec.ProvisionFee,
				FidusiaFee:          tc.beforeRec.FidusiaFee,
				AssetInsuranceFee:   tc.beforeRec.AssetInsuranceFee,
				LifeInsuranceFee:    tc.beforeRec.LifeInsuranceFee,
				NTF:                 tc.beforeRec.NTF,
				NTFAkumulasi:        tc.beforeRec.NTFAkumulasi,
				InterestRate:        tc.beforeRec.InterestRate,
				InterestAmount:      tc.beforeRec.InterestAmount,
				DSRFMF:              tc.beforeRec.DSRFMF,
				TotalDSR:            tc.beforeRec.TotalDSR,
			}

			mockRepository.On("GetRecalculate", tc.req.ProspectID).Return(tc.beforeRec, tc.errGetRecalculate)
			mockRepository.On("SaveRecalculate", saveBeforeRecalculate, tc.afterRec, tc.req).Return(tc.errFinal)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, err := usecase.Recalculate(ctx, tc.req)
			require.Equal(t, tc.result, data)
			require.Equal(t, tc.errFinal, err)
		})
	}
}
