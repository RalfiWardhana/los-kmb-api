package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common/platformevent"
	mockplatformevent "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPrinciplePembiayaan(t *testing.T) {
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	accessToken := ""

	birthDateStr := "2000-01-01"
	birthDate, _ := time.Parse("2006-01-02", birthDateStr)

	var monthlyVariableIncome float64 = 100000.0

	testcases := []struct {
		name                                string
		request                             request.PrinciplePembiayaan
		resGetPrincipleStepOne              entity.TrxPrincipleStepOne
		errGetPrincipleStepOne              error
		resGetPrincipleStepTwo              entity.TrxPrincipleStepTwo
		errGetPrincipleStepTwo              error
		resGetFiltering                     entity.FilteringKMB
		errGetFiltering                     error
		resGetConfig                        entity.AppConfig
		errGetConfig                        error
		resMasterMappingCluster             entity.MasterMappingCluster
		errMasterMappingCluster             error
		resVehicleCheck                     response.UsecaseApi
		errVehicleCheck                     error
		resRejectTenor36                    response.UsecaseApi
		errRejectTenor36                    error
		resDupcheckIntegrator               response.SpDupCekCustomerByID
		errDupcheckIntegrator               error
		resAgreementChassisNumberIntegrator response.AgreementChassisNumber
		errAgreementChassisNumberIntegrator error
		resScorepro                         response.IntegratorScorePro
		resMetricsScorepro                  response.ScorePro
		resPefindoIDXScorepro               response.PefindoIDX
		errScorepro                         error
		resDsrCheck                         response.UsecaseApi
		resMappingDsr                       response.Dsr
		resInstOther                        float64
		resInstOtherSpouse                  float64
		resInstTopUp                        float64
		errDsrCheck                         error
		resTotalDsrFmfPbk                   response.UsecaseApi
		resTrxFMFTotalDsrFmfPbk             response.TrxFMF
		errTotalDsrFmfPbk                   error
		errSavePrincipleStepThree           error
		errUpdatePrincipleStepTwo           error
		result                              response.UsecaseApi
		err                                 error
		expectPublishEvent                  bool
	}{
		{
			name: "error get principle step one",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			errGetPrincipleStepOne: errors.New("something wrong"),
			err:                    errors.New("something wrong"),
		},
		{
			name: "error get principle step two",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			errGetPrincipleStepTwo: errors.New("something wrong"),
			err:                    errors.New("something wrong"),
		},
		{
			name: "error get filtering",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			errGetFiltering: errors.New("something wrong"),
			err:             errors.New("something wrong"),
		},
		{
			name: "error get config",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			errGetConfig: errors.New("something wrong"),
			err:          errors.New("something wrong"),
		},
		{
			name: "error get config",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				Decision: constant.DECISION_REJECT,
			},
			err: errors.New(constant.PRINCIPLE_ALREADY_REJECTED_MESSAGE),
		},
		{
			name: "error get master mapping cluster",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			errMasterMappingCluster: errors.New("something wrong"),
			err:                     errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error"),
		},
		{
			name: "error vehicle check",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			errVehicleCheck: errors.New("something wrong"),
			err:             errors.New("something wrong"),
		},
		{
			name: "reject vehicle check",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     nil,
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			resVehicleCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_VEHICLE_AGE_MAX,
				Reason: fmt.Sprintf("%s Ketentuan", constant.REASON_VEHICLE_AGE_MAX),
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_VEHICLE_AGE_MAX,
				Reason: "Data pembiayaan tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error check reject tenor 36",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
				Tenor:             36,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			errRejectTenor36: errors.New("something wrong"),
			err:              errors.New("something wrong"),
		},
		{
			name: "reject tenor > 36",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
				Tenor:             37,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_REJECT_TENOR,
				Reason: "Data pembiayaan tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error dupcheck integrator",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
				Tenor:             12,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName: "K",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			errDupcheckIntegrator: errors.New("something wrong"),
			err:                   errors.New("something wrong"),
		},
		{
			name: "error get agreement chassis number integrator",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
				Tenor:             12,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			errAgreementChassisNumberIntegrator: errors.New("something wrong"),
			err:                                 errors.New("something wrong"),
		},
		{
			name: "error scorepro",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
				Tenor:             12,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			errScorepro: errors.New("something wrong"),
			err:         errors.New("something wrong"),
		},
		{
			name: "reject scorepro",
			request: request.PrinciplePembiayaan{
				ProspectID:        "SAL-123",
				InstallmentAmount: 10000000,
				Tenor:             12,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
			},
			resGetFiltering: entity.FilteringKMB{
				CMOCluster:      "Cluster C",
				CustomerSegment: constant.RO_AO_REGULAR,
				ScoreBiro:       "HIGH",
				CustomerStatus:  constant.STATUS_KONSUMEN_AO,
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			resAgreementChassisNumberIntegrator: response.AgreementChassisNumber{
				InstallmentAmount: 1000000,
			},
			resMetricsScorepro: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: "Data pembiayaan tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error check dsr",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			errDsrCheck: errors.New("something wrong"),
			err:         errors.New("something wrong"),
		},
		{
			name: "error unmarshal dupcheck data",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
				DupcheckData:            `invalid json`,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			err: errors.New("invalid character 'i' looking for beginning of value"),
		},
		{
			name: "reject check dsr",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
				DupcheckData:            `{"max_overduedays_roao":60}`,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CMOCluster:     "Cluster C",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			resDsrCheck: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_DSRGT35,
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_DSRGT35,
				Reason: "Data pembiayaan tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error parse float instalment biro",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
				DupcheckData:            `{"max_overduedays_roao":60}`,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus:             constant.STATUS_KONSUMEN_RO,
				CMOCluster:                 "Cluster C",
				TotalInstallmentAmountBiro: "error",
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			err: errors.New(constant.ERROR_UPSTREAM + " - GetFloat TotalInstallmentAmountBiro Error"),
		},
		{
			name: "error check total dsr fmf pbk",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
				DupcheckData:            `{"max_overduedays_roao":60}`,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus:             constant.STATUS_KONSUMEN_RO,
				CMOCluster:                 "Cluster C",
				TotalInstallmentAmountBiro: 100000.0,
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			errTotalDsrFmfPbk: errors.New("something wrong"),
			err:               errors.New("something wrong"),
		},
		{
			name: "reject check total dsr fmf pbk",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
				DupcheckData:            `{"max_overduedays_roao":60}`,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus:             constant.STATUS_KONSUMEN_RO,
				CMOCluster:                 "Cluster C",
				TotalInstallmentAmountBiro: 100000.0,
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			resTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_TOTAL_DSRGT35,
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_TOTAL_DSRGT35,
				Reason: "Data pembiayaan tidak lolos verifikasi",
			},
			expectPublishEvent: true,
		},
		{
			name: "error save principle step three",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
				DupcheckData:            `{"max_overduedays_roao":60}`,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus:             constant.STATUS_KONSUMEN_RO,
				CMOCluster:                 "Cluster C",
				TotalInstallmentAmountBiro: 100000.0,
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			resTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_TOTAL_DSRGT35,
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_TOTAL_DSRGT35,
				Reason: "Data pembiayaan tidak lolos verifikasi",
			},
			errSavePrincipleStepThree: errors.New("something wrong"),
			err:                       errors.New("something wrong"),
		},
		{
			name: "error update principle step two",
			request: request.PrinciplePembiayaan{
				ProspectID:            "SAL-123",
				InstallmentAmount:     10000000,
				Tenor:                 12,
				MonthlyVariableIncome: &monthlyVariableIncome,
			},
			resGetPrincipleStepOne: entity.TrxPrincipleStepOne{
				BPKBName:  "K",
				NoChassis: "TEST123",
			},
			resGetPrincipleStepTwo: entity.TrxPrincipleStepTwo{
				IDNumber:                "123456",
				LegalName:               "Test",
				BirthDate:               birthDate,
				SurgateMotherName:       "Test Mother Name",
				MaritalStatus:           constant.MARRIED,
				SpouseIDNumber:          "123456",
				SpouseLegalName:         "Test Spouse Name",
				SpouseBirthDate:         birthDate,
				SpouseSurgateMotherName: "Test Spouse Mother Name",
				SpouseIncome:            1000000.0,
				DupcheckData:            `{"max_overduedays_roao":60}`,
			},
			resGetFiltering: entity.FilteringKMB{
				CustomerStatus:             constant.STATUS_KONSUMEN_RO,
				CMOCluster:                 "Cluster C",
				TotalInstallmentAmountBiro: 100000.0,
			},
			resGetConfig: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			resMasterMappingCluster: entity.MasterMappingCluster{
				BranchID: "426",
				Cluster:  "Cluster C",
			},
			resTotalDsrFmfPbk: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_TOTAL_DSRGT35,
			},
			result: response.UsecaseApi{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_TOTAL_DSRGT35,
				Reason: "Data pembiayaan tidak lolos verifikasi",
			},
			errUpdatePrincipleStepTwo: errors.New("something wrong"),
			err:                       errors.New("something wrong"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			mockRepository.On("GetPrincipleStepOne", tc.request.ProspectID).Return(tc.resGetPrincipleStepOne, tc.errGetPrincipleStepOne)
			mockRepository.On("GetPrincipleStepTwo", tc.request.ProspectID).Return(tc.resGetPrincipleStepTwo, tc.errGetPrincipleStepTwo)
			mockRepository.On("GetFilteringResult", tc.request.ProspectID).Return(tc.resGetFiltering, tc.errGetFiltering)
			mockRepository.On("GetConfig", "dupcheck", constant.LOB_KMB_OFF, "dupcheck_kmb_config").Return(tc.resGetConfig, tc.errGetConfig)

			mockRepository.On("MasterMappingCluster", mock.AnythingOfType("entity.MasterMappingCluster")).Return(tc.resMasterMappingCluster, tc.errMasterMappingCluster)
			mockRepository.On("SavePrincipleStepThree", mock.AnythingOfType("entity.TrxPrincipleStepThree")).Return(tc.errSavePrincipleStepThree)
			mockRepository.On("UpdatePrincipleStepTwo", tc.request.ProspectID, mock.AnythingOfType("entity.TrxPrincipleStepTwo")).Return(tc.errUpdatePrincipleStepTwo)

			mockUsecase := new(mocks.Usecase)
			mockUsecase.On("VehicleCheck", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resVehicleCheck, tc.errVehicleCheck)
			mockUsecase.On("RejectTenor36", mock.Anything).Return(tc.resRejectTenor36, tc.errRejectTenor36)
			mockUsecase.On("DupcheckIntegrator", ctx, tc.request.ProspectID, mock.Anything, mock.Anything, mock.Anything, mock.Anything, "").Return(tc.resDupcheckIntegrator, tc.errDupcheckIntegrator)
			mockUsecase.On("AgreementChassisNumberIntegrator", ctx, mock.Anything, tc.request.ProspectID, mock.Anything).Return(tc.resAgreementChassisNumberIntegrator, tc.errAgreementChassisNumberIntegrator)
			mockUsecase.On("Scorepro", ctx, tc.request, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resScorepro, tc.resMetricsScorepro, tc.resPefindoIDXScorepro, tc.errScorepro)
			mockUsecase.On("DsrCheck", ctx, tc.request, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resDsrCheck, tc.resMappingDsr, tc.resInstOther, tc.resInstOtherSpouse, tc.resInstTopUp, tc.errDsrCheck)
			mockUsecase.On("TotalDsrFmfPbk", ctx, mock.Anything, mock.Anything, mock.Anything, tc.request.ProspectID, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tc.resTotalDsrFmfPbk, tc.resTrxFMFTotalDsrFmfPbk, tc.errTotalDsrFmfPbk)

			if tc.expectPublishEvent {
				mockPlatformEvent.On("PublishEvent", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, 0).Return(nil).Once()
			}

			multiUsecase := NewMultiUsecase(mockRepository, mockHttpClient, platformEvent, mockUsecase)

			result, err := multiUsecase.PrinciplePembiayaan(ctx, tc.request, accessToken)

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
		filtering               entity.FilteringKMB
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

			mockRepository.On("GetMappingVehicleAge", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.resGetMappingVehicleAge, test.errGetMappingVehicleAge)

			service := NewUsecase(mockRepository, mockHttpClient, nil)
			result, err := service.VehicleCheck(test.year, test.cmoCluster, test.bpkbName, test.tenor, configValue, test.filtering, 100)

			require.Equal(t, test.errExpected, err)
			require.Equal(t, test.vehicle.Result, result.Result)
			require.Equal(t, test.vehicle.Code, result.Code)
			require.Equal(t, test.vehicle.Reason, result.Reason)
		})
	}

}

