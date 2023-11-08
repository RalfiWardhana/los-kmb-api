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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedPhoto := []entity.DataPhoto{{PhotoID: "1", Label: "KTP", Url: "http://example.com/photo1.jpg"}}

	// Mock SQL query and result
	mock.ExpectQuery(`SELECT tcp.photo_id, CASE WHEN lpi.Name IS NULL THEN 'LAINNYA' ELSE lpi.Name END AS label, tcp.url FROM trx_customer_photo tcp WITH \(nolock\) LEFT JOIN m_label_photo_inquiry lpi ON lpi.LabelPhotoID = tcp.photo_id WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"photo_id", "label", "url"}).
			AddRow(1, "KTP", "http://example.com/photo1.jpg"))

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT tcp.photo_id, CASE WHEN lpi.Name IS NULL THEN 'LAINNYA' ELSE lpi.Name END AS label, tcp.url FROM trx_customer_photo tcp WITH \(nolock\) LEFT JOIN m_label_photo_inquiry lpi ON lpi.LabelPhotoID = tcp.photo_id WHERE ProspectID = \?`).WithArgs(prospectID).
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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedStatus := entity.TrxStatus{
		Activity:       constant.ACTIVITY_UNPROCESS,
		SourceDecision: constant.PRESCREENING,
	}

	// Mock SQL query and result
	mock.ExpectQuery(`SELECT activity, decision, source_decision FROM trx_status WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnRows(sqlmock.NewRows([]string{"activity", "source_decision"}).
			AddRow(constant.ACTIVITY_UNPROCESS, constant.PRESCREENING))

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
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(`SELECT activity, decision, source_decision FROM trx_status WITH \(nolock\) WHERE ProspectID = \?`).WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)

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
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (SELECT Code, ReasonID, ReasonMessage FROM m_reason_message WITH (nolock)) AS tt WHERE ReasonID NOT IN ('99','100','101','102') ORDER BY tt.ReasonID asc OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
		WillReturnRows(sqlmock.NewRows([]string{"Code", "ReasonID", "ReasonMessage"}).
			AddRow("12", "11", "Akte Jual Beli Tidak Sesuai"))

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
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqReasonPrescreening{
		ReasonID: "99,100,101,102",
	}

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (SELECT Code, ReasonID, ReasonMessage FROM m_reason_message WITH (nolock)) AS tt WHERE ReasonID NOT IN ('99','100','101','102') ORDER BY tt.ReasonID asc`)).
		WillReturnError(gorm.ErrRecordNotFound)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqInquiryPrescreening{
		Search:   "aprospectid",
		BranchID: "426,903",
	}

	expectedInquiry := []entity.InquiryPrescreening{{CmoRecommendation: 0, Activity: "", SourceDecision: "", Decision: "", Reason: "", DecisionBy: "", DecisionAt: "", ProspectID: "", BranchName: "", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	// Mock SQL query and result
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT
	COUNT(tt.ProspectID) AS totalRow
	FROM
	(
		SELECT
		cb.BranchID,
		tm.ProspectID,
		tm.created_at,
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
	) AS tt WHERE tt.BranchID IN ('426','903') AND (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%')`)).
		WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
			AddRow("27"))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (
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
		) pr2 ON tcs.ProfessionID = pr2.[key] ) AS tt WHERE tt.BranchID IN ('426','903') AND (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%') ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
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

func TestGetInquiryPrescreeningWithout(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	expectedInquiry := []entity.InquiryPrescreening{{CmoRecommendation: 0, Activity: "", SourceDecision: "", Decision: "", Reason: "", DecisionBy: "", DecisionAt: "", ProspectID: "", BranchName: "", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	// Mock SQL query and result

	t.Run("without param branch", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryPrescreening{
			Search: "aprospectid",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(
			SELECT
			cb.BranchID,
			tm.ProspectID,
			tm.created_at,
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
		) AS tt WHERE (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%')`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (
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
			) pr2 ON tcs.ProfessionID = pr2.[key] ) AS tt WHERE (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%') ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
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

	t.Run("with one param branch", func(t *testing.T) {
		// Expected input and output
		req := request.ReqInquiryPrescreening{
			Search:   "aprospectid",
			BranchID: "426",
		}

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(
			SELECT
			cb.BranchID,
			tm.ProspectID,
			tm.created_at,
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
		) AS tt WHERE tt.BranchID IN ('426') AND (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%')`)).
			WillReturnRows(sqlmock.NewRows([]string{"totalRow"}).
				AddRow("27"))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT tt.* FROM (
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
			) pr2 ON tcs.ProfessionID = pr2.[key] ) AS tt WHERE tt.BranchID IN ('426') AND (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%') ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
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

func TestGetInquiryPrescreening_RecordNotFound(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqInquiryPrescreening{}

	// Mock SQL query to simulate record not found
	mock.ExpectQuery(regexp.QuoteMeta(`
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
        ) pr2 ON tcs.ProfessionID = pr2.[key] ) AS tt WHERE CAST(tt.created_at AS date) >= DATEADD(day, , CAST(GETDATE() AS date)) ORDER BY tt.created_at DESC`)).
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

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB)
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

	t.Run("success", func(t *testing.T) {

		mock.ExpectBegin()
		query := `UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`
		queryRegex := regexp.QuoteMeta(query)
		mock.ExpectExec(queryRegex).WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(`INSERT INTO "trx_details" (.*)`).
			WithArgs(detail...).
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

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
		mock.ExpectQuery(`SELECT thas.decision_by, thas.next_final_approval_flag, CASE WHEN thas.decision = 'APR' THEN 'Approve' WHEN thas.decision = 'REJ' THEN 'Reject' WHEN thas.decision = 'CAN' THEN 'Cancel' ELSE '-' END AS decision, CASE WHEN thas.need_escalation = 1 THEN 'Yes' ELSE 'No' END AS need_escalation, thas.source_decision, CASE WHEN thas.next_step<>'' THEN thas.next_step ELSE '-' END AS next_step, CASE WHEN thas.note<>'' THEN thas.note ELSE '-' END AS note, thas.created_at, CASE WHEN thas.source_decision = 'CRA' AND tcd.slik_result<>'' THEN tcd.slik_result ELSE '-' END AS slik_result FROM trx_history_approval_scheme thas WITH \(nolock\) LEFT JOIN trx_ca_decision tcd on thas.ProspectID = tcd.ProspectID WHERE thas.ProspectID = \? ORDER BY thas.created_at DESC`).WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"decision", "decision_by", "next_final_approval_flag", "need_escalation", "source_decision", "next_step", "note", "created_at", "slik_result"}).
				AddRow("APR", "User CA KMB", 1, "No", "CRA", "CBM", "Ok dari CA", time.Time{}, "Lancar"))

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
		mock.ExpectQuery(`SELECT thas.decision_by, thas.next_final_approval_flag, CASE WHEN thas.decision = 'APR' THEN 'Approve' WHEN thas.decision = 'REJ' THEN 'Reject' WHEN thas.decision = 'CAN' THEN 'Cancel' ELSE '-' END AS decision, CASE WHEN thas.need_escalation = 1 THEN 'Yes' ELSE 'No' END AS need_escalation, thas.source_decision, CASE WHEN thas.next_step<>'' THEN thas.next_step ELSE '-' END AS next_step, CASE WHEN thas.note<>'' THEN thas.note ELSE '-' END AS note, thas.created_at, CASE WHEN thas.source_decision = 'CRA' AND tcd.slik_result<>'' THEN tcd.slik_result ELSE '-' END AS slik_result FROM trx_history_approval_scheme thas WITH \(nolock\) LEFT JOIN trx_ca_decision tcd on thas.ProspectID = tcd.ProspectID WHERE thas.ProspectID = \? ORDER BY thas.created_at DESC`).WithArgs(prospectID).
			WillReturnError(gorm.ErrRecordNotFound)

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
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM trx_internal_record WITH (nolock) WHERE ProspectID = ? ORDER BY created_at DESC`)).WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"ApplicationID", "ProductType", "AgreementDate", "AssetCode", "Tenor", "OutstandingPrincipal", "InstallmentAmount", "ContractStatus", "CurrentCondition"}).
				AddRow("426A202201124155", "KMB", time.Time{}, "K-YMH.MOTOR.NMAX (B6H A/T)", 26, 0, 1866000, "LIV", "OVD 204 hari"))

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM trx_internal_record WITH (nolock) WHERE ProspectID = ? ORDER BY created_at DESC`)).WithArgs(prospectID).
			WillReturnError(gorm.ErrRecordNotFound)

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
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	ntf := 10000.65
	expectedData := entity.MappingLimitApprovalScheme{
		Alias: "CBM",
	}

	t.Run("success", func(t *testing.T) {

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT [alias] FROM m_limit_approval_scheme WITH (nolock) WHERE ? between coverage_ntf_start AND coverage_ntf_end`)).WithArgs(ntf).
			WillReturnRows(sqlmock.NewRows([]string{"alias"}).
				AddRow("CBM"))

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT [alias] FROM m_limit_approval_scheme WITH (nolock) WHERE ? between coverage_ntf_start AND coverage_ntf_end`)).WithArgs(ntf).
			WillReturnError(gorm.ErrRecordNotFound)

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

func TestGetHistoryProcess(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	prospectID := "12345"
	expectedData := []entity.TrxDetail{
		{
			Decision:       "PASS",
			SourceDecision: "PRE SCREENING",
			Info:           "Dokumen Sesuai",
			CreatedAt:      time.Time{},
		},
	}

	t.Run("success", func(t *testing.T) {

		// Mock SQL query and result
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		CASE
		 WHEN td.source_decision = 'PSI' THEN 'PRE SCREENING'
		 WHEN td.source_decision = 'DCK' THEN 'DUPLICATION CHECKING'
		 WHEN td.source_decision = 'DCP'
		 OR td.source_decision = 'ARI'
		 OR td.source_decision = 'KTP' THEN 'EKYC'
		 WHEN td.source_decision = 'PBK' THEN 'PEFINDO'
		 WHEN td.source_decision = 'SCS' THEN 'SCOREPRO'
		 WHEN td.source_decision = 'DSR' THEN 'DSR'
		 WHEN td.source_decision = 'CRA' THEN 'CREDIT ANALYSIS'
		 WHEN td.source_decision = 'CBM'
		  OR td.source_decision = 'DRM'
		  OR td.source_decision = 'GMO'
		  OR td.source_decision = 'COM'
		  OR td.source_decision = 'GMC'
		  OR td.source_decision = 'UCC' THEN 'CREDIT COMMITEE'
		 ELSE '-'
		END AS source_decision,
		CASE
		 WHEN td.decision = 'PAS' THEN 'PASS'
		 WHEN td.decision = 'REJ' THEN 'REJECT'
		 WHEN td.decision = 'CAN' THEN 'CANCEL'
		 WHEN td.decision = 'CPR' THEN 'CREDIT PROCESS'
		 ELSE '-'
		END AS decision,
		ap.reason AS info,
		td.created_at
	FROM
		trx_details td WITH (nolock)
		LEFT JOIN app_rules ap ON ap.rule_code = td.rule_code
	WHERE td.ProspectID = ? AND td.source_decision IN('PSI','DCK','DCP','ARI','KTP','PBK','SCS','DSR','CRA','CBM','DRM','GMO','COM','GMC','UCC')
	AND td.decision <> 'CTG' ORDER BY td.created_at ASC`)).WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"source_decision", "decision", "info", "created_at"}).
				AddRow("PRE SCREENING", "PASS", "Dokumen Sesuai", time.Time{}))

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

		// Mock SQL query to simulate record not found
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT
		CASE
		 WHEN td.source_decision = 'PSI' THEN 'PRE SCREENING'
		 WHEN td.source_decision = 'DCK' THEN 'DUPLICATION CHECKING'
		 WHEN td.source_decision = 'DCP'
		 OR td.source_decision = 'ARI'
		 OR td.source_decision = 'KTP' THEN 'EKYC'
		 WHEN td.source_decision = 'PBK' THEN 'PEFINDO'
		 WHEN td.source_decision = 'SCS' THEN 'SCOREPRO'
		 WHEN td.source_decision = 'DSR' THEN 'DSR'
		 WHEN td.source_decision = 'CRA' THEN 'CREDIT ANALYSIS'
		 WHEN td.source_decision = 'CBM'
		  OR td.source_decision = 'DRM'
		  OR td.source_decision = 'GMO'
		  OR td.source_decision = 'COM'
		  OR td.source_decision = 'GMC'
		  OR td.source_decision = 'UCC' THEN 'CREDIT COMMITEE'
		 ELSE '-'
		END AS source_decision,
		CASE
		 WHEN td.decision = 'PAS' THEN 'PASS'
		 WHEN td.decision = 'REJ' THEN 'REJECT'
		 WHEN td.decision = 'CAN' THEN 'CANCEL'
		 WHEN td.decision = 'CPR' THEN 'CREDIT PROCESS'
		 ELSE '-'
		END AS decision,
		ap.reason AS info,
		td.created_at
	FROM
		trx_details td WITH (nolock)
		LEFT JOIN app_rules ap ON ap.rule_code = td.rule_code
	WHERE td.ProspectID = ? AND td.source_decision IN('PSI','DCK','DCP','ARI','KTP','PBK','SCS','DSR','CRA','CBM','DRM','GMO','COM','GMC','UCC')
	AND td.decision <> 'CTG' ORDER BY td.created_at ASC`)).WithArgs(prospectID).
			WillReturnError(gorm.ErrRecordNotFound)

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
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM m_cancel_reason with (nolock) WHERE show = '1' ORDER BY id_cancel_reason ASC`)).
			WillReturnRows(sqlmock.NewRows([]string{"id_cancel_reason", "reason", "show"}).
				AddRow("1", "Ganti Program Marketing", "1"))

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM m_cancel_reason with (nolock) WHERE show = '1' ORDER BY id_cancel_reason ASC`)).
			WillReturnError(gorm.ErrRecordNotFound)

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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqInquiryCa{
		Search:   "aprospectid",
		BranchID: "426,903",
		Filter:   "APPROVE",
		UserID:   "5XeZs9PCeiPcZGS6azt",
	}

	expectedInquiry := []entity.InquiryCa{entity.InquiryCa{ShowAction: false, ActionDate: "", Activity: "", SourceDecision: "", StatusDecision: "", StatusReason: "", CaDecision: "", CANote: "", ScsDate: "", ScsScore: "", ScsStatus: "", BiroCustomerResult: "", BiroSpouseResult: "", DraftDecision: "", DraftSlikResult: "", DraftNote: "", DraftCreatedAt: time.Time{}, DraftCreatedBy: "", DraftDecisionBy: "", ProspectID: "EFM03406412522151347", BranchName: "BANDUNG", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerID: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), SurveyResult: "", Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	// Mock SQL query and result
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT
	COUNT(tt.ProspectID) AS totalRow
	FROM
	(
		SELECT
		cb.BranchID,
		tm.ProspectID,
		tm.lob,
		tm.created_at,
		tst.activity,
		tst.source_decision,
		tst.decision,
		tcd.decision as decision_ca,
		tdd.created_by AS draft_created_by,
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
	INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
	LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
	LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
	LEFT JOIN (
	  SELECT
		ProspectID,
		decision,
		created_at
	  FROM
		trx_ca_decision WITH (nolock)
	) tcd ON tm.ProspectID = tcd.ProspectID
	LEFT JOIN (
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
			  max(created_at)
			from
			  trx_draft_ca_decision WITH (NOLOCK)
			WHERE
			  ProspectID = x.ProspectID
		  )
	) tdd ON tm.ProspectID = tdd.ProspectID
	) AS tt WHERE tt.BranchID IN ('426','903') AND (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%') AND tt.decision= 'APR' AND tt.source_decision<>'PSI'`)).
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
	tst.reason,
	tcd.decision as decision_ca,
	CASE
	  WHEN tcd.decision='APR' THEN 'APPROVE'
	  WHEN tcd.decision='REJ' THEN 'REJECT'
	  WHEN tcd.decision='CAN' THEN 'CANCEL'
	  ELSE tcd.decision
	END AS ca_decision,
	tcd.note AS ca_note,
	CASE
	  WHEN tcd.created_at IS NOT NULL
	  AND tfa.created_at IS NULL THEN tcd.created_at
	  WHEN tfa.created_at IS NOT NULL THEN tfa.created_at
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
	tak.ScsDate,
	tak.ScsScore,
	tak.ScsStatus,
	tdb.BiroCustomerResult,
	tdb.BiroSpouseResult

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
	LEFT JOIN (
	  SELECT
		ProspectID,
		decision,
		note,
		created_at
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
	LEFT JOIN (
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
			max(created_at)
		  from
			trx_draft_ca_decision WITH (NOLOCK)
		  WHERE
			ProspectID = x.ProspectID
		)
	) tdd ON tm.ProspectID = tdd.ProspectID
) AS tt WHERE tt.BranchID IN ('426','903') AND (tt.ProspectID LIKE '%aprospectid%' OR tt.IDNumber LIKE '%aprospectid%' OR tt.LegalName LIKE '%aprospectid%') AND tt.decision= 'APR' AND tt.source_decision<>'PSI' ORDER BY tt.created_at DESC OFFSET 0 ROWS FETCH FIRST 0 ROWS ONLY`)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "BranchName", "BranchID"}).
			AddRow("EFM03406412522151347", "BANDUNG", "426"))

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
}

func TestSaveDraftData(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	_ = gormDB

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_draft_ca_decision" SET "ProspectID" = ?, "created_at" = ?, "created_by" = ?, "decision" = ?, "decision_by" = ?, "note" = ?, "slik_result" = ? WHERE (ProspectID = ?)`)).
			WithArgs(data.ProspectID, sqlmock.AnyArg(), data.CreatedBy, data.Decision, data.DecisionBy, data.Note, data.SlikResult, data.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.SaveDraftData(data)
		if err != nil {
			t.Errorf("error '%s' was not expected, but got: ", err)
		}
	})

	t.Run("success insert", func(t *testing.T) {

		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_draft_ca_decision" SET "ProspectID" = ?, "created_at" = ?, "created_by" = ?, "decision" = ?, "decision_by" = ?, "note" = ?, "slik_result" = ? WHERE (ProspectID = ?)`)).
			WithArgs(data.ProspectID, sqlmock.AnyArg(), data.CreatedBy, data.Decision, data.DecisionBy, data.Note, data.SlikResult, data.ProspectID).
			WillReturnResult(sqlmock.NewResult(0, 0))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_draft_ca_decision" ("ProspectID","decision","slik_result","note","created_at","created_by","decision_by") VALUES (?,?,?,?,?,?,?)`)).
			WithArgs(data.ProspectID, data.Decision, data.SlikResult, data.Note, sqlmock.AnyArg(), data.CreatedBy, data.DecisionBy).
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

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
	}

	trxHistoryApproval := entity.TrxHistoryApprovalScheme{
		ID:                    "xxxxx",
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
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_ca_decision" ("ProspectID","decision","slik_result","note","created_at","created_by","decision_by","final_approval") VALUES (?,?,?,?,?,?,?,?)`)).
			WithArgs(trxCaDecision.ProspectID, trxCaDecision.Decision, trxCaDecision.SlikResult, trxCaDecision.Note, sqlmock.AnyArg(), trxCaDecision.CreatedBy, trxCaDecision.DecisionBy, trxCaDecision.FinalApproval).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "reason" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ?  WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.Reason, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_history_approval_scheme" ("id","ProspectID","decision","reason","note","created_at","created_by","decision_by","need_escalation","next_final_approval_flag","source_decision","next_step") VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(sqlmock.AnyArg(), trxCaDecision.ProspectID, trxCaDecision.Decision, trxCaDecision.SlikResult.(string), trxCaDecision.Note, sqlmock.AnyArg(), trxCaDecision.CreatedBy, trxCaDecision.DecisionBy, sqlmock.AnyArg(), sqlmock.AnyArg(), trxDetail.SourceDecision, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_draft_ca_decision"  WHERE (ProspectID = ?)`)).
			WithArgs(trxCaDecision.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := newDB.ProcessTransaction(trxCaDecision, trxHistoryApproval, trxStatus, trxDetail)
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

	newDB := NewRepository(gormDB, gormDB, gormDB, gormDB)

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
	}

	t.Run("success update", func(t *testing.T) {

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_status" SET "ProspectID" = ?, "activity" = ?, "created_at" = ?, "decision" = ?, "rule_code" = ?, "source_decision" = ?, "status_process" = ? WHERE "trx_status"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(trxStatus.ProspectID, trxStatus.Activity, sqlmock.AnyArg(), trxStatus.Decision, trxStatus.RuleCode, trxStatus.SourceDecision, trxStatus.StatusProcess, trxStatus.ProspectID, trxStatus.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_details" WHERE "trx_details"."ProspectID" = ? AND ((ProspectID = ?))`)).
			WithArgs(ppid, ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_details" ("ProspectID","status_process","activity","decision","rule_code","source_decision","next_step","type","info","created_by","created_at") VALUES (?,?,?,?,?,?,?,?,?,?,?)`)).
			WithArgs(trxDetail.ProspectID, trxDetail.StatusProcess, trxDetail.Activity, trxDetail.Decision, trxDetail.RuleCode, trxDetail.SourceDecision, trxDetail.NextStep, sqlmock.AnyArg(), trxDetail.Info, trxDetail.CreatedBy, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_prescreening" WHERE (ProspectID = ?)`)).
			WithArgs(ppid).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "trx_draft_ca_decision"  WHERE (ProspectID = ?)`)).
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
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	// Expected input and output
	req := request.ReqSearchInquiry{
		UserID: "5XeZs9PCeiPcZGS6azt",
		Search: "aprospectid",
	}

	expectedInquiry := []entity.InquirySearch{entity.InquirySearch{ProspectID: "EFM03406412522151347", BranchName: "BANDUNG", IncomingSource: "", CreatedAt: "", OrderAt: "", CustomerID: "", CustomerStatus: "", IDNumber: "", LegalName: "", BirthPlace: "", BirthDate: time.Time{}, SurgateMotherName: "", Gender: "", MobilePhone: "", Email: "", Education: "", MaritalStatus: "", NumOfDependence: 0, HomeStatus: "", StaySinceMonth: "", StaySinceYear: "", ExtCompanyPhone: (*string)(nil), SourceOtherIncome: (*string)(nil), Supplier: "", ProductOfferingID: "", AssetType: "", AssetDescription: "", ManufacturingYear: "", Color: "", ChassisNumber: "", EngineNumber: "", InterestRate: 0, InstallmentPeriod: 0, OTR: 0, DPAmount: 0, FinanceAmount: 0, InterestAmount: 0, LifeInsuranceFee: 0, AssetInsuranceFee: 0, InsuranceAmount: 0, AdminFee: 0, ProvisionFee: 0, NTF: 0, NTFAkumulasi: 0, Total: 0, MonthlyInstallment: 0, FirstInstallment: "", ProfessionID: "", JobTypeID: "", JobPosition: "", CompanyName: "", IndustryTypeID: "", EmploymentSinceYear: "", EmploymentSinceMonth: "", MonthlyFixedIncome: 0, MonthlyVariableIncome: 0, SpouseIncome: 0, SpouseIDNumber: "", SpouseLegalName: "", SpouseCompanyName: "", SpouseCompanyPhone: "", SpouseMobilePhone: "", SpouseProfession: "", EmconName: "", Relationship: "", EmconMobilePhone: "", LegalAddress: "", LegalRTRW: "", LegalKelurahan: "", LegalKecamatan: "", LegalZipCode: "", LegalCity: "", ResidenceAddress: "", ResidenceRTRW: "", ResidenceKelurahan: "", ResidenceKecamatan: "", ResidenceZipCode: "", ResidenceCity: "", CompanyAddress: "", CompanyRTRW: "", CompanyKelurahan: "", CompanyKecamatan: "", CompanyZipCode: "", CompanyCity: "", CompanyAreaPhone: "", CompanyPhone: "", EmergencyAddress: "", EmergencyRTRW: "", EmergencyKelurahan: "", EmergencyKecamatan: "", EmergencyZipcode: "", EmergencyCity: "", EmergencyAreaPhone: "", EmergencyPhone: ""}}

	// Mock SQL query and result
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT
	COUNT(tt.ProspectID) AS totalRow
	FROM
	(
		SELECT
		cb.BranchID,
		tm.ProspectID,
		tm.lob,
		tm.created_at,
		tst.activity,
		tst.source_decision,
		tst.decision,
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
	INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
	LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
	LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
	LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
	) AS tt`)).
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
	  WHEN tst.decision='APR' THEN 'Approve'
	  WHEN tst.decision='REJ' THEN 'Reject'
	  WHEN tst.decision='CAN' THEN 'Cancel'
	  ELSE '-'
	END AS FinalStatus,
	CASE
	  WHEN tps.ProspectID IS NOT NULL
	  AND tst.status_process='ONP' THEN 1
	  ELSE 0
	END AS ActionReturn,
	CASE
	  WHEN tst.status_process='FIN'
	  AND tst.activity='STOP'
	  AND tst.decision='REJ' THEN 0
	  ELSE 1
	END AS ActionCancel,
	CASE
	  WHEN tcd.decision='CAN' THEN 0
	  ELSE 1
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
	LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
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
) AS tt`)).
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
}

func TestGetApprovalReason(t *testing.T) {
	// Setup mock database connection
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	// Create a repository instance
	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

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

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT CONCAT(ReasonID, '|', Type, '|', Description) AS 'id', Description AS 'value', [Type] FROM tblApprovalReason WHERE IsActive = 'True'  AND [Type] = 'APR' ORDER BY ReasonID ASC`)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "value", "Type"}).
				AddRow("1|APR|Oke", "Oke", "APR"))

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT CONCAT(ReasonID, '|', Type, '|', Description) AS 'id', Description AS 'value', [Type] FROM tblApprovalReason WHERE IsActive = 'True'  AND [Type] = 'APR' ORDER BY ReasonID ASC`)).
			WillReturnError(gorm.ErrRecordNotFound)

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
