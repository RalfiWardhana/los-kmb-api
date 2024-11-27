package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/cms/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
)

type repoHandler struct {
	NewKmb    *gorm.DB
	core      *gorm.DB
	confins   *gorm.DB
	losDB     *gorm.DB
	KpLosLogs *gorm.DB
}

func NewRepository(core, confins, NewKmb, kpLos, KpLosLogs *gorm.DB) interfaces.Repository {
	return &repoHandler{
		core:      core,
		confins:   confins,
		NewKmb:    NewKmb,
		losDB:     kpLos,
		KpLosLogs: KpLosLogs,
	}
}

func (r repoHandler) GetAFMobilePhone(prospectID string) (data entity.AFMobilePhone, err error) {

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT AF, SCP.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone, OTR, DPAmount FROM trx_apk apk WITH (nolock) INNER JOIN trx_customer_personal tcp WITH (nolock) ON apk.ProspectID = tcp.ProspectID WHERE apk.ProspectID = '%s'`, prospectID)).Scan(&data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetRegionBranch(userId string) (data []entity.RegionBranch, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT region_name, branch_member FROM region_branch a WITH (nolock)
	INNER JOIN region b WITH (nolock) ON a.region = b.region_id WHERE region IN 
	(	SELECT value 
		FROM region_user ru WITH (nolock)
		cross apply STRING_SPLIT(REPLACE(REPLACE(REPLACE(region,'[',''),']',''), '"',''),',')
		WHERE ru.user_id = '%s' 
	)
	AND b.lob_id='%s'`, userId, constant.LOB_ID_NEW_KMB)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return
	}

	return
}

func (r repoHandler) GetSpIndustryTypeMaster() (data []entity.SpIndustryTypeMaster, err error) {

	if err = r.core.Raw("exec[spIndustryTypeMaster] '01/01/2007'").Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		err = fmt.Errorf(constant.RECORD_NOT_FOUND)
		return
	}

	return
}

