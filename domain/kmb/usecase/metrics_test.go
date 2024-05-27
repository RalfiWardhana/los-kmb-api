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
	"los-kmb-api/shared/utils"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	os.Setenv("BIRO_VALID_DAYS", "4")
	os.Setenv("INTERNAL_RECORD_URL", "http://localhost/")
	MonthlyVariableIncome := float64(10000000)
	SpouseIncome := float64(0)
	ctx := context.Background()

	info, _ := json.Marshal(response.SpDupcheckMap{
		CustomerID: "123456",
	})

	testcases := []struct {
		name                      string
		reqMetrics                request.Metrics
		resultMetrics             interface{}
		err                       error
		trxMaster                 int
		errScanTrxMaster          error
		countTrx                  int
		errScanTrxPrescreening    error
		filtering                 entity.FilteringKMB
		errGetFilteringResult     error
		errGetFilteringForJourney error
		errGetElaborateLtv        error
		details                   []entity.TrxDetail
		errSaveTransaction        error
		trxPrescreening           entity.TrxPrescreening
		trxFMF                    response.TrxFMF
		errPrescreening           error
		trxPrescreeningDetail     entity.TrxDetail
		trxTenor                  response.UsecaseApi
		errRejectTenor36          error
		config                    entity.AppConfig
		errGetConfig              error
		configValue               response.DupcheckConfig
		dupcheckData              response.SpDupcheckMap
		trxFMFDupcheck            response.TrxFMF
		trxDetailDupcheck         []entity.TrxDetail
		customerStatus            string
		customerSegment           string
		metricsDupcheck           response.UsecaseApi
		errDupcheck               error
		codeInternalRecord        int
		bodyInternalRecord        string
		errInternalRecord         error
		mappingCluster            entity.MasterMappingCluster
		mappingMaxDSR             entity.MasterMappingMaxDSR
		errmappingCluster         error
		errmappingMaxDSR          error
	}{
		{
			name: "test metrics errScanTrxMaster",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:              errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error"),
			errScanTrxMaster: errors.New(constant.ERROR_UPSTREAM + " - Get Transaction Error"),
		},
		{
			name: "test metrics trxMaster > 0",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:       errors.New(constant.ERROR_BAD_REQUEST + " - ProspectID Already Exist"),
			trxMaster: 1,
		},
		{
			name: "test metrics ScanTrxPrescreening",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:                    errors.New(constant.ERROR_UPSTREAM + " - Get Prescreening Error"),
			trxMaster:              0,
			errScanTrxPrescreening: errors.New(constant.ERROR_UPSTREAM + " - Get Prescreening Error"),
		},
		{
			name: "test metrics errGetFilteringResult Belum melakukan filtering",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:                   errors.New(fmt.Sprintf("%s - Belum melakukan filtering atau hasil filtering sudah lebih dari %s hari", constant.ERROR_BAD_REQUEST, os.Getenv("BIRO_VALID_DAYS"))),
			trxMaster:             0,
			errGetFilteringResult: errors.New(constant.RECORD_NOT_FOUND),
		},
		{
			name: "test metrics errGetFilteringResult selain Belum melakukan filtering",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:                   errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
			trxMaster:             0,
			errGetFilteringResult: errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
		},
		{
			name: "test metrics errGetFilteringResult Tidak bisa lanjut proses",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:       errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses"),
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    0,
				CustomerStatus: "NEW",
			},
		},
		{
			name: "test metrics errGetFilteringForJourney",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:       errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			errGetFilteringForJourney: errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error"),
		},
		{
			name: "test metrics errGetElaborateLtv",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			err:       errors.New(constant.ERROR_BAD_REQUEST + " - Belum melakukan pengecekan LTV"),
			trxMaster: 0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				CustomerStatus: "NEW",
			},
			errGetElaborateLtv: errors.New(constant.ERROR_BAD_REQUEST + " - Belum melakukan pengecekan LTV"),
		},
		{
			name: "test metrics CMO not recommend errSaveTransactionSaveTransaction",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Agent: request.Agent{
					CmoRecom: constant.CMO_NOT_RECOMMEDED,
				},
			},
			err:       errors.New(constant.ERROR_UPSTREAM + " - Save Transaction Error"),
			trxMaster: 0,
			countTrx:  0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
					Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
					SourceDecision: constant.CMO_AGENT,
					NextStep:       constant.PRESCREENING,
				},
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					SourceDecision: constant.PRESCREENING,
					RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
					Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
					CreatedBy:      constant.SYSTEM_CREATED,
				},
			},
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			resultMetrics:      response.Metrics{},
			errSaveTransaction: errors.New(constant.ERROR_UPSTREAM + " - Save Transaction Error"),
		},
		{
			name: "test metrics CMO not recommend save success",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Agent: request.Agent{
					CmoRecom: constant.CMO_NOT_RECOMMEDED,
				},
			},
			trxMaster: 0,
			countTrx:  0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
					Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
					SourceDecision: constant.CMO_AGENT,
					NextStep:       constant.PRESCREENING,
				},
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					SourceDecision: constant.PRESCREENING,
					RuleCode:       constant.CODE_CMO_NOT_RECOMMEDED,
					Reason:         constant.REASON_CMO_NOT_RECOMMENDED,
					CreatedBy:      constant.SYSTEM_CREATED,
				},
			},
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			resultMetrics: response.Metrics{},
		},
		{
			name: "test metrics errPrescreening",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster: 0,
			countTrx:  0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       constant.CODE_CMO_RECOMMENDED,
					Reason:         constant.REASON_CMO_RECOMMENDED,
					SourceDecision: constant.CMO_AGENT,
					NextStep:       constant.PRESCREENING,
				},
			},
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			err:             errors.New("error prescreening"),
			errPrescreening: errors.New("error prescreening"),
		},
		{
			name: "test metrics Prescreening errSaveTransaction",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster: 0,
			countTrx:  0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       constant.CODE_CMO_RECOMMENDED,
					Reason:         constant.REASON_CMO_RECOMMENDED,
					SourceDecision: constant.CMO_AGENT,
					NextStep:       constant.PRESCREENING,
				},
				{},
			},
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			resultMetrics:      response.Metrics{},
			err:                errors.New("error prescreening"),
			errSaveTransaction: errors.New("error prescreening"),
		},
		{
			name: "test metrics Prescreening save success",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
			},
			trxMaster: 0,
			countTrx:  0,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       constant.CODE_CMO_RECOMMENDED,
					Reason:         constant.REASON_CMO_RECOMMENDED,
					SourceDecision: constant.CMO_AGENT,
					NextStep:       constant.PRESCREENING,
				},
				{},
			},
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			resultMetrics: response.Metrics{},
		},
		{
			name: "test metrics tenor 36 err",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       constant.CODE_CMO_RECOMMENDED,
					Reason:         constant.REASON_CMO_RECOMMENDED,
					SourceDecision: constant.CMO_AGENT,
					NextStep:       constant.PRESCREENING,
				},
				{},
			},
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			err:              errors.New("error reject tenor 36"),
			errRejectTenor36: errors.New("error reject tenor 36"),
		},
		{
			name: "test metrics tenor 36 err cluster CMO",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
				CMOCluster:     "Cluster C",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       constant.CODE_CMO_RECOMMENDED,
					Reason:         constant.REASON_CMO_RECOMMENDED,
					SourceDecision: constant.CMO_AGENT,
					NextStep:       constant.PRESCREENING,
				},
				{},
			},
			trxPrescreening: entity.TrxPrescreening{
				ProspectID: "TEST1",
				Decision:   constant.DB_DECISION_REJECT,
				Reason:     constant.REASON_CMO_NOT_RECOMMENDED,
				CreatedBy:  constant.SYSTEM_CREATED,
				DecisionBy: constant.SYSTEM_CREATED,
			},
			err:              errors.New("error reject tenor 36"),
			errRejectTenor36: errors.New("error reject tenor 36"),
		},
		{
			name: "test metrics tenor 36 reject errsave",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "REJECT TENOR 36",
					Info:           fmt.Sprintf("Cluster : "),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_REJECT,
				Reason: "REJECT TENOR 36",
			},
			resultMetrics:      response.Metrics{},
			err:                errors.New("error save"),
			errSaveTransaction: errors.New("error save"),
		},
		{
			name: "test metrics tenor 36 reject errsave Cluster CMO",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
				CMOCluster:     "Cluster C",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "REJECT TENOR 36",
					Info:           fmt.Sprintf("Cluster : Cluster C"),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_REJECT,
				Reason: "REJECT TENOR 36",
			},
			resultMetrics:      response.Metrics{},
			err:                errors.New("error save"),
			errSaveTransaction: errors.New("error save"),
		},
		{
			name: "test metrics tenor 36 reject",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 48,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "013",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         constant.REASON_REJECT_TENOR,
					Info:           fmt.Sprintf("Cluster : "),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "013",
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_TENOR,
			},
			resultMetrics: response.Metrics{},
		},
		{
			name: "test metrics tenor 36 reject Cluster CMO",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 48,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
				CMOCluster:     "Cluster C",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "013",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         constant.REASON_REJECT_TENOR,
					Info:           fmt.Sprintf("Cluster : Cluster C"),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "013",
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_TENOR,
			},
			resultMetrics: response.Metrics{},
		},
		{
			name: "test metrics tenor errGetConfig",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			err:          errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Config Error"),
			errGetConfig: errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Config Error"),
		},
		{
			name: "test metrics tenor errGetConfig Cluster CMO",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:    1,
				ScoreBiro:      "AVERAGE RISK",
				CustomerStatus: "NEW",
				CMOCluster:     "Cluster C",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			err:          errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Config Error"),
			errGetConfig: errors.New(constant.ERROR_UPSTREAM + " - Get Dupcheck Config Error"),
		},
		{
			name: "test metrics err Dupcheck ",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:     1,
				ScoreBiro:       "AVERAGE RISK",
				CustomerSegment: constant.RO_AO_PRIME,
				CustomerStatus:  "RO",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			mappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			mappingMaxDSR: entity.MasterMappingMaxDSR{
				DSRThreshold: 35,
			},
			config: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
			},
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					VehicleAge: 17,
					MaxOvd:     60,
					MaxDsr:     35,
				},
			},
			err:         errors.New(constant.ERROR_UPSTREAM + " - errDupcheck Error"),
			errDupcheck: errors.New(constant.ERROR_UPSTREAM + " - errDupcheck Error"),
		},
		{
			name: "test metrics err internal record ",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:     1,
				ScoreBiro:       "AVERAGE RISK",
				CustomerSegment: constant.RO_AO_PRIME,
				CustomerStatus:  "RO",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			mappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			mappingMaxDSR: entity.MasterMappingMaxDSR{
				DSRThreshold: 35,
			},
			config: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
			},
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					VehicleAge: 17,
					MaxOvd:     60,
					MaxDsr:     35,
				},
			},
			dupcheckData: response.SpDupcheckMap{
				CustomerID: "123456",
			},
			err:               errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Interal Record Error"),
			errInternalRecord: errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Get Interal Record Error"),
		},
		{
			name: "test metrics dupcheck err save ",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:     1,
				ScoreBiro:       "AVERAGE RISK",
				CustomerSegment: constant.RO_AO_PRIME,
				CustomerStatus:  "RO",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
					Info:           fmt.Sprintf("Cluster : "),
				},
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "123",
					SourceDecision: "DUPCHECK",
					Reason:         "dupcheck reject",
					Info:           string(utils.SafeEncoding(info)),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			mappingCluster: entity.MasterMappingCluster{},
			mappingMaxDSR: entity.MasterMappingMaxDSR{
				DSRThreshold: 35,
			},
			config: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
			},
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					VehicleAge: 17,
					MaxOvd:     60,
					MaxDsr:     35,
				},
			},
			dupcheckData: response.SpDupcheckMap{
				CustomerID: "123456",
			},
			metricsDupcheck: response.UsecaseApi{
				Code:           "123",
				Result:         constant.DECISION_REJECT,
				Reason:         "dupcheck reject",
				SourceDecision: "DUPCHECK",
			},
			trxFMF: response.TrxFMF{
				DupcheckData: response.SpDupcheckMap{
					CustomerID: "123456",
				},
				DSRFMF: float64(0),
				AgreementCONFINS: []response.AgreementCONFINS{
					{
						ApplicationID:     "426A202212124023",
						ProductType:       "WG",
						AgreementDate:     "12/20/2022",
						AssetCode:         "APPLE.HP/SMARTPHONES.HPIPHONE6S32",
						Tenor:             12,
						InstallmentAmount: 371000,
						ContractStatus:    "EXP",
						CurrentCondition:  "Current",
					},
					{
						ApplicationID:     "426A202306124184",
						ProductType:       "WG",
						AgreementDate:     "06/20/2023",
						AssetCode:         "OPPO.HP/SMARTPHONES.PHABLETF56GB",
						Tenor:             11,
						InstallmentAmount: 161000,
						ContractStatus:    "LIV",
						CurrentCondition:  "Current",
					},
				},
			},
			codeInternalRecord: 200,
			bodyInternalRecord: `{"messages":"LOS-ListAgreements","errors":null,"data":[{"application_id":"426A202212124023","product_type":"WG","agreement_date":"12/20/2022","asset_code":"APPLE.HP/SMARTPHONES.HPIPHONE6S32","period":12,"outstanding_principal":0,"installment_amount":371000,"contract_status":"EXP","current_condition":"Current"},{"application_id":"426A202306124184","product_type":"WG","agreement_date":"06/20/2023","asset_code":"OPPO.HP/SMARTPHONES.PHABLETF56GB","period":11,"outstanding_principal":0,"installment_amount":161000,"contract_status":"LIV","current_condition":"Current"}],"server_time":"2023-11-27T17:00:19+07:00","request_id":"db8b5f93-242d-4b9b-9932-0c8e99bcac00"}`,
			resultMetrics:      response.Metrics{},
			err:                errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - metricsDupcheck Error"),
			errSaveTransaction: errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - metricsDupcheck Error"),
		},
		{
			name: "test metrics dupcheck err mapping cluster ",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:     1,
				ScoreBiro:       "AVERAGE RISK",
				CustomerSegment: constant.RO_AO_PRIME,
				CustomerStatus:  "RO",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
				},
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "123",
					SourceDecision: "DUPCHECK",
					Reason:         "dupcheck reject",
					Info:           string(utils.SafeEncoding(info)),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			config: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
			},
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					VehicleAge: 17,
					MaxOvd:     60,
					MaxDsr:     35,
				},
			},
			dupcheckData: response.SpDupcheckMap{
				CustomerID: "123456",
			},
			metricsDupcheck: response.UsecaseApi{
				Code:           "123",
				Result:         constant.DECISION_REJECT,
				Reason:         "dupcheck reject",
				SourceDecision: "DUPCHECK",
			},
			trxFMF: response.TrxFMF{
				DupcheckData: response.SpDupcheckMap{
					CustomerID: "123456",
				},
				DSRFMF: float64(0),
				AgreementCONFINS: []response.AgreementCONFINS{
					{
						ApplicationID:     "426A202212124023",
						ProductType:       "WG",
						AgreementDate:     "12/20/2022",
						AssetCode:         "APPLE.HP/SMARTPHONES.HPIPHONE6S32",
						Tenor:             12,
						InstallmentAmount: 371000,
						ContractStatus:    "EXP",
						CurrentCondition:  "Current",
					},
					{
						ApplicationID:     "426A202306124184",
						ProductType:       "WG",
						AgreementDate:     "06/20/2023",
						AssetCode:         "OPPO.HP/SMARTPHONES.PHABLETF56GB",
						Tenor:             11,
						InstallmentAmount: 161000,
						ContractStatus:    "LIV",
						CurrentCondition:  "Current",
					},
				},
			},
			codeInternalRecord: 200,
			bodyInternalRecord: `{"messages":"LOS-ListAgreements","errors":null,"data":[{"application_id":"426A202212124023","product_type":"WG","agreement_date":"12/20/2022","asset_code":"APPLE.HP/SMARTPHONES.HPIPHONE6S32","period":12,"outstanding_principal":0,"installment_amount":371000,"contract_status":"EXP","current_condition":"Current"},{"application_id":"426A202306124184","product_type":"WG","agreement_date":"06/20/2023","asset_code":"OPPO.HP/SMARTPHONES.PHABLETF56GB","period":11,"outstanding_principal":0,"installment_amount":161000,"contract_status":"LIV","current_condition":"Current"}],"server_time":"2023-11-27T17:00:19+07:00","request_id":"db8b5f93-242d-4b9b-9932-0c8e99bcac00"}`,
			err:                errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error"),
			errmappingCluster:  errors.New(constant.ERROR_UPSTREAM + " - Get Mapping cluster error"),
		},
		{
			name: "test metrics dupcheck err mapping max dsr ",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:     1,
				ScoreBiro:       "AVERAGE RISK",
				CustomerSegment: constant.RO_AO_PRIME,
				CustomerStatus:  "RO",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
				},
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "123",
					SourceDecision: "DUPCHECK",
					Reason:         "dupcheck reject",
					Info:           string(utils.SafeEncoding(info)),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			config: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
			},
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					VehicleAge: 17,
					MaxOvd:     60,
					MaxDsr:     35,
				},
			},
			dupcheckData: response.SpDupcheckMap{
				CustomerID: "123456",
			},
			metricsDupcheck: response.UsecaseApi{
				Code:           "123",
				Result:         constant.DECISION_REJECT,
				Reason:         "dupcheck reject",
				SourceDecision: "DUPCHECK",
			},
			trxFMF: response.TrxFMF{
				DupcheckData: response.SpDupcheckMap{
					CustomerID: "123456",
				},
				DSRFMF: float64(0),
				AgreementCONFINS: []response.AgreementCONFINS{
					{
						ApplicationID:     "426A202212124023",
						ProductType:       "WG",
						AgreementDate:     "12/20/2022",
						AssetCode:         "APPLE.HP/SMARTPHONES.HPIPHONE6S32",
						Tenor:             12,
						InstallmentAmount: 371000,
						ContractStatus:    "EXP",
						CurrentCondition:  "Current",
					},
					{
						ApplicationID:     "426A202306124184",
						ProductType:       "WG",
						AgreementDate:     "06/20/2023",
						AssetCode:         "OPPO.HP/SMARTPHONES.PHABLETF56GB",
						Tenor:             11,
						InstallmentAmount: 161000,
						ContractStatus:    "LIV",
						CurrentCondition:  "Current",
					},
				},
			},
			mappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			codeInternalRecord: 200,
			bodyInternalRecord: `{"messages":"LOS-ListAgreements","errors":null,"data":[{"application_id":"426A202212124023","product_type":"WG","agreement_date":"12/20/2022","asset_code":"APPLE.HP/SMARTPHONES.HPIPHONE6S32","period":12,"outstanding_principal":0,"installment_amount":371000,"contract_status":"EXP","current_condition":"Current"},{"application_id":"426A202306124184","product_type":"WG","agreement_date":"06/20/2023","asset_code":"OPPO.HP/SMARTPHONES.PHABLETF56GB","period":11,"outstanding_principal":0,"installment_amount":161000,"contract_status":"LIV","current_condition":"Current"}],"server_time":"2023-11-27T17:00:19+07:00","request_id":"db8b5f93-242d-4b9b-9932-0c8e99bcac00"}`,
			err:                errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Max DSR error"),
			errmappingMaxDSR:   errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Max DSR error"),
		},
		{
			name: "test metrics dupcheck reject ",
			reqMetrics: request.Metrics{
				Transaction: request.Transaction{
					ProspectID: "TEST1",
				},
				Apk: request.Apk{
					Tenor: 36,
				},
				CustomerPersonal: request.CustomerPersonal{
					IDNumber: "123456",
				},
				CustomerSpouse: &request.CustomerSpouse{
					IDNumber:  "123456",
					LegalName: "SPOUSE",
				},
				CustomerPhoto: []request.CustomerPhoto{
					{
						ID:  constant.TAG_KTP_PHOTO,
						Url: "URL KTP",
					},
					{
						ID:  constant.TAG_SELFIE_PHOTO,
						Url: "URL SELFIE",
					},
				},
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_LEGAL,
						ZipCode: "12345",
					},
					{
						Type: constant.ADDRESS_TYPE_COMPANY,
					},
				},
				CustomerEmployment: request.CustomerEmployment{
					MonthlyVariableIncome: &MonthlyVariableIncome,
					SpouseIncome:          &SpouseIncome,
				},
			},
			trxMaster: 0,
			countTrx:  1,
			filtering: entity.FilteringKMB{
				NextProcess:     1,
				ScoreBiro:       "AVERAGE RISK",
				CustomerSegment: constant.RO_AO_PRIME,
				CustomerStatus:  "RO",
			},
			details: []entity.TrxDetail{
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_ONPROCESS,
					Activity:       constant.ACTIVITY_PROCESS,
					Decision:       constant.DB_DECISION_PASS,
					RuleCode:       "123",
					SourceDecision: constant.SOURCE_DECISION_TENOR,
					NextStep:       constant.SOURCE_DECISION_DUPCHECK,
					CreatedBy:      constant.SYSTEM_CREATED,
					Reason:         "PASS TENOR 36",
					Info:           fmt.Sprintf("Cluster : Cluster A"),
				},
				{
					ProspectID:     "TEST1",
					StatusProcess:  constant.STATUS_FINAL,
					Activity:       constant.ACTIVITY_STOP,
					Decision:       constant.DB_DECISION_REJECT,
					RuleCode:       "123",
					SourceDecision: "DUPCHECK",
					Reason:         "dupcheck reject",
					Info:           string(utils.SafeEncoding(info)),
				},
			},
			trxTenor: response.UsecaseApi{
				Code:   "123",
				Result: constant.DECISION_PASS,
				Reason: "PASS TENOR 36",
			},
			config: entity.AppConfig{
				Key:   "parameterize",
				Value: `{"data":{"vehicle_age":17,"max_ovd":60,"max_dsr":35}}`,
			},
			configValue: response.DupcheckConfig{
				Data: response.DataDupcheckConfig{
					VehicleAge: 17,
					MaxOvd:     60,
					MaxDsr:     35,
				},
			},
			mappingCluster: entity.MasterMappingCluster{
				Cluster: "Cluster A",
			},
			mappingMaxDSR: entity.MasterMappingMaxDSR{
				DSRThreshold: 35,
			},
			dupcheckData: response.SpDupcheckMap{
				CustomerID: "123456",
			},
			codeInternalRecord: 200,
			bodyInternalRecord: `{"messages":"LOS-ListAgreements","errors":null,"data":[{"application_id":"426A202212124023","product_type":"WG","agreement_date":"12/20/2022","asset_code":"APPLE.HP/SMARTPHONES.HPIPHONE6S32","period":12,"outstanding_principal":0,"installment_amount":371000,"contract_status":"EXP","current_condition":"Current"},{"application_id":"426A202306124184","product_type":"WG","agreement_date":"06/20/2023","asset_code":"OPPO.HP/SMARTPHONES.PHABLETF56GB","period":11,"outstanding_principal":0,"installment_amount":161000,"contract_status":"LIV","current_condition":"Current"}],"server_time":"2023-11-27T17:00:19+07:00","request_id":"db8b5f93-242d-4b9b-9932-0c8e99bcac00"}`,
			metricsDupcheck: response.UsecaseApi{
				Code:           "123",
				Result:         constant.DECISION_REJECT,
				Reason:         "dupcheck reject",
				SourceDecision: "DUPCHECK",
			},
			trxFMF: response.TrxFMF{
				DupcheckData: response.SpDupcheckMap{
					CustomerID: "123456",
				},
				DSRFMF: float64(0),
				AgreementCONFINS: []response.AgreementCONFINS{
					{
						ApplicationID:     "426A202212124023",
						ProductType:       "WG",
						AgreementDate:     "12/20/2022",
						AssetCode:         "APPLE.HP/SMARTPHONES.HPIPHONE6S32",
						Tenor:             12,
						InstallmentAmount: 371000,
						ContractStatus:    "EXP",
						CurrentCondition:  "Current",
					},
					{
						ApplicationID:     "426A202306124184",
						ProductType:       "WG",
						AgreementDate:     "06/20/2023",
						AssetCode:         "OPPO.HP/SMARTPHONES.PHABLETF56GB",
						Tenor:             11,
						InstallmentAmount: 161000,
						ContractStatus:    "LIV",
						CurrentCondition:  "Current",
					},
				},
			},
			resultMetrics: response.Metrics{},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockUsecase := new(mocks.Usecase)
			mockMultiUsecase := new(mocks.MultiUsecase)

			mockRepository.On("ScanTrxMaster", tc.reqMetrics.Transaction.ProspectID).Return(tc.trxMaster, tc.errScanTrxMaster)
			mockRepository.On("ScanTrxPrescreening", tc.reqMetrics.Transaction.ProspectID).Return(tc.countTrx, tc.errScanTrxPrescreening)
			mockRepository.On("GetFilteringResult", tc.reqMetrics.Transaction.ProspectID).Return(tc.filtering, tc.errGetFilteringResult)
			mockRepository.On("GetFilteringForJourney", tc.reqMetrics.Transaction.ProspectID).Return(tc.filtering, tc.errGetFilteringForJourney)
			mockRepository.On("GetElaborateLtv", tc.reqMetrics.Transaction.ProspectID).Return(entity.MappingElaborateLTV{}, tc.errGetElaborateLtv)
			mockRepository.On("MasterMappingCluster", mock.Anything).Return(tc.mappingCluster, tc.errmappingCluster)
			mockRepository.On("MasterMappingMaxDSR", mock.Anything).Return(tc.mappingMaxDSR, tc.errmappingMaxDSR)
			mockUsecase.On("SaveTransaction", tc.countTrx, tc.reqMetrics, tc.trxPrescreening, tc.trxFMF, tc.details, mock.Anything).Return(tc.resultMetrics, tc.errSaveTransaction)
			mockUsecase.On("Prescreening", ctx, tc.reqMetrics, tc.filtering, "token").Return(tc.trxPrescreening, tc.trxFMF, tc.trxPrescreeningDetail, tc.errPrescreening)
			mockUsecase.On("RejectTenor36", mock.Anything).Return(tc.trxTenor, tc.errRejectTenor36)
			mockRepository.On("GetConfig", "dupcheck", "KMB-OFF", "dupcheck_kmb_config").Return(tc.config, tc.errGetConfig)
			mockMultiUsecase.On("Dupcheck", ctx, mock.Anything, true, "token", tc.configValue).Return(tc.dupcheckData, tc.customerStatus, tc.metricsDupcheck, tc.trxFMFDupcheck, tc.trxDetailDupcheck, tc.errDupcheck)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			if tc.dupcheckData.CustomerID != nil {
				httpmock.RegisterResponder(constant.METHOD_GET, os.Getenv("INTERNAL_RECORD_URL"), httpmock.NewStringResponder(tc.codeInternalRecord, tc.bodyInternalRecord))
				resp, _ := rst.R().Get(os.Getenv("INTERNAL_RECORD_URL"))

				mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("INTERNAL_RECORD_URL")+tc.dupcheckData.CustomerID.(string), mock.Anything, map[string]string{}, constant.METHOD_GET, true, 3, 60, tc.reqMetrics.Transaction.ProspectID, "token").Return(resp, tc.errInternalRecord).Once()
			}

			metrics := NewMetrics(mockRepository, mockHttpClient, mockUsecase, mockMultiUsecase)
			result, err := metrics.MetricsLos(ctx, tc.reqMetrics, "token")
			require.Equal(t, tc.resultMetrics, result)
			require.Equal(t, tc.err, err)
		})
	}
}

