package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
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

func TestDsrCheck(t *testing.T) {
	// always set the valid url
	os.Setenv("INSTALLMENT_PENDING_URL", "http://localhost/")
	os.Setenv("AGREEMENT_OF_CHASSIS_NUMBER_URL", "http://localhost/")

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

	konsumenOnlyRo := []request.CustomerData{
		{
			StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			IDNumber:       idNumber,
			LegalName:      legalName,
			BirthDate:      birthDate,
			MotherName:     motherName,
		},
	}

	konsumenOnlyRoPrime := []request.CustomerData{
		{
			StatusKonsumen:  constant.STATUS_KONSUMEN_RO,
			IDNumber:        idNumber,
			LegalName:       legalName,
			BirthDate:       birthDate,
			MotherName:      motherName,
			CustomerSegment: constant.RO_AO_PRIME,
		},
	}

	konsumenOnlyAO := []request.CustomerData{
		{
			StatusKonsumen: constant.STATUS_KONSUMEN_AO,
			IDNumber:       idNumber,
			LegalName:      legalName,
			BirthDate:      birthDate,
			MotherName:     motherName,
		},
	}

	konsumenOnlyAoPrime := []request.CustomerData{
		{
			StatusKonsumen:  constant.STATUS_KONSUMEN_AO,
			IDNumber:        idNumber,
			LegalName:       legalName,
			BirthDate:       birthDate,
			MotherName:      motherName,
			CustomerSegment: constant.RO_AO_PRIME,
		},
	}

	konsumenOnlyAoPrio := []request.CustomerData{
		{
			StatusKonsumen:  constant.STATUS_KONSUMEN_AO,
			IDNumber:        idNumber,
			LegalName:       legalName,
			BirthDate:       birthDate,
			MotherName:      motherName,
			CustomerSegment: constant.RO_AO_PRIORITY,
		},
	}

	konsumenPasanganRo := []request.CustomerData{
		{
			StatusKonsumen: constant.STATUS_KONSUMEN_RO,
			IDNumber:       idNumber,
			LegalName:      legalName,
			BirthDate:      birthDate,
			MotherName:     motherName,
		},
		{
			StatusKonsumen: "",
			IDNumber:       idNumber,
			LegalName:      legalName,
			BirthDate:      birthDate,
			MotherName:     motherName,
		},
	}

	adminFee := float64(10000)

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
		bodyChassisNumber                                                       string
		codeChassisNumber                                                       int
		errRespChassisNumber                                                    error
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
				Reason:         "NEW - Confins DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            2.6,
			},
			installmentAmount: 260000,
			income:            10000000,
		},
		{
			name:           "DsrCheck reject",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData: konsumenOnly,
			code:         200,
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_DSRGT35,
				Reason:         "NEW - Confins DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            65,
			},
			installmentAmount: 260000,
			income:            400000,
		},
		{
			name:           "DsrCheck ro Call Installment Pending API Error",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData:       konsumenOnlyRo,
			code:               200,
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
			errResp:   errors.New(constant.ERROR_UPSTREAM + " - Call Installment Pending API Error"),
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Call Installment Pending API Error"),
		},
		{
			name:           "DsrCheck ro",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData: konsumenPasanganRo,
			code:         200,
			body:         `{"messages":"LOS - Installment Pending","errors":null,"data":{"new_wg":75000,"wg_offline":62000,"kmb":241000,"kmob":0,"uc":0,"new_kmb":0},"server_time":"2023-10-18T14:07:34+07:00","request_id":"aa533888-e940-4ac6-8fc2-aee5238d075d"}`,
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_DSRGT35,
				Reason:         "RO - Confins DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            304,
			},
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
		},
		{
			name:           "DsrCheck ro Call Get Agreement of Chassis Number Timeout",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData:       konsumenPasanganRo,
			code:               200,
			body:               `{"messages":"LOS - Installment Pending","errors":null,"data":{"new_wg":75000,"wg_offline":62000,"kmb":241000,"kmob":0,"uc":0,"new_kmb":0},"server_time":"2023-10-18T14:07:34+07:00","request_id":"aa533888-e940-4ac6-8fc2-aee5238d075d"}`,
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
			errRespChassisNumber: errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - DsrCheck Call Get Agreement of Chassis Number Timeout"),
			errResult:            errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - DsrCheck Call Get Agreement of Chassis Number Timeout"),
		},
		{
			name:           "DsrCheck ro Call Get Agreement of Chassis Number Error",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData:       konsumenPasanganRo,
			code:               200,
			body:               `{"messages":"LOS - Installment Pending","errors":null,"data":{"new_wg":75000,"wg_offline":62000,"kmb":241000,"kmob":0,"uc":0,"new_kmb":0},"server_time":"2023-10-18T14:07:34+07:00","request_id":"aa533888-e940-4ac6-8fc2-aee5238d075d"}`,
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  500,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - DsrCheck Call Get Agreement of Chassis Number Error"),
		},
		{
			name:           "DsrCheck ro Unmarshal Get Agreement of Chassis Number Error",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			customerData:       konsumenPasanganRo,
			code:               200,
			body:               `{"messages":"LOS - Installment Pending","errors":null,"data":{"new_wg":75000,"wg_offline":62000,"kmb":241000,"kmob":0,"uc":0,"new_kmb":0},"server_time":"2023-10-18T14:07:34+07:00","request_id":"aa533888-e940-4ac6-8fc2-aee5238d075d"}`,
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber:  `{ "response salah" }`,
			errResult:          errors.New(constant.ERROR_UPSTREAM + " - DsrCheck Unmarshal Get Agreement of Chassis Number Error"),
		},
		{
			name:           "DsrCheck ro prime",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID:      "TEST198091461892",
				RangkaNo:        "198091461892",
				CustomerSegment: constant.RO_AO_PRIME,
			},
			customerData: konsumenOnlyRoPrime,
			code:         200,
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "RO PRIME",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            114.99999999999999,
			},
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 0, "is_active": false, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
		},
		{
			name:           "DsrCheck ao top up",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID:      "TEST198091461892",
				RangkaNo:        "198091461892",
				CustomerSegment: constant.RO_AO_PRIME,
				OTRPrice:        17000000,
				DPAmount:        2000000,
			},
			customerData: konsumenOnlyAO,
			code:         200,
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_DSRGT35,
				Reason:         "AO Top Up - Confins DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            40,
			},
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 300000, "is_active": true, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 0, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
		},
		{
			name: "DsrCheck ao top up menunggak",
			dupcheckConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":13,"max_ovd":60,"max_ovd_ao_prime_priority":30,"max_ovd_ao_regular":0,"max_dsr":35,"angsuran_berjalan":6,"attempt_pmk_dsr":2,"attempt_nik":2,"attempt_chassis_number":3,"minimum_pencairan_ro_top_up":{"prime":20,"priority":30,"regular":30}}}`,
			},
			req: request.DupcheckApi{
				ProspectID:      "TEST198091461892",
				RangkaNo:        "198091461892",
				CustomerSegment: constant.RO_AO_PRIME,
				OTRPrice:        17000000,
				DPAmount:        2000000,
				AdminFee:        &adminFee,
				Dealer:          constant.DEALER_PSA,
			},
			customerData: konsumenOnlyAO,
			code:         200,
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PENCAIRAN_TOPUP,
				Reason:         "AO Top Up " + constant.REASON_PENCAIRAN_TOPUP,
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				Dsr:            40,
			},
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 300000, "is_active": true, "is_registered": false, "lc_installment": 3000000, "legal_name": "", "outstanding_interest": 3000000, 
			"outstanding_principal": 6000000, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
		},
		{
			name: "DsrCheck ao top up menunggak err Perhitungan OTR - DP harus lebih dari 0",
			dupcheckConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":13,"max_ovd":60,"max_ovd_ao_prime_priority":30,"max_ovd_ao_regular":0,"max_dsr":35,"angsuran_berjalan":6,"attempt_pmk_dsr":2,"attempt_nik":2,"attempt_chassis_number":3,"minimum_pencairan_ro_top_up":{"prime":20,"priority":30,"regular":30}}}`,
			},
			req: request.DupcheckApi{
				ProspectID:      "TEST198091461892",
				RangkaNo:        "198091461892",
				CustomerSegment: constant.RO_AO_PRIME,
				OTRPrice:        190000,
				DPAmount:        2000000,
				AdminFee:        &adminFee,
				Dealer:          constant.DEALER_PSA,
			},
			customerData:       konsumenOnlyAO,
			code:               200,
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 300000, "is_active": true, "is_registered": false, "lc_installment": 3000000, "legal_name": "", "outstanding_interest": 3000000, 
			"outstanding_principal": 6000000, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Perhitungan OTR - DP harus lebih dari 0"),
		},
		{
			name:           "DsrCheck ao prime",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID:      "TEST198091461892",
				RangkaNo:        "198091461892",
				CustomerSegment: constant.RO_AO_PRIME,
				OTRPrice:        17000000,
				DPAmount:        2000000,
			},
			customerData: konsumenOnlyAoPrime,
			code:         200,
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "AO PRIME Top Up",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            65,
			},
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             400000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 200000, "is_active": true, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 12000000, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
		},
		{
			name:           "DsrCheck ao priority",
			dupcheckConfig: config,
			req: request.DupcheckApi{
				ProspectID:      "TEST198091461892",
				RangkaNo:        "198091461892",
				CustomerSegment: constant.RO_AO_PRIORITY,
				OTRPrice:        17000000,
				DPAmount:        2000000,
			},
			customerData: konsumenOnlyAoPrio,
			code:         200,
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "AO PRIORITY Top Up - Confins DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            6.5,
			},
			installmentAmount:  260000,
			installmentConfins: 200000,
			income:             4000000,
			codeChassisNumber:  200,
			bodyChassisNumber: `{ "code": "OK", "message": "operasi berhasil dieksekusi.", "data": { "go_live_date": null, "id_number": "", 
			"installment_amount": 200000, "is_active": true, "is_registered": false, "lc_installment": 0, "legal_name": "", "outstanding_interest": 0, 
			"outstanding_principal": 12000000, "status": "" }, "errors": null, "request_id": "e818eaf9-cc7b-40cb-b707-37ab3006ae5c", 
			"timestamp": "2023-11-02 06:22:17" }`,
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

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("INSTALLMENT_PENDING_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, true, 2, 60, tc.req.ProspectID, accessToken).Return(resp, tc.errResp).Once()

			if len(tc.customerData) > 1 {
				mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("INSTALLMENT_PENDING_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, true, 2, 60, tc.req.ProspectID, accessToken).Return(resp, tc.errResp).Once()
			}

			rst = resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"), httpmock.NewStringResponder(tc.codeChassisNumber, tc.bodyChassisNumber))
			resp, _ = rst.R().Get(os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+tc.req.RangkaNo, mock.Anything, mock.Anything, constant.METHOD_GET, true, 2, 60, tc.req.ProspectID, accessToken).Return(resp, tc.errRespChassisNumber).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, _, _, _, _, err := usecase.DsrCheck(ctx, tc.req, tc.customerData, tc.installmentAmount, tc.installmentConfins, tc.installmentConfinsSpouse, tc.income, accessToken, configValue)
			require.Equal(t, tc.result, data)
			// require.Equal(t, tc.result, dsr)
			require.Equal(t, tc.errResult, err)
		})
	}

}

