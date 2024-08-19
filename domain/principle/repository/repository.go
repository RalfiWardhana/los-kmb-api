package repository

import (
	"context"
	"database/sql"
	"fmt"
	"los-kmb-api/domain/principle/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

type repoHandler struct {
	newKmb *gorm.DB
	los    *gorm.DB
}

func NewRepository(newKmb, los *gorm.DB) interfaces.Repository {
	return &repoHandler{
		newKmb: newKmb,
		los:    los,
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
