package tracers

import (
	"github.com/entropyx/soul/context"
	"github.com/sirupsen/logrus"
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
