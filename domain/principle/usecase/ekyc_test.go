package usecase

import (
	"context"
	"errors"
	"fmt"
	"los-kmb-api/domain/principle/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDukcapil(t *testing.T) {

	type expectedResult struct {
		data response.Ekyc
		err  error
	}

	type respDukcapil struct {
		code     int
		response string
	}

	type respGetMappingDukcapil struct {
		data entity.MappingResultDukcapil
		err  error
	}

	testcases := []struct {
		label                          string
		request                        request.PrinciplePemohon
		expected                       expectedResult
		respAppConfig                  entity.AppConfig
		respGetMappingDukcapil         respGetMappingDukcapil
		MappingResultDukcapilVD        entity.MappingResultDukcapilVD
		errMappingResultDukcapilVD     error
		errGetConfig                   error
		respDukcapilVD, respDukcapilFR respDukcapil
		reqMetricsEkyc                 request.MetricsEkyc
	}{
		{
			label: "Test PASS dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 100,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 100,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respDukcapilFR: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"rule_code": "6018",
						"reason": "EKYC Sesuai",
						"threshold": "5.0",
						"ref_id": "7301010xxxxxxxxx",
						"matchScore": "8.331"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "PASS",
					RuleCode: "1621",
				},
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result:      "PASS",
					Code:        "1621",
					Reason:      "Ekyc Valid",
					Source:      "DCP",
					Similiarity: "8.331",
					Info:        "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":100,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":100,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":{\"transaction_id\":\"EFM01108902308030001\",\"rule_code\":\"6018\",\"reason\":\"EKYC Sesuai\",\"threshold\":\"5.0\",\"ref_id\":\"7301010xxxxxxxxx\",\"matchScore\":\"8.331\"},\"fr_service\":\"dukcapil\",\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test PASS FR izidata",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"izidata","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 100,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 100,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respDukcapilFR: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"rule_code": "6058",
						"reason": "EKYC Sesuai",
						"threshold": "5.0",
						"ref_id": "7301010xxxxxxxxx",
						"matchScore": "8.331"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "PASS",
					RuleCode: "1621",
				},
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result:      "PASS",
					Code:        "1621",
					Reason:      "Ekyc Valid",
					Source:      "DCP",
					Similiarity: "8.331",
					Info:        "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":100,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":100,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":{\"transaction_id\":\"EFM01108902308030001\",\"rule_code\":\"6058\",\"reason\":\"EKYC Sesuai\",\"threshold\":\"5.0\",\"ref_id\":\"7301010xxxxxxxxx\",\"matchScore\":\"8.331\"},\"fr_service\":\"izidata\",\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test RTO - RTO",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 504,
			},
			respDukcapilFR: respDukcapil{
				code: 504,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "CONTINGENCY",
					RuleCode: "1626",
				},
			},
			expected: expectedResult{
				err: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
				data: response.Ekyc{
					Result: "CONTINGENCY",
					Code:   "1626",
					Reason: "CONTINGENCY",
					Source: "DCP",
					Info:   "{\"vd\":null,\"vd_service\":\"\",\"vd_error\":\"Request Timed Out\",\"fr\":null,\"fr_service\":\"\",\"fr_error\":\"Request Timed Out\",\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test Not Check - Not Check",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 500,
				response: `{
					"messages": "bypass Dukcapil VD",
				}`,
			},
			respDukcapilFR: respDukcapil{
				code: 500,
				response: `{
					"messages": "bypass Dukcapil VD",
				}`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "CONTINGENCY",
				},
			},
			expected: expectedResult{
				err: fmt.Errorf("%s - Dukcapil", constant.TYPE_CONTINGENCY),
				data: response.Ekyc{
					Result: "CONTINGENCY",
					Code:   "",
					Reason: "CONTINGENCY",
					Source: "DCP",
					Info:   "{\"vd\":null,\"vd_service\":\"izidata\",\"vd_error\":\"\",\"fr\":null,\"fr_service\":\"dukcapil\",\"fr_error\":\"\",\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test Error Mapping Dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 500,
				response: `{
					"messages": "bypass Dukcapil VD",
				}`,
			},
			respDukcapilFR: respDukcapil{
				code: 500,
				response: `{
					"messages": "bypass Dukcapil VD",
				}`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				err: fmt.Errorf("error"),
			},
			expected: expectedResult{
				err: fmt.Errorf("error"),
			},
		},
		{
			label: "Test VD REJECT BYPASS RO PRIME",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "RO",
				CustomerSegment: "PRIME",
				CBFound:         true,
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.EKYC_BYPASS,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "PASS",
					RuleCode: "1621",
				},
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 80,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 45,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respDukcapilFR: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"rule_code": "6018",
						"reason": "EKYC Sesuai",
						"threshold": "5.0",
						"ref_id": "7301010xxxxxxxxx",
						"matchScore": "8.331"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result:      "PASS",
					Code:        "1621",
					Reason:      "Ekyc Valid",
					Source:      "DCP",
					Info:        "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":80,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":45,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":{\"transaction_id\":\"EFM01108902308030001\",\"rule_code\":\"6018\",\"reason\":\"EKYC Sesuai\",\"threshold\":\"5.0\",\"ref_id\":\"7301010xxxxxxxxx\",\"matchScore\":\"8.331\"},\"fr_service\":\"dukcapil\",\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
					Similiarity: "8.331",
				},
			},
		},
		{
			label: "Test VD REJECT nik dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 50,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 60,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Tidak Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "RO",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1612",
					Reason: "Ekyc Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":50,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":60,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Tidak Sesuai\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT alamat dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":90}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 100,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 60,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "RO",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1612",
					Reason: "Ekyc Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":100,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":60,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT nik izidata",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 50,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 60,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Tidak Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "RO",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1652",
					Reason: "Izi Data Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":50,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":60,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Tidak Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT nama_lgkp dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 50,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 60,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "RO",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1612",
					Reason: "Ekyc Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":50,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":60,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT nama_lgkp izidata",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 50,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 60,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "RO",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1652",
					Reason: "Izi Data Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":50,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":60,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT meninggal",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": false,
						"reason": "Customer Meninggal Dunia"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1613",
					Reason: "Ekyc Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":false,\"reason\":\"Customer Meninggal Dunia\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT Data Ganda",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": false,
						"reason": "Data Ganda"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1614",
					Reason: "Ekyc Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":false,\"reason\":\"Data Ganda\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT Data Inactive",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": false,
						"reason": "Data Inactive"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1615",
					Reason: "Ekyc Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":false,\"reason\":\"Data Inactive\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT Data Invalid izidata fix: reason izi data",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": false,
						"reason": "Data Invalid"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1652",
					Reason: "Izi Data Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":false,\"reason\":\"Data Invalid\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD REJECT Data Not Found dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"dukcapil","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": false,
						"reason": "Data Not Found"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result: "REJECT",
					Code:   "1616",
					Reason: "Ekyc Invalid",
					Source: "DCP",
					Info:   "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":false,\"reason\":\"Data Not Found\"},\"vd_service\":\"dukcapil\",\"vd_error\":null,\"fr\":null,\"fr_service\":null,\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test FR REJECT nik dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 100,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 100,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respDukcapilFR: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"rule_code": "6020",
						"reason": "EKYC Tidak Sesuai",
						"threshold": "5.0",
						"ref_id": "7301010xxxxxxxxx",
						"matchScore": "8.331"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "REJECT",
					RuleCode: "1623",
				},
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_PASS,
				Decision: constant.DECISION_PASS,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result:      "REJECT",
					Code:        "1623",
					Reason:      "Ekyc Invalid",
					Source:      "DCP",
					Similiarity: "8.331",
					Info:        "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":100,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":100,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":{\"transaction_id\":\"EFM01108902308030001\",\"rule_code\":\"6020\",\"reason\":\"EKYC Tidak Sesuai\",\"threshold\":\"5.0\",\"ref_id\":\"7301010xxxxxxxxx\",\"matchScore\":\"8.331\"},\"fr_service\":\"dukcapil\",\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test FR REJECT nik izidata",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"izidata","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 100,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 100,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respDukcapilFR: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"rule_code": "6060",
						"reason": "EKYC Tidak Sesuai",
						"threshold": "5.0",
						"ref_id": "7301010xxxxxxxxx",
						"matchScore": "8.331"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "REJECT",
					RuleCode: "1623",
				},
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_PASS,
				Decision: constant.DECISION_PASS,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result:      "REJECT",
					Code:        "1623",
					Reason:      "Ekyc Invalid",
					Source:      "DCP",
					Similiarity: "8.331",
					Info:        "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":100,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":100,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":{\"transaction_id\":\"EFM01108902308030001\",\"rule_code\":\"6060\",\"reason\":\"EKYC Tidak Sesuai\",\"threshold\":\"5.0\",\"ref_id\":\"7301010xxxxxxxxx\",\"matchScore\":\"8.331\"},\"fr_service\":\"izidata\",\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test FR REJECT foto dukcapil",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 100,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 100,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respDukcapilFR: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"rule_code": "6019",
						"reason": "EKYC Tidak Sesuai",
						"threshold": "5.0",
						"ref_id": "7301010xxxxxxxxx",
						"matchScore": "8.331"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "REJECT",
					RuleCode: "1622",
				},
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_PASS,
				Decision: constant.DECISION_PASS,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result:      "REJECT",
					Code:        "1622",
					Reason:      "Ekyc Invalid",
					Source:      "DCP",
					Similiarity: "8.331",
					Info:        "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":100,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":100,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":{\"transaction_id\":\"EFM01108902308030001\",\"rule_code\":\"6019\",\"reason\":\"EKYC Tidak Sesuai\",\"threshold\":\"5.0\",\"ref_id\":\"7301010xxxxxxxxx\",\"matchScore\":\"8.331\"},\"fr_service\":\"dukcapil\",\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test FR REJECT foto izidata",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"izidata","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": true,
						"no_kk": "Sesuai",
						"nama_lgkp": 100,
						"tmpt_lhr": 100,
						"tgl_lhr": "Sesuai",
						"prop_name": "Sesuai",
						"kab_name": "Sesuai",
						"kec_name": "Sesuai",
						"kel_name": "Sesuai",
						"no_rt": "Sesuai",
						"no_rw": "Sesuai",
						"alamat": 100,
						"nama_lgkp_ibu": 100,
						"status_kawin": "Sesuai",
						"jenis_pkrjn": "Sesuai",
						"jenis_klmin": "Sesuai",
						"no_prop": "Sesuai",
						"no_kab": "Sesuai",
						"no_kec": "Sesuai",
						"no_kel": "Sesuai",
						"nik": "Sesuai"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respDukcapilFR: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"rule_code": "6059",
						"reason": "EKYC Tidak Sesuai",
						"threshold": "5.0",
						"ref_id": "7301010xxxxxxxxx",
						"matchScore": "8.331"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			respGetMappingDukcapil: respGetMappingDukcapil{
				data: entity.MappingResultDukcapil{
					Decision: "REJECT",
					RuleCode: "1622",
				},
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_PASS,
				Decision: constant.DECISION_PASS,
			},
			expected: expectedResult{
				data: response.Ekyc{
					Result:      "REJECT",
					Code:        "1622",
					Reason:      "Ekyc Invalid",
					Source:      "DCP",
					Similiarity: "8.331",
					Info:        "{\"vd\":{\"transaction_id\":\"EFM01108902308030001\",\"threshold\":\"0\",\"ref_id\":\"1000338d-208e-4e06-80f0-cbe8c1358a20\",\"is_valid\":true,\"no_kk\":\"Sesuai\",\"nama_lgkp\":100,\"tmpt_lhr\":100,\"tgl_lhr\":\"Sesuai\",\"prop_name\":\"Sesuai\",\"kab_name\":\"Sesuai\",\"kec_name\":\"Sesuai\",\"kel_name\":\"Sesuai\",\"no_rt\":\"Sesuai\",\"no_rw\":\"Sesuai\",\"alamat\":100,\"nama_lgkp_ibu\":100,\"status_kawin\":\"Sesuai\",\"jenis_pkrjn\":\"Sesuai\",\"jenis_klmin\":\"Sesuai\",\"no_prop\":\"Sesuai\",\"no_kab\":\"Sesuai\",\"no_kec\":\"Sesuai\",\"no_kel\":\"Sesuai\",\"nik\":\"Sesuai\"},\"vd_service\":\"izidata\",\"vd_error\":null,\"fr\":{\"transaction_id\":\"EFM01108902308030001\",\"rule_code\":\"6059\",\"reason\":\"EKYC Tidak Sesuai\",\"threshold\":\"5.0\",\"ref_id\":\"7301010xxxxxxxxx\",\"matchScore\":\"8.331\"},\"fr_service\":\"izidata\",\"fr_error\":null,\"asliri\":null,\"ktp\":null}",
				},
			},
		},
		{
			label: "Test VD MAPPING ERROR data not found izidata",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": false,
						"reason": "Data Not Found"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			errMappingResultDukcapilVD: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Verify Dukcapil Error"),
			expected: expectedResult{
				err: errors.New(constant.ERROR_UPSTREAM + " - Get Mapping Verify Dukcapil Error"),
			},
		},
		{
			label: "Test Get config ERROR",
			request: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             "123",
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            "Test Legal Name",
				FullName:             "Test Full Name",
				BirthDate:            "1999-09-08",
				BirthPlace:           "JAKARTA",
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			respAppConfig: entity.AppConfig{
				Value: `{"data":{"verify_data":{"service_on":"izidata","izidata":{"nama_lengkap":80},"dukcapil":{"nama_lengkap":80,"alamat":0}},"face_recognition":{"service_on":"dukcapil","izidata":{"threshold":80},"dukcapil":{"threshold":5}}}}`,
			},
			respDukcapilVD: respDukcapil{
				code: 200,
				response: `{
					"data": {
						"transaction_id": "EFM01108902308030001",
						"threshold": "0",
						"ref_id": "1000338d-208e-4e06-80f0-cbe8c1358a20",
						"is_valid": false,
						"reason": "Data Not Found"
					},
					"errors": {},
					"messages": "string",
					"request_id": "string",
					"server_time": "string"
				  }`,
			},
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus: "NEW",
				CBFound:        true,
			},
			MappingResultDukcapilVD: entity.MappingResultDukcapilVD{
				ResultVD: constant.DECISION_REJECT,
				Decision: constant.DECISION_REJECT,
			},
			errGetConfig: errors.New(constant.ERROR_UPSTREAM + " - Get Dukcapil Config Error"),
			expected: expectedResult{
				err: errors.New(constant.ERROR_UPSTREAM + " - Get Dukcapil Config Error"),
			},
		},
	}

	for _, test := range testcases {

		ctx := context.Background()

		mockRepository := new(mocks.Repository)
		mockHttpClient := new(httpclient.MockHttpClient)

		mockRepository.On("GetConfig", "ekyc", "KMB-OFF", "threshold_ekyc").Return(test.respAppConfig, test.errGetConfig)
		mockRepository.On("GetMappingDukcapil", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.respGetMappingDukcapil.data, test.respGetMappingDukcapil.err)
		mockRepository.On("GetMappingDukcapilVD", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(test.MappingResultDukcapilVD, test.errMappingResultDukcapilVD)

		//httpclient Dukcapil VD
		rst := resty.New()
		httpmock.ActivateNonDefault(rst.GetClient())
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("DUKCAPIL_VD_URL"), httpmock.NewStringResponder(test.respDukcapilVD.code, test.respDukcapilVD.response))
		resp, _ := rst.R().Post(os.Getenv("DUKCAPIL_VD_URL"))
		mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("DUKCAPIL_VD_URL"), mock.Anything, map[string]string{}, constant.METHOD_POST, true, 2, mock.Anything, test.request.ProspectID, mock.Anything).Return(resp, nil).Once()

		//httpclient Dukcapil FR
		rst = resty.New()
		httpmock.ActivateNonDefault(rst.GetClient())
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("DUKCAPIL_FR_URL"), httpmock.NewStringResponder(test.respDukcapilFR.code, test.respDukcapilFR.response))
		resp, _ = rst.R().Post(os.Getenv("DUKCAPIL_FR_URL"))
		mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("DUKCAPIL_FR_URL"), mock.Anything, map[string]string{}, constant.METHOD_POST, true, 2, mock.Anything, test.request.ProspectID, mock.Anything).Return(resp, nil).Once()

		usecase := NewUsecase(mockRepository, mockHttpClient, nil)

		data, err := usecase.Dukcapil(ctx, test.request, test.reqMetricsEkyc, "token")

		fmt.Println(test.label)

		require.Equal(t, test.expected.data, data)
		require.Equal(t, test.expected.err, err)

	}
}

