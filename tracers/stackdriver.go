package tracers

import (
	"fmt"
	"os"

	propagation "github.com/entropyx/opencensus-propagation"
	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/log"
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

var _ Tracer = &Stackdriver{}

func (*Stackdriver) LogFields(m context.M, logger log.Logger) log.Logger {
	traceID := m[propagation.HeaderTraceID]
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	newLogger := logger.WithField(stackdriverTraceID, fmt.Sprintf("projects/%s/traces/%s", project, traceID))
	newLogger = newLogger.WithField("trace_id", traceID)
	newLogger = newLogger.WithField(stackdriverSpanID, m[propagation.HeaderSpanID])
	return newLogger
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
