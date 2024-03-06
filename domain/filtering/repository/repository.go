package repository

import (
	"context"
	"database/sql"
	"los-kmb-api/domain/filtering/interfaces"
	"los-kmb-api/models/entity"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	minilosKmb *gorm.DB
	KpLos      *gorm.DB
	dummy      *gorm.DB
}

func NewRepository(minilosKmb, KpLos, dummy *gorm.DB) interfaces.Repository {
	return &repoHandler{
		minilosKmb: minilosKmb,
		KpLos:      KpLos,
		dummy:      dummy,
	}
}

func (r repoHandler) DummyDataPbk(noktp string) (data entity.DummyPBK, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.dummy.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.dummy.Raw("SELECT * FROM dbo.dummy_pefindo_kmb WHERE IDNumber = ?", noktp).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) DataGetMappingDp(branchID, statusKonsumen string) (data []entity.RangeBranchDp, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw("SELECT mbd.* FROM dbo.mapping_branch_dp mdp LEFT JOIN dbo.mapping_baki_debet mbd ON mdp.baki_debet = mbd.id LEFT JOIN dbo.master_list_dp mld ON mdp.master_list_dp = mld.id WHERE mdp.branch = ? AND mdp.customer_status = ?", branchID, statusKonsumen).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) BranchDpData(query string) (data entity.BranchDp, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

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
	if err = r.minilosKmb.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateData(data entity.ApiDupcheckKmbUpdate) (err error) {
	data.DtmResponse = time.Now()
	data.Timestamp = time.Now()
	if err = r.minilosKmb.Table("api_dupcheck_kmb").Where("RequestID = ?", data.RequestID).UpdateColumns(data).Error; err != nil {
		return
	}

	return

}
