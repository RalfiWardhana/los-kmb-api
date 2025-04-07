package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	losDB      *gorm.DB
	logsDB     *gorm.DB
	confinsDB  *gorm.DB
	stagingDB  *gorm.DB
	wgOffDB    *gorm.DB
	kmbOffDB   *gorm.DB
	newKmbDB   *gorm.DB
	scoreProDB *gorm.DB
}

func NewRepository(los, logs, confins, staging, newKmbDB, scorePro *gorm.DB) interfaces.Repository {
	return &repoHandler{
		losDB:      los,
		logsDB:     logs,
		confinsDB:  confins,
		stagingDB:  staging,
		newKmbDB:   newKmbDB,
		scoreProDB: scorePro,
	}
}

func (r repoHandler) ScanTrxMaster(prospectID string) (countMaster int, err error) {

	var (
		master []entity.TrxMaster
	)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`
		SELECT tm.ProspectID FROM trx_master tm WITH (nolock) 
		LEFT JOIN trx_status ts WITH (nolock) ON tm.ProspectID = ts.ProspectID 
		WHERE tm.ProspectID = '%s' AND ((ts.activity = '%s' AND ts.source_decision = '%s') OR (ts.activity != '%s' OR ts.source_decision != '%s'))`,
		prospectID, constant.ACTIVITY_UNPROCESS, constant.PRESCREENING, constant.ACTIVITY_UNPROCESS, constant.SOURCE_DECISION_DUPCHECK)).Scan(&master).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	countMaster = len(master)

	return
}

func (r repoHandler) ScanTrxPrescreening(prospectID string) (count int, err error) {

	var (
		trxPrescreening []entity.TrxPrescreening
	)

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT ProspectID FROM trx_prescreening WITH (nolock) WHERE ProspectID = '%s'", prospectID)).Scan(&trxPrescreening).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	count = len(trxPrescreening)

	return
}

