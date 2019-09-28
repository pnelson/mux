package mux

import (
	"context"
	"net/http"

	"golang.org/x/text/language"
)

// key represents http context.Context keys.
type key int

// requestContextKey represents the key holding the request context.
var requestContextKey interface{} = key(0)

// requestContext represents the mux-specific request context.
type requestContext struct {
	id     string
	route  *Route
	params Params
	locale language.Tag
}

func getContext(req *http.Request) *requestContext {
	return req.Context().Value(requestContextKey).(*requestContext)
}

func setContext(req *http.Request, rc *requestContext) *http.Request {
	ctx := req.Context()
	ctx = context.WithValue(ctx, requestContextKey, rc)
	return req.WithContext(ctx)
}

// RequestID returns the request identifier.
func RequestID(req *http.Request) string {
	rc := getContext(req)
	return rc.id
}

// Match returns the matching route for the request, or nil if no match.
func Match(req *http.Request) *Route {
	rc := getContext(req)
	return rc.route
}

// Locale returns the best match BCP 47 language tag
// parsed from the Accept-Language header.
func Locale(req *http.Request) language.Tag {
	rc := getContext(req)
	return rc.locale
}

// SetLocale sets the BCP 47 language tag on the request context.
func SetLocale(req *http.Request, tag language.Tag) {
	rc := getContext(req)
	rc.locale = tag
}

// Param returns the named parameter.
func Param(req *http.Request, name string) string {
	rc := getContext(req)
	return rc.params[name]
}
