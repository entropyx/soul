package middlewares

import (
	"fmt"
	"os"

	"github.com/entropyx/opencensus-propagation"
	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/env"
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

		_, span := trace.StartSpanWithRemoteParent(c, fmt.Sprintf("%s new request", os.Getenv("SERVICE_NAME")), spanCtx, trace.WithSpanKind(trace.SpanKindServer), trace.WithSampler(setSampler()))
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

func setSampler() trace.Sampler {
	switch env.Mode {
	case env.ModeProduction, env.ModeStaging:
		return trace.AlwaysSample()
	case env.ModeTest, env.ModeDebug:
		return trace.NeverSample()
	default:
		return trace.AlwaysSample()
	}
}
