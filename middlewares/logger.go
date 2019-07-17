package middlewares

import (
	"time"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/log"
)

func Logger(logger log.Logger) context.Handler {
	return func(c *context.Context) {
		c.SetLogger(logger.WithField("routing_key", c.Request.RoutingKey))
		t := time.Now()
		c.Log().Info("Incoming request")
		c.Next()
		withDuration := c.Log().WithField("duration", time.Since(t).String())
		if c.Error != nil {
			withDuration.WithField("error.message", c.Error.Error()).Error("Request aborted with error")
			return
		}
		withDuration.Info("Request completed")
	}
}
