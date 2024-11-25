package platformevent

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/KB-FMF/platform-library/event"
)

type EventMiddlewareFunc func(next event.ConsumerProcessor) event.ConsumerProcessor

type ConsumerRouter struct {
	consumerClient *event.Client
	routes         map[string]event.ConsumerProcessor
	middlewares    []EventMiddlewareFunc
	topic          []string
	consumerGroup  string
	auth           map[string]interface{}
}

func NewConsumerRouter(topic string, consumerGroup string, auth map[string]interface{}) *ConsumerRouter {
	appEnv := os.Getenv("APP_ENV")

	var brokerEnv string
	if strings.Contains(strings.ToLower(appEnv), "production") {
		brokerEnv = event.ENV_PRODUCTION
	} else if strings.Contains(strings.ToLower(appEnv), "staging") {
		brokerEnv = event.ENV_STAGING
	} else {
		brokerEnv = event.ENV_DEVELOPMENT
	}

	client := event.NewConsumer(brokerEnv)
	return &ConsumerRouter{
		consumerClient: client,
		routes:         map[string]event.ConsumerProcessor{},
		topic:          []string{topic},
		consumerGroup:  consumerGroup,
		auth:           auth,
	}
}

func (c *ConsumerRouter) Use(middlewareFuncs ...EventMiddlewareFunc) {
	c.middlewares = append(c.middlewares, middlewareFuncs...)
}

func (c *ConsumerRouter) Handle(key string, processorFunc event.ConsumerProcessor) {
	c.routes[key] = processorFunc
}

func (c *ConsumerRouter) StartConsume() error {
	err := c.consumerClient.StartConsume(c.topic, c.consumerGroup, c.auth, func(ctx context.Context, event event.Event) error {
		key := string(event.GetKey())
		key = strings.ReplaceAll(key, "\"", "")
		key = strings.ReplaceAll(key, "\\", "")
		re := regexp.MustCompile(`^(.+)_\d`)
		strSubmatch := re.FindStringSubmatch(key)

		if len(strSubmatch) > 0 {
			if val, ok := c.routes[strSubmatch[1]]; ok {
				processorFunc := applyMiddlewares(val, c.middlewares...)
				go processorFunc(ctx, event)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("consumer process error: %w", err)
	}

	return nil
}

func (c *ConsumerRouter) StopConsume() error {
	return c.consumerClient.CloseConsumer()
}

func applyMiddlewares(processorFunc event.ConsumerProcessor, middlewares ...EventMiddlewareFunc) event.ConsumerProcessor {
	for i := len(middlewares) - 1; i >= 0; i-- {
		processorFunc = middlewares[i](processorFunc)
	}

	return processorFunc
}
