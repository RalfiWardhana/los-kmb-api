package usecase

import (
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

func (u multiUsecase) Elaborate(reqs request.BodyRequestElaborate, accessToken string) (data response.ElaborateResult, err error) {

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
			result_elaborate, errs := u.usecase.ResultElaborate(reqs)
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
		result_elaborate, errs := u.usecase.ResultElaborate(reqs)
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

func (u usecase) ResultElaborate(reqs request.BodyRequestElaborate) (data response.ElaborateResult, err error) {

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
	baki_debet := utils.ToFixed(reqs.Data.TotalBakiDebet, 0)

	var reason = constant.REASON_PASS_ELABORATE

	// Hitung LTV
	ltv := utils.ToFixed((NTF/OTR)*100, 0)

	// Get Cluster Branch
	cluster_branch, err := u.GetClusterBranchElaborate(branch_id, status_konsumen, bpkbNameType)
	if err != nil {
		err = fmt.Errorf("failed get cluster branch elaborate")
		return
	}

	if reqs.Data.TotalBakiDebet > constant.RANGE_CLUSTER_BAKI_DEBET_REJECT {
		if cluster_branch.Cluster == constant.CLUSTER_E || cluster_branch.Cluster == constant.CLUSTER_F {
			data.Code = constant.CODE_REJECT_ELABORATE
			data.Reason = constant.REASON_REJECT_ELABORATE
			data.Decision = constant.DECISION_REJECT
			return
		}
	} else if reqs.Data.TotalBakiDebet > constant.BAKI_DEBET { //This should be reject in Filtering tho
		data.Code = constant.CODE_REJECT_ELABORATE
		data.Reason = constant.REASON_REJECT_ELABORATE
		data.Decision = constant.DECISION_REJECT
		return
	}

	// Get Result from Mapping Elaborate
	result_elaborate, err := u.GetResultElaborate(branch_id, status_konsumen, bpkbNameType, result_pefindo, tenor, age_vehicle, ltv, baki_debet)
	if err != nil {
		err = fmt.Errorf("failed get result elaborate")
		return
	}

	if result_elaborate.Decision == constant.DECISION_PASS {
		data.Code = constant.CODE_PASS_ELABORATE
	} else {
		data.Code = constant.CODE_REJECT_NTF_ELABORATE
		if result_elaborate.LTV > 0 {
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

func (u usecase) GetClusterBranchElaborate(branch_id string, cust_status string, bpkb int) (cluster entity.ClusterBranch, err error) {

	query := fmt.Sprintf("SELECT cluster FROM kmb_mapping_cluster_branch WHERE branch_id = '%s' AND customer_status = '%s' AND bpkb_name_type = %d", branch_id, cust_status, bpkb)

	cluster, err = u.repository.GetClusterBranchElaborate(query)

	if err != nil {
		return
	}

	return
}

func (u usecase) GetResultElaborate(branch_id string, cust_status string, bpkb int, result_pefindo string, tenor int, age_vehicle string, ltv float64, baki_debet float64) (data entity.ResultElaborate, err error) {

	var queryAdd string
	var ltv_range int = int(ltv)
	var total_baki_debet int = int(baki_debet)

	// PEFINDO PASS
	if result_pefindo == "PASS" {
		if tenor >= 36 {
			queryAdd = fmt.Sprintf("AND mes.bpkb_name_type = %d AND mes.tenor_start >= 36 AND mes.tenor_end = 0", bpkb)
		} else {
			queryAdd = fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if tenor >= 36 && bpkb == 1 {
			queryAdd += fmt.Sprintf(" AND mes.age_vehicle = '%s'", age_vehicle)
		}

		if age_vehicle == "<=12" && bpkb == 1 {

			if ltv_range != 0 && ltv_range <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", ltv_range, ltv_range)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", ltv_range)
			}
		}
	}

	// PEFINDO NO HIT
	if result_pefindo == "NO HIT" {
		if tenor >= 24 {
			queryAdd = fmt.Sprintf("AND mes.bpkb_name_type = %d AND mes.tenor_start >= 24 AND mes.tenor_end = 0", bpkb)
		} else {
			queryAdd = fmt.Sprintf("AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if ltv_range != 0 && ltv_range <= 1000 {
			queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", ltv_range, ltv_range)
		} else {
			queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= 1000", ltv_range)
		}
	}

	// PEFINDO REJECT
	if result_pefindo == "REJECT" {

		queryAdd = fmt.Sprintf("AND mes.total_baki_debet_start <= %d AND mes.total_baki_debet_end >= %d", total_baki_debet, total_baki_debet)

		if tenor >= 24 {
			queryAdd += fmt.Sprintf(" AND mes.bpkb_name_type = %d AND mes.tenor_start >= '24' AND mes.tenor_end = 0", bpkb)
		} else {
			queryAdd += fmt.Sprintf(" AND mes.tenor_start <= %d AND mes.tenor_end >= %d", tenor, tenor)
		}

		if tenor < 24 {
			if ltv_range != 0 && ltv_range <= 1000 {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= %d", ltv_range, ltv_range)
			} else {
				queryAdd += fmt.Sprintf(" AND mes.ltv_start <= %d AND mes.ltv_end >= '1000'", ltv_range)
			}
		}
	}

	query := fmt.Sprintf("SELECT mcb.cluster, mes.decision, mes.ltv_start FROM kmb_mapping_cluster_branch mcb JOIN kmb_mapping_elaborate_scheme mes ON mcb.cluster = mes.cluster WHERE mcb.branch_id = '%s' AND mcb.customer_status = '%s' AND mcb.bpkb_name_type = %d AND mes.result_pefindo = '%s' %s", branch_id, cust_status, bpkb, result_pefindo, queryAdd)

	data, err = u.repository.GetResultElaborate(query)

	if err != nil {
		return
	}

	return
}
