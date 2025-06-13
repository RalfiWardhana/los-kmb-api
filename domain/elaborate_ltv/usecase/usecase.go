package usecase

import (
	"context"
	"encoding/json"
	"errors"
	cache "los-kmb-api/domain/cache/interfaces"
	"los-kmb-api/domain/elaborate_ltv/interfaces"
	"los-kmb-api/models/entity"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type (
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		cache      cache.Repository
	}
)

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, cache cache.Repository) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
		cache:      cache,
	}
}

func (u usecase) Elaborate(ctx context.Context, reqs request.ElaborateLTV, accessToken string) (data response.ElaborateLTV, err error) {

	var (
		filteringKMB            entity.FilteringKMB
		ageS                    string
		bakiDebet               float64
		bpkbNameType            int
		manufacturingYear       time.Time
		mappingElaborateLTV     []entity.MappingElaborateLTV
		cluster                 string
		RrdDateString           string
		CreatedAtString         string
		RrdDate                 time.Time
		CreatedAt               time.Time
		MonthsOfExpiredContract int
		OverrideFlowLikeRegular bool
		expiredContractConfig   entity.AppConfig
	)

	// Try to get filteringKMB from cache first
	filteringCacheKey := "filtering_result:" + reqs.ProspectID
	filteringCacheData, err := u.cache.GetWithExpiration(filteringCacheKey)
	// fmt.Printf("GET Cache Key: %s, Cache Data Length: %d\n", filteringCacheKey, len(filteringCacheData))
	if err == nil && len(filteringCacheData) > 0 {
		// Data found in cache
		if err = json.Unmarshal(filteringCacheData, &filteringKMB); err != nil {
			// If unmarshal fails, proceed to get from repository
			filteringKMB, err = u.repository.GetFilteringResult(reqs.ProspectID)
		}
	} else {
		// Not found in cache, get from repository
		filteringKMB, err = u.repository.GetFilteringResult(reqs.ProspectID)
		if err == nil {
			cacheData, jsonErr := json.Marshal(filteringKMB)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(filteringCacheKey, cacheData, 5*time.Minute)
			}
		}
	}

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

	if filteringKMB.CustomerSegment != nil && strings.Contains("PRIME PRIORITY", filteringKMB.CustomerSegment.(string)) {
		cluster = constant.CLUSTER_PRIME_PRIORITY

		// Cek apakah customer RO PRIME/PRIORITY ini termasuk jalur `expired_contract tidak <= 6 bulan`
		if filteringKMB.CustomerStatus == constant.STATUS_KONSUMEN_RO {
			if filteringKMB.RrdDate == nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Customer RO then rrd_date should not be empty")
				return
			}

			RrdDateString = filteringKMB.RrdDate.(string)
			CreatedAtString = filteringKMB.CreatedAt.Format(time.RFC3339)

			RrdDate, err = time.Parse(time.RFC3339, RrdDateString)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of RrdDate (" + RrdDateString + ")")
				return
			}

			CreatedAt, err = time.Parse(time.RFC3339, CreatedAtString)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Error parsing date of CreatedAt (" + CreatedAtString + ")")
				return
			}

			MonthsOfExpiredContract, err = utils.PreciseMonthsDifference(RrdDate, CreatedAt)
			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Difference of months RrdDate and CreatedAt is negative (-)")
				return
			}

			// Get config expired_contract
			configCacheKey := "ConfigCheckExpiredContractKMB"
			configCacheData, cacheErr := u.cache.GetWithExpiration(configCacheKey)
			// fmt.Printf("GET Cache Key: %s, Cache Data Length: %d\n", configCacheKey, len(configCacheData))
			if cacheErr == nil && len(configCacheData) > 0 {
				if err = json.Unmarshal(configCacheData, &expiredContractConfig); err != nil {
					// If unmarshal fails, proceed to get from repository
					expiredContractConfig, err = u.repository.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
				}
			} else {
				// Not found in cache, get from repository
				expiredContractConfig, err = u.repository.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
				if err == nil {
					cacheData, jsonErr := json.Marshal(expiredContractConfig)
					if jsonErr == nil {
						_ = u.cache.SetWithExpiration(configCacheKey, cacheData, 5*time.Minute)
					}
				}
			}

			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Get Expired Contract Config Error")
				return
			}

			var configValueExpContract response.ExpiredContractConfig
			json.Unmarshal([]byte(expiredContractConfig.Value), &configValueExpContract)

			if configValueExpContract.Data.ExpiredContractCheckEnabled && !(MonthsOfExpiredContract <= configValueExpContract.Data.ExpiredContractMaxMonths) {
				// Jalur mirip seperti customer segment "REGULAR"
				OverrideFlowLikeRegular = true
			}
		}
	} else {
		cluster = filteringKMB.CMOCluster.(string)
	}

	if (filteringKMB.CustomerSegment != nil && !strings.Contains("PRIME PRIORITY", filteringKMB.CustomerSegment.(string))) || (OverrideFlowLikeRegular) {
		if filteringKMB.ScoreBiro == nil || filteringKMB.ScoreBiro == "" || filteringKMB.ScoreBiro == constant.UNSCORE_PBK {
			resultPefindo = constant.DECISION_PBK_NO_HIT
		} else if filteringKMB.MaxOverdueBiro != nil || filteringKMB.MaxOverdueLast12monthsBiro != nil {
			// KO Rules Include All
			ovdCurrent, _ := filteringKMB.MaxOverdueBiro.(int64)
			ovd12, _ := filteringKMB.MaxOverdueLast12monthsBiro.(int64)

			if ovdCurrent > constant.PBK_OVD_CURRENT || ovd12 > constant.PBK_OVD_LAST_12 {
				resultPefindo = constant.DECISION_REJECT
			}
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

	if OverrideFlowLikeRegular && resultPefindo == constant.DECISION_REJECT {
		cluster = filteringKMB.CustomerStatus.(string) + " " + constant.CLUSTER_PRIME_PRIORITY
		if int(bakiDebet) > constant.RANGE_CLUSTER_BAKI_DEBET_REJECT {
			cluster = filteringKMB.CustomerStatus.(string) + " " + filteringKMB.CustomerSegment.(string)
		}
	}

	var (
		filteringDetail      []entity.TrxDetailBiro
		mappingPBKScoreGrade []entity.MappingPBKScoreGrade
		gradePBK             string
		mappingBranch        entity.MappingBranchByPBKScore
	)

	// Try to get filteringDetail from cache
	detailCacheKey := "filtering_detail:" + reqs.ProspectID
	detailCacheData, err := u.cache.GetWithExpiration(detailCacheKey)
	// fmt.Printf("GET Cache Key: %s, Cache Data Length: %d\n", detailCacheKey, len(detailCacheData))
	if err == nil && len(detailCacheData) > 0 {
		if err = json.Unmarshal(detailCacheData, &filteringDetail); err != nil {
			// If unmarshal fails, proceed to get from repository
			filteringDetail, err = u.repository.GetFilteringDetail(reqs.ProspectID)
		}
	} else {
		// Not found in cache, get from repository
		filteringDetail, err = u.repository.GetFilteringDetail(reqs.ProspectID)
		if err == nil && len(filteringDetail) > 0 {
			cacheData, jsonErr := json.Marshal(filteringDetail)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(detailCacheKey, cacheData, 5*time.Minute)
			}
		}
	}

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetFilteringDetail error - " + err.Error())
		return
	}

	if len(filteringDetail) == 0 {
		gradePBK = constant.DECISION_PBK_NO_HIT

	} else {
		// Try to get mappingPBKScoreGrade from cache
		pbkScoreCacheKey := "mapping_pbk_score_grade"
		pbkScoreCacheData, cacheErr := u.cache.GetWithExpiration(pbkScoreCacheKey)
		// fmt.Printf("GET Cache Key: %s, Cache Data Length: %d\n", pbkScoreCacheKey, len(pbkScoreCacheData))
		if cacheErr == nil && len(pbkScoreCacheData) > 0 {
			if err = json.Unmarshal(pbkScoreCacheData, &mappingPBKScoreGrade); err != nil {
				// If unmarshal fails, proceed to get from repository
				mappingPBKScoreGrade, err = u.repository.GetMappingPBKScoreGrade()
			}
		} else {
			// Not found in cache, get from repository
			mappingPBKScoreGrade, err = u.repository.GetMappingPBKScoreGrade()
			if err == nil {
				cacheData, jsonErr := json.Marshal(mappingPBKScoreGrade)
				if jsonErr == nil {
					_ = u.cache.SetWithExpiration(pbkScoreCacheKey, cacheData, 5*time.Hour)
				}
			}
		}

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetMappingPBKScoreGrade error - " + err.Error())
			return
		}

		var gradeRisk int
		for _, v := range filteringDetail {
			for _, mapp := range mappingPBKScoreGrade {
				if strings.EqualFold(v.Score, mapp.Score) && mapp.GradeRisk > gradeRisk {
					gradeRisk = mapp.GradeRisk
					gradePBK = mapp.GradeScore
				}
			}
		}
	}

	// Try to get mappingBranch from cache
	branchCacheKey := "mapping_branch_pbk:" + filteringKMB.BranchID + ":" + gradePBK
	branchCacheData, err := u.cache.GetWithExpiration(branchCacheKey)
	// fmt.Printf("GET Cache Key: %s, Cache Data Length: %d\n", branchCacheKey, len(branchCacheData))
	if err == nil && len(branchCacheData) > 0 {
		if err = json.Unmarshal(branchCacheData, &mappingBranch); err != nil {
			// If unmarshal fails, proceed to get from repository
			mappingBranch, err = u.repository.GetMappingBranchPBK(filteringKMB.BranchID, gradePBK)
		}
	} else {
		// Not found in cache, get from repository
		mappingBranch, err = u.repository.GetMappingBranchPBK(filteringKMB.BranchID, gradePBK)
		if err == nil {
			cacheData, jsonErr := json.Marshal(mappingBranch)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(branchCacheKey, cacheData, 5*time.Minute)
			}
		}
	}

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetMappingBranchPBK error - " + err.Error())
		return
	}

	if mappingBranch.GradeBranch == "" {
		mappingBranch.GradeBranch = constant.GOOD
	}

	// Try to get mappingElaborateLTV from cache
	elaborateCacheKey := "mapping_elaborate_ltv:" + resultPefindo + ":" + cluster + ":" +
		strconv.Itoa(bpkbNameType) + ":" + filteringKMB.CustomerStatus.(string) + ":" +
		gradePBK + ":" + mappingBranch.GradeBranch
	elaborateCacheData, err := u.cache.GetWithExpiration(elaborateCacheKey)
	// fmt.Printf("GET Cache Key: %s, Cache Data Length: %d\n", elaborateCacheKey, len(elaborateCacheData))
	if err == nil && len(elaborateCacheData) > 0 {
		if err = json.Unmarshal(elaborateCacheData, &mappingElaborateLTV); err != nil {
			// If unmarshal fails, proceed to get from repository
			mappingElaborateLTV, err = u.repository.GetMappingElaborateLTV(resultPefindo, cluster, bpkbNameType, filteringKMB.CustomerStatus.(string), gradePBK, mappingBranch.GradeBranch)
		}
	} else {
		// Not found in cache, get from repository
		mappingElaborateLTV, err = u.repository.GetMappingElaborateLTV(resultPefindo, cluster, bpkbNameType, filteringKMB.CustomerStatus.(string), gradePBK, mappingBranch.GradeBranch)
		if err == nil {
			cacheData, jsonErr := json.Marshal(mappingElaborateLTV)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(elaborateCacheKey, cacheData, 5*time.Minute)
			}
		}
	}

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
				if m.BPKBNameType == 1 {
					if bpkbNameType == m.BPKBNameType {
						data.LTV = m.LTV
						trxElaborateLTV.MappingElaborateLTVID = m.ID
					}
				} else {
					data.LTV = m.LTV
					trxElaborateLTV.MappingElaborateLTVID = m.ID
				}
			}

			//reject
			if resultPefindo == constant.DECISION_REJECT && m.TotalBakiDebetStart <= int(bakiDebet) && int(bakiDebet) <= m.TotalBakiDebetEnd && m.TenorStart <= reqs.Tenor && reqs.Tenor <= m.TenorEnd {
				data.LTV = m.LTV
				trxElaborateLTV.MappingElaborateLTVID = m.ID
			}
		}

		// max tenor
		if m.TenorEnd >= data.MaxTenor && m.LTV > 0 {
			if m.AgeVehicle != "" {
				if bpkbNameType == m.BPKBNameType && ageS == m.AgeVehicle {
					data.MaxTenor = m.TenorEnd
					data.AdjustTenor = true
				}
			} else if m.BPKBNameType == 1 {
				if bpkbNameType == m.BPKBNameType {
					data.MaxTenor = m.TenorEnd
					data.AdjustTenor = true
				}
			} else {
				data.MaxTenor = m.TenorEnd
				data.AdjustTenor = true
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
