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

func TestFilteringProspectID(t *testing.T) {
	testcases := []struct {
		name       string
		prospectID string
		row        int
		result     request.OrderIDCheck
		errGet     error
		errFinal   error
	}{
		{
			name:       "test error",
			prospectID: "TEST01",
			result: request.OrderIDCheck{
				ProspectID: "TEST01 - true",
			},
			errGet:   errors.New("upstream_service_error - Get Filtering Order ID"),
			errFinal: errors.New("upstream_service_error - Get Filtering Order ID"),
		},
		{
			name:       "test ppid exist",
			prospectID: "TEST01",
			result: request.OrderIDCheck{
				ProspectID: "TEST01 - false",
			},
			row: 1,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetFilteringByID", mock.Anything).Return(tc.row, tc.errGet)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.FilteringProspectID(tc.prospectID)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestSaveFiltering(t *testing.T) {
	testcases := []struct {
		name          string
		transaction   entity.FilteringKMB
		trxDetailBiro []entity.TrxDetailBiro
		errSave       error
		errFinal      error
	}{
		{
			name:     "test error timeout",
			errSave:  errors.New("connection deadline"),
			errFinal: errors.New("upstream_service_timeout - Save Filtering Timeout"),
		},
		{
			name:     "test error save",
			errSave:  errors.New("save error"),
			errFinal: errors.New(constant.ERROR_BAD_REQUEST + " - Save Filtering Error ProspectID Already Exist"),
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("SaveFiltering", mock.Anything, mock.Anything).Return(tc.errSave)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			err := usecase.SaveFiltering(tc.transaction, tc.trxDetailBiro)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestBlacklistCheck(t *testing.T) {
	testcases := []struct {
		name         string
		spCustomer   response.SpDupCekCustomerByID
		index        int
		customerType string
		res          response.UsecaseApi
	}{
		{
			name:         "test new",
			customerType: constant.MESSAGE_BERSIH,
			res: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_NON_BLACKLIST_ALL,
				Reason:         constant.REASON_NON_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
		},
		{
			name: "test ro BadType",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				BadType:               constant.BADTYPE_B,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name:  "test ro pasangan BadType",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				BadType:               constant.BADTYPE_B,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name: "test ro MaxOverdueDays",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST_OVD_90DAYS,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name:  "test ro pasangan MaxOverdueDays",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				MaxOverdueDays:        91,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST_OVD_90DAYS,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name: "test ro NumOfAssetInventoried",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST_ASSET_INVENTORY,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name:  "test ro pasangan NumOfAssetInventoried",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus:        constant.STATUS_KONSUMEN_RO,
				NumOfAssetInventoried: 1,
				IsRestructure:         1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST_ASSET_INVENTORY,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name: "test ro IsRestructure",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				IsRestructure:  1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_KONSUMEN_BLACKLIST,
				Reason:         constant.REASON_KONSUMEN_BLACKLIST_RESTRUCTURE,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name:  "test ro pasangan IsRestructure",
			index: 1,
			spCustomer: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				IsRestructure:  1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			res: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PASANGAN_BLACKLIST,
				Reason:         constant.REASON_PASANGAN_BLACKLIST_RESTRUCTURE,
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			},
		},
		{
			name: "test new sp",
			spCustomer: response.SpDupCekCustomerByID{
				CustomerID: 12345,
			},
			customerType: constant.MESSAGE_BERSIH,
			res: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_NON_BLACKLIST_ALL,
				Reason:         constant.REASON_NON_BLACKLIST,
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, ct := usecase.BlacklistCheck(tc.index, tc.spCustomer)
			require.Equal(t, tc.res, result)
			require.Equal(t, tc.customerType, ct)
		})
	}
}

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
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
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
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
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
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0,
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
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
				Reason:            "NAMA SAMA (I) & Baki Debet Sesuai Ketentuan",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(0),
				MaxOverdueLast12Months:        float64(0),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject 12 bpkb sama",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":70,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 3,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 70},
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
				Code:              "9107",
				Decision:          constant.DECISION_REJECT,
				Reason:            "NAMA SAMA (III) & Baki Debet <= 3 Juta",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(0),
				MaxOverdueLast12Months:        float64(70),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(3),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(70),
			},
		},
		{
			name:           "test pefindo reject current bpkb sama",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 3,
			"max_overdue_ko_rules": 40,
			"max_overdue_last12months_ko_rules": 0},
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
				Code:              "9107",
				Decision:          constant.DECISION_REJECT,
				Reason:            "NAMA SAMA (III) & Baki Debet <= 3 Juta",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(0),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(3),
				MaxOverdueKORules:             float64(40),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo pass bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
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
				Reason:            "NAMA BEDA (I) & Baki Debet Sesuai Ketentuan",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(0),
				MaxOverdueLast12Months:        float64(0),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":61,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
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
				Code:              "9096",
				Decision:          constant.DECISION_REJECT,
				Reason:            "NAMA BEDA & Baki Debet > Threshold",
				CustomerStatus:    "RO",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(0),
				MaxOverdueLast12Months:        float64(61),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 3,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
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
				Code:              "9096",
				Decision:          constant.DECISION_REJECT,
				Reason:            "NAMA BEDA & Baki Debet > Threshold",
				CustomerStatus:    "RO",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(0),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(3),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":0,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
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
				Reason:            "NAMA BEDA (I) & Baki Debet Sesuai Ketentuan",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    nil,
				MaxOverdueLast12Months:        float64(0),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":null,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":null,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":null,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
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
				Reason:            "NAMA BEDA (I) & Baki Debet Sesuai Ketentuan",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    nil,
				MaxOverdueLast12Months:        nil,
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":0,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
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
				Reason:            "NAMA SAMA (I) & Baki Debet Sesuai Ketentuan",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    nil,
				MaxOverdueLast12Months:        float64(0),
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":null,"angsuran_aktif_pbk":702009,"wo_contract":false,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":null,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":null,"max_overdue_last12months":null,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
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
				Reason:            "NAMA SAMA (I) & Baki Debet Sesuai Ketentuan",
				CustomerStatus:    "RO",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    nil,
				MaxOverdueLast12Months:        nil,
				AngsuranAktifPbk:              float64(702009),
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":true,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
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
				Reason:            "NAMA SAMA (I) & Ada Fasilitas WO Agunan",
				CustomerStatus:    "RO",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				WoContract:                    true,
				WoAdaAgunan:                   true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":true,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9103",
				Decision:          constant.DECISION_PASS,
				Reason:            "NAMA SAMA (I) & Ada Fasilitas WO Agunan",
				CustomerStatus:    "NEW",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				WoContract:                    true,
				WoAdaAgunan:                   true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":true,
			"total_baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9103",
				Decision:          constant.DECISION_PASS,
				Reason:            "NAMA SAMA (I) & Baki Debet > 20 Juta",
				CustomerStatus:    "NEW",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(21000000),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				TotalBakiDebetNonAgunan:       float64(21000000),
				WoContract:                    true,
				WoAdaAgunan:                   true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":true,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9096",
				Decision:          constant.DECISION_REJECT,
				Reason:            "NAMA BEDA & Baki Debet > Threshold",
				CustomerStatus:    "NEW",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				WoContract:                    true,
				WoAdaAgunan:                   true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":true,
			"total_baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9096",
				Decision:          constant.DECISION_REJECT,
				Reason:            "NAMA BEDA & Baki Debet > Threshold",
				CustomerStatus:    "NEW",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(21000000),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				TotalBakiDebetNonAgunan:       float64(21000000),
				WoContract:                    true,
				WoAdaAgunan:                   true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9103",
				Decision:          constant.DECISION_PASS,
				Reason:            "NAMA SAMA (I) & Baki Debet Sesuai Ketentuan",
				CustomerStatus:    "NEW",
				NextProcess:       true,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				WoContract:                    true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "K",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   1,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9103",
				Decision:          constant.DECISION_PASS,
				Reason:            "NAMA SAMA (I) & Baki Debet > 20 Juta",
				CustomerStatus:    "NEW",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(21000000),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				TotalBakiDebetNonAgunan:       float64(21000000),
				WoContract:                    true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"200","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:        "SAL02400020230727001",
				Code:              "9096",
				Decision:          constant.DECISION_REJECT,
				Reason:            "NAMA BEDA & Baki Debet > Threshold",
				CustomerStatus:    "NEW",
				NextProcess:       false,
				PbkReportCustomer: &pbkCustomer,
				PbkReportSpouse:   &pbkSpouse,
				TotalBakiDebet:    float64(0),
				Cluster:           "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				WoContract:                    true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_RO,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"201","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
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
				ProspectID:     "SAL02400020230727001",
				Code:           "9190",
				Decision:       constant.DECISION_PASS,
				Reason:         "PBK Tidak Ditemukan - RO",
				CustomerStatus: "RO",
				NextProcess:    true,
				Cluster:        "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				TotalBakiDebetNonAgunan:       float64(21000000),
				WoContract:                    true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"201","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:     "SAL02400020230727001",
				Code:           "9192",
				Decision:       constant.DECISION_PASS,
				Reason:         "PBK Tidak Ditemukan - NEW",
				CustomerStatus: "NEW",
				NextProcess:    true,
				Cluster:        "Cluster B",
			},
			resPefindo: response.PefindoResult{
				SearchID:                      "kp_64bb79a65a904",
				PefindoID:                     "6108521441",
				Score:                         "AVERAGE RISK",
				MaxOverdue:                    float64(40),
				MaxOverdueLast12Months:        float64(45),
				AngsuranAktifPbk:              float64(702009),
				TotalBakiDebetNonAgunan:       float64(21000000),
				WoContract:                    true,
				DetailReport:                  "http://10.0.0.170/los-symlink/pefindo/pdf/pdf_kp_64bb79a65a904_6108521441.pdf",
				Category:                      float64(1),
				MaxOverdueKORules:             float64(0),
				MaxOverdueLast12MonthsKORules: float64(0),
			},
		},
		{
			name:           "test pefindo reject current bpkb beda",
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			reqPefindo: request.Pefindo{
				ClientKey:               os.Getenv("CLIENTKEY_CORE_PBK"),
				IDMember:                constant.USER_PBK_KMB_FILTEERING,
				User:                    constant.USER_PBK_KMB_FILTEERING,
				ProspectID:              "SAL02400020230727001",
				BPKBName:                "KK",
				BranchID:                "426",
				IDNumber:                "3275066006789999",
				LegalName:               "EMI LegalName",
				BirthDate:               "1971-04-15",
				Gender:                  "M",
				SurgateMotherName:       "HAROEMI MotherName",
				SpouseIDNumber:          "3345270510910123",
				SpouseLegalName:         "DIANA LegalName",
				SpouseBirthDate:         "1995-08-28",
				SpouseGender:            "F",
				SpouseSurgateMotherName: "ELSA",
				MaritalStatus:           "M",
			},
			rPefindoCode: 200,
			rPefindoBody: `{"code":"202","status":"SUCCESS",
			"result":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":702009,"wo_contract":true,"wo_ada_agunan":false,
			"total_baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf",
			"category": 1,
			"max_overdue_ko_rules": 0,
			"max_overdue_last12months_ko_rules": 0},
			"konsumen":{"search_id":"kp_64bb79a65a904","pefindo_id":"6108521441","score":"AVERAGE RISK",
			"max_overdue":40,"max_overdue_last12months":45,"angsuran_aktif_pbk":0,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":21000000,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79a65a904_6108521441.pdf"},
			"pasangan":{"search_id":"kp_64bb79ab36b6d","pefindo_id":"5793480682","score":"AVERAGE RISK",
			"max_overdue":0,"max_overdue_last12months":0,"angsuran_aktif_pbk":702009,"wo_contract":0,"wo_ada_agunan":0,
			"baki_debet_non_agunan":0,"detail_report":"http:\/\/10.0.0.170\/los-symlink\/pefindo\/pdf\/pdf_kp_64bb79ab36b6d_5793480682.pdf"},
			"server_time":"2023-07-22T13:39:45+07:00","duration_time":"11000 ms"}`,
			reqMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
			},
			resMappingCluster: entity.MasterMappingCluster{
				BranchID:       "426",
				CustomerStatus: "NEW",
				BpkbNameType:   0,
				Cluster:        "Cluster B",
			},
			respFilteringPefindo: response.Filtering{
				ProspectID:     "SAL02400020230727001",
				Code:           "9195",
				Reason:         "No Hit PBK",
				CustomerStatus: "NEW",
				Decision:       constant.DECISION_PASS,
				NextProcess:    true,
				Cluster:        "Cluster B",
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

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("NEW_KMB_DUPCHECK_URL"), httpmock.NewStringResponder(tc.rDupcheckCode, tc.rDupcheckBody))
			resp, _ := rst.R().Post(os.Getenv("NEW_KMB_DUPCHECK_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("NEW_KMB_DUPCHECK_URL"), mock.Anything, map[string]string{}, constant.METHOD_POST, false, 0, timeout, ProspectID, accessToken).Return(resp, tc.errHttp).Once()
			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.DupcheckIntegrator(ctx, ProspectID, IDNumber, LegalName, BirthDate, MotherName, accessToken)
			require.Equal(t, tc.resDupcheck, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}

func TestGetResultFiltering(t *testing.T) {
	testcases := []struct {
		name                  string
		prospectID            string
		getResultFiltering    []entity.ResultFiltering
		errgetResultFiltering error
		errFinal              error
		respFiltering         response.Filtering
	}{
		{
			name:       "test_get_result_filtering_err",
			prospectID: "SAL-1234567890",
			errFinal:   errors.New(constant.ERROR_UPSTREAM + " - Get Result Filtering Error"),
		},
		{
			name:       "test_get_result_filtering_success",
			prospectID: "SAL-1234567890",
			getResultFiltering: []entity.ResultFiltering{
				{
					ProspectID:                      "SAL-1234567890",
					TotalBakiDebetNonCollateralBiro: 100000,
					Subject:                         "CUSTOMER",
					UrlPdfReport:                    "http://urlpdf.com/customer.pdf",
				},
				{
					ProspectID:                      "SAL-1234567890",
					TotalBakiDebetNonCollateralBiro: 100000,
					Subject:                         "SPOUSE",
					UrlPdfReport:                    "http://urlpdf.com/spouse.pdf",
				},
			},
			respFiltering: response.Filtering{
				ProspectID:        "SAL-1234567890",
				TotalBakiDebet:    float64(100000),
				PbkReportCustomer: "http://urlpdf.com/customer.pdf",
				PbkReportSpouse:   "http://urlpdf.com/spouse.pdf",
			},
		},
		{
			name:       "test_get_result_filtering_err_conv_float",
			prospectID: "SAL-1234567890",
			getResultFiltering: []entity.ResultFiltering{
				{
					ProspectID:                      "SAL-1234567890",
					TotalBakiDebetNonCollateralBiro: "abcd",
					Subject:                         "CUSTOMER",
					UrlPdfReport:                    "http://urlpdf.com/customer.pdf",
				},
			},
			respFiltering: response.Filtering{
				ProspectID: "SAL-1234567890",
			},
			errFinal: errors.New(constant.ERROR_UPSTREAM + " - GetResultFiltering GetFloat Error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetResultFiltering", tc.prospectID).Return(tc.getResultFiltering, tc.errgetResultFiltering)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.GetResultFiltering(tc.prospectID)
			require.Equal(t, tc.respFiltering, result)
			require.Equal(t, tc.errFinal, err)
		})
	}
}
