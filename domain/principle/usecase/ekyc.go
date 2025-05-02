package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (u usecase) Dukcapil(ctx context.Context, r request.PrinciplePemohon, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error) {

	var (
		codeVD, _, decisionVD, decisionFR string
		infoDukcapil                      response.InfoEkyc
		verify                            response.VerifyDataIntegratorResponse
		face                              response.FaceRecognitionIntegratorData
		thresholdDukcapil                 entity.ConfigThresholdDukcapil
		timeout                           int
		statusVD, statusFR                string
		endpointVd, endpointFr            string
		thresholdFr                       float64
	)

	config, err := u.repository.GetConfig("ekyc", constant.LOB_KMB_OFF, "threshold_ekyc")

	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Get Dukcapil Config Error")
		return
	}

	if err = json.Unmarshal([]byte(config.Value), &thresholdDukcapil); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data config threshold ekyc")
		return
	}

	timeout, err = strconv.Atoi(os.Getenv("DUKCAPIL_TIMEOUT"))
	if err != nil {
		timeout, _ = strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	}

	// Verify Data
	paramVd, _ := json.Marshal(map[string]interface{}{
		"address":             r.LegalAddress,
		"birth_date":          r.BirthDate,
		"birth_place":         r.BirthPlace,
		"city":                r.LegalCity,
		"gender":              r.Gender,
		"marital_status":      r.MaritalStatus,
		"id_number":           r.IDNumber,
		"kabupaten":           r.LegalCity,
		"kecamatan":           r.LegalKecamatan,
		"kelurahan":           r.LegalKelurahan,
		"legal_name":          r.LegalName,
		"profession_id":       r.ProfessionID,
		"province":            r.LegalProvince,
		"rt":                  r.LegalRT,
		"rw":                  r.LegalRW,
		"surgate_mother_name": r.SurgateMotherName,
		"threshold":           "0",
		"transaction_id":      r.ProspectID,
	})

	serviceVD := thresholdDukcapil.Data.VerifyData.Service

	switch serviceVD {
	case constant.SERVICE_IZIDATA:
		endpointVd = os.Getenv("IZIDATA_VD_URL")
	case constant.SERVICE_DUKCAPIL:
		endpointVd = os.Getenv("DUKCAPIL_VD_URL")
	default:
		endpointVd = os.Getenv("DUKCAPIL_VD_URL")
	}

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, endpointVd, paramVd, map[string]string{}, constant.METHOD_POST, true, 2, timeout, r.ProspectID, accessToken)

	infoDukcapil.VdService = serviceVD

	if resp.StatusCode() == 504 || resp.StatusCode() == 502 {
		statusVD = constant.EKYC_RTO

		infoDukcapil.VdError = "Request Timed Out"
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 504 && resp.StatusCode() != 502 {
		statusVD = constant.EKYC_NOT_CHECK

		var responseIntegrator response.ApiResponse
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &responseIntegrator); err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data response error vd ekyc")
			return
		}

		infoDukcapil.VdError = responseIntegrator.Message
	}

	if err == nil && resp.StatusCode() == 200 {

		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &verify); err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data response vd ekyc")
			return
		}

		if serviceVD == constant.SERVICE_IZIDATA {
			codeVD, _, decisionVD = checkEKYCIzidata(verify, thresholdDukcapil)
		} else {
			codeVD, _, decisionVD = checkEKYCDukcapil(verify, thresholdDukcapil)
		}

		infoDukcapil.Vd = verify

		if decisionVD == constant.DECISION_REJECT {
			statusVD = constant.DECISION_REJECT
		} else {
			statusVD = constant.DECISION_PASS
		}
	}

	resultDukcapilVD, err := u.repository.GetMappingDukcapilVD(statusVD, reqMetricsEkyc.CustomerStatus, reqMetricsEkyc.CustomerSegment, verify.IsValid)
	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Verify Dukcapil Error")
		return
	}

	if resultDukcapilVD.Decision == constant.DECISION_REJECT {
		data.Result = resultDukcapilVD.Decision
		data.Code = codeVD
		switch serviceVD {
		case constant.SERVICE_IZIDATA:
			data.Reason = constant.REASON_IZIDATA_INVALID
		case constant.SERVICE_DUKCAPIL:
			data.Reason = constant.REASON_EKYC_INVALID
		default:
			data.Reason = constant.REASON_EKYC_INVALID
		}
		data.Source = constant.SOURCE_DECISION_DUKCAPIL

		info, _ := json.Marshal(infoDukcapil)
		data.Info = string(info)
		return
	}

	statusVD = resultDukcapilVD.Decision

	if resultDukcapilVD.Decision == constant.EKYC_BYPASS {
		statusVD = constant.DECISION_PASS
	}

	serviceFR := thresholdDukcapil.Data.FaceRecognition.Service

	switch serviceFR {
	case constant.SERVICE_IZIDATA:
		endpointFr = os.Getenv("IZIDATA_FR_URL")
		thresholdFr = thresholdDukcapil.Data.FaceRecognition.FRIziData.Threshold
	case constant.SERVICE_DUKCAPIL:
		endpointFr = os.Getenv("DUKCAPIL_FR_URL")
		thresholdFr = thresholdDukcapil.Data.FaceRecognition.FRDukcapil.Threshold
	default:
		endpointFr = os.Getenv("DUKCAPIL_FR_URL")
		parseThreshold, _ := strconv.ParseFloat(strings.TrimSpace(os.Getenv("THRESHOLD_FR")), 64)
		thresholdFr = parseThreshold
	}

	//Face Recog
	paramFr, _ := json.Marshal(map[string]interface{}{
		"id_number":      r.IDNumber,
		"selfie_image":   r.SelfiePhoto,
		"threshold":      fmt.Sprintf("%.1f", thresholdFr),
		"transaction_id": r.ProspectID,
	})

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, endpointFr, paramFr, map[string]string{}, constant.METHOD_POST, true, 2, timeout, r.ProspectID, accessToken)

	infoDukcapil.FrService = serviceFR

	if resp.StatusCode() == 504 || resp.StatusCode() == 502 {
		statusFR = constant.EKYC_RTO
		infoDukcapil.FrError = "Request Timed Out"
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 504 && resp.StatusCode() != 502 {
		statusFR = constant.EKYC_NOT_CHECK
		var responseIntegrator response.ApiResponse
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &responseIntegrator); err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data response error fr ekyc")
			return
		}

		infoDukcapil.FrError = responseIntegrator.Message
	}

	if err == nil && resp.StatusCode() == 200 {

		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &face); err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data response fr ekyc")
			return
		}

		if serviceFR == constant.SERVICE_IZIDATA {
			_, _, decisionFR = checkRuleCodeIzidata(face)
		} else {
			_, _, decisionFR = checkRuleCodeDukcapil(face)
		}

		infoDukcapil.Fr = face

		statusFR = decisionFR
		data.Similiarity = face.MatchScore
	}

	resultDukcapil, err := u.repository.GetMappingDukcapil(statusVD, statusFR, reqMetricsEkyc.CustomerStatus, reqMetricsEkyc.CustomerSegment)
	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Result Dukcapil Error")
		return
	}

	switch resultDukcapil.Decision {
	case constant.TYPE_CONTINGENCY:
		data.Result = resultDukcapil.Decision
		data.Code = resultDukcapil.RuleCode
		data.Reason = constant.TYPE_CONTINGENCY
		data.Source = constant.SOURCE_DECISION_DUKCAPIL
		info, _ := json.Marshal(infoDukcapil)
		data.Info = string(info)
		err = fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY)
		return
	case constant.DECISION_REJECT:
		data.Result = resultDukcapil.Decision
		data.Code = resultDukcapil.RuleCode
		data.Reason = constant.REASON_EKYC_INVALID
		data.Source = constant.SOURCE_DECISION_DUKCAPIL
		info, _ := json.Marshal(infoDukcapil)
		data.Info = string(info)
		return
	}

	data.Result = constant.DECISION_PASS
	data.Code = resultDukcapil.RuleCode
	data.Reason = constant.REASON_EKYC_VALID
	data.Source = constant.SOURCE_DECISION_DUKCAPIL
	info, _ := json.Marshal(infoDukcapil)
	data.Info = string(info)

	return
}

