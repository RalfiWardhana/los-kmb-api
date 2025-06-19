package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMaxLoanAmount(t *testing.T) {
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
	accessToken := "test-token"
	referralCode := "test"
	os.Setenv("MI_NUMBER_WHITELIST", "123,1234")
	os.Setenv("NAMA_SAMA", "K,P")

	testCases := []struct {
		name                      string
		request                   request.GetMaxLoanAmount
		config                    entity.AppConfig
		errConfig                 error
		trxKPM                    entity.TrxKPM
		errTrxKPM                 error
		customerDetailResponse    response.MDMGetDetailCustomerKPMResponse
		errCustomerDetail         error
		dupcheckResponse          response.SpDupCekCustomerByID
		errDupcheck               error
		assetResponse             response.AssetList
		errAsset                  error
		marsevFilterProgramRes    response.MarsevFilterProgramResponse
		errMarsevFilterProgram    error
		assetYearResponse         response.AssetYearList
		errAssetYear              error
		mappingBranchResponse     response.MDMMasterMappingBranchEmployeeResponse
		errMappingBranch          error
		hrisResponse              response.EmployeeCMOResponse
		errHris                   error
		fpdCMOResponse            response.FpdCMOResponse
		errFpdCMO                 error
		mappingFpdCluster         entity.MasterMappingFpdCluster
		errMappingFpdCluster      error
		savedClusterCheckCmoNoFPD string
		entitySaveTrxNoFPd        entity.TrxCmoNoFPD
		errCheckCmoNoFPD          error
		mappingElaborateLTV       []entity.MappingElaborateLTV
		errMappingElaborateLTV    error
		mappingBranch             entity.MappingBranchByPBKScore
		errMappingBranchEntity    error
		trxDetailBiro             []entity.TrxDetailBiro
		pbkScore                  string
		errTrxDetailBiro          error
		getLTVResponse            int
		adjustTenorResponse       bool
		isSimulasi                bool
		errGetLTV                 error
		marsevLoanAmountRes       response.MarsevLoanAmountResponse
		errMarsevLoanAmount       error
		rejectTenorResponse       response.UsecaseApi
		errRejectTenor            error
		expectedResponse          response.GetMaxLoanAmountData
		mappingPbkScoreGrade      entity.MappingPBKScoreGrade
		errMappingPbkScoreGrade   error
		expectedError             error
	}{
		{
			name: "success simulation case",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			pbkScore: "NO HOT",
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "AVERAGE RISK",
				},
			},
			mappingBranch: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						Tenors: []response.TenorInfo{
							{Tenor: 12},
							{Tenor: 24},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 80,
				},
			},
			getLTVResponse: 80,
			marsevLoanAmountRes: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 40000000,
				},
			},
			expectedResponse: response.GetMaxLoanAmountData{
				MaxLoanAmount: 40000000,
			},
		},
		{
			name: "success with referral code simulation case",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				ReferralCode:       &referralCode,
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "NO HIT",
				},
			},
			mappingPbkScoreGrade: entity.MappingPBKScoreGrade{
				GradeScore: "GOOD",
			},
			mappingBranch: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 12345,
						Tenors: []response.TenorInfo{
							{Tenor: 12},
							{Tenor: 24},
						},
					},
					{
						MINumber: 123,
						Tenors: []response.TenorInfo{
							{Tenor: 3},
							{Tenor: 6},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.NEW,
			},
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 80,
				},
			},
			getLTVResponse: 80,
			marsevLoanAmountRes: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 40000000,
				},
			},
			expectedResponse: response.GetMaxLoanAmountData{
				MaxLoanAmount: 40000000,
			},
		},
		{
			name: "error with referral code simulation case",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				ReferralCode:       &referralCode,
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "NO HIT",
				},
			},
			mappingBranch: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 1,
						Tenors: []response.TenorInfo{
							{Tenor: 12},
							{Tenor: 24},
						},
					},
					{
						MINumber: 1234123,
						Tenors: []response.TenorInfo{
							{Tenor: 3},
							{Tenor: 6},
						},
					},
				},
			},
			expectedError: errors.New(constant.ERROR_BAD_REQUEST + " - No matching MI_NUMBER found"),
		},
		{
			name: "error get detail customer kpm",
			request: request.GetMaxLoanAmount{
				ProspectID: "SIM-123",
				KPMID:      6278,
			},
			errCustomerDetail: errors.New("get detail customer kpm error"),
			expectedError:     errors.New("get detail customer kpm error"),
		},
		{
			name: "error get config",
			request: request.GetMaxLoanAmount{
				ProspectID: "SIM-123",
				KPMID:      6278,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			errConfig:     errors.New("config error"),
			expectedError: errors.New("config error"),
		},
		{
			name: "error get trx kpm",
			request: request.GetMaxLoanAmount{
				ProspectID: "123",
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			errTrxKPM:     errors.New("trx kpm error"),
			expectedError: errors.New("trx kpm error"),
		},
		{
			name: "error dupcheck",
			request: request.GetMaxLoanAmount{
				ProspectID: "123",
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			errDupcheck:   errors.New("dupcheck error"),
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Get Data Customer Error"),
		},
		{
			name: "error get master asset",
			request: request.GetMaxLoanAmount{
				ProspectID: "123",
				BranchID:   "123",
				AssetCode:  "MOT",
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
			},
			errAsset:      errors.New("asset error"),
			expectedError: errors.New("asset error"),
		},
		{
			name: "error get marketing program",
			request: request.GetMaxLoanAmount{
				ProspectID: "123",
				BranchID:   "123",
				AssetCode:  "MOT",
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			errMarsevFilterProgram: errors.New("marketing program error"),
			expectedError:          errors.New("marketing program error"),
		},
		{
			name: "error get asset year list",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			errAssetYear:  errors.New("failed to get asset year list"),
			expectedError: errors.New("failed to get asset year list"),
		},
		{
			name: "error get mapping branch employee",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			errMappingBranch: errors.New("failed to get mapping branch employee"),
			expectedError:    errors.New("failed to get mapping branch employee"),
		},
		{
			name: "error get cmo dedicated not found",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{},
			},
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - CMO Dedicated Not Found"),
		},
		{
			name: "error get employee data",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			errHris:       errors.New("failed to get employee data"),
			expectedError: errors.New("failed to get employee data"),
		},
		{
			name: "error cmo category not found",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "",
			},
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - CMO Not Found"),
		},
		{
			name: "error get fpd cmo",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			errFpdCMO:     errors.New("failed to get fpd cmo"),
			expectedError: errors.New("failed to get fpd cmo"),
		},
		{
			name: "error get master mapping fpd cluster",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			errMappingFpdCluster: errors.New("failed to get master mapping fpd cluster"),
			expectedError:        errors.New("failed to get master mapping fpd cluster"),
		},
		{
			name: "error check cmo no fpd",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			mappingFpdCluster: entity.MasterMappingFpdCluster{
				Cluster: "",
			},
			errCheckCmoNoFPD: errors.New("failed to check cmo no fpd"),
			expectedError:    errors.New("failed to check cmo no fpd"),
		},
		{
			name: "error get trx detail biro",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			errTrxDetailBiro: errors.New(constant.ERROR_UPSTREAM + " - Get Trx Detail Biro Error"),
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			mappingFpdCluster: entity.MasterMappingFpdCluster{
				Cluster: "Cluster C",
			},
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Get Trx Detail Biro Error"),
		},
		{
			name:       "error get mapping pbk score",
			isSimulasi: false,
			request: request.GetMaxLoanAmount{
				ProspectID:         "sSIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			errMappingPbkScoreGrade: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Pbk Score Error"),
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "LOW RISK",
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			mappingFpdCluster: entity.MasterMappingFpdCluster{
				Cluster: "Cluster C",
			},
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Pbk Score Error"),
		},
		{
			name: "error get mapping branch",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "NO HIT",
				},
			},
			errMappingBranchEntity: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Branch Error"),
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			mappingFpdCluster: entity.MasterMappingFpdCluster{
				Cluster: "Cluster C",
			},
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Branch Error"),
		},
		{
			name: "error get mapping elaborate ltv",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "NO HIT",
				},
			},
			mappingBranch: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			mappingFpdCluster: entity.MasterMappingFpdCluster{
				Cluster: "Cluster C",
			},
			errMappingElaborateLTV: errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error"),
			expectedError:          errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error"),
		},
		{
			name: "error get ltv",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "NO HIT",
				},
			},
			mappingBranch: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			errGetLTV:     errors.New("failed to get ltv"),
			expectedError: errors.New("failed to get ltv"),
		},
		{
			name: "error get ltv 0",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SAL-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "NO HIT",
				},
			},
			mappingBranch: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			getLTVResponse: 0,
		},
		{
			name: "error marsev get loan amount",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "NO HIT",
				},
			},
			mappingBranch: entity.MappingBranchByPBKScore{
				GradeBranch: "GOOD",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			getLTVResponse:      80,
			errMarsevLoanAmount: errors.New("failed to get loan amount"),
			expectedError:       errors.New("failed to get loan amount"),
		},
		{
			name: "error reject tenor 36",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 36,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			errRejectTenor: errors.New("failed to get reject tenor"),
			expectedError:  errors.New("failed to get reject tenor"),
		},
		{
			name: "error tenor > 36",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 40,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
		},
		{
			name: "error tenor 36 get result reject",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				AssetUsageTypeCode: "P",
				KPMID:              6287,
			},
			customerDetailResponse: response.MDMGetDetailCustomerKPMResponse{
				Data: response.GetCustomerByKpmID{
					Customer: response.Customer{
						IdNumber:          "1234567890",
						LegalName:         "Test User",
						BirthDate:         "1990-01-01",
						SurgateMotherName: "Mother Name",
					},
				},
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			assetResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160",
						AssetTypeID:         "2W",
						BranchID:            "123",
						Brand:               "HONDA",
						CategoryID:          "CAT1",
						CategoryDescription: "Sport",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			marsevFilterProgramRes: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID: "1234",
						Tenors: []response.TenorInfo{
							{
								Tenor: 36,
							},
						},
					},
				},
			},
			assetYearResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT",
						BranchID:         "123",
						Brand:            "HONDA",
						ManufactureYear:  2020,
						MarketPriceValue: 60000000,
					},
				},
			},
			mappingBranchResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			fpdCMOResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			mappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			rejectTenorResponse: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockUsecase := new(mocks.Usecase)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockUsecase.On("MDMGetDetailCustomerKPM", ctx, tc.request.ProspectID, tc.request.KPMID, accessToken).Return(tc.customerDetailResponse, tc.errCustomerDetail)

			if tc.errCustomerDetail == nil {
				mockRepository.On("GetConfig", constant.GROUP_2WILEN, "KMB-OFF", constant.KEY_PPID_SIMULASI).Return(tc.config, tc.errConfig)

				if tc.errConfig == nil {
					re := regexp.MustCompile(tc.config.Value)
					isSimulasi := re.MatchString(tc.request.ProspectID)

					if !isSimulasi {
						mockRepository.On("GetTrxKPM", tc.request.ProspectID).Return(tc.trxKPM, tc.errTrxKPM)
					}

					if tc.errTrxKPM == nil {
						mockUsecase.On("DupcheckIntegrator", ctx, tc.request.ProspectID, mock.Anything,
							mock.Anything, mock.Anything, mock.Anything, accessToken).Return(tc.dupcheckResponse, tc.errDupcheck)

						if tc.errDupcheck == nil {
							mockUsecase.On("MDMGetMasterAsset", ctx, tc.request.BranchID, tc.request.AssetCode,
								tc.request.ProspectID, accessToken).Return(tc.assetResponse, tc.errAsset)

							if tc.errAsset == nil {
								mockUsecase.On("MarsevGetMarketingProgram", ctx, mock.Anything, tc.request.ProspectID,
									accessToken).Return(tc.marsevFilterProgramRes, tc.errMarsevFilterProgram)

								if tc.errMarsevFilterProgram == nil && len(tc.marsevFilterProgramRes.Data) > 0 {
									mockUsecase.On("MDMGetAssetYear", ctx, tc.request.BranchID, tc.request.AssetCode,
										tc.request.ManufactureYear, tc.request.ProspectID, accessToken).Return(tc.assetYearResponse, tc.errAssetYear)

									if tc.errAssetYear == nil {
										mockUsecase.On("MDMGetMasterMappingBranchEmployee", ctx, tc.request.ProspectID,
											tc.request.BranchID, accessToken).Return(tc.mappingBranchResponse, tc.errMappingBranch)

										if tc.errMappingBranch == nil && len(tc.mappingBranchResponse.Data) > 0 {
											mockUsecase.On("GetEmployeeData", ctx, tc.mappingBranchResponse.Data[0].CMOID).Return(tc.hrisResponse, tc.errHris)

											if tc.errHris == nil {
												if tc.hrisResponse.CMOCategory != constant.NEW {
													mockUsecase.On("GetFpdCMO", ctx, tc.mappingBranchResponse.Data[0].CMOID, mock.Anything).Return(tc.fpdCMOResponse, tc.errFpdCMO)
												}

												if tc.fpdCMOResponse.FpdExist {
													mockRepository.On("MasterMappingFpdCluster", mock.Anything).Return(tc.mappingFpdCluster, tc.errMappingFpdCluster)
												}

												if !tc.fpdCMOResponse.FpdExist || (tc.fpdCMOResponse.FpdExist && tc.mappingFpdCluster.Cluster == "") {
													mockUsecase.On("CheckCmoNoFPD", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.savedClusterCheckCmoNoFPD, tc.entitySaveTrxNoFPd, tc.errCheckCmoNoFPD)
												}

												if tc.errCheckCmoNoFPD == nil {
													mockRepository.On("GetTrxDetailBIro", tc.request.ProspectID).Return(tc.trxDetailBiro, tc.errTrxDetailBiro)

													if tc.errTrxDetailBiro == nil {
														if !isSimulasi {
															mockRepository.On("GetMappingPbkScore", mock.Anything).Return(tc.mappingPbkScoreGrade, tc.errMappingPbkScoreGrade)
														}

														if tc.errMappingPbkScoreGrade == nil {
															mockRepository.On("GetMappingBranchByBranchID", tc.request.BranchID, mock.Anything).
																Return(tc.mappingBranch, tc.errMappingBranchEntity)

															if tc.errMappingBranchEntity == nil {
																mockRepository.On("GetMappingElaborateLTV", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.mappingElaborateLTV, tc.errMappingElaborateLTV)

																if tc.errMappingElaborateLTV == nil {
																	mockUsecase.On("GetLTV", ctx, tc.mappingElaborateLTV, tc.request.ProspectID, mock.Anything, tc.request.BPKBNameType, tc.request.ManufactureYear, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.getLTVResponse, tc.adjustTenorResponse, tc.errGetLTV)

																	if tc.marsevFilterProgramRes.Data[0].Tenors[0].Tenor == 36 {
																		mockUsecase.On("RejectTenor36", mock.Anything).Return(tc.rejectTenorResponse, tc.errRejectTenor)
																	}

																	if tc.errGetLTV == nil && tc.getLTVResponse > 0 {
																		mockUsecase.On("MarsevGetLoanAmount", ctx, mock.Anything, tc.request.ProspectID, accessToken).Return(tc.marsevLoanAmountRes, tc.errMarsevLoanAmount)
																	}
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}

			multiUsecase := NewMultiUsecase(mockRepository, mockHttpClient, nil, mockUsecase)
			result, err := multiUsecase.GetMaxLoanAmout(ctx, tc.request, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}
		})
	}
}

func TestMarsevGetLoanAmount(t *testing.T) {
	os.Setenv("MARSEV_LOAN_AMOUNT_URL", "https://dev-marsev-api.kbfinansia.com/api/v1/calculate/loan-amount")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("MARSEV_AUTHORIZATION_KEY", "marsev-test-key")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		prospectID       string
		request          request.ReqMarsevLoanAmount
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.MarsevLoanAmountResponse
		expectedError    error
	}{
		{
			name:       "success",
			prospectID: "SAL-123",
			request: request.ReqMarsevLoanAmount{
				BranchID:      "426",
				OTR:           100000000,
				MaxLTV:        80,
				IsRecalculate: false,
			},
			resEngineAPICode: 200,
			resEngineAPIBody: `{
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
			expectedResponse: response.MarsevLoanAmountResponse{
				Code:    200,
				Message: "Success",
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum:  80000000,
					AmountOfFinance:    75000000,
					DpAmount:           20000000,
					DpPercentFinal:     20,
					LtvPercentFinal:    80,
					AdminFeeAmount:     2500000,
					ProvisionFeeAmount: 1500000,
					LoanAmountFinal:    75000000,
					IsPsa:              true,
				},
				Errors: nil,
			},
			expectedError: nil,
		},
		{
			name:       "error api response",
			prospectID: "SAL-123",
			request: request.ReqMarsevLoanAmount{
				BranchID:      "426",
				OTR:           100000000,
				MaxLTV:        80,
				IsRecalculate: false,
			},
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New("network error"),
		},
		{
			name:       "not 200 status code",
			prospectID: "SAL-123",
			request: request.ReqMarsevLoanAmount{
				BranchID:      "426",
				OTR:           100000000,
				MaxLTV:        80,
				IsRecalculate: false,
			},
			resEngineAPICode: 400,
			resEngineAPIBody: `{
				"code": 400,
				"message": "Bad Request",
				"data": null,
				"errors": {
					"branch_id": "branch_id is required"
				}
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Loan Amount Error"),
		},
		{
			name:       "invalid json response",
			prospectID: "SAL-123",
			request: request.ReqMarsevLoanAmount{
				BranchID:      "426",
				OTR:           100000000,
				MaxLTV:        80,
				IsRecalculate: false,
			},
			resEngineAPICode: 200,
			resEngineAPIBody: `invalid json`,
			expectedError:    errors.New("unexpected end of JSON input"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("MARSEV_LOAN_AMOUNT_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Post(url)

			param, _ := json.Marshal(tc.request)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, param, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
			}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MarsevGetLoanAmount(ctx, tc.request, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}

func TestMarsevGetMarketingProgram(t *testing.T) {
	os.Setenv("MARSEV_FILTER_PROGRAM_URL", "https://dev-marsev-api.kbfinansia.com/api/v1/marketing-programs/filter")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("MARSEV_AUTHORIZATION_KEY", "marsev-test-key")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		prospectID       string
		request          request.ReqMarsevFilterProgram
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.MarsevFilterProgramResponse
		expectedError    error
	}{
		{
			name:       "success",
			prospectID: "SAL-123",
			request: request.ReqMarsevFilterProgram{
				Page:                   1,
				Limit:                  10,
				BranchID:               "426",
				FinancingTypeCode:      "PM",
				CustomerOccupationCode: "KRYSW",
				BpkbStatusCode:         "DN",
				SourceApplication:      "KPM",
				CustomerType:           "NEW",
				AssetUsageTypeCode:     "P",
				AssetCategory:          "MOT",
				AssetBrand:             "HONDA",
				AssetYear:              2024,
				LoanAmount:             50000000,
				SalesMethodID:          5,
			},
			resEngineAPICode: 200,
			resEngineAPIBody: `{
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
						},
						{
							"tenor": 24,
							"interest_rate": 11.5,
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
			expectedResponse: response.MarsevFilterProgramResponse{
				Code:    200,
				Message: "Success",
				Data: []response.MarsevFilterProgramData{
					{
						ID:                         "PROG001",
						ProgramName:                "Program Test",
						MINumber:                   123,
						PeriodStart:                "2024-01-01",
						PeriodEnd:                  "2024-12-31",
						Priority:                   1,
						Description:                "Test Program Description",
						ProductID:                  "PROD001",
						ProductOfferingID:          "OFF001",
						ProductOfferingDescription: "Standard Offering",
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
				},
				PageInfo: map[string]interface{}{
					"total": float64(1),
					"page":  float64(1),
					"limit": float64(10),
				},
				Errors: nil,
			},
			expectedError: nil,
		},
		{
			name:       "error api response",
			prospectID: "SAL-123",
			request: request.ReqMarsevFilterProgram{
				Page:     1,
				Limit:    10,
				BranchID: "426",
			},
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New("network error"),
		},
		{
			name:       "not 200 status code",
			prospectID: "SAL-123",
			request: request.ReqMarsevFilterProgram{
				Page:     1,
				Limit:    10,
				BranchID: "426",
			},
			resEngineAPICode: 400,
			resEngineAPIBody: `{
				"code": 400,
				"message": "Bad Request",
				"data": null,
				"errors": {
					"branch_id": "branch_id is required"
				}
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Filter Program Error"),
		},
		{
			name:       "invalid json response",
			prospectID: "SAL-123",
			request: request.ReqMarsevFilterProgram{
				Page:     1,
				Limit:    10,
				BranchID: "426",
			},
			resEngineAPICode: 200,
			resEngineAPIBody: `invalid json`,
			expectedError:    errors.New("unexpected end of JSON input"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("MARSEV_FILTER_PROGRAM_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Post(url)

			param, _ := json.Marshal(tc.request)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, param, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
			}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MarsevGetMarketingProgram(ctx, tc.request, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}

func TestMDMGetMasterAsset(t *testing.T) {
	os.Setenv("MDM_ASSET_URL", "https://dev-core-masterdata-asset-api.kbfinansia.com/api/v1/master-data/asset/search")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		branchID         string
		search           string
		prospectID       string
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.AssetList
		expectedError    error
	}{
		{
			name:             "success",
			branchID:         "426",
			search:           "VARIO",
			prospectID:       "SAL-123",
			resEngineAPICode: 200,
			resEngineAPIBody: `{
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
						},
						{
							"asset_code": "MOT002",
							"asset_description": "HONDA VARIO 125",
							"asset_display": "HONDA VARIO 125 CBS",
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
			expectedResponse: response.AssetList{
				Records: []struct {
					AssetCode           string `json:"asset_code"`
					AssetDescription    string `json:"asset_description"`
					AssetDisplay        string `json:"asset_display"`
					AssetTypeID         string `json:"asset_type_id"`
					BranchID            string `json:"branch_id"`
					Brand               string `json:"brand"`
					CategoryID          string `json:"category_id"`
					CategoryDescription string `json:"category_description"`
					IsElectric          bool   `json:"is_electric"`
					Model               string `json:"model"`
				}{
					{
						AssetCode:           "MOT001",
						AssetDescription:    "HONDA VARIO 160",
						AssetDisplay:        "HONDA VARIO 160 CBS",
						AssetTypeID:         "2W",
						BranchID:            "426",
						Brand:               "HONDA",
						CategoryID:          "SPM",
						CategoryDescription: "SPORT MATIC",
						IsElectric:          false,
						Model:               "VARIO",
					},
					{
						AssetCode:           "MOT002",
						AssetDescription:    "HONDA VARIO 125",
						AssetDisplay:        "HONDA VARIO 125 CBS",
						AssetTypeID:         "2W",
						BranchID:            "426",
						Brand:               "HONDA",
						CategoryID:          "SPM",
						CategoryDescription: "SPORT MATIC",
						IsElectric:          false,
						Model:               "VARIO",
					},
				},
			},
			expectedError: nil,
		},
		{
			name:          "error api response",
			branchID:      "426",
			search:        "VARIO",
			prospectID:    "SAL-123",
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout"),
		},
		{
			name:             "not 200 status code",
			branchID:         "426",
			search:           "VARIO",
			prospectID:       "SAL-123",
			resEngineAPICode: 400,
			resEngineAPIBody: `{
				"code": 400,
				"message": "Bad Request",
				"errors": {
					"branch_id": "branch_id is required"
				}
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
		{
			name:             "not 200 status code",
			branchID:         "426",
			search:           "VARIO",
			prospectID:       "SAL-123",
			resEngineAPICode: 200,
			resEngineAPIBody: `-`,
			expectedError:    errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data asset list"),
		},
		{
			name:             "empty records",
			branchID:         "426",
			search:           "NONEXISTENT",
			prospectID:       "SAL-123",
			resEngineAPICode: 200,
			resEngineAPIBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": []
				},
				"errors": null
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("MDM_ASSET_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Post(url)

			payloadAsset, _ := json.Marshal(map[string]interface{}{
				"branch_id": tc.branchID,
				"lob_id":    11,
				"page_size": 10,
				"search":    tc.search,
			})

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, payloadAsset, map[string]string{}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MDMGetMasterAsset(ctx, tc.branchID, tc.search, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}

func TestMDMGetAssetYear(t *testing.T) {
	os.Setenv("MDM_MARKETPRICE_URL", "https://dev-core-masterdata-asset-api.kbfinansia.com/api/v1/master-data/market-price/search")
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		branchID         string
		assetCode        string
		search           string
		prospectID       string
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.AssetYearList
		expectedError    error
	}{
		{
			name:             "success",
			branchID:         "426",
			assetCode:        "MOT001",
			search:           "2023",
			prospectID:       "SAL-123",
			resEngineAPICode: 200,
			resEngineAPIBody: `{
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
			expectedResponse: response.AssetYearList{
				Records: []struct {
					AssetCode        string `json:"asset_code"`
					BranchID         string `json:"branch_id"`
					Brand            string `json:"brand"`
					ManufactureYear  int    `json:"manufacturing_year"`
					MarketPriceValue int    `json:"market_price_value"`
				}{
					{
						AssetCode:        "MOT001",
						BranchID:         "426",
						Brand:            "HONDA",
						ManufactureYear:  2023,
						MarketPriceValue: 25000000,
					},
					{
						AssetCode:        "MOT001",
						BranchID:         "426",
						Brand:            "HONDA",
						ManufactureYear:  2022,
						MarketPriceValue: 22000000,
					},
				},
			},
			expectedError: nil,
		},
		{
			name:          "error api response",
			branchID:      "426",
			assetCode:     "MOT001",
			search:        "2023",
			prospectID:    "SAL-123",
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Timeout"),
		},
		{
			name:             "not 200 status code",
			branchID:         "426",
			assetCode:        "MOT001",
			search:           "2023",
			prospectID:       "SAL-123",
			resEngineAPICode: 400,
			resEngineAPIBody: `{
				"code": 400,
				"message": "Bad Request",
				"errors": {
					"branch_id": "branch_id is required"
				}
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error"),
		},
		{
			name:             "error unmarshal",
			branchID:         "426",
			assetCode:        "MOT001",
			search:           "2023",
			prospectID:       "SAL-123",
			resEngineAPICode: 200,
			resEngineAPIBody: `-`,
			expectedError:    errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data marketprice"),
		},
		{
			name:             "empty records",
			branchID:         "426",
			assetCode:        "MOT001",
			search:           "2030",
			prospectID:       "SAL-123",
			resEngineAPICode: 200,
			resEngineAPIBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": []
				},
				"errors": null
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("MDM_MARKETPRICE_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Post(url)

			param, _ := json.Marshal(map[string]interface{}{
				"branch_id":  tc.branchID,
				"asset_code": tc.assetCode,
				"search":     tc.search,
			})

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, param, map[string]string{}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MDMGetAssetYear(ctx, tc.branchID, tc.assetCode, tc.search, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}

func TestGetLTV(t *testing.T) {
	os.Setenv("NAMA_SAMA", "K,L,M")
	ctx := context.Background()
	ctx = context.WithValue(ctx, echo.HeaderXRequestID, "test-request-id")

	testcases := []struct {
		name            string
		prospectID      string
		resultPefindo   string
		pbkScore        string
		statusKonsumen  string
		bpkbName        string
		manufactureYear string
		tenor           int
		bakiDebet       float64
		mappingLTV      []entity.MappingElaborateLTV
		errSaveTrx      error
		expectedLTV     int
		expectedAdjust  bool
		expectedError   error
		gradeBranch     string
		shouldSaveTrx   bool
	}{
		{
			name:            "success no hit tenor < 36 nama sama",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PBK_NO_HIT,
			bpkbName:        "K",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PBK_NO_HIT,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  1,
					LTV:           75,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success no hit tenor < 36 nama beda, 18 tenor, bad branch, konsumen status new",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PBK_NO_HIT,
			bpkbName:        "RAMA",
			manufactureYear: "2022",
			pbkScore:        "BAD",
			gradeBranch:     "BAD",
			statusKonsumen:  "NEW",
			tenor:           18,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:             1,
					ResultPefindo:  constant.DECISION_PBK_NO_HIT,
					TenorStart:     12,
					TenorEnd:       18,
					GradeBranch:    "BAD",
					BPKBNameType:   0,
					StatusKonsumen: "NEW",
					PbkScore:       "BAD",
					LTV:            75,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success no hit tenor < 36 nama beda",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PBK_NO_HIT,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PBK_NO_HIT,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  0,
					LTV:           70,
				},
			},
			expectedLTV:    70,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success no hit tenor < 36 with multiple mapping",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PBK_NO_HIT,
			bpkbName:        "K",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PBK_NO_HIT,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  1,
					LTV:           75,
				},
				{
					ID:            2,
					ResultPefindo: constant.DECISION_PBK_NO_HIT,
					TenorStart:    25,
					TenorEnd:      36,
					BPKBNameType:  1,
					LTV:           70,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass nama sama tenor >= 36",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "K",
			manufactureYear: "2022",
			tenor:           36,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    36,
					TenorEnd:      48,
					BPKBNameType:  1,
					AgeVehicle:    "<=12",
					LTV:           80,
				},
			},
			expectedLTV:    80,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass nama sama tenor >= 36",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "K",
			manufactureYear: "2022",
			tenor:           36,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    36,
					TenorEnd:      48,
					BPKBNameType:  1,
					AgeVehicle:    "<=12",
					LTV:           80,
				},
			},
			expectedLTV:    80,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass nama beda tenor >= 36",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           36,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    36,
					TenorEnd:      48,
					BPKBNameType:  0,
					AgeVehicle:    "<=12",
					LTV:           70,
				},
			},
			expectedLTV:    70,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass nama sama tenor < 36",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "K",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  1,
					LTV:           85,
				},
			},
			expectedLTV:    85,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success reject with baki debet tenor >= 36",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_REJECT,
			bpkbName:        "K",
			manufactureYear: "2022",
			tenor:           36,
			bakiDebet:       5000000,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					TotalBakiDebetStart: 1000000,
					TotalBakiDebetEnd:   10000000,
					TenorStart:          36,
					TenorEnd:            48,
					BPKBNameType:        1,
					AgeVehicle:          "<=12",
					LTV:                 60,
				},
			},
			expectedLTV:    60,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success no hit tenor >= 36",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PBK_NO_HIT,
			bpkbName:        "K",
			manufactureYear: "2022",
			tenor:           36,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PBK_NO_HIT,
					TenorStart:    36,
					TenorEnd:      48,
					BPKBNameType:  1,
					AgeVehicle:    "<=12",
					LTV:           75,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "invalid manufacture year",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "K",
			manufactureYear: "invalid",
			tenor:           36,
			bakiDebet:       0,
			expectedError:   errors.New(constant.ERROR_BAD_REQUEST + " - Format tahun kendaraan tidak sesuai"),
			shouldSaveTrx:   false,
		},
		{
			name:            "error save trx elaborate ltv",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "K",
			manufactureYear: "2002",
			tenor:           36,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    36,
					TenorEnd:      48,
					BPKBNameType:  1,
					AgeVehicle:    "<=12",
					LTV:           80,
				},
			},
			errSaveTrx:    errors.New("database error"),
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - save elaborate ltv error"),
			shouldSaveTrx: true,
		},
		{
			name:            "success pass tenor < 36 - default LTV assignment without checking BPKBNameType",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  0,
					LTV:           75,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass tenor < 36 - multiple LTV records with different BPKBNameType",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  0,
					LTV:           75,
				},
				{
					ID:            2,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  1,
					LTV:           80,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass tenor < 36 - ignore baki debet criteria",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       5000000,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PASS,
					TenorStart:          12,
					TenorEnd:            24,
					BPKBNameType:        0,
					TotalBakiDebetStart: 1000000,
					TotalBakiDebetEnd:   10000000,
					LTV:                 75,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass tenor < 36 - with out of range LTV mapping",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    25,
					TenorEnd:      36,
					BPKBNameType:  0,
					LTV:           75,
				},
				{
					ID:            2,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  0,
					LTV:           80,
				},
			},
			expectedLTV:    80,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success reject with baki debet tenor < 36",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_REJECT,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       5000000,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_REJECT,
					TotalBakiDebetStart: 1000000,
					TotalBakiDebetEnd:   10000000,
					TenorStart:          12,
					TenorEnd:            24,
					BPKBNameType:        0,
					LTV:                 65,
				},
				{
					ID:                  2,
					ResultPefindo:       constant.DECISION_REJECT,
					TotalBakiDebetStart: 10000001,
					TotalBakiDebetEnd:   20000000,
					TenorStart:          12,
					TenorEnd:            24,
					BPKBNameType:        0,
					LTV:                 60,
				},
			},
			expectedLTV:    65,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass tenor < 36 - default LTV assignment without checking BPKBNameType",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  0,
					LTV:           75,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass tenor < 36 - multiple LTV records with different BPKBNameType",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  0,
					LTV:           75,
				},
				{
					ID:            2,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  1,
					LTV:           80,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass tenor < 36 - ignore baki debet criteria",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       5000000,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:                  1,
					ResultPefindo:       constant.DECISION_PASS,
					TenorStart:          12,
					TenorEnd:            24,
					BPKBNameType:        0,
					TotalBakiDebetStart: 1000000,
					TotalBakiDebetEnd:   10000000,
					LTV:                 75,
				},
			},
			expectedLTV:    75,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
		{
			name:            "success pass tenor < 36 - with out of range LTV mapping",
			prospectID:      "SAL-123",
			resultPefindo:   constant.DECISION_PASS,
			bpkbName:        "X",
			manufactureYear: "2022",
			tenor:           24,
			bakiDebet:       0,
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID:            1,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    25,
					TenorEnd:      36,
					BPKBNameType:  0,
					LTV:           75,
				},
				{
					ID:            2,
					ResultPefindo: constant.DECISION_PASS,
					TenorStart:    12,
					TenorEnd:      24,
					BPKBNameType:  0,
					LTV:           80,
				},
			},
			expectedLTV:    80,
			expectedAdjust: true,
			shouldSaveTrx:  true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)

			if tc.shouldSaveTrx {
				mockRepository.On("SaveTrxElaborateLTV", mock.MatchedBy(func(trx entity.TrxElaborateLTV) bool {
					return trx.ProspectID == tc.prospectID &&
						trx.RequestID == "test-request-id" &&
						trx.Tenor == tc.tenor &&
						trx.ManufacturingYear == tc.manufactureYear
				})).Return(tc.errSaveTrx)
			}

			usecase := NewUsecase(mockRepository, nil, nil)

			ltv, adjustTenor, err := usecase.GetLTV(ctx, tc.mappingLTV, tc.prospectID, tc.resultPefindo, tc.bpkbName, tc.manufactureYear, tc.tenor, tc.bakiDebet, false, tc.pbkScore, tc.statusKonsumen, tc.gradeBranch)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedLTV, ltv)
				require.Equal(t, tc.expectedAdjust, adjustTenor)
			}

			mockRepository.AssertExpectations(t)
		})
	}
}

