package middlewares

import (
	"time"

	"github.com/entropyx/errors"
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
			errorFields := fields.WithField("error.message", c.Error.Error())
			err, ok := c.Error.(errors.Error)
			if ok {
				errorFields = fields.WithFields(log.Fields{"error.stack": err.StackTrace, "error.kind": err.Code})
			}
			errorFields.Error("Request aborted with error")
			return
		}
		fields.Info("Request completed")
	}
}
