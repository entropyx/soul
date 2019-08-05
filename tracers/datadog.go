package tracers

import (
	"github.com/entropyx/dd-trace-go/ddtrace/ext"
	"github.com/entropyx/errors"
	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/log"
	opentracing "github.com/opentracing/opentracing-go"
	"go.opencensus.io/trace"
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
		e, ok := err.(errors.Error)
		if !ok {
			s.AddAttributes(trace.StringAttribute("error.msg", "undefined"))
			return
		}
		s.AddAttributes(trace.StringAttribute("error.msg", e.Message))
	}
}
