package repository

import (
	"database/sql/driver"
	"encoding/json"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"

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

func TestDummyDataPbk(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM dbo.dummy_pefindo_kmb WITH (nolock) WHERE IDNumber = ?")).
		WithArgs("TST001").
		WillReturnRows(sqlmock.NewRows([]string{"prospect_id"}).
			AddRow("TST001"))
	mock.ExpectCommit()

	_, err := newDB.DummyDataPbk("TST001")
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestSaveFiltering(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB)
	trxFiltering := entity.FilteringKMB{
		ProspectID:        "TST001",
		Decision:          "PASS",
		IDNumber:          "123456",
		LegalName:         "JOHN DOE",
		SurgateMotherName: "IBU",
	}
	filtering := structToSlice(trxFiltering)
	trxDetailBiro := []entity.TrxDetailBiro{
		{
			ProspectID: "TST001",
		},
	}
	detailBiro := structToSlice(trxDetailBiro)
	trxCmoNoFPD := entity.TrxCmoNoFPD{
		ProspectID:              "TST001",
		CMOID:                   "105394",
		CmoCategory:             "OLD",
		CmoJoinDate:             "2020-06-12",
		DefaultCluster:          "Cluster C",
		DefaultClusterStartDate: "2024-05-14",
		DefaultClusterEndDate:   "2024-07-31",
		CreatedAt:               time.Time{},
	}
	historyAssetCheck := []entity.TrxHistoryCheckingAsset{
		{
			ProspectID:              "SAL001",
			NumberOfRetry:           1,
			FinalDecision:           "REJ",
			Reason:                  constant.REASON_REJECT_ASSET_CHECK,
			SourceService:           "SALLY",
			SourceDecisionCreatedAt: time.Time{},
			IsDataChanged:           1,
			IsAssetLocked:           1,
			ChassisNumber:           "NOKA123456",
			EngineNumber:            "NOMESIN123",
			IDNumber:                "3578102808920099",
			LegalName:               "RONALD FAGUNDES",
			BirthDate:               time.Time{},
			SurgateMotherName:       "SULASTRI",
			CreatedAt:               time.Time{},
		},
	}
	historyAsset := structToSlice(historyAssetCheck)
	var date, _ = time.Parse("2006-01-02", "2025-03-07")
	chassisNumber := "NOKA123456"
	engineNumber := "NOMESIN123"
	lockingSystem := entity.TrxLockSystem{
		ProspectID:    "SAL011",
		IDNumber:      "3578102808920088",
		ChassisNumber: &chassisNumber,
		EngineNumber:  &engineNumber,
		Reason:        constant.ASSET_PERNAH_REJECT,
		CreatedAt:     time.Now(),
		UnbanDate:     date,
	}
	lockData := structToSlice(lockingSystem)
	cmoNoFPD := structToSlice(trxCmoNoFPD)
	mock.ExpectBegin()

	// Expect encryption query
	encryptQuery := regexp.QuoteMeta("SELECT SCP.dbo.ENC_B64('SEC', ?) AS encrypt0, SCP.dbo.ENC_B64('SEC', ?) AS encrypt1, SCP.dbo.ENC_B64('SEC', ?) AS encrypt2")
	encryptRows := sqlmock.NewRows([]string{"encrypt0", "encrypt1", "encrypt2"}).
		AddRow("123456", "JOHN DOE", "IBU")

	mock.ExpectQuery(encryptQuery).
		WithArgs(trxFiltering.IDNumber, trxFiltering.LegalName, trxFiltering.SurgateMotherName).
		WillReturnRows(encryptRows)

	mock.ExpectExec(`INSERT INTO "trx_filtering" (.*)`).
		WithArgs(filtering...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "trx_cmo_no_fpd" (.*)`).
		WithArgs(cmoNoFPD...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "trx_detail_biro" (.*)`).
		WithArgs(detailBiro...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Expect encryption query
	encryptQuery2 := regexp.QuoteMeta("SELECT SCP.dbo.ENC_B64('SEC', ?) AS encrypt0, SCP.dbo.ENC_B64('SEC', ?) AS encrypt1, SCP.dbo.ENC_B64('SEC', ?) AS encrypt2")
	encryptRows2 := sqlmock.NewRows([]string{"encrypt0", "encrypt1", "encrypt2"}).
		AddRow("3578102808920099", "RONALD FAGUNDES", "SULASTRI")

	mock.ExpectQuery(encryptQuery2).
		WithArgs(historyAssetCheck[0].IDNumber, historyAssetCheck[0].LegalName, historyAssetCheck[0].SurgateMotherName).
		WillReturnRows(encryptRows2)

	mock.ExpectExec(`INSERT INTO "trx_history_checking_asset" (.*)`).
		WithArgs(historyAsset...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	updateRegex := regexp.QuoteMeta(`UPDATE "trx_history_checking_asset" SET "is_asset_locked" = ?, "updated_at" = ? WHERE ((chassis_number = ? OR engine_number = ?) AND is_asset_locked = 0)`)
	mock.ExpectExec(updateRegex).
		WithArgs(1, sqlmock.AnyArg(), historyAssetCheck[0].ChassisNumber, historyAssetCheck[0].EngineNumber).
		WillReturnResult(sqlmock.NewResult(0, 1)) // No LastInsertId, 1 row affected

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','3578102808920088') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("3578102808920088"))

	mock.ExpectExec(`INSERT INTO "trx_lock_system" (.*)`).
		WithArgs(lockData...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := newDB.SaveFiltering(trxFiltering, trxDetailBiro, trxCmoNoFPD, historyAssetCheck, lockingSystem)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetFilteringByID(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT prospect_id FROM dbo.trx_filtering WITH (nolock) WHERE prospect_id = 'TST001'")).
		WillReturnRows(sqlmock.NewRows([]string{"prospect_id"}).
			AddRow("TST001"))
	_, err := newDB.GetFilteringByID("TST001")
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestMasterMappingCluster(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)
	req := entity.MasterMappingCluster{
		BranchID:       "426",
		CustomerStatus: constant.STATUS_KONSUMEN_NEW,
		BpkbNameType:   1,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM dbo.kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?")).
		WithArgs(req.BranchID, req.CustomerStatus, req.BpkbNameType).
		WillReturnRows(sqlmock.NewRows([]string{"branch_id", "customer_status", "bpkb_name_type", "cluster"}).
			AddRow("426", "NEW", 1, "Cluster C"))
	mock.ExpectCommit()

	_, err := newDB.MasterMappingCluster(req)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestSaveLogOrchestrator(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)

	header := "{'header':'header-id'}"
	request := "{'request':'request-id'}"
	response := "{'response':'response-id'}"
	headerByte, _ := json.Marshal(header)
	requestByte, _ := json.Marshal(request)
	responseByte, _ := json.Marshal(response)
	path := "path"
	method := "method"
	prospectID := "prospectID"
	requestID := "requestID"

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "log_orchestrators" (.*)`).
		WithArgs(requestID, prospectID, sqlmock.AnyArg(), string(headerByte), sqlmock.AnyArg(), method, string(requestByte), string(utils.SafeEncoding(responseByte)), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	err := newDB.SaveLogOrchestrator(header, request, response, path, method, prospectID, requestID)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestMasterMappingFpdCluster(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)
	FpdValue := 5.0

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT cluster FROM m_mapping_fpd_cluster WITH (nolock) 
                            WHERE (fpd_start_hte <= ? OR fpd_start_hte IS NULL) 
                            AND (fpd_end_lt > ? OR fpd_end_lt IS NULL)`)).
		WithArgs(FpdValue, FpdValue).
		WillReturnRows(sqlmock.NewRows([]string{"cluster"}).
			AddRow("Cluster A"))
	mock.ExpectCommit()

	_, err := newDB.MasterMappingFpdCluster(FpdValue)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestCheckCMONoFPD(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)
	cmoID := "CMO001"
	bpkbName := "NAMA SAMA"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 prospect_id, cmo_id, cmo_category, 
                            FORMAT(CONVERT(datetime, cmo_join_date, 127), 'yyyy-MM-dd') AS cmo_join_date, 
                            default_cluster, 
                            FORMAT(CONVERT(datetime, default_cluster_start_date, 127), 'yyyy-MM-dd') AS default_cluster_start_date, 
                            FORMAT(CONVERT(datetime, default_cluster_end_date, 127), 'yyyy-MM-dd') AS default_cluster_end_date
                          FROM dbo.trx_cmo_no_fpd WITH (nolock) 
                          WHERE cmo_id = ? AND bpkb_name = ?
                          ORDER BY created_at DESC`)).
		WithArgs(cmoID, bpkbName).
		WillReturnRows(sqlmock.NewRows([]string{"prospect_id", "cmo_id", "cmo_category", "cmo_join_date", "default_cluster", "default_cluster_start_date", "default_cluster_end_date"}).
			AddRow("TST001", cmoID, "CAT1", "2023-01-01", "Cluster1", "2023-01-01", "2023-03-31"))
	mock.ExpectCommit()

	data, err := newDB.CheckCMONoFPD(cmoID, bpkbName)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	expected := entity.TrxCmoNoFPD{
		ProspectID:              "TST001",
		CMOID:                   cmoID,
		CmoCategory:             "CAT1",
		CmoJoinDate:             "2023-01-01",
		DefaultCluster:          "Cluster1",
		DefaultClusterStartDate: "2023-01-01",
		DefaultClusterEndDate:   "2023-03-31",
	}

	if data != expected {
		t.Errorf("expected %v, got %v", expected, data)
	}
}

func TestGetResultFiltering(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB)

	prospectID := "TST001"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tf.prospect_id, tf.decision, tf.reason, tf.customer_status, tf.customer_status_kmb, tf.customer_segment, tf.is_blacklist, tf.next_process,
										tf.total_baki_debet_non_collateral_biro, tdb.url_pdf_report, tdb.subject FROM trx_filtering tf 
										LEFT JOIN trx_detail_biro tdb ON tf.prospect_id = tdb.prospect_id 
										WHERE tf.prospect_id = ?`)).
		WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"prospect_id", "decision", "reason", "customer_status", "customer_segment", "is_blacklist", "next_process", "total_baki_debet_non_collateral_biro", "url_pdf_report", "subject"}).
			AddRow("TST001", "APPROVE", "Good Credit", "Active", "Premium", false, true, 1000000, "http://example.com/report.pdf", "Credit Report"))
	mock.ExpectCommit()

	_, err := repo.GetResultFiltering(prospectID)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetConfig(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	jsonValue := `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`

	newDB := NewRepository(gormDB, gormDB, gormDB)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = 'expired_contract' AND lob = 'KMB-OFF' AND [key]= 'expired_contract_check' AND is_active = 1")).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).
			AddRow(jsonValue))
	result, err := newDB.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}

	expected := entity.AppConfig{Value: jsonValue}
	if result != expected {
		t.Errorf("expected '%v', but got '%v'", expected, result)
	}
}
