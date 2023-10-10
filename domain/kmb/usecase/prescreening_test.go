package usecase

import (
	"context"
	"encoding/json"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrescreening(t *testing.T) {
	accessToken := "token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	os.Setenv("NTF_PENDING_URL", "/")

	testcases := []struct {
		name                 string
		req                  request.Metrics
		filtering            entity.FilteringKMB
		trxPrescreening      entity.TrxPrescreening
		trxFMF               response.TrxFMF
		trxDetail            entity.TrxDetail
		errGet               error
		errFinal             error
		codeNtfOther         int
		respNtfOther         string
		errNtfOtherApi       error
		ntfOther             response.NTFOther
		ntfDetails           response.NTFDetails
		ntfOtherAmount       float64
		ntfOtherAmountSpouse float64
		ntfConfinsAmount     response.OutstandingConfins
		confins              response.OutstandingConfins
		topup                response.IntegratorAgreementChassisNumber
		ntfAmount            float64
		customerID           string
	}{
		{
			name: "test prescreening",
			trxPrescreening: entity.TrxPrescreening{
				Decision:   "APR",
				Reason:     "Dokumen Sesuai",
				CreatedBy:  "SYSTEM",
				DecisionBy: "SYSTEM",
			},
			req: request.Metrics{
				CustomerPersonal: request.CustomerPersonal{
					IDNumber:          "123456",
					LegalName:         "TEST",
					BirthDate:         "1999-09-09",
					SurgateMotherName: "TEST",
				},
			},
			codeNtfOther: 200,
			respNtfOther: `{ "messages": "LOS - NTF Pending", "errors": null, "data": { "new_wg": 0, "wg_offline": 0, "kmb": 0, "kmob": 0, "uc": 0, "new_kmb": 0, "outstanding": 0 }, "server_time": "2023-09-21T07:40:06+07:00", "request_id": "f2d13daf-5d48-461e-9e36-cbd75a7f3847" }`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("NTF_PENDING_URL"), httpmock.NewStringResponder(tc.codeNtfOther, tc.respNtfOther))
			resp, _ := rst.R().Post(os.Getenv("NTF_PENDING_URL"))

			var customerID string
			if tc.filtering.CustomerID != nil {
				customerID = tc.filtering.CustomerID.(string)
			}

			var customerData []request.CustomerData
			customerData = append(customerData, request.CustomerData{
				IDNumber:   tc.req.CustomerPersonal.IDNumber,
				LegalName:  tc.req.CustomerPersonal.LegalName,
				BirthDate:  tc.req.CustomerPersonal.BirthDate,
				MotherName: tc.req.CustomerPersonal.SurgateMotherName,
				CustomerID: customerID,
			})

			if tc.req.CustomerPersonal.MaritalStatus == constant.MARRIED && tc.req.CustomerSpouse != nil {
				spouse := *tc.req.CustomerSpouse
				customerData = append(customerData, request.CustomerData{
					IDNumber:   spouse.IDNumber,
					LegalName:  spouse.LegalName,
					BirthDate:  spouse.BirthDate,
					MotherName: spouse.SurgateMotherName,
				})
			}

			jsonCustomer, _ := json.Marshal(customerData[0])
			if tc.req.CustomerPersonal.MaritalStatus == constant.MARRIED && tc.req.CustomerSpouse != nil {
				jsonCustomer, _ = json.Marshal(customerData[1])
			}

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_PENDING_URL"), jsonCustomer, map[string]string{}, constant.METHOD_POST, true, 3, 60, "", accessToken).Return(resp, tc.errNtfOtherApi).Once()

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_PENDING_URL"), mock.Anything, map[string]string{}, constant.METHOD_GET, true, 3, 60, "", accessToken).Return(resp, tc.errNtfOtherApi).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			trxPrescreening, _, _, err := usecase.Prescreening(ctx, tc.req, tc.filtering, accessToken)
			require.Equal(t, tc.trxPrescreening, trxPrescreening)
			// require.Equal(t, tc.trxDetail, trxPrescreeningDetail)
			// require.Equal(t, tc.trxFMF, trxFMF)
			require.Equal(t, tc.errFinal, err)
		})
	}

}
