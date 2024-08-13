package usecase

import (
	"fmt"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPefindo(t *testing.T) {
	os.Setenv("NAMA_SAMA", "K,P")
	testcases := []struct {
		name       string
		cbFound    bool
		bpkbName   string
		filtering  entity.FilteringKMB
		spDupcheck response.SpDupcheckMap
		req        request.Metrics
		result     response.UsecaseApi
		errResult  error
	}{
		{
			name: "Pefindo prime",
			filtering: entity.FilteringKMB{
				CustomerSegment: constant.RO_AO_PRIME,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen:          constant.STATUS_KONSUMEN_AO,
				InstallmentTopup:        0,
				NumberOfPaidInstallment: 6,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_PRIME_PRIORITY,
				Reason:         fmt.Sprintf("%s %s >= 6 bulan - PBK Pass", constant.STATUS_KONSUMEN_AO, constant.RO_AO_PRIME),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name: "Pefindo prime",
			filtering: entity.FilteringKMB{
				CustomerSegment: constant.RO_AO_PRIORITY,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_PRIME_PRIORITY,
				Reason:         fmt.Sprintf("%s %s - PBK Pass", constant.STATUS_KONSUMEN_RO, constant.RO_AO_PRIORITY),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo Reject BPKB nama sama pass",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                13,
				MaxOverdueLast12monthsBiro:    65,
				MaxOverdueKORules:             13,
				MaxOverdueLast12MonthsKORules: 65,
				Category:                      2,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, "(II)", constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo Reject BPKB nama sama baki debet reject",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  13,
				MaxOverdueLast12monthsBiro:      65,
				TotalBakiDebetNonCollateralBiro: 22000000,
				MaxOverdueKORules:               13,
				MaxOverdueLast12MonthsKORules:   65,
				Category:                        1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
				Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, "(I)"),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo PASS cat != 3 BPKB nama sama baki debet reject",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  13,
				MaxOverdueLast12monthsBiro:      65,
				TotalBakiDebetNonCollateralBiro: 22000000,
				MaxOverdueKORules:               67,
				MaxOverdueLast12MonthsKORules:   12,
				Category:                        1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
				Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, "(I)"),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo Reject BPKB nama sama ada wo agunan reject",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                13,
				MaxOverdueLast12monthsBiro:    65,
				IsWoContractBiro:              1,
				IsWoWithCollateralBiro:        1,
				MaxOverdueKORules:             13,
				MaxOverdueLast12MonthsKORules: 65,
				Category:                      1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, "(I)", constant.ADA_FASILITAS_WO_AGUNAN),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo Reject BPKB nama sama ada wo tidak ada agunan baki debet reject",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  13,
				MaxOverdueLast12monthsBiro:      65,
				IsWoContractBiro:                1,
				TotalBakiDebetNonCollateralBiro: 22000000,
				MaxOverdueKORules:               13,
				MaxOverdueLast12MonthsKORules:   65,
				Category:                        1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
				Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, "(I)"),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo Reject BPKB nama sama ada wo tidak ada agunan baki debet pass",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                13,
				MaxOverdueLast12monthsBiro:    65,
				IsWoContractBiro:              1,
				MaxOverdueKORules:             13,
				MaxOverdueLast12MonthsKORules: 65,
				Category:                      1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
				Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, "(I)"),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo Reject BPKB nama beda reject",
			cbFound:  true,
			bpkbName: "KK",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                31,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             31,
				MaxOverdueLast12MonthsKORules: 9,
				Category:                      1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_BPKB_BEDA,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, "(I)", fmt.Sprintf(constant.REJECT_REASON_OVD_PEFINDO)),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo PASS tidak ada WO",
			cbFound:  true,
			bpkbName: "KK",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 9,
				Category:                      1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_BEDA_NO_FACILITY_WO_CODE,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, "(I)", constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo not found",
			cbFound:  false,
			bpkbName: "KK",
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_NO,
				Reason:         constant.REASON_PEFINDO_NOTFOUND,
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_PASS_konsumen_NEW_nama_sama_ada_WO",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 9,
				Category:                      1,
				IsWoContractBiro:              1,
				IsWoWithCollateralBiro:        1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, "(I)", constant.ADA_FASILITAS_WO_AGUNAN),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_reject_cat3_konsumen_NEW_nama_sama_ada_WO",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 90,
				Category:                      3,
				IsWoContractBiro:              1,
				IsWoWithCollateralBiro:        1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, "(III)", constant.ADA_FASILITAS_WO_AGUNAN),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_reject_cat3_konsumen_NEW_nama_sama_ada_wo_baki_debet_tidak_sesuai",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  9,
				MaxOverdueLast12monthsBiro:      9,
				MaxOverdueKORules:               9,
				MaxOverdueLast12MonthsKORules:   90,
				Category:                        3,
				IsWoContractBiro:                1,
				TotalBakiDebetNonCollateralBiro: 9000000000,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
				Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, "(III)"),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_reject_cat3_konsumen_NEW_nama_sama_ada_wo_baki_debet_sesuai",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 90,
				Category:                      3,
				IsWoContractBiro:              1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
				Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_SESUAI_BNPL, "(III)"),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_reject_cat3_konsumen_NEW_nama_sama_tidak_ada_wo_baki_debet_tidak_sesuai",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  9,
				MaxOverdueLast12monthsBiro:      9,
				MaxOverdueKORules:               9,
				MaxOverdueLast12MonthsKORules:   90,
				Category:                        3,
				TotalBakiDebetNonCollateralBiro: 900000000000000000,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
				Reason:         fmt.Sprintf(constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI_BNPL, "(III)"),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_reject_cat3_konsumen_NEW_nama_sama_tidak_ada_wo_baki_debet_sesuai",
			cbFound:  true,
			bpkbName: "K",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 90,
				Category:                      3,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_SAMA, "(III)", constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_pass_cat1_konsumen_NEW_nama_beda_ada_wo_ada_agunan",
			cbFound:  true,
			bpkbName: "KK",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 9,
				Category:                      3,
				IsWoContractBiro:              1,
				IsWoWithCollateralBiro:        1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_BEDA_WO_AGUNAN_REJECT_CODE,
				Reason:         fmt.Sprintf("%s %s & %s", constant.REASON_BPKB_BEDA, "(III)", constant.ADA_FASILITAS_WO_AGUNAN),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_pass_cat1_konsumen_NEW_nama_beda_ada_wo_baki_debet_tidak_sesuai",
			cbFound:  true,
			bpkbName: "KK",
			filtering: entity.FilteringKMB{
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  9,
				MaxOverdueLast12monthsBiro:      9,
				MaxOverdueKORules:               9,
				MaxOverdueLast12MonthsKORules:   9,
				Category:                        3,
				IsWoContractBiro:                1,
				TotalBakiDebetNonCollateralBiro: 90000000000000,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_GT20J,
				Reason:         fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, "(III)"),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_pass_cat1_konsumen_NEW_nama_beda_ada_wo_baki_debet_sesuai",
			cbFound:  true,
			bpkbName: "KK",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 9,
				Category:                      3,
				IsWoContractBiro:              1,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_LTE20J,
				Reason:         fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_SESUAI_BNPL, "(III)"),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo_pass_cat1_konsumen_NEW_nama_beda_tidak_ada_wo_baki_debet_tidak_sesuai",
			cbFound:  true,
			bpkbName: "KK",
			filtering: entity.FilteringKMB{
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  9,
				MaxOverdueLast12monthsBiro:      9,
				MaxOverdueKORules:               9,
				MaxOverdueLast12MonthsKORules:   9,
				Category:                        3,
				TotalBakiDebetNonCollateralBiro: 900000000000000,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_BEDA_BAKI_DEBET_GT20J,
				Reason:         fmt.Sprintf(constant.NAMA_BEDA_BAKI_DEBET_TIDAK_SESUAI_BNPL, "(III)"),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			// mockRepository.On("GetElaborateLtv", tc.req.Transaction.ProspectID).Return(tc.trxElaborateLtv, tc.errTrxElaborateLtv)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, err := usecase.Pefindo(tc.cbFound, tc.bpkbName, tc.filtering, tc.spDupcheck)
			require.Equal(t, tc.result, data)
			require.Equal(t, tc.errResult, err)
		})
	}
}
