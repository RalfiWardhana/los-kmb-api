package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
)

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
