package repository

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
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

func TestGetCustomerPhoto(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedPhoto := []entity.DataPhoto{{PhotoID: "1", Label: "KTP", Url: "http://example.com/photo1.jpg"}}

	mock.ExpectBegin()

	// Mock SQL query and result
	mock.ExpectQuery(`SELECT tcp.photo_id, CASE WHEN lpi.Name IS NULL THEN 'LAINNYA' ELSE lpi.Name END AS label, tcp.url FROM trx_customer_photo tcp WITH \(nolock\) LEFT JOIN m_label_photo_inquiry lpi ON lpi.LabelPhotoID = tcp.photo_id WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"photo_id", "label", "url"}).
			AddRow(1, "KTP", "http://example.com/photo1.jpg"))
	mock.ExpectCommit()

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
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	mock.ExpectBegin()

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT tcp.photo_id, CASE WHEN lpi.Name IS NULL THEN 'LAINNYA' ELSE lpi.Name END AS label, tcp.url FROM trx_customer_photo tcp WITH \(nolock\) LEFT JOIN m_label_photo_inquiry lpi ON lpi.LabelPhotoID = tcp.photo_id WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectCommit()

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
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedSurveyor := []entity.TrxSurveyor{
		{
			ProspectID:   prospectID,
			Destination:  "HOME",
			RequestDate:  time.Time{},
			AssignDate:   time.Time{},
			SurveyorName: "RONY ACHMAD MOQAROBIN",
			ResultDate:   time.Time{},
			Status:       "APPROVE",
			SurveyorNote: nil,
		},
	}

	// Mock SQL query and result
	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT destination, request_date, assign_date, surveyor_name, result_date, status, surveyor_note FROM trx_surveyor WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "destination", "request_date", "assign_date", "surveyor_name", "result_date", "status", "surveyor_note"}).
			AddRow("12345", "HOME", time.Time{}, time.Time{}, "RONY ACHMAD MOQAROBIN", time.Time{}, "APPROVE", nil))
	mock.ExpectCommit()

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
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	mock.ExpectBegin()

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT destination, request_date, assign_date, surveyor_name, result_date, status, surveyor_note FROM trx_surveyor WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectCommit()

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
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedStatus := entity.TrxStatus{
		Activity:       constant.ACTIVITY_UNPROCESS,
		SourceDecision: constant.PRESCREENING,
	}

	mock.ExpectBegin()

	// Mock SQL query and result
	mock.ExpectQuery(`SELECT activity, decision, source_decision FROM trx_status WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"activity", "source_decision"}).
			AddRow(constant.ACTIVITY_UNPROCESS, constant.PRESCREENING))
	mock.ExpectCommit()

	// Call the function
	photo, err := repo.GetTrxStatus(prospectID)

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
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"

	mock.ExpectBegin()

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT activity, decision, source_decision FROM trx_status WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectCommit()

	// Call the function
	_, err := repo.GetTrxStatus(prospectID)

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
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output

	req := request.ReqReasonPrescreening{
		ReasonID: "99,100,101,102",
	}
	expectedReason := []entity.ReasonMessage{
		{
			ReasonID:      "11",
			ReasonMessage: "Akte Jual Beli Tidak Sesuai",
			Code:          "12",
		},
	}

	// Mock SQL query and result

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(tt.ReasonID) AS totalRow FROM (SELECT ReasonID FROM m_reason_message WITH (nolock)) AS tt WHERE ReasonID NOT IN ('99','100','101','102')`)).
		WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
			AddRow("27"))

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (SELECT Code, ReasonID, ReasonMessage FROM m_reason_message WITH (nolock)) AS tt WHERE ReasonID NOT IN ('99','100','101','102') ORDER BY tt.ReasonID asc OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
		WillReturnRows(sqlmock.NewRows([]string{"Code", "ReasonID", "ReasonMessage"}).
			AddRow("12", "11", "Akte Jual Beli Tidak Sesuai"))
	mock.ExpectCommit()

	// Call the function
	reason, _, err := repo.GetReasonPrescreening(req, 1)

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
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqReasonPrescreening{
		ReasonID: "99,100,101,102",
	}
	mock.ExpectBegin()

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (SELECT Code, ReasonID, ReasonMessage FROM m_reason_message WITH (nolock)) AS tt WHERE ReasonID NOT IN ('99','100','101','102') ORDER BY tt.ReasonID asc`)).
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectCommit()

	// Call the function
	_, _, err := repo.GetReasonPrescreening(req, nil)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

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

func TestGetInquiryPrescreening(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqInquiryPrescreening{
		Search:      "my name",
		BranchID:    "426",
		MultiBranch: "1",
		UserID:      "abc123",
	}

	expectedInquiry := []entity.InquiryPrescreening{{CmoRecommendation: 0, Activity: "", SourceDecision: "", Decision: "", Reason: "", DecisionBy: "", DecisionAt: "", ProspectID: "", BranchName: "", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	// Mock SQL query and result

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123' 
		)
		AND b.lob_id='125'`)).
		WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
			AddRow("WEST JAVA", `["426","436","429","431","442","428","430"]`))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','my name') AS encrypt`)).
		WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).
			AddRow("xxxxxx"))

	mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'LEGAL'
		) cal ON tm.ProspectID = cal.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'RESIDENCE'
		) car ON tm.ProspectID = car.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'COMPANY'
		) cac ON tm.ProspectID = cac.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'EMERGENCY'
		) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx')) AS tt`)).
		WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
			AddRow("27"))

	mock.ExpectQuery(regexp.QuoteMeta(`WITH 
	cte_app_config_mn AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'MonthName'
	),
	cte_app_config_pr AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'ProfessionID'
	)
	SELECT tt.* FROM (
	SELECT
	tm.ProspectID,
	cb.BranchName,
	cb.BranchID,
	tia.info AS CMORecommend,
	tst.activity,
	tst.source_decision,
	tps.decision,
	tps.reason,
	tps.created_by AS DecisionBy,
	tps.decision_by AS DecisionName,
	tps.created_at AS DecisionAt,
	CASE
	  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
	  ELSE 'NE'
	END AS incoming_source,
	tf.customer_status,
	tm.created_at,
	tm.order_at,
	scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
	scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
	scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
	tcp.BirthDate,
	scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
	CASE
	  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
	  WHEN tcp.Gender = 'F' THEN 'Perempuan'
	END AS 'Gender',
	scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
	CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
	cal.Kelurahan AS LegalKelurahan,
	cal.Kecamatan AS LegalKecamatan,
	cal.ZipCode AS LegalZipcode,
	cal.City AS LegalCity,
	scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
	scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
	edu.value AS Education,
	mst.value AS MaritalStatus,
	tcp.NumOfDependence,
	scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
	CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
	car.Kelurahan AS ResidenceKelurahan,
	car.Kecamatan AS ResidenceKecamatan,
	car.ZipCode AS ResidenceZipcode,
	car.City AS ResidenceCity,
	hst.value AS HomeStatus,
	mn.value AS StaySinceMonth,
	tcp.StaySinceYear,
	ta.ProductOfferingID,
	ta.dealer,
	ta.LifeInsuranceFee,
	ta.AssetInsuranceFee,
	'KMB MOTOR' AS AssetType,
	ti.asset_description,
	ti.manufacture_year,
	ti.color,
	chassis_number,
	engine_number,
	CASE
		WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
	END AS bpkb_name,
	ti.owner_asset,
	ti.license_plate,
	interest_rate,
	Tenor AS InstallmentPeriod,
	OTR,
	DPAmount,
	AF AS FinanceAmount,
	interest_amount,
	insurance_amount,
	AdminFee,
	provision_fee,
	NTF,
	NTFAkumulasi,
	(NTF + interest_amount) AS Total,
	InstallmentAmount AS MonthlyInstallment,
	FirstInstallment,
	pr.value AS ProfessionID,
	jt.value AS JobType,
	jb.value AS JobPosition,
	mn2.value AS EmploymentSinceMonth,
	tce.EmploymentSinceYear,
	tce.CompanyName,
	cac.AreaPhone AS CompanyAreaPhone,
	cac.Phone AS CompanyPhone,
	tcp.ExtCompanyPhone,
	scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
	CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
	cac.Kelurahan AS CompanyKelurahan,
	cac.Kecamatan AS CompanyKecamatan,
	car.ZipCode AS CompanyZipcode,
	car.City AS CompanyCity,
	tce.MonthlyFixedIncome,
	tce.MonthlyVariableIncome,
	tce.SpouseIncome,
	tcp.SourceOtherIncome,
	tcs.FullName AS SpouseLegalName,
	tcs.CompanyName AS SpouseCompanyName,
	tcs.CompanyPhone AS SpouseCompanyPhone,
	tcs.MobilePhone AS SpouseMobilePhone,
	tcs.IDNumber AS SpouseIDNumber,
	pr2.value AS SpouseProfession,
	em.Name AS EmconName,
	em.Relationship,
	em.MobilePhone AS EmconMobilePhone,
	scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
	CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
	cae.Kelurahan AS EmergencyKelurahan,
	cae.Kecamatan AS EmergencyKecamatan,
	cae.ZipCode AS EmergencyZipcode,
	cae.City AS EmergencyCity,
	cae.AreaPhone AS EmergencyAreaPhone,
	cae.Phone AS EmergencyPhone,
	tce.IndustryTypeID
  FROM
	trx_master tm WITH (nolock)
	INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
	INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
	INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
	INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
	INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
	INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
	INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
	INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'LEGAL'
	) cal ON tm.ProspectID = cal.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'RESIDENCE'
	) car ON tm.ProspectID = car.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'COMPANY'
	) cac ON tm.ProspectID = cac.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'EMERGENCY'
	) cae ON tm.ProspectID = cae.ProspectID
	INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
	LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
	LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'Education'
	) edu ON tcp.Education = edu.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'MaritalStatus'
	) mst ON tcp.MaritalStatus = mst.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'HomeStatus'
	) hst ON tcp.HomeStatus = hst.[key]
	LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
	LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobType'
	) jt ON tce.JobType = jt.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobPosition'
	) jb ON tce.JobPosition = jb.[key]
	LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
	LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key] WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
		WillReturnRows(sqlmock.NewRows([]string{"Code", "ReasonID", "ReasonMessage"}).
			AddRow("12", "11", "Akte Jual Beli Tidak Sesuai"))

	// Call the function
	reason, _, err := repo.GetInquiryPrescreening(req, 1)

	// Verify the result
	if err != nil {
		t.Fatalf("Expected no error, but got: %v", err)
	}
	assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestGetInquiryPrescreeningWithoutParam(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.InquiryPrescreening{{CmoRecommendation: 0, Activity: "", SourceDecision: "", Decision: "", Reason: "", DecisionBy: "", DecisionAt: "", ProspectID: "", BranchName: "", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	// Mock SQL query and result

	t.Run("without param branch", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryPrescreening{
			Search: "my name",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','my name') AS encrypt`)).
			WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).
				AddRow("xxxxxx"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'LEGAL'
		) cal ON tm.ProspectID = cal.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'RESIDENCE'
		) car ON tm.ProspectID = car.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'COMPANY'
		) cac ON tm.ProspectID = cac.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'EMERGENCY'
		) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		 WHERE (tcp.LegalName = 'xxxxxx')) AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
	cte_app_config_mn AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'MonthName'
	),
	cte_app_config_pr AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'ProfessionID'
	)
	SELECT tt.* FROM (
	SELECT
	tm.ProspectID,
	cb.BranchName,
	cb.BranchID,
	tia.info AS CMORecommend,
	tst.activity,
	tst.source_decision,
	tps.decision,
	tps.reason,
	tps.created_by AS DecisionBy,
	tps.decision_by AS DecisionName,
	tps.created_at AS DecisionAt,
	CASE
	  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
	  ELSE 'NE'
	END AS incoming_source,
	tf.customer_status,
	tm.created_at,
	tm.order_at,
	scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
	scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
	scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
	tcp.BirthDate,
	scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
	CASE
	  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
	  WHEN tcp.Gender = 'F' THEN 'Perempuan'
	END AS 'Gender',
	scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
	CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
	cal.Kelurahan AS LegalKelurahan,
	cal.Kecamatan AS LegalKecamatan,
	cal.ZipCode AS LegalZipcode,
	cal.City AS LegalCity,
	scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
	scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
	edu.value AS Education,
	mst.value AS MaritalStatus,
	tcp.NumOfDependence,
	scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
	CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
	car.Kelurahan AS ResidenceKelurahan,
	car.Kecamatan AS ResidenceKecamatan,
	car.ZipCode AS ResidenceZipcode,
	car.City AS ResidenceCity,
	hst.value AS HomeStatus,
	mn.value AS StaySinceMonth,
	tcp.StaySinceYear,
	ta.ProductOfferingID,
	ta.dealer,
	ta.LifeInsuranceFee,
	ta.AssetInsuranceFee,
	'KMB MOTOR' AS AssetType,
	ti.asset_description,
	ti.manufacture_year,
	ti.color,
	chassis_number,
	engine_number,
	CASE
		WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
	END AS bpkb_name,
	ti.owner_asset,
	ti.license_plate,
	interest_rate,
	Tenor AS InstallmentPeriod,
	OTR,
	DPAmount,
	AF AS FinanceAmount,
	interest_amount,
	insurance_amount,
	AdminFee,
	provision_fee,
	NTF,
	NTFAkumulasi,
	(NTF + interest_amount) AS Total,
	InstallmentAmount AS MonthlyInstallment,
	FirstInstallment,
	pr.value AS ProfessionID,
	jt.value AS JobType,
	jb.value AS JobPosition,
	mn2.value AS EmploymentSinceMonth,
	tce.EmploymentSinceYear,
	tce.CompanyName,
	cac.AreaPhone AS CompanyAreaPhone,
	cac.Phone AS CompanyPhone,
	tcp.ExtCompanyPhone,
	scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
	CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
	cac.Kelurahan AS CompanyKelurahan,
	cac.Kecamatan AS CompanyKecamatan,
	car.ZipCode AS CompanyZipcode,
	car.City AS CompanyCity,
	tce.MonthlyFixedIncome,
	tce.MonthlyVariableIncome,
	tce.SpouseIncome,
	tcp.SourceOtherIncome,
	tcs.FullName AS SpouseLegalName,
	tcs.CompanyName AS SpouseCompanyName,
	tcs.CompanyPhone AS SpouseCompanyPhone,
	tcs.MobilePhone AS SpouseMobilePhone,
	tcs.IDNumber AS SpouseIDNumber,
	pr2.value AS SpouseProfession,
	em.Name AS EmconName,
	em.Relationship,
	em.MobilePhone AS EmconMobilePhone,
	scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
	CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
	cae.Kelurahan AS EmergencyKelurahan,
	cae.Kecamatan AS EmergencyKecamatan,
	cae.ZipCode AS EmergencyZipcode,
	cae.City AS EmergencyCity,
	cae.AreaPhone AS EmergencyAreaPhone,
	cae.Phone AS EmergencyPhone,
	tce.IndustryTypeID
  FROM
	trx_master tm WITH (nolock)
	INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
	INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
	INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
	INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
	INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
	INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
	INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
	INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'LEGAL'
	) cal ON tm.ProspectID = cal.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'RESIDENCE'
	) car ON tm.ProspectID = car.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'COMPANY'
	) cac ON tm.ProspectID = cac.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'EMERGENCY'
	) cae ON tm.ProspectID = cae.ProspectID
	INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
	LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
	LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'Education'
	) edu ON tcp.Education = edu.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'MaritalStatus'
	) mst ON tcp.MaritalStatus = mst.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'HomeStatus'
	) hst ON tcp.HomeStatus = hst.[key]
	LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
	LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobType'
	) jt ON tce.JobType = jt.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobPosition'
	) jb ON tce.JobPosition = jb.[key]
	LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
	LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key] WHERE (tcp.LegalName = 'xxxxxx')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"Code", "ReasonID", "ReasonMessage"}).
				AddRow("12", "11", "Akte Jual Beli Tidak Sesuai"))

		// Call the function
		reason, _, err := repo.GetInquiryPrescreening(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("with region west java", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryPrescreening{
			Search:      "my name",
			BranchID:    "426",
			MultiBranch: "1",
			UserID:      "abc123",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN
		(	SELECT value
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123'
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("WEST JAVA", `["426","436","429","431","442","428","430"]`))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','my name') AS encrypt`)).
			WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).
				AddRow("xxxxxx"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'LEGAL'
		) cal ON tm.ProspectID = cal.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'RESIDENCE'
		) car ON tm.ProspectID = car.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'COMPANY'
		) cac ON tm.ProspectID = cac.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'EMERGENCY'
		) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		 WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx')) AS tt `)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
	cte_app_config_mn AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'MonthName'
	),
	cte_app_config_pr AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'ProfessionID'
	)
	SELECT tt.* FROM (
	SELECT
	tm.ProspectID,
	cb.BranchName,
	cb.BranchID,
	tia.info AS CMORecommend,
	tst.activity,
	tst.source_decision,
	tps.decision,
	tps.reason,
	tps.created_by AS DecisionBy,
	tps.decision_by AS DecisionName,
	tps.created_at AS DecisionAt,
	CASE
	  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
	  ELSE 'NE'
	END AS incoming_source,
	tf.customer_status,
	tm.created_at,
	tm.order_at,
	scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
	scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
	scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
	tcp.BirthDate,
	scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
	CASE
	  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
	  WHEN tcp.Gender = 'F' THEN 'Perempuan'
	END AS 'Gender',
	scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
	CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
	cal.Kelurahan AS LegalKelurahan,
	cal.Kecamatan AS LegalKecamatan,
	cal.ZipCode AS LegalZipcode,
	cal.City AS LegalCity,
	scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
	scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
	edu.value AS Education,
	mst.value AS MaritalStatus,
	tcp.NumOfDependence,
	scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
	CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
	car.Kelurahan AS ResidenceKelurahan,
	car.Kecamatan AS ResidenceKecamatan,
	car.ZipCode AS ResidenceZipcode,
	car.City AS ResidenceCity,
	hst.value AS HomeStatus,
	mn.value AS StaySinceMonth,
	tcp.StaySinceYear,
	ta.ProductOfferingID,
	ta.dealer,
	ta.LifeInsuranceFee,
	ta.AssetInsuranceFee,
	'KMB MOTOR' AS AssetType,
	ti.asset_description,
	ti.manufacture_year,
	ti.color,
	chassis_number,
	engine_number,
	CASE
		WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
	END AS bpkb_name,
	ti.owner_asset,
	ti.license_plate,
	interest_rate,
	Tenor AS InstallmentPeriod,
	OTR,
	DPAmount,
	AF AS FinanceAmount,
	interest_amount,
	insurance_amount,
	AdminFee,
	provision_fee,
	NTF,
	NTFAkumulasi,
	(NTF + interest_amount) AS Total,
	InstallmentAmount AS MonthlyInstallment,
	FirstInstallment,
	pr.value AS ProfessionID,
	jt.value AS JobType,
	jb.value AS JobPosition,
	mn2.value AS EmploymentSinceMonth,
	tce.EmploymentSinceYear,
	tce.CompanyName,
	cac.AreaPhone AS CompanyAreaPhone,
	cac.Phone AS CompanyPhone,
	tcp.ExtCompanyPhone,
	scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
	CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
	cac.Kelurahan AS CompanyKelurahan,
	cac.Kecamatan AS CompanyKecamatan,
	car.ZipCode AS CompanyZipcode,
	car.City AS CompanyCity,
	tce.MonthlyFixedIncome,
	tce.MonthlyVariableIncome,
	tce.SpouseIncome,
	tcp.SourceOtherIncome,
	tcs.FullName AS SpouseLegalName,
	tcs.CompanyName AS SpouseCompanyName,
	tcs.CompanyPhone AS SpouseCompanyPhone,
	tcs.MobilePhone AS SpouseMobilePhone,
	tcs.IDNumber AS SpouseIDNumber,
	pr2.value AS SpouseProfession,
	em.Name AS EmconName,
	em.Relationship,
	em.MobilePhone AS EmconMobilePhone,
	scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
	CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
	cae.Kelurahan AS EmergencyKelurahan,
	cae.Kecamatan AS EmergencyKecamatan,
	cae.ZipCode AS EmergencyZipcode,
	cae.City AS EmergencyCity,
	cae.AreaPhone AS EmergencyAreaPhone,
	cae.Phone AS EmergencyPhone,
	tce.IndustryTypeID
  FROM
	trx_master tm WITH (nolock)
	INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
	INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
	INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
	INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
	INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
	INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
	INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
	INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'LEGAL'
	) cal ON tm.ProspectID = cal.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'RESIDENCE'
	) car ON tm.ProspectID = car.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'COMPANY'
	) cac ON tm.ProspectID = cac.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'EMERGENCY'
	) cae ON tm.ProspectID = cae.ProspectID
	INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
	LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
	LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'Education'
	) edu ON tcp.Education = edu.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'MaritalStatus'
	) mst ON tcp.MaritalStatus = mst.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'HomeStatus'
	) hst ON tcp.HomeStatus = hst.[key]
	LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
	LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobType'
	) jt ON tce.JobType = jt.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobPosition'
	) jb ON tce.JobPosition = jb.[key]
	LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
	LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key] WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"Code", "ReasonID", "ReasonMessage"}).
				AddRow("12", "11", "Akte Jual Beli Tidak Sesuai"))

		// Call the function
		reason, _, err := repo.GetInquiryPrescreening(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("with region ALL", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryPrescreening{
			Search:      "my name",
			BranchID:    "426",
			MultiBranch: "1",
			UserID:      "abc123",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN
		(	SELECT value
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123'
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("ALL", `["426","436","429","431","442","428","430"]`))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','my name') AS encrypt`)).
			WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).
				AddRow("xxxxxx"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'LEGAL'
		) cal ON tm.ProspectID = cal.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'RESIDENCE'
		) car ON tm.ProspectID = car.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'COMPANY'
		) cac ON tm.ProspectID = cac.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'EMERGENCY'
		) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		 WHERE (tcp.LegalName = 'xxxxxx')) AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
	cte_app_config_mn AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'MonthName'
	),
	cte_app_config_pr AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'ProfessionID'
	)
	SELECT tt.* FROM (
	SELECT
	tm.ProspectID,
	cb.BranchName,
	cb.BranchID,
	tia.info AS CMORecommend,
	tst.activity,
	tst.source_decision,
	tps.decision,
	tps.reason,
	tps.created_by AS DecisionBy,
	tps.decision_by AS DecisionName,
	tps.created_at AS DecisionAt,
	CASE
	  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
	  ELSE 'NE'
	END AS incoming_source,
	tf.customer_status,
	tm.created_at,
	tm.order_at,
	scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
	scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
	scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
	tcp.BirthDate,
	scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
	CASE
	  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
	  WHEN tcp.Gender = 'F' THEN 'Perempuan'
	END AS 'Gender',
	scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
	CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
	cal.Kelurahan AS LegalKelurahan,
	cal.Kecamatan AS LegalKecamatan,
	cal.ZipCode AS LegalZipcode,
	cal.City AS LegalCity,
	scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
	scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
	edu.value AS Education,
	mst.value AS MaritalStatus,
	tcp.NumOfDependence,
	scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
	CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
	car.Kelurahan AS ResidenceKelurahan,
	car.Kecamatan AS ResidenceKecamatan,
	car.ZipCode AS ResidenceZipcode,
	car.City AS ResidenceCity,
	hst.value AS HomeStatus,
	mn.value AS StaySinceMonth,
	tcp.StaySinceYear,
	ta.ProductOfferingID,
	ta.dealer,
	ta.LifeInsuranceFee,
	ta.AssetInsuranceFee,
	'KMB MOTOR' AS AssetType,
	ti.asset_description,
	ti.manufacture_year,
	ti.color,
	chassis_number,
	engine_number,
	CASE
		WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
	END AS bpkb_name,
	ti.owner_asset,
	ti.license_plate,
	interest_rate,
	Tenor AS InstallmentPeriod,
	OTR,
	DPAmount,
	AF AS FinanceAmount,
	interest_amount,
	insurance_amount,
	AdminFee,
	provision_fee,
	NTF,
	NTFAkumulasi,
	(NTF + interest_amount) AS Total,
	InstallmentAmount AS MonthlyInstallment,
	FirstInstallment,
	pr.value AS ProfessionID,
	jt.value AS JobType,
	jb.value AS JobPosition,
	mn2.value AS EmploymentSinceMonth,
	tce.EmploymentSinceYear,
	tce.CompanyName,
	cac.AreaPhone AS CompanyAreaPhone,
	cac.Phone AS CompanyPhone,
	tcp.ExtCompanyPhone,
	scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
	CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
	cac.Kelurahan AS CompanyKelurahan,
	cac.Kecamatan AS CompanyKecamatan,
	car.ZipCode AS CompanyZipcode,
	car.City AS CompanyCity,
	tce.MonthlyFixedIncome,
	tce.MonthlyVariableIncome,
	tce.SpouseIncome,
	tcp.SourceOtherIncome,
	tcs.FullName AS SpouseLegalName,
	tcs.CompanyName AS SpouseCompanyName,
	tcs.CompanyPhone AS SpouseCompanyPhone,
	tcs.MobilePhone AS SpouseMobilePhone,
	tcs.IDNumber AS SpouseIDNumber,
	pr2.value AS SpouseProfession,
	em.Name AS EmconName,
	em.Relationship,
	em.MobilePhone AS EmconMobilePhone,
	scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
	CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
	cae.Kelurahan AS EmergencyKelurahan,
	cae.Kecamatan AS EmergencyKecamatan,
	cae.ZipCode AS EmergencyZipcode,
	cae.City AS EmergencyCity,
	cae.AreaPhone AS EmergencyAreaPhone,
	cae.Phone AS EmergencyPhone,
	tce.IndustryTypeID
  FROM
	trx_master tm WITH (nolock)
	INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
	INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
	INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
	INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
	INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
	INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
	INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
	INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'LEGAL'
	) cal ON tm.ProspectID = cal.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'RESIDENCE'
	) car ON tm.ProspectID = car.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'COMPANY'
	) cac ON tm.ProspectID = cac.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'EMERGENCY'
	) cae ON tm.ProspectID = cae.ProspectID
	INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
	LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
	LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'Education'
	) edu ON tcp.Education = edu.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'MaritalStatus'
	) mst ON tcp.MaritalStatus = mst.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'HomeStatus'
	) hst ON tcp.HomeStatus = hst.[key]
	LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
	LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobType'
	) jt ON tce.JobType = jt.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobPosition'
	) jb ON tce.JobPosition = jb.[key]
	LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
	LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key] WHERE (tcp.LegalName = 'xxxxxx')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"Code", "ReasonID", "ReasonMessage"}).
				AddRow("12", "11", "Akte Jual Beli Tidak Sesuai"))

		// Call the function
		reason, _, err := repo.GetInquiryPrescreening(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetInquiryPrescreeningRecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqInquiryPrescreening{}

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(regexp.QuoteMeta(`WITH
        cte_app_config_mn AS (
                SELECT
                        [key],
                        value
                FROM
                        app_config ap WITH (nolock)
                WHERE
                        group_name = 'MonthName'
        ),
        cte_app_config_pr AS (
                SELECT
                        [key],
                        value
                FROM
                        app_config ap WITH (nolock)
                WHERE
                        group_name = 'ProfessionID'
        )
        SELECT tt.* FROM (
        SELECT
        tm.ProspectID,
        cb.BranchName,
        cb.BranchID,
        tia.info AS CMORecommend,
        tst.activity,
        tst.source_decision,
        tps.decision,
        tps.reason,
        tps.created_by AS DecisionBy,
        tps.decision_by AS DecisionName,
        tps.created_at AS DecisionAt,
        CASE
          WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
          ELSE 'NE'
        END AS incoming_source,
        tf.customer_status,
        tm.created_at,
        tm.order_at,
        scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
        scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
        scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
        tcp.BirthDate,
        scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
        CASE
          WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
          WHEN tcp.Gender = 'F' THEN 'Perempuan'
        END AS 'Gender',
        scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
        CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
        cal.Kelurahan AS LegalKelurahan,
        cal.Kecamatan AS LegalKecamatan,
        cal.ZipCode AS LegalZipcode,
        cal.City AS LegalCity,
        scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
        scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
        edu.value AS Education,
        mst.value AS MaritalStatus,
        tcp.NumOfDependence,
        scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
        CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
        car.Kelurahan AS ResidenceKelurahan,
        car.Kecamatan AS ResidenceKecamatan,
        car.ZipCode AS ResidenceZipcode,
        car.City AS ResidenceCity,
        hst.value AS HomeStatus,
        mn.value AS StaySinceMonth,
        tcp.StaySinceYear,
        ta.ProductOfferingID,
        ta.dealer,
        ta.LifeInsuranceFee,
        ta.AssetInsuranceFee,
        'KMB MOTOR' AS AssetType,
        ti.asset_description,
        ti.manufacture_year,
        ti.color,
        chassis_number,
        engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
        interest_rate,
        Tenor AS InstallmentPeriod,
        OTR,
        DPAmount,
        AF AS FinanceAmount,
        interest_amount,
        insurance_amount,
        AdminFee,
        provision_fee,
        NTF,
        NTFAkumulasi,
        (NTF + interest_amount) AS Total,
        InstallmentAmount AS MonthlyInstallment,
        FirstInstallment,
        pr.value AS ProfessionID,
        jt.value AS JobType,
        jb.value AS JobPosition,
        mn2.value AS EmploymentSinceMonth,
        tce.EmploymentSinceYear,
        tce.CompanyName,
        cac.AreaPhone AS CompanyAreaPhone,
        cac.Phone AS CompanyPhone,
        tcp.ExtCompanyPhone,
        scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
        CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
        cac.Kelurahan AS CompanyKelurahan,
        cac.Kecamatan AS CompanyKecamatan,
        car.ZipCode AS CompanyZipcode,
        car.City AS CompanyCity,
        tce.MonthlyFixedIncome,
        tce.MonthlyVariableIncome,
        tce.SpouseIncome,
        tcp.SourceOtherIncome,
        tcs.FullName AS SpouseLegalName,
        tcs.CompanyName AS SpouseCompanyName,
        tcs.CompanyPhone AS SpouseCompanyPhone,
        tcs.MobilePhone AS SpouseMobilePhone,
        tcs.IDNumber AS SpouseIDNumber,
        pr2.value AS SpouseProfession,
        em.Name AS EmconName,
        em.Relationship,
        em.MobilePhone AS EmconMobilePhone,
        scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
        CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
        cae.Kelurahan AS EmergencyKelurahan,
        cae.Kecamatan AS EmergencyKecamatan,
        cae.ZipCode AS EmergencyZipcode,
        cae.City AS EmergencyCity,
        cae.AreaPhone AS EmergencyAreaPhone,
        cae.Phone AS EmergencyPhone,
        tce.IndustryTypeID
  FROM
        trx_master tm WITH (nolock)
        INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
        INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
        INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
        INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
        INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
        INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
        INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
        INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
        INNER JOIN (
          SELECT
                ProspectID,
                Address,
                RT,
                RW,
                Kelurahan,
                Kecamatan,
                ZipCode,
                City
          FROM
                trx_customer_address WITH (nolock)
          WHERE
                "Type" = 'LEGAL'
        ) cal ON tm.ProspectID = cal.ProspectID
        INNER JOIN (
          SELECT
                ProspectID,
                Address,
                RT,
                RW,
                Kelurahan,
                Kecamatan,
                ZipCode,
                City
          FROM
                trx_customer_address WITH (nolock)
          WHERE
                "Type" = 'RESIDENCE'
        ) car ON tm.ProspectID = car.ProspectID
        INNER JOIN (
          SELECT
                ProspectID,
                Address,
                RT,
                RW,
                Kelurahan,
                Kecamatan,
                ZipCode,
                City,
                Phone,
                AreaPhone
          FROM
                trx_customer_address WITH (nolock)
          WHERE
                "Type" = 'COMPANY'
        ) cac ON tm.ProspectID = cac.ProspectID
        INNER JOIN (
          SELECT
                ProspectID,
                Address,
                RT,
                RW,
                Kelurahan,
                Kecamatan,
                ZipCode,
                City,
                Phone,
                AreaPhone
          FROM
                trx_customer_address WITH (nolock)
          WHERE
                "Type" = 'EMERGENCY'
        ) cae ON tm.ProspectID = cae.ProspectID
        INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
        LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
        LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
        LEFT JOIN (
          SELECT
                [key],
                value
          FROM
                app_config ap WITH (nolock)
          WHERE
                group_name = 'Education'
        ) edu ON tcp.Education = edu.[key]
        LEFT JOIN (
          SELECT
                [key],
                value
          FROM
                app_config ap WITH (nolock)
          WHERE
                group_name = 'MaritalStatus'
        ) mst ON tcp.MaritalStatus = mst.[key]
        LEFT JOIN (
          SELECT
                [key],
                value
          FROM
                app_config ap WITH (nolock)
          WHERE
                group_name = 'HomeStatus'
        ) hst ON tcp.HomeStatus = hst.[key]
        LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
        LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
        LEFT JOIN (
          SELECT
                [key],
                value
          FROM
                app_config ap WITH (nolock)
          WHERE
                group_name = 'JobType'
        ) jt ON tce.JobType = jt.[key]
        LEFT JOIN (
          SELECT
                [key],
                value
          FROM
                app_config ap WITH (nolock)
          WHERE
                group_name = 'JobPosition'
        ) jb ON tce.JobPosition = jb.[key]
        LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
        LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key] WHERE CAST(tm.created_at AS date) >= DATEADD(day, , CAST(GETDATE() AS date))) AS tt ORDER BY tt.created_at DESC`)).
		WillReturnError(gorm.ErrRecordNotFound)

	// Call the function
	_, _, err := repo.GetInquiryPrescreening(req, nil)

	// Verify the error message
	expectedErr := errors.New(constant.RECORD_NOT_FOUND)
	assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expectations: %s", err)
	}
}

