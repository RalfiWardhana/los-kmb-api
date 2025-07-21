package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	platformEventMockery "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrincipleMarketingProgram(t *testing.T) {
	ctx := context.Background()
	accessToken := "test-token"
	middlewares.UserInfoData.AccessToken = accessToken

	sampleTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("MDM_MASTER_BRANCH_URL", "http://test-mdm/branch/")
	os.Setenv("SALLY_SUBMISSION_2W_PRINCIPLE_URL", "http://test-sally/submit")

	defaultFilteringKMB := entity.FilteringKMB{
		Decision:        "PASS",
		Reason:          "OK",
		CustomerStatus:  interface{}(constant.STATUS_KONSUMEN_NEW),
		CustomerSegment: interface{}(constant.RO_AO_REGULAR),
		IsBlacklist:     0,
		NextProcess:     1,
	}

	testCases := []struct {
		name                string
		prospectID          string
		principleStepOne    entity.TrxPrincipleStepOne
		errPrincipleOne     error
		principleStepTwo    entity.TrxPrincipleStepTwo
		errPrincipleTwo     error
		principleStepThree  entity.TrxPrincipleStepThree
		errPrincipleThree   error
		emergencyContact    entity.TrxPrincipleEmergencyContact
		errEmergencyContact error
		marketingProgram    entity.TrxPrincipleMarketingProgram
		errMarketingProgram error
		filteringKMB        entity.FilteringKMB
		errFilteringKMB     error
		elaborateLTV        entity.MappingElaborateLTV
		errElaborateLTV     error
		detailBiro          []entity.TrxDetailBiro
		errDetailBiro       error
		branchResStatus     int
		branchResBody       string
		branchErr           error
		sallyResStatus      int
		sallyResBody        string
		sallyErr            error
		errPublishEvent     error
		expectedError       error
	}{
		{
			name:       "success",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				BranchID:        "BR001",
				CMOID:           "CMO123",
				CMOName:         "Test CMO",
				LicensePlate:    "B1234CD",
				BPKBName:        "K",
				OwnerAsset:      "Test Owner",
				STNKExpiredDate: sampleTime,
				TaxDate:         sampleTime,
				ManufactureYear: "2023",
				AssetCode:       "MOT",
				Color:           "BLACK",
				CC:              "150",
				NoChassis:       "CHS123",
				NoEngine:        "ENG123",
				STNKPhoto:       "http://test/stnk.jpg",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				KtpPhoto:    "http://test/ktp.jpg",
				SelfiePhoto: "http://test/selfie.jpg",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				InstallmentAmount: 1000000,
				AssetCategoryID:   "CAT1",
				Tenor:             24,
				OTR:               25000000,
				FinancePurpose:    "PRODUCTIVE",
				Dealer:            "NON-PSA",
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				CustomerID: 12345,
			},
			marketingProgram: entity.TrxPrincipleMarketingProgram{
				ProgramID:                  "PRG001",
				ProgramName:                "Test Program",
				ProductOfferingID:          "OFF001",
				ProductOfferingDescription: "Test Offering",
				LoanAmount:                 20000000,
				LoanAmountMaximum:          22000000,
				AdminFee:                   500000,
				ProvisionFee:               200000,
				DPAmount:                   5000000,
				FinanceAmount:              20000000,
			},
			filteringKMB: entity.FilteringKMB{
				Decision:        "PASS",
				Reason:          "OK", // Initialize with non-nil value
				CustomerStatus:  interface{}(constant.STATUS_KONSUMEN_NEW),
				CustomerSegment: interface{}(constant.RO_AO_REGULAR), // Changed from stringPtr
				IsBlacklist:     0,
				NextProcess:     1,
			},
			elaborateLTV: entity.MappingElaborateLTV{
				LTV: 80,
			},
			detailBiro: []entity.TrxDetailBiro{
				{
					Subject:      constant.CUSTOMER,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
			},
			branchResStatus: 200,
			branchResBody: `{
				"data": {
					"branch_name": "Test Branch"
				}
			}`,
			sallyResStatus: 200,
			sallyResBody: `{
				"message": "Success"
			}`,
		},
		{
			name:       "success - psa",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				BranchID:        "BR001",
				CMOID:           "CMO123",
				CMOName:         "Test CMO",
				LicensePlate:    "B1234CD",
				BPKBName:        "P",
				OwnerAsset:      "Test Owner",
				STNKExpiredDate: sampleTime,
				TaxDate:         sampleTime,
				ManufactureYear: "2023",
				AssetCode:       "MOT",
				Color:           "BLACK",
				CC:              "150",
				NoChassis:       "CHS123",
				NoEngine:        "ENG123",
				STNKPhoto:       "http://test/stnk.jpg",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				KtpPhoto:    "http://test/ktp.jpg",
				SelfiePhoto: "http://test/selfie.jpg",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				InstallmentAmount: 1000000,
				AssetCategoryID:   "CAT1",
				Tenor:             24,
				OTR:               25000000,
				FinancePurpose:    "PRODUCTIVE",
				Dealer:            "PSA",
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				CustomerID: 12345,
			},
			marketingProgram: entity.TrxPrincipleMarketingProgram{
				ProgramID:                  "PRG001",
				ProgramName:                "Test Program",
				ProductOfferingID:          "OFF001",
				ProductOfferingDescription: "Test Offering",
				LoanAmount:                 20000000,
				LoanAmountMaximum:          22000000,
				AdminFee:                   500000,
				ProvisionFee:               200000,
				DPAmount:                   5000000,
				FinanceAmount:              20000000,
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        "PASS",
				Reason:                          "OK",
				IsBlacklist:                     0,
				NextProcess:                     1,
				CustomerStatusKMB:               interface{}(constant.STATUS_KONSUMEN_NEW),
				TotalBakiDebetNonCollateralBiro: float64Ptr(1000000),
			},
			elaborateLTV: entity.MappingElaborateLTV{
				LTV: 80,
			},
			detailBiro: []entity.TrxDetailBiro{
				{
					Subject:      constant.CUSTOMER,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
				{
					Subject:      constant.SPOUSE,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
			},
			branchResStatus: 200,
			branchResBody: `{
				"data": {
					"branch_name": "Test Branch"
				}
			}`,
			sallyResStatus: 200,
			sallyResBody: `{
				"message": "Success"
			}`,
		},
		{
			name:       "error get float total baki debet",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				BranchID:        "BR001",
				CMOID:           "CMO123",
				CMOName:         "Test CMO",
				LicensePlate:    "B1234CD",
				BPKBName:        "P",
				OwnerAsset:      "Test Owner",
				STNKExpiredDate: sampleTime,
				TaxDate:         sampleTime,
				ManufactureYear: "2023",
				AssetCode:       "MOT",
				Color:           "BLACK",
				CC:              "150",
				NoChassis:       "CHS123",
				NoEngine:        "ENG123",
				STNKPhoto:       "http://test/stnk.jpg",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				KtpPhoto:    "http://test/ktp.jpg",
				SelfiePhoto: "http://test/selfie.jpg",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				InstallmentAmount: 1000000,
				AssetCategoryID:   "CAT1",
				Tenor:             24,
				OTR:               25000000,
				FinancePurpose:    "PRODUCTIVE",
				Dealer:            "PSA",
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				CustomerID: 12345,
			},
			marketingProgram: entity.TrxPrincipleMarketingProgram{
				ProgramID:                  "PRG001",
				ProgramName:                "Test Program",
				ProductOfferingID:          "OFF001",
				ProductOfferingDescription: "Test Offering",
				LoanAmount:                 20000000,
				LoanAmountMaximum:          22000000,
				AdminFee:                   500000,
				ProvisionFee:               200000,
				DPAmount:                   5000000,
				FinanceAmount:              20000000,
			},
			filteringKMB: entity.FilteringKMB{
				Decision:                        "PASS",
				Reason:                          "OK",
				IsBlacklist:                     0,
				NextProcess:                     1,
				CustomerStatusKMB:               interface{}(constant.STATUS_KONSUMEN_NEW),
				TotalBakiDebetNonCollateralBiro: "invalid",
			},
			elaborateLTV: entity.MappingElaborateLTV{
				LTV: 80,
			},
			detailBiro: []entity.TrxDetailBiro{
				{
					Subject:      constant.CUSTOMER,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
				{
					Subject:      constant.SPOUSE,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
			},
			branchResStatus: 200,
			branchResBody: `{
				"data": {
					"branch_name": "Test Branch"
				}
			}`,
			sallyResStatus: 200,
			sallyResBody: `{
				"message": "Success"
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " baki debet strconv.ParseFloat: parsing \"invalid\": invalid syntax"),
		},
		{
			name:       "error submit to sally",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				BranchID:        "BR001",
				CMOID:           "CMO123",
				CMOName:         "Test CMO",
				LicensePlate:    "B1234CD",
				BPKBName:        "KK",
				OwnerAsset:      "Test Owner",
				STNKExpiredDate: sampleTime,
				TaxDate:         sampleTime,
				ManufactureYear: "2023",
				AssetCode:       "MOT",
				Color:           "BLACK",
				CC:              "150",
				NoChassis:       "CHS123",
				NoEngine:        "ENG123",
				STNKPhoto:       "http://test/stnk.jpg",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				KtpPhoto:    "http://test/ktp.jpg",
				SelfiePhoto: "http://test/selfie.jpg",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				InstallmentAmount: 1000000,
				AssetCategoryID:   "CAT1",
				Tenor:             24,
				OTR:               25000000,
				FinancePurpose:    "PRODUCTIVE",
				Dealer:            "PSA",
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				CustomerID: 12345,
			},
			marketingProgram: entity.TrxPrincipleMarketingProgram{
				ProgramID:                  "PRG001",
				ProgramName:                "Test Program",
				ProductOfferingID:          "OFF001",
				ProductOfferingDescription: "Test Offering",
				LoanAmount:                 20000000,
				LoanAmountMaximum:          22000000,
				AdminFee:                   500000,
				ProvisionFee:               200000,
				DPAmount:                   5000000,
				FinanceAmount:              20000000,
			},
			filteringKMB: entity.FilteringKMB{
				Decision:    "PASS",
				Reason:      "OK",
				IsBlacklist: 0,
				NextProcess: 1,
			},
			elaborateLTV: entity.MappingElaborateLTV{
				LTV: 80,
			},
			detailBiro: []entity.TrxDetailBiro{
				{
					Subject:      constant.CUSTOMER,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
				{
					Subject:      constant.SPOUSE,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
			},
			branchResStatus: 200,
			branchResBody: `{
				"data": {
					"branch_name": "Test Branch"
				}
			}`,
			sallyErr:      errors.New("something went wrong"),
			expectedError: errors.New("something went wrong"),
		},
		{
			name:       "error code submit to sally",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				BranchID:        "BR001",
				CMOID:           "CMO123",
				CMOName:         "Test CMO",
				LicensePlate:    "B1234CD",
				BPKBName:        "KK",
				OwnerAsset:      "Test Owner",
				STNKExpiredDate: sampleTime,
				TaxDate:         sampleTime,
				ManufactureYear: "2023",
				AssetCode:       "MOT",
				Color:           "BLACK",
				CC:              "150",
				NoChassis:       "CHS123",
				NoEngine:        "ENG123",
				STNKPhoto:       "http://test/stnk.jpg",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				KtpPhoto:    "http://test/ktp.jpg",
				SelfiePhoto: "http://test/selfie.jpg",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				InstallmentAmount: 1000000,
				AssetCategoryID:   "CAT1",
				Tenor:             24,
				OTR:               25000000,
				FinancePurpose:    "PRODUCTIVE",
				Dealer:            "PSA",
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				CustomerID: 12345,
			},
			marketingProgram: entity.TrxPrincipleMarketingProgram{
				ProgramID:                  "PRG001",
				ProgramName:                "Test Program",
				ProductOfferingID:          "OFF001",
				ProductOfferingDescription: "Test Offering",
				LoanAmount:                 20000000,
				LoanAmountMaximum:          22000000,
				AdminFee:                   500000,
				ProvisionFee:               200000,
				DPAmount:                   5000000,
				FinanceAmount:              20000000,
			},
			filteringKMB: entity.FilteringKMB{
				Decision:    "PASS",
				Reason:      "OK",
				IsBlacklist: 0,
				NextProcess: 1,
			},
			elaborateLTV: entity.MappingElaborateLTV{
				LTV: 80,
			},
			detailBiro: []entity.TrxDetailBiro{
				{
					Subject:      constant.CUSTOMER,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
				{
					Subject:      constant.SPOUSE,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
			},
			branchResStatus: 200,
			branchResBody: `{
				"data": {
					"branch_name": "Test Branch"
				}
			}`,
			sallyResStatus: 500,
			expectedError:  errors.New(constant.ERROR_UPSTREAM + " - Sally Submit 2W Principle Error"),
		},
		{
			name:       "error unmarshal submit to sally",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				BranchID:        "BR001",
				CMOID:           "CMO123",
				CMOName:         "Test CMO",
				LicensePlate:    "B1234CD",
				BPKBName:        "KK",
				OwnerAsset:      "Test Owner",
				STNKExpiredDate: sampleTime,
				TaxDate:         sampleTime,
				ManufactureYear: "2023",
				AssetCode:       "MOT",
				Color:           "BLACK",
				CC:              "150",
				NoChassis:       "CHS123",
				NoEngine:        "ENG123",
				STNKPhoto:       "http://test/stnk.jpg",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				KtpPhoto:    "http://test/ktp.jpg",
				SelfiePhoto: "http://test/selfie.jpg",
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				InstallmentAmount: 1000000,
				AssetCategoryID:   "CAT1",
				Tenor:             24,
				OTR:               25000000,
				FinancePurpose:    "PRODUCTIVE",
				Dealer:            "PSA",
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				CustomerID: 12345,
			},
			marketingProgram: entity.TrxPrincipleMarketingProgram{
				ProgramID:                  "PRG001",
				ProgramName:                "Test Program",
				ProductOfferingID:          "OFF001",
				ProductOfferingDescription: "Test Offering",
				LoanAmount:                 20000000,
				LoanAmountMaximum:          22000000,
				AdminFee:                   500000,
				ProvisionFee:               200000,
				DPAmount:                   5000000,
				FinanceAmount:              20000000,
			},
			filteringKMB: entity.FilteringKMB{
				Decision:    "PASS",
				Reason:      "OK",
				IsBlacklist: 0,
				NextProcess: 1,
			},
			elaborateLTV: entity.MappingElaborateLTV{
				LTV: 80,
			},
			detailBiro: []entity.TrxDetailBiro{
				{
					Subject:      constant.CUSTOMER,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
				{
					Subject:      constant.SPOUSE,
					Score:        "750",
					UrlPdfReport: "http://test/report.pdf",
				},
			},
			branchResStatus: 200,
			branchResBody: `{
				"data": {
					"branch_name": "Test Branch"
				}
			}`,
			sallyResStatus: 200,
			sallyResBody:   `-`,
			expectedError:  errors.New("invalid character ' ' in numeric literal"),
		},
		{
			name:            "error get principle step one",
			prospectID:      "PROS-123",
			errPrincipleOne: errors.New("database error"),
			filteringKMB:    defaultFilteringKMB, // Add default value
			expectedError:   errors.New("database error"),
		},
		{
			name:            "error get principle step two",
			prospectID:      "PROS-123",
			errPrincipleTwo: errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
		{
			name:              "error get principle step three",
			prospectID:        "PROS-123",
			errPrincipleThree: errors.New("database error"),
			expectedError:     errors.New("database error"),
		},
		{
			name:                "error get emergency contact",
			prospectID:          "PROS-123",
			errEmergencyContact: errors.New("database error"),
			expectedError:       errors.New("database error"),
		},
		{
			name:                "error get marketing program",
			prospectID:          "PROS-123",
			errMarketingProgram: errors.New("database error"),
			expectedError:       errors.New("database error"),
		},
		{
			name:            "error get filtering result",
			prospectID:      "PROS-123",
			errFilteringKMB: errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
		{
			name:            "error get elaborate ltv",
			prospectID:      "PROS-123",
			errElaborateLTV: errors.New("database error"),
			expectedError:   errors.New("database error"),
		},
		{
			name:          "error get detail biro",
			prospectID:    "PROS-123",
			errDetailBiro: errors.New("database error"),
			expectedError: errors.New("database error"),
		},
		{
			name:            "error branch api - internal server error",
			prospectID:      "PROS-123",
			branchResStatus: 500,
			expectedError:   errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Branch Error"),
		},
		{
			name:          "error branch api",
			prospectID:    "PROS-123",
			branchErr:     errors.New("something wrong"),
			expectedError: errors.New("something wrong"),
		},
		{
			name:            "error branch api unmarshal",
			prospectID:      "PROS-123",
			branchResStatus: 200,
			branchResBody:   "invalid json",
			expectedError:   errors.New("unexpected end of JSON input"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockProducer := new(platformEventMockery.PlatformEventInterface)

			mockRepository.On("GetPrincipleStepOne", tc.prospectID).Return(tc.principleStepOne, tc.errPrincipleOne)
			mockRepository.On("GetPrincipleStepTwo", tc.prospectID).Return(tc.principleStepTwo, tc.errPrincipleTwo)
			mockRepository.On("GetPrincipleStepThree", tc.prospectID).Return(tc.principleStepThree, tc.errPrincipleThree)
			mockRepository.On("GetPrincipleEmergencyContact", tc.prospectID).Return(tc.emergencyContact, tc.errEmergencyContact)
			mockRepository.On("GetPrincipleMarketingProgram", tc.prospectID).Return(tc.marketingProgram, tc.errMarketingProgram)
			mockRepository.On("GetFilteringResult", tc.prospectID).Return(tc.filteringKMB, tc.errFilteringKMB)
			mockRepository.On("GetElaborateLtv", tc.prospectID).Return(tc.elaborateLTV, tc.errElaborateLTV)
			mockRepository.On("GetTrxDetailBIro", tc.prospectID).Return(tc.detailBiro, tc.errDetailBiro)

			if tc.errPrincipleOne == nil && tc.errPrincipleTwo == nil && tc.errPrincipleThree == nil &&
				tc.errEmergencyContact == nil && tc.errMarketingProgram == nil && tc.errFilteringKMB == nil &&
				tc.errElaborateLTV == nil && tc.errDetailBiro == nil {

				branchURL := os.Getenv("MDM_MASTER_BRANCH_URL") + tc.principleStepOne.BranchID
				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_GET, branchURL,
					httpmock.NewStringResponder(tc.branchResStatus, tc.branchResBody))
				branchResp, _ := rst.R().Get(branchURL)

				mockHttpClient.On("EngineAPI",
					ctx,
					constant.DILEN_KMB_LOG,
					branchURL,
					[]byte(nil),
					map[string]string{
						"Content-Type":  "application/json",
						"Authorization": accessToken,
					},
					constant.METHOD_GET,
					false,
					0,
					mock.AnythingOfType("int"),
					tc.prospectID,
					accessToken,
				).Return(branchResp, tc.branchErr)

				if tc.branchResStatus == 200 {
					sallyURL := os.Getenv("SALLY_SUBMISSION_2W_PRINCIPLE_URL")
					httpmock.RegisterResponder(constant.METHOD_POST, sallyURL,
						httpmock.NewStringResponder(tc.sallyResStatus, tc.sallyResBody))
					sallyResp, _ := rst.R().Post(sallyURL)

					mockHttpClient.On("EngineAPI",
						ctx,
						constant.DILEN_KMB_LOG,
						sallyURL,
						mock.MatchedBy(func(param []byte) bool {
							var js map[string]interface{}
							return json.Unmarshal(param, &js) == nil
						}),
						map[string]string{
							"Content-Type":  "application/json",
							"Authorization": accessToken,
						},
						constant.METHOD_POST,
						false,
						0,
						mock.AnythingOfType("int"),
						tc.prospectID,
						accessToken,
					).Return(sallyResp, tc.sallyErr)

					if tc.sallyResStatus == 200 {
						mockProducer.On("PublishEvent",
							ctx,
							accessToken,
							constant.TOPIC_SUBMISSION_PRINCIPLE,
							constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE,
							tc.prospectID,
							mock.MatchedBy(func(data map[string]interface{}) bool {
								return data["order_id"] == tc.prospectID &&
									data["source"] == 3 &&
									data["status_code"] == constant.PRINCIPLE_STATUS_SUBMIT_SALLY &&
									data["product_name"] == tc.principleStepOne.AssetCode &&
									data["branch_code"] == tc.principleStepOne.BranchID &&
									data["asset_type_code"] == constant.KPM_ASSET_TYPE_CODE_MOTOR
							}),
							0,
						).Return(tc.errPublishEvent)
					}
				}
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, mockProducer)

			err := usecase.PrincipleMarketingProgram(ctx, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMapperBPKBOwnershipStatusID(t *testing.T) {
	testCases := []struct {
		name          string
		bpkbName      string
		expectedValue int
	}{
		{
			name:          "K returns 1",
			bpkbName:      "K",
			expectedValue: 1,
		},
		{
			name:          "P returns 2",
			bpkbName:      "P",
			expectedValue: 2,
		},
		{
			name:          "KK returns 3",
			bpkbName:      "KK",
			expectedValue: 3,
		},
		{
			name:          "O returns 4",
			bpkbName:      "O",
			expectedValue: 4,
		},
		{
			name:          "unknown returns 4",
			bpkbName:      "UNKNOWN",
			expectedValue: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := MapperBPKBOwnershipStatusID(tc.bpkbName)
			require.Equal(t, tc.expectedValue, result)
		})
	}
}
