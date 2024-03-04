package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"los-kmb-api/domain/elaborate/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
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

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

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

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.minilosKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.minilosKmb.Raw(`SELECT
			PefindoID,
			PefindoIDSpouse,
			CASE
			 WHEN PefindoScore IS NULL then 'UNSCORE'
			 ELSE PefindoScore
			END AS PefindoScore,
			CAST(
			JSON_EXTRACT(ResultPefindo, '$.result.max_overdue') AS SIGNED
			) AS MaxOverdue,
			JSON_EXTRACT(ResultPefindo, '$.result.max_overdue') = CAST('null' AS JSON) AS IsNullMaxOverdue,
			CAST(
			JSON_EXTRACT(
				ResultPefindo,
				'$.result.max_overdue_last12months'
			) AS SIGNED
			) AS MaxOverdueLast12Months,
			JSON_EXTRACT(
			ResultPefindo,
			'$.result.max_overdue_last12months'
			) = CAST('null' AS JSON) AS IsNullMaxOverdueLast12Months
		FROM
			api_dupcheck_kmb
		WHERE
			ProspectID = ?
		ORDER BY
			Timestamp DESC
		LIMIT
			1`, prospect_id).Scan(&filtering).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetResultElaborate(branchId string, customerStatus string, bpkb int, resultPefindo string, tenor int, ageVehicle string, ltv float64, bakiDebet float64) (data entity.ResultElaborate, err error) {

	var queryAdd string
	var rangeLtv int = int(ltv)
	var total_baki_debet int = int(bakiDebet)

	// PEFINDO PASS
	if resultPefindo == constant.DECISION_PASS {
		if tenor >= 36 {
			queryAdd = fmt.Sprintf("AND mes.bpkb_name_type = %d AND mes.tenor_start >= 36 AND mes.tenor_end = 0", bpkb)
		} else {
			queryAdd = fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if tenor >= 36 && bpkb == 1 {
			queryAdd += fmt.Sprintf(" AND mes.age_vehicle = '%s'", ageVehicle)
		}

		if (ageVehicle == "<=12" && bpkb == 1) || (ageVehicle == "<=12" && bpkb == 0 && tenor < 36) {

			if rangeLtv != 0 && rangeLtv <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", rangeLtv, rangeLtv)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", rangeLtv)
			}
		} else if tenor < 36 {
			if rangeLtv != 0 && rangeLtv <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", rangeLtv, rangeLtv)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", rangeLtv)
			}
		}
	}

	// PEFINDO NO HIT
	if resultPefindo == constant.DECISION_PBK_NO_HIT {
		if tenor >= 24 {
			queryAdd = "AND mes.tenor_start >= 24 AND mes.tenor_end = 0"
		} else {
			queryAdd = fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if rangeLtv != 0 && rangeLtv <= 1000 {
			queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", rangeLtv, rangeLtv)
		} else {
			queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", rangeLtv)
		}
	}

	// PEFINDO REJECT
	if resultPefindo == constant.DECISION_REJECT {

		queryAdd = fmt.Sprintf("AND mes.total_baki_debet_start <= %d AND mes.total_baki_debet_end >= %d", total_baki_debet, total_baki_debet)

		if tenor >= 24 {
			queryAdd += " AND mes.tenor_start >= '24' AND mes.tenor_end = 0"
		} else {
			queryAdd += fmt.Sprintf(" AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if tenor < 24 {
			if rangeLtv != 0 && rangeLtv <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", rangeLtv, rangeLtv)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= '1000'", rangeLtv)
			}
		}
	}

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.KpLos.Raw("SELECT mcb.cluster, mes.decision, mes.ltv_start FROM kmb_mapping_cluster_branch mcb WITH (nolock) JOIN kmb_mapping_elaborate_scheme mes ON mcb.cluster = mes.cluster WHERE mcb.branch_id = ? AND mcb.customer_status = ? AND mcb.bpkb_name_type = ? AND mes.result_pefindo = ? "+queryAdd, branchId, customerStatus, bpkb, resultPefindo).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetMappingLtvOvd(cluster, resultPefindo string, tenor int, ltv float64) (data entity.ResultElaborate, err error) {

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ltvRange := int(ltv)

	// Common query building functions
	buildTenorQuery := func() string {
		return fmt.Sprintf("AND tenor_start <= %d AND tenor_end >= %d", tenor, tenor)
	}

	buildLTVQuery := func() string {
		if ltvRange != 0 && ltvRange <= 1000 {
			return fmt.Sprintf("AND ltv_start <= %d AND ltv_end >= %d", ltvRange, ltvRange)
		}
		return fmt.Sprintf("AND ltv_start <= %d AND ltv_end >= 1000", ltvRange)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.KpLos.BeginTx(ctx, &sql.TxOptions{})
	defer db.Commit()

	if err = r.KpLos.Raw(fmt.Sprintf("SELECT cluster, decision, ltv_start FROM kmb_mapping_elaborate_scheme_ovd WITH (nolock) WHERE result_pefindo = '%s' AND cluster = '%s' %s %s", resultPefindo, cluster, buildTenorQuery(), buildLTVQuery())).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}
