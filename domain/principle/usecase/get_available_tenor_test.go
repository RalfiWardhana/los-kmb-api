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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAvailableTenor(t *testing.T) {
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
	referralCode := "test"
	os.Setenv("MI_NUMBER_WHITELIST", "123")
	accessToken := "test-token"

	testCases := []struct {
		name                      string
		request                   request.GetAvailableTenor
		config                    entity.AppConfig
		errConfig                 error
		trxKPM                    entity.TrxKPM
		errTrxKPM                 error
		dupcheckResponse          response.SpDupCekCustomerByID
		errDupcheck               error
		assetResponse             response.AssetList
		errAsset                  error
		cmoResponse               response.MDMMasterMappingBranchEmployeeResponse
		errCmo                    error
		hrisResponse              response.EmployeeCMOResponse
		errHris                   error
		mappingFpdCluster         entity.MasterMappingFpdCluster
		errFpdCluster             error
		plateResponse             response.MDMMasterMappingLicensePlateResponse
		errPlate                  error
		marsevResponse            response.MarsevFilterProgramResponse
		errMarsev                 error
		assetYearListResponse     response.AssetYearList
		errAssetYearList          error
		getFpdCmoResponse         response.FpdCMOResponse
		errGetFpdCmo              error
		savedClusterCheckCmoNoFPD string
		entitySaveTrxNoFPd        entity.TrxCmoNoFPD
		errCheckCmoNoFPD          error
		calculationResponse       response.MarsevCalculateInstallmentResponse
		errCalculation            error
		mappingLTV                []entity.MappingElaborateLTV
		errMappingLTV             error
		ltvResponse               int
		adjustTenorResponse       bool
		errGetLTV                 error
		loanAmountResponse        response.MarsevLoanAmountResponse
		errLoanAmount             error
		rejectTenorResponse       response.UsecaseApi
		errRejectTenor            error
		expectedTenors            []response.GetAvailableTenorData
		expectedError             error
		mappingBranch             entity.MappingBranchByPBKScore
		errMappingBranchEntity    error
		trxDetailBiro             []entity.TrxDetailBiro
		pbkScore                  string
		errTrxDetailBiro          error
		mappingPbkScoreGrade      entity.MappingPBKScoreGrade
		errMappingPbkScoreGrade   error
	}{
		{
			name: "success case - simulation",
			request: request.GetAvailableTenor{
				ProspectID:      "SIM-123",
				BranchID:        "123",
				AssetCode:       "MOT",
				LicensePlate:    "B1234XX",
				BPKBNameType:    "K",
				ManufactureYear: "2020",
				LoanAmount:      50000000,
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			pbkScore: "GOOD",
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "AVERAGE RISK",
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						MonthlyInstallment: 5000000,
						AmountOfFinance:    45000000,
						AdminFee:           1000000,
						DPAmount:           5000000,
						NTF:                50000000,
						IsPSA:              true,
					},
					{
						Tenor:              24,
						MonthlyInstallment: 3000000,
						AmountOfFinance:    45000000,
						AdminFee:           1000000,
						DPAmount:           5000000,
						NTF:                50000000,
						IsPSA:              true,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
				{
					LTV: 80,
				},
			},
			ltvResponse: 80,
			loanAmountResponse: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 60000000,
					IsPsa:             true,
				},
			},
			expectedTenors: []response.GetAvailableTenorData{
				{
					Tenor:             12,
					IsPsa:             true,
					Dealer:            "PSA",
					InstallmentAmount: 5000000,
					AF:                45000000,
					AdminFee:          1000000,
					DPAmount:          5000000,
					NTF:               50000000,
					AssetCategoryID:   "CAT1",
					OTR:               60000000,
				},
				{
					Tenor:             24,
					IsPsa:             true,
					Dealer:            "PSA",
					InstallmentAmount: 3000000,
					AF:                45000000,
					AdminFee:          1000000,
					DPAmount:          5000000,
					NTF:               50000000,
					AssetCategoryID:   "CAT1",
					OTR:               60000000,
				},
			},
		},
		{
			name: "success case with referral cde - simulation",
			request: request.GetAvailableTenor{
				ProspectID:      "SIM-123",
				BranchID:        "123",
				AssetCode:       "MOT",
				LicensePlate:    "B1234XX",
				BPKBNameType:    "K",
				ManufactureYear: "2020",
				LoanAmount:      50000000,
				ReferralCode:    &referralCode,
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID:       "1234",
						MINumber: 1,
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
					{
						ID:       "12345",
						MINumber: 123,
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
			assetYearListResponse: response.AssetYearList{
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
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						MonthlyInstallment: 5000000,
						AmountOfFinance:    45000000,
						AdminFee:           1000000,
						DPAmount:           5000000,
						NTF:                50000000,
						IsPSA:              true,
					},
					{
						Tenor:              24,
						MonthlyInstallment: 3000000,
						AmountOfFinance:    45000000,
						AdminFee:           1000000,
						DPAmount:           5000000,
						NTF:                50000000,
						IsPSA:              true,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
				{
					LTV: 80,
				},
			},
			ltvResponse: 80,
			loanAmountResponse: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 60000000,
					IsPsa:             true,
				},
			},
			expectedTenors: []response.GetAvailableTenorData{
				{
					Tenor:             12,
					IsPsa:             true,
					Dealer:            "PSA",
					InstallmentAmount: 5000000,
					AF:                45000000,
					AdminFee:          1000000,
					DPAmount:          5000000,
					NTF:               50000000,
					AssetCategoryID:   "CAT1",
					OTR:               60000000,
				},
				{
					Tenor:             24,
					IsPsa:             true,
					Dealer:            "PSA",
					InstallmentAmount: 3000000,
					AF:                45000000,
					AdminFee:          1000000,
					DPAmount:          5000000,
					NTF:               50000000,
					AssetCategoryID:   "CAT1",
					OTR:               60000000,
				},
			},
		},

		{
			name: "error case with referral cde - simulation",
			request: request.GetAvailableTenor{
				ProspectID:      "SIM-123",
				BranchID:        "123",
				AssetCode:       "MOT",
				LicensePlate:    "B1234XX",
				BPKBNameType:    "K",
				ManufactureYear: "2020",
				LoanAmount:      50000000,
				ReferralCode:    &referralCode,
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						ID:       "1234",
						MINumber: 1,
						Tenors: []response.TenorInfo{
							{
								Tenor: 12,
							},
							{
								Tenor: 24,
							},
						},
					},
					{
						ID:       "12345",
						MINumber: 1223,
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
			expectedError: errors.New(constant.ERROR_BAD_REQUEST + " - No matching MI_NUMBER found"),
		},
		{
			name: "error get config",
			request: request.GetAvailableTenor{
				ProspectID: "SIM-123",
			},
			errConfig:     errors.New("config error"),
			expectedError: errors.New("config error"),
		},
		{
			name: "error get trx kpm",
			request: request.GetAvailableTenor{
				ProspectID: "REG-123",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			errTrxKPM:     errors.New("trx kpm error"),
			expectedError: errors.New("trx kpm error"),
		},
		{
			name: "error dupcheck",
			request: request.GetAvailableTenor{
				ProspectID: "REG-123",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			errDupcheck:   errors.New("dupcheck error"),
			expectedError: errors.New("upstream_service_error - Get Data Customer Error"),
		},
		{
			name: "error get master asset",
			request: request.GetAvailableTenor{
				ProspectID:      "REG-123",
				BranchID:        "123",
				AssetCode:       "MOT",
				LicensePlate:    "B1234XX",
				BPKBNameType:    "K",
				ManufactureYear: "2020",
				LoanAmount:      50000000,
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
			},
			errAsset:      errors.New("failed to get master asset"),
			expectedError: errors.New("failed to get master asset"),
		},
		{
			name: "error get marketing program",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			errMarsev:     errors.New("failed to get marketing program"),
			expectedError: errors.New("failed to get marketing program"),
		},
		{
			name: "error get asset year list",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			errAssetYearList: errors.New("failed to get asset year list"),
			expectedError:    errors.New("failed to get asset year list"),
		},
		{
			name: "error get mapping branch employee",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "CMO123",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: "NEW",
			},
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			errCmo:        errors.New("failed to get mapping branch employee"),
			expectedError: errors.New("failed to get mapping branch employee"),
		},
		{
			name: "error get cmo dedicated not found",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{},
			},
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - CMO Dedicated Not Found"),
		},
		{
			name: "error get employee data",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
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
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse:  response.EmployeeCMOResponse{},
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - CMO Not Found"),
		},
		{
			name: "error get fpd cmo",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			errGetFpdCmo:  errors.New("failed to get fpd cmo"),
			expectedError: errors.New("failed to get fpd cmo"),
		},
		{
			name: "error get master mapping fpd cluster",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			errFpdCluster: errors.New("failed to get master mapping fpd cluster"),
			expectedError: errors.New("failed to get master mapping fpd cluster"),
		},
		{
			name: "error check cmo no fpd",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			mappingFpdCluster: entity.MasterMappingFpdCluster{
				Cluster: "",
			},
			errCheckCmoNoFPD: errors.New("failed to check cmo no fpd"),
			expectedError:    errors.New("failed to check cmo no fpd"),
		},
		{
			name: "error get mapping license plate",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: true,
			},
			mappingFpdCluster: entity.MasterMappingFpdCluster{
				Cluster: "Cluster C",
			},
			errPlate:      errors.New("failed to get mapping license plate"),
			expectedError: errors.New("failed to get mapping license plate"),
		},
		{
			name: "error marsev calculate installment",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			errCalculation:            errors.New("failed to calculate installment"),
		},
		{
			name: "error marsev empty installment data",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse:       response.MarsevCalculateInstallmentResponse{},
		},
		{
			name: "error get trx detail biro",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			errTrxDetailBiro: errors.New(constant.ERROR_UPSTREAM + " - Get Trx Detail Biro Error"),
			expectedError:    errors.New(constant.ERROR_UPSTREAM + " - Get Trx Detail Biro Error"),
		},
		{
			name: "error get mapping pbk score",
			request: request.GetAvailableTenor{
				ProspectID:         "SSREG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
			},
			config: entity.AppConfig{
				Value: "^SIM-.*",
			},
			dupcheckResponse: response.SpDupCekCustomerByID{
				CustomerStatus:  constant.STATUS_KONSUMEN_RO_AO,
				CustomerSegment: constant.RO_AO_REGULAR,
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score: "Low Risk",
				},
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			errMappingPbkScoreGrade: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Pbk Score Error"),
			expectedError:           errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Pbk Score Error"),
		},
		{
			name: "error get mapping branch",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			errMappingBranchEntity: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Branch Error"),
			expectedError:          errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Branch Error"),
		},
		{
			name: "error get mapping elaborate ltv",
			request: request.GetAvailableTenor{
				ProspectID:         "REG-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			errMappingLTV: errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error"),
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error"),
		},
		{
			name: "error get ltv",
			request: request.GetAvailableTenor{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			errGetLTV:     errors.New("failed to get ltv"),
			expectedError: errors.New("failed to get ltv"),
		},
		{
			name: "error get ltv 0",
			request: request.GetAvailableTenor{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			ltvResponse: 0,
		},
		{
			name: "error marsev get loan amount",
			request: request.GetAvailableTenor{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			ltvResponse:   80,
			errLoanAmount: errors.New("failed to get loan amount"),
			expectedError: errors.New("failed to get loan amount"),
		},
		{
			name: "error reject tenor 36",
			request: request.GetAvailableTenor{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
			errRejectTenor: errors.New("failed to get reject tenor"),
			expectedError:  errors.New("failed to get reject tenor"),
		},
		{
			name: "error tenor > 36",
			request: request.GetAvailableTenor{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
				{
					ID: 1,
				},
			},
		},
		{
			name: "error tenor 36 get result reject",
			request: request.GetAvailableTenor{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				LicensePlate:       "B1234XX",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
				LoanAmount:         50000000,
				AssetUsageTypeCode: "P",
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
			plateResponse: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "AREA1",
						},
					},
				},
			},
			marsevResponse: response.MarsevFilterProgramResponse{
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
			assetYearListResponse: response.AssetYearList{
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
			cmoResponse: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID: "12434",
					},
				},
			},
			hrisResponse: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			getFpdCmoResponse: response.FpdCMOResponse{
				FpdExist: false,
			},
			savedClusterCheckCmoNoFPD: "Cluster C",
			calculationResponse: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						Tenor:              12,
						IsPSA:              true,
						MonthlyInstallment: 1000000,
						AmountOfFinance:    1000000,
						AdminFee:           100000,
						DPAmount:           1000000,
						NTF:                100000,
					},
				},
			},
			mappingLTV: []entity.MappingElaborateLTV{
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
			mockHttpClient := new(httpclient.MockHttpClient)
			mockUsecase := new(mocks.Usecase)

			mockRepository.On("GetConfig", constant.GROUP_2WILEN, "KMB-OFF", constant.KEY_PPID_SIMULASI).Return(tc.config, tc.errConfig)

			if tc.errConfig == nil {
				re := regexp.MustCompile(tc.config.Value)
				isSimulasi := re.MatchString(tc.request.ProspectID)
				mockRepository.On("GetTrxKPM", tc.request.ProspectID).Return(tc.trxKPM, tc.errTrxKPM)

				if tc.errTrxKPM == nil {
					mockUsecase.On("DupcheckIntegrator", ctx, tc.request.ProspectID, tc.request.IDNumber, tc.request.LegalName, tc.request.BirthDate, tc.request.SurgateMotherName, accessToken).Return(tc.dupcheckResponse, tc.errDupcheck)

					if tc.errDupcheck == nil {
						mockUsecase.On("MDMGetMasterAsset", ctx, tc.request.BranchID, tc.request.AssetCode, tc.request.ProspectID, accessToken).Return(tc.assetResponse, tc.errAsset)

						if tc.errAsset == nil {
							mockUsecase.On("MarsevGetMarketingProgram", ctx, mock.MatchedBy(func(req request.ReqMarsevFilterProgram) bool {
								return req.BranchID == tc.request.BranchID &&
									req.AssetCategory == "CAT1" &&
									req.AssetUsageTypeCode == tc.request.AssetUsageTypeCode
							}), tc.request.ProspectID, accessToken).Return(tc.marsevResponse, tc.errMarsev)

							if tc.errMarsev == nil {
								mockUsecase.On("MDMGetAssetYear", ctx, tc.request.BranchID, tc.request.AssetCode, tc.request.ManufactureYear, tc.request.ProspectID, accessToken).Return(tc.assetYearListResponse, tc.errAssetYearList)

								if tc.errAssetYearList == nil {
									mockUsecase.On("MDMGetMasterMappingBranchEmployee", ctx, tc.request.ProspectID, tc.request.BranchID, accessToken).Return(tc.cmoResponse, tc.errCmo)

									if tc.errCmo == nil {
										mockUsecase.On("GetEmployeeData", ctx, mock.Anything).Return(tc.hrisResponse, tc.errHris)

										if tc.errHris == nil {
											if tc.hrisResponse.CMOCategory != constant.NEW {
												mockUsecase.On("GetFpdCMO", ctx, mock.Anything, mock.Anything).Return(tc.getFpdCmoResponse, tc.errGetFpdCmo)
											}

											if tc.getFpdCmoResponse.FpdExist {
												mockRepository.On("MasterMappingFpdCluster", mock.Anything).Return(tc.mappingFpdCluster, tc.errFpdCluster)
											}

											if !tc.getFpdCmoResponse.FpdExist || (tc.getFpdCmoResponse.FpdExist && tc.mappingFpdCluster.Cluster == "") {
												mockUsecase.On("CheckCmoNoFPD", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.savedClusterCheckCmoNoFPD, tc.entitySaveTrxNoFPd, tc.errCheckCmoNoFPD)
											}

											if tc.errCheckCmoNoFPD == nil {

												mockUsecase.On("MDMGetMappingLicensePlate", ctx, tc.request.LicensePlate, tc.request.ProspectID, accessToken).Return(tc.plateResponse, tc.errPlate)

												mockUsecase.On("MarsevCalculateInstallment", ctx, mock.Anything, tc.request.ProspectID, accessToken).Return(tc.calculationResponse, tc.errCalculation)
												mockRepository.On("GetTrxDetailBIro", tc.request.ProspectID).Return(tc.trxDetailBiro, tc.errTrxDetailBiro)

												if tc.errTrxDetailBiro == nil {
													if !isSimulasi {
														mockRepository.On("GetMappingPbkScore", mock.Anything).Return(tc.mappingPbkScoreGrade, tc.errMappingPbkScoreGrade)
													}

													if tc.errMappingPbkScoreGrade == nil {
														mockRepository.On("GetMappingBranchByBranchID", tc.request.BranchID, mock.Anything).
															Return(tc.mappingBranch, tc.errMappingBranchEntity)

														if tc.errMappingBranchEntity == nil {
															mockRepository.On("GetMappingElaborateLTV", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.mappingLTV, tc.errMappingLTV)
															mockUsecase.On("MarsevGetLoanAmount", ctx, mock.Anything, tc.request.ProspectID, accessToken).Return(tc.loanAmountResponse, tc.errLoanAmount)
															mockUsecase.On("GetLTV", ctx, tc.mappingLTV, tc.request.ProspectID, "PASS", tc.request.BPKBNameType, tc.request.ManufactureYear, mock.AnythingOfType("int"), float64(0), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.ltvResponse, tc.adjustTenorResponse, tc.errGetLTV)

															if tc.marsevResponse.Data[0].Tenors[0].Tenor == 36 {
																mockUsecase.On("RejectTenor36", mock.Anything).Return(tc.rejectTenorResponse, tc.errRejectTenor)
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

			result, err := multiUsecase.GetAvailableTenor(ctx, tc.request, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedTenors, result)
			}
		})
	}
}

func TestMDMGetMappingLicensePlate(t *testing.T) {
	os.Setenv("MDM_MASTER_MAPPING_LICENSE_PLATE_URL", "https://dev-core-masterdata-area-api.kbfinansia.com/api/v1/master-data/area/mapping/license-plate")
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		prospectID       string
		licensePlate     string
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.MDMMasterMappingLicensePlateResponse
		expectedError    error
	}{
		{
			name:             "success",
			prospectID:       "SAL-123",
			licensePlate:     "B1234CD",
			resEngineAPICode: 200,
			resEngineAPIBody: `{
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
			expectedResponse: response.MDMMasterMappingLicensePlateResponse{
				Code:    "200",
				Message: "Success",
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							PlateAreaID:     1,
							PlateID:         101,
							PlateCode:       "B",
							AreaID:          "426",
							AreaDescription: "JAKARTA PUSAT",
							LobID:           1,
							CreatedAt:       "2024-01-01T00:00:00Z",
							CreatedBy:       "system",
							UpdatedAt:       strPtr("2024-01-02T00:00:00Z"),
							UpdatedBy:       strPtr("admin"),
							DeletedAt:       nil,
							DeletedBy:       nil,
						},
					},
					MaxPage:     1,
					Total:       1,
					PageSize:    10,
					CurrentPage: 1,
				},
				Errors:    nil,
				RequestID: "req-123",
				Timestamp: "2024-01-01T00:00:00Z",
			},
			expectedError: nil,
		},
		{
			name:          "error api response",
			prospectID:    "SAL-123",
			licensePlate:  "B1234CD",
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New("network error"),
		},
		{
			name:             "not 200 status code",
			prospectID:       "SAL-123",
			licensePlate:     "B1234CD",
			resEngineAPICode: 400,
			resEngineAPIBody: `{
				"code": "400",
				"message": "Bad Request",
				"data": null,
				"errors": {"field": "plate_code", "message": "Invalid plate code"},
				"request_id": "req-123",
				"timestamp": "2024-01-01T00:00:00Z"
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping License Plate Error"),
		},
		{
			name:             "invalid json response",
			prospectID:       "SAL-123",
			licensePlate:     "B1234CD",
			resEngineAPICode: 200,
			resEngineAPIBody: `invalid json`,
			expectedError:    errors.New("unexpected end of JSON input"),
		},
		{
			name:             "empty records",
			prospectID:       "SAL-123",
			licensePlate:     "B1234CD",
			resEngineAPICode: 200,
			resEngineAPIBody: `{
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
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping License Plate Error Not Found Data"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			licensePlateCode := utils.GetLicensePlateCode(tc.licensePlate)
			url := os.Getenv("MDM_MASTER_MAPPING_LICENSE_PLATE_URL") + "?lob_id=" + strconv.Itoa(constant.LOBID_KMB) + "&plate_code=" + licensePlateCode
			httpmock.RegisterResponder(constant.METHOD_GET, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Get(url)

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, []byte(nil), map[string]string{
				"Content-Type":  "application/json",
				"Authorization": accessToken,
			}, constant.METHOD_GET, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MDMGetMappingLicensePlate(ctx, tc.licensePlate, tc.prospectID, accessToken)

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

func TestMarsevCalculateInstallment(t *testing.T) {
	os.Setenv("MARSEV_CALCULATE_INSTALLMENT_URL", "https://dev-marsev-api.kbfinansia.com/api/v1/calculate-installment")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("MARSEV_AUTHORIZATION_KEY", "marsev-test-key")
	accessToken := "test-token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	testcases := []struct {
		name             string
		prospectID       string
		request          request.ReqMarsevCalculateInstallment
		errEngineAPI     error
		resEngineAPICode int
		resEngineAPIBody string
		expectedResponse response.MarsevCalculateInstallmentResponse
		expectedError    error
	}{
		{
			name:       "success",
			prospectID: "SAL-123",
			request: request.ReqMarsevCalculateInstallment{
				ProgramID:              "PROG001",
				BranchID:               "BR001",
				CustomerOccupationCode: "EMP",
				AssetUsageTypeCode:     "PERSONAL",
				AssetYear:              2023,
				BpkbStatusCode:         "OWN",
				LoanAmount:             100000000,
				Otr:                    120000000,
				RegionCode:             "JKT",
				AssetCategory:          "CAR",
				CustomerBirthDate:      "1990-01-01",
				Tenor:                  36,
			},
			resEngineAPICode: 200,
			resEngineAPIBody: `{
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
			expectedResponse: response.MarsevCalculateInstallmentResponse{
				Code:    200,
				Message: "Success",
				Data: []response.MarsevCalculateInstallmentData{
					{
						InstallmentTypeCode:        "ADDB",
						IsPSA:                      false,
						Tenor:                      36,
						AdminFee:                   2500000,
						AdminFeePSA:                0,
						ProvisionFee:               1200000,
						AmountOfFinance:            100000000,
						DPAmount:                   20000000,
						DPPercent:                  20,
						AdditionalRate:             0.5,
						EffectiveRate:              12.5,
						LifeInsurance:              1500000,
						AssetInsurance:             2000000,
						TotalInsurance:             3500000,
						FiduciaFee:                 500000,
						NTF:                        107700000,
						MonthlyInstallment:         3500000,
						MonthlyInstallmentMin:      3400000,
						MonthlyInstallmentMax:      3600000,
						TotalLoan:                  126000000,
						AmountOfInterest:           26000000,
						FlatRateYearlyPercent:      6.5,
						FlatRateMonthlyPercent:     0.54,
						ProductID:                  "PRD001",
						ProductOfferingID:          "OFF001",
						ProductOfferingDescription: "Standard Car Loan",
						SubsidyAmountScheme:        0,
						FineAmount:                 175000,
						FineAmountFormula:          "5% * installment",
						FineAmountDetail:           "5% penalty from monthly installment",
						NTFFormula:                 "loan_amount + total_insurance + admin_fee",
						NTFDetail:                  "Total loan calculation details",
						AmountOfInterestFormula:    "flat_rate * tenor * loan_amount",
						AmountOfInterestDetail:     "Interest calculation details",
						WanprestasiFreightFee:      1000000,
						ExternalFreightFee:         500000,
						WanprestasiFreightFormula:  "fixed_amount",
						WanprestasiFreightDetail:   "Fixed penalty amount",
						ExternalFreightFormula:     "fixed_amount",
						ExternalFreightDetail:      "Fixed external fee",
						IsStampDutyAsLoan:          boolPtr(false),
						StampDutyFee:               10000,
					},
				},
				Errors: nil,
			},
			expectedError: nil,
		},
		{
			name:       "error api response",
			prospectID: "SAL-123",
			request: request.ReqMarsevCalculateInstallment{
				ProgramID:  "PROG001",
				LoanAmount: 100000000,
				Tenor:      36,
			},
			errEngineAPI:  errors.New("network error"),
			expectedError: errors.New("network error"),
		},
		{
			name:       "not 200 status code",
			prospectID: "SAL-123",
			request: request.ReqMarsevCalculateInstallment{
				ProgramID:  "PROG001",
				LoanAmount: 100000000,
				Tenor:      36,
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
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Marsev Calculate Installment Error"),
		},
		{
			name:       "invalid json response",
			prospectID: "SAL-123",
			request: request.ReqMarsevCalculateInstallment{
				ProgramID:  "PROG001",
				LoanAmount: 100000000,
				Tenor:      36,
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

			url := os.Getenv("MARSEV_CALCULATE_INSTALLMENT_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resEngineAPICode, tc.resEngineAPIBody))
			resp, _ := rst.R().Post(url)

			param, _ := json.Marshal(tc.request)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, param, map[string]string{
				"Content-Type":  "application/json",
				"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
			}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.prospectID, accessToken).Return(resp, tc.errEngineAPI)

			usecase := NewUsecase(nil, mockHttpClient, nil)

			result, err := usecase.MarsevCalculateInstallment(ctx, tc.request, tc.prospectID, accessToken)

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

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}
