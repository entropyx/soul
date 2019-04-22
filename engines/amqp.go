package engines

import (
	"log"
	"os"
	"strconv"

	"github.com/entropyx/rabbitgo"
	"github.com/entropyx/soul/context"
	"github.com/streadway/amqp"
)

type AckOptions struct {
	Multiple bool
	Requeue  bool
}

type AMQP struct {
	ExchangeName    string
	ExchangeType    string
	ExchangeDurable bool
	PrefetchCount   uint8
	AutoAck         bool
	conn            *rabbitgo.Connection
}

type AMQPConsumer struct {
	*rabbitgo.Consumer
}

type Context struct {
	*rabbitgo.Delivery
}

func (a *AMQP) Close() error {
	if a.conn != nil {
		return a.conn.Close()
	}
	return nil
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

func (a *AMQP) Consumer(routingKey string) (Consumer, error) {
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
		return nil, err
	}
	return &AMQPConsumer{consumer}, nil
}

func (a *AMQP) IsConnected() bool {
	if a.conn == nil {
		return false
	}
	return a.conn.IsConnected && !a.conn.IsBlocked
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

func (a *AMQPConsumer) Consume(handlers []context.Handler) error {
	h := func(d *rabbitgo.Delivery) {
		c := &Context{d}
		context := context.NewContext(c)
		context.RunHandlers(handlers...)
	}
	return a.ConsumeRPC(h)
}

func (a *AMQPConsumer) Close() error {
	return a.Shutdown()
}

func (c *Context) Ack(args ...interface{}) {
	opts := args[0].(*AckOptions)
	c.Delivery.Ack(opts.Multiple)
}

func (c *Context) Nack(args ...interface{}) {
	opts := args[0].(*AckOptions)
	c.Delivery.Nack(opts.Multiple, opts.Requeue)
}

func (c *Context) Publish(r *context.R) {
	c.Headers = amqp.Table(r.Headers)
	c.Data(r.Body, r.ContentType)
}

func (c *Context) Request() *context.R {
	return &context.R{
		ContentType: c.ContentType,
		Headers:     context.M(c.Headers),
		RoutingKey:  c.RoutingKey,
		Body:        c.Body,
	}
}