func (r repoHandler) GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error) {
	var x sql.TxOptions

	resultValid := os.Getenv("BIRO_VALID_DAYS")

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmbDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT * FROM trx_filtering WITH (nolock) WHERE prospect_id = '%s' AND DATEADD(day, -%s, CAST(GETDATE() AS date)) <= CAST(created_at AS date) ORDER BY created_at DESC", prospectID, resultValid)).Scan(&filtering).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetFilteringForJourney(prospectID string) (filtering entity.FilteringKMB, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmbDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT * FROM trx_filtering WITH (nolock) WHERE prospect_id = '%s'", prospectID)).Scan(&filtering).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetTrxDetailBIro(prospectID string) (trxDetailBiro []entity.TrxDetailBiro, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmbDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT * FROM trx_detail_biro WITH (nolock) WHERE prospect_id = '%s'", prospectID)).Scan(&trxDetailBiro).Error; err != nil {
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
			if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s'`, statusVD, customerStatus)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		} else {
			if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND is_valid=0`, customerStatus)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		}
	} else {
		if isValid || statusVD == constant.EKYC_RTO || statusVD == constant.EKYC_NOT_CHECK {
			if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE result_vd='%s' AND status_konsumen='%s' AND kategori_status_konsumen='%s'`, statusVD, customerStatus, customerSegment)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		} else {
			if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_verify_result_v2 WITH (nolock) WHERE status_konsumen='%s' AND kategori_status_konsumen='%s' AND is_valid=0`, customerStatus, customerSegment)).Scan(&resultDukcapil).Error; err != nil {
				return
			}
		}
	}

	return
}

func (r repoHandler) GetMappingDukcapil(statusVD, statusFR, customerStatus, customerSegment string) (resultDukcapil entity.MappingResultDukcapil, err error) {
	if customerStatus == constant.STATUS_KONSUMEN_NEW {
		if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s'`, statusVD, statusFR, customerStatus)).Scan(&resultDukcapil).Error; err != nil {
			return
		}
	} else {
		if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM kmb_dukcapil_mapping_result_v2 WITH (nolock) WHERE result_vd='%s' AND result_fr='%s' AND status_konsumen='%s' AND kategori_status_konsumen='%s'`, statusVD, statusFR, customerStatus, customerSegment)).Scan(&resultDukcapil).Error; err != nil {
			return
		}
	}

	return
}

func (r repoHandler) MasterMappingCluster(req entity.MasterMappingCluster) (data entity.MasterMappingCluster, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.losDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT * FROM dbo.kmb_mapping_cluster_branch WITH (nolock) WHERE branch_id = ? AND customer_status = ? AND bpkb_name_type = ?", req.BranchID, req.CustomerStatus, req.BpkbNameType).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) MasterMappingMaxDSR(req entity.MasterMappingMaxDSR) (data entity.MasterMappingMaxDSR, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.losDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("SELECT * FROM dbo.kmb_mapping_cluster_dsr WITH (nolock) WHERE cluster = ?", req.Cluster).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.DATA_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetScoreGenerator(zipCode string) (score entity.ScoreGenerator, err error) {

	if err = r.scoreProDB.Raw(fmt.Sprintf(`SELECT TOP 1 x.* 
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

	if err = r.scoreProDB.Raw(`SELECT TOP 1 x.* 
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

func (r repoHandler) GetActiveLoanTypeLast6M(customerID string) (score entity.GetActiveLoanTypeLast6M, err error) {

	if err = r.confinsDB.Raw(fmt.Sprintf(`SELECT CustomerID, Concat([1],' ',';',' ',[2],' ',';',' ',[3]) AS active_loanType_last6m FROM
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

	if err = r.confinsDB.Raw(fmt.Sprintf(`SELECT a.AgreementNo, MIN(DATEDIFF(MM, GETDATE(), s.AgingDate)) AS 'MOB' FROM Agreement a WITH (NOLOCK)
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

	if err = r.confinsDB.Raw(fmt.Sprintf(`SELECT TOP 1 DATEDIFF(MM, GoLiveDate, GETDATE()) AS 'moblast' FROM Agreement a WITH (NOLOCK) 
	WHERE a.CustomerID = '%s' AND a.GoLiveDate IS NOT NULL ORDER BY a.GoLiveDate DESC`, customerID)).Scan(&score).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetMappingDeviasi(prospectID string) (confirmDeviasi entity.ConfirmDeviasi, err error) {
	// cek kuota deviasi
	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT ta.NTF, mbd.*,
			CASE 
				WHEN mbd.balance_amount >= ta.NTF AND mbd.balance_account > 0 THEN 1
				ELSE 0
			END AS deviasi
			FROM trx_master tm 
			LEFT JOIN trx_apk ta ON tm.ProspectID = ta.ProspectID 
			LEFT JOIN m_branch_deviasi mbd ON tm.BranchID = mbd.BranchID 
			WHERE tm.ProspectID = '%s'`, prospectID)).Scan(&confirmDeviasi).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetElaborateLtv(prospectID string) (elaborateLTV entity.MappingElaborateLTV, err error) {

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT CASE WHEN mmel.ltv IS NULL THEN mmelovd.ltv ELSE mmel.ltv END AS ltv FROM trx_elaborate_ltv tel WITH (nolock) 
	LEFT JOIN m_mapping_elaborate_ltv mmel WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmel.id
	LEFT JOIN m_mapping_elaborate_ltv_ovd mmelovd WITH (nolock) ON tel.m_mapping_elaborate_ltv_id = mmelovd.id 
	WHERE tel.prospect_id ='%s'`, prospectID)).Scan(&elaborateLTV).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetMasterBranch(branchID string) (masterBranch entity.MasterBranch, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT branch_category FROM kmb_master_branch_category kmbc WITH (nolock) WHERE branch_id = '%s'`, branchID)).Scan(&masterBranch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetMappingElaborateIncome(mappingElaborateIncome entity.MappingElaborateIncome) (result entity.MappingElaborateIncome, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT branch_category, estimation_income, status_konsumen, bpkb_name_type, scoreband, worst_24mth, [result] 
	FROM kmb_mapping_treatment_elaborated_income WITH (nolock) WHERE estimation_income='%s' AND status_konsumen='%s' AND bpkb_name_type=%d AND scoreband='%s' AND worst_24mth='%s' 
	AND branch_category = '%s'`, mappingElaborateIncome.EstimationIncome, mappingElaborateIncome.StatusKonsumen, mappingElaborateIncome.BPKBNameType, mappingElaborateIncome.Scoreband, mappingElaborateIncome.Worst24Mth, mappingElaborateIncome.BranchCategory)).Scan(&result).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (newErr error) {

	location, _ := time.LoadLocation("Asia/Jakarta")

	formatBirthDate, _ := time.ParseInLocation("2006-01-02", data.CustomerPersonal.BirthDate, location)

	var formatIDExpired interface{}

	if data.CustomerPersonal.ExpiredDate != nil {
		formatIDExpired, _ = time.ParseInLocation("2006-01-02", *data.CustomerPersonal.ExpiredDate, location)
	}

	formatTaxDate, _ := time.ParseInLocation("2006-01-02", data.Item.TaxDate, location)
	formatSTNKExpired, _ := time.ParseInLocation("2006-01-02", data.Item.STNKExpiredDate, location)

	var logInfo interface{}

	newErr = r.newKmbDB.Transaction(func(tx *gorm.DB) error {

		//save data from payload
		if countTrx == 0 {
			var legalAddress, residenceAddress, companyAddress, emergencyAddress, mailingAddress, ownerAddress, locationAddress string
			for _, v := range data.Address {
				if v.Type == "LEGAL" {
					legalAddress = v.Address
				} else if v.Type == "RESIDENCE" {
					residenceAddress = v.Address
				} else if v.Type == "COMPANY" {
					companyAddress = v.Address
				} else if v.Type == "EMERGENCY" {
					emergencyAddress = v.Address
				} else if v.Type == "LOCATION" {
					locationAddress = v.Address
				} else if v.Type == "MAILING" {
					mailingAddress = v.Address
				} else {
					ownerAddress = v.Address
				}
			}

			var encrypted entity.Encrypted

			if err := tx.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName, SCP.dbo.ENC_B64('SEC','%s') AS SurgateMotherName, SCP.dbo.ENC_B64('SEC','%s') AS Email,
			SCP.dbo.ENC_B64('SEC','%s') AS MobilePhone, SCP.dbo.ENC_B64('SEC','%s') AS BirthPlace, SCP.dbo.ENC_B64('SEC','%s') AS ResidenceAddress, SCP.dbo.ENC_B64('SEC','%s') AS LegalAddress,
			SCP.dbo.ENC_B64('SEC','%s') AS IDNumber,SCP.dbo.ENC_B64('SEC','%s') AS FullName, SCP.dbo.ENC_B64('SEC','%s') AS CompanyAddress, SCP.dbo.ENC_B64('SEC','%s') AS EmergencyAddress,
			SCP.dbo.ENC_B64('SEC','%s') AS LocationAddress, SCP.dbo.ENC_B64('SEC','%s') AS MailingAddress, SCP.dbo.ENC_B64('SEC','%s') AS OwnerAddress`,
				data.CustomerPersonal.LegalName, data.CustomerPersonal.SurgateMotherName, data.CustomerPersonal.Email, data.CustomerPersonal.MobilePhone, data.CustomerPersonal.BirthPlace,
				residenceAddress, legalAddress, data.CustomerPersonal.IDNumber, data.CustomerPersonal.FullName, companyAddress, emergencyAddress,
				locationAddress, mailingAddress, ownerAddress)).Scan(&encrypted).Error; err != nil {
				return err
			}

			master := entity.TrxMaster{
				ProspectID:        data.Transaction.ProspectID,
				BranchID:          data.Transaction.BranchID,
				Channel:           data.Transaction.Channel,
				Lob:               data.Transaction.Lob,
				IncomingSource:    data.Transaction.IncomingSource,
				ApplicationSource: data.Transaction.ApplicationSource,
				OrderAt:           data.Transaction.OrderAt,
			}

			logInfo = master

			if err := tx.Create(&master).Error; err != nil {
				return err
			}

			if trxPrescreening.Decision != "" {
				logInfo = trxPrescreening

				if err := tx.Create(&trxPrescreening).Error; err != nil {
					return err
				}
			}

			for idx, v := range data.Address {
				if v.Type == "LEGAL" {
					data.Address[idx].Address = encrypted.LegalAddress
				} else if v.Type == "RESIDENCE" {
					data.Address[idx].Address = encrypted.ResidenceAddress
				} else if v.Type == "COMPANY" {
					data.Address[idx].Address = encrypted.CompanyAddress
				} else if v.Type == "EMERGENCY" {
					data.Address[idx].Address = encrypted.EmergencyAddress
				} else if v.Type == "MAILING" {
					data.Address[idx].Address = encrypted.MailingAddress
				} else if v.Type == "LOCATION" {
					data.Address[idx].Address = encrypted.LocationAddress
				} else {
					data.Address[idx].Address = encrypted.OwnerAddress
				}
			}

			for i := 0; i < len(data.Address); i++ {

				address := entity.CustomerAddress{
					ProspectID: data.Transaction.ProspectID,
					Type:       data.Address[i].Type,
					Address:    data.Address[i].Address,
					Rt:         data.Address[i].Rt,
					Rw:         data.Address[i].Rw,
					Kelurahan:  data.Address[i].Kelurahan,
					Kecamatan:  data.Address[i].Kecamatan,
					City:       data.Address[i].City,
					ZipCode:    data.Address[i].ZipCode,
					AreaPhone:  data.Address[i].AreaPhone,
					Phone:      data.Address[i].Phone,
				}

				logInfo = address

				if err := tx.Create(&address).Error; err != nil {
					return err
				}
			}

			var surveyResult string

			for i := 0; i < len(data.CustomerPhoto); i++ {
				if data.CustomerPhoto[i].ID == "RESULT_SURVEY" {
					surveyResult = data.CustomerPhoto[i].Url
					continue
				}

				photo := entity.CustomerPhoto{
					ProspectID: data.Transaction.ProspectID,
					PhotoID:    data.CustomerPhoto[i].ID,
					Url:        data.CustomerPhoto[i].Url,
				}

				logInfo = photo

				if err := tx.Create(&photo).Error; err != nil {
					return err
				}
			}

			var livingCostAmount float64
			if data.CustomerPersonal.LivingCostAmount != nil {
				livingCostAmount = *data.CustomerPersonal.LivingCostAmount
			}

			var counterpart int
			if data.CustomerPersonal.Counterpart != nil {
				counterpart = *data.CustomerPersonal.Counterpart
			}

			personal := entity.CustomerPersonal{
				ProspectID:                 data.Transaction.ProspectID,
				IDType:                     data.CustomerPersonal.IDType,
				IDNumber:                   encrypted.IDNumber,
				FullName:                   encrypted.FullName,
				LegalName:                  encrypted.LegalName,
				BirthPlace:                 encrypted.BirthPlace,
				Education:                  data.CustomerPersonal.Education,
				BirthDate:                  formatBirthDate,
				MaritalStatus:              data.CustomerPersonal.MaritalStatus,
				SurgateMotherName:          encrypted.SurgateMotherName,
				Religion:                   data.CustomerPersonal.Religion,
				Gender:                     data.CustomerPersonal.Gender,
				PersonalNPWP:               data.CustomerPersonal.NPWP,
				MobilePhone:                encrypted.MobilePhone,
				Email:                      encrypted.Email,
				HomeStatus:                 data.CustomerPersonal.HomeStatus,
				StaySinceYear:              data.CustomerPersonal.StaySinceYear,
				StaySinceMonth:             data.CustomerPersonal.StaySinceMonth,
				ExpiredDate:                formatIDExpired,
				LivingCostAmount:           livingCostAmount,
				ExtCompanyPhone:            data.CustomerEmployment.ExtCompanyPhone,
				SourceOtherIncome:          data.CustomerEmployment.SourceOtherIncome,
				EmergencyOfficeAreaPhone:   data.CustomerEmcon.AreaPhone,
				EmergencyOfficePhone:       data.CustomerEmcon.Phone,
				PersonalCustomerType:       data.CustomerPersonal.PersonalCustomerType,
				Nationality:                data.CustomerPersonal.Nationality,
				WNACountry:                 data.CustomerPersonal.WNACountry,
				HomeLocation:               data.CustomerPersonal.HomeLocation,
				CustomerGroup:              data.CustomerPersonal.CustomerGroup,
				KKNo:                       data.CustomerPersonal.KKNo,
				BankID:                     data.CustomerPersonal.BankID,
				AccountNo:                  data.CustomerPersonal.AccountNo,
				AccountName:                data.CustomerPersonal.AccountName,
				Counterpart:                counterpart,
				DebtBusinessScale:          data.CustomerPersonal.DebtBusinessScale,
				DebtGroup:                  data.CustomerPersonal.DebtGroup,
				IsAffiliateWithPP:          data.CustomerPersonal.IsAffiliateWithPP,
				AgreetoAcceptOtherOffering: data.CustomerPersonal.AgreetoAcceptOtherOffering,
				DataType:                   data.CustomerPersonal.DataType,
				Status:                     data.CustomerPersonal.Status,
				SurveyResult:               surveyResult,
				RentFinishDate:             data.CustomerPersonal.RentFinishDate,
			}

			if trxFMF.DupcheckData.CustomerID != nil {
				personal.CustomerID = trxFMF.DupcheckData.CustomerID.(string)
			}

			if trxFMF.DupcheckData.StatusKonsumen != "" {
				personal.CustomerStatus = trxFMF.DupcheckData.StatusKonsumen
			}

			if data.CustomerPersonal.NumOfDependence != nil {
				personal.NumOfDependence = *data.CustomerPersonal.NumOfDependence
			}

			logInfo = personal

			if err := tx.Create(&personal).Error; err != nil {
				return err
			}

			var monthlyVariableIncome float64
			if data.CustomerEmployment.MonthlyVariableIncome != nil {
				monthlyVariableIncome = *data.CustomerEmployment.MonthlyVariableIncome
			}

			var spouseIncome float64
			if data.CustomerEmployment.SpouseIncome != nil {
				spouseIncome = *data.CustomerEmployment.SpouseIncome
			}

			employment := entity.CustomerEmployment{
				ProspectID:            data.Transaction.ProspectID,
				ProfessionID:          data.CustomerEmployment.ProfessionID,
				JobType:               data.CustomerEmployment.JobType,
				JobPosition:           data.CustomerEmployment.JobPosition,
				CompanyName:           data.CustomerEmployment.CompanyName,
				IndustryTypeID:        data.CustomerEmployment.IndustryTypeID,
				EmploymentSinceYear:   data.CustomerEmployment.EmploymentSinceYear,
				EmploymentSinceMonth:  data.CustomerEmployment.EmploymentSinceMonth,
				MonthlyFixedIncome:    data.CustomerEmployment.MonthlyFixedIncome,
				MonthlyVariableIncome: monthlyVariableIncome,
				SpouseIncome:          spouseIncome,
			}

			logInfo = employment

			if err := tx.Create(&employment).Error; err != nil {
				return err
			}

			var (
				adminFee float64
			)
			if data.Apk.AdminFee != nil {
				adminFee = *data.Apk.AdminFee
			}

			var (
				percentDP      float64
				fidusiaFee     float64
				interestRate   float64
				interestAmount float64
				provisionFee   float64
			)
			if data.Apk.PercentDP != nil {
				percentDP = *data.Apk.PercentDP
			}
			if data.Apk.FidusiaFee != nil {
				fidusiaFee = *data.Apk.FidusiaFee
			}
			if data.Apk.InterestRate != nil {
				interestRate = *data.Apk.InterestRate
			}
			if data.Apk.InterestAmount != nil {
				interestAmount = *data.Apk.InterestAmount
			}
			if data.Apk.ProvisionFee != nil {
				provisionFee = *data.Apk.ProvisionFee
			}

			var surveyFee float64 = 0
			if data.Apk.SurveyFee != nil {
				surveyFee = *data.Apk.SurveyFee
			}

			apk := entity.TrxApk{
				ProspectID:                  data.Transaction.ProspectID,
				Tenor:                       &data.Apk.Tenor,
				ProductOfferingID:           data.Apk.ProductOfferingID,
				ProductID:                   data.Apk.ProductID,
				NTF:                         data.Apk.NTF,
				AF:                          data.Apk.AF,
				AoID:                        data.Apk.AoID,
				DPAmount:                    data.Apk.DPAmount,
				AdminFee:                    adminFee,
				OTR:                         data.Apk.OTR,
				InstallmentAmount:           data.Apk.InstallmentAmount,
				FirstInstallment:            data.Apk.FirstInstallment,
				OtherFee:                    data.Apk.OtherFee,
				PercentDP:                   percentDP,
				AssetInsuranceFee:           data.Item.PremiumAmountToCustomer,
				LifeInsuranceFee:            data.Apk.PremiumAmountToCustomer,
				FidusiaFee:                  fidusiaFee,
				InterestRate:                interestRate,
				InsuranceAmount:             data.Apk.InsuranceAmount,
				InterestAmount:              interestAmount,
				PaymentMethod:               data.Apk.PaymentMethod,
				SurveyFee:                   surveyFee,
				IsFidusiaCovered:            data.Apk.IsFidusiaCovered,
				ProvisionFee:                provisionFee,
				InsAssetPaidBy:              data.Apk.InsAssetPaidBy,
				InsAssetPeriod:              data.Apk.InsAssetPeriod,
				EffectiveRate:               data.Apk.EffectiveRate,
				SalesmanID:                  data.Apk.SalesmanID,
				SupplierBankAccountID:       data.Apk.SupplierBankAccountID,
				LifeInsuranceCoyBranchID:    data.Apk.LifeInsuranceCoyBranchID,
				LifeInsuranceAmountCoverage: data.Apk.LifeInsuranceAmountCoverage,
				CommisionSubsidi:            data.Apk.CommisionSubsidy,
				ProductOfferingDesc:         data.Apk.ProductOfferingDesc,
				Dealer:                      data.Apk.Dealer,
				LoanAmount:                  data.Apk.LoanAmount,
				FinancePurpose:              data.Apk.FinancePurpose,
				NTFAkumulasi:                trxFMF.NTFAkumulasi,
				NTFOtherAmount:              trxFMF.NTFOtherAmount,
				NTFOtherAmountSpouse:        trxFMF.NTFOtherAmountSpouse,
				NTFOtherAmountDetail:        trxFMF.NTFOtherAmountDetail,
				NTFConfinsAmount:            trxFMF.NTFConfinsAmount,
				NTFConfins:                  trxFMF.NTFConfins,
				NTFTopup:                    trxFMF.NTFTopup,
				WayOfPayment:                data.Apk.WayOfPayment,
				StampDutyFee:                data.Apk.StampDutyFee,
			}

			logInfo = apk

			if err := tx.Create(&apk).Error; err != nil {
				return err
			}

			if data.CustomerSpouse != nil {
				spouseBirthdate, _ := time.ParseInLocation("2006-01-02", data.CustomerSpouse.BirthDate, location)

				customerSpouse := entity.CustomerSpouse{
					ProspectID:        data.Transaction.ProspectID,
					IDNumber:          data.CustomerSpouse.IDNumber,
					FullName:          data.CustomerSpouse.FullName,
					LegalName:         data.CustomerSpouse.LegalName,
					BirthDate:         spouseBirthdate,
					BirthPlace:        data.CustomerSpouse.BirthPlace,
					SurgateMotherName: data.CustomerSpouse.SurgateMotherName,
					Gender:            data.CustomerSpouse.Gender,
					CompanyPhone:      data.CustomerSpouse.CompanyPhone,
					CompanyName:       data.CustomerSpouse.CompanyName,
					MobilePhone:       data.CustomerSpouse.MobilePhone,
					ProfessionID:      data.CustomerSpouse.ProfessionID,
				}

				logInfo = customerSpouse

				if err := tx.Create(&customerSpouse).Error; err != nil {
					return err
				}

			}

			emcon := entity.CustomerEmcon{
				ProspectID:           data.Transaction.ProspectID,
				Name:                 data.CustomerEmcon.Name,
				Relationship:         data.CustomerEmcon.Relationship,
				MobilePhone:          data.CustomerEmcon.MobilePhone,
				EmconVerified:        data.CustomerEmcon.ApplicationEmconSesuai,
				VerifyBy:             data.CustomerEmcon.VerifyBy,
				VerificationWith:     data.CustomerEmcon.VerificationWith,
				KnownCustomerAddress: data.CustomerEmcon.KnownCustomerAddress,
				KnownCustomerJob:     data.CustomerEmcon.KnownCustomerJob,
			}

			logInfo = emcon

			if err := tx.Create(&emcon).Error; err != nil {
				return err
			}

			item := entity.TrxItem{
				ProspectID:                   data.Transaction.ProspectID,
				CategoryID:                   data.Item.CategoryID,
				SupplierID:                   data.Item.SupplierID,
				Qty:                          data.Item.Qty,
				AssetCode:                    data.Item.AssetCode,
				AssetDescription:             data.Item.AssetDescription,
				ManufactureYear:              data.Item.ManufactureYear,
				BPKBName:                     data.Item.BPKBName,
				OwnerAsset:                   data.Item.OwnerAsset,
				LicensePlate:                 data.Item.LicensePlate,
				Color:                        data.Item.Color,
				EngineNo:                     data.Item.NoEngine,
				ChassisNo:                    data.Item.NoChassis,
				Pos:                          data.Item.POS,
				Cc:                           data.Item.CC,
				Condition:                    data.Item.Condition,
				Region:                       data.Item.Region,
				TaxDate:                      formatTaxDate,
				STNKExpiredDate:              formatSTNKExpired,
				AssetInsuranceAmountCoverage: data.Item.AssetInsuranceAmountCoverage,
				InsAssetInsuredBy:            data.Item.InsAssetInsuredBy,
				InsuranceCoyBranchID:         data.Item.InsuranceCoyBranchID,
				CoverageType:                 data.Item.CoverageType,
				OwnerKTP:                     data.Item.OwnerKTP,
				AssetUsage:                   data.Item.AssetUsage,
				Brand:                        data.Item.Brand,
			}

			logInfo = item

			if err := tx.Create(&item).Error; err != nil {
				return err
			}

			agent := entity.TrxInfoAgent{
				ProspectID: data.Transaction.ProspectID,
				NIK:        data.Agent.CmoNik,
				Name:       data.Agent.CmoName,
				Info:       data.Agent.CmoRecom,
				RecomDate:  data.Agent.RecomDate,
			}

			logInfo = agent

			if err := tx.Create(&agent).Error; err != nil {
				return err
			}

			for i := 0; i < len(data.Surveyor); i++ {

				requestDate, _ := time.Parse("2006-01-02", data.Surveyor[i].RequestDate)
				assignDate, _ := time.Parse("2006-01-02", data.Surveyor[i].AssignDate)
				resultDate, _ := time.Parse("2006-01-02", data.Surveyor[i].ResultDate)

				surveyor := entity.TrxSurveyor{
					ProspectID:   data.Transaction.ProspectID,
					Destination:  data.Surveyor[i].Destination,
					RequestDate:  requestDate,
					AssignDate:   assignDate,
					SurveyorName: data.Surveyor[i].SurveyorName,
					ResultDate:   resultDate,
					Status:       data.Surveyor[i].Status,
				}

				logInfo = surveyor

				if err := tx.Create(&surveyor).Error; err != nil {
					return err
				}

			}

			if data.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_WRST || data.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_PRO {
				if data.CustomerOmset != nil {
					omset := *data.CustomerOmset
					for i := 0; i < len(omset); i++ {

						cusOmset := entity.CustomerOmset{
							ProspectID:        data.Transaction.ProspectID,
							SeqNo:             i + 1,
							MonthlyOmsetYear:  omset[i].MonthlyOmsetYear,
							MonthlyOmsetMonth: omset[i].MonthlyOmsetMonth,
							MonthlyOmset:      omset[i].MonthlyOmset,
						}

						logInfo = cusOmset

						if err := tx.Create(&cusOmset).Error; err != nil {
							return err
						}
					}
				}
			}
		} else {
			//update data from metrics
			var updateMap = make(map[string]interface{})
			if trxFMF.DupcheckData.CustomerID != nil {
				updateMap["CustomerID"] = trxFMF.DupcheckData.CustomerID.(string)
			}
			if trxFMF.DupcheckData.StatusKonsumen != "" {
				updateMap["CustomerStatus"] = trxFMF.DupcheckData.StatusKonsumen
			}
			if updateMap != nil {
				if err := tx.Model(&entity.CustomerPersonal{}).Where("ProspectID = ?", data.Transaction.ProspectID).Updates(updateMap).Error; err != nil {
					return err
				}
			}
		}

		// proses sudah melewati prescreening
		if countTrx > 0 || len(details) > 2 {

			// save data metrics
			// insert trx edd
			logInfo = trxFMF.TrxEDD
			if trxFMF.TrxEDD.ProspectID != "" {
				if err := tx.Create(&trxFMF.TrxEDD).Error; err != nil {
					return err
				}
			}

			// insert trx deviasi
			if trxFMF.TrxDeviasi != (entity.TrxDeviasi{}) {

				logInfo = trxFMF.TrxDeviasi

				if err := tx.Create(&trxFMF.TrxDeviasi).Error; err != nil {
					return err
				}
			}

			// insert ban pmk dsr
			if trxFMF.TrxBannedPMKDSR != (entity.TrxBannedPMKDSR{}) {

				logInfo = trxFMF.TrxBannedPMKDSR

				if err := tx.Create(&trxFMF.TrxBannedPMKDSR).Error; err != nil {
					return err
				}
			}

			// insert ban chassis number
			if trxFMF.TrxBannedChassisNumber != (entity.TrxBannedChassisNumber{}) {

				logInfo = trxFMF.TrxBannedChassisNumber

				if err := tx.Create(&trxFMF.TrxBannedChassisNumber).Error; err != nil {
					return err
				}
			}

			// internal record
			if len(trxFMF.AgreementCONFINS) > 0 {
				for _, agr := range trxFMF.AgreementCONFINS {
					dateStr := agr.AgreementDate
					agreementDate, _ := time.Parse("01/02/2006", dateStr)
					if agr.ApplicationID != "" {
						internalRecord := entity.TrxInternalRecord{
							ProspectID:           data.Transaction.ProspectID,
							CustomerID:           trxFMF.DupcheckData.CustomerID.(string),
							ApplicationID:        agr.ApplicationID,
							ProductType:          agr.ProductType,
							AgreementDate:        agreementDate,
							AssetCode:            agr.AssetCode,
							Tenor:                agr.Tenor,
							OutstandingPrincipal: agr.OutstandingPrincipal,
							InstallmentAmount:    agr.InstallmentAmount,
							ContractStatus:       agr.ContractStatus,
							CurrentCondition:     agr.CurrentCondition,
						}

						logInfo = internalRecord

						if err := tx.Create(&internalRecord).Error; err != nil {
							return err
						}
					}
				}
			}

			var ekycReason interface{}

			if trxFMF.EkycReason == constant.REASON_EKYC_INVALID {
				ekycReason = constant.REASON_TIDAK_SESUAI
			} else if trxFMF.EkycReason != nil {
				ekycReason = constant.REASON_SESUAI
			}

			var roaoAkkk response.RoaoAkkk
			roaoAkkk.InstallmentAmountOther = trxFMF.DupcheckData.InstallmentAmountOther
			roaoAkkk.InstallmentAmountOtherSpouse = trxFMF.DupcheckData.InstallmentAmountOtherSpouse
			roaoAkkk.InstallmentAmountSpouseFMF = trxFMF.DupcheckData.InstallmentAmountSpouseFMF

			if trxFMF.DupcheckData.StatusKonsumen != constant.STATUS_KONSUMEN_NEW {
				roaoAkkk.MaxOverdueDaysROAO = trxFMF.DupcheckData.MaxOverdueDaysROAO
				roaoAkkk.MaxOverdueDaysforActiveAgreement = trxFMF.DupcheckData.MaxOverdueDaysforActiveAgreement
				roaoAkkk.NumberofAgreement = trxFMF.DupcheckData.NumberofAgreement
				roaoAkkk.AgreementStatus = trxFMF.DupcheckData.AgreementStatus
				roaoAkkk.NumberOfPaidInstallment = trxFMF.DupcheckData.NumberOfPaidInstallment
				roaoAkkk.OSInstallmentDue = trxFMF.DupcheckData.OSInstallmentDue
				roaoAkkk.InstallmentAmountFMF = trxFMF.DupcheckData.InstallmentAmountFMF
				roaoAkkk.InstallmentTopup = trxFMF.DupcheckData.InstallmentTopup
			}

			akkk := entity.TrxAkkk{
				ProspectID:                   data.Transaction.ProspectID,
				ScsDate:                      trxFMF.ScsDecision.ScsDate,
				ScsScore:                     trxFMF.ScsDecision.ScsScore,
				ScsStatus:                    trxFMF.ScsDecision.ScsStatus,
				CustomerType:                 trxFMF.DupcheckData.CustomerType,
				SpouseType:                   trxFMF.DupcheckData.SpouseType,
				AgreementStatus:              roaoAkkk.AgreementStatus,
				TotalAgreementAktif:          roaoAkkk.NumberofAgreement,
				MaxOVDAgreementAktif:         roaoAkkk.MaxOverdueDaysROAO,
				LastMaxOVDAgreement:          roaoAkkk.MaxOverdueDaysforActiveAgreement,
				DSRFMF:                       trxFMF.DSRFMF,
				DSRPBK:                       trxFMF.DSRPBK,
				TotalDSR:                     trxFMF.TotalDSR,
				EkycSource:                   trxFMF.EkycSource,
				EkycSimiliarity:              trxFMF.EkycSimiliarity,
				EkycReason:                   ekycReason,
				NumberOfPaidInstallment:      roaoAkkk.NumberOfPaidInstallment,
				OSInstallmentDue:             roaoAkkk.OSInstallmentDue,
				InstallmentAmountFMF:         roaoAkkk.InstallmentAmountFMF,
				InstallmentAmountSpouseFMF:   roaoAkkk.InstallmentAmountSpouseFMF,
				InstallmentAmountOther:       roaoAkkk.InstallmentAmountOther,
				InstallmentAmountOtherSpouse: roaoAkkk.InstallmentAmountOtherSpouse,
				InstallmentTopup:             roaoAkkk.InstallmentTopup,
			}

			logInfo = akkk

			if err := tx.Create(&akkk).Error; err != nil {
				return err
			}

		}

		// quick approve
		if trxFMF.TrxCaDecision.Decision == constant.DB_DECISION_APR {
			logInfo = trxFMF.TrxCaDecision

			if err := tx.Create(&trxFMF.TrxCaDecision).Error; err != nil {
				return err
			}

			trxAgreement := entity.TrxAgreement{
				ProspectID:         data.Transaction.ProspectID,
				CheckingStatus:     constant.ACTIVITY_UNPROCESS,
				ContractStatus:     "0",
				AF:                 data.Apk.AF,
				MobilePhone:        data.CustomerPersonal.MobilePhone,
				CustomerIDKreditmu: constant.LOB_NEW_KMB,
			}

			logInfo = trxAgreement

			if err := tx.Create(&trxAgreement).Error; err != nil {
				return err
			}

		}

		for i := 0; i < len(details); i++ {
			// skip prescreening unpr
			if details[i].SourceDecision != constant.PRESCREENING || details[i].Activity != constant.ACTIVITY_UNPROCESS {
				detail := entity.TrxDetail{
					ProspectID:     details[i].ProspectID,
					StatusProcess:  details[i].StatusProcess,
					Activity:       details[i].Activity,
					Decision:       details[i].Decision,
					RuleCode:       details[i].RuleCode,
					SourceDecision: details[i].SourceDecision,
					NextStep:       details[i].NextStep,
					Info:           details[i].Info,
					Reason:         details[i].Reason,
					CreatedBy:      constant.SYSTEM_CREATED,
				}

				logInfo = detail

				if err := tx.Create(&detail).Error; err != nil {
					return err
				}
			}
		}

		last := details[len(details)-1]

		status := entity.TrxStatus{
			ProspectID:     last.ProspectID,
			StatusProcess:  last.StatusProcess,
			Activity:       last.Activity,
			Decision:       last.Decision,
			RuleCode:       last.RuleCode,
			SourceDecision: last.SourceDecision,
			NextStep:       last.NextStep,
			Reason:         reason,
		}

		logInfo = status

		if rowsAffected := tx.Model(&status).Where("ProspectID = ?", status.ProspectID).Updates(status).RowsAffected; rowsAffected == 0 {
			if err := tx.Create(&status).Error; err != nil {
				return err
			}
		}

		return nil

	})

	if newErr != nil {
		newErr = errors.New(fmt.Sprintf("%s - %s - %s", constant.ERROR_UPSTREAM, newErr.Error(), logInfo))
	}

	return
}

