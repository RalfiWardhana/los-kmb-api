package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCheckBannedPMKDSR(t *testing.T) {
	testcases := []struct {
		name              string
		idNumber          string
		result            response.UsecaseApi
		errResult         error
		errEnc            error
		encryptedIDNumber entity.EncryptedString
		errGetTrx         error
		TrxBannedPMKDSR   entity.TrxBannedPMKDSR
	}{
		{
			name:              "CheckBannedPMKDSR error encrypt",
			idNumber:          "198091461892",
			errEnc:            errors.New("Encrypt Error"),
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - GetEncB64 ID Number Error"),
		},
		{
			name:              "CheckBannedPMKDSR error Trx",
			idNumber:          "198091461892",
			errGetTrx:         errors.New("Get Trx Error"),
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - Get Banned PMK DSR Error"),
		},
		{
			name:              "CheckBannedPMKDSR reject",
			idNumber:          "198091461892",
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			result: response.UsecaseApi{
				Code:           constant.CODE_PERNAH_REJECT_PMK_DSR,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_PERNAH_REJECT_PMK_DSR,
				SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
			},
			TrxBannedPMKDSR: entity.TrxBannedPMKDSR{
				ProspectID: "SLY-123457678",
				IDNumber:   "123456",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetEncB64", tc.idNumber).Return(tc.encryptedIDNumber, tc.errEnc)
			mockRepository.On("GetBannedPMKDSR", tc.encryptedIDNumber.MyString).Return(tc.TrxBannedPMKDSR, tc.errGetTrx)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.CheckBannedPMKDSR(tc.idNumber)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errResult, err)
		})
	}

}

func TestCheckRejection(t *testing.T) {
	testcases := []struct {
		name              string
		idNumber          string
		result            response.UsecaseApi
		errResult         error
		errEnc            error
		encryptedIDNumber entity.EncryptedString
		errGetTrx         error
		trxReject         entity.TrxReject
		TrxBannedPMKDSR   entity.TrxBannedPMKDSR
	}{
		{
			name:              "CheckRejection error encrypt",
			idNumber:          "198091461892",
			errEnc:            errors.New("Encrypt Error"),
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - GetEncB64 ID Number Error"),
		},
		{
			name:              "CheckRejection error Trx",
			idNumber:          "198091461892",
			errGetTrx:         errors.New("Get Trx Error"),
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - Get Trx Reject Error"),
		},
		{
			name:              "CheckRejection REASON_PERNAH_REJECT_PMK_DSR",
			idNumber:          "198091461892",
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			trxReject:         entity.TrxReject{RejectPMKDSR: 3},
			result: response.UsecaseApi{
				Code:           constant.CODE_PERNAH_REJECT_PMK_DSR,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_PERNAH_REJECT_PMK_DSR,
				SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
			},
			TrxBannedPMKDSR: entity.TrxBannedPMKDSR{
				ProspectID: "SLY-123457678",
				IDNumber:   "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI=",
			},
		},
		{
			name:              "CheckRejection REASON_PERNAH_REJECT_NIK",
			idNumber:          "198091461892",
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			trxReject:         entity.TrxReject{RejectNIK: 3},
			result: response.UsecaseApi{
				Code:           constant.CODE_PERNAH_REJECT_NIK,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_PERNAH_REJECT_NIK,
				SourceDecision: constant.SOURCE_DECISION_NIK,
			},
		},
		{
			name:              "CheckRejection REASON_BELUM_PERNAH_REJECT",
			idNumber:          "198091461892",
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			result: response.UsecaseApi{
				Code:           constant.CODE_BELUM_PERNAH_REJECT,
				Result:         constant.DECISION_PASS,
				Reason:         constant.REASON_BELUM_PERNAH_REJECT,
				SourceDecision: constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			prospectID := "SLY-123457678"
			configValue := response.DupcheckConfig{
				Data: response.DataDupcheckConfig{AttemptPMKDSR: 2},
			}
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetEncB64", tc.idNumber).Return(tc.encryptedIDNumber, tc.errEnc)
			mockRepository.On("GetCurrentTrxWithReject", tc.encryptedIDNumber.MyString).Return(tc.trxReject, tc.errGetTrx)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, trxBanned, err := usecase.CheckRejection(tc.idNumber, prospectID, configValue)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.TrxBannedPMKDSR, trxBanned)
			require.Equal(t, tc.errResult, err)
		})
	}

}