func TestSavePrescreening(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)
	data := response.ReviewPrescreening{
		ProspectID: "TST001",
		Code:       constant.CODE_REJECT_PRESCREENING,
		Decision:   constant.DB_DECISION_REJECT,
		Reason:     "reject reason",
	}

	info, _ := json.Marshal(data)

	trxPrescreening := entity.TrxPrescreening{
		ProspectID: "TST001",
		Decision:   constant.DB_DECISION_REJECT,
		Reason:     "reject reason",
		CreatedBy:  "SYSTEM",
		DecisionBy: "5XeZs9PCeiPcZGS6azt",
	}

	trxDetail := entity.TrxDetail{
		ProspectID:     "TST001",
		RuleCode:       constant.CODE_REJECT_PRESCREENING,
		StatusProcess:  constant.STATUS_FINAL,
		Activity:       constant.ACTIVITY_STOP,
		Decision:       constant.DB_DECISION_REJECT,
		SourceDecision: constant.PRESCREENING,
		Info:           string(info),
		CreatedBy:      "SYSTEM",
	}
	detail := structToSlice(trxDetail)

	trxStatus := entity.TrxStatus{
		ProspectID:     "TST001",
		StatusProcess:  constant.STATUS_FINAL,
		Activity:       constant.ACTIVITY_STOP,
		Decision:       constant.DB_DECISION_REJECT,
		SourceDecision: constant.PRESCREENING,
		RuleCode:       constant.CODE_REJECT_PRESCREENING,
		Reason:         "reject reason",
	}

	t.Run("success update", func(t *testing.T) {

		mock.ExpectBegin()
		query := `UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`
		queryRegex := regexp.QuoteMeta(query)
		mock.ExpectExec(queryRegex).WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "trx_details" (.*)`).
			WithArgs(detail...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_akkk" ("ProspectID","ScsDate","ScsScore","ScsStatus","CustomerType","SpouseType","AgreementStatus","TotalAgreementAktif","MaxOVDAgreementAktif","LastMaxOVDAgreement","DSRFMF","DSRPBK","TotalDSR","EkycSource","EkycSimiliarity","EkycReason","NumberOfPaidInstallment","OSInstallmentDue","InstallmentAmountFMF","InstallmentAmountSpouseFMF","InstallmentAmountOther","InstallmentAmountOtherSpouse","InstallmentTopup","LatestInstallment","UrlFormAkkk","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxStatus.ProspectID, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_prescreening" ("ProspectID","decision","reason","created_at","created_by","decision_by") VALUES (?,?,?,?,?,?)`)).
			WithArgs(trxPrescreening.ProspectID, trxPrescreening.Decision, trxPrescreening.Reason, sqlmock.AnyArg(), trxPrescreening.CreatedBy, trxPrescreening.DecisionBy).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.SavePrescreening(trxPrescreening, trxDetail, trxStatus)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("success insert", func(t *testing.T) {

		mock.ExpectBegin()
		query := `UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`
		queryRegex := regexp.QuoteMeta(query)
		mock.ExpectExec(queryRegex).WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_status" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","created_at","reason") VALUES (?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxStatus.ProspectID, trxStatus.StatusProcess, trxStatus.Activity, trxStatus.Decision, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.NextStep, sqlmock.AnyArg(), trxStatus.Reason).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "trx_details" (.*)`).
			WithArgs(detail...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_akkk" ("ProspectID","ScsDate","ScsScore","ScsStatus","CustomerType","SpouseType","AgreementStatus","TotalAgreementAktif","MaxOVDAgreementAktif","LastMaxOVDAgreement","DSRFMF","DSRPBK","TotalDSR","EkycSource","EkycSimiliarity","EkycReason","NumberOfPaidInstallment","OSInstallmentDue","InstallmentAmountFMF","InstallmentAmountSpouseFMF","InstallmentAmountOther","InstallmentAmountOtherSpouse","InstallmentTopup","LatestInstallment","UrlFormAkkk","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxStatus.ProspectID, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_prescreening" ("ProspectID","decision","reason","created_at","created_by","decision_by") VALUES (?,?,?,?,?,?)`)).
			WithArgs(trxPrescreening.ProspectID, trxPrescreening.Decision, trxPrescreening.Reason, sqlmock.AnyArg(), trxPrescreening.CreatedBy, trxPrescreening.DecisionBy).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.SavePrescreening(trxPrescreening, trxDetail, trxStatus)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})
}

