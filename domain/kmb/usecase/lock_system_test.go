package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/kmb/interfaces/mocks"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLockSystem(t *testing.T) {
	var date, _ = time.Parse("2006-01-02", "2025-01-02")
	testcases := []struct {
		name                 string
		idNumber             string
		config               entity.AppConfig
		configValue          response.LockSystemConfig
		encryptedIDNumber    entity.EncryptedString
		trxReject            []entity.TrxLockSystem
		trxCancel            []entity.TrxLockSystem
		trxLockSystem        entity.TrxLockSystem
		errGetEncB64         error
		errGetTrxLockSystem  error
		errGetConfig         error
		errGetTrxReject      error
		errSaveTrxLockSystem error
		errGetTrxCancel      error
		result               response.LockSystem
		err                  error
	}{
		{
			name:     "test lock system GetEncB64 error",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			errGetEncB64: errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetEncB64 Error"),
			err:          errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetEncB64 Error"),
		},
		{
			name:     "test lock system GetTrxLockSystem error",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			errGetTrxLockSystem: errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxLockSystem Error"),
			err:                 errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxLockSystem Error"),
		},
		{
			name:     "test lock system banned idnumber true",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			trxLockSystem: entity.TrxLockSystem{
				ProspectID: "TEST1",
				UnbanDate:  date,
			},
			result: response.LockSystem{
				IsBanned:  true,
				UnbanDate: "2025-01-02",
			},
		},
		{
			name:     "test lock system GetConfig error",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			errGetConfig: errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetConfig Error"),
			err:          errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetConfig Error"),
		},
		{
			name:     "test lock system GetTrxReject error",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			config: entity.AppConfig{
				Value: `{"data":{"lock_reject_attempt":2,"lock_reject_ban":30,"lock_reject_check":30,"lock_cancel_attempt":2,"lock_cancel_ban":1,"lock_cancel_check":1,"lock_start_date":"2024-12-01"}}`,
			},
			errGetTrxReject: errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxReject Error"),
			err:             errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxReject Error"),
		},
		{
			name:     "test lock system RejectAttempt more than threshold",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			config: entity.AppConfig{
				Value: `{"data":{"lock_reject_attempt":2,"lock_reject_ban":30,"lock_reject_check":30,"lock_cancel_attempt":2,"lock_cancel_ban":1,"lock_cancel_check":1,"lock_start_date":"2024-12-01"}}`,
			},
			trxReject: []entity.TrxLockSystem{
				{
					ProspectID: "reject1",
					UnbanDate:  date,
				},
				{
					ProspectID: "reject2",
				},
			},
			result: response.LockSystem{
				IsBanned:  true,
				Reason:    constant.PERNAH_REJECT,
				UnbanDate: "2025-01-02",
			},
		},
		{
			name:     "test lock system RejectAttempt more than threshold SaveTrxLockSystem error",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			config: entity.AppConfig{
				Value: `{"data":{"lock_reject_attempt":2,"lock_reject_ban":30,"lock_reject_check":30,"lock_cancel_attempt":2,"lock_cancel_ban":1,"lock_cancel_check":1,"lock_start_date":"2024-12-01"}}`,
			},
			trxReject: []entity.TrxLockSystem{
				{
					ProspectID: "reject1",
					UnbanDate:  date,
				},
				{
					ProspectID: "reject2",
				},
			},
			result: response.LockSystem{
				IsBanned:  true,
				Reason:    constant.PERNAH_REJECT,
				UnbanDate: "2025-01-02",
			},
			errSaveTrxLockSystem: errors.New(constant.ERROR_UPSTREAM + " - LockSystem SaveTrxLockSystem trxReject Error"),
			err:                  errors.New(constant.ERROR_UPSTREAM + " - LockSystem SaveTrxLockSystem trxReject Error"),
		},
		{
			name:     "test lock system GetTrxCancel error",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			config: entity.AppConfig{
				Value: `{"data":{"lock_reject_attempt":2,"lock_reject_ban":30,"lock_reject_check":30,"lock_cancel_attempt":2,"lock_cancel_ban":1,"lock_cancel_check":1,"lock_start_date":"2024-12-01"}}`,
			},
			errGetTrxCancel: errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxCancel Error"),
			err:             errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxCancel Error"),
		},
		{
			name:     "test lock system CancelAttempt more than threshold",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			config: entity.AppConfig{
				Value: `{"data":{"lock_reject_attempt":2,"lock_reject_ban":30,"lock_reject_check":30,"lock_cancel_attempt":2,"lock_cancel_ban":1,"lock_cancel_check":1,"lock_start_date":"2024-12-01"}}`,
			},
			trxCancel: []entity.TrxLockSystem{
				{
					ProspectID: "cancel1",
					UnbanDate:  date,
				},
				{
					ProspectID: "cancel1",
				},
			},
			result: response.LockSystem{
				IsBanned:  true,
				Reason:    constant.PERNAH_CANCEL,
				UnbanDate: "2025-01-02",
			},
		},
		{
			name:     "test lock system CancelAttempt more than threshold SaveTrxLockSystem error",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			config: entity.AppConfig{
				Value: `{"data":{"lock_reject_attempt":2,"lock_reject_ban":30,"lock_reject_check":30,"lock_cancel_attempt":2,"lock_cancel_ban":1,"lock_cancel_check":1,"lock_start_date":"2024-12-01"}}`,
			},
			trxCancel: []entity.TrxLockSystem{
				{
					ProspectID: "cancel1",
					UnbanDate:  date,
				},
				{
					ProspectID: "cancel1",
				},
			},
			result: response.LockSystem{
				IsBanned:  true,
				Reason:    constant.PERNAH_CANCEL,
				UnbanDate: "2025-01-02",
			},
			errSaveTrxLockSystem: errors.New(constant.ERROR_UPSTREAM + " - LockSystem SaveTrxLockSystem trxCancel Error"),
			err:                  errors.New(constant.ERROR_UPSTREAM + " - LockSystem SaveTrxLockSystem trxCancel Error"),
		},
		{
			name:     "test lock system done",
			idNumber: "1234567",
			encryptedIDNumber: entity.EncryptedString{
				MyString: "TESTIDNUMBER",
			},
			config: entity.AppConfig{
				Value: `{"data":{"lock_reject_attempt":2,"lock_reject_ban":30,"lock_reject_check":30,"lock_cancel_attempt":2,"lock_cancel_ban":1,"lock_cancel_check":1,"lock_start_date":"2024-12-01"}}`,
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepository := new(mocks.Repository)
			mockHttpClient := new(httpclient.MockHttpClient)

			mockRepository.On("GetEncB64", tc.idNumber).Return(tc.encryptedIDNumber, tc.errGetEncB64)
			mockRepository.On("GetTrxLockSystem", tc.encryptedIDNumber.MyString).Return(tc.trxLockSystem, tc.errGetTrxLockSystem)
			mockRepository.On("GetConfig", "lock_system", "KMB-OFF", "lock_system_kmb").Return(tc.config, tc.errGetConfig)
			mockRepository.On("GetTrxReject", tc.encryptedIDNumber.MyString, mock.Anything).Return(tc.trxReject, tc.errGetTrxReject)
			mockRepository.On("SaveTrxLockSystem", mock.Anything).Return(tc.errSaveTrxLockSystem)
			mockRepository.On("GetTrxCancel", tc.encryptedIDNumber.MyString, mock.Anything).Return(tc.trxCancel, tc.errGetTrxCancel)

			usecase := NewUsecase(mockRepository, mockHttpClient)

			result, err := usecase.LockSystem(context.Background(), tc.idNumber)
			require.Equal(t, tc.result, result)
			require.Equal(t, tc.err, err)
		})
	}

}
