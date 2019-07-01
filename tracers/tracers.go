package tracers

import (
	"github.com/entropyx/soul/context"
	"github.com/sirupsen/logrus"
)

var tracer Tracer

type Tracer interface {
	LogFields(context.M) logrus.Fields
}

func GlobalTracer() Tracer {
	return tracer
}

func SetGlobalTracer(t Tracer) {
	tracer = t
}