func TestSaveLogOrchestrator(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

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

func TestGetHistoryApproval(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedData := []entity.HistoryApproval{
		{
			Decision:              "APR",
			Note:                  "Ok dari CA",
			CreatedAt:             time.Time{},
			DecisionBy:            "User CA KMB",
			NeedEscalation:        "No",
			NextFinalApprovalFlag: 1,
			SourceDecision:        "CRA",
			NextStep:              "CBM",
			SlikResult:            "Lancar",
		},
	}

	t.Run("success", func(t *testing.T) {

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT thas.decision_by, thas.next_final_approval_flag, CASE WHEN thas.decision = 'APR' THEN 'Approve' WHEN thas.decision = 'REJ' THEN 'Reject' WHEN thas.decision = 'CAN' THEN 'Cancel' WHEN thas.decision = 'RTN' THEN 'Return' WHEN thas.decision = 'SDP' THEN 'Submit Perubahan Data Pembiayaan' ELSE '-' END AS decision, CASE WHEN thas.need_escalation = 1 THEN 'Yes' ELSE 'No' END AS need_escalation, thas.source_decision, CASE WHEN thas.next_step<>'' THEN thas.next_step ELSE '-' END AS next_step, CASE WHEN thas.note<>'' THEN thas.note ELSE '-' END AS note, thas.created_at, CASE WHEN thas.source_decision = 'CRA' AND tcd.slik_result<>'' AND thas.decision<>'SDP' THEN tcd.slik_result ELSE '-' END AS slik_result FROM trx_history_approval_scheme thas WITH \(nolock\) LEFT JOIN trx_ca_decision tcd on thas.ProspectID = tcd.ProspectID WHERE thas.ProspectID = \? ORDER BY thas.created_at DESC`).WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"decision", "decision_by", "next_final_approval_flag", "need_escalation", "source_decision", "next_step", "note", "created_at", "slik_result"}).
				AddRow("APR", "User CA KMB", 1, "No", "CRA", "CBM", "Ok dari CA", time.Time{}, "Lancar"))
		mock.ExpectCommit()

		// Call the function
		data, err := repo.GetHistoryApproval(prospectID)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {

		// Mock SQL query to simulate record not found
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT thas.decision_by, thas.next_final_approval_flag, CASE WHEN thas.decision = 'APR' THEN 'Approve' WHEN thas.decision = 'REJ' THEN 'Reject' WHEN thas.decision = 'CAN' THEN 'Cancel' WHEN thas.decision = 'RTN' THEN 'Return' WHEN thas.decision = 'SDP' THEN 'Submit Perubahan Data Pembiayaan' ELSE '-' END AS decision, CASE WHEN thas.need_escalation = 1 THEN 'Yes' ELSE 'No' END AS need_escalation, thas.source_decision, CASE WHEN thas.next_step<>'' THEN thas.next_step ELSE '-' END AS next_step, CASE WHEN thas.note<>'' THEN thas.note ELSE '-' END AS note, thas.created_at, CASE WHEN thas.source_decision = 'CRA' AND tcd.slik_result<>'' AND thas.decision<>'SDP' THEN tcd.slik_result ELSE '-' END AS slik_result FROM trx_history_approval_scheme thas WITH \(nolock\) LEFT JOIN trx_ca_decision tcd on thas.ProspectID = tcd.ProspectID WHERE thas.ProspectID = \? ORDER BY thas.created_at DESC`).WithArgs(prospectID).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		// Call the function
		_, err := repo.GetHistoryApproval(prospectID)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetInternalRecord(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedData := []entity.TrxInternalRecord{
		{
			ApplicationID:        "426A202201124155",
			ProductType:          "KMB",
			AgreementDate:        time.Time{},
			AssetCode:            "K-YMH.MOTOR.NMAX (B6H A/T)",
			Tenor:                26,
			OutstandingPrincipal: 0,
			InstallmentAmount:    1866000,
			ContractStatus:       "LIV",
			CurrentCondition:     "OVD 204 hari",
		},
	}

	t.Run("success", func(t *testing.T) {

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM trx_internal_record WITH (nolock) WHERE ProspectID = ? ORDER BY created_at DESC`)).WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"ApplicationID", "ProductType", "AgreementDate", "AssetCode", "Tenor", "OutstandingPrincipal", "InstallmentAmount", "ContractStatus", "CurrentCondition"}).
				AddRow("426A202201124155", "KMB", time.Time{}, "K-YMH.MOTOR.NMAX (B6H A/T)", 26, 0, 1866000, "LIV", "OVD 204 hari"))
		mock.ExpectCommit()

		// Call the function
		data, err := repo.GetInternalRecord(prospectID)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {

		// Mock SQL query to simulate record not found
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM trx_internal_record WITH (nolock) WHERE ProspectID = ? ORDER BY created_at DESC`)).WithArgs(prospectID).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		// Call the function
		_, err := repo.GetInternalRecord(prospectID)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetLimitApproval(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	ntf := 10000.65
	expectedData := entity.MappingLimitApprovalScheme{
		Alias: "CBM",
	}

	t.Run("success", func(t *testing.T) {

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT [alias] FROM m_limit_approval_scheme WITH (nolock) WHERE ? between coverage_ntf_start AND coverage_ntf_end`)).WithArgs(ntf).
			WillReturnRows(sqlmock.NewRows([]string{"alias"}).
				AddRow("CBM"))
		mock.ExpectCommit()

		// Call the function
		data, err := repo.GetLimitApproval(ntf)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {

		// Mock SQL query to simulate record not found
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT [alias] FROM m_limit_approval_scheme WITH (nolock) WHERE ? between coverage_ntf_start AND coverage_ntf_end`)).WithArgs(ntf).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		// Call the function
		_, err := repo.GetLimitApproval(ntf)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetLimitApprovalDeviasi(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "SAL-000001"
	expectedData := entity.MappingLimitApprovalScheme{
		Alias: "CBM",
	}

	t.Run("success", func(t *testing.T) {

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		CASE 
			WHEN final_approval IS NULL THEN 'CBM'
			ELSE final_approval
			END AS alias
		FROM trx_deviasi td
		LEFT JOIN trx_master tm ON td.ProspectID = tm.ProspectID 
		LEFT JOIN m_branch_deviasi mbd ON tm.BranchID = mbd.BranchID 
		WHERE td.ProspectID = ?`)).WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"alias"}).
				AddRow("CBM"))
		mock.ExpectCommit()

		// Call the function
		data, err := repo.GetLimitApprovalDeviasi(prospectID)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

}

func TestGetHistoryProcess(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedData := []entity.HistoryProcess{
		{
			Decision:       "PASS",
			SourceDecision: "PRE SCREENING",
			Reason:         "Dokumen Sesuai",
			CreatedAt:      "",
		},
	}

	t.Run("success", func(t *testing.T) {

		mock.ExpectBegin()

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		CASE
			WHEN td.source_decision = 'PSI' THEN 'PRE SCREENING'
			WHEN td.source_decision IN ('TNR','PRJ','NIK','NKA','BLK','PMK') THEN 'DUPLICATION CHECKING'
			WHEN td.source_decision = 'DCK' THEN 'DUPLICATION CHECKING'
			WHEN td.source_decision = 'DCP'
			OR td.source_decision = 'ARI'
			OR td.source_decision = 'KTP' THEN 'EKYC'
			WHEN td.source_decision = 'PBK' THEN 'PEFINDO'
			WHEN td.source_decision = 'SCP' THEN 'SCOREPRO'
			WHEN td.source_decision = 'DSR' THEN 'DSR'
			WHEN td.source_decision = 'LTV' THEN 'LTV'
			WHEN td.source_decision = 'DEV' THEN 'DEVIASI'
			WHEN td.source_decision = 'CRA' THEN 'CREDIT ANALYSIS'
			WHEN td.source_decision = 'NRC' THEN 'RECALCULATE PROCESS'
			WHEN td.source_decision = 'CBM'
			OR td.source_decision = 'DRM'
			OR td.source_decision = 'GMO'
			OR td.source_decision = 'COM'
			OR td.source_decision = 'GMC'
			OR td.source_decision = 'UCC' THEN 'CREDIT COMMITEE'
			ELSE '-'
			END AS source_decision,
			CASE
			WHEN td.source_decision = 'CRA' THEN 'CA'
			WHEN td.source_decision = 'CBM' THEN 'BM'
			WHEN td.source_decision = 'DRM' THEN 'RM'
			WHEN td.source_decision = 'GMO' THEN 'GMO'
			WHEN td.source_decision = 'COM' THEN 'COM'
			WHEN td.source_decision = 'GMC' THEN 'GMC'
			WHEN td.source_decision = 'UCC' THEN 'UCC'
			ELSE td.source_decision
			END AS alias,
			CASE
			WHEN td.source_decision = 'DEV' THEN '-'
			WHEN td.decision = 'PAS' THEN 'PASS'
			WHEN td.decision = 'REJ' THEN 'REJECT'
			WHEN td.decision = 'CAN' THEN 'CANCEL'
			WHEN td.decision = 'RTN' THEN 'RETURN'
			WHEN td.decision = 'CPR' THEN 'CREDIT PROCESS'
			ELSE '-'
			END AS decision,
			CASE
			WHEN ap.reason IS NULL THEN td.reason 
			ELSE ap.reason 
			END AS reason,
			FORMAT(td.created_at,'yyyy-MM-dd HH:mm:ss') as created_at,
			td.next_step
		FROM
			trx_details td WITH (nolock)
			LEFT JOIN app_rules ap ON ap.rule_code = td.rule_code
		WHERE td.ProspectID = ? AND (td.source_decision IN('PSI','DCK','DCP','ARI','KTP','PBK','SCP','DSR','CRA','CBM','DRM','GMO','COM','GMC','UCC','NRC','DEV') OR 
		(td.source_decision IN('TNR','PRJ','NIK','NKA','BLK','PMK','LTV') AND td.decision = 'REJ'))
		AND td.decision <> 'CTG' AND td.activity <> 'UNPR' ORDER BY td.created_at ASC`)).WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"source_decision", "decision", "reason", "created_at"}).
				AddRow("PRE SCREENING", "PASS", "Dokumen Sesuai", ""))
		mock.ExpectCommit()

		// Call the function
		data, err := repo.GetHistoryProcess(prospectID)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {

		mock.ExpectBegin()

		// Mock SQL query to simulate record not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		CASE
			WHEN td.source_decision = 'PSI' THEN 'PRE SCREENING'
			WHEN td.source_decision IN ('TNR','PRJ','NIK','NKA','BLK','PMK') THEN 'DUPLICATION CHECKING'
			WHEN td.source_decision = 'DCK' THEN 'DUPLICATION CHECKING'
			WHEN td.source_decision = 'DCP'
			OR td.source_decision = 'ARI'
			OR td.source_decision = 'KTP' THEN 'EKYC'
			WHEN td.source_decision = 'PBK' THEN 'PEFINDO'
			WHEN td.source_decision = 'SCP' THEN 'SCOREPRO'
			WHEN td.source_decision = 'DSR' THEN 'DSR'
			WHEN td.source_decision = 'LTV' THEN 'LTV'
			WHEN td.source_decision = 'DEV' THEN 'DEVIASI'
			WHEN td.source_decision = 'CRA' THEN 'CREDIT ANALYSIS'
			WHEN td.source_decision = 'NRC' THEN 'RECALCULATE PROCESS'
			WHEN td.source_decision = 'CBM'
			OR td.source_decision = 'DRM'
			OR td.source_decision = 'GMO'
			OR td.source_decision = 'COM'
			OR td.source_decision = 'GMC'
			OR td.source_decision = 'UCC' THEN 'CREDIT COMMITEE'
			ELSE '-'
			END AS source_decision,
			CASE
			WHEN td.source_decision = 'CRA' THEN 'CA'
			WHEN td.source_decision = 'CBM' THEN 'BM'
			WHEN td.source_decision = 'DRM' THEN 'RM'
			WHEN td.source_decision = 'GMO' THEN 'GMO'
			WHEN td.source_decision = 'COM' THEN 'COM'
			WHEN td.source_decision = 'GMC' THEN 'GMC'
			WHEN td.source_decision = 'UCC' THEN 'UCC'
			ELSE td.source_decision
			END AS alias,
			CASE
			WHEN td.source_decision = 'DEV' THEN '-'
			WHEN td.decision = 'PAS' THEN 'PASS'
			WHEN td.decision = 'REJ' THEN 'REJECT'
			WHEN td.decision = 'CAN' THEN 'CANCEL'
			WHEN td.decision = 'RTN' THEN 'RETURN'
			WHEN td.decision = 'CPR' THEN 'CREDIT PROCESS'
			ELSE '-'
			END AS decision,
			CASE
			WHEN ap.reason IS NULL THEN td.reason 
			ELSE ap.reason 
			END AS reason,
			FORMAT(td.created_at,'yyyy-MM-dd HH:mm:ss') as created_at,
			td.next_step
		FROM
			trx_details td WITH (nolock)
			LEFT JOIN app_rules ap ON ap.rule_code = td.rule_code
		WHERE td.ProspectID = ? AND (td.source_decision IN('PSI','DCK','DCP','ARI','KTP','PBK','SCP','DSR','CRA','CBM','DRM','GMO','COM','GMC','UCC','NRC','DEV') OR 
		(td.source_decision IN('TNR','PRJ','NIK','NKA','BLK','PMK','LTV') AND td.decision = 'REJ'))
		AND td.decision <> 'CTG' AND td.activity <> 'UNPR' ORDER BY td.created_at ASC`)).WithArgs(prospectID).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		// Call the function
		_, err := repo.GetHistoryProcess(prospectID)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetCancelReason(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedReason := []entity.CancelReason{
		{
			ReasonID: "1",
			Show:     "1",
			Reason:   "Ganti Program Marketing",
		},
	}

	t.Run("success", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.id_cancel_reason) AS totalRow
		FROM
		(SELECT * FROM m_cancel_reason with (nolock) WHERE show = '1') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("8"))

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM m_cancel_reason with (nolock) WHERE show = '1' ORDER BY id_cancel_reason ASC`)).
			WillReturnRows(sqlmock.NewRows([]string{"id_cancel_reason", "reason", "show"}).
				AddRow("1", "Ganti Program Marketing", "1"))
		mock.ExpectCommit()

		// Call the function
		reason, _, err := repo.GetCancelReason(1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedReason, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success not found", func(t *testing.T) {
		// Mock SQL query to simulate record not found
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM m_cancel_reason with (nolock) WHERE show = '1' ORDER BY id_cancel_reason ASC`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetCancelReason(nil)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

}

func TestGetInquiryCa(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.InquiryCa{{ShowAction: false, ActionDate: "", Activity: "", SourceDecision: "", StatusDecision: "", StatusReason: "", CaDecision: "", CANote: "", ScsDate: "", ScsScore: "", ScsStatus: "", BiroCustomerResult: "", BiroSpouseResult: "", DraftDecision: "", DraftSlikResult: "", DraftNote: "", DraftCreatedAt: time.Time{}, DraftCreatedBy: "", DraftDecisionBy: "", ProspectID: "EFM03406412522151347", BranchName: "BANDUNG", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerID: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), SurveyResult: "", Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	t.Run("success with multi branch and need decision", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryCa{
			Search:      "aprospectid",
			BranchID:    "426",
			MultiBranch: "1",
			Filter:      "NEED_DECISION",
			UserID:      "abc123",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock) INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN ( SELECT value FROM region_user ru WITH (nolock) cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',') WHERE ru.user_id = 'abc123' ) AND b.lob_id='125'`)).WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','aprospectid') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
		 WHERE tm.BranchID = '426' AND (tcp.LegalName = 'xxxxxx') AND tst.activity= 'UNPR' AND tst.decision= 'CPR' AND tst.source_decision = 'CRA' AND (tcd.decision IS NULL OR (rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN')) AND tst.source_decision<>'PSI') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by,
				x.pernyataan_1,
				x.pernyataan_2,
				x.pernyataan_3,
				x.pernyataan_4,
				x.pernyataan_5,
				x.pernyataan_6
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_detail_biro AS (
			SELECT
				prospect_id, url_pdf_report AS BiroCustomerResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'CUSTOMER'
		),
		cte_trx_detail_biro2 AS (
			SELECT
				prospect_id, url_pdf_report AS BiroSpouseResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'SPOUSE'
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		),
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as decision_ca,
		tcd.created_by as decision_by_ca,
		tr.additional_dp,
		CASE
		  WHEN tcd.decision='APR' THEN 'APPROVE'
		  WHEN tcd.decision='REJ' THEN 'REJECT'
		  WHEN tcd.decision='CAN' THEN 'CANCEL'
		  ELSE tcd.decision
		END AS ca_decision,
		tcd.note AS ca_note,
		CASE
		  WHEN tcd.created_at IS NOT NULL
		  AND tfa.created_at IS NULL THEN FORMAT(tcd.created_at,'yyyy-MM-dd HH:mm:ss')
		  WHEN tfa.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE NULL
		END AS ActionDate,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		
		tdd.decision AS draft_decision,
		tdd.slik_result AS draft_slik_result,
		tdd.note AS draft_note,
		tdd.created_at AS draft_created_at,
		tdd.created_by AS draft_created_by,
		tdd.decision_by AS draft_decision_by,
		tdd.pernyataan_1 AS draft_pernyataan_1,
		tdd.pernyataan_2 AS draft_pernyataan_2,
		tdd.pernyataan_3 AS draft_pernyataan_3,
		tdd.pernyataan_4 AS draft_pernyataan_4,
		tdd.pernyataan_5 AS draft_pernyataan_5,
		tdd.pernyataan_6 AS draft_pernyataan_6,

		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		ta.interest_rate,
		ta.Tenor AS InstallmentPeriod,
		OTR,
		ta.DPAmount,
		ta.AF AS FinanceAmount,
		ta.interest_amount,
		ta.insurance_amount,
		ta.AdminFee,
		ta.provision_fee,
		ta.NTF,
		ta.NTFAkumulasi,
		(ta.NTF + ta.interest_amount) AS Total,
		ta.InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb2.BiroSpouseResult,
		CASE
		 WHEN rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN' THEN 1
		 ELSE 0
		END AS ActionEditData,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN
			cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN
			cte_trx_detail_biro tdb ON tm.ProspectID = tdb.prospect_id
		LEFT JOIN
			cte_trx_detail_biro2 tdb2 ON tm.ProspectID = tdb2.prospect_id
		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		LEFT JOIN
			cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
	 WHERE tm.BranchID = '426' AND (tcp.LegalName = 'xxxxxx') AND tst.activity= 'UNPR' AND tst.decision= 'CPR' AND tst.source_decision = 'CRA' AND (tcd.decision IS NULL OR (rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN')) AND tst.source_decision<>'PSI') AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryCa(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with multi branch and saved as draft", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryCa{
			Search:      "SAL-XXX",
			BranchID:    "426",
			MultiBranch: "1",
			Filter:      "SAVED_AS_DRAFT",
			UserID:      "abc123",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock) INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN ( SELECT value FROM region_user ru WITH (nolock) cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',') WHERE ru.user_id = 'abc123' ) AND b.lob_id='125'`)).WillReturnError(gorm.ErrRecordNotFound)

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
		 WHERE tm.BranchID = '426' AND (tm.ProspectID = 'SAL-XXX') AND tdd.draft_created_by= 'abc123'  AND tst.source_decision<>'PSI') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by,
				x.pernyataan_1,
				x.pernyataan_2,
				x.pernyataan_3,
				x.pernyataan_4,
				x.pernyataan_5,
				x.pernyataan_6
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_detail_biro AS (
			SELECT
				prospect_id, url_pdf_report AS BiroCustomerResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'CUSTOMER'
		),
		cte_trx_detail_biro2 AS (
			SELECT
				prospect_id, url_pdf_report AS BiroSpouseResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'SPOUSE'
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		),
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as decision_ca,
		tcd.created_by as decision_by_ca,
		tr.additional_dp,
		CASE
		  WHEN tcd.decision='APR' THEN 'APPROVE'
		  WHEN tcd.decision='REJ' THEN 'REJECT'
		  WHEN tcd.decision='CAN' THEN 'CANCEL'
		  ELSE tcd.decision
		END AS ca_decision,
		tcd.note AS ca_note,
		CASE
		  WHEN tcd.created_at IS NOT NULL
		  AND tfa.created_at IS NULL THEN FORMAT(tcd.created_at,'yyyy-MM-dd HH:mm:ss')
		  WHEN tfa.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE NULL
		END AS ActionDate,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		
		tdd.decision AS draft_decision,
		tdd.slik_result AS draft_slik_result,
		tdd.note AS draft_note,
		tdd.created_at AS draft_created_at,
		tdd.created_by AS draft_created_by,
		tdd.decision_by AS draft_decision_by,
		tdd.pernyataan_1 AS draft_pernyataan_1,
		tdd.pernyataan_2 AS draft_pernyataan_2,
		tdd.pernyataan_3 AS draft_pernyataan_3,
		tdd.pernyataan_4 AS draft_pernyataan_4,
		tdd.pernyataan_5 AS draft_pernyataan_5,
		tdd.pernyataan_6 AS draft_pernyataan_6,

		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		ta.interest_rate,
		ta.Tenor AS InstallmentPeriod,
		OTR,
		ta.DPAmount,
		ta.AF AS FinanceAmount,
		ta.interest_amount,
		ta.insurance_amount,
		ta.AdminFee,
		ta.provision_fee,
		ta.NTF,
		ta.NTFAkumulasi,
		(ta.NTF + ta.interest_amount) AS Total,
		ta.InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb2.BiroSpouseResult,
		CASE
		 WHEN rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN' THEN 1
		 ELSE 0
		END AS ActionEditData,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN
			cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN
			cte_trx_detail_biro tdb ON tm.ProspectID = tdb.prospect_id
		LEFT JOIN
			cte_trx_detail_biro2 tdb2 ON tm.ProspectID = tdb2.prospect_id
		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		LEFT JOIN
			cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
	 WHERE tm.BranchID = '426' AND (tm.ProspectID = 'SAL-XXX') AND tdd.draft_created_by= 'abc123'  AND tst.source_decision<>'PSI') AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryCa(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success without multi branch", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryCa{
			Search:      "6104",
			BranchID:    "426",
			MultiBranch: "0",
			Filter:      "REJECT",
			UserID:      "db1f4044e1dc574",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','6104') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
		 WHERE tm.BranchID IN ('426') AND (tcp.IDNumber = 'xxxxxx') AND tst.decision = 'REJ' AND tst.status_process='FIN' AND tst.source_decision<>'PSI') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by,
				x.pernyataan_1,
				x.pernyataan_2,
				x.pernyataan_3,
				x.pernyataan_4,
				x.pernyataan_5,
				x.pernyataan_6
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_detail_biro AS (
			SELECT
				prospect_id, url_pdf_report AS BiroCustomerResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'CUSTOMER'
		),
		cte_trx_detail_biro2 AS (
			SELECT
				prospect_id, url_pdf_report AS BiroSpouseResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'SPOUSE'
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		),
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as decision_ca,
		tcd.created_by as decision_by_ca,
		tr.additional_dp,
		CASE
		  WHEN tcd.decision='APR' THEN 'APPROVE'
		  WHEN tcd.decision='REJ' THEN 'REJECT'
		  WHEN tcd.decision='CAN' THEN 'CANCEL'
		  ELSE tcd.decision
		END AS ca_decision,
		tcd.note AS ca_note,
		CASE
		  WHEN tcd.created_at IS NOT NULL
		  AND tfa.created_at IS NULL THEN FORMAT(tcd.created_at,'yyyy-MM-dd HH:mm:ss')
		  WHEN tfa.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE NULL
		END AS ActionDate,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		
		tdd.decision AS draft_decision,
		tdd.slik_result AS draft_slik_result,
		tdd.note AS draft_note,
		tdd.created_at AS draft_created_at,
		tdd.created_by AS draft_created_by,
		tdd.decision_by AS draft_decision_by,
		tdd.pernyataan_1 AS draft_pernyataan_1,
		tdd.pernyataan_2 AS draft_pernyataan_2,
		tdd.pernyataan_3 AS draft_pernyataan_3,
		tdd.pernyataan_4 AS draft_pernyataan_4,
		tdd.pernyataan_5 AS draft_pernyataan_5,
		tdd.pernyataan_6 AS draft_pernyataan_6,

		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		ta.interest_rate,
		ta.Tenor AS InstallmentPeriod,
		OTR,
		ta.DPAmount,
		ta.AF AS FinanceAmount,
		ta.interest_amount,
		ta.insurance_amount,
		ta.AdminFee,
		ta.provision_fee,
		ta.NTF,
		ta.NTFAkumulasi,
		(ta.NTF + ta.interest_amount) AS Total,
		ta.InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb2.BiroSpouseResult,
		CASE
		 WHEN rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN' THEN 1
		 ELSE 0
		END AS ActionEditData,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN
			cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN
			cte_trx_detail_biro tdb ON tm.ProspectID = tdb.prospect_id
		LEFT JOIN
			cte_trx_detail_biro2 tdb2 ON tm.ProspectID = tdb2.prospect_id
		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		LEFT JOIN
			cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
	 WHERE tm.BranchID IN ('426') AND (tcp.IDNumber = 'xxxxxx') AND tst.decision = 'REJ' AND tst.status_process='FIN' AND tst.source_decision<>'PSI') AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryCa(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with region west java", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryCa{
			Search:      "aprospectid",
			BranchID:    "426",
			MultiBranch: "1",
			Filter:      "CANCEL",
			UserID:      "db1f4044e1dc574",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'db1f4044e1dc574' 
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("WEST JAVA", `["426","436","429","431","442","428","430"]`))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','aprospectid') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
		 WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx') AND tst.decision = 'CAN' AND tst.status_process='FIN' AND tst.source_decision<>'PSI') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by,
				x.pernyataan_1,
				x.pernyataan_2,
				x.pernyataan_3,
				x.pernyataan_4,
				x.pernyataan_5,
				x.pernyataan_6
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_detail_biro AS (
			SELECT
				prospect_id, url_pdf_report AS BiroCustomerResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'CUSTOMER'
		),
		cte_trx_detail_biro2 AS (
			SELECT
				prospect_id, url_pdf_report AS BiroSpouseResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'SPOUSE'
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		),
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
			SELECT
				tt.*
				FROM
				(
				SELECT
				tm.ProspectID,
				cb.BranchName,
				cb.BranchID,
				tst.activity,
				tst.source_decision,
				tst.status_process,
				tst.decision,
				tst.reason,
				tcd.decision as decision_ca,
				tcd.created_by as decision_by_ca,
				tr.additional_dp,
				CASE
				WHEN tcd.decision='APR' THEN 'APPROVE'
				WHEN tcd.decision='REJ' THEN 'REJECT'
				WHEN tcd.decision='CAN' THEN 'CANCEL'
				ELSE tcd.decision
				END AS ca_decision,
				tcd.note AS ca_note,
				CASE
				WHEN tcd.created_at IS NOT NULL
				AND tfa.created_at IS NULL THEN FORMAT(tcd.created_at,'yyyy-MM-dd HH:mm:ss')
				WHEN tfa.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
				ELSE NULL
				END AS ActionDate,
				CASE
				WHEN tst.decision = 'CPR'
				AND tst.source_decision = 'CRA'
				AND tst.activity = 'UNPR'
				AND tcd.decision IS NULL THEN 1
				ELSE 0
				END AS ShowAction,
				CASE
				WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
				ELSE 'NE'
				END AS incoming_source,
				
				tdd.decision AS draft_decision,
				tdd.slik_result AS draft_slik_result,
				tdd.note AS draft_note,
				tdd.created_at AS draft_created_at,
				tdd.created_by AS draft_created_by,
				tdd.decision_by AS draft_decision_by,
				tdd.pernyataan_1 AS draft_pernyataan_1,
				tdd.pernyataan_2 AS draft_pernyataan_2,
				tdd.pernyataan_3 AS draft_pernyataan_3,
				tdd.pernyataan_4 AS draft_pernyataan_4,
				tdd.pernyataan_5 AS draft_pernyataan_5,
				tdd.pernyataan_6 AS draft_pernyataan_6,

				tcp.CustomerID,
				tcp.CustomerStatus,
				tcp.SurveyResult,
				tm.created_at,
				tm.order_at,
				tm.lob,
				scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
				scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
				scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
				tcp.BirthDate,
				scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
				CASE
				WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
				WHEN tcp.Gender = 'F' THEN 'Perempuan'
				END AS 'Gender',
				scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
				CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
				cal.Kelurahan AS LegalKelurahan,
				cal.Kecamatan AS LegalKecamatan,
				cal.ZipCode AS LegalZipcode,
				cal.City AS LegalCity,
				scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
				CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
				car.Kelurahan AS ResidenceKelurahan,
				car.Kecamatan AS ResidenceKecamatan,
				car.ZipCode AS ResidenceZipcode,
				car.City AS ResidenceCity,
				scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
				scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
				edu.value AS Education,
				mst.value AS MaritalStatus,
				tcp.NumOfDependence,
				hst.value AS HomeStatus,
				mn.value AS StaySinceMonth,
				tcp.StaySinceYear,
				ta.ProductOfferingID,
				ta.dealer,
				ta.LifeInsuranceFee,
				ta.AssetInsuranceFee,
				'KMB MOTOR' AS AssetType,
				ti.asset_description,
				ti.manufacture_year,
				ti.color,
				chassis_number,
				engine_number,
				CASE
				WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
				WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
				WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
				WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
				END AS bpkb_name,
				ti.owner_asset,
				ti.license_plate,
				ta.interest_rate,
				ta.Tenor AS InstallmentPeriod,
				OTR,
				ta.DPAmount,
				ta.AF AS FinanceAmount,
				ta.interest_amount,
				ta.insurance_amount,
				ta.AdminFee,
				ta.provision_fee,
				ta.NTF,
				ta.NTFAkumulasi,
				(ta.NTF + ta.interest_amount) AS Total,
				ta.InstallmentAmount AS MonthlyInstallment,
				FirstInstallment,
				pr.value AS ProfessionID,
				jt.value AS JobType,
				jb.value AS JobPosition,
				mn2.value AS EmploymentSinceMonth,
				tce.EmploymentSinceYear,
				tce.CompanyName,
				cac.AreaPhone AS CompanyAreaPhone,
				cac.Phone AS CompanyPhone,
				tcp.ExtCompanyPhone,
				scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
				CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
				cac.Kelurahan AS CompanyKelurahan,
				cac.Kecamatan AS CompanyKecamatan,
				car.ZipCode AS CompanyZipcode,
				car.City AS CompanyCity,
				tce.MonthlyFixedIncome,
				tce.MonthlyVariableIncome,
				tce.SpouseIncome,
				tcp.SourceOtherIncome,
				tcs.FullName AS SpouseLegalName,
				tcs.CompanyName AS SpouseCompanyName,
				tcs.CompanyPhone AS SpouseCompanyPhone,
				tcs.MobilePhone AS SpouseMobilePhone,
				tcs.IDNumber AS SpouseIDNumber,
				pr2.value AS SpouseProfession,
				em.Name AS EmconName,
				em.Relationship,
				em.MobilePhone AS EmconMobilePhone,
				scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
				CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
				cae.Kelurahan AS EmergencyKelurahan,
				cae.Kecamatan AS EmergencyKecamatan,
				cae.ZipCode AS EmergencyZipcode,
				cae.City AS EmergencyCity,
				cae.AreaPhone AS EmergencyAreaPhone,
				cae.Phone AS EmergencyPhone,
				tce.IndustryTypeID,
				FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
				tak.ScsScore,
				tak.ScsStatus,
				tdb.BiroCustomerResult,
				tdb2.BiroSpouseResult,
				CASE
				 WHEN rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN' THEN 1
				ELSE 0
				END AS ActionEditData,
				tde.deviasi_id,
				mkd.deskripsi AS deviasi_description,
				'REJECT' AS deviasi_decision,
				tde.reason AS deviasi_reason,
				CASE
				WHEN ted.ProspectID IS NOT NULL THEN 1
				ELSE 0
				END AS is_edd,
				ted.is_highrisk,
				ted.pernyataan_1,
				ted.pernyataan_2,
				ted.pernyataan_3,
				ted.pernyataan_4,
				ted.pernyataan_5,
				ted.pernyataan_6
			FROM
				trx_master tm WITH (nolock)
				INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
				INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
				INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
				INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
				INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
				INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
				INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
				INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
				LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
				LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
				LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
				LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
				LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
				LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
				LEFT JOIN
					cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
				LEFT JOIN
					cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
				LEFT JOIN
					cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
				LEFT JOIN
					cte_trx_detail_biro tdb ON tm.ProspectID = tdb.prospect_id
				LEFT JOIN
					cte_trx_detail_biro2 tdb2 ON tm.ProspectID = tdb2.prospect_id
				INNER JOIN (
					SELECT
					ProspectID,
					Address,
					RT,
					RW,
					Kelurahan,
					Kecamatan,
					ZipCode,
					City
					FROM
					trx_customer_address WITH (nolock)
					WHERE
					"Type" = 'LEGAL'
				) cal ON tm.ProspectID = cal.ProspectID
				INNER JOIN (
					SELECT
					ProspectID,
					Address,
					RT,
					RW,
					Kelurahan,
					Kecamatan,
					ZipCode,
					City
					FROM
					trx_customer_address WITH (nolock)
					WHERE
					"Type" = 'RESIDENCE'
				) car ON tm.ProspectID = car.ProspectID
				INNER JOIN (
					SELECT
					ProspectID,
					Address,
					RT,
					RW,
					Kelurahan,
					Kecamatan,
					ZipCode,
					City,
					Phone,
					AreaPhone
					FROM
					trx_customer_address WITH (nolock)
					WHERE
					"Type" = 'COMPANY'
				) cac ON tm.ProspectID = cac.ProspectID
				INNER JOIN (
					SELECT
					ProspectID,
					Address,
					RT,
					RW,
					Kelurahan,
					Kecamatan,
					ZipCode,
					City,
					Phone,
					AreaPhone
					FROM
					trx_customer_address WITH (nolock)
					WHERE
					"Type" = 'EMERGENCY'
				) cae ON tm.ProspectID = cae.ProspectID

				INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
				LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
				LEFT JOIN (
				SELECT
					[key],
					value
				FROM
					app_config ap WITH (nolock)
				WHERE
					group_name = 'Education'
				) edu ON tcp.Education = edu.[key]
				LEFT JOIN (
				SELECT
					[key],
					value
				FROM
					app_config ap WITH (nolock)
				WHERE
					group_name = 'MaritalStatus'
				) mst ON tcp.MaritalStatus = mst.[key]
				LEFT JOIN (
				SELECT
					[key],
					value
				FROM
					app_config ap WITH (nolock)
				WHERE
					group_name = 'HomeStatus'
				) hst ON tcp.HomeStatus = hst.[key]
				LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
				LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
				LEFT JOIN (
				SELECT
					[key],
					value
				FROM
					app_config ap WITH (nolock)
				WHERE
					group_name = 'JobType'
				) jt ON tce.JobType = jt.[key]
				LEFT JOIN (
				SELECT
					[key],
					value
				FROM
					app_config ap WITH (nolock)
				WHERE
					group_name = 'JobPosition'
				) jb ON tce.JobPosition = jb.[key]
				LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
				LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
				LEFT JOIN
					cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
			 WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx') AND tst.decision = 'CAN' AND tst.status_process='FIN' AND tst.source_decision<>'PSI') AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryCa(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with region ALL", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryCa{
			Search:      "aprospectid",
			BranchID:    "426",
			MultiBranch: "1",
			Filter:      "APPROVE",
			UserID:      "abc123",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123' 
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("ALL", `["426","436","429","431","442","428","430"]`))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','aprospectid') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
		 WHERE (tcp.LegalName = 'xxxxxx') AND tst.decision = 'APR' AND tst.status_process='FIN' AND tst.source_decision<>'PSI') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by,
				x.pernyataan_1,
				x.pernyataan_2,
				x.pernyataan_3,
				x.pernyataan_4,
				x.pernyataan_5,
				x.pernyataan_6
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_detail_biro AS (
			SELECT
				prospect_id, url_pdf_report AS BiroCustomerResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'CUSTOMER'
		),
		cte_trx_detail_biro2 AS (
			SELECT
				prospect_id, url_pdf_report AS BiroSpouseResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'SPOUSE'
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		),
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as decision_ca,
		tcd.created_by as decision_by_ca,
		tr.additional_dp,
		CASE
		  WHEN tcd.decision='APR' THEN 'APPROVE'
		  WHEN tcd.decision='REJ' THEN 'REJECT'
		  WHEN tcd.decision='CAN' THEN 'CANCEL'
		  ELSE tcd.decision
		END AS ca_decision,
		tcd.note AS ca_note,
		CASE
		  WHEN tcd.created_at IS NOT NULL
		  AND tfa.created_at IS NULL THEN FORMAT(tcd.created_at,'yyyy-MM-dd HH:mm:ss')
		  WHEN tfa.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE NULL
		END AS ActionDate,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		
		tdd.decision AS draft_decision,
		tdd.slik_result AS draft_slik_result,
		tdd.note AS draft_note,
		tdd.created_at AS draft_created_at,
		tdd.created_by AS draft_created_by,
		tdd.decision_by AS draft_decision_by,
		tdd.pernyataan_1 AS draft_pernyataan_1,
		tdd.pernyataan_2 AS draft_pernyataan_2,
		tdd.pernyataan_3 AS draft_pernyataan_3,
		tdd.pernyataan_4 AS draft_pernyataan_4,
		tdd.pernyataan_5 AS draft_pernyataan_5,
		tdd.pernyataan_6 AS draft_pernyataan_6,

		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		ta.interest_rate,
		ta.Tenor AS InstallmentPeriod,
		OTR,
		ta.DPAmount,
		ta.AF AS FinanceAmount,
		ta.interest_amount,
		ta.insurance_amount,
		ta.AdminFee,
		ta.provision_fee,
		ta.NTF,
		ta.NTFAkumulasi,
		(ta.NTF + ta.interest_amount) AS Total,
		ta.InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb2.BiroSpouseResult,
		CASE
		 WHEN rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN' THEN 1
		 ELSE 0
		END AS ActionEditData,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN
			cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN
			cte_trx_detail_biro tdb ON tm.ProspectID = tdb.prospect_id
		LEFT JOIN
			cte_trx_detail_biro2 tdb2 ON tm.ProspectID = tdb2.prospect_id
		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		LEFT JOIN
			cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
	 WHERE (tcp.LegalName = 'xxxxxx') AND tst.decision = 'APR' AND tst.status_process='FIN' AND tst.source_decision<>'PSI') AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryCa(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestSaveDraftData(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	data := entity.TrxDraftCaDecision{
		ProspectID: "TST001",
		Decision:   constant.DB_DECISION_REJECT,
		SlikResult: "lancar",
		Note:       "Catatan",
		CreatedBy:  "SYSTEM",
		DecisionBy: "5XeZs9PCeiPcZGS6azt",
	}
	t.Run("success update", func(t *testing.T) {

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_draft_ca_decision" SET "created_at" = ?, "created_by" = ?, "decision" = ?, "decision_by" = ?, "note" = ?, "pernyataan_1" = ?, "pernyataan_2" = ?, "pernyataan_3" = ?, "pernyataan_4" = ?, "pernyataan_5" = ?, "pernyataan_6" = ?, "slik_result" = ? WHERE (ProspectID = ?)`)).
			WithArgs(sqlmock.AnyArg(), data.CreatedBy, data.Decision, data.DecisionBy, data.Note, data.Pernyataan1, data.Pernyataan2, data.Pernyataan3, data.Pernyataan4, data.Pernyataan5, data.Pernyataan6, data.SlikResult, data.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.SaveDraftData(data)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("success insert", func(t *testing.T) {

		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_draft_ca_decision" SET "created_at" = ?, "created_by" = ?, "decision" = ?, "decision_by" = ?, "note" = ?, "pernyataan_1" = ?, "pernyataan_2" = ?, "pernyataan_3" = ?, "pernyataan_4" = ?, "pernyataan_5" = ?, "pernyataan_6" = ?, "slik_result" = ? WHERE (ProspectID = ?)`)).
			WithArgs(sqlmock.AnyArg(), data.CreatedBy, data.Decision, data.DecisionBy, data.Note, data.Pernyataan1, data.Pernyataan2, data.Pernyataan3, data.Pernyataan4, data.Pernyataan5, data.Pernyataan6, data.SlikResult, data.ProspectID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_draft_ca_decision" ("ProspectID","decision","slik_result","note","created_at","created_by","decision_by","pernyataan_1","pernyataan_2","pernyataan_3","pernyataan_4","pernyataan_5","pernyataan_6") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(data.ProspectID, data.Decision, data.SlikResult, data.Note, sqlmock.AnyArg(), data.CreatedBy, data.DecisionBy, data.Pernyataan1, data.Pernyataan2, data.Pernyataan3, data.Pernyataan4, data.Pernyataan5, data.Pernyataan6).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.SaveDraftData(data)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})
}

func TestProcessTransaction(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	trxCaDecision := entity.TrxCaDecision{
		ProspectID:    "TST001",
		Decision:      constant.DB_DECISION_REJECT,
		SlikResult:    "lancar",
		Note:          "Catatan",
		CreatedBy:     "SYSTEM",
		DecisionBy:    "5XeZs9PCeiPcZGS6azt",
		FinalApproval: "CBM",
	}

	trxStatus := entity.TrxStatus{
		ProspectID:     trxCaDecision.ProspectID,
		StatusProcess:  constant.ACTIVITY_UNPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		Reason:         trxCaDecision.SlikResult.(string),
	}

	trxDetail := entity.TrxDetail{
		ProspectID:     trxCaDecision.ProspectID,
		StatusProcess:  constant.ACTIVITY_UNPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_APR,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		NextStep:       constant.DB_DECISION_BRANCH_MANAGER,
		Info:           trxCaDecision.SlikResult,
		CreatedBy:      trxCaDecision.CreatedBy,
		Reason:         trxCaDecision.SlikResult,
	}

	trxHistoryApproval := entity.TrxHistoryApprovalScheme{
		ProspectID:            trxCaDecision.ProspectID,
		Decision:              trxCaDecision.Decision,
		Reason:                "lancar",
		Note:                  trxCaDecision.Note,
		CreatedBy:             trxCaDecision.CreatedBy,
		DecisionBy:            trxCaDecision.DecisionBy,
		NeedEscalation:        0,
		NextFinalApprovalFlag: 0,
		SourceDecision:        trxDetail.SourceDecision,
	}
	t.Run("success update", func(t *testing.T) {

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ?  WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ? AND source_decision = 'CRA'))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_ca_decision" ("ProspectID","decision","slik_result","note","created_at","created_by","decision_by","final_approval") VALUES (?,?,?,?,?,?,?,?)`)).
			WithArgs(trxCaDecision.ProspectID, trxCaDecision.Decision, trxCaDecision.SlikResult, trxCaDecision.Note, sqlmock.AnyArg(), trxCaDecision.CreatedBy, trxCaDecision.DecisionBy, trxCaDecision.FinalApproval).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxCaDecision.ProspectID, trxCaDecision.Decision, trxCaDecision.SlikResult.(string), trxCaDecision.Note, sqlmock.AnyArg(), trxCaDecision.CreatedBy, trxCaDecision.DecisionBy, sqlmock.AnyArg(), sqlmock.AnyArg(), trxDetail.SourceDecision, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_draft_ca_decision"  WHERE (ProspectID = ?)`)).
			WithArgs(trxCaDecision.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.ProcessTransaction(trxCaDecision, trxHistoryApproval, trxStatus, trxDetail, false, entity.TrxEDD{})
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})
}

func TestProcessReturnOrder(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	ppid := "TST001"

	trxStatus := entity.TrxStatus{
		ProspectID:     ppid,
		StatusProcess:  constant.ACTIVITY_UNPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
	}

	trxDetail := entity.TrxDetail{
		ProspectID:     ppid,
		StatusProcess:  constant.ACTIVITY_UNPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_APR,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		NextStep:       constant.DB_DECISION_BRANCH_MANAGER,
		Reason:         constant.REASON_RETURN_ORDER,
	}

	t.Run("success update", func(t *testing.T) {

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_details" WHERE "trx_details"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(ppid, ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_deviasi" WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_edd" WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_final_approval" WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_agreements" WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_prescreening" WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_ca_decision"  WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_draft_ca_decision"  WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_history_approval_scheme"  WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_akkk"  WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.ProcessReturnOrder(ppid, trxStatus, trxDetail)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})
}

func TestGetInquirySearch(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.InquirySearch{{ProspectID: "EFM03406412522151347", BranchName: "BANDUNG", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerID: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	t.Run("success with region all", func(t *testing.T) {
		// Expected input and output
		req := request.ReqSearchInquiry{
			BranchID:    "426",
			MultiBranch: "1",
			UserID:      "abc123",
			Search:      "SAL-12345",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123' 
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("ALL", `["426","436","429","431","442","428","430"]`))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
			COUNT(tt.ProspectID) AS totalRow
			FROM
			(SELECT tm.ProspectID
			FROM
			trx_master tm WITH (nolock)
			INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
			INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
			INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
			INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
			INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
			INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
			INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
			INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
			INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
			LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
			LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
			LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
			WHERE (tm.ProspectID = 'SAL-12345')) AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.decision,
		CASE
		  WHEN tst.status_process='FIN' AND tst.decision='APR' THEN 'Approve'
		  WHEN tst.status_process='FIN' AND tst.decision='REJ' THEN 'Reject'
		  WHEN tst.status_process='FIN' AND tst.decision='CAN' THEN 'Cancel'
		  ELSE '-'
		END AS FinalStatus,
		CASE
		  WHEN tps.ProspectID IS NOT NULL
		  AND tcd.decision IS NULL
		  AND tst.source_decision NOT IN('CBM','DRM','GMO','COM','GMC','UCC')
		  AND tst.status_process <> 'FIN' THEN 1
		  ELSE 0
		END AS ActionReturn,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP'
		  AND tst.decision='REJ' OR tst.decision='CAN' THEN 0
		  ELSE 1
		END AS ActionCancel,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	WHERE (tm.ProspectID = 'SAL-12345')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).
				AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquirySearch(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with region west java", func(t *testing.T) {
		// Expected input and output
		req := request.ReqSearchInquiry{
			BranchID:    "426",
			MultiBranch: "1",
			UserID:      "abc123",
			Search:      "ahmad name",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123' 
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("WEST JAVA", `["426","436","429","431","442","428","430"]`))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','ahmad name') AS encrypt`)).
			WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).
				AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx')) AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.decision,
		CASE
		  WHEN tst.status_process='FIN' AND tst.decision='APR' THEN 'Approve'
		  WHEN tst.status_process='FIN' AND tst.decision='REJ' THEN 'Reject'
		  WHEN tst.status_process='FIN' AND tst.decision='CAN' THEN 'Cancel'
		  ELSE '-'
		END AS FinalStatus,
		CASE
		  WHEN tps.ProspectID IS NOT NULL
		  AND tcd.decision IS NULL
		  AND tst.source_decision NOT IN('CBM','DRM','GMO','COM','GMC','UCC')
		  AND tst.status_process <> 'FIN' THEN 1
		  ELSE 0
		END AS ActionReturn,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP'
		  AND tst.decision='REJ' OR tst.decision='CAN' THEN 0
		  ELSE 1
		END AS ActionCancel,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.LegalName = 'xxxxxx')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).
				AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquirySearch(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with region null", func(t *testing.T) {
		// Expected input and output
		req := request.ReqSearchInquiry{
			BranchID:    "426",
			MultiBranch: "1",
			UserID:      "abc123",
			Search:      "7171072102760001",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123' 
		)
		AND b.lob_id='125'`)).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','7171072102760001') AS encrypt`)).
			WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).
				AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		WHERE tm.BranchID = '426' AND (tcp.IDNumber = 'xxxxxx')) AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.decision,
		CASE
		  WHEN tst.status_process='FIN' AND tst.decision='APR' THEN 'Approve'
		  WHEN tst.status_process='FIN' AND tst.decision='REJ' THEN 'Reject'
		  WHEN tst.status_process='FIN' AND tst.decision='CAN' THEN 'Cancel'
		  ELSE '-'
		END AS FinalStatus,
		CASE
		  WHEN tps.ProspectID IS NOT NULL
		  AND tcd.decision IS NULL
		  AND tst.source_decision NOT IN('CBM','DRM','GMO','COM','GMC','UCC')
		  AND tst.status_process <> 'FIN' THEN 1
		  ELSE 0
		END AS ActionReturn,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP'
		  AND tst.decision='REJ' OR tst.decision='CAN' THEN 0
		  ELSE 1
		END AS ActionCancel,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	WHERE tm.BranchID = '426' AND (tcp.IDNumber = 'xxxxxx')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).
				AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquirySearch(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success without multi branch", func(t *testing.T) {
		// Expected input and output
		req := request.ReqSearchInquiry{
			BranchID:    "426",
			MultiBranch: "0",
			UserID:      "abc123",
			Search:      "random123",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','random123') AS encrypt`)).
			WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).
				AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		WHERE tm.BranchID IN ('426') AND (tm.ProspectID = 'random123' OR tcp.IDNumber = 'xxxxxx' OR tcp.LegalName = 'xxxxxx')) AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.decision,
		CASE
		  WHEN tst.status_process='FIN' AND tst.decision='APR' THEN 'Approve'
		  WHEN tst.status_process='FIN' AND tst.decision='REJ' THEN 'Reject'
		  WHEN tst.status_process='FIN' AND tst.decision='CAN' THEN 'Cancel'
		  ELSE '-'
		END AS FinalStatus,
		CASE
		  WHEN tps.ProspectID IS NOT NULL
		  AND tcd.decision IS NULL
		  AND tst.source_decision NOT IN('CBM','DRM','GMO','COM','GMC','UCC')
		  AND tst.status_process <> 'FIN' THEN 1
		  ELSE 0
		END AS ActionReturn,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP'
		  AND tst.decision='REJ' OR tst.decision='CAN' THEN 0
		  ELSE 1
		END AS ActionCancel,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	WHERE tm.BranchID IN ('426') AND (tm.ProspectID = 'random123' OR tcp.IDNumber = 'xxxxxx' OR tcp.LegalName = 'xxxxxx')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).
				AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquirySearch(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

}

