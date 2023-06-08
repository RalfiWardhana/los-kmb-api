package repository

import (
	"context"
	"database/sql"
	"los-kmb-api/domain/filtering/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/config"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	kmbFiltering *gorm.DB
	KpLos        *gorm.DB
	dummy        *gorm.DB
}

func NewRepository(kmbFiltering, KpLos, dummy *gorm.DB) interfaces.Repository {
	return &repoHandler{
		dummy:        dummy,
		kmbFiltering: kmbFiltering,
		KpLos:        KpLos,
	}
}

func (r repoHandler) DummyData(query string) (data entity.DummyColumn, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.dummy.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.dummy.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) DummyDataPbk(query string) (data entity.DummyPBK, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.dummy.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.dummy.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) DataProfessionGroup(query string) (data entity.ProfessionGroup, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.kmbFiltering.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.kmbFiltering.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}
func (r repoHandler) DataGetMappingDp(query string) (data []entity.RangeBranchDp, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) BranchDpData(query string) (data entity.BranchDp, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveData(data entity.ApiDupcheckKmb) (err error) {
	data.DtmRequest = DtmRequest
	data.Timestamp = time.Now()
	if err = r.kmbFiltering.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateData(data entity.ApiDupcheckKmbUpdate) (err error) {
	data.DtmResponse = time.Now()
	data.Timestamp = time.Now()
	if err = r.kmbFiltering.Table("api_dupcheck_kmb").Where("RequestID = ?", data.RequestID).UpdateColumns(data).Error; err != nil {
		return
	}

	return

}