func TestCheckBannedChassisNumber(t *testing.T) {
	testcases := []struct {
		name                   string
		ChassisNo              string
		result                 response.UsecaseApi
		errResult              error
		errGetTrx              error
		TrxBannedChassisNumber entity.TrxBannedChassisNumber
	}{
		{
			name:      "CheckBannedChassisNumber error Trx",
			ChassisNo: "198091461892",
			errGetTrx: errors.New("Get Trx Error"),
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Get Banned Chassis Number Error"),
		},
		{
			name:      "CheckBannedChassisNumber reject",
			ChassisNo: "198091461892",
			result: response.UsecaseApi{
				Code:           constant.CODE_REJECT_NOKA_NOSIN,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_NOKA_NOSIN,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
			TrxBannedChassisNumber: entity.TrxBannedChassisNumber{
				ProspectID: "SLY-123457678",
				ChassisNo:  "198091461892",
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetBannedChassisNumber", tc.ChassisNo).Return(tc.TrxBannedChassisNumber, tc.errGetTrx)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.CheckBannedChassisNumber(tc.ChassisNo)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errResult, err)
		})
	}

}

func TestCheckRejectChassisNumber(t *testing.T) {
	testcases := []struct {
		name                   string
		result                 response.UsecaseApi
		errResult              error
		errGetTrx              error
		RejectChassisNumber    []entity.RejectChassisNumber
		req                    request.DupcheckApi
		trxBannedChassisNumber entity.TrxBannedChassisNumber
	}{
		{
			name: "CheckRejectChassisNumber error Trx",
			req: request.DupcheckApi{
				RangkaNo: "198091461892",
			},
			errGetTrx: errors.New("Get Trx Error"),
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Get Reject Chassis Number Error"),
		},
		{
			name: "CheckRejectChassisNumber pass",
			req: request.DupcheckApi{
				RangkaNo: "198091461892",
			},
		},
		{
			name: "CheckRejectChassisNumber reject",
			req: request.DupcheckApi{
				ProspectID: "TEST12346",
				RangkaNo:   "198091461892",
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_REJECT_NOKA_NOSIN,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_NOKA_NOSIN,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
			RejectChassisNumber: []entity.RejectChassisNumber{{ProspectID: "123456"}},
		},
		{
			name: "CheckRejectChassisNumber reject tanpa perubahan data",
			req: request.DupcheckApi{
				ProspectID: "TEST12346",
				RangkaNo:   "198091461892",
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_REJECT_NOKA_NOSIN,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_NOKA_NOSIN,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
			RejectChassisNumber: []entity.RejectChassisNumber{{ProspectID: "TEST1"}, {ProspectID: "TEST2", ChassisNo: "198091461892"}},
		},
		{
			name: "CheckRejectChassisNumber banned",
			req: request.DupcheckApi{
				ProspectID: "TEST12346",
				RangkaNo:   "198091461892",
			},
			result: response.UsecaseApi{
				Code:           constant.CODE_REJECT_NOKA_NOSIN,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_NOKA_NOSIN,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
			RejectChassisNumber:    []entity.RejectChassisNumber{{ProspectID: "TEST1"}, {ProspectID: "TEST2"}, {ProspectID: "TEST3"}},
			trxBannedChassisNumber: entity.TrxBannedChassisNumber{ProspectID: "TEST12346", ChassisNo: "198091461892"},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			configValue := response.DupcheckConfig{
				Data: response.DataDupcheckConfig{AttemptChassisNumber: 3},
			}
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetCurrentTrxWithRejectChassisNumber", tc.req.RangkaNo).Return(tc.RejectChassisNumber, tc.errGetTrx)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, banned, err := usecase.CheckRejectChassisNumber(tc.req, configValue)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.trxBannedChassisNumber, banned)
			require.Equal(t, tc.errResult, err)
		})
	}

}

func TestCheckAgreementChassisNumber(t *testing.T) {
	// always set the valid url
	os.Setenv("AGREEMENT_OF_CHASSIS_NUMBER_URL", "http://localhost/")

	testcases := []struct {
		name                           string
		body                           string
		code                           int
		result                         response.UsecaseApi
		errResp                        error
		errResult                      error
		errGetTrx                      error
		req                            request.DupcheckApi
		responseAgreementChassisNumber response.AgreementChassisNumber
	}{
		{
			name: "CheckAgreementChassisNumber error EngineAPI",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			errResp:   errors.New("Get Error"),
			errResult: errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Get Agreement of Chassis Number Timeout"),
		},
		{
			name: "CheckAgreementChassisNumber error EngineAPI != 200",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			code:      502,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Call Get Agreement of Chassis Number Error"),
		},
		{
			name: "CheckAgreementChassisNumber error Unmarshal",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			code:      200,
			errResult: errors.New(constant.ERROR_UPSTREAM + " - Unmarshal Get Agreement of Chassis Number Error"),
			body:      `{"code"}`,
		},
		{
			name: "CheckAgreementChassisNumber REASON_AGREEMENT_NOT_FOUND",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
			},
			code: 200,
			body: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"","installment_amount":0,"is_active":false,"is_registered":false,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
				Code:           constant.CODE_AGREEMENT_NOT_FOUND,
				Result:         constant.DECISION_PASS,
				Reason:         constant.REASON_AGREEMENT_NOT_FOUND,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
		{
			name: "CheckAgreementChassisNumber REASON_OK_CONSUMEN_MATCH",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
				IDNumber:   "7612379",
			},
			code: 200,
			body: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"7612379","installment_amount":0,"is_active":true,"is_registered":true,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
				Code:           constant.CODE_OK_CONSUMEN_MATCH,
				Result:         constant.DECISION_PASS,
				Reason:         constant.REASON_OK_CONSUMEN_MATCH,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
		{
			name: "CheckAgreementChassisNumber REASON_REJECT_CHASSIS_NUMBER",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
				IDNumber:   "7612379",
			},
			code: 200,
			body: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"161234339","installment_amount":0,"is_active":true,"is_registered":true,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
				Code:           constant.CODE_REJECT_CHASSIS_NUMBER,
				Result:         constant.DECISION_REJECT,
				Reason:         constant.REASON_REJECT_CHASSIS_NUMBER,
				SourceDecision: constant.SOURCE_DECISION_NOKANOSIN,
			},
		},
		{
			name: "CheckAgreementChassisNumber REASON_REJECTION_FRAUD_POTENTIAL",
			req: request.DupcheckApi{
				ProspectID: "TEST198091461892",
				RangkaNo:   "198091461892",
				IDNumber:   "7612379",
				Spouse:     &request.DupcheckApiSpouse{IDNumber: "161234339"},
			},
			code: 200,
			body: `{"code":"OK","message":"operasi berhasil dieksekusi.","data":
			{"go_live_date":null,"id_number":"161234339","installment_amount":0,"is_active":true,"is_registered":true,
			"lc_installment":0,"legal_name":"","outstanding_interest":0,"outstanding_principal":0,"status":""},
			"errors":null,"request_id":"230e9356-ef14-45bd-a41f-96e98c16b5fb","timestamp":"2023-10-10 11:48:50"}`,
			result: response.UsecaseApi{
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
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+tc.req.RangkaNo, httpmock.NewStringResponder(tc.code, tc.body))
			resp, _ := rst.R().Get(os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL") + tc.req.RangkaNo)

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+tc.req.RangkaNo, []byte(nil), map[string]string{}, constant.METHOD_GET, true, 6, 60, tc.req.ProspectID, accessToken).Return(resp, tc.errResp).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.CheckAgreementChassisNumber(ctx, tc.req, accessToken)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errResult, err)
		})
	}

}

