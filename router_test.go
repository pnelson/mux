package mux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouteName(t *testing.T) {
	name := "test"
	r := NewRoute("/", nil, WithName(name))
	assertString(t, "name", r.Name(), name)
}

func TestRoutePattern(t *testing.T) {
	pattern := "/test"
	r := NewRoute(pattern, nil)
	assertString(t, "pattern", r.Pattern(), pattern)
}

func TestRouteMethods(t *testing.T) {
	method := http.MethodPost
	r := NewRoute("/", nil, WithMethod(method))
	assertDeepEqual(t, "methods", r.Methods(), []string{method})
}

func TestRouteMethodsAutomaticHEAD(t *testing.T) {
	method := http.MethodGet
	r := NewRoute("/", nil, WithMethod(method))
	assertDeepEqual(t, "methods", r.Methods(), []string{method, http.MethodHead})
}

func TestRouteMiddleware(t *testing.T) {
	have := ""
	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			have += "a"
			next.ServeHTTP(w, req)
			have += "b"
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			have += "c"
			next.ServeHTTP(w, req)
			have += "d"
		})
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {})
	r := NewRoute("/", h, WithMiddleware(m1, m2))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, (*http.Request)(nil))
	assertString(t, "middleware", have, "acdb")
}
