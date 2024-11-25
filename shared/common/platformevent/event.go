package platformevent

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

import (
	"context"
	"fmt"
	"os"
	"time"

	"los-kmb-api/shared/common"
	"los-kmb-api/shared/constant"
	"los-kmb-api/shared/utils"

	"github.com/KB-FMF/platform-library/event"
)

type platformEvent struct {
	producerSubmission     *event.Client
	producerSubmissionLOS  *event.Client
	producerInsertCustomer *event.Client
}

//counterfeiter:generate . PlatformEventInterface
type PlatformEventInterface interface {
	PublishEvent(ctx context.Context, accessToken, topicName, key, id string, value map[string]interface{}, countRetry int) error
}

func NewPlatformEvent(producerSubmission, producerSubmissionLOS, producerInsertCustomer *event.Client) PlatformEventInterface {
	return &platformEvent{producerSubmission, producerSubmissionLOS, producerInsertCustomer}
}

func (pe platformEvent) PublishEvent(ctx context.Context, accessToken, topicName, key, id string, value map[string]interface{}, countRetry int) error {
	var (
		err      error
		producer *event.Client
	)

	keyMessage := fmt.Sprintf("%v_%v_%v", key, utils.GenerateUnixTimeNow(), id)

	value["topic_key"] = keyMessage
	value["topic_name"] = topicName

	switch topicName {
	case constant.TOPIC_SUBMISSION:
		producer = pe.producerSubmission
	case constant.TOPIC_SUBMISSION_LOS:
		producer = pe.producerSubmissionLOS
	case constant.TOPIC_INSERT_CUSTOMER:
		producer = pe.producerInsertCustomer
	default:
		err = fmt.Errorf("producer for topic %s was not created", topicName)

		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       os.Getenv("DUMMY_URL_LOGS"),
			Action:     "PUBLISH_EVENT",
			Type:       "EVENT_PLATFORM_LIBRARY",
			LogFile:    constant.NEW_KMB_LOG,
			MsgLogFile: constant.MSG_PUBLISH_DATA_STREAM,
			LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
			Request:    value,
			Response:   map[string]interface{}{"errors": err.Error()},
		})
	}

	if err != nil {
		return err
	}

	err = producer.Publish(accessToken, keyMessage, value)
	if err != nil {

		// Write Error Log
		common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
			Link:       os.Getenv("DUMMY_URL_LOGS"),
			Action:     "PUBLISH_EVENT",
			Type:       "EVENT_PLATFORM_LIBRARY",
			LogFile:    constant.NEW_KMB_LOG,
			MsgLogFile: constant.MSG_PUBLISH_DATA_STREAM,
			LevelLog:   constant.PLATFORM_LOG_LEVEL_ERROR,
			Request:    value,
			Response:   map[string]interface{}{"errors": err.Error()},
		})

		if countRetry < constant.MAX_RETRY_PUBLISH {
			countRetry = countRetry + 1
			time.Sleep(time.Second * time.Duration(countRetry*10))
			err = pe.PublishEvent(ctx, accessToken, topicName, key, id, value, countRetry)
			return err
		} else {
			return err
		}

	}

	// Write Success Log
	common.CentralizeLog(ctx, accessToken, common.CentralizeLogParameter{
		Link:       os.Getenv("DUMMY_URL_LOGS"),
		Action:     "PUBLISH_EVENT",
		Type:       "EVENT_PLATFORM_LIBRARY",
		LogFile:    constant.NEW_KMB_LOG,
		MsgLogFile: constant.MSG_PUBLISH_DATA_STREAM,
		LevelLog:   constant.PLATFORM_LOG_LEVEL_INFO,
		Request:    value,
		Response: map[string]interface{}{
			"messages": "success publish data stream",
		},
	})

	return err
}