func TestBlacklistCheck(t *testing.T) {

	t.Parallel()

	sisaJumlahAngsuran := 1

	testcase := []struct {
		spDupcheck   response.SpDupCekCustomerByID
		customerType string
		expected     response.UsecaseApi
		index        int
		label        string
	}{
		{
			spDupcheck:   response.SpDupCekCustomerByID{},
			customerType: constant.MESSAGE_BERSIH,
			expected: response.UsecaseApi{
				Result: constant.DECISION_PASS, Code: constant.CODE_NON_BLACKLIST_ALL, StatusKonsumen: constant.STATUS_KONSUMEN_NEW, Reason: constant.REASON_NON_BLACKLIST,
			},
			index: 0,
			label: "TEST_BLACKLIST_KONSUMEN BERSIH",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", BadType: constant.BADTYPE_B, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_KONSUMEN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_KONSUMEN_BLACKLIST,
			},
			index: 0,
			label: "TEST_RO_BLACKLIST_KONSUMEN_TYPE_B",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", BadType: constant.BADTYPE_B, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_PASANGAN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_PASANGAN_BLACKLIST,
			},
			index: 1,
			label: "TEST_RO_BLACKLIST_PASANGAN_TYPE_B",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", BadType: constant.BADTYPE_W, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_WARNING,
			expected: response.UsecaseApi{
				Result: constant.DECISION_PASS, Code: constant.CODE_NON_BLACKLIST_ALL, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_NON_BLACKLIST,
			},
			index: 0,
			label: "TEST_RO_BLACKLIST_TYPE_W",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", MaxOverdueDays: 92, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_KONSUMEN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_KONSUMEN_BLACKLIST_OVD_90DAYS,
			},
			index: 0,
			label: "TEST_RO_BLACKLIST_KONSUMEN_MAX_OVERDUE",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", MaxOverdueDays: 92, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_PASANGAN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_PASANGAN_BLACKLIST_OVD_90DAYS,
			},
			index: 1,
			label: "TEST_RO_BLACKLIST_PASANGAN_MAX_OVERDUE",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", NumOfAssetInventoried: 1, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_KONSUMEN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_KONSUMEN_BLACKLIST_ASSET_INVENTORY,
			},
			index: 0,
			label: "TEST_RO_BLACKLIST_KONSUMEN_ASSET_INVENTORY",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", NumOfAssetInventoried: 1, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_PASANGAN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_PASANGAN_BLACKLIST_ASSET_INVENTORY,
			},
			index: 1,
			label: "TEST_RO_BLACKLIST_PASANGAN_ASSET_INVENTORY",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", IsRestructure: 1, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_KONSUMEN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_KONSUMEN_BLACKLIST_RESTRUCTURE,
			},
			index: 0,
			label: "TEST_RO_BLACKLIST_KONSUMEN_RESTRUCTURE",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				RRDDate: "2021-03-29", IsRestructure: 1, TotalInstallment: 0,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_PASANGAN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_RO, Reason: constant.REASON_PASANGAN_BLACKLIST_RESTRUCTURE,
			},
			index: 1,
			label: "TEST_RO_BLACKLIST_PASANGAN_RESTRUCTURE",
		},
		{
			spDupcheck: response.SpDupCekCustomerByID{
				NumberOfPaidInstallment: &sisaJumlahAngsuran, IsRestructure: 1, TotalInstallment: 1,
			},
			customerType: constant.MESSAGE_BLACKLIST,
			expected: response.UsecaseApi{
				Result: constant.DECISION_REJECT, Code: constant.CODE_PASANGAN_BLACKLIST, StatusKonsumen: constant.STATUS_KONSUMEN_AO, Reason: constant.REASON_PASANGAN_BLACKLIST_RESTRUCTURE,
			},
			index: 1,
			label: "TEST_AO_BLACKLIST_PASANGAN_RESTRUCTURE",
		},
	}

	for _, test := range testcase {
		t.Run(test.label, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)
			mockRepository := new(mocks.Repository)

			service := NewUsecase(mockRepository, mockHttpClient)
			result, customerType := service.BlacklistCheck(test.index, test.spDupcheck)

			require.Equal(t, test.expected.Result, result.Result)
			require.Equal(t, test.expected.Code, result.Code)
			require.Equal(t, test.expected.StatusKonsumen, result.StatusKonsumen)
			require.Equal(t, test.expected.Reason, result.Reason)
			require.Equal(t, test.customerType, customerType)
		})
	}

}

