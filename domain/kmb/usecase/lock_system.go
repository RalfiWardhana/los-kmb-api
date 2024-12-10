package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
)

func (u usecase) LockSystem(ctx context.Context, idNumber string) (data response.LockSystem, err error) {
	var (
		config            entity.AppConfig
		configValue       response.LockSystemConfig
		encryptedIDNumber entity.EncryptedString
		trxReject         []entity.TrxLockSystem
		trxCancel         []entity.TrxLockSystem
		trxLockSystem     entity.TrxLockSystem
	)

	encryptedIDNumber, err = u.repository.GetEncB64(idNumber)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetEncB64 Error")
		return
	}

	//scan banned IDNumber
	trxLockSystem, err = u.repository.GetTrxLockSystem(encryptedIDNumber.MyString)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxLockSystem Error")
		return
	}

	if trxLockSystem.ProspectID != "" {
		data.IsBanned = true
		data.Reason = trxLockSystem.Reason
		data.UnbanDate = trxLockSystem.UnbanDate.Format(constant.FORMAT_DATE)
		return
	}

	//Get parameterize config
	config, err = u.repository.GetConfig("lock_system", "KMB-OFF", "lock_system_kmb")
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetConfig Error")
		return
	}

	json.Unmarshal([]byte(config.Value), &configValue)

	if configValue.Data.LockRejectBan > 0 {
		configValue.Data.LockRejectBan -= 1
	}

	if configValue.Data.LockRejectCheck > 0 {
		configValue.Data.LockRejectCheck -= 1
	}

	if configValue.Data.LockCancelBan > 0 {
		configValue.Data.LockCancelBan -= 1
	}

	if configValue.Data.LockCancelCheck > 0 {
		configValue.Data.LockCancelCheck -= 1
	}

	trxReject, err = u.repository.GetTrxReject(encryptedIDNumber.MyString, configValue)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxReject Error")
		return
	}

	if len(trxReject) >= configValue.Data.LockRejectAttempt {
		trxReject[0].Reason = constant.PERNAH_REJECT
		data.IsBanned = true
		data.Reason = trxReject[0].Reason
		data.UnbanDate = trxReject[0].UnbanDate.Format(constant.FORMAT_DATE)

		err = u.repository.SaveTrxLockSystem(trxReject[0])
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - LockSystem SaveTrxLockSystem trxReject Error")
			return
		}
		return
	}

	trxCancel, err = u.repository.GetTrxCancel(encryptedIDNumber.MyString, configValue)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - LockSystem GetTrxCancel Error")
		return
	}

	if len(trxCancel) >= configValue.Data.LockCancelAttempt {
		trxCancel[0].Reason = constant.PERNAH_CANCEL
		data.IsBanned = true
		data.Reason = trxCancel[0].Reason
		data.UnbanDate = trxCancel[0].UnbanDate.Format(constant.FORMAT_DATE)

		err = u.repository.SaveTrxLockSystem(trxCancel[0])
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - LockSystem SaveTrxLockSystem trxCancel Error")
			return
		}
		return
	}

	return
}
