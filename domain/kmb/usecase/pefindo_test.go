package usecase

import (
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPefindo(t *testing.T) {
	os.Setenv("NAMA_SAMA", "K,P")

	// Get the current time
	currentTime := time.Now().UTC()

	// Sample older date from the current time to test "RrdDate"
	sevenMonthsAgo := currentTime.AddDate(0, -7, 0)
	sixMonthsAgo := currentTime.AddDate(0, -6, 0)

	testcases := []struct {
		name         string
		cbFound      bool
		bpkbName     string
		filtering    entity.FilteringKMB
		spDupcheck   response.SpDupcheckMap
		req          request.Metrics
		result       response.UsecaseApi
		errResult    error
		config       entity.AppConfig
		errGetConfig error
	}{
		{
			name: "Pefindo prime 1",
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
			name: "Pefindo prime 2",
			filtering: entity.FilteringKMB{
				CustomerSegment: constant.RO_AO_PRIORITY,
				RrdDate:         sixMonthsAgo,
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
			name: "Pefindo - CR Perbaikan Flow RO PrimePriority PASS",
			filtering: entity.FilteringKMB{
				CustomerSegment: constant.RO_AO_PRIME,
				RrdDate:         sevenMonthsAgo,
				CreatedAt:       currentTime,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
			},
			config: entity.AppConfig{
				Key:   "expired_contract_check",
				Value: `{"data":{"expired_contract_check_enabled":true,"expired_contract_max_months":6}}`,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_PRIME_PRIORITY_EXP_CONTRACT_6MONTHS,
				Reason:         constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + fmt.Sprintf("%s %s - PBK Pass", constant.STATUS_KONSUMEN_RO, constant.RO_AO_PRIME),
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name: "Pefindo - CR Perbaikan Flow RO PrimePriority RrdDate NULL",
			filtering: entity.FilteringKMB{
				CustomerSegment: constant.RO_AO_PRIME,
				RrdDate:         nil,
				CreatedAt:       currentTime,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
			},
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty"),
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
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
				Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_SAMA, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
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
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
				Reason:         constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI,
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
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_WO_AGUNAN_REJECT_CODE,
				Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_SAMA, constant.ADA_FASILITAS_WO_AGUNAN),
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
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_GT20J,
				Reason:         constant.NAMA_SAMA_BAKI_DEBET_TIDAK_SESUAI,
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
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_BPKB_SAMA_BAKI_DEBET_LTE20J,
				Reason:         constant.NAMA_SAMA_BAKI_DEBET_SESUAI,
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
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_PEFINDO_BPKB_BEDA,
				Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_BEDA, fmt.Sprintf(constant.REASON_PEFINDO_CURRENT_GT30, constant.PBK_OVD_CURRENT)),
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BIRO,
			},
		},
		{
			name:     "Pefindo PASS",
			cbFound:  true,
			bpkbName: "KK",
			filtering: entity.FilteringKMB{
				CustomerSegment:               constant.RO_AO_REGULAR,
				MaxOverdueBiro:                9,
				MaxOverdueLast12monthsBiro:    9,
				MaxOverdueKORules:             9,
				MaxOverdueLast12MonthsKORules: 9,
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
			result: response.UsecaseApi{
				Code:           constant.NAMA_SAMA_NO_FACILITY_WO_CODE,
				Reason:         fmt.Sprintf("%s & %s", constant.REASON_BPKB_BEDA, constant.TIDAK_ADA_FASILITAS_WO_AGUNAN),
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
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			// mockRepository.On("GetElaborateLtv", tc.req.Transaction.ProspectID).Return(tc.trxElaborateLtv, tc.errTrxElaborateLtv)
			mockRepository.On("GetConfig", "expired_contract", "KMB-OFF", "expired_contract_check").Return(tc.config, tc.errGetConfig)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, err := usecase.Pefindo(tc.cbFound, tc.bpkbName, tc.filtering, tc.spDupcheck)
			require.Equal(t, tc.result, data)
			require.Equal(t, tc.errResult, err)
		})
	}
}
