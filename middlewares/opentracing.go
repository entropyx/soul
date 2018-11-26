package middlewares

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/env"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

const (
	jaegerHeaderName        = "uber-trace-id"
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

func ConfigureJaegerSimple(hostPort, service string) (opentracing.Tracer, error) {
	sender, err := jaeger.NewUDPTransport(hostPort, 0)
	if err != nil {
		return nil, err
	}
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	tracer, _, err := cfg.New(
		service,
		jaegercfg.Reporter(jaeger.NewRemoteReporter(
			sender,
			jaeger.ReporterOptions.BufferFlushInterval(1*time.Second),
			jaeger.ReporterOptions.Logger(jLogger),
		)),
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
		jaegercfg.Observer(rpcmetrics.NewObserver(jMetricsFactory, rpcmetrics.DefaultNameNormalizer)),
	)
	if err != nil {
		return nil, err
	}
	return tracer, nil
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
