package eventhandlers

import (
	"context"
	"errors"
	"los-kmb-api/domain/filtering_new/interfaces/mocks"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/common"
	mocksJson "los-kmb-api/shared/common/json/mocks"
	mockplatformcache "los-kmb-api/shared/common/platformcache/mocks"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/mock"
)

// MockEvent is a mock implementation for the Event interface
type MockEvent struct {
	mock.Mock
}

func (m *MockEvent) GetBody() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *MockEvent) GetKey() []byte {
	args := m.Called()
	return args.Get(0).([]byte)
}

func (m *MockEvent) GetTopicPartition() kafka.TopicPartition {
	args := m.Called()
	return args.Get(0).(kafka.TopicPartition)
}

func (m *MockEvent) GetTimestamp() time.Time {
	args := m.Called()
	return args.Get(0).(time.Time)
}

func (m *MockEvent) GetOpaque() interface{} {
	args := m.Called()
	return args.Get(0)
}

func (m *MockEvent) GetHeaders() []kafka.Header {
	args := m.Called()
	return args.Get(0).([]kafka.Header)
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
				"id_number": "123",
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
				"legal_name": "MGwNDewJ8H=",
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
				"surgate_mother_name": " ",
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
					"spouse_id_number": "Tcrz5",
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
				"gender": "M",
				"surgate_mother_name": "1LUjPy3GQdAs4E9rPuLVuKjGLjZqm/AqoglB5g==",
				"bpkb_name": "K",
				"spouse": {
					"spouse_id_number": "Tcrz599clw886iyL3A5Boc1yM+LOVGGHBnaW9vgSvOY=",
					"spouse_legal_name": "MGwNDewJ8HdHwdnOHXeNCVUKXoGh2Vm/f6uO8nOPpCClwUc=",
					"spouse_birth_date": "1971-04-15",
					"spouse_gender": "M",
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
			validator := common.NewValidator()

			mockMultiUsecase := new(mocks.MultiUsecase)
			mockUsecase := new(mocks.Usecase)
			mockRepository := new(mocks.Repository)
			mockJson := new(mocksJson.JSON)
			mockEvent := new(MockEvent)
			mockPlatformCache := new(mockplatformcache.PlatformCacheInterface)

			handler := &handlers{
				multiusecase:  mockMultiUsecase,
				usecase:       mockUsecase,
				repository:    mockRepository,
				validator:     validator,
				Json:          mockJson,
				platformCache: mockPlatformCache,
			}
			ctx := context.Background()
			startTime := utils.GenerateTimeInMilisecond()
			reqID := utils.GenerateUUID()

			ctx = context.WithValue(ctx, constant.CTX_KEY_REQUEST_TIME, startTime)
			ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
			ctx = context.WithValue(ctx, constant.CTX_KEY_IS_CONSUMER, true)

			mockEvent.On("GetBody").Return([]byte(tc.reqbody))
			mockEvent.On("GetKey").Return([]byte("filtering_12131421414_EFM0TST0020230809013"))
			mockRepository.On("SaveLogOrchestrator", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			mockUsecase.On("FilteringProspectID", "EFM0TST0020230809013").Return(request.OrderIDCheck{ProspectID: tc.checkPpid}, tc.errCheckPpid).Once()
			mockMultiUsecase.On("Filtering", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.Filtering{}, tc.errFiltering).Once()
			mockJson.On("EventServiceError", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ApiResponse{})
			mockJson.On("EventBadRequestErrorValidation", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ApiResponse{})
			mockJson.On("EventSuccess", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(response.ApiResponse{})
			mockPlatformCache.On("SetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			mockPlatformCache.On("GetCache", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
			err := handler.Filtering(ctx, mockEvent)
			if err != nil {
				t.Errorf("error '%s' was not expected, but got: ", err)
			}

		})
	}
}
