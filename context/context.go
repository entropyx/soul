package context

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

const (
	TypeJson   = "application/json"
	TypePlain  = "text/plain"
	TypeProto  = "application/protobuf"
	TypeXProto = "application/x-protobuf"
)

type key uint

const (
	keyEntry key = iota
	keyServiceName
)

type Handler func(*Context)

type C interface {
	Publish(*R)
	Request() *R
	Ack(args ...interface{})
	Nack(args ...interface{})
}

type Context struct {
	C     C
	Error error
	// Log      *logrus.Entry
	TraceID     string
	SpanID      string
	ServiceName string
	Request     *R
	Headers     M
	handlers    []Handler
	index       int8
	m           mi
}

type R struct {
	Body        []byte
	Headers     M
	RoutingKey  string
	ContentType string
}

type M map[string]interface{}

type mi map[interface{}]interface{}

func HTTPHeaderToM(header http.Header) M {
	m := M{}
	for k, v := range header {
		m[k] = v
	}
	return m
}

func MtoHeader(m M) http.Header {
	header := http.Header{}
	for k, v := range m {
		header[k] = v.([]string)
	}
	return header
}

func NewContext(c C) *Context {
	context := &Context{C: c, Headers: M{}, ServiceName: os.Getenv("SERVICE_NAME")}
	context.SetLog(logrus.NewEntry(logrus.StandardLogger()))
	context.setRequest()
	return context
}

func NewAndRun(c C, handlers ...Handler) *Context {
	ctx := NewContext(c)
	ctx.RunHandlers(handlers...)
	return ctx
}

func (c *Context) Ack(args ...interface{}) {
	c.C.Ack(args...)
}

func (c *Context) Bind(v interface{}) error {
	var err error
	r := c.Request
	body := r.Body
	switch r.ContentType {
	case TypeJson:
		err = json.Unmarshal(body, v)
	case TypeProto, TypeXProto:
		err = proto.Unmarshal(body, v.(proto.Message))
	default:
		err = errors.New(fmt.Sprintf("unknown type: %s", r.ContentType))
	}
	return err
}

func (c *Context) Abort(v interface{}) {
	c.index = math.MaxInt8 / 2
	r := c.Request
	if v == nil {
		return
	}
	switch r.ContentType {
	case TypeJson:
		c.JSON(v)
	case TypeProto:
		c.Proto(v.(proto.Message))
	case TypePlain:
		c.String(v.(string))
	}
}

func (c *Context) AbortWithError(v interface{}, err error) {
	c.Error = err
	c.Abort(v)
}

func (c *Context) Deadline() (time.Time, bool) {
	return time.Now(), false
}

func (c *Context) Done() <-chan struct{} {
	return nil
}

func (c *Context) Err() error {
	return nil
}

func (c *Context) Get(key interface{}) interface{} {
	return c.m[key]
}

func (c *Context) JSON(v interface{}) {
	body, _ := json.Marshal(v)
	c.publish(body, TypeJson)
}

func (c *Context) Log() *logrus.Entry {
	return c.Get(keyEntry).(*logrus.Entry)
}

func (c *Context) Nack(args ...interface{}) {
	c.C.Nack(args...)
}

func (c *Context) Next() {
	c.index++

	for s := int8(len(c.handlers)); c.index < s; {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Proto(pb proto.Message) {
	body, _ := proto.Marshal(pb)
	c.publish(body, TypeProto)
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Get(key)
}

func (c *Context) reset() {
	c.index = -1
}

func (c *Context) RunHandlers(handlers ...Handler) {
	c.reset()
	c.handlers = handlers
	c.Next()
}

func (c *Context) Set(key interface{}, value interface{}) {
	if c.m == nil {
		c.m = mi{}
	}
	c.m[key] = value
}

func (c *Context) SetLog(entry *logrus.Entry) {
	c.Set(keyEntry, entry)
}

func (c *Context) String(text string) {
	body := []byte(text)
	c.publish(body, TypePlain)
}

func (c *Context) publish(body []byte, t string) {
	r := &R{
		Body:        body,
		Headers:     c.Headers,
		ContentType: t,
	}
	c.C.Publish(r)
}

func (c *Context) setRequest() {
	r := c.C.Request()
	c.Request = r
}

// ForeachKey conforms to the TextMapReader interface.
func (m M) ForeachKey(handler func(key, val string) error) error {
	for k, v := range m {
		if err := handler(k, v.(string)); err != nil {
			return err
		}
	}
	return nil
}

// Set conforms to the TextMapWriter interface.
func (m M) Set(key, value string) {
	m[key] = value
}

func LogFromContext(c context.Context) *logrus.Entry {
	log, ok := c.Value(keyEntry).(*logrus.Entry)
	if ok {
		return log
	}
	return logrus.NewEntry(logrus.StandardLogger())
}

func ServiceNameFromContext(c context.Context) string {
	name, ok := c.Value(keyServiceName).(string)
	if ok {
		return name
	}
	return ""
}

func WithLog(c context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(c, keyEntry, entry)
}

func WithServiceName(c context.Context, name string) context.Context {
	return context.WithValue(c, keyServiceName, name)
}
