package tracers

import (
	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/log"
	opentracing "github.com/opentracing/opentracing-go"
	"go.opencensus.io/trace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
)

const (
	datadogTraceHeaderName  = "trace_id"
	datadogParentHeaderName = "parent_id"
)

type Datadog struct{}

var _ Tracer = &Datadog{}

func (*Datadog) LogFields(m context.M, logger log.Logger) log.Logger {
	newLogger := logger.WithField(datadogTraceHeaderName, m[datadogTraceHeaderName])
	newLogger = newLogger.WithField(datadogParentHeaderName, m[datadogParentHeaderName])

	return newLogger
}

func (*Datadog) SetErrorTag(span interface{}, err error) {
	switch s := span.(type) {
	case opentracing.Span:
		s.SetTag(ext.Error, err)
	case trace.Span:
		s.AddAttributes(trace.StringAttribute("error.msg", err.Error()))
	}
}
