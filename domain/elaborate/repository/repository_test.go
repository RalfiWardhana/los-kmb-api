package repository

import (
	"database/sql/driver"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
)

func structToSlice(data interface{}) []driver.Value {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Struct {
		return nil
	}

	values := make([]driver.Value, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		if i == (v.NumField() - 1) {
			values[i] = sqlmock.AnyArg()
		} else {
			values[i] = v.Field(i).Interface()

		}
	}

	return values
}

func TestSaveDataElaborate(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB)

	saveData := entity.ApiElaborateKmb{
		ProspectID: "TEST0001",
		RequestID:  "a6c09ce7-9d6d-4962-b32d-61a9e64d9be7",
		Request:    `{"client_key":"$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6","data":{"ProspectID":"EFM0TSTRT2023112900","BranchID":"426","BPKBName":"K","CustomerStatus":"NEW","CategoryCustomer":"","ResultPefindo":"PASS","TotalBakiDebet":0,"Tenor":12,"ManufacturingYear":"2020","OTR":10000000,"NTF":36000000}}`,
		Code:       constant.CODE_REJECT_NTF_ELABORATE,
		Decision:   constant.DECISION_REJECT,
		Reason:     constant.REASON_REJECT_NTF_ELABORATE,
	}
	elaborateScheme := structToSlice(saveData)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "api_elaborate_scheme" (.*)`).
		WithArgs(elaborateScheme[0], elaborateScheme[1], elaborateScheme[2], elaborateScheme[3], elaborateScheme[4], elaborateScheme[5], sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := newDB.SaveDataElaborate(saveData)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestUpdateDataElaborate(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB)

	updateData := entity.ApiElaborateKmbUpdate{
		ProspectID: "TEST0001",
		RequestID:  `{"client_key":"$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6","data":{"ProspectID":"EFM11621202310300008","BranchID":"407","BPKBName":"P","CustomerStatus":"NEW","CategoryCustomer":"","ResultPefindo":"PASS","TotalBakiDebet":0,"Tenor":55,"ManufacturingYear":"2006","OTR":75900000,"NTF":6200000}}`,
		Response:   `{"code":9601,"decision":"PASS","reason":"PASS - Elaborated Scheme"}`,
		Code:       constant.CODE_PASS_ELABORATE,
		Decision:   constant.DECISION_PASS,
		Reason:     constant.REASON_PASS_ELABORATE,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "api_elaborate_scheme" SET "Code" = ?, "Decision" = ?, "DtmResponse" = ?, "ProspectID" = ?, "Reason" = ?, "RequestID" = ?, "Response" = ?, "Timestamp" = ? WHERE (RequestID = ?)`)).
		WithArgs(updateData.Code, updateData.Decision, sqlmock.AnyArg(), updateData.ProspectID, updateData.Reason, updateData.RequestID, updateData.Response, sqlmock.AnyArg(), updateData.RequestID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := newDB.UpdateDataElaborate(updateData)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetClusterBranchElaborate(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB)

	query := "SELECT cluster FROM kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?"

	branchID := "426"
	statusKonsumen := constant.STATUS_KONSUMEN_NEW
	bpkbNameType := 0

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(branchID, statusKonsumen, bpkbNameType).
		WillReturnRows(sqlmock.NewRows([]string{"branch_id", "customer_status", "bpkb_name_type", "cluster"}).
			AddRow("426", "NEW", 0, "Cluster F"))
	mock.ExpectCommit()

	_, err := newDB.GetClusterBranchElaborate(branchID, statusKonsumen, bpkbNameType)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetClusterBranchElaborateErrorNotFound(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB)

	query := "SELECT cluster FROM kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?"

	branchID := "426"
	statusKonsumen := constant.STATUS_KONSUMEN_NEW
	bpkbNameType := 0

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(branchID, statusKonsumen, bpkbNameType).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectCommit()

	_, err := newDB.GetClusterBranchElaborate(branchID, statusKonsumen, bpkbNameType)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetFilteringResult(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB)

	query := `SELECT
				PefindoID,
				PefindoIDSpouse,
				CASE
				WHEN PefindoScore IS NULL then 'UNSCORE'
				ELSE PefindoScore
				END AS PefindoScore,
				CAST(
				JSON_EXTRACT(ResultPefindo, '$.result.max_overdue') AS SIGNED
				) AS MaxOverdue,
				JSON_EXTRACT(ResultPefindo, '$.result.max_overdue') = CAST('null' AS JSON) AS IsNullMaxOverdue,
				CAST(
				JSON_EXTRACT(
					ResultPefindo,
					'$.result.max_overdue_last12months'
				) AS SIGNED
				) AS MaxOverdueLast12Months,
				JSON_EXTRACT(
				ResultPefindo,
				'$.result.max_overdue_last12months'
				) = CAST('null' AS JSON) AS IsNullMaxOverdueLast12Months
			FROM
				api_dupcheck_kmb
			WHERE
				ProspectID = ?
			ORDER BY
				Timestamp DESC
			LIMIT
				1`

	prospectID := "TEST0001"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"PefindoID", "PefindoIDSpouse", "PefindoScore"}).
			AddRow("2467775893", "7838158537", "AVERAGE RISK"))
	mock.ExpectCommit()

	_, err := newDB.GetFilteringResult(prospectID)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetResultElaborate(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB)

	testCases := []struct {
		name           string
		branch_id      string
		cust_status    string
		bpkb           int
		result_pefindo string
		tenor          int
		age_vehicle    string
		ltv            float64
		baki_debet     float64
		queryAdd       string
	}{
		{
			name:           "TEST_GetResultElaborate_PefindoPass_Tenor>=36_BPKB=0",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_PASS,
			tenor:          36,
			age_vehicle:    "<=12",
			ltv:            0.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.bpkb_name_type = %d AND mes.tenor_start >= 36 AND mes.tenor_end = 0", 0),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoPass_Tenor>=36_BPKB=1",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           1,
			result_pefindo: constant.DECISION_PASS,
			tenor:          36,
			age_vehicle:    "<=12",
			ltv:            0.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.bpkb_name_type = %d AND mes.tenor_start >= 36 AND mes.tenor_end = 0", 1) + fmt.Sprintf(" AND mes.age_vehicle = '%s'", "<=12"),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoPass_Tenor<36_BPKB=0_AgeVehicle<=12_LTV=0",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_PASS,
			tenor:          24,
			age_vehicle:    "<=12",
			ltv:            0.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", 24, 24) + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", int(0.00)),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoPass_Tenor<36_BPKB=0_AgeVehicle<=12_LTV<=1000",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_PASS,
			tenor:          24,
			age_vehicle:    "<=12",
			ltv:            1000.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", 24, 24) + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", int(1000.00), int(1000.00)),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoPass_Tenor<36_BPKB=0_AgeVehicle>12_LTV=0",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_PASS,
			tenor:          24,
			age_vehicle:    ">12",
			ltv:            0.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", 24, 24) + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", int(0.00)),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoPass_Tenor<36_BPKB=0_AgeVehicle>12_LTV<=1000",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_PASS,
			tenor:          24,
			age_vehicle:    ">12",
			ltv:            1000.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", 24, 24) + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", int(1000.00), int(1000.00)),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoNoHit_Tenor>=24_LTV=0",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_PBK_NO_HIT,
			tenor:          24,
			age_vehicle:    ">12",
			ltv:            0.00,
			baki_debet:     10000000.00,
			queryAdd:       "AND mes.tenor_start >= 24 AND mes.tenor_end = 0" + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", int(0.00)),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoNoHit_Tenor<24_LTV<=1000",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_PBK_NO_HIT,
			tenor:          12,
			age_vehicle:    ">12",
			ltv:            1000.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", 12, 12) + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", int(1000.00), int(1000.00)),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoReject_Tenor>=24",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_REJECT,
			tenor:          24,
			age_vehicle:    ">12",
			ltv:            1000.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.total_baki_debet_start <= %d AND mes.total_baki_debet_end >= %d", int(10000000.00), int(10000000.00)) + " AND mes.tenor_start >= '24' AND mes.tenor_end = 0",
		},
		{
			name:           "TEST_GetResultElaborate_PefindoReject_Tenor<24_LTV=0",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_REJECT,
			tenor:          12,
			age_vehicle:    ">12",
			ltv:            0.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.total_baki_debet_start <= %d AND mes.total_baki_debet_end >= %d", int(10000000.00), int(10000000.00)) + fmt.Sprintf(" AND mes.tenor_start <= %d AND mes.tenor_end >= %d", 12, 12) + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= '1000'", int(0.00)),
		},
		{
			name:           "TEST_GetResultElaborate_PefindoReject_Tenor<24_LTV<=1000",
			branch_id:      "426",
			cust_status:    constant.CUSTOMER_NEW,
			bpkb:           0,
			result_pefindo: constant.DECISION_REJECT,
			tenor:          12,
			age_vehicle:    ">12",
			ltv:            1000.00,
			baki_debet:     10000000.00,
			queryAdd:       fmt.Sprintf("AND mes.total_baki_debet_start <= %d AND mes.total_baki_debet_end >= %d", int(10000000.00), int(10000000.00)) + fmt.Sprintf(" AND mes.tenor_start <= %d AND mes.tenor_end >= %d", 12, 12) + fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", int(1000.00), int(1000.00)),
		},
	}

	for _, tc := range testCases {
		testFailed := false
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectBegin()
			mock.ExpectQuery(regexp.QuoteMeta("SELECT mcb.cluster, mes.decision, mes.ltv_start FROM kmb_mapping_cluster_branch mcb JOIN kmb_mapping_elaborate_scheme mes ON mcb.cluster = mes.cluster WHERE mcb.branch_id = ? AND mcb.customer_status = ? AND mcb.bpkb_name_type = ? AND mes.result_pefindo = ? "+tc.queryAdd)).
				WithArgs(tc.branch_id, tc.cust_status, tc.bpkb, tc.result_pefindo).
				WillReturnRows(sqlmock.NewRows([]string{"cluster", "decision", "ltv_start"}).
					AddRow("Cluster A", constant.DECISION_PASS, 0))
			mock.ExpectCommit()

			_, err := newDB.GetResultElaborate(tc.branch_id, tc.cust_status, tc.bpkb, tc.result_pefindo, tc.tenor, tc.age_vehicle, tc.ltv, tc.baki_debet)
			if err != nil {
				t.Errorf("error '%s' was not expected, but got: ", err)
				testFailed = true
			}
		})

		if testFailed {
			t.Fail()
			break
		}
	}
}