func TestRejectTenor36(t *testing.T) {
	responseAppConfig := entity.AppConfig{
		Value: `{"data":["Cluster A", "Cluster B"]}`,
	}
	testcases := []struct {
		name              string
		cluster           string
		result            response.UsecaseApi
		errResult         error
		responseAppConfig error
		trxStatus         entity.TrxStatus
	}{
		{
			name:              "handling tenor error config",
			cluster:           "Cluster A",
			responseAppConfig: errors.New(constant.ERROR_UPSTREAM + " - GetConfig exclusion_tenor36 Error"),
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - GetConfig exclusion_tenor36 Error"),
		},
		{
			name:    "handling tenor reject",
			cluster: "Cluster C",
			result: response.UsecaseApi{
				Code:   constant.CODE_REJECT_TENOR,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_TENOR},
		},
		{
			name:    "handling tenor pass",
			cluster: "Cluster A",
			result: response.UsecaseApi{
				Code:   constant.CODE_PASS_TENOR,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_PASS_TENOR},
			trxStatus: entity.TrxStatus{
				ProspectID:     "SLY-123457678",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Decision:       constant.DB_DECISION_REJECT,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetConfig", "tenor36", "KMB-OFF", "exclusion_tenor36").Return(responseAppConfig, tc.responseAppConfig)

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			result, err := usecase.RejectTenor36(tc.cluster)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errResult, err)
		})
	}

}

