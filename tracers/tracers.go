package tracers

import (
	"bytes"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/log"
	"github.com/sirupsen/logrus"
)

var tracer Tracer

type Tracer interface {
	LogFields(headers context.M, logger log.Logger) log.Logger
	SetErrorTag(span interface{}, err error)
}

func GlobalTracer() Tracer {
	if tracer == nil {
		return &noop{}
	}
	return tracer
}

func SetGlobalTracer(t Tracer) {
	tracer = t
}

func entryBuffer(entry *logrus.Entry) *bytes.Buffer {
	if entry.Buffer != nil {
		return entry.Buffer
	} else {
		return &bytes.Buffer{}
	}
}