func (r repoHandler) SaveTrxJourney(prospectID string, request interface{}) (err error) {

	requestByte, _ := json.Marshal(request)
	payload := string(utils.SafeEncoding(requestByte))

	trxJourney := entity.TrxJourney{
		ProspectID: prospectID,
	}

	asRunes := []rune(payload)
	if len(asRunes) > 7900 {
		trxJourney.Request = string(asRunes[:7900])
		trxJourney.Request2 = string(asRunes[7900:])
	} else {
		trxJourney.Request = payload
	}

	if err = r.newKmbDB.Model(&entity.TrxJourney{}).Create(&trxJourney).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetTrxJourney(prospectID string) (trxJourney entity.TrxJourney, err error) {

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT ProspectID, request, request2 from trx_journey with (nolock) where ProspectID = '%s'`, prospectID)).Scan(&trxJourney).Error; err != nil {
		return
	}

	if trxJourney.Request2 != nil {
		trxJourney.Request += trxJourney.Request2.(string)
	}

	return
}

func (r repoHandler) SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error) {

	headerByte, _ := json.Marshal(header)
	requestByte, _ := json.Marshal(request)
	responseByte, _ := json.Marshal(response)

	if err = r.logsDB.Model(&entity.LogOrchestrator{}).Create(&entity.LogOrchestrator{
		ID:           requestID,
		ProspectID:   prospectID,
		Owner:        "LOS-KMB",
		Header:       string(headerByte),
		Url:          path,
		Method:       method,
		RequestData:  string(utils.SafeEncoding(requestByte)),
		ResponseData: string(utils.SafeEncoding(responseByte)),
	}).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetLogOrchestrator(prospectID string) (logOrchestrator entity.LogOrchestrator, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT TOP 1 ProspectID, request_data from log_orchestrators with (nolock) where ProspectID = '%s' AND url = '/api/v3/kmb/consume/journey' ORDER BY created_at DESC", prospectID)).Scan(&logOrchestrator).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetEncryptedValue(idNumber string, legalName string, motherName string) (encrypted entity.Encrypted, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName, SCP.dbo.ENC_B64('SEC','%s') AS SurgateMotherName, SCP.dbo.ENC_B64('SEC','%s') AS IDNumber`,
		legalName, motherName, idNumber)).Scan(&encrypted).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) ScanWgOff(query string) (data entity.ScanInstallmentAmount, err error) {
	if err = r.wgOffDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanKmbOff(query string) (data entity.ScanInstallmentAmount, err error) {
	if err = r.kmbOffDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanKmobOff(query string) (data entity.ScanInstallmentAmount, err error) {

	if err = r.losDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanWgOnl(query string) (data entity.ScanInstallmentAmount, err error) {

	if err = r.losDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) GetMinimalIncomePMK(branchID string, statusKonsumen string) (responseIncomePMK entity.MappingIncomePMK, err error) {
	if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)).Scan(&responseIncomePMK).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM mapping_income_pmk WITH (nolock) WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, constant.DEFAULT_BRANCH_ID, statusKonsumen)).Scan(&responseIncomePMK).Error; err != nil {
				return
			}
		}
	}
	return
}

