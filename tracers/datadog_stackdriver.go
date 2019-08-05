package tracers

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/entropyx/errors"
	propagation "github.com/entropyx/opencensus-propagation"
	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/log"
	opentracing "github.com/opentracing/opentracing-go"
	"go.opencensus.io/trace"
)

type DatadogSd struct{}

var _ Tracer = &DatadogSd{}

func (*DatadogSd) LogFields(m context.M, logger log.Logger) log.Logger {
	var traceID [16]byte
	var spanID [8]byte
	decodeAndCopyString(m[propagation.HeaderTraceID].(string), traceID[:])
	decodeAndCopyString(m[propagation.HeaderSpanID].(string), spanID[:])
	uTraceID := binary.BigEndian.Uint64(traceID[8:])
	uSpanID := binary.BigEndian.Uint64(spanID[:])
	newLogger := logger.WithField("trace_id", fmt.Sprintf("%d", uTraceID))
	newLogger = newLogger.WithField("span_id", fmt.Sprintf("%d", uSpanID))
	return newLogger
}

func (*DatadogSd) SetErrorTag(span interface{}, err error) {
	switch s := span.(type) {
	case opentracing.Span:
		// TODO: OpenTracing
	case trace.Span:
		e, ok := err.(errors.Error)
		if !ok {
			s.AddAttributes(trace.StringAttribute("error.msg", "undefined"))
			return
		}
		s.AddAttributes(trace.StringAttribute("error.msg", e.Message))
		s.SetStatus(trace.Status{
			Code:    trace.StatusCodeUnknown, // TODO: code strategy
			Message: err.Error(),
		})
	}
}

func decodeAndCopyString(s string, dst []byte) error {
	buf, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	copy(dst, buf)
	return nil
}
