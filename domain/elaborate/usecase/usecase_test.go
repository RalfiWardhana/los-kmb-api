package usecase

import (
	"context"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"testing"

	"los-kmb-api/domain/elaborate/interfaces/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestElaborate(t *testing.T) {
	accessToken := "token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testCases := []struct {
		name                   string
		req                    request.BodyRequestElaborate
		errSaveDataElaborate   error
		resResultElaborate     response.ElaborateResult
		errResultElaborate     error
		errUpdateDataElaborate error
		resFinal               response.ElaborateResult
		errFinal               error
	}{
		{
			name: "TEST_ERROR_Elaborate_SaveDataElaborate",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_NEW,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "202",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			errSaveDataElaborate: errors.New("failed process elaborate"),
			errFinal:             errors.New("failed process elaborate"),
		},
		{
			name: "TEST_PASS_Elaborate_NotFoundMapping",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  constant.RO_AO_PRIME,
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "202",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resResultElaborate: response.ElaborateResult{
				Decision: "",
			},
			resFinal: response.ElaborateResult{
				Code:     constant.CODE_PASS_ELABORATE,
				Decision: constant.DECISION_PASS,
				Reason:   constant.REASON_PASS_ELABORATE,
			},
		},
		{
			name: "TEST_PASS_Elaborate_FoundMapping",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  constant.RO_AO_PRIME,
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "202",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resResultElaborate: response.ElaborateResult{
				Code:     constant.CODE_PASS_ELABORATE,
				Decision: constant.DECISION_PASS,
				Reason:   constant.REASON_PASS_ELABORATE,
				LTV:      50,
			},
			resFinal: response.ElaborateResult{
				Code:     constant.CODE_PASS_ELABORATE,
				Decision: constant.DECISION_PASS,
				Reason:   constant.REASON_PASS_ELABORATE,
			},
		},
		{
			name: "TEST_REJECT_Elaborate_FoundMapping",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  constant.RO_AO_PRIME,
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "202",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resResultElaborate: response.ElaborateResult{
				Code:     constant.CODE_REJECT_NTF_ELABORATE,
				Decision: constant.DECISION_REJECT,
				Reason:   constant.REASON_REJECT_NTF_ELABORATE,
				LTV:      50,
			},
			resFinal: response.ElaborateResult{
				Code:     constant.CODE_REJECT_NTF_ELABORATE,
				Decision: constant.DECISION_REJECT,
				Reason:   constant.REASON_REJECT_NTF_ELABORATE,
				LTV:      50,
			},
		},
		{
			name: "TEST_ERROR_Elaborate_ResultElaborate",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  constant.RO_AO_PRIME,
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "202",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resResultElaborate: response.ElaborateResult{
				Code:     constant.CODE_REJECT_NTF_ELABORATE,
				Decision: constant.DECISION_REJECT,
				Reason:   constant.REASON_REJECT_NTF_ELABORATE,
				LTV:      50,
			},
			errResultElaborate: errors.New("failed get result elaborate"),
			errFinal:           errors.New("failed get result elaborate"),
		},
		{
			name: "TEST_ERROR_Elaborate_UpdateData",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  constant.RO_AO_PRIME,
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "202",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resResultElaborate: response.ElaborateResult{
				Code:         constant.CODE_REJECT_NTF_ELABORATE,
				Decision:     constant.DECISION_REJECT,
				Reason:       constant.REASON_REJECT_NTF_ELABORATE,
				LTV:          50,
				IsMappingOvd: true,
			},
			errUpdateDataElaborate: errors.New("failed update data api elaborate"),
			resFinal: response.ElaborateResult{
				Code:     constant.CODE_REJECT_NTF_ELABORATE,
				Decision: constant.DECISION_REJECT,
				Reason:   constant.REASON_REJECT_NTF_ELABORATE,
				LTV:      50,
			},
			errFinal: errors.New("failed update data api elaborate"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockUsecase := new(mocks.Usecase)

			mockRepository.On("SaveDataElaborate", mock.Anything).Return(tc.errSaveDataElaborate).Once()
			mockRepository.On("UpdateDataElaborate", mock.Anything).Return(tc.errUpdateDataElaborate).Once()

			mockUsecase.On("ResultElaborate", mock.Anything, mock.Anything).Return(tc.resResultElaborate, tc.errResultElaborate).Once()

			multiUsecase, _ := NewMultiUsecase(mockRepository, mockHttpClient, mockUsecase)

			result, err := multiUsecase.Elaborate(ctx, tc.req, accessToken)

			require.Equal(t, tc.resFinal, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestResultElaborate(t *testing.T) {
	os.Setenv("NAMA_SAMA", "K,P")
	os.Setenv("NAMA_BEDA", "O,KK")

	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	pefindoUnscoreValue := constant.UNSCORE_PBK
	pefindoNotMatchValue := constant.NOT_MATCH_PBK

	testCases := []struct {
		name                         string
		req                          request.BodyRequestElaborate
		resGetClusterBranchElaborate entity.ClusterBranch
		errGetClusterBranchElaborate error
		resGetFilteringResult        entity.ApiDupcheckKmbUpdate
		errGetFilteringResult        error
		resGetResultElaborate        entity.ResultElaborate
		errGetResultElaborate        error
		resGetMappingLtvOvd          entity.ResultElaborate
		errGetMappingLtvOvd          error
		resFinal                     response.ElaborateResult
		errFinal                     error
	}{
		{
			name: "TEST_ERROR_ResultElaborate_TimeParse",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_NEW,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "202",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			errFinal: errors.New("error parsing manufacturing year"),
		},
		{
			name: "TEST_ERROR_ResultElaborate_GetClusterBranchElaborate",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_NEW,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    20000000,
					Tenor:             12,
					ManufacturingYear: "2023",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			errGetClusterBranchElaborate: errors.New("failed get cluster branch elaborate"),
			resFinal: response.ElaborateResult{
				ResultPefindo: constant.DECISION_PASS,
				BPKBNameType:  1,
				AgeVehicle:    "<=12",
				LTVOrigin:     67,
			},
			errFinal: errors.New("failed get cluster branch elaborate"),
		},
		{
			name: "TEST_REJECT_ResultElaborate_BakiDebet10jt_ClusterE",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_NEW,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    10000000.00,
					Tenor:             12,
					ManufacturingYear: "2023",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resGetClusterBranchElaborate: entity.ClusterBranch{
				Cluster: "Cluster E",
			},
			resGetResultElaborate: entity.ResultElaborate{
				Decision: constant.DECISION_REJECT,
			},
			resFinal: response.ElaborateResult{
				Code:           constant.CODE_REJECT_ELABORATE,
				Decision:       constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_ELABORATE,
				ResultPefindo:  constant.DECISION_PASS,
				Cluster:        "Cluster E",
				BPKBNameType:   1,
				AgeVehicle:     "<=12",
				LTVOrigin:      67,
				TotalBakiDebet: 10000000.00,
			},
		},
		{
			name: "TEST_REJECT_ResultElaborate_BakiDebet>20jt",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_NEW,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    21000000.00,
					Tenor:             12,
					ManufacturingYear: "2023",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resGetClusterBranchElaborate: entity.ClusterBranch{
				Cluster: "Cluster E",
			},
			resGetResultElaborate: entity.ResultElaborate{
				Decision: constant.DECISION_REJECT,
			},
			resFinal: response.ElaborateResult{
				Code:           constant.CODE_REJECT_ELABORATE,
				Decision:       constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_ELABORATE,
				ResultPefindo:  constant.DECISION_PASS,
				Cluster:        "Cluster E",
				BPKBNameType:   1,
				AgeVehicle:     "<=12",
				LTVOrigin:      67,
				TotalBakiDebet: 21000000.00,
			},
		},
		{
			name: "TEST_REJECT_ResultElaborate_NTFElaborate_LTV>0",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "K",
					CustomerStatus:    constant.STATUS_KONSUMEN_NEW,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    11000000.00,
					Tenor:             12,
					ManufacturingYear: "2023",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resGetClusterBranchElaborate: entity.ClusterBranch{
				Cluster: "Cluster A",
			},
			resGetFilteringResult: entity.ApiDupcheckKmbUpdate{
				RequestID:    "TEST123",
				PefindoScore: &pefindoUnscoreValue,
			},
			resGetResultElaborate: entity.ResultElaborate{
				Decision: constant.DECISION_REJECT,
				LTV:      50,
			},
			resFinal: response.ElaborateResult{
				Code:           constant.CODE_REJECT_NTF_ELABORATE,
				Decision:       constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_NTF_ELABORATE,
				ResultPefindo:  constant.DECISION_PBK_NO_HIT,
				Cluster:        "Cluster A",
				BPKBNameType:   1,
				AgeVehicle:     "<=12",
				LTVOrigin:      67,
				LTV:            49,
				TotalBakiDebet: 11000000.00,
			},
		},
		{
			name: "TEST_REJECT_ResultElaborate_NTFElaborate_LTV=0",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "O",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    11000000.00,
					Tenor:             12,
					ManufacturingYear: "2010",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resGetClusterBranchElaborate: entity.ClusterBranch{
				Cluster: "Cluster A",
			},
			resGetFilteringResult: entity.ApiDupcheckKmbUpdate{
				RequestID:    "TEST123",
				PefindoID:    "1676593952",
				PefindoScore: &pefindoUnscoreValue,
			},
			resGetResultElaborate: entity.ResultElaborate{
				Decision: constant.DECISION_REJECT,
				LTV:      0,
			},
			resFinal: response.ElaborateResult{
				Code:           constant.CODE_REJECT_ELABORATE,
				Decision:       constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_ELABORATE,
				ResultPefindo:  constant.DECISION_PBK_NO_HIT,
				Cluster:        "Cluster A",
				BPKBNameType:   0,
				AgeVehicle:     ">12",
				LTVOrigin:      67,
				LTV:            0,
				TotalBakiDebet: 11000000.00,
			},
		},
		{
			name: "TEST_ERROR_GetResultElaborate",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "O",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    11000000.00,
					Tenor:             12,
					ManufacturingYear: "2010",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resGetClusterBranchElaborate: entity.ClusterBranch{
				Cluster: "Cluster A",
			},
			resGetFilteringResult: entity.ApiDupcheckKmbUpdate{},
			resGetResultElaborate: entity.ResultElaborate{
				Decision: constant.DECISION_REJECT,
				LTV:      0,
			},
			errGetResultElaborate: errors.New("failed get mapping elaborate ltv"),
			resFinal: response.ElaborateResult{
				ResultPefindo:  constant.DECISION_PASS,
				Cluster:        "Cluster A",
				BPKBNameType:   0,
				AgeVehicle:     ">12",
				LTVOrigin:      67,
				LTV:            0,
				TotalBakiDebet: 11000000.00,
			},
			errFinal: errors.New("failed get mapping elaborate ltv"),
		},
		{
			name: "TEST_ERROR_GetFilteringResult",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "O",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    11000000.00,
					Tenor:             12,
					ManufacturingYear: "2010",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resGetClusterBranchElaborate: entity.ClusterBranch{
				Cluster: "Cluster A",
			},
			resGetFilteringResult: entity.ApiDupcheckKmbUpdate{
				RequestID:    "TEST123",
				PefindoID:    "1676593952",
				PefindoScore: &pefindoUnscoreValue,
			},
			errGetFilteringResult: errors.New("failed retrieve filtering result"),
			resGetResultElaborate: entity.ResultElaborate{
				Decision: constant.DECISION_PASS,
				LTV:      0,
			},
			resFinal: response.ElaborateResult{
				Code:           constant.CODE_PASS_ELABORATE,
				Decision:       constant.DECISION_PASS,
				Reason:         constant.REASON_PASS_ELABORATE,
				ResultPefindo:  constant.DECISION_PBK_NO_HIT,
				Cluster:        "Cluster A",
				BPKBNameType:   0,
				AgeVehicle:     ">12",
				LTVOrigin:      67,
				LTV:            0,
				TotalBakiDebet: 11000000.00,
			},
			errFinal: errors.New("failed retrieve filtering result"),
		},
		{
			name: "TEST_PASS_ResultElaborate",
			req: request.BodyRequestElaborate{
				ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
				Data: request.DataElaborate{
					ProspectID:        "NE65499A4AA35CF",
					BranchID:          "426",
					BPKBName:          "O",
					CustomerStatus:    constant.STATUS_KONSUMEN_RO_AO,
					CategoryCustomer:  "",
					ResultPefindo:     constant.DECISION_PASS,
					TotalBakiDebet:    11000000.00,
					Tenor:             12,
					ManufacturingYear: "2010",
					OTR:               35355000,
					NTF:               23571178,
				},
			},
			resGetClusterBranchElaborate: entity.ClusterBranch{
				Cluster: "Cluster A",
			},
			resGetFilteringResult: entity.ApiDupcheckKmbUpdate{
				RequestID:    "TEST123",
				PefindoID:    "1676593952",
				PefindoScore: &pefindoNotMatchValue,
			},
			resGetResultElaborate: entity.ResultElaborate{
				Decision: constant.DECISION_PASS,
				LTV:      0,
			},
			resGetMappingLtvOvd: entity.ResultElaborate{
				Cluster:  "Cluster D",
				Decision: constant.DECISION_PASS,
				LTV:      60,
			},
			resFinal: response.ElaborateResult{
				Code:           constant.CODE_PASS_ELABORATE,
				Decision:       constant.DECISION_PASS,
				Reason:         constant.REASON_PASS_ELABORATE,
				ResultPefindo:  constant.DECISION_PASS,
				Cluster:        "Cluster A",
				BPKBNameType:   0,
				AgeVehicle:     ">12",
				LTVOrigin:      67,
				LTV:            0,
				TotalBakiDebet: 11000000.00,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetClusterBranchElaborate", mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetClusterBranchElaborate, tc.errGetClusterBranchElaborate).Once()
			mockRepository.On("GetFilteringResult", mock.Anything).Return(tc.resGetFilteringResult, tc.errGetFilteringResult).Once()
			mockRepository.On("GetResultElaborate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetResultElaborate, tc.errGetResultElaborate).Once()
			mockRepository.On("GetMappingLtvOvd", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetMappingLtvOvd, tc.errGetMappingLtvOvd).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.ResultElaborate(ctx, tc.req)

			require.Equal(t, tc.resFinal, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}
