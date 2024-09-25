package usecase

import (
	"context"
	"errors"
	"fmt"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCheckPMK(t *testing.T) {
	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	currentYearInt := int(time.Now().Year())
	currentMonthInt := int(time.Now().Month())

	testcases := []struct {
		name                      string
		branchID, customerKMB     string
		income                    float64
		homeStatus, professionID  string
		birthDate                 string
		tenor                     int
		maritalStatus             string
		empYear, empMonth         int
		stayYear, stayMonth       int
		resGetConfig              entity.AppConfig
		errGetConfig              error
		resGetMinimalIncomePMK    entity.MappingIncomePMK
		errGetMinimalIncomePMK    error
		result                    response.UsecaseApi
		err                       error
		expectGetMinimalIncomePMK bool
	}{
		{
			name:          "success",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "SD",
			professionID:  "PRO",
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       2015,
			empMonth:      11,
			stayYear:      2018,
			stayMonth:     6,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PMK_SESUAI,
				Reason:         constant.REASON_PMK_SESUAI,
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject low income",
			branchID:      "426",
			customerKMB:   "12345",
			income:        2000000,
			homeStatus:    "SD",
			professionID:  "PRO",
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       2015,
			empMonth:      1,
			stayYear:      2018,
			stayMonth:     6,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_INCOME,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_INCOME),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject work experience",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "SD",
			professionID:  constant.PROFESSION_ID_PRO,
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       currentYearInt,
			empMonth:      currentMonthInt,
			stayYear:      2018,
			stayMonth:     6,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_WORK_EXPERIENCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_WORK_EXPERIENCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject work experience other",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "SD",
			professionID:  "Other",
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       currentYearInt,
			empMonth:      currentMonthInt,
			stayYear:      2018,
			stayMonth:     6,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_WORK_EXPERIENCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_WORK_EXPERIENCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject home since PE",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "PE",
			professionID:  constant.PROFESSION_ID_PRO,
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       2015,
			empMonth:      1,
			stayYear:      currentYearInt,
			stayMonth:     currentMonthInt,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_HOME_SINCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_HOME_SINCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject home since KS",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "KS",
			professionID:  constant.PROFESSION_ID_PRO,
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       2015,
			empMonth:      1,
			stayYear:      currentYearInt,
			stayMonth:     currentMonthInt,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_HOME_SINCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_HOME_SINCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject home since other",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "other",
			professionID:  constant.PROFESSION_ID_PRO,
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       2015,
			empMonth:      1,
			stayYear:      currentYearInt,
			stayMonth:     currentMonthInt,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_HOME_SINCE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_HOME_SINCE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject min age",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "other",
			professionID:  constant.PROFESSION_ID_PRO,
			birthDate:     strconv.Itoa(currentYearInt) + "-01-01",
			tenor:         36,
			maritalStatus: "S",
			empYear:       2015,
			empMonth:      1,
			stayYear:      2015,
			stayMonth:     1,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_MIN_AGE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_MIN_AGE_THRESHOLD),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "reject max age",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "other",
			professionID:  constant.PROFESSION_ID_PRO,
			birthDate:     "1880-01-01",
			tenor:         36,
			maritalStatus: "S",
			empYear:       2015,
			empMonth:      1,
			stayYear:      2015,
			stayMonth:     1,
			resGetConfig: entity.AppConfig{
				GroupName: constant.GROUP_PMK,
				Lob:       constant.LOB_KMB,
				Key:       constant.KEY_PMK,
				Value:     `{"data":{"min_age_marital_status_m":19,"min_age_marital_status_s":21,"marital_checking":true,"max_age_limit":58,"length_of_business":2,"length_of_work":1,"length_of_stay":{"sd":1,"rd":1,"rk":2},"minimal_income":0,"manufacturing_year":0}}`,
				IsActive:  1,
			},
			resGetMinimalIncomePMK: entity.MappingIncomePMK{
				BranchID:       "426",
				StatusKonsumen: "12345",
				Income:         3000000,
				Lob:            constant.LOB_KMB,
			},
			result: response.UsecaseApi{
				Result:         constant.DECISION_REJECT,
				Code:           constant.CODE_REJECT_MAX_AGE,
				Reason:         fmt.Sprintf(" %s", constant.REASON_REJECT_MAX_AGE),
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			expectGetMinimalIncomePMK: true,
		},
		{
			name:          "error get config",
			branchID:      "426",
			customerKMB:   "12345",
			income:        5000000,
			homeStatus:    "SD",
			professionID:  "PRO",
			birthDate:     "1990-01-01",
			tenor:         36,
			maritalStatus: "M",
			empYear:       2015,
			empMonth:      1,
			stayYear:      2018,
			stayMonth:     6,
			resGetConfig:  entity.AppConfig{},
			errGetConfig:  errors.New("something wrong"),
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PMK_SESUAI,
				Reason:         constant.REASON_PMK_SESUAI,
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			err:                       errors.New("something wrong"),
			expectGetMinimalIncomePMK: false,
		},
		{
			name:                   "error get minimal income PMK",
			branchID:               "426",
			customerKMB:            "12345",
			income:                 5000000,
			homeStatus:             "SD",
			professionID:           "PRO",
			birthDate:              "1990-01-01",
			tenor:                  36,
			maritalStatus:          "M",
			empYear:                2015,
			empMonth:               1,
			stayYear:               2018,
			stayMonth:              6,
			resGetMinimalIncomePMK: entity.MappingIncomePMK{},
			errGetMinimalIncomePMK: errors.New("something wrong"),
			result: response.UsecaseApi{
				Result:         constant.DECISION_PASS,
				Code:           constant.CODE_PMK_SESUAI,
				Reason:         constant.REASON_PMK_SESUAI,
				SourceDecision: constant.SOURCE_DECISION_PMK,
			},
			err:                       errors.New("something wrong"),
			expectGetMinimalIncomePMK: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetConfig", constant.GROUP_PMK, constant.LOB_KMB, constant.KEY_PMK).Return(tc.resGetConfig, tc.errGetConfig).Once()

			if tc.expectGetMinimalIncomePMK {
				mockRepository.On("GetMinimalIncomePMK", tc.branchID, tc.customerKMB).Return(tc.resGetMinimalIncomePMK, tc.errGetMinimalIncomePMK).Once()
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			result, err := usecase.CheckPMK(tc.branchID, tc.customerKMB, tc.income, tc.homeStatus, tc.professionID, tc.birthDate, tc.tenor, tc.maritalStatus, tc.empYear, tc.empMonth, tc.stayYear, tc.stayMonth)

			require.Equal(t, tc.result, result)
			if tc.err != nil {
				require.Error(t, err)
				require.IsType(t, tc.err, err)
			} else {
				require.NoError(t, err)
			}

			mockRepository.AssertExpectations(t)
			mockHttpClient.AssertExpectations(t)
		})
	}
}
