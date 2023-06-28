package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	request "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (u multiUsecase) GetPhoto(ctx context.Context, req request.FaceCompareRequest, accessToken string) (selfie1 string, selfie2 string, err error) {

	selfie1Media := utils.GetIsMedia(req.ImageSelfie1)

	if selfie1Media {
		selfie1, err = u.usecase.DecodeMedia(ctx, req.ImageSelfie1, req.CustomerID, accessToken)
		if err != nil {
			return
		}

	} else {
		selfie1, err = utils.DecodeNonMedia(req.ImageSelfie1)
		if err != nil {
			return
		}
	}

	selfie2Media := utils.GetIsMedia(req.ImageSelfie2)

	if selfie2Media {
		selfie2, err = u.usecase.DecodeMedia(ctx, req.ImageSelfie2, req.CustomerID, accessToken)
		if err != nil {
			return
		}

	} else {
		selfie2, err = utils.DecodeNonMedia(req.ImageSelfie2)
		if err != nil {
			return
		}
	}

	return
}

func (u usecase) DecodeMedia(ctx context.Context, url string, customerID int, accessToken string) (base64Image string, err error) {

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

	image, err := u.httpclient.MediaClient(ctx, constant.LOG_JOURNEY_LOG, url+"?type=base64", constant.METHOD_GET, nil, header, timeOut, customerID, accessToken)

	if image.StatusCode() != 200 || err != nil {
		err = errors.New(constant.CANNOT_GET_IMAGE)
		return
	}

	err = json.Unmarshal([]byte(image.Body()), &decode)

	if err != nil {
		return
	}

	base64Image = decode.Data.Encode
	return
}

func (u usecase) FacePlus(ctx context.Context, selfie1 string, selfie2 string, req request.FaceCompareRequest, accessToken string) (result response.FaceCompareResponse, err error) {

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
			return
		}

		err = fmt.Errorf("400-" + faceplusResponseBody.Meta.Error)
		return
	}

	if resp.StatusCode() != 200 || err != nil {
		return
	}

	if err = json.Unmarshal(resp.Body(), &faceplusResponseBody); err != nil {
		return
	}

	config := u.repository.GetConfig(constant.GROUP_FPP, constant.LOB_NEW_KMB, "face_plus_confidence")

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
	case constant.LOB_KMB:
		confidence = faceConfig.Data.Kmb
	}

	dataConfidence, _ := strconv.ParseFloat(faceplusResponseBody.Data.Confidence, 64)

	if int(dataConfidence) < confidence {

		result = response.FaceCompareResponse{
			CustomerID: req.CustomerID,
			RequestID:  requestID,
			Result:     constant.DECISION_REJECT,
			Reason:     constant.REASON_CONFIDENCE_BELOW_THRESHOLD,
		}

		return
	}

	result = response.FaceCompareResponse{
		CustomerID: req.CustomerID,
		RequestID:  requestID,
		Result:     constant.DECISION_PASS,
		Reason:     constant.REASON_CONFIDENCE_UPPER_THRESHOLD,
	}

	return
}
