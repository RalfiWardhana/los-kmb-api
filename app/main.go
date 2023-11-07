package main

import (
	"context"
	"fmt"
	"log"
	"los-kmb-api/docs"
	cacheRepository "los-kmb-api/domain/cache/repository"
	cmsDelivery "los-kmb-api/domain/cms/delivery/http"
	cmsRepository "los-kmb-api/domain/cms/repository"
	cmsUsecase "los-kmb-api/domain/cms/usecase"
	elaborateDelivery "los-kmb-api/domain/elaborate/delivery/http"
	elaborateRepository "los-kmb-api/domain/elaborate/repository"
	elaborateUsecase "los-kmb-api/domain/elaborate/usecase"
	elaborateLTVDelivery "los-kmb-api/domain/elaborate_ltv/delivery/http"
	elaborateLTVRepository "los-kmb-api/domain/elaborate_ltv/repository"
	elaborateLTVUsecase "los-kmb-api/domain/elaborate_ltv/usecase"
	filteringDelivery "los-kmb-api/domain/filtering/delivery/http"
	filteringRepository "los-kmb-api/domain/filtering/repository"
	filteringUsecase "los-kmb-api/domain/filtering/usecase"
	eventhandlers "los-kmb-api/domain/filtering_new/delivery/event"
	newKmbFilteringDelivery "los-kmb-api/domain/filtering_new/delivery/http"
	newKmbFilteringRepository "los-kmb-api/domain/filtering_new/repository"
	newKmbFilteringUsecase "los-kmb-api/domain/filtering_new/usecase"
	eventHandler "los-kmb-api/domain/kmb/delivery/event"
	kmbDelivery "los-kmb-api/domain/kmb/delivery/http"
	kmbRepository "los-kmb-api/domain/kmb/repository"
	kmbUsecase "los-kmb-api/domain/kmb/usecase"
	toolsDelivery "los-kmb-api/domain/tools/delivery/http"
	"los-kmb-api/middlewares"
	"los-kmb-api/shared/authorization"
	authRepository "los-kmb-api/shared/authorization/repository"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/json"
	"los-kmb-api/shared/common/platformlog"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/database"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"los-kmb-api/shared/common/platformevent"

	"github.com/KB-FMF/platform-library/event"
	"github.com/allegro/bigcache/v3"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/newrelic/go-agent/v3/integrations/nrecho-v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @contact.name Kredit Plus
// @contact.url https://kreditplus.com
// @contact.email support@kreditplus.com

// @host localhost:9710
// @BasePath /api/v2/pbk
// @query.collection.format multi

func main() {
	e := echo.New()

	e.Validator = common.NewValidator()

	validator := common.NewValidator()

	config.SetupTimezone()
	config.LoadEnv()

	env := strings.ToLower(os.Getenv("APP_ENV"))

	config.NewConfiguration(env)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestID())
	e.Debug = config.IsDevelopment

	if config.IsDevelopment {
		docs.SwaggerInfo.Title = "LOS-KMB-API"
		docs.SwaggerInfo.Description = "This is a orchestrator api server."
		docs.SwaggerInfo.Version = "3.0"
		docs.SwaggerInfo.Host = os.Getenv("SWAGGER_HOST")
		docs.SwaggerInfo.BasePath = "/"
		docs.SwaggerInfo.Schemes = []string{"http", "https"}
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	} else {
		e.HideBanner = true
	}

	// Newrelic
	app, err := config.InitNewrelic()
	if err == nil {
		e.Use(nrecho.Middleware(app))
	}

	// LOS_KMB_BASE_URL
	constant.LOS_KMB_BASE_URL = os.Getenv("SWAGGER_HOST")
	if !strings.Contains(constant.LOS_KMB_BASE_URL, "/los-kmb-api") && !strings.Contains(constant.LOS_KMB_BASE_URL, ":") {
		constant.LOS_KMB_BASE_URL = constant.LOS_KMB_BASE_URL + ":" + os.Getenv("APP_PORT")
	}
	if !strings.Contains(constant.LOS_KMB_BASE_URL, "http") {
		constant.LOS_KMB_BASE_URL = "http://" + constant.LOS_KMB_BASE_URL
	}

	//topic kafka
	constant.TOPIC_SUBMISSION = os.Getenv("TOPIC_SUBMISSION")
	constant.TOPIC_SUBMISSION_LOS = os.Getenv("TOPIC_SUBMISSION_LOS")

	minilosWG, err := database.OpenMinilosWG()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	minilosKMB, err := database.OpenMinilosKMB()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	kpLos, err := database.OpenKpLos()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	kpLosLogs, err := database.OpenKpLosLogs()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	newKMB, err := database.OpenNewKmb()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	confins, err := database.OpenConfins()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	core, err := database.OpenCore()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	staging, err := database.OpenStaging()
	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	scorePro, err := database.OpenScorepro()
	if err != nil {
		log.Fatal(err)
	}

	var cache *bigcache.BigCache
	isCacheActive, _ := strconv.ParseBool(config.Env("CACHE_ACTIVE"))
	if isCacheActive {
		cacheExp, _ := strconv.Atoi(config.Env("CACHE_EXPIRED_DEFAULT"))
		cache, _ = bigcache.NewBigCache(bigcache.DefaultConfig(time.Duration(cacheExp) * time.Second))
	}

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowMethods: []string{echo.POST},
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "KREDITPLUS LOS-KMB-API")
	})

	accessToken := middlewares.NewAccessMiddleware()
	e.Use(accessToken.SetupHeadersAndContext())

	config.CreateCustomLogFile(constant.LOG_FILTERING_LOG)
	config.CreateCustomLogFile(constant.NEW_KMB_LOG)

	utils.NewCache(cache, kpLos, config.IsDevelopment)

	jsonResponse := json.NewResponse()
	authRepo := authRepository.NewRepository(newKMB)
	authorization := authorization.NewAuth(authRepo)
	apiGroup := e.Group("/api/v2/kmb")
	apiGroupv3 := e.Group("/api/v3/kmb")
	httpClient := httpclient.NewHttpClient()

	useLogPlatform, _ := strconv.ParseBool(os.Getenv("USE_LOG_PLATFORM"))
	if useLogPlatform {
		platformLog := platformlog.NewPlatformLog()
		platformlog.Log = platformLog
		platformLog.CreateLogger()
	}

	producer := platformevent.NewPlatformEvent()

	// define kmb filtering domain
	kmbFilteringRepo := filteringRepository.NewRepository(minilosKMB, kpLos, kpLosLogs)
	kmbFilteringMultiCase, kmbFilteringCase := filteringUsecase.NewMultiUsecase(kmbFilteringRepo, httpClient)
	filteringDelivery.FilteringHandler(apiGroup, kmbFilteringMultiCase, kmbFilteringCase, kmbFilteringRepo, jsonResponse, accessToken)

	// define kmb elaborate domain
	kmbElaborateRepo := elaborateRepository.NewRepository(minilosKMB, kpLos)
	kmbElaborateMultiCase, kmbElaborateCase := elaborateUsecase.NewMultiUsecase(kmbElaborateRepo, httpClient)
	elaborateDelivery.ElaborateHandler(apiGroup, kmbElaborateMultiCase, kmbElaborateCase, kmbElaborateRepo, jsonResponse, accessToken)

	// define new kmb filtering domain
	newKmbFilteringRepo := newKmbFilteringRepository.NewRepository(kpLos, kpLosLogs, newKMB)
	newKmbFilteringCase := newKmbFilteringUsecase.NewUsecase(newKmbFilteringRepo, httpClient)
	newKmbFilteringMultiCase := newKmbFilteringUsecase.NewMultiUsecase(newKmbFilteringRepo, httpClient, newKmbFilteringCase)
	newKmbFilteringDelivery.FilteringHandler(apiGroupv3, newKmbFilteringMultiCase, newKmbFilteringCase, newKmbFilteringRepo, jsonResponse, accessToken, producer)

	// define new kmb elaborate domain
	newElaborateLTVRepo := elaborateLTVRepository.NewRepository(kpLosLogs, newKMB)
	newElaborateLTVUsecase := elaborateLTVUsecase.NewUsecase(newElaborateLTVRepo, httpClient)
	elaborateLTVDelivery.ElaborateHandler(apiGroupv3, newElaborateLTVUsecase, newElaborateLTVRepo, authorization, jsonResponse, accessToken)

	// define new kmb cms
	cacheRepository := cacheRepository.NewRepository(cache)
	cmsRepositories := cmsRepository.NewRepository(core, confins, newKMB, kpLosLogs)
	cmsUsecases := cmsUsecase.NewUsecase(cmsRepositories, httpClient, cacheRepository)
	cmsDelivery.CMSHandler(apiGroupv3, cmsUsecases, cmsRepositories, jsonResponse, producer, accessToken)

	// define new kmb journey
	kmbRepositories := kmbRepository.NewRepository(kpLos, kpLosLogs, core, staging, minilosWG, minilosKMB, newKMB, scorePro)
	kmbUsecases := kmbUsecase.NewUsecase(kmbRepositories, httpClient)
	kmbMultiUsecases := kmbUsecase.NewMultiUsecase(kmbRepositories, httpClient, kmbUsecases)
	kmbMetrics := kmbUsecase.NewMetrics(kmbRepositories, httpClient, kmbUsecases, kmbMultiUsecases)
	kmbDelivery.KMBHandler(apiGroupv3, kmbMetrics, kmbUsecases, kmbRepositories, jsonResponse, accessToken, producer)

	toolsDelivery.ToolsHandler(apiGroupv3, jsonResponse, accessToken)

	auth := map[string]interface{}{
		"secret_key":         os.Getenv("PLATFORM_SECRET_KEY"),
		"source_application": constant.FLAG_LOS,
	}
	consumerRouter := platformevent.NewConsumerRouter(constant.TOPIC_SUBMISSION, os.Getenv("LOS_SUBMISSION_FILTERING"), auth)

	consumerRouter.Use(func(next event.ConsumerProcessor) event.ConsumerProcessor {
		return func(ctx context.Context, event event.Event) error {
			startTime := utils.GenerateTimeInMilisecond()
			reqID := utils.GenerateUUID()

			ctx = context.WithValue(ctx, constant.CTX_KEY_REQUEST_TIME, startTime)
			ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
			ctx = context.WithValue(ctx, constant.CTX_KEY_IS_CONSUMER, true)

			return next(ctx, event)
		}
	})

	eventhandlers.NewServiceFiltering(consumerRouter, newKmbFilteringRepo, newKmbFilteringCase, newKmbFilteringMultiCase, validator, producer, jsonResponse)

	if err := consumerRouter.StartConsume(); err != nil {
		panic(err)
	}

	consumerJourneyRouter := platformevent.NewConsumerRouter(constant.TOPIC_SUBMISSION_LOS, os.Getenv("LOS_SUBMISSION_KMB"), auth)

	consumerJourneyRouter.Use(func(next event.ConsumerProcessor) event.ConsumerProcessor {
		return func(ctx context.Context, event event.Event) error {
			startTime := utils.GenerateTimeInMilisecond()
			reqID := utils.GenerateUUID()

			ctx = context.WithValue(ctx, constant.CTX_KEY_REQUEST_TIME, startTime)
			ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
			ctx = context.WithValue(ctx, constant.CTX_KEY_IS_CONSUMER, true)

			return next(ctx, event)
		}
	})

	eventHandler.NewServiceKMB(consumerJourneyRouter, kmbRepositories, kmbUsecases, kmbMetrics, validator, producer, jsonResponse)

	if err := consumerJourneyRouter.StartConsume(); err != nil {
		panic(err)
	}

	// Setup Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("APP_PORT")),
		Handler:      e,
		WriteTimeout: 20 * time.Minute,
		ReadTimeout:  20 * time.Minute,
	}

	go func() {
		if err := e.StartServer(srv); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error start server %v \n", err)
		}
	}()

	// platform shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-quit:
		fmt.Println("========== Shutdown signal received ==========")

		if err := consumerRouter.StopConsume(); err != nil {
			panic(err)
		}

		fmt.Println("========== Shutdown Completed ==========")
		os.Exit(0)
	}
}
