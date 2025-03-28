package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common/platformevent"
	mockplatformevent "los-kmb-api/shared/common/platformevent/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDataPrinciple(t *testing.T) {
	os.Setenv("MDM_ASSET_URL", "https://dev-core-masterdata-asset-api.kbfinansia.com/api/v1/master-data/asset/search")
	os.Setenv("MARSEV_LOAN_AMOUNT_URL", "https://dev-marsev-api.kbfinansia.com/api/v1/calculate/loan-amount")
	os.Setenv("MARSEV_FILTER_PROGRAM_URL", "https://dev-marsev-api.kbfinansia.com/api/v1/marketing-programs/filter")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("MARSEV_AUTHORIZATION_KEY", "marsev-test-key")

	ctx := context.Background()
	reqID := utils.GenerateUUID()
	ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)

	accessToken := "test-token"

	birthDateStr := "1990-01-01"
	birthDate, _ := time.Parse("2006-01-02", birthDateStr)

	spouseBirthDateStr := "1991-01-01"
	spouseBirthDate, _ := time.Parse("2006-01-02", spouseBirthDateStr)

	samplePrincipleStepOne := entity.TrxPrincipleStepOne{
		KPMID:              12345,
		ProspectID:         "SAL-123",
		IDNumber:           "1234567890",
		SpouseIDNumber:     "0987654321",
		ResidenceAddress:   "Jl. Test No. 123",
		ResidenceRT:        "001",
		ResidenceRW:        "002",
		ResidenceProvince:  "DKI Jakarta",
		ResidenceCity:      "Jakarta Selatan",
		ResidenceKecamatan: "Kebayoran Baru",
		ResidenceKelurahan: "Melawai",
		ResidenceZipCode:   "12160",
		ResidenceAreaPhone: "021",
		ResidencePhone:     "5555555",
		HomeStatus:         "Own",
		StaySinceYear:      2015,
		StaySinceMonth:     1,
		BranchID:           "426",
		AssetCode:          "AST001",
		BPKBName:           "John Doe",
		ManufactureYear:    "2020",
	}

	samplePrincipleStepTwo := entity.TrxPrincipleStepTwo{
		ProspectID:         "SAL-123",
		IDNumber:           "1234567890",
		LegalName:          "John Doe",
		MobilePhone:        "08123456789",
		FullName:           "John Doe",
		BirthDate:          birthDate,
		BirthPlace:         "Jakarta",
		SurgateMotherName:  "Jane Doe",
		Gender:             "M",
		SpouseIDNumber:     "0987654321",
		Email:              "john.doe@example.com",
		Religion:           "Other",
		MaritalStatus:      "Married",
		SpouseIncome:       5000000,
		MonthlyFixedIncome: 10000000,
		ProfessionID:       "PROF001",
	}

	samplePrincipleStepThree := entity.TrxPrincipleStepThree{
		ProspectID:            "SAL-123",
		IDNumber:              "1234567890",
		Tenor:                 12,
		AF:                    1000000,
		NTF:                   10000000,
		OTR:                   12000000,
		DPAmount:              5000000,
		AdminFee:              250000,
		InstallmentAmount:     2000000,
		Dealer:                "PSA",
		MonthlyVariableIncome: 5000000,
		AssetCategoryID:       "BEBEK",
		FinancePurpose:        "Multiguna Pembayaran dengan Angsuran",
		TipeUsaha:             "Bahan Baku",
	}

	sampleFilteringKMB := entity.FilteringKMB{
		ProspectID:      "SAL-123",
		CustomerStatus:  "NEW",
		CustomerSegment: "REGULAR",
	}

	samplePrincipleEmergencyContact := entity.TrxPrincipleEmergencyContact{
		ProspectID:   "SAL-123",
		Name:         "Emergency Contact",
		Relationship: "Family",
		MobilePhone:  "08987654321",
		Address:      "Jl. Emergency No. 456",
		Rt:           "003",
		Rw:           "004",
		Kelurahan:    "Pondok Indah",
		Kecamatan:    "Kebayoran Lama",
		City:         "Jakarta Selatan",
		Province:     "DKI Jakarta",
		ZipCode:      "12310",
		AreaPhone:    "021",
		Phone:        "7777777",
	}

	sampleTrxKPM := entity.TrxKPM{
		IDNumber:                "3506124382000087",
		LegalName:               "Customer Wilen Testt",
		MobilePhone:             "085113218192",
		Email:                   "test.email6@gmail.com",
		BirthPlace:              "Jakarta",
		BirthDate:               birthDate,
		SurgateMotherName:       "IBU",
		Gender:                  "M",
		ResidenceAddress:        "Dermaga Baru",
		ResidenceRT:             "001",
		ResidenceRW:             "002",
		ResidenceProvince:       "Jakarta",
		ResidenceCity:           "Jakarta Timur",
		ResidenceKecamatan:      "Duren Sawit",
		ResidenceKelurahan:      "Klender",
		ResidenceZipCode:        "13470",
		BranchID:                "426",
		AssetCode:               "K-HND.MOTOR.ABSOLUTE REVO",
		ManufactureYear:         "2016",
		LicensePlate:            "B1839XVB",
		AssetUsageTypeCode:      "C",
		BPKBName:                "K",
		OwnerAsset:              "JONATHAN",
		LoanAmount:              2000000,
		MaxLoanAmount:           4000000,
		Tenor:                   12,
		InstallmentAmount:       356000,
		NumOfDependence:         1,
		MaritalStatus:           "M",
		SpouseIDNumber:          "3506126712000002",
		SpouseLegalName:         "YULINAR NIATI",
		SpouseBirthDate:         spouseBirthDate,
		SpouseBirthPlace:        "Jakarta",
		SpouseSurgateMotherName: "MAMA",
		SpouseMobilePhone:       "085880529111",
		Education:               "S1",
		ProfessionID:            "KRYSW",
		JobType:                 "003",
		JobPosition:             "M",
		EmploymentSinceMonth:    12,
		EmploymentSinceYear:     2020,
		MonthlyFixedIncome:      5000000,
		SpouseIncome:            5000000,
		NoChassis:               "MHKV1AA2JBK107322",
		HomeStatus:              "SD",
		StaySinceYear:           2020,
		StaySinceMonth:          4,
		KtpPhoto:                "https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/ktp_SAL-1140024081400003.jpg",
		SelfiePhoto:             "https://dev-platform-media.kbfinansia.com/media/reference/120000/SAL-1140024081400003/selfie_SAL-1140024081400003.jpg",
		AF:                      2915000,
		NTF:                     2943000,
		OTR:                     5300000,
		DPAmount:                2385000,
		AdminFee:                760000,
		Dealer:                  "PSA",
		AssetCategoryID:         "BEBEK",
		KPMID:                   83072,
		ResultPefindo:           "PASS",
		BakiDebet:               0,
		ReadjustContext:         "tenor",
		ReferralCode:            "TQ72AJ",
	}

	testcases := []struct {
		name                       string
		request                    request.PrincipleGetData
		resGetPrincipleStepOne     entity.TrxPrincipleStepOne
		errGetPrincipleStepOne     error
		resGetPrincipleStepTwo     entity.TrxPrincipleStepTwo
		errGetPrincipleStepTwo     error
		resGetPrincipleStepThree   entity.TrxPrincipleStepThree
		errGetPrincipleStepThree   error
		resGetFilteringResult      entity.FilteringKMB
		errGetFilteringResult      error
		resGetPrincipleEmergency   entity.TrxPrincipleEmergencyContact
		errGetPrincipleEmergency   error
		resGetTrxKPM               entity.TrxKPM
		errGetTrxKPM               error
		resMarsevLoanAmountCode    int
		resMarsevLoanAmountBody    string
		errMarsevLoanAmount        error
		resMDMGetAssetCode         int
		resMDMGetAssetBody         string
		errMDMGetAsset             error
		resMarsevFilterProgramCode int
		resMarsevFilterProgramBody string
		errMarsevFilterProgram     error
		expectedResult             map[string]interface{}
		expectedError              error
	}{
		{
			name: "success get domisili data",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Domisili",
			},
			resGetPrincipleStepOne: samplePrincipleStepOne,
			expectedResult: map[string]interface{}{
				"id_number":            "1234567890",
				"id_number_spouse":     "0987654321",
				"residence_address":    "Jl. Test No. 123",
				"residence_rt":         "001",
				"residence_rw":         "002",
				"residence_province":   "DKI Jakarta",
				"residence_city":       "Jakarta Selatan",
				"residence_kecamatan":  "Kebayoran Baru",
				"residence_kelurahan":  "Melawai",
				"residence_zipcode":    "12160",
				"residence_area_phone": "021",
				"residence_phone":      "5555555",
				"home_status":          "Own",
				"stay_since_year":      2015,
				"stay_since_month":     1,
			},
		},
		{
			name: "error get domisili data - KPM ID mismatch",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Domisili",
				KPMID:      54321,
			},
			resGetPrincipleStepOne: samplePrincipleStepOne,
			expectedError:          errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match"),
		},
		{
			name: "success get pemohon data",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Pemohon",
			},
			resGetPrincipleStepOne: samplePrincipleStepOne,
			resGetPrincipleStepTwo: samplePrincipleStepTwo,
			expectedResult: map[string]interface{}{
				"id_number":                  "1234567890",
				"legal_name":                 "John Doe",
				"mobile_phone":               "08123456789",
				"full_name":                  "John Doe",
				"birth_date":                 birthDate,
				"birth_place":                "Jakarta",
				"surgate_mother_name":        "Jane Doe",
				"gender":                     "M",
				"spouse_id_number":           "0987654321",
				"legal_address":              "",
				"legal_rt":                   "",
				"legal_rw":                   "",
				"legal_province":             "",
				"legal_city":                 "",
				"legal_kecamatan":            "",
				"legal_kelurahan":            "",
				"legal_zipcode":              "",
				"legal_area_phone":           "",
				"legal_phone":                "",
				"company_address":            "",
				"company_rt":                 "",
				"company_rw":                 "",
				"company_province":           "",
				"company_city":               "",
				"company_kecamatan":          "",
				"company_kelurahan":          "",
				"company_zipcode":            "",
				"company_area_phone":         "",
				"company_phone":              "",
				"monthly_fixed_income":       1e+07,
				"marital_status":             "Married",
				"spouse_income":              5000000,
				"selfie_photo":               nil,
				"ktp_photo":                  nil,
				"spouse_full_name":           nil,
				"spouse_birth_date":          nil,
				"spouse_birth_place":         nil,
				"spouse_gender":              nil,
				"spouse_legal_name":          nil,
				"spouse_mobile_phone":        nil,
				"spouse_surgate_mother_name": nil,
				"employment_since_month":     0,
				"employment_since_year":      0,
				"email":                      "john.doe@example.com",
				"religion":                   "Other",
			},
		},
		{
			name: "error get pemohon data - KPM ID mismatch",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Pemohon",
				KPMID:      54321,
			},
			resGetPrincipleStepOne: samplePrincipleStepOne,
			resGetPrincipleStepTwo: samplePrincipleStepTwo,
			expectedError:          errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match"),
		},
		{
			name: "success get emergency contact data",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Emergency",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleEmergency: samplePrincipleEmergencyContact,
			expectedResult: map[string]interface{}{
				"name":         "Emergency Contact",
				"relationship": "Family",
				"mobile_phone": "08987654321",
				"address":      "Jl. Emergency No. 456",
				"rt":           "003",
				"rw":           "004",
				"kelurahan":    "Pondok Indah",
				"kecamatan":    "Kebayoran Lama",
				"city":         "Jakarta Selatan",
				"province":     "DKI Jakarta",
				"zip_code":     "12310",
				"area_phone":   "021",
				"phone":        "7777777",
			},
		},
		{
			name: "error get emergency contact data - KPM ID mismatch",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Emergency",
				KPMID:      54321,
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleEmergency: samplePrincipleEmergencyContact,
			expectedError:            errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match"),
		},
		{
			name: "error get biaya data - KPM ID mismatch",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
				KPMID:      54321,
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			expectedError:            errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match"),
		},
		{
			name: "error get biaya data - get filtering result",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			errGetFilteringResult: errors.New("error get data"),
			expectedError:         errors.New("error get data"),
		},
		{
			name: "error get biaya data - marsev loan amount resp err",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			errMarsevLoanAmount:      errors.New("error get data"),
			expectedError:            errors.New("error get data"),
		},
		{
			name: "error get biaya data - marsev loan amount resp code",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  500,
			expectedError:            errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Loan Amount Error"),
		},
		{
			name: "error get biaya data - marsev loan amount resp unmarshal",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody:  "-",
			expectedError:            errors.New("invalid character ' ' in numeric literal"),
		},
		{
			name: "error get biaya data - get error asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			errMDMGetAsset: errors.New("error get asset mdm data"),
			expectedError:  errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout"),
		},
		{
			name: "error get biaya data - get error code asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resMDMGetAssetCode: 500,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
		{
			name: "error get biaya data - get error unmarshal asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `-`,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data asset list"),
		},
		{
			name: "error get biaya data - get empty record asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Biaya",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": []
				},
				"errors": null
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
		{
			name: "error get biaya data - get error marsev filter program",
			request: request.PrincipleGetData{
				ProspectID:     "SAL-123",
				Context:        "Biaya",
				FinancePurpose: "Multiguna Pembayaran dengan Angsuran",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			errMarsevFilterProgram: errors.New("error get data"),
			expectedError:          errors.New("error get data"),
		},
		{
			name: "error get biaya data - get error marsev filter program code",
			request: request.PrincipleGetData{
				ProspectID:     "SAL-123",
				Context:        "Biaya",
				FinancePurpose: "Multiguna Pembayaran dengan Angsuran",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resMarsevFilterProgramCode: 500,
			expectedError:              errors.New(constant.ERROR_UPSTREAM + " - Marsev Get Filter Program Error"),
		},
		{
			name: "error get biaya data - get error marsev filter program unmarshal",
			request: request.PrincipleGetData{
				ProspectID:     "SAL-123",
				Context:        "Biaya",
				FinancePurpose: "Multiguna Pembayaran dengan Angsuran",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resMarsevFilterProgramCode: 200,
			resMarsevFilterProgramBody: "-",
			expectedError:              errors.New("invalid character ' ' in numeric literal"),
		},
		{
			name: "success get biaya data",
			request: request.PrincipleGetData{
				ProspectID:     "SAL-123",
				Context:        "Biaya",
				FinancePurpose: "Multiguna Pembayaran dengan Angsuran",
			},
			resGetPrincipleStepOne:   samplePrincipleStepOne,
			resGetPrincipleStepTwo:   samplePrincipleStepTwo,
			resGetPrincipleStepThree: samplePrincipleStepThree,
			resGetFilteringResult:    sampleFilteringKMB,
			resMarsevLoanAmountCode:  200,
			resMarsevLoanAmountBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"loan_amount_maximum": 80000000,
					"amount_of_finance": 75000000,
					"dp_amount": 20000000,
					"dp_percent_final": 20,
					"ltv_percent_final": 80,
					"admin_fee_amount": 2500000,
					"provision_fee_amount": 1500000,
					"loan_amount_final": 75000000,
					"is_psa": true
				},
				"errors": null
			}`,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			resMarsevFilterProgramCode: 200,
			resMarsevFilterProgramBody: `{
				"code": 200,
				"message": "Success",
				"data": [{
					"id": "PROG001",
					"program_name": "Program Test",
					"mi_number": 123,
					"period_start": "2024-01-01",
					"period_end": "2024-12-31",
					"priority": 1,
					"description": "Test Program Description",
					"product_id": "PROD001",
					"product_offering_id": "OFF001",
					"product_offering_description": "Standard Offering",
					"tenors": [
						{
							"tenor": 12,
							"interest_rate": 10.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						},
						{
							"tenor": 24,
							"interest_rate": 11.5,
							"admin_fee": 2500000,
							"provision_amount": 500000
						}
					]
				}],
				"page_info": {
					"total": 1,
					"page": 1,
					"limit": 10
				},
				"errors": null
			}`,
			expectedResult: map[string]interface{}{
				"brand":            "HONDA",
				"dealer":           "PSA",
				"finance_purpose":  "Multiguna Pembayaran dengan Angsuran",
				"is_psa":           true,
				"manufacture_year": "2020",
				"model":            "HONDA VARIO 160 CBS",
				"tenors": []response.TenorInfo{
					{Tenor: 12},
					{Tenor: 24},
				},
				"type": "VARIO",
			},
		},
		{
			name: "success get readjust data",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Readjust",
			},
			resGetTrxKPM:       sampleTrxKPM,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": [
						{
							"asset_code": "MOT001",
							"asset_description": "HONDA VARIO 160",
							"asset_display": "HONDA VARIO 160 CBS",
							"asset_type_id": "2W",
							"branch_id": "426",
							"brand": "HONDA",
							"category_id": "SPM",
							"category_description": "SPORT MATIC",
							"is_electric": false,
							"model": "VARIO"
						}
					]
				},
				"errors": null
			}`,
			expectedResult: map[string]interface{}{
				"id_number":                  sampleTrxKPM.IDNumber,
				"legal_name":                 sampleTrxKPM.LegalName,
				"mobile_phone":               sampleTrxKPM.MobilePhone,
				"email":                      sampleTrxKPM.Email,
				"birth_place":                sampleTrxKPM.BirthPlace,
				"birth_date":                 birthDateStr,
				"surgate_mother_name":        sampleTrxKPM.SurgateMotherName,
				"gender":                     sampleTrxKPM.Gender,
				"residence_address":          sampleTrxKPM.ResidenceAddress,
				"residence_rt":               sampleTrxKPM.ResidenceRT,
				"residence_rw":               sampleTrxKPM.ResidenceRW,
				"residence_province":         sampleTrxKPM.ResidenceProvince,
				"residence_city":             sampleTrxKPM.ResidenceCity,
				"residence_kecamatan":        sampleTrxKPM.ResidenceKecamatan,
				"residence_kelurahan":        sampleTrxKPM.ResidenceKelurahan,
				"residence_zipcode":          sampleTrxKPM.ResidenceZipCode,
				"branch_id":                  sampleTrxKPM.BranchID,
				"asset_code":                 sampleTrxKPM.AssetCode,
				"asset_display":              "HONDA VARIO 160 CBS",
				"manufacture_year":           sampleTrxKPM.ManufactureYear,
				"license_plate":              sampleTrxKPM.LicensePlate,
				"asset_usage_type_code":      sampleTrxKPM.AssetUsageTypeCode,
				"bpkb_name_type":             sampleTrxKPM.BPKBName,
				"owner_asset":                sampleTrxKPM.OwnerAsset,
				"loan_amount":                sampleTrxKPM.LoanAmount,
				"max_loan_amount":            sampleTrxKPM.MaxLoanAmount,
				"tenor":                      sampleTrxKPM.Tenor,
				"installment_amount":         sampleTrxKPM.InstallmentAmount,
				"num_of_dependence":          sampleTrxKPM.NumOfDependence,
				"marital_status":             sampleTrxKPM.MaritalStatus,
				"spouse_id_number":           sampleTrxKPM.SpouseIDNumber,
				"spouse_legal_name":          sampleTrxKPM.SpouseLegalName,
				"spouse_birth_date":          spouseBirthDateStr,
				"spouse_birth_place":         sampleTrxKPM.SpouseBirthPlace,
				"spouse_surgate_mother_name": sampleTrxKPM.SpouseSurgateMotherName,
				"spouse_mobile_phone":        sampleTrxKPM.SpouseMobilePhone,
				"education":                  sampleTrxKPM.Education,
				"profession_id":              sampleTrxKPM.ProfessionID,
				"job_type":                   sampleTrxKPM.JobType,
				"job_position":               sampleTrxKPM.JobPosition,
				"employment_since_month":     sampleTrxKPM.EmploymentSinceMonth,
				"employment_since_year":      sampleTrxKPM.EmploymentSinceYear,
				"monthly_fixed_income":       sampleTrxKPM.MonthlyFixedIncome,
				"spouse_income":              sampleTrxKPM.SpouseIncome,
				"chassis_number":             sampleTrxKPM.NoChassis,
				"home_status":                sampleTrxKPM.HomeStatus,
				"stay_since_year":            sampleTrxKPM.StaySinceYear,
				"stay_since_month":           sampleTrxKPM.StaySinceMonth,
				"ktp_photo":                  sampleTrxKPM.KtpPhoto,
				"selfie_photo":               sampleTrxKPM.SelfiePhoto,
				"af":                         sampleTrxKPM.AF,
				"ntf":                        sampleTrxKPM.NTF,
				"otr":                        sampleTrxKPM.OTR,
				"down_payment_amount":        sampleTrxKPM.DPAmount,
				"admin_fee":                  sampleTrxKPM.AdminFee,
				"dealer":                     sampleTrxKPM.Dealer,
				"asset_category_id":          sampleTrxKPM.AssetCategoryID,
				"kpm_id":                     sampleTrxKPM.KPMID,
				"result_pefindo":             sampleTrxKPM.ResultPefindo,
				"baki_debet":                 sampleTrxKPM.BakiDebet,
				"readjust_context":           sampleTrxKPM.ReadjustContext,
				"rent_finish_date":           sampleTrxKPM.RentFinishDate,
				"referral_code":              sampleTrxKPM.ReferralCode,
			},
		},
		{
			name: "error get readjust data - error get trx kpm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Readjust",
			},
			errGetTrxKPM:  errors.New("error get data"),
			expectedError: errors.New("error get data"),
		},
		{
			name: "error get readjust data - KPM ID mismatch",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Readjust",
				KPMID:      54321,
			},
			resGetTrxKPM:  sampleTrxKPM,
			expectedError: errors.New(constant.INTERNAL_SERVER_ERROR + " - KPM ID does not match"),
		},
		{
			name: "error get readjust data - get error asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Readjust",
			},
			resGetTrxKPM:   sampleTrxKPM,
			errMDMGetAsset: errors.New("error get asset mdm data"),
			expectedError:  errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Timeout"),
		},
		{
			name: "error get readjust data - get error code asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Readjust",
			},
			resGetTrxKPM:       sampleTrxKPM,
			resMDMGetAssetCode: 500,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
		{
			name: "error get readjust data - get error unmarshal asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Readjust",
			},
			resGetTrxKPM:       sampleTrxKPM,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `-`,
			expectedError:      errors.New(constant.ERROR_UPSTREAM + " - error unmarshal data asset list"),
		},
		{
			name: "error get readjust data - get empty record asset mdm",
			request: request.PrincipleGetData{
				ProspectID: "SAL-123",
				Context:    "Readjust",
			},
			resGetTrxKPM:       sampleTrxKPM,
			resMDMGetAssetCode: 200,
			resMDMGetAssetBody: `{
				"code": 200,
				"message": "Success",
				"data": {
					"records": []
				},
				"errors": null
			}`,
			expectedError: errors.New(constant.ERROR_UPSTREAM + " - Call Asset MDM Error"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)
			mockPlatformEvent := mockplatformevent.NewPlatformEventInterface(t)
			var platformEvent platformevent.PlatformEventInterface = mockPlatformEvent

			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			mockRepository.On("GetPrincipleStepOne", tc.request.ProspectID).Return(tc.resGetPrincipleStepOne, tc.errGetPrincipleStepOne)
			mockRepository.On("GetPrincipleStepTwo", tc.request.ProspectID).Return(tc.resGetPrincipleStepTwo, tc.errGetPrincipleStepTwo)
			mockRepository.On("GetPrincipleStepThree", tc.request.ProspectID).Return(tc.resGetPrincipleStepThree, tc.errGetPrincipleStepThree)
			mockRepository.On("GetFilteringResult", tc.request.ProspectID).Return(tc.resGetFilteringResult, tc.errGetFilteringResult)
			mockRepository.On("GetPrincipleEmergencyContact", tc.request.ProspectID).Return(tc.resGetPrincipleEmergency, tc.errGetPrincipleEmergency)
			mockRepository.On("GetTrxKPM", tc.request.ProspectID).Return(tc.resGetTrxKPM, tc.errGetTrxKPM)

			if tc.request.Context == "Biaya" {
				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				url := os.Getenv("MARSEV_LOAN_AMOUNT_URL")
				httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resMarsevLoanAmountCode, tc.resMarsevLoanAmountBody))
				resp, _ := rst.R().Post(url)

				mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, mock.Anything, map[string]string{
					"Content-Type":  "application/json",
					"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
				}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.request.ProspectID, accessToken).Return(resp, tc.errMarsevLoanAmount)

				rst2 := resty.New()
				httpmock.ActivateNonDefault(rst2.GetClient())
				defer httpmock.DeactivateAndReset()

				url2 := os.Getenv("MDM_ASSET_URL")
				httpmock.RegisterResponder(constant.METHOD_POST, url2, httpmock.NewStringResponder(tc.resMDMGetAssetCode, tc.resMDMGetAssetBody))
				resp2, _ := rst2.R().Post(url2)

				mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url2, mock.Anything, map[string]string{}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.request.ProspectID, accessToken).Return(resp2, tc.errMDMGetAsset)

				rst3 := resty.New()
				httpmock.ActivateNonDefault(rst3.GetClient())
				defer httpmock.DeactivateAndReset()

				url3 := os.Getenv("MARSEV_FILTER_PROGRAM_URL")
				httpmock.RegisterResponder(constant.METHOD_POST, url3, httpmock.NewStringResponder(tc.resMarsevFilterProgramCode, tc.resMarsevFilterProgramBody))
				resp3, _ := rst.R().Post(url3)

				mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url3, mock.Anything, map[string]string{
					"Content-Type":  "application/json",
					"Authorization": os.Getenv("MARSEV_AUTHORIZATION_KEY"),
				}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.request.ProspectID, accessToken).Return(resp3, tc.errMarsevFilterProgram)
			}

			if tc.request.Context == "Readjust" {
				rst := resty.New()
				httpmock.ActivateNonDefault(rst.GetClient())
				defer httpmock.DeactivateAndReset()

				url := os.Getenv("MDM_ASSET_URL")
				httpmock.RegisterResponder(constant.METHOD_POST, url, httpmock.NewStringResponder(tc.resMDMGetAssetCode, tc.resMDMGetAssetBody))
				resp, _ := rst.R().Post(url)

				payloadAsset, _ := json.Marshal(map[string]interface{}{
					"branch_id": tc.resGetTrxKPM.BranchID,
					"lob_id":    11,
					"page_size": 10,
					"search":    tc.resGetTrxKPM.AssetCode,
				})

				mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, url, payloadAsset, map[string]string{}, constant.METHOD_POST, false, 0, mock.AnythingOfType("int"), tc.request.ProspectID, accessToken).Return(resp, tc.errMDMGetAsset)
			}

			usecase := NewUsecase(mockRepository, mockHttpClient, platformEvent)

			result, err := usecase.GetDataPrinciple(ctx, tc.request, accessToken)

			if tc.expectedError != nil {
				require.Error(t, err)
				require.EqualError(t, err, tc.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedResult, result)
			}
		})
	}
}