func TestAsliri(t *testing.T) {

	os.Setenv("DEFAULT_TIMEOUT_30S", "30")

	var (
		prospectID        string = "SAL-123"
		idNumber          string = "32030143096XXXX6"
		legalName         string = "SOE***E"
		birthDate         string = "1966-09-03"
		birthPlace        string = "JAKARTA"
		responseAppConfig        = entity.AppConfig{
			GroupName: "",
			Lob:       "",
			Key:       "",
			Value: `{
				"data": {
					"asliri_service_active": true,
					"asliri_threshold_selfie_photo": 70,
					"asliri_threshold_name": 80,
					"asliri_threshold_pdob": 80
				}
			}`,
			IsActive:  0,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		}
	)

	testcase := []struct {
		payload      request.PrinciplePemohon
		expected     response.Ekyc
		body         string
		code         int
		err          error
		errGetConfig error
		errExpected  error
		label        string
	}{
		{
			payload: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             idNumber,
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            legalName,
				FullName:             "Test Full Name",
				BirthDate:            birthDate,
				BirthPlace:           birthPlace,
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			body: `{
				"message": "LOS - ASLIRI",
				"errors": null,
				"data": {
				  "name": 100,
				  "pdob": 100,
				  "selfie_photo": 85,
				  "not_found": false,
				  "ref_id": "aW50ZXJuYWw=-1675666687732"
				},
				"server_time": "2022-11-30T16:48:45+07:00"
			  }`,
			code: 200,
			expected: response.Ekyc{
				Result: constant.DECISION_PASS, Code: "1525", Source: "ARI", Reason: "Ekyc Valid", Info: `{"vd":null,"vd_service":null,"vd_error":null,"fr":null,"fr_service":null,"fr_error":null,"asliri":{"name":100,"pdob":100,"selfie_photo":85,"not_found":false,"ref_id":"aW50ZXJuYWw=-1675666687732"},"ktp":null}`,
			},
			label: "TEST_ASLIRI_RESPONSE_OK_PASS",
		},
		{
			payload: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             idNumber,
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            legalName,
				FullName:             "Test Full Name",
				BirthDate:            birthDate,
				BirthPlace:           birthPlace,
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			body: `{
				"message": "LOS - ASLIRI",
				"errors": null,
				"data": {
				  "name": 0,
				  "pdob": 0,
				  "selfie_photo": 0,
				  "not_found": true,
				  "ref_id": "aW50ZXJuYWw=-1675666687732"
				},
				"server_time": "2022-11-30T16:48:45+07:00"
			  }`,
			code: 200,
			expected: response.Ekyc{
				Result: constant.DECISION_REJECT, Code: "1528", Source: "ARI", Reason: "Ekyc Invalid", Info: `{"vd":null,"vd_service":null,"vd_error":null,"fr":null,"fr_service":null,"fr_error":null,"asliri":{"name":0,"pdob":0,"selfie_photo":0,"not_found":true,"ref_id":"aW50ZXJuYWw=-1675666687732"},"ktp":null}`,
			},
			label: "TEST_ASLIRI_RESPONSE_REJECT_NOT_FOUND",
		},
		{
			payload: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             idNumber,
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            legalName,
				FullName:             "Test Full Name",
				BirthDate:            birthDate,
				BirthPlace:           birthPlace,
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			body: `{
				"message": "LOS - ASLIRI",
				"errors": null,
				"data": {
				  "name": 0,
				  "pdob": 0,
				  "selfie_photo": 0,
				  "not_found": false,
				  "ref_id": "aW50ZXJuYWw=-1675666687732"
				},
				"server_time": "2022-11-30T16:48:45+07:00"
			  }`,
			code: 200,
			expected: response.Ekyc{
				Result: constant.DECISION_REJECT, Code: "1527", Source: "ARI", Reason: "Ekyc Invalid", Info: `{"vd":null,"vd_service":null,"vd_error":null,"fr":null,"fr_service":null,"fr_error":null,"asliri":{"name":0,"pdob":0,"selfie_photo":0,"not_found":false,"ref_id":"aW50ZXJuYWw=-1675666687732"},"ktp":null}`,
			},
			label: "TEST_ASLIRI_RESPONSE_REJECT_NAME_OR_PDOB",
		},
		{
			payload: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             idNumber,
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            legalName,
				FullName:             "Test Full Name",
				BirthDate:            birthDate,
				BirthPlace:           birthPlace,
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			body: `{
				"message": "LOS - ASLIRI",
				"errors": null,
				"data": {
				  "name": 100,
				  "pdob": 100,
				  "selfie_photo": 0,
				  "not_found": false,
				  "ref_id": "aW50ZXJuYWw=-1675666687732"
				},
				"server_time": "2022-11-30T16:48:45+07:00"
			  }`,
			code: 200,
			expected: response.Ekyc{
				Result: constant.DECISION_REJECT, Code: "1526", Source: "ARI", Reason: "Ekyc Invalid", Info: `{"vd":null,"vd_service":null,"vd_error":null,"fr":null,"fr_service":null,"fr_error":null,"asliri":{"name":100,"pdob":100,"selfie_photo":0,"not_found":false,"ref_id":"aW50ZXJuYWw=-1675666687732"},"ktp":null}`,
			},
			label: "TEST_ASLIRI_RESPONSE_REJECT_SELFIE",
		},
		{
			payload: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             idNumber,
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            legalName,
				FullName:             "Test Full Name",
				BirthDate:            birthDate,
				BirthPlace:           birthPlace,
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			err:         fmt.Errorf("err"),
			errExpected: fmt.Errorf("err"),
			label:       "TEST_ASLIRI_RESPONSE_ERROR",
		},
		{
			payload: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             idNumber,
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            legalName,
				FullName:             "Test Full Name",
				BirthDate:            birthDate,
				BirthPlace:           birthPlace,
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			code:  503,
			label: "TEST_ASLIRI_RESPONSE_ERROR_CODE",
		},
		{
			payload: request.PrinciplePemohon{
				ProspectID:           "SAL-123",
				IDNumber:             idNumber,
				SpouseIDNumber:       "1234",
				MobilePhone:          "08123743211",
				LegalName:            legalName,
				FullName:             "Test Full Name",
				BirthDate:            birthDate,
				BirthPlace:           birthPlace,
				SurgateMotherName:    "Test",
				Gender:               "M",
				LegalAddress:         "Test",
				LegalRT:              "001",
				LegalRW:              "001",
				LegalProvince:        "JAWA TIMUR",
				LegalCity:            "MALANG",
				LegalKecamatan:       "MALANG",
				LegalKelurahan:       "MALANG",
				LegalZipCode:         "66192",
				LegalPhoneArea:       "021",
				LegalPhone:           "12345",
				Education:            "SLTA",
				ProfessionID:         "WRST",
				JobType:              "TEST",
				JobPosition:          "Test",
				EmploymentSinceMonth: 2,
				EmploymentSinceYear:  2024,
				CompanyName:          "Test",
				EconomySectorID:      "001",
				IndustryTypeID:       "11",
				KtpPhoto:             "http://www.example.com",
				SelfiePhoto:          "http://www.example.com",
			},
			body: `{
				"message": "LOS - ASLIRI",
				"errors": null,
				"data": {
				  "name": 100,
				  "pdob": 100,
				  "selfie_photo": 0,
				  "not_found": false,
				  "ref_id": "aW50ZXJuYWw=-1675666687732"
				},
				"server_time": "2022-11-30T16:48:45+07:00"
			  }`,
			code:         200,
			errGetConfig: errors.New(constant.ERROR_UPSTREAM + " - Get ASLI RI Config Error"),
			errExpected:  errors.New(constant.ERROR_UPSTREAM + " - Get ASLI RI Config Error"),
			label:        "TEST_GET_CONFIG_ERROR",
		},
	}

	for _, test := range testcase {

		timeout, _ := strconv.Atoi(os.Getenv("DEFAULT_TIMEOUT_30S"))
		mockRepository := new(mocks.Repository)
		mockRepository.On("GetConfig", mock.Anything, mock.Anything, mock.Anything).Return(responseAppConfig, test.errGetConfig)

		rst := resty.New()
		ctx := context.Background()

		httpmock.ActivateNonDefault(rst.GetClient())
		defer httpmock.DeactivateAndReset()

		httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("ASLIRI_URL"), httpmock.NewStringResponder(test.code, test.body))
		resp, _ := rst.R().Post(os.Getenv("ASLIRI_URL"))

		mockHttpClient := new(httpclient.MockHttpClient)
		mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("ASLIRI_URL"), mock.Anything, map[string]string{}, constant.METHOD_POST, false, 0, timeout, prospectID, "token").Return(resp, test.err).Once()

		service := NewUsecase(mockRepository, mockHttpClient, nil)

		result, err := service.Asliri(ctx, test.payload, "token")

		fmt.Println(test.label)

		require.Equal(t, test.errExpected, err)
		require.Equal(t, test.expected.Result, result.Result)
		require.Equal(t, test.expected.Code, result.Code)
		require.Equal(t, test.expected.Reason, result.Reason)
		require.Equal(t, test.expected.Info, result.Info)
	}

}

