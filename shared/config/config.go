package config

import (
	"io"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	logger "github.com/labstack/gommon/log"
	"github.com/newrelic/go-agent/v3/newrelic"
	log "github.com/sirupsen/logrus"
)

var (
	DateLogFile   map[string]string
	GetLogFile    map[string]*os.File
	IsDevelopment bool
)

func LoadEnv() {
	err := godotenv.Load("conf/config.env")
	if err != nil {
		log.Fatal("Error loading env file")
	}
}

func SetupTimezone() {
	location, _ := time.LoadLocation("Asia/Jakarta")
	time.Local = location
}

func NewConfiguration(appEnv string) {

	if strings.ToLower(appEnv) != "prod" && strings.ToLower(appEnv) != "production" {
		IsDevelopment = true
	} else {
		IsDevelopment = false
	}

	GetLogFile = make(map[string]*os.File)
	DateLogFile = make(map[string]string)

}

func Env(key string) string {
	env, err := godotenv.Read("conf/config.env")
	if err != nil {
		logger.Fatalf("Error %v", err)
	}
	v := env[key]
	return v
}

func InitNewrelic() (*newrelic.Application, error) {
	newrelicActive, _ := strconv.ParseBool(Env("NEWRELIC_ACTIVE"))
	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(Env("APP_NAME")),
		newrelic.ConfigLicense(Env("NEWRELIC_CONFIG_LICENSE")),
		newrelic.ConfigDistributedTracerEnabled(true),
		func(addConfig *newrelic.Config) {
			addConfig.Enabled = newrelicActive
			addConfig.Labels = map[string]string{
				"Env": strings.ToLower(Env("APP_ENV")),
				"Tag": strings.ToLower(Env("APP_VERSION")),
			}
		},
	)
	return app, err
}

func CreateCustomLogFile(keyConfig string) {

	loc, _ := time.LoadLocation("Asia/Jakarta")
	currentTime := time.Now().In(loc)

	// create folder general
	logPath := Env("LOG_FILE") + keyConfig + "/"

	active, _ := strconv.ParseBool(Env(keyConfig))
	if logPath != "" && active {
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			err = os.MkdirAll(logPath, 0755)
			if err != nil {
				panic(err)
			}
		}

		logFileName := strings.ToLower(keyConfig) + "-" + currentTime.Format(constant.TIME_FORMAT) + ".log"
		logFile, err := os.OpenFile(logPath+logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger.Fatalf("error opening file: %v", err)
		}
		GetLogFile[keyConfig] = logFile
		DateLogFile[keyConfig] = currentTime.Format(constant.TIME_FORMAT)
	}
}

func SetCustomLog(keyConfig string, isError bool, data map[string]interface{}, msg string) {

	loc, _ := time.LoadLocation("Asia/Jakarta")

	dateNow := time.Now().In(loc).Format(constant.TIME_FORMAT)
	if DateLogFile[keyConfig] != dateNow {
		CreateCustomLogFile(keyConfig)
	}
	logPath := Env("LOG_FILE")
	active, _ := strconv.ParseBool(Env(keyConfig))
	if logPath != "" && active {

		logFile := GetLogFile[keyConfig]

		log.SetOutput(io.MultiWriter(logFile, os.Stdout))
		log.SetFormatter(&log.JSONFormatter{})
		if isError {
			log.WithFields(data).Error(msg)
			return
		}
		log.WithFields(data).Info(msg)
		return
	}
}

func GetMinilosWgDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("MINILOS_WG_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("MINILOS_WG_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("MINILOS_WG_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("MINILOS_WG_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("MINILOS_WG_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetMinilosKmbDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("MINILOS_KMB_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("MINILOS_KMB_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("MINILOS_KMB_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("MINILOS_KMB_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("MINILOS_KMB_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetKpLosDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("KP_LOS_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("KP_LOS_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("KP_LOS_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("KP_LOS_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("KP_LOS_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetKpLosLogDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("KP_LOS_LOG_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("KP_LOS_LOG_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("KP_LOS_LOG_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("KP_LOS_LOG_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("KP_LOS_LOG_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetNewKmbDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("NEW_KMB_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("NEW_KMB_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("NEW_KMB_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("NEW_KMB_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("NEW_KMB_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetCoreDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("CORE_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetConfinsDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("CONFINS_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetStagingDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("STAGING_DB_USERNAME"))
	pwd, _ := utils.DecryptCredential(os.Getenv("STAGING_DB_PASSWORD"))
	host, _ := utils.DecryptCredential(os.Getenv("STAGING_DB_HOST"))
	strPort, _ := utils.DecryptCredential(os.Getenv("STAGING_DB_PORT"))
	port, _ := strconv.Atoi(strPort)
	database, _ := utils.DecryptCredential(os.Getenv("STAGING_DB_DATABASE"))

	return user, pwd, host, port, database
}

func GetScoreProDB() (string, string, string, int, string) {
	user, _ := utils.DecryptCredential(os.Getenv("DB_SCOREPRO_USERNAME"))
	strPort, _ := utils.DecryptCredential(os.Getenv("DB_SCOREPRO_PORT"))
	port, _ := strconv.Atoi(strPort)
	host, _ := utils.DecryptCredential(os.Getenv("DB_SCOREPRO_HOST"))
	pwd, _ := utils.DecryptCredential(os.Getenv("DB_SCOREPRO_PASSWORD"))
	database, _ := utils.DecryptCredential(os.Getenv("DB_SCOREPRO_DATABASE"))

	return user, pwd, host, port, database
}
