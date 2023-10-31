package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDsrCheck(t *testing.T) {
	// always set the valid url
	os.Setenv("INSTALLMENT_PENDING_URL", "http://localhost/")
	config := entity.AppConfig{
		Key:   "parameterize",
		Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
	}
	var (
		idNumber   string = "32030143096XXXX6"
		legalName  string = "SOE***E"
		birthDate  string = "1966-09-03"
		motherName string = "IBU KANDUNG"
	)

	konsumenOnly := []request.CustomerData{
		{
			StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
			IDNumber:       idNumber,
			LegalName:      legalName,
			BirthDate:      birthDate,
			MotherName:     motherName,
		},
	}

	// konsumenOnlyRo := []request.CustomerData{
	// 	{
	// 		StatusKonsumen: constant.STATUS_KONSUMEN_RO,
	// 		IDNumber:       idNumber,
	// 		LegalName:      legalName,
	// 		BirthDate:      birthDate,
	// 		MotherName:     motherName,
	// 	},
	// }

	// konsumenSpouseAo := []request.CustomerData{
	// 	{
	// 		StatusKonsumen: constant.STATUS_KONSUMEN_AO,
	// 		IDNumber:       idNumber,
	// 		LegalName:      legalName,
	// 		BirthDate:      birthDate,
	// 		MotherName:     motherName,
	// 	},
	// 	{
	// 		IDNumber:   idNumber,
	// 		LegalName:  legalName,
	// 		BirthDate:  birthDate,
	// 		MotherName: motherName,
	// 	},
	// }

	testcases := []struct {
		name                                                                    string
		body                                                                    string
		dupcheckConfig                                                          entity.AppConfig
		code                                                                    int
		result                                                                  response.UsecaseApi
		errResp                                                                 error
		errResult                                                               error
		errGetTrx                                                               error
		req                                                                     request.DupcheckApi
		responseAgreementChassisNumber                                          response.AgreementChassisNumber
		customerData                                                            []request.CustomerData
		installmentAmount, installmentConfins, installmentConfinsSpouse, income float64
	}{
		{
			name:           "DsrCheck error EngineAPI",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData: konsumenOnly,
			errResult:    errors.New(constant.ERROR_UPSTREAM + " - Call Installment Pending API Error"),
		},
		{
			name:           "DsrCheck pass",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData: konsumenOnly,
			code:         200,
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "NEW - Confins DSR <= 35",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            2.6,
			},
			installmentAmount: 260000,
			income:            10000000,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			accessToken := "access-token"
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			var configValue response.DupcheckConfig
			json.Unmarshal([]byte(tc.dupcheckConfig.Value), &configValue)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("INSTALLMENT_PENDING_URL"), httpmock.NewStringResponder(tc.code, tc.body))
			resp, _ := rst.R().Post(os.Getenv("INSTALLMENT_PENDING_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("INSTALLMENT_PENDING_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, true, 3, 60, tc.req.ProspectID, accessToken).Return(resp, tc.errResp).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, _, _, _, _, err := usecase.DsrCheck(ctx, tc.req, tc.customerData, tc.installmentAmount, tc.installmentConfins, tc.installmentConfinsSpouse, tc.income, accessToken, configValue)
			require.Equal(t, tc.result, data)
			// require.Equal(t, tc.result, dsr)
			require.Equal(t, tc.errResult, err)
		})
	}

}