func TestGetApprovalReason(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedReason := []entity.ApprovalReason{
		{
			ReasonID: "1|APR|Oke",
			Value:    "Oke",
			Type:     "APR",
		},
	}

	req := request.ReqApprovalReason{
		Type: "APR",
	}

	t.Run("success", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(tt.id) AS totalRow FROM (SELECT CONCAT(ReasonID, '|', Type, '|', Description) AS 'id', Description AS 'value', [Type] FROM tblApprovalReason WHERE IsActive = 'True' AND [Type] = 'APR') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("8"))
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT CONCAT(ReasonID, '|', Type, '|', Description) AS 'id', Description AS 'value', [Type] FROM tblApprovalReason WHERE IsActive = 'True'  AND [Type] = 'APR' ORDER BY ReasonID ASC`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "value", "Type"}).
				AddRow("1|APR|Oke", "Oke", "APR"))
		mock.ExpectCommit()

		// Call the function
		reason, _, err := repo.GetApprovalReason(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedReason, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success not found", func(t *testing.T) {
		// Mock SQL query to simulate record not found
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT CONCAT(ReasonID, '|', Type, '|', Description) AS 'id', Description AS 'value', [Type] FROM tblApprovalReason WHERE IsActive = 'True'  AND [Type] = 'APR' ORDER BY ReasonID ASC`)).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetApprovalReason(req, nil)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

}

