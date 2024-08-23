package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"los-kmb-api/domain/principle/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

type repoHandler struct {
	newKmb   *gorm.DB
	los      *gorm.DB
	scorePro *gorm.DB
	confins  *gorm.DB
}

func NewRepository(newKmb, los, scorePro, confins *gorm.DB) interfaces.Repository {
	return &repoHandler{
		newKmb:   newKmb,
		los:      los,
		scorePro: scorePro,
		confins:  confins,
	}
}

func (r *repoHandler) GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error) {
	if err := r.los.Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s' AND lob = '%s' AND [key]= '%s' AND is_active = 1", groupName, lob, key)).Scan(&appConfig).Error; err != nil {
		return appConfig, err
	}

	return appConfig, err
}

func (r repoHandler) GetMinimalIncomePMK(branchID string, statusKonsumen string) (responseIncomePMK entity.MappingIncomePMK, err error) {
	if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)).Scan(&responseIncomePMK).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, constant.DEFAULT_BRANCH_ID, statusKonsumen)).Scan(&responseIncomePMK).Error; err != nil {
				return
			}
		}
	}
	return
}

func (r repoHandler) GetDraftPrinciple(prospectID string) (data entity.DraftPrinciple, err error) {

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT * FROM trx_draft_principle WITH (nolock) WHERE ProspectID = '%s'`, prospectID)).Scan(&data).Error; err != nil {
		err = nil
		return
	}

	return
}

func (r repoHandler) MasterMappingFpdCluster(FpdValue float64) (data entity.MasterMappingFpdCluster, err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmb.BeginTx(ctx, &x)
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

func (r repoHandler) MasterMappingCluster(req entity.MasterMappingCluster) (data entity.MasterMappingCluster, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.los.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT * FROM dbo.kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?", req.BranchID, req.CustomerStatus, req.BpkbNameType).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) SaveFiltering(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, dataCMOnoFPD entity.TrxCmoNoFPD) (err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Create(&data).Error; err != nil {
		err = nil // dummy
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

	return
}

func (r repoHandler) GetBannedPMKDSR(idNumber string) (data entity.TrxBannedPMKDSR, err error) {

	date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT * FROM trx_banned_pmk_dsr WITH (nolock) WHERE IDNumber = '%s' AND CAST(created_at as DATE) >= '%s'`, idNumber, date)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetEncB64(myString string) (encryptedString entity.EncryptedString, err error) {

	if err = r.los.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS my_string`, myString)).Scan(&encryptedString).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetRejection(idNumber string) (data entity.TrxReject, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT 
	COUNT(CASE WHEN ts.source_decision = 'PMK' OR ts.source_decision = 'DSR' OR ts.source_decision = 'PRJ' THEN 1 END) as reject_pmk_dsr,
	COUNT(CASE WHEN ts.source_decision != 'PMK' AND ts.source_decision != 'DSR' AND ts.source_decision != 'PRJ' AND ts.source_decision != 'NKA' THEN 1 END) as reject_nik 
	FROM trx_status ts WITH (nolock) LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID
	WHERE ts.decision = 'REJ' AND tcp.IDNumber = '%s' AND CAST(ts.created_at as DATE) = '%s'`, idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetMappingDukcapilVD(statusVD, customerStatus, customerSegment string, isValid bool) (resultDukcapil entity.MappingResultDukcapilVD, err error) {
	if customerStatus == constant.STATUS_KONSUMEN_NEW {
		if isValid || statusVD == constant.EKYC_RTO || statusVD == constant.EKYC_NOT_CHECK {
			if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s'`, statusVD, customerStatus)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		} else {
			if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND is_valid=0`, customerStatus)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		}
	} else {
		if isValid || statusVD == constant.EKYC_RTO || statusVD == constant.EKYC_NOT_CHECK {
			if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s' AND kategori_status_konsumen='%s'`, statusVD, customerStatus, customerSegment)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		} else {
			if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND kategori_status_konsumen='%s' AND is_valid=0`, customerStatus, customerSegment)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		}
	}

	return
}

func (r repoHandler) GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment string) (resultDukcapil entity.MappingResultDukcapil, err error) {
	if customerStatus == constant.STATUS_KONSUMEN_NEW {
		if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s'`, statusVD, statusFR, customerStatus)).Scan(&resultDukcapil).Error; err != nil {
			return
		}
	} else {
		if err = r.los.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s' AND kategori_status_konsumen='%s'`, statusVD, statusFR, customerStatus, customerSegment)).Scan(&resultDukcapil).Error; err != nil {
			return
		}
	}

	return
}

