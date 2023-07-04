package usecase

import (
	"context"
	"errors"
	"los-kmb-api/models/other"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"strings"
)

func (u usecase) RejectTenor36(ctx context.Context, prospectID, idNumber, accessToken string) (result response.UsecaseApi, err error) {

	result.Code = constant.CODE_REJECT_TENOR
	result.Result = constant.DECISION_REJECT
	result.Reason = constant.REASON_REJECT_TENOR

	dataInquiry, err := u.repository.GetDataInquiry(idNumber)

	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = errors.New(constant.ERROR_UPSTREAM + " - Error Get Data Inquiry")
		CentralizeLog(constant.DUPCHECK_LOG, "Get Data Inquiry", constant.MESSAGE_E, "DUPCHECK SERVICE", true, other.CustomLog{ProspectID: prospectID, Error: strings.Split(err.Error(), " - ")[1]})
		return
	}

	if len(dataInquiry) > 0 {
		isRejectDSRToday := false
		for _, Inquiry := range dataInquiry {
			if Inquiry.FinalApproval == 0 && Inquiry.RejectDSR == 1 {
				isRejectDSRToday = true
				break
			}
		}

		if isRejectDSRToday {
			result.Code = constant.CODE_PASS_TENOR
			result.Result = constant.RESULT_OK
			result.Reason = constant.REASON_PASS_TENOR
			return
		}
	}

	return
}
