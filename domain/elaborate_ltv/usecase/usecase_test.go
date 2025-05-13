package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/elaborate_ltv/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestElaborate(t *testing.T) {
	os.Setenv("NAMA_SAMA", "K,P")
	accessToken := "token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
	testcases := []struct {
		name                     string
		reqs                     request.ElaborateLTV
		filteringKMB             entity.FilteringKMB
		errGetGetFilteringResult error
		mappingElaborateLTV      []entity.MappingElaborateLTV
		errMapping               error
		errSaveTrxElaborateLTV   error
		result                   response.ElaborateLTV
		errFinal                 error
		filteringDetail          []entity.TrxDetailBiro
		errGetGetFilteringDetail error
		mappingPBKScoreGrade     []entity.MappingPBKScoreGrade
		errmappingPBKScoreGrade  error
		mappingBranch            entity.MappingBranchByPBKScore
		errmappingBranch         error
	}{
		{
			name: "test elaborate err1",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: "2000",
			},
			errGetGetFilteringResult: errors.New(constant.RECORD_NOT_FOUND),
			errFinal:                 errors.New(constant.ERROR_BAD_REQUEST + " - Silahkan melakukan filtering terlebih dahulu"),
		},
		{
			name: "test elaborate err2",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: "2000",
			},
			errGetGetFilteringResult: errors.New("error"),
			errFinal:                 errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
		},
		{
			name: "test elaborate err3",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: "2000",
			},
			filteringKMB: entity.FilteringKMB{
				Decision: constant.DECISION_REJECT,
			},
			errFinal: errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses"),
		},
		{
			name: "test elaborate err4",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: "2000",
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				CMOCluster:                      "Cluster A",
				TotalBakiDebetNonCollateralBiro: 30000,
				NextProcess:                     1,
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			errMapping: errors.New("error"),
			errFinal:   errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error"),
		},
		{
			name: "test elaborate no hit",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: "2000",
			},
			filteringKMB: entity.FilteringKMB{
				Decision:        constant.DECISION_PASS,
				CMOCluster:      "Cluster A",
				NextProcess:     1,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         90,
				AdjustTenor: true,
				MaxTenor:    35,
			},
		},
		{
			name: "test elaborate no hit 36",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             36,
				ManufacturingYear: "2019",
			},
			filteringKMB: entity.FilteringKMB{
				Decision:        constant.DECISION_PASS,
				CMOCluster:      "Cluster A",
				NextProcess:     1,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         85,
				AdjustTenor: true,
				MaxTenor:    36,
			},
		},
		{
			name: "test elaborate pbk reject",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: "2009",
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				CMOCluster:                      "Cluster A",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				ScoreBiro:                       "VERY HIGH RISK",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			filteringDetail: []entity.TrxDetailBiro{
				{
					ProspectID: "SAL-12345",
					Score:      "HIGH RISK",
				},
				{
					ProspectID: "SAL-12345",
					Score:      "VERY HIGH RISK",
				},
			},
			mappingPBKScoreGrade: []entity.MappingPBKScoreGrade{
				{
					Score:      "VERY HIGH RISK",
					GradeRisk:  6,
					GradeScore: "BAD",
				},
				{
					Score:      "UNSCORE",
					GradeRisk:  5,
					GradeScore: "BAD",
				},
				{
					Score:      "HIGH RISK",
					GradeRisk:  4,
					GradeScore: "BAD",
				},
				{
					Score:      "AVERAGE RISK",
					GradeRisk:  3,
					GradeScore: "BAD",
				},
				{
					Score:      "LOW RISK",
					GradeRisk:  2,
					GradeScore: "BAD",
				},
				{
					Score:      "VERY LOW RISK",
					GradeRisk:  1,
					GradeScore: "BAD",
				},
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         85,
				AdjustTenor: true,
				MaxTenor:    23,
			},
		},
		{
			name: "test elaborate pbk pass 36",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             36,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				CMOCluster:                      "Cluster A",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         85,
				AdjustTenor: true,
				MaxTenor:    36,
			},
		},
		{
			name: "test elaborate pbk pass cluster A",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             24,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				CMOCluster:                      "Cluster A",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         85,
				AdjustTenor: true,
				MaxTenor:    36,
			},
		},
		{
			name: "test elaborate pbk pass max tenor 36",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_PASS,
				CMOCluster:                      "Cluster A",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         90,
				AdjustTenor: true,
				MaxTenor:    36,
			},
			errSaveTrxElaborateLTV: errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
			errFinal:               errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error"),
		},
		{
			name: "test elaborate pbk reject cluster E",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             9,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				CMOCluster:                      "Cluster E",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         55,
				AdjustTenor: true,
				MaxTenor:    23,
			},
		},
		{
			name: "test elaborate pbk no hit tenor 36",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             36,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_PBK_NO_HIT,
				CMOCluster:                      "Cluster E",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         50,
				AdjustTenor: true,
				MaxTenor:    36,
			},
		},
		{
			name: "test elaborate pbk no hit tenor <36",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             24,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_PBK_NO_HIT,
				CMOCluster:                      "Cluster E",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT + 1,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         50,
				AdjustTenor: true,
				MaxTenor:    35,
			},
		},
		{
			name: "test elaborate pbk reject tenor 36",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             36,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				CMOCluster:                      "Cluster E",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_REGULAR,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         50,
				AdjustTenor: true,
				MaxTenor:    36,
			},
		},
		{
			name: "test elaborate pbk reject prime priority",
			reqs: request.ElaborateLTV{
				ProspectID:        "EFM0TST0020230809011",
				Tenor:             36,
				ManufacturingYear: time.Now().AddDate(-9, 0, 0).Format("2006"),
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        constant.DECISION_REJECT,
				CMOCluster:                      "Cluster E",
				NextProcess:                     1,
				TotalBakiDebetNonCollateralBiro: constant.RANGE_CLUSTER_BAKI_DEBET_REJECT,
				ScoreBiro:                       "HIGH RISK",
				BpkbName:                        "K",
				CustomerSegment:                 constant.RO_AO_PRIORITY,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
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
			result: response.ElaborateLTV{
				LTV:         50,
				AdjustTenor: true,
				MaxTenor:    36,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			tc.filteringKMB.CustomerStatus = "NEW"
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetFilteringResult", tc.reqs.ProspectID).Return(tc.filteringKMB, tc.errGetGetFilteringResult)
			mockRepository.On("GetFilteringDetail", tc.reqs.ProspectID).Return(tc.filteringDetail, tc.errGetGetFilteringDetail)
			mockRepository.On("GetMappingPBKScoreGrade").Return(tc.mappingPBKScoreGrade, tc.errmappingPBKScoreGrade)
			mockRepository.On("GetMappingBranchPBK", mock.Anything, mock.Anything).Return(tc.mappingBranch, tc.errmappingBranch)
			mockRepository.On("GetMappingElaborateLTV", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.mappingElaborateLTV, tc.errMapping)
			mockRepository.On("SaveTrxElaborateLTV", mock.Anything).Return(tc.errSaveTrxElaborateLTV)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.Elaborate(ctx, tc.reqs, accessToken)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errFinal, err)
		})
	}

}