func TestGetAFMobilePhone(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expected := entity.AFMobilePhone{
		AFValue:     1000,
		OTR:         10000,
		DPAmount:    1000,
		MobilePhone: "0896245242",
	}

	t.Run("success", func(t *testing.T) {

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT AF, SCP.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone, OTR, DPAmount FROM trx_apk apk WITH (nolock) INNER JOIN trx_customer_personal tcp WITH (nolock) ON apk.ProspectID = tcp.ProspectID WHERE apk.ProspectID = 'ppid'`)).
			WillReturnRows(sqlmock.NewRows([]string{"AF", "MobilePhone", "OTR", "DPAmount"}).
				AddRow(1000, "0896245242", 10000, 1000))
		// Call the function
		data, err := repo.GetAFMobilePhone("ppid")

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expected, data, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		// Mock SQL query to simulate record not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT AF, SCP.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone, OTR, DPAmount FROM trx_apk apk WITH (nolock) INNER JOIN trx_customer_personal tcp WITH (nolock) ON apk.ProspectID = tcp.ProspectID WHERE apk.ProspectID = 'ppid'`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Call the function
		_, err := repo.GetAFMobilePhone("ppid")

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetRegionBranch(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expected := []entity.RegionBranch{
		{
			RegionName:   "WEST JAVA",
			BranchMember: `["426","436","429","431","442","428","430"]`,
		},
	}

	t.Run("success", func(t *testing.T) {

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock) INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN ( SELECT value FROM region_user ru WITH (nolock) cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',') WHERE ru.user_id = 'user_id' ) AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("WEST JAVA", `["426","436","429","431","442","428","430"]`))
		// Call the function
		data, err := repo.GetRegionBranch("user_id")

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expected, data, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		// Mock SQL query to simulate record not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock) INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN ( SELECT value FROM region_user ru WITH (nolock) cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',') WHERE ru.user_id = 'user_id' ) AND b.lob_id='125'`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Call the function
		_, err := repo.GetRegionBranch("user_id")

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetAkkk(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expected := entity.Akkk{
		ProspectID: "ppid",
	}

	t.Run("success", func(t *testing.T) {

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT ts.ProspectID, 
		ta2.FinancePurpose,
		scp.dbo.DEC_B64('SEC',tcp.LegalName) as LegalName,
		scp.dbo.DEC_B64('SEC',tcp.IDNumber) as IDNumber,  
		tcp.PersonalNPWP,
		scp.dbo.DEC_B64('SEC',tcp.SurgateMotherName) as SurgateMotherName,
		tce.ProfessionID,
		tcp.CustomerStatus,
		ta.CustomerType,
		tcp.Gender,
		scp.dbo.DEC_B64('SEC',tcp.BirthPlace) as BirthPlace,
		tcp.BirthDate,
		tcp.Education,
		scp.dbo.DEC_B64('SEC',tcp.MobilePhone) as MobilePhone,
		scp.dbo.DEC_B64('SEC',tcp.Email) as Email,
		tcs.LegalName as SpouseLegalName,
		tcs.IDNumber as SpouseIDNumber,
		tcs.SurgateMotherName as SpouseSurgateMotherName,
		tcs.ProfessionID as SpouseProfessionID,
		ta.SpouseType,
		tcs.Gender as SpouseGender,
		tcs.BirthPlace as SpouseBirthPlace,
		tcs.BirthDate as SpouseBirthDate,
		tcs.MobilePhone as SpouseMobilePhone,
		tce2.VerificationWith,
		tce2.Relationship as EmconRelationship,
		tce2.EmconVerified,
		tca.Address,
		tce2.MobilePhone as EmconMobilePhone,
		tce2.VerifyBy,
		tce2.KnownCustomerAddress,
		tcp.StaySinceYear,
		tcp.StaySinceMonth,
		tce2.KnownCustomerJob,
		CASE tce.ProfessionID 
			WHEN 'WRST' THEN 'WIRASWASTA'
			WHEN 'PRO' THEN 'PROFESSIONAL'
			WHEN 'KRYSW' THEN 'KARYAWAN SWASTA'
			WHEN 'PNS' THEN 'PEGAWAI NEGERI SIPIL'
			WHEN 'ANG' THEN 'ANGKATAN'
			ELSE '-'
		END as Job,
		tce.EmploymentSinceYear,
		tce.EmploymentSinceMonth,
		tce.IndustryTypeID,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		CASE ti.bpkb_name
			WHEN 'K' THEN 'NAMA SAMA'
			WHEN 'P' THEN 'NAMA SAMA'
			WHEN 'KK' THEN 'NAMA BEDA'
			WHEN 'O' THEN 'NAMA BEDA'
			ELSE '-'
		END as BpkbName,
		tdb.plafon as Plafond,
		tdb.baki_debet_non_collateral as BakiDebet,
		tdb.fasilitas_aktif as FasilitasAktif,
		CASE
			WHEN tdb.kualitas_kredit_terburuk IS NOT NULL THEN CONCAT(tdb.kualitas_kredit_terburuk,' ',tdb.bulan_kualitas_terburuk)
			ELSE NULL
		END as ColTerburuk,
		tdb.baki_debet_kualitas_terburuk as BakiDebetTerburuk,
		CASE
			WHEN tdb.kualitas_kredit_terakhir IS NOT NULL THEN CONCAT(tdb.kualitas_kredit_terakhir,' ',tdb.bulan_kualitas_kredit_terakhir)
			ELSE NULL
		END as ColTerakhirAktif,
		tdb2.plafon as SpousePlafond,
		tdb2.baki_debet_non_collateral as SpouseBakiDebet,
		tdb2.fasilitas_aktif as SpouseFasilitasAktif,
		CASE
			WHEN tdb2.kualitas_kredit_terburuk IS NOT NULL THEN CONCAT(tdb2.kualitas_kredit_terburuk,' ',tdb2.bulan_kualitas_terburuk)
			ELSE NULL
		END as SpouseColTerburuk,
		tdb2.baki_debet_kualitas_terburuk as SpouseBakiDebetTerburuk,
		CASE
			WHEN tdb2.kualitas_kredit_terakhir IS NOT NULL THEN CONCAT(tdb2.kualitas_kredit_terakhir,' ',tdb2.bulan_kualitas_kredit_terakhir)
			ELSE NULL
		END as SpouseColTerakhirAktif,
		ta.ScsScore,
		ta.AgreementStatus,
		ta.TotalAgreementAktif,
		ta.MaxOVDAgreementAktif,
		ta.LastMaxOVDAgreement,
		tf.customer_segment,
		ta.LatestInstallment,
		ta2.NTFAkumulasi,
		(CAST(ISNULL(ta2.InstallmentAmount, 0) AS NUMERIC(17,2)) +
		CASE 
			WHEN ta.TotalDSR = ta.DSRFMF THEN 0
			WHEN ta.TotalDSR IS NULL THEN 0
			ELSE CAST(ISNULL(tf.total_installment_amount_biro, 0) AS NUMERIC(17,2))
		END +
		CAST(ISNULL(ta.InstallmentAmountFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountSpouseFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOther, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOtherSpouse, 0) AS NUMERIC(17,2)) -
		CAST(ISNULL(ta.InstallmentTopup, 0) AS NUMERIC(17,2)) ) as TotalInstallment,
		(CAST(ISNULL(tce.MonthlyFixedIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.MonthlyVariableIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.SpouseIncome, 0) AS NUMERIC(17,2)) ) as TotalIncome,
		CASE 
			WHEN ta.TotalDSR IS NULL THEN CAST(ISNULL(ta.DSRFMF, 0) AS NUMERIC(17,2)) + CAST(ISNULL(ta.DSRPBK, 0) AS NUMERIC(17,2))
			ELSE ta.TotalDSR
		END as TotalDSR,
		CASE
			WHEN ta.EkycSource IS NOT NULL THEN CONCAT(ta.EkycSource,' - ',ta.EkycReason)
			ELSE NULL
		END as EkycSource,
		ta.EkycSimiliarity,
		ta.EkycReason,
		CASE tia.info 
			WHEN '1' THEN 'APR'
			ELSE 'REJ'
		END cmo_decision,
		tia.name as cmo_name,
		tia.recom_date cmo_date,
		tcd.decision as ca_decision,
		tcd.note as ca_note,
		tcd.decision_by as ca_name,
		FORMAT(tcd.created_at,'yyyy-MM-dd') as ca_date,
		cbm.decision as cbm_decision,
		cbm.note as cbm_note,
		cbm.decision_by as cbm_name,
		FORMAT(cbm.created_at,'yyyy-MM-dd') as cbm_date,
		drm.decision as drm_decision,
		drm.note as drm_note,
		drm.decision_by as drm_name,
		FORMAT(drm.created_at,'yyyy-MM-dd') as drm_date,
		gmo.decision as gmo_decision,
		gmo.note as gmo_note,
		gmo.decision_by as gmo_name,
		FORMAT(gmo.created_at,'yyyy-MM-dd') as gmo_date
		FROM trx_status ts WITH (nolock)
		LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID 
		LEFT JOIN trx_customer_employment tce WITH (nolock) ON ts.ProspectID = tce.ProspectID
		LEFT JOIN trx_akkk ta WITH (nolock) ON ts.ProspectID = ta.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON ts.ProspectID = tcs.ProspectID
		LEFT JOIN trx_customer_emcon tce2 WITH (nolock) ON ts.ProspectID = tce2.ProspectID
		LEFT OUTER JOIN ( 
			SELECT scp.dbo.DEC_B64('SEC', Address) AS Address, ProspectID, Phone 
			FROM trx_customer_address WITH (nolock)
			WHERE Type = 'EMERGENCY' 
		) AS tca ON ts.ProspectID = tca.ProspectID 
		LEFT JOIN trx_item ti WITH (nolock) ON ts.ProspectID = ti.ProspectID
		LEFT OUTER JOIN ( 
			SELECT * FROM trx_detail_biro WITH (nolock)
			WHERE subject = 'CUSTOMER' 
		) AS tdb ON ts.ProspectID = tdb.prospect_id
		LEFT JOIN trx_filtering tf WITH (nolock) ON ts.ProspectID = tf.prospect_id 
		LEFT JOIN trx_apk ta2 WITH (nolock) ON ts.ProspectID = ta2.ProspectID
		LEFT OUTER JOIN ( 
			SELECT * FROM trx_detail_biro WITH (nolock)
			WHERE subject = 'SPOUSE' 
		) AS tdb2 ON ts.ProspectID = tdb2.prospect_id
		LEFT JOIN trx_ca_decision tcd WITH (nolock) ON ts.ProspectID = tcd.ProspectID
		LEFT JOIN trx_info_agent tia WITH (nolock) ON ts.ProspectID = tia.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas1 WITH (nolock)
			WHERE thas1.source_decision = 'CBM' AND thas1.created_at = (SELECT MAX(tha1.created_at) From trx_history_approval_scheme tha1 WHERE tha1.source_decision = thas1.source_decision AND tha1.ProspectID = thas1.ProspectID)
		) AS cbm ON ts.ProspectID = cbm.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas2 WITH (nolock)
			WHERE thas2.source_decision = 'DRM' AND thas2.created_at = (SELECT MAX(tha2.created_at) From trx_history_approval_scheme tha2 WHERE tha2.source_decision = thas2.source_decision AND tha2.ProspectID = thas2.ProspectID)
		) AS drm ON ts.ProspectID = drm.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas3 WITH (nolock)
			WHERE thas3.source_decision = 'GMO' AND thas3.created_at = (SELECT MAX(tha3.created_at) From trx_history_approval_scheme tha3 WHERE tha3.source_decision = thas3.source_decision AND tha3.ProspectID = thas3.ProspectID)
		) AS gmo ON ts.ProspectID = gmo.ProspectID
		WHERE ts.ProspectID = 'ppid'`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID"}).
				AddRow("ppid"))
		// Call the function
		data, err := repo.GetAkkk("ppid")

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expected, data, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error", func(t *testing.T) {
		// Mock SQL query to simulate record not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT ts.ProspectID, 
		ta2.FinancePurpose,
		scp.dbo.DEC_B64('SEC',tcp.LegalName) as LegalName,
		scp.dbo.DEC_B64('SEC',tcp.IDNumber) as IDNumber,  
		tcp.PersonalNPWP,
		scp.dbo.DEC_B64('SEC',tcp.SurgateMotherName) as SurgateMotherName,
		tce.ProfessionID,
		tcp.CustomerStatus,
		ta.CustomerType,
		tcp.Gender,
		scp.dbo.DEC_B64('SEC',tcp.BirthPlace) as BirthPlace,
		tcp.BirthDate,
		tcp.Education,
		scp.dbo.DEC_B64('SEC',tcp.MobilePhone) as MobilePhone,
		scp.dbo.DEC_B64('SEC',tcp.Email) as Email,
		tcs.LegalName as SpouseLegalName,
		tcs.IDNumber as SpouseIDNumber,
		tcs.SurgateMotherName as SpouseSurgateMotherName,
		tcs.ProfessionID as SpouseProfessionID,
		ta.SpouseType,
		tcs.Gender as SpouseGender,
		tcs.BirthPlace as SpouseBirthPlace,
		tcs.BirthDate as SpouseBirthDate,
		tcs.MobilePhone as SpouseMobilePhone,
		tce2.VerificationWith,
		tce2.Relationship as EmconRelationship,
		tce2.EmconVerified,
		tca.Address,
		tce2.MobilePhone as EmconMobilePhone,
		tce2.VerifyBy,
		tce2.KnownCustomerAddress,
		tcp.StaySinceYear,
		tcp.StaySinceMonth,
		tce2.KnownCustomerJob,
		CASE tce.ProfessionID 
			WHEN 'WRST' THEN 'WIRASWASTA'
			WHEN 'PRO' THEN 'PROFESSIONAL'
			WHEN 'KRYSW' THEN 'KARYAWAN SWASTA'
			WHEN 'PNS' THEN 'PEGAWAI NEGERI SIPIL'
			WHEN 'ANG' THEN 'ANGKATAN'
			ELSE '-'
		END as Job,
		tce.EmploymentSinceYear,
		tce.EmploymentSinceMonth,
		tce.IndustryTypeID,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		CASE ti.bpkb_name
			WHEN 'K' THEN 'NAMA SAMA'
			WHEN 'P' THEN 'NAMA SAMA'
			WHEN 'KK' THEN 'NAMA BEDA'
			WHEN 'O' THEN 'NAMA BEDA'
			ELSE '-'
		END as BpkbName,
		tdb.plafon as Plafond,
		tdb.baki_debet_non_collateral as BakiDebet,
		tdb.fasilitas_aktif as FasilitasAktif,
		CASE
			WHEN tdb.kualitas_kredit_terburuk IS NOT NULL THEN CONCAT(tdb.kualitas_kredit_terburuk,' ',tdb.bulan_kualitas_terburuk)
			ELSE NULL
		END as ColTerburuk,
		tdb.baki_debet_kualitas_terburuk as BakiDebetTerburuk,
		CASE
			WHEN tdb.kualitas_kredit_terakhir IS NOT NULL THEN CONCAT(tdb.kualitas_kredit_terakhir,' ',tdb.bulan_kualitas_kredit_terakhir)
			ELSE NULL
		END as ColTerakhirAktif,
		tdb2.plafon as SpousePlafond,
		tdb2.baki_debet_non_collateral as SpouseBakiDebet,
		tdb2.fasilitas_aktif as SpouseFasilitasAktif,
		CASE
			WHEN tdb2.kualitas_kredit_terburuk IS NOT NULL THEN CONCAT(tdb2.kualitas_kredit_terburuk,' ',tdb2.bulan_kualitas_terburuk)
			ELSE NULL
		END as SpouseColTerburuk,
		tdb2.baki_debet_kualitas_terburuk as SpouseBakiDebetTerburuk,
		CASE
			WHEN tdb2.kualitas_kredit_terakhir IS NOT NULL THEN CONCAT(tdb2.kualitas_kredit_terakhir,' ',tdb2.bulan_kualitas_kredit_terakhir)
			ELSE NULL
		END as SpouseColTerakhirAktif,
		ta.ScsScore,
		ta.AgreementStatus,
		ta.TotalAgreementAktif,
		ta.MaxOVDAgreementAktif,
		ta.LastMaxOVDAgreement,
		tf.customer_segment,
		ta.LatestInstallment,
		ta2.NTFAkumulasi,
		(CAST(ISNULL(ta2.InstallmentAmount, 0) AS NUMERIC(17,2)) +
		CASE 
			WHEN ta.TotalDSR = ta.DSRFMF THEN 0
			WHEN ta.TotalDSR IS NULL THEN 0
			ELSE CAST(ISNULL(tf.total_installment_amount_biro, 0) AS NUMERIC(17,2))
		END +
		CAST(ISNULL(ta.InstallmentAmountFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountSpouseFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOther, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOtherSpouse, 0) AS NUMERIC(17,2)) -
		CAST(ISNULL(ta.InstallmentTopup, 0) AS NUMERIC(17,2)) ) as TotalInstallment,
		(CAST(ISNULL(tce.MonthlyFixedIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.MonthlyVariableIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.SpouseIncome, 0) AS NUMERIC(17,2)) ) as TotalIncome,
		CASE 
			WHEN ta.TotalDSR IS NULL THEN CAST(ISNULL(ta.DSRFMF, 0) AS NUMERIC(17,2)) + CAST(ISNULL(ta.DSRPBK, 0) AS NUMERIC(17,2))
			ELSE ta.TotalDSR
		END as TotalDSR,
		CASE
			WHEN ta.EkycSource IS NOT NULL THEN CONCAT(ta.EkycSource,' - ',ta.EkycReason)
			ELSE NULL
		END as EkycSource,
		ta.EkycSimiliarity,
		ta.EkycReason,
		CASE tia.info 
			WHEN '1' THEN 'APR'
			ELSE 'REJ'
		END cmo_decision,
		tia.name as cmo_name,
		tia.recom_date cmo_date,
		tcd.decision as ca_decision,
		tcd.note as ca_note,
		tcd.decision_by as ca_name,
		FORMAT(tcd.created_at,'yyyy-MM-dd') as ca_date,
		cbm.decision as cbm_decision,
		cbm.note as cbm_note,
		cbm.decision_by as cbm_name,
		FORMAT(cbm.created_at,'yyyy-MM-dd') as cbm_date,
		drm.decision as drm_decision,
		drm.note as drm_note,
		drm.decision_by as drm_name,
		FORMAT(drm.created_at,'yyyy-MM-dd') as drm_date,
		gmo.decision as gmo_decision,
		gmo.note as gmo_note,
		gmo.decision_by as gmo_name,
		FORMAT(gmo.created_at,'yyyy-MM-dd') as gmo_date
		FROM trx_status ts WITH (nolock)
		LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID 
		LEFT JOIN trx_customer_employment tce WITH (nolock) ON ts.ProspectID = tce.ProspectID
		LEFT JOIN trx_akkk ta WITH (nolock) ON ts.ProspectID = ta.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON ts.ProspectID = tcs.ProspectID
		LEFT JOIN trx_customer_emcon tce2 WITH (nolock) ON ts.ProspectID = tce2.ProspectID
		LEFT OUTER JOIN ( 
			SELECT scp.dbo.DEC_B64('SEC', Address) AS Address, ProspectID, Phone 
			FROM trx_customer_address WITH (nolock)
			WHERE Type = 'EMERGENCY' 
		) AS tca ON ts.ProspectID = tca.ProspectID 
		LEFT JOIN trx_item ti WITH (nolock) ON ts.ProspectID = ti.ProspectID
		LEFT OUTER JOIN ( 
			SELECT * FROM trx_detail_biro WITH (nolock)
			WHERE subject = 'CUSTOMER' 
		) AS tdb ON ts.ProspectID = tdb.prospect_id
		LEFT JOIN trx_filtering tf WITH (nolock) ON ts.ProspectID = tf.prospect_id 
		LEFT JOIN trx_apk ta2 WITH (nolock) ON ts.ProspectID = ta2.ProspectID
		LEFT OUTER JOIN ( 
			SELECT * FROM trx_detail_biro WITH (nolock)
			WHERE subject = 'SPOUSE' 
		) AS tdb2 ON ts.ProspectID = tdb2.prospect_id
		LEFT JOIN trx_ca_decision tcd WITH (nolock) ON ts.ProspectID = tcd.ProspectID
		LEFT JOIN trx_info_agent tia WITH (nolock) ON ts.ProspectID = tia.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas1 WITH (nolock)
			WHERE thas1.source_decision = 'CBM' AND thas1.created_at = (SELECT MAX(tha1.created_at) From trx_history_approval_scheme tha1 WHERE tha1.source_decision = thas1.source_decision AND tha1.ProspectID = thas1.ProspectID)
		) AS cbm ON ts.ProspectID = cbm.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas2 WITH (nolock)
			WHERE thas2.source_decision = 'DRM' AND thas2.created_at = (SELECT MAX(tha2.created_at) From trx_history_approval_scheme tha2 WHERE tha2.source_decision = thas2.source_decision AND tha2.ProspectID = thas2.ProspectID)
		) AS drm ON ts.ProspectID = drm.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas3 WITH (nolock)
			WHERE thas3.source_decision = 'GMO' AND thas3.created_at = (SELECT MAX(tha3.created_at) From trx_history_approval_scheme tha3 WHERE tha3.source_decision = thas3.source_decision AND tha3.ProspectID = thas3.ProspectID)
		) AS gmo ON ts.ProspectID = gmo.ProspectID
		WHERE ts.ProspectID = 'ppid'`)).
			WillReturnError(gorm.ErrRecordNotFound)

		// Call the function
		_, err := repo.GetAkkk("ppid")

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetInquiryApproval(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.InquiryCa{{ShowAction: false, ActionDate: "", ActionFormAkk: false, ActionEditData: false, AdditionalDP: 0, Activity: "", SourceDecision: "", StatusDecision: "", StatusReason: "", CaDecision: "", FinalApproval: "", CANote: "", ScsDate: "", ScsScore: "", ScsStatus: "", BiroCustomerResult: "", BiroSpouseResult: "", IsLastApproval: false, HasReturn: false, DraftDecision: "", DraftSlikResult: "", DraftNote: "", DraftCreatedAt: time.Time{}, DraftCreatedBy: "", DraftDecisionBy: "", ProspectID: "EFM03406412522151347", BranchName: "BANDUNG", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerID: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), SurveyResult: "", Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	t.Run("success with multi branch and without filter", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryApproval{
			Search:      "aprospectid",
			BranchID:    "426",
			MultiBranch: "1",
			UserID:      "abc123",
			Alias:       "CBM",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock) INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN ( SELECT value FROM region_user ru WITH (nolock) cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',') WHERE ru.user_id = 'abc123' ) AND b.lob_id='125'`)).WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','aprospectid') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID   	
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		 WHERE tm.BranchID = '426' AND (tcp.LegalName = 'xxxxxx') AND (has.next_step = 'CBM' OR has.source_decision='CBM')) AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as ca_decision,
		tcd.decision_by,
		tcd.final_approval,
		has.next_step,
		has.decision AS approval_decision,
		has.source_decision AS approval_source_decision,
		CASE
		  WHEN tcd.final_approval='CBM' THEN 1
		  ELSE 0
		END AS is_last_approval,
		CASE
		  WHEN rtn.decision IS NOT NULL THEN 1
		  ELSE 0
		END AS HasReturn,
		tcd.note AS ca_note,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tcd.decision = 'CAN' THEN tcd.created_at 
		  WHEN tcd.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE FORMAT(tst.created_at,'yyyy-MM-dd HH:mm:ss')
		END AS ActionDate,
		CASE
		  WHEN (tfa.decision IS NULL)
		  AND (tcd.decision <> 'CAN') 
		  AND (tst.source_decision='CBM') THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb.BiroSpouseResult,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6

	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (SELECT ProspectID, decision FROM trx_history_approval_scheme has WITH (nolock) WHERE has.decision = 'RTN') rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			note,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN (
			SELECT prospect_id, 
			MAX(Case [subject] When 'CUSTOMER' Then url_pdf_report End) BiroCustomerResult,
			MAX(Case [subject] When 'SPOUSE' Then url_pdf_report End) BiroSpouseResult
			FROM trx_detail_biro
			GROUP BY prospect_id
		) tdb ON tm.ProspectID = tdb.prospect_id 

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	 WHERE tm.BranchID = '426' AND (tcp.LegalName = 'xxxxxx') AND (has.next_step = 'CBM' OR has.source_decision='CBM')) AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryApproval(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with multi branch and need decision", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryApproval{
			Search:      "aprospectid",
			BranchID:    "426",
			MultiBranch: "1",
			Filter:      "NEED_DECISION",
			UserID:      "abc123",
			Alias:       "CBM",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock) INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN ( SELECT value FROM region_user ru WITH (nolock) cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',') WHERE ru.user_id = 'abc123' ) AND b.lob_id='125'`)).WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','aprospectid') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID   	
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		 WHERE tm.BranchID = '426' AND (tcp.LegalName = 'xxxxxx') AND tst.activity= 'UNPR' AND tst.decision= 'CPR' AND tst.source_decision = 'CBM') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as ca_decision,
		tcd.decision_by,
		tcd.final_approval,
		has.next_step,
		has.decision AS approval_decision,
		has.source_decision AS approval_source_decision,
		CASE
		  WHEN tcd.final_approval='CBM' THEN 1
		  ELSE 0
		END AS is_last_approval,
		CASE
		  WHEN rtn.decision IS NOT NULL THEN 1
		  ELSE 0
		END AS HasReturn,
		tcd.note AS ca_note,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tcd.decision = 'CAN' THEN tcd.created_at 
		  WHEN tcd.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE FORMAT(tst.created_at,'yyyy-MM-dd HH:mm:ss')
		END AS ActionDate,
		CASE
		  WHEN (tfa.decision IS NULL)
		  AND (tcd.decision <> 'CAN') 
		  AND (tst.source_decision='CBM') THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb.BiroSpouseResult,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6

	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (SELECT ProspectID, decision FROM trx_history_approval_scheme has WITH (nolock) WHERE has.decision = 'RTN') rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			note,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN (
			SELECT prospect_id, 
			MAX(Case [subject] When 'CUSTOMER' Then url_pdf_report End) BiroCustomerResult,
			MAX(Case [subject] When 'SPOUSE' Then url_pdf_report End) BiroSpouseResult
			FROM trx_detail_biro
			GROUP BY prospect_id
		) tdb ON tm.ProspectID = tdb.prospect_id 

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	 WHERE tm.BranchID = '426' AND (tcp.LegalName = 'xxxxxx') AND tst.activity= 'UNPR' AND tst.decision= 'CPR' AND tst.source_decision = 'CBM') AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryApproval(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success without multi branch", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryApproval{
			Search:      "NE-XXX",
			BranchID:    "426",
			MultiBranch: "0",
			Filter:      "REJECT",
			UserID:      "db1f4044e1dc574",
			Alias:       "CBM",
		}

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID   	
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		 WHERE tm.BranchID IN ('426') AND (tm.ProspectID = 'NE-XXX') AND tst.decision = 'REJ' AND tst.status_process='FIN' AND has.source_decision='CBM') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as ca_decision,
		tcd.decision_by,
		tcd.final_approval,
		has.next_step,
		has.decision AS approval_decision,
		has.source_decision AS approval_source_decision,
		CASE
		  WHEN tcd.final_approval='CBM' THEN 1
		  ELSE 0
		END AS is_last_approval,
		CASE
		  WHEN rtn.decision IS NOT NULL THEN 1
		  ELSE 0
		END AS HasReturn,
		tcd.note AS ca_note,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tcd.decision = 'CAN' THEN tcd.created_at 
		  WHEN tcd.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE FORMAT(tst.created_at,'yyyy-MM-dd HH:mm:ss')
		END AS ActionDate,
		CASE
		  WHEN (tfa.decision IS NULL)
		  AND (tcd.decision <> 'CAN') 
		  AND (tst.source_decision='CBM') THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb.BiroSpouseResult,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6

	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (SELECT ProspectID, decision FROM trx_history_approval_scheme has WITH (nolock) WHERE has.decision = 'RTN') rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			note,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN (
			SELECT prospect_id, 
			MAX(Case [subject] When 'CUSTOMER' Then url_pdf_report End) BiroCustomerResult,
			MAX(Case [subject] When 'SPOUSE' Then url_pdf_report End) BiroSpouseResult
			FROM trx_detail_biro
			GROUP BY prospect_id
		) tdb ON tm.ProspectID = tdb.prospect_id 

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
		  WHERE tm.BranchID IN ('426') AND (tm.ProspectID = 'NE-XXX') AND tst.decision = 'REJ' AND tst.status_process='FIN' AND has.source_decision='CBM') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryApproval(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with region west java", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryApproval{
			Search:      "76457",
			BranchID:    "426",
			MultiBranch: "1",
			Filter:      "CANCEL",
			UserID:      "db1f4044e1dc574",
			Alias:       "CBM",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'db1f4044e1dc574' 
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("WEST JAVA", `["426","436","429","431","442","428","430"]`))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','76457') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID   	
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		 WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.IDNumber = 'xxxxxx') AND tst.decision = 'CAN' AND tst.status_process='FIN' AND has.source_decision='CBM') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as ca_decision,
		tcd.decision_by,
		tcd.final_approval,
		has.next_step,
		has.decision AS approval_decision,
		has.source_decision AS approval_source_decision,
		CASE
		  WHEN tcd.final_approval='CBM' THEN 1
		  ELSE 0
		END AS is_last_approval,
		CASE
		  WHEN rtn.decision IS NOT NULL THEN 1
		  ELSE 0
		END AS HasReturn,
		tcd.note AS ca_note,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tcd.decision = 'CAN' THEN tcd.created_at 
		  WHEN tcd.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE FORMAT(tst.created_at,'yyyy-MM-dd HH:mm:ss')
		END AS ActionDate,
		CASE
		  WHEN (tfa.decision IS NULL)
		  AND (tcd.decision <> 'CAN') 
		  AND (tst.source_decision='CBM') THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb.BiroSpouseResult,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6

	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (SELECT ProspectID, decision FROM trx_history_approval_scheme has WITH (nolock) WHERE has.decision = 'RTN') rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			note,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN (
			SELECT prospect_id, 
			MAX(Case [subject] When 'CUSTOMER' Then url_pdf_report End) BiroCustomerResult,
			MAX(Case [subject] When 'SPOUSE' Then url_pdf_report End) BiroSpouseResult
			FROM trx_detail_biro
			GROUP BY prospect_id
		) tdb ON tm.ProspectID = tdb.prospect_id 

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	WHERE tm.BranchID IN ('426','436','429','431','442','428','430') AND (tcp.IDNumber = 'xxxxxx') AND tst.decision = 'CAN' AND tst.status_process='FIN' AND has.source_decision='CBM') AS tt  ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryApproval(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with region ALL", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryApproval{
			Search:      "aprospectid",
			BranchID:    "426",
			MultiBranch: "1",
			Filter:      "APPROVE",
			UserID:      "abc123",
			Alias:       "CBM",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
		INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
		(	SELECT value 
			FROM region_user ru WITH (nolock)
			cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
			WHERE ru.user_id = 'abc123' 
		)
		AND b.lob_id='125'`)).
			WillReturnRows(sqlmock.NewRows([]string{"region_name", "branch_member"}).
				AddRow("ALL", `["426","436","429","431","442","428","430"]`))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT SCP.dbo.ENC_B64('SEC','aprospectid') AS encrypt`)).WillReturnRows(sqlmock.NewRows([]string{"encrypt"}).AddRow("xxxxxx"))

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID   	
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		 WHERE (tcp.LegalName = 'xxxxxx') AND tst.decision = 'APR' AND tst.status_process='FIN' AND has.source_decision='CBM') AS tt`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as ca_decision,
		tcd.decision_by,
		tcd.final_approval,
		has.next_step,
		has.decision AS approval_decision,
		has.source_decision AS approval_source_decision,
		CASE
		  WHEN tcd.final_approval='CBM' THEN 1
		  ELSE 0
		END AS is_last_approval,
		CASE
		  WHEN rtn.decision IS NOT NULL THEN 1
		  ELSE 0
		END AS HasReturn,
		tcd.note AS ca_note,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tcd.decision = 'CAN' THEN tcd.created_at 
		  WHEN tcd.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE FORMAT(tst.created_at,'yyyy-MM-dd HH:mm:ss')
		END AS ActionDate,
		CASE
		  WHEN (tfa.decision IS NULL)
		  AND (tcd.decision <> 'CAN') 
		  AND (tst.source_decision='CBM') THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb.BiroSpouseResult,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6

	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = 'CBM'
				OR has.source_decision = 'CBM'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (SELECT ProspectID, decision FROM trx_history_approval_scheme has WITH (nolock) WHERE has.decision = 'RTN') rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			note,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN (
			SELECT prospect_id, 
			MAX(Case [subject] When 'CUSTOMER' Then url_pdf_report End) BiroCustomerResult,
			MAX(Case [subject] When 'SPOUSE' Then url_pdf_report End) BiroSpouseResult
			FROM trx_detail_biro
			GROUP BY prospect_id
		) tdb ON tm.ProspectID = tdb.prospect_id 

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	 WHERE (tcp.LegalName = 'xxxxxx') AND tst.decision = 'APR' AND tst.status_process='FIN' AND has.source_decision='CBM') AS tt ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).AddRow("EFM03406412522151347", "BANDUNG", "426"))

		// Call the function
		reason, _, err := repo.GetInquiryApproval(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestProcessRecalculateOrder(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	ppid := "TST001"

	trxStatus := entity.TrxStatus{
		ProspectID:     ppid,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_UNPROCESS,
		Decision:       constant.DB_DECISION_CREDIT_PROCESS,
		SourceDecision: constant.NEED_RECALCULATE,
		Reason:         constant.REASON_NEED_RECALCULATE,
	}

	infoMap := map[string]float64{
		"dp_amount": 1000.45,
	}
	info, _ := json.Marshal(infoMap)

	trxDetail := entity.TrxDetail{
		ProspectID:     ppid,
		StatusProcess:  constant.STATUS_ONPROCESS,
		Activity:       constant.ACTIVITY_PROCESS,
		Decision:       constant.DB_DECISION_PASS,
		RuleCode:       constant.CODE_CREDIT_COMMITTEE,
		SourceDecision: constant.DB_DECISION_CREDIT_ANALYST,
		NextStep:       constant.NEED_RECALCULATE,
		Info:           string(info),
		CreatedBy:      "abc123",
		Reason:         constant.REASON_NEED_RECALCULATE,
	}

	trxHistoryApproval := entity.TrxHistoryApprovalScheme{
		ProspectID:            ppid,
		Decision:              constant.DB_DECISION_SDP,
		Reason:                trxStatus.Reason,
		Note:                  fmt.Sprintf("Nilai DP: %.0f", 1000.45),
		CreatedBy:             "abc123",
		DecisionBy:            "CBM KMB",
		NeedEscalation:        0,
		NextFinalApprovalFlag: 1,
		SourceDecision:        trxDetail.SourceDecision,
	}

	t.Run("success update", func(t *testing.T) {

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_ca_decision" SET "created_at" = ? WHERE (ProspectID = ?)`)).
			WithArgs(sqlmock.AnyArg(), ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxHistoryApproval.ProspectID, trxHistoryApproval.Decision, trxHistoryApproval.Reason, trxHistoryApproval.Note, sqlmock.AnyArg(), trxHistoryApproval.CreatedBy, trxHistoryApproval.DecisionBy, trxHistoryApproval.NeedEscalation, trxHistoryApproval.NextFinalApprovalFlag, trxHistoryApproval.SourceDecision, trxHistoryApproval.NextStep).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := newDB.ProcessRecalculateOrder(ppid, trxStatus, trxDetail, trxHistoryApproval)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})
}

