package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

func (u multiUsecase) GetMaxLoanAmout(ctx context.Context, req request.GetMaxLoanAmount, accessToken string) (data response.GetMaxLoanAmountData, err error) {

	var (
		resultPefindo string
		bakiDebet     float64
	)

	config, err := u.repository.GetConfig(constant.GROUP_2WILEN, "KMB-OFF", constant.KEY_PPID_SIMULASI)
	if err != nil {
		return
	}

	re := regexp.MustCompile(config.Value)
	isSimulasi := re.MatchString(req.ProspectID)

	if isSimulasi {
		resultPefindo = constant.DECISION_PASS
		bakiDebet = 0
	} else {
		trxKPM, err := u.repository.GetTrxKPM(req.ProspectID)
		if err != nil {
			return data, err
		}

		resultPefindo = trxKPM.ResultPefindo
		bakiDebet = trxKPM.BakiDebet
	}

	// get data customer
	dataCustomer, err := u.usecase.DupcheckIntegrator(ctx, req.ProspectID, req.IDNumber, req.LegalName, req.BirthDate, req.SurgateMotherName, accessToken)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get Data Customer Error")
		return data, err
	}

	if dataCustomer.CustomerStatus == "" || dataCustomer.CustomerStatus == constant.STATUS_KONSUMEN_NEW {
		dataCustomer.CustomerStatus = constant.STATUS_KONSUMEN_NEW
		dataCustomer.CustomerSegment = constant.RO_AO_REGULAR
	}

	// get asset
	assetList, err := u.usecase.MDMGetMasterAsset(ctx, req.BranchID, req.AssetCode, req.ProspectID, accessToken)
	if err != nil {
		return data, err
	}

	var categoryId, brand string
	if len(assetList.Records) > 0 {
		categoryId = assetList.Records[0].CategoryID
		brand = assetList.Records[0].Brand
	}

	// get marketing program
	bpkbStatusCode := "DN"
	if strings.Contains(os.Getenv("NAMA_SAMA"), req.BPKBNameType) {
		bpkbStatusCode = "SN"
	}

	customerType := utils.CapitalizeEachWord(dataCustomer.CustomerStatus)
	if dataCustomer.CustomerStatus != constant.STATUS_KONSUMEN_NEW {
		customerType = constant.STATUS_KONSUMEN_RO_AO + " " + utils.CapitalizeEachWord(dataCustomer.CustomerSegment)
		if dataCustomer.CustomerSegment == constant.RO_AO_REGULAR {
			customerType = constant.STATUS_KONSUMEN_RO_AO + " Standard"
		}
	}

	manufactureYear, _ := strconv.Atoi(req.ManufactureYear)

	financeType := "PM"

	payloadFilterProgram := request.ReqMarsevFilterProgram{
		Page:                   1,
		Limit:                  10,
		BranchID:               req.BranchID,
		FinancingTypeCode:      financeType,
		CustomerOccupationCode: "KRYSW",
		BpkbStatusCode:         bpkbStatusCode,
		SourceApplication:      constant.MARSEV_SOURCE_APPLICATION_KPM,
		CustomerType:           customerType,
		AssetUsageTypeCode:     req.AssetUsageTypeCode,
		AssetCategory:          categoryId,
		AssetBrand:             brand,
		AssetYear:              manufactureYear,
		LoanAmount:             2000000,
		SalesMethodID:          5,
	}

	marsevFilterProgramRes, err := u.usecase.MarsevGetMarketingProgram(ctx, payloadFilterProgram, req.ProspectID, accessToken)
	if err != nil {
		return data, err
	}

	var maxLoanAmount float64
	if len(marsevFilterProgramRes.Data) > 0 {
		marsevProgramData := marsevFilterProgramRes.Data[0]

		if req.ReferralCode != nil && *req.ReferralCode != "" {
			miNumbers := strings.Split(os.Getenv("MI_NUMBER_WHITELIST"), ",")
			miNumberSet := make(map[int]struct{}, len(miNumbers))
			for _, miNumber := range miNumbers {
				miNumberInt, _ := strconv.Atoi(miNumber)
				miNumberSet[miNumberInt] = struct{}{}
			}

			found := false
			for _, datum := range marsevFilterProgramRes.Data {
				if _, exists := miNumberSet[datum.MINumber]; exists {
					marsevProgramData = datum
					found = true
					break
				}
			}

			if !found {
				return data, errors.New(constant.ERROR_BAD_REQUEST + " - No matching MI_NUMBER found")
			}
		}

		var assetMP response.AssetYearList
		assetMP, err = u.usecase.MDMGetAssetYear(ctx, req.BranchID, req.AssetCode, req.ManufactureYear, req.ProspectID, accessToken)
		if err != nil {
			return data, err
		}

		var otr float64
		if len(assetMP.Records) > 0 {
			otr = float64(assetMP.Records[0].MarketPriceValue)
		}

		var respCmoBranch response.MDMMasterMappingBranchEmployeeResponse
		respCmoBranch, err = u.usecase.MDMGetMasterMappingBranchEmployee(ctx, req.ProspectID, req.BranchID, accessToken)
		if err != nil {
			return data, err
		}

		var cmoID string
		if len(respCmoBranch.Data) > 0 {
			cmoID = respCmoBranch.Data[0].CMOID
		} else {
			err = errors.New(constant.ERROR_UPSTREAM + " - CMO Dedicated Not Found")
			return data, err
		}

		var hrisCMO response.EmployeeCMOResponse
		hrisCMO, err = u.usecase.GetEmployeeData(ctx, cmoID)
		if err != nil {
			return data, err
		}

		if hrisCMO.CMOCategory == "" {
			err = errors.New(constant.ERROR_UPSTREAM + " - CMO Not Found")
			return data, err
		}

		bpkbName := strings.Contains(os.Getenv("NAMA_SAMA"), req.BPKBNameType)

		cluster := constant.CLUSTER_C
		bpkb := constant.BPKB_NAMA_BEDA
		if bpkbName {
			bpkb = constant.BPKB_NAMA_SAMA
			cluster = constant.CLUSTER_B
		}

		var clusterCMO string
		var useDefaultCluster bool

		var mdmFPD response.FpdCMOResponse
		if hrisCMO.CMOCategory == constant.NEW {
			clusterCMO = cluster
			// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak bulan join_date nya
			useDefaultCluster = true
		} else {
			// Mendapatkan value FPD dari masing-masing jenis BPKB
			mdmFPD, err = u.usecase.GetFpdCMO(ctx, cmoID, bpkb)
			if err != nil {
				return data, err
			}

			if !mdmFPD.FpdExist {
				clusterCMO = cluster
				// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak tanggal hit filtering nya (assume: today)
				useDefaultCluster = true
			} else {
				// Check Cluster
				var mappingFpdCluster entity.MasterMappingFpdCluster
				mappingFpdCluster, err = u.repository.MasterMappingFpdCluster(mdmFPD.CmoFpd)
				if err != nil {
					return data, err
				}

				if mappingFpdCluster.Cluster == "" {
					clusterCMO = cluster
					// set cluster menggunakan Default Cluster selama 3 bulan, terhitung sejak tanggal hit filtering nya (assume: today)
					useDefaultCluster = true
				} else {
					clusterCMO = mappingFpdCluster.Cluster
				}
			}
		}

		var savedCluster string
		if useDefaultCluster {
			savedCluster, _, err = u.usecase.CheckCmoNoFPD(req.ProspectID, cmoID, hrisCMO.CMOCategory, hrisCMO.JoinDate, clusterCMO, bpkb)
			if err != nil {
				return data, err
			}
			if savedCluster != "" {
				clusterCMO = savedCluster
			}
		}

		detailTrxBiro, err := u.repository.GetTrxDetailBIro(req.ProspectID)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Trx Detail Biro Error")
			return data, err
		}

		pbkScore := "BAD"
		for _, v := range detailTrxBiro {
			if v.Score == "NO HIT" {
				pbkScore = "NO HIT"
				break
			}
		}
		if pbkScore == "BAD" {
			for _, v := range detailTrxBiro {
				if v.Score == "AVERAGE RISK" || v.Score == "LOW RISK" || v.Score == "VERY LOW RISK" {
					pbkScore = "GOOD"
					break
				}
			}
		}

		branch, err := u.repository.GetMappingBranchByBranchID(req.BranchID, pbkScore)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Branch Error")
			return data, err
		}

		customerStatus := dataCustomer.CustomerStatus

		var mappingElaborateLTV []entity.MappingElaborateLTV
		mappingElaborateLTV, err = u.repository.GetMappingElaborateLTV(resultPefindo, clusterCMO, branch.GradeBranch)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error")
			return data, err
		}

		var (
			wg             sync.WaitGroup
			loanAmountChan = make(chan float64, len(marsevProgramData.Tenors))
			errChan        = make(chan error, len(marsevProgramData.Tenors))
		)

		for _, tenorInfo := range marsevProgramData.Tenors {
			wg.Add(1)
			go func(tenorInfo response.TenorInfo) {
				defer wg.Done()

				if tenorInfo.Tenor >= 36 {
					var trxTenor response.UsecaseApi
					if tenorInfo.Tenor == 36 {
						trxTenor, err = u.usecase.RejectTenor36(clusterCMO)
						if err != nil {
							errChan <- err
							return
						}
					} else if tenorInfo.Tenor > 36 {
						return
					}
					if trxTenor.Result == constant.DECISION_REJECT {
						return
					}
				}

				ltv, _, err := u.usecase.GetLTV(ctx, mappingElaborateLTV, req.ProspectID, resultPefindo, req.BPKBNameType, req.ManufactureYear, tenorInfo.Tenor, bakiDebet, isSimulasi, pbkScore, customerStatus)
				if err != nil {
					errChan <- err
					return
				}

				if ltv == 0 {
					return
				}

				// get loan amount
				payloadMaxLoan := request.ReqMarsevLoanAmount{
					BranchID:      req.BranchID,
					OTR:           otr,
					MaxLTV:        ltv,
					IsRecalculate: false,
				}

				marsevLoanAmountRes, err := u.usecase.MarsevGetLoanAmount(ctx, payloadMaxLoan, req.ProspectID, accessToken)
				if err != nil {
					errChan <- err
					return
				}

				loanAmountChan <- marsevLoanAmountRes.Data.LoanAmountMaximum
			}(tenorInfo)
		}

		go func() {
			wg.Wait()
			close(loanAmountChan)
			close(errChan)
		}()

		if err := <-errChan; err != nil {
			return response.GetMaxLoanAmountData{}, err
		}

		for loanAmount := range loanAmountChan {
			if loanAmount > maxLoanAmount {
				maxLoanAmount = loanAmount
			}
		}
	}

	data.MaxLoanAmount = maxLoanAmount

	return
}

