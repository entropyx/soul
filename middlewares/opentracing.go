package middlewares

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/entropyx/dd-trace-go/ddtrace/ext"
	"github.com/entropyx/dd-trace-go/ddtrace/opentracer"
	"github.com/entropyx/dd-trace-go/ddtrace/tracer"
	opentracing "github.com/opentracing/opentracing-go"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/env"
)

const (
	datadogTraceHeaderName  = "trace_id"
	datadogParentHeaderName = "parent_id"
)

func ConfigureDatadog(service string) (opentracing.Tracer, error) {
	cfg := &tracer.PropagatorConfig{
		TraceHeader:  "trace_id",
		ParentHeader: "parent_id",
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
		requestHeaders := c.Request.Headers

		c.Log().Debugf("Tracing info from headers: span_id:%s trace_id:%s", requestHeaders[datadogParentHeaderName], requestHeaders[datadogTraceHeaderName])

		t := opentracing.GlobalTracer()
		spanCtx, _ := t.Extract(opentracing.HTTPHeaders, c.Request.Headers)
		span := t.StartSpan("new-request", opentracing.ChildOf(spanCtx))
		defer span.Finish()
		t.Inject(span.Context(), opentracing.HTTPHeaders, headers)
		for k, v := range headers {
			fk := strings.Replace(k, "-", "_", -1)
			c.Headers[fk] = v
			fields[fk] = v
		}

		c.Log().Debugf("Tracing info in response headers: span_id:%s trace_id:%s", c.Headers[datadogParentHeaderName], c.Headers[datadogTraceHeaderName])

		// span.SetTag(ext.SamplingPriority, ext.PriorityAutoKeep)
		c.SetLog(c.Log().WithFields(fields))
		c.Set("span", span)
		c.Next()
		if err := c.Err; err != nil {
			span.SetTag(ext.Error, err)
		}
		// Inject the client span context into the headers
	}
}
