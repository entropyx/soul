package middlewares

import (
	"os"

	log "github.com/sirupsen/logrus"

	opentracing "github.com/opentracing/opentracing-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/env"
)

const (
	datadogTraceHeaderName  = "trace-id"
	datadogParentHeaderName = "parent-id"
)

func ConfigureDatadog(service string) (opentracing.Tracer, error) {
	cfg := &tracer.PropagatorConfig{
		TraceHeader:  "trace-id",
		ParentHeader: "parent-id",
	}
	propagator := tracer.NewPropagator(cfg)
	t := opentracer.New(
		tracer.WithPropagator(propagator),
		tracer.WithAgentAddr(os.Getenv("DD_AGENT_HOST")),
		tracer.WithServiceName(service),
		tracer.WithGlobalTag("env", env.Mode),
	)
	return t, nil
}

func ConfigureOpenTracing(tracer opentracing.Tracer) {
	opentracing.SetGlobalTracer(tracer)
}

func Opentracing() context.Handler {
	return func(c *context.Context) {
		headers := context.M{}
		fields := log.Fields{}
		t := opentracing.GlobalTracer()
		spanCtx, _ := t.Extract(opentracing.HTTPHeaders, c.Request.Headers)
		span := t.StartSpan("new-request", opentracing.ChildOf(spanCtx))
		defer span.Finish()
		t.Inject(span.Context(), opentracing.HTTPHeaders, headers)
		for k, v := range headers {
			c.Headers[k] = v
			fields[k] = v
		}
		// span.SetTag(ext.SamplingPriority, ext.PriorityAutoKeep)
		c.Log = c.Log.WithFields(fields)
		c.Set("span", span)
		c.Next()
		if err := c.Err; err != nil {
			span.SetTag(ext.Error, err)
		}
		// Inject the client span context into the headers
	}
}
