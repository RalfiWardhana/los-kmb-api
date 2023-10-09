package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"los-kmb-api/domain/kmb/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

var (
	DtmRequest = time.Now()
)

type repoHandler struct {
	losDB     *gorm.DB
	logsDB    *gorm.DB
	confinsDB *gorm.DB
	stagingDB *gorm.DB
	wgOffDB   *gorm.DB
	kmbOffDB  *gorm.DB
	newKmbDB  *gorm.DB
}

func NewRepository(los, logs, confins, staging, wgOffDB, kmbOff, newKmbDB *gorm.DB) interfaces.Repository {
	return &repoHandler{
		losDB:     los,
		logsDB:    logs,
		confinsDB: confins,
		stagingDB: staging,
		wgOffDB:   wgOffDB,
		kmbOffDB:  kmbOff,
		newKmbDB:  newKmbDB,
	}
}

func (r repoHandler) ScanTrxMaster(prospectID string) (countMaster int, err error) {

	var (
		master []entity.TrxMaster
	)

	if err = r.newKmbDB.Raw(fmt.Sprintf(`
		SELECT tm.ProspectID FROM trx_master tm WITH (nolock) 
		LEFT JOIN trx_status ts ON tm.ProspectID = ts.ProspectID 
		WHERE tm.ProspectID = '%s' AND ((ts.activity = '%s' AND ts.source_decision = '%s') OR (ts.activity != '%s' OR ts.source_decision != '%s'))`,
		prospectID, constant.ACTIVITY_UNPROCESS, constant.PRESCREENING, constant.ACTIVITY_UNPROCESS, constant.SOURCE_DECISION_DUPCHECK)).Scan(&master).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	countMaster = len(master)

	return
}

func (r repoHandler) ScanTrxPrescreening(prospectID string) (count int, err error) {

	var (
		trxPrescreening []entity.TrxPrescreening
	)

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT ProspectID FROM trx_prescreening WITH (nolock) WHERE ProspectID = '%s'", prospectID)).Scan(&trxPrescreening).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	count = len(trxPrescreening)

	return
}

