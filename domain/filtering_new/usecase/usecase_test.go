package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/filtering_new/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFiltering(t *testing.T) {
	accessToken := "token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name                 string
		req                  request.Filtering
		married              bool
		spCustomer           response.SpDupCekCustomerByID
		errspCustomer        error
		spSpouse             response.SpDupCekCustomerByID
		errspSpouse          error
		resBlackList         response.UsecaseApi
		respFilteringPefindo response.Filtering
		reqPefindo           request.Pefindo
		resPefindo           response.PefindoResult
		errpefindo           error
		trxDetailBiro        []entity.TrxDetailBiro
		resFinal             response.Filtering
		errFinal             error
	}{
		{
			name: "TEST_ERROR_DupcheckIntegrator",
			req: request.Filtering{
				ProspectID: "SAL02400020230727001",
				BPKBName:   "K",
				BranchID:   "426",
				IDNumber:   "3275066006789999",
				LegalName:  "EMI LegalName",
				BirthDate:  "1971-04-15",
				Gender:     "M",
				MotherName: "HAROEMI MotherName",
				Spouse: &request.FilteringSpouse{
					IDNumber:   "3345270510910123",
					LegalName:  "DIANA LegalName",
					BirthDate:  "1995-08-28",
					Gender:     "F",
					MotherName: "ELSA",
				},
			},
			married:       true,
			errspCustomer: errors.New("error sp"),
			errFinal:      errors.New("error sp"),
		},
		{
			name: "TEST_Reject_BlacklistCheck",
			req: request.Filtering{
				ProspectID: "SAL02400020230727001",
				BPKBName:   "K",
				BranchID:   "426",
				IDNumber:   "3275066006789999",
				LegalName:  "EMI LegalName",
				BirthDate:  "1971-04-15",
				Gender:     "M",
				MotherName: "HAROEMI MotherName",
				Spouse: &request.FilteringSpouse{
					IDNumber:   "3345270510910123",
					LegalName:  "DIANA LegalName",
					BirthDate:  "1995-08-28",
					Gender:     "F",
					MotherName: "ELSA",
				},
			},
			married: true,
			resBlackList: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   "123",
			},
			resFinal: response.Filtering{
				ProspectID:  "SAL02400020230727001",
				Decision:    constant.DECISION_REJECT,
				Code:        "123",
				IsBlacklist: true,
			},
		},
		{
			name: "TEST_err_FilteringPefindo",
			req: request.Filtering{
				ProspectID: "SAL02400020230727001",
				BPKBName:   "K",
				BranchID:   "426",
				IDNumber:   "3275066006789999",
				LegalName:  "EMI LegalName",
				BirthDate:  "1971-04-15",
				Gender:     "M",
				MotherName: "HAROEMI MotherName",
				Spouse: &request.FilteringSpouse{
					IDNumber:   "3345270510910123",
					LegalName:  "DIANA LegalName",
					BirthDate:  "1995-08-28",
					Gender:     "F",
					MotherName: "ELSA",
				},
			},
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "K",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			married:    true,
			errpefindo: errors.New("error pefindo"),
			errFinal:   errors.New("error pefindo"),
		},
		{
			name: "TEST_err_FilteringPefindo",
			req: request.Filtering{
				ProspectID: "SAL02400020230727001",
				BPKBName:   "K",
				BranchID:   "426",
				IDNumber:   "3275066006789999",
				LegalName:  "EMI LegalName",
				BirthDate:  "1971-04-15",
				Gender:     "M",
				MotherName: "HAROEMI MotherName",
				Spouse: &request.FilteringSpouse{
					IDNumber:   "3345270510910123",
					LegalName:  "DIANA LegalName",
					BirthDate:  "1995-08-28",
					Gender:     "F",
					MotherName: "ELSA",
				},
			},
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "K",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_PRIME,
			},
			resBlackList: response.UsecaseApi{
				Result: constant.DECISION_PASS,
				Code:   "123",
			},
			married: true,
			respFilteringPefindo: response.Filtering{
				ProspectID:      "SAL02400020230727001",
				Decision:        constant.DECISION_PASS,
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_REGULAR,
				NextProcess:     true,
				Code:            "123",
			},
			resPefindo: response.PefindoResult{
				Score:       "Average Risk",
				WoContract:  true,
				WoAdaAgunan: true,
			},
			resFinal: response.Filtering{
				ProspectID:      "SAL02400020230727001",
				Decision:        constant.DECISION_PASS,
				CustomerStatus:  constant.STATUS_KONSUMEN_RO,
				CustomerSegment: constant.RO_AO_PRIME,
				NextProcess:     true,
				Code:            "123",
				Reason:          "RO PRIME",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockUsecase := new(mocks.Usecase)
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockUsecase.On("SaveFiltering", mock.Anything, mock.Anything).Return(nil)

			mockUsecase.On("DupcheckIntegrator", ctx, tc.req.ProspectID, tc.req.IDNumber, tc.req.LegalName, tc.req.BirthDate, tc.req.MotherName, accessToken).Return(tc.spCustomer, tc.errspCustomer).Once()
			if tc.married {
				mockUsecase.On("DupcheckIntegrator", ctx, tc.req.ProspectID, tc.req.Spouse.IDNumber, tc.req.Spouse.LegalName, tc.req.Spouse.BirthDate, tc.req.Spouse.MotherName, accessToken).Return(tc.spSpouse, tc.errspSpouse).Once()
			}

			mockUsecase.On("BlacklistCheck", 0, tc.spCustomer).Return(tc.resBlackList, mock.Anything).Once()
			if tc.married {
				mockUsecase.On("BlacklistCheck", 1, tc.spSpouse).Return(tc.resBlackList, mock.Anything).Once()
			}

			mockUsecase.On("FilteringPefindo", ctx, tc.reqPefindo, mock.Anything, accessToken).Return(tc.respFilteringPefindo, tc.resPefindo, tc.trxDetailBiro, tc.errpefindo).Once()

			multiUsecase := NewMultiUsecase(mockRepository, mockHttpClient, mockUsecase)

			result, err := multiUsecase.Filtering(ctx, tc.req, tc.married, accessToken)

			require.Equal(t, tc.resFinal, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestFilteringPefindo(t *testing.T) {
	os.Setenv("NAMA_SAMA", "K,P")
	os.Setenv("ACTIVE_PBK", "true")
	os.Setenv("DUMMY_PBK", "false")
	timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

	accessToken := "token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	pbkCustomer := "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf"
	pbkSpouse := "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79ab36b6d_5793480682.pdf"

	testcases := []struct {
		name                 string
		checkPefindo         response.ResponsePefindo
		pefindoResult        response.PefindoResult
		customerStatus       string
		reqPefindo           request.Pefindo
		respFilteringPefindo response.Filtering
		resPefindo           response.PefindoResult
		trxDetailBiro        []entity.TrxDetailBiro
		errPefindo           error
		rPefindoCode         int
		rPefindoBody         string
		reqMappingCluster    entity.MasterMappingCluster
		resMappingCluster    entity.MasterMappingCluster
		errMappingCluster    error
		errFinal             error
	}{
		{
			name:           "test pefindo bpkb sama",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "K",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   1,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   1,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9103",
				Decision:          constant.DECISION_PASS,
				Reason:            "Nama Sama & PBK OVD 12 Bulan Terakhir <= 60 & OVD Current <= 30",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:               "kp_64bb79a65a904",
				PefindoID:              "6108521441",
				Score:                  "AVERAGE RISK",
				MaxOverdue:             float64(0),
				MaxOverdueLast12Months: float64(0),
				AngsuranAktifPbk:       float64(702009),
				DetailReport:           "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
			},
		},
		{
			name:           "test pefindo reject 12 bpkb sama",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "K",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":70,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":70,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   1,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   1,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9108",
				Decision:          constant.DECISION_REJECT,
				Reason:            "Nama Sama & Baki Debet <= 3 Juta",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:               "kp_64bb79a65a904",
				PefindoID:              "6108521441",
				Score:                  "AVERAGE RISK",
				MaxOverdue:             float64(0),
				MaxOverdueLast12Months: float64(70),
				AngsuranAktifPbk:       float64(702009),
				DetailReport:           "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
			},
		},
		{
			name:           "test pefindo reject current bpkb sama",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "K",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":0,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   1,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   1,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9108",
				Decision:          constant.DECISION_REJECT,
				Reason:            "Nama Sama & Baki Debet <= 3 Juta",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:               "kp_64bb79a65a904",
				PefindoID:              "6108521441",
				Score:                  "AVERAGE RISK",
				MaxOverdue:             float64(40),
				MaxOverdueLast12Months: float64(0),
				AngsuranAktifPbk:       float64(702009),
				DetailReport:           "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
			},
		},
		{
			name:           "test pefindo pass bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "KK",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9096",
				Decision:          constant.DECISION_PASS,
				Reason:            "Nama Beda & PBK OVD 12 Bulan Terakhir <= 60 & OVD Current <= 30",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:               "kp_64bb79a65a904",
				PefindoID:              "6108521441",
				Score:                  "AVERAGE RISK",
				MaxOverdue:             float64(0),
				MaxOverdueLast12Months: float64(0),
				AngsuranAktifPbk:       float64(702009),
				DetailReport:           "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
			},
		},
		{
			name:           "test pefindo reject bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "KK",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":61,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":61,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9108",
				Decision:          constant.DECISION_REJECT,
				Reason:            "Nama Beda & Baki Debet <= 3 Juta",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:               "kp_64bb79a65a904",
				PefindoID:              "6108521441",
				Score:                  "AVERAGE RISK",
				MaxOverdue:             float64(0),
				MaxOverdueLast12Months: float64(61),
				AngsuranAktifPbk:       float64(702009),
				DetailReport:           "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:         os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:          constant.USER_PBK_KMB_FILTEERING,
				User:              constant.USER_PBK_KMB_FILTEERING,
				ProspectID:        "SAL02400020230727001",
				BPKBName:          "KK",
				BranchID:          "426",
				IDNumber:          "3275066006789999",
				LegalName:         "EMI LegalName",
				BirthDate:         "1971-04-15",
				Gender:            "M",
				SurgateMotherName: "HAROEMI MotherName",
				Spouse: &request.SpousePefindo{
					IDNumber:          "3345270510910123",
					LegalName:         "DIANA LegalName",
					BirthDate:         "1995-08-28",
					Gender:            "F",
					SurgateMotherName: "ELSA",
				},
				MaritalStatus: "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":0,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "AO/RO",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9108",
				Decision:          constant.DECISION_REJECT,
				Reason:            "Nama Beda & Baki Debet <= 3 Juta",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:               "kp_64bb79a65a904",
				PefindoID:              "6108521441",
				Score:                  "AVERAGE RISK",
				MaxOverdue:             float64(40),
				MaxOverdueLast12Months: float64(0),
				AngsuranAktifPbk:       float64(702009),
				DetailReport:           "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("MasterMappingCluster", tc.reqMappingCluster).Return(tc.resMappingCluster, tc.errMappingCluster)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("NEW_KMB_PBK_URL"), httpmock.NewStringResponder(tc.rPefindoCode, tc.rPefindoBody))
			resp, _ := rst.R().Post(os.Getenv("NEW_KMB_PBK_URL"))

			param, _ := json.Marshal(tc.reqPefindo)
			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("NEW_KMB_PBK_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, tc.reqPefindo.ProspectID, accessToken).Return(resp, tc.errPefindo).Once()
			usecase := NewUsecase(mockRepository, mockHttpClient)

			rFilteringPefindo, rPefindo, _, err := usecase.FilteringPefindo(ctx, tc.reqPefindo, tc.customerStatus, accessToken)
			require.Equal(t, tc.respFilteringPefindo, rFilteringPefindo)
			require.Equal(t, tc.resPefindo, rPefindo)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestDupcheckIntegrator(t *testing.T) {
	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	accessToken := "token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	ProspectID := "SAL02400020230727001"
	IDNumber := "3275066006789999"
	LegalName := "EMI LegalName"
	BirthDate := "1971-04-15"
	MotherName := "HAROEMI MotherName"

	testcases := []struct {
		name          string
		rDupcheckCode int
		rDupcheckBody string
		errHttp       error
		resDupcheck   response.SpDupCekCustomerByID
		errFinal      error
	}{
		{
			name:     "test error",
			errHttp:  errors.New("upstream_service_timeout - Call Dupcheck Timeout"),
			errFinal: errors.New("upstream_service_timeout - Call Dupcheck Timeout"),
		},
		{
			name:          "test error > 200",
			rDupcheckCode: 400,
			errFinal:      errors.New("upstream_service_error - Call Dupcheck Error"),
		},
		{
			name:          "test success",
			rDupcheckCode: 200,
			rDupcheckBody: `{ "messages": "LOS DUPCHECK", "errors": null, 
			"data": { "customer_id": null, "id_number": "", "full_name": "", "birth_date": "", "surgate_mother_name": "", "birth_place": "", "gender": "", 
			"emergency_contact_address": "", "legal_address": "", "legal_kelurahan": "", "legal_kecamatan": "", "legal_city": "", "lagal_zipcode": "", 
			"residence_address": "", "residence_kelurahan": "", "residence_kecamatan": "", "residence_city": "", "residence_zipcode": "", "company_address": "", 
			"company_kelurahan": "", "company_kecamatan": "", "company_city": "", "company_zipcode": "", "personal_npwp": "", "education": "", "marital_status": "", 
			"num_of_dependence": 0, "home_status": "", "profession_id": "", "job_type_id": "", "job_pos": null, "monthly_fixed_income": 0, "spouse_income": null, 
			"monthly_variable_income": 0, "total_installment": 0, "total_installment_nap": 0, "bad_type": null, "max_overduedays": 0, "max_overduedays_roao": null, 
			"num_of_asset_inventoried": 0, "overduedays_aging": null, "max_overduedays_for_active_agreement": null, "max_overduedays_for_prev_eom": null, 
			"sisa_jumlah_angsuran": null, "rrd_date": null, "number_of_agreement": 0, "work_since_year": null, "outstanding_principal": 0, "os_installmentdue": 0, 
			"is_restructure": 0, "is_similiar": 0, "customer_status": "", "customer_segment": "" }, "server_time": "2023-08-07T08:28:13+07:00", "request_id": "" }`,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("DUPCHECK_URL"), httpmock.NewStringResponder(tc.rDupcheckCode, tc.rDupcheckBody))
			resp, _ := rst.R().Post(os.Getenv("DUPCHECK_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("DUPCHECK_URL"), mock.Anything, map[string]string{}, constant.METHOD_POST, false, 0, timeout, ProspectID, accessToken).Return(resp, tc.errHttp).Once()
			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.DupcheckIntegrator(ctx, ProspectID, IDNumber, LegalName, BirthDate, MotherName, accessToken)
			require.Equal(t, tc.resDupcheck, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}