func (r repoHandler) CheckCMONoFPD(cmoID string, bpkbName string) (data entity.TrxCmoNoFPD, err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmb.BeginTx(ctx, &x)
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

func (r repoHandler) GetTrxPrincipleStatus(nik string) (data entity.TrxPrincipleStatus, err error) {

	if err = r.newKmb.Raw(fmt.Sprintf("SELECT TOP 1 * FROM trx_principle_status WITH (nolock) WHERE IDNumber = '%s' ORDER BY created_at DESC", nik)).Scan(&data).Error; err != nil {
		return
	}

	fmt.Println(data)

	return
}

func (r repoHandler) SavePrincipleStepOne(data entity.TrxPrincipleStepOne) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {
		// worker insert staging

		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.TrxPrincipleStatus{
			ProspectID: data.ProspectID,
			IDNumber:   data.IDNumber,
			Step:       1,
			Decision:   data.Decision,
			UpdatedAt:  time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}

func (r repoHandler) GetPrincipleStepOne(prospectID string) (data entity.TrxPrincipleStepOne, err error) {

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT * FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = '%s'`, prospectID)).Scan(&data).Error; err != nil {
		err = nil
		return
	}

	return
}

func (r repoHandler) SavePrincipleStepTwo(data entity.TrxPrincipleStepTwo) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.TrxPrincipleStatus{
			ProspectID: data.ProspectID,
			IDNumber:   data.IDNumber,
			Step:       2,
			Decision:   data.Decision,
			UpdatedAt:  time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}

func (r repoHandler) GetPrincipleStepTwo(prospectID string) (data entity.TrxPrincipleStepTwo, err error) {

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT * FROM trx_principle_step_two WITH (nolock) WHERE ProspectID = '%s'`, prospectID)).Scan(&data).Error; err != nil {
		err = nil
		return
	}

	return
}

