package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	docs "los-kmb-api/docs"
	delivery "los-kmb-api/domain/filtering/delivery/http"
	"los-kmb-api/domain/filtering/repository"
	"los-kmb-api/domain/filtering/usecase"

	middlewares "los-kmb-api/middlewares"
	"los-kmb-api/shared/common"
	"los-kmb-api/shared/common/json"
	"los-kmb-api/shared/config"
	"los-kmb-api/shared/database"
	"los-kmb-api/shared/httpclient"
	"los-kmb-api/shared/utils"

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

	kmbFiltering, err := database.OpenDatabase()

	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	dummykmbFiltering, err := database.OpenDummyDatabase()

	if err != nil {
		panic(fmt.Sprintf("Failed to open database connection: %s", err))
	}

	kpLos, err := database.OpenDatabaseKpLos()

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
		return c.String(http.StatusOK, "KREDITPLUS INTEGERATOR INCOME PREDICTION")
	})

	accessToken := middlewares.NewAccessMiddleware()

	config.CreateCustomLogFile("API_INT_KMB_FILTERING_LOG")

	utils.NewCache(cache, kpLos, config.IsDevelopment)

	// define kmb filtering domain
	kmbFilteringRepo := repository.NewRepository(kmbFiltering, kpLos, dummykmbFiltering)
	kmbFilteringHttpClient := httpclient.NewHttpClient()
	kmbFilteringMultiCase, kmbFilteringCase := usecase.NewMultiUsecase(kmbFilteringRepo, kmbFilteringHttpClient)
	kmbFilteringJson := json.NewResponse()
	kmbFilteringkGroup := e.Group("/api/v2/kmb")
	delivery.FilteringHandler(kmbFilteringkGroup, kmbFilteringMultiCase, kmbFilteringCase, kmbFilteringRepo, kmbFilteringJson, accessToken)

	if config.IsDevelopment {
		docs.SwaggerInfo.Title = "LOS-KMB-API"
		docs.SwaggerInfo.Description = "This is a los kmb api server."
		docs.SwaggerInfo.Version = "2.0"
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", config.Env("APP_HOST"), config.Env("APP_PORT"))
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
	e.Start(fmt.Sprintf(":%s", config.Env("APP_PORT")))
}
