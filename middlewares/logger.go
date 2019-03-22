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
		t := time.Now()
		fields := entry.WithField("routing_key", c.Request.RoutingKey)
		c.SetLog(fields)
		c.Next()
		durationField := c.Log().WithField("duration", time.Since(t).String())
		if c.Error != nil {
			errorField := durationField.WithField("error.message", c.Error.Error())
			errorField.Error("Request aborted with error")
			return
		}
		durationField.Info("Request completed")
	}
}
