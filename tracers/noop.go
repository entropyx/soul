package tracers

import (
	"github.com/entropyx/soul/context"
	"github.com/sirupsen/logrus"
)

type noop struct {
}

func (*noop) LogFields(context.M) logrus.Fields {
	noopWarning()
	return logrus.Fields{}
}

func (*noop) SetErrorTag(span interface{}, err error) {
	noopWarning()
}

func noopWarning() {
	logrus.Warning("NOOP tracer")
}
