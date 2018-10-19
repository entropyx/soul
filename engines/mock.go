package engines

import (
	"github.com/soul-go/soul/context"
)

type Mock struct {
	IsConnected bool
	RoutingKey  string
	Handlers    []context.Handler
}

func (m *Mock) Connect() error {
	m.IsConnected = true
	return nil
}

func (m *Mock) Consume(routingKey string, handlers []context.Handler) error {
	m.RoutingKey = routingKey
	m.Handlers = handlers
	return nil
}

func (m *Mock) MergeRoutingKeys(absolute, relative string) string {
	amqp := &AMQP{}
	return amqp.MergeRoutingKeys(absolute, relative)
}
