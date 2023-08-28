package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/elaborate_ltv/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	KpLosLogs *gorm.DB
	NewKmb    *gorm.DB
}

func NewRepository(KpLosLogs, NewKmb *gorm.DB) interfaces.Repository {
	return &repoHandler{
		KpLosLogs: KpLosLogs,
		NewKmb:    NewKmb,
	}
}

func (r repoHandler) SaveTrxElaborateLTV(data entity.TrxElaborateLTV) (err error) {
	data.CreatedAt = time.Now()
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	result := db.Model(&data).Where("prospect_id = ?", data.ProspectID).Updates(entity.TrxElaborateLTV{
		RequestID:             data.RequestID,
		Tenor:                 data.Tenor,
		ManufacturingYear:     data.ManufacturingYear,
		MappingElaborateLTVID: data.MappingElaborateLTVID,
		CreatedAt:             data.CreatedAt,
	})

	if err = result.Error; err != nil {
		return
	}

	if result.RowsAffected == 0 {
		// record not found...
		if err = db.Create(data).Error; err != nil {
			return
		}
	}

	return
}

func (r repoHandler) GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT bpkb_name, customer_status, decision, customer_segment, total_baki_debet_non_collateral_biro, score_biro, cluster FROM trx_filtering WITH (nolock) WHERE prospect_id = ?", prospectID).Scan(&filtering).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetMappingElaborateLTV(resultPefindo, cluster string) (data []entity.MappingElaborateLTV, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT * FROM m_mapping_elaborate_ltv WITH (nolock) WHERE result_pefindo = ? AND cluster = ? ", resultPefindo, cluster).Scan(&data).Error; err != nil {
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
