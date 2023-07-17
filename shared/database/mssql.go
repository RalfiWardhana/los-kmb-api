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

func OpenMinilosKMB() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetMinilosKmbDB()

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s"+"?charset=utf8mb4&parseTime=True&loc=Local",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("MINILOS_KMB_DB_MAX_IDLE_CONNECTION"))
	maxOpen, _ := strconv.Atoi(os.Getenv("MINILOS_KMB_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxOpen)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.LogMode(config.IsDevelopment)

	return db, nil
}

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

func OpenKpLosLog() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetKpLosLogDB()

	connString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mssql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("KP_LOS_LOG_DB_MAX_IDLE_CONNECTION"))
	maxOpen, _ := strconv.Atoi(os.Getenv("KP_LOS_LOG_DB_MAX_OPEN_CONNECTION"))

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
