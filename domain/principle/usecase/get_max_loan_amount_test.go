package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"regexp"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMaxLoanAmount(t *testing.T) {
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
	accessToken := "test-token"

	testCases := []struct {
		name                      string
		request                   request.GetMaxLoanAmount
		config                    entity.AppConfig
		errConfig                 error
		trxKPM                    entity.TrxKPM
		errTrxKPM                 error
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
		getLTVResponse            int
		adjustTenorResponse       bool
		errGetLTV                 error
		marsevLoanAmountRes       response.MarsevLoanAmountResponse
		errMarsevLoanAmount       error
		expectedResponse          response.GetMaxLoanAmountData
		expectedError             error
	}{
		{
			name: "success simulation case",
			request: request.GetMaxLoanAmount{
				ProspectID:         "SIM-123",
				BranchID:           "123",
				AssetCode:          "MOT",
				IDNumber:           "1234567890",
				LegalName:          "Test User",
				BirthDate:          "1990-01-01",
				SurgateMotherName:  "Mother Name",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
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
			name: "error get config",
			request: request.GetMaxLoanAmount{
				ProspectID: "SIM-123",
			},
			errConfig:     errors.New("config error"),
			expectedError: errors.New("config error"),
		},
		{
			name: "error get trx kpm",
			request: request.GetMaxLoanAmount{
				ProspectID: "123",
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
				IDNumber:           "1234567890",
				LegalName:          "Test User",
				BirthDate:          "1990-01-01",
				SurgateMotherName:  "Mother Name",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
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
				IDNumber:           "1234567890",
				LegalName:          "Test User",
				BirthDate:          "1990-01-01",
				SurgateMotherName:  "Mother Name",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
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
				IDNumber:           "1234567890",
				LegalName:          "Test User",
				BirthDate:          "1990-01-01",
				SurgateMotherName:  "Mother Name",
				BPKBNameType:       "K",
				ManufactureYear:    "2020",
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockUsecase := new(mocks.Usecase)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetConfig", constant.GROUP_2WILEN, "KMB-OFF", constant.KEY_PPID_SIMULASI).Return(tc.config, tc.errConfig)

			if tc.errConfig == nil {
				re := regexp.MustCompile(tc.config.Value)
				isSimulasi := re.MatchString(tc.request.ProspectID)

				if !isSimulasi {
					mockRepository.On("GetTrxKPM", tc.request.ProspectID).Return(tc.trxKPM, tc.errTrxKPM)
				}

				if tc.errTrxKPM == nil {
					mockUsecase.On("DupcheckIntegrator", ctx, tc.request.ProspectID, tc.request.IDNumber,
						tc.request.LegalName, tc.request.BirthDate, tc.request.SurgateMotherName, accessToken).Return(tc.dupcheckResponse, tc.errDupcheck)

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

												if tc.errFpdCMO == nil && tc.fpdCMOResponse.FpdExist {
													mockRepository.On("MasterMappingFpdCluster", tc.fpdCMOResponse.CmoFpd).Return(tc.mappingFpdCluster, tc.errMappingFpdCluster)
												}
											}

											if tc.hrisResponse.CMOCategory == constant.NEW || (!tc.fpdCMOResponse.FpdExist || tc.mappingFpdCluster.Cluster == "") {
												mockUsecase.On("CheckCmoNoFPD", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.savedClusterCheckCmoNoFPD, tc.entitySaveTrxNoFPd, tc.errCheckCmoNoFPD)
											}

											mockRepository.On("GetMappingElaborateLTV", mock.Anything, mock.Anything).Return(tc.mappingElaborateLTV, tc.errMappingElaborateLTV)

											if tc.errMappingElaborateLTV == nil {
												mockUsecase.On("GetLTV", ctx, tc.mappingElaborateLTV, tc.request.ProspectID, mock.Anything, tc.request.BPKBNameType, tc.request.ManufactureYear, mock.Anything, mock.Anything).Return(tc.getLTVResponse, tc.adjustTenorResponse, tc.errGetLTV)

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

			multiUsecase := NewMultiUsecase(mockRepository, mockHttpClient, nil, mockUsecase)
			result, err := multiUsecase.GetMaxLoanAmout(ctx, tc.request, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResponse, result)
			}

			mockRepository.AssertExpectations(t)
			mockUsecase.AssertExpectations(t)
		})
	}
}
