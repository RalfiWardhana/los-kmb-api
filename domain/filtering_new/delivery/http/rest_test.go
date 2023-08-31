package http

import (
	"errors"
	"los-kmb-api/domain/filtering_new/interfaces/mocks"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	mocksJson "los-kmb-api/shared/common/json/mocks"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

type MockEchoContext struct {
	mock.Mock
}

func (m *MockEchoContext) Validate(obj interface{}) error {
	args := m.Called(obj)
	return args.Error(0)
}

func TestFiltering(t *testing.T) {
	os.Setenv("PLATFORM_LIBRARY_KEY", "PLATFORMS-APIToEncryptDecryptAPI")
	os.Setenv("NAMA_SAMA", "K,P")
	os.Setenv("NAMA_BEDA", "O,KK")

	testcases := []struct {
		name         string
		reqbody      string
		checkPpid    string
		errCheckPpid error
		errFiltering error
	}{
		{
			name: "test err bind",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": nil
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test customer spouse",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test spouse",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": null
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errCIDnumber",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw8",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": null
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errCLegalName",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": null
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errCMotherName",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1=",
				"bpkb_name": "K",
				"spouse": null
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errSIDnumber",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errSLegalName",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "M",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errSMotherName",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
				}
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errValidateCS",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": null
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errValidateCS",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errValidateGender",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "F",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			checkPpid: "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errValidatePpid",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			checkPpid: "EFM0TST0020230809013 - false",
		},
		{
			name: "test err errCheckPpid",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			errCheckPpid: errors.New("error"),
			checkPpid:    "EFM0TST0020230809013 - true",
		},
		{
			name: "test err errCheckPpid",
			reqbody: `{
				"prospect_id": "EFM0TST0020230809013",
				"branch_id": "426",
				"id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
				"legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
				"birth_date": "1971-04-15",
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "F",
					"spouse_surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g=="
				}
			}`,
			errFiltering: errors.New("error"),
			checkPpid:    "EFM0TST0020230809013 - true",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a request body

			// Create a new Echo instance
			e := echo.New()
			e.Validator = common.NewValidator()

			req := httptest.NewRequest(http.MethodPost, "/api/v3/kmb/produce/filtering", strings.NewReader(tc.reqbody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			mockMultiUsecase := new(mocks.MultiUsecase)
			mockUsecase := new(mocks.Usecase)
			mockRepository := new(mocks.Repository)
			mockJson := new(mocksJson.JSON)
			mockUsecase.On("FilteringProspectID", "EFM0TST0020230809013").Return(request.OrderIDCheck{ProspectID: tc.checkPpid}, tc.errCheckPpid).Once()
			mockMultiUsecase.On("Filtering", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.Filtering{}, tc.errFiltering).Once()
			mockJson.On("InternalServerErrorCustomV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockJson.On("ServerSideErrorV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockJson.On("BadRequestErrorValidationV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockJson.On("SuccessV2", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

			// Create an instance of the handler
			handler := &handlerKmbFiltering{
				multiusecase: mockMultiUsecase,
				usecase:      mockUsecase,
				repository:   mockRepository,
				Json:         mockJson,
			}

			// Call the handler
			err := handler.Filtering(c)
			if err != nil {
				t.Errorf("error '%s' was not expected, but got: ", err)
			}
		})
	}
}
