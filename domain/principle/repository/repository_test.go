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
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
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

func TestGetConfig(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	jsonValue := `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = 'expired_contract' AND lob = 'KMB-OFF' AND [key]= 'expired_contract_check' AND is_active = 1")).
		WillReturnRows(sqlmock.NewRows([]string{"value"}).
			AddRow(jsonValue))
	result, err := repo.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}

	expected := entity.AppConfig{Value: jsonValue}
	if result != expected {
		t.Errorf("expected '%v', but got '%v'", expected, result)
	}
}

func TestGetMinimalIncomePMK(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	branchID := "BR123"
	statusKonsumen := "new"
	response := entity.MappingIncomePMK{
		Income:         3000000,
		BranchID:       "426",
		ID:             "123",
		StatusKonsumen: "NEW",
		Lob:            "los_kmb_off",
	}

	query := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "branch_id", "status_konsumen", "income", "lob"}).
			AddRow(response.ID, response.BranchID, response.StatusKonsumen, response.Income, response.Lob))

	result, err := repo.GetMinimalIncomePMK(branchID, statusKonsumen)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMinimalIncomePMK_DefaultBranch(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	branchID := "BR123"
	statusKonsumen := "existing"
	defaultBranchID := constant.DEFAULT_BRANCH_ID
	response := entity.MappingIncomePMK{
		Income:         3000000,
		BranchID:       "426",
		ID:             "123",
		StatusKonsumen: "NEW",
		Lob:            "los_kmb_off",
	}

	query := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnError(gorm.ErrRecordNotFound)

	defaultQuery := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, defaultBranchID, statusKonsumen)
	mock.ExpectQuery(regexp.QuoteMeta(defaultQuery)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "branch_id", "status_konsumen", "income", "lob"}).
			AddRow(response.ID, response.BranchID, response.StatusKonsumen, response.Income, response.Lob))

	result, err := repo.GetMinimalIncomePMK(branchID, statusKonsumen)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestMasterMappingFpdCluster(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	fpdValue := 80.0
	response := entity.MasterMappingFpdCluster{
		Cluster: "Cluster A",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT cluster FROM m_mapping_fpd_cluster WITH (nolock) 
                            WHERE (fpd_start_hte <= ? OR fpd_start_hte IS NULL) 
                            AND (fpd_end_lt > ? OR fpd_end_lt IS NULL)`)).
		WithArgs(fpdValue, fpdValue).
		WillReturnRows(sqlmock.NewRows([]string{"cluster"}).
			AddRow(response.Cluster))
	mock.ExpectCommit()

	result, err := repo.MasterMappingFpdCluster(fpdValue)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestMasterMappingCluster(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)
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

	_, err := repo.MasterMappingCluster(req)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSaveFiltering(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)
	trxFiltering := entity.FilteringKMB{
		ProspectID: "TST001",
		Decision:   "PASS",
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
	cmoNoFPD := structToSlice(trxCmoNoFPD)
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "trx_filtering" (.*)`).
		WithArgs(filtering...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "trx_cmo_no_fpd" (.*)`).
		WithArgs(cmoNoFPD...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "trx_detail_biro" (.*)`).
		WithArgs(detailBiro...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SaveFiltering(trxFiltering, trxDetailBiro, trxCmoNoFPD)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetBannedPMKDSR(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	idNumber := "1140024080800016"
	date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)
	response := entity.TrxBannedPMKDSR{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
	}

	query := fmt.Sprintf(`SELECT * FROM trx_banned_pmk_dsr WITH (nolock) WHERE IDNumber = '%s' AND CAST(created_at as DATE) >= '%s'`, idNumber, date)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber"}).
			AddRow(response.ProspectID, response.IDNumber))

	result, err := repo.GetBannedPMKDSR(idNumber)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetEncB64(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	myString := "1140024080800016"
	response := entity.EncryptedString{
		MyString: myString,
	}

	query := fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS my_string`, myString)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"my_string"}).
			AddRow(response.MyString))

	result, err := repo.GetEncB64(myString)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetRejection(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	idNumber := "1140024080800016"
	currentDate := time.Now().Format(constant.FORMAT_DATE)

	response := entity.TrxReject{
		RejectPMKDSR: 10,
		RejectNIK:    10,
	}

	query := fmt.Sprintf(`SELECT 
	COUNT(CASE WHEN ts.source_decision = 'PMK' OR ts.source_decision = 'DSR' OR ts.source_decision = 'PRJ' THEN 1 END) as reject_pmk_dsr,
	COUNT(CASE WHEN ts.source_decision != 'PMK' AND ts.source_decision != 'DSR' AND ts.source_decision != 'PRJ' AND ts.source_decision != 'NKA' THEN 1 END) as reject_nik 
	FROM trx_status ts WITH (nolock) LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID
	WHERE ts.decision = 'REJ' AND tcp.IDNumber = '%s' AND CAST(ts.created_at as DATE) = '%s'`, idNumber, currentDate)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"reject_pmk_dsr", "reject_nik"}).
			AddRow(response.RejectPMKDSR, response.RejectNIK))

	result, err := repo.GetRejection(idNumber)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapilVD_NewCustomerValid(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := constant.EKYC_RTO
	customerStatus := constant.STATUS_KONSUMEN_NEW
	customerSegment := "segment1"
	isValid := true

	response := entity.MappingResultDukcapilVD{
		ResultVD:       "valid",
		StatusKonsumen: "new",
	}

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s'`, statusVD, customerStatus)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"result_vd", "status_konsumen"}).
			AddRow(response.ResultVD, response.StatusKonsumen))

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapilVD_NewCustomerInvalid(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := "some_invalid_status"
	customerStatus := constant.STATUS_KONSUMEN_NEW
	customerSegment := "segment1"
	isValid := false

	response := entity.MappingResultDukcapilVD{
		ResultVD:       "invalid",
		StatusKonsumen: "new",
	}

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND is_valid=0`, customerStatus)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"result_vd", "status_konsumen"}).
			AddRow(response.ResultVD, response.StatusKonsumen))

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapilVD_ExistingCustomerValid(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := constant.EKYC_NOT_CHECK
	customerStatus := "existing"
	customerSegment := "segment1"
	isValid := true

	response := entity.MappingResultDukcapilVD{
		ResultVD:               "valid",
		StatusKonsumen:         "existing",
		KategoriStatusKonsumen: "segment1",
	}

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s' AND kategori_status_konsumen='%s'`, statusVD, customerStatus, customerSegment)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"result_vd", "status_konsumen", "kategori_status_konsumen"}).
			AddRow(response.ResultVD, response.StatusKonsumen, response.KategoriStatusKonsumen))

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapilVD_ExistingCustomerInvalid(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := "some_invalid_status"
	customerStatus := "existing"
	customerSegment := "segment1"
	isValid := false

	response := entity.MappingResultDukcapilVD{
		ResultVD:               "invalid",
		StatusKonsumen:         "existing",
		KategoriStatusKonsumen: "segment1",
	}

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND kategori_status_konsumen='%s' AND is_valid=0`, customerStatus, customerSegment)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"result_vd", "status_konsumen", "kategori_status_konsumen"}).
			AddRow(response.ResultVD, response.StatusKonsumen, response.KategoriStatusKonsumen))

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapil_NewCustomer(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := constant.EKYC_RTO
	customerStatus := constant.STATUS_KONSUMEN_NEW
	customerSegment := "segment1"
	statusFR := "test"

	response := entity.MappingResultDukcapil{
		ResultVD:       "valid",
		StatusKonsumen: "new",
	}

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s'`, statusVD, statusFR, customerStatus)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"result_vd", "status_konsumen"}).
			AddRow(response.ResultVD, response.StatusKonsumen))

	result, err := repo.GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapil_AOROCustomer(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := constant.EKYC_RTO
	customerStatus := constant.STATUS_KONSUMEN_AO
	customerSegment := "segment1"
	statusFR := "test"

	response := entity.MappingResultDukcapil{
		ResultVD:       "valid",
		StatusKonsumen: "new",
	}

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s'`, statusVD, statusFR, customerStatus)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"result_vd", "status_konsumen"}).
			AddRow(response.ResultVD, response.StatusKonsumen))

	result, err := repo.GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestCheckCMONoFPD(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)
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

	data, err := repo.CheckCMONoFPD(cmoID, bpkbName)
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

