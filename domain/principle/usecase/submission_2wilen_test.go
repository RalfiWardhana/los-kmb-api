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
	"os"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckAgreementChassisNumber(t *testing.T) {
	os.Setenv("AGREEMENT_OF_CHASSIS_NUMBER_URL", "http://localhost/")

	testcases := []struct {
		name           string
		prospectID     string
		chassisNumber  string
		idNumber       string
		spouseIDNumber string
		code           int
		body           string
		errResp        error
		expectedErr    error
		expectedResult response.UsecaseApi
	}{
		{
			name:          "error EngineAPI",
			prospectID:    "TEST198091461892",
			chassisNumber: "198091461892",
			errResp:       errors.New("Get Error"),
			expectedErr:   errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Get Agreement of Chassis Number Timeout"),
		},
		{
			name:          "EngineAPI status != 200",
			prospectID:    "TEST198091461892",
			chassisNumber: "198091461892",
			code:          502,
			body:          `{"code":"FAIL"}`,
			expectedErr:   errors.New(constant.ERROR_UPSTREAM + " - Call Get Agreement of Chassis Number Error"),
		},
		{
			name:          "Unmarshal error",
			prospectID:    "TEST198091461892",
			chassisNumber: "198091461892",
			code:          200,
			body:          `{"data": {invalid}}`,
			expectedErr:   errors.New(constant.ERROR_UPSTREAM + " - Unmarshal Get Agreement of Chassis Number Error"),
		},
		{
			name:          "Agreement not found",
			prospectID:    "TEST198091461892",
			chassisNumber: "198091461892",
			code:          200,
			body:          `{"data":{"id_number":"","is_active":false,"is_registered":false}}`,
			expectedResult: response.UsecaseApi{
				Code:           constant.CODE_AGREEMENT_NOT_FOUND,
				Result:         constant.DECISION_PASS,
				Reason:         constant.REASON_AGREEMENT_NOT_FOUND,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
		{
			name:          "Consumer match",
			prospectID:    "TEST198091461892",
			chassisNumber: "198091461892",
			idNumber:      "7612379",
			code:          200,
			body:          `{"data":{"id_number":"7612379","is_active":true,"is_registered":true}}`,
			expectedResult: response.UsecaseApi{
				Code:           constant.CODE_OK_CONSUMEN_MATCH,
				Result:         constant.DECISION_PASS,
				Reason:         constant.REASON_OK_CONSUMEN_MATCH,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
		{
			name:          "Reject chassis number",
			prospectID:    "TEST198091461892",
			chassisNumber: "198091461892",
			idNumber:      "7612379",
			code:          200,
			body:          `{"data":{"id_number":"161234339","is_active":true,"is_registered":true}}`,
			expectedResult: response.UsecaseApi{
				Code:           constant.CODE_REJECT_CHASSIS_NUMBER,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_CHASSIS_NUMBER,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
		{
			name:           "Reject fraud potential",
			prospectID:     "TEST198091461892",
			chassisNumber:  "198091461892",
			idNumber:       "7612379",
			spouseIDNumber: "161234339",
			code:           200,
			body:           `{"data":{"id_number":"161234339","is_active":true,"is_registered":true}}`,
			expectedResult: response.UsecaseApi{
				Code:           constant.CODE_REJECTION_FRAUD_POTENTIAL,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECTION_FRAUD_POTENTIAL,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			accessToken := "access-token"

			mockRepo := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL") + tc.chassisNumber
			httpmock.RegisterResponder(constant.METHOD_GET, url, httpmock.NewStringResponder(tc.code, tc.body))

			resp, _ := rst.R().Get(url)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, []byte(nil), mock.Anything, constant.METHOD_GET, true, 6, 60, tc.prospectID, accessToken).Return(resp, tc.errResp).Once()

			uc := NewUsecase(mockRepo, mockHttpClient, nil)
			_, data, err := uc.CheckAgreementChassisNumber(ctx, tc.prospectID, tc.chassisNumber, tc.idNumber, tc.spouseIDNumber, accessToken)

			assert.Equal(t, tc.expectedResult, data)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}

func TestCheckBannedChassisNumber(t *testing.T) {
	testcases := []struct {
		name           string
		chassisNo      string
		expectedResult response.UsecaseApi
		expectedErr    error
		mockErr        error
		mockResult     entity.TrxBannedChassisNumber
	}{
		{
			name:        "error on repository call",
			chassisNo:   "198091461892",
			mockErr:     errors.New("Get Trx Error"),
			expectedErr: errors.New(constant.ERROR_UPSTREAM + " - Get Banned Chassis Number Error"),
		},
		{
			name:      "chassis number is banned",
			chassisNo: "198091461892",
			mockResult: entity.TrxBannedChassisNumber{
				ProspectID: "SLY-123457678",
				ChassisNo:  "198091461892",
			},
			expectedResult: response.UsecaseApi{
				Code:           constant.CODE_REJECT_NOKA_NOSIN,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_NOKA_NOSIN,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
		{
			name:           "chassis number is not banned",
			chassisNo:      "111122223333",
			mockResult:     entity.TrxBannedChassisNumber{},
			expectedResult: response.UsecaseApi{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockHttp := new(httpclient.MockHttpClient)

			mockRepo.On("GetBannedChassisNumber", tc.chassisNo).Return(tc.mockResult, tc.mockErr).Once()

			uc := NewUsecase(mockRepo, mockHttp, nil)
			result, err := uc.CheckBannedChassisNumber(tc.chassisNo)

			assert.Equal(t, tc.expectedResult, result)

			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestNegativeCustomerCheck(t *testing.T) {
	os.Setenv("API_NEGATIVE_CUSTOMER", "http://localhost/")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	accessToken := "token"
	header := map[string]string{
		"Authorization": accessToken,
	}

	testcases := []struct {
		name                    string
		respBody                string
		result                  response.UsecaseApi
		respNegativeCustomer    response.NegativeCustomer
		negativeCustomer        response.NegativeCustomer
		errResp                 error
		errResult               error
		mappingNegativeCustomer entity.MappingNegativeCustomer
		errRepo                 error
		req                     request.DupcheckApi
	}{
		{
			name: "NegativeCustomerCheck error EngineAPI",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
			},
			errResp:   errors.New("Get Error"),
			errResult: errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call API Negative Customer Error"),
		},
		{
			name: "NegativeCustomerCheck GetMappingNegativeCustomer error",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
			},
			respBody: `{
				"code": "OK",
				"message": "operasi berhasil dieksekusi.",
				"data": {
					"is_active":1,
					"is_blacklist":1,
					"is_highrisk":1,
					"bad_type":"B",
					"result":"BLACKLIST APU-PPT"
				},
				"errors": null,
				"request_id": "c240772b-5f78-489b-bdb4-6ed796dadaf6",
				"timestamp": "2023-03-26 21:29:07"
			}`,
			errRepo: errors.New("Get Error"),
			negativeCustomer: response.NegativeCustomer{
				IsActive:    1,
				IsBlacklist: 1,
				IsHighrisk:  1,
				BadType:     "B",
				Result:      "BLACKLIST APU-PPT",
				Decision:    "",
			},
			errResult: errors.New(constant.ERROR_UPSTREAM + " - GetMappingNegativeCustomer Error - Get Error"),
		},
		{
			name: "NegativeCustomerCheck reject",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
			},
			respBody: `{
				"code": "OK",
				"message": "operasi berhasil dieksekusi.",
				"data": {
					"is_active":1,
					"is_blacklist":1,
					"is_highrisk":1,
					"bad_type":"B",
					"result":"BLACKLIST APU-PPT"
				},
				"errors": null,
				"request_id": "c240772b-5f78-489b-bdb4-6ed796dadaf6",
				"timestamp": "2023-03-26 21:29:07"
			}`,
			mappingNegativeCustomer: entity.MappingNegativeCustomer{
				Decision: constant.DECISION_REJECT,
				Reason:   "BLACKLIST APU-PPT",
			},
			negativeCustomer: response.NegativeCustomer{
				IsActive:    1,
				IsBlacklist: 1,
				IsHighrisk:  1,
				BadType:     "B",
				Result:      "BLACKLIST APU-PPT",
				Decision:    "REJECT",
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_NEGATIVE_CUSTOMER,
				Reason:         "BLACKLIST APU-PPT",
				Result:         constant.DECISION_REJECT,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
				Info:           "{\"is_active\":1,\"is_blacklist\":1,\"is_highrisk\":1,\"bad_type\":\"B\",\"result\":\"BLACKLIST APU-PPT\",\"decision\":\"REJECT\"}",
			},
		},
		{
			name: "NegativeCustomerCheck pass",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
			},
			respBody: `{
				"code": "OK",
				"message": "operasi berhasil dieksekusi.",
				"data": {
					"is_active":1,
					"is_blacklist":0,
					"is_highrisk":1,
					"bad_type":"",
					"result":"HIGHRISK APU-PPT"
				},
				"errors": null,
				"request_id": "c240772b-5f78-489b-bdb4-6ed796dadaf6",
				"timestamp": "2023-03-26 21:29:07"
			}`,
			mappingNegativeCustomer: entity.MappingNegativeCustomer{
				Decision: "YES",
				Reason:   "HIGHRISK APU-PPT",
			},
			negativeCustomer: response.NegativeCustomer{
				IsActive:    1,
				IsBlacklist: 0,
				IsHighrisk:  1,
				BadType:     "0",
				Result:      "HIGHRISK APU-PPT",
				Decision:    "YES",
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_NEGATIVE_CUSTOMER,
				Reason:         constant.REASON_NON_BLACKLIST,
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
				Info:           "{\"is_active\":1,\"is_blacklist\":0,\"is_highrisk\":1,\"bad_type\":\"0\",\"result\":\"HIGHRISK APU-PPT\",\"decision\":\"YES\"}",
			},
		},
		{
			name: "NegativeCustomerCheck pass no data",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
			},
			respBody: `{
				"code": "OK",
				"message": "operasi berhasil dieksekusi.",
				"data": {
					"is_active":0,
					"is_blacklist":0,
					"is_highrisk":0,
					"bad_type":"",
					"result":""
				},
				"errors": null,
				"request_id": "c240772b-5f78-489b-bdb4-6ed796dadaf6",
				"timestamp": "2023-03-26 21:29:07"
			}`,
			result: response.UsecaseApi{
				Code:           constant.CODE_NEGATIVE_CUSTOMER,
				Reason:         constant.REASON_NON_BLACKLIST,
				Result:         constant.DECISION_PASS,
				SourceDecision: constant.SOURCE_DECISION_BLACKLIST,
				Info:           "{\"is_active\":0,\"is_blacklist\":0,\"is_highrisk\":0,\"bad_type\":\"\",\"result\":\"\",\"decision\":\"\"}",
			},
		},
		{
			name: "NegativeCustomerCheck error unmarshal",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
			},
			respBody: `{
				"code": "OK",
				"message": "operasi berhasil dieksekusi.",
				"data": "this should be an object, not a string",
				"errors": null,
				"request_id": "c240772b-5f78-489b-bdb4-6ed796dadaf6",
				"timestamp": "2023-03-26 21:29:07"
			}`,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data response negative customer"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			req, _ := json.Marshal(request.NegativeCustomer{
				ProspectID:        tc.req.ProspectID,
				IDNumber:          tc.req.IDNumber,
				LegalName:         tc.req.LegalName,
				BirthDate:         tc.req.BirthDate,
				SurgateMotherName: tc.req.MotherName,
				ProfessionID:      tc.req.ProfessionID,
				JobType:           tc.req.JobType,
				JobPosition:       tc.req.JobPosition,
			})

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("API_NEGATIVE_CUSTOMER"), httpmock.NewStringResponder(200, tc.respBody))
			resp, _ := rst.R().Post(os.Getenv("API_NEGATIVE_CUSTOMER"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("API_NEGATIVE_CUSTOMER"), req, header, constant.METHOD_POST, true, 6, timeout, tc.req.ProspectID, accessToken).Return(resp, tc.errResp).Once()
			mockRepository.On("GetMappingNegativeCustomer", mock.Anything).Return(tc.mappingNegativeCustomer, tc.errRepo)

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			result, negativeCustomer, err := usecase.NegativeCustomerCheck(ctx, tc.req, accessToken)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.negativeCustomer, negativeCustomer)
			require.Equal(t, tc.errResult, err)
		})
	}
}

func TestCheckMobilePhoneFMF(t *testing.T) {
	os.Setenv("HRIS_LIST_EMPLOYEE", "http://localhost/hris-list-employee")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	accessToken := "token"
	hrisAccessToken := "token"

	testcases := []struct {
		name        string
		prospectID  string
		mobilePhone string
		idNumber    string
		httpCode    int
		httpBody    string
		httpErr     error
		expected    response.UsecaseApi
		expectedErr error
	}{
		{
			name:        "EngineAPI HRIS_LIST_EMPLOYEE error",
			prospectID:  "SAL-1234567",
			mobilePhone: "081234567",
			httpCode:    500,
			httpErr:     errors.New(constant.ERROR_UPSTREAM + " - Call API HRIS List Employee Error"),
			expectedErr: errors.New(constant.ERROR_UPSTREAM + " - Call API HRIS List Employee Error"),
		},
		{
			name:        "Reject when phone matches and NIK differs and is_resign is false",
			prospectID:  "SAL-1234567",
			mobilePhone: "08161970587",
			idNumber:    "3173063101700003",
			httpCode:    200,
			httpBody: `{
				"data": [{
					"id_number": "3173063101700002",
					"phone_number": "08161970587",
					"is_resign": false
				}]
			}`,
			expected: response.UsecaseApi{
				SourceDecision: constant.SOURCE_DECISION_NOHP,
				Code:           constant.CODE_NOHP,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_NOHP,
			},
		},
		{
			name:        "Pass when phone doesn't match",
			prospectID:  "SAL-1234567",
			mobilePhone: "081234567",
			idNumber:    "3173063101700002",
			httpCode:    200,
			httpBody: `{
				"data": [{
					"id_number": "3173063101700002",
					"phone_number": "08161970587",
					"is_resign": false
				}]
			}`,
			expected: response.UsecaseApi{
				SourceDecision: constant.SOURCE_DECISION_NOHP,
				Code:           constant.CODE_NOHP,
				Result:         constant.DECISION_PASS,
			},
		},
		{
			name:        "Unmarshal error on list employee data",
			prospectID:  "SAL-1234567",
			mobilePhone: "081234567",
			idNumber:    "3173063101700002",
			httpCode:    200,
			httpBody: `{
				"data": "invalid_data_format"
			}`,
			expected: response.UsecaseApi{
				Info: nil,
			},
			expectedErr: errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data response list employee"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(mocks.Repository)
			mockHttp := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("HRIS_LIST_EMPLOYEE"),
				httpmock.NewStringResponder(tc.httpCode, tc.httpBody))

			resp, _ := rst.R().Post(os.Getenv("HRIS_LIST_EMPLOYEE"))

			ctx := context.Background()
			mockHttp.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("HRIS_LIST_EMPLOYEE"),
				mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, timeout, "", accessToken).
				Return(resp, tc.httpErr).Once()

			if tc.httpErr == nil && tc.httpBody != "" && tc.name != "Unmarshal error on list employee data" {
				var list []response.HrisListEmployee
				_ = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &list)
				info, _ := json.Marshal(list)
				tc.expected.Info = string(info)
			}

			usecase := NewUsecase(mockRepo, mockHttp, nil)

			result, err := usecase.CheckMobilePhoneFMF(ctx, tc.prospectID, tc.mobilePhone, tc.idNumber, accessToken, hrisAccessToken)

			require.Equal(t, tc.expected, result)
			if tc.expectedErr != nil {
				require.EqualError(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}

			mockHttp.AssertExpectations(t)
		})
	}
}
