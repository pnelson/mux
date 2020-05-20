package mux

import (
	"net/http"

	"golang.org/x/text/language"
)

// Option represents a functional option for configuration.
type Option func(*Handler)

// WithRouter sets the request router.
func WithRouter(router Router) Option {
	return func(h *Handler) {
		h.router = router
	}
}

// WithLocales sets the supported locales.
func WithLocales(matcher language.Matcher) Option {
	return func(h *Handler) {
		h.locales = matcher
	}
}

// WithDecoders sets the content-type/encoder mapping.
func WithDecoders(decoders map[string]Decoder) Option {
	return func(h *Handler) {
		h.decoders = decoders
	}
}

// WithEncoder sets the encoder negotiation function.
func WithEncoder(fn EncoderFunc) Option {
	return func(h *Handler) {
		h.encoder = fn
	}
}

// WithResolver sets the error resolver.
func WithResolver(resolver Resolver) Option {
	return func(h *Handler) {
		h.resolver = resolver
	}
}

// WithPool sets the buffer pool for encoding responses.
func WithPool(pool Pool) Option {
	return func(h *Handler) {
		h.pool = pool
	}
}

// WithLogger sets the logger.
func WithLogger(logger Logger) Option {
	return func(h *Handler) {
		h.log = logger
	}
}

// RouteOption represents a functional option for configuration.
type RouteOption func(*Route)

// WithName sets the route name.
func WithName(name string) RouteOption {
	return func(r *Route) {
		r.name = name
	}
}

// WithMethod sets the methods for which the route is valid.
func WithMethod(method ...string) RouteOption {
	return func(r *Route) {
		for _, m := range method {
			r.methods[m] = struct{}{}
		}
		_, ok := r.methods[http.MethodGet]
		if ok {
			r.methods[http.MethodHead] = struct{}{}
		}
	}
}

// WithMiddleware appends middleware to the middleware stack.
func WithMiddleware(middleware ...func(http.Handler) http.Handler) RouteOption {
	return func(r *Route) {
		r.middleware = append(r.middleware, middleware...)
	}
}
