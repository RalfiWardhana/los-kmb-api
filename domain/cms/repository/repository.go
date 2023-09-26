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
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	jsoniter "github.com/json-iterator/go"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	NewKmb    *gorm.DB
	core      *gorm.DB
	KpLosLogs *gorm.DB
}

func NewRepository(core, NewKmb, KpLosLogs *gorm.DB) interfaces.Repository {
	return &repoHandler{
		core:      core,
		NewKmb:    NewKmb,
		KpLosLogs: KpLosLogs,
	}
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

func (r repoHandler) GetCustomerPhoto(prospectID string) (photo []entity.CustomerPhoto, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT photo_id, url FROM trx_customer_photo WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&photo).Error; err != nil {

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
	)

	if req.BranchID != "" {
		arrBranch := strings.Split(req.BranchID, ",")
		if len(arrBranch) == 1 {
			// spesific branch non HO
			if arrBranch[0] != "999" {
				filterBranch = fmt.Sprintf("WHERE tt.BranchID IN ('%s')", req.BranchID)
			}
		} else {
			// multi branch
			var branch string
			for i, val := range arrBranch {
				branch = fmt.Sprintf("%s'%s'", branch, val)
				if i < len(arrBranch)-1 {
					branch = fmt.Sprintf("%s,", branch)
				}
			}

			filterBranch = fmt.Sprintf("WHERE tt.BranchID IN (%s)", branch)
		}
	}

	filter = filterBranch

	if req.Search != "" && filterBranch != "" {
		filter = filterBranch + " AND (tt.ProspectID LIKE '%" + req.Search + "%' OR tt.IDNumber LIKE '%" + req.Search + "%' OR tt.LegalName LIKE '%" + req.Search + "%')"
	} else if req.Search != "" {
		filter = "WHERE (tt.ProspectID LIKE '%" + req.Search + "%' OR tt.IDNumber LIKE '%" + req.Search + "%' OR tt.LegalName LIKE '%" + req.Search + "%')"
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

func (r repoHandler) GetStatusPrescreening(prospectID string) (status entity.TrxStatus, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT activity, source_decision FROM trx_status WITH (nolock) WHERE ProspectID = ?", prospectID).Scan(&status).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SavePrescreening(prescreening entity.TrxPrescreening, detail entity.TrxDetail, status entity.TrxStatus) (err error) {
	var x sql.TxOptions

	prescreening.CreatedAt = DtmRequest
	detail.CreatedAt = DtmRequest
	status.CreatedAt = DtmRequest

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	// update trx_status
	result := db.Model(&status).Where("ProspectID = ?", status.ProspectID).Updates(status)

	if err = result.Error; err != nil {
		return
	}

	if result.RowsAffected == 0 {
		// record not found...
		if err = db.Create(&status).Error; err != nil {
			return
		}
	}

	// insert trx_details
	if err = db.Create(&detail).Error; err != nil {
		return
	}

	// insert trx_prescreening
	if err = db.Create(&prescreening).Error; err != nil {
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
