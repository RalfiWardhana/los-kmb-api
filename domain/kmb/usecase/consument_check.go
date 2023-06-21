package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	entity "los-kmb-api/models/dupcheck"
	response "los-kmb-api/models/dupcheck"
	"los-kmb-api/shared/constant"
)

func (u usecase) ConsumentCheck(req response.SpDupCekCustomerByID) (data response.UsecaseApi, mapping response.SpDupcheckMap, err error) {

	if req != (response.SpDupCekCustomerByID{}) {

		var config entity.AppConfig

		config, err = u.repository.GetDupcheckConfig()

		if err != nil {
			err = errors.New("upstream_service_error - Error Get Parameterize Config")
			return
		}

		var configValue response.DupcheckConfig

		json.Unmarshal([]byte(config.Value), &configValue)

		if req.MaxOverdueDaysROAO != nil {
			mapping.MaxOverdueDaysROAO = *req.MaxOverdueDaysROAO
		}

		if req.NumberOfPaidInstallment != nil {
			mapping.NumberOfPaidInstallment = *req.NumberOfPaidInstallment
		}

		if req.MaxOverdueDaysforActiveAgreement != nil {
			mapping.MaxOverdueDaysforActiveAgreement = *req.MaxOverdueDaysforActiveAgreement
		}

		mapping.CustomerID = req.CustomerID

		if (req.TotalInstallment <= 0 && req.RRDDate != nil) || (req.TotalInstallment > 0 && req.RRDDate != nil && req.NumberOfPaidInstallment == nil) {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_RO

			if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_RO_OVDLTE30
				data.Reason = fmt.Sprintf("RO - OVD Maks <= %d days", configValue.Data.MinOvd)

			} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MaxOvd {

				data.Result = constant.DECISION_PASS
				data.Code = constant.CODE_RO_OVDGT30_LTE90
				data.Reason = fmt.Sprintf("RO - OVD Maks > %d-%d days", configValue.Data.MinOvd, configValue.Data.MaxOvd)

			} else {

				data.Result = constant.DECISION_REJECT
				data.Code = constant.CODE_RO_OVDGT90
				data.Reason = fmt.Sprintf("RO - OVD Maks > %d days", configValue.Data.MaxOvd)

			}

			return

		} else if req.TotalInstallment > 0 {

			if req.NumberOfPaidInstallment != nil && mapping.NumberOfPaidInstallment >= 0 {

				data.StatusKonsumen = constant.STATUS_KONSUMEN_AO

				if req.OSInstallmentDue <= 0 {

					if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment >= configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSGT6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Lancar >= %d bulan Angsuran - OVD Maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MaxOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSGT6_OVDGT30_LTE90
							data.Reason = fmt.Sprintf("AO - Lancar >= %d bulan Angsuran - OVD maks > %d-%d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd, configValue.Data.MaxOvd)

						} else {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_LANCAR_ANGSGT6_OVDGT90
							data.Reason = fmt.Sprintf("AO - Lancar >= %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MaxOvd)

						}

					} else if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment < configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_OVDGT90
							data.Reason = fmt.Sprintf("AO - OVD Maks > %d days", configValue.Data.MaxOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSLTE6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Lancar < %d bulan Angsuran - OVD maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)
						} else {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_LANCAR_ANGSLTE6_OVDGT30
							data.Reason = fmt.Sprintf("AO - Lancar < %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)
						}
					}

				} else {

					if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment >= configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSGT6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Menunggak >= %d bulan Angsuran - OVD Maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MaxOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSGT6_OVDGT30_LTE90
							data.Reason = fmt.Sprintf("AO - Menunggak >= %d bulan Angsuran - OVD maks > %d - %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd, configValue.Data.MaxOvd)
						} else {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSGT6_OVDGT90
							data.Reason = fmt.Sprintf("AO - Menunggak >= %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MaxOvd)

						}

					} else if req.MaxOverdueDaysROAO != nil && mapping.NumberOfPaidInstallment < configValue.Data.AngsuranBerjalan {

						if mapping.MaxOverdueDaysROAO > configValue.Data.MaxOvd {

							data.Result = constant.DECISION_REJECT
							data.Code = constant.CODE_AO_OVDGT90
							data.Reason = fmt.Sprintf("AO - OVD Maks > %d days", configValue.Data.MaxOvd)

						} else if mapping.MaxOverdueDaysROAO <= configValue.Data.MinOvd {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSLTE6_OVDLTE30
							data.Reason = fmt.Sprintf("AO - Menunggak < %d bulan Angsuran - OVD maks <= %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)

						} else {

							data.Result = constant.DECISION_PASS
							data.Code = constant.CODE_AO_MENUNGGAK_ANGSLTE6_OVDGT30
							data.Reason = fmt.Sprintf("AO - Menunggak < %d bulan Angsuran - OVD maks > %d days", configValue.Data.AngsuranBerjalan, configValue.Data.MinOvd)
						}
					}
				}

			}

		} else {
			data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
			data.Result = constant.DECISION_PASS
			data.Code = constant.CODE_KONSUMEN_UNIDENTIFIED
			data.Reason = constant.REASON_KONSUMEN_UNIDENTIFIED

		}

	} else {
		data.StatusKonsumen = constant.STATUS_KONSUMEN_NEW
		data.Result = constant.DECISION_PASS
		data.Code = constant.CODE_KONSUMEN_NEW
		data.Reason = constant.REASON_KONSUMEN_NEW
	}

	return
}