func TestGetTrxPrincipleStatus(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	nik := "1140024080800016"
	response := entity.TrxPrincipleStatus{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
	}

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_status WITH (nolock) WHERE IDNumber = '%s' ORDER BY created_at DESC`, nik)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber"}).
			AddRow(response.ProspectID, response.IDNumber))

	result, err := repo.GetTrxPrincipleStatus(nik)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepOne(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	dataStepOne := entity.TrxPrincipleStepOne{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	dataStatus := entity.TrxPrincipleStatus{
		ProspectID: dataStepOne.ProspectID,
		IDNumber:   dataStepOne.IDNumber,
		Step:       1,
		Decision:   dataStepOne.Decision,
		UpdatedAt:  time.Now(),
	}

	mock.ExpectBegin()

	stepOne := structToSlice(dataStepOne)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_one" (.*)`).
		WithArgs(stepOne...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO "trx_principle_status" (.*)`).
		WithArgs(dataStatus.ProspectID, dataStatus.IDNumber, dataStatus.Step, dataStatus.Decision, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.SavePrincipleStepOne(dataStepOne)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPrincipleStepOne(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	response := entity.TrxPrincipleStepOne{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
	}

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber"}).
			AddRow(response.ProspectID, response.IDNumber))

	result, err := repo.GetPrincipleStepOne(prospectID)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePrincipleStepOne(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	existingData := entity.TrxPrincipleStepOne{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "pending",
		CreatedAt:  time.Now().Add(-time.Hour),
	}

	updatedData := entity.TrxPrincipleStepOne{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "trx_principle_step_one" WHERE (ProspectID = ?) ORDER BY created_at DESC LIMIT 1`)).
		WithArgs(existingData.ProspectID).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber", "Decision", "created_at"}).
			AddRow(existingData.ProspectID, existingData.IDNumber, existingData.Decision, existingData.CreatedAt))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_principle_step_one" SET "Decision" = ?, "IDNumber" = ?, "ProspectID" = ? WHERE (ProspectID = ?)`)).
		WithArgs(updatedData.Decision, updatedData.IDNumber, updatedData.ProspectID, updatedData.ProspectID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.UpdatePrincipleStepOne(existingData.ProspectID, updatedData)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepTwo(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	dataStepTwo := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	dataStatus := entity.TrxPrincipleStatus{
		ProspectID: dataStepTwo.ProspectID,
		IDNumber:   dataStepTwo.IDNumber,
		Step:       2,
		Decision:   dataStepTwo.Decision,
		UpdatedAt:  time.Now(),
	}

	mock.ExpectBegin()

	stepOne := structToSlice(dataStepTwo)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_two" (.*)`).
		WithArgs(stepOne...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO "trx_principle_status" (.*)`).
		WithArgs(dataStatus.ProspectID, dataStatus.IDNumber, dataStatus.Step, dataStatus.Decision, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.SavePrincipleStepTwo(dataStepTwo)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPrincipleStepTwo(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	response := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
	}

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_two WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber"}).
			AddRow(response.ProspectID, response.IDNumber))

	result, err := repo.GetPrincipleStepTwo(prospectID)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetFilteringResult(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	query := `SELECT bpkb_name, customer_status, decision, reason, is_blacklist, next_process, max_overdue_biro, max_overdue_last12months_biro, customer_segment, total_baki_debet_non_collateral_biro, score_biro, cluster, cmo_cluster, FORMAT(rrd_date, 'yyyy-MM-ddTHH:mm:ss') + 'Z' AS rrd_date, created_at FROM trx_filtering WITH (nolock) WHERE prospect_id = ?`

	prospectID := "TEST0001"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"bpkb_name", "customer_status", "decision"}).
			AddRow("O", "NEW", "PASS"))
	mock.ExpectCommit()

	_, err := repo.GetFilteringResult(prospectID)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetMappingElaborateLTV(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	query := `SELECT * FROM m_mapping_elaborate_ltv WITH (nolock) WHERE result_pefindo = ? AND cluster = ? `

	resultPefindo := "PASS"
	cluster := "Cluster A"

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(resultPefindo, cluster).
		WillReturnRows(sqlmock.NewRows([]string{"result_pefindo", "cluster", "total_baki_debet_start"}).
			AddRow("AVERAGE RISK", "Cluster A", 0))
	mock.ExpectCommit()

	_, err := repo.GetMappingElaborateLTV(resultPefindo, cluster)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestSaveTrxElaborateLTV(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	data := entity.TrxElaborateLTV{
		ProspectID:        "SAL-1140024080800016",
		RequestID:         "REQ-001",
		Tenor:             12,
		ManufacturingYear: "2023",
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_elaborate_ltv" SET "created_at" = ?, "manufacturing_year" = ?, "request_id" = ?, "tenor" = ? WHERE (prospect_id = ?)`)).
		WithArgs(sqlmock.AnyArg(), data.ManufacturingYear, data.RequestID, data.Tenor, data.ProspectID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.SaveTrxElaborateLTV(data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetMappingVehicleAge(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	vehicleAge := 12
	cluster := "Cluster A"
	bpkbNameType := 1
	tenor := 12
	resultPefindo := "PASS"
	af := 1200000.0

	expectedQuery := `SELECT TOP 1 \* FROM m_mapping_vehicle_age WHERE vehicle_age_start <= \? AND vehicle_age_end >= \? AND cluster LIKE \? AND bpkb_name_type = \? AND tenor_start <= \? AND tenor_end >= \? AND result_pbk LIKE \? AND af_start < \? AND af_end >= \?`

	mock.ExpectQuery(expectedQuery).
		WithArgs(vehicleAge, vehicleAge, fmt.Sprintf("%%%s%%", cluster), bpkbNameType, tenor, tenor, fmt.Sprintf("%%%s%%", resultPefindo), af, af).
		WillReturnRows(sqlmock.NewRows([]string{"vehicle_age_start", "vehicle_age_end", "cluster", "decision"}).
			AddRow(12, 24, "Cluster A", "PASS"))

	result, err := repo.GetMappingVehicleAge(vehicleAge, cluster, bpkbNameType, tenor, resultPefindo, af)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedResult := entity.MappingVehicleAge{
		VehicleAgeStart: 12,
		VehicleAgeEnd:   24,
		Cluster:         "Cluster A",
		Decision:        "PASS",
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("expected %+v, but got %+v", expectedResult, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetScoreGenerator(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	zipCode := "12345"

	expectedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT TOP 1 x.* 
	FROM
	(
		SELECT 
		a.[key],
		b.id AS score_generators_id
		FROM [dbo].[score_models_rules_data] a
		INNER JOIN score_generators b
		ON b.id = a.score_generators
		WHERE (a.[key] = 'first_residence_zipcode_2w_jabo' AND a.[value] = '%s')
		OR (a.[key] = 'first_residence_zipcode_2w_others' AND a.[value] = '%s')
	)x`, zipCode, zipCode))

	mock.ExpectQuery(expectedQuery).
		WillReturnRows(sqlmock.NewRows([]string{"key", "score_generators_id"}).
			AddRow("first_residence_zipcode_2w_jabo", "TEST"))

	result, err := repo.GetScoreGenerator(zipCode)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedResult := entity.ScoreGenerator{
		Key:               "first_residence_zipcode_2w_jabo",
		ScoreGeneratorsID: "TEST",
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("expected %+v, but got %+v", expectedResult, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetScoreGeneratorROAO(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	expectedQuery := regexp.QuoteMeta(`SELECT TOP 1 x.* 
	FROM
	(
		SELECT 
		a.[key],
		b.id AS score_generators_id
		FROM [dbo].[score_models_rules_data] a
		INNER JOIN score_generators b
		ON b.id = a.score_generators
		WHERE a.[key] = 'first_residence_zipcode_2w_aoro'
	)x`)

	mock.ExpectQuery(expectedQuery).
		WillReturnRows(sqlmock.NewRows([]string{"key", "score_generators_id"}).
			AddRow("first_residence_zipcode_2w_jabo", "TEST"))

	result, err := repo.GetScoreGeneratorROAO()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedResult := entity.ScoreGenerator{
		Key:               "first_residence_zipcode_2w_jabo",
		ScoreGeneratorsID: "TEST",
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("expected %+v, but got %+v", expectedResult, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetTrxDetailBIro(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "PROSPECT123"

	mock.ExpectBegin()

	expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT * FROM trx_detail_biro WITH (nolock) WHERE prospect_id = '%s'", prospectID))

	rows := sqlmock.NewRows([]string{"prospect_id"}).
		AddRow(prospectID)

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

	mock.ExpectCommit()

	result, err := repo.GetTrxDetailBIro(prospectID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectedResult := []entity.TrxDetailBiro{
		{
			ProspectID: prospectID,
		},
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Errorf("expected %+v, but got %+v", expectedResult, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetActiveLoanTypeLast6M(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	customerID := "CUSTOMER123"

	expectedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT CustomerID, Concat([1],' ',';',' ',[2],' ',';',' ',[3]) AS active_loanType_last6m FROM
	( SELECT * FROM
		( SELECT  CustomerID, PRODUCT, Seq_PRODUCT FROM
			( SELECT DISTINCT CustomerID,PRODUCT, ROW_NUMBER() OVER (PARTITION BY CustomerID Order By PRODUCT DESC) AS Seq_PRODUCT FROM
				( SELECT DISTINCT CustomerID,
						CASE WHEN ContractStatus in ('ICP','PRP','LIV','RRD','ICL','INV') and MOB<=-1 and MOB>=-6 THEN PRODUCT
						END AS 'PRODUCT'
					FROM
					( SELECT CustomerID,A.ApplicationID,DATEDIFF(MM, GETDATE(), AgingDate) AS 'MOB', CAST(aa.AssetTypeID AS int) AS PRODUCT, ContractStatus FROM																	   
						( SELECT * FROM Agreement a WITH (NOLOCK) WHERE a.CustomerID = '%s' 
						)A
						LEFT JOIN
						( SELECT DISTINCT ApplicationID,AgingDate,EndPastDueDays FROM SBOAging WITH (NOLOCK)
							WHERE ApplicationID IN (SELECT DISTINCT a.ApplicationID  FROM Agreement a WITH (NOLOCK)) AND AgingDate=EOMONTH(AgingDate)
						)B ON A.ApplicationID=B.ApplicationID
						LEFT JOIN AgreementAsset aa WITH (NOLOCK) ON A.ApplicationID = aa.ApplicationID
					)S
				)T
			)U
		) AS SourceTable PIVOT(AVG(PRODUCT) FOR Seq_PRODUCT IN([1],[2],[3])) AS PivotTable
	)V`, customerID))

	rows := sqlmock.NewRows([]string{"CustomerID", "active_loanType_last6m"}).
		AddRow(customerID, "test")

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

	result, err := repo.GetActiveLoanTypeLast6M(customerID)

	assert.NoError(t, err)

	expectedResult := entity.GetActiveLoanTypeLast6M{
		CustomerID:           customerID,
		ActiveLoanTypeLast6M: "test",
	}

	assert.Equal(t, expectedResult, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetActiveLoanTypeLast24M(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := repoHandler{
		confins: gormDB,
	}

	customerID := "CUSTOMER123"

	expectedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT a.AgreementNo, MIN(DATEDIFF(MM, GETDATE(), s.AgingDate)) AS 'MOB' FROM Agreement a WITH (NOLOCK)
	LEFT JOIN SBOAging s WITH (NOLOCK) ON s.ApplicationId = a.ApplicationID
	WHERE a.ContractStatus in ('ICP','PRP','LIV','RRD','ICL','INV') 
	AND DATEDIFF(MM, GETDATE(), s.AgingDate)<=-7 AND DATEDIFF(MM, GETDATE(), s.AgingDate)>=-24
	AND a.CustomerID = '%s' GROUP BY a.AgreementNo`, customerID))

	rows := sqlmock.NewRows([]string{"AgreementNo", "MOB"}).
		AddRow("AGR001", "test")

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

	result, err := repo.GetActiveLoanTypeLast24M(customerID)

	assert.NoError(t, err)

	expectedResult := entity.GetActiveLoanTypeLast24M{
		AgreementNo: "AGR001",
		MOB:         "test",
	}

	assert.Equal(t, expectedResult, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetMoblast(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := repoHandler{
		confins: gormDB,
	}

	customerID := "CUSTOMER123"

	expectedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT TOP 1 DATEDIFF(MM, GoLiveDate, GETDATE()) AS 'moblast' FROM Agreement a WITH (NOLOCK) 
	WHERE a.CustomerID = '%s' AND a.GoLiveDate IS NOT NULL ORDER BY a.GoLiveDate DESC`, customerID))

	rows := sqlmock.NewRows([]string{"moblast"}).
		AddRow("test")

	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)

	result, err := repo.GetMoblast(customerID)

	assert.NoError(t, err)

	expectedResult := entity.GetMoblast{
		Moblast: "test",
	}

	assert.Equal(t, expectedResult, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestSavePrincipleStepThree(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	dataStepThree := entity.TrxPrincipleStepThree{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	dataStatus := entity.TrxPrincipleStatus{
		ProspectID: dataStepThree.ProspectID,
		IDNumber:   dataStepThree.IDNumber,
		Step:       3,
		Decision:   dataStepThree.Decision,
		UpdatedAt:  time.Now(),
	}

	mock.ExpectBegin()

	stepOne := structToSlice(dataStepThree)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_three" (.*)`).
		WithArgs(stepOne...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO "trx_principle_status" (.*)`).
		WithArgs(dataStatus.ProspectID, dataStatus.IDNumber, dataStatus.Step, dataStatus.Decision, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.SavePrincipleStepThree(dataStepThree)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPrincipleStepThree(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	response := entity.TrxPrincipleStepThree{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
	}

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_three WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber"}).
			AddRow(response.ProspectID, response.IDNumber))

	result, err := repo.GetPrincipleStepThree(prospectID)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleEmergencyContact(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	data := entity.TrxPrincipleEmergencyContact{
		ProspectID:        "SAL-1140024080800016",
		Name:              "Test",
		Relationship:      "",
		MobilePhone:       "",
		CompanyStreetName: "",
		HomeNumber:        "",
		LocationDetails:   "",
		Rt:                "",
		Rw:                "",
		Kelurahan:         "",
		Kecamatan:         "",
		City:              "",
		Province:          "",
		ZipCode:           "",
		AreaPhone:         "",
		Phone:             "",
		CustomerID:        0,
		KPMID:             0,
	}

	idNumber := "123456"

	t.Run("Create New Record", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectExec(`INSERT INTO "trx_principle_emergency_contact" (.*)`).
			WithArgs(
				data.ProspectID, data.Name, data.Relationship, data.MobilePhone,
				data.CompanyStreetName, data.HomeNumber, data.LocationDetails,
				data.Rt, data.Rw, data.Kelurahan, data.Kecamatan, data.City,
				data.Province, data.ZipCode, data.AreaPhone, data.Phone,
				data.CustomerID, data.KPMID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`INSERT INTO "trx_principle_status" (.*)`).
			WithArgs(data.ProspectID, idNumber, 4, constant.DECISION_CREDIT_PROCESS, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)

		if err != nil {
			t.Errorf("error was not expected while saving new record: %s", err)
		}
	})

	t.Run("Update Existing Record", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID"}).AddRow(data.ProspectID))

		mock.ExpectExec(`UPDATE "trx_principle_emergency_contact" SET (.*) WHERE .*ProspectID = \?`).
			WithArgs(
				data.Name, data.ProspectID, sqlmock.AnyArg(), data.ProspectID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`UPDATE "trx_principle_status" SET (.*) WHERE .*ProspectID = \? AND Step = \?`).
			WithArgs(constant.DECISION_CREDIT_PROCESS, idNumber, sqlmock.AnyArg(), data.ProspectID, 4).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)

		if err != nil {
			t.Errorf("error was not expected while updating existing record: %s", err)
		}
	})

	t.Run("Database Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnError(fmt.Errorf("database error"))

		mock.ExpectRollback()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)

		if err == nil {
			t.Error("error was expected but got nil")
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPrincipleEmergencyContact(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	response := entity.TrxPrincipleEmergencyContact{
		ProspectID: "SAL-1140024080800016",
		Name:       "Test",
	}

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_emergency_contact WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "Name"}).
			AddRow(response.ProspectID, response.Name))

	result, err := repo.GetPrincipleEmergencyContact(prospectID)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if result != response {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSaveToWorker(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	workers := []entity.TrxWorker{
		{ProspectID: "SAL-1140024080800016", Activity: "Test"},
		{ProspectID: "SAL-1140024080800017", Activity: "Test"},
	}

	mock.ExpectBegin()
	for _, worker := range workers {
		mock.ExpectExec(`INSERT INTO "trx_worker" (.+) VALUES (.+)`).
			WithArgs(
				worker.ProspectID,
				worker.Activity,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectCommit()

	err := repo.SaveToWorker(workers)

	if err != nil {
		t.Errorf("error was not expected while saving workers: %s", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetElaborateLtv(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	response := entity.MappingElaborateLTV{
		LTV: 80,
	}

	query := fmt.Sprintf(`SELECT CASE WHEN mmel.ltv IS NULL THEN mmelovd.ltv ELSE mmel.ltv END AS ltv FROM trx_elaborate_ltv tel WITH (nolock) 
	LEFT JOIN m_mapping_elaborate_ltv mmel WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmel.id
	LEFT JOIN m_mapping_elaborate_ltv_ovd mmelovd WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmelovd.id 
	WHERE tel.prospect_id ='%s'`, prospectID)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ltv"}).
			AddRow(response.LTV))

	result, err := repo.GetElaborateLtv(prospectID)

	if err != nil {
		t.Errorf("error was not expected while getting elaborate LTV: %s", err)
	}

	if result != response {
		t.Errorf("expected %v, but got %v", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleMarketingProgram(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	data := entity.TrxPrincipleMarketingProgram{
		ProspectID:                 "PMP-1234567890",
		ProgramID:                  "",
		ProgramName:                "Sample Marketing Program",
		ProductOfferingID:          "",
		ProductOfferingDescription: "",
		LoanAmount:                 0,
		LoanAmountMaximum:          0,
		AdminFee:                   0,
		ProvisionFee:               0,
		DPAmount:                   0,
		FinanceAmount:              0,
	}

	mock.ExpectBegin()

	mock.ExpectExec(`INSERT INTO "trx_principle_marketing_program" (.*)`).
		WithArgs(
			data.ProspectID,
			data.ProgramID,
			data.ProgramName,
			data.ProductOfferingID,
			data.ProductOfferingDescription,
			data.LoanAmount,
			data.LoanAmountMaximum,
			data.AdminFee,
			data.ProvisionFee,
			data.DPAmount,
			data.FinanceAmount,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.SavePrincipleMarketingProgram(data)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPrincipleMarketingProgram(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "PMP-1234567890"
	response := entity.TrxPrincipleMarketingProgram{
		ProspectID:                 "PMP-1234567890",
		ProgramID:                  "PROG001",
		ProgramName:                "Sample Marketing Program",
		ProductOfferingID:          "PROD001",
		ProductOfferingDescription: "Sample Product Offering",
		LoanAmount:                 10000,
		LoanAmountMaximum:          15000,
		AdminFee:                   100,
		ProvisionFee:               50,
		DPAmount:                   1000,
		FinanceAmount:              9000,
	}

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{
			"ProspectID", "ProgramID", "ProgramName", "ProductOfferingID", "ProductOfferingDescription",
			"LoanAmount", "LoanAmountMaximum", "AdminFee", "ProvisionFee", "DPAmount", "FinanceAmount",
		}).AddRow(
			response.ProspectID, response.ProgramID, response.ProgramName, response.ProductOfferingID,
			response.ProductOfferingDescription, response.LoanAmount, response.LoanAmountMaximum,
			response.AdminFee, response.ProvisionFee, response.DPAmount, response.FinanceAmount,
		))

	result, err := repo.GetPrincipleMarketingProgram(prospectID)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if !reflect.DeepEqual(result, response) {
		t.Errorf("expected '%v', but got '%v'", response, result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
