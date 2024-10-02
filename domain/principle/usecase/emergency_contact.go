package usecase

import (
	"context"
	"encoding/json"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"
	"sync"
)

func (u usecase) PrincipleEmergencyContact(ctx context.Context, req request.PrincipleEmergencyContact, accessToken string) (err error) {
	var (
		principleStepThree           entity.TrxPrincipleStepThree
		trxPrincipleEmergencyContact entity.TrxPrincipleEmergencyContact
		wg                           sync.WaitGroup
		errChan                      = make(chan error, 2)
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		principleStepThree, err = u.repository.GetPrincipleStepThree(req.ProspectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		trxPrincipleEmergencyContact = entity.TrxPrincipleEmergencyContact{
			ProspectID:   req.ProspectID,
			Name:         req.Name,
			Relationship: req.Relationship,
			MobilePhone:  req.MobilePhone,
			Address:      req.Address,
			Rt:           req.Rt,
			Rw:           req.Rw,
			Kelurahan:    req.Kelurahan,
			Kecamatan:    req.Kecamatan,
			City:         req.City,
			Province:     req.Province,
			ZipCode:      req.ZipCode,
			AreaPhone:    req.AreaPhone,
			Phone:        req.Phone,
		}

		err = u.repository.SavePrincipleEmergencyContact(trxPrincipleEmergencyContact, principleStepThree.IDNumber)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return err
	}

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	var worker []entity.TrxWorker

	headerParamLos, _ := json.Marshal(
		map[string]string{
			"X-Client-ID":   os.Getenv("CLIENT_LOS"),
			"Authorization": os.Getenv("AUTH_LOS"),
		})

	// insert customer
	sequence := 1

	worker = append(worker, entity.TrxWorker{ProspectID: req.ProspectID, Activity: constant.WORKER_UNPROCESS, EndPointTarget: os.Getenv("PRINCIPLE_CORE_CUSTOMER_URL") + req.ProspectID,
		EndPointMethod: constant.METHOD_POST, Header: string(headerParamLos), Payload: "",
		ResponseTimeout: timeOut, APIType: constant.WORKER_TYPE_RAW, MaxRetry: 6, CountRetry: 0,
		Category: constant.WORKER_CATEGORY_PRINCIPLE_KMB, Action: constant.WORKER_ACTION_UPDATE_CORE_CUSTOMER, Sequence: sequence,
	})

	// get marketing program
	sequence += 1

	worker = append(worker, entity.TrxWorker{ProspectID: req.ProspectID, Activity: constant.WORKER_IDLE, EndPointTarget: os.Getenv("PRINCIPLE_MARKETING_PROGRAM_URL") + req.ProspectID,
		EndPointMethod: constant.METHOD_POST, Header: string(headerParamLos), Payload: "",
		ResponseTimeout: timeOut, APIType: constant.WORKER_TYPE_RAW, MaxRetry: 6, CountRetry: 0,
		Category: constant.WORKER_CATEGORY_PRINCIPLE_KMB, Action: constant.WORKER_ACTION_GET_MARKETING_PROGRAM, Sequence: sequence,
	})

	go u.repository.SaveToWorker(worker)

	return
}