func TestSubmitApproval(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("CLIENT_LOS", "2ck21b02")
	os.Setenv("AUTH_LOS", "xYtKHAWHn2sZLm1IbXXu")
	os.Setenv("INSERT_STAGING_URL", "http://10.9.100.131/los-kmb-api/api/v3/kmb/insert-staging")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	var (
		decision        string
		decision_detail string
		trxRecalculate  entity.TrxRecalculate
		approvalScheme  response.RespApprovalScheme
	)
	rej := constant.DB_DECISION_REJECT
	apr := constant.DB_DECISION_APR
	pas := constant.DB_DECISION_PASS
	rtn := constant.DB_DECISION_RTN
	cpr := constant.DB_DECISION_CREDIT_PROCESS
	cra := constant.DB_DECISION_CREDIT_ANALYST
	onp := constant.STATUS_ONPROCESS
	unpr := constant.ACTIVITY_UNPROCESS
	prcd := constant.ACTIVITY_PROCESS

	t.Run("success approve from cbm to drm", func(t *testing.T) {
		req := request.ReqSubmitApproval{
			ProspectID:    "ppid",
			FinalApproval: "DRM",
			Decision:      "APPROVE",
			RuleCode:      "3741",
			Alias:         "CBM",
			Reason:        "Oke",
			Note:          "lanjut ke drm",
			CreatedBy:     "abc123",
			DecisionBy:    "BM KMB",
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = rej
			decision_detail = rej

		case constant.DECISION_APPROVE:
			decision = apr
			decision_detail = pas

		case constant.DECISION_RETURN:
			decision = rtn
			decision_detail = rtn
		}

		approvalScheme, _ = utils.ApprovalScheme(req)

		trxStatus := entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       unpr,
			Decision:       cpr,
			RuleCode:       req.RuleCode,
			SourceDecision: approvalScheme.NextStep,
			Reason:         req.Reason,
		}

		trxDetail := entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       prcd,
			Decision:       decision_detail,
			RuleCode:       req.RuleCode,
			SourceDecision: req.Alias,
			Info:           req.Reason,
			CreatedBy:      req.CreatedBy,
			Reason:         req.Reason,
		}

		trxHistoryApproval := entity.TrxHistoryApprovalScheme{
			ProspectID:            req.ProspectID,
			Decision:              decision,
			Reason:                req.Reason,
			Note:                  req.Note,
			CreatedBy:             req.CreatedBy,
			DecisionBy:            req.DecisionBy,
			NextFinalApprovalFlag: 1,
			NeedEscalation:        0,
			SourceDecision:        trxDetail.SourceDecision,
			NextStep:              approvalScheme.NextStep,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT ts.ProspectID, 
			CASE 
				WHEN td.ProspectID IS NOT NULL AND tcp.CustomerStatus = 'NEW' THEN 'DEV'
				ELSE NULL
			END AS activity 
			FROM trx_status ts
			LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID
			LEFT JOIN trx_deviasi td ON ts.ProspectID = td.ProspectID 
			WHERE ts.ProspectID = 'ppid' AND ts.status_process = 'ONP'`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "activity"}).
				AddRow("ppid", ""))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxHistoryApproval.ProspectID, trxHistoryApproval.Decision, trxHistoryApproval.Reason, trxHistoryApproval.Note, sqlmock.AnyArg(), trxHistoryApproval.CreatedBy, trxHistoryApproval.DecisionBy, trxHistoryApproval.NeedEscalation, sqlmock.AnyArg(), trxHistoryApproval.SourceDecision, trxHistoryApproval.NextStep).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		trxStatus, err := newDB.SubmitApproval(req, trxStatus, trxDetail, trxRecalculate, approvalScheme)
		if err != nil {
			t.Errorf("error '%s'", err.Error())
		}
	})

	t.Run("success submit rej cbm to drm escalation to gmo", func(t *testing.T) {
		req := request.ReqSubmitApproval{
			ProspectID:     "ppid",
			FinalApproval:  "DRM",
			Decision:       "REJECT",
			NeedEscalation: true,
			RuleCode:       "3747",
			Alias:          "DRM",
			Reason:         "Oke",
			Note:           "eskalasi ke gmo",
			CreatedBy:      "abc123",
			DecisionBy:     "RM KMB",
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = rej
			decision_detail = rej

		case constant.DECISION_APPROVE:
			decision = apr
			decision_detail = pas

		case constant.DECISION_RETURN:
			decision = rtn
			decision_detail = rtn
		}

		approvalScheme, _ = utils.ApprovalScheme(req)

		trxStatus := entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       unpr,
			Decision:       cpr,
			RuleCode:       req.RuleCode,
			SourceDecision: approvalScheme.NextStep,
			Reason:         req.Reason,
		}

		trxDetail := entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       prcd,
			Decision:       decision_detail,
			RuleCode:       req.RuleCode,
			SourceDecision: req.Alias,
			Info:           req.Reason,
			CreatedBy:      req.CreatedBy,
			Reason:         req.Reason,
		}

		trxHistoryApproval := entity.TrxHistoryApprovalScheme{
			ProspectID:     req.ProspectID,
			Decision:       decision,
			Reason:         req.Reason,
			Note:           req.Note,
			CreatedBy:      req.CreatedBy,
			DecisionBy:     req.DecisionBy,
			SourceDecision: trxDetail.SourceDecision,
			NextStep:       approvalScheme.NextStep,
		}

		trxCaDecision := entity.TrxCaDecision{
			FinalApproval: approvalScheme.NextStep,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT ts.ProspectID, 
			CASE 
				WHEN td.ProspectID IS NOT NULL AND tcp.CustomerStatus = 'NEW' THEN 'DEV'
				ELSE NULL
			END AS activity 
			FROM trx_status ts
			LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID
			LEFT JOIN trx_deviasi td ON ts.ProspectID = td.ProspectID 
			WHERE ts.ProspectID = 'ppid' AND ts.status_process = 'ONP'`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "activity"}).
				AddRow("ppid", ""))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_ca_decision" SET "final_approval" = ? WHERE (ProspectID = ?)`)).
			WithArgs(trxCaDecision.FinalApproval, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxHistoryApproval.ProspectID, trxHistoryApproval.Decision, trxHistoryApproval.Reason, trxHistoryApproval.Note, sqlmock.AnyArg(), trxHistoryApproval.CreatedBy, trxHistoryApproval.DecisionBy, sqlmock.AnyArg(), sqlmock.AnyArg(), trxHistoryApproval.SourceDecision, trxHistoryApproval.NextStep).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		trxStatus, err := newDB.SubmitApproval(req, trxStatus, trxDetail, trxRecalculate, approvalScheme)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("success submit return approval to ca", func(t *testing.T) {
		req := request.ReqSubmitApproval{
			ProspectID:     "ppid",
			FinalApproval:  "GMO",
			Decision:       "RETURN",
			NeedEscalation: false,
			RuleCode:       "3750",
			Alias:          "GMO",
			Reason:         "Final Approval Return",
			Note:           "RETURN",
			CreatedBy:      "abc123",
			DecisionBy:     "GMO KMB",
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = rej
			decision_detail = rej

		case constant.DECISION_APPROVE:
			decision = apr
			decision_detail = pas

		case constant.DECISION_RETURN:
			decision = rtn
			decision_detail = rtn
		}

		approvalScheme, _ = utils.ApprovalScheme(req)

		trxStatus := entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       unpr,
			Decision:       cpr,
			RuleCode:       req.RuleCode,
			SourceDecision: approvalScheme.NextStep,
			Reason:         req.Reason,
		}

		trxDetail := entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       prcd,
			Decision:       decision_detail,
			RuleCode:       req.RuleCode,
			SourceDecision: req.Alias,
			Info:           req.Reason,
			CreatedBy:      req.CreatedBy,
			Reason:         req.Reason,
		}

		trxHistoryApproval := entity.TrxHistoryApprovalScheme{
			ProspectID:     req.ProspectID,
			Decision:       decision,
			Reason:         req.Reason,
			Note:           req.Note,
			CreatedBy:      req.CreatedBy,
			DecisionBy:     req.DecisionBy,
			SourceDecision: trxDetail.SourceDecision,
			NextStep:       cra,
		}

		trxRecalculate := entity.TrxRecalculate{
			ProspectID:   req.ProspectID,
			AdditionalDP: req.DPAmount,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT ts.ProspectID, 
			CASE 
				WHEN td.ProspectID IS NOT NULL AND tcp.CustomerStatus = 'NEW' THEN 'DEV'
				ELSE NULL
			END AS activity 
			FROM trx_status ts
			LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID
			LEFT JOIN trx_deviasi td ON ts.ProspectID = td.ProspectID 
			WHERE ts.ProspectID = 'ppid' AND ts.status_process = 'ONP'`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "activity"}).
				AddRow("ppid", ""))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_recalculate" ("ProspectID","ProductOfferingID","product_offering_desc","Tenor","loan_amount","AF","InstallmentAmount","DPAmount","percent_dp","AdminFee","provision_fee","fidusia_fee","AssetInsuranceFee","LifeInsuranceFee","NTF","NTFAkumulasi","interest_rate","interest_amount","additional_dp","DSRFMF","TotalDSR","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxRecalculate.ProspectID, trxRecalculate.ProductOfferingID, trxRecalculate.ProductOfferingDesc, trxRecalculate.Tenor, trxRecalculate.LoanAmount, trxRecalculate.AF, trxRecalculate.InstallmentAmount, trxRecalculate.DPAmount, trxRecalculate.PercentDP, trxRecalculate.AdminFee, trxRecalculate.ProvisionFee, trxRecalculate.FidusiaFee, trxRecalculate.AssetInsuranceFee, trxRecalculate.LifeInsuranceFee, trxRecalculate.NTF, trxRecalculate.NTFAkumulasi, trxRecalculate.InterestRate, trxRecalculate.InterestAmount, trxRecalculate.AdditionalDP, trxRecalculate.DSRFMF, trxRecalculate.TotalDSR, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxHistoryApproval.ProspectID, trxHistoryApproval.Decision, trxHistoryApproval.Reason, trxHistoryApproval.Note, sqlmock.AnyArg(), trxHistoryApproval.CreatedBy, trxHistoryApproval.DecisionBy, sqlmock.AnyArg(), sqlmock.AnyArg(), trxHistoryApproval.SourceDecision, trxHistoryApproval.NextStep).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		trxStatus, err := newDB.SubmitApproval(req, trxStatus, trxDetail, trxRecalculate, approvalScheme)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("success submit apr final approval", func(t *testing.T) {
		os.Setenv("CLIENT_LOS", "2ck21b02")
		os.Setenv("AUTH_LOS", "xYtKHAWHn2sZLm1IbXXu")
		os.Setenv("INSERT_STAGING_URL", "http://10.9.100.131/los-kmb-api/api/v3/kmb/insert-staging")

		req := request.ReqSubmitApproval{
			ProspectID:     "ppid",
			FinalApproval:  "GMO",
			Decision:       "APPROVE",
			NeedEscalation: false,
			RuleCode:       "3750",
			Alias:          "GMO",
			Reason:         "Oke",
			Note:           "final di gmo",
			CreatedBy:      "abc123",
			DecisionBy:     "GMO KMB",
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = rej
			decision_detail = rej

		case constant.DECISION_APPROVE:
			decision = apr
			decision_detail = pas

		case constant.DECISION_RETURN:
			decision = rtn
			decision_detail = rtn
		}

		approvalScheme, _ = utils.ApprovalScheme(req)

		trxStatus := entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       unpr,
			Decision:       cpr,
			RuleCode:       req.RuleCode,
			SourceDecision: approvalScheme.NextStep,
			Reason:         req.Reason,
		}

		trxDetail := entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       prcd,
			Decision:       decision_detail,
			RuleCode:       req.RuleCode,
			SourceDecision: req.Alias,
			Info:           req.Reason,
			CreatedBy:      req.CreatedBy,
			Reason:         req.Reason,
		}

		trxHistoryApproval := entity.TrxHistoryApprovalScheme{
			ProspectID:     req.ProspectID,
			Decision:       decision,
			Reason:         req.Reason,
			Note:           req.Note,
			CreatedBy:      req.CreatedBy,
			DecisionBy:     req.DecisionBy,
			SourceDecision: trxDetail.SourceDecision,
			NextStep:       approvalScheme.NextStep,
		}

		trxFinalApproval := entity.TrxFinalApproval{
			ProspectID: req.ProspectID,
			Decision:   decision,
			Reason:     req.Reason,
			Note:       req.Note,
			CreatedBy:  req.CreatedBy,
			DecisionBy: req.DecisionBy,
		}

		getAFPhone := entity.AFMobilePhone{
			AFValue:     1000,
			MobilePhone: "08989898",
		}
		trxAgreement := entity.TrxAgreement{
			ProspectID:         req.ProspectID,
			CheckingStatus:     constant.ACTIVITY_UNPROCESS,
			ContractStatus:     "0",
			AF:                 getAFPhone.AFValue,
			MobilePhone:        getAFPhone.MobilePhone,
			CustomerIDKreditmu: constant.LOB_NEW_KMB,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT ts.ProspectID, 
			CASE 
				WHEN td.ProspectID IS NOT NULL AND tcp.CustomerStatus = 'NEW' THEN 'DEV'
				ELSE NULL
			END AS activity 
			FROM trx_status ts
			LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID
			LEFT JOIN trx_deviasi td ON ts.ProspectID = td.ProspectID 
			WHERE ts.ProspectID = 'ppid' AND ts.status_process = 'ONP'`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "activity"}).
				AddRow("ppid", ""))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxHistoryApproval.ProspectID, trxHistoryApproval.Decision, trxHistoryApproval.Reason, trxHistoryApproval.Note, sqlmock.AnyArg(), trxHistoryApproval.CreatedBy, trxHistoryApproval.DecisionBy, sqlmock.AnyArg(), sqlmock.AnyArg(), trxHistoryApproval.SourceDecision, trxHistoryApproval.NextStep).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_final_approval" ("ProspectID","decision","reason","note","created_at","created_by","decision_by") VALUES (?,?,?,?,?,?,?)`)).
			WithArgs(trxFinalApproval.ProspectID, trxFinalApproval.Decision, trxFinalApproval.Reason, trxFinalApproval.Note, sqlmock.AnyArg(), trxFinalApproval.CreatedBy, trxFinalApproval.DecisionBy).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT AF, SCP.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone, OTR, DPAmount FROM trx_apk apk WITH (nolock) INNER JOIN trx_customer_personal tcp WITH (nolock) ON apk.ProspectID = tcp.ProspectID WHERE apk.ProspectID =`)).
			WillReturnRows(sqlmock.NewRows([]string{"AF", "MobilePhone", "OTR", "DPAmount"}).
				AddRow(1000, "08989898", 10000, 1000))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_agreements" ("ProspectID","BranchID","CustomerID","ApplicationID","AgreementNo","AgreementDate","NextInstallmentDate","MaturityDate","ContractStatus","NewApplicationDate","ApprovalDate","PurchaseOrderDate","GoLiveDate","ProductID","ProductOfferingID","TotalOTR","DownPayment","NTF","PayToDealerAmount","PayToDealerDate","checking_status","last_checking_at","created_at","updated_at","AF","MobilePhone","customer_id_kreditmu") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxAgreement.ProspectID, trxAgreement.BranchID, trxAgreement.CustomerID, trxAgreement.ApplicationID, trxAgreement.AgreementNo, trxAgreement.AgreementDate, trxAgreement.NextInstallmentDate, trxAgreement.MaturityDate, trxAgreement.ContractStatus, trxAgreement.NewApplicationDate, trxAgreement.ApprovalDate, trxAgreement.PurchaseOrderDate, trxAgreement.GoLiveDate, trxAgreement.ProductID, trxAgreement.ProductOfferingID, trxAgreement.TotalOTR, trxAgreement.DownPayment, trxAgreement.NTF, trxAgreement.PayToDealerAmount, trxAgreement.PayToDealerDate, trxAgreement.CheckingStatus, trxAgreement.LastCheckingAt, sqlmock.AnyArg(), sqlmock.AnyArg(), trxAgreement.AF, trxAgreement.MobilePhone, trxAgreement.CustomerIDKreditmu).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_worker" ("ProspectID","activity","endpoint_target","endpoint_method","payload","header","response_timeout","api_type","max_retry","count_retry","created_at","category","action","status_code","sequence") VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs("ppid", "UNPR", "http://10.9.100.131/los-kmb-api/api/v3/kmb/insert-staging/ppid", "POST", sqlmock.AnyArg(), sqlmock.AnyArg(), 30, "RAW", 6, 0, sqlmock.AnyArg(), "CONFINS", "INSERT_STAGING_KMB", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		mock.ExpectCommit()

		trxStatus, err := newDB.SubmitApproval(req, trxStatus, trxDetail, trxRecalculate, approvalScheme)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("success submit rej final approval", func(t *testing.T) {
		os.Setenv("CLIENT_LOS", "2ck21b02")
		os.Setenv("AUTH_LOS", "xYtKHAWHn2sZLm1IbXXu")
		os.Setenv("INSERT_STAGING_URL", "http://10.9.100.131/los-kmb-api/api/v3/kmb/insert-staging")

		req := request.ReqSubmitApproval{
			ProspectID:     "ppid",
			FinalApproval:  "GMO",
			Decision:       "REJECT",
			NeedEscalation: false,
			RuleCode:       "3750",
			Alias:          "GMO",
			Reason:         "Oke",
			Note:           "final di gmo",
			CreatedBy:      "abc123",
			DecisionBy:     "GMO KMB",
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = rej
			decision_detail = rej

		case constant.DECISION_APPROVE:
			decision = apr
			decision_detail = pas

		case constant.DECISION_RETURN:
			decision = rtn
			decision_detail = rtn
		}

		approvalScheme, _ = utils.ApprovalScheme(req)

		trxStatus := entity.TrxStatus{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       unpr,
			Decision:       cpr,
			RuleCode:       req.RuleCode,
			SourceDecision: approvalScheme.NextStep,
			Reason:         req.Reason,
		}

		trxDetail := entity.TrxDetail{
			ProspectID:     req.ProspectID,
			StatusProcess:  onp,
			Activity:       prcd,
			Decision:       decision_detail,
			RuleCode:       req.RuleCode,
			SourceDecision: req.Alias,
			Info:           req.Reason,
			CreatedBy:      req.CreatedBy,
			Reason:         req.Reason,
		}

		trxHistoryApproval := entity.TrxHistoryApprovalScheme{
			ProspectID:     req.ProspectID,
			Decision:       decision,
			Reason:         req.Reason,
			Note:           req.Note,
			CreatedBy:      req.CreatedBy,
			DecisionBy:     req.DecisionBy,
			SourceDecision: trxDetail.SourceDecision,
			NextStep:       approvalScheme.NextStep,
		}

		trxFinalApproval := entity.TrxFinalApproval{
			ProspectID: req.ProspectID,
			Decision:   decision,
			Reason:     req.Reason,
			Note:       req.Note,
			CreatedBy:  req.CreatedBy,
			DecisionBy: req.DecisionBy,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT ts.ProspectID, 
			CASE 
				WHEN td.ProspectID IS NOT NULL AND tcp.CustomerStatus = 'NEW' THEN 'DEV'
				ELSE NULL
			END AS activity 
			FROM trx_status ts
			LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID
			LEFT JOIN trx_deviasi td ON ts.ProspectID = td.ProspectID 
			WHERE ts.ProspectID = 'ppid' AND ts.status_process = 'ONP'`)).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "activity"}).
				AddRow("ppid", ""))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","reason","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.Reason, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxHistoryApproval.ProspectID, trxHistoryApproval.Decision, trxHistoryApproval.Reason, trxHistoryApproval.Note, sqlmock.AnyArg(), trxHistoryApproval.CreatedBy, trxHistoryApproval.DecisionBy, sqlmock.AnyArg(), sqlmock.AnyArg(), trxHistoryApproval.SourceDecision, trxHistoryApproval.NextStep).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_final_approval" ("ProspectID","decision","reason","note","created_at","created_by","decision_by") VALUES (?,?,?,?,?,?,?)`)).
			WithArgs(trxFinalApproval.ProspectID, trxFinalApproval.Decision, trxFinalApproval.Reason, trxFinalApproval.Note, sqlmock.AnyArg(), trxFinalApproval.CreatedBy, trxFinalApproval.DecisionBy).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		trxStatus, err := newDB.SubmitApproval(req, trxStatus, trxDetail, trxRecalculate, approvalScheme)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

}

