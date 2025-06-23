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
	"golang.org/x/sync/singleflight"
)

type (
	usecase struct {
		repository interfaces.Repository
		httpclient httpclient.HttpClient
		cache      cache.Repository
		sfGroup    *singleflight.Group
	}
)

func NewUsecase(repository interfaces.Repository, httpclient httpclient.HttpClient, cache cache.Repository) interfaces.Usecase {
	return &usecase{
		repository: repository,
		httpclient: httpclient,
		cache:      cache,
		sfGroup:    &singleflight.Group{},
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

	cacheTTL := 5 * time.Minute

	filteringCacheKey := "filtering_result:" + reqs.ProspectID

	filteringResult, err, _ := u.sfGroup.Do(filteringCacheKey, func() (interface{}, error) {
		// Try to get from cache first
		filteringCacheData, cacheErr := u.cache.GetWithExpiration(filteringCacheKey)
		if cacheErr == nil && len(filteringCacheData) > 0 {
			var cachedFiltering entity.FilteringKMB
			if jsonErr := json.Unmarshal(filteringCacheData, &cachedFiltering); jsonErr == nil {
				return cachedFiltering, nil
			}
		}

		// Not found in cache or unmarshal error, get from repository
		repoFiltering, repoErr := u.repository.GetFilteringResult(reqs.ProspectID)
		if repoErr == nil {
			cacheData, jsonErr := json.Marshal(repoFiltering)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(filteringCacheKey, cacheData, cacheTTL)
			}
		}
		return repoFiltering, repoErr
	})

	if err != nil {
		if err.Error() == constant.RECORD_NOT_FOUND {
			err = errors.New(constant.ERROR_BAD_REQUEST + " - Silahkan melakukan filtering terlebih dahulu")
		} else {
			err = errors.New(constant.ERROR_UPSTREAM + " - Get Filtering Error")
		}
		return
	}

	filteringKMB = filteringResult.(entity.FilteringKMB)

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

			var configResult interface{}
			configResult, err, _ = u.sfGroup.Do(configCacheKey, func() (interface{}, error) {
				// Try to get from cache first
				configCacheData, cacheErr := u.cache.GetWithExpiration(configCacheKey)
				if cacheErr == nil && len(configCacheData) > 0 {
					var cachedConfig entity.AppConfig
					if jsonErr := json.Unmarshal(configCacheData, &cachedConfig); jsonErr == nil {
						return cachedConfig, nil
					}
				}

				// Not found in cache or unmarshal error, get from repository
				repoConfig, repoErr := u.repository.GetConfig("expired_contract", "KMB-OFF", "expired_contract_check")
				if repoErr == nil {
					cacheData, jsonErr := json.Marshal(repoConfig)
					if jsonErr == nil {
						_ = u.cache.SetWithExpiration(configCacheKey, cacheData, cacheTTL)
					}
				}
				return repoConfig, repoErr
			})

			if err != nil {
				err = errors.New(constant.ERROR_UPSTREAM + " - Get Expired Contract Config Error")
				return
			}

			expiredContractConfig = configResult.(entity.AppConfig)
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

	detailResult, err, _ := u.sfGroup.Do(detailCacheKey, func() (interface{}, error) {
		// Try to get from cache first
		detailCacheData, cacheErr := u.cache.GetWithExpiration(detailCacheKey)
		if cacheErr == nil && len(detailCacheData) > 0 {
			var cachedDetails []entity.TrxDetailBiro
			if jsonErr := json.Unmarshal(detailCacheData, &cachedDetails); jsonErr == nil {
				return cachedDetails, nil
			}
		}

		// Not found in cache or unmarshal error, get from repository
		repoDetails, repoErr := u.repository.GetFilteringDetail(reqs.ProspectID)
		if repoErr == nil && len(repoDetails) > 0 {
			cacheData, jsonErr := json.Marshal(repoDetails)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(detailCacheKey, cacheData, cacheTTL)
			}
		}
		return repoDetails, repoErr
	})

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetFilteringDetail error - " + err.Error())
		return
	}

	filteringDetail = detailResult.([]entity.TrxDetailBiro)

	if len(filteringDetail) == 0 {
		gradePBK = constant.DECISION_PBK_NO_HIT

	} else {
		// Try to get mappingPBKScoreGrade from cache
		pbkScoreCacheKey := "mapping_pbk_score_grade"

		var pbkScoreResult interface{}
		pbkScoreResult, err, _ = u.sfGroup.Do(pbkScoreCacheKey, func() (interface{}, error) {
			// Try to get from cache first
			pbkScoreCacheData, cacheErr := u.cache.GetWithExpiration(pbkScoreCacheKey)
			if cacheErr == nil && len(pbkScoreCacheData) > 0 {
				var cachedPBKScore []entity.MappingPBKScoreGrade
				if jsonErr := json.Unmarshal(pbkScoreCacheData, &cachedPBKScore); jsonErr == nil {
					return cachedPBKScore, nil
				}
			}

			// Not found in cache or unmarshal error, get from repository
			repoPBKScore, repoErr := u.repository.GetMappingPBKScoreGrade()
			if repoErr == nil {
				cacheData, jsonErr := json.Marshal(repoPBKScore)
				if jsonErr == nil {
					_ = u.cache.SetWithExpiration(pbkScoreCacheKey, cacheData, cacheTTL)
				}
			}
			return repoPBKScore, repoErr
		})

		if err != nil {
			err = errors.New(constant.ERROR_UPSTREAM + " - GetMappingPBKScoreGrade error - " + err.Error())
			return
		}

		mappingPBKScoreGrade = pbkScoreResult.([]entity.MappingPBKScoreGrade)

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

	branchResult, err, _ := u.sfGroup.Do(branchCacheKey, func() (interface{}, error) {
		// Try to get from cache first
		branchCacheData, cacheErr := u.cache.GetWithExpiration(branchCacheKey)
		if cacheErr == nil && len(branchCacheData) > 0 {
			var cachedBranch entity.MappingBranchByPBKScore
			if jsonErr := json.Unmarshal(branchCacheData, &cachedBranch); jsonErr == nil {
				return cachedBranch, nil
			}
		}

		// Not found in cache or unmarshal error, get from repository
		repoBranch, repoErr := u.repository.GetMappingBranchPBK(filteringKMB.BranchID, gradePBK)
		if repoErr == nil {
			cacheData, jsonErr := json.Marshal(repoBranch)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(branchCacheKey, cacheData, cacheTTL)
			}
		}
		return repoBranch, repoErr
	})

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - GetMappingBranchPBK error - " + err.Error())
		return
	}

	mappingBranch = branchResult.(entity.MappingBranchByPBKScore)

	if mappingBranch.GradeBranch == "" {
		mappingBranch.GradeBranch = constant.GOOD
	}

	// Try to get mappingElaborateLTV from cache
	elaborateCacheKey := "mapping_elaborate_ltv:" + resultPefindo + ":" + cluster + ":" +
		strconv.Itoa(bpkbNameType) + ":" + filteringKMB.CustomerStatus.(string) + ":" +
		gradePBK + ":" + mappingBranch.GradeBranch

	elaborateResult, err, _ := u.sfGroup.Do(elaborateCacheKey, func() (interface{}, error) {
		// Try to get from cache first
		elaborateCacheData, cacheErr := u.cache.GetWithExpiration(elaborateCacheKey)
		if cacheErr == nil && len(elaborateCacheData) > 0 {
			var cachedElaborate []entity.MappingElaborateLTV
			if jsonErr := json.Unmarshal(elaborateCacheData, &cachedElaborate); jsonErr == nil {
				return cachedElaborate, nil
			}
		}

		// Not found in cache or unmarshal error, get from repository
		repoElaborate, repoErr := u.repository.GetMappingElaborateLTV(resultPefindo, cluster, bpkbNameType, filteringKMB.CustomerStatus.(string), gradePBK, mappingBranch.GradeBranch)
		if repoErr == nil {
			cacheData, jsonErr := json.Marshal(repoElaborate)
			if jsonErr == nil {
				_ = u.cache.SetWithExpiration(elaborateCacheKey, cacheData, cacheTTL)
			}
		}
		return repoElaborate, repoErr
	})

	if err != nil {
		err = errors.New(constant.ERROR_UPSTREAM + " - Get mapping elaborate error")
		return
	}

	mappingElaborateLTV = elaborateResult.([]entity.MappingElaborateLTV)

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
