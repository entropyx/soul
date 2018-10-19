package soul

import (
	"time"

	"github.com/dsmontoya/soul/context"
	log "github.com/sirupsen/logrus"
)

type Router struct {
	RouteGroup
	engine Engine
	routes map[string][]context.Handler
}

type RouteGroup struct {
	routingKey string
	handlers   []context.Handler
	router     *Router
}

type Engine interface {
	MergeRoutingKeys(string, string) string
	Connect() error
	Consume(string, []context.Handler) error
}

func (r *RouteGroup) Group(routingKey string) *RouteGroup {
	return &RouteGroup{router: r.router, routingKey: r.mergeRoutingKey(routingKey), handlers: r.handlers}
}

func (r *RouteGroup) Include(handlers ...context.Handler) {
	r.handlers = append(r.handlers, handlers...)
}

func (r *RouteGroup) Listen(routingKey string, handler context.Handler) {
	handlers := r.handlers
	handlers = append(handlers, handler)
	r.router.routes[r.mergeRoutingKey(routingKey)] = handlers
}

func (r *RouteGroup) mergeRoutingKey(relativeRoutingKey string) string {
	return r.router.engine.MergeRoutingKeys(r.routingKey, relativeRoutingKey)
}

func (r *Router) connect() {
	for {
		if err := r.engine.Connect(); err != nil {
			log.Error("Unable to connect. Retrying...")
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
}