func TestKtp(t *testing.T) {
	os.Setenv("KTP_VALIDATOR_URL", "/")
	os.Setenv("DEFAULT_TIMEOUT_30S", "30")
	os.Setenv("NAMA_SAMA", "K,P")

	testcases := []struct {
		name             string
		req              request.PrinciplePemohon
		reqMetricsEkyc   request.MetricsEkyc
		codeKtpValidator int
		respKtpValidator string
		errKtpValidator  error
		data             response.Ekyc
		err              error
	}{
		{
			name: "test KTP pass",
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "NEW",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			req: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				BirthDate:  "1999-08-09",
				Gender:     "M",
				IDNumber:   "123456",
				BpkbName:   "O",
			},
			codeKtpValidator: 200,
			respKtpValidator: `{ "messages": "LOS - KTP Validator", "errors": null, "data": { "code": "2600", "result": "PASS", "reason": "eKYC Sesuai - No KTP Valid" }, "server_time": "2023-11-25T23:10:02+07:00", "request_id": "d5b16870-86b9-4ebc-a334-889ac4da7773" }`,
			data: response.Ekyc{
				Result: constant.DECISION_PASS,
				Code:   "2600",
				Reason: "eKYC Sesuai - No KTP Valid",
				Source: constant.KTP,
				Info:   `{"vd":null,"vd_service":null,"vd_error":null,"fr":null,"fr_service":null,"fr_error":null,"asliri":null,"ktp":{"code":"2600","result":"PASS","reason":"eKYC Sesuai - No KTP Valid"}}`,
			},
		},
		{
			name: "test KTP err api ktp",
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "NEW",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			req: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				BirthDate:  "1999-08-09",
				Gender:     "M",
				IDNumber:   "123456",
				BpkbName:   "O",
			},
			codeKtpValidator: 200,
			errKtpValidator:  errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call KTP Validator"),
			err:              errors.New(constant.ERROR_UPSTREAM_TIMEOUT + " - Call KTP Validator"),
		},
		{
			name: "test KTP err api ktp",
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "NEW",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			req: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				BirthDate:  "1999-08-09",
				Gender:     "M",
				IDNumber:   "123456",
				BpkbName:   "O",
			},
			codeKtpValidator: 500,
		},
		{
			name: "test KTP pass",
			reqMetricsEkyc: request.MetricsEkyc{
				CustomerStatus:  "NEW",
				CustomerSegment: "REGULAR",
				CBFound:         true,
			},
			req: request.PrinciplePemohon{
				ProspectID: "SAL-123",
				BirthDate:  "1999-08-09",
				Gender:     "M",
				IDNumber:   "123456",
				BpkbName:   "K",
			},
			codeKtpValidator: 200,
			respKtpValidator: `{ "messages": "LOS - KTP Validator", "errors": null, "data": { "code": "2602", "result": "REJECT", "reason": "eKYC Tidak Sesuai - Format KTP Tidak Valid" }, "server_time": "2023-11-25T23:23:41+07:00", "request_id": "44e68164-3173-4220-ac69-d14bc345b9de" }`,
			data: response.Ekyc{
				Result: constant.DECISION_PASS,
				Code:   "2600",
				Reason: "eKYC Sesuai - No KTP Valid",
				Source: constant.KTP,
				Info:   `{"vd":null,"vd_service":null,"vd_error":null,"fr":null,"fr_service":null,"fr_error":null,"asliri":null,"ktp":{"code":"2602","result":"REJECT","reason":"eKYC Tidak Sesuai - Format KTP Tidak Valid"}}`,
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			rst := resty.New()
			httpmock.ActivateNonDefault(rst.GetClient())
			defer httpmock.DeactivateAndReset()

			httpmock.RegisterResponder(constant.METHOD_POST, os.Getenv("KTP_VALIDATOR_URL"), httpmock.NewStringResponder(tc.codeKtpValidator, tc.respKtpValidator))
			resp, _ := rst.R().Post(os.Getenv("KTP_VALIDATOR_URL"))

			mockHttpClient.On("EngineAPI", ctx, constant.DILEN_KMB_LOG, os.Getenv("KTP_VALIDATOR_URL"), mock.Anything, map[string]string{}, constant.METHOD_POST, false, 0, 30, tc.req.ProspectID, "token").Return(resp, tc.errKtpValidator).Once()

			usecase := NewUsecase(mockRepository, mockHttpClient, nil)

			result, err := usecase.Ktp(ctx, tc.req, tc.reqMetricsEkyc, "token")

			require.Equal(t, tc.data, result)
			require.Equal(t, tc.err, err)
		})
	}

}