func TestSaveTransaction(t *testing.T) {
	testcases := []struct {
		name            string
		countTrx        int
		data            request.Metrics
		trxPrescreening entity.TrxPrescreening
		trxFMF          response.TrxFMF
		details         []entity.TrxDetail
		reason          string
		resp            response.Metrics
		err, errsave    error
	}{
		{
			name: "test save transaction apr",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_APR,
				},
			},
			resp: response.Metrics{
				Decision: constant.DECISION_APPROVE,
			},
		},
		{
			name: "test save transaction rej",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_REJECT,
				},
			},
			resp: response.Metrics{
				Decision: constant.JSON_DECISION_REJECT,
			},
		},
		{
			name: "test save transaction pas",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_PASS,
				},
			},
			resp: response.Metrics{
				Decision: constant.JSON_DECISION_PASS,
			},
		},
		{
			name: "test save transaction cpr",
			details: []entity.TrxDetail{
				{
					Decision: constant.DB_DECISION_CREDIT_PROCESS,
				},
			},
			resp: response.Metrics{
				Decision: constant.JSON_DECISION_CREDIT_PROCESS,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("SaveTransaction", tc.countTrx, tc.data, tc.trxPrescreening, tc.trxFMF, tc.details, tc.reason).Return(tc.errsave)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			data, err := usecase.SaveTransaction(tc.countTrx, tc.data, tc.trxPrescreening, tc.trxFMF, tc.details, tc.reason)
			require.Equal(t, tc.resp, data)
			require.Equal(t, tc.err, err)
		})
	}
}
