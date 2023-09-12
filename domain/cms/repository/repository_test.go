package repository

import (
	"database/sql/driver"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
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

func TestGetCustomerPhoto(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedPhoto := []entity.CustomerPhoto{{PhotoID: "1", Url: "http://example.com/photo1.jpg"}}

	// Mock SQL query and result
	mock.ExpectQuery(`SELECT photo_id, url FROM trx_customer_photo WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"photo_id", "url"}).
			AddRow(1, "http://example.com/photo1.jpg"))

	// Call the function
	photo, err := repo.GetCustomerPhoto(prospectID)

	// Verify the result
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	assert.Equal(t, expectedPhoto, photo, "Expected photo slice to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetCustomerPhoto_RecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT photo_id, url FROM trx_customer_photo WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)

	// Call the function
	_, err := repo.GetCustomerPhoto(prospectID)

	// Verify the error message
	expectedErr := errors.New(constant.RECORD_NOT_FOUND)
	assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetSurveyorData(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedSurveyor := []entity.TrxSurveyor{
		{
			ProspectID:   prospectID,
			Destination:  "HOME",
			RequestDate:  time.Now(),
			AssignDate:   time.Now(),
			SurveyorName: "RONY ACHMAD MOQAROBIN",
			ResultDate:   time.Now(),
			Status:       "APPROVE",
			SurveyorNote: nil,
		},
	}

	// Mock SQL query and result
	mock.ExpectQuery(`SELECT destination, request_date, assign_date, surveyor_name, result_date, status, surveyor_note FROM trx_surveyor WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "destination", "request_date", "assign_date", "surveyor_name", "result_date", "status", "surveyor_note"}).
			AddRow("12345", "HOME", time.Now(), time.Now(), "RONY ACHMAD MOQAROBIN", time.Now(), "APPROVE", nil))

	// Call the function
	photo, err := repo.GetSurveyorData(prospectID)

	// Verify the result
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	assert.Equal(t, expectedSurveyor, photo, "Expected surveyor slice to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetSurveyorData_RecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT destination, request_date, assign_date, surveyor_name, result_date, status, surveyor_note FROM trx_surveyor WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)

	// Call the function
	_, err := repo.GetSurveyorData(prospectID)

	// Verify the error message
	expectedErr := errors.New(constant.RECORD_NOT_FOUND)
	assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetStatusPrescreening(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedStatus := entity.TrxStatus{
		Activity:       constant.ACTIVITY_UNPROCESS,
		SourceDecision: constant.PRESCREENING,
	}

	// Mock SQL query and result
	mock.ExpectQuery(`SELECT activity, source_decision FROM trx_status WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"activity", "source_decision"}).
			AddRow(constant.ACTIVITY_UNPROCESS, constant.PRESCREENING))

	// Call the function
	photo, err := repo.GetStatusPrescreening(prospectID)

	// Verify the result
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	assert.Equal(t, expectedStatus, photo, "Expected surveyor slice to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetStatusPrescreening_RecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT activity, source_decision FROM trx_status WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)

	// Call the function
	_, err := repo.GetStatusPrescreening(prospectID)

	// Verify the error message
	expectedErr := errors.New(constant.RECORD_NOT_FOUND)
	assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetReasonPrescreening(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output

	reasonID := "99"
	expectedReason := []entity.ReasonMessage{
		{
			ReasonID:      "11",
			ReasonMessage: "Akte Jual Beli Tidak Sesuai",
			Code:          "12",
		},
	}

	// Mock SQL query and result
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(tt.ReasonID) AS totalRow FROM (SELECT ReasonID FROM reason_message WITH (nolock)) AS tt WHERE ReasonID != '99'`)).
		WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
			AddRow("27"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (SELECT Code, ReasonID, ReasonMessage FROM reason_message WITH (nolock)) AS tt WHERE ReasonID != '99' ORDER BY tt.ReasonID asc OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
		WillReturnRows(sqlmock.NewRows([]string{"Code", "ReasonID", "ReasonMessage"}).
			AddRow("12", "11", "Akte Jual Beli Tidak Sesuai"))

	// Call the function
	reason, _, err := repo.GetReasonPrescreening(reasonID, 1)

	// Verify the result
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	assert.Equal(t, expectedReason, reason, "Expected reason slice to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetReasonPrescreening_RecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output
	reasonID := "99"

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (SELECT Code, ReasonID, ReasonMessage FROM reason_message WITH (nolock)) AS tt WHERE ReasonID != '99' ORDER BY tt.ReasonID asc`)).
		WillReturnError(gorm.ErrRecordNotFound)

	// Call the function
	_, _, err := repo.GetReasonPrescreening(reasonID, nil)

	// Verify the error message
	expectedErr := errors.New(constant.RECORD_NOT_FOUND)
	assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetSpIndustryTypeMaster(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Expected input and output

	expectedIndustry := []entity.SpIndustryTypeMaster{
		{
			IndustryTypeID: "SE_e6611  ",
			Description:    "Administrasi Pasar Keuangan",
			IsActive:       true,
		},
	}

	// Mock SQL query and result

	mock.ExpectQuery(regexp.QuoteMeta(`exec[spIndustryTypeMaster] '01/01/2007'`)).
		WillReturnRows(sqlmock.NewRows([]string{"IndustryTypeID", "Description", "IsActive"}).
			AddRow("SE_e6611  ", "Administrasi Pasar Keuangan", "True"))

	// Call the function
	industry, err := repo.GetSpIndustryTypeMaster()

	// Verify the result
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	assert.Equal(t, expectedIndustry, industry, "Expected industry slice to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetSpIndustryTypeMaster_RecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB)

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(regexp.QuoteMeta(`exec[spIndustryTypeMaster] '01/01/2007'`)).
		WillReturnError(gorm.ErrRecordNotFound)

	// Call the function
	_, err := repo.GetSpIndustryTypeMaster()

	// Verify the error message
	expectedErr := errors.New(constant.RECORD_NOT_FOUND)
	assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}
