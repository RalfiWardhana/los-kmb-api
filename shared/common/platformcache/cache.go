package platformcache

import (
	"context"
	"fmt"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"os"
	"strconv"
	"strings"

	"github.com/KB-FMF/platform-library/cache"
)

type PlatformCache struct{}

type PlatformCacheInterface interface {
	SetCache(ctx context.Context, accessToken, collectionName, documentName string, value interface{}, expiredAt string) (result interface{}, err error)
	GetCache(ctx context.Context, accessToken, collectionName, documentName string) (result interface{}, err error)
}

func NewPlatformCache() PlatformCache {
	return PlatformCache{}
}

func (pc PlatformCache) SetCache(ctx context.Context, accessToken, collectionName, documentName string, value interface{}, expiredAt string) (result interface{}, err error) {
	env := os.Getenv("APP_ENV")

	if strings.Contains(strings.ToLower(env), "production") {
		env = cache.ENV_PRODUCTION
	} else if strings.Contains(strings.ToLower(env), "staging") {
		env = cache.ENV_STAGING
	} else {
		env = cache.ENV_DEVELOPMENT
	}

	cache := cache.New(env)

	data := map[string]interface{}{
		"document": value,
	}

	// optional if expired_at set in request, will be overwrite default expired of collection_name when you set in platform-cms.
	// default 300 (second) set in platform-cms
	v, errAtoi := strconv.Atoi(expiredAt)
	if errAtoi == nil {
		data["expired_at"] = v
	}

	set, errCache := cache.SetCache(accessToken, collectionName, documentName, data)
	if errCache != nil {
		// Write Error Log
		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       os.Getenv("DUMMY_URL_LOGS"),
			Action:     "SET_CACHE",
			Type:       "CACHE_PLATFORM_LIBRARY",
			LogFile:    constant.NEW_KMB_LOG,
			MsgLogFile: constant.MSG_SET_DATA_CACHE,
			LevelLog:   constant.PLATFORM_LOG_LEVEL_WARNING,
			Request:    value,
			Response: map[string]interface{}{
				"errors": errCache.ErrorMessage(),
				"code":   errCache.GetErrorCode(),
			},
		})
		err = fmt.Errorf("msg: %v, code: %v", errCache.ErrorMessage(), errCache.GetErrorCode())
		return
	}

	// Write Success Log
	common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Action:     "SET_CACHE",
		Type:       "CACHE_PLATFORM_LIBRARY",
		LogFile:    constant.NEW_KMB_LOG,
		MsgLogFile: constant.MSG_SET_DATA_CACHE,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    value,
		Response: map[string]interface{}{
			"messages": "success set data cache",
		},
	})

	result = set.Data
	return
}

func (pc PlatformCache) GetCache(ctx context.Context, accessToken, collectionName, documentName string) (result interface{}, err error) {
	env := os.Getenv("APP_ENV")

	if strings.Contains(strings.ToLower(env), "production") {
		env = cache.ENV_PRODUCTION
	} else if strings.Contains(strings.ToLower(env), "staging") {
		env = cache.ENV_STAGING
	} else {
		env = cache.ENV_DEVELOPMENT
	}

	cache := cache.New(env)

	get, errCache := cache.GetCache(accessToken, collectionName, documentName)
	if errCache != nil {
		// Write Error Log
		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       os.Getenv("DUMMY_URL_LOGS"),
			Action:     "GET_CACHE",
			Type:       "CACHE_PLATFORM_LIBRARY",
			LogFile:    constant.NEW_KMB_LOG,
			MsgLogFile: constant.MSG_GET_DATA_CACHE,
			LevelLog:   constant.PLATFORM_LOG_LEVEL_WARNING,
			Request:    documentName,
			Response: map[string]interface{}{
				"errors": errCache.ErrorMessage(),
				"code":   errCache.GetErrorCode(),
			},
		})
		err = fmt.Errorf("msg: %v, code: %v", errCache.ErrorMessage(), errCache.GetErrorCode())
		return
	}

	// Write Success Log
	common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Action:     "GET_CACHE",
		Type:       "CACHE_PLATFORM_LIBRARY",
		LogFile:    constant.NEW_KMB_LOG,
		MsgLogFile: constant.MSG_GET_DATA_CACHE,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    documentName,
		Response:   get.Data,
	})

	result = get.Data["document"]

	return
}