func (u usecase) Asliri(ctx context.Context, r request.PrinciplePemohon, accessToken string) (data response.Ekyc, err error) {

	var (
		resp *resty.Response
	)

	paramARI, _ := json.Marshal(map[string]interface{}{
		"transaction_id": r.ProspectID,
		"id_number":      r.IDNumber,
		"name":           r.LegalName,
		"birth_place":    r.BirthPlace,
		"birth_date":     r.BirthDate,
		"ktp_photo":      r.KtpPhoto,
		"selfie_photo":   r.SelfiePhoto,
		"request_id":     uuid.New().String(),
	})

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("ASLIRI_URL"), paramARI, map[string]string{}, constant.METHOD_POST, false, 0, timeout, r.ProspectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Asliri")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Asliri")
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

	config, err := u.repository.GetConfig("asliri", constant.LOB_KMB_OFF, "threshold_asliri")

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get ASLI RI Config Error")
		return
	}

	if err = json.Unmarshal([]byte(config.Value), &asliriConfig); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data config asliri")
		return
	}

	if asliri.NotFound {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_ASLIRI_NOT_FOUND
		data.Reason = constant.REASON_EKYC_INVALID
		data.Source = constant.ASLIRI
		data.Info = string(infoAsliri)
		return

	}

	data.Similiarity = asliriSelfie

	if name < asliriConfig.Data.AsliriName || pdob < asliriConfig.Data.AsliriPDOB {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_REJECT_ASLIRI_NAME
		data.Reason = constant.REASON_EKYC_INVALID
		data.Source = constant.ASLIRI
		data.Info = string(infoAsliri)
		return

	}

	if asliriSelfie < asliriConfig.Data.AsliriPhoto {
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

func (u usecase) Ktp(ctx context.Context, r request.PrinciplePemohon, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error) {

	paramKtp, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"birth_date": r.BirthDate,
			"gender":     r.Gender,
			"id_number":  r.IDNumber,
			"is_pefindo": reqMetricsEkyc.CBFound,
			"request_id": uuid.New().String(),
		},
	})

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("KTP_VALIDATOR_URL"), paramKtp, map[string]string{}, constant.METHOD_POST, false, 0, timeout, r.ProspectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call KTP Validator")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call KTP Validator")
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
	data.Source = constant.SOURCE_DECISION_KTP_VALIDATOR
	data.Info = string(infoKtp)

	if data.Result == constant.DECISION_REJECT {
		if strings.Contains(os.Getenv("NAMA_SAMA"), r.BpkbName) {
			data.Result = constant.DECISION_PASS
			data.Code = "2600"
			data.Reason = "eKYC Sesuai - No KTP Valid"
		}
		if reqMetricsEkyc.CBFound {
			data.Result = constant.DECISION_PASS
			data.Code = "2600"
			data.Reason = "eKYC Sesuai - No KTP Valid"
		}
	}

	return
}

