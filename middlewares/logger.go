package middlewares

import (
	"time"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/env"
	log "github.com/sirupsen/logrus"
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
		c.SetLog(entry)
		t := time.Now()
		c.Next()
		fields := c.Log().WithFields(log.Fields{
			"routing_key": c.Request.RoutingKey,
			"duration":    time.Since(t).String(),
		})
		if c.Error != nil {
			fields.Errorf("Request aborted with error: %s", c.Error)
			return
		}
		fields.Info("Request completed")
	}
}
