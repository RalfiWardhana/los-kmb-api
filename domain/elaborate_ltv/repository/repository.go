package repository

import (
	"context"
	"database/sql"
	"fmt"
	"los-kmb-api/domain/elaborate_ltv/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
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
}

func NewRepository(minilosKmb, KpLos *gorm.DB) interfaces.Repository {
	return &repoHandler{
		minilosKmb: minilosKmb,
		KpLos:      KpLos,
	}
}

func (r repoHandler) SaveDataElaborate(data entity.ApiElaborateKmb) (err error) {
	data.DtmRequest = DtmRequest
	data.Timestamp = time.Now()
	if err = r.minilosKmb.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateDataElaborate(data entity.ApiElaborateKmbUpdate) (err error) {
	data.DtmResponse = time.Now()
	data.Timestamp = time.Now()
	if err = r.minilosKmb.Table("api_elaborate_scheme").Where("RequestID = ?", data.RequestID).UpdateColumns(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetClusterBranchElaborate(branch_id string, cust_status string, bpkb int) (cluster entity.ClusterBranch, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw("SELECT cluster FROM kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?", branch_id, cust_status, bpkb).Scan(&cluster).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) GetFilteringResult(prospect_id string) (filtering entity.ApiDupcheckKmbUpdate, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.minilosKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.minilosKmb.Raw("SELECT PefindoID, PefindoIDSpouse, PefindoScore FROM api_dupcheck_kmb WHERE ProspectID = ?", prospect_id).Scan(&filtering).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetResultElaborate(branch_id string, cust_status string, bpkb int, result_pefindo string, tenor int, age_vehicle string, ltv float64, baki_debet float64) (data entity.ResultElaborate, err error) {

	var queryAdd string
	var ltv_range int = int(ltv)
	var total_baki_debet int = int(baki_debet)

	// PEFINDO PASS
	if result_pefindo == constant.DECISION_PASS {
		if tenor >= 36 {
			queryAdd = fmt.Sprintf("AND mes.bpkb_name_type = %d AND mes.tenor_start >= 36 AND mes.tenor_end = 0", bpkb)
		} else {
			queryAdd = fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if tenor >= 36 && bpkb == 1 {
			queryAdd += fmt.Sprintf(" AND mes.age_vehicle = '%s'", age_vehicle)
		}

		if (age_vehicle == "<=12" && bpkb == 1) || (age_vehicle == "<=12" && bpkb == 0 && tenor < 36) {

			if ltv_range != 0 && ltv_range <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", ltv_range, ltv_range)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", ltv_range)
			}
		} else if tenor < 36 {
			if ltv_range != 0 && ltv_range <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", ltv_range, ltv_range)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", ltv_range)
			}
		}
	}

	// PEFINDO NO HIT
	if result_pefindo == constant.DECISION_PBK_NO_HIT {
		if tenor >= 24 {
			queryAdd = "AND mes.tenor_start >= 24 AND mes.tenor_end = 0"
		} else {
			queryAdd = fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if ltv_range != 0 && ltv_range <= 1000 {
			queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", ltv_range, ltv_range)
		} else {
			queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", ltv_range)
		}
	}

	// PEFINDO REJECT
	if result_pefindo == constant.DECISION_REJECT {

		queryAdd = fmt.Sprintf("AND mes.total_baki_debet_start <= %d AND mes.total_baki_debet_end >= %d", total_baki_debet, total_baki_debet)

		if tenor >= 24 {
			queryAdd += " AND mes.tenor_start >= '24' AND mes.tenor_end = 0"
		} else {
			queryAdd += fmt.Sprintf(" AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if tenor < 24 {
			if ltv_range != 0 && ltv_range <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", ltv_range, ltv_range)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= '1000'", ltv_range)
			}
		}
	}

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw("SELECT mcb.cluster, mes.decision, mes.ltv_start FROM kmb_mapping_cluster_branch mcb JOIN kmb_mapping_elaborate_scheme mes ON mcb.cluster = mes.cluster WHERE mcb.branch_id = ? AND mcb.customer_status = ? AND mcb.bpkb_name_type = ? AND mes.result_pefindo = ? "+queryAdd, branch_id, cust_status, bpkb, result_pefindo).Scan(&data).Error; err != nil {
		return
	}

	return
}
