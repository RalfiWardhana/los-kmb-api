package repository

import (
	"database/sql/driver"
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

func TestDummyDataPbk(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM dbo.dummy_pefindo_kmb WHERE IDNumber = ?")).
		WithArgs("TST001").
		WillReturnRows(sqlmock.NewRows([]string{"prospect_id"}).
			AddRow("TST001"))

	_, err := newDB.DummyDataPbk("TST001")
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestGetMappingDp(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)

	branchID := "426"
	statusKonsumen := constant.STATUS_KONSUMEN_NEW

	mock.ExpectQuery(regexp.QuoteMeta("SELECT mbd.* FROM dbo.mapping_branch_dp mdp LEFT JOIN dbo.mapping_baki_debet mbd ON mdp.baki_debet = mbd.id LEFT JOIN dbo.master_list_dp mld ON mdp.master_list_dp = mld.id WHERE mdp.branch = ? AND mdp.customer_status = ?")).
		WithArgs(branchID, statusKonsumen).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "range_start", "range_end", "created_at"}).
			AddRow("379f6c45-8baf-4152-a3a6-c47e3452ecac", "baki_debet_1", 0, 3000000, "2022-09-19 11:46:32.000"))

	_, err := newDB.DataGetMappingDp(branchID, statusKonsumen)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestBranchDpData(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)

	query := "SELECT mdp.branch, mdp.customer_status, mdp.profession_group, mbd.name AS minimal_dp_name, mbd.range_start AS minimal_dp_value FROM dbo.mapping_branch_dp mdp LEFT JOIN dbo.mapping_baki_debet mbd ON mdp.baki_debet = mbd.id LEFT JOIN dbo.master_list_dp mld ON mdp.master_list_dp = mld.id"

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"branch", "customer_status", "profession_group", "minimal_dp_name", "minimal_dp_value"}).
			AddRow("912", "NEW", "KARYAWAN", "baki_debet_1", "3000001"))

	_, err := newDB.BranchDpData(query)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestSaveData(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)

	reqFiltering := entity.ApiDupcheckKmb{
		ProspectID: "TST001",
		RequestID:  "a6c09ce7-9d6d-4962-b32d-61a9e64d9be7",
		Request:    `{"client_key":"$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6","data":{"BPKBName":"P","ProspectID":"3201020902820017","BranchID":"621","IDNumber":"3275066006789999","LegalName":"INDAH KIKI DEWANTI","BirthPlace":"MEDAN","BirthDate":"1989-12-13","SurgateMotherName":"SUSILAWATI","Gender":"F","MaritalStatus":"M","ProfessionID":"KRYSW","Spouse":{"Spouse_IDNumber":"1207211506880005","Spouse_LegalName":"BASIR ERIK SAHMANA","Spouse_BirthPlace":"PATUMBAK","Spouse_BirthDate":"1988-06-15","Spouse_SurgateMotherName":"SITI HABSAH","Spouse_Gender":"M"},"MobilePhone":"082362557491"}}`,
		Code:       constant.WO_AGUNAN_PASS_CODE,
		Decision:   constant.DECISION_REJECT,
		Reason:     constant.TIDAK_ADA_FASILITAS_WO_AGUNAN,
	}
	filtering := structToSlice(reqFiltering)

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "api_dupcheck_kmb" (.*)`).
		WithArgs(filtering[0], filtering[1], filtering[2], filtering[3], filtering[4], filtering[5], sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := newDB.SaveData(reqFiltering)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}

func TestUpdateData(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB)

	req := entity.ApiDupcheckKmbUpdate{
		ProspectID: "TST001",
		RequestID:  "a6c09ce7-9d6d-4962-b32d-61a9e64d9be7",
		Code:       constant.WO_AGUNAN_PASS_CODE,
		Decision:   constant.DECISION_REJECT,
		Reason:     constant.TIDAK_ADA_FASILITAS_WO_AGUNAN,
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "api_dupcheck_kmb" SET "Code" = ?, "Decision" = ?, "DtmResponse" = ?, "ProspectID" = ?, "Reason" = ?, "RequestID" = ?, "Timestamp" = ? WHERE (RequestID = ?)`)).
		WithArgs(req.Code, req.Decision, sqlmock.AnyArg(), req.ProspectID, req.Reason, req.RequestID, sqlmock.AnyArg(), req.RequestID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := newDB.UpdateData(req)
	if err != nil {
		t.Errorf("error '%s' was not expected, but got: ", err)
	}
}
