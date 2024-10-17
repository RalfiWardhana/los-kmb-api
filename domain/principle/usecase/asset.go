package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/principle/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common/platformevent"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type (
	multiUsecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		usecase    interfaces.Usecase
		producer   platformevent.PlatformEventInterface
	}
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		producer   platformevent.PlatformEventInterface
	}
)

func NewMultiUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, producer platformevent.PlatformEventInterface, usecase interfaces.Usecase) interfaces.MultiUsecase {

	return &multiUsecase{
		usecase:    usecase,
		repository: repository,
		httpclient: httpclient,
		producer:   producer,
	}
}

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, producer platformevent.PlatformEventInterface) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
		producer:   producer,
	}
}

func (u usecase) CheckNokaNosin(ctx context.Context, r request.PrincipleAsset) (data response.UsecaseApi, err error) {

	var (
		mdmChassis     response.AgreementChassisNumber
		cmoID, cmoName string
	)

	hitChassisNumber, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("AGREEMENT_OF_CHASSIS_NUMBER_URL")+r.NoChassis, nil, map[string]string{}, constant.METHOD_GET, true, 6, 60, r.ProspectID, middlewares.UserInfoData.AccessToken)
	if err != nil {
		return
	}

	if hitChassisNumber.StatusCode() != 200 {
		return
	}

	err = json.Unmarshal([]byte(jsoniter.Get(hitChassisNumber.Body(), "data").ToString()), &mdmChassis)

	if err != nil {
		return
	}

	if mdmChassis.IsRegistered && mdmChassis.IsActive && len(mdmChassis.IDNumber) > 0 {
		listNikKonsumenDanPasangan := make([]string, 0)

		listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, r.IDNumber)
		if r.SpouseIDNumber != "" {
			listNikKonsumenDanPasangan = append(listNikKonsumenDanPasangan, r.SpouseIDNumber)
		}

		if !utils.Contains(listNikKonsumenDanPasangan, mdmChassis.IDNumber) {
			data.Code = constant.CODE_REJECT_CHASSIS_NUMBER
			data.Result = constant.DECISION_REJECT
			data.Reason = constant.REASON_REJECT_CHASSIS_NUMBER

		} else {
			if mdmChassis.IDNumber == r.IDNumber {
				data.Code = constant.CODE_OK_CONSUMEN_MATCH
				data.Result = constant.DECISION_PASS
				data.Reason = constant.REASON_OK_CONSUMEN_MATCH
			} else {
				data.Code = constant.CODE_REJECTION_FRAUD_POTENTIAL
				data.Result = constant.DECISION_REJECT
				data.Reason = constant.REASON_REJECTION_FRAUD_POTENTIAL
			}
		}
	} else {
		data.Code = constant.CODE_AGREEMENT_NOT_FOUND
		data.Result = constant.DECISION_PASS
		data.Reason = constant.REASON_AGREEMENT_NOT_FOUND
	}

	taxDate, _ := time.Parse(constant.FORMAT_DATE, r.TaxDate)
	stnkExpired, _ := time.Parse(constant.FORMAT_DATE, r.STNKExpiredDate)

	respCmoBranch, err := u.MDMGetMasterMappingBranchEmployee(ctx, r.ProspectID, r.BranchID, middlewares.UserInfoData.AccessToken)
	if err != nil {
		return
	}

	if len(respCmoBranch.Data) > 0 {
		cmoData := respCmoBranch.Data[0]
		cmoID = cmoData.CMOID
		cmoName = cmoData.CMOName
	}

	_ = u.repository.SavePrincipleStepOne(entity.TrxPrincipleStepOne{
		ProspectID:         r.ProspectID,
		IDNumber:           r.IDNumber,
		SpouseIDNumber:     utils.CheckEmptyString(r.SpouseIDNumber),
		ManufactureYear:    strconv.Itoa(r.ManufactureYear),
		NoChassis:          r.NoChassis,
		NoEngine:           r.NoEngine,
		BranchID:           r.BranchID,
		CMOID:              cmoID,
		CMOName:            cmoName,
		CC:                 strconv.Itoa(r.CC),
		TaxDate:            taxDate,
		STNKExpiredDate:    stnkExpired,
		OwnerAsset:         r.OwnerAsset,
		LicensePlate:       r.LicensePlate,
		Color:              r.Color,
		Brand:              r.Brand,
		ResidenceAddress:   r.ResidenceAddress,
		ResidenceRT:        r.ResidenceRT,
		ResidenceRW:        r.ResidenceRW,
		ResidenceProvince:  r.ResidenceProvince,
		ResidenceCity:      r.ResidenceCity,
		ResidenceKecamatan: r.ResidenceKecamatan,
		ResidenceKelurahan: r.ResidenceKelurahan,
		ResidenceZipCode:   r.ResidenceZipCode,
		ResidenceAreaPhone: r.ResidenceAreaPhone,
		ResidencePhone:     r.ResidencePhone,
		HomeStatus:         r.HomeStatus,
		StaySinceYear:      r.StaySinceYear,
		StaySinceMonth:     r.StaySinceMonth,
		Decision:           data.Result,
		RuleCode:           data.Code,
		Reason:             data.Reason,
		AssetCode:          r.AssetCode,
		STNKPhoto:          r.STNKPhoto,
		KPMID:              r.KPMID,
	})

	//masking reason
	statusCode := constant.PRINCIPLE_STATUS_ASSET_APPROVE
	data.Reason = "Verifikasi data aset berhasil"
	if data.Result == constant.DECISION_REJECT {
		statusCode = constant.PRINCIPLE_STATUS_ASSET_REJECT
		data.Reason = "Data STNK tidak lolos verifikasi"
	}

	u.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_PRINCIPLE, constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE, r.ProspectID, utils.StructToMap(request.Update2wPrincipleTransaction{
		KpmID:         r.KPMID,
		OrderID:       r.ProspectID,
		Source:        3,
		StatusCode:    statusCode,
		ProductName:   r.AssetCode,
		BranchCode:    r.BranchID,
		AssetTypeCode: constant.KPM_ASSET_TYPE_CODE_MOTOR,
	}), 0)

	return
}

func (u usecase) MDMGetMasterMappingBranchEmployee(ctx context.Context, prospectID, branchID, accessToken string) (data response.MDMMasterMappingBranchEmployeeResponse, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	headerMDM := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": accessToken,
	}

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MASTER_MAPPING_BRANCH_EMPLOYEE_URL")+"?lob_id="+strconv.Itoa(constant.LOBID_KMB)+"&branch_id="+branchID, nil, headerMDM, constant.METHOD_GET, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping Branch Employee Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &data); err != nil {
			return
		}
	}

	return
}
