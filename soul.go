package soul

import (
	"strings"
	"time"

	"github.com/entropyx/soul/context"
	"github.com/entropyx/soul/engines"
	log "github.com/sirupsen/logrus"
)

type Router struct {
	RouteGroup
	engine engines.Engine
	routes map[string][]context.Handler
}

type RouteGroup struct {
	routingKey string
	handlers   []context.Handler
	router     *Router
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

func GetValue(key string) string {
	for _, vs := range vars {
		split := strings.Split(vs, "=")
		k, v := split[0], split[1]
		if k == key {
			return v
		}
	}
	return ""
}
