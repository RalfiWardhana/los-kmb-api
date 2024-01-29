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

func (u multiUsecase) Ekyc(ctx context.Context, req request.Metrics, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, trxDetail []entity.TrxDetail, trxFMF response.TrxFMF, err error) {

	data, err = u.usecase.Dukcapil(ctx, req, reqMetricsEkyc, accessToken)

	if err != nil && err.Error() != fmt.Sprintf("%s - Dukcapil", constant.TYPE_CONTINGENCY) {
		return
	}

	if err != nil && err.Error() == fmt.Sprintf("%s - Dukcapil", constant.TYPE_CONTINGENCY) {

		trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.Transaction.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_CONTINGENCY, RuleCode: data.Code, SourceDecision: data.Source, Info: data.Info, NextStep: constant.SOURCE_DECISION_ASLIRI})

		data, err = u.usecase.Asliri(ctx, req, accessToken)

		if err != nil {

			trxDetail = append(trxDetail, entity.TrxDetail{ProspectID: req.Transaction.ProspectID, StatusProcess: constant.STATUS_ONPROCESS, Activity: constant.ACTIVITY_PROCESS, Decision: constant.DB_DECISION_CONTINGENCY, RuleCode: constant.CODE_CONTINGENCY, SourceDecision: constant.SOURCE_DECISION_ASLIRI, Info: constant.TYPE_CONTINGENCY, NextStep: constant.SOURCE_DECISION_KTP_VALIDATOR})

			data, err = u.usecase.Ktp(ctx, req, reqMetricsEkyc, accessToken)

			trxFMF.EkycSource = "KTP VALIDATOR"
			trxFMF.EkycSimiliarity = data.Similiarity
			trxFMF.EkycReason = data.Reason
			return
		}

		trxFMF.EkycSource = "ASLI RI"
		trxFMF.EkycSimiliarity = data.Similiarity
		trxFMF.EkycReason = data.Reason
		return
	}

	trxFMF.EkycSource = "DUKCAPIL"
	trxFMF.EkycSimiliarity = data.Similiarity
	trxFMF.EkycReason = data.Reason

	return

}

func (u usecase) Dukcapil(ctx context.Context, req request.Metrics, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error) {

	var (
		selfie, codeVD, _, decisionVD, decisionFR   string
		address, city, kelurahan, kecamatan, rt, rw string
		infoDukcapil                                response.InfoEkyc
		verify                                      response.VerifyDataIntegratorResponse
		face                                        response.FaceRecognitionIntegratorData
		thresholdDukcapil                           entity.ConfigThresholdDukcapil
		timeout                                     int
		statusVD, statusFR                          string
		endpointVd, endpointFr                      string
		thresholdFr                                 float64
	)

	config, err := u.repository.GetConfig("ekyc", "NKMB", "threshold_ekyc")

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Dukcapil Config Error")
		return
	}

	json.Unmarshal([]byte(config.Value), &thresholdDukcapil)

	for i := 0; i < len(req.Address); i++ {

		if req.Address[i].Type == constant.ADDRESS_TYPE_LEGAL {
			address = req.Address[i].Address
			city = req.Address[i].City
			kelurahan = req.Address[i].Kelurahan
			kecamatan = req.Address[i].Kecamatan
			rt = req.Address[i].Rt
			rw = req.Address[i].Rw
			break
		}
	}

	for _, photo := range req.CustomerPhoto {

		if photo.ID == constant.TAG_SELFIE_PHOTO {
			selfie = photo.Url
			break
		}
	}

	timeout, err = strconv.Atoi(os.Getenv("DUKCAPIL_TIMEOUT"))
	if err != nil {
		timeout, _ = strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	}

	// Verify Data
	paramVd, _ := json.Marshal(map[string]interface{}{
		"address":             address,
		"birth_date":          req.CustomerPersonal.BirthDate,
		"birth_place":         req.CustomerPersonal.BirthPlace,
		"city":                city,
		"gender":              req.CustomerPersonal.Gender,
		"marital_status":      req.CustomerPersonal.MaritalStatus,
		"id_number":           req.CustomerPersonal.IDNumber,
		"kabupaten":           city,
		"kecamatan":           kecamatan,
		"kelurahan":           kelurahan,
		"legal_name":          req.CustomerPersonal.LegalName,
		"profession_id":       req.CustomerEmployment.ProfessionID,
		"province":            city,
		"rt":                  rt,
		"rw":                  rw,
		"surgate_mother_name": req.CustomerPersonal.SurgateMotherName,
		"threshold":           "0",
		"transaction_id":      req.Transaction.ProspectID,
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

	resp, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, endpointVd, paramVd, map[string]string{}, constant.METHOD_POST, true, 2, timeout, req.Transaction.ProspectID, accessToken)

	infoDukcapil.VdService = serviceVD

	if resp.StatusCode() == 504 || resp.StatusCode() == 502 {
		statusVD = constant.EKYC_RTO

		infoDukcapil.VdError = "Request Timed Out"
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 504 && resp.StatusCode() != 502 {
		statusVD = constant.EKYC_NOT_CHECK

		var responseIntegrator response.ApiResponse
		json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &responseIntegrator)
		infoDukcapil.VdError = responseIntegrator.Message
	}

	if err == nil && resp.StatusCode() == 200 {

		json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &verify)

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
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Verify Dukcapil Error")
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
		"id_number":      req.CustomerPersonal.IDNumber,
		"selfie_image":   selfie,
		"threshold":      fmt.Sprintf("%.1f", thresholdFr),
		"transaction_id": req.Transaction.ProspectID,
	})

	resp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, endpointFr, paramFr, map[string]string{}, constant.METHOD_POST, true, 2, timeout, req.Transaction.ProspectID, accessToken)

	infoDukcapil.FrService = serviceFR

	if resp.StatusCode() == 504 || resp.StatusCode() == 502 {
		statusFR = constant.EKYC_RTO
		infoDukcapil.FrError = "Request Timed Out"
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 504 && resp.StatusCode() != 502 {
		statusFR = constant.EKYC_NOT_CHECK
		var responseIntegrator response.ApiResponse
		json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &responseIntegrator)
		infoDukcapil.FrError = responseIntegrator.Message
	}

	if err == nil && resp.StatusCode() == 200 {

		json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &face)
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
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Result Dukcapil Error")
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

func (u usecase) Asliri(ctx context.Context, req request.Metrics, accessToken string) (data response.Ekyc, err error) {

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

	paramARI, _ := json.Marshal(map[string]interface{}{
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

	resp, err = u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("ASLIRI_URL"), paramARI, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.Transaction.ProspectID, accessToken)

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

	config, err := u.repository.GetConfig("asliri", "KMB-OFF", "threshold_asliri")

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get ASLI RI Config Error")
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

func (u usecase) Ktp(ctx context.Context, req request.Metrics, reqMetricsEkyc request.MetricsEkyc, accessToken string) (data response.Ekyc, err error) {

	paramKtp, _ := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"birth_date": req.CustomerPersonal.BirthDate,
			"gender":     req.CustomerPersonal.Gender,
			"id_number":  req.CustomerPersonal.IDNumber,
			"is_pefindo": reqMetricsEkyc.CBFound,
			"request_id": uuid.New().String(),
		},
	})

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	resp, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("KTP_VALIDATOR_URL"), paramKtp, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.Transaction.ProspectID, accessToken)

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
		if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
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
