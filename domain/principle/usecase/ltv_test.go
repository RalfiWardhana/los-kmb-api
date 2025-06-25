package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common/platformevent"
	mockplatformevent "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrincipleElaborateLTV(t *testing.T) {
	os.Setenv("NAMA_SAMA", "K,P")
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	os.Setenv("MDM_MARKETPRICE_URL", "http://example-mdm-market-price.com")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("MARSEV_AUTHORIZATION_KEY", "marsev-auth-key")
	os.Setenv("MARSEV_LOAN_AMOUNT_URL", "http://example-marsev-loan-amount.com")
	os.Setenv("MDM_ASSET_URL", "http://example-mdm-asset-url.com")
	os.Setenv("MARSEV_FILTER_PROGRAM_URL", "http://example-marsev-filter-program.com")
	os.Setenv("MDM_MASTER_MAPPING_LICENSE_PLATE_URL", "http://example-mdm-license-plate.com")
	os.Setenv("MARSEV_CALCULATE_INSTALLMENT_URL", "http://example-marsev-calculate-installment.com")

	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	accessToken := ""

	testcases := []struct {
		name                              string
		request                           request.PrincipleElaborateLTV
		resGetPrincipleStepOne            entity.TrxPrincipleStepOne
		errGetPrincipleStepOne            error
		resGetPrincipleStepTwo            entity.TrxPrincipleStepTwo
		errGetPrincipleStepTwo            error
		resGetFilteringResult             entity.FilteringKMB
		errGetFilteringResult             error
		resGetConfig                      entity.AppConfig
		errGetConfig                      error
		resGetTrxDetailBIro               []entity.TrxDetailBiro
		errGetTrxDetailBIro               error
		resGetMappingPbkScore             entity.MappingPBKScoreGrade
		errGetMappingPbkScore             error
		resGetMappingBranchByBranchID     entity.MappingBranchByPBKScore
		errGetMappingBranchByBranchID     error
		resGetMappingElaborateLTV         []entity.MappingElaborateLTV
		errGetMappingElaborateLTV         error
		errSaveTrxElaborateLTV            error
		resCodeMDMMarketPrice             int
		resBodyMDMMarketPrice             string
		errMDMMarketPrice                 error
		resCodeMarsevLoanAmount           int
		resBodyMarsevLoanAmount           string
		errMarsevLoanAmount               error
		resCodeMDMAsset                   int
		resBodyMDMAsset                   string
		errMDMAsset                       error
		resCodeMarsevFilterProgram        int
		resBodyMarsevFilterProgram        string
		errMarsevFilterProgram            error
		resCodeMDMLicensePlate            int
		resBodyMDMLicensePlate            string
		errMDMLicensePlate                error
		resCodeMarsevCalculateInstallment int
		resBodyMarsevCalculateInstallment string
		errMarsevCalculateInstallment     error
		errSavePrincipleMarketingProgram  error
		result                            response.PrincipleElaborateLTV
		err                               error
	}{
		{
			name:                   "error get priniple step one",
			errGetPrincipleStepOne: errors.New("something wrong"),
			err:                    errors.New("something wrong"),
		},
		{
			name:                   "error get priniple step two",
			errGetPrincipleStepTwo: errors.New("something wrong"),
			err:                    errors.New("something wrong"),
		},
		{
			name:                  "error not found get filtering result",
			errGetFilteringResult: errors.New(constant.RECORD_NOT_FOUND),
			err:                   errors.New(constant.ERROR_BAD_REQUEST + " - Silahkan melakukan filtering terlebih dahulu"),
		},
		{
			name:                  "error get filtering result",
			errGetFilteringResult: errors.New("something wrong"),
			err:                   errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
		},
		{
			name: "error next process filtering result",
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess: 0,
			},
			err: errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses"),
		},
		{
			name: "error rrd date nil",
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:     1,
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_PRIME,
				RrdDate:         nil,
			},
			err: errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty"),
		},
		{
			name: "error parsing rrd date",
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:     1,
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_PRIME,
				RrdDate:         "-",
			},
			err: errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of RrdDate (-)"),
		},
		{
			name: "error different month created at & rrd date (negative result)",
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:     1,
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_PRIME,
				RrdDate:         "2025-01-31T00:00:00Z",
				CreatedAt:       time.Date(2024, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			err: errors.New(constant.ERROR_UPSTREAM + " - Difference of months RrdDate and CreatedAt is negative (-)"),
		},
		{
			name: "error get config expired contract",
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:     1,
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_PRIME,
				RrdDate:         "2025-01-31T00:00:00Z",
				CreatedAt:       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			errGetConfig: errors.New("something wrong"),
			err:          errors.New(constant.ERROR_UPSTREAM + " - Get Expired Contract Config Error"),
		},
		{
			name: "error parsing manufacture year",
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "-",
			},
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:     1,
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_PRIME,
				RrdDate:         "2025-01-31T00:00:00Z",
				CreatedAt:       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			err: errors.New(constant.ERROR_BAD_REQUEST + " - Format tahun kendaraan tidak sesuai"),
		},
		{
			name: "error parsing baki debet",
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2012",
			},
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_PRIME,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(65),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: "-",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			err: errors.New(constant.ERROR_UPSTREAM + " baki debet strconv.ParseFloat: parsing \"-\": invalid syntax"),
		},
		{
			name: "error get trx detail biro",
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2012",
			},
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_PRIME,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(65),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: float64(19999999),
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			errGetTrxDetailBIro: errors.New("something wrong"),
			err:                 errors.New(constant.ERROR_UPSTREAM + " - Get Trx Detail Biro Error"),
		},
		{
			name: "error get mapping pbk score",
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: float64(8999999),
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			errGetMappingPbkScore: errors.New("something wrong"),
			err:                   errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Pbk Score Error"),
		},
		{
			name: "error get mapping branch by branch id",
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: float64(8999999),
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			errGetMappingBranchByBranchID: errors.New("something wrong"),
			err:                           errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Branch Error"),
		},
		{
			name: "error get mapping elaborate ltv",
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: float64(8999999),
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			errGetMappingElaborateLTV: errors.New("something wrong"),
			err:                       errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error"),
		},
		{
			name: "test elaborate no hit",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      9,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: float64(8999999),
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          1,
					TenorEnd:            15,
					BPKBNameType:        0,
					LTV:                 90,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          16,
					TenorEnd:            23,
					BPKBNameType:        0,
					LTV:                 90,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          24,
					TenorEnd:            35,
					BPKBNameType:        0,
					LTV:                 85,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        0,
					LTV:                 0,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 85,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          ">12",
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate no hit 36",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: float64(8999999),
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          1,
					TenorEnd:            15,
					BPKBNameType:        0,
					LTV:                 90,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          16,
					TenorEnd:            23,
					BPKBNameType:        0,
					LTV:                 90,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          24,
					TenorEnd:            35,
					BPKBNameType:        0,
					LTV:                 85,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        0,
					LTV:                 0,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 85,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          ">12",
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk reject",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      9,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "VERY HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 85,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          16,
					TenorEnd:            23,
					LTV:                 85,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          24,
					TenorEnd:            36,
					LTV:                 0,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
					TotalBakiDebetEnd:   constant.BAKI_DEBET,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 60,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
					TotalBakiDebetEnd:   constant.BAKI_DEBET,
					TenorStart:          16,
					TenorEnd:            23,
					LTV:                 60,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
					TotalBakiDebetEnd:   constant.BAKI_DEBET,
					TenorStart:          24,
					TenorEnd:            36,
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk pass 36",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 90,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          16,
					TenorEnd:            23,
					LTV:                 90,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          24,
					TenorEnd:            35,
					LTV:                 85,
					BPKBNameType:        1,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					LTV:                 0,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 85,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          ">12",
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk pass cluster A",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      24,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 90,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          16,
					TenorEnd:            23,
					LTV:                 90,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          24,
					TenorEnd:            35,
					LTV:                 85,
					BPKBNameType:        1,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					LTV:                 0,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 85,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          ">12",
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk pass cluster A nama beda",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      24,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "O",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 90,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          16,
					TenorEnd:            23,
					BPKBNameType:        0,
					LTV:                 90,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          24,
					TenorEnd:            35,
					LTV:                 85,
					BPKBNameType:        1,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					LTV:                 0,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 85,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          ">12",
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk pass max tenor 36",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster A",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 90,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          16,
					TenorEnd:            23,
					LTV:                 90,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          24,
					TenorEnd:            35,
					LTV:                 85,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					LTV:                 0,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 85,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_PASS,
					Cluster:             "Cluster A",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   0,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          ">12",
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk reject cluster E",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      9,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetTrxDetailBIro: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-123",
					Score:      "HIGH RISK",
				},
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 50,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          16,
					TenorEnd:            23,
					LTV:                 50,
				},
				{
					ID:                  3,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          24,
					TenorEnd:            36,
					LTV:                 0,
				},
				{
					ID:                  4,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
					TotalBakiDebetEnd:   constant.BAKI_DEBET,
					TenorStart:          1,
					TenorEnd:            15,
					LTV:                 55,
				},
				{
					ID:                  5,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
					TotalBakiDebetEnd:   constant.BAKI_DEBET,
					TenorStart:          16,
					TenorEnd:            23,
					LTV:                 55,
				},
				{
					ID:                  6,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
					TotalBakiDebetEnd:   constant.BAKI_DEBET,
					TenorStart:          24,
					TenorEnd:            36,
					LTV:                 0,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk no hit tenor 36",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PBK_NO_HIT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 50,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk no hit tenor <36",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      24,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_PBK_NO_HIT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PBK_NO_HIT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          24,
					TenorEnd:            35,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 50,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk reject tenor 36",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             "Cluster E",
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 50,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk reject prime priority",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 50,
				},
			},
			errSaveTrxElaborateLTV: errors.New("something wrong"),
			err:                    errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "pass adjust tenor",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          12,
					TenorEnd:            24,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 50,
				},
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 0,
				},
			},
			result: response.PrincipleElaborateLTV{
				AdjustTenor: true,
				MaxTenor:    24,
				Reason:      "Lama Angsuran Tidak Tersedia, Silahkan coba lama angsuran yang lain",
			},
		},
		{
			name: "reject adjust tenor",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 0,
				},
			},
			result: response.PrincipleElaborateLTV{
				AdjustTenor: false,
				MaxTenor:    0,
				Reason:      "Mohon maaf, Lama Angsuran Tidak Tersedia",
			},
		},
		{
			name: "error get mdm market price",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			errMDMMarketPrice: errors.New("something new"),
			err:               errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Timeout"),
		},
		{
			name: "error code get mdm market price",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 500,
			err:                   errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error"),
		},
		{
			name: "error empty response get mdm market price",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": []
				},
				"errors": null
			}`,
			err: errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error"),
		},
		{
			name: "error marsev get loan amount",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			errMarsevLoanAmount: errors.New("something wrong"),
			err:                 errors.New("something wrong"),
		},
		{
			name: "error code marsev get loan amount",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 500,
			err:                     errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Loan Amount Error"),
		},
		{
			name: "error unmarshal marsev get loan amount",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `-`,
			err:                     errors.New("invalid character ' ' in numeric literal"),
		},
		{
			name: "error mdm asset",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
				LoanAmount: 1000000,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			errMDMAsset: errors.New("something wrong"),
			err:         errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout"),
		},
		{
			name: "error code mdm asset",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
				LoanAmount: 1000000,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 500,
			err:             errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
		{
			name: "error empty data mdm asset",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
				LoanAmount: 1000000,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": []
				},
				"errors": null
			}`,
			err: errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
		{
			name: "error marsev filter program",
			request: request.PrincipleElaborateLTV{
				ProspectID: "SAL-123",
				Tenor:      36,
				LoanAmount: 1000000,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			errMarsevFilterProgram: errors.New("something wrong"),
			err:                    errors.New("something wrong"),
		},
		{
			name: "error code marsev filter program",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 500,
			err:                        errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Filter Program Error"),
		},
		{
			name: "error unmarshal marsev filter program",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `-`,
			err:                        errors.New("invalid character ' ' in numeric literal"),
		},
		{
			name: "error not found marsev filter program",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			err: errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Filter Program Error Not Found Data"),
		},
		{
			name: "error mdm license plate",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			errMDMLicensePlate: errors.New("something wrong"),
			err:                errors.New("something wrong"),
		},
		{
			name: "error code mdm license plate",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 500,
			err:                    errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping License Plate Error"),
		},
		{
			name: "error unmarshal mdm license plate",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 200,
			resBodyMDMLicensePlate: `-`,
			err:                    errors.New("invalid character ' ' in numeric literal"),
		},
		{
			name: "error not found mdm license plate",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 200,
			resBodyMDMLicensePlate: `{
				"code": "200",
				"message": "Success",
				"data": {
					"records": [],
					"max_page": 0,
					"total": 0,
					"page_size": 10,
					"current_page": 1
				},
				"errors": null,
				"request_id": "req-123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			err: errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping License Plate Error Not Found Data"),
		},
		{
			name: "error marsev calculate installment",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 200,
			resBodyMDMLicensePlate: `{
				"code": "200",
				"message": "Success",
				"data": {
					"records": [{
						"plate_area_id": 1,
						"plate_id": 101,
						"plate_code": "B",
						"area_id": "426",
						"area_description": "JAKARTA PUSAT",
						"lob_id": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"created_by": "system",
						"updated_at": "2024-01-02T00:00:00Z",
						"updated_by": "admin",
						"deleted_at": null,
						"deleted_by": null
					}],
					"max_page": 1,
					"total": 1,
					"page_size": 10,
					"current_page": 1
				},
				"errors": null,
				"request_id": "req-123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			errMarsevCalculateInstallment: errors.New("something wrong"),
			err:                           errors.New("something wrong"),
		},
		{
			name: "error code marsev calculate installment",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 200,
			resBodyMDMLicensePlate: `{
				"code": "200",
				"message": "Success",
				"data": {
					"records": [{
						"plate_area_id": 1,
						"plate_id": 101,
						"plate_code": "B",
						"area_id": "426",
						"area_description": "JAKARTA PUSAT",
						"lob_id": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"created_by": "system",
						"updated_at": "2024-01-02T00:00:00Z",
						"updated_by": "admin",
						"deleted_at": null,
						"deleted_by": null
					}],
					"max_page": 1,
					"total": 1,
					"page_size": 10,
					"current_page": 1
				},
				"errors": null,
				"request_id": "req-123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			resCodeMarsevCalculateInstallment: 500,
			err:                               errors.New(constant.ERROR_UPSTREAM + " - Marsev Calculate Installment Error"),
		},
		{
			name: "error unmarshal marsev calculate installment",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 200,
			resBodyMDMLicensePlate: `{
				"code": "200",
				"message": "Success",
				"data": {
					"records": [{
						"plate_area_id": 1,
						"plate_id": 101,
						"plate_code": "B",
						"area_id": "426",
						"area_description": "JAKARTA PUSAT",
						"lob_id": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"created_by": "system",
						"updated_at": "2024-01-02T00:00:00Z",
						"updated_by": "admin",
						"deleted_at": null,
						"deleted_by": null
					}],
					"max_page": 1,
					"total": 1,
					"page_size": 10,
					"current_page": 1
				},
				"errors": null,
				"request_id": "req-123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			resCodeMarsevCalculateInstallment: 200,
			resBodyMarsevCalculateInstallment: `-`,
			err:                               errors.New("invalid character ' ' in numeric literal"),
		},
		{
			name: "error save principle marketing program",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 200,
			resBodyMDMLicensePlate: `{
				"code": "200",
				"message": "Success",
				"data": {
					"records": [{
						"plate_area_id": 1,
						"plate_id": 101,
						"plate_code": "B",
						"area_id": "426",
						"area_description": "JAKARTA PUSAT",
						"lob_id": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"created_by": "system",
						"updated_at": "2024-01-02T00:00:00Z",
						"updated_by": "admin",
						"deleted_at": null,
						"deleted_by": null
					}],
					"max_page": 1,
					"total": 1,
					"page_size": 10,
					"current_page": 1
				},
				"errors": null,
				"request_id": "req-123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			resCodeMarsevCalculateInstallment: 200,
			resBodyMarsevCalculateInstallment: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"installment_type_code": "ADDB",
					"is_psa": false,
					"tenor": 36,
					"admin_fee": 2500000,
					"admin_fee_psa": 0,
					"provision_fee": 1200000,
					"amount_of_finance": 100000000,
					"dp_amount": 20000000,
					"dp_percent": 20,
					"additional_rate": 0.5,
					"effective_rate": 12.5,
					"life_insurance": 1500000,
					"asset_insurance": 2000000,
					"total_insurance": 3500000,
					"fiducia_fee": 500000,
					"ntf": 107700000,
					"monthly_installment": 3500000,
					"monthly_installment_min": 3400000,
					"monthly_installment_max": 3600000,
					"total_loan": 126000000,
					"amount_of_interest": 26000000,
					"flat_rate_yearly_percent": 6.5,
					"flat_rate_monthly_percent": 0.54,
					"product_id": "PRD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Car Loan",
					"subsidy_amount_scheme": 0,
					"fine_amount": 175000,
					"fine_amount_formula": "5% * installment",
					"fine_amount_detail": "5% penalty from monthly installment",
					"ntf_formula": "loan_amount + total_insurance + admin_fee",
					"ntf_detail": "Total loan calculation details",
					"amount_of_interest_formula": "flat_rate * tenor * loan_amount",
					"amount_of_interest_detail": "Interest calculation details",
					"wanprestasi_freight_fee": 1000000,
					"external_freight_fee": 500000,
					"wanprestasi_freight_formula": "fixed_amount",
					"wanprestasi_freight_detail": "Fixed penalty amount",
					"external_freight_formula": "fixed_amount",
					"external_freight_detail": "Fixed external fee",
					"is_stamp_duty_as_loan": false,
					"stamp_duty_fee": 10000
				}],
				"errors": null
			}`,
			errSavePrincipleMarketingProgram: errors.New("something wrong"),
			err:                              errors.New("something wrong"),
		},
		{
			name: "success",
			request: request.PrincipleElaborateLTV{
				ProspectID:     "SAL-123",
				Tenor:          36,
				LoanAmount:     1000000,
				FinancePurpose: constant.FINANCE_PURPOSE_MODAL_KERJA,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				ManufactureYear: "2018",
				LicensePlate:    "B1234CD",
			},
			resGetFilteringResult: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				NextProcess:                     1,
				CustomerStatus:                  constant.STATUS_KONSUMEN_RO,
				CustomerSegment:                 constant.RO_AO_REGULAR,
				ScoreBiro:                       "HIGH RISK",
				MaxOverdueBiro:                  int64(29),
				BpkbName:                        "K",
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				CMOCluster:                      "Cluster E",
				RrdDate:                         "2025-01-31T00:00:00Z",
				CreatedAt:                       time.Date(2025, 12, 30, 0, 0, 0, 0, time.UTC),
			},
			resGetConfig: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			resGetMappingPbkScore: entity.MappingPBKScoreGrade{
				GradeScore: "",
			},
			resGetMappingBranchByBranchID: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					Cluster:             constant.CLUSTER_PRIME_PRIORITY,
					TotalBakiDebetStart: 0,
					TotalBakiDebetEnd:   constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
					TenorStart:          36,
					TenorEnd:            36,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			resCodeMDMMarketPrice: 200,
			resBodyMDMMarketPrice: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2023,
							"market_price_value": 25000000
						},
						{
							"asset_code": "MOT001",
							"branch_id": "426",
							"brand": "HONDA",
							"manufacturing_year": 2022,
							"market_price_value": 22000000
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevLoanAmount: 200,
			resBodyMarsevLoanAmount: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resCodeMDMAsset: 200,
			resBodyMDMAsset: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resCodeMarsevFilterProgram: 200,
			resBodyMarsevFilterProgram: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			resCodeMDMLicensePlate: 200,
			resBodyMDMLicensePlate: `{
				"code": "200",
				"message": "Success",
				"data": {
					"records": [{
						"plate_area_id": 1,
						"plate_id": 101,
						"plate_code": "B",
						"area_id": "426",
						"area_description": "JAKARTA PUSAT",
						"lob_id": 1,
						"created_at": "2024-01-01T00:00:00Z",
						"created_by": "system",
						"updated_at": "2024-01-02T00:00:00Z",
						"updated_by": "admin",
						"deleted_at": null,
						"deleted_by": null
					}],
					"max_page": 1,
					"total": 1,
					"page_size": 10,
					"current_page": 1
				},
				"errors": null,
				"request_id": "req-123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			resCodeMarsevCalculateInstallment: 200,
			resBodyMarsevCalculateInstallment: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"installment_type_code": "ADDB",
					"is_psa": false,
					"tenor": 36,
					"admin_fee": 2500000,
					"admin_fee_psa": 0,
					"provision_fee": 1200000,
					"amount_of_finance": 100000000,
					"dp_amount": 20000000,
					"dp_percent": 20,
					"additional_rate": 0.5,
					"effective_rate": 12.5,
					"life_insurance": 1500000,
					"asset_insurance": 2000000,
					"total_insurance": 3500000,
					"fiducia_fee": 500000,
					"ntf": 107700000,
					"monthly_installment": 3500000,
					"monthly_installment_min": 3400000,
					"monthly_installment_max": 3600000,
					"total_loan": 126000000,
					"amount_of_interest": 26000000,
					"flat_rate_yearly_percent": 6.5,
					"flat_rate_monthly_percent": 0.54,
					"product_id": "PRD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Car Loan",
					"subsidy_amount_scheme": 0,
					"fine_amount": 175000,
					"fine_amount_formula": "5% * installment",
					"fine_amount_detail": "5% penalty from monthly installment",
					"ntf_formula": "loan_amount + total_insurance + admin_fee",
					"ntf_detail": "Total loan calculation details",
					"amount_of_interest_formula": "flat_rate * tenor * loan_amount",
					"amount_of_interest_detail": "Interest calculation details",
					"wanprestasi_freight_fee": 1000000,
					"external_freight_fee": 500000,
					"wanprestasi_freight_formula": "fixed_amount",
					"wanprestasi_freight_detail": "Fixed penalty amount",
					"external_freight_formula": "fixed_amount",
					"external_freight_detail": "Fixed external fee",
					"is_stamp_duty_as_loan": false,
					"stamp_duty_fee": 10000
				}],
				"errors": null
			}`,
			result: response.PrincipleElaborateLTV{
				LTV:               60,
				AdjustTenor:       true,
				MaxTenor:          36,
				LoanAmountMaximum: float64(80000000),
				IsPsa:             true,
				Dealer:            "PSA",
				InstallmentAmount: float64(3500000),
				AF:                float64(100000000),
				AdminFee:          float64(2500000),
				NTF:               float64(107700000),
				AssetCategoryID:   "SPM",
				Otr:               float64(25000000),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			mockRepository.On("GetPrincipleStepOne", mock.Anything).Return(tc.resGetPrincipleStepOne, tc.errGetPrincipleStepOne)
			mockRepository.On("GetPrincipleStepTwo", mock.Anything).Return(tc.resGetPrincipleStepTwo, tc.errGetPrincipleStepTwo)
			mockRepository.On("GetFilteringResult", mock.Anything).Return(tc.resGetFilteringResult, tc.errGetFilteringResult)
			mockRepository.On("GetConfig", mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetConfig, tc.errGetConfig)
			mockRepository.On("GetTrxDetailBIro", mock.Anything).Return(tc.resGetTrxDetailBIro, tc.errGetTrxDetailBIro)
			mockRepository.On("GetMappingPbkScore", mock.Anything).Return(tc.resGetMappingPbkScore, tc.errGetMappingPbkScore)
			mockRepository.On("GetMappingBranchByBranchID", mock.Anything, mock.Anything).Return(tc.resGetMappingBranchByBranchID, tc.errGetMappingBranchByBranchID)
			mockRepository.On("GetMappingElaborateLTV", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetMappingElaborateLTV, tc.errGetMappingElaborateLTV)
			mockRepository.On("SaveTrxElaborateLTV", mock.Anything).Return(tc.errSaveTrxElaborateLTV)
			mockRepository.On("SavePrincipleMarketingProgram", mock.Anything).Return(tc.errSavePrincipleMarketingProgram)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("MDM_MARKETPRICE_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resCodeMDMMarketPrice, tc.resBodyMDMMarketPrice))
			resp, _ := rst.R().Post(url)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 10, tc.request.ProspectID, accessToken).Return(resp, tc.errMDMMarketPrice).Once()

			url2 := os.Getenv("MARSEV_LOAN_AMOUNT_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url2, httpmock.NewStringResponder(tc.resCodeMarsevLoanAmount, tc.resBodyMarsevLoanAmount))
			resp2, _ := rst.R().Post(url2)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url2, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.request.ProspectID, accessToken).Return(resp2, tc.errMarsevLoanAmount).Once()

			url3 := os.Getenv("MDM_ASSET_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url3, httpmock.NewStringResponder(tc.resCodeMDMAsset, tc.resBodyMDMAsset))
			resp3, _ := rst.R().Post(url3)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url3, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 10, tc.request.ProspectID, accessToken).Return(resp3, tc.errMDMAsset).Once()

			url4 := os.Getenv("MARSEV_FILTER_PROGRAM_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url4, httpmock.NewStringResponder(tc.resCodeMarsevFilterProgram, tc.resBodyMarsevFilterProgram))
			resp4, _ := rst.R().Post(url4)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url4, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 10, tc.request.ProspectID, accessToken).Return(resp4, tc.errMarsevFilterProgram).Once()

			licensePlateCode := utils.GetLicensePlateCode(tc.resGetPrincipleStepOne.LicensePlate)
			url5 := os.Getenv("MDM_MASTER_MAPPING_LICENSE_PLATE_URL") + "?lob_id=" + strconv.Itoa(constant.LOBID_KMB) + "&plate_code=" + licensePlateCode
			httpmock.RegisterResponder(constant.METHOD_GET, url5, httpmock.NewStringResponder(tc.resCodeMDMLicensePlate, tc.resBodyMDMLicensePlate))
			resp5, _ := rst.R().Get(url5)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url5, mock.Anything, mock.Anything, constant.METHOD_GET, false, 0, 10, tc.request.ProspectID, accessToken).Return(resp5, tc.errMDMLicensePlate).Once()

			url6 := os.Getenv("MARSEV_CALCULATE_INSTALLMENT_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url6, httpmock.NewStringResponder(tc.resCodeMarsevCalculateInstallment, tc.resBodyMarsevCalculateInstallment))
			resp6, _ := rst.R().Post(url6)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url6, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 10, tc.request.ProspectID, accessToken).Return(resp6, tc.errMarsevCalculateInstallment).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, platformEvent)

			result, err := usecase.PrincipleElaborateLTV(ctx, tc.request, accessToken)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.result, result)
			}
		})
	}
}
