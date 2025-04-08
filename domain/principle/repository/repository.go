package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"los-kmb-api/domain/principle/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
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
			return err
		}
		return nil
	})

}

func (r repoHandler) GetPrincipleStepOne(prospectID string) (data entity.TrxPrincipleStepOne, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmb.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdatePrincipleStepOne(prospectID string, data entity.TrxPrincipleStepOne) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {
		var existing entity.TrxPrincipleStepOne
		if err := tx.Where("ProspectID = ?", prospectID).Order("created_at DESC").First(&existing).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.TrxPrincipleStepOne{}).
			Where("ProspectID = ?", data.ProspectID).
			Updates(&data).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r repoHandler) SavePrincipleStepTwo(data entity.TrxPrincipleStepTwo) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.TrxPrincipleStatus{}).
			Where("ProspectID = ?", data.ProspectID).
			Updates(&entity.TrxPrincipleStatus{
				ProspectID: data.ProspectID,
				IDNumber:   data.IDNumber,
				Step:       2,
				Decision:   data.Decision,
				UpdatedAt:  time.Now(),
			}).Error; err != nil {
			return err
		}
		return nil
	})

}

func (r repoHandler) GetPrincipleStepTwo(prospectID string) (data entity.TrxPrincipleStepTwo, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_two WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmb.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdatePrincipleStepTwo(prospectID string, data entity.TrxPrincipleStepTwo) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {
		var existing entity.TrxPrincipleStepTwo
		if err := tx.Where("ProspectID = ?", prospectID).Order("created_at DESC").First(&existing).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.TrxPrincipleStepTwo{}).
			Where("ProspectID = ?", data.ProspectID).
			Updates(&data).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r repoHandler) GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmb.Raw("SELECT bpkb_name, customer_status, customer_status_kmb, decision, reason, is_blacklist, next_process, max_overdue_biro, max_overdue_last12months_biro, customer_segment, total_baki_debet_non_collateral_biro, total_installment_amount_biro, score_biro, cluster, cmo_cluster, FORMAT(rrd_date, 'yyyy-MM-ddTHH:mm:ss') + 'Z' AS rrd_date, created_at FROM trx_filtering WITH (nolock) WHERE prospect_id = ?", prospectID).Scan(&filtering).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetMappingElaborateLTV(resultPefindo, cluster string) (data []entity.MappingElaborateLTV, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

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

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

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

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

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

		if err := tx.Model(&entity.TrxPrincipleStatus{}).
			Where("ProspectID = ?", data.ProspectID).
			Updates(&entity.TrxPrincipleStatus{
				ProspectID: data.ProspectID,
				IDNumber:   data.IDNumber,
				Step:       3,
				Decision:   data.Decision,
				UpdatedAt:  time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})

}

