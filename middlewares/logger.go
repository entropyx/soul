package middlewares

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/soul-go/soul/context"
	"github.com/soul-go/soul/env"
)

type LoggerOptions struct {
	Formatter log.Formatter
	Hook      log.Hook
}

func Logger(options *LoggerOptions) context.Handler {
	logger := log.New()
	if hook := options.Hook; hook != nil {
		logger.AddHook(hook)
	}
	if formatter := options.Formatter; formatter != nil {
		logger.SetFormatter(formatter)
	}
	switch env.Mode {
	case env.ModeDebug:
		logger.SetLevel(log.DebugLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}
	entry := log.NewEntry(logger)

	return func(c *context.Context) {
		c.Log = entry
		t := time.Now()
		c.Next()
		fields := c.Log.WithFields(log.Fields{
			"routing_key": c.Request.RoutingKey,
			"duration":    time.Since(t),
		})
		if c.Err != nil {
			fields.Errorf("Request aborted with error: %s", c.Err)
			return
		}
		fields.Info("Request completed")
	}
}