func (r repoHandler) GetReasonPrescreening(req request.ReqReasonPrescreening, pagination interface{}) (reason []entity.ReasonMessage, rowTotal int, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	var (
		filter         string
		filterPaginate string
	)

	if req.ReasonID != "" {
		arrReason := strings.Split(req.ReasonID, ",")
		var reason string
		for i, val := range arrReason {
			reason = fmt.Sprintf("%s'%s'", reason, val)
			if i < len(arrReason)-1 {
				reason = fmt.Sprintf("%s,", reason)
			}
		}

		filter = fmt.Sprintf("WHERE ReasonID NOT IN (%s)", reason)
	}

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`
		SELECT
		COUNT(tt.ReasonID) AS totalRow
		FROM
		(SELECT ReasonID FROM m_reason_message WITH (nolock)) AS tt %s`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT tt.* FROM (SELECT Code, ReasonID, ReasonMessage FROM m_reason_message WITH (nolock)) AS tt %s ORDER BY tt.ReasonID asc %s`, filter, filterPaginate)).Scan(&reason).Error; err != nil {
		return
	}

	if len(reason) == 0 {
		return reason, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) GetCancelReason(pagination interface{}) (reason []entity.CancelReason, rowTotal int, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	var (
		filterPaginate string
	)

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(`
		SELECT
		COUNT(tt.id_cancel_reason) AS totalRow
		FROM
		(SELECT * FROM m_cancel_reason with (nolock) WHERE show = '1') AS tt`).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT * FROM m_cancel_reason with (nolock) WHERE show = '1' ORDER BY id_cancel_reason ASC %s`, filterPaginate)).Scan(&reason).Error; err != nil {
		return
	}

	if len(reason) == 0 {
		return reason, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) GetApprovalReason(req request.ReqApprovalReason, pagination interface{}) (reason []entity.ApprovalReason, rowTotal int, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	var (
		filterPaginate string
		filter         string
	)

	if req.Type != "" {
		filter = fmt.Sprintf(" AND [Type] = '%s'", req.Type)
	}

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.confins.Raw(fmt.Sprintf(`
		SELECT
		COUNT(tt.id) AS totalRow
		FROM
		(SELECT CONCAT(ReasonID, '|', Type, '|', Description) AS 'id', Description AS 'value', [Type] FROM tblApprovalReason WHERE IsActive = 'True' %s) AS tt`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.confins.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.confins.Raw(fmt.Sprintf(`SELECT CONCAT(ReasonID, '|', Type, '|', Description) AS 'id', Description AS 'value', [Type] FROM tblApprovalReason WHERE IsActive = 'True' %s ORDER BY ReasonID ASC %s`, filter, filterPaginate)).Scan(&reason).Error; err != nil {
		return
	}

	if len(reason) == 0 {
		return reason, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) GetCustomerPhoto(prospectID string) (photo []entity.DataPhoto, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT tcp.photo_id, CASE WHEN lpi.Name IS NULL THEN 'LAINNYA' ELSE lpi.Name END AS label, tcp.url FROM trx_customer_photo tcp WITH (nolock) LEFT JOIN m_label_photo_inquiry lpi ON lpi.LabelPhotoID = tcp.photo_id WHERE ProspectID = ?", prospectID).Scan(&photo).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetSurveyorData(prospectID string) (surveyor []entity.TrxSurveyor, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT destination, request_date, assign_date, surveyor_name, result_date, status, surveyor_note FROM trx_surveyor WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&surveyor).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetInquiryPrescreening(req request.ReqInquiryPrescreening, pagination interface{}) (data []entity.InquiryPrescreening, rowTotal int, err error) {

	var (
		filter         string
		filterBranch   string
		filterPaginate string
		getRegion      []entity.RegionBranch
		encrypted      entity.EncryptString
	)

	rangeDays := os.Getenv("DEFAULT_RANGE_DAYS")

	if req.MultiBranch == "1" {
		getRegion, _ = r.GetRegionBranch(req.UserID)

		if len(getRegion) > 0 {
			extractBranchIDUser := ""
			userAllRegion := false
			for _, value := range getRegion {
				if strings.ToUpper(value.RegionName) == constant.REGION_ALL {
					userAllRegion = true
					break
				} else if value.BranchMember != "" {
					branch := strings.Trim(strings.ReplaceAll(value.BranchMember, `"`, `'`), "'")
					replace := strings.ReplaceAll(branch, `[`, ``)
					branchMember := strings.ReplaceAll(replace, `]`, ``)
					extractBranchIDUser += branchMember
					if value != getRegion[len(getRegion)-1] {
						extractBranchIDUser += ","
					}
				}
			}
			if userAllRegion {
				filterBranch += ""
			} else {
				filterBranch += "WHERE tm.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += ""
			if req.BranchID != "999" {
				filterBranch += "WHERE tm.BranchID = '" + req.BranchID + "'"
			}
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	if req.Search != "" {
		encrypted, _ = r.EncryptString(req.Search)
	}

	filter = utils.GenerateFilter(req.Search, encrypted.Encrypt, filterBranch, rangeDays, "")

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`WITH 
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'LEGAL'
		) cal ON tm.ProspectID = cal.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'RESIDENCE'
		) car ON tm.ProspectID = car.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'COMPANY'
		) cac ON tm.ProspectID = cac.ProspectID
		INNER JOIN (
			SELECT
			ProspectID,
			Address,
			RT,
			RW,
			Kelurahan,
			Kecamatan,
			ZipCode,
			City,
			Phone,
			AreaPhone
			FROM
			trx_customer_address WITH (nolock)
			WHERE
			"Type" = 'EMERGENCY'
		) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
			SELECT
			[key],
			value
			FROM
			app_config ap WITH (nolock)
			WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		 %s) AS tt`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`WITH 
	cte_app_config_mn AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'MonthName'
	),
	cte_app_config_pr AS (
		SELECT
			[key],
			value
		FROM
			app_config ap WITH (nolock)
		WHERE
			group_name = 'ProfessionID'
	)
	SELECT tt.* FROM (
	SELECT
	tm.ProspectID,
	cb.BranchName,
	cb.BranchID,
	tia.info AS CMORecommend,
	tst.activity,
	tst.source_decision,
	tps.decision,
	tps.reason,
	tps.created_by AS DecisionBy,
	tps.decision_by AS DecisionName,
	tps.created_at AS DecisionAt,
	CASE
	  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
	  ELSE 'NE'
	END AS incoming_source,
	tf.customer_status,
	tm.created_at,
	tm.order_at,
	scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
	scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
	scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
	tcp.BirthDate,
	scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
	CASE
	  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
	  WHEN tcp.Gender = 'F' THEN 'Perempuan'
	END AS 'Gender',
	scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
	CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
	cal.Kelurahan AS LegalKelurahan,
	cal.Kecamatan AS LegalKecamatan,
	cal.ZipCode AS LegalZipcode,
	cal.City AS LegalCity,
	scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
	scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
	edu.value AS Education,
	mst.value AS MaritalStatus,
	tcp.NumOfDependence,
	scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
	CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
	car.Kelurahan AS ResidenceKelurahan,
	car.Kecamatan AS ResidenceKecamatan,
	car.ZipCode AS ResidenceZipcode,
	car.City AS ResidenceCity,
	hst.value AS HomeStatus,
	mn.value AS StaySinceMonth,
	tcp.StaySinceYear,
	ta.ProductOfferingID,
	ta.dealer,
	ta.LifeInsuranceFee,
	ta.AssetInsuranceFee,
	'KMB MOTOR' AS AssetType,
	ti.asset_description,
	ti.manufacture_year,
	ti.color,
	chassis_number,
	engine_number,
	CASE
		WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
	END AS bpkb_name,
	ti.owner_asset,
	ti.license_plate,
	interest_rate,
	Tenor AS InstallmentPeriod,
	OTR,
	DPAmount,
	AF AS FinanceAmount,
	interest_amount,
	insurance_amount,
	AdminFee,
	provision_fee,
	NTF,
	NTFAkumulasi,
	(NTF + interest_amount) AS Total,
	InstallmentAmount AS MonthlyInstallment,
	FirstInstallment,
	pr.value AS ProfessionID,
	jt.value AS JobType,
	jb.value AS JobPosition,
	mn2.value AS EmploymentSinceMonth,
	tce.EmploymentSinceYear,
	tce.CompanyName,
	cac.AreaPhone AS CompanyAreaPhone,
	cac.Phone AS CompanyPhone,
	tcp.ExtCompanyPhone,
	scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
	CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
	cac.Kelurahan AS CompanyKelurahan,
	cac.Kecamatan AS CompanyKecamatan,
	car.ZipCode AS CompanyZipcode,
	car.City AS CompanyCity,
	tce.MonthlyFixedIncome,
	tce.MonthlyVariableIncome,
	tce.SpouseIncome,
	tcp.SourceOtherIncome,
	tcs.FullName AS SpouseLegalName,
	tcs.CompanyName AS SpouseCompanyName,
	tcs.CompanyPhone AS SpouseCompanyPhone,
	tcs.MobilePhone AS SpouseMobilePhone,
	tcs.IDNumber AS SpouseIDNumber,
	pr2.value AS SpouseProfession,
	em.Name AS EmconName,
	em.Relationship,
	em.MobilePhone AS EmconMobilePhone,
	scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
	CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
	cae.Kelurahan AS EmergencyKelurahan,
	cae.Kecamatan AS EmergencyKecamatan,
	cae.ZipCode AS EmergencyZipcode,
	cae.City AS EmergencyCity,
	cae.AreaPhone AS EmergencyAreaPhone,
	cae.Phone AS EmergencyPhone,
	tce.IndustryTypeID
  FROM
	trx_master tm WITH (nolock)
	INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
	INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
	INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
	INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
	INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
	INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
	INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
	INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'LEGAL'
	) cal ON tm.ProspectID = cal.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'RESIDENCE'
	) car ON tm.ProspectID = car.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'COMPANY'
	) cac ON tm.ProspectID = cac.ProspectID
	INNER JOIN (
	  SELECT
		ProspectID,
		Address,
		RT,
		RW,
		Kelurahan,
		Kecamatan,
		ZipCode,
		City,
		Phone,
		AreaPhone
	  FROM
		trx_customer_address WITH (nolock)
	  WHERE
		"Type" = 'EMERGENCY'
	) cae ON tm.ProspectID = cae.ProspectID
	INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
	LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
	LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'Education'
	) edu ON tcp.Education = edu.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'MaritalStatus'
	) mst ON tcp.MaritalStatus = mst.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'HomeStatus'
	) hst ON tcp.HomeStatus = hst.[key]
	LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
	LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobType'
	) jt ON tce.JobType = jt.[key]
	LEFT JOIN (
	  SELECT
		[key],
		value
	  FROM
		app_config ap WITH (nolock)
	  WHERE
		group_name = 'JobPosition'
	) jb ON tce.JobPosition = jb.[key]
	LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
	LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key] %s) AS tt ORDER BY tt.created_at DESC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) GetTrxStatus(prospectID string) (status entity.TrxStatus, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT activity, decision, source_decision FROM trx_status WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&status).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetTrxEDD(prospectID string) (trxEDD entity.TrxEDD, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT * FROM trx_edd WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&trxEDD).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) SavePrescreening(prescreening entity.TrxPrescreening, detail entity.TrxDetail, status entity.TrxStatus) (err error) {

	prescreening.CreatedAt = time.Now()
	detail.CreatedAt = time.Now()
	status.CreatedAt = time.Now()

	return r.NewKmb.Transaction(func(tx *gorm.DB) error {

		// update trx_status
		result := tx.Model(&status).Where("ProspectID = ?", status.ProspectID).Updates(status)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found...
			if err = tx.Create(&status).Error; err != nil {
				return err
			}
		}

		// insert trx_details
		if err = tx.Create(&detail).Error; err != nil {
			return err
		}

		// if stop insert trx_akkk
		if status.Activity == constant.ACTIVITY_STOP {
			trxAkkk := entity.TrxAkkk{
				ProspectID: status.ProspectID,
			}
			if err = tx.Create(&trxAkkk).Error; err != nil {
				return err
			}
		}

		// insert trx_prescreening
		if err = tx.Create(&prescreening).Error; err != nil {
			return err
		}
		return nil
	})
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

func (r repoHandler) GetHistoryApproval(prospectID string) (history []entity.HistoryApproval, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(`SELECT
				thas.decision_by,
				thas.next_final_approval_flag,
				CASE
					WHEN thas.decision = 'APR' THEN 'Approve'
					WHEN thas.decision = 'REJ' THEN 'Reject'
					WHEN thas.decision = 'CAN' THEN 'Cancel'
					WHEN thas.decision = 'RTN' THEN 'Return'
					WHEN thas.decision = 'SDP' THEN 'Submit Perubahan Data Pembiayaan'
					ELSE '-'
				END AS decision,
				CASE
					WHEN thas.need_escalation = 1 THEN 'Yes'
					ELSE 'No'
				END AS need_escalation,
				thas.source_decision,
				CASE
					WHEN thas.next_step<>'' THEN thas.next_step
					ELSE '-'
				END AS next_step,
				CASE
					WHEN thas.note<>'' THEN thas.note
					ELSE '-'
				END AS note,
				thas.created_at,
				CASE
				  WHEN thas.source_decision = 'CRA' AND tcd.slik_result<>'' AND thas.decision<>'SDP' THEN tcd.slik_result
				  ELSE
				  '-'
				END AS slik_result
			FROM trx_history_approval_scheme thas WITH (nolock) LEFT JOIN trx_ca_decision tcd on thas.ProspectID = tcd.ProspectID WHERE thas.ProspectID = ? ORDER BY thas.created_at DESC`, prospectID).Scan(&history).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetAkkk(prospectID string) (data entity.Akkk, err error) {
	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT ts.ProspectID, 
		ta2.FinancePurpose,
		scp.dbo.DEC_B64('SEC',tcp.LegalName) as LegalName,
		scp.dbo.DEC_B64('SEC',tcp.IDNumber) as IDNumber,  
		tcp.PersonalNPWP,
		scp.dbo.DEC_B64('SEC',tcp.SurgateMotherName) as SurgateMotherName,
		tce.ProfessionID,
		tcp.CustomerStatus,
		ta.CustomerType,
		tcp.Gender,
		scp.dbo.DEC_B64('SEC',tcp.BirthPlace) as BirthPlace,
		tcp.BirthDate,
		tcp.Education,
		scp.dbo.DEC_B64('SEC',tcp.MobilePhone) as MobilePhone,
		scp.dbo.DEC_B64('SEC',tcp.Email) as Email,
		tcs.LegalName as SpouseLegalName,
		tcs.IDNumber as SpouseIDNumber,
		tcs.SurgateMotherName as SpouseSurgateMotherName,
		tcs.ProfessionID as SpouseProfessionID,
		ta.SpouseType,
		tcs.Gender as SpouseGender,
		tcs.BirthPlace as SpouseBirthPlace,
		tcs.BirthDate as SpouseBirthDate,
		tcs.MobilePhone as SpouseMobilePhone,
		tce2.VerificationWith,
		tce2.Relationship as EmconRelationship,
		tce2.EmconVerified,
		tca.Address,
		tce2.MobilePhone as EmconMobilePhone,
		tce2.VerifyBy,
		tce2.KnownCustomerAddress,
		tcp.StaySinceYear,
		tcp.StaySinceMonth,
		tce2.KnownCustomerJob,
		CASE tce.ProfessionID 
			WHEN 'WRST' THEN 'WIRASWASTA'
			WHEN 'PRO' THEN 'PROFESSIONAL'
			WHEN 'KRYSW' THEN 'KARYAWAN SWASTA'
			WHEN 'PNS' THEN 'PEGAWAI NEGERI SIPIL'
			WHEN 'ANG' THEN 'ANGKATAN'
			ELSE '-'
		END as Job,
		tce.EmploymentSinceYear,
		tce.EmploymentSinceMonth,
		tce.IndustryTypeID,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		CASE ti.bpkb_name
			WHEN 'K' THEN 'NAMA SAMA'
			WHEN 'P' THEN 'NAMA SAMA'
			WHEN 'KK' THEN 'NAMA BEDA'
			WHEN 'O' THEN 'NAMA BEDA'
			ELSE '-'
		END as BpkbName,
		tdb.plafon as Plafond,
		tdb.baki_debet_non_collateral as BakiDebet,
		tdb.fasilitas_aktif as FasilitasAktif,
		CASE
			WHEN tdb.kualitas_kredit_terburuk IS NOT NULL THEN CONCAT(tdb.kualitas_kredit_terburuk,' ',tdb.bulan_kualitas_terburuk)
			ELSE NULL
		END as ColTerburuk,
		tdb.baki_debet_kualitas_terburuk as BakiDebetTerburuk,
		CASE
			WHEN tdb.kualitas_kredit_terakhir IS NOT NULL THEN CONCAT(tdb.kualitas_kredit_terakhir,' ',tdb.bulan_kualitas_kredit_terakhir)
			ELSE NULL
		END as ColTerakhirAktif,
		tdb2.plafon as SpousePlafond,
		tdb2.baki_debet_non_collateral as SpouseBakiDebet,
		tdb2.fasilitas_aktif as SpouseFasilitasAktif,
		CASE
			WHEN tdb2.kualitas_kredit_terburuk IS NOT NULL THEN CONCAT(tdb2.kualitas_kredit_terburuk,' ',tdb2.bulan_kualitas_terburuk)
			ELSE NULL
		END as SpouseColTerburuk,
		tdb2.baki_debet_kualitas_terburuk as SpouseBakiDebetTerburuk,
		CASE
			WHEN tdb2.kualitas_kredit_terakhir IS NOT NULL THEN CONCAT(tdb2.kualitas_kredit_terakhir,' ',tdb2.bulan_kualitas_kredit_terakhir)
			ELSE NULL
		END as SpouseColTerakhirAktif,
		ta.ScsScore,
		ta.AgreementStatus,
		ta.TotalAgreementAktif,
		ta.MaxOVDAgreementAktif,
		ta.LastMaxOVDAgreement,
		tf.customer_segment,
		ta.LatestInstallment,
		ta2.NTFAkumulasi,
		(CAST(ISNULL(ta2.InstallmentAmount, 0) AS NUMERIC(17,2)) +
		CASE 
			WHEN ta.TotalDSR = ta.DSRFMF THEN 0
			WHEN ta.TotalDSR IS NULL THEN 0
			ELSE CAST(ISNULL(tf.total_installment_amount_biro, 0) AS NUMERIC(17,2))
		END +
		CAST(ISNULL(ta.InstallmentAmountFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountSpouseFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOther, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOtherSpouse, 0) AS NUMERIC(17,2)) -
		CAST(ISNULL(ta.InstallmentTopup, 0) AS NUMERIC(17,2)) ) as TotalInstallment,
		(CAST(ISNULL(tce.MonthlyFixedIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.MonthlyVariableIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.SpouseIncome, 0) AS NUMERIC(17,2)) ) as TotalIncome,
		CASE 
			WHEN ta.TotalDSR IS NULL THEN CAST(ISNULL(ta.DSRFMF, 0) AS NUMERIC(17,2)) + CAST(ISNULL(ta.DSRPBK, 0) AS NUMERIC(17,2))
			ELSE ta.TotalDSR
		END as TotalDSR,
		CASE
			WHEN ta.EkycSource IS NOT NULL THEN CONCAT(ta.EkycSource,' - ',ta.EkycReason)
			ELSE NULL
		END as EkycSource,
		ta.EkycSimiliarity,
		ta.EkycReason,
		CASE tia.info 
			WHEN '1' THEN 'APR'
			ELSE 'REJ'
		END cmo_decision,
		tia.name as cmo_name,
		tia.recom_date cmo_date,
		tcd.decision as ca_decision,
		tcd.note as ca_note,
		tcd.decision_by as ca_name,
		FORMAT(tcd.created_at,'yyyy-MM-dd') as ca_date,
		cbm.decision as cbm_decision,
		cbm.note as cbm_note,
		cbm.decision_by as cbm_name,
		FORMAT(cbm.created_at,'yyyy-MM-dd') as cbm_date,
		drm.decision as drm_decision,
		drm.note as drm_note,
		drm.decision_by as drm_name,
		FORMAT(drm.created_at,'yyyy-MM-dd') as drm_date,
		gmo.decision as gmo_decision,
		gmo.note as gmo_note,
		gmo.decision_by as gmo_name,
		FORMAT(gmo.created_at,'yyyy-MM-dd') as gmo_date
		FROM trx_status ts WITH (nolock)
		LEFT JOIN trx_customer_personal tcp WITH (nolock) ON ts.ProspectID = tcp.ProspectID 
		LEFT JOIN trx_customer_employment tce WITH (nolock) ON ts.ProspectID = tce.ProspectID
		LEFT JOIN trx_akkk ta WITH (nolock) ON ts.ProspectID = ta.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON ts.ProspectID = tcs.ProspectID
		LEFT JOIN trx_customer_emcon tce2 WITH (nolock) ON ts.ProspectID = tce2.ProspectID
		LEFT OUTER JOIN ( 
			SELECT scp.dbo.DEC_B64('SEC', Address) AS Address, ProspectID, Phone 
			FROM trx_customer_address WITH (nolock)
			WHERE Type = 'EMERGENCY' 
		) AS tca ON ts.ProspectID = tca.ProspectID 
		LEFT JOIN trx_item ti WITH (nolock) ON ts.ProspectID = ti.ProspectID
		LEFT OUTER JOIN ( 
			SELECT * FROM trx_detail_biro WITH (nolock)
			WHERE subject = 'CUSTOMER' 
		) AS tdb ON ts.ProspectID = tdb.prospect_id
		LEFT JOIN trx_filtering tf WITH (nolock) ON ts.ProspectID = tf.prospect_id 
		LEFT JOIN trx_apk ta2 WITH (nolock) ON ts.ProspectID = ta2.ProspectID
		LEFT OUTER JOIN ( 
			SELECT * FROM trx_detail_biro WITH (nolock)
			WHERE subject = 'SPOUSE' 
		) AS tdb2 ON ts.ProspectID = tdb2.prospect_id
		LEFT JOIN trx_ca_decision tcd WITH (nolock) ON ts.ProspectID = tcd.ProspectID
		LEFT JOIN trx_info_agent tia WITH (nolock) ON ts.ProspectID = tia.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas1 WITH (nolock)
			WHERE thas1.source_decision = 'CBM' AND thas1.created_at = (SELECT MAX(tha1.created_at) From trx_history_approval_scheme tha1 WHERE tha1.source_decision = thas1.source_decision AND tha1.ProspectID = thas1.ProspectID)
		) AS cbm ON ts.ProspectID = cbm.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas2 WITH (nolock)
			WHERE thas2.source_decision = 'DRM' AND thas2.created_at = (SELECT MAX(tha2.created_at) From trx_history_approval_scheme tha2 WHERE tha2.source_decision = thas2.source_decision AND tha2.ProspectID = thas2.ProspectID)
		) AS drm ON ts.ProspectID = drm.ProspectID
		LEFT JOIN ( 
			SELECT * FROM trx_history_approval_scheme thas3 WITH (nolock)
			WHERE thas3.source_decision = 'GMO' AND thas3.created_at = (SELECT MAX(tha3.created_at) From trx_history_approval_scheme tha3 WHERE tha3.source_decision = thas3.source_decision AND tha3.ProspectID = thas3.ProspectID)
		) AS gmo ON ts.ProspectID = gmo.ProspectID
		WHERE ts.ProspectID = '%s'`, prospectID)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SubmitNE(req request.MetricsNE, filtering request.Filtering, elaboreateLTV request.ElaborateLTV, journey request.Metrics) (err error) {
	err = r.NewKmb.Transaction(func(tx *gorm.DB) error {
		var encrypted entity.Encrypted

		if err := tx.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName,  SCP.dbo.ENC_B64('SEC','%s') AS IDNumber`,
			req.CustomerPersonal.LegalName, req.CustomerPersonal.IDNumber)).Scan(&encrypted).Error; err != nil {
			return err
		}

		PayloadNE, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(req)
		PayloadFiltering, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(filtering)
		PayloadLTV, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(elaboreateLTV)
		PayloadJourney, _ := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(journey)
		ne := entity.NewEntry{
			ProspectID:       req.Transaction.ProspectID,
			BranchID:         req.Transaction.BranchID,
			IDNumber:         encrypted.IDNumber,
			LegalName:        encrypted.LegalName,
			BirthDate:        req.CustomerPersonal.BirthDate,
			CreatedByID:      req.CreatedBy.CreatedByID,
			CreatedByName:    req.CreatedBy.CreatedByName,
			PayloadNE:        string(PayloadNE),
			PayloadFiltering: string(PayloadFiltering),
			PayloadLTV:       string(PayloadLTV),
			PayloadJourney:   string(PayloadJourney),
		}

		if err := tx.Create(&ne).Error; err != nil {
			return err
		}
		return nil
	})

	return
}

func (r repoHandler) GetInquiryNE(req request.ReqInquiryNE, pagination interface{}) (data []entity.InquiryDataNE, rowTotal int, err error) {

	var (
		filter         string
		filterBranch   string
		filterPaginate string
		getRegion      []entity.RegionBranch
		encrypted      entity.EncryptString
	)

	rangeDays := os.Getenv("DEFAULT_RANGE_DAYS")

	if req.MultiBranch == "1" {
		getRegion, _ = r.GetRegionBranch(req.UserID)

		if len(getRegion) > 0 {
			extractBranchIDUser := ""
			userAllRegion := false
			for _, value := range getRegion {
				if strings.ToUpper(value.RegionName) == constant.REGION_ALL {
					userAllRegion = true
					break
				} else if value.BranchMember != "" {
					branch := strings.Trim(strings.ReplaceAll(value.BranchMember, `"`, `'`), "'")
					replace := strings.ReplaceAll(branch, `[`, ``)
					branchMember := strings.ReplaceAll(replace, `]`, ``)
					extractBranchIDUser += branchMember
					if value != getRegion[len(getRegion)-1] {
						extractBranchIDUser += ","
					}
				}
			}
			if userAllRegion {
				filterBranch += ""
			} else {
				filterBranch += "WHERE tm.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += ""
			if req.BranchID != "999" {
				filterBranch += "WHERE tm.BranchID = '" + req.BranchID + "'"
			}
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	if req.Search != "" {
		encrypted, _ = r.EncryptString(req.Search)
	}

	filter = utils.GenerateFilter(req.Search, encrypted.Encrypt, filterBranch, rangeDays, "NE")

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(
			SELECT
			tm.BranchID,
			tm.ProspectID,
			tm.created_at,
			scp.dbo.DEC_B64('SEC', tm.IDNumber) AS IDNumber,
			scp.dbo.DEC_B64('SEC', tm.LegalName) AS LegalName
		FROM
		trx_new_entry tm WITH (nolock)
		 %s) AS tt`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT tt.* FROM (
	SELECT
	tm.ProspectID,
	tm.BranchID,
	tm.created_at,
	scp.dbo.DEC_B64('SEC', tm.IDNumber) AS IDNumber,
	scp.dbo.DEC_B64('SEC', tm.LegalName) AS LegalName,
	tm.BirthDate,
	CASE
	WHEN tf.next_process = 1 THEN 'PASS'
	WHEN tf.next_process = 0 THEN 'REJECT'
	ELSE NULL END AS ResultFiltering,
	tf.reason as Reason
  	FROM
	trx_new_entry tm WITH (nolock)
	LEFT JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
	 %s) AS tt ORDER BY tt.created_at DESC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) GetInquiryNEDetail(prospectID string) (data entity.NewEntry, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT payload_ne FROM trx_new_entry WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetInternalRecord(prospectID string) (record []entity.TrxInternalRecord, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT * FROM trx_internal_record WITH (nolock) WHERE ProspectID = ? ORDER BY created_at DESC", prospectID).Scan(&record).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetInquiryCa(req request.ReqInquiryCa, pagination interface{}) (data []entity.InquiryCa, rowTotal int, err error) {

	var (
		filter         string
		filterBranch   string
		filterPaginate string
		query          string
		getRegion      []entity.RegionBranch
		encrypted      entity.EncryptString
	)

	rangeDays := os.Getenv("DEFAULT_RANGE_DAYS")

	if req.MultiBranch == "1" {
		getRegion, _ = r.GetRegionBranch(req.UserID)

		if len(getRegion) > 0 {
			extractBranchIDUser := ""
			userAllRegion := false
			for _, value := range getRegion {
				if strings.ToUpper(value.RegionName) == constant.REGION_ALL {
					userAllRegion = true
					break
				} else if value.BranchMember != "" {
					branch := strings.Trim(strings.ReplaceAll(value.BranchMember, `"`, `'`), "'")
					replace := strings.ReplaceAll(branch, `[`, ``)
					branchMember := strings.ReplaceAll(replace, `]`, ``)
					extractBranchIDUser += branchMember
					if value != getRegion[len(getRegion)-1] {
						extractBranchIDUser += ","
					}
				}
			}
			if userAllRegion {
				filterBranch += ""
			} else {
				filterBranch += "WHERE tm.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += ""
			if req.BranchID != "999" {
				filterBranch += "WHERE tm.BranchID = '" + req.BranchID + "'"
			}
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	if req.Search != "" {
		encrypted, _ = r.EncryptString(req.Search)
	}

	filter = utils.GenerateFilter(req.Search, encrypted.Encrypt, filterBranch, rangeDays, "")

	// Filter By
	if req.Filter != "" {
		var (
			activity string
		)
		switch req.Filter {
		case constant.DECISION_APPROVE:
			query = fmt.Sprintf(" AND tst.decision = '%s' AND tst.status_process='%s'", constant.DB_DECISION_APR, constant.STATUS_FINAL)

		case constant.DECISION_REJECT:
			query = fmt.Sprintf(" AND tst.decision = '%s' AND tst.status_process='%s'", constant.DB_DECISION_REJECT, constant.STATUS_FINAL)

		case constant.DECISION_CANCEL:
			query = fmt.Sprintf(" AND tst.decision = '%s' AND tst.status_process='%s'", constant.DB_DECISION_CANCEL, constant.STATUS_FINAL)

		case constant.NEED_DECISION:
			activity = constant.ACTIVITY_UNPROCESS
			source := constant.DB_DECISION_CREDIT_ANALYST
			query = fmt.Sprintf(" AND tst.activity= '%s' AND tst.decision= '%s' AND tst.source_decision = '%s' AND (tcd.decision IS NULL OR (rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'%s'))", activity, constant.DB_DECISION_CREDIT_PROCESS, source, constant.STATUS_FINAL)

		case constant.SAVED_AS_DRAFT:
			if req.UserID != "" {
				query = fmt.Sprintf(" AND tdd.draft_created_by= '%s' ", req.UserID)
			}
		}
	}

	filter = filter + query

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		)
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
		 %s AND tst.source_decision<>'%s') AS tt`, filter, constant.PRESCREENING)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`WITH 
		cte_trx_ca_decision AS (
			SELECT
				ProspectID,
				decision,
				note,
				created_at,
				created_by
			FROM
				trx_ca_decision WITH (nolock)
		),
		cte_trx_draft_ca_decision AS (
			SELECT
				x.ProspectID,
				x.decision,
				x.slik_result,
				x.note,
				x.created_at,
				x.created_by,
				x.decision_by,
				x.pernyataan_1,
				x.pernyataan_2,
				x.pernyataan_3,
				x.pernyataan_4,
				x.pernyataan_5,
				x.pernyataan_6
			FROM
				trx_draft_ca_decision x WITH (nolock)
			WHERE
				x.created_at = (
					SELECT
						MAX(created_at)
					FROM
						trx_draft_ca_decision WITH (NOLOCK)
					WHERE
						ProspectID = x.ProspectID
				)
		),
		cte_trx_detail_biro AS (
			SELECT
				prospect_id, url_pdf_report AS BiroCustomerResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'CUSTOMER'
		),
		cte_trx_detail_biro2 AS (
			SELECT
				prospect_id, url_pdf_report AS BiroSpouseResult
			FROM
				trx_detail_biro WITH (nolock)
			WHERE
				[subject] = 'SPOUSE'
		),
		cte_trx_history_approval_scheme AS (
			SELECT
				ProspectID,
				decision AS decision_rtn
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'RTN'
		),
		cte_trx_history_approval_scheme_sdp AS (
			SELECT
				ProspectID,
				decision AS decision_sdp
			FROM
				trx_history_approval_scheme WITH (nolock)
			WHERE
				decision = 'SDP'
		),
		cte_app_config_mn AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'MonthName'
		),
		cte_app_config_pr AS (
			SELECT
				[key],
				value
			FROM
				app_config ap WITH (nolock)
			WHERE
				group_name = 'ProfessionID'
		)
		SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as decision_ca,
		tcd.created_by as decision_by_ca,
		tr.additional_dp,
		CASE
		  WHEN tcd.decision='APR' THEN 'APPROVE'
		  WHEN tcd.decision='REJ' THEN 'REJECT'
		  WHEN tcd.decision='CAN' THEN 'CANCEL'
		  ELSE tcd.decision
		END AS ca_decision,
		tcd.note AS ca_note,
		CASE
		  WHEN tcd.created_at IS NOT NULL
		  AND tfa.created_at IS NULL THEN FORMAT(tcd.created_at,'yyyy-MM-dd HH:mm:ss')
		  WHEN tfa.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE NULL
		END AS ActionDate,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tdd.decision AS draft_decision,
		tdd.slik_result AS draft_slik_result,
		tdd.note AS draft_note,
		tdd.created_at AS draft_created_at,
		tdd.created_by AS draft_created_by,
		tdd.decision_by AS draft_decision_by,
		tdd.pernyataan_1 AS draft_pernyataan_1,
		tdd.pernyataan_2 AS draft_pernyataan_2,
		tdd.pernyataan_3 AS draft_pernyataan_3,
		tdd.pernyataan_4 AS draft_pernyataan_4,
		tdd.pernyataan_5 AS draft_pernyataan_5,
		tdd.pernyataan_6 AS draft_pernyataan_6,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
		  WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
		  WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
		  WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
		  WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		ta.interest_rate,
		ta.Tenor AS InstallmentPeriod,
		OTR,
		ta.DPAmount,
		ta.AF AS FinanceAmount,
		ta.interest_amount,
		ta.insurance_amount,
		ta.AdminFee,
		ta.provision_fee,
		ta.NTF,
		ta.NTFAkumulasi,
		(ta.NTF + ta.interest_amount) AS Total,
		ta.InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb2.BiroSpouseResult,
		CASE
		 WHEN rtn.decision_rtn IS NOT NULL AND sdp.decision_sdp IS NULL AND tst.status_process<>'FIN' THEN 1
		 ELSE 0
		END AS ActionEditData,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  	FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_recalculate tr WITH (nolock) ON tm.ProspectID = tr.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN
			cte_trx_history_approval_scheme rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_history_approval_scheme_sdp sdp ON sdp.ProspectID = tm.ProspectID
		LEFT JOIN
			cte_trx_ca_decision tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN
			cte_trx_detail_biro tdb ON tm.ProspectID = tdb.prospect_id
		LEFT JOIN
			cte_trx_detail_biro2 tdb2 ON tm.ProspectID = tdb2.prospect_id
		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN cte_app_config_mn mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN cte_app_config_pr pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN cte_app_config_mn mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN cte_app_config_pr pr2 ON tcs.ProfessionID = pr2.[key]
		LEFT JOIN
			cte_trx_draft_ca_decision tdd ON tm.ProspectID = tdd.ProspectID
		 %s AND tst.source_decision<>'%s') AS tt ORDER BY tt.created_at DESC %s`, filter, constant.PRESCREENING, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) SaveDraftData(draft entity.TrxDraftCaDecision) (err error) {

	draft.CreatedAt = time.Now()

	return r.NewKmb.Transaction(func(tx *gorm.DB) error {

		var inInterface map[string]interface{}
		inrec, _ := json.Marshal(draft)
		json.Unmarshal(inrec, &inInterface)

		// update trx_status
		result := tx.Model(&draft).Where("ProspectID = ?", draft.ProspectID).Updates(inInterface)

		if err = result.Error; err != nil {
			return err
		}

		if result.RowsAffected == 0 {
			// record not found will be create draft
			if err = tx.Create(&draft).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r repoHandler) GetLimitApproval(ntf float64) (limit entity.MappingLimitApprovalScheme, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT [alias] FROM m_limit_approval_scheme WITH (nolock) WHERE ? between coverage_ntf_start AND coverage_ntf_end", ntf).Scan(&limit).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetLimitApprovalDeviasi(prospectID string) (limit entity.MappingLimitApprovalScheme, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(`SELECT
		CASE 
			WHEN final_approval IS NULL THEN 'CBM'
			ELSE final_approval
			END AS alias
		FROM trx_deviasi td
		LEFT JOIN trx_master tm ON td.ProspectID = tm.ProspectID 
		LEFT JOIN m_branch_deviasi mbd ON tm.BranchID = mbd.BranchID 
		WHERE td.ProspectID = ?`, prospectID).Scan(&limit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetHistoryProcess(prospectID string) (detail []entity.HistoryProcess, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(`SELECT
			CASE
			 WHEN td.source_decision = 'PSI' THEN 'PRE SCREENING'
			 WHEN td.source_decision IN ('TNR','PRJ','NIK','NKA','BLK','PMK') THEN 'DUPLICATION CHECKING'
			 WHEN td.source_decision = 'DCK' THEN 'DUPLICATION CHECKING'
			 WHEN td.source_decision = 'DCP'
			 OR td.source_decision = 'ARI'
			 OR td.source_decision = 'KTP' THEN 'EKYC'
			 WHEN td.source_decision = 'PBK' THEN 'PEFINDO'
			 WHEN td.source_decision = 'SCP' THEN 'SCOREPRO'
			 WHEN td.source_decision = 'DSR' THEN 'DSR'
			 WHEN td.source_decision = 'LTV' THEN 'LTV'
			 WHEN td.source_decision = 'DEV' THEN 'DEVIASI'
			 WHEN td.source_decision = 'CRA' THEN 'CREDIT ANALYSIS'
			 WHEN td.source_decision = 'NRC' THEN 'RECALCULATE PROCESS'
			 WHEN td.source_decision = 'CBM'
			  OR td.source_decision = 'DRM'
			  OR td.source_decision = 'GMO'
			  OR td.source_decision = 'COM'
			  OR td.source_decision = 'GMC'
			  OR td.source_decision = 'UCC' THEN 'CREDIT COMMITEE'
			 ELSE '-'
			END AS source_decision,
			CASE
			 WHEN td.source_decision = 'CRA' THEN 'CA'
			 WHEN td.source_decision = 'CBM' THEN 'BM'
			 WHEN td.source_decision = 'DRM' THEN 'RM'
			 WHEN td.source_decision = 'GMO' THEN 'GMO'
			 WHEN td.source_decision = 'COM' THEN 'COM'
			 WHEN td.source_decision = 'GMC' THEN 'GMC'
			 WHEN td.source_decision = 'UCC' THEN 'UCC'
			 ELSE td.source_decision
			END AS alias,
			CASE
			 WHEN td.source_decision = 'DEV' THEN '-'
			 WHEN td.decision = 'PAS' THEN 'PASS'
			 WHEN td.decision = 'REJ' THEN 'REJECT'
			 WHEN td.decision = 'CAN' THEN 'CANCEL'
			 WHEN td.decision = 'RTN' THEN 'RETURN'
			 WHEN td.decision = 'CPR' THEN 'CREDIT PROCESS'
			 ELSE '-'
			END AS decision,
			CASE
			 WHEN ap.reason IS NULL THEN td.reason 
			 ELSE ap.reason 
			END AS reason,
			FORMAT(td.created_at,'yyyy-MM-dd HH:mm:ss') as created_at,
			td.next_step
		FROM
			trx_details td WITH (nolock)
			LEFT JOIN app_rules ap ON ap.rule_code = td.rule_code
		WHERE td.ProspectID = ? AND (td.source_decision IN('PSI','DCK','DCP','ARI','KTP','PBK','SCP','DSR','CRA','CBM','DRM','GMO','COM','GMC','UCC','NRC','DEV') OR 
		(td.source_decision IN('TNR','PRJ','NIK','NKA','BLK','PMK','LTV') AND td.decision = 'REJ'))
		AND td.decision <> 'CTG' AND td.activity <> 'UNPR' ORDER BY td.created_at ASC`, prospectID).Scan(&detail).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetInquirySearch(req request.ReqSearchInquiry, pagination interface{}) (data []entity.InquirySearch, rowTotal int, err error) {

	var (
		filter         string
		filterPaginate string
		filterBranch   string
		query          string
		getRegion      []entity.RegionBranch
	)
	if req.MultiBranch == "1" {
		getRegion, _ = r.GetRegionBranch(req.UserID)

		if len(getRegion) > 0 {
			extractBranchIDUser := ""
			userAllRegion := false
			for _, value := range getRegion {
				if strings.ToUpper(value.RegionName) == constant.REGION_ALL {
					userAllRegion = true
					break
				} else if value.BranchMember != "" {
					branch := strings.Trim(strings.ReplaceAll(value.BranchMember, `"`, `'`), "'")
					replace := strings.ReplaceAll(branch, `[`, ``)
					branchMember := strings.ReplaceAll(replace, `]`, ``)
					extractBranchIDUser += branchMember
					if value != getRegion[len(getRegion)-1] {
						extractBranchIDUser += ","
					}
				}
			}
			if userAllRegion {
				filterBranch += ""
			} else {
				filterBranch += "WHERE tm.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += ""
			if req.BranchID != "999" {
				filterBranch += "WHERE tm.BranchID = '" + req.BranchID + "'"
			}
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	filter = filterBranch

	search := req.Search

	var qSearch string
	var regexpPpid = regexp.MustCompile(`SAL-|NE-`)
	var regexpIDNumber = regexp.MustCompile(`^[0-9]*$`)
	var regexpLegalName = regexp.MustCompile("^[a-zA-Z.,'` ]*$")

	if search != "" && regexpPpid.MatchString(search) {
		//query prospect id only
		qSearch = fmt.Sprintf("(tm.ProspectID = '%s')", search)
	} else if search != "" && regexpIDNumber.MatchString(search) {
		//query id number only
		encrypted, _ := r.EncryptString(search)
		qSearch = fmt.Sprintf("(tcp.IDNumber = '%s')", encrypted.Encrypt)
	} else if search != "" && regexpLegalName.MatchString(search) {
		//query legal name only
		encrypted, _ := r.EncryptString(search)
		qSearch = fmt.Sprintf("(tcp.LegalName = '%s')", encrypted.Encrypt)
	} else {
		//query default
		encrypted, _ := r.EncryptString(search)
		qSearch = fmt.Sprintf("(tm.ProspectID = '%s' OR tcp.IDNumber = '%s' OR tcp.LegalName = '%s')", search, encrypted.Encrypt, encrypted.Encrypt)
	}

	if search != "" {
		query = fmt.Sprintf("WHERE %s", qSearch)
	}

	if filter == "" {
		filter = query
	} else {
		filter = filterBranch + fmt.Sprintf(" AND %s", qSearch)
	}

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		%s) AS tt`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.decision,
		CASE
		  WHEN tst.status_process='FIN' AND tst.decision='APR' THEN 'Approve'
		  WHEN tst.status_process='FIN' AND tst.decision='REJ' THEN 'Reject'
		  WHEN tst.status_process='FIN' AND tst.decision='CAN' THEN 'Cancel'
		  ELSE '-'
		END AS FinalStatus,
		CASE
		  WHEN tps.ProspectID IS NOT NULL
		  AND tcd.decision IS NULL
		  AND tst.source_decision NOT IN('CBM','DRM','GMO','COM','GMC','UCC')
		  AND tst.status_process <> 'FIN' THEN 1
		  ELSE 0
		END AS ActionReturn,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP'
		  AND tst.decision='REJ' OR tst.decision='CAN' THEN 0
		  ELSE 1
		END AS ActionCancel,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tst.decision = 'CPR'
		  AND tst.source_decision = 'CRA'
		  AND tst.activity = 'UNPR'
		  AND tcd.decision IS NULL THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
			WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
			WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
			WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
			WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6
	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
		 %s) AS tt ORDER BY tt.created_at DESC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) ProcessTransaction(trxCaDecision entity.TrxCaDecision, trxHistoryApproval entity.TrxHistoryApprovalScheme, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail, isCancel bool, trxEdd entity.TrxEDD) (err error) {

	trxCaDecision.CreatedAt = time.Now()
	trxStatus.CreatedAt = time.Now()
	trxDetail.CreatedAt = time.Now()
	trxHistoryApproval.ID = uuid.New().String()
	trxHistoryApproval.CreatedAt = time.Now()

	return r.NewKmb.Transaction(func(tx *gorm.DB) error {

		// if trx not cancel/stop then check source_decision CRA
		if trxStatus.Activity != constant.ACTIVITY_STOP {
			// trx_status
			result := tx.Model(&trxStatus).Where("ProspectID = ? AND source_decision = 'CRA'", trxStatus.ProspectID).Updates(trxStatus)
			if result.Error != nil {
				err = result.Error
				return err
			}
			if result.RowsAffected == 0 {
				err = errors.New(constant.ERROR_ROWS_AFFECTED)
				return err
			}
		} else {
			// trx_status
			if err := tx.Model(&trxStatus).Where("ProspectID = ?", trxStatus.ProspectID).Updates(trxStatus).Error; err != nil {
				return err
			}
		}

		// trx_ca_decision
		result := tx.Model(&trxCaDecision).Where("ProspectID = ?", trxCaDecision.ProspectID).Updates(trxCaDecision)

		if result.RowsAffected == 0 {
			// record not found...
			if err = tx.Create(&trxCaDecision).Error; err != nil {
				return err
			}
		}

		// trx_details
		if err := tx.Create(&trxDetail).Error; err != nil {
			return err
		}

		// trx_history_approval_scheme
		if err := tx.Create(&trxHistoryApproval).Error; err != nil {
			return err
		}

		// trx edd
		if trxEdd != (entity.TrxEDD{}) {
			if err := tx.Model(&trxEdd).Where("ProspectID = ?", trxStatus.ProspectID).Updates(trxEdd).Error; err != nil {
				return err
			}
		}

		// trx_draft_ca_decision
		var draft entity.TrxDraftCaDecision
		if err := tx.Where("ProspectID = ?", trxCaDecision.ProspectID).Delete(&draft).Error; err != nil {
			return err
		}

		// if stop insert trx_akkk
		if trxStatus.Activity == constant.ACTIVITY_STOP {
			trxAkkk := entity.TrxAkkk{
				ProspectID: trxStatus.ProspectID,
			}
			tx.Create(&trxAkkk)
		}

		if isCancel {
			var resultCheckDeviation entity.ResultCheckDeviation

			selectQuery := `
                SELECT mbd.BranchID, ta.NTF, tf.customer_status, tfa.decision
				FROM trx_deviasi AS td WITH (nolock)
				LEFT JOIN trx_apk AS ta ON (td.ProspectID = ta.ProspectID)
				LEFT JOIN trx_master AS tm ON (td.ProspectID = tm.ProspectID)
				LEFT JOIN m_branch_deviasi AS mbd ON (tm.BranchID = mbd.BranchID)
				LEFT JOIN trx_filtering AS tf ON (td.ProspectID = tf.prospect_id)
				LEFT JOIN trx_final_approval AS tfa ON (td.ProspectID = tfa.ProspectID)
                WHERE ta.ProspectID = ? AND mbd.is_active = 1
            `
			if err = tx.Raw(selectQuery, trxCaDecision.ProspectID).Scan(&resultCheckDeviation).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}
			}

			if resultCheckDeviation.BranchID != "" && resultCheckDeviation.CustomerStatus != "" && resultCheckDeviation.CustomerStatus == constant.STATUS_KONSUMEN_NEW && resultCheckDeviation.Decision != nil {
				if decisionStr, ok := resultCheckDeviation.Decision.(string); ok && decisionStr == constant.DB_DECISION_APR {
					updateQuery := `
						UPDATE m_branch_deviasi
						SET booking_amount = booking_amount - ?,
							booking_account = booking_account - 1,
							balance_amount = (quota_amount - booking_amount) + ?,
							balance_account = (quota_account - booking_account) + 1
						WHERE BranchID = ? AND is_active = 1
					`
					if err = tx.Exec(updateQuery, resultCheckDeviation.NTF, resultCheckDeviation.NTF, resultCheckDeviation.BranchID).Error; err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func (r repoHandler) ProcessReturnOrder(prospectID string, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail) (err error) {

	trxStatus.CreatedAt = time.Now()
	trxDetail.CreatedAt = time.Now()

	return r.NewKmb.Transaction(func(tx *gorm.DB) error {

		// update trx_status
		if err := tx.Model(&trxStatus).Where("ProspectID = ?", prospectID).Updates(trxStatus).Error; err != nil {
			return err
		}

		// truncate the order from trx_details
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&trxDetail).Error; err != nil {
			return err
		}

		// truncate the order from trx_deviasi
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&entity.TrxDeviasi{}).Error; err != nil {
			return err
		}

		// truncate the order from trx_edd
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&entity.TrxEDD{}).Error; err != nil {
			return err
		}

		// truncate the trx_final_approval
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&entity.TrxFinalApproval{}).Error; err != nil {
			return err
		}

		// truncate the dbo.trx_agreements
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&entity.TrxAgreement{}).Error; err != nil {
			return err
		}

		// insert trx_details
		if err := tx.Create(&trxDetail).Error; err != nil {
			return err
		}

		// delete trx_prescreening
		var prescreening entity.TrxPrescreening
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&prescreening).Error; err != nil {
			return err
		}

		// delete trx_ca_decision
		var ca entity.TrxCaDecision
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&ca).Error; err != nil {
			return err
		}

		// delete trx_draft_ca_decision
		var draft entity.TrxDraftCaDecision
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&draft).Error; err != nil {
			return err
		}

		// delete trx_history_approval_scheme
		var history entity.TrxHistoryApprovalScheme
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&history).Error; err != nil {
			return err
		}

		// delete trx_akkk
		var akkk entity.TrxAkkk
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&akkk).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r repoHandler) ProcessRecalculateOrder(prospectID string, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail, trxHistoryApproval entity.TrxHistoryApprovalScheme) (err error) {

	trxStatus.CreatedAt = time.Now()
	trxDetail.CreatedAt = time.Now()
	trxHistoryApproval.ID = uuid.New().String()
	trxHistoryApproval.CreatedAt = time.Now()

	return r.NewKmb.Transaction(func(tx *gorm.DB) error {

		// update trx_status
		if err := tx.Model(&trxStatus).Where("ProspectID = ?", prospectID).Updates(trxStatus).Error; err != nil {
			return err
		}

		// update trx_ca_decision
		var ca entity.TrxCaDecision
		ca.CreatedAt = time.Now()
		if err := tx.Model(&ca).Where("ProspectID = ?", prospectID).Updates(ca).Error; err != nil {
			return err
		}

		// insert trx_details
		if err := tx.Create(&trxDetail).Error; err != nil {
			return err
		}

		// insert trx_history_approval_scheme
		if err := tx.Create(&trxHistoryApproval).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r repoHandler) GetInquiryApproval(req request.ReqInquiryApproval, pagination interface{}) (data []entity.InquiryCa, rowTotal int, err error) {

	var (
		filter         string
		filterBranch   string
		filterPaginate string
		query          string
		alias          string
		getRegion      []entity.RegionBranch
		encrypted      entity.EncryptString
	)

	alias = req.Alias

	rangeDays := os.Getenv("DEFAULT_RANGE_DAYS")

	if req.MultiBranch == "1" {
		getRegion, _ = r.GetRegionBranch(req.UserID)

		if len(getRegion) > 0 {
			extractBranchIDUser := ""
			userAllRegion := false
			for _, value := range getRegion {
				if strings.ToUpper(value.RegionName) == constant.REGION_ALL {
					userAllRegion = true
					break
				} else if value.BranchMember != "" {
					branch := strings.Trim(strings.ReplaceAll(value.BranchMember, `"`, `'`), "'")
					replace := strings.ReplaceAll(branch, `[`, ``)
					branchMember := strings.ReplaceAll(replace, `]`, ``)
					extractBranchIDUser += branchMember
					if value != getRegion[len(getRegion)-1] {
						extractBranchIDUser += ","
					}
				}
			}
			if userAllRegion {
				filterBranch += ""
			} else {
				filterBranch += "WHERE tm.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += ""
			if req.BranchID != "999" {
				filterBranch += "WHERE tm.BranchID = '" + req.BranchID + "'"
			}
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	if req.Search != "" {
		encrypted, _ = r.EncryptString(req.Search)
	}

	filter = utils.GenerateFilter(req.Search, encrypted.Encrypt, filterBranch, rangeDays, "")

	// Filter By
	if req.Filter != "" {
		var (
			activity string
		)
		switch req.Filter {
		case constant.DECISION_APPROVE:
			query = fmt.Sprintf(" AND tst.decision = '%s' AND tst.status_process='%s' AND has.source_decision='%s'", constant.DB_DECISION_APR, constant.STATUS_FINAL, alias)

		case constant.DECISION_REJECT:

			query = fmt.Sprintf(" AND tst.decision = '%s' AND tst.status_process='%s' AND has.source_decision='%s'", constant.DB_DECISION_REJECT, constant.STATUS_FINAL, alias)

		case constant.DECISION_CANCEL:
			query = fmt.Sprintf(" AND tst.decision = '%s' AND tst.status_process='%s' AND has.source_decision='%s'", constant.DB_DECISION_CANCEL, constant.STATUS_FINAL, alias)

		case constant.NEED_DECISION:
			activity = constant.ACTIVITY_UNPROCESS
			query = fmt.Sprintf(" AND tst.activity= '%s' AND tst.decision= '%s' AND tst.source_decision = '%s'", activity, constant.DB_DECISION_CREDIT_PROCESS, alias)
		}
	} else {
		query = fmt.Sprintf(" AND (has.next_step = '%s' OR has.source_decision='%s')", alias, alias)
	}

	filter = filter + query

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(SELECT tm.ProspectID
		FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID   	
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = '%s'
				OR has.source_decision = '%s'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		 %s) AS tt`, alias, alias, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT
		tt.*
		FROM
		(
		SELECT
		tm.ProspectID,
		cb.BranchName,
		cb.BranchID,
		tst.activity,
		tst.source_decision,
		tst.status_process,
		tst.decision,
		tst.reason,
		tcd.decision as ca_decision,
		tcd.decision_by,
		tcd.final_approval,
		has.next_step,
		has.decision AS approval_decision,
		has.source_decision AS approval_source_decision,
		CASE
		  WHEN tcd.final_approval='%s' THEN 1
		  ELSE 0
		END AS is_last_approval,
		CASE
		  WHEN rtn.decision IS NOT NULL THEN 1
		  ELSE 0
		END AS HasReturn,
		tcd.note AS ca_note,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP' THEN 1
		  ELSE 0
		END AS ActionFormAkk,
		CASE
		  WHEN tcd.decision = 'CAN' THEN tcd.created_at 
		  WHEN tcd.created_at IS NOT NULL THEN FORMAT(tfa.created_at,'yyyy-MM-dd HH:mm:ss')
		  ELSE FORMAT(tst.created_at,'yyyy-MM-dd HH:mm:ss')
		END AS ActionDate,
		CASE
		  WHEN (tfa.decision IS NULL)
		  AND (tcd.decision <> 'CAN') 
		  AND (tst.source_decision='%s') THEN 1
		  ELSE 0
		END AS ShowAction,
		CASE
		  WHEN tm.incoming_source = 'SLY' THEN 'SALLY'
		  ELSE 'NE'
		END AS incoming_source,
		tcp.CustomerID,
		tcp.CustomerStatus,
		tcp.SurveyResult,
		tm.created_at,
		tm.order_at,
		tm.lob,
		scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
		scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
		scp.dbo.DEC_B64('SEC', tcp.BirthPlace) AS BirthPlace,
		tcp.BirthDate,
		scp.dbo.DEC_B64('SEC', tcp.SurgateMotherName) AS SurgateMotherName,
		CASE
		  WHEN tcp.Gender = 'M' THEN 'Laki-Laki'
		  WHEN tcp.Gender = 'F' THEN 'Perempuan'
		END AS 'Gender',
		scp.dbo.DEC_B64('SEC', cal.Address) AS LegalAddress,
		CONCAT(cal.RT, '/', cal.RW) AS LegalRTRW,
		cal.Kelurahan AS LegalKelurahan,
		cal.Kecamatan AS LegalKecamatan,
		cal.ZipCode AS LegalZipcode,
		cal.City AS LegalCity,
		scp.dbo.DEC_B64('SEC', car.Address) AS ResidenceAddress,
		CONCAT(car.RT, '/', cal.RW) AS ResidenceRTRW,
		car.Kelurahan AS ResidenceKelurahan,
		car.Kecamatan AS ResidenceKecamatan,
		car.ZipCode AS ResidenceZipcode,
		car.City AS ResidenceCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		hst.value AS HomeStatus,
		mn.value AS StaySinceMonth,
		tcp.StaySinceYear,
		ta.ProductOfferingID,
		ta.dealer,
		ta.LifeInsuranceFee,
		ta.AssetInsuranceFee,
		'KMB MOTOR' AS AssetType,
		ti.asset_description,
		ti.manufacture_year,
		ti.color,
		chassis_number,
		engine_number,
		CASE
			WHEN ti.bpkb_name = 'K' THEN 'Sendiri'
			WHEN ti.bpkb_name = 'P' THEN 'Pasangan'
			WHEN ti.bpkb_name = 'KK' THEN 'Nama Satu KK'
			WHEN ti.bpkb_name = 'O' THEN 'Orang Lain'
		END AS bpkb_name,
		ti.owner_asset,
		ti.license_plate,
		interest_rate,
		Tenor AS InstallmentPeriod,
		OTR,
		DPAmount,
		AF AS FinanceAmount,
		interest_amount,
		insurance_amount,
		AdminFee,
		provision_fee,
		NTF,
		NTFAkumulasi,
		(NTF + interest_amount) AS Total,
		InstallmentAmount AS MonthlyInstallment,
		FirstInstallment,
		pr.value AS ProfessionID,
		jt.value AS JobType,
		jb.value AS JobPosition,
		mn2.value AS EmploymentSinceMonth,
		tce.EmploymentSinceYear,
		tce.CompanyName,
		cac.AreaPhone AS CompanyAreaPhone,
		cac.Phone AS CompanyPhone,
		tcp.ExtCompanyPhone,
		scp.dbo.DEC_B64('SEC', cac.Address) AS CompanyAddress,
		CONCAT(cac.RT, '/', cac.RW) AS CompanyRTRW,
		cac.Kelurahan AS CompanyKelurahan,
		cac.Kecamatan AS CompanyKecamatan,
		car.ZipCode AS CompanyZipcode,
		car.City AS CompanyCity,
		tce.MonthlyFixedIncome,
		tce.MonthlyVariableIncome,
		tce.SpouseIncome,
		tcp.SourceOtherIncome,
		tcs.FullName AS SpouseLegalName,
		tcs.CompanyName AS SpouseCompanyName,
		tcs.CompanyPhone AS SpouseCompanyPhone,
		tcs.MobilePhone AS SpouseMobilePhone,
		tcs.IDNumber AS SpouseIDNumber,
		pr2.value AS SpouseProfession,
		em.Name AS EmconName,
		em.Relationship,
		em.MobilePhone AS EmconMobilePhone,
	    scp.dbo.DEC_B64('SEC', cae.Address) AS EmergencyAddress,
		CONCAT(cae.RT, '/', cae.RW) AS EmergencyRTRW,
		cae.Kelurahan AS EmergencyKelurahan,
		cae.Kecamatan AS EmergencyKecamatan,
		cae.ZipCode AS EmergencyZipcode,
		cae.City AS EmergencyCity,
		cae.AreaPhone AS EmergencyAreaPhone,
		cae.Phone AS EmergencyPhone,
		tce.IndustryTypeID,
		FORMAT(tak.ScsDate,'dd-MM-yyyy') as ScsDate,
		tak.ScsScore,
		tak.ScsStatus,
		tdb.BiroCustomerResult,
		tdb.BiroSpouseResult,
		tak.UrlFormAkkk,
		tde.deviasi_id,
		mkd.deskripsi AS deviasi_description,
		'REJECT' AS deviasi_decision,
		tde.reason AS deviasi_reason,
		CASE
		  WHEN ted.ProspectID IS NOT NULL THEN 1
		  ELSE 0
		END AS is_edd,
		ted.is_highrisk,
		ted.pernyataan_1,
		ted.pernyataan_2,
		ted.pernyataan_3,
		ted.pernyataan_4,
		ted.pernyataan_5,
		ted.pernyataan_6

	  FROM
		trx_master tm WITH (nolock)
		INNER JOIN confins_branch cb WITH (nolock) ON tm.BranchID = cb.BranchID
		INNER JOIN trx_filtering tf WITH (nolock) ON tm.ProspectID = tf.prospect_id
		INNER JOIN trx_customer_personal tcp (nolock) ON tm.ProspectID = tcp.ProspectID
		INNER JOIN trx_apk ta WITH (nolock) ON tm.ProspectID = ta.ProspectID
		INNER JOIN trx_item ti WITH (nolock) ON tm.ProspectID = ti.ProspectID
		INNER JOIN trx_customer_employment tce WITH (nolock) ON tm.ProspectID = tce.ProspectID
		INNER JOIN trx_status tst WITH (nolock) ON tm.ProspectID = tst.ProspectID
		INNER JOIN trx_info_agent tia WITH (nolock) ON tm.ProspectID = tia.ProspectID
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN trx_akkk tak WITH (nolock) ON tm.ProspectID = tak.ProspectID
		LEFT JOIN trx_edd ted WITH (nolock) ON tm.ProspectID = ted.ProspectID
		LEFT JOIN trx_deviasi tde WITH (nolock) ON tm.ProspectID = tde.ProspectID
		LEFT JOIN m_kode_deviasi mkd WITH (nolock) ON tde.deviasi_id = mkd.deviasi_id
		OUTER APPLY (
			SELECT
			  TOP 1 *
			FROM
			  trx_history_approval_scheme has
			WHERE
			  (
				has.next_step = '%s'
				OR has.source_decision = '%s'
			  )
			  AND tm.ProspectID = has.ProspectID
			ORDER BY
			  has.created_at DESC
		  ) has
		LEFT JOIN (SELECT ProspectID, decision FROM trx_history_approval_scheme has WITH (nolock) WHERE has.decision = 'RTN') rtn ON rtn.ProspectID = tm.ProspectID
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			note,
			created_at,
			final_approval,
			decision_by
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN (
			SELECT prospect_id, 
			MAX(Case [subject] When 'CUSTOMER' Then url_pdf_report End) BiroCustomerResult,
			MAX(Case [subject] When 'SPOUSE' Then url_pdf_report End) BiroSpouseResult
			FROM trx_detail_biro
			GROUP BY prospect_id
		) tdb ON tm.ProspectID = tdb.prospect_id 

		INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'LEGAL'
		  ) cal ON tm.ProspectID = cal.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'RESIDENCE'
		  ) car ON tm.ProspectID = car.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'COMPANY'
		  ) cac ON tm.ProspectID = cac.ProspectID
		  INNER JOIN (
			SELECT
			  ProspectID,
			  Address,
			  RT,
			  RW,
			  Kelurahan,
			  Kecamatan,
			  ZipCode,
			  City,
			  Phone,
			  AreaPhone
			FROM
			  trx_customer_address WITH (nolock)
			WHERE
			  "Type" = 'EMERGENCY'
		  ) cae ON tm.ProspectID = cae.ProspectID

		INNER JOIN trx_customer_emcon em WITH (nolock) ON tm.ProspectID = em.ProspectID
		LEFT JOIN trx_customer_spouse tcs WITH (nolock) ON tm.ProspectID = tcs.ProspectID
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'Education'
		) edu ON tcp.Education = edu.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MaritalStatus'
		) mst ON tcp.MaritalStatus = mst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'HomeStatus'
		) hst ON tcp.HomeStatus = hst.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn ON tcp.StaySinceMonth = mn.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr ON tce.ProfessionID = pr.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobType'
		) jt ON tce.JobType = jt.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'JobPosition'
		) jb ON tce.JobPosition = jb.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'MonthName'
		) mn2 ON tce.EmploymentSinceMonth = mn2.[key]
		LEFT JOIN (
		  SELECT
			[key],
			value
		  FROM
			app_config ap WITH (nolock)
		  WHERE
			group_name = 'ProfessionID'
		) pr2 ON tcs.ProfessionID = pr2.[key]
	 %s) AS tt ORDER BY tt.created_at DESC %s`, alias, alias, alias, alias, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) SubmitApproval(req request.ReqSubmitApproval, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail, trxRecalculate entity.TrxRecalculate, approval response.RespApprovalScheme) (status entity.TrxStatus, err error) {

	trxStatus.CreatedAt = time.Now()
	trxDetail.CreatedAt = time.Now()
	trxRecalculate.CreatedAt = time.Now()

	err = r.NewKmb.Transaction(func(tx *gorm.DB) error {

		// cek trx status terbaru dan pengajuan deviasi atau bukan
		var cekstatus entity.TrxStatus
		if err := tx.Raw(fmt.Sprintf(`SELECT ts.ProspectID, 
				CASE 
 					WHEN td.ProspectID IS NOT NULL AND tcp.CustomerStatus = 'NEW' THEN 'DEV'
					ELSE NULL
				END AS activity 
				FROM trx_status ts
				LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID 
				LEFT JOIN trx_deviasi td ON ts.ProspectID = td.ProspectID 
				WHERE ts.ProspectID = '%s' AND ts.status_process = '%s'`, trxStatus.ProspectID, constant.STATUS_ONPROCESS)).Scan(&cekstatus).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				err = errors.New(constant.RECORD_NOT_FOUND)
			}
			return err
		}

		// jika pengajuan deviasi approve maka cek kuota dan kurangi kuota deviasi
		if approval.IsFinal && !approval.IsEscalation && req.Decision == constant.DECISION_APPROVE && cekstatus.Activity == constant.SOURCE_DECISION_DEVIASI {

			// cek kuota deviasi
			// kuota tersedia, kurangi kuota deviasi
			var confirmDeviasi entity.ConfirmDeviasi
			if err = tx.Raw(fmt.Sprintf(`UPDATE m_branch_deviasi 
				SET booking_amount = q.booking_amount+q.NTF, booking_account = q.booking_account+1, balance_amount = q.balance_amount-q.NTF, balance_account = q.balance_account-1
				OUTPUT q.NTF, inserted.*,
				CASE 
					WHEN inserted.booking_account = deleted.booking_account+1 THEN 1
					ELSE 0
				END as deviasi
				FROM (
					SELECT ta.NTF, mbd.*
					FROM m_branch_deviasi mbd
					LEFT JOIN trx_master tm ON mbd.BranchID = tm.BranchID 
					LEFT JOIN trx_apk ta ON tm.ProspectID = ta.ProspectID 
					WHERE tm.ProspectID = '%s'
				) as q
				WHERE m_branch_deviasi.BranchID = q.BranchID AND m_branch_deviasi.is_active = 1 AND q.balance_amount >= q.NTF AND q.balance_account > 0
				`, trxStatus.ProspectID)).Scan(&confirmDeviasi).Error; err != nil {
				// record not found artinya kuota deviasi tidak tersedia
				if err != gorm.ErrRecordNotFound {
					return err
				}
			}

			if confirmDeviasi.Deviasi {

				info, _ := json.Marshal(confirmDeviasi)
				trxDetail.Info = string(info)

			} else {

				// kuota tidak tersedia, reject deviasi
				// tidak mengubah rule_code karena rule_code digunakan didetail transaksi (rule_code credit committe)
				trxStatus.Decision = constant.DB_DECISION_REJECT
				trxStatus.Reason = constant.REASON_REJECT_KUOTA_DEVIASI

				trxDetail.Decision = constant.DB_DECISION_REJECT
				trxDetail.Reason = constant.REASON_REJECT_KUOTA_DEVIASI

				req.Decision = constant.DECISION_REJECT
				req.Reason = constant.REASON_REJECT_KUOTA_DEVIASI
				req.Note = constant.REASON_REJECT_KUOTA_DEVIASI
			}

		}

		// trx_status
		if err := tx.Model(&trxStatus).Where("ProspectID = ?", trxStatus.ProspectID).Updates(trxStatus).Error; err != nil {
			return err
		}

		// trx_details
		if err := tx.Create(&trxDetail).Error; err != nil {
			return err
		}

		var (
			trxHistoryApproval entity.TrxHistoryApprovalScheme
			nextFinal          int
			isEscalation       int
			decision           string
		)

		isEscalation = 0
		if approval.IsEscalation {
			isEscalation = 1
			nextFinal = 1
		}

		if approval.IsFinal {
			nextFinal = 1
		}

		switch req.Decision {
		case constant.DECISION_REJECT:
			decision = constant.DB_DECISION_REJECT
		case constant.DECISION_APPROVE:
			decision = constant.DB_DECISION_APR
		}

		if approval.IsEscalation {
			trxCaDecision := entity.TrxCaDecision{
				FinalApproval: approval.NextStep,
			}

			// trx_ca_decision
			if err := tx.Model(&trxCaDecision).Where("ProspectID = ?", req.ProspectID).Updates(trxCaDecision).Error; err != nil {
				return err
			}
		}

		trxHistoryApproval = entity.TrxHistoryApprovalScheme{
			ID:                    uuid.New().String(),
			ProspectID:            req.ProspectID,
			Decision:              decision,
			Reason:                req.Reason,
			Note:                  req.Note,
			CreatedAt:             time.Now(),
			CreatedBy:             req.CreatedBy,
			DecisionBy:            req.DecisionBy,
			NextFinalApprovalFlag: nextFinal,
			NeedEscalation:        isEscalation,
			SourceDecision:        trxDetail.SourceDecision,
			NextStep:              approval.NextStep,
		}

		if req.Decision == constant.DECISION_RETURN {
			trxHistoryApproval.Decision = constant.DB_DECISION_RTN
			trxHistoryApproval.Reason = constant.REASON_RETURN_APPROVAL
			trxHistoryApproval.Note = constant.DECISION_RETURN
			trxHistoryApproval.NextFinalApprovalFlag = 0
			trxHistoryApproval.NextStep = constant.DB_DECISION_CREDIT_ANALYST

			// insert trx_recalculate
			if err := tx.Create(&trxRecalculate).Error; err != nil {
				return err
			}
		}

		// trx_history_approval_scheme
		if err := tx.Create(&trxHistoryApproval).Error; err != nil {
			return err
		}

		if approval.IsFinal && !approval.IsEscalation && req.Decision != constant.DECISION_RETURN {
			trxFinalApproval := entity.TrxFinalApproval{
				ProspectID: req.ProspectID,
				Decision:   decision,
				Reason:     req.Reason,
				Note:       req.Note,
				CreatedAt:  time.Now(),
				CreatedBy:  req.CreatedBy,
				DecisionBy: req.DecisionBy,
			}

			// trx_final_approval
			if err := tx.Create(&trxFinalApproval).Error; err != nil {
				return err
			}

			// will insert trx_agreement
			if decision == constant.DB_DECISION_APR {

				getAFPhone, _ := r.GetAFMobilePhone(req.ProspectID)

				trxAgreement := entity.TrxAgreement{
					ProspectID:         req.ProspectID,
					CheckingStatus:     constant.ACTIVITY_UNPROCESS,
					ContractStatus:     "0",
					AF:                 getAFPhone.AFValue,
					MobilePhone:        getAFPhone.MobilePhone,
					CustomerIDKreditmu: constant.LOB_NEW_KMB,
				}

				if err := tx.Create(&trxAgreement).Error; err != nil {
					return err
				}

				if err == nil && (len(req.ProspectID) > 2 && req.ProspectID[0:2] != "NE") {
					return r.losDB.Transaction(func(tx2 *gorm.DB) error {
						// worker insert staging
						callbackHeaderLos, _ := json.Marshal(
							map[string]string{
								"X-Client-ID":   os.Getenv("CLIENT_LOS"),
								"Authorization": os.Getenv("AUTH_LOS"),
							})

						if err := tx2.Create(&entity.TrxWorker{
							ProspectID:      req.ProspectID,
							Category:        "CONFINS",
							Action:          "INSERT_STAGING_KMB",
							APIType:         "RAW",
							EndPointTarget:  fmt.Sprintf("%s/%s", os.Getenv("INSERT_STAGING_URL"), req.ProspectID),
							EndPointMethod:  constant.METHOD_POST,
							Header:          string(callbackHeaderLos),
							ResponseTimeout: 30,
							MaxRetry:        6,
							CountRetry:      0,
							Activity:        constant.ACTIVITY_UNPROCESS,
						}).Error; err != nil {
							tx.Rollback()
							return err
						}
						return nil
					})
				}

			}
		}

		return nil
	})

	return trxStatus, err
}

