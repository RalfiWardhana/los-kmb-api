package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"los-kmb-api/domain/elaborate/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"time"

	"github.com/google/uuid"
)

type (
	multiUsecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		usecase    interfaces.Usecase
	}
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
	}
)

func NewMultiUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) (interfaces.MultiUsecase, interfaces.Usecase) {
	usecase := NewUsecase(repository, httpclient)

	return &multiUsecase{
		usecase:    usecase,
		repository: repository,
		httpclient: httpclient,
	}, usecase
}

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
	}
}

func (u multiUsecase) Elaborate(ctx context.Context, reqs request.BodyRequestElaborate, accessToken string) (data response.ElaborateResult, err error) {

	var savedata entity.ApiElaborateKmb

	id := uuid.New()

	savedata.RequestID = id.String()

	request, _ := json.Marshal(reqs)
	savedata.Request = string(request)
	savedata.ProspectID = reqs.Data.ProspectID
	if err = u.repository.SaveDataElaborate(savedata); err != nil {
		err = fmt.Errorf("failed process elaborate")
		return
	}

	var updateElaborate entity.ApiElaborateKmbUpdate

	status_konsumen := reqs.Data.CustomerStatus
	if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
		status_konsumen = "AO/RO"
	}
	kategori_status_konsumen := reqs.Data.CategoryCustomer
	check_prime_priority, _ := utils.ItemExists(kategori_status_konsumen, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

	tenor := reqs.Data.Tenor

	var max_ltv int

	if reqs.Data.CustomerStatus == constant.STATUS_KONSUMEN_RO_AO && check_prime_priority { //AO/RO PRIME PRIORITY
		if tenor < 36 {
			updateElaborate.Code = constant.CODE_PASS_ELABORATE
			updateElaborate.Reason = constant.REASON_PASS_ELABORATE
			updateElaborate.Decision = constant.DECISION_PASS
		} else {
			result_elaborate, errs := u.usecase.ResultElaborate(ctx, reqs)
			if errs != nil {
				err = fmt.Errorf("failed get result elaborate")
				return
			}

			if result_elaborate.Decision == constant.DECISION_REJECT {
				max_ltv = result_elaborate.LTV
			}

			updateElaborate.Code = result_elaborate.Code
			updateElaborate.Reason = result_elaborate.Reason
			updateElaborate.Decision = result_elaborate.Decision
		}
	} else { //NONE OF AO/RO PRIME PRIORITY
		result_elaborate, errs := u.usecase.ResultElaborate(ctx, reqs)
		if errs != nil {
			err = fmt.Errorf("failed get result elaborate")
			return
		}

		if result_elaborate.Decision == constant.DECISION_REJECT {
			max_ltv = result_elaborate.LTV
		}

		updateElaborate.Code = result_elaborate.Code
		updateElaborate.Reason = result_elaborate.Reason
		updateElaborate.Decision = result_elaborate.Decision
	}

	data.Code = updateElaborate.Code
	data.Decision = updateElaborate.Decision
	data.Reason = updateElaborate.Reason
	data.LTV = max_ltv

	resp, _ := json.Marshal(data)

	updateElaborate.RequestID = savedata.RequestID
	updateElaborate.Response = string(resp)

	if err = u.repository.UpdateDataElaborate(updateElaborate); err != nil {
		err = fmt.Errorf("failed update data api elaborate")
		return
	}

	return
}