func (r repoHandler) GetLatestBannedRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error) {

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT * FROM dupcheck_rejection_nokanosin WHERE NoRangka = '%s' AND IsBanned = 1 ORDER BY created_at DESC LIMIT 1", noRangka)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetLatestRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT * FROM dupcheck_rejection_nokanosin WHERE NoRangka = '%s' AND CAST(created_at as DATE) = '%s' ORDER BY created_at DESC LIMIT 1", noRangka, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetAllReject(idNumber string) (data []entity.DupcheckRejectionPMK, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT di.BranchID, di.ProspectID, di.IDNumber, di.DtmUpd, dr.reject_pmk, dr.reject_dsr, fi.final_approval, fi.dtm_final_approval FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) = '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) ORDER BY di.DtmUpd ASC", idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetHistoryRejectAttempt(idNumber string) (data []entity.DupcheckRejectionPMK, err error) {

	searchRange := time.Now().AddDate(0, 0, -30)
	searchRangeString := searchRange.Format("2006-01-02")

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT x.* FROM (SELECT COUNT(*) AS reject_attempt, DATE(fi.dtm_final_approval) as date FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND (dr.reject_pmk IS NOT NULL OR dr.reject_dsr IS NOT NULL) AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) >= '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) GROUP BY date) x ORDER BY x.date DESC LIMIT 1", idNumber, searchRangeString)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetCheckingRejectAttempt(idNumber, blackListDate string) (data entity.DupcheckRejectionPMK, err error) {

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT x.* FROM (SELECT COUNT(*) AS reject_attempt, DATE(fi.dtm_final_approval) as date FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) = '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) GROUP BY date) x ORDER BY x.date DESC", idNumber, blackListDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SaveDataNoka(data entity.DupcheckRejectionNokaNosin) (err error) {
	data.CreatedAt = time.Now()
	if err = r.kmbOffDB.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveDataApiLog(data entity.TrxApiLog) (err error) {
	data.Timestamps = time.Now()
	if err = r.kmbOffDB.Create(data).Error; err != nil {
		return
	}

	return
}

func (r *repoHandler) GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig, err error) {
	if err := r.losDB.Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s' AND lob = '%s' AND [key]= '%s' AND is_active = 1", groupName, lob, key)).Scan(&appConfig).Error; err != nil {
		return appConfig, err
	}

	return appConfig, err
}