func (r repoHandler) GetInquiryQuotaDeviasi(req request.ReqListQuotaDeviasi, pagination interface{}) (data []entity.InquirySettingQuotaDeviasi, rowTotal int, err error) {

	var (
		filterBuilder  strings.Builder
		conditions     []string
		filterPaginate string
		x              sql.TxOptions
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(mbd.BranchID LIKE '%%%[1]s%%' OR cb.BranchName LIKE '%%%[1]s%%')", req.Search))
	}

	if req.BranchID != "" {
		numbers := strings.Split(req.BranchID, ",")
		for i, number := range numbers {
			numbers[i] = "'" + number + "'"
		}
		conditions = append(conditions, fmt.Sprintf("mbd.BranchID IN (%s)", strings.Join(numbers, ",")))
	}

	if req.IsActive != "" {
		conditions = append(conditions, fmt.Sprintf("mbd.is_active = '%s'", req.IsActive))
	}

	if len(conditions) > 0 {
		filterBuilder.WriteString("WHERE ")
		filterBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	filter := filterBuilder.String()

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT
				COUNT(*) AS totalRow
			FROM (
				SELECT mbd.*, cb.BranchName AS branch_name
				FROM m_branch_deviasi AS mbd WITH (nolock)
				JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID) %s
			) AS y`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT mbd.BranchID, cb.BranchName AS branch_name, mbd.quota_amount, mbd.quota_account, mbd.booking_amount, mbd.booking_account, mbd.balance_amount, mbd.balance_account, mbd.is_active, mbd.updated_by, ISNULL(FORMAT(mbd.updated_at, 'yyyy-MM-dd HH:mm:ss'), '') AS updated_at
			FROM m_branch_deviasi AS mbd WITH (nolock)
			JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID) %s ORDER BY mbd.is_active DESC, mbd.BranchID ASC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) GetQuotaDeviasiBranch(req request.ReqListQuotaDeviasiBranch) (data []entity.ConfinsBranch, err error) {
	var (
		filterBuilder strings.Builder
		conditions    []string
		x             sql.TxOptions
	)

	if req.BranchID != "" {
		numbers := strings.Split(req.BranchID, ",")
		for i, number := range numbers {
			numbers[i] = "'" + number + "'"
		}
		conditions = append(conditions, fmt.Sprintf("mbd.BranchID IN (%s)", strings.Join(numbers, ",")))
	}

	if req.BranchName != "" {
		conditions = append(conditions, fmt.Sprintf("cb.BranchName LIKE '%%%[1]s%%'", req.BranchName))
	}

	if len(conditions) > 0 {
		filterBuilder.WriteString("WHERE ")
		filterBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	filter := filterBuilder.String()

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT DISTINCT mbd.BranchID, cb.BranchName
			FROM m_branch_deviasi AS mbd WITH (nolock)
			JOIN confins_branch AS cb ON (mbd.BranchID = cb.BranchID) %s
			ORDER BY cb.BranchName ASC`, filter)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) ProcessUpdateQuotaDeviasiBranch(branchID string, mBranchDeviasi entity.MappingBranchDeviasi) (dataBefore entity.DataQuotaDeviasiBranch, dataAfter entity.DataQuotaDeviasiBranch, err error) {
	err = r.NewKmb.Transaction(func(tx *gorm.DB) error {
		// Step 1: Retrieve current values for dataBefore
		if err := tx.Raw("SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?", branchID).Scan(&dataBefore).Error; err != nil {
			return err
		}

		// Check if BookingAmount or BookingAccount exceeds the new Quota
		if dataBefore.BookingAmount > mBranchDeviasi.QuotaAmount {
			return fmt.Errorf("BookingAmount > QuotaAmount")
		}
		if dataBefore.BookingAccount > mBranchDeviasi.QuotaAccount {
			return fmt.Errorf("BookingAccount > QuotaAccount")
		}

		// Step 2: Update quota and calculate new balances
		newBalanceAmount := mBranchDeviasi.QuotaAmount - dataBefore.BookingAmount
		newBalanceAccount := mBranchDeviasi.QuotaAccount - dataBefore.BookingAccount

		if err := tx.Model(&entity.MappingBranchDeviasi{}).
			Where("BranchID = ?", branchID).
			Updates(map[string]interface{}{
				"quota_amount":    mBranchDeviasi.QuotaAmount,
				"quota_account":   mBranchDeviasi.QuotaAccount,
				"balance_amount":  newBalanceAmount,
				"balance_account": newBalanceAccount,
				"is_active":       mBranchDeviasi.IsActive,
				"updated_at":      time.Now(),
				"updated_by":      mBranchDeviasi.UpdatedBy,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Step 3: Retrieve updated values for dataAfter
		if err := tx.Raw("SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?", branchID).Scan(&dataAfter).Error; err != nil {
			return err
		}

		return nil
	})

	return
}

func (r repoHandler) BatchUpdateQuotaDeviasi(data []entity.MappingBranchDeviasi) (dataBeforeList []entity.MappingBranchDeviasi, dataAfterList []entity.MappingBranchDeviasi, err error) {

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	txOptions := &sql.TxOptions{}

	db := r.NewKmb.BeginTx(ctx, txOptions)
	defer db.Commit()

	defer func() {
		if r := recover(); r != nil || err != nil {
			db.Rollback()
		}
	}()

	// Create a map to store the updates and a slice to store the branch IDs
	updatesMap := make(map[string]map[string]interface{})
	var branchIDs []string
	var updatedBranchIDs []string

	// Collect branch IDs
	for _, newData := range data {
		branchIDs = append(branchIDs, newData.BranchID)
	}

	// Fetch all current data in one query
	var currentDataList []entity.MappingBranchDeviasi
	if err := db.Raw("SELECT BranchID, final_approval, quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID IN (?)", branchIDs).Scan(&currentDataList).Error; err != nil {
		return nil, nil, err
	}

	// Create a map for quick lookup of current data by BranchID
	currentDataMap := make(map[string]entity.MappingBranchDeviasi)
	for _, currentData := range currentDataList {
		currentDataMap[currentData.BranchID] = currentData
	}

	// Prepare updates
	for _, newData := range data {
		currentData, exists := currentDataMap[newData.BranchID]
		if !exists {
			continue
		}

		updates := make(map[string]interface{})

		if currentData.QuotaAmount != newData.QuotaAmount {
			updates["quota_amount"] = newData.QuotaAmount
		}
		if currentData.QuotaAccount != newData.QuotaAccount {
			updates["quota_account"] = newData.QuotaAccount
		}
		if currentData.IsActive != newData.IsActive {
			updates["is_active"] = newData.IsActive
		}

		if len(updates) > 0 {
			// Store dataBefore only if there are updates
			dataBeforeList = append(dataBeforeList, currentData)

			// Calculate new balances
			newBalanceAmount := newData.QuotaAmount - currentData.BookingAmount
			newBalanceAccount := newData.QuotaAccount - currentData.BookingAccount

			// Check for negative balances
			if newBalanceAmount < 0 {
				return nil, nil, fmt.Errorf(constant.ERROR_BAD_REQUEST + " - BookingAmount > QuotaAmount")
			}
			if newBalanceAccount < 0 {
				return nil, nil, fmt.Errorf(constant.ERROR_BAD_REQUEST + " - BookingAccount > QuotaAccount")
			}

			updates["balance_amount"] = newBalanceAmount
			updates["balance_account"] = newBalanceAccount

			updates["updated_at"] = time.Now()
			updates["updated_by"] = newData.UpdatedBy

			// Store updates in the map
			updatesMap[newData.BranchID] = updates

			// Add to updatedBranchIDs
			updatedBranchIDs = append(updatedBranchIDs, newData.BranchID)
		}
	}

	if len(updatesMap) > 0 {
		// Perform bulk update
		for branchID, updates := range updatesMap {
			if err := db.Model(&entity.MappingBranchDeviasi{}).
				Where("BranchID = ?", branchID).
				Updates(updates).Error; err != nil {
				return nil, nil, err
			}
		}

		// Retrieve updated data in bulk
		if err := db.Raw("SELECT BranchID, final_approval, quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID IN (?)", updatedBranchIDs).Scan(&dataAfterList).Error; err != nil {
			return nil, nil, err
		}
	}

	return dataBeforeList, dataAfterList, nil
}

func (r repoHandler) ProcessResetQuotaDeviasiBranch(branchID string, updatedBy string) (dataBefore entity.DataQuotaDeviasiBranch, dataAfter entity.DataQuotaDeviasiBranch, err error) {
	err = r.NewKmb.Transaction(func(tx *gorm.DB) error {
		// Step 1: Retrieve current values for dataBefore
		if err := tx.Raw("SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?", branchID).Scan(&dataBefore).Error; err != nil {
			return err
		}

		if err := tx.Model(&entity.MappingBranchDeviasi{}).
			Where("BranchID = ?", branchID).
			Updates(map[string]interface{}{
				"quota_amount":    0,
				"quota_account":   0,
				"booking_amount":  0,
				"booking_account": 0,
				"balance_amount":  0,
				"balance_account": 0,
				"is_active":       false,
				"updated_at":      time.Now(),
				"updated_by":      updatedBy,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Step 3: Retrieve updated values for dataAfter
		if err := tx.Raw("SELECT TOP 1 quota_amount, quota_account, booking_amount, booking_account, balance_amount, balance_account, is_active, updated_at, updated_by FROM m_branch_deviasi WITH (nolock) WHERE BranchID = ?", branchID).Scan(&dataAfter).Error; err != nil {
			return err
		}

		return nil
	})

	return
}

func (r repoHandler) ProcessResetAllQuotaDeviasi(updatedBy string) (err error) {
	err = r.NewKmb.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&entity.MappingBranchDeviasi{}).
			Updates(map[string]interface{}{
				"quota_amount":    0,
				"quota_account":   0,
				"booking_amount":  0,
				"booking_account": 0,
				"balance_amount":  0,
				"balance_account": 0,
				"is_active":       false,
				"updated_at":      time.Now(),
				"updated_by":      updatedBy,
			}).Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	})

	return
}

func (r repoHandler) GetInquiryListOrder(req request.ReqInquiryListOrder, pagination interface{}) (data []entity.InquiryDataListOrder, rowTotal int, err error) {

	var (
		filterBuilder  strings.Builder
		conditions     []string
		filterPaginate string
		x              sql.TxOptions
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if req.BranchID != "" && req.BranchID != "999" {
		conditions = append(conditions, fmt.Sprintf("tm.BranchID = '%s'", req.BranchID))
	}

	if req.Decision != "" && req.Decision != "ALL" {
		conditions = append(conditions, fmt.Sprintf("sts.decision = '%s'", req.Decision))
	}

	if req.IsHighRisk != "" && req.IsHighRisk != "ALL" {
		conditions = append(conditions, fmt.Sprintf("edd.is_highrisk = %s", req.IsHighRisk))
	}

	if req.ProspectID != "" || req.IDNumber != "" || req.LegalName != "" {
		if req.ProspectID != "" {
			conditions = append(conditions, fmt.Sprintf("tm.ProspectID = '%s'", req.ProspectID))
		}

		if req.IDNumber != "" {
			encrypted, _ := r.EncryptString(req.IDNumber)
			conditions = append(conditions, fmt.Sprintf("tcp.IDNumber = '%s'", encrypted.Encrypt))
		}

		if req.LegalName != "" {
			encrypted, _ := r.EncryptString(req.LegalName)
			conditions = append(conditions, fmt.Sprintf("tcp.LegalName = '%s'", encrypted.Encrypt))
		}
	} else {
		startDate, _ := time.Parse("2006-01-02", req.OrderDateStart)
		endDate, _ := time.Parse("2006-01-02", req.OrderDateEnd)

		startDate = startDate.Add(time.Hour * 0).Add(time.Minute * 0).Add(time.Second * 0)
		endDate = endDate.Add(time.Hour * 23).Add(time.Minute * 59).Add(time.Second * 59)

		startDateFormatted := startDate.Format(time.RFC3339)
		endDateFormatted := endDate.Format(time.RFC3339)

		conditions = append(conditions, fmt.Sprintf("tm.created_at BETWEEN '%s' AND '%s'", startDateFormatted, endDateFormatted))
	}

	if len(conditions) > 0 {
		filterBuilder.WriteString("WHERE ")
		filterBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	filter := filterBuilder.String()

	rawQuery := `SELECT 
					tm.created_at AS OrderAt,
					b.BranchName,
					tm.ProspectID, 
					scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
					scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
					tcp.BirthDate,
					prf.[value] AS Profession,
					jt.[value] AS JobType,
					jp.[value] AS JobPosition,
					edd.is_highrisk AS IsHighRisk,
					edd.pernyataan_1 AS Pernyataan1,
					edd.pernyataan_2 AS Pernyataan2,
					edd.pernyataan_3 AS Pernyataan3,
					edd.pernyataan_4 AS Pernyataan4,
					edd.pernyataan_5 AS Pernyataan5,
					edd.pernyataan_6 AS Pernyataan6,
					sts.decision AS Decision,
					tcd.decision_by AS DecisionBy,
					edd.created_at AS DecisionAt
				FROM 
				trx_master AS tm WITH (nolock)
				JOIN confins_branch AS b ON (tm.BranchID = b.BranchID)
				JOIN trx_status AS sts ON (tm.ProspectID = sts.ProspectID)
				JOIN trx_customer_personal AS tcp ON (tm.ProspectID = tcp.ProspectID)
				JOIN trx_customer_employment AS emp ON (tm.ProspectID = emp.ProspectID)
				LEFT JOIN trx_ca_decision AS tcd ON (tm.ProspectID = tcd.ProspectID) 
				LEFT JOIN trx_edd AS edd ON (tm.ProspectID = edd.ProspectID)
				LEFT JOIN (
					SELECT [key], value
					FROM app_config ap WITH (nolock)
					WHERE group_name = 'ProfessionID'
				) AS prf ON (emp.ProfessionID = prf.[key])
				LEFT JOIN (
					SELECT [key], value
					FROM app_config ap WITH (nolock)
					WHERE group_name = 'JobType'
				) AS jt ON (emp.JobType = jt.[key])
				LEFT JOIN (
					SELECT [key], value
					FROM app_config ap WITH (nolock)
					WHERE group_name = 'JobPosition'
				) AS jp ON (emp.JobPosition = jp.[key])`

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT
				COUNT(*) AS totalRow
			FROM (
				%s %s
			) AS y`, rawQuery, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`%s %s ORDER BY tm.created_at DESC %s`, rawQuery, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}

	return
}

func (r repoHandler) GetInquiryListOrderDetail(prospectID string) (data entity.InquiryDataListOrder, err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	rawQuery := `SELECT 
					tm.created_at AS OrderAt,
					b.BranchName,
					tm.ProspectID, 
					scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName,
					scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
					tcp.BirthDate,
					prf.[value] AS Profession,
					jt.[value] AS JobType,
					jp.[value] AS JobPosition,
					edd.is_highrisk AS IsHighRisk,
					edd.pernyataan_1 AS Pernyataan1,
					edd.pernyataan_2 AS Pernyataan2,
					edd.pernyataan_3 AS Pernyataan3,
					edd.pernyataan_4 AS Pernyataan4,
					edd.pernyataan_5 AS Pernyataan5,
					edd.pernyataan_6 AS Pernyataan6,
					sts.decision AS Decision,
					tcd.decision_by AS DecisionBy,
					edd.created_at AS DecisionAt
				FROM 
				trx_master AS tm WITH (nolock)
				JOIN confins_branch AS b ON (tm.BranchID = b.BranchID)
				JOIN trx_status AS sts ON (tm.ProspectID = sts.ProspectID)
				JOIN trx_customer_personal AS tcp ON (tm.ProspectID = tcp.ProspectID)
				JOIN trx_customer_employment AS emp ON (tm.ProspectID = emp.ProspectID)
				LEFT JOIN trx_ca_decision AS tcd ON (tm.ProspectID = tcd.ProspectID) 
				LEFT JOIN trx_edd AS edd ON (tm.ProspectID = edd.ProspectID)
				LEFT JOIN (
					SELECT [key], value
					FROM app_config ap WITH (nolock)
					WHERE group_name = 'ProfessionID'
				) AS prf ON (emp.ProfessionID = prf.[key])
				LEFT JOIN (
					SELECT [key], value
					FROM app_config ap WITH (nolock)
					WHERE group_name = 'JobType'
				) AS jt ON (emp.JobType = jt.[key])
				LEFT JOIN (
					SELECT [key], value
					FROM app_config ap WITH (nolock)
					WHERE group_name = 'JobPosition'
				) AS jp ON (emp.JobPosition = jp.[key])
				WHERE tm.ProspectID = ?`

	if err = r.NewKmb.Raw(rawQuery, prospectID).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetMappingCluster() (data []entity.MasterMappingCluster, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.losDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.losDB.Raw("SELECT * FROM kmb_mapping_cluster_branch WITH (nolock) ORDER BY branch_id ASC").Scan(&data).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetInquiryMappingCluster(req request.ReqListMappingCluster, pagination interface{}) (data []entity.InquiryMappingCluster, rowTotal int, err error) {

	var (
		filterBuilder  strings.Builder
		conditions     []string
		filterPaginate string
		x              sql.TxOptions
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.losDB.BeginTx(ctx, &x)
	defer db.Commit()

	if req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(kmcb.branch_id LIKE '%%%[1]s%%' OR cb.BranchName LIKE '%%%[1]s%%')", req.Search))
	}

	if req.BranchID != "" {
		numbers := strings.Split(req.BranchID, ",")
		for i, number := range numbers {
			numbers[i] = "'" + number + "'"
		}
		conditions = append(conditions, fmt.Sprintf("kmcb.branch_id IN (%s)", strings.Join(numbers, ",")))
	}

	if req.CustomerStatus != "" {
		conditions = append(conditions, fmt.Sprintf("kmcb.customer_status = '%s'", req.CustomerStatus))
	}

	if req.BPKBNameType != "" {
		conditions = append(conditions, fmt.Sprintf("kmcb.bpkb_name_type = '%s'", req.BPKBNameType))
	}

	if req.Cluster != "" {
		conditions = append(conditions, fmt.Sprintf("kmcb.cluster = '%s'", req.Cluster))
	}

	if len(conditions) > 0 {
		filterBuilder.WriteString("WHERE ")
		filterBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	filter := filterBuilder.String()

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.losDB.Raw(fmt.Sprintf(`SELECT
				COUNT(*) AS totalRow
			FROM (
				SELECT kmcb.*, cb.BranchName AS branch_name 
				FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
				LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID %s
			) AS y`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT
		kmcb.*, 
		cb.BranchName AS branch_name
		FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
		LEFT JOIN confins_branch cb ON kmcb.branch_id = cb.BranchID %s ORDER BY kmcb.branch_id ASC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) BatchUpdateMappingCluster(data []entity.MasterMappingCluster, history entity.HistoryConfigChanges) (err error) {

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	txOptions := &sql.TxOptions{}

	db := r.losDB.BeginTx(ctx, txOptions)
	defer db.Commit()

	defer func() {
		if r := recover(); r != nil || err != nil {
			db.Rollback()
		}
	}()

	var cluster entity.MasterMappingCluster
	if err = db.Delete(&cluster).Error; err != nil {
		return err
	}

	for _, val := range data {
		if err = db.Create(&val).Error; err != nil {
			return err
		}
	}

	if err = db.Create(&history).Error; err != nil {
		return err
	}

	return err
}

func (r repoHandler) GetMappingClusterBranch(req request.ReqListMappingClusterBranch) (data []entity.ConfinsBranch, err error) {
	var (
		filterBuilder strings.Builder
		conditions    []string
		x             sql.TxOptions
	)

	if req.BranchID != "" {
		numbers := strings.Split(req.BranchID, ",")
		for i, number := range numbers {
			numbers[i] = "'" + number + "'"
		}
		conditions = append(conditions, fmt.Sprintf("kmcb.branch_id IN (%s)", strings.Join(numbers, ",")))
	}

	if req.BranchName != "" {
		conditions = append(conditions, fmt.Sprintf("cb.BranchName LIKE '%%%[1]s%%'", req.BranchName))
	}

	if len(conditions) > 0 {
		filterBuilder.WriteString("WHERE ")
		filterBuilder.WriteString(strings.Join(conditions, " AND "))
	}

	filter := filterBuilder.String()

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.losDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT DISTINCT 
		kmcb.branch_id AS BranchID, 
		CASE 
			WHEN kmcb.branch_id = '000' THEN 'PRIME PRIORITY'
			ELSE cb.BranchName 
		END AS BranchName 
		FROM kmb_mapping_cluster_branch kmcb WITH (nolock)
		LEFT JOIN confins_branch cb ON cb.BranchID = kmcb.branch_id %s 
		ORDER BY kmcb.branch_id ASC`, filter)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SaveWorker(trxworker entity.TrxWorker) (err error) {
	err = r.losDB.Create(&trxworker).Error
	return
}

func (r repoHandler) SaveUrlFormAKKK(prospectID, urlFormAKKK string) (err error) {

	var (
		data entity.TrxAkkk
		x    sql.TxOptions
	)

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	result := db.Model(&data).Where("ProspectID = ?", prospectID).Updates(entity.TrxAkkk{
		UrlFormAkkk: urlFormAKKK,
	})

	if err = result.Error; err != nil {
		return
	}

	if result.RowsAffected == 0 {
		err = errors.New("RowsAffected 0")
	}
	return
}

func (r repoHandler) GetMappingClusterChangeLog(pagination interface{}) (data []entity.MappingClusterChangeLog, rowTotal int, err error) {
	var (
		filterPaginate string
		x              sql.TxOptions
	)

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.losDB.BeginTx(ctx, &x)
	defer db.Commit()

	if pagination != nil {
		page, _ := json.Marshal(pagination)
		var paginationFilter request.RequestPagination
		jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal(page, &paginationFilter)
		if paginationFilter.Page == 0 {
			paginationFilter.Page = 1
		}

		offset := paginationFilter.Limit * (paginationFilter.Page - 1)

		var row entity.TotalRow

		if err = r.losDB.Raw(`SELECT 
				COUNT(*) AS totalRow
			FROM history_config_changes hcc 
			LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
			WHERE hcc.config_id = 'kmb_mapping_cluster_branch'
		`).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT
			hcc.id, hcc.data_before, hcc.data_after, hcc.created_at, ud.name AS user_name
		FROM history_config_changes hcc 
		LEFT JOIN user_details ud ON ud.user_id = hcc.created_by 
		WHERE hcc.config_id = 'kmb_mapping_cluster_branch'
		ORDER BY hcc.created_at DESC %s`, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) EncryptString(data string) (encrypted entity.EncryptString, err error) {

	var regexpPpid = regexp.MustCompile(`SAL-|NE-`)

	// check if data is not ProspectID
	if data != "" && !regexpPpid.MatchString(data) {
		if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS encrypt`, data)).Scan(&encrypted).Error; err != nil {
			return
		}
	} else {
		encrypted.Encrypt = data
	}

	return
}