func TestMDMGetDetailCustomerKPM(t *testing.T) {
	os.Setenv("MDM_GET_DETAIL_CUSTOMER_KPM_URL", "https://dev-core-masterdata-customer-api.kbfinansia.com/api/v1/customer/kpm")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		prospectID       string
		KPMID            int
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.MDMGetDetailCustomerKPMResponse
		expectedError    error
	}{
		{
			name:             "success",
			prospectID:       "SAL-123",
			KPMID:            6287,
			resEngineAPICode: 200,
			resEngineAPIBody: `{
				"code": "OK",
				"message": "operasi berhasil dieksekusi.",
				"data": {
					"customer_kpm": {
						"id": 76203,
						"kmp_id": 6287,
						"customer_id": 5779373,
						"full_name": "Hendri Christianto",
						"mobile_phone": "08111244076",
						"migrated_at": "2023-02-07 10:12:52"
					},
					"customer": {
						"id": 5779373,
						"customer_id_confins": "42600089387",
						"id_type": "KTP",
						"id_number": "3173012404900011",
						"expired_date": "",
						"legal_name": "HENDRI CHRISTIANTO",
						"full_name": "Hendri Christianto",
						"gender": "M",
						"birth_date": "1990-05-24",
						"birth_place": "jakarta",
						"surgate_mother_name": "IBU",
						"personal_customer_type": "N",
						"personal_npwp_number": "000000000000000",
						"mobile_phone": "08111244076",
						"email": "2flv35ka@example.com",
						"religion": "3",
						"marital_status": "S",
						"num_of_dependence": 0,
						"no_kk": "",
						"education": "S1",
						"nationality": "",
						"wna_country": "",
						"home_status": "KL",
						"rent_finish_date": "",
						"home_location": "",
						"home_price": null,
						"stay_since_month": 4,
						"stay_since_year": 2023,
						"num_of_asset_owned": 0,
						"living_cost_amount": 5000000,
						"bank_id": "",
						"account_no": "567857855",
						"account_name": "TEST NAMA",
						"profession_id": "WRST",
						"industry_type_id": "9950",
						"main_business_since_year": 2023,
						"job_position": "W",
						"job_type": "0141",
						"employment_since_year": 2023,
						"monthly_fixed_income": 4000000,
						"monthly_variable_income": 5000000,
						"spouse_income": 500000,
						"company_name": "Pt Finansia",
						"employment_since_month": 4,
						"employee_status_id": "",
						"source_variable_income_id": "",
						"bank_branch": "",
						"job_title": "",
						"others_profession_id": "",
						"customer_address": {
							"residence_address": "Jalan Testing",
							"residence_rt": "004",
							"residence_rw": "001",
							"residence_kelurahan": "BINONG",
							"residence_kecamatan": "BATUNUNGGAL",
							"residence_city": "KOTA BANDUNG",
							"residence_zip_code": "40275",
							"residence_area_phone": "0000",
							"residence_phone": "00000",
							"legal_address": "Jl. Bandung",
							"legal_rt": "005",
							"legal_rw": "004",
							"legal_kelurahan": "Sumber Makmur",
							"legal_kecamatan": "Lubuk Pinang",
							"legal_city": "Kab. Mukomuko",
							"legal_zip_code": "38767",
							"legal_area_phone": "0000",
							"legal_phone": "00000",
							"company_address": "Jalan Jalan",
							"company_rt": "004",
							"company_rw": "005",
							"company_kelurahan": "Pasar Minggu",
							"company_kecamatan": "Pasar Minggu",
							"company_city": "Kota Jakarta Selatan",
							"company_zip_code": "12520",
							"company_area_phone": "021",
							"company_phone": "55556666",
							"emergency_contact_address": "Jalan Jalan No.08 (Mesjid)",
							"emergency_contact_rt": "004",
							"emergency_contact_rw": "001",
							"emergency_contact_kelurahan": "Pesanggrahan",
							"emergency_contact_kecamatan": "Pesanggrahan",
							"emergency_contact_city": "Kota Jakarta Selatan",
							"emergency_contact_zip_code": "12320",
							"emergency_contact_home_phone_area": "0000",
							"emergency_contact_home_phone": "00000",
							"emergency_contact_office_phone_area": "",
							"emergency_contact_office_phone": "",
							"legal_province": "",
							"residence_province": "Dki Jakarta",
							"company_province": "Dki Jakarta",
							"emergency_contact_province": ""
						},
						"customer_emcon": null,
						"customer_omset": null,
						"customer_photo": null,
						"customer_spouse": null,
						"customer_kmp_segment": null
					}
				},
				"errors": null,
				"request_id": "d97a4c41468a0bc7289c236a7a40040c",
				"timestamp": "2025-06-17 00:57:37"
			}`,
			expectedResponse: response.MDMGetDetailCustomerKPMResponse{
				Code:    "OK",
				Message: "operasi berhasil dieksekusi.",
				Data: response.GetCustomerByKpmID{
					CustomerKpm: response.CustomerKpm{
						KpmId:       6287,
						CustomerId:  5779373,
						FullName:    "Hendri Christianto",
						MobilePhone: "08111244076",
						MigratedAt:  "2023-02-07 10:12:52",
					},
					Customer: response.Customer{
						Id:                     5779373,
						CustomerIdConfins:      "42600089387",
						IdType:                 "KTP",
						IdNumber:               "3173012404900011",
						ExpiredDate:            "",
						LegalName:              "HENDRI CHRISTIANTO",
						FullName:               "Hendri Christianto",
						Gender:                 "M",
						BirthDate:              "1990-05-24",
						BirthPlace:             "jakarta",
						SurgateMotherName:      "IBU",
						PersonalCustomerType:   "N",
						PersonalNpwpNumber:     "000000000000000",
						MobilePhone:            "08111244076",
						Email:                  "2flv35ka@example.com",
						Religion:               "3",
						MaritalStatus:          "S",
						NumOfDependence:        0,
						NoKk:                   "",
						Education:              "S1",
						Nationality:            "",
						WnaCountry:             "",
						HomeStatus:             "KL",
						RentFinishDate:         "",
						HomeLocation:           "",
						HomePrice:              nil,
						StaySinceMonth:         intPtr(4),
						StaySinceYear:          intPtr(2023),
						NumOfAssetOwned:        0,
						LivingCostAmount:       intPtr(5000000),
						BankId:                 "",
						AccountNo:              "567857855",
						AccountName:            "TEST NAMA",
						ProfessionId:           "WRST",
						IndustryTypeId:         "9950",
						MainBusinessSinceYear:  float64(2023),
						JobPosition:            "W",
						JobType:                "0141",
						EmploymentSinceYear:    intPtr(2023),
						MonthlyFixedIncome:     float64Ptr(4000000),
						MonthlyVariableIncome:  float64Ptr(5000000),
						SpouseIncome:           float64Ptr(500000),
						CompanyName:            "Pt Finansia",
						EmploymentSinceMonth:   intPtr(4),
						EmployeeStatusId:       "",
						SourceVariableIncomeId: "",
						BankBranch:             "",
						JobTitle:               "",
						OthersProfessionId:     "",
						CustomerAddress: response.CustomerKpmAddress{
							ResidenceAddress:                "Jalan Testing",
							ResidenceRt:                     "004",
							ResidenceRw:                     "001",
							ResidenceKelurahan:              "BINONG",
							ResidenceKecamatan:              "BATUNUNGGAL",
							ResidenceCity:                   "KOTA BANDUNG",
							ResidenceZipCode:                "40275",
							ResidenceAreaPhone:              "0000",
							ResidencePhone:                  "00000",
							LegalAddress:                    "Jl. Bandung",
							LegalRt:                         "005",
							LegalRw:                         "004",
							LegalKelurahan:                  "Sumber Makmur",
							LegalKecamatan:                  "Lubuk Pinang",
							LegalCity:                       "Kab. Mukomuko",
							LegalZipCode:                    "38767",
							LegalAreaPhone:                  "0000",
							LegalPhone:                      "00000",
							CompanyAddress:                  "Jalan Jalan",
							CompanyRt:                       "004",
							CompanyRw:                       "005",
							CompanyKelurahan:                "Pasar Minggu",
							CompanyKecamatan:                "Pasar Minggu",
							CompanyCity:                     "Kota Jakarta Selatan",
							CompanyZipCode:                  "12520",
							CompanyAreaPhone:                "021",
							CompanyPhone:                    "55556666",
							EmergencyContactAddress:         "Jalan Jalan No.08 (Mesjid)",
							EmergencyContactRt:              "004",
							EmergencyContactRw:              "001",
							EmergencyContactKelurahan:       "Pesanggrahan",
							EmergencyContactKecamatan:       "Pesanggrahan",
							EmergencyContactCity:            "Kota Jakarta Selatan",
							EmergencyContactZipCode:         "12320",
							EmergencyContactHomePhoneArea:   "0000",
							EmergencyContactHomePhone:       "00000",
							EmergencyContactOfficePhoneArea: "",
							EmergencyContactOfficePhone:     "",
							LegalProvince:                   "",
							ResidenceProvince:               "Dki Jakarta",
							CompanyProvince:                 "Dki Jakarta",
							EmergencyContactProvince:        "",
						},
						CustomerEmcon:   nil,
						CustomerOmset:   nil,
						CustomerPhoto:   nil,
						CustomerSpouse:  nil,
						CustomerSegment: nil,
					},
				},
				Errors:    nil,
				RequestID: "d97a4c41468a0bc7289c236a7a40040c",
				Timestamp: "2025-06-17 00:57:37",
			},
			expectedError: nil,
		},
		{
			name:          "error api response",
			prospectID:    "SAL-123",
			KPMID:         6287,
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New("network error"),
		},
		{
			name:             "not 200 status code",
			prospectID:       "SAL-123",
			KPMID:            6287,
			resEngineAPICode: 400,
			resEngineAPIBody: `{
				"code": "BAD_REQUEST",
				"message": "Invalid request",
				"errors": {
					"kmp_id": "kmp_id is required"
				}
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - MDM Get Detail Customer KPM Error"),
		},
		{
			name:             "error unmarshal",
			prospectID:       "SAL-123",
			KPMID:            6287,
			resEngineAPICode: 200,
			resEngineAPIBody: `invalid-json`,
			expectedError:    errors.New("unexpected end of JSON input"),
		},
		{
			name:             "empty response body",
			prospectID:       "SAL-123",
			KPMID:            6287,
			resEngineAPICode: 200,
			resEngineAPIBody: `{}`,
			expectedResponse: response.MDMGetDetailCustomerKPMResponse{},
			expectedError:    nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("MDM_GET_DETAIL_CUSTOMER_KPM_URL") + "/" + strconv.Itoa(tc.KPMID)
			httpmock.RegisterResponder(constant.METHOD_GET, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Get(url)

			headerMDM := map[string]string{
				"Content-Type":  "application/json",
				"Authorization": accessToken,
			}

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, mock.Anything, headerMDM, constant.METHOD_GET, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MDMGetDetailCustomerKPM(ctx, tc.prospectID, tc.KPMID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}

// Helper functions for pointer values
func intPtr(i int) *int {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}
