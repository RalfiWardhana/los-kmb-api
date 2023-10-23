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
	)

	rangeDays := os.Getenv("DEFAULT_RANGE_DAYS")

	filterBranch = utils.GenerateBranchFilter(req.BranchID)

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

func (r repoHandler) GetHistoryApproval(prospectID string) (history []entity.TrxHistoryApprovalScheme, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.NewKmb.Raw("SELECT decision, decision_by, next_final_approval_flag, need_escalation, source_decision, next_step, created_at FROM trx_history_approval_scheme WITH (nolock) WHERE ProspectID = ? ORDER BY created_at DESC", prospectID).Scan(&history).Error; err != nil {

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
	)

	rangeDays := os.Getenv("DEFAULT_RANGE_DAYS")

	filterBranch = utils.GenerateBranchFilter(req.BranchID)

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
			query = fmt.Sprintf(" AND tt.activity= '%s' AND tt.decision= '%s'", activity, decision)

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
		LEFT JOIN trx_prescreening tps WITH (nolock) ON tm.ProspectID = tps.ProspectID
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
		  WHEN tcd.created_at IS NOT NULL
		  AND tfa.created_at IS NULL THEN tcd.created_at
		  WHEN tfa.created_at IS NOT NULL THEN tfa.created_at
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
		tca.LegalAddress,
		tca.LegalRTRW,
		tca.LegalKelurahan,
		tca.LegalKecamatan,
		tca.LegalZipcode,
		tca.LegalCity,
		scp.dbo.DEC_B64('SEC', tcp.MobilePhone) AS MobilePhone,
		scp.dbo.DEC_B64('SEC', tcp.Email) AS Email,
		edu.value AS Education,
		mst.value AS MaritalStatus,
		tcp.NumOfDependence,
		tca.ResidenceAddress,
		tca.ResidenceRTRW,
		tca.ResidenceKelurahan,
		tca.ResidenceKecamatan,
		tca.ResidenceZipcode,
		tca.ResidenceCity,
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
		tca.CompanyAreaPhone,
		tca.CompanyPhone,
		tcp.ExtCompanyPhone,
		tca.CompanyAddress,
		tca.CompanyRTRW,
		tca.CompanyKelurahan,
		tca.CompanyKecamatan,
		tca.CompanyZipcode,
		tca.CompanyCity,
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
	    tca.EmergencyAddress,
		tca.EmergencyRTRW,
		tca.EmergencyKelurahan,
		tca.EmergencyKecamatan,
		tca.EmergencyZipcode,
		tca.EmergencyCity,
		tca.EmergencyAreaPhone,
		tca.EmergencyPhone,
		tce.IndustryTypeID,
		tak.ScsDate,
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
			SELECT ProspectID,
			MAX(Case [Type] When 'LEGAL' Then scp.dbo.DEC_B64('SEC', Address) End) LegalAddress,
			MAX(Case [Type] When 'LEGAL' Then CONCAT(RT, '/', RW) End) LegalRTRW,
			MAX(Case [Type] When 'LEGAL' Then Kelurahan End) LegalKelurahan,
			MAX(Case [Type] When 'LEGAL' Then Kecamatan End) LegalKecamatan,
			MAX(Case [Type] When 'LEGAL' Then ZipCode End) LegalZipCode,
			MAX(Case [Type] When 'LEGAL' Then City End) LegalCity,
			MAX(Case [Type] When 'LEGAL' Then Phone End) LegalPhone,
			MAX(Case [Type] When 'LEGAL' Then AreaPhone End) LegalAreaPhone,

			MAX(Case [Type] When 'COMPANY' Then scp.dbo.DEC_B64('SEC', Address) End) CompanyAddress,
			MAX(Case [Type] When 'COMPANY' Then CONCAT(RT, '/', RW) End) CompanyRTRW,
			MAX(Case [Type] When 'COMPANY' Then Kelurahan End) CompanyKelurahan,
			MAX(Case [Type] When 'COMPANY' Then Kecamatan End) CompanyKecamatan,
			MAX(Case [Type] When 'COMPANY' Then ZipCode End) CompanyZipCode,
			MAX(Case [Type] When 'COMPANY' Then City End) CompanyCity,
			MAX(Case [Type] When 'COMPANY' Then Phone End) CompanyPhone,
			MAX(Case [Type] When 'COMPANY' Then AreaPhone End) CompanyAreaPhone,

			MAX(Case [Type] When 'RESIDENCE' Then scp.dbo.DEC_B64('SEC', Address) End) ResidenceAddress,
			MAX(Case [Type] When 'RESIDENCE' Then CONCAT(RT, '/', RW) End) ResidenceRTRW,
			MAX(Case [Type] When 'RESIDENCE' Then Kelurahan End) ResidenceKelurahan,
			MAX(Case [Type] When 'RESIDENCE' Then Kecamatan End) ResidenceKecamatan,
			MAX(Case [Type] When 'RESIDENCE' Then ZipCode End) ResidenceZipCode,
			MAX(Case [Type] When 'RESIDENCE' Then City End) ResidenceCity,
			MAX(Case [Type] When 'RESIDENCE' Then Phone End) ResidencePhone,
			MAX(Case [Type] When 'RESIDENCE' Then AreaPhone End) ResidenceAreaPhone,

			MAX(Case [Type] When 'EMERGENCY' Then scp.dbo.DEC_B64('SEC', Address) End) EmergencyAddress,
			MAX(Case [Type] When 'EMERGENCY' Then CONCAT(RT, '/', RW) End) EmergencyRTRW,
			MAX(Case [Type] When 'EMERGENCY' Then Kelurahan End) EmergencyKelurahan,
			MAX(Case [Type] When 'EMERGENCY' Then Kecamatan End) EmergencyKecamatan,
			MAX(Case [Type] When 'EMERGENCY' Then ZipCode End) EmergencyZipcode,
			MAX(Case [Type] When 'EMERGENCY' Then City End) EmergencyCity,
			MAX(Case [Type] When 'EMERGENCY' Then Phone End) EmergencyPhone,
			MAX(Case [Type] When 'EMERGENCY' Then AreaPhone End) EmergencyAreaPhone
			FROM trx_customer_address
			GROUP BY ProspectID
		) tca ON tm.ProspectID = tca.ProspectID 

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
	) AS tt %s ORDER BY tt.created_at DESC %s`, filter, filterPaginate)).Scan(&data).Error; err != nil {
		return
	}

	if len(data) == 0 {
		return data, 0, fmt.Errorf(constant.RECORD_NOT_FOUND)
	}
	return
}

func (r repoHandler) SaveDraftData(draft entity.TrxDraftCaDecision) (err error) {
	var x sql.TxOptions

	draft.CreatedAt = DtmRequest

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	// update trx_status
	result := db.Model(&draft).Where("ProspectID = ?", draft.ProspectID).Updates(draft)

	if err = result.Error; err != nil {
		return
	}

	if result.RowsAffected == 0 {
		// record not found will be create draft
		if err = db.Create(&draft).Error; err != nil {
			return
		}
	}

	return
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

func (r repoHandler) SaveCADecionData(trxCaDecision entity.TrxCaDecision) (err error) {
	var x sql.TxOptions

	trxCaDecision.CreatedAt = DtmRequest

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	// insert trx_ca_decision
	if err = db.Create(&trxCaDecision).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) UpdateTrxStatus(trxStatus entity.TrxStatus) (err error) {
	var x sql.TxOptions

	trxStatus.CreatedAt = DtmRequest

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	// update trx_status
	result := db.Model(&trxStatus).Where("ProspectID = ?", trxStatus.ProspectID).Updates(trxStatus)

	if err = result.Error; err != nil {
		return
	}

	if result.RowsAffected == 0 {
		// record not found...
		if err = db.Create(&trxStatus).Error; err != nil {
			return
		}
	}

	return
}

func (r repoHandler) SaveTrxDetail(trxDetail entity.TrxDetail) (err error) {
	var x sql.TxOptions

	trxDetail.CreatedAt = DtmRequest

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	// insert trx_ca_decision
	if err = db.Create(&trxDetail).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) DeleteDraft(prospectID string) (err error) {
	var x sql.TxOptions

	var txrDraft entity.TrxDraftCaDecision

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.NewKmb.BeginTx(ctx, &x)
	defer db.Commit()

	// insert trx_ca_decision
	result := db.Where("ProspectID = ?", prospectID).Delete(&txrDraft)

	if err = result.Error; err != nil {
		return
	}

	return
}
