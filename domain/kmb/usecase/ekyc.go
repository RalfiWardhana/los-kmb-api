package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
)

func (u usecase) Asliri(ctx context.Context, req request.Metrics, cb_found bool, accessToken string) (data response.Ekyc, err error) {

	var (
		resp        *resty.Response
		selfie, ktp string
	)

	for _, photo := range req.CustomerPhoto {

		if photo.ID == constant.TAG_KTP_PHOTO {
			ktp = photo.Url
		} else if photo.ID == constant.TAG_SELFIE_PHOTO {
			selfie = photo.Url
		}
	}

	param, _ := json.Marshal(map[string]interface{}{
		"transaction_id": req.Transaction.ProspectID,
		"id_number":      req.CustomerPersonal.IDNumber,
		"name":           req.CustomerPersonal.LegalName,
		"birth_place":    req.CustomerPersonal.BirthPlace,
		"birth_date":     req.CustomerPersonal.BirthDate,
		"ktp_photo":      ktp,
		"selfie_photo":   selfie,
		"request_id":     uuid.New().String(),
	})

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	resp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("ASLIRI_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.Transaction.ProspectID, accessToken)

	if err != nil {
		err = errors.New("upstream_service_timeout - Call Asliri")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New("upstream_service_error - Call Asliri")
		return
	}

	var asliri response.AsliriIntegrator

	jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &asliri)

	var name, pdob, asliriSelfie int

	if asliri.Name != nil {
		name = int(asliri.Name.(float64))
	}

	if asliri.PDOB != nil {
		pdob = int(asliri.PDOB.(float64))
	}

	if asliri.SelfiePhoto != nil {
		asliriSelfie = int(asliri.SelfiePhoto.(float64))
	}

	var ekycInfo = response.InfoEkyc{
		Asliri: asliri,
	}

	infoAsliri, _ := json.Marshal(ekycInfo)

	var asliriConfig entity.AsliriConfig

	config, err := u.repository.GetConfig("asliri", "KMB", "asliri_tier2_parameter")
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Asliri Config Error")
		return
	}

	json.Unmarshal([]byte(config.Value), &asliriConfig)

	if asliri.NotFound {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_ASLIRI_NOT_FOUND
		data.Reason = constant.REASON_EKYC_INVALID
		data.Source = constant.ASLIRI
		data.Info = string(infoAsliri)
		return

	}

	if name < asliriConfig.Data.KMB.AsliriName || pdob < asliriConfig.Data.KMB.AsliriPDOB {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_ASLIRI_NAME
		data.Reason = constant.REASON_EKYC_INVALID
		data.Source = constant.ASLIRI
		data.Info = string(infoAsliri)
		return

	}

	if asliriSelfie < asliriConfig.Data.KMB.AsliriPhoto {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_ASLIRI_SELFIE
		data.Reason = constant.REASON_EKYC_INVALID
		data.Source = constant.ASLIRI
		data.Info = string(infoAsliri)
		return
	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_PASS_ASLIRI
	data.Reason = constant.REASON_EKYC_VALID
	data.Source = constant.ASLIRI
	data.Info = string(infoAsliri)

	return
}

func (u usecase) Ktp(ctx context.Context, req request.Metrics, cb_found bool, accessToken string) (data response.Ekyc, err error) {

	var (
		resp *resty.Response
	)
	param, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"birth_date": req.CustomerPersonal.BirthDate,
			"gender":     req.CustomerPersonal.Gender,
			"id_number":  req.CustomerPersonal.IDNumber,
			"is_pefindo": cb_found,
			"request_id": uuid.New().String(),
		},
	})

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	resp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("KTP_VALIDATOR_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.Transaction.ProspectID, accessToken)

	if err != nil {
		err = errors.New("upstream_service_timeout - Call KTP Validator")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New("upstream_service_error - Call KTP Validator")
		return
	}

	var ktp response.KtpValidator

	jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &ktp)

	var ekycInfo = response.InfoEkyc{
		Ktp: ktp,
	}

	infoKtp, _ := json.Marshal(ekycInfo)

	data.Result = ktp.Result
	data.Code = ktp.Code
	data.Reason = ktp.Reason
	data.Source = constant.KTP
	data.Info = string(infoKtp)

	return
}

func CheckEKYC(data response.VerifyDataIntegratorResponse) (code, reason, decision string) {

	if data.IsValid {
		if strings.Contains(data.Nik, "Tidak Sesuai") {
			return constant.CODE_VERIFICATION_REJECT_EKYC, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}
		if strings.Contains(data.TglLhr, "Tidak Sesuai") {
			return constant.CODE_VERIFICATION_REJECT_EKYC, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}
		if strings.Contains(data.NamaLgkp, "Tidak Sesuai") {
			return constant.CODE_VERIFICATION_REJECT_EKYC, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}
		if strings.Contains(data.NamaLgkpIbu, "Tidak Sesuai") {
			return constant.CODE_VERIFICATION_REJECT_EKYC, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}

		return constant.CODE_VERIFICATION_PASS_EKYC, "EKYC Sesuai", constant.DECISION_PASS

	} else {

		switch *data.Reason {
		case constant.CUSTOMER_MENINGGAL:
			return constant.CODE_VERIFICATION_REJECT_MENINGGAL, *data.Reason, constant.DECISION_REJECT

		case constant.DATA_GANDA:
			return constant.CODE_VERIFICATION_REJECT_DATA_GANDA, *data.Reason, constant.DECISION_REJECT

		case constant.DATA_INACTIVE:
			return constant.CODE_VERIFICATION_REJECT_INACTIVE, *data.Reason, constant.DECISION_REJECT

		case constant.DATA_NOT_FOUND:
			return constant.CODE_VERIFICATION_REJECT_NOT_FOUND, *data.Reason, constant.DECISION_REJECT
		}

	}

	return
}

func CheckThreshold(data response.FaceRecognitionIntegratorData) (code, reason, decision string) {

	switch data.RuleCode {
	case "6020":
		return constant.CODE_FACERECOGNITION_REJECT_NIK, data.Reason, constant.DECISION_REJECT
	case "6019":
		return constant.CODE_FACERECOGNITION_REJECT_FOTO, data.Reason, constant.DECISION_REJECT
	case "6018":
		return constant.CODE_FACERECOGNITION_PASS, data.Reason, constant.DECISION_PASS
	}
	return
}
