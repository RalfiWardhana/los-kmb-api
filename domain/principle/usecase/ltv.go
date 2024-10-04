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
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
)

func (u usecase) PrincipleElaborateLTV(ctx context.Context, reqs request.PrincipleElaborateLTV, accessToken string) (data response.PrincipleElaborateLTV, err error) {

	var (
		filteringKMB                    entity.FilteringKMB
		ageS                            string
		bakiDebet                       float64
		bpkbNameType                    int
		manufacturingYear               time.Time
		mappingElaborateLTV             []entity.MappingElaborateLTV
		cluster                         string
		RrdDateString                   string
		CreatedAtString                 string
		RrdDate                         time.Time
		CreatedAt                       time.Time
		MonthsOfExpiredContract         int
		OverrideFlowLikeRegular         bool
		expiredContractConfig           entity.AppConfig
		principleStepOne                entity.TrxPrincipleStepOne
		principleStepTwo                entity.TrxPrincipleStepTwo
		assetMP                         response.AssetYearList
		assetList                       response.AssetList
		marsevLoanAmountRes             response.MarsevLoanAmountResponse
		marsevFilterProgramRes          response.MarsevFilterProgramResponse
		marsevCalculateInstallmentRes   response.MarsevCalculateInstallmentResponse
		mdmMasterMappingLicensePlateRes response.MDMMasterMappingLicensePlateResponse
		otr                             float64
	)

	principleStepOne, err = u.repository.GetPrincipleStepOne(reqs.ProspectID)
	if err != nil {
		return
	}

	principleStepTwo, err = u.repository.GetPrincipleStepTwo(reqs.ProspectID)
	if err != nil {
		return
	}

	filteringKMB, err = u.repository.GetFilteringResult(reqs.ProspectID)
	if err != nil {
		if err.Error() == constant.RECORD_NOT_FOUND {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Silahkan melakukan filtering terlebih dahulu")
		} else {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error")
		}
		return
	}

	if filteringKMB.NextProcess != 1 {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses")
		return
	}

	resultPefindo := filteringKMB.Decision

	if filteringKMB.CustomerSegment != nil && strings.Contains("PRIME PRIORITY", filteringKMB.CustomerSegment.(string)) {
		cluster = constant.CLUSTER_PRIME_PRIORITY

		// Cek apakah customer RO PRIME/PRIORITY ini termasuk jalur `expired_contract tidak <= 6 bulan`
		if filteringKMB.CustomerStatus == constant.STATUS_KONSUMEN_RO {
			if filteringKMB.RrdDate == nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty")
				return
			}

			RrdDateString = filteringKMB.RrdDate.(string)
			CreatedAtString = filteringKMB.CreatedAt.Format(time.RFC3339)

			RrdDate, err = time.Parse(time.RFC3339, RrdDateString)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of RrdDate (" + RrdDateString + ")")
				return
			}

			CreatedAt, err = time.Parse(time.RFC3339, CreatedAtString)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of CreatedAt (" + CreatedAtString + ")")
				return
			}

			MonthsOfExpiredContract, err = utils.PreciseMonthsDifference(RrdDate, CreatedAt)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Difference of months RrdDate and CreatedAt is negative (-)")
				return
			}

			// Get config expired_contract
			expiredContractConfig, err = u.repository.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Get Expired Contract Config Error")
				return
			}

			var configValueExpContract response.ExpiredContractConfig
			json.Unmarshal([]byte(expiredContractConfig.Value), &configValueExpContract)

			if configValueExpContract.Data.ExpiredContractCheckEnabled && !(MonthsOfExpiredContract <= configValueExpContract.Data.ExpiredContractMaxMonths) {
				// Jalur mirip seperti customer segment "REGULAR"
				OverrideFlowLikeRegular = true
			}
		}
	} else {
		cluster = filteringKMB.CMOCluster.(string)
	}

	if (filteringKMB.CustomerSegment != nil && !strings.Contains("PRIME PRIORITY", filteringKMB.CustomerSegment.(string))) || (OverrideFlowLikeRegular) {
		if filteringKMB.ScoreBiro == nil || filteringKMB.ScoreBiro == "" || filteringKMB.ScoreBiro == constant.UNSCORE_PBK {
			resultPefindo = constant.DECISION_PBK_NO_HIT
		} else if filteringKMB.MaxOverdueBiro != nil || filteringKMB.MaxOverdueLast12monthsBiro != nil {
			// KO Rules Include All
			ovdCurrent, _ := filteringKMB.MaxOverdueBiro.(int64)
			ovd12, _ := filteringKMB.MaxOverdueLast12monthsBiro.(int64)

			if ovdCurrent > constant.PBK_OVD_CURRENT || ovd12 > constant.PBK_OVD_LAST_12 {
				resultPefindo = constant.DECISION_REJECT
			}
		}
	}

	if strings.Contains(os.Getenv("NAMA_SAMA"), filteringKMB.BpkbName) {
		bpkbNameType = 1
	}

	now := time.Now()
	// convert date to year for age_vehicle
	manufacturingYear, err = time.Parse("2006", principleStepOne.ManufactureYear)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Format tahun kendaraan tidak sesuai")
		return
	}
	subManufacturingYear := now.Sub(manufacturingYear)
	age := int((subManufacturingYear.Hours()/24)/365) + (reqs.Tenor / 12)
	if age <= 12 {
		ageS = "<=12"
	} else {
		ageS = ">12"
	}

	if filteringKMB.TotalBakiDebetNonCollateralBiro != nil {
		bakiDebet, err = utils.GetFloat(filteringKMB.TotalBakiDebetNonCollateralBiro)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " baki debet " + err.Error())
			return
		}
	}

	trxElaborateLTV := entity.TrxElaborateLTV{
		ProspectID:        reqs.ProspectID,
		RequestID:         ctx.Value(echo.HeaderXRequestID).(string),
		Tenor:             reqs.Tenor,
		ManufacturingYear: principleStepOne.ManufactureYear,
	}

	if OverrideFlowLikeRegular && resultPefindo == constant.DECISION_REJECT {
		cluster = filteringKMB.CustomerStatus.(string) + " " + constant.CLUSTER_PRIME_PRIORITY
		if int(bakiDebet) > constant.RANGE_CLUSTER_BAKI_DEBET_REJECT {
			cluster = filteringKMB.CustomerStatus.(string) + " " + filteringKMB.CustomerSegment.(string)
		}
	}

	mappingElaborateLTV, err = u.repository.GetMappingElaborateLTV(resultPefindo, cluster)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error")
		return
	}

	for _, m := range mappingElaborateLTV {
		if reqs.Tenor >= 36 {
			//no hit
			if resultPefindo == constant.DECISION_PBK_NO_HIT && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}

			//pass
			if resultPefindo == constant.DECISION_PASS && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}

			//reject
			if resultPefindo == constant.DECISION_REJECT && m.TotalBakiDebetStart <= int(bakiDebet) && int(bakiDebet) <= m.TotalBakiDebetEnd && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}
		} else {
			//no hit
			if resultPefindo == constant.DECISION_PBK_NO_HIT && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}

			//pass
			if resultPefindo == constant.DECISION_PASS && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd {
				if m.BPKBNameType == 1 {
					if bpkbNameType == m.BPKBNameType {
						data.LTV = m.LTV
						trxElaborateLTV.MappingElaborateLTVID = m.ID
					}
				} else {
					data.LTV = m.LTV
					trxElaborateLTV.MappingElaborateLTVID = m.ID
				}
			}

			//reject
			if resultPefindo == constant.DECISION_REJECT && m.TotalBakiDebetStart <= int(bakiDebet) && int(bakiDebet) <= m.TotalBakiDebetEnd && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}
		}

		// max tenor
		if m.TenorEnd >= data.MaxTenor && m.LTV > 0 {
			if m.AgeVehicle != "" {
				if bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
					data.MaxTenor = m.TenorEnd
					data.AdjustTenor = true
				}
			} else if m.BPKBNameType == 1 {
				if bpkbNameType == m.BPKBNameType {
					data.MaxTenor = m.TenorEnd
					data.AdjustTenor = true
				}
			} else {
				data.MaxTenor = m.TenorEnd
				data.AdjustTenor = true
			}
		}
	}

	err = u.repository.SaveTrxElaborateLTV(trxElaborateLTV)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error")
		return
	}

	if data.LTV == 0 && data.AdjustTenor {
		data.Reason = "Lama Angsuran Tidak Tersedia, Silahkan coba lama angsuran yang lain"
		return
	}

	if data.LTV == 0 && !data.AdjustTenor {
		data.Reason = "Mohon maaf, Lama Angsuran Tidak Tersedia"
		return
	}

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

	req, _ := json.Marshal(map[string]interface{}{
		"branch_id":  principleStepOne.BranchID,
		"asset_code": principleStepOne.AssetCode,
		"search":     principleStepOne.ManufactureYear,
	})

	var resp *resty.Response

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MARKETPRICE_URL"), req, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, reqs.ProspectID, accessToken)

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Timeout")
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error")
		return
	}

	json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &assetMP)

	if len(assetMP.Records) == 0 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Call Marketprice MDM Error")
		return
	}

	if len(assetMP.Records) > 0 {
		otr = float64(assetMP.Records[0].MarketPriceValue)
	}

	timeOut, _ = strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
	}

	// get loan amount
	payload := request.ReqMarsevLoanAmount{
		BranchID:      principleStepOne.BranchID,
		OTR:           otr,
		MaxLTV:        data.LTV,
		IsRecalculate: false,
		LoanAmount:    2000000,
		DPAmount:      0,
	}

	param, _ := json.Marshal(payload)

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_LOAN_AMOUNT_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, reqs.ProspectID, accessToken)
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

		data.LoanAmountMaximum = marsevLoanAmountRes.Data.LoanAmountMaximum
	}

	if reqs.LoanAmount > 0 {

		var resp *resty.Response

		timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_10S"))

		req, _ := json.Marshal(map[string]interface{}{
			"branch_id": principleStepOne.BranchID,
			"lob_id":    11,
			"page_size": 10,
			"search":    principleStepOne.AssetCode,
		})

		resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_ASSET_URL"), req, map[string]string{}, constant.METHOD_POST, false, 0, timeOut, reqs.ProspectID, accessToken)

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout")
			return
		}

		if resp.StatusCode() != 200 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
			return
		}

		json.Unmarshal([]byte(jsoniter.Get(resp.Body(), "data").ToString()), &assetList)

		if len(assetList.Records) == 0 {
			err = errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error")
			return
		}

		var categoryId string

		if len(assetList.Records) > 0 {
			categoryId = assetList.Records[0].CategoryID
		}

		// get marketing program
		bpkbStatusCode := "DN"
		if strings.Contains(os.Getenv("NAMA_SAMA"), principleStepOne.BPKBName) {
			bpkbStatusCode = "SN"
		}

		var customerStatus string
		if filteringKMB.CustomerStatus == nil {
			customerStatus = constant.STATUS_KONSUMEN_NEW
		} else {
			customerStatus = filteringKMB.CustomerStatus.(string)
		}

		var customerSegment string
		if filteringKMB.CustomerSegment == nil {
			customerSegment = constant.RO_AO_REGULAR
		} else {
			customerSegment = filteringKMB.CustomerSegment.(string)
		}

		customerType := utils.CapitalizeEachWord(customerStatus)
		if customerStatus != constant.STATUS_KONSUMEN_NEW {
			customerType = constant.STATUS_KONSUMEN_RO_AO + " " + utils.CapitalizeEachWord(customerSegment)
			if customerSegment == constant.RO_AO_REGULAR {
				customerType = constant.STATUS_KONSUMEN_RO_AO + " Standard"
			}
		}

		manufactureYear, _ := strconv.Atoi(principleStepOne.ManufactureYear)

		financeType := "PM"
		if reqs.FinancePurpose == constant.FINANCE_PURPOSE_MODAL_KERJA {
			financeType = "PMK"
		}

		payloadFilterProgram := request.ReqMarsevFilterProgram{
			Page:                   1,
			Limit:                  10,
			BranchID:               principleStepOne.BranchID,
			FinancingTypeCode:      financeType,
			CustomerOccupationCode: principleStepTwo.ProfessionID,
			BpkbStatusCode:         bpkbStatusCode,
			SourceApplication:      constant.MARSEV_SOURCE_APPLICATION_KPM,
			CustomerType:           customerType,
			AssetUsageTypeCode:     "C",
			AssetCategory:          categoryId,
			AssetBrand:             principleStepOne.Brand,
			AssetYear:              manufactureYear,
			LoanAmount:             reqs.LoanAmount,
			Tenor:                  reqs.Tenor,
			SalesMethodID:          5,
		}

		param, _ = json.Marshal(payloadFilterProgram)

		resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_FILTER_PROGRAM_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, reqs.ProspectID, accessToken)
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

			if len(marsevFilterProgramRes.Data) == 0 {
				err = errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Filter Program Error Not Found Data")
				return
			}
		}

		filterProgramData := marsevFilterProgramRes.Data[0]

		// calculate installment
		headerMDM := map[string]string{
			"Content-Type":  "application/json",
			"Authorization": accessToken,
		}

		licensePlateCode := utils.GetLicensePlateCode(principleStepOne.LicensePlate)
		resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MASTER_MAPPING_LICENSE_PLATE_URL")+"?lob_id="+strconv.Itoa(constant.LOBID_KMB)+"&plate_code="+licensePlateCode, nil, headerMDM, constant.METHOD_GET, false, 0, timeOut, reqs.ProspectID, accessToken)
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

		mappingLicensePlate := mdmMasterMappingLicensePlateRes.Data.Records[0]

		birthDateStr := principleStepTwo.BirthDate.Format(constant.FORMAT_DATE)
		payloadCalculate := request.ReqMarsevCalculateInstallment{
			ProgramID:              filterProgramData.ID,
			BranchID:               principleStepOne.BranchID,
			CustomerOccupationCode: principleStepTwo.ProfessionID,
			AssetUsageTypeCode:     "C",
			AssetYear:              manufactureYear,
			BpkbStatusCode:         bpkbStatusCode,
			LoanAmount:             reqs.LoanAmount,
			Otr:                    otr,
			RegionCode:             mappingLicensePlate.AreaID,
			AssetCategory:          categoryId,
			CustomerBirthDate:      birthDateStr,
			Tenor:                  reqs.Tenor,
		}

		param, _ = json.Marshal(payloadCalculate)

		resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_CALCULATE_INSTALLMENT_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, reqs.ProspectID, accessToken)
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

		if len(marsevCalculateInstallmentRes.Data) > 0 {
			data.InstallmentAmount = marsevCalculateInstallmentRes.Data[0].MonthlyInstallment
			data.AF = marsevCalculateInstallmentRes.Data[0].AmountOfFinance
			data.NTF = marsevCalculateInstallmentRes.Data[0].NTF
			data.IsPsa = marsevLoanAmountRes.Data.IsPsa
			data.AdminFee = marsevCalculateInstallmentRes.Data[0].AdminFee
			data.AssetCategoryID = categoryId
			data.Otr = otr
		}

	}

	return
}
