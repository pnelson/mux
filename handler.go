package mux

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"sync/atomic"

	"golang.org/x/text/language"
)

// Handler is a http.Handler with application lifecycle helpers.
type Handler struct {
	router     Router
	middleware []func(http.Handler) http.Handler
	locales    language.Matcher
	decoder    DecoderFunc
	encoder    EncoderFunc
	resolver   Resolver
	pool       Pool
	log        Logger
}

// Logger represents the ability to log errors.
type Logger func(req *http.Request, err error)

// HandlerFunc represents a HTTP handler with error handling.
type HandlerFunc func(w http.ResponseWriter, req *http.Request) error

// New returns a new handler.
func New(opts ...Option) *Handler {
	h := &Handler{}
	for _, option := range opts {
		option(h)
	}
	if h.router == nil {
		h.router = &tree{}
	}
	if h.locales == nil {
		h.locales = language.NewMatcher([]language.Tag{language.English})
	}
	if h.decoder == nil {
		h.decoder = NewContentTypeDecoder(map[string]Decoder{
			"application/json": &jsonDecoder{},
		})
	}
	if h.encoder == nil {
		encoder := &jsonEncoder{}
		h.encoder = NewAcceptEncoder(map[string]Encoder{
			"":                 encoder,
			"*/*":              encoder,
			"application/*":    encoder,
			"application/json": encoder,
		})
	}
	if h.resolver == nil {
		h.resolver = ResolverFunc(defaultResolver)
	}
	if h.pool == nil {
		h.pool = &pool{free: make(chan *bytes.Buffer, 1<<6)}
	}
	if h.log == nil {
		h.log = defaultLogger
	}
	return h
}

// Add registers a HandlerFunc.
func (h *Handler) Add(pattern string, handler HandlerFunc, opts ...RouteOption) *Route {
	fn := func(w http.ResponseWriter, req *http.Request) {
		err := handler(w, req)
		if err != nil {
			h.Abort(w, req, err)
		}
	}
	return h.Handle(pattern, http.Handler(http.HandlerFunc(fn)), opts...)
}

// Build returns the URL for the named route.
func (h *Handler) Build(name string, params Params) (string, error) {
	b, ok := h.router.(Builder)
	if !ok {
		return "", errors.New("mux: router is not a Builder")
	}
	return b.Build(name, params)
}

// FileServer registers a file system.
// The pattern is expected to be a prefix wildcard route.
// The pattern prefix is removed from the request URL before handled.
func (h *Handler) FileServer(pattern string, fs http.FileSystem) {
	prefix := pattern[:len(pattern)-1]
	handler := http.StripPrefix(prefix, http.FileServer(fs))
	h.Handle(pattern, handler, WithMethod(http.MethodGet))
}

// Handle registers a standard net/http Handler.
func (h *Handler) Handle(pattern string, handler http.Handler, opts ...RouteOption) *Route {
	opt := WithMiddleware(h.middleware...)
	opts = append([]RouteOption{opt}, opts...)
	r := NewRoute(pattern, handler, opts...)
	err := h.router.Add(r)
	if err != nil {
		panic(err)
	}
	return r
}

// ServeHTTP implements the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req = h.init(req)
	defer h.abort(w, req)
	h.dispatch(w, req)
}

// init initializes a new request context.
func (h *Handler) init(req *http.Request) *http.Request {
	n := atomic.AddUint64(&seq, 1)
	rc := &requestContext{seq: n, locale: h.locale(req)}
	return setContext(req, rc)
}

// dispatch finds a matching route and calls its prepared handler.
func (h *Handler) dispatch(w http.ResponseWriter, req *http.Request) {
	r, params, err := h.router.Match(req)
	if err != nil {
		merr, ok := err.(ErrMethodNotAllowed)
		if ok && req.Method == http.MethodOptions {
			allowed := merr.Error()
			w.Header().Set("Allow", allowed)
			return
		}
		h.Abort(w, req, err)
		return
	}
	rc := getContext(req)
	rc.route = r
	rc.params = params
	r.ServeHTTP(w, req)
}

// abort resolves an error if the application panics.
func (h *Handler) abort(w http.ResponseWriter, req *http.Request) {
	err := recover()
	if err != nil {
		p := Panic{err: err, stack: debug.Stack()}
		h.Abort(w, req, p)
	}
}

// locale parses the Accept-Language header and returns
// the BCP 47 language tag for the request.
func (h *Handler) locale(req *http.Request) language.Tag {
	accept := req.Header.Get("Accept-Language")
	// Intentionally ignored error as the default language will be matched.
	locales, _, _ := language.ParseAcceptLanguage(accept)
	tag, _, _ := h.locales.Match(locales...)
	return tag
}

// Use appends middleware to the global middleware stack.
func (h *Handler) Use(middleware ...func(http.Handler) http.Handler) {
	h.middleware = append(h.middleware, middleware...)
}

// Walk walks the named routes if the Router is a Walker.
func (h *Handler) Walk(fn WalkFunc) error {
	w, ok := h.router.(Walker)
	if !ok {
		return errors.New("mux: router is not a Walker")
	}
	return w.Walk(fn)
}

// Export walks the named routes and applies the exporter to the response body.
// A nil exporter writes to the dist directory within the current working
// directory. See FileSystemExporter documentation for more details.
func (h *Handler) Export(exporter Exporter) error {
	server := httptest.NewServer(h)
	defer server.Close()
	if exporter == nil {
		exporter = FileSystemExporter("dist")
	}
	fn := func(r *Route) error {
		route := r.Pattern()
		resp, err := http.Get(server.URL + route)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return exporter.Export(r, b)
	}
	return h.Walk(fn)
}

// Query returns the first query value associated with the given key.
// If there are no values associated with the key, Query returns the
// empty string.
func Query(req *http.Request, name string) string {
	return req.URL.Query().Get(name)
}

// RequestID returns the request identifier from the X-Request-ID header.
// The formatted request sequence number is returned if the header is not set.
func RequestID(req *http.Request) string {
	id := req.Header.Get("X-Request-ID")
	if id == "" {
		n := Sequence(req)
		id = fmt.Sprintf("%d", n)
	}
	return id
}

// defaultResolver represents the default resolver.
func defaultResolver(req *http.Request, code int, err error) Error {
	return defaultErrorView{
		Code:      code,
		Title:     http.StatusText(code),
		Message:   ErrorText(code, err),
		RequestID: RequestID(req),
	}
}

// defaultLogger represents the default logger.
func defaultLogger(req *http.Request, err error) {
	message := err.Error()
	perr, ok := err.(Panic)
	if ok {
		message += "\n" + perr.String()
	}
	log.Println(message)
}
