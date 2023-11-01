package usecase

// import (
// 	"context"
// 	"los-kmb-api/domain/kmb/interfaces/mocks"
// 	"los-kmb-api/models/entity"
// 	"los-kmb-api/models/request"
// 	"los-kmb-api/models/response"
// 	"los-kmb-api/shared/constant"
// 	"los-kmb-api/shared/httpclient"
// 	"os"
// 	"testing"

// 	"github.com/go-resty/resty/v2"
// 	"github.com/jarcoal/httpmock"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/require"
// )

// func TestScorepro(t *testing.T) {

// 	os.Setenv("SCOREPRO_DEFAULT_KEY", "first_residence_zipcode_2w_others")
// 	os.Setenv("SCOREPRO_DEFAULT_SCORE_GENERATOR_ID", "37fe1525-1be1-48d1-aab5-6adf05305a0a")
// 	os.Setenv("SCOREPRO_REQUESTID", "107b280e-12e4-4c92-80e8-38a7422cb9bc")
// 	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
// 	os.Setenv("PEFINDO_IDX_URL", "http://10.9.100.231/los-int-pbk/usecase/new-kmb/idx")
// 	os.Setenv("KEY_SCOREPRO_IDX_2W_JABOJABAR", "first_residence_zipcode_2w_jabo")
// 	os.Setenv("KEY_SCOREPRO_IDX_2W_OTHERS", "first_residence_zipcode_2w_others")
// 	os.Setenv("KEY_SCOREPRO_IDX_2W_AORO", "first_residence_zipcode_2w_aoro")
// 	os.Setenv("SCOREPRO_IDX_URL", "http://10.9.100.122:9105/api/v1/scorepro/kmb/idx")
// 	os.Setenv("SCOREPRO_SEGMEN_ASS_SCORE", "2")
// 	os.Setenv("NAMA_SAMA", "K,P")

// 	testcases := []struct {
// 		name               string
// 		req                request.Metrics
// 		pefindoScore       string
// 		customerSegment    string
// 		spDupcheck         response.SpDupcheckMap
// 		accessToken        string
// 		scoreGenerator     entity.ScoreGenerator
// 		errscoreGenerator  error
// 		trxDetailBiro      []entity.TrxDetailBiro
// 		errtrxDetailBiro   error
// 		codePefindoIDX     int
// 		bodyPefindoIDX     string
// 		errRespPefindoIDX  error
// 		codeScoreproIDX    int
// 		bodyScoreproIDX    string
// 		errRespScoreproIDX error
// 		responseScs        response.IntegratorScorePro
// 		data               response.ScorePro
// 		respPefindoIDX     response.PefindoIDX
// 		err                error
// 		result             response.ScorePro
// 		errResult          error
// 	}{
// 		{
// 			name: "scorepro ",
// 			req: request.Metrics{
// 				Address: []request.Address{
// 					{
// 						Type:    constant.ADDRESS_TYPE_RESIDENCE,
// 						ZipCode: "12908",
// 					},
// 				},
// 				Item: request.Item{BPKBName: "K"},
// 			},
// 			spDupcheck: response.SpDupcheckMap{
// 				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
// 			},
// 			codePefindoIDX: 200,
// 			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
// 			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
// 			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
// 			scoreGenerator: entity.ScoreGenerator{
// 				Key: "first_residence_zipcode_2w_jabo",
// 			},
// 			trxDetailBiro: []entity.TrxDetailBiro{
// 				{
// 					Score:   "LOW RISK",
// 					Subject: constant.CUSTOMER,
// 					BiroID:  "123456",
// 				},
// 				{
// 					Score:   "LOW RISK",
// 					Subject: constant.SPOUSE,
// 					BiroID:  "234567",
// 				},
// 			},
// 			codeScoreproIDX: 200,
// 			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
// 			result: response.ScorePro{
// 				Result: constant.DECISION_PASS,
// 				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
// 				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
// 				Source: constant.SOURCE_DECISION_SCOREPRO,
// 				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
// 			},
// 		},
// 		{
// 			name: "scorepro ",
// 			req: request.Metrics{
// 				Address: []request.Address{
// 					{
// 						Type:    constant.ADDRESS_TYPE_RESIDENCE,
// 						ZipCode: "12908",
// 					},
// 				},
// 				Item: request.Item{BPKBName: "K"},
// 			},
// 			spDupcheck: response.SpDupcheckMap{
// 				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
// 				CustomerID:     "123456",
// 			},
// 			codePefindoIDX: 200,
// 			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
// 			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
// 			"server_time":"2023-11-01 13:14:35"}`,
// 			scoreGenerator: entity.ScoreGenerator{
// 				Key: "first_residence_zipcode_2w_aoro",
// 			},
// 			trxDetailBiro: []entity.TrxDetailBiro{
// 				{
// 					Score:   "LOW RISK",
// 					Subject: constant.CUSTOMER,
// 					BiroID:  "123456",
// 				},
// 				{
// 					Score:   "LOW RISK",
// 					Subject: constant.SPOUSE,
// 					BiroID:  "234567",
// 				},
// 			},
// 			codeScoreproIDX: 200,
// 			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
// 			result: response.ScorePro{
// 				Result: constant.DECISION_PASS,
// 				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
// 				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
// 				Source: constant.SOURCE_DECISION_SCOREPRO,
// 				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
// 			},
// 		},
// 	}