func TestVehicleCheck(t *testing.T) {

	os.Setenv("NAMA_SAMA", "K,P")

	config := entity.AppConfig{
		Key:   "parameterize",
		Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
	}

	yearPass := time.Now().AddDate(-1, 0, 0).Format("2006")
	yearReject := time.Now().AddDate(-18, 0, 0).Format("2006")

	testcase := []struct {
		vehicle                 response.UsecaseApi
		err, errExpected        error
		dupcheckConfig          entity.AppConfig
		year                    string
		cmoCluster              string
		bpkbName                string
		tenor                   int
		resGetMappingVehicleAge entity.MappingVehicleAge
		errGetMappingVehicleAge error
		label                   string
	}{
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_VEHICLE_SESUAI,
				Reason: constant.REASON_VEHICLE_SESUAI,
			},
			year:  yearPass,
			label: "TEST_VEHICLE_PASS",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_VEHICLE_AGE_MAX,
				Reason: fmt.Sprintf("%s %d Tahun", constant.REASON_VEHICLE_AGE_MAX, 17),
			},
			year:  yearReject,
			label: "TEST_VEHICLE_REJECT",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_VEHICLE_SESUAI,
				Reason: constant.REASON_VEHICLE_SESUAI,
			},
			year:       time.Now().AddDate(-10, 0, 0).Format("2006"),
			cmoCluster: "Cluster C",
			bpkbName:   "KK",
			tenor:      12,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 11,
				VehicleAgeEnd:   12,
				Cluster:         "Cluster C",
				BPKBNameType:    0,
				TenorStart:      1,
				TenorEnd:        23,
				Decision:        constant.DECISION_PASS,
			},
			label: "test pass vehicle age 11-12 cluster A-C tenor <24",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_VEHICLE_SESUAI,
				Reason: constant.REASON_VEHICLE_SESUAI,
			},
			year:       time.Now().AddDate(-9, 0, 0).Format("2006"),
			cmoCluster: "Cluster A",
			bpkbName:   "K",
			tenor:      24,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 11,
				VehicleAgeEnd:   12,
				Cluster:         "Cluster A",
				BPKBNameType:    1,
				TenorStart:      24,
				TenorEnd:        36,
				Decision:        constant.DECISION_PASS,
			},
			label: "test pass vehicle age 11-12 cluster A-C tenor >=24",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_VEHICLE_AGE_MAX,
				Reason: fmt.Sprintf("%s Ketentuan", constant.REASON_VEHICLE_AGE_MAX),
			},
			year:       time.Now().AddDate(-8, 0, 0).Format("2006"),
			cmoCluster: "Cluster B",
			bpkbName:   "KK",
			tenor:      36,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 11,
				VehicleAgeEnd:   12,
				Cluster:         "Cluster B",
				BPKBNameType:    0,
				TenorStart:      24,
				TenorEnd:        36,
				Decision:        constant.DECISION_REJECT,
			},
			label: "test reject vehicle age 11-12 cluster A-C tenor >=24",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_VEHICLE_AGE_MAX,
				Reason: fmt.Sprintf("%s Ketentuan", constant.REASON_VEHICLE_AGE_MAX),
			},
			year:       time.Now().AddDate(-11, 0, 0).Format("2006"),
			cmoCluster: "Cluster D",
			bpkbName:   "K",
			tenor:      1,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 11,
				VehicleAgeEnd:   12,
				Cluster:         "Cluster D",
				BPKBNameType:    1,
				TenorStart:      1,
				TenorEnd:        23,
				Decision:        constant.DECISION_REJECT,
			},
			label: "test reject vehicle age 11-12 cluster D-F all tenor",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_VEHICLE_SESUAI,
				Reason: constant.REASON_VEHICLE_SESUAI,
			},
			year:       time.Now().AddDate(-12, 0, 0).Format("2006"),
			cmoCluster: "Cluster B",
			bpkbName:   "KK",
			tenor:      12,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 13,
				VehicleAgeEnd:   13,
				Cluster:         "Cluster B",
				BPKBNameType:    0,
				TenorStart:      1,
				TenorEnd:        23,
				Decision:        constant.DECISION_PASS,
			},
			label: "test pass vehicle age 13 cluster A-C tenor <24",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_VEHICLE_SESUAI,
				Reason: constant.REASON_VEHICLE_SESUAI,
			},
			year:       time.Now().AddDate(-11, 0, 0).Format("2006"),
			cmoCluster: "Cluster A",
			bpkbName:   "K",
			tenor:      24,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 13,
				VehicleAgeEnd:   13,
				Cluster:         "Cluster A",
				BPKBNameType:    1,
				TenorStart:      24,
				TenorEnd:        36,
				Decision:        constant.DECISION_PASS,
			},
			label: "test pass vehicle age 13 cluster A-C tenor >=24",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_VEHICLE_AGE_MAX,
				Reason: fmt.Sprintf("%s Ketentuan", constant.REASON_VEHICLE_AGE_MAX),
			},
			year:       time.Now().AddDate(-11, 0, 0).Format("2006"),
			cmoCluster: "Cluster C",
			bpkbName:   "KK",
			tenor:      24,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 13,
				VehicleAgeEnd:   13,
				Cluster:         "Cluster C",
				BPKBNameType:    0,
				TenorStart:      24,
				TenorEnd:        36,
				Decision:        constant.DECISION_REJECT,
			},
			label: "test reject vehicle age 13 cluster A-C tenor >=24",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_VEHICLE_AGE_MAX,
				Reason: fmt.Sprintf("%s Ketentuan", constant.REASON_VEHICLE_AGE_MAX),
			},
			year:       time.Now().AddDate(-13, 0, 0).Format("2006"),
			cmoCluster: "Cluster F",
			bpkbName:   "K",
			tenor:      1,
			resGetMappingVehicleAge: entity.MappingVehicleAge{
				VehicleAgeStart: 13,
				VehicleAgeEnd:   13,
				Cluster:         "Cluster F",
				BPKBNameType:    1,
				TenorStart:      1,
				TenorEnd:        23,
				Decision:        constant.DECISION_REJECT,
			},
			label: "test reject vehicle age 13 cluster D-F all tenor",
		},
		{
			dupcheckConfig:          config,
			year:                    time.Now().AddDate(-13, 0, 0).Format("2006"),
			cmoCluster:              "Cluster F",
			bpkbName:                "K",
			tenor:                   1,
			errGetMappingVehicleAge: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Vehicle Age Error"),
			errExpected:             errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Vehicle Age Error"),
			label:                   "test error get mapping vehicle age",
		},
		{
			dupcheckConfig: config,
			vehicle: response.UsecaseApi{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_VEHICLE_SESUAI,
				Reason: constant.REASON_VEHICLE_SESUAI,
			},
			year:                    time.Now().AddDate(-12, 0, 0).Format("2006"),
			cmoCluster:              "Cluster B",
			bpkbName:                "KK",
			tenor:                   12,
			resGetMappingVehicleAge: entity.MappingVehicleAge{},
			label:                   "test pass mapping empty",
		},
	}

	for _, test := range testcase {
		t.Run(test.label, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)
			mockRepository := new(mocks.Repository)

			var configValue response.DupcheckConfig
			json.Unmarshal([]byte(test.dupcheckConfig.Value), &configValue)

			mockRepository.On("GetMappingVehicleAge", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.resGetMappingVehicleAge, test.errGetMappingVehicleAge)

			service := NewUsecase(mockRepository, mockHttpClient)
			result, err := service.VehicleCheck(test.year, test.cmoCluster, test.bpkbName, test.tenor, configValue)

			require.Equal(t, test.errExpected, err)
			require.Equal(t, test.vehicle.Result, result.Result)
			require.Equal(t, test.vehicle.Code, result.Code)
			require.Equal(t, test.vehicle.Reason, result.Reason)
		})
	}

}

