package usecase

import (
	"context"
	"encoding/json"
	"errors"
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

func NewMultiUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, usecase interfaces.Usecase) (interfaces.MultiUsecase, interfaces.Usecase) {
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

	var (
		max_ltv         int
		savedata        entity.ApiElaborateKmb
		updateElaborate entity.ApiElaborateKmbUpdate
		parameter       response.ResponseMappingElaborateScheme
	)

	id := uuid.New()

	savedata.RequestID = id.String()

	request, _ := json.Marshal(reqs)
	savedata.Request = string(request)
	savedata.ProspectID = reqs.Data.ProspectID
	if err = u.repository.SaveDataElaborate(savedata); err != nil {
		err = fmt.Errorf("failed process elaborate")
		return
	}

	status_konsumen := reqs.Data.CustomerStatus
	if status_konsumen == constant.STATUS_KONSUMEN_RO_AO {
		status_konsumen = "AO/RO"
	}

	kategoriStatusKonsumen := reqs.Data.CategoryCustomer
	checkPrimePriority, _ := utils.ItemExists(kategoriStatusKonsumen, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

	parameter.BranchID = reqs.Data.BranchID

	if reqs.Data.CustomerStatus == constant.STATUS_KONSUMEN_RO_AO && checkPrimePriority {
		parameter.BranchIDMask = constant.BRANCH_ID_PRIME_PRIORITY
		reqs.Data.BranchID = constant.BRANCH_ID_PRIME_PRIORITY
	}

	tenor := reqs.Data.Tenor

	// get the result elaborated scheme
	resultElaborate, err := u.usecase.ResultElaborate(ctx, reqs)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - " + err.Error())
		return
	}

	updateElaborate.IsMapping = 1 //default, for mapping is exist

	// not found mapping elaborate scheme
	if resultElaborate.Decision == "" {
		updateElaborate.IsMapping = 0 //set flag for mapping not found
		updateElaborate.Code = constant.CODE_PASS_ELABORATE
		updateElaborate.Decision = constant.DECISION_PASS
		updateElaborate.Reason = constant.REASON_PASS_ELABORATE
	} else {
		if resultElaborate.Decision == constant.DECISION_REJECT {
			max_ltv = resultElaborate.LTV
		}

		updateElaborate.Code = resultElaborate.Code
		updateElaborate.Reason = resultElaborate.Reason
		updateElaborate.Decision = resultElaborate.Decision
	}

	data.Code = updateElaborate.Code
	data.Decision = updateElaborate.Decision
	data.Reason = updateElaborate.Reason
	data.LTV = max_ltv

	resp, _ := json.Marshal(data)

	updateElaborate.RequestID = savedata.RequestID
	updateElaborate.Response = string(resp)

	parameter.ResultPefindo = resultElaborate.ResultPefindo
	parameter.CustomerStatus = status_konsumen
	parameter.BPKBNameType = resultElaborate.BPKBNameType
	parameter.Cluster = resultElaborate.Cluster
	parameter.Tenor = tenor
	parameter.AgeVehicle = resultElaborate.AgeVehicle
	parameter.LTV = resultElaborate.LTVOrigin
	parameter.TotalBakiDebet = int(resultElaborate.TotalBakiDebet)
	parameter.Decision = resultElaborate.Decision

	if resultElaborate.IsMappingOvd {
		parameter.IsMappingOvd = "Yes"
	} else {
		parameter.IsMappingOvd = "No"
	}

	mapping, _ := json.Marshal(parameter)
	updateElaborate.MappingParameter = string(mapping)

	if err = u.repository.UpdateDataElaborate(updateElaborate); err != nil {
		err = fmt.Errorf("failed update data api elaborate")
		return
	}

	return
}