func (u usecase) GetLTV(ctx context.Context, mappingElaborateLTV []entity.MappingElaborateLTV, prospectID, resultPefindo, bpkbName, manufactureYear string, tenor int, bakiDebet float64, isSimulasi bool, pbkScore, customerStatus string) (ltv int, adjustTenor bool, err error) {
	var bpkbNameType int
	if strings.Contains(os.Getenv("NAMA_SAMA"), bpkbName) {
		bpkbNameType = 1
	}

	now := time.Now()
	// convert date to year for age_vehicle
	manufacturingYear, err := time.Parse("2006", manufactureYear)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Format tahun kendaraan tidak sesuai")
		return
	}

	subManufacturingYear := now.Sub(manufacturingYear)
	age := int((subManufacturingYear.Hours()/24)/365) + (tenor / 12)

	var ageS string
	if age <= 12 {
		ageS = "<=12"
	} else {
		ageS = ">12"
	}

	trxElaborateLTV := entity.TrxElaborateLTV{
		ProspectID:        prospectID,
		RequestID:         ctx.Value(echo.HeaderXRequestID).(string),
		Tenor:             tenor,
		ManufacturingYear: manufactureYear,
	}

	isFixedLtv := false
	maxTenor := 0
	for _, m := range mappingElaborateLTV {
		if !isFixedLtv {
			if tenor >= 36 {
				//no hit
				if resultPefindo == constant.DECISION_PBK_NO_HIT && m.TenorStart <= tenor && tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
					ltv = m.LTV
					trxElaborateLTV.MappingElaborateLTVID = m.ID
				}

				//pass
				if resultPefindo == constant.DECISION_PASS && m.TenorStart <= tenor && tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
					ltv = m.LTV
					trxElaborateLTV.MappingElaborateLTVID = m.ID
				}

				//reject
				if resultPefindo == constant.DECISION_REJECT && m.TotalBakiDebetStart <= int(bakiDebet) && int(bakiDebet) <= m.TotalBakiDebetEnd && m.TenorStart <= tenor && tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
					ltv = m.LTV
					trxElaborateLTV.MappingElaborateLTVID = m.ID
				}

			} else if tenor == 18 && m.TenorStart <= tenor && tenor <= m.TenorEnd && m.StatusKonsumen == "NEW" && bpkbNameType == 0 && m.PbkScore == pbkScore &&
				m.GradeBranch == "BAD" {
				ltv = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
				isFixedLtv = true
			} else {
				//no hit
				if resultPefindo == constant.DECISION_PBK_NO_HIT && m.TenorStart <= tenor && tenor <= m.TenorEnd {
					ltv = m.LTV
					trxElaborateLTV.MappingElaborateLTVID = m.ID
				}

				//pass
				if resultPefindo == constant.DECISION_PASS && m.TenorStart <= tenor && tenor <= m.TenorEnd {
					if m.BPKBNameType == 1 {
						if bpkbNameType == m.BPKBNameType {
							ltv = m.LTV
							trxElaborateLTV.MappingElaborateLTVID = m.ID
						}
					} else {
						ltv = m.LTV
						trxElaborateLTV.MappingElaborateLTVID = m.ID
					}
				}

				//reject
				if resultPefindo == constant.DECISION_REJECT && m.TotalBakiDebetStart <= int(bakiDebet) && int(bakiDebet) <= m.TotalBakiDebetEnd && m.TenorStart <= tenor && tenor <= m.TenorEnd {
					ltv = m.LTV
					trxElaborateLTV.MappingElaborateLTVID = m.ID
				}
			}
		}

		// max tenor
		if m.TenorEnd >= maxTenor && m.LTV > 0 {
			if m.AgeVehicle != "" {
				if bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
					maxTenor = m.TenorEnd
					adjustTenor = true
				}
			} else if m.BPKBNameType == 1 {
				if bpkbNameType == m.BPKBNameType {
					maxTenor = m.TenorEnd
					adjustTenor = true
				}
			} else {
				maxTenor = m.TenorEnd
				adjustTenor = true
			}
		}
	}

	if !isSimulasi {
		err = u.repository.SaveTrxElaborateLTV(trxElaborateLTV)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - save elaborate ltv error")
			return
		}
	}

	return
}

