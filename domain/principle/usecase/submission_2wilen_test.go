package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common/platformevent"
	mockplatformevent "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
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

func TestSubmission2Wilen(t *testing.T) {
	os.Setenv("MI_NUMBER_WHITELIST", "321,768")
	os.Setenv("CUSTOMER_V3_BASE_URL", "http://example.com")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	accessToken := ""

	birthDateStr := "2000-01-01"

	contextReadjustLoanAmount := constant.READJUST_LOAN_AMOUNT_CONTEXT_2WILEN
	contextTenor := constant.READJUST_TENOR_CONTEXT_2WILEN

	testcases := []struct {
		name                                 string
		request                              request.Submission2Wilen
		resExceedErrorTrxKPM                 int
		resGetTrxKPM                         entity.TrxKPM
		errGetTrxKPM                         error
		errSaveTrxKPMStatus                  error
		customerDetailResponse               response.MDMGetDetailCustomerKPMResponse
		errCustomerDetail                    error
		resGetConfig                         entity.AppConfig
		errGetConfig                         error
		errSaveTrxKPM                        error
		resGetReadjustCountTrxKPM            int
		resGetAvailableTenor                 []response.GetAvailableTenorData
		errGetAvailableTenor                 error
		resCheckBannedChassisNumber          response.UsecaseApi
		errCheckBannedChassisNumber          error
		resAgreementChassisNumber            response.AgreementChassisNumber
		resCheckAgreementChassisNumber       response.UsecaseApi
		errCheckAgreementChassisNumber       error
		resDupcheckIntegrator                response.SpDupCekCustomerByID
		errDupcheckIntegrator                error
		resBlacklistCheck                    response.UsecaseApi
		resCustomerType                      string
		errSave                              error
		resNegativeCustomerCheck             response.UsecaseApi
		resNegativeCustomer                  response.NegativeCustomer
		errNegativeCustomerCheck             error
		resCheckMobilePhoneFMF               response.UsecaseApi
		errCheckMobilePhoneFMF               error
		resCustomerKMB                       string
		errCustomerKMB                       error
		resCheckPMK                          response.UsecaseApi
		errCheckPMK                          error
		resMDMGetMasterMappingBranchEmployee response.MDMMasterMappingBranchEmployeeResponse
		errMDMGetMasterMappingBranchEmployee error
		resGetEmployeeData                   response.EmployeeCMOResponse
		errGetEmployeeData                   error
		resGetFpdCMO                         response.FpdCMOResponse
		errGetFpdCMO                         error
		resMasterMappingFpdCluster           entity.MasterMappingFpdCluster
		errMasterMappingFpdCluster           error
		resSavedClusterCheckCmoNoFPD         string
		resEntityCheckCmoNoFPD               entity.TrxCmoNoFPD
		errCheckCmoNoFPD                     error
		resFilteringPefindo                  response.Filtering
		resPefindo                           response.PefindoResult
		resTrxDetailBiroPefindo              []entity.TrxDetailBiro
		errPefindo                           error
		resDukcapil                          response.Ekyc
		errDukcapil                          error
		resAsliri                            response.Ekyc
		errAsliri                            error
		resKtp                               response.Ekyc
		errKtp                               error
		resScsScorepro                       response.IntegratorScorePro
		resScorepro                          response.ScorePro
		resPefindoIDX                        response.PefindoIDX
		errorScorepro                        error
		resGetMappingElaborateLTV            []entity.MappingElaborateLTV
		errGetMappingElaborateLTV            error
		resMarsevGetLoanAmount               response.MarsevLoanAmountResponse
		errMarsevGetLoanAmount               error
		resGetLTV                            int
		resAdjustTenorGetLTV                 bool
		errGetLTV                            error
		resMDMGetMasterAsset                 response.AssetList
		errMDMGetMasterAsset                 error
		resMarsevGetMarketingProgram         response.MarsevFilterProgramResponse
		errMarsevGetMarketingProgram         error
		resMDMGetMappingLicensePlate         response.MDMMasterMappingLicensePlateResponse
		errMDMGetMappingLicensePlate         error
		resMarsevCalculateInstallment        response.MarsevCalculateInstallmentResponse
		errMarsevCalculateInstallment        error
		isGetConfigDupcheck                  bool
		resGetConfigDupcheck                 entity.AppConfig
		errGetConfigDupcheck                 error
		resDsrCheck                          response.UsecaseApi
		resMappingDsrCheck                   response.Dsr
		resInstOtherDsrCheck                 float64
		resInstOtherSpouseDsrCheck           float64
		resInstTopUpDsrCheck                 float64
		errDsrCheck                          error
		resMetricTotalDsrFmfPbk              response.UsecaseApi
		resTrxFMFTotalDsrFmfPbk              response.TrxFMF
		errTotalDsrFmfPbk                    error
		resMasterMappingCluster              entity.MasterMappingCluster
		errMasterMappingCluster              error
		resVehicleCheck                      response.UsecaseApi
		errVehicleCheck                      error
		resCodeCustomerValidateData          int
		resBodyCustomerValidateData          string
		errCustomerValidateData              error
		resCodeInsertCustomerTransaction     int
		resBodyInsertCustomerTransaction     string
		errInsertCustomerTransaction         error
		resCodeUpdateCustomerTransaction     int
		resBodyUpdateCustomerTransaction     string
		errUpdateCustomerTransaction         error
		resCodeMDMMasterBranch               int
		resBodyMDMMasterBranch               string
		errMDMMasterBranch                   error
		resCodeSubmitSally                   int
		resBodySubmitSally                   string
		errSubmitSally                       error
		result                               response.Submission2Wilen
		err                                  error
		expectPublishEventKPMWait            bool
		expectPublishEventKPMApprove         bool
		expectPublishEvent                   bool
		mappingBranch                        entity.MappingBranchByPBKScore
		errMappingBranchEntity               error
		trxDetailBiro                        []entity.TrxDetailBiro
		pbkScore                             string
		errTrxDetailBiro                     error
		mappingPbkScoreGrade                 entity.MappingPBKScoreGrade
		errMappingPbkScoreGrade              error
	}{
		{
			name:                 "error max exceed",
			resExceedErrorTrxKPM: 4,
			err:                  errors.New(constant.ERROR_MAX_EXCEED),
		},
		{
			name:         "error get trx kpm",
			errGetTrxKPM: errors.New("failed get data trx kpm"),
			err:          errors.New("failed get data trx kpm"),
		},
		{
			name: "error trx kpm already rejected",
			resGetTrxKPM: entity.TrxKPM{
				Decision: constant.DECISION_KPM_REJECT,
			},
			err: errors.New(constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE),
		},
		{
			name:                      "error save trx kpm status KPM-WAIT",
			errSaveTrxKPMStatus:       errors.New("something wrong"),
			err:                       errors.New("something wrong"),
			expectPublishEventKPMWait: true,
		},
		{
			name:                      "error get detail customer kpm",
			errCustomerDetail:         errors.New("get detail customer kpm error"),
			err:                       errors.New("get detail customer kpm error"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name:                      "error get config",
			errGetConfig:              errors.New("something wrong"),
			err:                       errors.New("something wrong"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error unmarshal config",
			resGetConfig: entity.AppConfig{
				Value: `-`,
			},
			err:                       errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data config"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error max exceed readjust count",
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetReadjustCountTrxKPM: 4,
			err:                       errors.New(constant.ERROR_MAX_EXCEED),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get available tenor",
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetReadjustCountTrxKPM: 0,
			errGetAvailableTenor:      errors.New("error get available tenor"),
			err:                       errors.New("error get available tenor"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error admin fee does not match",
			request: request.Submission2Wilen{
				AdminFee: 100000,
				Tenor:    12,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetReadjustCountTrxKPM: 0,
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 110000,
				},
			},
			err:                       errors.New(constant.INTERNAL_SERVER_ERROR + " - Admin fee does not match"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check banned chassis number",
			request: request.Submission2Wilen{
				AdminFee: 100000,
				Tenor:    12,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			errCheckBannedChassisNumber: errors.New("error check banned chassis number"),
			err:                         errors.New("error check banned chassis number"),
			expectPublishEventKPMWait:   true,
			expectPublishEvent:          true,
		},
		{
			name: "reject check banned chassis number",
			request: request.Submission2Wilen{
				AdminFee: 100000,
				Tenor:    12,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check agreement chassis number",
			request: request.Submission2Wilen{
				AdminFee: 100000,
				Tenor:    12,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errCheckAgreementChassisNumber: errors.New("error check agreement chassis number"),
			err:                            errors.New("error check agreement chassis number"),
			expectPublishEventKPMWait:      true,
			expectPublishEvent:             true,
		},
		{
			name: "reject check agreement chassis number",
			request: request.Submission2Wilen{
				AdminFee: 100000,
				Tenor:    12,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error dupcheck integrator",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errDupcheckIntegrator:     errors.New("error dupcheck integrator"),
			err:                       errors.New("error dupcheck integrator"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check blacklist",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error save reject check blacklist",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			errSave:                   errors.New("failed save"),
			err:                       errors.New("failed save"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check negative customer",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errNegativeCustomerCheck:  errors.New("error check negative customer"),
			err:                       errors.New("error check negative customer"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check negative customer",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check mobile phone fmf",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errCheckMobilePhoneFMF:    errors.New("error check mobile phone fmf"),
			err:                       errors.New("error check mobile phone fmf"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check mobile phone fmf",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get customer kmb",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errCustomerKMB:            errors.New("error get customer kmb"),
			err:                       errors.New("error get customer kmb"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check pmk",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB:            constant.STATUS_KONSUMEN_AO,
			errCheckPMK:               errors.New("error check pmk"),
			err:                       errors.New("error check pmk"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check pmk",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get mdm master mapping branch employee",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errMDMGetMasterMappingBranchEmployee: errors.New("failed get master mapping branch employee"),
			err:                                  errors.New("failed get master mapping branch employee"),
			expectPublishEventKPMWait:            true,
			expectPublishEvent:                   true,
		},
		{
			name: "error cmo dedicated not found",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{},
			err:                                  errors.New(constant.ERROR_UPSTREAM + " - CMO Dedicated Not Found"),
			expectPublishEventKPMWait:            true,
			expectPublishEvent:                   true,
		},
		{
			name: "error get employee data",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			errGetEmployeeData:        errors.New("failed get employee data"),
			err:                       errors.New("failed get employee data"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error cmo not found",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: "",
			},
			err:                       errors.New(constant.ERROR_UPSTREAM + " - CMO Not Found"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get fpd cmo",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			errGetFpdCMO:              errors.New("failed get fpd cmo"),
			err:                       errors.New("failed get fpd cmo"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get master mapping fpd cluster",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: true,
			},
			errMasterMappingFpdCluster: errors.New("failed get master mapping fpd cluster"),
			err:                        errors.New("failed get master mapping fpd cluster"),
			expectPublishEventKPMWait:  true,
			expectPublishEvent:         true,
		},
		{
			name: "error check cmo no fpd",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			errCheckCmoNoFPD:          errors.New("failed check cmo no fpd"),
			err:                       errors.New("failed check cmo no fpd"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error pefindo",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			errPefindo:                   errors.New("failed check pefindo"),
			err:                          errors.New("failed check pefindo"),
			expectPublishEventKPMWait:    true,
			expectPublishEvent:           true,
		},
		{
			name: "error check ekyc dukcapil",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				NewKoRules: response.ResultNewKoRules{
					CategoryPBK:       "lunas_diskon",
					ContractCode:      "1070007119003",
					ContractStatus:    "Settled",
					CreditorType:      "NotSpecified",
					Creditor:          "PT.MANDIRI Finance",
					ConditionDate:     "2023-10-11",
					RestructuringDate: "2024-11-11",
				},
			},
			errDukcapil:               errors.New("failed check dukcapil"),
			err:                       errors.New("failed check dukcapil"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check ekyc asliri",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			resAsliri: response.Ekyc{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check ekyc ktp",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil:               fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:                 fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			errKtp:                    errors.New("failed check ekyc ktp"),
			err:                       errors.New("failed check ekyc ktp"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check ekyc ktp",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error save after check ekyc",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			errSave:                   errors.New("failed to save after check ekyc"),
			err:                       errors.New("failed to save after check ekyc"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject pefindo",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: false,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error scorepro",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			errorScorepro:             errors.New("failed check scorepro"),
			err:                       errors.New("failed check scorepro"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject scorepro",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get mapping elaborate ltv",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			errGetMappingElaborateLTV: errors.New("failed get mapping elaborate ltv"),
			err:                       errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get ltv",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			errGetLTV:                 errors.New("failed to get ltv"),
			err:                       errors.New("failed to get ltv"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject max readjust attempt",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 0,
			result: response.Submission2Wilen{
				Code:            constant.READJUST_LOAN_AMOUNT_CODE_2WILEN,
				Result:          constant.DECISION_KPM_REJECT,
				ReadjustContext: &contextReadjustLoanAmount,
				Reason:          constant.READJUST_LOAN_AMOUNT_REASON_2WILEN,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject marsev get loan amount",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV:                 90,
			errMarsevGetLoanAmount:    errors.New("failed get marsev loan amount"),
			err:                       errors.New("failed get marsev loan amount"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject max readjust attempt loan amount",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 900000,
				},
			},
			result: response.Submission2Wilen{
				Code:            constant.READJUST_LOAN_AMOUNT_CODE_2WILEN,
				Result:          constant.DECISION_KPM_REJECT,
				ReadjustContext: &contextReadjustLoanAmount,
				Reason:          constant.READJUST_LOAN_AMOUNT_REASON_2WILEN,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error mdm get master asset",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			errMDMGetMasterAsset:      errors.New("failed to get mdm master asset"),
			err:                       errors.New("failed to get mdm master asset"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error marsev get program",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			errMarsevGetMarketingProgram: errors.New("failed to get mdm master asset"),
			err:                          errors.New("failed to get mdm master asset"),
			expectPublishEventKPMWait:    true,
			expectPublishEvent:           true,
		},
		{
			name: "error marsev get program not found",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			err:                       errors.New(constant.ERROR_UPSTREAM + " - Marsev Filter Program Not Found"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error marsev get program not matching mi number found",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				KPMID:                   6287,
				ReferralCode:            "TX92XS",
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 123,
					},
				},
			},
			err:                       errors.New(constant.ERROR_BAD_REQUEST + " - No matching MI_NUMBER found"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error mdm get mapping license plate",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			errMDMGetMappingLicensePlate: errors.New("failed to get mapping license plate"),
			err:                          errors.New("failed to get mapping license plate"),
			expectPublishEventKPMWait:    true,
			expectPublishEvent:           true,
		},
		{
			name: "error marsev calculate installment",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			errMarsevCalculateInstallment: errors.New("failed to calculate installment marsev"),
			err:                           errors.New("failed to calculate installment marsev"),
			expectPublishEventKPMWait:     true,
			expectPublishEvent:            true,
		},
		{
			name: "readjust after marsev calculate installment",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 1,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			result: response.Submission2Wilen{
				Result:          constant.DECISION_KPM_READJUST,
				Code:            constant.READJUST_TENOR_CODE_2WILEN,
				Reason:          constant.READJUST_TENOR_REASON_2WILEN,
				ReadjustContext: &contextTenor,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject readjust after marsev calculate installment",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			result: response.Submission2Wilen{
				Result:          constant.DECISION_KPM_REJECT,
				Code:            constant.READJUST_TENOR_CODE_2WILEN,
				Reason:          constant.READJUST_TENOR_REASON_2WILEN,
				ReadjustContext: &contextTenor,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get config dupcheck",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck:       true,
			errGetConfigDupcheck:      errors.New("failed get config dupcheck"),
			err:                       errors.New("failed get config dupcheck"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error unmarshal get config dupcheck",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Value: `-`,
			},
			err:                       errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data config dupcheck"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check dsr",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			errDsrCheck:               errors.New("failed check dsr"),
			err:                       errors.New("failed check dsr"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check dsr",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error check dsr fmf pbk",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errTotalDsrFmfPbk:         errors.New("failed to check dsr fmf pbk"),
			err:                       errors.New("failed to check dsr fmf pbk"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject check dsr fmf pbk",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error get master mapping cluster",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errMasterMappingCluster:   errors.New("failed get master mapping cluster"),
			err:                       errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error vehicle check",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			errVehicleCheck:           errors.New("failed vehicle check"),
			err:                       errors.New("failed vehicle check"),
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "reject vehicle check",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
			},
			result: response.Submission2Wilen{
				Result: constant.DECISION_KPM_REJECT,
			},
			expectPublishEventKPMWait: true,
			expectPublishEvent:        true,
		},
		{
			name: "error customer validate data",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			errCustomerValidateData:      errors.New("failed customer validate data"),
			err:                          errors.New("failed customer validate data"),
			expectPublishEventKPMWait:    true,
			expectPublishEventKPMApprove: true,
			expectPublishEvent:           true,
		},
		{
			name: "error upstream customer validate data",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:  500,
			err:                          errors.New(constant.ERROR_UPSTREAM + " - Customer Validate Data Error"),
			expectPublishEventKPMWait:    true,
			expectPublishEventKPMApprove: true,
			expectPublishEvent:           true,
		},
		{
			name: "error unmarshal customer validate data",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:  200,
			resBodyCustomerValidateData:  `-`,
			err:                          errors.New(`invalid character ' ' in numeric literal`),
			expectPublishEventKPMWait:    true,
			expectPublishEventKPMApprove: true,
			expectPublishEvent:           true,
		},
		{
			name: "error insert customer transaction",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:  200,
			resBodyCustomerValidateData:  `{"data":{"customer_id":0,"kpm_id":0}}`,
			errInsertCustomerTransaction: errors.New(`failed insert customer transaction`),
			err:                          errors.New(`failed insert customer transaction`),
			expectPublishEventKPMWait:    true,
			expectPublishEventKPMApprove: true,
			expectPublishEvent:           true,
		},
		{
			name: "error code insert customer transaction",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 500,
			err:                              errors.New(constant.ERROR_UPSTREAM + " - Insert Customer Data Error"),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error unmarshal insert customer transaction",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `-`,
			err:                              errors.New(`invalid character ' ' in numeric literal`),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error update customer transaction",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			errUpdateCustomerTransaction:     errors.New("failed update customer transaction"),
			err:                              errors.New("failed update customer transaction"),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error code update customer transaction",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 500,
			err:                              errors.New(constant.ERROR_UPSTREAM + " - Update Customer Transaction Error"),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error unmarshal update customer transaction",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 200,
			resBodyUpdateCustomerTransaction: `-`,
			err:                              errors.New(`invalid character ' ' in numeric literal`),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error mdm master branch",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 200,
			resBodyUpdateCustomerTransaction: `{"data":{}}`,
			errMDMMasterBranch:               errors.New("failed mdm master branch"),
			err:                              errors.New("failed mdm master branch"),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error code mdm master branch",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 200,
			resBodyUpdateCustomerTransaction: `{"data":{}}`,
			resCodeMDMMasterBranch:           500,
			err:                              errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Branch Error"),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error unmarshal mdm master branch",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 200,
			resBodyUpdateCustomerTransaction: `{"data":{}}`,
			resCodeMDMMasterBranch:           200,
			resBodyMDMMasterBranch:           `-`,
			err:                              errors.New(`invalid character ' ' in numeric literal`),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error submit to sally",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KtpPhoto:                "http://image.com",
				SelfiePhoto:             "http://image.com",
				Dealer:                  constant.DEALER_PSA,
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 200,
			resBodyUpdateCustomerTransaction: `{"data":{}}`,
			resCodeMDMMasterBranch:           200,
			resBodyMDMMasterBranch:           `{"data":{"branch_id":"426","branch_name":"BANDUNG"}}`,
			errSubmitSally:                   errors.New("failed submit to sally"),
			err:                              errors.New("failed submit to sally"),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error code submit to sally",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 200,
			resBodyUpdateCustomerTransaction: `{"data":{}}`,
			resCodeMDMMasterBranch:           200,
			resBodyMDMMasterBranch:           `{"data":{"branch_id":"426","branch_name":"BANDUNG"}}`,
			resCodeSubmitSally:               500,
			err:                              errors.New(constant.ERROR_UPSTREAM + " - Sally Submit 2W Principle Error"),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
		{
			name: "error unmarshal submit to sally",
			request: request.Submission2Wilen{
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "456",
				SpouseLegalName:         "Test Spouse Legal Name",
				SpouseBirthDate:         birthDateStr,
				SpouseSurgateMotherName: "Test Spouse Surgate Mother Name",
				BPKBNameType:            "K",
				AdminFee:                100000,
				Tenor:                   12,
				LoanAmount:              1000000,
				ReferralCode:            "TX92XS",
				KPMID:                   6287,
			},
			resGetReadjustCountTrxKPM: 2,
			resGetConfig: entity.AppConfig{
				Value: `{"data":{"max_readjust_attempt":3}}`,
			},
			resGetAvailableTenor: []response.GetAvailableTenorData{
				{
					Tenor:    12,
					AdminFee: 100000,
				},
			},
			resCheckBannedChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckAgreementChassisNumber: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resBlacklistCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resNegativeCustomerCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCheckMobilePhoneFMF: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCustomerKMB: constant.STATUS_KONSUMEN_AO,
			resCheckPMK: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMDMGetMasterMappingBranchEmployee: response.MDMMasterMappingBranchEmployeeResponse{
				Data: []response.MDMMasterMappingBranchEmployeeRecord{
					{
						CMOID:   "123",
						CMOName: "Test CMO Name",
					},
				},
			},
			resGetEmployeeData: response.EmployeeCMOResponse{
				CMOCategory: constant.CMO_LAMA,
			},
			resGetFpdCMO: response.FpdCMOResponse{
				FpdExist: false,
			},
			resSavedClusterCheckCmoNoFPD: "Cluster A",
			resPefindo: response.PefindoResult{
				Score:                         "LOW RISK",
				Category:                      1,
				WoContract:                    true,
				WoAdaAgunan:                   true,
				MaxOverdueKORules:             20,
				MaxOverdueLast12MonthsKORules: 12,
				AngsuranAktifPbk:              1000000,
			},
			errDukcapil: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
			errAsliri:   fmt.Errorf("%s - Asliri", constant.TYPE_CONTINGENCY),
			resKtp: response.Ekyc{
				Result: constant.DECISION_PASS,
			},
			resFilteringPefindo: response.Filtering{
				Code:        "",
				NextProcess: true,
			},
			resScorepro: response.ScorePro{
				Result: constant.DECISION_PASS,
			},
			resGetMappingElaborateLTV: []entity.MappingElaborateLTV{
				{
					LTV: 90,
				},
			},
			resGetLTV: 90,
			resMarsevGetLoanAmount: response.MarsevLoanAmountResponse{
				Data: response.MarsevLoanAmountData{
					LoanAmountMaximum: 1000000,
				},
			},
			resMDMGetMasterAsset: response.AssetList{
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
			resMarsevGetMarketingProgram: response.MarsevFilterProgramResponse{
				Data: []response.MarsevFilterProgramData{
					{
						MINumber: 321,
					},
				},
			},
			resMDMGetMappingLicensePlate: response.MDMMasterMappingLicensePlateResponse{
				Data: response.MDMMasterMappingLicensePlateData{
					Records: []response.MDMMasterMappingLicensePlateRecord{
						{
							AreaID: "Wilayah 2",
						},
					},
				},
			},
			resMarsevCalculateInstallment: response.MarsevCalculateInstallmentResponse{
				Data: []response.MarsevCalculateInstallmentData{
					{
						ProvisionFee: 1100000,
					},
				},
			},
			isGetConfigDupcheck: true,
			resGetConfigDupcheck: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":{"prime":5000000,"priority":5000000,"regular":5000000}}}`,
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMetricTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_PASS,
			},
			resCodeCustomerValidateData:      200,
			resBodyCustomerValidateData:      `{"data":{"customer_id":123,"kpm_id":0}}`,
			resCodeInsertCustomerTransaction: 200,
			resBodyInsertCustomerTransaction: `{"data":{"customer_id":0,"is_new_customer":true}}`,
			resCodeUpdateCustomerTransaction: 200,
			resBodyUpdateCustomerTransaction: `{"data":{}}`,
			resCodeMDMMasterBranch:           200,
			resBodyMDMMasterBranch:           `{"data":{"branch_id":"426","branch_name":"BANDUNG"}}`,
			resCodeSubmitSally:               200,
			resBodySubmitSally:               `-`,
			err:                              errors.New(`invalid character ' ' in numeric literal`),
			expectPublishEventKPMWait:        true,
			expectPublishEventKPMApprove:     true,
			expectPublishEvent:               true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockUsecase := new(mocks.Usecase)
			mockMultiUsecase := new(mocks.MultiUsecase)
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			mockRepository.On("ExceedErrorTrxKPM", mock.Anything).Return(tc.resExceedErrorTrxKPM)
			mockRepository.On("GetTrxKPM", mock.Anything).Return(tc.resGetTrxKPM, tc.errGetTrxKPM)
			mockRepository.On("SaveTrxKPMStatus", mock.Anything).Return(tc.errSaveTrxKPMStatus)
			mockRepository.On("GetConfig", mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetConfig, tc.errGetConfig).Once()
			mockRepository.On("SaveTrxKPM", mock.Anything).Return(tc.errSaveTrxKPM)
			mockRepository.On("GetReadjustCountTrxKPM", mock.Anything).Return(tc.resGetReadjustCountTrxKPM)
			mockRepository.On("MasterMappingFpdCluster", mock.Anything).Return(tc.resMasterMappingFpdCluster, tc.errMasterMappingFpdCluster)
			mockRepository.On("GetTrxDetailBIro", tc.request.ProspectID).Return(tc.trxDetailBiro, tc.errTrxDetailBiro)
			mockRepository.On("GetMappingPbkScore", mock.Anything).Return(tc.mappingPbkScoreGrade, tc.errMappingPbkScoreGrade)
			mockRepository.On("GetMappingBranchByBranchID", tc.request.BranchID, mock.Anything).
				Return(tc.mappingBranch, tc.errMappingBranchEntity)
			mockRepository.On("GetMappingElaborateLTV", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetMappingElaborateLTV, tc.errGetMappingElaborateLTV)
			if tc.isGetConfigDupcheck {
				mockRepository.On("GetConfig", mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetConfigDupcheck, tc.errGetConfigDupcheck).Once()
			}
			mockRepository.On("MasterMappingCluster", mock.Anything).Return(tc.resMasterMappingCluster, tc.errMasterMappingCluster)

			mockUsecase.On("MDMGetDetailCustomerKPM", ctx, tc.request.ProspectID, tc.request.KPMID, accessToken).Return(tc.customerDetailResponse, tc.errCustomerDetail)
			mockUsecase.On("CheckBannedChassisNumber", mock.Anything).Return(tc.resCheckBannedChassisNumber, tc.errCheckBannedChassisNumber)
			mockUsecase.On("CheckAgreementChassisNumber", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resAgreementChassisNumber, tc.resCheckAgreementChassisNumber, tc.errCheckAgreementChassisNumber)
			mockUsecase.On("DupcheckIntegrator", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resDupcheckIntegrator, tc.errDupcheckIntegrator)
			mockUsecase.On("BlacklistCheck", mock.Anything, mock.Anything).Return(tc.resBlacklistCheck, tc.resCustomerType)
			mockUsecase.On("Save", mock.Anything, mock.Anything, mock.Anything).Return(tc.errSave)
			mockUsecase.On("NegativeCustomerCheck", ctx, mock.Anything, mock.Anything).Return(tc.resNegativeCustomerCheck, tc.resNegativeCustomer, tc.errNegativeCustomerCheck)
			mockUsecase.On("CheckMobilePhoneFMF", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resCheckMobilePhoneFMF, tc.errCheckMobilePhoneFMF)
			mockUsecase.On("CustomerKMB", mock.Anything).Return(tc.resCustomerKMB, tc.errCustomerKMB)
			mockUsecase.On("CheckPMK", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resCheckPMK, tc.errCheckPMK)
			mockUsecase.On("MDMGetMasterMappingBranchEmployee", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resMDMGetMasterMappingBranchEmployee, tc.errMDMGetMasterMappingBranchEmployee)
			mockUsecase.On("GetEmployeeData", ctx, mock.Anything).Return(tc.resGetEmployeeData, tc.errGetEmployeeData)
			mockUsecase.On("GetFpdCMO", ctx, mock.Anything, mock.Anything).Return(tc.resGetFpdCMO, tc.errGetFpdCMO)
			mockUsecase.On("CheckCmoNoFPD", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resSavedClusterCheckCmoNoFPD, tc.resEntityCheckCmoNoFPD, tc.errCheckCmoNoFPD)
			mockUsecase.On("Pefindo", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resFilteringPefindo, tc.resPefindo, tc.resTrxDetailBiroPefindo, tc.errPefindo)
			mockUsecase.On("Dukcapil", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resDukcapil, tc.errDukcapil)
			mockUsecase.On("Asliri", ctx, mock.Anything, mock.Anything).Return(tc.resAsliri, tc.errAsliri)
			mockUsecase.On("Ktp", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resKtp, tc.errKtp)
			mockUsecase.On("Scorepro", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resScsScorepro, tc.resScorepro, tc.resPefindoIDX, tc.errorScorepro)
			mockUsecase.On("GetLTV", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resGetLTV, tc.resAdjustTenorGetLTV, tc.errGetLTV)
			mockUsecase.On("MarsevGetLoanAmount", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resMarsevGetLoanAmount, tc.errMarsevGetLoanAmount)
			mockUsecase.On("MDMGetMasterAsset", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resMDMGetMasterAsset, tc.errMDMGetMasterAsset)
			mockUsecase.On("MarsevGetMarketingProgram", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resMarsevGetMarketingProgram, tc.errMarsevGetMarketingProgram)
			mockUsecase.On("MDMGetMappingLicensePlate", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resMDMGetMappingLicensePlate, tc.errMDMGetMappingLicensePlate)
			mockUsecase.On("MarsevCalculateInstallment", ctx, mock.Anything, mock.Anything, mock.Anything).Return(tc.resMarsevCalculateInstallment, tc.errMarsevCalculateInstallment)
			mockUsecase.On("DsrCheck", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resDsrCheck, tc.resMappingDsrCheck, tc.resInstOtherDsrCheck, tc.resInstOtherSpouseDsrCheck, tc.resInstTopUpDsrCheck, tc.errDsrCheck)
			mockUsecase.On("TotalDsrFmfPbk", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resMetricTotalDsrFmfPbk, tc.resTrxFMFTotalDsrFmfPbk, tc.errTotalDsrFmfPbk)
			mockUsecase.On("VehicleCheck", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resVehicleCheck, tc.errVehicleCheck)

			mockMultiUsecase.On("GetAvailableTenor", ctx, mock.Anything, mock.Anything).Return(tc.resGetAvailableTenor, tc.errGetAvailableTenor)

			if tc.expectPublishEventKPMWait {
				mockPlatformEvent.On("PublishEvent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}

			if tc.expectPublishEventKPMApprove {
				mockPlatformEvent.On("PublishEvent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}

			if tc.expectPublishEvent {
				mockPlatformEvent.On("PublishEvent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			url := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/validate-data"
			httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resCodeCustomerValidateData, tc.resBodyCustomerValidateData))
			resp, _ := rst.R().Post(url)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.request.ProspectID, accessToken).Return(resp, tc.errCustomerValidateData).Once()

			url2 := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/transaction"
			httpmock.RegisterResponder(constant.METHOD_POST, url2, httpmock.NewStringResponder(tc.resCodeInsertCustomerTransaction, tc.resBodyInsertCustomerTransaction))
			resp2, _ := rst.R().Post(url2)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url2, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.request.ProspectID, accessToken).Return(resp2, tc.errInsertCustomerTransaction).Once()

			url3 := os.Getenv("CUSTOMER_V3_BASE_URL") + "/api/v3/customer/transaction/" + tc.request.ProspectID
			httpmock.RegisterResponder(constant.METHOD_PUT, url3, httpmock.NewStringResponder(tc.resCodeUpdateCustomerTransaction, tc.resBodyUpdateCustomerTransaction))
			resp3, _ := rst.R().Put(url3)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url3, mock.Anything, mock.Anything, constant.METHOD_PUT, false, 0, 30, tc.request.ProspectID, accessToken).Return(resp3, tc.errUpdateCustomerTransaction).Once()

			url4 := os.Getenv("MDM_MASTER_BRANCH_URL") + tc.request.BranchID
			httpmock.RegisterResponder(constant.METHOD_GET, url4, httpmock.NewStringResponder(tc.resCodeMDMMasterBranch, tc.resBodyMDMMasterBranch))
			resp4, _ := rst.R().Get(url4)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url4, mock.Anything, mock.Anything, constant.METHOD_GET, false, 0, 30, tc.request.ProspectID, accessToken).Return(resp4, tc.errMDMMasterBranch).Once()

			url5 := os.Getenv("SALLY_SUBMISSION_2W_PRINCIPLE_URL")
			httpmock.RegisterResponder(constant.METHOD_POST, url5, httpmock.NewStringResponder(tc.resCodeSubmitSally, tc.resBodySubmitSally))
			resp5, _ := rst.R().Post(url5)
			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url5, mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.request.ProspectID, accessToken).Return(resp5, tc.errSubmitSally).Once()

			metrics := NewMetrics(mockRepository, mockHttpClient, platformEvent, mockUsecase, mockMultiUsecase)

			result, err := metrics.Submission2Wilen(ctx, tc.request, accessToken)

			if tc.err != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.result, result)
			}
		})
	}
}

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
