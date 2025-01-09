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

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) GenerateFormAKKK(ctx context.Context, req request.RequestGenerateFormAKKK, accessToken string) (data interface{}, err error) {

	// check trx status
	trxStatus, err := u.repository.GetTrxStatus(req.ProspectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK GetTrxStatus Error")
		return
	}

	if trxStatus.Activity != constant.ACTIVITY_STOP {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - GenerateFormAKKK GetTrxStatus Status Pengajuan Belum Selesai")
		return
	}

	// call generator form akkk api
	payload, _ := json.Marshal(req)
	respAPI, errAPI := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("GENERATOR_FORM_AKKK_URL"), payload, map[string]string{}, constant.METHOD_POST, false, 0, 60, req.ProspectID, accessToken)

	if errAPI != nil || respAPI.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Failed Generate Form AKKK")
	}

	// set response
	if err == nil {
		var responseAkkk response.ResponseGenerateFormAKKK
		json.Unmarshal([]byte(jsoniter.Get(respAPI.Body(), "data").ToString()), &responseAkkk)

		if responseAkkk.MediaUrl == "" {

			err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal MediaUrl Form AKKK Error")

		} else {

			// save url form akkk
			err = u.repository.SaveUrlFormAKKK(req.ProspectID, responseAkkk.MediaUrl)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK SaveUrlFormAKKK Error")
			} else {
				data = responseAkkk
			}

		}
	}

	// if requested by system then do retry via worker
	if err != nil && req.Source == constant.SYSTEM {
		// set payload for worker
		req.Source = "WORKER"
		payloadWorker, _ := json.Marshal(req)

		trxWorker := entity.TrxWorker{
			ProspectID:      req.ProspectID,
			Category:        "FORM_AKKK_NKMB",
			Action:          "GENERATE_FORM_AKKK",
			APIType:         "RAW",
			EndPointTarget:  os.Getenv("RETRY_GENERATE_FORM_AKKK_URL"),
			EndPointMethod:  constant.METHOD_POST,
			Header:          `{}`,
			Payload:         string(payloadWorker),
			ResponseTimeout: 60,
			MaxRetry:        6,
			CountRetry:      0,
			Activity:        constant.ACTIVITY_UNPROCESS,
		}

		err = u.repository.SaveWorker(trxWorker)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GenerateFormAKKK SaveWorker Error")
			return
		}
	}

	return
}
