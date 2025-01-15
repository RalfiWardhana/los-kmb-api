package usecase

import "los-kmb-api/models/response"

func (u usecase) History2Wilen(prospectID string) (data []response.History2Wilen, err error) {

	history, err := u.repository.GetTrxKPMStatusHistory(prospectID)
	if err != nil {
		return
	}

	if len(history) > 0 {
		data = make([]response.History2Wilen, len(history))

		for i, h := range history {
			data[i] = response.History2Wilen{
				ID:              h.ID,
				ProspectID:      h.ProspectID,
				OrderStatusName: h.Decision,
				CreatedAt:       h.CreatedAt.Format("2006-01-02 15:04:05"),
			}
		}
	}

	return
}
