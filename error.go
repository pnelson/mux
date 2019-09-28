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
type Resolver func(req *http.Request, code int, err error) Error

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
	if err == nil {
		return
	}
	switch err {
	case ErrEncodeAccept:
		abort(w, http.StatusNotAcceptable)
		return
	case ErrNotFound:
		err = h.resolve(req, http.StatusNotFound, err)
	case ErrDecodeContentType:
		err = h.resolve(req, http.StatusUnsupportedMediaType, err)
	case ErrDecodeRequestData:
		err = h.resolve(req, http.StatusBadRequest, err)
	}
	var view Error
	switch e := err.(type) {
	case Error:
		view = e
	case ErrMethodNotAllowed:
		allowed := err.Error()
		w.Header().Set("Allow", allowed)
		view = h.resolve(req, http.StatusMethodNotAllowed, err)
	case ValidationError:
		view = h.resolve(req, http.StatusUnprocessableEntity, err)
	default:
		view = h.resolve(req, http.StatusInternalServerError, err)
		h.log(req, err)
	}
	code := view.StatusCode()
	err = h.Encode(w, req, view, code)
	if err != nil {
		h.log(req, err)
	}
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

type defaultErrorView struct {
	Code      int    `json:"code"`
	Title     string `json:"title"`
	Message   string `json:"message,omitempty"`
	RequestID string `json:"request_id"`
}

func (v defaultErrorView) Error() string {
	return v.Message
}

func (v defaultErrorView) StatusCode() int {
	return v.Code
}
