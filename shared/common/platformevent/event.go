package platformevent

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"
	"os"
	"strings"

	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/KB-FMF/platform-library/event"
)

type PlatformEvent struct{}

//counterfeiter:generate . PlatformEventInterface
type PlatformEventInterface interface {
	PublishEvent(ctx context.Context, accessToken, topicName, key, id string, value map[string]interface{}, countRetry int) error
}

func NewPlatformEvent() PlatformEvent {
	return PlatformEvent{}
}

func (pe PlatformEvent) PublishEvent(ctx context.Context, accessToken, topicName, key, id string, value map[string]interface{}, countRetry int) error {
	var logEnv string

	env := os.Getenv("APP_ENV")

	if strings.Contains(strings.ToLower(env), "production") {
		logEnv = event.ENV_PRODUCTION
	} else if strings.Contains(strings.ToLower(env), "staging") {
		logEnv = event.ENV_STAGING
	} else {
		logEnv = event.ENV_DEVELOPMENT
	}

	timestamp := utils.GenerateUnixTimeNow()
	keyMessage := fmt.Sprintf("%v_%v_%v", key, timestamp, id)

	config := event.ProducerConfig{Topic: topicName}
	producer, errCreateProducer := event.NewProducer(logEnv, config)
	if errCreateProducer != nil {
		return errCreateProducer
	}

	//don't forget to close producer
	defer producer.CloseProducer()

	errPublish := producer.Publish(accessToken, keyMessage, value)

	value["topic_key"] = keyMessage
	value["topic_name"] = topicName

	if errPublish != nil {

		// Write Error Log
		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       os.Getenv("DUMMY_URL_LOGS"),
			Action:     "PUBLISH_EVENT",
			Type:       "EVENT_PLATFORM_LIBRARY",
			LogFile:    constant.LOG_EVENT,
			MsgLogFile: constant.MSG_PUBLISH_DATA_STREAM,
			LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
			Request:    value,
			Response:   map[string]interface{}{"errors": errPublish.Error()},
		})

		if countRetry < constant.MAX_RETRY_PUBLISH {
			countRetry = countRetry + 1
			pe.PublishEvent(ctx, accessToken, topicName, key, id, value, countRetry)
			return nil
		} else {
			return errPublish
		}

	}

	// Write Success Log
	common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Action:     "PUBLISH_EVENT",
		Type:       "EVENT_PLATFORM_LIBRARY",
		LogFile:    constant.LOG_EVENT,
		MsgLogFile: constant.MSG_PUBLISH_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    value,
		Response: map[string]interface{}{
			"messages": "success publish data stream",
		},
	})

	return nil
}
