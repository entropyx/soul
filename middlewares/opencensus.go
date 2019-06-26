package middlewares

import (
	"github.com/entropyx/opencensus-propagation"
	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/tracers"
	"go.opencensus.io/trace"
)

type openCensuskey uint

const (
	keyOpenCensusSpan openCensuskey = iota
)

func OpenCensus() context.Handler {
	return func(c *context.Context) {
		tracer := tracers.GlobalTracer()

		spanCtx, _ := propagation.Extract(propagation.FormatTextMap, c.Request.Headers)

		_, span := trace.StartSpanWithRemoteParent(c, "new-request", spanCtx, trace.WithSpanKind(trace.SpanKindServer))
		defer span.End()

		propagation.Inject(span.SpanContext(), propagation.FormatTextMap, c.Headers)
		fields := tracer.LogFields(c.Headers)

		c.SetLog(c.Log().WithFields(fields))
		c.Set("span", span)

		c.Next()

		if err := c.Error; err != nil {
			tracer.SetErrorTag(span, err)
		}
	}
}
