package usecase

import (
	"errors"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRejectTenor36(t *testing.T) {
	testcases := []struct {
		name              string
		idNumber          string
		result            response.UsecaseApi
		errResult         error
		errEnc            error
		encryptedIDNumber entity.EncryptedString
		errGetTrx         error
		trxStatus         entity.TrxStatus
	}{
		{
			name:              "handling tenor error encrypt",
			idNumber:          "198091461892",
			errEnc:            errors.New("Encrypt Error"),
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - GetEncB64 ID Number Error"),
		},
		{
			name:              "handling tenor error Trx",
			idNumber:          "198091461892",
			errGetTrx:         errors.New("Get Trx Error"),
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			errResult:         errors.New(constant.ERROR_UPSTREAM + " - GetCurrentTrxWithRejectDSR Error"),
		},
		{
			name:              "handling tenor reject",
			idNumber:          "198091461892",
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			result: response.UsecaseApi{
				Code:   constant.CODE_REJECT_TENOR,
				Result: constant.DECISION_REJECT,
				Reason: constant.REASON_REJECT_TENOR},
		},
		{
			name:              "handling tenor pass",
			idNumber:          "198091461892",
			encryptedIDNumber: entity.EncryptedString{MyString: "e6FXjuesjmzPsQlG+JRkm28vK9NoqXY3NQg7qJn4nFI="},
			result: response.UsecaseApi{
				Code:   constant.CODE_PASS_TENOR,
				Result: constant.DECISION_PASS,
				Reason: constant.REASON_PASS_TENOR},
			trxStatus: entity.TrxStatus{
				ProspectID:     "SLY-123457678",
				SourceDecision: constant.SOURCE_DECISION_DSR,
				Decision:       constant.DB_DECISION_REJECT,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetEncB64", tc.idNumber).Return(tc.encryptedIDNumber, tc.errEnc)
			mockRepository.On("GetCurrentTrxWithRejectDSR", tc.encryptedIDNumber.MyString).Return(tc.trxStatus, tc.errGetTrx)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.RejectTenor36(tc.idNumber)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.errResult, err)
		})
	}

}