func (u usecase) ResultElaborate(ctx context.Context, reqs request.BodyRequestElaborate) (data response.ElaborateResult, err error) {

	var (
		ageVehicle       string
		bpkbNameType     int
		bakiDebet        float64
		filteringResult  entity.ApiDupcheckKmbUpdate
		getMappingLtvOvd entity.ResultElaborate
	)

	// convert date to year for age_vehicle
	dateString := reqs.Data.ManufacturingYear
	dateParse, err := time.Parse("2006", dateString)
	if err != nil {
		err = fmt.Errorf("error parsing manufacturing year")
		return
	}

	dateNow := time.Now()
	subDate := dateNow.Sub(dateParse)
	age := int((subDate.Hours()/24)/365) + (reqs.Data.Tenor / 12)

	// age vehicle
	if age <= 12 {
		ageVehicle = "<=12"
	} else {
		ageVehicle = ">12"
	}

	// BPKB Name
	namaSama := utils.AizuArrayString(os.Getenv("NAMA_SAMA"))
	namaBeda := utils.AizuArrayString(os.Getenv("NAMA_BEDA"))

	bpkbNamaSama, _ := utils.ItemExists(reqs.Data.BPKBName, namaSama)
	bpkbNamaBeda, _ := utils.ItemExists(reqs.Data.BPKBName, namaBeda)

	if bpkbNamaSama {
		bpkbNameType = 1
	} else if bpkbNamaBeda {
		bpkbNameType = 0
	}

	statusKonsumen := reqs.Data.CustomerStatus
	if statusKonsumen == constant.STATUS_KONSUMEN_RO_AO {
		statusKonsumen = "AO/RO"
	}

	branchId := reqs.Data.BranchID
	resultPefindo := reqs.Data.ResultPefindo
	tenor := reqs.Data.Tenor
	NTF := reqs.Data.NTF
	OTR := reqs.Data.OTR

	var ok bool
	if reqs.Data.TotalBakiDebet != nil {
		bakiDebet, ok = reqs.Data.TotalBakiDebet.(float64)
		if !ok {
			bakiDebet = 0
		}
	}

	var reason = constant.REASON_PASS_ELABORATE

	// Hitung LTV
	ltv := utils.ToFixed((NTF/OTR)*100, 0)

	data.ResultPefindo = resultPefindo
	data.BPKBNameType = bpkbNameType
	data.AgeVehicle = ageVehicle
	data.LTVOrigin = ltv
	data.TotalBakiDebet = bakiDebet

	// Get Cluster Branch
	clusterBranch, err := u.repository.GetClusterBranchElaborate(branchId, statusKonsumen, bpkbNameType)
	if err != nil && err.Error() != constant.ERROR_NOT_FOUND {
		err = fmt.Errorf("failed get cluster branch elaborate")
		return
	}

	if clusterBranch != (entity.ClusterBranch{}) {
		data.Cluster = clusterBranch.Cluster
	}

	// Set Result Pefindo for HIT/NO HIT based on Filtering Result
	kategoriStatusKonsumen := reqs.Data.CategoryCustomer
	checkPrimePriority, _ := utils.ItemExists(kategoriStatusKonsumen, []string{constant.RO_AO_PRIME, constant.RO_AO_PRIORITY})

	filteringResult, err = u.repository.GetFilteringResult(reqs.Data.ProspectID)
	if err != nil {
		err = fmt.Errorf("failed retrieve filtering result")
		return
	}

	if !checkPrimePriority {
		if filteringResult == (entity.ApiDupcheckKmbUpdate{}) {
			resultPefindo = reqs.Data.ResultPefindo
		} else {
			if filteringResult.ResultPefindoIncludeAll != "" {
				resultPefindo = filteringResult.ResultPefindoIncludeAll
			} else {
				if *filteringResult.PefindoScore == constant.UNSCORE_PBK {
					resultPefindo = constant.DECISION_PBK_NO_HIT
				} else {
					resultPefindo = constant.DECISION_PASS
				}
			}
		}
	}
	data.ResultPefindo = resultPefindo

	// Get Result from Mapping Elaborate
	resultElaborate, err := u.repository.GetResultElaborate(branchId, statusKonsumen, bpkbNameType, resultPefindo, tenor, ageVehicle, ltv, bakiDebet)
	if err != nil && err.Error() != constant.RECORD_NOT_FOUND {
		err = fmt.Errorf("failed get mapping elaborate ltv")
		return
	}

	data.IsMappingOvd = false
	if (tenor <= 35) || (tenor >= 36 && bpkbNamaSama && age <= 12) { //excepting check age vehicle
		maxOverdueBiro := filteringResult.MaxOverdue
		maxOverdueLast12 := filteringResult.MaxOverdueLast12Months

		// Check max OVD 12 & max OVD current
		if !filteringResult.IsNullMaxOverdue && !filteringResult.IsNullMaxOverdueLast12Months {
			if maxOverdueBiro == 0 && maxOverdueLast12 <= 10 {
				// Get OVD Mapping LTV
				getMappingLtvOvd, _ = u.repository.GetMappingLtvOvd(clusterBranch.Cluster, resultPefindo, tenor, ltv)

				// recent mapping ltv will replaced with ovd mapping ltv
				if getMappingLtvOvd != (entity.ResultElaborate{}) {
					resultElaborate = getMappingLtvOvd
					data.IsMappingOvd = true
				}
			}
		}
	}

	if resultElaborate.Decision == constant.DECISION_PASS {
		data.Code = constant.CODE_PASS_ELABORATE
	} else {
		data.Code = constant.CODE_REJECT_ELABORATE
		if resultElaborate.LTV > 0 {
			data.Code = constant.CODE_REJECT_NTF_ELABORATE
			reason = constant.REASON_REJECT_NTF_ELABORATE
			data.LTV = resultElaborate.LTV - 1
		} else {
			reason = constant.REASON_REJECT_ELABORATE
		}
	}
	data.Decision = resultElaborate.Decision
	data.Reason = reason
	return
}
