package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
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
	KpLos     *gorm.DB
	KpLosLogs *gorm.DB
	NewKmb    *gorm.DB
}

func NewRepository(kpLos, KpLosLogs, NewKmb *gorm.DB) interfaces.Repository {
	return &repoHandler{
		KpLos:     kpLos,
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

	if err = r.NewKmb.Raw("SELECT branch_id, bpkb_name, customer_status, decision, next_process, max_overdue_biro, max_overdue_last12months_biro, customer_segment, total_baki_debet_non_collateral_biro, score_biro, cluster, cmo_cluster, FORMAT(rrd_date, 'yyyy-MM-ddTHH:mm:ss') + 'Z' AS rrd_date, created_at FROM trx_filtering WITH (nolock) WHERE prospect_id = ?", prospectID).Scan(&filtering).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetFilteringDetail(prospectID string) (filtering []entity.TrxDetailBiro, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT prospect_id, score FROM trx_detail_biro WITH (nolock) WHERE prospect_id = ?", prospectID).Scan(&filtering).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetMappingPBKScoreGrade() (mappingPBKScoreGrade []entity.MappingPBKScoreGrade, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT score, grade_risk, grade_score FROM m_mapping_pbk_score_grade WITH (nolock) WHERE deleted_at IS NULL").Scan(&mappingPBKScoreGrade).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetMappingBranchPBK(branchID string, gradePBK string) (mappingBranchByPBKScore entity.MappingBranchByPBKScore, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT branch_id, score, grade_branch FROM m_mapping_branch mmb WHERE deleted_at IS NULL AND  branch_id = ? AND score = ?", branchID, gradePBK).Scan(&mappingBranchByPBKScore).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetMappingElaborateLTV(resultPefindo, cluster string, bpkb_name_type int, customerStatus, gradePBK, gradeBranch string) (data []entity.MappingElaborateLTV, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(fmt.Sprintf("SELECT * FROM m_mapping_elaborate_ltv WITH (nolock) WHERE deleted_at IS NULL AND result_pefindo = '%s' AND cluster = '%s' AND bpkb_name_type = %d AND status_konsumen IN ('ALL','%s') AND pbk_score IN ('ALL','%s') AND grade_branch IN ('ALL','%s')", resultPefindo, cluster, bpkb_name_type, customerStatus, gradePBK, gradeBranch)).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetMappingElaborateLTVOvd(resultPefindo, cluster string) (data []entity.MappingElaborateLTV, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT * FROM m_mapping_elaborate_ltv_ovd WITH (nolock) WHERE result_pefindo = ? AND cluster = ? ", resultPefindo, cluster).Scan(&data).Error; err != nil {
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

func (r *repoHandler) GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error) {
	if err := r.KpLos.Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s' AND lob = '%s' AND [key]= '%s' AND is_active = 1", groupName, lob, key)).Scan(&appConfig).Error; err != nil {
		return appConfig, err
	}

	return appConfig, err
}