func (r repoHandler) GetFilteringResult(prospectID string) (filtering entity.FilteringKMB, err error) {
	var x sql.TxOptions

	resultValid := os.Getenv("BIRO_VALID_DAYS")

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmbDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT * FROM trx_filtering WITH (nolock) WHERE prospect_id = '%s' AND DATEADD(day, -%s, CAST(GETDATE() AS date)) <= CAST(created_at AS date) ORDER BY created_at DESC", prospectID, resultValid)).Scan(&filtering).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.RECORD_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetFilteringForJourney(prospectID string) (filtering entity.FilteringKMB, err error) {
	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_10S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.newKmbDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = r.newKmbDB.Raw(fmt.Sprintf("SELECT * FROM trx_filtering WITH (nolock) WHERE prospect_id = '%s'", prospectID)).Scan(&filtering).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveTransaction(countTrx int, data request.Metrics, trxPrescreening entity.TrxPrescreening, trxFMF response.TrxFMF, details []entity.TrxDetail, reason string) (newErr error) {

	location, _ := time.LoadLocation("Asia/Jakarta")

	formatBirthDate, _ := time.ParseInLocation("2006-01-02", data.CustomerPersonal.BirthDate, location)

	var formatIDExpired interface{}

	if data.CustomerPersonal.ExpiredDate != nil {
		formatIDExpired, _ = time.ParseInLocation("2006-01-02", *data.CustomerPersonal.ExpiredDate, location)
	}

	formatTaxDate, _ := time.ParseInLocation("2006-01-02", data.Item.TaxDate, location)
	formatSTNKExpired, _ := time.ParseInLocation("2006-01-02", data.Item.STNKExpiredDate, location)

	var logInfo interface{}

	newErr = r.newKmbDB.Transaction(func(tx *gorm.DB) error {

		//save data from payload
		if countTrx == 0 {
			var legalAddress, residenceAddress, companyAddress, emergencyAddress, mailingAddress, ownerAddress, locationAddress string
			for _, v := range data.Address {
				if v.Type == "LEGAL" {
					legalAddress = v.Address
				} else if v.Type == "RESIDENCE" {
					residenceAddress = v.Address
				} else if v.Type == "COMPANY" {
					companyAddress = v.Address
				} else if v.Type == "EMERGENCY" {
					emergencyAddress = v.Address
				} else if v.Type == "LOCATION" {
					locationAddress = v.Address
				} else if v.Type == "MAILING" {
					mailingAddress = v.Address
				} else {
					ownerAddress = v.Address
				}
			}

			var encrypted entity.Encrypted

			if err := tx.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName, SCP.dbo.ENC_B64('SEC','%s') AS SurgateMotherName, SCP.dbo.ENC_B64('SEC','%s') AS Email,
			SCP.dbo.ENC_B64('SEC','%s') AS MobilePhone, SCP.dbo.ENC_B64('SEC','%s') AS BirthPlace, SCP.dbo.ENC_B64('SEC','%s') AS ResidenceAddress, SCP.dbo.ENC_B64('SEC','%s') AS LegalAddress,
			SCP.dbo.ENC_B64('SEC','%s') AS IDNumber,SCP.dbo.ENC_B64('SEC','%s') AS FullName, SCP.dbo.ENC_B64('SEC','%s') AS CompanyAddress, SCP.dbo.ENC_B64('SEC','%s') AS EmergencyAddress,
			SCP.dbo.ENC_B64('SEC','%s') AS LocationAddress, SCP.dbo.ENC_B64('SEC','%s') AS MailingAddress, SCP.dbo.ENC_B64('SEC','%s') AS OwnerAddress`,
				data.CustomerPersonal.LegalName, data.CustomerPersonal.SurgateMotherName, data.CustomerPersonal.Email, data.CustomerPersonal.MobilePhone, data.CustomerPersonal.BirthPlace,
				residenceAddress, legalAddress, data.CustomerPersonal.IDNumber, data.CustomerPersonal.FullName, companyAddress, emergencyAddress,
				locationAddress, mailingAddress, ownerAddress)).Scan(&encrypted).Error; err != nil {
				return err
			}

			master := entity.TrxMaster{
				ProspectID:        data.Transaction.ProspectID,
				BranchID:          data.Transaction.BranchID,
				Channel:           data.Transaction.Channel,
				Lob:               data.Transaction.Lob,
				IncomingSource:    data.Transaction.IncomingSource,
				ApplicationSource: data.Transaction.ApplicationSource,
				OrderAt:           data.Transaction.OrderAt,
			}

			logInfo = master

			if err := tx.Create(&master).Error; err != nil {
				return err
			}

			if trxPrescreening.Decision != "" {
				logInfo = trxPrescreening

				if err := tx.Create(&trxPrescreening).Error; err != nil {
					return err
				}
			}

			for idx, v := range data.Address {
				if v.Type == "LEGAL" {
					data.Address[idx].Address = encrypted.LegalAddress
				} else if v.Type == "RESIDENCE" {
					data.Address[idx].Address = encrypted.ResidenceAddress
				} else if v.Type == "COMPANY" {
					data.Address[idx].Address = encrypted.CompanyAddress
				} else if v.Type == "EMERGENCY" {
					data.Address[idx].Address = encrypted.EmergencyAddress
				} else if v.Type == "MAILING" {
					data.Address[idx].Address = encrypted.MailingAddress
				} else if v.Type == "LOCATION" {
					data.Address[idx].Address = encrypted.LocationAddress
				} else {
					data.Address[idx].Address = encrypted.OwnerAddress
				}
			}

			for i := 0; i < len(data.Address); i++ {

				address := entity.CustomerAddress{
					ProspectID: data.Transaction.ProspectID,
					Type:       data.Address[i].Type,
					Address:    data.Address[i].Address,
					Rt:         data.Address[i].Rt,
					Rw:         data.Address[i].Rw,
					Kelurahan:  data.Address[i].Kelurahan,
					Kecamatan:  data.Address[i].Kecamatan,
					City:       data.Address[i].City,
					ZipCode:    data.Address[i].ZipCode,
					AreaPhone:  data.Address[i].AreaPhone,
					Phone:      data.Address[i].Phone,
				}

				logInfo = address

				if err := tx.Create(&address).Error; err != nil {
					return err
				}
			}

			var surveyResult string

			for i := 0; i < len(data.CustomerPhoto); i++ {
				if data.CustomerPhoto[i].ID == "RESULT_SURVEY" {
					surveyResult = data.CustomerPhoto[i].Url
					continue
				}

				photo := entity.CustomerPhoto{
					ProspectID: data.Transaction.ProspectID,
					PhotoID:    data.CustomerPhoto[i].ID,
					Url:        data.CustomerPhoto[i].Url,
				}

				logInfo = photo

				if err := tx.Create(&photo).Error; err != nil {
					return err
				}
			}

			var living_cost_amount float64
			if data.CustomerPersonal.LivingCostAmount != nil {
				living_cost_amount = *data.CustomerPersonal.LivingCostAmount
			}

			personal := entity.CustomerPersonal{
				ProspectID:                 data.Transaction.ProspectID,
				IDType:                     data.CustomerPersonal.IDType,
				IDNumber:                   encrypted.IDNumber,
				FullName:                   encrypted.FullName,
				LegalName:                  encrypted.LegalName,
				BirthPlace:                 encrypted.BirthPlace,
				Education:                  data.CustomerPersonal.Education,
				BirthDate:                  formatBirthDate,
				MaritalStatus:              data.CustomerPersonal.MaritalStatus,
				SurgateMotherName:          encrypted.SurgateMotherName,
				Religion:                   data.CustomerPersonal.Religion,
				Gender:                     data.CustomerPersonal.Gender,
				PersonalNPWP:               data.CustomerPersonal.NPWP,
				MobilePhone:                encrypted.MobilePhone,
				Email:                      encrypted.Email,
				HomeStatus:                 data.CustomerPersonal.HomeStatus,
				StaySinceYear:              data.CustomerPersonal.StaySinceYear,
				StaySinceMonth:             data.CustomerPersonal.StaySinceMonth,
				ExpiredDate:                formatIDExpired,
				LivingCostAmount:           living_cost_amount,
				ExtCompanyPhone:            data.CustomerEmployment.ExtCompanyPhone,
				SourceOtherIncome:          data.CustomerEmployment.SourceOtherIncome,
				EmergencyOfficeAreaPhone:   data.CustomerEmcon.AreaPhone,
				EmergencyOfficePhone:       data.CustomerEmcon.Phone,
				PersonalCustomerType:       data.CustomerPersonal.PersonalCustomerType,
				Nationality:                data.CustomerPersonal.Nationality,
				WNACountry:                 data.CustomerPersonal.WNACountry,
				HomeLocation:               data.CustomerPersonal.HomeLocation,
				CustomerGroup:              data.CustomerPersonal.CustomerGroup,
				KKNo:                       data.CustomerPersonal.KKNo,
				BankID:                     data.CustomerPersonal.BankID,
				AccountNo:                  data.CustomerPersonal.AccountNo,
				AccountName:                data.CustomerPersonal.AccountName,
				Counterpart:                data.CustomerPersonal.Counterpart,
				DebtBusinessScale:          data.CustomerPersonal.DebtBusinessScale,
				DebtGroup:                  data.CustomerPersonal.DebtGroup,
				IsAffiliateWithPP:          data.CustomerPersonal.IsAffiliateWithPP,
				AgreetoAcceptOtherOffering: data.CustomerPersonal.AgreetoAcceptOtherOffering,
				DataType:                   data.CustomerPersonal.DataType,
				Status:                     data.CustomerPersonal.Status,
				SurveyResult:               surveyResult,
				RentFinishDate:             data.CustomerPersonal.RentFinishDate,
			}

			if trxFMF.DupcheckData.CustomerID != nil {
				personal.CustomerID = trxFMF.DupcheckData.CustomerID.(string)
			}

			if trxFMF.DupcheckData.StatusKonsumen != "" {
				personal.CustomerStatus = trxFMF.DupcheckData.StatusKonsumen
			}

			logInfo = personal

			if err := tx.Create(&personal).Error; err != nil {
				return err
			}

			employment := entity.CustomerEmployment{
				ProspectID:            data.Transaction.ProspectID,
				ProfessionID:          data.CustomerEmployment.ProfessionID,
				JobType:               data.CustomerEmployment.JobType,
				JobPosition:           data.CustomerEmployment.JobPosition,
				CompanyName:           data.CustomerEmployment.CompanyName,
				IndustryTypeID:        data.CustomerEmployment.IndustryTypeID,
				EmploymentSinceYear:   data.CustomerEmployment.EmploymentSinceYear,
				EmploymentSinceMonth:  data.CustomerEmployment.EmploymentSinceMonth,
				MonthlyFixedIncome:    data.CustomerEmployment.MonthlyFixedIncome,
				MonthlyVariableIncome: *data.CustomerEmployment.MonthlyVariableIncome,
				SpouseIncome:          *data.CustomerEmployment.SpouseIncome,
			}

			logInfo = employment

			if err := tx.Create(&employment).Error; err != nil {
				return err
			}

			var (
				adminFee float64
			)
			if data.Apk.AdminFee != nil {
				adminFee = *data.Apk.AdminFee
			}

			var (
				percentDP      float64
				fidusiaFee     float64
				interestRate   float64
				interestAmount float64
				provisionFee   float64
			)
			if data.Apk.PercentDP != nil {
				percentDP = *data.Apk.PercentDP
			}
			if data.Apk.FidusiaFee != nil {
				fidusiaFee = *data.Apk.FidusiaFee
			}
			if data.Apk.InterestRate != nil {
				interestRate = *data.Apk.InterestRate
			}
			if data.Apk.InterestAmount != nil {
				interestAmount = *data.Apk.InterestAmount
			}
			if data.Apk.ProvisionFee != nil {
				provisionFee = *data.Apk.ProvisionFee
			}

			var surveyFee float64 = 0
			if data.Apk.SurveyFee != nil {
				surveyFee = *data.Apk.SurveyFee
			}

			apk := entity.TrxApk{
				ProspectID:                  data.Transaction.ProspectID,
				Tenor:                       &data.Apk.Tenor,
				ProductOfferingID:           data.Apk.ProductOfferingID,
				ProductID:                   data.Apk.ProductID,
				NTF:                         data.Apk.NTF,
				AF:                          data.Apk.AF,
				AoID:                        data.Apk.AoID,
				DPAmount:                    data.Apk.DPAmount,
				AdminFee:                    adminFee,
				OTR:                         data.Apk.OTR,
				InstallmentAmount:           data.Apk.InstallmentAmount,
				FirstInstallment:            data.Apk.FirstInstallment,
				OtherFee:                    data.Apk.OtherFee,
				PercentDP:                   percentDP,
				AssetInsuranceFee:           data.Apk.PremiumAmountToCustomer,
				LifeInsuranceFee:            data.Item.PremiumAmountToCustomer,
				FidusiaFee:                  fidusiaFee,
				InterestRate:                interestRate,
				InsuranceAmount:             data.Apk.InsuranceAmount,
				InterestAmount:              interestAmount,
				PaymentMethod:               data.Apk.PaymentMethod,
				SurveyFee:                   surveyFee,
				IsFidusiaCovered:            data.Apk.IsFidusiaCovered,
				ProvisionFee:                provisionFee,
				InsAssetPaidBy:              data.Apk.InsAssetPaidBy,
				InsAssetPeriod:              data.Apk.InsAssetPeriod,
				EffectiveRate:               data.Apk.EffectiveRate,
				SalesmanID:                  data.Apk.SalesmanID,
				SupplierBankAccountID:       data.Apk.SupplierBankAccountID,
				LifeInsuranceCoyBranchID:    data.Apk.LifeInsuranceCoyBranchID,
				LifeInsuranceAmountCoverage: data.Apk.LifeInsuranceAmountCoverage,
				CommisionSubsidi:            data.Apk.CommisionSubsidy,
				ProductOfferingDesc:         data.Apk.ProductOfferingDesc,
				Dealer:                      data.Apk.Dealer,
				LoanAmount:                  data.Apk.LoanAmount,
				FinancePurpose:              data.Apk.FinancePurpose,
				NTFAkumulasi:                trxFMF.NTFAkumulasi,
				NTFOtherAmount:              trxFMF.NTFOtherAmount,
				NTFOtherAmountSpouse:        trxFMF.NTFOtherAmountSpouse,
				NTFOtherAmountDetail:        trxFMF.NTFOtherAmountDetail,
				NTFConfinsAmount:            trxFMF.NTFConfinsAmount,
				NTFConfins:                  trxFMF.NTFConfins,
				NTFTopup:                    trxFMF.NTFTopup,
				WayOfPayment:                data.Apk.WayOfPayment,
			}

			logInfo = apk

			if err := tx.Create(&apk).Error; err != nil {
				return err
			}

			if data.CustomerSpouse != nil {
				spouseBirthdate, _ := time.ParseInLocation("2006-01-02", data.CustomerSpouse.BirthDate, location)

				customerSpouse := entity.CustomerSpouse{
					ProspectID:        data.Transaction.ProspectID,
					IDNumber:          data.CustomerSpouse.IDNumber,
					FullName:          data.CustomerSpouse.FullName,
					LegalName:         data.CustomerSpouse.LegalName,
					BirthDate:         spouseBirthdate,
					BirthPlace:        data.CustomerSpouse.BirthPlace,
					SurgateMotherName: data.CustomerSpouse.SurgateMotherName,
					Gender:            data.CustomerSpouse.Gender,
					CompanyPhone:      data.CustomerSpouse.CompanyPhone,
					CompanyName:       data.CustomerSpouse.CompanyName,
					MobilePhone:       data.CustomerSpouse.MobilePhone,
					ProfessionID:      data.CustomerSpouse.ProfessionID,
				}

				logInfo = customerSpouse

				if err := tx.Create(&customerSpouse).Error; err != nil {
					return err
				}

			}

			emcon := entity.CustomerEmcon{
				ProspectID:           data.Transaction.ProspectID,
				Name:                 data.CustomerEmcon.Name,
				Relationship:         data.CustomerEmcon.Relationship,
				MobilePhone:          data.CustomerEmcon.MobilePhone,
				EmconVerified:        data.CustomerEmcon.ApplicationEmconSesuai,
				VerifyBy:             data.CustomerEmcon.VerifyBy,
				VerificationWith:     data.CustomerEmcon.VerificationWith,
				KnownCustomerAddress: data.CustomerEmcon.KnownCustomerAddress,
				KnownCustomerJob:     data.CustomerEmcon.KnownCustomerJob,
			}

			logInfo = emcon

			if err := tx.Create(&emcon).Error; err != nil {
				return err
			}

			item := entity.TrxItem{
				ProspectID:                   data.Transaction.ProspectID,
				CategoryID:                   data.Item.CategoryID,
				SupplierID:                   data.Item.SupplierID,
				Qty:                          data.Item.Qty,
				AssetCode:                    data.Item.AssetCode,
				AssetDescription:             data.Item.AssetDescription,
				ManufactureYear:              data.Item.ManufactureYear,
				BPKBName:                     data.Item.BPKBName,
				OwnerAsset:                   data.Item.OwnerAsset,
				LicensePlate:                 data.Item.LicensePlate,
				Color:                        data.Item.Color,
				EngineNo:                     data.Item.NoEngine,
				ChassisNo:                    data.Item.NoChassis,
				Pos:                          data.Item.POS,
				Cc:                           data.Item.CC,
				Condition:                    data.Item.Condition,
				Region:                       data.Item.Region,
				TaxDate:                      formatTaxDate,
				STNKExpiredDate:              formatSTNKExpired,
				AssetInsuranceAmountCoverage: data.Item.AssetInsuranceAmountCoverage,
				InsAssetInsuredBy:            data.Item.InsAssetInsuredBy,
				InsuranceCoyBranchID:         data.Item.InsuranceCoyBranchID,
				CoverageType:                 data.Item.CoverageType,
				OwnerKTP:                     data.Item.OwnerKTP,
				AssetUsage:                   data.Item.AssetUsage,
				Brand:                        data.Item.Brand,
			}

			logInfo = item

			if err := tx.Create(&item).Error; err != nil {
				return err
			}

			agent := entity.TrxInfoAgent{
				ProspectID: data.Transaction.ProspectID,
				NIK:        data.Agent.CmoNik,
				Name:       data.Agent.CmoName,
				Info:       data.Agent.CmoRecom,
				RecomDate:  data.Agent.RecomDate,
			}

			logInfo = agent

			if err := tx.Create(&agent).Error; err != nil {
				return err
			}

			for i := 0; i < len(data.Surveyor); i++ {

				requestDate, _ := time.Parse("2006-01-02", data.Surveyor[i].RequestDate)
				assignDate, _ := time.Parse("2006-01-02", data.Surveyor[i].AssignDate)
				resultDate, _ := time.Parse("2006-01-02", data.Surveyor[i].ResultDate)

				surveyor := entity.TrxSurveyor{
					ProspectID:   data.Transaction.ProspectID,
					Destination:  data.Surveyor[i].Destination,
					RequestDate:  requestDate,
					AssignDate:   assignDate,
					SurveyorName: data.Surveyor[i].SurveyorName,
					ResultDate:   resultDate,
					Status:       data.Surveyor[i].Status,
				}

				logInfo = surveyor

				if err := tx.Create(&surveyor).Error; err != nil {
					return err
				}

			}

			if data.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_WRST || data.CustomerEmployment.ProfessionID == constant.PROFESSION_ID_PRO {
				if data.CustomerOmset != nil {
					omset := *data.CustomerOmset
					for i := 0; i < len(omset); i++ {

						cusOmset := entity.CustomerOmset{
							ProspectID:        data.Transaction.ProspectID,
							SeqNo:             i + 1,
							MonthlyOmsetYear:  omset[i].MonthlyOmsetYear,
							MonthlyOmsetMonth: omset[i].MonthlyOmsetMonth,
							MonthlyOmset:      omset[i].MonthlyOmset,
						}

						logInfo = cusOmset

						if err := tx.Create(&cusOmset).Error; err != nil {
							return err
						}
					}
				}
			}
		} else {
			//update data from metrics
			var updateMap = make(map[string]interface{})
			if trxFMF.DupcheckData.CustomerID != nil {
				updateMap["CustomerID"] = trxFMF.DupcheckData.CustomerID.(string)
			}
			if trxFMF.DupcheckData.StatusKonsumen != "" {
				updateMap["CustomerStatus"] = trxFMF.DupcheckData.StatusKonsumen
			}
			if updateMap != nil {
				if err := tx.Model(&entity.CustomerPersonal{}).Where("ProspectID = ?", data.Transaction.ProspectID).Updates(updateMap).Error; err != nil {
					return err
				}
			}
		}

		//save data form metrics

		for i := 0; i < len(details); i++ {
			// skip prescreening unpr
			if details[i].SourceDecision != constant.PRESCREENING || details[i].Activity != constant.ACTIVITY_UNPROCESS {
				detail := entity.TrxDetail{
					ProspectID:     details[i].ProspectID,
					StatusProcess:  details[i].StatusProcess,
					Activity:       details[i].Activity,
					Decision:       details[i].Decision,
					RuleCode:       details[i].RuleCode,
					SourceDecision: details[i].SourceDecision,
					NextStep:       details[i].NextStep,
					Info:           details[i].Info,
					CreatedBy:      constant.SYSTEM_CREATED,
				}

				logInfo = detail

				if err := tx.Create(&detail).Error; err != nil {
					return err
				}
			}
		}

		last := details[len(details)-1]

		status := entity.TrxStatus{
			ProspectID:     last.ProspectID,
			StatusProcess:  last.StatusProcess,
			Activity:       last.Activity,
			Decision:       last.Decision,
			RuleCode:       last.RuleCode,
			SourceDecision: last.SourceDecision,
			NextStep:       last.NextStep,
			Reason:         reason,
		}

		logInfo = status

		if rowsAffected := tx.Model(&status).Where("ProspectID = ?", status.ProspectID).Updates(status).RowsAffected; rowsAffected == 0 {
			if err := tx.Create(&status).Error; err != nil {
				return err
			}
		}

		return nil

	})

	if newErr != nil {
		newErr = errors.New(fmt.Sprintf("%s - %s - %s", constant.ERROR_UPSTREAM, newErr.Error(), logInfo))
	}

	return
}

func (r repoHandler) SaveTrxJourney(prospectID string, request interface{}) (err error) {

	requestByte, _ := json.Marshal(request)

	if err = r.newKmbDB.Model(&entity.TrxJourney{}).Create(&entity.TrxJourney{
		ProspectID: prospectID,
		Request:    string(utils.SafeEncoding(requestByte)),
	}).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetTrxJourney(prospectID string) (trxJourney entity.TrxJourney, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT ProspectID, request from trx_journey with (nolock) where ProspectID = '%s'", prospectID)).Scan(&trxJourney).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveLogOrchestrator(header, request, response interface{}, path, method, prospectID string, requestID string) (err error) {

	headerByte, _ := json.Marshal(header)
	requestByte, _ := json.Marshal(request)
	responseByte, _ := json.Marshal(response)

	if err = r.logsDB.Model(&entity.LogOrchestrator{}).Create(&entity.LogOrchestrator{
		ID:           requestID,
		ProspectID:   prospectID,
		Owner:        "LOS-KMB",
		Header:       string(headerByte),
		Url:          path,
		Method:       method,
		RequestData:  string(utils.SafeEncoding(requestByte)),
		ResponseData: string(utils.SafeEncoding(responseByte)),
	}).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetLogOrchestrator(prospectID string) (logOrchestrator entity.LogOrchestrator, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT TOP 1 ProspectID, request_data from log_orchestrators with (nolock) where ProspectID = '%s' AND url = '/api/v3/kmb/consume/journey' ORDER BY created_at DESC", prospectID)).Scan(&logOrchestrator).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetDupcheckConfig() (config entity.AppConfig, err error) {

	if err = r.losDB.Raw("SELECT [key], [value] FROM app_config WHERE lob = 'KMB-OFF' AND [key] = 'dupcheck_kmb_config' AND group_name = 'dupcheck'").Scan(&config).Error; err != nil {
		return
	}

	if config == (entity.AppConfig{}) {
		err = errors.New(constant.ERROR_NOT_FOUND)
		return
	}
	return
}

func (r repoHandler) GetNewDupcheck(prospectID string) (data entity.NewDupcheck, err error) {

	if err = r.losDB.Raw(fmt.Sprintf("SELECT TOP 1 ProspectID, customer_status, customer_type FROM new_dupcheck WITH (nolock) WHERE ProspectID = '%s' ORDER BY created_at DESC", prospectID)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SaveNewDupcheck(newDupcheck entity.NewDupcheck) (err error) {

	if err = r.losDB.Create(&newDupcheck).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetDummyCustomerDomain(idNumber string) (data entity.DummyCustomerDomain, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT * FROM dummy_customer_domain WITH (nolock) WHERE id_number = '%s'", idNumber)).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetDummyLatestPaidInstallment(idNumber string) (data entity.DummyLatestPaidInstallment, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT * FROM dummy_latest_paid_installment WITH (nolock) WHERE id_number = '%s'", idNumber)).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r repoHandler) GetEncryptedValue(idNumber string, legalName string, motherName string) (encrypted entity.Encrypted, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS LegalName, SCP.dbo.ENC_B64('SEC','%s') AS SurgateMotherName, SCP.dbo.ENC_B64('SEC','%s') AS IDNumber`,
		legalName, motherName, idNumber)).Scan(&encrypted).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetDSRBypass() (config entity.AppConfig, err error) {

	if err = r.losDB.Raw("SELECT [key], [value] FROM app_config WHERE lob = 'KMOB-OFF' AND [key] = 'dsr-bypass' AND group_name = 'dsr_setting'").Scan(&config).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) ScanWgOff(query string) (data entity.ScanInstallmentAmount, err error) {
	if err = r.wgOffDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanKmbOff(query string) (data entity.ScanInstallmentAmount, err error) {
	if err = r.kmbOffDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanKmobOff(query string) (data entity.ScanInstallmentAmount, err error) {

	if err = r.losDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) ScanWgOnl(query string) (data entity.ScanInstallmentAmount, err error) {

	if err = r.losDB.Raw(query).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}

		return
	}

	return
}

func (r repoHandler) GetKMBOff() (config entity.AppConfig, err error) {

	if err = r.losDB.Raw("SELECT [key], [value] FROM app_config WHERE lob = 'KMB-OFF' AND [key] = 'pmk_kmb_off' AND group_name = 'pmk_config'").Scan(&config).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetMinimalIncomePMK(branchID string, statusKonsumen string) (responseIncomePMK entity.MappingIncomePMK, err error) {
	if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM mapping_income_pmk WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, branchID, statusKonsumen)).Scan(&responseIncomePMK).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			if err = r.losDB.Raw(fmt.Sprintf(`SELECT * FROM mapping_income_pmk WHERE lob='los_kmb_off' AND branch_id='%s' AND status_konsumen='%s'`, constant.DEFAULT_BRANCH_ID, statusKonsumen)).Scan(&responseIncomePMK).Error; err != nil {
				return
			}
		}
	}
	return
}