func TestScorepro(t *testing.T) {

	os.Setenv("SCOREPRO_DEFAULT_KEY", "first_residence_zipcode_2w_others")
	os.Setenv("SCOREPRO_DEFAULT_SCORE_GENERATOR_ID", "37fe1525-1be1-48d1-aab5-6adf05305a0a")
	os.Setenv("SCOREPRO_REQUESTID", "107b280e-12e4-4c92-80e8-38a7422cb9bc")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("PEFINDO_IDX_URL", "http://10.9.100.231/los-int-pbk/usecase/new-kmb/idx")
	os.Setenv("KEY_SCOREPRO_IDX_2W_JABOJABAR", "first_residence_zipcode_2w_jabo")
	os.Setenv("KEY_SCOREPRO_IDX_2W_OTHERS", "first_residence_zipcode_2w_others")
	os.Setenv("KEY_SCOREPRO_IDX_2W_AORO", "first_residence_zipcode_2w_aoro")
	os.Setenv("SCOREPRO_IDX_URL", "http://10.9.100.122:9105/api/v1/scorepro/kmb/idx")
	os.Setenv("SCOREPRO_SEGMEN_ASS_SCORE", "2")
	os.Setenv("NAMA_SAMA", "K,P")

	// Get the current time
	currentTime := time.Now().UTC()

	// Sample older date from the current time to test "RrdDate"
	sevenMonthsAgo := currentTime.AddDate(0, -7, 0)

	birthDateStr := "2000-01-01"

	birthDate, _ := time.Parse("2006-01-02", birthDateStr)

	maxOverdueDaysforActiveAgreement := 31
	numberOfPaidInstallment := 6

	testcases := []struct {
		name               string
		req                request.PrinciplePembiayaan
		principleStepOne   entity.TrxPrincipleStepOne
		principleStepTwo   entity.TrxPrincipleStepTwo
		filtering          entity.FilteringKMB
		pefindoScore       string
		customerStatus     string
		customerSegment    string
		installmentTopUp   float64
		spDupcheck         response.SpDupCekCustomerByID
		accessToken        string
		scoreGenerator     entity.ScoreGenerator
		errscoreGenerator  error
		trxDetailBiro      []entity.TrxDetailBiro
		errtrxDetailBiro   error
		codePefindoIDX     int
		bodyPefindoIDX     string
		errRespPefindoIDX  error
		codeScoreproIDX    int
		bodyScoreproIDX    string
		errRespScoreproIDX error
		responseScs        response.IntegratorScorePro
		data               response.ScorePro
		respPefindoIDX     response.PefindoIDX
		err                error
		result             response.ScorePro
		errResult          error
		config             entity.AppConfig
		errGetConfig       error
	}{
		{
			name: "scorepro jabo",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro jabo nama sama no pefindo reject",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFM0TSTRT87183109505","score":200,"result":"REJECT","score_result":"LOW","status":"ASS-LOW","phone_number":"0817344026205","segmen":"1","is_tsi":false,"score_bin":"","deviasi":null},"errors":null,"server_time":"2023-11-07T17:48:36+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFM0TSTRT87183109505","score":200,"result":"REJECT","score_band":"","score_result":"LOW","status":"ASS-LOW","segmen":"1","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda low",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk reject",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			pefindoScore:   "VERY HIGH RISK",
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "VERY HIGH RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk pass",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			pefindoScore:   "VERY HIGH RISK",
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "VERY HIGH RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk segmen 1 score low",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			pefindoScore:   "VERY HIGH RISK",
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "VERY HIGH RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"1","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"1","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk segmen 1 score high",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_NEW,
			pefindoScore:   "VERY HIGH RISK",
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "VERY HIGH RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"13","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"13","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro others",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_others",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro ro prime",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 1000000,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus:  constant.STATUS_KONSUMEN_RO,
			customerSegment: constant.RO_AO_PRIME,
			codePefindoIDX:  200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "CR perbaikan flow RO PrimePriority PASS",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus:                   constant.STATUS_KONSUMEN_RO,
				CustomerID:                       "123456",
				MaxOverdueDaysforActiveAgreement: &maxOverdueDaysforActiveAgreement,
			},
			filtering: entity.FilteringKMB{
				RrdDate:   sevenMonthsAgo,
				CreatedAt: currentTime,
			},
			customerStatus:  constant.STATUS_KONSUMEN_RO,
			customerSegment: constant.RO_AO_PRIME,
			codePefindoIDX:  200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			config: entity.AppConfig{
				Key:   "expired_contract_check",
				Value: `{"data":{"expired_contract_check_enabled":true,"expired_contract_max_months":6}}`,
			},
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "CR perbaikan flow RO PrimePriority RrdDate NULL",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus:                   constant.STATUS_KONSUMEN_RO,
				CustomerID:                       "123456",
				MaxOverdueDaysforActiveAgreement: &maxOverdueDaysforActiveAgreement,
			},
			customerStatus:  constant.STATUS_KONSUMEN_RO,
			customerSegment: constant.RO_AO_PRIME,
			filtering: entity.FilteringKMB{
				RrdDate:   nil,
				CreatedAt: currentTime,
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 500,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			errResult:       errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty"),
		},
		{
			name: "scorepro ao prime",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus:          constant.STATUS_KONSUMEN_AO,
				CustomerID:              "123456",
				NumberOfPaidInstallment: &numberOfPaidInstallment,
			},
			customerStatus:  constant.STATUS_KONSUMEN_AO,
			customerSegment: constant.RO_AO_PRIME,
			codePefindoIDX:  200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda low",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			trxDetailBiro: []entity.TrxDetailBiro{
				{
					Score:   "LOW RISK",
					Subject: constant.CUSTOMER,
					BiroID:  "123456",
				},
				{
					Score:   "LOW RISK",
					Subject: constant.SPOUSE,
					BiroID:  "234567",
				},
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda low no pefindo reject",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda low no pefindo pass",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo pass",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo pass bukan ASS-SCORE",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"2","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"2","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo reject ASS-SCORE",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_result":"HIGH","status":"ASS-HIGH","phone_number":"085716728933","segmen":"2","is_tsi":false,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_band":"","score_result":"HIGH","status":"ASS-HIGH","segmen":"2","is_tsi":false,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo tsi reject",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_result":"MEDIUM","status":"ASSTSH-S04","phone_number":"085716728933","segmen":"4","is_tsi":true,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result:    constant.DECISION_REJECT,
				Code:      constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason:    constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source:    constant.SOURCE_DECISION_SCOREPRO,
				Info:      `{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_band":"","score_result":"MEDIUM","status":"ASSTSH-S04","segmen":"4","is_tsi":true,"score_bin":"","deviasi":null}`,
				IsDeviasi: false,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo tsi pass",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "K",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_result":"MEDIUM","status":"ASSTSH-S04","phone_number":"085716728933","segmen":"4","is_tsi":true,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_band":"","score_result":"MEDIUM","status":"ASSTSH-S04","segmen":"4","is_tsi":true,"score_bin":"","deviasi":null}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo tsi pass",
			req: request.PrinciplePembiayaan{
				ProspectID: "SAL-123",
				NTF:        1000000,
				OTR:        800000,
				Tenor:      12,
			},
			principleStepOne: entity.TrxPrincipleStepOne{
				ResidenceZipCode: "12908",
				BPKBName:         "KK",
				ManufactureYear:  "2021",
				HomeStatus:       "K",
			},
			principleStepTwo: entity.TrxPrincipleStepTwo{
				MobilePhone:          "08123456789",
				Gender:               "M",
				MaritalStatus:        "M",
				ProfessionID:         "KRYS",
				BirthDate:            birthDate,
				EmploymentSinceMonth: 10,
				EmploymentSinceYear:  2016,
			},
			installmentTopUp: 0,
			spDupcheck: response.SpDupCekCustomerByID{
				CustomerStatus: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			customerStatus: constant.STATUS_KONSUMEN_RO,
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":"800-899","result":"PASS","score_result":"MEDIUM","status":"ASSTSH-S04","phone_number":"085716728933","segmen":"4","is_tsi":true,"score_band":"","score_bin":"","deviasi":null},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":"800-899","result":"PASS","score_band":"","score_result":"MEDIUM","status":"ASSTSH-S04","segmen":"4","is_tsi":true,"score_bin":"","deviasi":null}`,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetScoreGenerator", mock.Anything).Return(tc.scoreGenerator, tc.errscoreGenerator)
			mockRepository.On("GetScoreGeneratorROAO").Return(tc.scoreGenerator, tc.errscoreGenerator)
			mockRepository.On("GetTrxDetailBIro", tc.req.ProspectID).Return(tc.trxDetailBiro, tc.errtrxDetailBiro)
			mockRepository.On("GetActiveLoanTypeLast6M", tc.spDupcheck.CustomerID.(string)).Return(entity.GetActiveLoanTypeLast6M{}, nil)
			mockRepository.On("GetActiveLoanTypeLast24M", tc.spDupcheck.CustomerID.(string)).Return(entity.GetActiveLoanTypeLast24M{}, nil)
			mockRepository.On("GetMoblast", tc.spDupcheck.CustomerID.(string)).Return(entity.GetMoblast{}, nil)
			mockRepository.On("GetConfig", "expired_contract", "KMB-OFF", "expired_contract_check").Return(tc.config, tc.errGetConfig)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("PEFINDO_IDX_URL"), httpmock.NewStringResponder(tc.codePefindoIDX, tc.bodyPefindoIDX))
			resp, _ := rst.R().Post(os.Getenv("PEFINDO_IDX_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("PEFINDO_IDX_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.req.ProspectID, tc.accessToken).Return(resp, tc.errRespPefindoIDX).Once()

			rst = resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("SCOREPRO_IDX_URL"), httpmock.NewStringResponder(tc.codeScoreproIDX, tc.bodyScoreproIDX))
			resp, _ = rst.R().Post(os.Getenv("SCOREPRO_IDX_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("SCOREPRO_IDX_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.req.ProspectID, tc.accessToken).Return(resp, tc.errRespScoreproIDX).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			_, data, _, err := usecase.Scorepro(ctx, tc.req, tc.principleStepOne, tc.principleStepTwo, tc.pefindoScore, tc.customerStatus, tc.customerSegment, tc.installmentTopUp, tc.spDupcheck, tc.filtering, tc.accessToken)
			require.Equal(t, tc.result, data)
			require.Equal(t, tc.errResult, err)
		})
	}
}

func TestDsrCheck(t *testing.T) {
	// Set environment variables
	os.Setenv("INSTALLMENT_PENDING_URL", "http://localhost/")
	os.Setenv("DEFAULT_TIMEOUT_60S", "60")

	var (
		idNumber   = "3203014309612346"
		legalName  = "TEST USER"
		birthDate  = "1990-01-01"
		motherName = "TEST MOTHER"
		adminFee   = float64(10000)
	)

	config := response.DupcheckConfig{
		Data: response.DataDupcheckConfig{
			MaxDsr: 35,
			MinimumPencairanROTopUp: struct {
				Prime    float64 `json:"prime"`
				Priority float64 `json:"priority"`
				Regular  float64 `json:"regular"`
			}{
				Prime:    20,
				Priority: 30,
				Regular:  30,
			},
		},
	}

	// Test customer data sets
	newCustomer := []request.CustomerData{{
		StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
		IDNumber:       idNumber,
		LegalName:      legalName,
		BirthDate:      birthDate,
		MotherName:     motherName,
	}}

	roCustomer := []request.CustomerData{{
		StatusKonsumen: constant.STATUS_KONSUMEN_RO,
		IDNumber:       idNumber,
		LegalName:      legalName,
		BirthDate:      birthDate,
		MotherName:     motherName,
	}}

	primeCustomer := []request.CustomerData{{
		StatusKonsumen:  constant.STATUS_KONSUMEN_RO,
		CustomerSegment: constant.RO_AO_PRIME,
		IDNumber:        idNumber,
		LegalName:       legalName,
		BirthDate:       birthDate,
		MotherName:      motherName,
	}}

	regularCustomer := []request.CustomerData{{
		StatusKonsumen:  constant.STATUS_KONSUMEN_RO,
		CustomerSegment: "REGULAR",
		IDNumber:        idNumber,
		LegalName:       legalName,
		BirthDate:       birthDate,
		MotherName:      motherName,
	}}

	aoRegularCustomer := []request.CustomerData{{
		StatusKonsumen:  constant.STATUS_KONSUMEN_AO,
		CustomerSegment: "REGULAR",
		IDNumber:        idNumber,
		LegalName:       legalName,
		BirthDate:       birthDate,
		MotherName:      motherName,
	}}

	aoPrimeCustomer := []request.CustomerData{{
		StatusKonsumen:  constant.STATUS_KONSUMEN_AO,
		CustomerSegment: constant.RO_AO_PRIME,
		IDNumber:        idNumber,
		LegalName:       legalName,
		BirthDate:       birthDate,
		MotherName:      motherName,
	}}

	testCases := []struct {
		name               string
		customerData       []request.CustomerData
		request            request.PrinciplePembiayaan
		installmentAmount  float64
		installmentConfins float64
		income             float64
		agreementResponse  response.AgreementChassisNumber
		httpStatus         int
		httpResponse       string
		httpError          error
		expectedResult     response.UsecaseApi
		expectedError      error
	}{
		{
			name:         "pass new dsr <= threshold",
			customerData: newCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
			},
			installmentAmount: 1000000,
			income:            5000000,
			httpStatus:        200,
			httpResponse:      `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "NEW - Confins DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            20,
			},
		},
		{
			name:         "reject new dsr > threshold",
			customerData: newCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
			},
			installmentAmount: 2000000,
			income:            5000000,
			httpStatus:        200,
			httpResponse:      `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_DSRGT35,
				Reason:         "NEW - Confins DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            40,
			},
		},
		{
			name:         "pass ro prime",
			customerData: primeCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
			},
			installmentAmount: 2000000,
			income:            5000000,
			httpStatus:        200,
			httpResponse:      `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "RO PRIME",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            40,
			},
		},
		{
			name:         "reject ro regular dsr > threshold",
			customerData: regularCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
			},
			installmentAmount:  2000000,
			installmentConfins: 1000000,
			income:             5000000,
			httpStatus:         200,
			httpResponse:       `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_DSRGT35,
				Reason:         "RO - Confins DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            60,
			},
		},
		{
			name:         "error http",
			customerData: newCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
			},
			httpError:     errors.New("connection error"),
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Installment Pending API Error"),
		},
		{
			name:         "error not success",
			customerData: newCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
			},
			httpStatus:    500,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Installment Pending API Error"),
		},
		{
			name:         "pass top up",
			customerData: roCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
				OTR:        100000000,
				DPAmount:   20000000,
				AdminFee:   adminFee,
				Dealer:     constant.DEALER_PSA,
			},
			installmentAmount:  1000000,
			installmentConfins: 2000000,
			income:             10000000,
			agreementResponse: response.AgreementChassisNumber{
				InstallmentAmount:    1500000,
				OutstandingPrincipal: 50000000,
				OutstandingInterest:  5000000,
				LcInstallment:        100000,
			},
			httpStatus:   200,
			httpResponse: `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "RO Top Up - Confins DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            15,
			},
		},
		{
			name:         "reject ro top up minimum pencairan < threshold",
			customerData: regularCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
				OTR:        100000000,
				DPAmount:   20000000,
				AdminFee:   adminFee,
				Dealer:     constant.DEALER_PSA,
			},
			installmentAmount:  1000000,
			installmentConfins: 2000000,
			income:             10000000,
			agreementResponse: response.AgreementChassisNumber{
				InstallmentAmount:    1500000,
				OutstandingPrincipal: 70000000,
				OutstandingInterest:  5000000,
				LcInstallment:        100000,
			},
			httpStatus:   200,
			httpResponse: `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PENCAIRAN_TOPUP,
				Reason:         "RO Top Up Persentase Minimum Pencairan yang diterima kurang dari Threshold",
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				Dsr:            15,
			},
		},
		{
			name:         "reject ro prime top up minimum pencairan < threshold",
			customerData: primeCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
				OTR:        100000000,
				DPAmount:   20000000,
				AdminFee:   adminFee,
				Dealer:     constant.DEALER_PSA,
			},
			installmentAmount:  1000000,
			installmentConfins: 2000000,
			income:             10000000,
			agreementResponse: response.AgreementChassisNumber{
				InstallmentAmount:    1500000,
				OutstandingPrincipal: 75000000,
				OutstandingInterest:  4000000,
				LcInstallment:        100000,
			},
			httpStatus:   200,
			httpResponse: `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_PENCAIRAN_TOPUP,
				Reason:         "RO PRIME Top Up Persentase Minimum Pencairan yang diterima kurang dari Threshold",
				SourceDecision: constant.SOURCE_DECISION_DUPCHECK,
				Dsr:            15,
			},
		},
		{
			name:         "pass ao prime top up",
			customerData: aoPrimeCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
				OTR:        100000000,
				DPAmount:   20000000,
				AdminFee:   adminFee,
				Dealer:     constant.DEALER_PSA,
			},
			installmentAmount:  1000000,
			installmentConfins: 2000000,
			income:             10000000,
			agreementResponse: response.AgreementChassisNumber{
				InstallmentAmount:    1500000,
				OutstandingPrincipal: 50000000,
				OutstandingInterest:  5000000,
				LcInstallment:        100000,
			},
			httpStatus:   200,
			httpResponse: `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "AO PRIME Top Up",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            15,
			},
		},
		{
			name:         "reject ao prime dsr > threshold",
			customerData: aoPrimeCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
				OTR:        100000000,
				DPAmount:   20000000,
				AdminFee:   adminFee,
				Dealer:     constant.DEALER_PSA,
			},
			installmentAmount:  2000000,
			installmentConfins: 2000000,
			income:             5000000,
			agreementResponse:  response.AgreementChassisNumber{},
			httpStatus:         200,
			httpResponse:       `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_DSRGT35,
				Reason:         "AO PRIME - Confins DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            80,
			},
		},
		{
			name:         "reject ao regular dsr > threshold",
			customerData: aoRegularCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
				OTR:        100000000,
				DPAmount:   20000000,
				AdminFee:   adminFee,
				Dealer:     constant.DEALER_PSA,
			},
			installmentAmount:  2000000,
			installmentConfins: 2000000,
			income:             5000000,
			agreementResponse:  response.AgreementChassisNumber{},
			httpStatus:         200,
			httpResponse:       `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_DSRGT35,
				Reason:         "AO - Confins DSR > Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            80,
			},
		},
		{
			name:         "pass ao regular",
			customerData: aoRegularCustomer,
			request: request.PrinciplePembiayaan{
				ProspectID: "TEST123",
				OTR:        100000000,
				DPAmount:   20000000,
				AdminFee:   adminFee,
				Dealer:     constant.DEALER_PSA,
			},
			installmentAmount:  500000,
			installmentConfins: 500000,
			income:             5000000,
			agreementResponse:  response.AgreementChassisNumber{},
			httpStatus:         200,
			httpResponse:       `{"data": {}}`,
			expectedResult: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_DSRLTE35,
				Reason:         "AO - Confins DSR <= Threshold",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Dsr:            20,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockResp := &resty.Response{}
			if tc.httpResponse != "" {
				mockResp = &resty.Response{
					RawResponse: &http.Response{
						StatusCode: tc.httpStatus,
					},
				}
				mockResp.SetBody([]byte(tc.httpResponse))
			}

			// Set up expectations
			mockHttpClient.On("EngineAPI",
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything,
				mock.Anything).Return(mockResp, tc.httpError)

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			result, _, _, _, _, err := usecase.DsrCheck(
				ctx,
				tc.request,
				tc.customerData,
				tc.installmentAmount,
				tc.installmentConfins,
				0, // installmentConfinsSpouse
				tc.income,
				tc.agreementResponse,
				"test-token",
				config,
			)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}

			mockHttpClient.AssertExpectations(t)
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
				IsDeviasi:      false,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				Dsr:            30,
				CustomerID:     "123456",
				ConfigMaxDSR:   35,
			},
			filtering: entity.FilteringKMB{
				RrdDate:   sixMonthsAgo,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				Dsr:                              30,
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
				CustomerID:                       "123456",
				ConfigMaxDSR:                     35,
			},
			filtering: entity.FilteringKMB{
				RrdDate:   sevenMonthsAgo,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				Dsr:                              30,
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
				CustomerID:                       "123456",
				ConfigMaxDSR:                     35,
			},
			filtering: entity.FilteringKMB{
				RrdDate:   nil,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{
				RrdDate: sixMonthsAgo,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{
				RrdDate: sixMonthsAgo,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{
				RrdDate: sixMonthsAgo,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{
				RrdDate: sixMonthsAgo,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
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
			configValue: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35,"minimum_pencairan_ro_top_up":5000000}}`,
			},
			filtering: entity.FilteringKMB{
				RrdDate: sixMonthsAgo,
			},
			SpDupcheckMap: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
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

			mockRepository.On("GetBranchDeviasi", "400").Return(tc.mappingBranchDeviasi, tc.errGetConfig)

			mockRepository.On("MasterMappingDeviasiDSR", tc.totalIncome).Return(tc.mappingDeviasiDSR, tc.errGetConfig)

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("LASTEST_PAID_INSTALLMENT_URL")+tc.SpDupcheckMap.CustomerID.(string)+"/2", mock.Anything, mock.Anything, constant.METHOD_GET, false, 0, 30, tc.prospectID, tc.accessToken).Return(resp, tc.errLatestInstallment).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			data, trx, err := usecase.TotalDsrFmfPbk(ctx, tc.totalIncome, tc.newInstallment, tc.totalInstallmentPBK, tc.prospectID, tc.customerSegment, tc.accessToken, tc.SpDupcheckMap, configValue, tc.filtering)

			require.Equal(t, tc.result, data)
			require.Equal(t, tc.trxFMF, trx)
			require.Equal(t, tc.errResult, err)
		})
	}
}

