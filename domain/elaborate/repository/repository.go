package repository

import (
	"context"
	"database/sql"
	"los-kmb-api/domain/elaborate/interfaces"
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
	kmbElaborate *gorm.DB
	KpLos        *gorm.DB
}

func NewRepository(kmbElaborate, KpLos *gorm.DB) interfaces.Repository {
	return &repoHandler{
		kmbElaborate: kmbElaborate,
		KpLos:        KpLos,
	}
}

func (r repoHandler) SaveDataElaborate(data entity.ApiElaborateKmb) (err error) {
	data.DtmRequest = DtmRequest
	data.Timestamp = time.Now()
	if err = r.kmbElaborate.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateDataElaborate(data entity.ApiElaborateKmbUpdate) (err error) {
	data.DtmResponse = time.Now()
	data.Timestamp = time.Now()
	if err = r.kmbElaborate.Table("api_elaborate_scheme").Where("RequestID = ?", data.RequestID).UpdateColumns(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetClusterBranchElaborate(query string) (cluster entity.ClusterBranch, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw(query).Scan(&cluster).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetResultElaborate(query string) (data entity.ResultElaborate, err error) {
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
