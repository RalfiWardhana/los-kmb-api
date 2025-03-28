package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) BlacklistCheck(index int, spDupcheck response.SpDupCekCustomerByID) (data response.UsecaseApi, customerType string) {

	customerType = constant.MESSAGE_BERSIH
	data.SourceDecision = constant.SOURCE_DECISION_BLACKLIST

	if spDupcheck != (response.SpDupCekCustomerByID{}) {

		data.StatusKonsumen = constant.STATUS_KONSUMEN_AO

		if (spDupcheck.TotalInstallment <= 0 && spDupcheck.RRDDate != nil) || (spDupcheck.TotalInstallment > 0 && spDupcheck.RRDDate != nil && spDupcheck.NumberOfPaidInstallment == nil) {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_RO
		}

		if spDupcheck.IsSimiliar == 1 && index == 0 {
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_KONSUMEN_SIMILIAR
			data.Reason = constant.REASON_KONSUMEN_SIMILIAR
			customerType = constant.MESSAGE_BLACKLIST
			return

		} else if spDupcheck.BadType == constant.BADTYPE_B {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST
			}
			return

		} else if spDupcheck.MaxOverdueDays > 90 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_OVD_90DAYS

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_OVD_90DAYS
			}
			return

		} else if spDupcheck.NumOfAssetInventoried > 0 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_ASSET_INVENTORY

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_ASSET_INVENTORY
			}
			return

		} else if spDupcheck.IsRestructure == 1 {
			data.Result = constant.DECISION_REJECT
			customerType = constant.MESSAGE_BLACKLIST
			if index == 0 {
				data.Code = constant.CODE_KONSUMEN_BLACKLIST
				data.Reason = constant.REASON_KONSUMEN_BLACKLIST_RESTRUCTURE

			} else {
				data.Code = constant.CODE_PASANGAN_BLACKLIST
				data.Reason = constant.REASON_PASANGAN_BLACKLIST_RESTRUCTURE
			}
			return

		} else if spDupcheck.BadType == constant.BADTYPE_W {
			customerType = constant.MESSAGE_WARNING
		}

	} else {
		data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
	}

	data = response.UsecaseApi{StatusKonsumen: data.StatusKonsumen, Code: constant.CODE_NON_BLACKLIST_ALL, Reason: constant.REASON_NON_BLACKLIST, Result: constant.DECISION_PASS}

	return
}

func (u usecase) DupcheckIntegrator(ctx context.Context, prospectID, idNumber, legalName, birthDate, surgateName string, accessToken string) (spDupcheck response.SpDupCekCustomerByID, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))

	req, _ := json.Marshal(map[string]interface{}{
		"transaction_id":      prospectID,
		"id_number":           idNumber,
		"legal_name":          legalName,
		"birth_date":          birthDate,
		"surgate_mother_name": surgateName,
	})

	custDupcheck, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("DUPCHECK_URL"), req, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)

	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Dupcheck Timeout")
		return
	}

	if custDupcheck.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Dupcheck Error")
		return
	}

	if err = json.Unmarshal([]byte(jsoniter.Get(custDupcheck.Body(), "data").ToString()), &spDupcheck); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data dupcheck")
		return
	}

	return

}

func (u usecase) BannedPMKOrDSR(encrypted string) (data response.UsecaseApi, err error) {

	var trxReject entity.TrxBannedPMKDSR
	trxReject, err = u.repository.GetBannedPMKDSR(encrypted)
	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Get Banned PMK DSR Error")
		return
	}

	if trxReject != (entity.TrxBannedPMKDSR{}) {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_PERNAH_REJECT_PMK_DSR
		data.Reason = constant.REASON_PERNAH_REJECT_PMK_DSR
		data.SourceDecision = constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR
		return
	}

	return
}

func (u usecase) CustomerKMB(spDupcheck response.SpDupCekCustomerByID) (statusKonsumen string, err error) {

	if spDupcheck == (response.SpDupCekCustomerByID{}) {
		statusKonsumen = constant.STATUS_KONSUMEN_NEW
		return
	}

	if (spDupcheck.TotalInstallment <= 0 && spDupcheck.RRDDate != nil) || (spDupcheck.TotalInstallment > 0 && spDupcheck.RRDDate != nil && spDupcheck.NumberOfPaidInstallment == nil) {
		statusKonsumen = constant.STATUS_KONSUMEN_RO
		return

	} else if spDupcheck.TotalInstallment > 0 {
		statusKonsumen = constant.STATUS_KONSUMEN_AO
		return

	} else {
		statusKonsumen = constant.STATUS_KONSUMEN_NEW
		return
	}

}

func (u usecase) Rejection(prospectID string, encrypted string, configValue response.DupcheckConfig) (data response.UsecaseApi, trxBannedPMKDSR entity.TrxBannedPMKDSR, err error) {

	trxReject, err := u.repository.GetRejection(encrypted)
	if err != nil {
		// err = errors.New(constant.ERROR_UPSTREAM + " - Get Trx Reject Error")
		return
	}

	if trxReject.RejectPMKDSR > 0 {
		if (trxReject.RejectPMKDSR + trxReject.RejectNIK) >= configValue.Data.AttemptPMKDSR {
			//banned 30 hari
			trxBannedPMKDSR = entity.TrxBannedPMKDSR{
				ProspectID: prospectID,
				IDNumber:   encrypted,
			}
			data.Result = constant.DECISION_REJECT
			data.Code = constant.CODE_PERNAH_REJECT_PMK_DSR
			data.Reason = constant.REASON_PERNAH_REJECT_PMK_DSR
			data.SourceDecision = constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR
			return
		}
	}

	if trxReject.RejectNIK >= configValue.Data.AttemptPMKDSR {
		data.Result = constant.DECISION_REJECT
		data.Code = constant.CODE_PERNAH_REJECT_NIK
		data.Reason = constant.REASON_PERNAH_REJECT_NIK
		data.SourceDecision = constant.SOURCE_DECISION_NIK
		return
	}

	data.Result = constant.DECISION_PASS
	data.Code = constant.CODE_BELUM_PERNAH_REJECT
	data.Reason = constant.REASON_BELUM_PERNAH_REJECT
	data.SourceDecision = constant.SOURCE_DECISION_PERNAH_REJECT_PMK_DSR

	return
}