func TestCustomerKMB(t *testing.T) {
	var testcase = []struct {
		spDupcheck     response.SpDupCekCustomerByID
		expectedError  error
		statusKonsumen string
		label          string
	}{
		{
			statusKonsumen: constant.STATUS_KONSUMEN_NEW,
			spDupcheck:     response.SpDupCekCustomerByID{},
		},
		{
			statusKonsumen: constant.STATUS_KONSUMEN_RO,
			spDupcheck: response.SpDupCekCustomerByID{
				TotalInstallment: 0,
				RRDDate:          "2020-01-01",
			},
		},
		{
			statusKonsumen: constant.STATUS_KONSUMEN_AO,
			spDupcheck: response.SpDupCekCustomerByID{
				TotalInstallment: 2000000,
			},
		},
		{
			statusKonsumen: constant.STATUS_KONSUMEN_NEW,
			spDupcheck: response.SpDupCekCustomerByID{
				TotalInstallment: 0,
				RRDDate:          nil,
				CustomerID:       "123",
			},
		},
	}
	for _, test := range testcase {
		t.Run(test.label, func(t *testing.T) {
			mockHttpClient := new(httpclient.MockHttpClient)
			mockRepository := new(mocks.Repository)
			service := NewUsecase(mockRepository, mockHttpClient)

			statusKonsumen, err := service.CustomerKMB(test.spDupcheck)
			require.Equal(t, test.expectedError, err)
			require.Equal(t, test.statusKonsumen, statusKonsumen)
		})
	}
}

