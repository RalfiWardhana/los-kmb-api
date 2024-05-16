package repository

import (
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestGetMappingVehicleAge(t *testing.T) {

	sqlDB, mock, _ := sqlmock.New()
	defer sqlDB.Close()

	gormDB, _ := gorm.Open("sqlite3", sqlDB)
	gormDB.LogMode(true)

	gormDB = gormDB.Debug()

	repo := NewRepository(gormDB, gormDB, gormDB, gormDB, gormDB, gormDB, gormDB, gormDB)

	expected := entity.MappingVehicleAge{
		VehicleAgeStart: 11,
		VehicleAgeEnd:   12,
		Cluster:         "Cluster C",
		BPKBNameType:    1,
		TenorStart:      1,
		TenorEnd:        12,
		Decision:        constant.DECISION_PASS,
	}

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT TOP 1 * FROM m_mapping_vehicle_age WHERE vehicle_age_start <= ? AND vehicle_age_end >= ? AND cluster = ? AND bpkb_name_type = ? AND tenor_start <= ? AND tenor_end >= ?`)).
			WithArgs(11, 11, "Cluster C", 1, 12, 12).
			WillReturnRows(sqlmock.NewRows([]string{"vehicle_age_start", "vehicle_age_end", "cluster", "bpkb_name_type", "tenor_start", "tenor_end", "decision"}).
				AddRow(11, 12, "Cluster C", 1, 1, 12, constant.DECISION_PASS))

		data, err := repo.GetMappingVehicleAge(11, "Cluster C", 1, 12)

		assert.NoError(t, err)
		assert.Equal(t, expected, data, "Expected mapping vehicle age to match")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})

	t.Run("record not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT TOP 1 * FROM m_mapping_vehicle_age WHERE vehicle_age_start <= ? AND vehicle_age_end >= ? AND cluster = ? AND bpkb_name_type = ? AND tenor_start <= ? AND tenor_end >= ?`)).
			WithArgs(11, 11, "Cluster C", 1, 12, 12).
			WillReturnError(gorm.ErrRecordNotFound)

		data, err := repo.GetMappingVehicleAge(11, "Cluster C", 1, 12)

		assert.NoError(t, err)
		assert.Equal(t, entity.MappingVehicleAge{}, data, "Expected empty mapping vehicle age")

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("There were unfulfilled expectations: %s", err)
		}
	})
}