func (r repoHandler) GetLatestBannedRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error) {

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT * FROM dupcheck_rejection_nokanosin WHERE NoRangka = '%s' AND IsBanned = 1 ORDER BY created_at DESC LIMIT 1", noRangka)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetLatestRejectionNoka(noRangka string) (data entity.DupcheckRejectionNokaNosin, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT * FROM dupcheck_rejection_nokanosin WHERE NoRangka = '%s' AND CAST(created_at as DATE) = '%s' ORDER BY created_at DESC LIMIT 1", noRangka, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetAllReject(idNumber string) (data []entity.DupcheckRejectionPMK, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT di.BranchID, di.ProspectID, di.IDNumber, di.DtmUpd, dr.reject_pmk, dr.reject_dsr, fi.final_approval, fi.dtm_final_approval FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) = '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) ORDER BY di.DtmUpd ASC", idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetHistoryRejectAttempt(idNumber string) (data []entity.DupcheckRejectionPMK, err error) {

	searchRange := time.Now().AddDate(0, 0, -30)
	searchRangeString := searchRange.Format("2006-01-02")

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT x.* FROM (SELECT COUNT(*) AS reject_attempt, DATE(fi.dtm_final_approval) as date FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND (dr.reject_pmk IS NOT NULL OR dr.reject_dsr IS NOT NULL) AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) >= '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) GROUP BY date) x ORDER BY x.date DESC LIMIT 1", idNumber, searchRangeString)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) GetCheckingRejectAttempt(idNumber, blackListDate string) (data entity.DupcheckRejectionPMK, err error) {

	if err = r.kmbOffDB.Raw(fmt.Sprintf("SELECT x.* FROM (SELECT COUNT(*) AS reject_attempt, DATE(fi.dtm_final_approval) as date FROM data_inquiry di LEFT JOIN final_inquiry fi ON di.ProspectID = fi.ProspectID LEFT JOIN dupcheck_rejection_pmk dr ON di.ProspectID = dr.ProspectID LEFT JOIN dupcheck_inquiry dup ON di.ProspectID = dup.ProspectID WHERE fi.final_approval = 0 AND di.IDNumber = '%s' AND CAST(di.DtmUpd as DATE) = '%s' AND (dup.code NOT IN ('653','655','667') OR dup.code IS NULL) GROUP BY date) x ORDER BY x.date DESC", idNumber, blackListDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}

func (r repoHandler) SaveDataNoka(data entity.DupcheckRejectionNokaNosin) (err error) {
	data.CreatedAt = time.Now()
	if err = r.kmbOffDB.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) SaveDataApiLog(data entity.TrxApiLog) (err error) {
	data.Timestamps = time.Now()
	if err = r.kmbOffDB.Create(data).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetInstallmentAmountChassisNumber(chassisNumber string) (data entity.SpDupcekChasisNo, err error) {

	var x sql.TxOptions

	timeout, _ := strconv.Atoi(config.Env("DEFAULT_TIMEOUT_30S"))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	db := r.confinsDB.BeginTx(ctx, &x)
	defer db.Commit()

	if err = db.Raw("exec [%s] '%s'", os.Getenv("SP_DUPCHECK_CHASSIS_NUMBER"), chassisNumber).Scan(&data).Error; err != nil {
		return
	}

	return

}

func (r repoHandler) GetDummyAgreementChassisNumber(idNumber string) (data entity.DummyAgreementChassisNumber, err error) {

	if err = r.logsDB.Raw(fmt.Sprintf("SELECT * FROM dummy_agreement_chassis_number WITH (nolock) WHERE id_number = '%s'", idNumber)).Scan(&data).Error; err != nil {
		return
	}
	return
}

func (r *repoHandler) GetConfig(groupName string, lob string, key string) (appConfig entity.AppConfig) {
	if lob == "" || key == "" {
		if err := r.losDB.
			Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s'", groupName)).
			Scan(&appConfig).Error; err != nil {
			return appConfig
		}

		return
	}
	if err := r.losDB.
		Raw(fmt.Sprintf("SELECT [value] FROM app_config WITH (nolock) WHERE group_name = '%s' AND lob = '%s' AND [key]= '%s' AND is_active = 1", groupName, lob, key)).
		Scan(&appConfig).Error; err != nil {
		return appConfig
	}

	return appConfig
}

func (r *repoHandler) SaveVerificationFaceCompare(data entity.VerificationFaceCompare) error {
	if err := r.losDB.Create(&data).Error; err != nil {
		return fmt.Errorf("save verification face compare error: %w", err)
	}
	return nil
}

func (r repoHandler) GetEncB64(myString string) (encryptedString entity.EncryptedString, err error) {

	if err = r.losDB.Raw(fmt.Sprintf(`SELECT SCP.dbo.ENC_B64('SEC','%s') AS my_string`, myString)).Scan(&encryptedString).Error; err != nil {
		return
	}

	return
}

func (r repoHandler) GetCurrentTrxWithRejectDSR(idNumber string) (data entity.TrxStatus, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf(`SELECT TOP 1 ts.* FROM trx_status ts LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID
	WHERE ts.decision = 'REJ' AND ts.source_decision = 'DSR' AND tcp.IDNumber = '%s' AND CAST(ts.created_at as DATE) = '%s'`, idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) GetCurrentTrxWithReject(idNumber string) (data entity.TrxReject, err error) {

	currentDate := time.Now().Format(constant.FORMAT_DATE)

	if err = r.kmbOffDB.Raw(fmt.Sprintf(`SELECT 
	COUNT(CASE WHEN ts.source_decision = 'PMK' OR ts.source_decision = 'DSR' THEN 1 END) as reject_pmk_dsr,
	COUNT(CASE WHEN ts.source_decision != 'PMK' AND ts.source_decision != 'DSR' AND ts.source_decision != 'NKA' THEN 1 END) as reject_nik 
	FROM trx_status ts LEFT JOIN trx_customer_personal tcp ON ts.ProspectID = tcp.ProspectID
	WHERE ts.decision = 'REJ' AND tcp.IDNumber = '%s' AND CAST(ts.created_at as DATE) = '%s'`, idNumber, currentDate)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		}
		return
	}

	return
}

func (r repoHandler) ScanPreTrxJourney(prospectID string) (countMaster, countFiltering int, err error) {

	var (
		ftr    []entity.FilteringKMB
		master []entity.TrxMaster
	)

	if err = r.losDB.Raw(fmt.Sprintf("SELECT ProspectID FROM filtering_kmob WITH (nolock) WHERE ProspectID = '%s'", prospectID)).Scan(&ftr).Error; err != nil {
		return
	}

	countFiltering = len(ftr)

	if err = r.losDB.Raw(fmt.Sprintf("SELECT ProspectID FROM trx_master WITH (nolock) WHERE ProspectID = '%s'", prospectID)).Scan(&master).Error; err != nil {
		return
	}

	countMaster = len(master)

	return
}

func (r repoHandler) GetBiroData(prospectID string) (data entity.FilteringKMB, err error) {

	resultValid := os.Getenv("BIRO_VALID_DAYS")

	if err = r.losDB.Raw(fmt.Sprintf("SELECT TOP 1 ProspectID, ResultBiro FROM filtering_kmob WITH (nolock) WHERE ProspectID = '%s' AND ResultBiro IS NOT NULL AND source_credit_biro IS NOT NULL AND DATEADD(day, -%s, CAST(GETDATE() AS date)) <= CAST(DtmResponse AS date) ORDER BY created_at DESC", prospectID, resultValid)).Scan(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = errors.New(constant.ERROR_NOT_FOUND)
		}
		return
	}

	return
}
