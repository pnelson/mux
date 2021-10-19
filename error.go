package mux

import (
	"fmt"
	"net/http"
)

// Error repesents an error view.
type Error interface {
	error
	StatusCode() int
}

// Resolver represents the ability to resolve an error to a view.
type Resolver interface {
	Resolve(req *http.Request, code int, err error) Error
}

// ResolverFunc is an adapter to allow the use of
// ordinary functions as Resolvers.
type ResolverFunc func(req *http.Request, code int, err error) Error

// Resolve implements the Resolver interface.
func (fn ResolverFunc) Resolve(req *http.Request, code int, err error) Error {
	return fn(req, code, err)
}

// ErrorView is the default error view.
type ErrorView struct {
	Code      int    `json:"code"`
	Title     string `json:"title"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id"`
}

// Error implements the error interface.
func (v ErrorView) Error() string {
	return v.Message
}

// StatusCode implements the mux.Error interface.
func (v ErrorView) StatusCode() int {
	return v.Code
}

// NewErrorView returns a new ErrorView.
func NewErrorView(req *http.Request, code int, err error) ErrorView {
	return ErrorView{
		Code:      code,
		Title:     http.StatusText(code),
		Message:   ErrorText(code, err),
		RequestID: RequestID(req),
	}
}

// ErrorText returns supplementary message text for errors.
//
// Explicit descriptions are returned for mux errors. The error text is
// returned for http.StatusUnprocessableEntity status codes and instances
// of mux.ValidationError. The empty string is returned for unknown errors.
func ErrorText(code int, err error) string {
	switch code {
	case http.StatusUnprocessableEntity:
		return err.Error()
	case http.StatusInternalServerError:
		return "An unexpected error has occurred."
	}
	switch err {
	case ErrDecodeContentType:
		return "Invalid content-type header."
	case ErrDecodeRequestData:
		return "Invalid request data."
	}
	switch err.(type) {
	case ErrMethodNotAllowed:
		return "The method is not allowed for the requested URL."
	case ValidationError:
		return err.Error()
	}
	return ""
}

// ErrRedirect represents a redirect response.
type ErrRedirect struct {
	URL  string
	Code int
}

// Error implements the error interface.
func (e ErrRedirect) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.URL)
}

// Panic is an error resolved from a panic with a stack trace.
type Panic struct {
	err   interface{}
	stack []byte
}

// Error implements the error interface.
func (e Panic) Error() string {
	err, ok := e.err.(error)
	if ok {
		return err.Error()
	}
	return e.err.(string)
}

// String implements the fmt.Stringer interface.
func (e Panic) String() string {
	return string(e.stack)
}

// Abort resolves an error to a view and encodes the response.
func (h *Handler) Abort(w http.ResponseWriter, req *http.Request, err error) {
	switch err {
	case nil:
		return
	case ErrEncodeMatch:
		abort(w, http.StatusNotAcceptable)
		return
	}
	redirect, ok := err.(ErrRedirect)
	if ok {
		http.Redirect(w, req, redirect.URL, redirect.Code)
		return
	}
	view := h.resolve(w, req, err)
	code := view.StatusCode()
	if code == http.StatusInternalServerError {
		h.log(req, err)
	}
	err = h.Encode(w, req, view, code)
	if err != nil {
		h.log(req, err)
	}
}

// resolve resolves errors to an error view.
func (h *Handler) resolve(w http.ResponseWriter, req *http.Request, err error) Error {
	switch err {
	case ErrNotFound:
		return h.resolver.Resolve(req, http.StatusNotFound, err)
	case ErrDecodeContentType:
		return h.resolver.Resolve(req, http.StatusUnsupportedMediaType, err)
	case ErrDecodeRequestData:
		return h.resolver.Resolve(req, http.StatusBadRequest, err)
	}
	switch e := err.(type) {
	case Error:
		return e
	case ErrMethodNotAllowed:
		allowed := err.Error()
		w.Header().Set("Allow", allowed)
		return h.resolver.Resolve(req, http.StatusMethodNotAllowed, err)
	case ValidationError:
		return h.resolver.Resolve(req, http.StatusUnprocessableEntity, err)
	}
	return h.resolver.Resolve(req, http.StatusInternalServerError, err)
}

// abort replies to the request with a plain text error message.
func abort(w http.ResponseWriter, code int) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	message := http.StatusText(code)
	_, err := fmt.Fprintln(w, message)
	return err
}
