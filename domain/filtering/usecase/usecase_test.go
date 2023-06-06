package usecase_test

// import (
// 	"encoding/json"
// 	"fmt"
// 	"los-kmb-api/domain/filtering/repository"
// 	"los-kmb-api/models/entity"
// 	"los-kmb-api/models/request"
// 	"los-kmb-api/models/response"
// 	constants "los-kmb-api/shared/constant"
// 	"los-kmb-api/shared/httpclient"
// 	"os"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/jarcoal/httpmock"
// 	"github.com/stretchr/testify/require"
// 	"gopkg.in/resty.v1"
// )

// func TestDummyData(t *testing.T) {
// 	dummyFound := entity.DummyColumn{
// 		NoKTP: "3276052009920006",
// 		Value: `{"NoKTP":"3276052009920006","BirthPlace":"JAKARTA","LegalAddress":"KRANJI","LegalKelurahan":"KRANJI","LegalKecamatan":"BEKASI BARAT","LegalCity":"BEKASI","LegalZipCode":"17135","ResidenceAddress":"KRANJI","ResidenceKelurahan":"KRANJI","ResidenceKecamatan":"BEKASI BARAT","ResidenceCity":"BEKASI","ResidenceZipCode":null,"CompanyAddress":null,"CompanyKelurahan":null,"CompanyKecamatan":null,"CompanyCity":null,"CompanyZipCode":null,"PersonalNPWP":null,"Education":null,"MaritalStatus":null,"NumOfDependence":null,"HomeStatus":null,"ProfessionID":null,"JobTypeID":null,"JobPos":null,"MonthlyFixedIncome":"10000000.00","SpouseIncome":"0.00","MonthlyVariableIncome":"0.00","TotalInstallment":"200000.00","TotalInstallmentNAP":"0.00","CustomerID":"","BadType":null,"MaxOverdueDays":"24","MaxOverdueDaysROAO":"5","NumOfAssetInventoried":"0","Gender":"F","EmergencyContactAddress":null,"MaxOverdueDaysforActiveAgreement":"0","MaxOverdueDaysforPrevEOM":null,"SisaJumlahAngsuran":"2","RRDDate":null,"NumberofAgreement":"1","WorkSinceYear":null,"OutstandingPrincipal":"1140685.91","OSInstallmentDue":"0.00","SpouseName":null,"SpouseIDNumber":null,"SpouseBirthDate":null,"SpouseBadType":null,"IsRestructure":"0","InstallmentAmount_ChassisNo":null,"DownPayment":null,"TotalOTR":null,"NoRangka":null,"IsSimiliar":null}`,
// 	}

// 	var testcase = []struct {
// 		dummy            entity.DummyColumn
// 		expected         entity.DummyColumn
// 		err, errExpected error
// 		label            string
// 	}{
// 		{
// 			dummy:    dummyFound,
// 			expected: dummyFound,
// 			label:    "TEST_DUMMY_FOUND",
// 		}, {
// 			err:         fmt.Errorf("timeout exceed"),
// 			errExpected: fmt.Errorf("timeout exceed"),
// 			label:       "TEST_DUMMY_ERROR",
// 		},
// 	}

// 	for _, test := range testcase {

// 		mockRepository := new(repository.MockRepository)

// 		mockRepository.On("DummyData", fmt.Sprintf("SELECT * FROM dupcheck_confins_new WHERE NoKTP = '%s'", dummyFound.NoKTP)).Return(test.dummy, test.err)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		usecase := NewApi(mockRepository, mockHttpClient)

// 		getDummy, err := usecase.GetDummyData(dummyFound.NoKTP)

// 		fmt.Println(test.label)

