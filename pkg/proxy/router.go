package proxy

import (
	"net/http"
	"sync"
)

type Router struct {
	routes    map[string]*Route
	routeLock sync.RWMutex
}

type Route struct {
	Path      string
	Handler   http.Handler
	RateLimit int
	Methods   []string
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]*Route),
	}
}

func (r *Router) AddRoute(path string, route *Route) {
	r.routeLock.Lock()
	defer r.routeLock.Unlock()
	r.routes[path] = route
}

func (r *Router) GetRoute(path string) (*Route, bool) {
	r.routeLock.RLock()
	defer r.routeLock.RUnlock()
	route, exists := r.routes[path]
	return route, exists
}
