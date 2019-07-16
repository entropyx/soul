package tracers

import (
	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/log"
	"github.com/sirupsen/logrus"
)

type noop struct {
}

var _ Tracer = &noop{}

func (*noop) LogFields(m context.M, l log.Logger) log.Logger {
	noopWarning()
	return l
}

func (*noop) SetErrorTag(span interface{}, err error) {
	noopWarning()
}

func noopWarning() {
	logrus.Warning("NOOP tracer")
}
