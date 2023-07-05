package main

import (
	"context"
	"fmt"
	"log"
	"los-kmb-api/docs"
	elaborateDelivery "los-kmb-api/domain/elaborate/delivery/http"
	elaborateRepository "los-kmb-api/domain/elaborate/repository"
	elaborateUsecase "los-kmb-api/domain/elaborate/usecase"
	filteringDelivery "los-kmb-api/domain/filtering/delivery/http"
	filteringRepository "los-kmb-api/domain/filtering/repository"
	filteringUsecase "los-kmb-api/domain/filtering/usecase"
	"los-kmb-api/middlewares"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/json"
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
	"time"

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

	config.LoadEnv()

	env := strings.ToLower(config.Env("APP_ENV"))

	config.NewConfiguration(env)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.RequestID())
	e.Debug = config.IsDevelopment

	constant.LOS_KMB_BASE_URL = os.Getenv("SWAGGER_HOST")

	minilosKMB, err := database.OpenMinilosKMB()

	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	dummy, err := database.OpenKpLosLog()

	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	kpLos, err := database.OpenKpLos()

	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
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

	config.CreateCustomLogFile("FILTERING_LOG")

	utils.NewCache(cache, kpLos, config.IsDevelopment)

	jsonResponse := json.NewResponse()
	apiGroup := e.Group("/api/v2/kmb")
	httpClient := httpclient.NewHttpClient()

	// define kmb filtering domain
	kmbFilteringRepo := filteringRepository.NewRepository(minilosKMB, kpLos, dummy)
	kmbFilteringMultiCase, kmbFilteringCase := filteringUsecase.NewMultiUsecase(kmbFilteringRepo, httpClient)
	filteringDelivery.FilteringHandler(apiGroup, kmbFilteringMultiCase, kmbFilteringCase, kmbFilteringRepo, jsonResponse, accessToken)

	// define kmb elaborate domain
	kmbElaborateRepo := elaborateRepository.NewRepository(minilosKMB, kpLos)
	kmbElaborateMultiCase, kmbElaborateCase := elaborateUsecase.NewMultiUsecase(kmbElaborateRepo, httpClient)
	elaborateDelivery.ElaborateHandler(apiGroup, kmbElaborateMultiCase, kmbElaborateCase, kmbElaborateRepo, jsonResponse, accessToken)

	if config.IsDevelopment {
		docs.SwaggerInfo.Title = "LOS-KMB-API"
		docs.SwaggerInfo.Description = "This is a orchestrator api server."
		docs.SwaggerInfo.Version = "1.0"
		docs.SwaggerInfo.Host = os.Getenv("SWAGGER_HOST")
		docs.SwaggerInfo.BasePath = "/api/v2/kmb"
		docs.SwaggerInfo.Schemes = []string{"http", "https"}
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	} else {
		e.HideBanner = true

		// Newrelic
		app, err := config.InitNewrelic()
		if err == nil {
			e.Use(nrecho.Middleware(app))
		}
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

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
