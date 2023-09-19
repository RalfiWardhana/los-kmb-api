package repository

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
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
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqInquiryPrescreening{
		Search: "aprospectid",
	}

	expectedInquiry := []entity.InquiryPrescreening{{CmoRecommendation: 0, Activity: "", SourceDecision: "", Decision: "", Reason: "", DecisionBy: "", DecisionAt: "", ProspectID: "", BranchName: "", IncomingSource: "", CreatedAt: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InsuranceRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, Total: 0, MonthlyInstallment: 0, FirstPayment: 0, FirstInstallment: "", FirstPaymentDate: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	// Mock SQL query and result
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT
	COUNT(tt.ProspectID) AS totalRow
	FROM
	(
		SELECT
		tm.ProspectID,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName
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
	) AS tt  WHERE (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%')`)).
		WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
			AddRow("27"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM ( SELECT tm.ProspectID, cb.BranchName, tia.info AS CMORecommend, tst.activity, tst.source_decision, tps.decision, tps.reason, tps.created_by AS DecisionBy, tps.created_at AS DecisionAt, CASE WHEN tm.incoming_source = 'SLY' THEN 'SALLY' ELSE 'NE' END AS incoming_source, tf.customer_status, tm.created_at, scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber, scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName, scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace, tcp.BirthDate, scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName, CASE WHEN tcp.Gender = 'M' THEN 'Laki-Laki' WHEN tcp.Gender = 'F' THEN 'Perempuan' END AS 'Gender', scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress, CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW, cal.Kelurahan AS LegalKelurahan, cal.Kecamatan AS LegalKecamatan, cal.ZipCode AS LegalZipcode, cal.City AS LegalCity, scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone, scp.dbo.DEC_B64('SEC', tcp.Email) AS Email, edu.value AS Education, mst.value AS MaritalStatus, tcp.NumOfDependence, scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress, CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW, car.Kelurahan AS ResidenceKelurahan, car.Kecamatan AS ResidenceKecamatan, car.ZipCode AS ResidenceZipcode, car.City AS ResidenceCity, hst.value AS HomeStatus, mn.value AS StaySinceMonth, tcp.StaySinceYear, ta.ProductOfferingID, ta.dealer, 'KMB MOTOR' AS AssetType, ti.asset_description, ti.manufacture_year, ti.color, chassis_number, engine_number, interest_rate, insurance_rate, Tenor AS InstallmentPeriod, OTR, DPAmount, AF AS FinanceAmount, interest_amount, insurance_amount, AdminFee, provision_fee, NTF, (NTF + interest_amount) AS Total, InstallmentAmount AS MonthlyInstallment, first_payment, FirstInstallment, first_payment_date, pr.value AS ProfessionID, jt.value AS JobType, jb.value AS JobPosition, mn2.value AS EmploymentSinceMonth, tce.EmploymentSinceYear, tce.CompanyName, cac.AreaPhone AS CompanyAreaPhone, cac.Phone AS CompanyPhone, tcp.ExtCompanyPhone, scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress, CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW, cac.Kelurahan AS CompanyKelurahan, cac.Kecamatan AS CompanyKecamatan, car.ZipCode AS CompanyZipcode, car.City AS CompanyCity, tce.MonthlyFixedIncome, tce.MonthlyVariableIncome, tce.SpouseIncome, tcp.SourceOtherIncome, tcs.FullName AS SpouseLegalName, tcs.CompanyName AS SpouseCompanyName, tcs.CompanyPhone AS SpouseCompanyPhone, tcs.MobilePhone AS SpouseMobilePhone, tcs.IDNumber AS SpouseIDNumber, pr2.value AS SpouseProfession, em.Name AS EmconName, em.Relationship, em.MobilePhone AS EmconMobilePhone, scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress, CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW, cae.Kelurahan AS EmergencyKelurahan, cae.Kecamatan AS EmergencyKecamatan, cae.ZipCode AS EmergencyZipcode, cae.City AS EmergencyCity, cae.AreaPhone AS EmergencyAreaPhone, cae.Phone AS EmergencyPhone, tce.IndustryTypeID FROM trx_master tm WITH (nolock) INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID INNER JOIN ( SELECT ProspectID, Address, RT, RW, Kelurahan, Kecamatan, ZipCode, City FROM trx_customer_address WITH (nolock) WHERE "Type" = 'LEGAL' ) cal ON tm.ProspectID = cal.ProspectID INNER JOIN ( SELECT ProspectID, Address, RT, RW, Kelurahan, Kecamatan, ZipCode, City FROM trx_customer_address WITH (nolock) WHERE "Type" = 'RESIDENCE' ) car ON tm.ProspectID = car.ProspectID INNER JOIN ( SELECT ProspectID, Address, RT, RW, Kelurahan, Kecamatan, ZipCode, City, Phone, AreaPhone FROM trx_customer_address WITH (nolock) WHERE "Type" = 'COMPANY' ) cac ON tm.ProspectID = cac.ProspectID INNER JOIN ( SELECT ProspectID, Address, RT, RW, Kelurahan, Kecamatan, ZipCode, City, Phone, AreaPhone FROM trx_customer_address WITH (nolock) WHERE "Type" = 'EMERGENCY' ) cae ON tm.ProspectID = cae.ProspectID INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'Education' ) edu ON tcp.Education = edu.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'MaritalStatus' ) mst ON tcp.MaritalStatus = mst.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'HomeStatus' ) hst ON tcp.HomeStatus = hst.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'MonthName' ) mn ON tcp.StaySinceMonth = mn.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'ProfessionID' ) pr ON tce.ProfessionID = pr.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'JobType' ) jt ON tce.JobType = jt.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'JobPosition' ) jb ON tce.JobPosition = jb.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'MonthName' ) mn2 ON tce.EmploymentSinceMonth = mn2.[key] LEFT JOIN ( SELECT [key], value FROM app_config ap WITH (nolock) WHERE group_name = 'ProfessionID' ) pr2 ON tcs.ProfessionID = pr2.[key] ) AS tt  WHERE (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%') ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
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

func TestGetInquiryPrescreening_RecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqInquiryPrescreening{}

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(regexp.QuoteMeta(`
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
	'KMB MOTOR' AS AssetType,
	ti.asset_description,
	ti.manufacture_year,
	ti.color,
	chassis_number,
	engine_number,
	interest_rate,
	insurance_rate,
	Tenor AS InstallmentPeriod,
	OTR,
	DPAmount,
	AF AS FinanceAmount,
	interest_amount,
	insurance_amount,
	AdminFee,
	provision_fee,
	NTF,
	(NTF + interest_amount) AS Total,
	InstallmentAmount AS MonthlyInstallment,
	first_payment,
	FirstInstallment,
	first_payment_date,
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
	) pr2 ON tcs.ProfessionID = pr2.[key] ) AS tt  ORDER BY tt.created_at DESC`)).
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

func Test_repoHandler_SavePrescreening(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB)
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

	t.Run("success", func(t *testing.T) {

		mock.ExpectBegin()
		query := `UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`
		queryRegex := regexp.QuoteMeta(query)
		mock.ExpectExec(queryRegex).WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "trx_details" (.*)`).
			WithArgs(detail...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_prescreening" ("ProspectID","decision","reason","created_at","created_by") VALUES (?,?,?,?,?)`)).
			WithArgs(trxPrescreening.ProspectID, trxPrescreening.Decision, trxPrescreening.Reason, sqlmock.AnyArg(), trxPrescreening.CreatedBy).
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