func TestGetQuotaDeviasiBranch(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedBranches := []entity.ConfinsBranch{
		{
			BranchID:   "100",
			BranchName: "JAKARTA",
		},
		{
			BranchID:   "200",
			BranchName: "SURABAYA",
		},
	}

	t.Run("success with filter", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListQuotaDeviasiBranch{
			BranchID:   "100,200",
			BranchName: "JAKARTA",
		}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT mbd.BranchID, cb.BranchName
			FROM m_branch_deviasi AS mbd WITH (nolock)
			JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            WHERE mbd.BranchID IN ('100','200') AND cb.BranchName LIKE '%JAKARTA%'
            ORDER BY cb.BranchName ASC`)).
			WillReturnRows(sqlmock.NewRows([]string{"BranchID", "BranchName"}).
				AddRow("100", "JAKARTA").
				AddRow("200", "SURABAYA"))

		mock.ExpectCommit()

		// Call the function
		branches, err := repo.GetQuotaDeviasiBranch(req)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedBranches, branches, "Expected branches slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListQuotaDeviasiBranch{}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT mbd.BranchID, cb.BranchName
			FROM m_branch_deviasi AS mbd WITH (nolock)
			JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
			ORDER BY cb.BranchName ASC`)).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectCommit()

		// Call the function
		_, err := repo.GetQuotaDeviasiBranch(req)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetInquiryQuotaDeviasi(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedData := []entity.InquirySettingQuotaDeviasi{
		{
			BranchID:     "400",
			BranchName:   "BEKASI",
			QuotaAmount:  100,
			QuotaAccount: 50,
			IsActive:     true,
		},
	}

	t.Run("success with search", func(t *testing.T) {
		req := request.ReqListQuotaDeviasi{
			Search: "abranchname",
		}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
                COUNT(*) AS totalRow
            FROM (
                SELECT mbd.*, cb.BranchName AS branch_name
                FROM m_branch_deviasi AS mbd WITH (nolock)
                JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
                WHERE (mbd.BranchID LIKE '%abranchname%' OR cb.BranchName LIKE '%abranchname%')
            ) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT mbd.BranchID, cb.BranchName AS branch_name, mbd.quota_amount, mbd.quota_account, mbd.booking_amount, mbd.booking_account, mbd.balance_amount, mbd.balance_account, mbd.is_active, mbd.updated_by, ISNULL(FORMAT(mbd.updated_at, 'yyyy-MM-dd HH:mm:ss'), '') AS updated_at
            FROM m_branch_deviasi AS mbd WITH (nolock)
            JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            WHERE (mbd.BranchID LIKE '%abranchname%' OR cb.BranchName LIKE '%abranchname%') ORDER BY mbd.is_active DESC, mbd.BranchID ASC OFFSET 0 ROWS FETCH FIRST 10 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"BranchID", "branch_name", "quota_amount", "quota_account", "is_active"}).AddRow("400", "BEKASI", 100, 50, true))

		mock.ExpectCommit()

		data, _, err := repo.GetInquiryQuotaDeviasi(req, request.RequestPagination{Page: 1, Limit: 10})

		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with filters", func(t *testing.T) {
		req := request.ReqListQuotaDeviasi{
			BranchID: "400",
			IsActive: "1",
		}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
                COUNT(*) AS totalRow
            FROM (
                SELECT mbd.*, cb.BranchName AS branch_name
                FROM m_branch_deviasi AS mbd WITH (nolock)
                JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
                WHERE mbd.BranchID IN ('400') AND mbd.is_active = '1'
            ) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT mbd.BranchID, cb.BranchName AS branch_name, mbd.quota_amount, mbd.quota_account, mbd.booking_amount, mbd.booking_account, mbd.balance_amount, mbd.balance_account, mbd.is_active, mbd.updated_by, ISNULL(FORMAT(mbd.updated_at, 'yyyy-MM-dd HH:mm:ss'), '') AS updated_at
            FROM m_branch_deviasi AS mbd WITH (nolock)
            JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            WHERE mbd.BranchID IN ('400') AND mbd.is_active = '1' ORDER BY mbd.is_active DESC, mbd.BranchID ASC OFFSET 0 ROWS FETCH FIRST 10 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"BranchID", "branch_name", "quota_amount", "quota_account", "is_active"}).AddRow("400", "BEKASI", 100, 50, "1"))

		mock.ExpectCommit()

		data, _, err := repo.GetInquiryQuotaDeviasi(req, request.RequestPagination{Page: 1, Limit: 10})

		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		req := request.ReqListQuotaDeviasi{}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
                COUNT(*) AS totalRow
            FROM (
                SELECT mbd.*, cb.BranchName AS branch_name
                FROM m_branch_deviasi AS mbd WITH (nolock)
                JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            ) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("0"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT mbd.BranchID, cb.BranchName AS branch_name, mbd.quota_amount, mbd.quota_account, mbd.booking_amount, mbd.booking_account, mbd.balance_amount, mbd.balance_account, mbd.is_active, mbd.updated_by, ISNULL(FORMAT(mbd.updated_at, 'yyyy-MM-dd HH:mm:ss'), '') AS updated_at
            FROM m_branch_deviasi AS mbd WITH (nolock)
            JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            ORDER BY mbd.is_active DESC, mbd.BranchID ASC OFFSET 0 ROWS FETCH FIRST 10 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"BranchID", "branch_name", "quota_amount", "quota_account", "is_active"}))

		mock.ExpectCommit()

		_, _, err := repo.GetInquiryQuotaDeviasi(req, request.RequestPagination{Page: 1, Limit: 10})

		expectedErr := fmt.Errorf(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error get count data", func(t *testing.T) {
		req := request.ReqListQuotaDeviasi{}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
                COUNT(*) AS totalRow
            FROM (
                SELECT mbd.*, cb.BranchName AS branch_name
                FROM m_branch_deviasi AS mbd WITH (nolock)
                JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            ) AS y`)).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectCommit()

		_, _, err := repo.GetInquiryQuotaDeviasi(req, request.RequestPagination{Page: 1, Limit: 10})

		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error get data", func(t *testing.T) {
		req := request.ReqListQuotaDeviasi{}

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
                COUNT(*) AS totalRow
            FROM (
                SELECT mbd.*, cb.BranchName AS branch_name
                FROM m_branch_deviasi AS mbd WITH (nolock)
                JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            ) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT mbd.BranchID, cb.BranchName AS branch_name, mbd.quota_amount, mbd.quota_account, mbd.booking_amount, mbd.booking_account, mbd.balance_amount, mbd.balance_account, mbd.is_active, mbd.updated_by, ISNULL(FORMAT(mbd.updated_at, 'yyyy-MM-dd HH:mm:ss'), '') AS updated_at
            FROM m_branch_deviasi AS mbd WITH (nolock)
            JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID)
            ORDER BY mbd.is_active DESC, mbd.BranchID ASC OFFSET 0 ROWS FETCH FIRST 10 ROWS ONLY`)).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectCommit()

		_, _, err := repo.GetInquiryQuotaDeviasi(req, request.RequestPagination{Page: 1, Limit: 10})

		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestProcessUpdateQuotaDeviasiBranch(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	branchID := "BR001"
	mBranchDeviasi := entity.MappingBranchDeviasi{
		BranchID:     branchID,
		QuotaAmount:  10000,
		QuotaAccount: 5,
		IsActive:     true,
		UpdatedBy:    "tester",
	}

	t.Run("success update", func(t *testing.T) {
		dataBefore := entity.DataQuotaDeviasiBranch{
			QuotaAmount:    8000,
			QuotaAccount:   3,
			BookingAmount:  2000,
			BookingAccount: 2,
			BalanceAmount:  6000,
			BalanceAccount: 1,
			IsActive:       true,
			UpdatedAt:      time.Now(),
			UpdatedBy:      "tester",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?`)).
			WithArgs(branchID).
			WillReturnRows(sqlmock.NewRows([]string{"quota_amount", "quota_account", "booking_amount", "booking_account", "balance_amount", "balance_account", "is_active", "updated_at", "updated_by"}).
				AddRow(dataBefore.QuotaAmount, dataBefore.QuotaAccount, dataBefore.BookingAmount, dataBefore.BookingAccount, 6000, 1, true, time.Now(), "tester"))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_branch_deviasi" SET "balance_account" = ?, "balance_amount" = ?, "is_active" = ?, "quota_account" = ?, "quota_amount" = ?, "updated_at" = ?, "updated_by" = ? WHERE (BranchID = ?)`)).
			WithArgs(3, 8000.00, mBranchDeviasi.IsActive, mBranchDeviasi.QuotaAccount, mBranchDeviasi.QuotaAmount, sqlmock.AnyArg(), mBranchDeviasi.UpdatedBy, branchID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?`)).
			WithArgs(branchID).
			WillReturnRows(sqlmock.NewRows([]string{"quota_amount", "quota_account", "booking_amount", "booking_account", "balance_amount", "balance_account", "is_active", "updated_at", "updated_by"}).
				AddRow(dataBefore.QuotaAmount, dataBefore.QuotaAccount, dataBefore.BookingAmount, dataBefore.BookingAccount, 8000, 3, true, time.Now(), "tester"))

		mock.ExpectCommit()

		dataBeforeResult, dataAfterResult, err := repo.ProcessUpdateQuotaDeviasiBranch(branchID, mBranchDeviasi)
		dataBeforeResult.UpdatedAt = dataBefore.UpdatedAt

		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if dataBeforeResult != dataBefore {
			t.Errorf("expected dataBefore: %v, got: %v", dataBefore, dataBeforeResult)
		}
		if dataAfterResult.BalanceAmount != mBranchDeviasi.QuotaAmount-dataBefore.BookingAmount {
			t.Errorf("expected balance amount: %v, got: %v", mBranchDeviasi.QuotaAmount-dataBefore.BookingAmount, dataAfterResult.BalanceAmount)
		}
		if dataAfterResult.BalanceAccount != mBranchDeviasi.QuotaAccount-dataBefore.BookingAccount {
			t.Errorf("expected balance account: %v, got: %v", mBranchDeviasi.QuotaAccount-dataBefore.BookingAccount, dataAfterResult.BalanceAccount)
		}
	})

	t.Run("error when booking amount exceeds quota", func(t *testing.T) {
		dataBefore := entity.DataQuotaDeviasiBranch{
			QuotaAmount:    8000,
			QuotaAccount:   3,
			BookingAmount:  11000, // Exceeds new quota
			BookingAccount: 2,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?`)).
			WithArgs(branchID).
			WillReturnRows(sqlmock.NewRows([]string{"quota_amount", "quota_account", "booking_amount", "booking_account", "balance_amount", "balance_account", "is_active", "updated_at", "updated_by"}).
				AddRow(dataBefore.QuotaAmount, dataBefore.QuotaAccount, dataBefore.BookingAmount, dataBefore.BookingAccount, 6000, 3, true, time.Now(), "tester"))

		mock.ExpectRollback()

		_, _, err := repo.ProcessUpdateQuotaDeviasiBranch(branchID, mBranchDeviasi)
		if err == nil || err.Error() != "BookingAmount > QuotaAmount" {
			t.Errorf("expected error: BookingAmount > QuotaAmount, got: %v", err)
		}
	})

	t.Run("error when booking account exceeds quota", func(t *testing.T) {
		dataBefore := entity.DataQuotaDeviasiBranch{
			QuotaAmount:    8000,
			QuotaAccount:   3,
			BookingAmount:  2000,
			BookingAccount: 6, // Exceeds new quota
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?`)).
			WithArgs(branchID).
			WillReturnRows(sqlmock.NewRows([]string{"quota_amount", "quota_account", "booking_amount", "booking_account", "balance_amount", "balance_account", "is_active", "updated_at", "updated_by"}).
				AddRow(dataBefore.QuotaAmount, dataBefore.QuotaAccount, dataBefore.BookingAmount, dataBefore.BookingAccount, 6000, 3, true, time.Now(), "tester"))

		mock.ExpectRollback()

		_, _, err := repo.ProcessUpdateQuotaDeviasiBranch(branchID, mBranchDeviasi)
		if err == nil || err.Error() != "BookingAccount > QuotaAccount" {
			t.Errorf("expected error: BookingAccount > QuotaAccount, got: %v", err)
		}
	})
}

func TestBatchUpdateQuotaDeviasi(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	data := []entity.MappingBranchDeviasi{
		{
			BranchID:       "400",
			QuotaAmount:    1000,
			QuotaAccount:   500,
			BookingAmount:  200,
			BookingAccount: 100,
			IsActive:       true,
			UpdatedBy:      "1234567",
		},
	}

	t.Run("success without rollback", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT BranchID, final_approval, quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID IN (?)`)).
			WithArgs(data[0].BranchID).
			WillReturnRows(sqlmock.NewRows([]string{"BranchID", "quota_amount", "quota_account", "booking_amount", "booking_account", "is_active"}).
				AddRow(data[0].BranchID, 800, 400, 200, 100, true))

		mock.ExpectExec(regexp.QuoteMeta(
			`UPDATE "m_branch_deviasi" SET "balance_account" = ?, "balance_amount" = ?, "quota_account" = ?, "quota_amount" = ?, "updated_at" = ?, "updated_by" = ? WHERE (BranchID = ?)`)).
			WithArgs(400, 800.00, data[0].QuotaAccount, data[0].QuotaAmount, sqlmock.AnyArg(), data[0].UpdatedBy, data[0].BranchID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT BranchID, final_approval, quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID IN (?)`)).
			WithArgs(data[0].BranchID).
			WillReturnRows(sqlmock.NewRows([]string{"BranchID", "quota_amount", "quota_account", "booking_amount", "booking_account", "is_active"}).
				AddRow(data[0].BranchID, 800, 400, 200, 100, true))

		mock.ExpectCommit()

		// Call the function
		dataBefore, dataAfter, err := repo.BatchUpdateQuotaDeviasi(data)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		if dataBefore == nil || dataAfter == nil {
			t.Fatalf("Expected non-nil dataBefore and dataAfter")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestProcessResetQuotaDeviasiBranch(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	branchID := "BR001"
	updatedBy := "tester"

	t.Run("success update", func(t *testing.T) {
		dataBefore := entity.DataQuotaDeviasiBranch{
			QuotaAmount:    8000,
			QuotaAccount:   3,
			BookingAmount:  2000,
			BookingAccount: 2,
			BalanceAmount:  6000,
			BalanceAccount: 1,
			IsActive:       true,
			UpdatedAt:      time.Now(),
			UpdatedBy:      "tester",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?`)).
			WithArgs(branchID).
			WillReturnRows(sqlmock.NewRows([]string{"quota_amount", "quota_account", "booking_amount", "booking_account", "balance_amount", "balance_account", "is_active", "updated_at", "updated_by"}).
				AddRow(dataBefore.QuotaAmount, dataBefore.QuotaAccount, dataBefore.BookingAmount, dataBefore.BookingAccount, dataBefore.BalanceAmount, dataBefore.BalanceAccount, true, time.Now(), "tester"))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_branch_deviasi" SET "balance_account" = ?, "balance_amount" = ?, "booking_account" = ?, "booking_amount" = ?, "is_active" = ?, "quota_account" = ?, "quota_amount" = ?, "updated_at" = ?, "updated_by" = ? WHERE (BranchID = ?)`)).
			WithArgs(0, 0.00, 0, 0.00, false, 0, 0.00, sqlmock.AnyArg(), updatedBy, branchID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?`)).
			WithArgs(branchID).
			WillReturnRows(sqlmock.NewRows([]string{"quota_amount", "quota_account", "booking_amount", "booking_account", "balance_amount", "balance_account", "is_active", "updated_at", "updated_by"}).
				AddRow(0.00, 0, 0.00, 0, 0.00, 0, false, time.Now(), "tester"))

		mock.ExpectCommit()

		dataBeforeResult, dataAfterResult, err := repo.ProcessResetQuotaDeviasiBranch(branchID, updatedBy)
		dataBeforeResult.UpdatedAt = dataBefore.UpdatedAt

		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
		if dataBeforeResult != dataBefore {
			t.Errorf("expected dataBefore: %v, got: %v", dataBefore, dataBeforeResult)
		}
		if dataAfterResult.BalanceAmount != 0.00 {
			t.Errorf("expected balance amount: %v, got: %v", 0.00, dataAfterResult.BalanceAmount)
		}
	})
}

func TestProcessResetAllQuotaDeviasi(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	updatedBy := "tester"

	t.Run("success update", func(t *testing.T) {

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "m_branch_deviasi" SET "balance_account" = ?, "balance_amount" = ?, "booking_account" = ?, "booking_amount" = ?, "is_active" = ?, "quota_account" = ?, "quota_amount" = ?, "updated_at" = ?, "updated_by" = ?`)).
			WithArgs(0, 0.00, 0, 0.00, false, 0, 0.00, sqlmock.AnyArg(), updatedBy).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.ProcessResetAllQuotaDeviasi(updatedBy)

		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}
	})
}

