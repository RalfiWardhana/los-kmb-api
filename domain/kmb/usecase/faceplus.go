package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/middlewares"
	entity "los-kmb-api/models/dupcheck"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (u *usecase) DecodeMedia(ctx context.Context, url string, customerID int, accessToken string) (base64Image string, err error) {
	timeOut, _ := strconv.Atoi(os.Getenv("MEDIA_TIMEOUT"))

	var decode response.ImageDecodeResponse

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	header := map[string]string{
		"Content-Type":        "application/json",
		"Authorization":       os.Getenv("MEDIA_KEY"),
		echo.HeaderXRequestID: requestID,
	}

	image, err := u.httpclient.MediaClient(ctx, constant.LOG_FACE_COMPARE_TRX, url+os.Getenv("MEDIA_PATH"), constant.METHOD_GET, nil, header, timeOut, customerID, accessToken)

	if image.StatusCode() != 200 || err != nil {
		err = errors.New(constant.CANNOT_GET_IMAGE)
		return
	}

	err = json.Unmarshal([]byte(image.Body()), &decode)

	if err != nil {
		err = fmt.Errorf("DECODE MEDIA - UNMARSHAL ERROR")
		return
	}

	base64Image = decode.Data.Encode

	return
}

func (u *usecase) DetectBlurness(ctx context.Context, selfie1 string) (decision bool, err error) {

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	config := u.repository.GetConfig(constant.GROUP_FPP, constant.LOB_KMB_NEW, "face_plus_confidence")

	var (
		faceConfig response.Config
		confidence int
	)

	json.Unmarshal([]byte(config.Value), &faceConfig)

	confidence = faceConfig.Data.Blur

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	param, _ := json.Marshal(map[string]interface{}{
		"image_base64": selfie1,
		"prospect_id":  requestID,
	})

	resp, err := u.httpclient.EngineAPI(ctx, constant.LOG_FACE_COMPARE_TRX, os.Getenv("DETECT_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeout, requestID, middlewares.UserInfoData.AccessToken)

	var data response.DetectImageResponse

	if resp.StatusCode() == 400 {

		if err = json.Unmarshal(resp.Body(), &data); err != nil {
			err = fmt.Errorf("DETECT BLUR - UNMARSHAL ERROR")
			return
		}

		err = fmt.Errorf("400-" + data.Meta.Error)
		return
	}

	if resp.StatusCode() != 200 || err != nil {
		err = fmt.Errorf("DETECT BLUR - API INTEGERATOR SERVICE UNAVAILABLE")
		return
	}

	if err = json.Unmarshal(resp.Body(), &data); err != nil {
		err = fmt.Errorf("DETECT BLUR - UNMARSHAL ERROR")
		return
	}

	if data.Data.BlurValue > float64(confidence) {
		decision = true
		return
	}

	return
}

func (u *usecase) FacePlus(ctx context.Context, selfie1 string, selfie2 string, req request.FaceCompareRequest, accessToken string, getPhotoInfo interface{}) (result response.FaceCompareResponse, err error) {

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	prospectID := strconv.Itoa(req.CustomerID)

	param, _ := json.Marshal(map[string]interface{}{
		"image_base64_1": selfie1,
		"image_base64_2": selfie2,
		"prospect_id":    prospectID,
	})

	resp, err := u.httpclient.EngineAPI(ctx, constant.LOG_FACE_COMPARE_TRX, os.Getenv("COMPARE_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeout, prospectID, accessToken)

	var faceplusResponseBody response.CompareResponse

	if resp.StatusCode() == 400 {

		if err = json.Unmarshal(resp.Body(), &faceplusResponseBody); err != nil {
			err = fmt.Errorf("UNMARSHAL ERROR")
			return
		}

		err = fmt.Errorf("400-" + faceplusResponseBody.Meta.Error)
		return
	}

	if resp.StatusCode() != 200 || err != nil {
		err = fmt.Errorf("API INTEGERATOR SERVICE UNAVAILABLE")
		return
	}

	if err = json.Unmarshal(resp.Body(), &faceplusResponseBody); err != nil {
		err = fmt.Errorf("UNMARSHAL ERROR")
		return
	}

	faceplusInfo, _ := json.Marshal(faceplusResponseBody.Data)
	resultGetPhoto, _ := json.Marshal(getPhotoInfo)

	config := u.repository.GetConfig(constant.GROUP_FPP, constant.LOB_KMB_NEW, "face_plus_confidence")

	if err != nil {
		return
	}

	var (
		faceConfig response.Config
		confidence int
	)

	err = json.Unmarshal([]byte(config.Value), &faceConfig)

	if err != nil {
		confidence, _ = strconv.Atoi(os.Getenv("DEFAULT_CONFIDENCE"))
	}

	switch req.Lob {
	case constant.LOB_WG:
		confidence = faceConfig.Data.WG.Online
	case constant.LOB_KMB:
		confidence = faceConfig.Data.Kmb
	case constant.LOB_KMOB:
		confidence = faceConfig.Data.Kmob
	}

	dataConfidence, _ := strconv.ParseFloat(faceplusResponseBody.Data.Confidence, 64)

	if int(dataConfidence) < confidence {

		result = response.FaceCompareResponse{
			CustomerID: req.CustomerID,
			RequestID:  requestID,
			Result:     constant.DECISION_REJECT,
			Reason:     constant.REASON_CONFIDENCE_BELOW_THRESHOLD,
		}

		resultInfo, _ := json.Marshal(result)

		insertData := entity.VerificationFaceCompare{
			ID:             requestID,
			CustomerID:     req.CustomerID,
			ResultGetPhoto: string(resultGetPhoto),
			ResultFacePlus: string(faceplusInfo),
			Decision:       constant.DECISION_REJECT,
			Result:         string(resultInfo),
		}

		if err := u.repository.SaveVerificationFaceCompare(insertData); err != nil {
			return response.FaceCompareResponse{}, fmt.Errorf("failed to save verification data: %w", err)
		}

		return
	}

	result = response.FaceCompareResponse{
		CustomerID: req.CustomerID,
		RequestID:  requestID,
		Result:     constant.DECISION_PASS,
		Reason:     constant.REASON_CONFIDENCE_UPPER_THRESHOLD,
	}

	resultInfo, _ := json.Marshal(result)

	insertData := entity.VerificationFaceCompare{
		ID:             requestID,
		CustomerID:     req.CustomerID,
		ResultGetPhoto: string(resultGetPhoto),
		ResultFacePlus: string(faceplusInfo),
		Decision:       constant.DECISION_PASS,
		Result:         string(resultInfo),
	}

	if err := u.repository.SaveVerificationFaceCompare(insertData); err != nil {
		return response.FaceCompareResponse{}, fmt.Errorf("failed to save verification data: %w", err)
	}

	// bodyValue := map[string]interface{}{
	// 	"result_compare":    constant.RESULT_COMPARE,
	// 	"verification_date": utils.GenerateTimeWithFormat(constant.FORMAT_DATE_TIME),
	// 	"verification_from": constant.FLAG_LOS,
	// 	"customer_id":       req.CustomerID,
	// }

	// if req.FaceType == nil {
	// 	if err = u.platformEvent.PublishEvent(ctx, accessToken, constant.TOPIC_VERIFICATION, constant.KEY_PREFIX_FACE_VERIFICATION, strconv.Itoa(req.CustomerID), bodyValue, 0); err != nil {
	// 		err = fmt.Errorf("publish event error: %w", err)
	// 		return
	// 	}
	// }

	return
}
