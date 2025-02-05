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
	elaborateLTVDelivery "los-kmb-api/domain/elaborate_ltv/delivery/http"
	elaborateLTVRepository "los-kmb-api/domain/elaborate_ltv/repository"
	elaborateLTVUsecase "los-kmb-api/domain/elaborate_ltv/usecase"
	eventhandlers "los-kmb-api/domain/filtering_new/delivery/event"
	newKmbFilteringDelivery "los-kmb-api/domain/filtering_new/delivery/http"
	newKmbFilteringRepository "los-kmb-api/domain/filtering_new/repository"
	newKmbFilteringUsecase "los-kmb-api/domain/filtering_new/usecase"
	eventHandler "los-kmb-api/domain/kmb/delivery/event"
	kmbDelivery "los-kmb-api/domain/kmb/delivery/http"
	kmbRepository "los-kmb-api/domain/kmb/repository"
	kmbUsecase "los-kmb-api/domain/kmb/usecase"
	eventPrincipleHandler "los-kmb-api/domain/principle/delivery/event"
	principleDelivery "los-kmb-api/domain/principle/delivery/http"
	principleRepository "los-kmb-api/domain/principle/repository"
	principleUsecase "los-kmb-api/domain/principle/usecase"
	toolsDelivery "los-kmb-api/domain/tools/delivery/http"
	"los-kmb-api/middlewares"
	"los-kmb-api/shared/authorization"
	authRepository "los-kmb-api/shared/authorization/repository"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/json"
	"los-kmb-api/shared/common/platformcache"
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

	"github.com/KB-FMF/los-common-library/loslog"
	"github.com/KB-FMF/los-common-library/platform/manager"

	"los-kmb-api/shared/common/platformevent"

	"github.com/KB-FMF/los-common-library/response"
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
	e.Use(middleware.Secure())
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
	constant.TOPIC_INSERT_CUSTOMER = os.Getenv("TOPIC_INSERT_CUSTOMER")
	constant.TOPIC_SUBMISSION_PRINCIPLE = os.Getenv("TOPIC_SUBMISSION_PRINCIPLE")
	constant.TOPIC_SUBMISSION_2WILEN = os.Getenv("TOPIC_SUBMISSION_2WILEN")

	//Platform Event key
	constant.KEY_PREFIX_FILTERING = os.Getenv("KEY_PREFIX_FILTERING")
	constant.KEY_PREFIX_UPDATE_STATUS_FILTERING = os.Getenv("KEY_PREFIX_UPDATE_STATUS_FILTERING")
	constant.KEY_PREFIX_SUBMIT_TO_LOS = os.Getenv("KEY_PREFIX_SUBMIT_TO_LOS")
	constant.KEY_PREFIX_AFTER_PRESCREENING = os.Getenv("KEY_PREFIX_AFTER_PRESCREENING")
	constant.KEY_PREFIX_CALLBACK = os.Getenv("KEY_PREFIX_CALLBACK")
	constant.KEY_PREFIX_CALLBACK_GOLIVE = os.Getenv("KEY_PREFIX_CALLBACK_GOLIVE")
	constant.KEY_PREFIX_UPDATE_CUSTOMER = os.Getenv("KEY_PREFIX_UPDATE_CUSTOMER")
	constant.KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE = os.Getenv("KEY_PREFIX_UPDATE_TRANSACTION_PRINCIPLE")
	constant.KEY_PREFIX_CANCEL_ORDER_2WILEN = os.Getenv("KEY_PREFIX_CANCEL_ORDER_2WILEN")

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

	e.Use(middleware.BodyDumpWithConfig(middlewares.NewBodyDumpMiddleware(newKMB).BodyDumpConfig()))

	common.SetDB(newKMB)

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
	config.CreateCustomLogFile(constant.DILEN_KMB_LOG)

	utils.NewCache(cache, kpLos, config.IsDevelopment)

	jsonResponse := json.NewResponse()
	authRepo := authRepository.NewRepository(newKMB)
	authorization := authorization.NewAuth(authRepo)
	apiGroupv3 := e.Group("/api/v3/kmb")
	httpClient := httpclient.NewHttpClient()

	useLogPlatform, _ := strconv.ParseBool(os.Getenv("USE_LOG_PLATFORM"))
	if useLogPlatform {
		platformLog := platformlog.NewPlatformLog()
		platformlog.Log = platformLog
		platformLog.CreateLogger()
	}

	// init producer topic submission
	producerSubmission, err := config.ProducerEvent(constant.TOPIC_SUBMISSION, 3)
	if err != nil {
		log.Fatalf("Failed Init Producer event %s with Error : %s", constant.TOPIC_SUBMISSION, err.Error())
	}

	// init producer topic submission-los
	producerSubmissionLOS, err := config.ProducerEvent(constant.TOPIC_SUBMISSION_LOS, 3)
	if err != nil {
		log.Fatalf("Failed Init Producer event %s with Error : %s", constant.TOPIC_SUBMISSION_LOS, err.Error())
	}

	// init producer topic insert-customer
	producerInsertCustomer, err := config.ProducerEvent(constant.TOPIC_INSERT_CUSTOMER, 3)
	if err != nil {
		log.Fatalf("Failed Init Producer event %s with Error : %s", constant.TOPIC_INSERT_CUSTOMER, err.Error())
	}

	producer := platformevent.NewPlatformEvent(producerSubmission, producerSubmissionLOS, producerInsertCustomer)
	platformCache := platformcache.NewPlatformCache()

	// define new kmb filtering domain
	newKmbFilteringRepo := newKmbFilteringRepository.NewRepository(kpLos, kpLosLogs, newKMB)
	newKmbFilteringCase := newKmbFilteringUsecase.NewUsecase(newKmbFilteringRepo, httpClient)
	newKmbFilteringMultiCase := newKmbFilteringUsecase.NewMultiUsecase(newKmbFilteringRepo, httpClient, newKmbFilteringCase)
	newKmbFilteringDelivery.FilteringHandler(apiGroupv3, newKmbFilteringMultiCase, newKmbFilteringCase, newKmbFilteringRepo, jsonResponse, accessToken, producer, platformCache)

	// define new kmb elaborate domain
	newElaborateLTVRepo := elaborateLTVRepository.NewRepository(kpLos, kpLosLogs, newKMB)
	newElaborateLTVUsecase := elaborateLTVUsecase.NewUsecase(newElaborateLTVRepo, httpClient)
	elaborateLTVDelivery.ElaborateHandler(apiGroupv3, newElaborateLTVUsecase, newElaborateLTVRepo, authorization, jsonResponse, accessToken)

	// define new kmb cms
	cacheRepository := cacheRepository.NewRepository(cache)
	cmsRepositories := cmsRepository.NewRepository(core, confins, newKMB, kpLos, kpLosLogs)
	cmsUsecases := cmsUsecase.NewUsecase(cmsRepositories, httpClient, cacheRepository)
	cmsDelivery.CMSHandler(apiGroupv3, cmsUsecases, cmsRepositories, jsonResponse, producer, accessToken)

	// define new kmb journey
	kmbRepositories := kmbRepository.NewRepository(kpLos, kpLosLogs, core, staging, newKMB, scorePro)
	kmbUsecases := kmbUsecase.NewUsecase(kmbRepositories, httpClient)
	kmbMultiUsecases := kmbUsecase.NewMultiUsecase(kmbRepositories, httpClient, kmbUsecases)
	kmbMetrics := kmbUsecase.NewMetrics(kmbRepositories, httpClient, kmbUsecases, kmbMultiUsecases)
	kmbDelivery.KMBHandler(apiGroupv3, kmbMetrics, kmbUsecases, kmbRepositories, authorization, jsonResponse, accessToken, producer)

	managers := manager.New(platformlog.GetPlatformEnv(), os.Getenv("PLATFORM_SECRET_KEY"), os.Getenv("PLATFORM_AUTH_BASE_URL")+"/v1/auth/login")

	libLog := loslog.NewConfig(
		"Orchestrator-kmb",
		managers,
		loslog.WithHookPlatform(true),
	)

	defer func() {
		_ = libLog.Sync()
	}()

	libResponse := response.NewResponse(os.Getenv("APP_PREFIX_NAME"), response.WithDebug(true))
	// libTrace := tracer.Initialize(os.Getenv("APP_NAME"), tracer.IsEnable(config.IsDebug), tracer.LicenseKey(os.Getenv("NEWRELIC_CONFIG_LICENSE")))

	// losLog
	logMiddleware := loslog.New(libLog)
	e.Use(logMiddleware.Log)

	principleRepo := principleRepository.NewRepository(newKMB, kpLos, scorePro, confins)
	principleCase := principleUsecase.NewUsecase(principleRepo, httpClient, producer)
	principleMultiCase := principleUsecase.NewMultiUsecase(principleRepo, httpClient, producer, principleCase)
	principleDelivery.Handler(apiGroupv3, principleMultiCase, principleCase, principleRepo, libResponse, accessToken)

	toolsDelivery.ToolsHandler(apiGroupv3, jsonResponse, accessToken, producer)

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

	eventhandlers.NewServiceFiltering(consumerRouter, newKmbFilteringRepo, newKmbFilteringCase, newKmbFilteringMultiCase, validator, producer, jsonResponse, platformCache)

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

	eventHandler.NewServiceKMB(consumerJourneyRouter, kmbRepositories, kmbUsecases, kmbMetrics, validator, producer, jsonResponse, cmsUsecases)

	if err := consumerJourneyRouter.StartConsume(); err != nil {
		panic(err)
	}

	consumerPrincipleRouter := platformevent.NewConsumerRouter(constant.TOPIC_SUBMISSION_PRINCIPLE, os.Getenv("LOS_SUBMISSION_PRINCIPLE"), auth)

	consumerPrincipleRouter.Use(func(next event.ConsumerProcessor) event.ConsumerProcessor {
		return func(ctx context.Context, event event.Event) error {
			startTime := utils.GenerateTimeInMilisecond()
			reqID := utils.GenerateUUID()

			ctx = context.WithValue(ctx, constant.CTX_KEY_REQUEST_TIME, startTime)
			ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
			ctx = context.WithValue(ctx, constant.CTX_KEY_IS_CONSUMER, true)

			return next(ctx, event)
		}
	})

	eventPrincipleHandler.NewServicePrinciple(consumerPrincipleRouter, principleRepo, principleCase, validator, producer, jsonResponse)

	if err := consumerPrincipleRouter.StartConsumeWithoutTimestamp(); err != nil {
		panic(err)
	}

	consumer2WilenRouter := platformevent.NewConsumerRouter(constant.TOPIC_SUBMISSION_2WILEN, os.Getenv("LOS_SUBMISSION_PRINCIPLE"), auth)

	consumer2WilenRouter.Use(func(next event.ConsumerProcessor) event.ConsumerProcessor {
		return func(ctx context.Context, event event.Event) error {
			startTime := utils.GenerateTimeInMilisecond()
			reqID := utils.GenerateUUID()

			ctx = context.WithValue(ctx, constant.CTX_KEY_REQUEST_TIME, startTime)
			ctx = context.WithValue(ctx, constant.HeaderXRequestID, reqID)
			ctx = context.WithValue(ctx, constant.CTX_KEY_IS_CONSUMER, true)

			return next(ctx, event)
		}
	})

	eventPrincipleHandler.NewService2Wilen(consumer2WilenRouter, principleRepo, principleCase, validator, producer, jsonResponse)

	if err := consumer2WilenRouter.StartConsumeProspectIDWithoutTimestamp(); err != nil {
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

		if err := consumerJourneyRouter.StopConsume(); err != nil {
			panic(err)
		}

		if err := consumerPrincipleRouter.StopConsume(); err != nil {
			panic(err)
		}

		if err := consumer2WilenRouter.StopConsume(); err != nil {
			panic(err)
		}

		// close producer topic submission
		if err := producerSubmission.CloseProducer(); err != nil {
			panic(err)
		}

		// close producer topic submission-los
		if err := producerSubmissionLOS.CloseProducer(); err != nil {
			panic(err)
		}

		// close producer topic insert-customer
		if err := producerInsertCustomer.CloseProducer(); err != nil {
			panic(err)
		}

		fmt.Println("========== Shutdown Completed ==========")
		os.Exit(0)
	}
}
