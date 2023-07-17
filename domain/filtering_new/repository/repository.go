package repository

import (
	"context"
	"database/sql"
	"fmt"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/config"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	KpLos *gorm.DB
	dummy *gorm.DB
}

func NewRepository(KpLos, dummy *gorm.DB) interfaces.Repository {
	return &repoHandler{
		KpLos: KpLos,
		dummy: dummy,
	}
}

func (r repoHandler) DummyDataPbk(noktp string) (data entity.DummyPBK, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.dummy.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.dummy.Raw("SELECT * FROM new_pefindo_kmb WHERE IDNumber = ?", noktp).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) DataGetMappingDp(branchID, statusKonsumen string) (data []entity.RangeBranchDp, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw("SELECT mbd.* FROM dbo.mapping_branch_dp mdp LEFT JOIN dbo.mapping_baki_debet mbd ON mdp.baki_debet = mbd.id LEFT JOIN dbo.master_list_dp mld ON mdp.master_list_dp = mld.id WHERE mdp.branch = ? AND mdp.customer_status = ?").Scan(&data).Error; err != nil {
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

func (r repoHandler) SaveDupcheckResult(data entity.FilteringKMB) (err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Create(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetFilteringByID(prospectID string) (row int, err error) {

	var data []entity.FilteringKMB

	if err = r.KpLos.Raw(fmt.Sprintf("SELECT ProspectID FROM filtering_kmob WITH (nolock) WHERE ProspectID = '%s'", prospectID)).Scan(&data).Error; err != nil {
		return
	}

	row = len(data)

	return
}
