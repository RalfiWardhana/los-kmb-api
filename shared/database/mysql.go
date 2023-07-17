package database

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"los-kmb-api/shared/config"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func OpenMinilosWG() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetMinilosWGDB()

	connString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true&loc=Local",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("MINILOS_WG_DB_MAX_IDLE_CONNECTION"))
	// maxOpen, _ := strconv.Atoi(os.Getenv("MINILOS_WG_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(100)
	db.DB().SetConnMaxLifetime(time.Hour)
	db.LogMode(config.IsDevelopment)

	return db, nil
}

func OpenMinilosKMB() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetMinilosKMBDB()

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

func OpenCatalogData() (*gorm.DB, error) {

	user, pwd, host, port, database := config.GetLosCatalogDataDB()

	connString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s"+"?charset=utf8mb4&parseTime=True&loc=Local",
		user, pwd, host, port, database,
	)

	db, err := gorm.Open("mysql", connString)
	if err != nil {
		return nil, err
	}

	maxIdle, _ := strconv.Atoi(os.Getenv("LOS_CATALOG_DATA_DB_MAX_IDLE_CONNECTION"))
	maxOpen, _ := strconv.Atoi(os.Getenv("LOS_CATALOG_DATA_DB_MAX_OPEN_CONNECTION"))

	db.DB().SetMaxIdleConns(maxIdle)
	db.DB().SetMaxOpenConns(maxOpen)
	db.LogMode(config.IsDevelopment)

	return db, nil
}
