package usecase

import (
	"context"
	"errors"
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

	testcases := []struct {
		name               string
		req                request.Metrics
		filtering          entity.FilteringKMB
		pefindoScore       string
		customerSegment    string
		spDupcheck         response.SpDupcheckMap
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
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro jabo nama sama no pefindo reject",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541d4b0a7ea1","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 11:31:44","oldestmob_pl":-999,"final_nom60_12mth":0,"tot_bakidebet_banks_active":-999,"tot_bakidebet_31_60dpd":0,
			"worst_24mth":0,"max_limit_oth":-999,"pefindo_add_info":null},"server_time":"2023-11-01 11:31:44"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_jabo",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFM0TSTRT87183109505","score":200,"result":"REJECT","score_result":"LOW","status":"ASS-LOW","phone_number":"0817344026205","segmen":"1","is_tsi":false,"score_bin":0},"errors":null,"server_time":"2023-11-07T17:48:36+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFM0TSTRT87183109505","score":200,"result":"REJECT","score_band":"","score_result":"LOW","status":"ASS-LOW","segmen":"1","is_tsi":false,"score_bin":0}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda low",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk reject",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk pass",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk segmen 1 score low",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"1","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"PASS","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"1","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro jabo bpkb nama beda very hisk risk segmen 1 score high",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_NEW,
				CustomerID:     "",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"13","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"13","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro others",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro ro prime",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "CR perbaikan flow RO PrimePriority PASS",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				CustomerID:                       "123456",
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
			},
			filtering: entity.FilteringKMB{
				RrdDate:   sevenMonthsAgo,
				CreatedAt: currentTime,
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			config: entity.AppConfig{
				Key:   "expired_contract_check",
				Value: `{"data":{"expired_contract_check_enabled":true,"expired_contract_max_months":6}}`,
			},
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "CR perbaikan flow RO PrimePriority RrdDate NULL",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen:                   constant.STATUS_KONSUMEN_RO,
				CustomerID:                       "123456",
				InstallmentTopup:                 0,
				MaxOverdueDaysforActiveAgreement: 31,
			},
			filtering: entity.FilteringKMB{
				RrdDate:   nil,
				CreatedAt: currentTime,
			},
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
			codeScoreproIDX: 500,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			errResult:       errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty"),
		},
		{
			name: "scorepro ao prime",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen:          constant.STATUS_KONSUMEN_AO,
				CustomerID:              "123456",
				InstallmentTopup:        0,
				NumberOfPaidInstallment: 6,
			},
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
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
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
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
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
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"12","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"12","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda low",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
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
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda low no pefindo reject",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_result":"LOW","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":230,"result":"REJECT","score_band":"","score_result":"LOW","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama beda low no pefindo pass",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo pass",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"3","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":800,"result":"PASS","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"3","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo pass bukan ASS-SCORE",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_result":"HIGH","status":"ASSCB-HIGH","phone_number":"085716728933","segmen":"2","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_band":"","score_result":"HIGH","status":"ASSCB-HIGH","segmen":"2","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo reject ASS-SCORE",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_result":"HIGH","status":"ASS-HIGH","phone_number":"085716728933","segmen":"2","is_tsi":false,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_REJECT,
				Code:   constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":400,"result":"REJECT","score_band":"","score_result":"HIGH","status":"ASS-HIGH","segmen":"2","is_tsi":false,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo tsi reject",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_result":"MEDIUM","status":"ASSTSH-S04","phone_number":"085716728933","segmen":"4","is_tsi":true,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result:    constant.DECISION_REJECT,
				Code:      constant.CODE_SCOREPRO_LTMIN_THRESHOLD,
				Reason:    constant.REASON_SCOREPRO_LTMIN_THRESHOLD,
				Source:    constant.SOURCE_DECISION_SCOREPRO,
				Info:      `{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_band":"","score_result":"MEDIUM","status":"ASSTSH-S04","segmen":"4","is_tsi":true,"score_bin":""}`,
				IsDeviasi: true,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo tsi pass",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "K"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_result":"MEDIUM","status":"ASSTSH-S04","phone_number":"085716728933","segmen":"4","is_tsi":true,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":"400-599","result":"REJECT","score_band":"","score_result":"MEDIUM","status":"ASSTSH-S04","segmen":"4","is_tsi":true,"score_bin":""}`,
			},
		},
		{
			name: "scorepro aoro bpkb nama sama low no pefindo tsi pass",
			req: request.Metrics{
				Address: []request.Address{
					{
						Type:    constant.ADDRESS_TYPE_RESIDENCE,
						ZipCode: "12908",
					},
				},
				Item: request.Item{BPKBName: "KK"},
			},
			spDupcheck: response.SpDupcheckMap{
				StatusKonsumen: constant.STATUS_KONSUMEN_RO,
				CustomerID:     "123456",
			},
			codePefindoIDX: 200,
			bodyPefindoIDX: `{"message":"success","errors":null,"data":{"id":"pbk_idx6541eccbb2b8a","prospect_id":"EFM01454202307020007",
			"created_at":"2023-11-01 13:14:35","nom03_12mth_all":0,"worst_24mth_auto":-999,"worst_24mth":0,"pefindo_add_info":null},
			"server_time":"2023-11-01 13:14:35"}`,
			scoreGenerator: entity.ScoreGenerator{
				Key: "first_residence_zipcode_2w_aoro",
			},
			codeScoreproIDX: 200,
			bodyScoreproIDX: `{"messages":"OK","data":{"prospect_id":"EFMTESTAKKK0161109","score":"800-899","result":"PASS","score_result":"MEDIUM","status":"ASSTSH-S04","phone_number":"085716728933","segmen":"4","is_tsi":true,"score_band":"","score_bin":""},"errors":null,"server_time":"2023-10-30T14:26:17+07:00"}`,
			result: response.ScorePro{
				Result: constant.DECISION_PASS,
				Code:   constant.CODE_SCOREPRO_GTEMIN_THRESHOLD,
				Reason: constant.REASON_SCOREPRO_GTEMIN_THRESHOLD,
				Source: constant.SOURCE_DECISION_SCOREPRO,
				Info:   `{"prospect_id":"EFMTESTAKKK0161109","score":"800-899","result":"PASS","score_band":"","score_result":"MEDIUM","status":"ASSTSH-S04","segmen":"4","is_tsi":true,"score_bin":""}`,
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
			mockRepository.On("GetTrxDetailBIro", tc.req.Transaction.ProspectID).Return(tc.trxDetailBiro, tc.errtrxDetailBiro)
			mockRepository.On("GetActiveLoanTypeLast6M", tc.spDupcheck.CustomerID.(string)).Return(entity.GetActiveLoanTypeLast6M{}, nil)
			mockRepository.On("GetActiveLoanTypeLast24M", tc.spDupcheck.CustomerID.(string)).Return(entity.GetActiveLoanTypeLast24M{}, nil)
			mockRepository.On("GetMoblast", tc.spDupcheck.CustomerID.(string)).Return(entity.GetMoblast{}, nil)
			mockRepository.On("GetConfig", "expired_contract", "KMB-OFF", "expired_contract_check").Return(tc.config, tc.errGetConfig)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("PEFINDO_IDX_URL"), httpmock.NewStringResponder(tc.codePefindoIDX, tc.bodyPefindoIDX))
			resp, _ := rst.R().Post(os.Getenv("PEFINDO_IDX_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("PEFINDO_IDX_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.req.Transaction.ProspectID, tc.accessToken).Return(resp, tc.errRespPefindoIDX).Once()

			rst = resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("SCOREPRO_IDX_URL"), httpmock.NewStringResponder(tc.codeScoreproIDX, tc.bodyScoreproIDX))
			resp, _ = rst.R().Post(os.Getenv("SCOREPRO_IDX_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.NEW_KMB_LOG, os.Getenv("SCOREPRO_IDX_URL"), mock.Anything, mock.Anything, constant.METHOD_POST, false, 0, 30, tc.req.Transaction.ProspectID, tc.accessToken).Return(resp, tc.errRespScoreproIDX).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient)

			_, data, _, err := usecase.Scorepro(ctx, tc.req, tc.pefindoScore, tc.customerSegment, tc.spDupcheck, tc.accessToken, tc.filtering)
			require.Equal(t, tc.result, data)
			require.Equal(t, tc.errResult, err)
		})
	}
}
