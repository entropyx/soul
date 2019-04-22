package engines

import (
	"github.com/entropyx/soul/context"
)

type Mock struct {
	isConnected bool
	RoutingKey  string
	Handlers    []context.Handler
}

type MockConsumer struct {
	*Mock
	IsConnected bool
}

func (m *Mock) Close() error {
	m.isConnected = false
	return nil
}

func (m *Mock) Connect() error {
	m.isConnected = true
	return nil
}

func (m *Mock) Consumer(routingKey string) (Consumer, error) {
	m.RoutingKey = routingKey
	return &MockConsumer{m, false}, nil
}

func (m *Mock) IsConnected() bool {
	return m.isConnected
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