func checkRuleCodeDukcapil(data response.FaceRecognitionIntegratorData) (code, reason, decision string) {

	switch data.RuleCode {
	case "6020":
		code = constant.CODE_FACERECOGNITION_REJECT_NIK
		decision = constant.DECISION_REJECT
	case "6019":
		code = constant.CODE_FACERECOGNITION_REJECT_FOTO
		decision = constant.DECISION_REJECT
	case "6018":
		code = constant.CODE_FACERECOGNITION_PASS
		decision = constant.DECISION_PASS
	}
	reason = data.Reason
	return
}

func checkRuleCodeIzidata(data response.FaceRecognitionIntegratorData) (code, reason, decision string) {

	switch data.RuleCode {
	case "6060":
		code = constant.CODE_FACERECOGNITION_IZIDATA_REJECT_NIK
		decision = constant.DECISION_REJECT
	case "6059":
		code = constant.CODE_FACERECOGNITION_IZIDATA_REJECT_FOTO
		decision = constant.DECISION_REJECT
	case "6058":
		code = constant.CODE_FACERECOGNITION_IZIDATA_PASS
		decision = constant.DECISION_PASS
	}
	reason = data.Reason
	return
}

func checkEKYCDukcapil(data response.VerifyDataIntegratorResponse, thresholdDukcapil entity.ConfigThresholdDukcapil) (code, reason, decision string) {

	if data.IsValid {
		if strings.Contains(data.Nik, "Tidak Sesuai") || strings.Contains(data.TglLhr, "Tidak Sesuai") || strings.Contains(data.JenisKlmin, "Tidak Sesuai") {
			return constant.CODE_VERIFICATION_REJECT_EKYC, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}

		if float64(data.NamaLgkp) < thresholdDukcapil.Data.VerifyData.VDDukcapil.NamaLengkap {
			return constant.CODE_VERIFICATION_REJECT_EKYC, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}

		if float64(data.Alamat) < thresholdDukcapil.Data.VerifyData.VDDukcapil.Alamat {
			return constant.CODE_VERIFICATION_REJECT_EKYC, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}

		return constant.CODE_VERIFICATION_PASS_EKYC, "EKYC Sesuai", constant.DECISION_PASS

	}
	switch *data.Reason {
	case constant.CUSTOMER_MENINGGAL:
		code = constant.CODE_VERIFICATION_REJECT_MENINGGAL
		decision = constant.DECISION_REJECT
	case constant.DATA_GANDA:
		code = constant.CODE_VERIFICATION_REJECT_DATA_GANDA
		decision = constant.DECISION_REJECT
	case constant.DATA_INACTIVE:
		code = constant.CODE_VERIFICATION_REJECT_INACTIVE
		decision = constant.DECISION_REJECT
	case constant.DATA_NOT_FOUND:
		code = constant.CODE_VERIFICATION_REJECT_NOT_FOUND
		decision = constant.DECISION_REJECT
	}
	reason = *data.Reason

	return

}

func checkEKYCIzidata(data response.VerifyDataIntegratorResponse, configEkyc entity.ConfigThresholdDukcapil) (code, reason, decision string) {

	if data.IsValid {
		if strings.Contains(data.Nik, "Tidak Sesuai") || strings.Contains(data.TglLhr, "Tidak Sesuai") {
			return constant.CODE_IZIDATA_REJECT_INVALID, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}

		if float64(data.NamaLgkp) < configEkyc.Data.VerifyData.VDIziData.NamaLengkap {
			return constant.CODE_IZIDATA_REJECT_INVALID, "EKYC Tidak Sesuai", constant.DECISION_REJECT
		}

		return constant.CODE_IZIDATA_PASS_VALID, "EKYC Sesuai", constant.DECISION_PASS

	}

	//Data Invalid
	switch *data.Reason {
	case constant.DATA_INVALID:
		code = constant.CODE_IZIDATA_REJECT_INVALID
		decision = constant.DECISION_REJECT
	case constant.DATA_NOT_FOUND:
		code = constant.CODE_IZIDATA_REJECT_NOT_FOUND
		decision = constant.DECISION_REJECT
	}
	reason = *data.Reason

	return

}