// 		require.Equal(t, test.expected, getDummy)
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestGetPBKDummy(t *testing.T) {

// 	dummyFound := entity.DummyPBK{
// 		IDNumber: "1234567890123456",
// 		Response: `{
// 			"code": "200",
// 			"status": "SUCCESS",
// 			"result": {
// 				"search_id": "kp_61e69179e9ac2",
// 				"pefindo_id": "1042512812",
// 				"score": "VERY HIGH RISK",
// 				"max_overdue": "0",
// 				"max_overdue_last12months": "7",
// 				"angsuran_aktif_pbk": "1733927.25",
// 				"angsuran_aktif_pbk_konsumen": "1733927.25",
// 				"wo_contract": true,
// 				"wo_ada_agunan": false,
// 				"detail_report": "http:\/\/10.0.0.161\/los-symlink\/pefindo\/pdf\/dummy.pdf",
// 				"total_baki_debet_non_agunan": 5000000
// 			},
// 			"konsumen": null,
// 			"pasangan": {
// 				"search_id": "kp_61e69179e9ac2",
// 				"pefindo_id": "1042512812",
// 				"score": "VERY HIGH RISK",
// 				"max_overdue": "0",
// 				"max_overdue_last12months": "7",
// 				"detail_report": "http:\/\/10.0.0.161\/los-symlink\/pefindo\/pdf\/dummy.pdf",
// 				"baki_debet_non_agunan": 5000000
// 			},
// 			"timestamp": "1642500540"
// 		}`,
// 	}

// 	var testcase = []struct {
// 		dummypbk         entity.DummyPBK
// 		expected         entity.DummyPBK
// 		err, errExpected error
// 		label            string
// 	}{
// 		{
// 			dummypbk: dummyFound,
// 			expected: dummyFound,
// 			label:    "TEST_DUMMY_PBK_FOUND",
// 		}, {
// 			err:         fmt.Errorf("timeout exceed"),
// 			errExpected: fmt.Errorf("timeout exceed"),
// 			label:       "TEST_DUMMY_PBK_ERROR",
// 		},
// 	}

// 	for _, test := range testcase {

// 		mockRepository := new(repository.MockRepository)

// 		mockRepository.On("DummyDataPbk", fmt.Sprintf("SELECT * FROM new_pefindo_kmb WHERE IDNumber = '%s'", dummyFound.IDNumber)).Return(test.dummypbk, test.err)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		usecase := NewApi(mockRepository, mockHttpClient)

// 		getDummyPBK, err := usecase.GetDummyPBK(dummyFound.IDNumber)

// 		fmt.Println(test.label)

// 		require.Equal(t, test.expected, getDummyPBK)
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestGetProfesionID(t *testing.T) {

// 	layoutFormat := "2006-01-02 15:04:05.000"
// 	value := "2023-01-05 15:06:10.800"

// 	date, _ := time.Parse(layoutFormat, value)

// 	professionid := entity.ProfessionGroup{
// 		ID:        "CC62AFEC-E283-4239-9935-ED1E49D33DF8",
// 		Prefix:    "ANG",
// 		Name:      "KARYAWAN",
// 		CreatedAt: date,
// 	}

// 	var testcase = []struct {
// 		professionid     entity.ProfessionGroup
// 		expected         entity.ProfessionGroup
// 		err, errExpected error
// 		label            string
// 	}{
// 		{
// 			professionid: professionid,
// 			expected:     professionid,
// 			label:        "TEST_ProfesionID_FOUND",
// 		}, {
// 			err:         fmt.Errorf("timeout exceed"),
// 			errExpected: fmt.Errorf("timeout exceed"),
// 			label:       "TEST_ProfesionID_ERROR",
// 		},
// 	}

// 	for _, test := range testcase {

// 		mockRepository := new(repository.MockRepository)

// 		mockRepository.On("DataProfessionGroup", fmt.Sprintf("SELECT * FROM profession_group WHERE prefix = '%s'", professionid.Prefix)).Return(test.professionid, test.err)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		usecase := NewApi(mockRepository, mockHttpClient)

// 		GetDataProfessionGroup, err := usecase.GetDataProfessionGroup(professionid.Prefix)

// 		fmt.Println(test.label)

// 		require.Equal(t, test.expected, GetDataProfessionGroup)
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestGetBranchDP(t *testing.T) {

// 	BranchID := "426"
// 	StatusKonsumen := "RO/AO"

// 	rangedb := []entity.RangeBranchDp{{ID: "379f6c45-8baf-4152-a3a6-c47e3452ecac",
// 		Name:       "baki_debet_1",
// 		RangeStart: 0,
// 		RangeEnd:   3000000,
// 		CreatedAt:  "2022-09-19 11:46:32.000"}}

// 	var testcase = []struct {
// 		rangedb          []entity.RangeBranchDp
// 		expected         []entity.RangeBranchDp
// 		err, errExpected error
// 		label            string
// 	}{
// 		{
// 			rangedb:  rangedb,
// 			expected: rangedb,
// 			label:    "TEST_RangeDB_FOUND",
// 		}, {
// 			err:         fmt.Errorf("timeout exceed"),
// 			errExpected: fmt.Errorf("timeout exceed"),
// 			label:       "TEST_RangeDB_ERROR",
// 		},
// 	}

// 	for _, test := range testcase {

// 		mockRepository := new(repository.MockRepository)

// 		mockRepository.On("DataGetMappingDp", fmt.Sprintf("SELECT mbd.* FROM dbo.mapping_branch_dp mdp LEFT JOIN dbo.mapping_baki_debet mbd ON mdp.baki_debet = mbd.id LEFT JOIN dbo.master_list_dp mld ON mdp.master_list_dp = mld.id WHERE mdp.branch = '%s' AND mdp.customer_status = '%s'", BranchID, StatusKonsumen)).Return(test.rangedb, test.err)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		usecase := NewApi(mockRepository, mockHttpClient)

// 		GetDataProfessionGroup, err := usecase.GetDataGetMappingDp(BranchID, StatusKonsumen)

// 		fmt.Println(test.label)

// 		require.Equal(t, test.expected, GetDataProfessionGroup)
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestBranchDpData(t *testing.T) {

// 	BranchID := "426"

// 	StatusKonsumen := "RO/AO"
// 	ProfessionGroup := "KARYAWAN"
// 	totalBakiDebetNonAgunan := 2000000

// 	queryAdd := fmt.Sprintf("AND a.customer_status = '%s'AND a.profession_group IS NULL", StatusKonsumen)

// 	branchdb := entity.BranchDp{
// 		Branch:          "426",
// 		CustomerStatus:  "RO/AO",
// 		ProfessionGroup: "",
// 		MinimalDpName:   "DP_20",
// 		MinimalDpValue:  20.00,
// 	}

// 	var testcase = []struct {
// 		branchdb         entity.BranchDp
// 		expected         entity.BranchDp
// 		err, errExpected error
// 		label            string
// 	}{
// 		{
// 			branchdb: branchdb,
// 			expected: branchdb,
// 			label:    "TEST_BranchDB_FOUND",
// 		}, {
// 			err:         fmt.Errorf("timeout exceed"),
// 			errExpected: fmt.Errorf("timeout exceed"),
// 			label:       "TEST_BranchDB_ERROR",
// 		},
// 	}

// 	for _, test := range testcase {

// 		mockRepository := new(repository.MockRepository)

// 		mockRepository.On("BranchDpData", fmt.Sprintf("SELECT TOP 1 a.branch,a.customer_status,a.profession_group,b.minimal_dp_name,b.minimal_dp_value FROM dbo.mapping_branch_dp a WITH (NOLOCK) INNER JOIN dbo.master_list_dp b WITH (NOLOCK) ON a.master_list_dp = b.id LEFT JOIN dbo.mapping_baki_debet c WITH (NOLOCK) ON a.baki_debet = c.id WHERE a.branch = '%s' %s ORDER BY a.created_at ASC", BranchID, queryAdd)).Return(test.branchdb, test.err)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		usecase := NewApi(mockRepository, mockHttpClient)

// 		GetDataProfessionGroup, err := usecase.GetBranchDpTest(BranchID, StatusKonsumen, ProfessionGroup, totalBakiDebetNonAgunan)

// 		fmt.Println(test.label)

// 		require.Equal(t, test.expected, GetDataProfessionGroup)
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestFilteringBlackList(t *testing.T) {

// 	body := request.BodyRequest{
// 		ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
// 		Data: request.Data{
// 			BPKBName:          "K",
// 			ProspectID:        "TESTNEWROAO000000073",
// 			BranchID:          "426",
// 			IDNumber:          "3276052009920006",
// 			LegalName:         "AGUS",
// 			BirthPlace:        "JAKARTA",
// 			BirthDate:         "1983-08-17",
// 			SurgateMotherName: "ENEH",
// 			Gender:            "M",
// 			MaritalStatus:     "S",
// 			ProfessionID:      "KRYSW",
// 			Spouse:            nil,
// 			MobilePhone:       "087770425933",
// 		},
// 	}

// 	UpdateData := entity.ApiDupcheckKmbUpdate{
// 		ResultDupcheckKonsumen: `{"bad_type":"","birth_date":"","birth_place":"","company_address":"","company_city":"","company_kecamatan":"","company_kelurahan":"","company_zipcode":"","customer_id":"","education":"","emergency_contact_address":"","full_name":"","gender":"M","home_status":"","id_number":"","is_restructure":0,"is_similiar":0,"job_pos":"","job_type_id":"","lagal_zipcode":"","legal_address":"","legal_city":"","legal_kecamatan":"","legal_kelurahan":"","marital_status":"","max_overduedays":0,"max_overduedays_for_active_agreement":0,"max_overduedays_for_prev_eom":0,"max_overduedays_roao":0,"monthly_fixed_income":0,"monthly_variable_income":0,"num_of_asset_inventoried":0,"num_of_dependence":0,"number_of_agreement":0,"os_installmentdue":0,"outstanding_principal":0,"overduedays_aging":0,"personal_npwp":"","profession_id":"","residence_address":"","residence_city":"","residence_kecamatan":"","residence_kelurahan":"","residence_zipcode":"","rrd_date":"","sisa_jumlah_angsuran":0,"spouse_income":0,"surgate_mother_name":"","total_installment":0,"total_installment_nap":0,"work_since_year":"","installment_amount_chassis_no":""}`,
// 		RequestID:              ID,
// 	}

// 	dupcheckbody, _ := json.Marshal(body)

// 	param_konsumen := map[string]string{
// 		"birth_date":          body.Data.BirthDate,
// 		"id_number":           body.Data.IDNumber,
// 		"legal_name":          body.Data.LegalName,
// 		"surgate_mother_name": body.Data.SurgateMotherName,
// 		"transaction_id":      body.Data.ProspectID,
// 	}

// 	resultDupcheck, _ := json.Marshal(response.DupcheckResult{
// 		Code:           constants.CODE_NEW_CUSTOMER,
// 		Decision:       constants.DECISION_PASS,
// 		Reason:         constants.REASON_NEW_CUSTOMER,
// 		StatusKonsumen: constants.STATUS_KONSUMEN_NEW,
// 	})

// 	var testcase = []struct {
// 		payload          string
// 		expected         interface{}
// 		code             int
// 		body             string
// 		err, errExpected error
// 		label            string
// 	}{
// 		{

// 			body:     string(dupcheckbody),
// 			expected: string(resultDupcheck),
// 			code:     200,
// 			label:    "TEST_FILTERING_BLACK_LIST_FOUND",
// 		},
// 		{
// 			body:        string(dupcheckbody),
// 			expected:    `{"code":0,"decision":"","reason":"","status_konsumen":""}`,
// 			err:         fmt.Errorf("FAILED FETCHING DATA CONFINS KONSUMEN"),
// 			errExpected: fmt.Errorf("FAILED FETCHING DATA CONFINS KONSUMEN"),
// 			label:       "TEST_FILTERING_BLACK_LIST_ERROR",
// 		},
// 	}

// 	for _, test := range testcase {

// 		rst := resty.New()
// 		httpmock.ActivateNonDefault(rst.GetClient())
// 		defer httpmock.DeactivateAndReset()

// 		mockRepository := new(repository.MockRepository)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		httpmock.RegisterResponder("POST", os.Getenv("DUPCHECK_URL"), httpmock.NewStringResponder(test.code, test.body))

// 		resp, _ := rst.R().SetBody(param_konsumen).Post(os.Getenv("DUPCHECK_URL"))

// 		timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

// 		usecase := NewApi(mockRepository, mockHttpClient)
// 		mockHttpClient.On("CallWebSocket", os.Getenv("DUPCHECK_URL"), param_konsumen, map[string]string{}, timeOut).Return(resp, test.err)

// 		mockRepository.On("FilteringBlackList").Return(resultDupcheck, nil)

// 		mockRepository.On("UpdateData", UpdateData).Return(nil)

// 		GetDataProfessionGroup, err := usecase.FilteringBlackList(body, ID)

// 		result, _ := json.Marshal(GetDataProfessionGroup)

// 		fmt.Println(test.label)
// 		require.Equal(t, test.expected, string(result))
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestFilteringKreditmu(t *testing.T) {

// 	body := request.BodyRequest{
// 		ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
// 		Data: request.Data{
// 			BPKBName:          "K",
// 			ProspectID:        "TESTNEWROAO000000073",
// 			BranchID:          "426",
// 			IDNumber:          "3276052009920006",
// 			LegalName:         "AGUS",
// 			BirthPlace:        "JAKARTA",
// 			BirthDate:         "1983-08-17",
// 			SurgateMotherName: "ENEH",
// 			Gender:            "M",
// 			MaritalStatus:     "S",
// 			ProfessionID:      "KRYSW",
// 			Spouse:            nil,
// 			MobilePhone:       "087770425933",
// 		},
// 	}

// 	UpdateData := entity.ApiDupcheckKmbUpdate{
// 		ResultKreditmu: `{
// 			"code": "CORE-API-005",
// 			"message": "data yang anda minta tidak ditemukan.",
// 			"data": null,
// 			"errors": null,
// 			"request_id": "b24fa8f6-983b-401a-8226-5b5d35cbab92",
// 			"timestamp": "2023-01-16 15:46:09"
// 		  }`,
// 		RequestID: ID,
// 	}

// 	dupcheckbody, _ := json.Marshal(body)

// 	paramKreditmu := map[string]string{
// 		"birth_date":          body.Data.BirthDate,
// 		"id_number":           body.Data.IDNumber,
// 		"legal_name":          body.Data.LegalName,
// 		"surgate_mother_name": body.Data.SurgateMotherName,
// 	}

// 	resultDupcheck, _ := json.Marshal(response.DupcheckResult{
// 		Code:           constants.CODE_NEW_CUSTOMER,
// 		Decision:       constants.DECISION_PASS,
// 		Reason:         constants.REASON_NEW_CUSTOMER,
// 		StatusKonsumen: constants.STATUS_KONSUMEN_NEW,
// 	})

// 	var testcase = []struct {
// 		payload          string
// 		expected         interface{}
// 		code             int
// 		body             string
// 		err, errExpected error
// 		label            string
// 	}{
// 		{

// 			body:     string(dupcheckbody),
// 			expected: string(resultDupcheck),
// 			code:     200,
// 			label:    "TEST_FILTERING_KREDITMU_FOUND",
// 		},
// 		{
// 			body:        string(dupcheckbody),
// 			expected:    `{"code":0,"decision":"","reason":"","status_konsumen":""}`,
// 			err:         fmt.Errorf("FAILED FETCHING DATA KREDITMU"),
// 			errExpected: fmt.Errorf("FAILED FETCHING DATA KREDITMU"),
// 			label:       "TEST_FILTERING_KREDITMU_ERROR",
// 		},
// 	}

// 	for _, test := range testcase {

// 		rst := resty.New()
// 		httpmock.ActivateNonDefault(rst.GetClient())
// 		defer httpmock.DeactivateAndReset()

// 		mockRepository := new(repository.MockRepository)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		httpmock.RegisterResponder("POST", os.Getenv("KREDITMU_URL"), httpmock.NewStringResponder(test.code, test.body))

// 		resp, _ := rst.R().SetBody(paramKreditmu).Post(os.Getenv("KREDITMU_URL"))

// 		timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

// 		usecase := NewApi(mockRepository, mockHttpClient)
// 		mockHttpClient.On("CallWebSocket", os.Getenv("KREDITMU_URL"), paramKreditmu, map[string]string{}, timeOut).Return(resp, test.err)

// 		mockRepository.On("FilteringKreditmu").Return(resultDupcheck, nil)

// 		mockRepository.On("UpdateData", UpdateData).Return(nil)

// 		GetDataProfessionGroup, err := usecase.FilteringBlackList(body, ID)

// 		result, _ := json.Marshal(GetDataProfessionGroup)

// 		fmt.Println(test.label)
// 		require.Equal(t, test.expected, string(result))
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestFilteringPefindo(t *testing.T) {

// 	body := request.BodyRequest{
// 		ClientKey: "$2y$10$5X1gt1p11.CWbm.Gtgg7E.bATsMup..KhU2HeY/RJRteoW7UwT9N6",
// 		Data: request.Data{
// 			BPKBName:          "K",
// 			ProspectID:        "TESTNEWROAO000000073",
// 			BranchID:          "426",
// 			IDNumber:          "3276052009920006",
// 			LegalName:         "AGUS",
// 			BirthPlace:        "JAKARTA",
// 			BirthDate:         "1983-08-17",
// 			SurgateMotherName: "ENEH",
// 			Gender:            "M",
// 			MaritalStatus:     "S",
// 			ProfessionID:      "KRYSW",
// 			Spouse:            nil,
// 			MobilePhone:       "087770425933",
// 		},
// 	}

// 	UpdateData := entity.ApiDupcheckKmbUpdate{
// 		RequestID: ID,
// 	}

// 	dupcheckbody, _ := json.Marshal(body)

// 	paramPefindo := map[string]string{
// 		"ClientKey":         os.Getenv("CLIENTKEY_CORE_PBK"),
// 		"IDMember":          constants.USER_PBK_KMB_FILTEERING,
// 		"user":              constants.USER_PBK_KMB_FILTEERING,
// 		"IDNumber":          body.Data.IDNumber,
// 		"ProspectID":        body.Data.ProspectID,
// 		"LegalName":         body.Data.LegalName,
// 		"BirthDate":         body.Data.BirthDate,
// 		"SurgateMotherName": body.Data.SurgateMotherName,
// 		"Gender":            body.Data.Gender,
// 		"MaritalStatus":     body.Data.MaritalStatus,
// 	}

// 	resultDupcheck, _ := json.Marshal(response.DupcheckResult{
// 		Code:           constants.PBK_NO_HIT,
// 		Decision:       "NOT HIT",
// 		Reason:         "Akses ke PBK ditutup",
// 		StatusKonsumen: "",
// 	})

// 	var testcase = []struct {
// 		payload          string
// 		expected         interface{}
// 		code             int
// 		body             string
// 		err, errExpected error
// 		label            string
// 	}{
// 		{

// 			body:     string(dupcheckbody),
// 			expected: string(resultDupcheck),
// 			code:     200,
// 			label:    "TEST_FILTERING_PBK_FOUND",
// 		},
// 	}

// 	for _, test := range testcase {

// 		rst := resty.New()
// 		httpmock.ActivateNonDefault(rst.GetClient())
// 		defer httpmock.DeactivateAndReset()

// 		mockRepository := new(repository.MockRepository)

// 		mockHttpClient := new(httpclient.MockHttpClient)

// 		httpmock.RegisterResponder("POST", os.Getenv("PBK_URL"), httpmock.NewStringResponder(test.code, test.body))

// 		resp, _ := rst.R().SetBody(paramPefindo).Post(os.Getenv("PBK_URL"))

// 		timeOut, _ := strconv.Atoi(os.Getenv("DUPCHECK_API_TIMEOUT"))

// 		usecase := NewApi(mockRepository, mockHttpClient)
// 		mockHttpClient.On("CallWebSocket", os.Getenv("PBK_URL"), paramPefindo, map[string]string{}, timeOut).Return(resp, test.err)

// 		mockRepository.On("FilteringPefindo").Return(resultDupcheck, nil)

// 		mockRepository.On("UpdateData", UpdateData).Return(nil)

// 		GetDataProfessionGroup, err := usecase.FilteringPefindo(body, constants.STATUS_KONSUMEN_NEW, ID)

// 		result, _ := json.Marshal(GetDataProfessionGroup)

// 		fmt.Println(test.label)
// 		require.Equal(t, test.expected, string(result))
// 		require.Equal(t, test.errExpected, err)

// 	}
// }

// func TestGetDupcheck(t *testing.T) {

// 	param_konsumen, _ := json.Marshal(map[string]string{
// 		"birth_date":          "1983-08-17",
// 		"id_number":           "3276052009920006",
// 		"legal_name":          "AGUS",
// 		"surgate_mother_name": "ENEH",
// 		"transaction_id":      "TESTNEWROAO000000073",
// 	})

// 	testcase := []struct {
// 		payload  string
// 		expected interface{}
// 		body     string
// 		code     int
// 		err      error
// 		label    string
// 	}{
// 		{
// 			payload: string(param_konsumen),
// 			expected: `{
// 				"id_number":                            "3276052009920006",
// 				"birth_date":                           "JAKARTA",
// 				"legal_address":                        "KRANJI",
// 				"legal_kelurahan":                      "KRANJI",
// 				"legal_kecamatan":                      "BEKASI BARAT",
// 				"legal_city":                           "BEKASI",
// 				"lagal_zipcode":                        "17135",
// 				"residence_address":                    "KRANJI",
// 				"residence_kelurahan":                  "KRANJI",
// 				"residence_kecamatan":                  "BEKASI BARAT",
// 				"residence_city":                       "BEKASI",
// 				"residence_zipcode":                    "",
// 				"company_address":                      "",
// 				"company_kelurahan":                    "",
// 				"company_kecamatan":                    "",
// 				"company_city":                         "",
// 				"company_zipcode":                      "",
// 				"personal_npwp":                        "",
// 				"education":                            "",
// 				"marital_status":                       "",
// 				"num_of_dependence":                    0,
// 				"home_status":                          "",
// 				"profession_id":                        "",
// 				"job_type_id":                          "",
// 				"job_pos":                              "",
// 				"monthly_fixed_income":                 10000000,
// 				"spouse_income":                        0,
// 				"monthly_variable_income":              0,
// 				"total_installment":                    200000,
// 				"total_installment_nap":                0,
// 				"customer_id":                          "",
// 				"bad_type":                             "",
// 				"max_overduedays":                      24,
// 				"max_overduedays_roao":                 5,
// 				"num_of_asset_inventoried":             0,
// 				"gender":                               "F",
// 				"emergency_contact_address":            "",
// 				"max_overduedays_for_active_agreement": 0,
// 				"max_overduedays_for_prev_eom":         0,
// 				"sisa_jumlah_angsuran":                 2,
// 				"rrd_date":                             "",
// 				"number_of_agreement":                  1,
// 				"work_since_year":                      "",
// 				"outstanding_principal":                1140686,
// 				"os_installmentdue":                    0,
// 				"is_restructure":                       0,
// 				"installment_amount_chassis_no":        "",
// 				"is_similiar":                          0,
// 			}`,
// 			body: `{
// 				"id_number":                            "3276052009920006",
// 				"birth_date":                           "JAKARTA",
// 				"legal_address":                        "KRANJI",
// 				"legal_kelurahan":                      "KRANJI",
// 				"legal_kecamatan":                      "BEKASI BARAT",
// 				"legal_city":                           "BEKASI",
// 				"lagal_zipcode":                        "17135",
// 				"residence_address":                    "KRANJI",
// 				"residence_kelurahan":                  "KRANJI",
// 				"residence_kecamatan":                  "BEKASI BARAT",
// 				"residence_city":                       "BEKASI",
// 				"residence_zipcode":                    "",
// 				"company_address":                      "",
// 				"company_kelurahan":                    "",
// 				"company_kecamatan":                    "",
// 				"company_city":                         "",
// 				"company_zipcode":                      "",
// 				"personal_npwp":                        "",
// 				"education":                            "",
// 				"marital_status":                       "",
// 				"num_of_dependence":                    0,
// 				"home_status":                          "",
// 				"profession_id":                        "",
// 				"job_type_id":                          "",
// 				"job_pos":                              "",
// 				"monthly_fixed_income":                 10000000,
// 				"spouse_income":                        0,
// 				"monthly_variable_income":              0,
// 				"total_installment":                    200000,
// 				"total_installment_nap":                0,
// 				"customer_id":                          "",
// 				"bad_type":                             "",
// 				"max_overduedays":                      24,
// 				"max_overduedays_roao":                 5,
// 				"num_of_asset_inventoried":             0,
// 				"gender":                               "F",
// 				"emergency_contact_address":            "",
// 				"max_overduedays_for_active_agreement": 0,
// 				"max_overduedays_for_prev_eom":         0,
// 				"sisa_jumlah_angsuran":                 2,
// 				"rrd_date":                             "",
// 				"number_of_agreement":                  1,
// 				"work_since_year":                      "",
// 				"outstanding_principal":                1140686,
// 				"os_installmentdue":                    0,
// 				"is_restructure":                       0,
// 				"installment_amount_chassis_no":        "",
// 				"is_similiar":                          0,
// 			}`,
// 			code:  200,
// 			label: "TEST_DUPCHECK_NEW",
// 		},
// 		{
// 			payload: string(param_konsumen),
// 			expected: `{
// 				"id_number": "3276052009920006",
// 				"birth_date": "JAKARTA",
// 				"legal_address": "KRANJI",
// 				"legal_kelurahan": "KRANJI",
// 				"legal_kecamatan": "BEKASI BARAT",
// 				"legal_city": "BEKASI",
// 				"lagal_zipcode": "17135",
// 				"residence_address": "KRANJI",
// 				"residence_kelurahan": "KRANJI",
// 				"residence_kecamatan": "BEKASI BARAT",
// 				"residence_city": "BEKASI",
// 				"residence_zipcode": "",
// 				"company_address": "",
// 				"company_kelurahan": "",
// 				"company_kecamatan": "",
// 				"company_city": "",
// 				"company_zipcode": "",
// 				"personal_npwp": "",
// 				"education": "",
// 				"marital_status": "",
// 				"num_of_dependence": 0,
// 				"home_status": "",
// 				"profession_id": "",
// 				"job_type_id": "",
// 				"job_pos": "",
// 				"monthly_fixed_income": 10000000,
// 				"spouse_income": 0,
// 				"monthly_variable_income": 0,
// 				"total_installment": 200000,
// 				"total_installment_nap": 0,
// 				"customer_id": "40600051394",
// 				"bad_type": "",
// 				"max_overduedays": 24,
// 				"max_overduedays_roao": 5,
// 				"num_of_asset_inventoried": 0,
// 				"gender": "F",
// 				"emergency_contact_address": "",
// 				"max_overduedays_for_active_agreement": 0,
// 				"max_overduedays_for_prev_eom": 0,
// 				"sisa_jumlah_angsuran": 0,
// 				"rrd_date": "2019-07-20 17:59:25",
// 				"number_of_agreement": 1,
// 				"work_since_year": "",
// 				"outstanding_principal": 1140686,
// 				"os_installmentdue": 0,
// 				"is_restructure": 0,
// 				"installment_amount_chassis_no": "",
// 				"is_similiar": 0
// 			}`,
// 			body: `{
// 				"id_number": "3276052009920006",
// 				"birth_date": "JAKARTA",
// 				"legal_address": "KRANJI",
// 				"legal_kelurahan": "KRANJI",
// 				"legal_kecamatan": "BEKASI BARAT",
// 				"legal_city": "BEKASI",
// 				"lagal_zipcode": "17135",
// 				"residence_address": "KRANJI",
// 				"residence_kelurahan": "KRANJI",
// 				"residence_kecamatan": "BEKASI BARAT",
// 				"residence_city": "BEKASI",
// 				"residence_zipcode": "",
// 				"company_address": "",
// 				"company_kelurahan": "",
// 				"company_kecamatan": "",
// 				"company_city": "",
// 				"company_zipcode": "",
// 				"personal_npwp": "",
// 				"education": "",
// 				"marital_status": "",
// 				"num_of_dependence": 0,
// 				"home_status": "",
// 				"profession_id": "",
// 				"job_type_id": "",
// 				"job_pos": "",
// 				"monthly_fixed_income": 10000000,
// 				"spouse_income": 0,
// 				"monthly_variable_income": 0,
// 				"total_installment": 200000,
// 				"total_installment_nap": 0,
// 				"customer_id": "40600051394",
// 				"bad_type": "",
// 				"max_overduedays": 24,
// 				"max_overduedays_roao": 5,
// 				"num_of_asset_inventoried": 0,
// 				"gender": "F",
// 				"emergency_contact_address": "",
// 				"max_overduedays_for_active_agreement": 0,
// 				"max_overduedays_for_prev_eom": 0,
// 				"sisa_jumlah_angsuran": 0,
// 				"rrd_date": "2019-07-20 17:59:25",
// 				"number_of_agreement": 1,
// 				"work_since_year": "",
// 				"outstanding_principal": 1140686,
// 				"os_installmentdue": 0,
// 				"is_restructure": 0,
// 				"installment_amount_chassis_no": "",
// 				"is_similiar": 0
// 			}`,
// 			code:  200,
// 			label: "TEST_DUPCHECK_RO_AO",
// 		},
// 		{
// 			payload: string(param_konsumen),
// 			expected: `{
// 				"id_number": "",
// 				"birth_date": "",
// 				"legal_address": "",
// 				"legal_kelurahan": "",
// 				"legal_kecamatan": "",
// 				"legal_city": "",
// 				"lagal_zipcode": "",
// 				"residence_address": "",
// 				"residence_kelurahan": "",
// 				"residence_kecamatan": "",
// 				"residence_city": "",
// 				"residence_zipcode": "",
// 				"company_address": "",
// 				"company_kelurahan": "",
// 				"company_kecamatan": "",
// 				"company_city": "",
// 				"company_zipcode": "",
// 				"personal_npwp": "",
// 				"education": "",
// 				"marital_status": "",
// 				"num_of_dependence": 0,
// 				"home_status": "",
// 				"profession_id": "",
// 				"job_type_id": "",
// 				"job_pos": "",
// 				"monthly_fixed_income": 0,
// 				"spouse_income": 0,
// 				"monthly_variable_income": 0,
// 				"total_installment": 0,
// 				"total_installment_nap": 0,
// 				"customer_id": "",
// 				"bad_type": "",
// 				"max_overduedays": 0,
// 				"max_overduedays_roao": 0,
// 				"num_of_asset_inventoried": 0,
// 				"gender": "F",
// 				"emergency_contact_address": "",
// 				"max_overduedays_for_active_agreement": 0,
// 				"max_overduedays_for_prev_eom": 0,
// 				"sisa_jumlah_angsuran": 0,
// 				"rrd_date": "",
// 				"number_of_agreement": 0,
// 				"work_since_year": "",
// 				"outstanding_principal": 0,
// 				"os_installmentdue": 0,
// 				"is_restructure": 0,
// 				"installment_amount_chassis_no": "",
// 				"is_similiar": 0
// 			}`,
// 			body: `{
// 				"id_number": "",
// 				"birth_date": "",
// 				"legal_address": "",
// 				"legal_kelurahan": "",
// 				"legal_kecamatan": "",
// 				"legal_city": "",
// 				"lagal_zipcode": "",
// 				"residence_address": "",
// 				"residence_kelurahan": "",
// 				"residence_kecamatan": "",
// 				"residence_city": "",
// 				"residence_zipcode": "",
// 				"company_address": "",
// 				"company_kelurahan": "",
// 				"company_kecamatan": "",
// 				"company_city": "",
// 				"company_zipcode": "",
// 				"personal_npwp": "",
// 				"education": "",
// 				"marital_status": "",
// 				"num_of_dependence": 0,
// 				"home_status": "",
// 				"profession_id": "",
// 				"job_type_id": "",
// 				"job_pos": "",
// 				"monthly_fixed_income": 0,
// 				"spouse_income": 0,
// 				"monthly_variable_income": 0,
// 				"total_installment": 0,
// 				"total_installment_nap": 0,
// 				"customer_id": "",
// 				"bad_type": "",
// 				"max_overduedays": 0,
// 				"max_overduedays_roao": 0,
// 				"num_of_asset_inventoried": 0,
// 				"gender": "F",
// 				"emergency_contact_address": "",
// 				"max_overduedays_for_active_agreement": 0,
// 				"max_overduedays_for_prev_eom": 0,
// 				"sisa_jumlah_angsuran": 0,
// 				"rrd_date": "",
// 				"number_of_agreement": 0,
// 				"work_since_year": "",
// 				"outstanding_principal": 0,
// 				"os_installmentdue": 0,
// 				"is_restructure": 0,
// 				"installment_amount_chassis_no": "",
// 				"is_similiar": 0
// 			}`,
// 			code:  200,
// 			label: "TEST_DUPCHECK_NOT_FOUND",
// 		},
// 	}

// 	for _, test := range testcase {

// 		rst := resty.New()
// 		httpmock.ActivateNonDefault(rst.GetClient())
// 		defer httpmock.DeactivateAndReset()

// 		httpmock.RegisterResponder("POST", os.Getenv("DUPCHECK_URL"), httpmock.NewStringResponder(test.code, test.body))

// 		resp, _ := rst.R().SetBody(test.payload).Post(os.Getenv("DUPCHECK_URL"))

// 		mockHttpClient := new(httpclient.MockHttpClient)
// 		mockHttpClient.On("CallWebSocket", os.Getenv("DUPCHECK_URL"), test.body, map[string]string{}, 10).Return(resp, test.err)

// 		require.Equal(t, test.expected, string(resp.Body()))
// 	}
// }

// func TestKreditmu(t *testing.T) {

// 	paramKreditmu, _ := json.Marshal(map[string]string{
// 		"birth_date":          "1983-08-17",
// 		"id_number":           "3276052009920006",
// 		"legal_name":          "AGUS",
// 		"surgate_mother_name": "ENEH",
// 	})

// 	testcase := []struct {
// 		payload  string
// 		expected interface{}
// 		body     string
// 		code     int
// 		err      error
// 		label    string
// 	}{
// 		{
// 			payload: string(paramKreditmu),
// 			expected: `{
// 				"code": "OK",
// 				"message": "operasi berhasil dieksekusi.",
// 				"data": {
// 					"customer_status": "VERIFY",
// 					"id": 5773010,
// 					"is_allowed_upgrade_limit": false,
// 					"limit": 17900000,
// 					"limit_available": [
// 						{
// 							"current_limit": 500000,
// 							"gross_limit": 500000,
// 							"tenor": [
// 								1
// 							],
// 							"tenor_limit": 1
// 						},
// 						{
// 							"current_limit": 2900000,
// 							"gross_limit": 5000000,
// 							"tenor": [
// 								2,
// 								3
// 							],
// 							"tenor_limit": 3
// 						},
// 						{
// 							"current_limit": 7900000,
// 							"gross_limit": 10000000,
// 							"tenor": [
// 								4,
// 								6,
// 								5,
// 								7
// 							],
// 							"tenor_limit": 6
// 						},
// 						{
// 							"current_limit": 17900000,
// 							"gross_limit": 20000000,
// 							"tenor": [
// 								36,
// 								8,
// 								11,
// 								12,
// 								9,
// 								10,
// 								14,
// 								24,
// 								13
// 							],
// 							"tenor_limit": 12
// 						}
// 					],
// 					"limit_status": "ACTIVE"
// 				},
// 				"errors": null,
// 				"request_id": "90d0fbfb-9bb0-4dd1-9838-c1861cbe06f8",
// 				"timestamp": "2022-12-28 14:29:07"
// 			}`,
// 			body: `{
// 				"code": "OK",
// 				"message": "operasi berhasil dieksekusi.",
// 				"data": {
// 					"customer_status": "VERIFY",
// 					"id": 5773010,
// 					"is_allowed_upgrade_limit": false,
// 					"limit": 17900000,
// 					"limit_available": [
// 						{
// 							"current_limit": 500000,
// 							"gross_limit": 500000,
// 							"tenor": [
// 								1
// 							],
// 							"tenor_limit": 1
// 						},
// 						{
// 							"current_limit": 2900000,
// 							"gross_limit": 5000000,
// 							"tenor": [
// 								2,
// 								3
// 							],
// 							"tenor_limit": 3
// 						},
// 						{
// 							"current_limit": 7900000,
// 							"gross_limit": 10000000,
// 							"tenor": [
// 								4,
// 								6,
// 								5,
// 								7
// 							],
// 							"tenor_limit": 6
// 						},
// 						{
// 							"current_limit": 17900000,
// 							"gross_limit": 20000000,
// 							"tenor": [
// 								36,
// 								8,
// 								11,
// 								12,
// 								9,
// 								10,
// 								14,
// 								24,
// 								13
// 							],
// 							"tenor_limit": 12
// 						}
// 					],
// 					"limit_status": "ACTIVE"
// 				},
// 				"errors": null,
// 				"request_id": "90d0fbfb-9bb0-4dd1-9838-c1861cbe06f8",
// 				"timestamp": "2022-12-28 14:29:07"
// 			}`,
// 			code:  200,
// 			label: "TEST_KREDITMU_FOUND",
// 		},
// 		{
// 			payload: string(paramKreditmu),
// 			expected: `{
// 				"code": "CORE-API-005",
// 				"message": "data yang anda minta tidak ditemukan.",
// 				"data": null,
// 				"errors": null,
// 				"request_id": "b24fa8f6-983b-401a-8226-5b5d35cbab92",
// 				"timestamp": "2023-01-16 15:46:09"
// 			  }
// 			  `,
// 			body: `{
// 				"code": "CORE-API-005",
// 				"message": "data yang anda minta tidak ditemukan.",
// 				"data": null,
// 				"errors": null,
// 				"request_id": "b24fa8f6-983b-401a-8226-5b5d35cbab92",
// 				"timestamp": "2023-01-16 15:46:09"
// 			  }
// 			  `,
// 			code:  400,
// 			label: "TEST_NOT_FOUND",
// 		},
// 	}

// 	for _, test := range testcase {

// 		rst := resty.New()
// 		httpmock.ActivateNonDefault(rst.GetClient())
// 		defer httpmock.DeactivateAndReset()

// 		httpmock.RegisterResponder("POST", os.Getenv("KREDITMU_URL"), httpmock.NewStringResponder(test.code, test.body))

// 		resp, _ := rst.R().SetBody(test.payload).Post(os.Getenv("KREDITMU_URL"))

// 		mockHttpClient := new(httpclient.MockHttpClient)
// 		mockHttpClient.On("CallWebSocket", os.Getenv("KREDITMU_URL"), test.body, map[string]string{}, 10).Return(resp, test.err)

// 		require.Equal(t, test.expected, string(resp.Body()))
// 	}
// }

// func TestPefindo(t *testing.T) {

// 	paramPefindo, _ := json.Marshal(map[string]string{
// 		"ClientKey":         os.Getenv("CLIENTKEY_CORE_PBK"),
// 		"IDMember":          constants.USER_PBK_KMB_FILTEERING,
// 		"user":              constants.USER_PBK_KMB_FILTEERING,
// 		"IDNumber":          "3276052009920006",
// 		"ProspectID":        "TESTNEWROAO000000073",
// 		"LegalName":         "AGUS",
// 		"BirthDate":         "1983-08-17",
// 		"SurgateMotherName": "ENEH",
// 		"Gender":            "M",
// 		"MaritalStatus":     "S",
// 	})

// 	testcase := []struct {
// 		payload  string
// 		expected interface{}
// 		body     string
// 		code     int
// 		err      error
// 		label    string
// 	}{
// 		{
// 			payload: string(paramPefindo),
// 			expected: `{
// 				"code": "200",
// 				"status": "SUCCESS",
// 				"result": {
// 					"search_id": "kp_61e69179e9ac2",
// 					"pefindo_id": "1042512812",
// 					"score": "VERY HIGH RISK",
// 					"max_overdue": "31",
// 					"max_overdue_last12months": "37",
// 					"angsuran_aktif_pbk": "1733927.25",
// 					"angsuran_aktif_pbk_konsumen": "1733927.25",
// 					"wo_contract": true,
// 					"wo_ada_agunan": false,
// 					"detail_report": "http://10.0.0.161/los-symlink/pefindo/pdf/dummy.pdf",
// 					"total_baki_debet_non_agunan": 4000000
// 				},
// 				"konsumen": {
// 					"search_id": "kp_61e69179e9ac2",
// 					"pefindo_id": "1042512812",
// 					"score": "VERY HIGH RISK",
// 					"max_overdue": "31",
// 					"max_overdue_last12months": "37",
// 					"detail_report": "http://10.0.0.161/los-symlink/pefindo/pdf/dummy.pdf",
// 					"baki_debet_non_agunan": 4000000
// 				},
// 				"pasangan": null,
// 				"timestamp": "1642500540"
// 			}`,
// 			body: `{
// 				"code": "200",
// 				"status": "SUCCESS",
// 				"result": {
// 					"search_id": "kp_61e69179e9ac2",
// 					"pefindo_id": "1042512812",
// 					"score": "VERY HIGH RISK",
// 					"max_overdue": "31",
// 					"max_overdue_last12months": "37",
// 					"angsuran_aktif_pbk": "1733927.25",
// 					"angsuran_aktif_pbk_konsumen": "1733927.25",
// 					"wo_contract": true,
// 					"wo_ada_agunan": false,
// 					"detail_report": "http://10.0.0.161/los-symlink/pefindo/pdf/dummy.pdf",
// 					"total_baki_debet_non_agunan": 4000000
// 				},
// 				"konsumen": {
// 					"search_id": "kp_61e69179e9ac2",
// 					"pefindo_id": "1042512812",
// 					"score": "VERY HIGH RISK",
// 					"max_overdue": "31",
// 					"max_overdue_last12months": "37",
// 					"detail_report": "http://10.0.0.161/los-symlink/pefindo/pdf/dummy.pdf",
// 					"baki_debet_non_agunan": 4000000
// 				},
// 				"pasangan": null,
// 				"timestamp": "1642500540"
// 			}`,
// 			code:  200,
// 			label: "TEST_KREDITMU_FOUND",
// 		},
// 		{
// 			payload:  string(paramPefindo),
// 			expected: `{"code":"201","result":"Pefindo Data Not Found"}`,
// 			body:     `{"code":"201","result":"Pefindo Data Not Found"}`,
// 			code:     200,
// 			label:    "TEST_NOT_FOUND",
// 		},
// 	}

// 	for _, test := range testcase {

// 		rst := resty.New()
// 		httpmock.ActivateNonDefault(rst.GetClient())
// 		defer httpmock.DeactivateAndReset()

// 		httpmock.RegisterResponder("POST", os.Getenv("PBK_URL"), httpmock.NewStringResponder(test.code, test.body))

// 		resp, _ := rst.R().SetBody(test.payload).Post(os.Getenv("PBK_URL"))

// 		mockHttpClient := new(httpclient.MockHttpClient)
// 		mockHttpClient.On("CallWebSocket", os.Getenv("PBK_URL"), test.body, map[string]string{}, 10).Return(resp, test.err)

// 		require.Equal(t, test.expected, string(resp.Body()))
// 	}
// }