func TestDupcheckIntegrator(t *testing.T) {

	var (
		prospectID  string = "TEST-001"
		idNumber    string = "32030143096XXXX6"
		legalName   string = "SOE***E"
		birthDate   string = "1966-09-03"
		surgateName string = "IBU KANDUNG"
	)

	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	param, _ := json.Marshal(map[string]interface{}{
		"transaction_id":      prospectID,
		"id_number":           idNumber,
		"legal_name":          legalName,
		"birth_date":          birthDate,
		"surgate_mother_name": surgateName,
	})

	testcase := []struct {
		body        string
		err         error
		errExpected error
		code        int
		label       string
		legalName   string
	}{
		{
			body: `{"messages":"DUPCHECK-UNIT_TEST","errors":null,"data":{"customer_id":"4290003XXX4","id_number":"32030143096XXXX6","full_name":"SOE***E","birth_date":"1966-09-03T00:00:00Z","surgate_mother_name":"IBU KANDUNG",
			"birth_place":"JAKARTA","gender":"F","emergency_contact_address":"JALAN BARU","legal_address":"JALAN BARU","legal_kelurahan":"CIPANGERANG","legal_kecamatan":"CIMAHI UTARA",
			"legal_city":"CIMAHI","lagal_zipcode":"14**1","residence_address":"JALAN BARU","residence_kelurahan":"CIPAGERANG","residence_kecamatan":"CIMAHI UTARA","residence_city":"CIMAHI",
			"residence_zipcode":"14**1","company_address":"JALAN BARU","company_kelurahan":"CIPAGERANG","company_kecamatan":"CIMAHI UTARA","company_city":"CIMAHI","company_zipcode":"14**1",
			"personal_npwp":"25386221340****","education":"SLTA","marital_status":"M","num_of_dependence":1,"home_status":"SD","profession_id":"WRST","job_type_id":"0104","job_pos":"W",
			"monthly_fixed_income":500000000,"spouse_income":null,"monthly_variable_income":0,"total_installment":3910002,"total_installment_nap":0,"bad_type":null,"max_overduedays":0,
			"max_overduedays_roao":0,"num_of_asset_inventoried":0,"overduedays_aging":null,"max_overduedays_for_active_agreement":0,"max_overduedays_for_prev_eom":0,"sisa_jumlah_angsuran":3,
			"rrd_date":null,"number_of_agreement":1,"work_since_year":"2010","outstanding_principal":97298480.93,"os_installmentdue":0,"is_restructure":0,"is_similiar":0},"server_time": "2021-12-17T11:32:51+07:00"`,
			legalName:   "SOE***E",
			err:         nil,
			errExpected: nil,
			code:        200,
			label:       "TEST_DUPCHECK_INTEGRATOR_SUCCESS",
		},
		{
			body:        "Internal Server Error",
			err:         fmt.Errorf("Timeout"),
			errExpected: fmt.Errorf("upstream_service_timeout - Call Dupcheck Timeout"),
			code:        500,
			label:       "TEST_DUPCHECK_INTEGRATOR_TIMEOUT",
		},
		{
			body:        "Data Not Found",
			err:         nil,
			errExpected: fmt.Errorf("upstream_service_error - Call Dupcheck Error"),
			code:        404,
			label:       "TEST_DUPCHECK_INTEGRATOR_SUCCESS",
		},
	}

	for _, test := range testcase {
		t.Run(test.label, func(t *testing.T) {
			ctx := context.Background()
			accessToken := "access-token"
			timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("DUPCHECK_URL"), httpmock.NewStringResponder(test.code, test.body))
			resp, _ := rst.R().Post(os.Getenv("DUPCHECK_URL"))

			mockHttpClient := new(httpclient.MockHttpClient)
			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("DUPCHECK_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeout, mock.Anything, accessToken).Return(resp, test.err).Once()

			mockRepository := new(mocks.Repository)
			service := NewUsecase(mockRepository, mockHttpClient)

			result, err := service.DupcheckIntegrator(ctx, prospectID, idNumber, legalName, birthDate, surgateName, accessToken)

			mockRepository.AssertExpectations(t)
			fmt.Println(test.label)
			require.Equal(t, test.errExpected, err)
			require.Equal(t, test.legalName, result.FullName)
		})
	}
}