func TestTotalDsrFmfPbk(t *testing.T) {

	ctx := context.Background()
	os.Setenv("LASTEST_PAID_INSTALLMENT_URL", "http://10.9.100.231/los-int-dupcheck-v2/api/v2/mdm/installment/")

	// Get the current time
	currentTime := time.Now().UTC()

	// Sample older date from the current time to test "RrdDate"
	sevenMonthsAgo := currentTime.AddDate(0, -7, 0)
	sixMonthsAgo := currentTime.AddDate(0, -6, 0)

	testcases := []struct {
		name                                             string
		totalIncome, newInstallment, totalInstallmentPBK float64
		prospectID, customerSegment, accessToken         string
		SpDupcheckMap                                    response.SpDupcheckMap
		codeLatestInstallment                            int
		bodyLatestInstallment                            string
		errLatestInstallment                             error
		configValue                                      entity.AppConfig
		result                                           response.UsecaseApi
		trxFMF                                           response.TrxFMF
		filtering                                        entity.FilteringKMB
		mappingBranchDeviasi                             entity.MappingBranchDeviasi
		mappingDeviasiDSR                                entity.MasterMappingDeviasiDSR
		trxApk                                           entity.TrxApk
		errResult                                        error
		config                                           entity.AppConfig
		errGetConfig                                     error
	}{
		{
			name:                "TotalDsrFmfPbk reject",
			totalIncome:         100000000,
			newInstallment:      60000000,
			totalInstallmentPBK: 5000000,
			prospectID:          "TEST1",
			customerSegment:     "",
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			filtering: entity.FilteringKMB{
				BranchID: "400",
			},
			mappingBranchDeviasi: entity.MappingBranchDeviasi{
				BranchID:      "400",
				FinalApproval: "CBM",
			},
			mappingDeviasiDSR: entity.MasterMappingDeviasiDSR{
				DSRThreshold: 70,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				Dsr:            60,
				CustomerID:     "",
				ConfigMaxDSR:   35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_TOTAL_DSRGT35,
				Reason:         "NEW - DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				IsDeviasi:      true,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(5),
				TotalDSR: float64(65),
			},
		},
		{
			name:                "TotalDsrFmfPbk pass new",
			totalIncome:         7200000,
			newInstallment:      2240000,
			totalInstallmentPBK: 1000000,
			prospectID:          "TEST1",
			customerSegment:     "",
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":45,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				Dsr:            31.11111111111111,
				CustomerID:     "",
				ConfigMaxDSR:   45,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "NEW - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(13.88888888888889),
				TotalDSR: float64(45),
			},
		},
		{
			name:                "TotalDsrFmfPbk ro prime",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				RRDDate:        sixMonthsAgo,
				Dsr:            30,
				CustomerID:     "123456",
				ConfigMaxDSR:   35,
			},
			filtering: entity.FilteringKMB{
				CreatedAt: currentTime,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "RO PRIME",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:                  float64(0.5),
				TotalDSR:                float64(30),
				LatestInstallmentAmount: 1000000,
				InstallmentThreshold:    1500000,
			},
			codeLatestInstallment: 200,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 1000000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
		},
		{
			name:                "TotalDsrFmfPbk CR perbaikan flow RO PrimePriority PASS",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				RRDDate:                          sevenMonthsAgo,
				Dsr:                              30,
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
				CustomerID:                       "123456",
				ConfigMaxDSR:                     35,
			},
			filtering: entity.FilteringKMB{
				CreatedAt: currentTime,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35_EXP_CONTRACT_6MONTHS,
				Reason:         constant.EXPIRED_CONTRACT_HIGHERTHAN_6MONTHS + "RO PRIME - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30),
			},
			codeLatestInstallment: 200,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 1000000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
			config: entity.AppConfig{
				Key:   "expired_contract_check",
				Value: `{"data":{"expired_contract_check_enabled":true,"expired_contract_max_months":6}}`,
			},
		},
		{
			name:                "TotalDsrFmfPbk CR perbaikan flow RO PrimePriority RrdDate NULL",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				RRDDate:                          nil,
				Dsr:                              30,
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
				CustomerID:                       "123456",
				ConfigMaxDSR:                     35,
			},
			filtering: entity.FilteringKMB{
				CreatedAt: currentTime,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30.5),
			},
			codeLatestInstallment: 500,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 1000000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
			errLatestInstallment: errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty"),
			errResult:            errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty"),
		},
		{
			name:                "TotalDsrFmfPbk ro prime < installmentThreshold",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				RRDDate:        sixMonthsAgo,
				Dsr:            30,
				CustomerID:     "123456",
				ConfigMaxDSR:   35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "RO PRIME - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:                  float64(0.5),
				TotalDSR:                float64(30),
				LatestInstallmentAmount: 10000,
				InstallmentThreshold:    15000,
			},
			codeLatestInstallment: 200,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 10000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "` + sixMonthsAgo.Format(time.RFC3339) + `" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
		},
		{
			name:                "TotalDsrFmfPbk ro priority",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIORITY,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				RRDDate:        sixMonthsAgo,
				Dsr:            30,
				CustomerID:     "123456",
				ConfigMaxDSR:   35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "RO PRIORITY - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30),
			},
		},
		{
			name:                "TotalDsrFmfPbk ao top up prime",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:   constant.STATUS_KONSUMEN_AO,
				Dsr:              30,
				CustomerID:       "123456",
				InstallmentTopup: 1000000,
				ConfigMaxDSR:     35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "AO PRIME Top Up",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:                  float64(0.5),
				TotalDSR:                float64(30),
				LatestInstallmentAmount: 1000000,
				InstallmentThreshold:    1500000,
			},
			codeLatestInstallment: 200,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 1000000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
		},
		{
			name:                "TotalDsrFmfPbk ao top up prime < installmentThreshold",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:   constant.STATUS_KONSUMEN_AO,
				Dsr:              30,
				CustomerID:       "123456",
				InstallmentTopup: 100000,
				ConfigMaxDSR:     35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "AO PRIME Top Up - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:                  float64(0.5),
				TotalDSR:                float64(30),
				LatestInstallmentAmount: 100000,
				InstallmentThreshold:    150000,
			},
			codeLatestInstallment: 200,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 100000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
		},
		{
			name:                "TotalDsrFmfPbk ao prime NumberOfPaidInstallment >= 6",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:          constant.STATUS_KONSUMEN_AO,
				Dsr:                     30,
				CustomerID:              "123456",
				NumberOfPaidInstallment: 7,
				ConfigMaxDSR:            35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "AO PRIME - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30),
			},
		},
		{
			name:                "TotalDsrFmfPbk ao top up priority < installmentThreshold",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIORITY,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:   constant.STATUS_KONSUMEN_AO,
				Dsr:              30,
				CustomerID:       "123456",
				InstallmentTopup: 100000,
				ConfigMaxDSR:     35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "AO PRIORITY Top Up - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30),
			},
		},
		{
			name:                "TotalDsrFmfPbk ao priority NumberOfPaidInstallment >= 6",
			totalIncome:         7200000,
			newInstallment:      2240000,
			totalInstallmentPBK: 1000000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIORITY,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:          constant.STATUS_KONSUMEN_AO,
				Dsr:                     30,
				CustomerID:              "123456",
				NumberOfPaidInstallment: 7,
				ConfigMaxDSR:            35,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_TOTAL_DSRLTE35,
				Reason:         "AO PRIORITY - DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(13.88888888888889),
				TotalDSR: float64(30),
			},
		},
		{
			name:                "TotalDsrFmfPbk ro prime err",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				RRDDate:        sixMonthsAgo,
				Dsr:            30,
				CustomerID:     "123456",
				ConfigMaxDSR:   35,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30.5),
			},
			codeLatestInstallment: 500,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 1000000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
			errLatestInstallment: errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call LatestPaidInstallmentData Timeout"),
			errResult:            errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call LatestPaidInstallmentData Timeout"),
		},
		{
			name:                "TotalDsrFmfPbk ro prime err 500",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				RRDDate:        sixMonthsAgo,
				Dsr:            30,
				CustomerID:     "123456",
				ConfigMaxDSR:   35,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30.5),
			},
			codeLatestInstallment: 500,
			bodyLatestInstallment: `{ "messages": "LOS - Latest Installment", "errors": null, 
			"data": { "customer_id": "", "application_id": "", "agreement_no": "", "installment_amount": 1000000, "contract_status": "", "outstanding_principal": 0, 
			"rrd_date": "" }, "server_time": "2023-11-02T07:38:54+07:00", "request_id": "e187d3ce-5d4b-4b1a-b078-f0d1900df9dd" }`,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Call LatestPaidInstallmentData Error"),
		},
		{
			name:                "TotalDsrFmfPbk ro prime err unmarshal",
			totalIncome:         6000000,
			newInstallment:      200000,
			totalInstallmentPBK: 30000,
			prospectID:          "TEST1",
			customerSegment:     constant.RO_AO_PRIME,
			accessToken:         "token",
			trxApk: entity.TrxApk{
				NTF: 10000000,
			},
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				RRDDate:        sixMonthsAgo,
				Dsr:            30,
				CustomerID:     "123456",
				ConfigMaxDSR:   35,
			},
			trxFMF: response.TrxFMF{
				DSRPBK:   float64(0.5),
				TotalDSR: float64(30.5),
			},
			codeLatestInstallment: 200,
			bodyLatestInstallment: `response body unmarshal error`,
			errResult:             errors.New(constant.ERROR_UPSTREAM + " - Unmarshal LatestPaidInstallmentData Error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			dsrPBK := tc.totalInstallmentPBK / tc.totalIncome * 100

			totalDSR := tc.SpDupcheckMap.Dsr + dsrPBK

			log.Println(tc.name)
			log.Println(totalDSR)

			var configValue response.DupcheckConfig
			json.Unmarshal([]byte(tc.configValue.Value), &configValue)

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("LASTEST_PAID_INSTALLMENT_URL"), httpmock.NewStringResponder(tc.codeLatestInstallment, tc.bodyLatestInstallment))
			resp, _ := rst.R().Get(os.Getenv("LASTEST_PAID_INSTALLMENT_URL"))

			mockRepository.On("GetConfig", "expired_contract", "KMB-OFF", "expired_contract_check").Return(tc.config, tc.errGetConfig)

			mockRepository.On("GetBranchDeviasi", "400", tc.SpDupcheckMap.StatusKonsumen, tc.trxApk.NTF).Return(tc.mappingBranchDeviasi, tc.errGetConfig)

			mockRepository.On("MasterMappingDeviasiDSR", tc.totalIncome).Return(tc.mappingDeviasiDSR, tc.errGetConfig)

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("LASTEST_PAID_INSTALLMENT_URL")+tc.SpDupcheckMap.CustomerID.(string)+"/2", mock.Anything, mock.Anything, constant.METHOD_GET, false, 0, 30, tc.prospectID, tc.accessToken).Return(resp, tc.errLatestInstallment).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, trx, err := usecase.TotalDsrFmfPbk(ctx, tc.totalIncome, tc.newInstallment, tc.totalInstallmentPBK, tc.prospectID, tc.customerSegment, tc.accessToken, tc.SpDupcheckMap, configValue, tc.filtering, tc.trxApk.NTF)

			require.Equal(t, tc.result, data)
			require.Equal(t, tc.trxFMF, trx)
			require.Equal(t, tc.errResult, err)
		})
	}
}
