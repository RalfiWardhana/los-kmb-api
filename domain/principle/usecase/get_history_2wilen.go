package usecase

import (
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
)

func (u usecase) History2Wilen(req request.History2Wilen) (data []response.History2Wilen, err error) {

	history, err := u.repository.GetTrxKPMStatusHistory(req)
	if err != nil {
		return
	}

	if len(history) > 0 {
		data = make([]response.History2Wilen, len(history))

		for i, h := range history {
			data[i] = response.History2Wilen{
				ID:              h.ID,
				IDNumber:        h.IDNumber,
				KPMID:           h.KpmID,
				ReferralCode:    h.ReferralCode,
				ProspectID:      h.ProspectID,
				LoanAmount:      h.LoanAmount,
				OrderStatusName: h.Decision,
				CreatedAt:       h.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}
	}

	return
}
