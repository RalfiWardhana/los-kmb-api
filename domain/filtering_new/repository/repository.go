package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	KpLos     *gorm.DB
	KpLosLogs *gorm.DB
	NewKmb    *gorm.DB
}

func NewRepository(kpLos, kpLosLogs, newKmb *gorm.DB) interfaces.Repository {
	return &repoHandler{
		KpLos:     kpLos,
		KpLosLogs: kpLosLogs,
		NewKmb:    newKmb,
	}
}

func (r repoHandler) DummyDataPbk(noktp string) (data entity.DummyPBK, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLosLogs.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT * FROM dbo.dummy_pefindo_kmb WITH (nolock) WHERE IDNumber = ?", noktp).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveFiltering(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro) (err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Create(&data).Error; err != nil {
		return
	}

	if len(trxDetailBiro) > 0 {
		for _, v := range trxDetailBiro {
			if err = db.Create(&v).Error; err != nil {
				return
			}
		}
	}

	return
}

func (r repoHandler) GetFilteringByID(prospectID string) (row int, err error) {

	var data []entity.FilteringKMB

	if err = r.NewKmb.Raw(fmt.Sprintf("SELECT prospect_id FROM dbo.trx_filtering WITH (nolock) WHERE prospect_id = '%s'", prospectID)).Scan(&data).Error; err != nil {
		return
	}

	row = len(data)

	return
}

func (r repoHandler) MasterMappingCluster(req entity.MasterMappingCluster) (data entity.MasterMappingCluster, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT * FROM dbo.kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?", req.BranchID, req.CustomerStatus, req.BpkbNameType).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error) {

	headerByte, _ := json.Marshal(header)
	requestByte, _ := json.Marshal(request)
	responseByte, _ := json.Marshal(response)

	if err = r.KpLosLogs.Model(&entity.LogOrchestrator{}).Create(&entity.LogOrchestrator{
		ID:           requestID,
		ProspectID:   prospectID,
		Owner:        "LOS-KMB",
		Header:       string(headerByte),
		Url:          path,
		Method:       method,
		RequestData:  string(requestByte),
		ResponseData: string(utils.SafeEncoding(responseByte)),
	}).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetResultFiltering(prospectID string) (data []entity.ResultFiltering, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw(`SELECT tf.prospect_id, tf.decision, tf.reason, tf.customer_status, tf.customer_segment, tf.is_blacklist, tf.next_process,
	tf.total_baki_debet_non_collateral_biro, tdb.url_pdf_report, tdb.subject FROM trx_filtering tf 
	LEFT JOIN trx_detail_biro tdb ON tf.prospect_id = tdb.prospect_id 
	WHERE tf.prospect_id = ?`, prospectID).Scan(&data).Error; err != nil {
		return
	}

	return
}
