package usecase

import (
	"context"
	"encoding/json"
	"errors"
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
			name:            "error get principle step one",
			prospectID:      "PROS-001",
			errPrincipleOne: errors.New("failed to get step one"),
			expectedError:   errors.New("failed to get step one"),
		},
		{
			name:             "error get principle step two",
			prospectID:       "PROS-002",
			principleStepOne: entity.TrxPrincipleStepOne{},
			errPrincipleTwo:  errors.New("failed to get step two"),
			expectedError:    errors.New("failed to get step two"),
		},
		{
			name:              "error get principle step three",
			prospectID:        "PROS-003",
			principleStepOne:  entity.TrxPrincipleStepOne{},
			principleStepTwo:  entity.TrxPrincipleStepTwo{},
			errPrincipleThree: errors.New("failed to get step three"),
			expectedError:     errors.New("failed to get step three"),
		},
		{
			name:                "error get emergency contact",
			prospectID:          "PROS-004",
			principleStepOne:    entity.TrxPrincipleStepOne{},
			principleStepTwo:    entity.TrxPrincipleStepTwo{},
			principleStepThree:  entity.TrxPrincipleStepThree{},
			errEmergencyContact: errors.New("failed to get emergency contact"),
			expectedError:       errors.New("failed to get emergency contact"),
		},
		{
			name:               "error validate customer EngineAPI",
			prospectID:         "PROS-005",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateErr:        errors.New("engineAPI error"),
			expectedError:      errors.New("engineAPI error"),
		},
		{
			name:               "validate response not 200 or 400",
			prospectID:         "PROS-006",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  500,
			validateResBody:    `{}`,
			expectedError:      errors.New("upstream_service_error - Customer Validate Data Error"),
		},
		{
			name:               "unmarshal validate response error",
			prospectID:         "PROS-007",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  200,
			validateResBody:    `invalid-json`,
			expectedError:      &json.SyntaxError{},
		},
		{
			name:                    "error save emergency contact after validate",
			prospectID:              "PROS-008",
			principleStepOne:        entity.TrxPrincipleStepOne{},
			principleStepTwo:        entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree:      entity.TrxPrincipleStepThree{IDNumber: "1234567890"},
			emergencyContact:        entity.TrxPrincipleEmergencyContact{},
			validateResStatus:       200,
			validateResBody:         `{"data":{"customer_id":100,"kpm_id":0}}`,
			errSaveEmergencyContact: errors.New("save failed"),
			expectedError:           errors.New("save failed"),
		},
		{
			name:               "error insert customer EngineAPI",
			prospectID:         "PROS-009",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  200,
			validateResBody:    `{"data":{"customer_id":0,"kpm_id":0}}`,
			insertErr:          errors.New("insert failed"),
			expectedError:      errors.New("insert failed"),
		},
		{
			name:               "insert customer response not 200 or 400",
			prospectID:         "PROS-010",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  200,
			validateResBody:    `{"data":{"customer_id":0}}`,
			insertResStatus:    500,
			insertResBody:      `{}`,
			expectedError:      errors.New("upstream_service_error - Insert Customer Data Error"),
		},
		{
			name:               "insert customer unmarshal error",
			prospectID:         "PROS-011",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  200,
			validateResBody:    `{"data":{"customer_id":0}}`,
			insertResStatus:    200,
			insertResBody:      `invalid-json`,
			expectedError:      &json.SyntaxError{},
		},
		{
			name:                    "error save emergency contact after insert",
			prospectID:              "PROS-012",
			principleStepOne:        entity.TrxPrincipleStepOne{},
			principleStepTwo:        entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree:      entity.TrxPrincipleStepThree{IDNumber: "1234567890"},
			emergencyContact:        entity.TrxPrincipleEmergencyContact{},
			validateResStatus:       200,
			validateResBody:         `{"data":{"customer_id":0}}`,
			insertResStatus:         200,
			insertResBody:           `{"data":{"customer_id":9999}}`,
			errSaveEmergencyContact: errors.New("save error after insert"),
			expectedError:           errors.New("save error after insert"),
		},
		{
			name:               "update customer EngineAPI error",
			prospectID:         "PROS-013",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{IDNumber: "1234567890"},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  200,
			validateResBody:    `{"data":{"customer_id":0}}`,
			insertResStatus:    200,
			insertResBody:      `{"data":{"customer_id":0}}`,
			updateErr:          errors.New("update error"),
			expectedError:      errors.New("update error"),
		},
		{
			name:               "update customer response not 200 or 400",
			prospectID:         "PROS-014",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{IDNumber: "1234567890"},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  200,
			validateResBody:    `{"data":{"customer_id":0}}`,
			insertResStatus:    200,
			insertResBody:      `{"data":{"customer_id":0}}`,
			updateResStatus:    500,
			updateResBody:      `{}`,
			expectedError:      errors.New("upstream_service_error - Update Customer Transaction Error"),
		},
		{
			name:               "unmarshal update customer response error",
			prospectID:         "PROS-015",
			principleStepOne:   entity.TrxPrincipleStepOne{},
			principleStepTwo:   entity.TrxPrincipleStepTwo{BirthDate: sampleTime, MaritalStatus: constant.MARITAL_SINGLE},
			principleStepThree: entity.TrxPrincipleStepThree{IDNumber: "1234567890"},
			emergencyContact:   entity.TrxPrincipleEmergencyContact{},
			validateResStatus:  200,
			validateResBody:    `{"data":{"customer_id":0}}`,
			insertResStatus:    200,
			insertResBody:      `{"data":{"customer_id":0}}`,
			updateResStatus:    200,
			updateResBody:      `invalid-json`,
			expectedError:      &json.SyntaxError{},
		},
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
				SpouseBirthDate:   sampleTime,
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

			// Always set up all repository method expectations regardless of errors
			mockRepository.On("GetPrincipleStepOne", tc.prospectID).Return(tc.principleStepOne, tc.errPrincipleOne)
			mockRepository.On("GetPrincipleStepTwo", tc.prospectID).Return(tc.principleStepTwo, tc.errPrincipleTwo)
			mockRepository.On("GetPrincipleStepThree", tc.prospectID).Return(tc.principleStepThree, tc.errPrincipleThree)
			mockRepository.On("GetPrincipleEmergencyContact", tc.prospectID).Return(tc.emergencyContact, tc.errEmergencyContact)

			// If all repository calls are expected to succeed, set up the API calls
			if tc.errPrincipleOne == nil && tc.errPrincipleTwo == nil && tc.errPrincipleThree == nil && tc.errEmergencyContact == nil {
				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				validateURL := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/validate-data"
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

				// Only proceed with further setup if validate call is expected to succeed
				if tc.validateErr == nil && tc.validateResStatus == 200 {
					var validateRes response.CustomerDomainValidate
					err := json.Unmarshal([]byte(tc.validateResBody), &validateRes)
					if err == nil {
						// Setup for customer with existing ID
						if validateRes.Data.CustomerID > 0 || validateRes.Data.KPMID > 0 {
							expectedEC := tc.emergencyContact
							expectedEC.CustomerID = validateRes.Data.CustomerID
							expectedEC.KPMID = validateRes.Data.KPMID
							mockRepository.On("SavePrincipleEmergencyContact", expectedEC, tc.principleStepThree.IDNumber).Return(tc.errSaveEmergencyContact)
						}

						// Setup for insert customer case
						insertURL := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/transaction"
						httpmock.RegisterResponder(constant.METHOD_POST, insertURL,
							httpmock.NewStringResponder(tc.insertResStatus, tc.insertResBody))
						insertResp, _ := rst.R().Post(insertURL)

						// Only setup insert API call if needed (customer_id is 0)
						if validateRes.Data.CustomerID == 0 {
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

							// Only proceed if insert call is expected to succeed
							if tc.insertErr == nil && tc.insertResStatus == 200 {
								var insertRes response.CustomerDomainInsert
								err := json.Unmarshal([]byte(tc.insertResBody), &insertRes)
								if err == nil && insertRes.Data.CustomerID > 0 && insertRes.Data.CustomerID != validateRes.Data.CustomerID {
									expectedEC := tc.emergencyContact
									expectedEC.CustomerID = insertRes.Data.CustomerID
									mockRepository.On("SavePrincipleEmergencyContact", expectedEC, tc.principleStepThree.IDNumber).Return(tc.errSaveEmergencyContact)
								}
							}
						}

						// Only setup update call if we expect to get this far
						// This means either validation shows existing customer or insert is expected to succeed
						shouldSetupUpdate := (validateRes.Data.CustomerID > 0 && tc.errSaveEmergencyContact == nil) ||
							(validateRes.Data.CustomerID == 0 && (tc.insertErr == nil && tc.insertResStatus == 200) &&
								(tc.insertResBody == `{"data":{"customer_id":0}}` || tc.errSaveEmergencyContact == nil))

						if shouldSetupUpdate {
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
				}
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			err := usecase.PrincipleCoreCustomer(ctx, tc.prospectID, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				if _, ok := tc.expectedError.(*json.SyntaxError); ok {
					require.IsType(t, tc.expectedError, err)
				} else {
					require.EqualError(t, err, tc.expectedError.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