func (r repoHandler) GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmb.Raw("SELECT bpkb_name, customer_status, decision, next_process, max_overdue_biro, max_overdue_last12months_biro, customer_segment, total_baki_debet_non_collateral_biro, score_biro, cluster, cmo_cluster FROM trx_filtering WITH (nolock) WHERE prospect_id = ?", prospectID).Scan(&filtering).Error; err != nil {
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

	db := r.newKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmb.Raw("SELECT * FROM m_mapping_elaborate_ltv WITH (nolock) WHERE result_pefindo = ? AND cluster = ? ", resultPefindo, cluster).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) SaveTrxElaborateLTV(data entity.TrxElaborateLTV) (err error) {
	data.CreatedAt = time.Now()
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmb.BeginTx(ctx, &x)
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

func (r repoHandler) GetMappingVehicleAge(vehicleAge int, cluster string, bpkbNameType, tenor int, resultPefindo string, af float64) (data entity.MappingVehicleAge, err error) {

	query := `SELECT TOP 1 * FROM m_mapping_vehicle_age WHERE vehicle_age_start <= ? AND vehicle_age_end >= ? AND cluster LIKE ? AND bpkb_name_type = ? AND tenor_start <= ? AND tenor_end >= ? AND result_pbk LIKE ? AND af_start < ? AND af_end >= ?`

	if err = r.newKmb.Raw(query, vehicleAge, vehicleAge, fmt.Sprintf("%%%s%%", cluster), bpkbNameType, tenor, tenor, fmt.Sprintf("%%%s%%", resultPefindo), af, af).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return
}

func (r repoHandler) GetScoreGenerator(zipCode string) (score entity.ScoreGenerator, err error) {

	if err = r.scorePro.Raw(fmt.Sprintf(`SELECT TOP 1 x.* 
	FROM
	(
		SELECT 
		a.[key],
		b.id AS score_generators_id
		FROM [dbo].[score_models_rules_data] a
		INNER JOIN score_generators b
		ON b.id = a.score_generators
		WHERE (a.[key] = 'first_residence_zipcode_2w_jabo' AND a.[value] = '%s')
		OR (a.[key] = 'first_residence_zipcode_2w_others' AND a.[value] = '%s')
	)x`, zipCode, zipCode)).Scan(&score).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetScoreGeneratorROAO() (score entity.ScoreGenerator, err error) {

	if err = r.scorePro.Raw(`SELECT TOP 1 x.* 
	FROM
	(
		SELECT 
		a.[key],
		b.id AS score_generators_id
		FROM [dbo].[score_models_rules_data] a
		INNER JOIN score_generators b
		ON b.id = a.score_generators
		WHERE a.[key] = 'first_residence_zipcode_2w_aoro'
	)x`).Scan(&score).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetTrxDetailBIro(prospectID string) (trxDetailBiro []entity.TrxDetailBiro, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmb.Raw(fmt.Sprintf("SELECT * FROM trx_detail_biro WITH (nolock) WHERE prospect_id = '%s'", prospectID)).Scan(&trxDetailBiro).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetActiveLoanTypeLast6M(customerID string) (score entity.GetActiveLoanTypeLast6M, err error) {

	if err = r.confins.Raw(fmt.Sprintf(`SELECT CustomerID, Concat([1],' ',';',' ',[2],' ',';',' ',[3]) AS active_loanType_last6m FROM
	( SELECT * FROM
		( SELECT  CustomerID, PRODUCT, Seq_PRODUCT FROM
			( SELECT DISTINCT CustomerID,PRODUCT, ROW_NUMBER() OVER (PARTITION BY CustomerID Order By PRODUCT DESC) AS Seq_PRODUCT FROM
				( SELECT DISTINCT CustomerID,
						CASE WHEN ContractStatus in ('ICP','PRP','LIV','RRD','ICL','INV') and MOB<=-1 and MOB>=-6 THEN PRODUCT
						END AS 'PRODUCT'
					FROM
					( SELECT CustomerID,A.ApplicationID,DATEDIFF(MM, GETDATE(), AgingDate) AS 'MOB', CAST(aa.AssetTypeID AS int) AS PRODUCT, ContractStatus FROM																	   
						( SELECT * FROM Agreement a WITH (NOLOCK) WHERE a.CustomerID = '%s' 
						)A
						LEFT JOIN
						( SELECT DISTINCT ApplicationID,AgingDate,EndPastDueDays FROM SBOAging WITH (NOLOCK)
							WHERE ApplicationID IN (SELECT DISTINCT a.ApplicationID  FROM Agreement a WITH (NOLOCK)) AND AgingDate=EOMONTH(AgingDate)
						)B ON A.ApplicationID=B.ApplicationID
						LEFT JOIN AgreementAsset aa WITH (NOLOCK) ON A.ApplicationID = aa.ApplicationID
					)S
				)T
			)U
		) AS SourceTable PIVOT(AVG(PRODUCT) FOR Seq_PRODUCT IN([1],[2],[3])) AS PivotTable
	)V`, customerID)).Scan(&score).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetActiveLoanTypeLast24M(customerID string) (score entity.GetActiveLoanTypeLast24M, err error) {

	if err = r.confins.Raw(fmt.Sprintf(`SELECT a.AgreementNo, MIN(DATEDIFF(MM, GETDATE(), s.AgingDate)) AS 'MOB' FROM Agreement a WITH (NOLOCK)
	LEFT JOIN SBOAging s WITH (NOLOCK) ON s.ApplicationId = a.ApplicationID
	WHERE a.ContractStatus in ('ICP','PRP','LIV','RRD','ICL','INV') 
	AND DATEDIFF(MM, GETDATE(), s.AgingDate)<=-7 AND DATEDIFF(MM, GETDATE(), s.AgingDate)>=-24
	AND a.CustomerID = '%s' GROUP BY a.AgreementNo`, customerID)).Scan(&score).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetMoblast(customerID string) (score entity.GetMoblast, err error) {

	if err = r.confins.Raw(fmt.Sprintf(`SELECT TOP 1 DATEDIFF(MM, GoLiveDate, GETDATE()) AS 'moblast' FROM Agreement a WITH (NOLOCK) 
	WHERE a.CustomerID = '%s' AND a.GoLiveDate IS NOT NULL ORDER BY a.GoLiveDate DESC`, customerID)).Scan(&score).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) SavePrincipleStepThree(data entity.TrxPrincipleStepThree) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.TrxPrincipleStatus{
			ProspectID: data.ProspectID,
			IDNumber:   data.IDNumber,
			Step:       3,
			Decision:   data.Decision,
			UpdatedAt:  time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return nil
	})

}
