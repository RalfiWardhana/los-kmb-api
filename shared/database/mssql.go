package database

import (
	"fmt"
	"los-kmb-api/shared/config"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func OpenKpLos() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetKpLosDB()

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mssql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("KP_LOS_DB_MAX_IDLE_CONNECTION"))
	maxOpen, _ := strconv.Atoi(os.Getenv("KP_LOS_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxOpen)
	db.LogMode(config.IsDevelopment)

	return db, nil
}

func OpenKpLosLogs() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetKpLosLogsDB()

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mssql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("KP_LOS_LOG_DB_MAX_IDLE_CONNECTION"))
	// maxOpen, _ := strconv.Atoi(os.Getenv("KP_LOS_LOG_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.LogMode(config.IsDevelopment)

	return db, nil
}

func OpenConfins() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetConfinsDB()

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mssql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("CONFINS_DB_MAX_IDLE_CONNECTION"))
	// maxOpen, _ := strconv.Atoi(os.Getenv("CONFINS_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.LogMode(config.IsDevelopment)

	return db, nil
}

func OpenStaging() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetStagingDB()

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mssql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("STAGING_DB_MAX_IDLE_CONNECTION"))
	// maxOpen, _ := strconv.Atoi(os.Getenv("STAGING_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.LogMode(config.IsDevelopment)

	return db, nil
}