func TestAgreementChassisNumberIntegrator(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_60S", "60")

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_60S"))
	accessToken := "test_token"
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	prospectID := "SAL-123456"
	chassisNumber := "ABCD1234567890"

	testcases := []struct {
		name               string
		rChassisNumberCode int
		rChassisNumberBody string
		errHttp            error
		resChassisNumber   response.AgreementChassisNumber
		errFinal           error
	}{
		{
			name:     "test HTTP error",
			errHttp:  errors.New("upstream_service_timeout - DsrCheck Call Get Agreement of Chassis Number Timeout"),
			errFinal: errors.New("upstream_service_timeout - DsrCheck Call Get Agreement of Chassis Number Timeout"),
		},
		{
			name:               "test non-200 response",
			rChassisNumberCode: 400,
			rChassisNumberBody: `{"error": "Bad Request"}`,
			errFinal:           errors.New("upstream_service_error - DsrCheck Call Get Agreement of Chassis Number Error"),
		},
		{
			name:               "test unmarshal error",
			rChassisNumberCode: 200,
			rChassisNumberBody: `{"data": {"go_live_date": true, "installment_amount": "invalid"}}`, // Intentionally wrong types
			errFinal:           errors.New("upstream_service_error - DsrCheck Unmarshal Get Agreement of Chassis Number Error"),
		},
		{
			name:               "test successful response",
			rChassisNumberCode: 200,
			rChassisNumberBody: `{"data": {"go_live_date": "2024-01-01", "id_number": "1234567890", "installment_amount": 1500000, "is_active": true, "is_registered": true, "lc_installment": 1600000, "legal_name": "John Doe", "outstanding_interest": 5000000, "outstanding_principal": 65000000, "status": "ACTIVE"}}`,
			resChassisNumber: response.AgreementChassisNumber{
				GoLiveDate:           "2024-01-01",
				IDNumber:             "1234567890",
				InstallmentAmount:    1500000,
				IsActive:             true,
				IsRegistered:         true,
				LcInstallment:        1600000,
				LegalName:            "John Doe",
				OutstandingInterest:  5000000,
				OutstandingPrincipal: 65000000,
				Status:               "ACTIVE",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockResp := &resty.Response{}
			if tc.rChassisNumberBody != "" {
				mockResp = &resty.Response{
					RawResponse: &http.Response{
						StatusCode: tc.rChassisNumberCode,
					},
				}
				mockResp.SetBody([]byte(tc.rChassisNumberBody))
			}

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG,
				os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+chassisNumber,
				mock.Anything, map[string]string{}, constant.METHOD_GET, true, 2, timeout,
				prospectID, accessToken).Return(mockResp, tc.errHttp)

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			result, err := usecase.AgreementChassisNumberIntegrator(ctx, prospectID, chassisNumber, accessToken)

			if tc.errFinal != nil {
				require.Error(t, err)
				require.Equal(t, tc.errFinal.Error(), err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.resChassisNumber, result)
			}

			mockHttpClient.AssertExpectations(t)
		})
	}
}
