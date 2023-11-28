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
	os.Setenv("AGREEMENT_OF_CHASSIS_NUMBER_URL", "/")

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
		codeNtfTopup         int
		respNtfTopup         string
		errNtfTopup          error
		topup                response.IntegratorAgreementChassisNumber
		ntfAmount            float64
		customerID           string
	}{
		{
			name: "test prescreening",
			req: request.Metrics{
				CustomerPersonal: request.CustomerPersonal{
					IDNumber:          "123456",
					LegalName:         "TEST",
					BirthDate:         "1999-09-09",
					SurgateMotherName: "TEST",
					MaritalStatus:     constant.MARRIED,
				},
				Apk: request.Apk{
					NTF: 100000,
				},
				Item: request.Item{
					NoChassis: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:          "234567",
					LegalName:         "SPOUSE",
					BirthDate:         "1997-09-09",
					SurgateMotherName: "MOTHER",
				},
			},
			codeNtfOther: 200,
			respNtfOther: `{ "messages": "LOS - NTF Pending", "errors": null, "data": { "new_wg": 0, "wg_offline": 0, "kmb": 0, "kmob": 0, "uc": 0, "new_kmb": 0, "outstanding": 0 }, "server_time": "2023-09-21T07:40:06+07:00", "request_id": "f2d13daf-5d48-461e-9e36-cbd75a7f3847" }`,
			codeNtfTopup: 200,
			respNtfTopup: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", "installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, "outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "1f4c0309-5230-4cf1-beda-34d357470301", "timestamp": "2023-11-25 13:18:43" }`,
		},
		{
			name: "test prescreening ntf pending error",
			req: request.Metrics{
				CustomerPersonal: request.CustomerPersonal{
					IDNumber:          "123456",
					LegalName:         "TEST",
					BirthDate:         "1999-09-09",
					SurgateMotherName: "TEST",
					MaritalStatus:     constant.MARRIED,
				},
				Apk: request.Apk{
					NTF: 100000,
				},
				Item: request.Item{
					NoChassis: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:          "234567",
					LegalName:         "SPOUSE",
					BirthDate:         "1997-09-09",
					SurgateMotherName: "MOTHER",
				},
			},
			errNtfOtherApi: errors.New("api error"),
			errFinal:       errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error"),
			respNtfOther:   `{ "messages": "LOS - NTF Pending", "errors": null, "data": { "new_wg": 0, "wg_offline": 0, "kmb": 0, "kmob": 0, "uc": 0, "new_kmb": 0, "outstanding": 0 }, "server_time": "2023-09-21T07:40:06+07:00", "request_id": "f2d13daf-5d48-461e-9e36-cbd75a7f3847" }`,
			codeNtfTopup:   200,
			respNtfTopup:   `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", "installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, "outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "1f4c0309-5230-4cf1-beda-34d357470301", "timestamp": "2023-11-25 13:18:43" }`,
		},
		{
			name: "test prescreening ntf pending api error",
			req: request.Metrics{
				CustomerPersonal: request.CustomerPersonal{
					IDNumber:          "123456",
					LegalName:         "TEST",
					BirthDate:         "1999-09-09",
					SurgateMotherName: "TEST",
					MaritalStatus:     constant.MARRIED,
				},
				Apk: request.Apk{
					NTF: 100000,
				},
				Item: request.Item{
					NoChassis: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:          "234567",
					LegalName:         "SPOUSE",
					BirthDate:         "1997-09-09",
					SurgateMotherName: "MOTHER",
				},
			},
			codeNtfOther: 500,
			errFinal:     errors.New(constant.ERROR_UPSTREAM + " - Call NTF Pending API Error"),
			respNtfOther: `{ "messages": "LOS - NTF Pending", "errors": null, "data": { "new_wg": 0, "wg_offline": 0, "kmb": 0, "kmob": 0, "uc": 0, "new_kmb": 0, "outstanding": 0 }, "server_time": "2023-09-21T07:40:06+07:00", "request_id": "f2d13daf-5d48-461e-9e36-cbd75a7f3847" }`,
			codeNtfTopup: 200,
			respNtfTopup: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", "installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, "outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "1f4c0309-5230-4cf1-beda-34d357470301", "timestamp": "2023-11-25 13:18:43" }`,
		},
		{
			name: "test prescreening ntf topup api error",
			req: request.Metrics{
				CustomerPersonal: request.CustomerPersonal{
					IDNumber:          "123456",
					LegalName:         "TEST",
					BirthDate:         "1999-09-09",
					SurgateMotherName: "TEST",
					MaritalStatus:     constant.MARRIED,
				},
				Apk: request.Apk{
					NTF: 100000,
				},
				Item: request.Item{
					NoChassis: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:          "234567",
					LegalName:         "SPOUSE",
					BirthDate:         "1997-09-09",
					SurgateMotherName: "MOTHER",
				},
			},
			codeNtfOther: 200,
			respNtfOther: `{ "messages": "LOS - NTF Pending", "errors": null, "data": { "new_wg": 0, "wg_offline": 0, "kmb": 0, "kmob": 0, "uc": 0, "new_kmb": 0, "outstanding": 0 }, "server_time": "2023-09-21T07:40:06+07:00", "request_id": "f2d13daf-5d48-461e-9e36-cbd75a7f3847" }`,
			codeNtfTopup: 500,
			respNtfTopup: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", "installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, "outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "1f4c0309-5230-4cf1-beda-34d357470301", "timestamp": "2023-11-25 13:18:43" }`,
			errFinal:     errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error"),
		},
		{
			name: "test prescreening ntf topup api error",
			req: request.Metrics{
				CustomerPersonal: request.CustomerPersonal{
					IDNumber:          "123456",
					LegalName:         "TEST",
					BirthDate:         "1999-09-09",
					SurgateMotherName: "TEST",
					MaritalStatus:     constant.MARRIED,
				},
				Apk: request.Apk{
					NTF: 100000,
				},
				Item: request.Item{
					NoChassis: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:          "234567",
					LegalName:         "SPOUSE",
					BirthDate:         "1997-09-09",
					SurgateMotherName: "MOTHER",
				},
			},
			codeNtfOther: 200,
			respNtfOther: `{ "messages": "LOS - NTF Pending", "errors": null, "data": { "new_wg": 0, "wg_offline": 0, "kmb": 0, "kmob": 0, "uc": 0, "new_kmb": 0, "outstanding": 0 }, "server_time": "2023-09-21T07:40:06+07:00", "request_id": "f2d13daf-5d48-461e-9e36-cbd75a7f3847" }`,
			errNtfTopup:  errors.New("api ntf topup error"),
			respNtfTopup: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", "installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, "outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "1f4c0309-5230-4cf1-beda-34d357470301", "timestamp": "2023-11-25 13:18:43" }`,
			errFinal:     errors.New(constant.ERROR_UPSTREAM + " - Call NTF Topup API Error"),
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

			jsonCustomer, _ := json.Marshal(customerData[0])

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_PENDING_URL"), jsonCustomer, map[string]string{}, constant.METHOD_POST, true, 3, 60, "", accessToken).Return(resp, tc.errNtfOtherApi).Once()

			if tc.req.CustomerPersonal.MaritalStatus == constant.MARRIED && tc.req.CustomerSpouse != nil {
				spouse := *tc.req.CustomerSpouse
				customerData = append(customerData, request.CustomerData{
					IDNumber:   spouse.IDNumber,
					LegalName:  spouse.LegalName,
					BirthDate:  spouse.BirthDate,
					MotherName: spouse.SurgateMotherName,
				})
				jsonCustomer, _ = json.Marshal(customerData[1])
				mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("NTF_PENDING_URL"), jsonCustomer, map[string]string{}, constant.METHOD_POST, true, 3, 60, "", accessToken).Return(resp, tc.errNtfOtherApi).Once()
			}

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"), httpmock.NewStringResponder(tc.codeNtfTopup, tc.respNtfTopup))
			resp, _ = rst.R().Post(os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"))
			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+tc.req.Item.NoChassis, mock.Anything, map[string]string{}, constant.METHOD_GET, true, 3, 60, "", accessToken).Return(resp, tc.errNtfTopup).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			trxPrescreening, _, _, err := usecase.Prescreening(ctx, tc.req, tc.filtering, accessToken)
			require.Equal(t, tc.trxPrescreening, trxPrescreening)
			require.Equal(t, tc.errFinal, err)
		})
	}

}
