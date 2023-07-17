package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

func (u usecase) FacePlus(ctx context.Context, imageKtp string, imageSelfie string, req request.FaceCompareRequest, accessToken string) (result response.FaceCompareResponse, err error) {

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	param, _ := json.Marshal(map[string]interface{}{
		"image_base64_1": imageKtp,
		"image_base64_2": imageSelfie,
		"prospect_id":    req.ProspectID,
	})

	resp, err := u.httpclient.EngineAPI(ctx, constant.LOG_FACE_COMPARE_TRX, os.Getenv("COMPARE_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeout, req.ProspectID, accessToken)

	var faceplusResponseBody response.CompareResponse

	if resp.StatusCode() == 400 {

		if err = json.Unmarshal(resp.Body(), &faceplusResponseBody); err != nil {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - FacePlus Unmarshal Error")
			return
		}

		err = fmt.Errorf(constant.ERROR_BAD_REQUEST + " - " + faceplusResponseBody.Meta.Error)
		return
	}

	if resp.StatusCode() != 200 || err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - FacePlus Request Error")
		return
	}

	if err = json.Unmarshal(resp.Body(), &faceplusResponseBody); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - FacePlus Unmarshal Error")
		return
	}

	config := u.repository.GetConfig(constant.GROUP_FPP, constant.LOB_NEW_KMB, "face_plus_confidence")

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - FacePlus Config Error")
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
			ProspectID: req.ProspectID,
			RequestID:  requestID,
			Result:     constant.DECISION_REJECT,
			Reason:     constant.REASON_CONFIDENCE_BELOW_THRESHOLD,
		}

		return
	}

	result = response.FaceCompareResponse{
		ProspectID: req.ProspectID,
		RequestID:  requestID,
		Result:     constant.DECISION_PASS,
		Reason:     constant.REASON_CONFIDENCE_UPPER_THRESHOLD,
	}

	return
}
