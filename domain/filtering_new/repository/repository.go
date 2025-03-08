package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/filtering_new/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
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

// Helper function to encrypt multiple strings in a single database call
func encryptBatch(db *gorm.DB, values []string) ([]string, error) {
	if len(values) == 0 {
		return []string{}, nil
	}

	// Create query that encrypts multiple values in one go
	query := "SELECT "
	params := []interface{}{}

	for i, val := range values {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("SCP.dbo.ENC_B64('SEC', ?) AS encrypt%d", i)
		params = append(params, val)
	}

	// Execute the query
	rows, err := db.Raw(query, params...).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Extract results
	if !rows.Next() {
		return nil, fmt.Errorf("no encryption results returned")
	}

	// Prepare to scan results
	scanArgs := make([]interface{}, len(values))
	results := make([]sql.NullString, len(values))
	for i := range values {
		scanArgs[i] = &results[i]
	}

	if err := rows.Scan(scanArgs...); err != nil {
		return nil, err
	}

	// Convert to string slice
	encrypted := make([]string, len(values))
	for i, v := range results {
		if v.Valid {
			encrypted[i] = v.String
		}
	}

	return encrypted, nil
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

func (r repoHandler) SaveFiltering(data entity.FilteringKMB, trxDetailBiro []entity.TrxDetailBiro, dataCMOnoFPD entity.TrxCmoNoFPD, historyCheckAsset []entity.TrxHistoryCheckingAsset, lockingSystem entity.TrxLockSystem) (err error) {

	var (
		x sql.TxOptions
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	toEncrypt := []string{}
	fieldMap := map[int]string{}
	pos := 0

	// Required customer fields
	toEncrypt = append(toEncrypt, data.IDNumber)
	fieldMap[pos] = "IDNumber"
	pos++

	toEncrypt = append(toEncrypt, data.LegalName)
	fieldMap[pos] = "LegalName"
	pos++

	toEncrypt = append(toEncrypt, data.SurgateMotherName)
	fieldMap[pos] = "SurgateMotherName"
	pos++

	// Optional spouse fields
	if data.SpouseIDNumber != nil && *data.SpouseIDNumber != "" {
		toEncrypt = append(toEncrypt, *data.SpouseIDNumber)
		fieldMap[pos] = "SpouseIDNumber"
		pos++
	}

	if data.SpouseLegalName != nil && *data.SpouseLegalName != "" {
		toEncrypt = append(toEncrypt, *data.SpouseLegalName)
		fieldMap[pos] = "SpouseLegalName"
		pos++
	}

	if data.SpouseSurgateMotherName != nil && *data.SpouseSurgateMotherName != "" {
		toEncrypt = append(toEncrypt, *data.SpouseSurgateMotherName)
		fieldMap[pos] = "SpouseSurgateMotherName"
		pos++
	}

	// Encrypt all values in a single DB call
	encrypted, err := encryptBatch(db, toEncrypt)
	if err != nil {
		return err
	}

	// Apply encrypted values back to the struct
	for pos, fieldName := range fieldMap {
		switch fieldName {
		case "IDNumber":
			data.IDNumber = encrypted[pos]
		case "LegalName":
			data.LegalName = encrypted[pos]
		case "SurgateMotherName":
			data.SurgateMotherName = encrypted[pos]
		case "SpouseIDNumber":
			encryptedVal := encrypted[pos]
			data.SpouseIDNumber = &encryptedVal
		case "SpouseLegalName":
			encryptedVal := encrypted[pos]
			data.SpouseLegalName = &encryptedVal
		case "SpouseSurgateMotherName":
			encryptedVal := encrypted[pos]
			data.SpouseSurgateMotherName = &encryptedVal
		}
	}

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

	if len(historyCheckAsset) > 0 {
		for i, v := range historyCheckAsset {
			// Collect non-empty string values that need encryption
			toEncrypt := []string{}
			fieldMap := map[int]string{} // Maps position in batch to field name
			pos := 0

			// Collect all required fields
			if v.IDNumber != "" {
				toEncrypt = append(toEncrypt, v.IDNumber)
				fieldMap[pos] = "IDNumber"
				pos++
			}

			if v.LegalName != "" {
				toEncrypt = append(toEncrypt, v.LegalName)
				fieldMap[pos] = "LegalName"
				pos++
			}

			if v.SurgateMotherName != "" {
				toEncrypt = append(toEncrypt, v.SurgateMotherName)
				fieldMap[pos] = "SurgateMotherName"
				pos++
			}

			// Collect pointer fields
			if v.IDNumberSpouse != nil && *v.IDNumberSpouse != "" {
				toEncrypt = append(toEncrypt, *v.IDNumberSpouse)
				fieldMap[pos] = "IDNumberSpouse"
				pos++
			}

			if v.LegalNameSpouse != nil && *v.LegalNameSpouse != "" {
				toEncrypt = append(toEncrypt, *v.LegalNameSpouse)
				fieldMap[pos] = "LegalNameSpouse"
				pos++
			}

			if v.SurgateMotherNameSpouse != nil && *v.SurgateMotherNameSpouse != "" {
				toEncrypt = append(toEncrypt, *v.SurgateMotherNameSpouse)
				fieldMap[pos] = "SurgateMotherNameSpouse"
				pos++
			}

			// Encrypt all values in a single DB call
			if len(toEncrypt) > 0 {
				encrypted, err := encryptBatch(db, toEncrypt)
				if err != nil {
					return err
				}

				// Apply encrypted values back to the struct
				for pos, fieldName := range fieldMap {
					switch fieldName {
					case "IDNumber":
						historyCheckAsset[i].IDNumber = encrypted[pos]
					case "LegalName":
						historyCheckAsset[i].LegalName = encrypted[pos]
					case "SurgateMotherName":
						historyCheckAsset[i].SurgateMotherName = encrypted[pos]
					case "IDNumberSpouse":
						encryptedVal := encrypted[pos]
						historyCheckAsset[i].IDNumberSpouse = &encryptedVal
					case "LegalNameSpouse":
						encryptedVal := encrypted[pos]
						historyCheckAsset[i].LegalNameSpouse = &encryptedVal
					case "SurgateMotherNameSpouse":
						encryptedVal := encrypted[pos]
						historyCheckAsset[i].SurgateMotherNameSpouse = &encryptedVal
					}
				}
			}

			// Create the record after encryption
			if err = db.Create(&historyCheckAsset[i]).Error; err != nil {
				return
			}

			// Update locked asset status if needed
			if historyCheckAsset[i].IsAssetLocked == 1 {
				if err = db.Model(&entity.TrxHistoryCheckingAsset{}).
					Where("(ChassisNumber = ? OR EngineNumber = ?) AND IsAssetLocked = 0", historyCheckAsset[i].ChassisNumber, historyCheckAsset[i].EngineNumber).
					Update("IsAssetLocked", 1).Error; err != nil {
					return
				}
			}
		}
	}

	if lockingSystem.Reason != "" {
		var encrypted entity.EncryptString
		if err = db.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS encrypt`, lockingSystem.IDNumber)).Scan(&encrypted).Error; err != nil {
			return
		}

		lockingSystem.IDNumber = encrypted.Encrypt

		if err = db.Create(&lockingSystem).Error; err != nil {
			return
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

func (r repoHandler) getLatestRetryNumber(db *gorm.DB, chassisNumber, engineNumber, decision string) (int, error) {
	var maxRetry struct {
		LatestRetryNumber int `gorm:"column:latest_retry_number"`
	}

	err := db.Raw(`
        SELECT COALESCE(MAX(NumberOfRetry), 0) AS latest_retry_number 
        FROM trx_history_checking_asset WITH (NOLOCK)
        WHERE (ChassisNumber = ? OR EngineNumber = ?)
        AND FinalDecision = ? AND IsAssetLocked = 0
    `, chassisNumber, engineNumber, decision).Scan(&maxRetry).Error

	if err != nil {
		return 0, err
	}

	return maxRetry.LatestRetryNumber, nil
}

func (r repoHandler) GetAssetCancel(chassisNumber string, engineNumber string, lockSystemConfig response.LockSystemConfig) (historyData response.DataCheckLockAsset, found bool, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	currentDate := time.Now()
	startDate := currentDate.AddDate(0, 0, -lockSystemConfig.Data.LockAssetCheck)

	startDateStr := startDate.Format("2006-01-02")

	query := `
        SELECT TOP 1
            tri.ProspectID, tri.chassis_number, tri.engine_number, 'JOURNEY' AS source_service, ts.decision, ts.reason, ts.created_at,
            scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
            scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
            tcp.BirthDate,
            scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
            tcs.IDNumber AS IDNumber_spouse,
            tcs.LegalName AS LegalName_spouse,
            tcs.BirthDate AS BirthDate_spouse,
            tcs.SurgateMotherName AS SurgateMotherName_spouse
        FROM trx_item AS tri (NOLOCK)
        JOIN trx_status AS ts ON (tri.ProspectID = ts.ProspectID)
        JOIN trx_customer_personal AS tcp ON (tri.ProspectID = tcp.ProspectID)
        LEFT JOIN trx_customer_spouse AS tcs ON (tri.ProspectID = tcs.ProspectID)
        WHERE (tri.chassis_number = ? OR tri.engine_number = ?)
        AND ts.decision = 'CAN'
        AND ts.created_at >= ?
        ORDER BY ts.created_at ASC
    `

	err = db.Raw(query, chassisNumber, engineNumber, startDateStr).Scan(&historyData).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			latestRetryNumber, retryErr := r.getLatestRetryNumber(db, chassisNumber, engineNumber, "CAN")
			if retryErr != nil {
				return historyData, false, retryErr
			}
			historyData.LatestRetryNumber = latestRetryNumber
			return historyData, false, nil
		}
		return historyData, false, err
	}

	found = historyData.ProspectID != ""

	// If data was found, get the latest retry number
	if found {
		latestRetryNumber, retryErr := r.getLatestRetryNumber(db, historyData.ChassisNumber, historyData.EngineNumber, "CAN")
		if retryErr != nil {
			return historyData, false, retryErr
		}
		historyData.LatestRetryNumber = latestRetryNumber
	}

	return historyData, found, nil
}

func (r repoHandler) GetAssetReject(chassisNumber string, engineNumber string, lockSystemConfig response.LockSystemConfig) (historyData response.DataCheckLockAsset, found bool, err error) {
	var x sql.TxOptions
	var results []response.DataCheckLockAsset

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	currentDate := time.Now()
	startDate := currentDate.AddDate(0, 0, -lockSystemConfig.Data.LockAssetCheck)

	startDateStr := startDate.Format("2006-01-02")

	query := `
		WITH journey_results AS (
			SELECT TOP 1
				tri.ProspectID, tri.chassis_number, tri.engine_number, 'JOURNEY' AS source_service, 
				ts.decision, ts.reason, ts.created_at,
				scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
				scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
				tcp.BirthDate,
				scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
				tcs.IDNumber AS IDNumber_spouse,
				tcs.LegalName AS LegalName_spouse,
				tcs.BirthDate AS BirthDate_spouse,
				tcs.SurgateMotherName AS SurgateMotherName_spouse
			FROM trx_item AS tri WITH (NOLOCK)
			JOIN trx_status AS ts WITH (NOLOCK) ON (tri.ProspectID = ts.ProspectID)
			JOIN trx_customer_personal AS tcp WITH (NOLOCK) ON (tri.ProspectID = tcp.ProspectID)
			LEFT JOIN trx_customer_spouse AS tcs WITH (NOLOCK) ON (tri.ProspectID = tcs.ProspectID)
			WHERE (tri.chassis_number = ? OR tri.engine_number = ?)
			AND ts.decision = 'REJ'
			AND ts.created_at >= ?
			ORDER BY ts.created_at ASC
		),
		filtering_results AS (
			SELECT TOP 1
				tf.prospect_id AS ProspectID, tf.chassis_number, tf.engine_number, 'FILTERING' AS source_service,
				CASE
					WHEN tf.next_process = 0 THEN 'REJ'
					ELSE NULL
				END AS decision, 
				tf.reason, tf.created_at,
				scp.dbo.DEC_B64('SEC', tf.id_number) AS IDNumber,
				scp.dbo.DEC_B64('SEC', tf.legal_name) AS LegalName,
				tf.birth_date AS BirthDate,
				scp.dbo.DEC_B64('SEC', tf.surgate_mother_name) AS SurgateMotherName,
				scp.dbo.DEC_B64('SEC', tf.spouse_id_number) AS IDNumber_spouse,
				scp.dbo.DEC_B64('SEC', tf.spouse_legal_name) AS LegalName_spouse,
				tf.spouse_birth_date AS BirthDate_spouse,
				scp.dbo.DEC_B64('SEC', tf.spouse_surgate_mother_name) AS SurgateMotherName_spouse
			FROM trx_filtering AS tf WITH (NOLOCK)
			WHERE (tf.chassis_number = ? OR tf.engine_number = ?)
			AND tf.next_process = 0
			AND tf.id_number IS NOT NULL
			AND tf.legal_name IS NOT NULL
			AND tf.surgate_mother_name IS NOT NULL
			AND tf.birth_date IS NOT NULL
			AND tf.created_at >= ?
			ORDER BY tf.created_at ASC
		)

		SELECT * FROM journey_results WITH (NOLOCK)
		UNION ALL
		SELECT * FROM filtering_results WITH (NOLOCK)
		ORDER BY created_at ASC
	`

	if err = db.Raw(query,
		chassisNumber, engineNumber, startDateStr, // Parameters for first query
		chassisNumber, engineNumber, startDateStr). // Parameters for second query
		Scan(&results).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return historyData, false, nil
		}
		return historyData, false, err
	}

	if len(results) > 0 {
		historyData = results[0]
		found = true

		latestRetryNumber, retryErr := r.getLatestRetryNumber(db, historyData.ChassisNumber, historyData.EngineNumber, "REJ")
		if retryErr != nil {
			return historyData, false, retryErr
		}
		historyData.LatestRetryNumber = latestRetryNumber
	}

	return historyData, found, nil
}