func (r repoHandler) GetPrincipleStepThree(prospectID string) (data entity.TrxPrincipleStepThree, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_three WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmb.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SavePrincipleEmergencyContact(data entity.TrxPrincipleEmergencyContact, idNumber string) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {
		var existing entity.TrxPrincipleEmergencyContact
		if err := tx.Raw("SELECT TOP 1 * FROM trx_principle_emergency_contact WHERE ProspectID = ?", data.ProspectID).Scan(&existing).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&data).Error; err != nil {
					return err
				}
				if err := tx.Model(&entity.TrxPrincipleStatus{}).
					Where("ProspectID = ?", data.ProspectID).
					Updates(&entity.TrxPrincipleStatus{
						ProspectID: data.ProspectID,
						IDNumber:   idNumber,
						Step:       4,
						Decision:   constant.DECISION_CREDIT_PROCESS,
						UpdatedAt:  time.Now(),
					}).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err := tx.Model(&existing).
				Where("ProspectID = ?", data.ProspectID).
				Updates(data).Error; err != nil {
				return err
			}

			if err := tx.Model(&entity.TrxPrincipleStatus{}).
				Where("ProspectID = ? AND Step = ?", data.ProspectID, 4).
				Updates(&entity.TrxPrincipleStatus{
					IDNumber:  idNumber,
					Decision:  constant.DECISION_CREDIT_PROCESS,
					UpdatedAt: time.Now(),
				}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r repoHandler) GetPrincipleEmergencyContact(prospectID string) (data entity.TrxPrincipleEmergencyContact, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_emergency_contact WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmb.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveToWorker(data []entity.TrxWorker) (err error) {

	return r.los.Transaction(func(tx *gorm.DB) error {
		for _, worker := range data {
			if err := tx.Create(&worker).Error; err != nil {
				return err
			}
		}

		return nil
	})

}

func (r repoHandler) GetTrxWorker(prospectID, category string) (data []entity.TrxWorker, err error) {
	query := fmt.Sprintf("SELECT * FROM trx_worker WITH (nolock) WHERE ProspectID = '%s' AND [category] = '%s'", prospectID, category)

	if err = r.los.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetElaborateLtv(prospectID string) (elaborateLTV entity.MappingElaborateLTV, err error) {

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT CASE WHEN mmel.ltv IS NULL THEN mmelovd.ltv ELSE mmel.ltv END AS ltv FROM trx_elaborate_ltv tel WITH (nolock) 
	LEFT JOIN m_mapping_elaborate_ltv mmel WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmel.id
	LEFT JOIN m_mapping_elaborate_ltv_ovd mmelovd WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmelovd.id 
	WHERE tel.prospect_id ='%s'`, prospectID)).Scan(&elaborateLTV).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SavePrincipleMarketingProgram(data entity.TrxPrincipleMarketingProgram) (err error) {

	if err := r.newKmb.Create(&data).Error; err != nil {
		return err
	}

	return nil

}

func (r repoHandler) GetPrincipleMarketingProgram(prospectID string) (data entity.TrxPrincipleMarketingProgram, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_marketing_program WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmb.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) ScanOrderPending() (data []entity.AutoCancel, err error) {

	query := `SELECT DISTINCT tps.ProspectID, tpso.KPMID, tpso.BranchID, tpso.AssetCode  
	FROM trx_principle_status tps WITH (nolock)
	INNER JOIN trx_principle_step_one tpso WITH (nolock)
	ON tps.ProspectID  = tpso.ProspectID 
	WHERE tps.created_at < DATEADD(day, -3, GETDATE())
	AND tps.Decision <> 'CANCEL' `

	if err = r.newKmb.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateToCancel(prospectID string) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&entity.TrxPrincipleStatus{}).
			Where("ProspectID = ?", prospectID).
			Updates(&entity.TrxPrincipleStatus{
				Decision:  constant.DECISION_CANCEL,
				UpdatedAt: time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})

}

func (r repoHandler) UpdateTrxPrincipleStatus(prospectID string, decision string, step int) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&entity.TrxPrincipleStatus{}).
			Where("ProspectID = ?", prospectID).
			Updates(&entity.TrxPrincipleStatus{
				Decision:  decision,
				Step:      step,
				UpdatedAt: time.Now(),
			}).Error; err != nil {
			return err
		}

		return nil
	})

}

func (r repoHandler) ExceedErrorStepOne(kpmId int) int {
	var trxError entity.TrxPrincipleError

	result := r.newKmb.Raw("SELECT KpmID FROM trx_principle_error WITH (nolock) WHERE KpmID = ? AND step = 1 AND created_at >= DATEADD (HOUR , -1 , GETDATE())", kpmId).Scan(&trxError)

	return int(result.RowsAffected)
}

func (r repoHandler) ExceedErrorStepTwo(prospectId string) int {

	var trxError entity.TrxPrincipleError

	result := r.newKmb.Raw("SELECT KpmID FROM trx_principle_error WITH (nolock) WHERE ProspectID = ? AND step = 2 AND created_at >= DATEADD (HOUR , -1 , GETDATE())", prospectId).Scan(&trxError)

	return int(result.RowsAffected)

}

func (r repoHandler) ExceedErrorStepThree(prospectId string) int {

	var trxError entity.TrxPrincipleError

	result := r.newKmb.Raw("SELECT KpmID FROM trx_principle_error WITH (nolock) WHERE ProspectID = ? AND step = 3 AND created_at >= DATEADD (HOUR , -1 , GETDATE())", prospectId).Scan(&trxError)

	return int(result.RowsAffected)

}

func (r repoHandler) GetTrxStatus(prospectID string) (status entity.TrxStatus, err error) {

	if err = r.newKmb.Raw("SELECT activity, decision, source_decision FROM trx_status WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&status).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetBannedChassisNumber(chassisNumber string) (data entity.TrxBannedChassisNumber, err error) {

	date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT * FROM trx_banned_chassis_number WITH (nolock) WHERE chassis_number = '%s' AND CAST(created_at as DATE) >= '%s'`, chassisNumber, date)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetMappingNegativeCustomer(req response.NegativeCustomer) (data entity.MappingNegativeCustomer, err error) {

	query := `SELECT TOP 1 * FROM m_mapping_negative_customer WHERE is_active = ? AND bad_type = ? AND is_blacklist = ? AND is_highrisk = ?`

	if err = r.newKmb.Raw(query, req.IsActive, req.BadType, req.IsBlacklist, req.IsHighrisk).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return
}

func (r repoHandler) SaveTrxKPM(data entity.TrxKPM) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {

		var encrypted entity.Encrypted
		if err := tx.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName, SCP.dbo.ENC_B64('SEC','%s') AS SurgateMotherName, SCP.dbo.ENC_B64('SEC','%s') AS MobilePhone, 
			SCP.dbo.ENC_B64('SEC','%s') AS Email, SCP.dbo.ENC_B64('SEC','%s') AS BirthPlace, SCP.dbo.ENC_B64('SEC','%s') AS ResidenceAddress, SCP.dbo.ENC_B64('SEC','%s') AS IDNumber`,
			data.LegalName, data.SurgateMotherName, data.MobilePhone, data.Email, data.BirthPlace, data.ResidenceAddress, data.IDNumber)).Scan(&encrypted).Error; err != nil {
			return err
		}

		data.LegalName = encrypted.LegalName
		data.SurgateMotherName = encrypted.SurgateMotherName
		data.MobilePhone = encrypted.MobilePhone
		data.Email = encrypted.Email
		data.BirthPlace = encrypted.BirthPlace
		data.ResidenceAddress = encrypted.ResidenceAddress
		data.IDNumber = encrypted.IDNumber

		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		return nil
	})

}

func (r repoHandler) SaveTrxKPMStatus(data entity.TrxKPMStatus) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r repoHandler) GetTrxKPM(prospectID string) (data entity.TrxKPM, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_kpm WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmb.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	var decrypted entity.Encrypted

	if err = r.newKmb.Raw(fmt.Sprintf(`SELECT scp.dbo.DEC_B64('SEC', '%s') AS LegalName, scp.dbo.DEC_B64('SEC','%s') AS SurgateMotherName,
		scp.dbo.DEC_B64('SEC', '%s') AS MobilePhone, scp.dbo.DEC_B64('SEC', '%s') AS Email,
		scp.dbo.DEC_B64('SEC', '%s') AS BirthPlace, scp.dbo.DEC_B64('SEC','%s') AS ResidenceAddress,
		scp.dbo.DEC_B64('SEC', '%s') AS IDNumber`, data.LegalName, data.SurgateMotherName, data.MobilePhone,
		data.Email, data.BirthPlace, data.ResidenceAddress, data.IDNumber)).Scan(&decrypted).Error; err != nil {
		return
	}

	data.LegalName = decrypted.LegalName
	data.SurgateMotherName = decrypted.SurgateMotherName
	data.MobilePhone = decrypted.MobilePhone
	data.Email = decrypted.Email
	data.BirthPlace = decrypted.BirthPlace
	data.ResidenceAddress = decrypted.ResidenceAddress
	data.IDNumber = decrypted.IDNumber

	return
}

func (r repoHandler) ExceedErrorTrxKPM(kpmId int) int {

	var trxError entity.TrxKPMError

	result := r.newKmb.Raw("SELECT KpmID FROM trx_kpm_error WITH (nolock) WHERE KpmID = ? AND created_at >= DATEADD (HOUR , -1 , GETDATE())", kpmId).Scan(&trxError)

	return int(result.RowsAffected)

}

func (r repoHandler) GetReadjustCountTrxKPM(prospectId string) int {

	var trxKPM entity.TrxKPM

	result := r.newKmb.Raw("SELECT id FROM trx_kpm WITH (nolock) WHERE ProspectID = ? AND Decision = ?", prospectId, constant.DECISION_KPM_READJUST).Scan(&trxKPM)

	return int(result.RowsAffected)

}

func (r repoHandler) GetTrxKPMStatus(IDNumber string) (data entity.TrxKPMStatus, err error) {

	if err = r.newKmb.Raw(fmt.Sprintf("SELECT TOP 1 tks.* FROM trx_kpm_status tks WITH (nolock) JOIN trx_kpm tk ON tks.ProspectID = tk.ProspectID WHERE tk.IDNumber = SCP.dbo.ENC_B64('SEC','%s') ORDER BY tks.created_at DESC", IDNumber)).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetTrxKPMStatusHistory(req request.History2Wilen) (data []entity.TrxKPMStatusHistory, err error) {
	args := make([]interface{}, 0)
	whereQuery := "WHERE 1=1"

	if req.ProspectID != nil {
		whereQuery += " AND s.ProspectID = ?"
		args = append(args, *req.ProspectID)
	}

	if req.StartDate != nil && req.EndDate != nil {
		whereQuery += " AND s.created_at BETWEEN ? AND ?"
		args = append(args, *req.StartDate, *req.EndDate)
	} else if req.StartDate != nil {
		whereQuery += " AND s.created_at >= ?"
		args = append(args, *req.StartDate)
	} else if req.EndDate != nil {
		whereQuery += " AND s.created_at <= ?"
		args = append(args, *req.EndDate)
	}

	if req.Status != nil {
		whereQuery += " AND s.Decision = ?"
		args = append(args, *req.Status)
	}

	query := fmt.Sprintf(`
		SELECT s.ProspectID as ProspectID, s.id as id, s.Decision as Decision, s.created_at as created_at, k.KpmID as KpmID, scp.dbo.DEC_B64('SEC', k.IDNumber) as IDNumber, k.ReferralCode as ReferralCode, k.LoanAmount as LoanAmount
		FROM trx_kpm_status AS s WITH (nolock)
		LEFT JOIN (
		  SELECT *, ROW_NUMBER() OVER (PARTITION BY ProspectID ORDER BY created_at DESC) AS rn
		  FROM trx_kpm
		) AS k ON s.ProspectID = k.ProspectID AND k.rn = 1 %s
	`, whereQuery)

	if err = r.newKmb.Raw(query, args...).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateTrxKPMDecision(id string, prospectID string, decision string) (err error) {

	return r.newKmb.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&entity.TrxKPM{}).
			Where("id = ?", id).
			Updates(&entity.TrxKPM{
				Decision:  decision,
				UpdatedAt: time.Now(),
			}).Error; err != nil {
			return err
		}

		data := entity.TrxKPMStatus{
			ID:         utils.GenerateUUID(),
			ProspectID: prospectID,
			Decision:   decision,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if err := tx.Create(&data).Error; err != nil {
			return err
		}

		return nil
	})

}
