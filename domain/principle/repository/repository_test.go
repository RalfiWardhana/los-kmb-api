package repository

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
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

func structToSliceWithTimeArgs(obj interface{}) []driver.Value {
	v := reflect.ValueOf(obj)
	values := make([]driver.Value, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Type() == reflect.TypeOf(time.Time{}) {
			values[i] = sqlmock.AnyArg()
		} else {
			values[i] = field.Interface()
		}
	}
	return values
}

func TestGetConfig(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		groupName     string
		lob           string
		key           string
		mockRows      *sqlmock.Rows
		mockError     error
		expectedData  entity.AppConfig
		expectedError error
	}{
		{
			name:      "Success Case",
			groupName: "expired_contract",
			lob:       "KMB-OFF",
			key:       "expired_contract_check",
			mockRows: sqlmock.NewRows([]string{"value"}).AddRow(
				`{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			),
			mockError: nil,
			expectedData: entity.AppConfig{
				Value: `{ "data": { "expired_contract_check_enabled": true, "expired_contract_max_months": 6 } }`,
			},
			expectedError: nil,
		},
		{
			name:          "Record Not Found",
			groupName:     "non_existent",
			lob:           "KMB-OFF",
			key:           "non_existent_key",
			mockRows:      sqlmock.NewRows([]string{"value"}),
			mockError:     gorm.ErrRecordNotFound,
			expectedData:  entity.AppConfig{},
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:          "Database Error",
			groupName:     "error_case",
			lob:           "KMB-OFF",
			key:           "error_key",
			mockRows:      nil,
			mockError:     fmt.Errorf("database connection error"),
			expectedData:  entity.AppConfig{},
			expectedError: fmt.Errorf("database connection error"),
		},
		{
			name:      "Empty Value",
			groupName: "empty_value",
			lob:       "KMB-OFF",
			key:       "empty_key",
			mockRows:  sqlmock.NewRows([]string{"value"}).AddRow(""),
			mockError: nil,
			expectedData: entity.AppConfig{
				Value: "",
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s' AND lob = '%s' AND [key]= '%s' AND is_active = 1",
				tc.groupName, tc.lob, tc.key)

			if tc.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(tc.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(tc.mockRows)
			}

			result, err := repo.GetConfig(tc.groupName, tc.lob, tc.key)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error(), "Error message mismatch")
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedData, result, "Result mismatch")

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetMinimalIncomePMK(t *testing.T) {
	t.Run("success - found with given branch ID", func(t *testing.T) {
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

		assert.NoError(t, err)
		assert.Equal(t, response, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success - fallback to default branch ID", func(t *testing.T) {
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
			BranchID:       constant.DEFAULT_BRANCH_ID,
			ID:             "123",
			StatusKonsumen: "NEW",
			Lob:            "los_kmb_off",
		}

		query1 := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)
		mock.ExpectQuery(regexp.QuoteMeta(query1)).
			WillReturnError(gorm.ErrRecordNotFound)

		query2 := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, constant.DEFAULT_BRANCH_ID, statusKonsumen)
		mock.ExpectQuery(regexp.QuoteMeta(query2)).
			WillReturnRows(sqlmock.NewRows([]string{"id", "branch_id", "status_konsumen", "income", "lob"}).
				AddRow(response.ID, response.BranchID, response.StatusKonsumen, response.Income, response.Lob))

		result, err := repo.GetMinimalIncomePMK(branchID, statusKonsumen)

		assert.NoError(t, err)
		assert.Equal(t, response, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - both queries fail", func(t *testing.T) {
		sqlDB, mock, _ := sqlmock.New()
		defer sqlDB.Close()

		gormDB, _ := gorm.Open("sqlite3", sqlDB)
		gormDB.LogMode(true)
		gormDB = gormDB.Debug()

		repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

		branchID := "BR123"
		statusKonsumen := "new"
		expectedErr := errors.New("database error")

		query1 := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)
		mock.ExpectQuery(regexp.QuoteMeta(query1)).
			WillReturnError(expectedErr)

		result, err := repo.GetMinimalIncomePMK(branchID, statusKonsumen)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error - default branch query fails", func(t *testing.T) {
		sqlDB, mock, _ := sqlmock.New()
		defer sqlDB.Close()

		gormDB, _ := gorm.Open("sqlite3", sqlDB)
		gormDB.LogMode(true)
		gormDB = gormDB.Debug()

		repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

		branchID := "BR123"
		statusKonsumen := "new"
		expectedErr := errors.New("database error")

		query1 := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)
		mock.ExpectQuery(regexp.QuoteMeta(query1)).
			WillReturnError(gorm.ErrRecordNotFound)

		query2 := fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, constant.DEFAULT_BRANCH_ID, statusKonsumen)
		mock.ExpectQuery(regexp.QuoteMeta(query2)).
			WillReturnError(expectedErr)

		result, err := repo.GetMinimalIncomePMK(branchID, statusKonsumen)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
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

func TestGetDraftPrinciple(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	fixedTime := time.Date(2025, 1, 20, 8, 48, 0, 0, time.Local)

	testCases := []struct {
		name          string
		prospectID    string
		mockRows      *sqlmock.Rows
		mockError     error
		expectedData  entity.DraftPrinciple
		expectNoError bool
	}{
		{
			name:       "Success Case",
			prospectID: "PROS-001",
			mockRows: sqlmock.NewRows([]string{
				"ProspectID",
				"IDNumber",
				"SpouseIDNumber",
				"ManufactureYear",
				"NoChassis",
				"NoEngine",
				"BranchID",
				"CMOID",
				"CMOName",
				"CC",
				"TaxDate",
				"STNKExpiredDate",
				"OwnerAsset",
				"LicensePlate",
				"Color",
				"Brand",
				"ResidenceAddress",
				"ResidenceRT",
				"ResidenceRW",
				"ResidenceProvice",
				"ResidenceCity",
				"ResidenceKecamatan",
				"ResidenceKelurahan",
				"ResidenceZipCode",
				"ResidenceAreaPhone",
				"ResidencePhone",
				"HomeStatus",
				"StaySinceYear",
				"StaySinceMonth",
				"Decision",
				"Reason",
				"BPKBName",
				"CreatedAt",
			}).AddRow(
				"PROS-001",  // ProspectID
				"",          // IDNumber
				nil,         // SpouseIDNumber
				"",          // ManufactureYear
				"ASSET-001", // NoChassis
				"",          // NoEngine
				"BR-001",    // BranchID
				"123",       // CMOID
				"",          // CMOName
				"",          // CC
				time.Time{}, // TaxDate
				time.Time{}, // STNKExpiredDate
				"",          // OwnerAsset
				"",          // LicensePlate
				"",          // Color
				"",          // Brand
				"",          // ResidenceAddress
				"",          // ResidenceRT
				"",          // ResidenceRW
				"",          // ResidenceProvice
				"",          // ResidenceCity
				"",          // ResidenceKecamatan
				"",          // ResidenceKelurahan
				"",          // ResidenceZipCode
				"",          // ResidenceAreaPhone
				"",          // ResidencePhone
				"",          // HomeStatus
				0,           // StaySinceYear
				0,           // StaySinceMonth
				"APPROVED",  // Decision
				"",          // Reason
				"",          // BPKBName
				fixedTime,   // CreatedAt
			),
			mockError: nil,
			expectedData: entity.DraftPrinciple{
				ProspectID: "PROS-001",
				NoChassis:  "ASSET-001",
				BranchID:   "BR-001",
				CMOID:      "123",
				Decision:   "APPROVED",
				CreatedAt:  fixedTime,
			},
			expectNoError: true,
		},
		{
			name:          "Record Not Found - Returns Empty Data With No Error",
			prospectID:    "PROS-002",
			mockRows:      sqlmock.NewRows([]string{"ProspectID"}),
			mockError:     gorm.ErrRecordNotFound,
			expectedData:  entity.DraftPrinciple{},
			expectNoError: true,
		},
		{
			name:          "Database Error - Returns Empty Data With No Error",
			prospectID:    "PROS-003",
			mockRows:      sqlmock.NewRows([]string{"ProspectID"}),
			mockError:     fmt.Errorf("database error"),
			expectedData:  entity.DraftPrinciple{},
			expectNoError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := `SELECT \* FROM trx_draft_principle WITH \(nolock\) WHERE ProspectID = '.*'`

			if tc.mockError != nil {
				mock.ExpectQuery(query).
					WillReturnError(tc.mockError)
			} else {
				mock.ExpectQuery(query).
					WillReturnRows(tc.mockRows)
			}

			_, err := repo.GetDraftPrinciple(tc.prospectID)

			if tc.expectNoError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMasterMappingFpdCluster(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
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

		assert.NoError(t, err)
		assert.Equal(t, response, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("record not found case", func(t *testing.T) {
		os.Setenv("DEFAULT_TIMEOUT_30S", "30")

		sqlDB, mock, _ := sqlmock.New()
		defer sqlDB.Close()

		gormDB, _ := gorm.Open("sqlite3", sqlDB)
		gormDB.LogMode(true)
		gormDB = gormDB.Debug()

		repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

		fpdValue := 80.0

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT cluster FROM m_mapping_fpd_cluster WITH (nolock) 
                            WHERE (fpd_start_hte <= ? OR fpd_start_hte IS NULL) 
                            AND (fpd_end_lt > ? OR fpd_end_lt IS NULL)`)).
			WithArgs(fpdValue, fpdValue).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		result, err := repo.MasterMappingFpdCluster(fpdValue)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error case", func(t *testing.T) {
		os.Setenv("DEFAULT_TIMEOUT_30S", "30")

		sqlDB, mock, _ := sqlmock.New()
		defer sqlDB.Close()

		gormDB, _ := gorm.Open("sqlite3", sqlDB)
		gormDB.LogMode(true)
		gormDB = gormDB.Debug()

		repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

		fpdValue := 80.0
		expectedErr := errors.New("database error")

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT cluster FROM m_mapping_fpd_cluster WITH (nolock) 
                            WHERE (fpd_start_hte <= ? OR fpd_start_hte IS NULL) 
                            AND (fpd_end_lt > ? OR fpd_end_lt IS NULL)`)).
			WithArgs(fpdValue, fpdValue).
			WillReturnError(expectedErr)
		mock.ExpectCommit()

		result, err := repo.MasterMappingFpdCluster(fpdValue)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestMasterMappingCluster(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
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

		expectedResponse := entity.MasterMappingCluster{
			BranchID:       "426",
			CustomerStatus: "NEW",
			BpkbNameType:   1,
			Cluster:        "Cluster C",
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM dbo.kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?")).
			WithArgs(req.BranchID, req.CustomerStatus, req.BpkbNameType).
			WillReturnRows(sqlmock.NewRows([]string{"branch_id", "customer_status", "bpkb_name_type", "cluster"}).
				AddRow(expectedResponse.BranchID, expectedResponse.CustomerStatus, expectedResponse.BpkbNameType, expectedResponse.Cluster))
		mock.ExpectCommit()

		result, err := repo.MasterMappingCluster(req)

		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("record not found case", func(t *testing.T) {
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
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		result, err := repo.MasterMappingCluster(req)

		assert.NoError(t, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error case", func(t *testing.T) {
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

		expectedErr := errors.New("database error")

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM dbo.kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?")).
			WithArgs(req.BranchID, req.CustomerStatus, req.BpkbNameType).
			WillReturnError(expectedErr)
		mock.ExpectCommit()

		result, err := repo.MasterMappingCluster(req)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Empty(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSaveFiltering(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	fixedTime := time.Date(2025, 1, 31, 9, 32, 13, 0, time.UTC)

	t.Run("success case - with CMO and biro data", func(t *testing.T) {
		trxFiltering := entity.FilteringKMB{
			ProspectID: "TST001",
			Decision:   "PASS",
			CreatedAt:  fixedTime,
		}
		filtering := structToSlice(trxFiltering)

		trxDetailBiro := []entity.TrxDetailBiro{
			{
				ProspectID: "TST001",
				CreatedAt:  fixedTime,
			},
		}
		detailBiro := structToSlice(trxDetailBiro[0])

		trxCmoNoFPD := entity.TrxCmoNoFPD{
			ProspectID:              "TST001",
			CMOID:                   "105394",
			CmoCategory:             "OLD",
			CmoJoinDate:             "2020-06-12",
			DefaultCluster:          "Cluster C",
			DefaultClusterStartDate: "2024-05-14",
			DefaultClusterEndDate:   "2024-07-31",
			CreatedAt:               fixedTime,
		}
		cmoNoFPD := structToSlice(trxCmoNoFPD)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_filtering"`)).
			WithArgs(filtering...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_cmo_no_fpd"`)).
			WithArgs(cmoNoFPD...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_detail_biro"`)).
			WithArgs(detailBiro...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.SaveFiltering(trxFiltering, trxDetailBiro, trxCmoNoFPD)
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})

	t.Run("success case - without CMO data", func(t *testing.T) {
		trxFiltering := entity.FilteringKMB{
			ProspectID: "TST002",
			Decision:   "PASS",
			CreatedAt:  fixedTime,
		}
		filtering := structToSlice(trxFiltering)

		trxDetailBiro := []entity.TrxDetailBiro{
			{
				ProspectID: "TST002",
				CreatedAt:  fixedTime,
			},
		}
		detailBiro := structToSlice(trxDetailBiro[0])

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_filtering"`)).
			WithArgs(filtering...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_detail_biro"`)).
			WithArgs(detailBiro...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.SaveFiltering(trxFiltering, trxDetailBiro, entity.TrxCmoNoFPD{})
		if err != nil {
			t.Errorf("expected no error, but got: %v", err)
		}
	})

	t.Run("error case - CMO creation fails", func(t *testing.T) {
		trxFiltering := entity.FilteringKMB{
			ProspectID: "TST004",
			Decision:   "PASS",
			CreatedAt:  fixedTime,
		}
		filtering := structToSlice(trxFiltering)

		trxCmoNoFPD := entity.TrxCmoNoFPD{
			ProspectID: "TST004",
			CMOID:      "105394",
			CreatedAt:  fixedTime,
		}
		cmoNoFPD := structToSlice(trxCmoNoFPD)

		expectedErr := fmt.Errorf("database error")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_filtering"`)).
			WithArgs(filtering...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_cmo_no_fpd"`)).
			WithArgs(cmoNoFPD...).
			WillReturnError(expectedErr)
		mock.ExpectRollback()

		err := repo.SaveFiltering(trxFiltering, nil, trxCmoNoFPD)
		if err == nil {
			t.Error("expected an error but got nil")
			return
		}

		if err.Error() != expectedErr.Error() {
			t.Errorf("expected error: %v, got: %v", expectedErr, err)
		}
	})

	t.Run("error case - insert biro data fails", func(t *testing.T) {
		trxFiltering := entity.FilteringKMB{
			ProspectID: "TST001",
			Decision:   "PASS",
			CreatedAt:  fixedTime,
		}
		filtering := structToSlice(trxFiltering)

		trxDetailBiro := []entity.TrxDetailBiro{
			{
				ProspectID: "TST001",
				CreatedAt:  fixedTime,
			},
		}
		detailBiro := structToSlice(trxDetailBiro[0])

		trxCmoNoFPD := entity.TrxCmoNoFPD{
			ProspectID:              "TST001",
			CMOID:                   "105394",
			CmoCategory:             "OLD",
			CmoJoinDate:             "2020-06-12",
			DefaultCluster:          "Cluster C",
			DefaultClusterStartDate: "2024-05-14",
			DefaultClusterEndDate:   "2024-07-31",
			CreatedAt:               fixedTime,
		}
		cmoNoFPD := structToSlice(trxCmoNoFPD)

		expectedErr := fmt.Errorf("database error")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_filtering"`)).
			WithArgs(filtering...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_cmo_no_fpd"`)).
			WithArgs(cmoNoFPD...).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_detail_biro"`)).
			WithArgs(detailBiro...).
			WillReturnError(expectedErr)
		mock.ExpectRollback()

		_ = repo.SaveFiltering(trxFiltering, trxDetailBiro, trxCmoNoFPD)
	})
}

func TestGetBannedPMKDSR(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	t.Run("success case", func(t *testing.T) {
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
	})

	t.Run("record not found case", func(t *testing.T) {
		idNumber := "nonexistent"
		date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)

		query := fmt.Sprintf(`SELECT * FROM trx_banned_pmk_dsr WITH (nolock) WHERE IDNumber = '%s' AND CAST(created_at as DATE) >= '%s'`, idNumber, date)
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := repo.GetBannedPMKDSR(idNumber)

		if err != nil {
			t.Errorf("expected nil error, but got: %s", err)
		}

		if result != (entity.TrxBannedPMKDSR{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	t.Run("general error case", func(t *testing.T) {
		idNumber := "1140024080800016"
		date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)

		query := fmt.Sprintf(`SELECT * FROM trx_banned_pmk_dsr WITH (nolock) WHERE IDNumber = '%s' AND CAST(created_at as DATE) >= '%s'`, idNumber, date)
		expectedErr := fmt.Errorf("database connection failed")
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WillReturnError(expectedErr)

		result, err := repo.GetBannedPMKDSR(idNumber)

		if err != expectedErr {
			t.Errorf("expected error '%v', but got: %v", expectedErr, err)
		}

		if result != (entity.TrxBannedPMKDSR{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

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

	t.Run("success case", func(t *testing.T) {
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
	})

	t.Run("database error case", func(t *testing.T) {
		myString := "1140024080800016"
		query := fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS my_string`, myString)
		expectedErr := fmt.Errorf("database function execution failed")

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WillReturnError(expectedErr)

		result, err := repo.GetEncB64(myString)

		if err != expectedErr {
			t.Errorf("expected error '%v', but got: %v", expectedErr, err)
		}

		if result != (entity.EncryptedString{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	t.Run("empty input case", func(t *testing.T) {
		myString := ""
		query := fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS my_string`, myString)

		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WillReturnRows(sqlmock.NewRows([]string{"my_string"}).
				AddRow(""))

		result, err := repo.GetEncB64(myString)

		if err != nil {
			t.Errorf("expected nil error, but got: %s", err)
		}

		expectedResponse := entity.EncryptedString{MyString: ""}
		if result != expectedResponse {
			t.Errorf("expected '%v', but got: '%v'", expectedResponse, result)
		}
	})

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

	t.Run("success case", func(t *testing.T) {
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
	})

	t.Run("record not found case", func(t *testing.T) {
		idNumber := "nonexistent"
		currentDate := time.Now().Format(constant.FORMAT_DATE)

		query := fmt.Sprintf(`SELECT 
		COUNT(CASE WHEN ts.source_decision = 'PMK' OR ts.source_decision = 'DSR' OR ts.source_decision = 'PRJ' THEN 1 END) as reject_pmk_dsr,
		COUNT(CASE WHEN ts.source_decision != 'PMK' AND ts.source_decision != 'DSR' AND ts.source_decision != 'PRJ' AND ts.source_decision != 'NKA' THEN 1 END) as reject_nik 
		FROM trx_status ts WITH (nolock) LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID
		WHERE ts.decision = 'REJ' AND tcp.IDNumber = '%s' AND CAST(ts.created_at as DATE) = '%s'`, idNumber, currentDate)
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WillReturnError(gorm.ErrRecordNotFound)

		result, err := repo.GetRejection(idNumber)

		if err != nil {
			t.Errorf("expected nil error, but got: %s", err)
		}

		if result != (entity.TrxReject{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	t.Run("database error case", func(t *testing.T) {
		idNumber := "1140024080800016"
		currentDate := time.Now().Format(constant.FORMAT_DATE)

		query := fmt.Sprintf(`SELECT 
		COUNT(CASE WHEN ts.source_decision = 'PMK' OR ts.source_decision = 'DSR' OR ts.source_decision = 'PRJ' THEN 1 END) as reject_pmk_dsr,
		COUNT(CASE WHEN ts.source_decision != 'PMK' AND ts.source_decision != 'DSR' AND ts.source_decision != 'PRJ' AND ts.source_decision != 'NKA' THEN 1 END) as reject_nik 
		FROM trx_status ts WITH (nolock) LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID
		WHERE ts.decision = 'REJ' AND tcp.IDNumber = '%s' AND CAST(ts.created_at as DATE) = '%s'`, idNumber, currentDate)
		expectedErr := fmt.Errorf("database connection failed")
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WillReturnError(expectedErr)

		result, err := repo.GetRejection(idNumber)

		if err != expectedErr {
			t.Errorf("expected error '%v', but got: %v", expectedErr, err)
		}

		if result != (entity.TrxReject{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

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

func TestGetMappingDukcapilVD_NewCustomerQueryError(t *testing.T) {
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

	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s'`, statusVD, customerStatus)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.MappingResultDukcapilVD{}) {
		t.Errorf("expected empty result, but got '%v'", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapilVD_ExistingCustomerQueryError(t *testing.T) {
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

	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s' AND kategori_status_konsumen='%s'`, statusVD, customerStatus, customerSegment)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.MappingResultDukcapilVD{}) {
		t.Errorf("expected empty result, but got '%v'", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapilVD_NewCustomerInvalidQueryError(t *testing.T) {
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

	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND is_valid=0`, customerStatus)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.MappingResultDukcapilVD{}) {
		t.Errorf("expected empty result, but got '%v'", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapilVD_ExistingCustomerInvalidQueryError(t *testing.T) {
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

	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND kategori_status_konsumen='%s' AND is_valid=0`, customerStatus, customerSegment)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetMappingDukcapilVD(statusVD, customerStatus, customerSegment, isValid)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.MappingResultDukcapilVD{}) {
		t.Errorf("expected empty result, but got '%v'", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapil_NewCustomerQueryError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := constant.EKYC_RTO
	statusFR := "MATCH"
	customerStatus := constant.STATUS_KONSUMEN_NEW
	customerSegment := "segment1"

	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s'`,
		statusVD, statusFR, customerStatus)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.MappingResultDukcapil{}) {
		t.Errorf("expected empty result, but got '%v'", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingDukcapil_ExistingCustomerQueryError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	statusVD := constant.EKYC_RTO
	statusFR := "MATCH"
	customerStatus := "existing"
	customerSegment := "segment1"

	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s' AND kategori_status_konsumen='%s'`,
		statusVD, statusFR, customerStatus, customerSegment)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.MappingResultDukcapil{}) {
		t.Errorf("expected empty result, but got '%v'", result)
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

	t.Run("success case", func(t *testing.T) {
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
	})

	t.Run("record not found case", func(t *testing.T) {
		cmoID := "NONEXISTENT"
		bpkbName := "NAMA TIDAK ADA"

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
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		result, err := repo.CheckCMONoFPD(cmoID, bpkbName)

		if err != nil {
			t.Errorf("expected nil error, but got: %s", err)
		}

		if result != (entity.TrxCmoNoFPD{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	t.Run("transaction begin error", func(t *testing.T) {
		cmoID := "CMO001"
		bpkbName := "NAMA SAMA"

		expectedErr := fmt.Errorf("failed to begin transaction")
		mock.ExpectBegin().WillReturnError(expectedErr)

		_, err := repo.CheckCMONoFPD(cmoID, bpkbName)

		if err != expectedErr {
			t.Errorf("expected error '%v', but got: %v", expectedErr, err)
		}
	})

	t.Run("database error case", func(t *testing.T) {
		cmoID := "CMO001"
		bpkbName := "NAMA SAMA"

		mock.ExpectBegin()
		expectedErr := fmt.Errorf("database error")
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT TOP 1 prospect_id, cmo_id, cmo_category, 
                            FORMAT(CONVERT(datetime, cmo_join_date, 127), 'yyyy-MM-dd') AS cmo_join_date, 
                            default_cluster, 
                            FORMAT(CONVERT(datetime, default_cluster_start_date, 127), 'yyyy-MM-dd') AS default_cluster_start_date, 
                            FORMAT(CONVERT(datetime, default_cluster_end_date, 127), 'yyyy-MM-dd') AS default_cluster_end_date
                          FROM dbo.trx_cmo_no_fpd WITH (nolock) 
                          WHERE cmo_id = ? AND bpkb_name = ?
                          ORDER BY created_at DESC`)).
			WithArgs(cmoID, bpkbName).
			WillReturnError(expectedErr)
		mock.ExpectCommit()

		result, err := repo.CheckCMONoFPD(cmoID, bpkbName)

		if err != expectedErr {
			t.Errorf("expected error '%v', but got: %v", expectedErr, err)
		}

		if result != (entity.TrxCmoNoFPD{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	t.Run("context timeout case", func(t *testing.T) {
		os.Setenv("DEFAULT_TIMEOUT_30S", "1")
		defer os.Setenv("DEFAULT_TIMEOUT_30S", "30")

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
			WillDelayFor(2 * time.Second)
		mock.ExpectCommit()

		result, err := repo.CheckCMONoFPD(cmoID, bpkbName)

		if err == nil {
			t.Error("expected context deadline exceeded error, but got nil")
		}

		if result != (entity.TrxCmoNoFPD{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
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

func TestGetTrxPrincipleStatus_QueryError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	nik := "1140024080800016"
	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_status WITH (nolock) WHERE IDNumber = '%s' ORDER BY created_at DESC`, nik)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetTrxPrincipleStatus(nik)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.TrxPrincipleStatus{}) {
		t.Errorf("expected empty result, but got '%v'", result)
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

func TestSavePrincipleStepOne_StepOneInsertError(t *testing.T) {
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

	expectedError := fmt.Errorf("insert step one error")

	mock.ExpectBegin()

	stepOne := structToSlice(dataStepOne)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_one" (.*)`).
		WithArgs(stepOne...).
		WillReturnError(expectedError)

	mock.ExpectRollback()

	err := repo.SavePrincipleStepOne(dataStepOne)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepOne_StatusInsertError(t *testing.T) {
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

	expectedError := fmt.Errorf("insert status error")

	mock.ExpectBegin()

	stepOne := structToSlice(dataStepOne)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_one" (.*)`).
		WithArgs(stepOne...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`INSERT INTO "trx_principle_status" (.*)`).
		WithArgs(dataStatus.ProspectID, dataStatus.IDNumber, dataStatus.Step, dataStatus.Decision, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(expectedError)

	mock.ExpectRollback()

	err := repo.SavePrincipleStepOne(dataStepOne)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepOne_TransactionError(t *testing.T) {
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

	expectedError := fmt.Errorf("transaction error")

	mock.ExpectBegin().WillReturnError(expectedError)

	err := repo.SavePrincipleStepOne(dataStepOne)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
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

func TestGetPrincipleStepOne_QueryError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	expectedError := fmt.Errorf("database error")

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)
	mock.ExpectQuery(regexp.QuoteMeta(query)).WillReturnError(expectedError)

	result, err := repo.GetPrincipleStepOne(prospectID)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if result != (entity.TrxPrincipleStepOne{}) {
		t.Errorf("expected empty result, but got '%v'", result)
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

func TestUpdatePrincipleStepOne_TransactionError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	updatedData := entity.TrxPrincipleStepOne{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	expectedError := fmt.Errorf("transaction error")
	mock.ExpectBegin().WillReturnError(expectedError)

	err := repo.UpdatePrincipleStepOne(prospectID, updatedData)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePrincipleStepOne_SelectError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	updatedData := entity.TrxPrincipleStepOne{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	expectedError := fmt.Errorf("select error")

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "trx_principle_step_one" WHERE (ProspectID = ?) ORDER BY created_at DESC LIMIT 1`)).
		WithArgs(prospectID).
		WillReturnError(expectedError)
	mock.ExpectRollback()

	err := repo.UpdatePrincipleStepOne(prospectID, updatedData)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePrincipleStepOne_UpdateError(t *testing.T) {
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

	expectedError := fmt.Errorf("update error")

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "trx_principle_step_one" WHERE (ProspectID = ?) ORDER BY created_at DESC LIMIT 1`)).
		WithArgs(existingData.ProspectID).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber", "Decision", "created_at"}).
			AddRow(existingData.ProspectID, existingData.IDNumber, existingData.Decision, existingData.CreatedAt))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_principle_step_one" SET "Decision" = ?, "IDNumber" = ?, "ProspectID" = ? WHERE (ProspectID = ?)`)).
		WithArgs(updatedData.Decision, updatedData.IDNumber, updatedData.ProspectID, updatedData.ProspectID).
		WillReturnError(expectedError)

	mock.ExpectRollback()

	err := repo.UpdatePrincipleStepOne(existingData.ProspectID, updatedData)

	if err == nil {
		t.Error("expected error but got nil")
	}

	if err.Error() != expectedError.Error() {
		t.Errorf("expected error '%v', but got '%v'", expectedError, err)
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

	fixedTime := time.Date(2024, 10, 7, 9, 48, 22, 0, time.UTC)

	dataStepTwo := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
		CreatedAt:  fixedTime,
	}

	mock.ExpectBegin()

	data := structToSliceWithTimeArgs(dataStepTwo)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_two" (.*)`).
		WithArgs(data...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE "trx_principle_status" SET "Decision" = \?, "IDNumber" = \?, "ProspectID" = \?, "Step" = \?, "updated_at" = \? WHERE \(ProspectID = \?\)`).
		WithArgs("approved", "123", "SAL-1140024080800016", 2, sqlmock.AnyArg(), "SAL-1140024080800016").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := repo.SavePrincipleStepTwo(dataStepTwo)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepTwo_CreateError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	fixedTime := time.Date(2024, 10, 7, 9, 48, 22, 0, time.UTC)

	dataStepTwo := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
		CreatedAt:  fixedTime,
	}

	mock.ExpectBegin()

	data := structToSliceWithTimeArgs(dataStepTwo)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_two" (.*)`).
		WithArgs(data...).
		WillReturnError(fmt.Errorf("database error"))

	mock.ExpectRollback()

	err := repo.SavePrincipleStepTwo(dataStepTwo)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("expected error 'database error', got '%v'", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepTwo_UpdateError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	fixedTime := time.Date(2024, 10, 7, 9, 48, 22, 0, time.UTC)

	dataStepTwo := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
		CreatedAt:  fixedTime,
	}

	mock.ExpectBegin()

	data := structToSliceWithTimeArgs(dataStepTwo)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_two" (.*)`).
		WithArgs(data...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE "trx_principle_status" SET "Decision" = \?, "IDNumber" = \?, "ProspectID" = \?, "Step" = \?, "updated_at" = \? WHERE \(ProspectID = \?\)`).
		WithArgs("approved", "123", "SAL-1140024080800016", 2, sqlmock.AnyArg(), "SAL-1140024080800016").
		WillReturnError(fmt.Errorf("update error"))

	mock.ExpectRollback()

	err := repo.SavePrincipleStepTwo(dataStepTwo)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "update error" {
		t.Errorf("expected error 'update error', got '%v'", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePrincipleStepTwo(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	existingData := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "pending",
		CreatedAt:  time.Now().Add(-time.Hour),
	}

	updatedData := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "trx_principle_step_two" WHERE (ProspectID = ?) ORDER BY created_at DESC LIMIT 1`)).
		WithArgs(existingData.ProspectID).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber", "Decision", "created_at"}).
			AddRow(existingData.ProspectID, existingData.IDNumber, existingData.Decision, existingData.CreatedAt))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_principle_step_two" SET "Decision" = ?, "IDNumber" = ?, "ProspectID" = ? WHERE (ProspectID = ?)`)).
		WithArgs(updatedData.Decision, updatedData.IDNumber, updatedData.ProspectID, updatedData.ProspectID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err := repo.UpdatePrincipleStepTwo(existingData.ProspectID, updatedData)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePrincipleStepTwo_RecordNotFound(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	updatedData := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "trx_principle_step_two" WHERE (ProspectID = ?) ORDER BY created_at DESC LIMIT 1`)).
		WithArgs(prospectID).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectRollback()

	err := repo.UpdatePrincipleStepTwo(prospectID, updatedData)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected error '%v', got '%v'", gorm.ErrRecordNotFound, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdatePrincipleStepTwo_UpdateError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	existingData := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "pending",
		CreatedAt:  time.Now().Add(-time.Hour),
	}

	updatedData := entity.TrxPrincipleStepTwo{
		ProspectID: "SAL-1140024080800016",
		IDNumber:   "123",
		Decision:   "approved",
	}

	mock.ExpectBegin()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "trx_principle_step_two" WHERE (ProspectID = ?) ORDER BY created_at DESC LIMIT 1`)).
		WithArgs(existingData.ProspectID).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber", "Decision", "created_at"}).
			AddRow(existingData.ProspectID, existingData.IDNumber, existingData.Decision, existingData.CreatedAt))

	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_principle_step_two" SET "Decision" = ?, "IDNumber" = ?, "ProspectID" = ? WHERE (ProspectID = ?)`)).
		WithArgs(updatedData.Decision, updatedData.IDNumber, updatedData.ProspectID, updatedData.ProspectID).
		WillReturnError(fmt.Errorf("update failed"))

	mock.ExpectRollback()

	err := repo.UpdatePrincipleStepTwo(existingData.ProspectID, updatedData)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "update failed" {
		t.Errorf("expected error 'update failed', got '%v'", err)
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

func TestGetPrincipleStepTwo_DatabaseError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_two WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnError(fmt.Errorf("database error"))

	result, err := repo.GetPrincipleStepTwo(prospectID)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("expected error 'database error', got '%v'", err)
	}

	if result != (entity.TrxPrincipleStepTwo{}) {
		t.Errorf("expected empty result, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetPrincipleStepTwo_NoRows(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"
	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_two WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(sqlmock.NewRows([]string{"ProspectID", "IDNumber"}))

	result, err := repo.GetPrincipleStepTwo(prospectID)

	if err == nil {
		t.Error("expected error for no rows, got nil")
	}

	if err != gorm.ErrRecordNotFound {
		t.Errorf("expected error '%v', got '%v'", gorm.ErrRecordNotFound, err)
	}

	if result != (entity.TrxPrincipleStepTwo{}) {
		t.Errorf("expected empty result, got %v", result)
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

	query := `SELECT bpkb_name, customer_status, customer_status_kmb, decision, reason, is_blacklist, next_process, max_overdue_biro, max_overdue_last12months_biro, customer_segment, total_baki_debet_non_collateral_biro, total_installment_amount_biro, score_biro, cluster, cmo_cluster, FORMAT(rrd_date, 'yyyy-MM-ddTHH:mm:ss') + 'Z' AS rrd_date, created_at FROM trx_filtering WITH (nolock) WHERE prospect_id = ?`

	t.Run("success case", func(t *testing.T) {
		prospectID := "TEST0001"

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(prospectID).
			WillReturnRows(sqlmock.NewRows([]string{"bpkb_name", "customer_status", "decision"}).
				AddRow("O", "NEW", "PASS"))
		mock.ExpectCommit()

		result, err := repo.GetFilteringResult(prospectID)
		if err != nil {
			t.Errorf("unexpected error: %s", err)
		}

		expected := entity.FilteringKMB{
			BpkbName:       "O",
			CustomerStatus: "NEW",
			Decision:       "PASS",
		}

		if result != expected {
			t.Errorf("expected %v, got %v", expected, result)
		}
	})

	t.Run("record not found case", func(t *testing.T) {
		prospectID := "NONEXISTENT"

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(prospectID).
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectCommit()

		result, err := repo.GetFilteringResult(prospectID)

		if err.Error() != constant.RECORD_NOT_FOUND {
			t.Errorf("expected error '%s', but got: %v", constant.RECORD_NOT_FOUND, err)
		}

		if result != (entity.FilteringKMB{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	t.Run("database error case", func(t *testing.T) {
		prospectID := "TEST0001"

		mock.ExpectBegin()
		expectedErr := fmt.Errorf("database error")
		mock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(prospectID).
			WillReturnError(expectedErr)
		mock.ExpectCommit()

		result, err := repo.GetFilteringResult(prospectID)

		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("expected error '%v', but got: %v", expectedErr, err)
		}

		if result != (entity.FilteringKMB{}) {
			t.Errorf("expected empty struct, but got: %v", result)
		}
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
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

func TestGetMappingElaborateLTV_DatabaseError(t *testing.T) {
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
		WillReturnError(fmt.Errorf("database error"))
	mock.ExpectCommit()

	result, err := repo.GetMappingElaborateLTV(resultPefindo, cluster)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("expected error 'database error', got '%v'", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
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

func TestSaveTrxElaborateLTV_ErrorConditions(t *testing.T) {
	os.Setenv("DEFAULT_TIMEOUT_10S", "10")

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	data := entity.TrxElaborateLTV{
		ProspectID:        "SAL-1140024080800016",
		RequestID:         "REQ-001",
		Tenor:             12,
		ManufacturingYear: "2023",
	}

	t.Run("Error on Update query", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_elaborate_ltv" SET "created_at" = ?, "manufacturing_year" = ?, "request_id" = ?, "tenor" = ? WHERE (prospect_id = ?)`)).
			WithArgs(sqlmock.AnyArg(), data.ManufacturingYear, data.RequestID, data.Tenor, data.ProspectID).
			WillReturnError(fmt.Errorf("update query failed"))
		mock.ExpectCommit()

		err := repo.SaveTrxElaborateLTV(data)
		if err == nil {
			t.Errorf("expected error but got none")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})

	t.Run("Error on Create query", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "trx_elaborate_ltv" SET "created_at" = ?, "manufacturing_year" = ?, "request_id" = ?, "tenor" = ? WHERE (prospect_id = ?)`)).
			WithArgs(sqlmock.AnyArg(), data.ManufacturingYear, data.RequestID, data.Tenor, data.ProspectID).
			WillReturnResult(sqlmock.NewResult(1, 0)) // No rows affected
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_elaborate_ltv" ("prospect_id","request_id","tenor","manufacturing_year","m_mapping_elaborate_ltv_id","created_at") VALUES (?,?,?,?,?,?)`)).
			WithArgs(data.ProspectID, data.RequestID, data.Tenor, data.ManufacturingYear, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(fmt.Errorf("create query failed"))
		mock.ExpectCommit()

		err := repo.SaveTrxElaborateLTV(data)
		if err == nil {
			t.Errorf("expected error but got none")
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %v", err)
		}
	})
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

func TestGetMappingVehicleAge_ErrorCondition(t *testing.T) {
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
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetMappingVehicleAge(vehicleAge, cluster, bpkbNameType, tenor, resultPefindo, af)

	if err != nil {
		t.Errorf("expected no error for gorm.ErrRecordNotFound, but got: %v", err)
	}

	if result != (entity.MappingVehicleAge{}) {
		t.Errorf("expected an empty result but got %+v", result)
	}

	mock.ExpectQuery(expectedQuery).
		WithArgs(vehicleAge, vehicleAge, fmt.Sprintf("%%%s%%", cluster), bpkbNameType, tenor, tenor, fmt.Sprintf("%%%s%%", resultPefindo), af, af).
		WillReturnError(fmt.Errorf("query failed"))

	result, err = repo.GetMappingVehicleAge(vehicleAge, cluster, bpkbNameType, tenor, resultPefindo, af)

	if err == nil {
		t.Errorf("expected an error but got none")
	}

	if result != (entity.MappingVehicleAge{}) {
		t.Errorf("expected an empty result but got %+v", result)
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

func TestGetScoreGenerator_DatabaseError(t *testing.T) {
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
		WillReturnError(fmt.Errorf("database error"))

	result, err := repo.GetScoreGenerator(zipCode)

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("expected error 'database error', got '%v'", err)
	}

	if result != (entity.ScoreGenerator{}) {
		t.Errorf("expected empty result, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
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

func TestGetScoreGeneratorROAO_DatabaseError(t *testing.T) {
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
		WillReturnError(fmt.Errorf("database error"))

	result, err := repo.GetScoreGeneratorROAO()

	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "database error" {
		t.Errorf("expected error 'database error', got '%v'", err)
	}

	if result != (entity.ScoreGenerator{}) {
		t.Errorf("expected empty result, got %v", result)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
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

func TestGetTrxDetailBIro_ErrorConditions(t *testing.T) {
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

	mock.ExpectQuery(expectedQuery).WillReturnError(errors.New("query failed"))

	mock.ExpectCommit()

	_, err := repo.GetTrxDetailBIro(prospectID)

	if err == nil {
		t.Errorf("expected an error, but got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetTrxDetailBIro_RecordNotFound(t *testing.T) {
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

	mock.ExpectQuery(expectedQuery).WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectCommit()

	result, err := repo.GetTrxDetailBIro(prospectID)

	if err != nil {
		t.Errorf("expected nil error, but got %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, but got %+v", result)
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

func TestGetActiveLoanTypeLast6M_Error(t *testing.T) {
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

	testCases := []struct {
		name          string
		mockBehavior  func()
		expectedErr   error
		expectedScore entity.GetActiveLoanTypeLast6M
	}{
		{
			name: "Database Error",
			mockBehavior: func() {
				mock.ExpectQuery(expectedQuery).WillReturnError(fmt.Errorf("database error"))
			},
			expectedErr:   fmt.Errorf("database error"),
			expectedScore: entity.GetActiveLoanTypeLast6M{},
		},
		{
			name: "Record Not Found",
			mockBehavior: func() {
				mock.ExpectQuery(expectedQuery).WillReturnError(gorm.ErrRecordNotFound)
			},
			expectedErr:   nil,
			expectedScore: entity.GetActiveLoanTypeLast6M{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.mockBehavior()

			result, err := repo.GetActiveLoanTypeLast6M(customerID)

			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.expectedScore, result)
		})
	}

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

func TestGetActiveLoanTypeLast24M_DatabaseError(t *testing.T) {
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

	dbError := fmt.Errorf("database connection error")
	mock.ExpectQuery(expectedQuery).WillReturnError(dbError)

	result, err := repo.GetActiveLoanTypeLast24M(customerID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	assert.Empty(t, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetActiveLoanTypeLast24M_RecordNotFound(t *testing.T) {
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

	mock.ExpectQuery(expectedQuery).WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetActiveLoanTypeLast24M(customerID)

	assert.NoError(t, err)
	assert.Empty(t, result)

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

func TestGetMoblast_DatabaseError(t *testing.T) {
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

	dbError := fmt.Errorf("database connection error")
	mock.ExpectQuery(expectedQuery).WillReturnError(dbError)

	result, err := repo.GetMoblast(customerID)

	assert.Error(t, err)
	assert.Equal(t, dbError, err)
	assert.Empty(t, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetMoblast_RecordNotFound(t *testing.T) {
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

	mock.ExpectQuery(expectedQuery).WillReturnError(gorm.ErrRecordNotFound)

	result, err := repo.GetMoblast(customerID)

	assert.NoError(t, err)
	assert.Empty(t, result)

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

	mock.ExpectBegin()

	stepThree := structToSlice(dataStepThree)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_three" (.*)`).
		WithArgs(stepThree...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE "trx_principle_status" SET "Decision" = \?, "IDNumber" = \?, "ProspectID" = \?, "Step" = \?, "updated_at" = \? WHERE \(ProspectID = \?\)`).
		WithArgs("approved", "123", "SAL-1140024080800016", 3, sqlmock.AnyArg(), "SAL-1140024080800016").
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := repo.SavePrincipleStepThree(dataStepThree)

	if err != nil {
		t.Errorf("error '%s' was not expected, but got: %s", err, err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepThree_CreateError(t *testing.T) {
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

	mock.ExpectBegin()

	// Simulate error during Create operation
	stepThree := structToSlice(dataStepThree)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_three" (.*)`).
		WithArgs(stepThree...).
		WillReturnError(fmt.Errorf("create operation failed"))

	// Expect rollback due to error
	mock.ExpectRollback()

	err := repo.SavePrincipleStepThree(dataStepThree)

	// Verify error is returned
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create operation failed")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepThree_UpdateError(t *testing.T) {
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

	mock.ExpectBegin()

	stepThree := structToSlice(dataStepThree)
	mock.ExpectExec(`INSERT INTO "trx_principle_step_three" (.*)`).
		WithArgs(stepThree...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE "trx_principle_status" SET "Decision" = \?, "IDNumber" = \?, "ProspectID" = \?, "Step" = \?, "updated_at" = \? WHERE \(ProspectID = \?\)`).
		WithArgs("approved", "123", "SAL-1140024080800016", 3, sqlmock.AnyArg(), "SAL-1140024080800016").
		WillReturnError(fmt.Errorf("update operation failed"))

	mock.ExpectRollback()

	err := repo.SavePrincipleStepThree(dataStepThree)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update operation failed")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSavePrincipleStepThree_TransactionError(t *testing.T) {
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

	mock.ExpectBegin().WillReturnError(fmt.Errorf("transaction begin failed"))

	err := repo.SavePrincipleStepThree(dataStepThree)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction begin failed")

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

func TestGetPrincipleStepThree_DatabaseError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_three WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	expectedError := fmt.Errorf("database error")
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnError(expectedError)

	result, err := repo.GetPrincipleStepThree(prospectID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)

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
		ProspectID:   "SAL-1140024080800016",
		Name:         "Test",
		Address:      "",
		Relationship: "",
		MobilePhone:  "",
		Rt:           "",
		Rw:           "",
		Kelurahan:    "",
		Kecamatan:    "",
		City:         "",
		Province:     "",
		ZipCode:      "",
		AreaPhone:    "",
		Phone:        "",
		CustomerID:   0,
		KPMID:        0,
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
				data.Address,
				data.Rt, data.Rw, data.Kelurahan, data.Kecamatan, data.City,
				data.Province, data.ZipCode, data.AreaPhone, data.Phone,
				data.CustomerID, data.KPMID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(`UPDATE "trx_principle_status" SET (.*) WHERE .*ProspectID = \?`).
			WithArgs(constant.DECISION_CREDIT_PROCESS, idNumber, data.ProspectID, 4, sqlmock.AnyArg(), data.ProspectID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)
		assert.NoError(t, err)
	})

	t.Run("Update Existing Record", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID"}).AddRow(data.ProspectID))

		mock.ExpectExec(`UPDATE "trx_principle_emergency_contact" SET "Name" = \?, "ProspectID" = \?, "updated_at" = \?  WHERE \(ProspectID = \?\)`).
			WithArgs(data.Name, data.ProspectID, sqlmock.AnyArg(), data.ProspectID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`UPDATE "trx_principle_status" SET (.*) WHERE .*ProspectID = \? AND Step = \?`).
			WithArgs(constant.DECISION_CREDIT_PROCESS, idNumber, sqlmock.AnyArg(), data.ProspectID, 4).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)
		assert.NoError(t, err)
	})

	t.Run("Error - Create Record Fails", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnError(gorm.ErrRecordNotFound)

		expectedError := fmt.Errorf("failed to create record")
		mock.ExpectExec(`INSERT INTO "trx_principle_emergency_contact" (.*)`).
			WithArgs(
				data.ProspectID, data.Name, data.Relationship, data.MobilePhone,
				data.Address,
				data.Rt, data.Rw, data.Kelurahan, data.Kecamatan, data.City,
				data.Province, data.ZipCode, data.AreaPhone, data.Phone,
				data.CustomerID, data.KPMID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnError(expectedError)

		mock.ExpectRollback()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error - Status Update Fails After Create", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectExec(`INSERT INTO "trx_principle_emergency_contact" (.*)`).
			WithArgs(
				data.ProspectID, data.Name, data.Relationship, data.MobilePhone,
				data.Address,
				data.Rt, data.Rw, data.Kelurahan, data.Kecamatan, data.City,
				data.Province, data.ZipCode, data.AreaPhone, data.Phone,
				data.CustomerID, data.KPMID, sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		expectedError := fmt.Errorf("failed to update status")
		mock.ExpectExec(`UPDATE "trx_principle_status" SET (.*) WHERE .*ProspectID = \?`).
			WithArgs(constant.DECISION_CREDIT_PROCESS, idNumber, data.ProspectID, 4, sqlmock.AnyArg(), data.ProspectID).
			WillReturnError(expectedError)

		mock.ExpectRollback()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error - Update Record Fails", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID"}).AddRow(data.ProspectID))

		expectedError := fmt.Errorf("failed to update record")
		mock.ExpectExec(`UPDATE "trx_principle_emergency_contact" SET "Name" = \?, "ProspectID" = \?, "updated_at" = \?  WHERE \(ProspectID = \?\)`).
			WithArgs(data.Name, data.ProspectID, sqlmock.AnyArg(), data.ProspectID).
			WillReturnError(expectedError)

		mock.ExpectRollback()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error - Status Update Fails After Update", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnRows(sqlmock.NewRows([]string{"ProspectID"}).AddRow(data.ProspectID))

		mock.ExpectExec(`UPDATE "trx_principle_emergency_contact" SET "Name" = \?, "ProspectID" = \?, "updated_at" = \?  WHERE \(ProspectID = \?\)`).
			WithArgs(data.Name, data.ProspectID, sqlmock.AnyArg(), data.ProspectID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		expectedError := fmt.Errorf("failed to update status")
		mock.ExpectExec(`UPDATE "trx_principle_status" SET (.*) WHERE .*ProspectID = \? AND Step = \?`).
			WithArgs(constant.DECISION_CREDIT_PROCESS, idNumber, sqlmock.AnyArg(), data.ProspectID, 4).
			WillReturnError(expectedError)

		mock.ExpectRollback()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error - Select Query Returns Non-NotFound Error", func(t *testing.T) {
		mock.ExpectBegin()

		expectedError := fmt.Errorf("database connection error")
		mock.ExpectQuery(`SELECT TOP 1 \* FROM trx_principle_emergency_contact WHERE ProspectID = \?`).
			WithArgs(data.ProspectID).
			WillReturnError(expectedError)

		mock.ExpectRollback()

		err := repo.SavePrincipleEmergencyContact(data, idNumber)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
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

func TestGetPrincipleEmergencyContact_DatabaseError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_emergency_contact WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	expectedError := fmt.Errorf("database error")
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnError(expectedError)

	result, err := repo.GetPrincipleEmergencyContact(prospectID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)

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

func TestSaveToWorker_DatabaseError(t *testing.T) {
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

	mock.ExpectExec(`INSERT INTO "trx_worker" (.+) VALUES (.+)`).
		WithArgs(
			workers[0].ProspectID,
			workers[0].Activity,
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
		WillReturnError(fmt.Errorf("database insert error"))

	mock.ExpectRollback()

	err := repo.SaveToWorker(workers)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database insert error")

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

func TestGetElaborateLtv_DatabaseError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"

	query := fmt.Sprintf(`SELECT CASE WHEN mmel.ltv IS NULL THEN mmelovd.ltv ELSE mmel.ltv END AS ltv FROM trx_elaborate_ltv tel WITH (nolock) 
	LEFT JOIN m_mapping_elaborate_ltv mmel WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmel.id
	LEFT JOIN m_mapping_elaborate_ltv_ovd mmelovd WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmelovd.id 
	WHERE tel.prospect_id ='%s'`, prospectID)

	expectedError := fmt.Errorf("database query error")
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnError(expectedError)

	result, err := repo.GetElaborateLtv(prospectID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)

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
		InstallmentAmount:          0,
		NTF:                        0,
		OTR:                        0,
		Dealer:                     "PSA",
		AssetCategoryID:            "MATIC",
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
			data.InstallmentAmount,
			data.NTF,
			data.OTR,
			data.Dealer,
			data.AssetCategoryID,
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

func TestSavePrincipleMarketingProgram_DatabaseError(t *testing.T) {
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
		InstallmentAmount:          0,
		NTF:                        0,
		OTR:                        0,
		Dealer:                     "PSA",
		AssetCategoryID:            "MATIC",
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
			data.InstallmentAmount,
			data.NTF,
			data.OTR,
			data.Dealer,
			data.AssetCategoryID,
			sqlmock.AnyArg(),
		).
		WillReturnError(fmt.Errorf("database insert error"))

	mock.ExpectRollback()

	err := repo.SavePrincipleMarketingProgram(data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database insert error")

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

func TestGetPrincipleMarketingProgram_DatabaseError(t *testing.T) {
	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	prospectID := "SAL-1140024080800016"

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	expectedError := fmt.Errorf("database error")
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnError(expectedError)

	result, err := repo.GetPrincipleMarketingProgram(prospectID)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, result)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetTrxWorker(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	fixedTime := time.Now()
	dbError := fmt.Errorf("database error")

	testCases := []struct {
		name          string
		prospectID    string
		category      string
		expectedQuery string
		mockRows      *sqlmock.Rows
		mockError     error
		expectedError error
		expectedData  []entity.TrxWorker
	}{
		{
			name:          "Success Case",
			prospectID:    "PROSPECT-123",
			category:      "TEST",
			expectedQuery: `SELECT \* FROM trx_worker WITH \(nolock\) WHERE ProspectID = '.+' AND \[category\] = '.+'`,
			mockRows: sqlmock.NewRows([]string{
				"ProspectID",
				"category",
				"activity",
				"EndPointTarget",
				"EndPointMethod",
				"Payload",
				"Header",
				"ResponseTimeout",
				"APIType",
				"MaxRetry",
				"CountRetry",
				"created_at",
				"Action",
				"StatusCode",
				"Sequence",
			}).AddRow(
				"PROSPECT-123",
				"TEST",
				"Sample Activity",
				"",
				"",
				"",
				"",
				0,
				"",
				0,
				0,
				fixedTime,
				"",
				"",
				nil,
			),
			mockError:     nil,
			expectedError: nil,
			expectedData: []entity.TrxWorker{
				{
					ProspectID: "PROSPECT-123",
					Category:   "TEST",
					Activity:   "Sample Activity",
					CreatedAt:  fixedTime,
				},
			},
		},
		{
			name:          "No Records Found",
			prospectID:    "PROSPECT-456",
			category:      "TEST",
			expectedQuery: `SELECT \* FROM trx_worker WITH \(nolock\) WHERE ProspectID = '.+' AND \[category\] = '.+'`,
			mockRows: sqlmock.NewRows([]string{
				"ProspectID",
				"category",
				"activity",
				"EndPointTarget",
				"EndPointMethod",
				"Payload",
				"Header",
				"ResponseTimeout",
				"APIType",
				"MaxRetry",
				"CountRetry",
				"created_at",
				"Action",
				"StatusCode",
				"Sequence",
			}),
			mockError:     nil,
			expectedError: nil,
			expectedData:  []entity.TrxWorker{},
		},
		{
			name:          "Database Error",
			prospectID:    "PROSPECT-789",
			category:      "TEST",
			expectedQuery: `SELECT \* FROM trx_worker WITH \(nolock\) WHERE ProspectID = '.+' AND \[category\] = '.+'`,
			mockRows:      sqlmock.NewRows([]string{}),
			mockError:     dbError,
			expectedError: dbError,
			expectedData:  nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryExp := mock.ExpectQuery(tc.expectedQuery)
			if tc.mockError != nil {
				queryExp.WillReturnError(tc.mockError)
			} else {
				queryExp.WillReturnRows(tc.mockRows)
			}

			result, err := repo.GetTrxWorker(tc.prospectID, tc.category)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				if len(result) > 0 {
					assert.Equal(t, tc.expectedData[0].ProspectID, result[0].ProspectID)
					assert.Equal(t, tc.expectedData[0].Category, result[0].Category)
					assert.Equal(t, tc.expectedData[0].Activity, result[0].Activity)
				} else {
					assert.Equal(t, tc.expectedData, result)
				}
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestScanOrderPending(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		mockRows      *sqlmock.Rows
		mockError     error
		expectedData  []entity.AutoCancel
		expectedError error
	}{
		{
			name: "Success with Multiple Records",
			mockRows: sqlmock.NewRows([]string{
				"ProspectID",
				"KPMID",
				"BranchID",
				"AssetCode",
			}).AddRow(
				"PROS-001",
				123,
				"BR001",
				"ASSET001",
			).AddRow(
				"PROS-002",
				124,
				"BR002",
				"ASSET002",
			),
			mockError: nil,
			expectedData: []entity.AutoCancel{
				{
					ProspectID: "PROS-001",
					KPMID:      123,
					BranchID:   "BR001",
					AssetCode:  "ASSET001",
				},
				{
					ProspectID: "PROS-002",
					KPMID:      124,
					BranchID:   "BR002",
					AssetCode:  "ASSET002",
				},
			},
			expectedError: nil,
		},
		{
			name:          "Success with No Records",
			mockRows:      sqlmock.NewRows([]string{"ProspectID", "KPMID", "BranchID", "AssetCode"}),
			mockError:     nil,
			expectedData:  []entity.AutoCancel{},
			expectedError: nil,
		},
		{
			name:          "Database Error",
			mockRows:      nil,
			mockError:     fmt.Errorf("database error"),
			expectedData:  nil,
			expectedError: fmt.Errorf("database error"),
		},
	}

	query := `SELECT DISTINCT tps.ProspectID, tpso.KPMID, tpso.BranchID, tpso.AssetCode  
    FROM trx_principle_status tps WITH (nolock)
    INNER JOIN trx_principle_step_one tpso WITH (nolock)
    ON tps.ProspectID  = tpso.ProspectID 
    WHERE tps.created_at < DATEADD(day, -3, GETDATE())
    AND tps.Decision <> 'CANCEL' `

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tc.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tc.mockRows)
			}

			result, err := repo.ScanOrderPending()

			if tc.expectedError != nil {
				if assert.Error(t, err) {
					assert.Equal(t, tc.expectedError.Error(), err.Error())
				}
				assert.Empty(t, result)
			} else {
				if !assert.NoError(t, err) {
					return
				}
				assert.Equal(t, len(tc.expectedData), len(result))

				if len(result) > 0 {
					for i, expectedItem := range tc.expectedData {
						assert.Equal(t, expectedItem.ProspectID, result[i].ProspectID)
						assert.Equal(t, expectedItem.KPMID, result[i].KPMID)
						assert.Equal(t, expectedItem.BranchID, result[i].BranchID)
						assert.Equal(t, expectedItem.AssetCode, result[i].AssetCode)
					}
				}
			}
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestUpdateToCancel(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		prospectID    string
		mockError     error
		expectedError error
	}{
		{
			name:          "Success Case",
			prospectID:    "PROS-001",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "Database Error",
			prospectID:    "PROS-002",
			mockError:     fmt.Errorf("database error"),
			expectedError: fmt.Errorf("database error"),
		},
	}

	updateSQL := `UPDATE "trx_principle_status" SET "Decision" = \?, "updated_at" = \? WHERE \(ProspectID = \?\)`

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectBegin()

			if tc.mockError != nil {
				mock.ExpectExec(updateSQL).
					WithArgs(constant.DECISION_CANCEL, sqlmock.AnyArg(), tc.prospectID).
					WillReturnError(tc.mockError)
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(updateSQL).
					WithArgs(constant.DECISION_CANCEL, sqlmock.AnyArg(), tc.prospectID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}

			err := repo.UpdateToCancel(tc.prospectID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdateTrxPrincipleStatus(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		prospectID    string
		decision      string
		step          int
		mockError     error
		expectedError error
	}{
		{
			name:          "Success Case",
			prospectID:    "PROS-001",
			decision:      "APPROVED",
			step:          2,
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "Database Error",
			prospectID:    "PROS-002",
			decision:      "REJECTED",
			step:          3,
			mockError:     fmt.Errorf("database error"),
			expectedError: fmt.Errorf("database error"),
		},
	}

	updateSQL := `UPDATE "trx_principle_status" SET "Decision" = \?, "Step" = \?, "updated_at" = \? WHERE \(ProspectID = \?\)`

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectBegin()

			if tc.mockError != nil {
				mock.ExpectExec(updateSQL).
					WithArgs(tc.decision, tc.step, sqlmock.AnyArg(), tc.prospectID).
					WillReturnError(tc.mockError)
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(updateSQL).
					WithArgs(tc.decision, tc.step, sqlmock.AnyArg(), tc.prospectID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}

			err := repo.UpdateTrxPrincipleStatus(tc.prospectID, tc.decision, tc.step)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestExceedErrorStepOne(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		kpmID         int
		rowsAffected  int64
		expectedCount int
	}{
		{
			name:          "Record Found",
			kpmID:         12345,
			rowsAffected:  1,
			expectedCount: 1,
		},
		{
			name:          "No Records Found",
			kpmID:         67890,
			rowsAffected:  0,
			expectedCount: 0,
		},
	}

	query := "SELECT KpmID FROM trx_principle_error WITH \\(nolock\\) WHERE KpmID = \\? AND step = 1 AND created_at >= DATEADD \\(HOUR , -1 , GETDATE\\(\\)\\)"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"KpmID"})
			if tc.rowsAffected > 0 {
				rows.AddRow(tc.kpmID)
			}

			mock.ExpectQuery(query).
				WithArgs(tc.kpmID).
				WillReturnRows(rows)

			result := repo.ExceedErrorStepOne(tc.kpmID)

			assert.Equal(t, tc.expectedCount, result)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestExceedErrorStepTwo(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		prospectID    string
		rowsAffected  int64
		expectedCount int
	}{
		{
			name:          "Record Found",
			prospectID:    "PROS-001",
			rowsAffected:  1,
			expectedCount: 1,
		},
		{
			name:          "No Records Found",
			prospectID:    "PROS-002",
			rowsAffected:  0,
			expectedCount: 0,
		},
	}

	query := "SELECT KpmID FROM trx_principle_error WITH \\(nolock\\) WHERE ProspectID = \\? AND step = 2 AND created_at >= DATEADD \\(HOUR , -1 , GETDATE\\(\\)\\)"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"KpmID"})
			if tc.rowsAffected > 0 {
				rows.AddRow(123)
			}

			mock.ExpectQuery(query).
				WithArgs(tc.prospectID).
				WillReturnRows(rows)

			result := repo.ExceedErrorStepTwo(tc.prospectID)

			assert.Equal(t, tc.expectedCount, result)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestExceedErrorStepThree(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		prospectID    string
		rowsAffected  int64
		expectedCount int
	}{
		{
			name:          "Record Found",
			prospectID:    "PROS-001",
			rowsAffected:  1,
			expectedCount: 1,
		},
		{
			name:          "No Records Found",
			prospectID:    "PROS-002",
			rowsAffected:  0,
			expectedCount: 0,
		},
	}

	query := "SELECT KpmID FROM trx_principle_error WITH \\(nolock\\) WHERE ProspectID = \\? AND step = 3 AND created_at >= DATEADD \\(HOUR , -1 , GETDATE\\(\\)\\)"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"KpmID"})
			if tc.rowsAffected > 0 {
				rows.AddRow(123)
			}

			mock.ExpectQuery(query).
				WithArgs(tc.prospectID).
				WillReturnRows(rows)

			result := repo.ExceedErrorStepThree(tc.prospectID)

			assert.Equal(t, tc.expectedCount, result)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetTrxStatus(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		prospectID    string
		mockRows      *sqlmock.Rows
		mockError     error
		expectedError error
		expectedData  entity.TrxStatus
	}{
		{
			name:       "Success Case",
			prospectID: "PROS-001",
			mockRows: sqlmock.NewRows([]string{"activity", "decision", "source_decision"}).
				AddRow("TEST_ACTIVITY", "APPROVED", "SOURCE1"),
			mockError:     nil,
			expectedError: nil,
			expectedData: entity.TrxStatus{
				Activity:       "TEST_ACTIVITY",
				Decision:       "APPROVED",
				SourceDecision: "SOURCE1",
			},
		},
		{
			name:          "Record Not Found",
			prospectID:    "PROS-002",
			mockRows:      sqlmock.NewRows([]string{"activity", "decision", "source_decision"}),
			mockError:     gorm.ErrRecordNotFound,
			expectedError: errors.New(constant.RECORD_NOT_FOUND),
			expectedData:  entity.TrxStatus{},
		},
	}

	query := "SELECT activity, decision, source_decision FROM trx_status WITH \\(nolock\\) WHERE ProspectID = \\?"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockError != nil {
				mock.ExpectQuery(query).
					WithArgs(tc.prospectID).
					WillReturnError(tc.mockError)
			} else {
				mock.ExpectQuery(query).
					WithArgs(tc.prospectID).
					WillReturnRows(tc.mockRows)
			}

			result, err := repo.GetTrxStatus(tc.prospectID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData, result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetBannedChassisNumber(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)

	testCases := []struct {
		name          string
		chassisNumber string
		mockRows      *sqlmock.Rows
		mockError     error
		expectedError error
		expectedData  entity.TrxBannedChassisNumber
	}{
		{
			name:          "Success Case",
			chassisNumber: "CHASSIS001",
			mockRows: sqlmock.NewRows([]string{"chassis_number", "created_at"}).
				AddRow("CHASSIS001", time.Now()),
			mockError:     nil,
			expectedError: nil,
			expectedData: entity.TrxBannedChassisNumber{
				ChassisNo: "CHASSIS001",
			},
		},
		{
			name:          "Record Not Found",
			chassisNumber: "CHASSIS002",
			mockRows:      sqlmock.NewRows([]string{"chassis_number", "created_at"}),
			mockError:     gorm.ErrRecordNotFound,
			expectedError: nil,
			expectedData:  entity.TrxBannedChassisNumber{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf(`SELECT * FROM trx_banned_chassis_number WITH (nolock) WHERE chassis_number = '%s' AND CAST(created_at as DATE) >= '%s'`, tc.chassisNumber, date)

			if tc.mockError != nil {
				mock.ExpectQuery(query).WillReturnError(tc.mockError)
			} else {
				mock.ExpectQuery(query).WillReturnRows(tc.mockRows)
			}

			result, err := repo.GetBannedChassisNumber(tc.chassisNumber)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData.ChassisNo, result.ChassisNo)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetMappingNegativeCustomer(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		request       response.NegativeCustomer
		mockRows      *sqlmock.Rows
		mockError     error
		expectedError error
		expectedData  entity.MappingNegativeCustomer
	}{
		{
			name: "Success Case",
			request: response.NegativeCustomer{
				IsActive:    1,
				BadType:     "BAD",
				IsBlacklist: 1,
				IsHighrisk:  1,
			},
			mockRows: sqlmock.NewRows([]string{"is_active", "bad_type", "is_blacklist", "is_highrisk"}).
				AddRow(1, "BAD", 1, 1),
			mockError:     nil,
			expectedError: nil,
			expectedData: entity.MappingNegativeCustomer{
				IsActive:    1,
				BadType:     "BAD",
				IsBlacklist: 1,
				IsHighrisk:  1,
			},
		},
		{
			name: "Record Not Found",
			request: response.NegativeCustomer{
				IsActive:    1,
				BadType:     "GOOD",
				IsBlacklist: 0,
				IsHighrisk:  0,
			},
			mockRows:      sqlmock.NewRows([]string{"is_active", "bad_type", "is_blacklist", "is_highrisk"}),
			mockError:     gorm.ErrRecordNotFound,
			expectedError: nil,
			expectedData:  entity.MappingNegativeCustomer{},
		},
	}

	query := "SELECT TOP 1 \\* FROM m_mapping_negative_customer WHERE is_active = \\? AND bad_type = \\? AND is_blacklist = \\? AND is_highrisk = \\?"

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.mockError != nil {
				mock.ExpectQuery(query).
					WithArgs(tc.request.IsActive, tc.request.BadType, tc.request.IsBlacklist, tc.request.IsHighrisk).
					WillReturnError(tc.mockError)
			} else {
				mock.ExpectQuery(query).
					WithArgs(tc.request.IsActive, tc.request.BadType, tc.request.IsBlacklist, tc.request.IsHighrisk).
					WillReturnRows(tc.mockRows)
			}

			result, err := repo.GetMappingNegativeCustomer(tc.request)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData, result)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSaveTrxKPM(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening gorm connection", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		inputData     entity.TrxKPM
		encryptedData entity.Encrypted
		mockError     error
		createError   error
		updateError   error
		expectedError error
	}{
		{
			name: "Success Case",
			inputData: entity.TrxKPM{
				ID:                "TRX-001",
				ProspectID:        "PROS-001",
				LegalName:         "John Doe",
				SurgateMotherName: "Jane Doe",
				MobilePhone:       "1234567890",
				Email:             "john@example.com",
				BirthPlace:        "New York",
				ResidenceAddress:  "123 Main St",
				IDNumber:          "ID123456",
				Decision:          "APPROVED",
			},
			encryptedData: entity.Encrypted{
				LegalName:         "ENC_JOHN_DOE",
				SurgateMotherName: "ENC_JANE_DOE",
				MobilePhone:       "ENC_1234567890",
				Email:             "ENC_EMAIL",
				BirthPlace:        "ENC_BIRTHPLACE",
				ResidenceAddress:  "ENC_ADDRESS",
				IDNumber:          "ENC_ID123456",
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "Encryption Error",
			inputData: entity.TrxKPM{
				ID:        "TRX-002",
				LegalName: "John Doe",
			},
			encryptedData: entity.Encrypted{},
			mockError:     fmt.Errorf("encryption error"),
			expectedError: fmt.Errorf("encryption error"),
		},
		{
			name: "Create Error",
			inputData: entity.TrxKPM{
				ID:         "TRX-003",
				ProspectID: "PROS-003",
				LegalName:  "John Doe",
				Decision:   "APPROVED",
			},
			encryptedData: entity.Encrypted{
				LegalName: "ENC_JOHN_DOE",
			},
			createError:   fmt.Errorf("create error: duplicate key"),
			expectedError: fmt.Errorf("create error: duplicate key"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectBegin()

			if tc.mockError != nil {
				encryptQuery := fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName, SCP.dbo.ENC_B64('SEC','%s') AS SurgateMotherName, SCP.dbo.ENC_B64('SEC','%s') AS MobilePhone, SCP.dbo.ENC_B64('SEC','%s') AS Email, SCP.dbo.ENC_B64('SEC','%s') AS BirthPlace, SCP.dbo.ENC_B64('SEC','%s') AS ResidenceAddress, SCP.dbo.ENC_B64('SEC','%s') AS IDNumber`,
					tc.inputData.LegalName, tc.inputData.SurgateMotherName, tc.inputData.MobilePhone,
					tc.inputData.Email, tc.inputData.BirthPlace, tc.inputData.ResidenceAddress, tc.inputData.IDNumber)

				mock.ExpectQuery(regexp.QuoteMeta(encryptQuery)).
					WillReturnError(tc.mockError)
				mock.ExpectRollback()
			} else if tc.createError != nil {
				mock.ExpectQuery(`SELECT SCP\.dbo\.ENC_B64\('SEC','.*'\) AS LegalName`).
					WillReturnRows(sqlmock.NewRows([]string{
						"LegalName", "SurgateMotherName", "MobilePhone",
						"Email", "BirthPlace", "ResidenceAddress", "IDNumber",
					}).AddRow(
						tc.encryptedData.LegalName,
						tc.encryptedData.SurgateMotherName,
						tc.encryptedData.MobilePhone,
						tc.encryptedData.Email,
						tc.encryptedData.BirthPlace,
						tc.encryptedData.ResidenceAddress,
						tc.encryptedData.IDNumber,
					))

				mock.ExpectExec(`INSERT INTO "trx_kpm"`).
					WillReturnError(tc.createError)
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(`SELECT SCP\.dbo\.ENC_B64\('SEC','.*'\) AS LegalName`).
					WillReturnRows(sqlmock.NewRows([]string{
						"LegalName", "SurgateMotherName", "MobilePhone",
						"Email", "BirthPlace", "ResidenceAddress", "IDNumber",
					}).AddRow(
						tc.encryptedData.LegalName,
						tc.encryptedData.SurgateMotherName,
						tc.encryptedData.MobilePhone,
						tc.encryptedData.Email,
						tc.encryptedData.BirthPlace,
						tc.encryptedData.ResidenceAddress,
						tc.encryptedData.IDNumber,
					))

				mock.ExpectExec(`INSERT INTO "trx_kpm"`).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			}

			err := repo.SaveTrxKPM(tc.inputData)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSaveTrxKPMStatus(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		input         entity.TrxKPMStatus
		mockError     error
		expectedError error
	}{
		{
			name: "Success Case",
			input: entity.TrxKPMStatus{
				ID:         "STATUS-001",
				ProspectID: "PROS-001",
				Decision:   "APPROVED",
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "Create Error",
			input: entity.TrxKPMStatus{
				ID:         "STATUS-002",
				ProspectID: "PROS-002",
				Decision:   "REJECTED",
			},
			mockError:     fmt.Errorf("duplicate entry"),
			expectedError: fmt.Errorf("duplicate entry"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock.ExpectBegin()

			if tc.mockError != nil {
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_kpm_status" ("id","ProspectID","Decision","created_at","created_by","updated_at","updated_by","deleted_at","deleted_by")`)).
					WithArgs(
						tc.input.ID,
						tc.input.ProspectID,
						tc.input.Decision,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnError(tc.mockError)
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "trx_kpm_status" ("id","ProspectID","Decision","created_at","created_by","updated_at","updated_by","deleted_at","deleted_by")`)).
					WithArgs(
						tc.input.ID,
						tc.input.ProspectID,
						tc.input.Decision,
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
						sqlmock.AnyArg(),
					).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err := repo.SaveTrxKPMStatus(tc.input)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetTrxKPM(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		prospectID    string
		mockData      entity.TrxKPM
		decryptedData entity.Encrypted
		mockError     error
		decryptError  error
		expectedError error
	}{
		{
			name:       "Success Case",
			prospectID: "PROS-001",
			mockData: entity.TrxKPM{
				ID:         "TRX-001",
				ProspectID: "PROS-001",
				LegalName:  "ENC_NAME",
				IDNumber:   "ENC_ID",
			},
			decryptedData: entity.Encrypted{
				LegalName:         "John Doe",
				SurgateMotherName: "Jane Doe",
				MobilePhone:       "1234567890",
				Email:             "john@example.com",
				BirthPlace:        "New York",
				ResidenceAddress:  "123 Main St",
				IDNumber:          "ID123456",
			},
			mockError:     nil,
			decryptError:  nil,
			expectedError: nil,
		},
		{
			name:          "Record Not Found",
			prospectID:    "PROS-002",
			mockData:      entity.TrxKPM{},
			mockError:     gorm.ErrRecordNotFound,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:       "Decrypt Error",
			prospectID: "PROS-003",
			mockData: entity.TrxKPM{
				ID:         "TRX-003",
				ProspectID: "PROS-003",
				LegalName:  "ENC_NAME",
				IDNumber:   "ENC_ID",
			},
			decryptError:  fmt.Errorf("decrypt error"),
			expectedError: fmt.Errorf("decrypt error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initialQuery := fmt.Sprintf("SELECT TOP 1 * FROM trx_kpm WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC", tc.prospectID)

			if tc.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(initialQuery)).
					WillReturnError(tc.mockError)
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(initialQuery)).
					WillReturnRows(sqlmock.NewRows([]string{
						"id", "ProspectID", "LegalName", "IDNumber",
					}).AddRow(
						tc.mockData.ID,
						tc.mockData.ProspectID,
						tc.mockData.LegalName,
						tc.mockData.IDNumber,
					))

				if tc.mockError == nil && tc.decryptError == nil {
					decryptQuery := fmt.Sprintf(`SELECT scp.dbo.DEC_B64('SEC', '%s') AS LegalName, scp.dbo.DEC_B64('SEC','%s') AS SurgateMotherName,
                        scp.dbo.DEC_B64('SEC', '%s') AS MobilePhone, scp.dbo.DEC_B64('SEC', '%s') AS Email,
                        scp.dbo.DEC_B64('SEC', '%s') AS BirthPlace, scp.dbo.DEC_B64('SEC','%s') AS ResidenceAddress,
                        scp.dbo.DEC_B64('SEC', '%s') AS IDNumber`,
						tc.mockData.LegalName,
						tc.mockData.SurgateMotherName,
						tc.mockData.MobilePhone,
						tc.mockData.Email,
						tc.mockData.BirthPlace,
						tc.mockData.ResidenceAddress,
						tc.mockData.IDNumber)

					mock.ExpectQuery(regexp.QuoteMeta(decryptQuery)).
						WillReturnRows(sqlmock.NewRows([]string{
							"LegalName", "SurgateMotherName", "MobilePhone",
							"Email", "BirthPlace", "ResidenceAddress", "IDNumber",
						}).AddRow(
							tc.decryptedData.LegalName,
							tc.decryptedData.SurgateMotherName,
							tc.decryptedData.MobilePhone,
							tc.decryptedData.Email,
							tc.decryptedData.BirthPlace,
							tc.decryptedData.ResidenceAddress,
							tc.decryptedData.IDNumber,
						))
				} else if tc.decryptError != nil {
					mock.ExpectQuery("SELECT scp.dbo.DEC_B64").
						WillReturnError(tc.decryptError)
				}
			}

			result, err := repo.GetTrxKPM(tc.prospectID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.decryptedData.LegalName, result.LegalName)
				assert.Equal(t, tc.decryptedData.IDNumber, result.IDNumber)
				assert.Equal(t, tc.decryptedData.MobilePhone, result.MobilePhone)
				assert.Equal(t, tc.decryptedData.Email, result.Email)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestExceedErrorTrxKPM(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		kpmID         int
		rowsAffected  int64
		expectedCount int
	}{
		{
			name:          "Records Found",
			kpmID:         123,
			rowsAffected:  2,
			expectedCount: 2,
		},
		{
			name:          "No Records",
			kpmID:         456,
			rowsAffected:  0,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"KpmID"})
			for i := int64(0); i < tc.rowsAffected; i++ {
				rows.AddRow(i)
			}

			mock.ExpectQuery(`SELECT KpmID FROM trx_kpm_error WITH \(nolock\) WHERE KpmID = \? AND created_at >= DATEADD \(HOUR , -1 , GETDATE\(\)\)`).
				WithArgs(tc.kpmID).
				WillReturnRows(rows)

			result := repo.ExceedErrorTrxKPM(tc.kpmID)
			assert.Equal(t, tc.expectedCount, result)

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetReadjustCountTrxKPM(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		prospectID    string
		rowsAffected  int64
		expectedCount int
	}{
		{
			name:          "Records Found",
			prospectID:    "PROS-001",
			rowsAffected:  3,
			expectedCount: 3,
		},
		{
			name:          "No Records",
			prospectID:    "PROS-002",
			rowsAffected:  0,
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"id"})
			for i := int64(0); i < tc.rowsAffected; i++ {
				rows.AddRow(i)
			}

			mock.ExpectQuery(`SELECT id FROM trx_kpm WITH \(nolock\) WHERE ProspectID = \? AND Decision = \?`).
				WithArgs(tc.prospectID, constant.DECISION_KPM_READJUST).
				WillReturnRows(rows)

			result := repo.GetReadjustCountTrxKPM(tc.prospectID)
			assert.Equal(t, tc.expectedCount, result)
		})
	}
}

func TestGetTrxKPMStatusHistory(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlserver", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := repoHandler{newKmb: gormDB}

	now := time.Now()
	startDate := now.Add(-48 * time.Hour)
	endDate := now
	status := "APPROVED"
	loanAmount := 10.000
	testCases := []struct {
		name          string
		req           request.History2Wilen
		mockData      []entity.TrxKPMStatusHistory
		mockError     error
		expectedError error
		expectedLen   int
	}{
		{
			name: "Success - Filter by ProspectID, Date Range, and Status",
			req: request.History2Wilen{
				ProspectID: ptr("PROS-001"),
				StartDate:  &startDate,
				EndDate:    &endDate,
				Status:     &status,
			},
			mockData: []entity.TrxKPMStatusHistory{
				{
					ID:         "HIST-001",
					ProspectID: "PROS-001",
					Decision:   "APPROVED",
					CreatedAt:  now.Add(-24 * time.Hour),
					LoanAmount: &loanAmount,
				},
			},
			mockError:     nil,
			expectedError: nil,
			expectedLen:   1,
		},
		{
			name: "Success - No Matching Records",
			req: request.History2Wilen{
				ProspectID: ptr("PROS-002"),
				StartDate:  &startDate,
				EndDate:    &endDate,
				Status:     &status,
			},
			mockData:      []entity.TrxKPMStatusHistory{},
			mockError:     nil,
			expectedError: nil,
			expectedLen:   0,
		},
		{
			name: "Database Error",
			req: request.History2Wilen{
				ProspectID: ptr("PROS-003"),
			},
			mockData:      nil,
			mockError:     fmt.Errorf("database connection error"),
			expectedError: fmt.Errorf("database connection error"),
			expectedLen:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Dynamically build query
			whereQuery := "WHERE 1=1"
			args := []driver.Value{}

			if tc.req.ProspectID != nil {
				whereQuery += " AND s.ProspectID = ?"
				args = append(args, *tc.req.ProspectID)
			}
			if tc.req.StartDate != nil && tc.req.EndDate != nil {
				whereQuery += " AND s.created_at BETWEEN ? AND ?"
				args = append(args, *tc.req.StartDate, *tc.req.EndDate)
			}
			if tc.req.Status != nil {
				whereQuery += " AND s.Decision = ?"
				args = append(args, *tc.req.Status)
			}

			query := fmt.Sprintf(`
				SELECT s.ProspectID as ProspectID, s.id as id, s.Decision as Decision, s.created_at as created_at, k.KpmID as KpmID, scp.dbo.DEC_B64('SEC', k.IDNumber) as IDNumber, k.ReferralCode as ReferralCode, k.LoanAmount as LoanAmount
				FROM trx_kpm_status AS s WITH (nolock)
				LEFT JOIN (
					SELECT *, ROW_NUMBER() OVER (PARTITION BY ProspectID ORDER BY created_at DESC) AS rn
					FROM trx_kpm
				) AS k ON s.ProspectID = k.ProspectID AND k.rn = 1 %s`, whereQuery)

			if tc.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnError(tc.mockError)
			} else {
				rows := sqlmock.NewRows([]string{"id", "ProspectID", "Decision", "created_at", "KpmID", "IDNumber", "ReferralCode", "LoanAmount"})
				for _, data := range tc.mockData {
					rows.AddRow(data.ID, data.ProspectID, data.Decision, data.CreatedAt, data.KpmID, data.IDNumber, data.ReferralCode, data.LoanAmount)
				}
				mock.ExpectQuery(regexp.QuoteMeta(query)).WithArgs(args...).WillReturnRows(rows)
			}

			result, err := repo.GetTrxKPMStatusHistory(tc.req)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedLen, len(result))
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %s", err)
			}
		})
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestGetTrxKPMStatus(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name          string
		idNumber      string
		mockData      entity.TrxKPMStatus
		mockError     error
		expectedError error
	}{
		{
			name:     "Success Case",
			idNumber: "1234567890",
			mockData: entity.TrxKPMStatus{
				ID:         "STATUS-001",
				ProspectID: "PROS-001",
				Decision:   "APPROVED",
				CreatedAt:  time.Now(),
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "Record Not Found",
			idNumber:      "9999999999",
			mockData:      entity.TrxKPMStatus{},
			mockError:     gorm.ErrRecordNotFound,
			expectedError: gorm.ErrRecordNotFound,
		},
		{
			name:          "Database Error",
			idNumber:      "8888888888",
			mockData:      entity.TrxKPMStatus{},
			mockError:     fmt.Errorf("database error"),
			expectedError: fmt.Errorf("database error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query := fmt.Sprintf("SELECT TOP 1 tks.* FROM trx_kpm_status tks WITH (nolock) JOIN trx_kpm tk ON tks.ProspectID = tk.ProspectID WHERE tk.IDNumber = SCP.dbo.ENC_B64('SEC','%s') ORDER BY tks.created_at DESC", tc.idNumber)

			if tc.mockError != nil {
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnError(tc.mockError)
			} else {
				rows := sqlmock.NewRows([]string{"id", "ProspectID", "Decision", "created_at"}).
					AddRow(tc.mockData.ID, tc.mockData.ProspectID, tc.mockData.Decision, tc.mockData.CreatedAt)
				mock.ExpectQuery(regexp.QuoteMeta(query)).
					WillReturnRows(rows)
			}

			result, err := repo.GetTrxKPMStatus(tc.idNumber)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockData.ID, result.ID)
				assert.Equal(t, tc.mockData.ProspectID, result.ProspectID)
				assert.Equal(t, tc.mockData.Decision, result.Decision)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestUpdateTrxKPMDecision(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock DB: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("Failed to open GORM DB: %v", err)
	}
	gormDB.LogMode(true)
	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testID := "123"
	testProspectID := "PROS-123"
	testDecision := "APPROVED"

	t.Run("Successful Update", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE "trx_kpm" SET "Decision" = \?, "updated_at" = \? WHERE "trx_kpm"."deleted_at" IS NULL AND \(\(id = \?\)\)`).
			WithArgs(testDecision, sqlmock.AnyArg(), testID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`INSERT INTO "trx_kpm_status" \("id","ProspectID","Decision","created_at","created_by","updated_at","updated_by","deleted_at","deleted_by"\) VALUES \(\?,\?,\?,\?,\?,\?,\?,\?,\?\)`).
			WithArgs(
				sqlmock.AnyArg(),
				testProspectID,
				testDecision,
				sqlmock.AnyArg(),
				"",
				sqlmock.AnyArg(),
				"",
				nil,
				"",
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		err := repo.UpdateTrxKPMDecision(testID, testProspectID, testDecision)
		assert.NoError(t, err)
	})

	t.Run("Error - TrxKPM Update Fails", func(t *testing.T) {
		mock.ExpectBegin()

		expectedError := fmt.Errorf("failed to update TrxKPM")
		mock.ExpectExec(`UPDATE "trx_kpm" SET "Decision" = \?, "updated_at" = \? WHERE "trx_kpm"."deleted_at" IS NULL AND \(\(id = \?\)\)`).
			WithArgs(testDecision, sqlmock.AnyArg(), testID).
			WillReturnError(expectedError)

		mock.ExpectRollback()

		err := repo.UpdateTrxKPMDecision(testID, testProspectID, testDecision)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error - TrxKPMStatus Creation Fails", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE "trx_kpm" SET "Decision" = \?, "updated_at" = \? WHERE "trx_kpm"."deleted_at" IS NULL AND \(\(id = \?\)\)`).
			WithArgs(testDecision, sqlmock.AnyArg(), testID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		expectedError := fmt.Errorf("failed to create TrxKPMStatus")
		mock.ExpectExec(`INSERT INTO "trx_kpm_status" \("id","ProspectID","Decision","created_at","created_by","updated_at","updated_by","deleted_at","deleted_by"\) VALUES \(\?,\?,\?,\?,\?,\?,\?,\?,\?\)`).
			WithArgs(
				sqlmock.AnyArg(),
				testProspectID,
				testDecision,
				sqlmock.AnyArg(),
				"",
				sqlmock.AnyArg(),
				"",
				nil,
				"",
			).
			WillReturnError(expectedError)

		mock.ExpectRollback()

		err := repo.UpdateTrxKPMDecision(testID, testProspectID, testDecision)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error - Empty ID", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE "trx_kpm" SET "Decision" = \?, "updated_at" = \? WHERE "trx_kpm"."deleted_at" IS NULL AND \(\(id = \?\)\)`).
			WithArgs(testDecision, sqlmock.AnyArg(), "").
			WillReturnError(fmt.Errorf("invalid ID"))

		mock.ExpectRollback()

		err := repo.UpdateTrxKPMDecision("", testProspectID, testDecision)
		assert.Error(t, err)
	})

	t.Run("Error - Empty ProspectID", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(`UPDATE "trx_kpm" SET "Decision" = \?, "updated_at" = \? WHERE "trx_kpm"."deleted_at" IS NULL AND \(\(id = \?\)\)`).
			WithArgs(testDecision, sqlmock.AnyArg(), testID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(`INSERT INTO "trx_kpm_status" \("id","ProspectID","Decision","created_at","created_by","updated_at","updated_by","deleted_at","deleted_by"\) VALUES \(\?,\?,\?,\?,\?,\?,\?,\?,\?\)`).
			WithArgs(
				sqlmock.AnyArg(),
				"",
				testDecision,
				sqlmock.AnyArg(),
				"",
				sqlmock.AnyArg(),
				"",
				nil,
				"",
			).
			WillReturnError(fmt.Errorf("invalid ProspectID"))

		mock.ExpectRollback()

		err := repo.UpdateTrxKPMDecision(testID, "", testDecision)
		assert.Error(t, err)
	})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestGetMappingRiskLevel(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating mock db: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open("sqlite3", sqlDB)
	if err != nil {
		t.Fatalf("error opening gorm db: %v", err)
	}
	gormDB.LogMode(true)

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB)

	testCases := []struct {
		name             string
		numberOfInquiry  int
		expectedData     entity.MappingRiskLevel
		shouldReturnData bool
		expectError      bool
	}{
		{
			name:            "Mapping Found - Low Risk",
			numberOfInquiry: 2,
			expectedData: entity.MappingRiskLevel{
				InquiryStart: 1,
				InquiryEnd:   5,
				RiskLevel:    "LOW",
				Decision:     "APPROVE",
			},
			shouldReturnData: true,
			expectError:      false,
		},
		{
			name:            "Mapping Found - High Risk",
			numberOfInquiry: 8,
			expectedData: entity.MappingRiskLevel{
				InquiryStart: 6,
				InquiryEnd:   10,
				RiskLevel:    "HIGH",
				Decision:     "REJECT",
			},
			shouldReturnData: true,
			expectError:      false,
		},
		{
			name:             "No Mapping Found",
			numberOfInquiry:  15,
			expectedData:     entity.MappingRiskLevel{},
			shouldReturnData: false,
			expectError:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"inquiry_start", "inquiry_end", "risk_level", "decision"})

			if tc.shouldReturnData {
				rows.AddRow(tc.expectedData.InquiryStart, tc.expectedData.InquiryEnd, tc.expectedData.RiskLevel, tc.expectedData.Decision)
			}

			mock.ExpectQuery(`SELECT TOP 1 inquiry_start, inquiry_end, risk_level, decision FROM dbo\.m_mapping_risk_level WHERE \d+ >= inquiry_start AND \d+ <= inquiry_end AND deleted_at IS NULL`).
				WillReturnRows(rows)

			result, err := repo.GetMappingRiskLevel(tc.numberOfInquiry)

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "record not found")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedData.InquiryStart, result.InquiryStart)
				assert.Equal(t, tc.expectedData.InquiryEnd, result.InquiryEnd)
				assert.Equal(t, tc.expectedData.RiskLevel, result.RiskLevel)
				assert.Equal(t, tc.expectedData.Decision, result.Decision)
			}
		})
	}
}
