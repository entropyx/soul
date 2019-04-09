package engines

import (
	ctx "context"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/entropyx/soul/context"
)

var servers = map[string]*Server{}
var mutex = &sync.Mutex{}

type HTTPSimple struct {
	Address        string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	MaxHeaderBytes uint
}

type HTTPSimpleConsumer struct {
	address string
	path    string
}

type HTTPSimpleContext struct {
	w http.ResponseWriter
	r *http.Request
}

type Server struct {
	*http.Server
	patterns map[string][]context.Handler
}

type httpHandler struct {
	handler func(http.ResponseWriter, *http.Request)
}

type responseWriter struct {
}

func (h *HTTPSimple) Close() error {
	server := servers[h.Address]
	return server.Shutdown(ctx.Background())
}

func (h *HTTPSimple) Connect() error {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := servers[h.Address]; !ok {
		s := &http.Server{
			Addr:           h.Address,
			ReadTimeout:    h.ReadTimeout,
			WriteTimeout:   h.WriteTimeout,
			MaxHeaderBytes: int(h.MaxHeaderBytes),
		}
		server := &Server{s, map[string][]context.Handler{}}
		server.setHandler()
		servers[h.Address] = server
		go server.ListenAndServe()
	}
	return nil
}

func (h *HTTPSimple) Consumer(routingKey string) (Consumer, error) {
	return &HTTPSimpleConsumer{h.Address, routingKey}, nil
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
	mutex.Lock()
	defer mutex.Unlock()
	server, ok := servers[h.address]
	if !ok {
		return errors.New("undefined server")
	}
	server.patterns[h.path] = handlers
	return nil
}

func (h *HTTPSimpleConsumer) Close() error {
	delete(servers[h.address].patterns, h.path)
	return nil
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

func (s *Server) setHandler() {
	f := func(w http.ResponseWriter, r *http.Request) {
		u := r.URL
		mutex.Lock()
		handlers, ok := s.patterns[u.Path]
		mutex.Unlock()
		if !ok {
			w.WriteHeader(404)
			return
		}
		c := &HTTPSimpleContext{w, r}
		context.NewAndRun(c, handlers...)
	}
	handler := &httpHandler{f}
	s.Handler = handler
}
