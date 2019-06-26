package tracers

import (
	"github.com/entropyx/dd-trace-go/ddtrace/ext"
	"github.com/entropyx/soul/context"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

const (
	datadogTraceHeaderName  = "trace_id"
	datadogParentHeaderName = "parent_id"
)

type Datadog struct{}

func (*Datadog) LogFields(m context.M) logrus.Fields {
	fields := logrus.Fields{
		datadogTraceHeaderName:  m[datadogTraceHeaderName],
		datadogParentHeaderName: m[datadogParentHeaderName],
	}
	return fields
}

func (*Datadog) SetErrorTag(span interface{}, err error) {
	switch s := span.(type) {
	case opentracing.Span:
		s.SetTag(ext.Error, err)
	case trace.Span:
		// TODO: implement OpenCensus
	}
}
