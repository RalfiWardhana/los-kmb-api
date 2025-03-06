package usecase

import (
	"context"
	"encoding/json"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
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

func TestPrincipleCoreCustomer(t *testing.T) {
	ctx := context.Background()
	accessToken := "test-token"

	sampleTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("CUSTOMER_V3_BASE_URL", "http://test-customer")

	testCases := []struct {
		name                    string
		prospectID              string
		principleStepOne        entity.TrxPrincipleStepOne
		errPrincipleOne         error
		principleStepTwo        entity.TrxPrincipleStepTwo
		errPrincipleTwo         error
		principleStepThree      entity.TrxPrincipleStepThree
		errPrincipleThree       error
		emergencyContact        entity.TrxPrincipleEmergencyContact
		errEmergencyContact     error
		validateResStatus       int
		validateResBody         string
		validateErr             error
		insertResStatus         int
		insertResBody           string
		insertErr               error
		updateResStatus         int
		updateResBody           string
		updateErr               error
		errSaveEmergencyContact error
		expectedError           error
	}{
		{
			name:       "success single customer",
			prospectID: "PROS-123",
			principleStepOne: entity.TrxPrincipleStepOne{
				CMOID:             "CMO123",
				HomeStatus:        "OWN",
				StaySinceMonth:    1,
				StaySinceYear:     2020,
				ResidenceAddress:  "Test Address",
				ResidenceCity:     "Test City",
				ResidenceProvince: "Test Province",
				ResidencePhone:    "021123456",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:          "1234567890",
				LegalName:         "Test Name",
				FullName:          "Test Full Name",
				BirthDate:         sampleTime,
				BirthPlace:        "Test City",
				Gender:            "M",
				SurgateMotherName: "Test Mother",
				MobilePhone:       "08123456789",
				KtpPhoto:          "http://test/ktp.jpg",
				SelfiePhoto:       "http://test/selfie.jpg",
				MaritalStatus:     constant.MARRIED,
			},
			principleStepThree: entity.TrxPrincipleStepThree{
				IDNumber:              "1234567890",
				MonthlyVariableIncome: 5000000,
			},
			emergencyContact: entity.TrxPrincipleEmergencyContact{
				Name:         "Emergency Test",
				MobilePhone:  "08111111111",
				Relationship: "SIBLING",
			},
			validateResStatus: 200,
			validateResBody: `{
				"data": {
					"customer_id": 0,
					"kpm_id": 0
				}
			}`,
			insertResStatus: 200,
			insertResBody: `{
				"data": {
					"customer_id": 12345
				}
			}`,
			updateResStatus: 200,
			updateResBody: `{
				"message": "Success"
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetPrincipleStepOne", tc.prospectID).Return(tc.principleStepOne, tc.errPrincipleOne)
			mockRepository.On("GetPrincipleStepTwo", tc.prospectID).Return(tc.principleStepTwo, tc.errPrincipleTwo)
			mockRepository.On("GetPrincipleStepThree", tc.prospectID).Return(tc.principleStepThree, tc.errPrincipleThree)
			mockRepository.On("GetPrincipleEmergencyContact", tc.prospectID).Return(tc.emergencyContact, tc.errEmergencyContact)

			if tc.errPrincipleOne == nil && tc.errPrincipleTwo == nil && tc.errPrincipleThree == nil && tc.errEmergencyContact == nil {
				validateURL := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/validate-data"

				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				httpmock.RegisterResponder(constant.METHOD_POST, validateURL,
					httpmock.NewStringResponder(tc.validateResStatus, tc.validateResBody))
				validateResp, _ := rst.R().Post(validateURL)

				mockHttpClient.On("EngineAPI",
					ctx,
					constant.DILEN_KMB_LOG,
					validateURL,
					mock.Anything,
					mock.Anything,
					constant.METHOD_POST,
					false,
					0,
					mock.AnythingOfType("int"),
					tc.prospectID,
					accessToken,
				).Return(validateResp, tc.validateErr)

				if tc.validateResStatus == 200 {
					var validateRes response.CustomerDomainValidate
					json.Unmarshal([]byte(tc.validateResBody), &validateRes)

					if validateRes.Data.CustomerID > 0 || validateRes.Data.KPMID > 0 {
						expectedEC := tc.emergencyContact
						expectedEC.CustomerID = validateRes.Data.CustomerID
						expectedEC.KPMID = validateRes.Data.KPMID

						mockRepository.On("SavePrincipleEmergencyContact", expectedEC, tc.principleStepThree.IDNumber).Return(tc.errSaveEmergencyContact)
					}

					if validateRes.Data.CustomerID == 0 {
						insertURL := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/transaction"
						httpmock.RegisterResponder(constant.METHOD_POST, insertURL,
							httpmock.NewStringResponder(tc.insertResStatus, tc.insertResBody))
						insertResp, _ := rst.R().Post(insertURL)

						mockHttpClient.On("EngineAPI",
							ctx,
							constant.DILEN_KMB_LOG,
							insertURL,
							mock.Anything,
							mock.Anything,
							constant.METHOD_POST,
							false,
							0,
							mock.AnythingOfType("int"),
							tc.prospectID,
							accessToken,
						).Return(insertResp, tc.insertErr)

						if tc.insertResStatus == 200 {
							var insertRes response.CustomerDomainInsert
							json.Unmarshal([]byte(tc.insertResBody), &insertRes)

							if insertRes.Data.CustomerID > 0 && insertRes.Data.CustomerID != validateRes.Data.CustomerID {
								expectedEC := tc.emergencyContact
								expectedEC.CustomerID = insertRes.Data.CustomerID

								mockRepository.On("SavePrincipleEmergencyContact", expectedEC, tc.principleStepThree.IDNumber).Return(tc.errSaveEmergencyContact)
							}
						}
					}

					updateURL := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/transaction/" + tc.prospectID
					httpmock.RegisterResponder(constant.METHOD_PUT, updateURL,
						httpmock.NewStringResponder(tc.updateResStatus, tc.updateResBody))
					updateResp, _ := rst.R().Put(updateURL)

					mockHttpClient.On("EngineAPI",
						ctx,
						constant.DILEN_KMB_LOG,
						updateURL,
						mock.Anything,
						mock.Anything,
						constant.METHOD_PUT,
						false,
						0,
						mock.AnythingOfType("int"),
						tc.prospectID,
						accessToken,
					).Return(updateResp, tc.updateErr)
				}
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			err := usecase.PrincipleCoreCustomer(ctx, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			mockRepository.AssertExpectations(t)
			mockHttpClient.AssertExpectations(t)
		})
	}
}
