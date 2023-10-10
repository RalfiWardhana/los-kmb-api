package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
	"github.com/labstack/echo/v4"
)

func (u usecase) GetBase64Media(ctx context.Context, url string, customerID int, accessToken string) (base64Media string, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("MEDIA_TIMEOUT"))

	var (
		image1 *resty.Response
		decode response.ImageDecodeResponse
	)

	requestID, ok := ctx.Value(echo.HeaderXRequestID).(string)
	if !ok {
		requestID = ""
	}

	header := map[string]string{
		"Content-Type":        "application/json",
		"Authorization":       os.Getenv("MEDIA_KEY"),
		echo.HeaderXRequestID: requestID,
	}

	myMedia := utils.GetIsMedia(url)

	if myMedia {

		image1, err = u.httpclient.MediaClient(ctx, constant.NEW_KMB_LOG, url+"?type=base64", constant.METHOD_GET, nil, header, timeOut, customerID, accessToken)

		if image1.StatusCode() != 200 || err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Platform Media Request Error")
			return
		}

		err = json.Unmarshal([]byte(image1.Body()), &decode)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Platform Media Unmarshal Error")
			return
		}

		base64Media = decode.Data.Encode

	} else {

		base64Media, err = utils.DecodeNonMedia(url)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Non Media Error")
			return
		}

	}

	return
}
