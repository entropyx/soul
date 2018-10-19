package middlewares

import (
	"fmt"
	"os"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	jaeger "github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-lib/metrics"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/dsmontoya/soul/context"
	"github.com/dsmontoya/soul/env"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

const (
	jaegerHeaderName  = "uber-trace-id"
	datadogHeaderName = "x-datadog-trace-id"
)

func ConfigureDatadog(service string) (opentracing.Tracer, error) {
	t := opentracer.New(
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
	fmt.Println("service", service)
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

func Opentracing(c *context.Context) {
	t := opentracing.GlobalTracer()
	spanCtx, _ := t.Extract(opentracing.HTTPHeaders, c.Request.Headers)
	span := t.StartSpan("new-request", ext.RPCServerOption(spanCtx))
	defer span.Finish()
	c.Set("span", span)
	c.Next()
	// Inject the client span context into the headers
	t.Inject(span.Context(), opentracing.HTTPHeaders, c.Headers)
}
