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
	"strings"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) ElaborateScheme(prospectID string, req request.Metrics) (data response.UsecaseApi, err error) {

	var (
		trxElaborateLtv entity.MappingElaborateLTV
	)

	trxElaborateLtv, err = u.repository.GetElaborateLtv(prospectID)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetElaborateLtv Error")
		return
	}

	ltv := (req.Apk.AF / req.Apk.OTR) * 100

	if ltv > float64(trxElaborateLtv.LTV) {
		data = response.UsecaseApi{
			Result:         constant.DECISION_REJECT,
			Code:           constant.STRING_CODE_REJECT_NTF_ELABORATE,
			Reason:         constant.REASON_REJECT_NTF_ELABORATE,
			SourceDecision: constant.SOURCE_DECISION_ELABORATE_LTV,
		}
	} else {
		data = response.UsecaseApi{
			Result:         constant.DECISION_PASS,
			Code:           constant.STRING_CODE_PASS_ELABORATE,
			Reason:         constant.REASON_PASS_ELABORATE,
			SourceDecision: constant.SOURCE_DECISION_ELABORATE_LTV,
		}
	}

	return
}

func (u usecase) ElaborateIncome(ctx context.Context, req request.Metrics, filtering entity.FilteringKMB, pefindoIDX response.PefindoIDX, spDupcheckMap response.SpDupcheckMap, responseScs response.IntegratorScorePro, accessToken string) (data response.UsecaseApi, err error) {

	var mappingElaborateIncome entity.MappingElaborateIncome

	mBranch, err := u.repository.GetMasterBranch(req.Transaction.BranchID)
	mappingElaborateIncome.BranchCategory = mBranch.BranchCategory
	if err != nil {
		mappingElaborateIncome.BranchCategory = "GOOD"
	}

	mappingElaborateIncome.Scoreband = "Segmen" + responseScs.Segmen
	if responseScs.Segmen == "" {
		mappingElaborateIncome.Scoreband = "Segmen7"
	}

	mappingElaborateIncome.BPKBNameType = 0
	if strings.Contains(os.Getenv("NAMA_SAMA"), req.Item.BPKBName) {
		mappingElaborateIncome.BPKBNameType = 1

	}

	if spDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_AO || spDupcheckMap.StatusKonsumen == constant.STATUS_KONSUMEN_RO {
		mappingElaborateIncome.StatusKonsumen = "AO/RO"
	} else {
		mappingElaborateIncome.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
	}

	param, _ := json.Marshal(map[string]interface{}{
		"prospect_id":        req.Transaction.ProspectID,
		"requestor_id":       os.Getenv("LOW_INCOME_REQUESTOR_ID"),
		"score_generator_id": os.Getenv("LOW_INCOME_SCORE_GENERATOR_ID"),
		"data": map[string]interface{}{
			"transaction_id":     req.Transaction.ProspectID,
			"total_angsuran_pbk": filtering.TotalInstallmentAmountBiro,
		},
	})

	resp, err := u.httpclient.EngineAPI(ctx, constant.NEW_KMB_LOG, os.Getenv("LOW_INCOME_API"), param, map[string]string{}, constant.METHOD_POST, false, 0, 60, req.Transaction.ProspectID, accessToken)
	if err != nil {
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call Low Income Timeout")
			return
		}
	}

	if resp.StatusCode() != 200 {
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Low Income Error")
			return
		}
	}

	var respLowIncome response.LowIncome
	err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &respLowIncome)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Unmarshal Low Income Error")
		return
	}

	mappingElaborateIncome.EstimationIncome = respLowIncome.Range

	mappingElaborateIncome.Worst24Mth = "<=30"
	if pefindoIDX.Worst24Mth > 3 {
		mappingElaborateIncome.Worst24Mth = ">30"
	}

	mappingElaborateIncome, err = u.repository.GetMappingElaborateIncome(mappingElaborateIncome)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetMappingElaborateIncome Error")
		return
	}

	if mappingElaborateIncome.Result == constant.DECISION_REJECT {
		data = response.UsecaseApi{
			Result:         constant.DECISION_REJECT,
			Code:           constant.CODE_REJECT_ELABORATE_INCOME,
			Reason:         constant.REASON_REJECT_ELABORATE_INCOME,
			SourceDecision: constant.SOURCE_DECISION_ELABORATE_INCOME,
		}
	} else {
		data = response.UsecaseApi{
			Result:         constant.DECISION_PASS,
			Code:           constant.CODE_PASS_ELABORATE_INCOME,
			Reason:         constant.REASON_PASS_ELABORATE_INCOME,
			SourceDecision: constant.SOURCE_DECISION_ELABORATE_INCOME,
		}
	}

	return
}