func (u usecase) MarsevGetLoanAmount(ctx context.Context, req request.ReqMarsevLoanAmount, prospectID string, accessToken string) (marsevLoanAmountRes response.MarsevLoanAmountResponse, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
	}

	param, _ := json.Marshal(req)

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_LOAN_AMOUNT_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Loan Amount Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &marsevLoanAmountRes); err != nil {
			return
		}
	}

	return
}

func (u usecase) MarsevGetMarketingProgram(ctx context.Context, req request.ReqMarsevFilterProgram, prospectID string, accessToken string) (marsevFilterProgramRes response.MarsevFilterProgramResponse, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
	}

	param, _ := json.Marshal(req)

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_FILTER_PROGRAM_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Filter Program Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &marsevFilterProgramRes); err != nil {
			return
		}
	}

	return
}

func (u usecase) MDMGetMasterAsset(ctx context.Context, branchID string, search string, prospectID string, accessToken string) (assetList response.AssetList, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	payloadAsset, _ := json.Marshal(map[string]interface{}{
		"branch_id": branchID,
		"lob_id":    11,
		"page_size": 10,
		"search":    search,
	})

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_ASSET_URL"), payloadAsset, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
		return
	}

	if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &assetList); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data asset list")
		return
	}

	if len(assetList.Records) == 0 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
		return
	}

	return
}

func (u usecase) MDMGetAssetYear(ctx context.Context, branchID string, assetCode string, search string, prospectID string, accessToken string) (assetMP response.AssetYearList, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))
	param, _ := json.Marshal(map[string]interface{}{
		"branch_id":  branchID,
		"asset_code": assetCode,
		"search":     search,
	})

	var resp *resty.Response

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MARKETPRICE_URL"), param, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Timeout")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error")
		return
	}

	if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &assetMP); err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data marketprice")
		return
	}

	if len(assetMP.Records) == 0 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error")
		return
	}

	return
}
