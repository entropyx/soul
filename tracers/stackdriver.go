package tracers

import (
	"fmt"
	"os"

	propagation "github.com/entropyx/opencensus-propagation"
	"github.com/entropyx/soul/context"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

const (
	stackdriverTraceID = "trace"
	stackdriverSpanID  = "spanId"
)

type Stackdriver struct{}

type StackdriverFormatter struct {
}

func (*Stackdriver) LogFields(m context.M) logrus.Fields {
	traceID := m[propagation.HeaderTraceID]
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	fields := logrus.Fields{
		stackdriverTraceID: fmt.Sprintf("project/%s/traces/%s", traceID, project),
		stackdriverSpanID:  m[propagation.HeaderSpanID],
	}
	return fields
}

func (*Stackdriver) SetErrorTag(span interface{}, err error) {
	switch s := span.(type) {
	case opentracing.Span:
		// TODO: OpenTracing
	case trace.Span:
		s.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown, // TODO: code strategy
			Message: err.Error(),
		})
	}
}

func (s *StackdriverFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	jsonFormatter := &logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
		},
	}
	return jsonFormatter.Format(entry)
}
