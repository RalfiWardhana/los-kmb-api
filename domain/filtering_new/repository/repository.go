package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
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

func (r repoHandler) SaveFiltering(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, dataCMOnoFPD entity.TrxCmoNoFPD) (err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Create(&data).Error; err != nil {
		return
	}

	if dataCMOnoFPD.CMOID != "" {
		if err = db.Create(&dataCMOnoFPD).Error; err != nil {
			return
		}
	}

	if len(trxDetailBiro) > 0 {
		for _, v := range trxDetailBiro {
			if err = db.Create(&v).Error; err != nil {
				return
			}
		}
	}

	// insert worker ne
	if data.ProspectID[0:2] == "NE" {
		var trxNewEntry entity.NewEntry
		if err = db.Raw("SELECT * FROM trx_new_entry WITH (nolock) WHERE ProspectID = ?", data.ProspectID).Scan(&trxNewEntry).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				err = errors.New(constant.RECORD_NOT_FOUND)
			}
			return
		}

		newKpLos := r.KpLos.Transaction(func(tx *gorm.DB) error {
			// header los kmb api
			callbackHeaderLos, _ := json.Marshal(
				map[string]string{
					"X-Client-ID":   os.Getenv("CLIENT_LOS"),
					"Authorization": os.Getenv("AUTH_LOS"),
				})

			// elaborate
			if err := tx.Create(&entity.TrxWorker{
				ProspectID:      data.ProspectID,
				Category:        "NE_KMB",
				Action:          "NE_ELABORATE",
				APIType:         "RAW",
				EndPointTarget:  os.Getenv("NE_ELABORATE_URL"),
				EndPointMethod:  constant.METHOD_POST,
				Header:          string(callbackHeaderLos),
				Payload:         trxNewEntry.PayloadLTV,
				ResponseTimeout: 30,
				MaxRetry:        6,
				CountRetry:      0,
				Activity:        constant.ACTIVITY_UNPROCESS,
				Sequence:        1,
			}).Error; err != nil {
				return err
			}

			// submit to los
			if err := tx.Create(&entity.TrxWorker{
				ProspectID:      data.ProspectID,
				Category:        "NE_KMB",
				Action:          "NE_JOURNEY",
				APIType:         "RAW",
				EndPointTarget:  os.Getenv("NE_JOURNEY_URL"),
				EndPointMethod:  constant.METHOD_POST,
				Payload:         trxNewEntry.PayloadJourney,
				ResponseTimeout: 30,
				MaxRetry:        6,
				CountRetry:      0,
				Activity:        constant.ACTIVITY_IDLE,
				Sequence:        2,
			}).Error; err != nil {
				return err
			}

			return nil
		})

		if newKpLos != nil {
			db.Rollback()
			err = newKpLos
			return
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

	if err = db.Raw(`SELECT tf.prospect_id, tf.decision, tf.reason, tf.customer_status, tf.customer_status_kmb, tf.customer_segment, tf.is_blacklist, tf.next_process,
	tf.total_baki_debet_non_collateral_biro, tdb.url_pdf_report, tdb.subject FROM trx_filtering tf 
	LEFT JOIN trx_detail_biro tdb ON tf.prospect_id = tdb.prospect_id 
	WHERE tf.prospect_id = ?`, prospectID).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) MasterMappingFpdCluster(FpdValue float64) (data entity.MasterMappingFpdCluster, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw(`SELECT cluster FROM m_mapping_fpd_cluster WITH (nolock) 
							WHERE (fpd_start_hte <= ? OR fpd_start_hte IS NULL) 
							AND (fpd_end_lt > ? OR fpd_end_lt IS NULL)`, FpdValue, FpdValue).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) CheckCMONoFPD(cmoID string, bpkbName string) (data entity.TrxCmoNoFPD, err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw(`SELECT TOP 1 prospect_id, cmo_id, cmo_category, 
							FORMAT(CONVERT(datetime, cmo_join_date, 127), 'yyyy-MM-dd') AS cmo_join_date, 
							default_cluster, 
							FORMAT(CONVERT(datetime, default_cluster_start_date, 127), 'yyyy-MM-dd') AS default_cluster_start_date, 
							FORMAT(CONVERT(datetime, default_cluster_end_date, 127), 'yyyy-MM-dd') AS default_cluster_end_date
						  FROM dbo.trx_cmo_no_fpd WITH (nolock) 
						  WHERE cmo_id = ? AND bpkb_name = ?
						  ORDER BY created_at DESC`, cmoID, bpkbName).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}
	return
}

func (r *repoHandler) GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error) {
	if err := r.KpLos.Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s' AND lob = '%s' AND [key]= '%s' AND is_active = 1", groupName, lob, key)).Scan(&appConfig).Error; err != nil {
		return appConfig, err
	}

	return appConfig, err
}
