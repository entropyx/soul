package engines

import (
	ctx "context"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/entropyx/soul/context"
)

type HTTPSimple struct {
	Address        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes uint
}

type HTTPSimpleConsumer struct {
	server *http.Server
	path   string
}

type HTTPSimpleContext struct {
	w http.ResponseWriter
	r *http.Request
}

type httpHandler struct {
	handler func(http.ResponseWriter, *http.Request)
}

type responseWriter struct {
}

func (h *HTTPSimple) Close() error {
	return nil
}

func (h *HTTPSimple) Connect() error {
	return nil
}

func (h *HTTPSimple) Consumer(routingKey string) (Consumer, error) {
	s := &http.Server{
		Addr:           h.Address,
		ReadTimeout:    h.ReadTimeout,
		WriteTimeout:   h.WriteTimeout,
		MaxHeaderBytes: int(h.MaxHeaderBytes),
	}
	return &HTTPSimpleConsumer{s, routingKey}, nil
}

func (h *HTTPSimple) MergeRoutingKeys(absolute, relative string) string {
	merge := absolute
	if absolute == "" {
		merge += "/"
	}
	if absolute != "" && relative != "" {
		merge += "/"
	}
	merge += relative
	return merge
}

func (h *HTTPSimpleConsumer) Consume(handlers []context.Handler) error {
	f := func(w http.ResponseWriter, r *http.Request) {
		u := r.URL
		if u.Path != h.path {
			w.WriteHeader(404)
			return
		}
		c := &HTTPSimpleContext{w, r}
		context.NewAndRun(c, handlers...)
	}
	handler := &httpHandler{f}
	h.server.Handler = handler
	return h.server.ListenAndServe()
}

func (h *HTTPSimpleConsumer) Close() error {
	return h.server.Shutdown(ctx.Background())
}

func (c *HTTPSimpleContext) Ack(args ...interface{}) {

}

func (c *HTTPSimpleContext) Nack(args ...interface{}) {

}

func (c *HTTPSimpleContext) Publish(r *context.R) {
	w := c.w
	header := w.Header()
	for k, v := range r.Headers {
		if k == "status" {
			continue
		}
		header[k] = []string{v.(string)}
	}
	code := r.Headers["status"].(int)
	w.WriteHeader(code)
	w.Write(r.Body)
}

func (c *HTTPSimpleContext) Request() *context.R {
	req := c.r
	u := req.URL
	contentType := req.Header.Get("Content-Type")
	b, _ := ioutil.ReadAll(req.Body)
	r := &context.R{
		Body:        b,
		Headers:     context.HTTPHeaderToM(req.Header),
		RoutingKey:  u.Path,
		ContentType: contentType,
	}
	return r
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler(w, r)
}