func (u usecase) ResultElaborate(ctx context.Context, reqs request.BodyRequestElaborate) (data response.ElaborateResult, err error) {

	// convert date to year for age_vehicle
	date_string := reqs.Data.ManufacturingYear
	date_parse, err := time.Parse("2006", date_string)
	if err != nil {
		err = fmt.Errorf("failed process elaborate")
		return
	}

	date_now := time.Now()
	sub_date := date_now.Sub(date_parse)
	age := int((sub_date.Hours()/24)/365) + (reqs.Data.Tenor / 12)
	var age_vehicle string
	if age <= 12 {
		age_vehicle = "<=12"
	} else {
		age_vehicle = ">12"
	}

	// BPKB Name
	var bpkbNameType int
	namaSama := utils.AizuArrayString(os.Getenv("NAMA_SAMA"))
	namaBeda := utils.AizuArrayString(os.Getenv("NAMA_BEDA"))

	bpkb_nama_sama, _ := utils.ItemExists(reqs.Data.BPKBName, namaSama)
	bpkb_nama_beda, _ := utils.ItemExists(reqs.Data.BPKBName, namaBeda)

	if bpkb_nama_sama {
		bpkbNameType = 1
	} else if bpkb_nama_beda {
		bpkbNameType = 0
	}

	status_konsumen := reqs.Data.CustomerStatus
	if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
		status_konsumen = "AO/RO"
	}

	branch_id := reqs.Data.BranchID
	result_pefindo := reqs.Data.ResultPefindo
	tenor := reqs.Data.Tenor
	NTF := reqs.Data.NTF
	OTR := reqs.Data.OTR

	var baki_debet float64
	var ok bool

	if reqs.Data.TotalBakiDebet != nil {
		baki_debet, ok = reqs.Data.TotalBakiDebet.(float64)
		if !ok {
			baki_debet = 0
		}
	}

	var reason = constant.REASON_PASS_ELABORATE

	// Hitung LTV
	ltv := utils.ToFixed((NTF/OTR)*100, 0)

	// Get Cluster Branch
	cluster_branch, err := u.repository.GetClusterBranchElaborate(branch_id, status_konsumen, bpkbNameType)
	if err != nil {
		err = fmt.Errorf("failed get cluster branch elaborate")
		return
	}

	if baki_debet > constant.RANGE_CLUSTER_BAKI_DEBET_REJECT {
		if cluster_branch.Cluster == constant.CLUSTER_E || cluster_branch.Cluster == constant.CLUSTER_F {
			data.Code = constant.CODE_REJECT_ELABORATE
			data.Reason = constant.REASON_REJECT_ELABORATE
			data.Decision = constant.DECISION_REJECT
			return
		}
	} else if baki_debet > constant.BAKI_DEBET { //This should be reject in Filtering tho
		data.Code = constant.CODE_REJECT_ELABORATE
		data.Reason = constant.REASON_REJECT_ELABORATE
		data.Decision = constant.DECISION_REJECT
		return
	}

	// Set Result Pefindo for HIT/NO HIT based on Filtering Result
	filtering_result, err := u.repository.GetFilteringResult(reqs.Data.ProspectID)
	if filtering_result == (entity.ApiDupcheckKmbUpdate{}) {
		result_pefindo = reqs.Data.ResultPefindo
	} else {
		if err != nil {
			err = fmt.Errorf("failed get result filtering")
			return
		}

		if result_pefindo == constant.DECISION_PASS {
			if (filtering_result.PefindoID != nil || filtering_result.PefindoIDSpouse != nil) && filtering_result.PefindoScore != constant.UNSCORE_PBK {
				result_pefindo = constant.DECISION_PASS
			} else {
				result_pefindo = constant.DECISION_PBK_NO_HIT
			}
		}
	}

	// Get Result from Mapping Elaborate
	result_elaborate, err := u.repository.GetResultElaborate(branch_id, status_konsumen, bpkbNameType, result_pefindo, tenor, age_vehicle, ltv, baki_debet)
	if err != nil {
		err = fmt.Errorf("failed get result elaborate")
		return
	}

	if result_elaborate.Decision == constant.DECISION_PASS {
		data.Code = constant.CODE_PASS_ELABORATE
	} else {
		data.Code = constant.CODE_REJECT_ELABORATE
		if result_elaborate.LTV > 0 {
			data.Code = constant.CODE_REJECT_NTF_ELABORATE
			reason = constant.REASON_REJECT_NTF_ELABORATE
			data.LTV = result_elaborate.LTV - 1
		} else {
			reason = constant.REASON_REJECT_ELABORATE
		}
	}
	data.Decision = result_elaborate.Decision
	data.Reason = reason
	return
}
