package engines

import (
	"log"
	"os"
	"strconv"

	"github.com/entropyx/rabbitgo"
	"github.com/soul-go/soul/context"
	"github.com/streadway/amqp"
)

type AMQP struct {
	ExchangeName    string
	ExchangeType    string
	ExchangeDurable bool
	Queue           string
	RoutingKey      string
	PrefetchCount   uint8
	AutoAck         bool
	conn            *rabbitgo.Connection
}

type Context struct {
	*rabbitgo.Delivery
}

func (a *AMQP) Connect() error {
	rabbitPort, err := strconv.Atoi(os.Getenv("AMQP_PORT"))
	if err != nil {
		log.Fatal("invalid amqp port")
	}
	config := &rabbitgo.Config{
		Host:     os.Getenv("AMQP_HOST"),
		Username: os.Getenv("AMQP_USER"),
		Password: os.Getenv("AMQP_PASS"),
		Vhost:    os.Getenv("AMQP_VHOST"),
		Port:     rabbitPort,
	}
	if a.conn, err = rabbitgo.NewConnection(config); err != nil {
		return err
	}
	return nil
}

func (a *AMQP) Consume(routingKey string, handlers []context.Handler) error {
	exchange := &rabbitgo.Exchange{
		Name:    a.ExchangeName,
		Type:    a.ExchangeType,
		Durable: a.ExchangeDurable,
	}

	queue := &rabbitgo.Queue{
		Name: routingKey,
	}
	binding := &rabbitgo.BindingConfig{
		RoutingKey: routingKey,
	}
	consumerConfig := &rabbitgo.ConsumerConfig{
		Tag:           routingKey,
		PrefetchCount: int(a.PrefetchCount),
		AutoAck:       a.AutoAck,
	}
	consumer, err := a.conn.NewConsumer(exchange, queue, binding, consumerConfig)
	if err != nil {
		return err
	}
	h := func(d *rabbitgo.Delivery) {
		c := &Context{d}
		context := context.NewContext(c)
		context.RunHandlers(handlers)
	}
	return consumer.ConsumeRPC(h)
}

func (a *AMQP) MergeRoutingKeys(absolute, relative string) string {
	merge := absolute
	if absolute != "" && relative != "" {
		merge += "."
	}
	merge += relative
	return merge
}

func (a *AMQP) Run() error {
	return nil
}

func (c *Context) Publish(r *context.R) {
	c.Headers = amqp.Table(r.Headers)
	c.Data(r.Body, r.Type)
}

func (c *Context) Request() *context.R {
	return &context.R{
		Type:       c.Type,
		Headers:    context.M(c.Headers),
		RoutingKey: c.RoutingKey,
		Body:       c.Body,
	}
}
