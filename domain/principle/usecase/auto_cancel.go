package usecase

import (
	"context"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
)

func (u usecase) CheckOrderPendingPrinciple(ctx context.Context) (err error) {

	data, err := u.repository.ScanOrderPending()

	for _, val := range data {

		err := u.repository.UpdateToCancel(val.ProspectID)

		if err == nil {
			u.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_PRINCIPLE, constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE, val.ProspectID, utils.StructToMap(request.Update2wPrincipleTransaction{
				KpmID:       val.KPMID,
				OrderID:     val.ProspectID,
				Source:      3,
				StatusCode:  constant.PRINCIPLE_STATUS_CANCEL_LOS,
				ProductName: val.AssetCode,
				BranchCode:  val.BranchID,
			}), 0)
		} else {
			continue
		}

	}

	return
}
