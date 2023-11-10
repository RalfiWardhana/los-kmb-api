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
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
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
				filterBranch += "WHERE tt.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += "WHERE tt.BranchID = '" + req.BranchID + "'"
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	filter = utils.GenerateFilter(req.Search, filterBranch, rangeDays)

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
			cb.BranchID,
			tm.ProspectID,
			tm.created_at,
			scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
			scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName
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
		) AS tt %s`, filter)).Scan(&row).Error; err != nil {
			return
		}

		rowTotal = row.Total

		filterPaginate = fmt.Sprintf("OFFSET %d ROWS FETCH FIRST %d ROWS ONLY", offset, paginationFilter.Limit)
	}

	if err = r.NewKmb.Raw(fmt.Sprintf(`SELECT tt.* FROM (
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
	) pr2 ON tcs.ProfessionID = pr2.[key] ) AS tt %s ORDER BY tt.created_at DESC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
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
				  WHEN thas.source_decision = 'CRA' AND tcd.slik_result<>'' THEN tcd.slik_result
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
		NULL as Plafond,
		tdb.baki_debet_non_collateral as BakiDebet,
		NULL as FasilitasAktif,
		NULL as ColTerburuk,
		NULL as BakiDebetTerburuk,
		NULL as ColTerakhirAktif,
		NULL as SpousePlafond,
		tdb2.baki_debet_non_collateral as SpouseBakiDebet,
		NULL as SpouseFasilitasAktif,
		NULL as SpouseColTerburuk,
		NULL as SpouseBakiDebetTerburuk,
		NULL as SpouseColTerakhirAktif,
		ta.ScsScore,
		ta.AgreementStatus,
		ta.TotalAgreementAktif,
		ta.MaxOVDAgreementAktif,
		ta.LastMaxOVDAgreement,
		tf.customer_segment,
		ta.LatestInstallment,
		ta2.NTFAkumulasi,
		(CAST(ISNULL(ta2.InstallmentAmount, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tf.total_installment_amount_biro, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountSpouseFMF, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOther, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(ta.InstallmentAmountOtherSpouse, 0) AS NUMERIC(17,2)) -
		CAST(ISNULL(ta.InstallmentTopup, 0) AS NUMERIC(17,2)) ) as TotalInstallment,
		(CAST(ISNULL(tce.MonthlyFixedIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.MonthlyVariableIncome, 0) AS NUMERIC(17,2)) +
		CAST(ISNULL(tce.SpouseIncome, 0) AS NUMERIC(17,2)) ) as TotalIncome,
		ta.TotalDSR,
		ta.EkycSource,
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
		LEFT OUTER JOIN ( 
			SELECT TOP 1 * FROM trx_history_approval_scheme thas WITH (nolock)
			WHERE source_decision = 'CBM'
			ORDER BY created_at DESC 
		) AS cbm ON ts.ProspectID = cbm.ProspectID
		LEFT OUTER JOIN ( 
			SELECT TOP 1 * FROM trx_history_approval_scheme thas WITH (nolock)
			WHERE source_decision = 'DRM'
			ORDER BY created_at DESC 
		) AS drm ON ts.ProspectID = drm.ProspectID
		LEFT OUTER JOIN ( 
			SELECT TOP 1 * FROM trx_history_approval_scheme thas WITH (nolock)
			WHERE source_decision = 'GMO'
			ORDER BY created_at DESC 
		) AS gmo ON ts.ProspectID = gmo.ProspectID
		WHERE ts.ProspectID = '%s'`, prospectID)).Scan(&data).Error; err != nil {
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
				filterBranch += "WHERE tt.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += "WHERE tt.BranchID = '" + req.BranchID + "'"
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	filter = utils.GenerateFilter(req.Search, filterBranch, rangeDays)

	// Filter By
	if req.Filter != "" {
		var (
			decision string
			activity string
		)
		switch req.Filter {
		case constant.DECISION_APPROVE:
			decision = constant.DB_DECISION_APR
			query = fmt.Sprintf(" AND tt.decision= '%s'", decision)

		case constant.DECISION_REJECT:
			decision = constant.DB_DECISION_REJECT
			query = fmt.Sprintf(" AND tt.decision= '%s'", decision)

		case constant.DECISION_CANCEL:
			decision = constant.DB_DECISION_CANCEL
			query = fmt.Sprintf(" AND tt.decision= '%s'", decision)

		case constant.NEED_DECISION:
			activity = constant.ACTIVITY_UNPROCESS
			decision = constant.DB_DECISION_CREDIT_PROCESS
			source := constant.DB_DECISION_CREDIT_ANALYST
			query = fmt.Sprintf(" AND tt.activity= '%s' AND tt.decision= '%s' AND tt.source_decision = '%s' AND tt.decision_ca IS NULL", activity, decision, source)

		case constant.SAVED_AS_DRAFT:
			if req.UserID != "" {
				query = fmt.Sprintf(" AND tt.draft_created_by= '%s' ", req.UserID)
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

		if err = r.NewKmb.Raw(fmt.Sprintf(`
		SELECT
		COUNT(tt.ProspectID) AS totalRow
		FROM
		(
			SELECT
			cb.BranchID,
			tm.ProspectID,
			tm.lob,
			tm.created_at,
			tst.activity,
			tst.source_decision,
			tst.decision,
			tcd.decision as decision_ca,
			tdd.created_by AS draft_created_by,
			scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
			scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName
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
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			created_at
		  FROM
			trx_ca_decision WITH (nolock)
		) tcd ON tm.ProspectID = tcd.ProspectID
		LEFT JOIN (
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
				  max(created_at)
				from
				  trx_draft_ca_decision WITH (NOLOCK)
				WHERE
				  ProspectID = x.ProspectID
			  )
		) tdd ON tm.ProspectID = tdd.ProspectID
		) AS tt %s AND tt.source_decision<>'%s'`, filter, constant.PRESCREENING)).Scan(&row).Error; err != nil {
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
		tst.reason,
		tcd.decision as decision_ca,
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
		tdb.BiroSpouseResult

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
		LEFT JOIN (
		  SELECT
			ProspectID,
			decision,
			note,
			created_at
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
		LEFT JOIN (
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
				max(created_at)
			  from
				trx_draft_ca_decision WITH (NOLOCK)
			  WHERE
				ProspectID = x.ProspectID
			)
		) tdd ON tm.ProspectID = tdd.ProspectID
	) AS tt %s AND tt.source_decision<>'%s' ORDER BY tt.created_at DESC %s`, filter, constant.PRESCREENING, filterPaginate)).Scan(&data).Error; err != nil {
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

		// update trx_status
		result := tx.Model(&draft).Where("ProspectID = ?", draft.ProspectID).Updates(draft)

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

func (r repoHandler) GetHistoryProcess(prospectID string) (detail []entity.TrxDetail, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw(`SELECT
			CASE
			 WHEN td.source_decision = 'PSI' THEN 'PRE SCREENING'
			 WHEN td.source_decision = 'DCK' THEN 'DUPLICATION CHECKING'
			 WHEN td.source_decision = 'DCP'
			 OR td.source_decision = 'ARI'
			 OR td.source_decision = 'KTP' THEN 'EKYC'
			 WHEN td.source_decision = 'PBK' THEN 'PEFINDO'
			 WHEN td.source_decision = 'SCS' THEN 'SCOREPRO'
			 WHEN td.source_decision = 'DSR' THEN 'DSR'
			 WHEN td.source_decision = 'CRA' THEN 'CREDIT ANALYSIS'
			 WHEN td.source_decision = 'CBM'
			  OR td.source_decision = 'DRM'
			  OR td.source_decision = 'GMO'
			  OR td.source_decision = 'COM'
			  OR td.source_decision = 'GMC'
			  OR td.source_decision = 'UCC' THEN 'CREDIT COMMITEE'
			 ELSE '-'
			END AS source_decision,
			CASE
			 WHEN td.decision = 'PAS' THEN 'PASS'
			 WHEN td.decision = 'REJ' THEN 'REJECT'
			 WHEN td.decision = 'CAN' THEN 'CANCEL'
			 WHEN td.decision = 'CPR' THEN 'CREDIT PROCESS'
			 ELSE '-'
			END AS decision,
			ap.reason AS info,
			td.created_at
		FROM
			trx_details td WITH (nolock)
			LEFT JOIN app_rules ap ON ap.rule_code = td.rule_code
		WHERE td.ProspectID = ? AND td.source_decision IN('PSI','DCK','DCP','ARI','KTP','PBK','SCS','DSR','CRA','CBM','DRM','GMO','COM','GMC','UCC')
		AND td.decision <> 'CTG' ORDER BY td.created_at ASC`, prospectID).Scan(&detail).Error; err != nil {

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
				filterBranch += "WHERE tt.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += "WHERE tt.BranchID = '" + req.BranchID + "'"
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	filter = filterBranch

	search := req.Search

	if search != "" {
		query = fmt.Sprintf("WHERE (tt.ProspectID LIKE '%%%s%%' OR tt.IDNumber LIKE '%%%s%%' OR tt.LegalName LIKE '%%%s%%')", search, search, search)
	}

	if filter == "" {
		filter = query
	} else {
		filter = filterBranch + fmt.Sprintf(" AND (tt.ProspectID LIKE '%%%s%%' OR tt.IDNumber LIKE '%%%s%%' OR tt.LegalName LIKE '%%%s%%')", search, search, search)
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
		(
			SELECT
			cb.BranchID,
			tm.ProspectID,
			tm.lob,
			tm.created_at,
			tst.activity,
			tst.source_decision,
			tst.decision,
			scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
			scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName
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
		) AS tt %s`, filter)).Scan(&row).Error; err != nil {
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
		  WHEN tst.decision='APR' THEN 'Approve'
		  WHEN tst.decision='REJ' THEN 'Reject'
		  WHEN tst.decision='CAN' THEN 'Cancel'
		  ELSE '-'
		END AS FinalStatus,
		CASE
		  WHEN tps.ProspectID IS NOT NULL
		  AND tst.status_process='ONP' THEN 1
		  ELSE 0
		END AS ActionReturn,
		CASE
		  WHEN tst.status_process='FIN'
		  AND tst.activity='STOP'
		  AND tst.decision='REJ' OR tst.decision='CAN' THEN 0
		  ELSE 1
		END AS ActionCancel,
		CASE
		  WHEN tcd.decision='CAN' THEN 0
		  ELSE 1
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
		LEFT JOIN trx_final_approval tfa WITH (nolock) ON tm.ProspectID = tfa.ProspectID
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
	) AS tt %s ORDER BY tt.created_at DESC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) ProcessTransaction(trxCaDecision entity.TrxCaDecision, trxHistoryApproval entity.TrxHistoryApprovalScheme, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail) (err error) {

	trxCaDecision.CreatedAt = time.Now()
	trxStatus.CreatedAt = time.Now()
	trxDetail.CreatedAt = time.Now()

	return r.NewKmb.Transaction(func(tx *gorm.DB) error {

		// trx_ca_decision
		result := tx.Model(&trxCaDecision).Where("ProspectID = ?", trxCaDecision.ProspectID).Updates(trxCaDecision)

		if result.RowsAffected == 0 {
			// record not found...
			if err = tx.Create(&trxCaDecision).Error; err != nil {
				return err
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

		// trx_history_approval_scheme
		if err := tx.Create(&trxHistoryApproval).Error; err != nil {
			return err
		}

		// trx_draft_ca_decision
		var draft entity.TrxDraftCaDecision
		if err := tx.Where("ProspectID = ?", trxCaDecision.ProspectID).Delete(&draft).Error; err != nil {
			return err
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

		// insert trx_details
		if err := tx.Create(&trxDetail).Error; err != nil {
			return err
		}

		// delete trx_prescreening
		var prescreening entity.TrxPrescreening
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&prescreening).Error; err != nil {
			return err
		}

		// delete trx_draft_ca_decision
		var draft entity.TrxDraftCaDecision
		if err := tx.Where("ProspectID = ?", prospectID).Delete(&draft).Error; err != nil {
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
				filterBranch += "WHERE tt.BranchID IN (" + extractBranchIDUser + ")"
			}
		} else {
			filterBranch += "WHERE tt.BranchID = '" + req.BranchID + "'"
		}
	} else {
		filterBranch = utils.GenerateBranchFilter(req.BranchID)
	}

	filter = utils.GenerateFilter(req.Search, filterBranch, rangeDays)

	// Filter By
	if req.Filter != "" {
		var (
			decision string
			activity string
		)
		switch req.Filter {
		case constant.DECISION_APPROVE:
			decision = constant.DB_DECISION_APR
			query = fmt.Sprintf(" AND tt.decision= '%s'", decision)

		case constant.DECISION_REJECT:
			decision = constant.DB_DECISION_REJECT
			query = fmt.Sprintf(" AND tt.decision= '%s'", decision)

		case constant.DECISION_CANCEL:
			decision = constant.DB_DECISION_CANCEL
			query = fmt.Sprintf(" AND tt.decision= '%s'", decision)

		case constant.NEED_DECISION:
			activity = constant.ACTIVITY_UNPROCESS
			decision = constant.DB_DECISION_CREDIT_PROCESS
			source := alias
			query = fmt.Sprintf(" AND tt.activity= '%s' AND tt.decision= '%s' AND tt.source_decision = '%s'", activity, decision, source)
		}
	}

	if req.Alias == constant.DB_DECISION_BRANCH_MANAGER {
		query += fmt.Sprintf(" AND (tt.next_step = '%s' OR tt.decision_by = '%s')", req.Alias, constant.SYSTEM_CREATED)
	} else {
		query += fmt.Sprintf(" AND tt.next_step= '%s'", req.Alias)
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
		(
			SELECT
			cb.BranchID,
			tm.ProspectID,
			tm.lob,
			tm.created_at,
			tst.activity,
			tst.source_decision,
			tst.decision,
			tcd.final_approval,
			tcd.decision_by,
			has.next_step,
			tcd.decision as decision_ca,
			scp.dbo.DEC_B64('SEC', tcp.IDNumber) AS IDNumber,
			scp.dbo.DEC_B64('SEC', tcp.LegalName) AS LegalName
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
		LEFT JOIN trx_history_approval_scheme has WITH (nolock) ON has.ProspectID = tm.ProspectID
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
		) AS tt %s`, filter)).Scan(&row).Error; err != nil {
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
		tst.reason,
		tcd.decision as decision_ca,
		tcd.decision_by,
		tcd.final_approval,
		has.next_step,
		CASE
		  WHEN tcd.final_approval='%s' THEN 1
		  ELSE 0
		END AS is_last_approval,
		CASE
		  WHEN tcd.decision = 'CAN' THEN tcd.decision
		  WHEN tcd.decision IS NOT NULL THEN tfa.decision
		  WHEN tst.decision = 'REJ' THEN 'REJECT'
		  ELSE NULL
		END AS ca_decision,
		tcd.note AS ca_note,
		CASE
		  WHEN tcd.decision='CAN' THEN 0
		  ELSE 1
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
		tdb.BiroSpouseResult

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
		LEFT JOIN trx_history_approval_scheme has WITH (nolock) ON has.ProspectID = tm.ProspectID
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
	) AS tt %s ORDER BY tt.created_at DESC %s`, alias, alias, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) SubmitApproval(req request.ReqSubmitApproval, trxStatus entity.TrxStatus, trxDetail entity.TrxDetail, approval response.RespApprovalScheme) (err error) {

	trxStatus.CreatedAt = time.Now()
	trxDetail.CreatedAt = time.Now()

	return r.NewKmb.Transaction(func(tx *gorm.DB) error {

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

		// trx_history_approval_scheme
		if err := tx.Create(&trxHistoryApproval).Error; err != nil {
			return err
		}

		if approval.IsFinal {
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
		}

		return nil
	})
}
