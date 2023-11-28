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
	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "trx_filtering" (.*)`).
		WithArgs(filtering...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO "trx_detail_biro" (.*)`).
		WithArgs(detailBiro...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := newDB.SaveFiltering(trxFiltering, trxDetailBiro)
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
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM dbo.m_mapping_cluster WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?")).
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
