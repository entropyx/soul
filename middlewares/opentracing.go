package middlewares

import (
	"os"

	"github.com/entropyx/dd-trace-go/ddtrace/opentracer"
	ddtracer "github.com/entropyx/dd-trace-go/ddtrace/tracer"
	opentracing "github.com/opentracing/opentracing-go"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/env"
	"github.com/entropyx/soul/tracers"
)

const (
	datadogTraceHeaderName  = "trace_id"
	datadogParentHeaderName = "parent_id"
)

func ConfigureDatadog(service string) (opentracing.Tracer, error) {
	cfg := &ddtracer.PropagatorConfig{
		TraceHeader:  datadogTraceHeaderName,
		ParentHeader: datadogParentHeaderName,
	}
	propagator := ddtracer.NewPropagator(cfg)
	t := opentracer.New(
		ddtracer.WithPropagator(propagator),
		ddtracer.WithAgentAddr(os.Getenv("DD_AGENT_HOST")),
		ddtracer.WithServiceName(service),
		ddtracer.WithGlobalTag("env", env.Mode),
	)
	tracers.SetGlobalTracer(&tracers.Datadog{})
	return t, nil
}

func ConfigureOpenTracing(tracer opentracing.Tracer) {
	opentracing.SetGlobalTracer(tracer)
}

func Opentracing() context.Handler {
	return func(c *context.Context) {
		headers := context.M{}
		t := opentracing.GlobalTracer()
		tracer := tracers.GlobalTracer()
		spanCtx, _ := t.Extract(opentracing.HTTPHeaders, c.Request.Headers)
		span := t.StartSpan("new-request", opentracing.ChildOf(spanCtx))
		defer span.Finish()
		t.Inject(span.Context(), opentracing.HTTPHeaders, headers)
		fields := tracer.LogFields(headers)
		c.Headers = headers
		// span.SetTag(ext.SamplingPriority, ext.PriorityAutoKeep)
		c.SetLog(c.Log().WithFields(fields))
		c.Set("span", span)
		c.Next()
		if err := c.Error; err != nil {
			tracer.SetErrorTag(span, err)
		}
		// Inject the client span context into the headers
	}
}