// 	for _, tc := range testcases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctx := context.Background()

// 			mockRepository := new(mocks.Repository)
// 			mockHttpClient := new(httpclient.MockHttpClient)

// 			mockRepository.On("GetScoreGenerator", mock.Anything).Return(tc.scoreGenerator, tc.errscoreGenerator)
// 			mockRepository.On("GetScoreGeneratorROAO").Return(tc.scoreGenerator, tc.errscoreGenerator)
// 			mockRepository.On("GetTrxDetailBIro", tc.req.Transaction.ProspectID).Return(tc.trxDetailBiro, tc.errtrxDetailBiro)
// 			mockRepository.On("GetActiveLoanTypeLast6M", mock.Anything).Return(entity.GetActiveLoanTypeLast6M{}, mock.Anything)
// 			mockRepository.On("GetActiveLoanTypeLast24M", mock.Anything).Return(entity.GetActiveLoanTypeLast24M{}, mock.Anything)
// 			mockRepository.On("GetMoblast", mock.Anything).Return(entity.GetMoblast{}, mock.Anything)

// 			rst := resty.New()
// 			httpmock.ActivateNonDefault(rst.GetClient())
// 			defer httpmock.DeactivateAndReset()

// 			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("PEFINDO_IDX_URL"), httpmock.NewStringResponder(tc.codePefindoIDX, tc.bodyPefindoIDX))
// 			resp, _ := rst.R().Post(os.Getenv("PEFINDO_IDX_URL"))

// 			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("PEFINDO_IDX_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.req.Transaction.ProspectID, tc.accessToken).Return(resp, tc.errRespPefindoIDX).Once()

// 			rst = resty.New()
// 			httpmock.ActivateNonDefault(rst.GetClient())
// 			defer httpmock.DeactivateAndReset()

// 			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("SCOREPRO_IDX_URL"), httpmock.NewStringResponder(tc.codeScoreproIDX, tc.bodyScoreproIDX))
// 			resp, _ = rst.R().Post(os.Getenv("SCOREPRO_IDX_URL"))

// 			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("SCOREPRO_IDX_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.req.Transaction.ProspectID, tc.accessToken).Return(resp, tc.errRespScoreproIDX).Once()

// 			usecase := NewUsecase(mockRepository, mockHttpClient)

// 			_, data, _, err := usecase.Scorepro(ctx, tc.req, tc.pefindoScore, tc.customerSegment, tc.spDupcheck, tc.accessToken)
// 			require.Equal(t, tc.result, data)
// 			require.Equal(t, tc.errResult, err)
// 		})
// 	}
// }
