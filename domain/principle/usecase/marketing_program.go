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
	"sync"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) PrincipleMarketingProgram(ctx context.Context, prospectID string, accessToken string) (err error) {

	var (
		principleStepOne                entity.TrxPrincipleStepOne
		principleStepTwo                entity.TrxPrincipleStepTwo
		principleStepThree              entity.TrxPrincipleStepThree
		filteringKMB                    entity.FilteringKMB
		mappingElaborateLTV             entity.MappingElaborateLTV
		marsevLoanAmountRes             response.MarsevLoanAmountResponse
		marsevFilterProgramRes          response.MarsevFilterProgramResponse
		marsevCalculateInstallmentRes   response.MarsevCalculateInstallmentResponse
		mdmMasterMappingLicensePlateRes response.MDMMasterMappingLicensePlateResponse
		trxPrincipleMarketingProgram    entity.TrxPrincipleMarketingProgram
		wg                              sync.WaitGroup
		errChan                         = make(chan error, 5)
	)

	wg.Add(5)
	go func() {
		defer wg.Done()
		principleStepOne, err = u.repository.GetPrincipleStepOne(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		principleStepTwo, err = u.repository.GetPrincipleStepTwo(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		principleStepThree, err = u.repository.GetPrincipleStepThree(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		filteringKMB, err = u.repository.GetFilteringResult(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		mappingElaborateLTV, err = u.repository.GetElaborateLtv(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	if err := <-errChan; err != nil {
		return err
	}

	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	header := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
	}

	payload := request.ReqMarsevLoanAmount{
		BranchID:      principleStepOne.BranchID,
		OTR:           principleStepThree.OTR,
		MaxLTV:        mappingElaborateLTV.LTV,
		IsRecalculate: false,
		LoanAmount:    2000000,
		DPAmount:      principleStepThree.DPAmount,
	}

	param, _ := json.Marshal(payload)

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
	if principleStepThree.FinancePurpose == constant.FINANCE_PURPOSE_MODAL_KERJA {
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
		AssetCategory:          principleStepThree.AssetCategoryID,
		AssetBrand:             principleStepOne.Brand,
		AssetYear:              manufactureYear,
		LoanAmount:             marsevLoanAmountRes.Data.LoanAmountFinal,
		Tenor:                  principleStepThree.Tenor,
		SalesMethodID:          5,
	}

	param, _ = json.Marshal(payloadFilterProgram)

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_FILTER_PROGRAM_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
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

	headerMDM := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": accessToken,
	}

	licensePlateCode := utils.GetLicensePlateCode(principleStepOne.LicensePlate)
	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MASTER_MAPPING_LICENSE_PLATE_URL")+"?lob_id="+strconv.Itoa(constant.LOBID_KMB)+"&plate_code="+licensePlateCode, param, headerMDM, constant.METHOD_GET, false, 0, timeOut, prospectID, accessToken)
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
		LoanAmount:             marsevLoanAmountRes.Data.LoanAmountFinal,
		Otr:                    principleStepThree.OTR,
		RegionCode:             mappingLicensePlate.AreaID,
		AssetCategory:          principleStepThree.AssetCategoryID,
		CustomerBirthDate:      birthDateStr,
		Tenor:                  principleStepThree.Tenor,
	}

	param, _ = json.Marshal(payloadCalculate)

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MARSEV_CALCULATE_INSTALLMENT_URL"), param, header, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
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

	calculateInstallmentData := marsevCalculateInstallmentRes.Data[0]

	trxPrincipleMarketingProgram = entity.TrxPrincipleMarketingProgram{
		ProspectID:                 prospectID,
		ProgramID:                  filterProgramData.ID,
		ProgramName:                filterProgramData.ProgramName,
		ProductOfferingID:          filterProgramData.ProductOfferingID,
		ProductOfferingDescription: filterProgramData.ProductOfferingDescription,
		LoanAmount:                 marsevLoanAmountRes.Data.LoanAmountFinal,
		LoanAmountMaximum:          marsevLoanAmountRes.Data.LoanAmountMaximum,
		AdminFee:                   calculateInstallmentData.AdminFee,
		ProvisionFee:               calculateInstallmentData.ProvisionFee,
		DPAmount:                   calculateInstallmentData.DPAmount,
		FinanceAmount:              calculateInstallmentData.AmountOfFinance,
	}

	err = u.repository.SavePrincipleMarketingProgram(trxPrincipleMarketingProgram)

	if err != nil {
		return
	}

	return
}
