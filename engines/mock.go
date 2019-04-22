package engines

import (
	"github.com/entropyx/soul/context"
)

type Mock struct {
	IsConnected bool
	RoutingKey  string
	Handlers    []context.Handler
}

type MockConsumer struct {
	*Mock
	IsConnected bool
}

func (m *Mock) Close() error {
	m.IsConnected = false
	return nil
}

func (m *Mock) Connect() error {
	m.IsConnected = true
	return nil
}

func (m *Mock) Consumer(routingKey string) (Consumer, error) {
	m.RoutingKey = routingKey
	return &MockConsumer{m, false}, nil
}

func (m *Mock) MergeRoutingKeys(absolute, relative string) string {
	amqp := &AMQP{}
	return amqp.MergeRoutingKeys(absolute, relative)
}

func (m *MockConsumer) Consume(handlers []context.Handler) error {
	m.Handlers = handlers
	m.IsConnected = true
	return nil
}

func (m *MockConsumer) Close() error {
	m.IsConnected = false
	return nil
}
