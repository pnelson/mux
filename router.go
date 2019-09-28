package mux

import (
	"errors"
	"net/http"
	"sort"
	"strings"
)

// Route represents a route.
type Route struct {
	name       string
	pattern    string
	methods    map[string]struct{}
	handler    http.Handler
	middleware []func(http.Handler) http.Handler
}

// NewRoute returns a new route.
func NewRoute(pattern string, handler http.Handler, opts ...RouteOption) *Route {
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	r := &Route{
		pattern:    pattern,
		methods:    make(map[string]struct{}),
		middleware: make([]func(http.Handler) http.Handler, 0),
	}
	for _, option := range opts {
		option(r)
	}
	r.build(handler)
	return r
}

// Name returns the route name.
func (r *Route) Name() string {
	return r.name
}

// Pattern returns the route pattern.
func (r *Route) Pattern() string {
	return r.pattern
}

// Methods returns the methods the route responds to.
func (r *Route) Methods() []string {
	methods := make([]string, 0)
	for method := range r.methods {
		methods = append(methods, method)
	}
	sort.Strings(methods)
	return methods
}

// ServeHTTP implements the http.Handler interface.
func (r *Route) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.handler.ServeHTTP(w, req)
}

// build wraps h with the configured middleware.
func (r *Route) build(h http.Handler) {
	r.handler = h
	for i := len(r.middleware) - 1; i >= 0; i-- {
		r.handler = r.middleware[i](r.handler)
	}
}

// Router represents the ability to match HTTP requests to handlers.
type Router interface {
	Add(r *Route) error
	Match(req *http.Request) (*Route, Params, error)
}

// Params represents the route parameters.
type Params map[string]string

// Builder represents the ability to build routes by name.
type Builder interface {
	Build(name string, params Params) (string, error)
}

// Walker represents the ability to walk the available routes.
type Walker interface {
	Walk(fn WalkFunc) error
}

// WalkFunc is called for each route visited.
// Return a non-nil error to terminate iteration.
type WalkFunc func(r *Route) error

// ErrBuild represents a Builder error.
var ErrBuild = errors.New("mux: named route does not exist")

// ErrNotFound represents a HTTP 404 Not Found error.
var ErrNotFound = errors.New("mux: no route matched request")

// ErrMethodNotAllowed represents a HTTP 405 Method Not Allowed error.
type ErrMethodNotAllowed []string

// Error implements the error interface.
func (e ErrMethodNotAllowed) Error() string {
	return strings.Join(e, ", ")
}
