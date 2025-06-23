package http

import (
	"fmt"
	"los-kmb-api/domain/elaborate_ltv/interfaces"
	"los-kmb-api/middlewares"
	"los-kmb-api/models/request"
	"los-kmb-api/models/response"
	"los-kmb-api/shared/authorization"
	"los-kmb-api/shared/common"
	authPlatform "los-kmb-api/shared/common/platformauth/adapter"
	"los-kmb-api/shared/constant"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/singleflight"
)

type handlerKmbElaborate struct {
	usecase          interfaces.Usecase
	repository       interfaces.Repository
	authorization    authorization.Authorization
	Json             common.JSON
	authPlatform     authPlatform.PlatformAuthInterface
	sfGroup          *singleflight.Group
	responseCache    map[string]*cachedResponse
	responseCacheMux sync.RWMutex
	responseCacheTTL time.Duration
}

type cachedResponse struct {
	data      response.ElaborateLTV
	timestamp time.Time
}

func ElaborateHandler(kmbroute *echo.Group, usecase interfaces.Usecase, repository interfaces.Repository, authorization authorization.Authorization, json common.JSON, middlewares *middlewares.AccessMiddleware, authPlatform authPlatform.PlatformAuthInterface) {
	handler := handlerKmbElaborate{
		usecase:          usecase,
		repository:       repository,
		authorization:    authorization,
		Json:             json,
		authPlatform:     authPlatform,
		sfGroup:          &singleflight.Group{},
		responseCache:    make(map[string]*cachedResponse),
		responseCacheTTL: 30 * time.Second, // Short-lived cache to handle burst traffic for concurrent identical requests
	}

	// Start a background cleaner for the cache
	go func() {
		for {
			time.Sleep(30 * time.Second)
			handler.cleanExpiredCache()
		}
	}()

	kmbroute.POST("/elaborate", handler.Elaborate, middlewares.AccessMiddleware())
}

func (c *handlerKmbElaborate) cleanExpiredCache() {
	now := time.Now()
	c.responseCacheMux.Lock()
	defer c.responseCacheMux.Unlock()

	for key, cached := range c.responseCache {
		if now.Sub(cached.timestamp) > c.responseCacheTTL {
			delete(c.responseCache, key)
		}
	}
}

func (c *handlerKmbElaborate) getCachedResponse(key string) (response.ElaborateLTV, bool) {
	c.responseCacheMux.RLock()
	defer c.responseCacheMux.RUnlock()

	if cached, exists := c.responseCache[key]; exists {
		if time.Since(cached.timestamp) < c.responseCacheTTL {
			return cached.data, true
		}
	}
	return response.ElaborateLTV{}, false
}

func (c *handlerKmbElaborate) setCachedResponse(key string, data response.ElaborateLTV) {
	c.responseCacheMux.Lock()
	defer c.responseCacheMux.Unlock()

	c.responseCache[key] = &cachedResponse{
		data:      data,
		timestamp: time.Now(),
	}
}

// ElaborateLTV Tools godoc
// @Description ElaborateLTV
// @Tags Filtering
// @Produce json
// @Param body body request.ElaborateLTV true "Body payload"
// @Success 200 {object} response.ApiResponse{data=response.ElaborateLTV}
// @Failure 400 {object} response.ApiResponse{error=response.ErrorValidation}
// @Failure 500 {object} response.ApiResponse{}
// @Router /api/v3/kmb/elaborate [post]
func (c *handlerKmbElaborate) Elaborate(ctx echo.Context) (err error) {

	var (
		req     request.ElaborateLTV
		resp    interface{}
		ctxJson error
	)

	// Accept and validate the request first
	if err := ctx.Bind(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorBindV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
		return ctxJson
	}

	if err := ctx.Validate(&req); err != nil {
		ctxJson, resp = c.Json.BadRequestErrorValidationV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
		return ctxJson
	}

	// Save Log Orchestrator
	defer func() {
		logKey := req.ProspectID
		go func() {
			c.sfGroup.Do("log:"+logKey, func() (interface{}, error) {
				return nil, c.repository.SaveLogOrchestrator(ctx.Request().Header, req, resp, "/api/v3/kmb/elaborate", constant.METHOD_POST, req.ProspectID, ctx.Get(constant.HeaderXRequestID).(string))
			})
		}()
	}()

	_, errAuth := c.authPlatform.Validation(ctx.Request().Header.Get(constant.HEADER_AUTHORIZATION), "")
	if errAuth != nil {
		if errAuth.GetErrorCode() == "401" {
			err = fmt.Errorf(constant.ERROR_UNAUTHORIZED + " - Invalid token")
			ctxJson, resp = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
			return ctxJson
		} else {
			err = fmt.Errorf("%s - %v", constant.ERROR_UNAUTHORIZED, errAuth.ErrorMessage())
			ctxJson, resp = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
			return ctxJson
		}
	}

	// Start performance optimization: Try local cache first (extremely fast)
	cacheKey := fmt.Sprintf("elaborate:%s:%d:%s", req.ProspectID, req.Tenor, req.ManufacturingYear)
	if cachedData, found := c.getCachedResponse(cacheKey); found {
		ctxJson, resp = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, cachedData)
		return ctxJson
	}

	// Not in cache, use singleflight for concurrent identical requests
	accessToken := middlewares.UserInfoData.AccessToken

	result, err, _ := c.sfGroup.Do(cacheKey, func() (interface{}, error) {
		// Double-check cache in case another request populated it while waiting
		if cachedData, found := c.getCachedResponse(cacheKey); found {
			return cachedData, nil
		}

		// Actually call the usecase
		return c.usecase.Elaborate(ctx.Request().Context(), req, accessToken)
	})

	if err != nil {
		ctxJson, resp = c.Json.ServerSideErrorV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, err)
		return ctxJson
	}

	// Cache successful results
	data := result.(response.ElaborateLTV)
	c.setCachedResponse(cacheKey, data)

	ctxJson, resp = c.Json.SuccessV3(ctx, middlewares.UserInfoData.AccessToken, constant.NEW_KMB_LOG, "LOS - KMB ELABORATE", req, data)
	return ctxJson
}