func (r *repoHandler) SaveVerificationFaceCompare(data entity.VerificationFaceCompare) error {
	if err := r.losDB.Create(&data).Error; err != nil {
		return fmt.Errorf("save verification face compare error: %w", err)
	}
	return nil
}

func (r repoHandler) GetEncB64(myString string) (encryptedString entity.EncryptedString, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS my_string`, myString)).Scan(&encryptedString).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetCurrentTrxWithRejectDSR(idNumber string) (data entity.TrxStatus, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT TOP 1 ts.* FROM trx_status ts WITH (nolock) LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID
	WHERE ts.decision = 'REJ' AND ts.source_decision = 'DSR' AND tcp.IDNumber = '%s' AND CAST(ts.created_at as DATE) = '%s'`, idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetBannedPMKDSR(idNumber string) (data entity.TrxBannedPMKDSR, err error) {

	date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT * FROM trx_banned_pmk_dsr WITH (nolock) WHERE IDNumber = '%s' AND CAST(created_at as DATE) >= '%s'`, idNumber, date)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetTrxReject(idNumber string, config response.LockSystemConfig) (data []entity.TrxLockSystem, err error) {

	qRangeJourney := fmt.Sprintf(`DECLARE @date_range DATE = (SELECT TOP 1 CAST(DATEADD(DAY, -%d, ts.created_at) as DATE) as date_range
			FROM trx_status ts with (nolock) 
			LEFT JOIN trx_customer_personal tcp with (nolock) ON ts.ProspectID = tcp.ProspectID
			WHERE ts.decision = 'REJ' 
			AND tcp.IDNumber = '%s'
			ORDER BY ts.created_at DESC)`, config.Data.LockRejectCheck, idNumber)

	qRangeFiltering := fmt.Sprintf(`DECLARE @date_range_f DATE = (SELECT TOP 1 CAST(DATEADD(DAY, -%d, tf.created_at) as DATE) as date_range_f
			FROM trx_filtering tf with (nolock) 
			WHERE tf.next_process = 0 
			AND tf.id_number = '%s'
			ORDER BY tf.created_at DESC)`, config.Data.LockRejectCheck, idNumber)

	qRejectJourney := fmt.Sprintf(`SELECT TOP %d CAST(DATEADD(DAY, %d, ts.created_at) as DATE) as unban_date, 
			ts.created_at, ts.ProspectID, tcp.IDNumber, ts.decision, ts.reason
			FROM trx_status ts with (nolock) 
			LEFT JOIN trx_customer_personal tcp with (nolock) ON ts.ProspectID = tcp.ProspectID
			WHERE ts.decision = 'REJ' 
			AND tcp.IDNumber = '%s'
			AND ts.created_at >= '%s'
			AND CAST(ts.created_at as DATE) >= @date_range AND @date_range <= CAST(ts.created_at as DATE) 
			AND CAST(ts.created_at as DATE) >= CAST(DATEADD(DAY, -%d, GETDATE()) as DATE)`, config.Data.LockRejectAttempt, config.Data.LockRejectBan+1, idNumber, config.Data.LockStartDate, config.Data.LockRejectBan)

	qRejectFiltering := fmt.Sprintf(`SELECT TOP %d CAST(DATEADD(DAY, %d, tf.created_at) as DATE) as unban_date,
			tf.created_at, tf.prospect_id as ProspectID, tf.id_number as IDNumber, 
			CASE 
				when tf.next_process = 0 then 'REJ'
			END decision, tf.reason
			FROM trx_filtering tf with (nolock)
			WHERE tf.next_process = 0 
			AND tf.id_number = '%s'
			AND tf.created_at >= '%s'
			AND CAST(tf.created_at as DATE) >= @date_range_f AND @date_range_f <= CAST(tf.created_at as DATE) 
			AND CAST(tf.created_at as DATE) >= CAST(DATEADD(DAY, -%d, GETDATE()) as DATE)`, config.Data.LockRejectAttempt, config.Data.LockRejectBan+1, idNumber, config.Data.LockStartDate, config.Data.LockRejectBan)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`%s %s %s UNION ALL %s ORDER BY created_at DESC`, qRangeJourney, qRangeFiltering, qRejectJourney, qRejectFiltering)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetTrxCancel(idNumber string, config response.LockSystemConfig) (data []entity.TrxLockSystem, err error) {

	if err = r.newKmbDB.Raw(fmt.Sprintf(`DECLARE @date_range DATE = (SELECT TOP 1 CAST(DATEADD(DAY, -%d, ts.created_at) as DATE) as date_range
			FROM trx_status ts with (nolock) 
			LEFT JOIN trx_customer_personal tcp with (nolock) ON ts.ProspectID = tcp.ProspectID
			WHERE ts.decision = 'CAN' 
			AND tcp.IDNumber = '%s'
			ORDER BY ts.created_at DESC)
			SELECT TOP %d CAST(DATEADD(DAY, %d, ts.created_at) as DATE) as unban_date, 
			ts.created_at, ts.ProspectID, tcp.IDNumber, ts.decision, ts.reason
			FROM trx_status ts with (nolock) 
			LEFT JOIN trx_customer_personal tcp with (nolock) ON ts.ProspectID = tcp.ProspectID
			WHERE ts.decision = 'CAN' 
			AND tcp.IDNumber = '%s'
			AND ts.created_at >= '%s'
			AND CAST(ts.created_at as DATE) >= @date_range AND @date_range <= CAST(ts.created_at as DATE) 
			AND CAST(ts.created_at as DATE) >= CAST(DATEADD(DAY, -%d, GETDATE()) as DATE)
			ORDER BY ts.created_at DESC`, config.Data.LockCancelCheck, idNumber, config.Data.LockCancelAttempt, config.Data.LockCancelBan+1, idNumber, config.Data.LockStartDate, config.Data.LockCancelBan)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) SaveTrxLockSystem(trxLockSystem entity.TrxLockSystem) (err error) {

	if err = r.newKmbDB.Model(&entity.TrxLockSystem{}).Create(&trxLockSystem).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetTrxLockSystem(idNumber string, chassisNumber string, engineNumber string) (data entity.TrxLockSystem, bannedType string, err error) {
	query1 := "SELECT TOP 1 * FROM trx_lock_system tls WHERE unban_date > CAST(GETDATE() as DATE) AND IDNumber = ? ORDER BY unban_date DESC"

	if err = r.newKmbDB.Raw(query1, idNumber).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			return
		}
	}

	if data.ProspectID != "" {
		data.Reason = strings.TrimPrefix(data.Reason, "Asset ")
		bannedType = constant.BANNED_TYPE_NIK
		return
	}

	if chassisNumber != "" || engineNumber != "" {
		query2 := "SELECT TOP 1 * FROM trx_lock_system tls WHERE unban_date > CAST(GETDATE() as DATE) AND "
		args := []interface{}{}

		if chassisNumber != "" {
			query2 += "chassis_number = ?"
			args = append(args, chassisNumber)
		}

		if engineNumber != "" {
			if chassisNumber != "" {
				query2 += " OR "
			}
			query2 += "engine_number = ?"
			args = append(args, engineNumber)
		}

		query2 += " ORDER BY unban_date DESC"

		if err = r.newKmbDB.Raw(query2, args...).Scan(&data).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				err = nil
			}
			return
		}

		if data.ProspectID != "" {
			bannedType = constant.BANNED_TYPE_ASSET
			if !strings.HasPrefix(data.Reason, "Asset ") {
				data.Reason = "Asset " + data.Reason
			}
		}
	}

	return
}

func (r repoHandler) GetCurrentTrxWithReject(idNumber string) (data entity.TrxReject, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT 
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

func (r repoHandler) GetBannedChassisNumber(chassisNumber string) (data entity.TrxBannedChassisNumber, err error) {

	date := time.Now().AddDate(0, 0, -30).Format(constant.FORMAT_DATE)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT * FROM trx_banned_chassis_number WITH (nolock) WHERE chassis_number = '%s' AND CAST(created_at as DATE) >= '%s'`, chassisNumber, date)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetCurrentTrxWithRejectChassisNumber(chassisNumber string) (data []entity.RejectChassisNumber, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT  
		tcp.ProspectID,
		SCP.dbo.DEC_B64('SEC',tcp.IDNumber) AS IDNumber,
		SCP.dbo.DEC_B64('SEC',tcp.LegalName) AS LegalName,
		SCP.dbo.DEC_B64('SEC',tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		tcp.Gender,
		tcp.MaritalStatus,
		tcp.NumOfDependence,
		tcp.StaySinceYear,
		tcp.StaySinceMonth,
		tcp.HomeStatus,
		tca1.LegalZipCode,
		tca2.CompanyZipCode,
		tce.ProfessionID,
		tce.MonthlyFixedIncome,
		tce.EmploymentSinceYear,
		tce.EmploymentSinceMonth,
		ti.engine_number,
		ti.chassis_number,
		ti.bpkb_name,
		ti.manufacture_year,
		ta.NTF,
		ta.OTR,
		ta.Tenor
		FROM trx_status ts 
		LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID
		INNER JOIN (
			SELECT ProspectID, ZipCode AS LegalZipCode
			FROM trx_customer_address WITH (nolock)
			WHERE "Type" = 'LEGAL'
		) tca1 ON ts.ProspectID = tca1.ProspectID
		INNER JOIN (
			SELECT ProspectID, ZipCode AS CompanyZipCode
			FROM trx_customer_address WITH (nolock)
			WHERE "Type" = 'COMPANY'
		) tca2 ON ts.ProspectID = tca2.ProspectID
		LEFT JOIN trx_customer_employment tce WITH (nolock) ON ts.ProspectID = tce.ProspectID
		LEFT JOIN trx_item ti WITH (nolock) ON ts.ProspectID = ti.ProspectID
		LEFT JOIN trx_apk ta WITH (nolock) ON ts.ProspectID = ta.ProspectID
		WHERE ts.decision = 'REJ' AND ts.source_decision = 'NKA' 
		AND ti.chassis_number = '%s' AND CAST(ts.created_at as DATE) = '%s'`, chassisNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetRecalculate(prospectID string) (getRecalculate entity.GetRecalculate, err error) {
	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT ta.Tenor, ta.ProductOfferingID, ta.product_offering_desc, ta.NTF, ta.DPAmount, ta.percent_dp, ta.InstallmentAmount, ta.AdminFee, ta.AF,
	fidusia_fee, ta.interest_rate, ta.interest_amount, ta.NTFAkumulasi, ta.loan_amount, ta.LifeInsuranceFee, ta.AssetInsuranceFee, ta.provision_fee, 
	( CAST(ISNULL(ta2.InstallmentAmountFMF, 0) AS NUMERIC(17,2)) +
	CAST(ISNULL(ta2.InstallmentAmountSpouseFMF, 0) AS NUMERIC(17,2)) +
	CAST(ISNULL(ta2.InstallmentAmountOther, 0) AS NUMERIC(17,2)) +
	CAST(ISNULL(ta2.InstallmentAmountOtherSpouse, 0) AS NUMERIC(17,2)) -
	CAST(ISNULL(ta2.InstallmentTopup, 0) AS NUMERIC(17,2)) ) as TotalInstallmentFMF,
	(CAST(ISNULL(tce.MonthlyFixedIncome, 0) AS NUMERIC(17,2)) +
	CAST(ISNULL(tce.MonthlyVariableIncome, 0) AS NUMERIC(17,2)) +
	CAST(ISNULL(tce.SpouseIncome, 0) AS NUMERIC(17,2)) ) as TotalIncome,
	ta2.DSRFMF,ta2.DSRPBK,ta2.TotalDSR
	FROM trx_apk ta WITH (nolock) 
	LEFT JOIN trx_akkk ta2 WITH (nolock) ON ta2.ProspectID = ta.ProspectID
	LEFT JOIN trx_customer_employment tce WITH (nolock) ON tce.ProspectID = ta.ProspectID 
	WHERE ta.ProspectID = '%s'`, prospectID)).Scan(&getRecalculate).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) SaveRecalculate(beforeRecalculate entity.TrxRecalculate, afterRecalculate entity.TrxRecalculate) (err error) {

	var logInfo interface{}
	err = r.newKmbDB.Transaction(func(tx *gorm.DB) error {

		// update trx_recalculate
		logInfo = beforeRecalculate
		result := tx.Model(&beforeRecalculate).Where("ProspectID = ?", beforeRecalculate.ProspectID).Updates(beforeRecalculate)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found...
			err = errors.New(constant.RECORD_NOT_FOUND)
			return err
		}

		// update trx_apk
		TrxApk := entity.TrxApk{
			ProductOfferingID:   afterRecalculate.ProductOfferingID,
			ProductOfferingDesc: afterRecalculate.ProductOfferingDesc,
			Tenor:               afterRecalculate.Tenor,
			LoanAmount:          afterRecalculate.LoanAmount,
			AF:                  afterRecalculate.AF,
			InstallmentAmount:   afterRecalculate.InstallmentAmount,
			DPAmount:            afterRecalculate.DPAmount,
			PercentDP:           afterRecalculate.PercentDP,
			AdminFee:            afterRecalculate.AdminFee,
			ProvisionFee:        afterRecalculate.ProvisionFee,
			FidusiaFee:          afterRecalculate.FidusiaFee,
			AssetInsuranceFee:   afterRecalculate.AssetInsuranceFee,
			LifeInsuranceFee:    afterRecalculate.LifeInsuranceFee,
			InsuranceAmount:     afterRecalculate.LifeInsuranceFee + afterRecalculate.AssetInsuranceFee,
			NTF:                 afterRecalculate.NTF,
			NTFAkumulasi:        afterRecalculate.NTFAkumulasi,
			InterestRate:        afterRecalculate.InterestRate,
			InterestAmount:      afterRecalculate.InterestAmount,
		}
		logInfo = TrxApk
		result = tx.Model(&entity.TrxApk{}).Where("ProspectID = ?", afterRecalculate.ProspectID).Updates(TrxApk)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found...
			err = errors.New(constant.RECORD_NOT_FOUND)
			return err
		}

		// update trx_akkk
		TrxAkkk := entity.TrxAkkk{
			DSRFMF:   afterRecalculate.DSRFMF,
			TotalDSR: afterRecalculate.TotalDSR,
		}
		logInfo = TrxAkkk
		result = tx.Model(&entity.TrxAkkk{}).Where("ProspectID = ?", afterRecalculate.ProspectID).Updates(TrxAkkk)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found...
			err = errors.New(constant.RECORD_NOT_FOUND)
			return err
		}

		// get limit
		var limit entity.MappingLimitApprovalScheme
		logInfo = limit
		if err = tx.Raw("SELECT [alias] FROM m_limit_approval_scheme WITH (nolock) WHERE ? between coverage_ntf_start AND coverage_ntf_end", afterRecalculate.NTFAkumulasi).Scan(&limit).Error; err != nil {
			return err
		}

		// update trx_ca_decision
		TrxCaDecision := entity.TrxCaDecision{
			FinalApproval: limit.Alias,
		}
		logInfo = TrxCaDecision
		result = tx.Model(&entity.TrxCaDecision{}).Where("ProspectID = ?", afterRecalculate.ProspectID).Updates(TrxCaDecision)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found...
			err = errors.New(constant.RECORD_NOT_FOUND)
			return err
		}

		// update trx_status
		TrxStatus := entity.TrxStatus{
			SourceDecision: limit.Alias,
		}
		logInfo = TrxStatus
		result = tx.Model(&entity.TrxStatus{}).Where("ProspectID = ?", afterRecalculate.ProspectID).Updates(TrxStatus)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found...
			err = errors.New(constant.RECORD_NOT_FOUND)
			return err
		}

		// update trx_history_approval_scheme
		TrxHistoryApprovalScheme := entity.TrxHistoryApprovalScheme{
			NextStep: limit.Alias,
		}
		logInfo = TrxHistoryApprovalScheme
		result = tx.Model(&entity.TrxHistoryApprovalScheme{}).Where("ProspectID = ? AND decision = 'SDP'", afterRecalculate.ProspectID).Updates(TrxHistoryApprovalScheme)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found...
			err = errors.New(constant.RECORD_NOT_FOUND)
			return err
		}

		return nil
	})

	if err != nil {
		err = errors.New(fmt.Sprintf("%s - %s - %s", constant.ERROR_UPSTREAM, err.Error(), logInfo))
	}

	return
}

