package usecase

import (
	"context"
	"errors"
	"los-kmb-api/domain/elaborate_ltv/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type (
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
	}
)

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
	}
}

func (u usecase) Elaborate(ctx context.Context, reqs request.ElaborateLTV, accessToken string) (data response.ElaborateLTV, err error) {

	var (
		filteringKMB        entity.FilteringKMB
		ageS                string
		bakiDebet           float64
		bpkbNameType        int
		manufacturingYear   time.Time
		mappingElaborateLTV []entity.MappingElaborateLTV
	)

	filteringKMB, err = u.repository.GetFilteringResult(reqs.ProspectID)
	if err != nil {
		if err.Error() == constant.RECORD_NOT_FOUND {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Silahkan melakukan filtering terlebih dahulu")
		} else {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error")
		}
		return
	}

	if filteringKMB.NextProcess != 1 {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Tidak bisa lanjut proses")
		return
	}

	resultPefindo := filteringKMB.Decision
	if filteringKMB.ScoreBiro == nil || filteringKMB.ScoreBiro == "" || filteringKMB.ScoreBiro == constant.UNSCORE_PBK {
		if filteringKMB.CustomerSegment != nil && !strings.Contains("PRIME PRIORITY", filteringKMB.CustomerSegment.(string)) {
			resultPefindo = constant.DECISION_PBK_NO_HIT
		}
	}

	if strings.Contains(os.Getenv("NAMA_SAMA"), filteringKMB.BpkbName) {
		bpkbNameType = 1
	}

	now := time.Now()
	// convert date to year for age_vehicle
	manufacturingYear, err = time.Parse("2006", reqs.ManufacturingYear)
	if err != nil {
		err = errors.New(constant.ERROR_BAD_REQUEST + " - Format tahun kendaraan tidak sesuai")
		return
	}
	subManufacturingYear := now.Sub(manufacturingYear)
	age := int((subManufacturingYear.Hours()/24)/365) + (reqs.Tenor / 12)
	if age <= 12 {
		ageS = "<=12"
	} else {
		ageS = ">12"
	}

	if filteringKMB.TotalBakiDebetNonCollateralBiro != nil {
		bakiDebet, err = utils.GetFloat(filteringKMB.TotalBakiDebetNonCollateralBiro)
		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " baki debet " + err.Error())
			return
		}
	}

	trxElaborateLTV := entity.TrxElaborateLTV{
		ProspectID:        reqs.ProspectID,
		RequestID:         ctx.Value(echo.HeaderXRequestID).(string),
		Tenor:             reqs.Tenor,
		ManufacturingYear: reqs.ManufacturingYear,
	}

	mappingElaborateLTV, err = u.repository.GetMappingElaborateLTV(resultPefindo, filteringKMB.Cluster.(string))
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error")
		return
	}
	for _, m := range mappingElaborateLTV {
		if reqs.Tenor >= 36 {
			//no hit
			if resultPefindo == constant.DECISION_PBK_NO_HIT && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}

			//pass
			if resultPefindo == constant.DECISION_PASS && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}

			//reject
			if resultPefindo == constant.DECISION_REJECT && m.TotalBakiDebetStart <= int(bakiDebet) && int(bakiDebet) <= m.TotalBakiDebetEnd && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd && bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}
		} else {
			//no hit
			if resultPefindo == constant.DECISION_PBK_NO_HIT && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}

			//pass
			if resultPefindo == constant.DECISION_PASS && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}

			//reject
			if resultPefindo == constant.DECISION_REJECT && m.TotalBakiDebetStart <= int(bakiDebet) && int(bakiDebet) <= m.TotalBakiDebetEnd && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}
		}

		// max tenor
		if resultPefindo == constant.DECISION_REJECT && int(bakiDebet) > constant.RANGE_CLUSTER_BAKI_DEBET_REJECT && strings.Contains("Cluster E Cluster F", filteringKMB.Cluster.(string)) {
			data.LTV = 0
			data.MaxTenor = 0
			data.AdjustTenor = false
			data.Reason = constant.REASON_REJECT_ELABORATE
		} else {
			if m.TenorEnd >= data.MaxTenor && m.LTV > 0 {
				if m.BPKBNameType == 1 && m.AgeVehicle != "" {
					if bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
						data.MaxTenor = m.TenorEnd
						data.AdjustTenor = true
					}
				} else {
					data.MaxTenor = m.TenorEnd
					data.AdjustTenor = true
				}
			}
		}
	}

	err = u.repository.SaveTrxElaborateLTV(trxElaborateLTV)
	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " Save elaborate ltv error")
		return
	}

	return
}