func TestGetMappingCluster(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	expectedData := []entity.MasterMappingCluster{
		{
			BranchID:       "400",
			CustomerStatus: "AO/RO",
			BpkbNameType:   1,
			Cluster:        "Cluster A",
		},
	}

	t.Run("success", func(t *testing.T) {

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT \* FROM kmb_mapping_cluster_branch WITH \(nolock\) ORDER BY branch_id ASC`).
			WillReturnRows(sqlmock.NewRows([]string{"branch_id", "customer_status", "bpkb_name_type", "cluster"}).
				AddRow("400", "AO/RO", 1, "Cluster A"))
		mock.ExpectCommit()

		// Call the function
		data, err := repo.GetMappingCluster()

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedData, data, "Expected data slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {

		// Mock SQL query to simulate record not found
		mock.ExpectBegin()

		mock.ExpectQuery(`SELECT \* FROM kmb_mapping_cluster_branch WITH \(nolock\) ORDER BY branch_id ASC`).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		// Call the function
		_, err := repo.GetMappingCluster()

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetInquiryMappingCluster(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.InquiryMappingCluster{
		{
			BranchID:       "400",
			BranchName:     "BEKASI",
			CustomerStatus: "AO/RO",
			BpkbNameType:   1,
			Cluster:        "Cluster A",
		},
	}

	t.Run("success with search", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListMappingCluster{
			Search: "abranchname",
		}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				COUNT(*) AS totalRow
			FROM (
				SELECT kmcb.*, cb.BranchName AS branch_name 
				FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
				LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
				WHERE (kmcb.branch_id LIKE '%abranchname%' OR cb.BranchName LIKE '%abranchname%')
			) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
			kmcb.*, 
			cb.BranchName AS branch_name
			FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
			LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
			WHERE (kmcb.branch_id LIKE '%abranchname%' OR cb.BranchName LIKE '%abranchname%') ORDER BY kmcb.branch_id ASC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"branch_id", "branch_name", "customer_status", "bpkb_name_type", "cluster"}).AddRow("400", "BEKASI", "AO/RO", 1, "Cluster A"))

		mock.ExpectCommit()

		// Call the function
		reason, _, err := repo.GetInquiryMappingCluster(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("success with filter", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListMappingCluster{
			BranchID:       "400,401,403",
			CustomerStatus: "NEW",
			BPKBNameType:   "1",
			Cluster:        "Cluster A",
		}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				COUNT(*) AS totalRow
			FROM (
				SELECT kmcb.*, cb.BranchName AS branch_name 
				FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
				LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
				WHERE kmcb.branch_id IN ('400','401','403') AND kmcb.customer_status = 'NEW' AND kmcb.bpkb_name_type = '1' AND kmcb.cluster = 'Cluster A'
			) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
			kmcb.*, 
			cb.BranchName AS branch_name
			FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
			LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
			WHERE kmcb.branch_id IN ('400','401','403') AND kmcb.customer_status = 'NEW' AND kmcb.bpkb_name_type = '1' AND kmcb.cluster = 'Cluster A' ORDER BY kmcb.branch_id ASC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"branch_id", "branch_name", "customer_status", "bpkb_name_type", "cluster"}).AddRow("400", "BEKASI", "AO/RO", 1, "Cluster A"))

		mock.ExpectCommit()

		// Call the function
		reason, _, err := repo.GetInquiryMappingCluster(req, 1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListMappingCluster{}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				COUNT(*) AS totalRow
			FROM (
				SELECT kmcb.*, cb.BranchName AS branch_name 
				FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
				LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
			) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
			kmcb.*, 
			cb.BranchName AS branch_name
			FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
			LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
			ORDER BY kmcb.branch_id ASC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"branch_id", "branch_name", "customer_status", "bpkb_name_type", "cluster"}))

		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetInquiryMappingCluster(req, 1)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error get count data", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListMappingCluster{}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				COUNT(*) AS totalRow
			FROM (
				SELECT kmcb.*, cb.BranchName AS branch_name 
				FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
				LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
			) AS y`)).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetInquiryMappingCluster(req, 1)

		// Verify that an error was returned as expected
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error get data", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListMappingCluster{}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				COUNT(*) AS totalRow
			FROM (
				SELECT kmcb.*, cb.BranchName AS branch_name 
				FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
				LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
			) AS y`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
			kmcb.*, 
			cb.BranchName AS branch_name
			FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
			LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID 
			ORDER BY kmcb.branch_id ASC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetInquiryMappingCluster(req, 1)

		// Verify that an error was returned as expected
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestBatchUpdateMappingCluster(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	mapping := []entity.MasterMappingCluster{
		{
			BranchID:       "400",
			CustomerStatus: "AO/RO",
			BpkbNameType:   1,
			Cluster:        "Cluster A",
		},
	}

	history := entity.HistoryConfigChanges{
		ID:         utils.GenerateUUID(),
		ConfigID:   "kmb_mapping_cluster_branch",
		ObjectName: "kmb_mapping_cluster_branch",
		Action:     "UPDATE",
		DataBefore: `[{"branch_id":"400","customer_status":"AO/RO","bpkb_name_type":1,"cluster":"Cluster C"}]`,
		DataAfter:  `[{"branch_id":"400","customer_status":"AO/RO","bpkb_name_type":1,"cluster":"Cluster A"}]`,
		CreatedBy:  "1234567",
		CreatedAt:  time.Now(),
	}

	t.Run("success without rollback", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "kmb_mapping_cluster_branch"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		for _, val := range mapping {
			mock.ExpectExec(regexp.QuoteMeta(
				`INSERT INTO "kmb_mapping_cluster_branch" ("branch_id","customer_status","bpkb_name_type","cluster") VALUES (?,?,?,?)`)).
				WithArgs(val.BranchID, val.CustomerStatus, val.BpkbNameType, val.Cluster).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "history_config_changes" ("id","config_id","object_name","action","data_before","data_after","created_by","created_at") VALUES (?,?,?,?,?,?,?,?)`)).
			WithArgs(history.ID, history.ConfigID, history.ObjectName, history.Action, history.DataBefore, history.DataAfter, history.CreatedBy, history.CreatedAt).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		// Call the function
		err := repo.BatchUpdateMappingCluster(mapping, history)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error delete rollback", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "kmb_mapping_cluster_branch"`)).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectRollback()

		// Call the function
		err := repo.BatchUpdateMappingCluster(mapping, history)

		// Verify that an error was returned as expected
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error insert mapping cluster rollback", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "kmb_mapping_cluster_branch"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		for _, val := range mapping {
			mock.ExpectExec(regexp.QuoteMeta(
				`INSERT INTO "kmb_mapping_cluster_branch" ("branch_id","customer_status","bpkb_name_type","cluster") VALUES (?,?,?,?)`)).
				WithArgs(val.BranchID, val.CustomerStatus, val.BpkbNameType, val.Cluster).
				WillReturnError(gorm.ErrInvalidTransaction)
		}

		mock.ExpectRollback()

		// Call the function
		err := repo.BatchUpdateMappingCluster(mapping, history)

		// Verify that an error was returned as expected
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error insert history change log rollback", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(
			`DELETE FROM "kmb_mapping_cluster_branch"`)).
			WillReturnResult(sqlmock.NewResult(1, 1))

		for _, val := range mapping {
			mock.ExpectExec(regexp.QuoteMeta(
				`INSERT INTO "kmb_mapping_cluster_branch" ("branch_id","customer_status","bpkb_name_type","cluster") VALUES (?,?,?,?)`)).
				WithArgs(val.BranchID, val.CustomerStatus, val.BpkbNameType, val.Cluster).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}

		mock.ExpectExec(regexp.QuoteMeta(
			`INSERT INTO "history_config_changes" ("id","config_id","object_name","action","data_before","data_after","created_by","created_at") VALUES (?,?,?,?,?,?,?,?)`)).
			WithArgs(history.ID, history.ConfigID, history.ObjectName, history.Action, history.DataBefore, history.DataAfter, history.CreatedBy, history.CreatedAt).
			WillReturnError(gorm.ErrInvalidTransaction)

		mock.ExpectRollback()

		// Call the function
		err := repo.BatchUpdateMappingCluster(mapping, history)

		// Verify that an error was returned as expected
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetMappingClusterBranch(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.ConfinsBranch{
		{
			BranchID:   "400",
			BranchName: "BEKASI",
		},
	}

	t.Run("success with filter", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListMappingClusterBranch{
			BranchID:   "400,401,403",
			BranchName: "BEKASI",
		}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT 
			kmcb.branch_id AS BranchID, 
			CASE 
				WHEN kmcb.branch_id = '000' THEN 'PRIME PRIORITY'
				ELSE cb.BranchName 
			END AS BranchName 
			FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
			LEFT JOIN confins_branch cb ON cb.BranchID = kmcb.branch_id 
			WHERE kmcb.branch_id IN ('400','401','403') AND cb.BranchName LIKE '%BEKASI%'
			ORDER BY kmcb.branch_id ASC`)).
			WillReturnRows(sqlmock.NewRows([]string{"BranchID", "BranchName"}).AddRow("400", "BEKASI"))

		mock.ExpectCommit()

		// Call the function
		reason, err := repo.GetMappingClusterBranch(req)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		// Expected input and output
		req := request.ReqListMappingClusterBranch{}

		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT 
			kmcb.branch_id AS BranchID, 
			CASE 
				WHEN kmcb.branch_id = '000' THEN 'PRIME PRIORITY'
				ELSE cb.BranchName 
			END AS BranchName 
			FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
			LEFT JOIN confins_branch cb ON cb.BranchID = kmcb.branch_id 
			ORDER BY kmcb.branch_id ASC`)).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectCommit()

		// Call the function
		_, err := repo.GetMappingClusterBranch(req)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}

func TestGetMappingClusterChangeLog(t *testing.T) {
	// Setup mock database connection
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.MappingClusterChangeLog{
		{
			ID:         "041b02ab-19b7-4670-8a98-df612a6a93f6",
			DataBefore: `[{"branch_id":"400","customer_status":"AO/RO","bpkb_name_type":1,"cluster":"Cluster C"}]`,
			DataAfter:  `[{"branch_id":"400","customer_status":"AO/RO","bpkb_name_type":1,"cluster":"Cluster A"}]`,
			UserName:   "user",
			CreatedAt:  "2024-02-28 08:04:05",
		},
	}

	t.Run("success", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
				COUNT(*) AS totalRow
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				hcc.id, hcc.data_before, hcc.data_after, hcc.created_at, ud.name AS user_name
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'
			ORDER BY hcc.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "data_before", "data_after", "created_at", "user_name"}).AddRow("041b02ab-19b7-4670-8a98-df612a6a93f6", `[{"branch_id":"400","customer_status":"AO/RO","bpkb_name_type":1,"cluster":"Cluster C"}]`, `[{"branch_id":"400","customer_status":"AO/RO","bpkb_name_type":1,"cluster":"Cluster A"}]`, "2024-02-28 08:04:05", "user"))

		mock.ExpectCommit()

		// Call the function
		reason, _, err := repo.GetMappingClusterChangeLog(1)

		// Verify the result
		if err != nil {
			t.Fatalf("Expected no error, but got: %v", err)
		}
		assert.Equal(t, expectedInquiry, reason, "Expected reason slice to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
				COUNT(*) AS totalRow
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				hcc.id, hcc.data_before, hcc.data_after, hcc.created_at, ud.name AS user_name
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'
			ORDER BY hcc.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "data_before", "data_after", "created_at", "user_name"}))

		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetMappingClusterChangeLog(1)

		// Verify the error message
		expectedErr := errors.New(constant.RECORD_NOT_FOUND)
		assert.EqualError(t, err, expectedErr.Error(), "Expected error to match")

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error get count data", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
				COUNT(*) AS totalRow
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'`)).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetMappingClusterChangeLog(1)

		// Verify that an error was returned as expected
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("error get data", func(t *testing.T) {
		// Mock SQL query and result
		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT 
				COUNT(*) AS totalRow
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
				hcc.id, hcc.data_before, hcc.data_after, hcc.created_at, ud.name AS user_name
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'
			ORDER BY hcc.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
			WillReturnError(sql.ErrNoRows)

		mock.ExpectCommit()

		// Call the function
		_, _, err := repo.GetMappingClusterChangeLog(1)

		// Verify that an error was returned as expected
		if err == nil {
			t.Fatalf("Expected an error, got nil")
		}

		// Ensure all expectations were met
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}
