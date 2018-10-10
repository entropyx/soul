package engines

import (
	"log"
	"os"
	"strconv"

	"github.com/dsmontoya/soul/context"
	"github.com/entropyx/rabbitgo"
	"github.com/streadway/amqp"
)

type AMQP struct {
	ExchangeName  string
	ExchangeTopic string
	Queue         string
	RoutingKey    string
	PrefetchCount uint8
	AutoAck       bool
	conn          *rabbitgo.Connection
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
		Name:    "entropy",
		Type:    "topic",
		Durable: true,
	}

	queue := &rabbitgo.Queue{
		Name: routingKey,
	}
	binding := &rabbitgo.BindingConfig{
		RoutingKey: routingKey,
	}
	consumerConfig := &rabbitgo.ConsumerConfig{
		Tag:           routingKey,
		PrefetchCount: 20,
		AutoAck:       true,
	}
	consumer, err := a.conn.NewConsumer(exchange, queue, binding, consumerConfig)
	if err != nil {
		return err
	}
	h := func(d *rabbitgo.Delivery) {
		c := &Context{d}
		context := context.NewContext(c)
		for _, handler := range handlers {
			handler(context)
		}
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
		Type:    c.Type,
		Headers: context.M(c.Headers),
		Body:    c.Body,
	}
}
