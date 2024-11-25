package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

func (u usecase) PrincipleMarketingProgram(ctx context.Context, prospectID string, accessToken string) (err error) {

	var (
		principleStepOne             entity.TrxPrincipleStepOne
		principleStepTwo             entity.TrxPrincipleStepTwo
		principleStepThree           entity.TrxPrincipleStepThree
		principleEmergencyContact    entity.TrxPrincipleEmergencyContact
		filteringKMB                 entity.FilteringKMB
		trxDetailBiro                []entity.TrxDetailBiro
		mappingElaborateLTV          entity.MappingElaborateLTV
		mdmMasterDetailBranchRes     response.MDMMasterDetailBranchResponse
		sallySubmit2wPrincipleRes    response.SallySubmit2wPrincipleResponse
		trxPrincipleMarketingProgram entity.TrxPrincipleMarketingProgram
		wg                           sync.WaitGroup
		errChan                      = make(chan error, 8)
	)

	wg.Add(8)
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
		principleEmergencyContact, err = u.repository.GetPrincipleEmergencyContact(prospectID)

		if err != nil {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		trxPrincipleMarketingProgram, err = u.repository.GetPrincipleMarketingProgram(prospectID)

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
		defer wg.Done()
		trxDetailBiro, err = u.repository.GetTrxDetailBIro(prospectID)

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

	// submit to sally
	timeOut, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
	headerMDM := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": accessToken,
	}
	resp, err := u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("MDM_MASTER_BRANCH_URL")+principleStepOne.BranchID, nil, headerMDM, constant.METHOD_GET, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - MDM Get Master Branch Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &mdmMasterDetailBranchRes); err != nil {
			return
		}
	}

	headerSally := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": accessToken,
	}

	var payloadSubmitSally request.ReqSallySubmit2wPrinciple

	customerID := strconv.Itoa(principleEmergencyContact.CustomerID)

	payloadSubmitSally.Order.Application = request.SallySubmit2wPrincipleApplication{
		BranchID:          principleStepOne.BranchID,
		BranchName:        mdmMasterDetailBranchRes.Data.BranchName,
		InstallmentAmount: principleStepThree.InstallmentAmount,
		ApplicationFormID: 1,
		OrderTypeID:       6,
		ProspectID:        prospectID,
	}

	if principleStepOne.CMOID != "" {
		payloadSubmitSally.Order.Application.CmoID = principleStepOne.CMOID
		payloadSubmitSally.Order.Application.CmoName = principleStepOne.CMOName
	}

	payloadSubmitSally.Order.Asset = request.SallySubmit2wPrincipleAsset{
		PoliceNo:              principleStepOne.LicensePlate,
		BPKBOwnershipStatusID: MapperBPKBOwnershipStatusID(principleStepOne.BPKBName),
		BPKBName:              principleStepOne.OwnerAsset,
	}

	payloadSubmitSally.Order.Customer = request.SallySubmit2wPrincipleCustomer{
		CustomerID: customerID,
	}

	var documents []request.SallySubmit2wPrincipleDocument

	if ktpPhoto, ok := principleStepTwo.KtpPhoto.(string); ok {
		documents = append(documents, request.SallySubmit2wPrincipleDocument{
			URL:  ktpPhoto,
			Type: "KTP",
		})
	}

	if selfiePhoto, ok := principleStepTwo.SelfiePhoto.(string); ok {
		documents = append(documents, request.SallySubmit2wPrincipleDocument{
			URL:  selfiePhoto,
			Type: "SELFIE",
		})
	}

	if stnkPhoto, ok := principleStepOne.STNKPhoto.(string); ok {
		documents = append(documents, request.SallySubmit2wPrincipleDocument{
			URL:  stnkPhoto,
			Type: "STNK",
		})
	}

	payloadSubmitSally.Document = documents

	isPsa := false
	if principleStepThree.Dealer == constant.DEALER_PSA {
		isPsa = true
	}

	payloadSubmitSally.Kop = request.SallySubmit2wPrincipleKop{
		IsPSA:              isPsa,
		PurposeOfFinancing: principleStepThree.FinancePurpose,
	}

	if !isPsa {
		payloadSubmitSally.Kop.FinancingObject = principleStepThree.TipeUsaha
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

	manufactureYear, _ := strconv.Atoi(principleStepOne.ManufactureYear)
	licensePlateCode := utils.GetLicensePlateCode(principleStepOne.LicensePlate)
	expiredSTNKDate := principleStepOne.STNKExpiredDate.Format(constant.FORMAT_DATE)
	expiredSTNKTaxDate := principleStepOne.TaxDate.Format(constant.FORMAT_DATE)
	cylinderVolume, _ := strconv.Atoi(principleStepOne.CC)

	payloadSubmitSally.ObjekSewa = request.SallySubmit2wPrincipleObjekSewa{
		AssetUsageID:       "C",
		CategoryID:         principleStepThree.AssetCategoryID,
		AssetCode:          principleStepOne.AssetCode,
		ManufacturingYear:  manufactureYear,
		Color:              principleStepOne.Color,
		CylinderVolume:     cylinderVolume,
		PlateAreaCode:      licensePlateCode,
		IsBBN:              false,
		ChassisNumber:      principleStepOne.NoChassis,
		MachineNumber:      principleStepOne.NoEngine,
		OTRAmount:          principleStepThree.OTR,
		ExpiredSTNKDate:    expiredSTNKDate,
		ExpiredSTNKTaxDate: expiredSTNKTaxDate,
		UpdatedBy:          customerID,
	}

	payloadSubmitSally.Biaya = request.SallySubmit2wPrincipleBiaya{
		TotalOTRAmount:        principleStepThree.OTR,
		Tenor:                 principleStepThree.Tenor,
		LoanAmount:            trxPrincipleMarketingProgram.LoanAmount,
		LoanAmountMaximum:     trxPrincipleMarketingProgram.LoanAmountMaximum,
		AdminFee:              trxPrincipleMarketingProgram.AdminFee,
		ProvisionFee:          trxPrincipleMarketingProgram.ProvisionFee,
		TotalDPAmount:         trxPrincipleMarketingProgram.DPAmount,
		AmountFinance:         trxPrincipleMarketingProgram.FinanceAmount,
		PaymentDay:            1,
		RentPaymentMethod:     "Payment Point",
		PersonalNPWPNumber:    "",
		CorrespondenceAddress: "Rumah",
		MaxLTVLOS:             mappingElaborateLTV.LTV,
		UpdatedBy:             customerID,
	}

	payloadSubmitSally.ProgramMarketing = request.SallySubmit2wPrincipleProgramMarketing{
		ProgramMarketingID:   trxPrincipleMarketingProgram.ProgramID,
		ProgramMarketingName: trxPrincipleMarketingProgram.ProgramName,
		ProductOfferingID:    trxPrincipleMarketingProgram.ProductOfferingID,
		ProductOfferingName:  trxPrincipleMarketingProgram.ProductOfferingDescription,
		UpdatedBy:            customerID,
	}

	isBlacklist := filteringKMB.IsBlacklist == 1
	nextProcess := filteringKMB.NextProcess == 1

	var PBKReportCustomer string
	var PBKReportSpouse string
	for _, v := range trxDetailBiro {
		if v.Score != "" && v.Score != constant.DECISION_PBK_NO_HIT && v.Score != constant.PEFINDO_UNSCORE {
			if v.Subject == constant.CUSTOMER {
				PBKReportCustomer = v.UrlPdfReport
			}
			if v.Subject == constant.SPOUSE {
				PBKReportSpouse = v.UrlPdfReport
			}
		}
	}

	var bakiDebet float64
	if filteringKMB.TotalBakiDebetNonCollateralBiro != nil {
		bakiDebet, err = utils.GetFloat(filteringKMB.TotalBakiDebetNonCollateralBiro)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " baki debet " + err.Error())
			return
		}
	}

	var customerStatusKMB string
	if filteringKMB.CustomerStatusKMB != nil {
		customerStatusKMB = filteringKMB.CustomerStatusKMB.(string)
	}

	payloadSubmitSally.Filtering = request.SallySubmit2wPrincipleFiltering{
		Decision:          filteringKMB.Decision,
		Reason:            filteringKMB.Reason.(string),
		CustomerStatus:    customerStatus,
		CustomerStatusKMB: customerStatusKMB,
		CustomerSegment:   customerSegment,
		IsBlacklist:       isBlacklist,
		NextProcess:       nextProcess,
		PBKReportCustomer: PBKReportCustomer,
		PBKReportSpouse:   PBKReportSpouse,
		BakiDebet:         bakiDebet,
	}

	param, _ := json.Marshal(payloadSubmitSally)

	resp, err = u.httpclient.EngineAPI(ctx, constant.DILEN_KMB_LOG, os.Getenv("SALLY_SUBMISSION_2W_PRINCIPLE_URL"), param, headerSally, constant.METHOD_POST, false, 0, timeOut, prospectID, accessToken)
	if err != nil {
		return
	}

	if resp.StatusCode() != 200 {
		err = errors.New(constant.ERROR_UPSTREAM + " - Sally Submit 2W Principle Error")
		return
	}

	if resp.StatusCode() == 200 {
		if err = json.Unmarshal([]byte(jsoniter.Get(resp.Body()).ToString()), &sallySubmit2wPrincipleRes); err != nil {
			return
		}

		statusCode := constant.PRINCIPLE_STATUS_SUBMIT_SALLY
		u.producer.PublishEvent(ctx, middlewares.UserInfoData.AccessToken, constant.TOPIC_SUBMISSION_PRINCIPLE, constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE, prospectID, utils.StructToMap(request.Update2wPrincipleTransaction{
			OrderID:       prospectID,
			KpmID:         principleStepOne.KPMID,
			Source:        3,
			StatusCode:    statusCode,
			ProductName:   principleStepOne.AssetCode,
			BranchCode:    principleStepOne.BranchID,
			AssetTypeCode: constant.KPM_ASSET_TYPE_CODE_MOTOR,
		}), 0)
	}

	return
}

func MapperBPKBOwnershipStatusID(bpkbName string) int {
	switch bpkbName {
	case "K":
		return 1
	case "P":
		return 2
	case "KK":
		return 3
	case "O":
		return 4
	default:
		return 4
	}
}