func (r repoHandler) SaveToStaging(prospectID string) (newErr error) {

	var (
		master     entity.TrxMaster
		addresses  []entity.CustomerAddress
		apk        entity.TrxApk
		item       entity.TrxItem
		personal   entity.CustomerPersonal
		emcon      entity.CustomerEmcon
		employment entity.CustomerEmployment
		omset      []entity.CustomerOmset
		spouse     entity.CustomerSpouse
	)

	if newErr := r.newKmbDB.Raw("SELECT * FROM trx_master WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&master).Error; newErr != nil {
		return newErr
	}

	if newErr := r.newKmbDB.Raw("SELECT * FROM trx_apk WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&apk).Error; newErr != nil {
		return newErr
	}

	if newErr := r.newKmbDB.Raw("SELECT * FROM trx_customer_address WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&addresses).Error; newErr != nil {
		return newErr
	}

	if newErr := r.newKmbDB.Raw("SELECT * FROM trx_item WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&item).Error; newErr != nil {
		return newErr
	}

	if newErr := r.newKmbDB.Raw(`SELECT scp.dbo.DEC_B64('SEC', a.IDNumber) AS IDNumber, scp.dbo.DEC_B64('SEC', a.LegalName) AS LegalName, 
	scp.dbo.DEC_B64('SEC', a.FullName) AS FullName, scp.dbo.DEC_B64('SEC', a.BirthPlace) AS BirthPlace,
	scp.dbo.DEC_B64('SEC', a.SurgateMotherName) AS SurgateMotherName, scp.dbo.DEC_B64('SEC', a.MobilePhone) AS MobilePhone, 
	scp.dbo.DEC_B64('SEC', a.Email) AS Email, * FROM trx_customer_personal a WITH (nolock) WHERE a.ProspectID = ?`, prospectID).Scan(&personal).Error; newErr != nil {
		return newErr
	}

	if newErr := r.newKmbDB.Raw("SELECT * FROM trx_customer_emcon WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&emcon).Error; newErr != nil {
		return newErr
	}

	if newErr := r.newKmbDB.Raw("SELECT * FROM trx_customer_employment WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&employment).Error; newErr != nil {
		return newErr
	}

	if employment.ProfessionID == constant.PROFESSION_ID_WRST || employment.ProfessionID == constant.PROFESSION_ID_PRO {
		if newErr := r.newKmbDB.Raw("SELECT * FROM trx_customer_omset WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&omset).Error; newErr != nil {
			return newErr
		}
	}

	if personal.MaritalStatus == constant.MARRIED {
		if newErr := r.newKmbDB.Raw("SELECT * FROM trx_customer_spouse WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&spouse).Error; newErr != nil {
			return newErr
		}
	}

	var legal, emergency, residence, company, location, owner, mailing entity.CustomerAddress

	for _, address := range addresses {
		switch address.Type {
		case "LEGAL":
			legal = address
		case "EMERGENCY":
			emergency = address
		case "RESIDENCE":
			residence = address
		case "COMPANY":
			company = address
		case "OWNER":
			owner = address
		case "MAILING":
			mailing = address
		case "LOCATION":
			location = address
		}
	}

	var decrypted entity.Encrypted

	if err := r.newKmbDB.Raw(fmt.Sprintf(`SELECT scp.dbo.DEC_B64('SEC', '%s') AS ResidenceAddress, scp.dbo.DEC_B64('SEC','%s') AS LegalAddress,
		scp.dbo.DEC_B64('SEC', '%s') AS CompanyAddress, scp.dbo.DEC_B64('SEC', '%s') AS EmergencyAddress,
		scp.dbo.DEC_B64('SEC', '%s') AS OwnerAddress, scp.dbo.DEC_B64('SEC','%s') AS MailingAddress,
		scp.dbo.DEC_B64('SEC', '%s') AS LocationAddress`, residence.Address, legal.Address, company.Address,
		emergency.Address, owner.Address, mailing.Address, mailing.Address)).Scan(&decrypted).Error; err != nil {
		return err
	}

	var (
		month1, month2, month3, year1, year2, year3 int
		omset1, omset2, omset3                      float64
	)

	if employment.ProfessionID == constant.PROFESSION_ID_WRST || employment.ProfessionID == constant.PROFESSION_ID_PRO {
		for i := 0; i < len(omset); i++ {
			switch i {
			case 0:
				month1, _ = strconv.Atoi(omset[i].MonthlyOmsetMonth)
				year1, _ = strconv.Atoi(omset[i].MonthlyOmsetYear)
				omset1 = omset[i].MonthlyOmset
			case 1:
				month2, _ = strconv.Atoi(omset[i].MonthlyOmsetMonth)
				year2, _ = strconv.Atoi(omset[i].MonthlyOmsetYear)
				omset2 = omset[i].MonthlyOmset
			case 2:
				month3, _ = strconv.Atoi(omset[i].MonthlyOmsetMonth)
				year3, _ = strconv.Atoi(omset[i].MonthlyOmsetYear)
				omset3 = omset[i].MonthlyOmset
			}
		}
	} else {
		month1, _ = strconv.Atoi(time.Now().AddDate(0, -1, 0).Format("01"))
		month2, _ = strconv.Atoi(time.Now().AddDate(0, -2, 0).Format("01"))
		month3, _ = strconv.Atoi(time.Now().AddDate(0, -3, 0).Format("01"))
		year1, _ = strconv.Atoi(time.Now().AddDate(0, -1, 0).Format("2006"))
		year2, _ = strconv.Atoi(time.Now().AddDate(0, -2, 0).Format("2006"))
		year3, _ = strconv.Atoi(time.Now().AddDate(0, -3, 0).Format("2006"))
		omset1 = employment.MonthlyFixedIncome
		omset2 = employment.MonthlyFixedIncome
		omset3 = employment.MonthlyFixedIncome
	}

	newErr = r.stagingDB.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(&entity.STG_MAIN{
			BranchID:      master.BranchID,
			ProspectID:    master.ProspectID,
			IsRCA:         0,
			IsPV:          0,
			DataType:      personal.DataType,
			CreatedDate:   time.Now(),
			Status:        personal.Status,
			ApplicationID: nil,
			CustomerID:    nil,
			UpdatedDate:   time.Now(),
			UsrCrt:        constant.LOS_CREATED,
			DtmCrt:        time.Now(),
			DtmUpd:        time.Now(),
		}).Error; err != nil {
			return err
		}

		haveInsurance := "0"

		if apk.LifeInsuranceFee > 0 {
			haveInsurance = "1"
		}

		if err := tx.Create(&entity.STG_GEN_APP{
			BranchID:            master.BranchID,
			ProspectID:          master.ProspectID,
			ProductID:           apk.ProductID,
			ProductOfferingID:   apk.ProductOfferingID,
			Tenor:               *apk.Tenor,
			NumOfAssetUnit:      item.Qty,
			POS:                 item.Pos,
			WayOfPayment:        apk.WayOfPayment,
			ApplicationSource:   master.ApplicationSource,
			AgreementDate:       time.Now(),
			InsAssetInsuredBy:   item.InsAssetInsuredBy,
			InsAssetPaidBy:      apk.InsAssetPaidBy,
			InsAssetPeriod:      apk.InsAssetPeriod,
			IsLifeInsurance:     haveInsurance,
			AOID:                apk.AoID,
			MailingAddress:      decrypted.MailingAddress,
			MailingRT:           mailing.Rt,
			MailingRW:           mailing.Rw,
			MailingKelurahan:    mailing.Kelurahan,
			MailingKecamatan:    mailing.Kecamatan,
			MailingCity:         mailing.City,
			MailingZipCode:      mailing.ZipCode,
			MailingAreaPhone1:   mailing.AreaPhone,
			MailingPhone1:       mailing.Phone,
			UsrCrt:              constant.LOS_CREATED,
			DtmCrt:              time.Now(),
			ApplicationPriority: constant.RG_PRIORITY,
		}).Error; err != nil {
			return err
		}

		manufacture, _ := strconv.Atoi(item.ManufactureYear)

		if err := tx.Create(&entity.STG_GEN_ASD{
			BranchID:          master.BranchID,
			ProspectID:        master.ProspectID,
			SuplierID:         item.SupplierID,
			AssetCode:         item.AssetCode,
			OTRPrice:          apk.OTR,
			DPAmount:          apk.DPAmount,
			ManufacturingYear: manufacture,
			NoRangka:          item.ChassisNo,
			NoMesin:           item.EngineNo,
			UsedNew:           item.Condition,
			AssetUsage:        item.AssetUsage,
			CC:                item.Cc,
			Color:             item.Color,
			LicensePlate:      item.LicensePlate,
			OwnerAsset:        item.OwnerAsset,
			OwnerKTP:          item.OwnerKTP,
			OwnerAddress:      decrypted.OwnerAddress,
			OwnerRT:           owner.Rt,
			OwnerRW:           owner.Rw,
			OwnerKelurahan:    owner.Kelurahan,
			OwnerKecamatan:    owner.Kecamatan,
			OwnerCity:         owner.City,
			OwnerZipCode:      owner.ZipCode,
			LocationAddress:   decrypted.LocationAddress,
			LocationKelurahan: location.Kelurahan,
			LocationKecamatan: location.Kecamatan,
			LocationCity:      location.City,
			LocationZipCode:   location.ZipCode,
			Region:            item.Region,
			SalesmanID:        apk.SalesmanID,
			UsrCrt:            constant.LOS_CREATED,
			DtmCrt:            time.Now(),
			PromoToCust:       0,
			TaxDate:           item.TaxDate,
			STNKExpiredDate:   item.STNKExpiredDate,
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.STG_GEN_COM{
			BranchID:              master.BranchID,
			ProspectID:            master.ProspectID,
			SupplierBankAccountID: apk.SupplierBankAccountID,
			UsrCrt:                constant.LOS_CREATED,
			DtmCrt:                time.Now(),
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.STG_GEN_FIN{
			BranchID:          master.BranchID,
			ProspectID:        master.ProspectID,
			FirstInstallment:  apk.FirstInstallment,
			AdminFee:          apk.AdminFee,
			OtherFee:          apk.OtherFee,
			SurveyFee:         apk.SurveyFee,
			FiduciaFee:        apk.FidusiaFee,
			IsFiduciaCovered:  apk.IsFidusiaCovered,
			ProvisionFee:      apk.ProvisionFee,
			InstallmentAmount: apk.InstallmentAmount,
			EffectiveRate:     apk.EffectiveRate,
			CommisionSubsidy:  apk.CommisionSubsidi,
			StampDutyFee:      apk.StampDutyFee,
			UsrCrt:            constant.LOS_CREATED,
			DtmCrt:            time.Now(),
		}).Error; err != nil {
			return err
		}

		var expiredDateIns interface{}

		if item.InsAssetInsuredBy == constant.CU {
			expiredDateIns = time.Now().AddDate(5, 0, 0)

		}

		if err := tx.Create(&entity.STG_GEN_INS_H{
			BranchID:                master.BranchID,
			ProspectID:              master.ProspectID,
			ApplicationType:         constant.NG_APPLICATION_TYPE,
			AmountCoverage:          item.AssetInsuranceAmountCoverage,
			InsAssetInsuredBy:       item.InsAssetInsuredBy,
			InsuranceCoyBranchID:    item.InsuranceCoyBranchID,
			PremiumAmountToCustomer: apk.AssetInsuranceFee,
			CoverageType:            item.CoverageType,
			ExpiredDate:             expiredDateIns,
			UsrCrt:                  constant.LOS_CREATED,
			DtmCrt:                  time.Now(),
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.STG_GEN_INS_D{
			BranchID:      master.BranchID,
			ProspectID:    master.ProspectID,
			InsSequenceNo: 1,
			SRCC:          "0",
			Flood:         "0",
			EQVET:         "0",
			CoverageType:  item.CoverageType,
			UsrCrt:        constant.LOS_CREATED,
			DtmCrt:        time.Now(),
		}).Error; err != nil {
			return err
		}

		if err := tx.Create(&entity.STG_GEN_LFI{
			BranchID:                 master.BranchID,
			ProspectID:               master.ProspectID,
			LifeInsuranceCoyBranchID: apk.LifeInsuranceCoyBranchID,
			AmountCoverage:           apk.LifeInsuranceAmountCoverage,
			PremiumAmountToCustomer:  apk.LifeInsuranceFee,
			PaymentMethod:            apk.PaymentMethod,
			UsrCrt:                   constant.LOS_CREATED,
			DtmCrt:                   time.Now(),
		}).Error; err != nil {
			return err
		}

		expiredDate := time.Now().AddDate(4, 0, 0)
		if personal.ExpiredDate != nil {
			expiredDate = personal.ExpiredDate.(time.Time)
			if expiredDate.Unix() <= time.Now().Unix() {
				expiredDate = time.Now().AddDate(4, 0, 0)
			}
		}
		npwp := "000000000000000"
		if personal.PersonalNPWP != nil {
			personalNPWP := *personal.PersonalNPWP
			if personalNPWP != "" {
				npwp = personalNPWP
			}
		}

		if err := tx.Create(&entity.STG_CUST_H{
			BranchID:             master.BranchID,
			ProspectID:           master.ProspectID,
			LegalName:            personal.LegalName,
			FullName:             personal.FullName,
			PersonalCustomerType: personal.PersonalCustomerType,
			IDType:               personal.IDType,
			IDNumber:             personal.IDNumber,
			ExpiredDate:          expiredDate,
			Gender:               personal.Gender,
			BirthPlace:           personal.BirthPlace,
			BirthDate:            personal.BirthDate,
			PersonalNPWP:         npwp,
			SurgateMotherName:    personal.SurgateMotherName,
			UsrCrt:               constant.LOS_CREATED,
			DtmCrt:               time.Now(),
		}).Error; err != nil {
			return err
		}

		issueDate := time.Now().AddDate(0, 0, 3)
		if personal.IDTypeIssueDate != nil {
			issueDate = personal.IDTypeIssueDate.(time.Time)
			if issueDate.Unix() <= time.Now().Unix() {
				issueDate = time.Now().AddDate(0, 0, 3)
			}
		}

		stayMonth, _ := strconv.Atoi(personal.StaySinceMonth)
		stayYear, _ := strconv.Atoi(personal.StaySinceYear)
		employmentYear, _ := strconv.Atoi(employment.EmploymentSinceYear)

		if err := tx.Create(&entity.STG_CUST_D{
			BranchID:                         master.BranchID,
			ProspectID:                       master.ProspectID,
			IDTypeIssuedDate:                 issueDate,
			Education:                        personal.Education,
			Nationality:                      personal.Nationality,
			WNACountry:                       personal.WNACountry,
			HomeStatus:                       personal.HomeStatus,
			HomeLocation:                     personal.HomeLocation,
			StaySinceMonth:                   stayMonth,
			StaySinceYear:                    stayYear,
			Religion:                         personal.Religion,
			MaritalStatus:                    personal.MaritalStatus,
			NumOfDependence:                  personal.NumOfDependence,
			MobilePhone:                      personal.MobilePhone,
			Email:                            personal.Email,
			CustomerGroup:                    personal.CustomerGroup,
			KKNo:                             personal.KKNo,
			LegalAddress:                     decrypted.LegalAddress,
			LegalRT:                          legal.Rt,
			LegalRW:                          legal.Rw,
			LegalKelurahan:                   legal.Kelurahan,
			LegalKecamatan:                   legal.Kecamatan,
			LegalCity:                        legal.City,
			LegalZipCode:                     legal.ZipCode,
			LegalAreaPhone1:                  legal.AreaPhone,
			LegalPhone1:                      legal.Phone,
			ResidenceAddress:                 decrypted.ResidenceAddress,
			ResidenceRT:                      residence.Rt,
			ResidenceRW:                      residence.Rw,
			ResidenceCity:                    residence.City,
			ResidenceKelurahan:               residence.Kelurahan,
			ResidenceKecamatan:               residence.Kecamatan,
			ResidenceZipCode:                 residence.ZipCode,
			ResidenceAreaPhone1:              residence.AreaPhone,
			ResidencePhone1:                  residence.Phone,
			EmergencyContactAddress:          decrypted.EmergencyAddress,
			EmergencyContactRT:               emergency.Rt,
			EmergencyContactRW:               emergency.Rw,
			EmergencyContactKelurahan:        emergency.Kelurahan,
			EmergencyContactKecamatan:        emergency.Kecamatan,
			EmergencyContactCity:             emergency.City,
			EmergencyContactZipCode:          emergency.ZipCode,
			EmergencyContactHomePhoneArea1:   emergency.AreaPhone,
			EmergencyContactHomePhone1:       emergency.Phone,
			EmergencyContactOfficePhoneArea1: personal.EmergencyOfficeAreaPhone,
			EmergencyContactOfficePhone1:     personal.EmergencyOfficePhone,
			EmergencyContactMobilePhone:      emcon.MobilePhone,
			EmergencyContactName:             emcon.Name,
			EmergencyContactRelationship:     emcon.Relationship,
			ProfessionID:                     employment.ProfessionID,
			JobType:                          employment.JobType,
			JobPosition:                      employment.JobPosition,
			CompanyName:                      employment.CompanyName,
			IndustryTypeID:                   employment.IndustryTypeID,
			CompanyAddress:                   decrypted.CompanyAddress,
			CompanyRT:                        company.Rt,
			CompanyRW:                        company.Rw,
			CompanyCity:                      company.City,
			CompanyKelurahan:                 company.Kelurahan,
			CompanyKecamatan:                 company.Kecamatan,
			CompanyZipCode:                   company.ZipCode,
			CompanyAreaPhone1:                company.AreaPhone,
			CompanyPhone1:                    company.Phone,
			EmploymentSinceYear:              employmentYear,
			MonthlyFixedIncome:               employment.MonthlyFixedIncome,
			MonthlyVariableIncome:            employment.MonthlyVariableIncome,
			LivingCostAmount:                 personal.LivingCostAmount,
			MonthlyOmset1Year:                year1,
			MonthlyOmset1Month:               month1,
			MonthlyOmset1Omset:               omset1,
			MonthlyOmset2Year:                year2,
			MonthlyOmset2Month:               month2,
			MonthlyOmset2Omset:               omset2,
			MonthlyOmset3Year:                year3,
			MonthlyOmset3Month:               month3,
			MonthlyOmset3Omset:               omset3,
			Counterpart:                      personal.Counterpart,
			DebtBusinessScale:                personal.DebtBusinessScale,
			DebtGroup:                        personal.DebtGroup,
			IsAffiliateWithPP:                personal.IsAffiliateWithPP,
			AgreetoAcceptOtherOffering:       personal.AgreetoAcceptOtherOffering,
			SpouseIncome:                     employment.SpouseIncome,
			BankID:                           personal.BankID,
			AccountNo:                        personal.AccountNo,
			AccountName:                      personal.AccountName,
			UsrCrt:                           constant.LOS_CREATED,
			DtmCrt:                           time.Now(),
		}).Error; err != nil {
			return err
		}

		if personal.MaritalStatus == constant.MARRIED {
			if err := tx.Create(&entity.STG_CUST_FAM{
				BranchID:       master.BranchID,
				ProspectID:     master.ProspectID,
				SeqNo:          1,
				Name:           spouse.LegalName,
				IDNumber:       spouse.IDNumber,
				BirthDate:      spouse.BirthDate,
				FamilyRelation: constant.CODE_SPOUSE,
				UsrCrt:         constant.LOS_CREATED,
				DtmCrt:         time.Now(),
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})

	return
}

func (r repoHandler) GetMappingVehicleAge(vehicleAge int, cluster string, bpkbNameType, tenor int, resultPefindo string, af float64) (data entity.MappingVehicleAge, err error) {

	query := `SELECT TOP 1 * FROM m_mapping_vehicle_age WHERE vehicle_age_start <= ? AND vehicle_age_end >= ? AND cluster LIKE ? AND bpkb_name_type = ? AND tenor_start <= ? AND tenor_end >= ? AND result_pbk LIKE ? AND af_start < ? AND af_end >= ?`

	if err = r.newKmbDB.Raw(query, vehicleAge, vehicleAge, fmt.Sprintf("%%%s%%", cluster), bpkbNameType, tenor, tenor, fmt.Sprintf("%%%s%%", resultPefindo), af, af).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return
}

func (r repoHandler) GetMappingNegativeCustomer(req response.NegativeCustomer) (data entity.MappingNegativeCustomer, err error) {

	query := `SELECT TOP 1 * FROM m_mapping_negative_customer WHERE is_active = ? AND bad_type = ? AND is_blacklist = ? AND is_highrisk = ?`

	if err = r.newKmbDB.Raw(query, req.IsActive, req.BadType, req.IsBlacklist, req.IsHighrisk).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}
	return
}

func (r repoHandler) MasterMappingIncomeMaxDSR(totalIncome float64) (data entity.MasterMappingIncomeMaxDSR, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.losDB.BeginTx(ctx, &x)
	defer db.Commit()

	query := `SELECT TOP 1 * FROM kmb_mapping_income_dsr WHERE total_income_start <= ? AND (total_income_end >= ? OR total_income_end IS NULL)`
	if err = r.losDB.Raw(query, totalIncome, totalIncome).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}

	return
}

func (r repoHandler) MasterMappingDeviasiDSR(totalIncome float64) (data entity.MasterMappingDeviasiDSR, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmbDB.BeginTx(ctx, &x)
	defer db.Commit()

	query := `SELECT TOP 1 * FROM m_mapping_deviasi_dsr WHERE total_income_start <= ? AND (total_income_end >= ? OR total_income_end IS NULL)`
	if err = r.newKmbDB.Raw(query, totalIncome, totalIncome).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
	}

	return
}

func (r repoHandler) GetBranchDeviasi(BranchID string, customerStatus string, NTF float64) (data entity.MappingBranchDeviasi, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmbDB.BeginTx(ctx, &x)
	defer func() {
		if r := recover(); r != nil {
			db.Rollback()
			panic(r)
		} else if err != nil {
			db.Rollback()
		} else {
			db.Commit()
		}
	}()

	query := "SELECT * FROM dbo.m_branch_deviasi WITH (nolock) WHERE BranchID = ? AND is_active = 1"
	args := []interface{}{BranchID}

	if customerStatus != constant.STATUS_KONSUMEN_RO && customerStatus != constant.STATUS_KONSUMEN_AO {
		query += " AND balance_amount >= ? AND balance_account >= 1"
		args = append(args, NTF)
	}

	if err = db.Raw(query, args...).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) ScanTrxPrinciple(prospectID string) (count int, err error) {

	var (
		trxEmcon []entity.TrxPrincipleEmergencyContact
	)

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT ProspectID FROM trx_principle_emergency_contact WITH (nolock) WHERE ProspectID = '%s'", prospectID)).Scan(&trxEmcon).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	count = len(trxEmcon)

	return
}

func (r repoHandler) GetPrincipleStepOne(prospectID string) (data entity.TrxPrincipleStepOne, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_one WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmbDB.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetPrincipleStepTwo(prospectID string) (data entity.TrxPrincipleStepTwo, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_two WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmbDB.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetPrincipleStepThree(prospectID string) (data entity.TrxPrincipleStepThree, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_step_three WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmbDB.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetPrincipleEmergencyContact(prospectID string) (data entity.TrxPrincipleEmergencyContact, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_principle_emergency_contact WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmbDB.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) ScanTrxKPM(prospectID string) (count int, err error) {

	var (
		trxKPM []entity.TrxKPM
	)

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT ProspectID FROM trx_kpm WITH (nolock) WHERE ProspectID = '%s'", prospectID)).Scan(&trxKPM).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	count = len(trxKPM)

	return
}

func (r repoHandler) GetTrxKPM(prospectID string) (data entity.TrxKPM, err error) {

	query := fmt.Sprintf(`SELECT TOP 1 * FROM trx_kpm WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC`, prospectID)

	if err = r.newKmbDB.Raw(query).Scan(&data).Error; err != nil {
		return
	}

	var decrypted entity.Encrypted

	if err = r.newKmbDB.Raw(fmt.Sprintf(`SELECT scp.dbo.DEC_B64('SEC', '%s') AS LegalName, scp.dbo.DEC_B64('SEC','%s') AS SurgateMotherName,
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

func (r repoHandler) GetTrxKPMStatus(prospectID string) (data entity.TrxKPMStatus, err error) {

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT TOP 1 tks.* FROM trx_kpm_status tks WITH (nolock) WHERE tks.ProspectID = '%s' ORDER BY tks.created_at DESC", prospectID)).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateTrxKPMDecision(id string, prospectID string, decision string) (err error) {

	return r.newKmbDB.Transaction(func(tx *gorm.DB) error {

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
