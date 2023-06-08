package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/KB-FMF/platform-library"
	"github.com/KB-FMF/platform-library/auth"
	"github.com/KB-FMF/platform-library/utils"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-playground/validator/v10"
)

const ENV_DEVELOPMENT = "DEVELOPMENT"
const ENV_STAGING = "STAGING"
const ENV_PRODUCTION = "PRODUCTION"

type Logger struct {
	config Config

	client      *kafka.Producer
	authService *auth.Auth
}

func New(environment string) *Logger {
	var config Config
	if environment == "DEVELOPMENT" {
		config = newDevelopmentConfig()
	} else if environment == "STAGING" {
		config = newStagingConfig()
	} else if environment == "PRODUCTION" {
		config = newProductionConfig()
	} else {
		panic("no selected environment (DEVELOPMENT | STAGING | PRODUCTION)")
	}

	cfgMap := &kafka.ConfigMap{
		"bootstrap.servers":       config.bootstrapServers,
		"security.protocol":       config.securityProtocol,
		"sasl.mechanisms":         config.saslMechanisms,
		"sasl.username":           config.username,
		"sasl.password":           config.password,
		"go.delivery.reports":     config.deliveryReports,
		"go.events.channel.size":  config.eventsChannelSize,
		"go.produce.channel.size": config.produceChannelSize,
	}

	producer, err := kafka.NewProducer(cfgMap)
	if err != nil {
		panic(err)
	}

	return &Logger{client: producer, config: config, authService: auth.New(environment)}
}

func (l *Logger) Log(token string, body map[string]interface{}) *platform.Error {
	bodyByte, errInternal := json.Marshal(body)
	if errInternal != nil {
		return platform.FromGoErr(errInternal)
	}

	var requests Request
	if err := json.Unmarshal(bodyByte, &requests); err != nil {
		return platform.FromGoErr(err)
	}

	if err := validate.Struct(requests); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var errs string
			for _, fe := range ve {
				if ok, errStr := getTag(fe); ok {
					errs += fmt.Sprintf("\n %s", errStr)

				}
			}
			return platform.FromGoErr(fmt.Errorf(errs))
		}
	}

	return l.logProduce(token, "log", body)
}

func (l *Logger) LogGeneral(token string, body map[string]interface{}) *platform.Error {
	bodyByte, errInternal := json.Marshal(body)
	if errInternal != nil {
		return platform.FromGoErr(errInternal)
	}

	var requests PayloadLogGeneral
	if err := json.Unmarshal(bodyByte, &requests); err != nil {
		return platform.FromGoErr(err)
	}

	if err := validate.Struct(requests); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			var errs string
			for _, fe := range ve {
				if ok, errStr := getTag(fe); ok {
					errs += fmt.Sprintf("\n %s", errStr)

				}
			}
			return platform.FromGoErr(fmt.Errorf(errs))
		}
	}

	return l.logProduce(token, "log_general", body)
}

func (l *Logger) Flush(timeout int) int {
	return l.client.Flush(timeout)
}

func (l *Logger) logProduce(token string, logType string, body map[string]interface{}) *platform.Error {
	validation, err := l.authService.Validation(token, l.config.applicationName)
	if err != nil {
		return err
	}

	appName, ok := validation.Data["name"]
	if !ok {
		return platform.FromGoErr(fmt.Errorf("invalid token"))
	}

	if body != nil {
		body["client_token"] = token
		body["source_application"] = appName
	}

	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(body); err != nil {
		return platform.FromGoErr(err)
	}

	uniqueID := utils.ComputeHmac("logApp_key", b.String())
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &l.config.topics[0],
			Partition: kafka.PartitionAny,
		},
		Value: b.Bytes(),
		Key:   []byte(fmt.Sprintf("%s/%s", logType, uniqueID)),
	}

	if err := l.client.Produce(message, nil); err != nil {
		return platform.FromGoErr(err)
	}

	return nil
}
