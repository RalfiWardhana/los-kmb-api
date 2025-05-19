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
	"sort"
	"strconv"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

func (u multiUsecase) GetAvailableTenor(ctx context.Context, req request.GetAvailableTenor, accessToken string) (data []response.GetAvailableTenorData, err error) {

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
		BirthDate:              req.BirthDate,
	}

	marsevFilterProgramRes, err := u.usecase.MarsevGetMarketingProgram(ctx, payloadFilterProgram, req.ProspectID, accessToken)
	if err != nil {
		return data, err
	}

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

		var mdmMasterMappingLicensePlateRes response.MDMMasterMappingLicensePlateResponse
		mdmMasterMappingLicensePlateRes, err = u.usecase.MDMGetMappingLicensePlate(ctx, req.LicensePlate, req.ProspectID, accessToken)
		if err != nil {
			return data, err
		}

		mappingLicensePlate := mdmMasterMappingLicensePlateRes.Data.Records[0]

		payloadCalculate := request.ReqMarsevCalculateInstallment{
			ProgramID:              marsevProgramData.ID,
			BranchID:               req.BranchID,
			CustomerOccupationCode: "KRYSW",
			AssetUsageTypeCode:     req.AssetUsageTypeCode,
			AssetYear:              manufactureYear,
			BpkbStatusCode:         bpkbStatusCode,
			LoanAmount:             req.LoanAmount,
			Otr:                    otr,
			RegionCode:             mappingLicensePlate.AreaID,
			AssetCategory:          categoryId,
			CustomerBirthDate:      req.BirthDate,
			UseAdditionalInsurance: req.IsUseAdditionalInsurance,
		}

		var marsevCalculateInstallmentRes response.MarsevCalculateInstallmentResponse
		marsevCalculateInstallmentRes, err = u.usecase.MarsevCalculateInstallment(ctx, payloadCalculate, req.ProspectID, accessToken)
		if err != nil {
			return data, nil
		}

		if len(marsevCalculateInstallmentRes.Data) == 0 {
			return data, nil
		}

		var mappingElaborateLTV []entity.MappingElaborateLTV
		mappingElaborateLTV, err = u.repository.GetMappingElaborateLTV(resultPefindo, clusterCMO)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error")
			return
		}

		var (
			wg                 sync.WaitGroup
			availableTenorChan = make(chan response.GetAvailableTenorData, len(marsevProgramData.Tenors))
			errChan            = make(chan error, len(marsevProgramData.Tenors))
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

				ltv, _, err := u.usecase.GetLTV(ctx, mappingElaborateLTV, req.ProspectID, resultPefindo, req.BPKBNameType, req.ManufactureYear, tenorInfo.Tenor, bakiDebet, isSimulasi)
				if err != nil {
					errChan <- err
					return
				}

				if ltv == 0 {
					return
				}

				// Get loan amount
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

				dealer := "NON PSA"
				if marsevLoanAmountRes.Data.IsPsa {
					dealer = "PSA"
				}

				if marsevLoanAmountRes.Data.LoanAmountMaximum >= req.LoanAmount {
					for _, installmentData := range marsevCalculateInstallmentRes.Data {
						if installmentData.Tenor == tenorInfo.Tenor {
							availableTenor := response.GetAvailableTenorData{
								Tenor:                    tenorInfo.Tenor,
								IsPsa:                    installmentData.IsPSA,
								Dealer:                   dealer,
								InstallmentAmount:        installmentData.MonthlyInstallment,
								AF:                       installmentData.AmountOfFinance,
								AdminFee:                 installmentData.AdminFee,
								DPAmount:                 installmentData.DPAmount,
								NTF:                      installmentData.NTF,
								AssetCategoryID:          categoryId,
								OTR:                      otr,
								ShowAdditionalInsurance:  tenorInfo.ShowAdditionalInsurance,
								UseAdditionalInsurance:   tenorInfo.UseAdditionalInsurance,
								InsuranceCompanyBranchID: tenorInfo.InsuranceCompanyBranchID,
							}
							availableTenorChan <- availableTenor
						}
					}
				}
			}(tenorInfo)
		}

		go func() {
			wg.Wait()
			close(availableTenorChan)
			close(errChan)
		}()

		if err := <-errChan; err != nil {
			return []response.GetAvailableTenorData{}, err
		}

		var availableTenorList []response.GetAvailableTenorData
		for tenorData := range availableTenorChan {
			availableTenorList = append(availableTenorList, tenorData)
		}

		sort.Slice(availableTenorList, func(i, j int) bool {
			return availableTenorList[i].Tenor < availableTenorList[j].Tenor
		})

		data = append(data, availableTenorList...)
	}

	return
}

func (u usecase) MDMGetMappingLicensePlate(ctx context.Context, licensePlate string, prospectID string, accessToken string) (mdmMasterMappingLicensePlateRes response.MDMMasterMappingLicensePlateResponse, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))
	headerMDM := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": accessToken,
	}

	licensePlateCode := utils.GetLicensePlateCode(licensePlate)
	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MASTER_MAPPING_LICENSE_PLATE_URL")+"?lob_id="+strconv.Itoa(constant.LOBID_KMB)+"&plate_code="+licensePlateCode, nil, headerMDM, constant.METHOD_GET, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping License Plate Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &mdmMasterMappingLicensePlateRes); err != nil {
			return
		}

		if len(mdmMasterMappingLicensePlateRes.Data.Records) == 0 {
			err = errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Mapping License Plate Error Not Found Data")
			return
		}
	}

	return
}

func (u usecase) MarsevCalculateInstallment(ctx context.Context, req request.ReqMarsevCalculateInstallment, prospectID string, accessToken string) (marsevCalculateInstallmentRes response.MarsevCalculateInstallmentResponse, err error) {

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
	}

	param, _ := json.Marshal(req)

	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_CALCULATE_INSTALLMENT_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Marsev Calculate Installment Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &marsevCalculateInstallmentRes); err != nil {
			return
		}
	}

	return
}
