package usecase

import (
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPmk(t *testing.T) {
	today := time.Now()
	responseAppConfig := entity.AppConfig{
		Value: `{"data":{
				"min_age_marital_status_m":19,
				"min_age_marital_status_s":21,
				"marital_checking":true,
				"max_age_limit":60,
				"length_of_business":2,
				"length_of_work":1,
				"length_of_stay":{
					"sd":1,"rd":1,"rk":2},
				"minimal_income":1800000,
				"manufacturing_year":13}}`,
	}

	testcases := []struct {
		name, branchID, customerKMB                                                 string
		income                                                                      float64
		homeStatus, professionID, empYear, empMonth, stayYear, stayMonth, birthDate string
		tenor                                                                       int
		maritalStatus                                                               string
		data                                                                        response.UsecaseApi
		err, responseAppConfig, errminimalIncome                                    error
		minimalIncome                                                               entity.MappingIncomePMK
	}{
		{
			name:          "test pmk err responseAppConfig",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        10000000,
			homeStatus:    "SD",
			professionID:  "KRSW",
			empYear:       "2021",
			empMonth:      "12",
			stayYear:      "2005",
			stayMonth:     "09",
			birthDate:     "1998-09-09",
			tenor:         12,
			maritalStatus: "S",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			responseAppConfig: errors.New(constant.ERROR_UPSTREAM + " - Get PMK Config Error"),
			err:               errors.New(constant.ERROR_UPSTREAM + " - Get PMK Config Error"),
			data: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PMK_SESUAI,
				Reason:         constant.REASON_PMK_SESUAI,
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk err errminimalIncome",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        10000000,
			homeStatus:    "SD",
			professionID:  "KRSW",
			empYear:       "2021",
			empMonth:      "12",
			stayYear:      "2005",
			stayMonth:     "09",
			birthDate:     "1998-09-09",
			tenor:         12,
			maritalStatus: "S",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			errminimalIncome: errors.New(constant.ERROR_UPSTREAM + " - Get Minimal Income PMK Error"),
			err:              errors.New(constant.ERROR_UPSTREAM + " - Get Minimal Income PMK Error"),
			data: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PMK_SESUAI,
				Reason:         constant.REASON_PMK_SESUAI,
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk pass",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        10000000,
			homeStatus:    "SD",
			professionID:  "KRSW",
			empYear:       "2021",
			empMonth:      "12",
			stayYear:      "2005",
			stayMonth:     "09",
			birthDate:     "1998-09-09",
			tenor:         12,
			maritalStatus: "S",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			data: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PMK_SESUAI,
				Reason:         constant.REASON_PMK_SESUAI,
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk reject minimalIncome",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        700000,
			homeStatus:    "SD",
			professionID:  "KRSW",
			empYear:       "2021",
			empMonth:      "12",
			stayYear:      "2005",
			stayMonth:     "09",
			birthDate:     "1998-09-09",
			tenor:         12,
			maritalStatus: "S",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			data: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_INCOME,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_INCOME),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk reject kerja",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        7000000,
			homeStatus:    "SD",
			professionID:  "KRSW",
			empYear:       time.Now().Format("2006"),
			empMonth:      time.Now().Format("01"),
			stayYear:      "2005",
			stayMonth:     "09",
			birthDate:     "1998-09-09",
			tenor:         12,
			maritalStatus: "S",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			data: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_WORK_EXPERIENCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_WORK_EXPERIENCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk reject kerja pro",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        7000000,
			homeStatus:    "SD",
			professionID:  "PRO",
			empYear:       "2022",
			empMonth:      "12",
			stayYear:      "2005",
			stayMonth:     "09",
			birthDate:     "1998-09-09",
			tenor:         12,
			maritalStatus: "S",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			data: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_WORK_EXPERIENCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_WORK_EXPERIENCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk reject rumah dinas",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        7000000,
			homeStatus:    "PE",
			professionID:  "PRO",
			empYear:       "2020",
			empMonth:      "12",
			stayYear:      strconv.Itoa(today.Year()),
			stayMonth:     "01",
			birthDate:     "1998-09-09",
			tenor:         12,
			maritalStatus: "S",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			data: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_HOME_SINCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_HOME_SINCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk reject usia married",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        7000000,
			homeStatus:    "KS",
			professionID:  "PRO",
			empYear:       "2020",
			empMonth:      "12",
			stayYear:      "2020",
			stayMonth:     "09",
			birthDate:     "2005-09-09",
			tenor:         12,
			maritalStatus: "M",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			data: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_MIN_AGE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_MIN_AGE_THRESHOLD),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
		{
			name:          "test pmk reject usia > threshold",
			branchID:      "426",
			customerKMB:   "NEW",
			income:        7000000,
			homeStatus:    "RANDOM",
			professionID:  "PRO",
			empYear:       "2020",
			empMonth:      "12",
			stayYear:      "2020",
			stayMonth:     "09",
			birthDate:     "1945-09-09",
			tenor:         12,
			maritalStatus: "M",
			minimalIncome: entity.MappingIncomePMK{
				Income: 5000000,
			},
			data: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_MAX_AGE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_MAX_AGE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetConfig", "pmk_config", "KMB-OFF", "pmk_kmb_off").Return(responseAppConfig, tc.responseAppConfig)
			mockRepository.On("GetMinimalIncomePMK", tc.branchID, tc.customerKMB).Return(tc.minimalIncome, tc.errminimalIncome)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			pmk, err := usecase.PMK(tc.branchID, tc.customerKMB, tc.income, tc.homeStatus, tc.professionID, tc.empYear, tc.empMonth, tc.stayYear, tc.stayMonth, tc.birthDate, tc.tenor, tc.maritalStatus)

			require.Equal(t, tc.data, pmk)
			require.Equal(t, tc.err, err)
		})
	}

}
