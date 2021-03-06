package engines

import "github.com/entropyx/soul/context"

type Consumer interface {
	Consume(handlers []context.Handler) error
	Close() error
}

type Engine interface {
	MergeRoutingKeys(string, string) string
	Connect() error
	IsConnected() bool
	Close() error
	Consumer(routingKey string) (Consumer, error)
}
