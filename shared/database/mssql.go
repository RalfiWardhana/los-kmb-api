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

func OpenDatabase() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetFilteringDB()

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s"+"?charset=utf8mb4&parseTime=True&loc=Local",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("KMB_FILTERING_DB_MAX_IDLE_CONNECTION"))
	maxOpen, _ := strconv.Atoi(os.Getenv("KMB_FILTERING_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxOpen)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.LogMode(config.IsDevelopment)

	return db, nil
}

func OpenDatabaseKpLos() (*gorm.DB, error) {

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

func OpenDummyDatabase() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetDummyDB()

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s"+"?charset=utf8mb4&parseTime=True&loc=Local",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("DUMMY_DB_MAX_IDLE_CONNECTION"))
	maxOpen, _ := strconv.Atoi(os.Getenv("DUMMY_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxOpen)
	db.LogMode(config.IsDevelopment)

	return db, nil
}
