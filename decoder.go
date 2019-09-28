package mux

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
)

// A Form represents a form with validation.
type Form interface {
	// Validate sanitizes and validates the form.
	Validate() error
}

// Decoder represents the ability to decode, sanitize
// and validate a request body.
type Decoder interface {
	Decode(req *http.Request, form Form) error
}

// ValidationError represents a form validation error.
type ValidationError struct {
	err error
}

// Error implements the error interface.
func (e ValidationError) Error() string {
	return e.err.Error()
}

// Decoder errors.
var (
	ErrDecodeContentType = errors.New("mux: no decoder matched request")
	ErrDecodeRequestData = errors.New("mux: bad request data for decoder")
)

// Decode decodes, sanitizes and validates the request body
// and stores the result in to the value pointed to by form.
func (h *Handler) Decode(req *http.Request, form Form) error {
	v := req.Header.Get("Content-Type")
	media, _, err := mime.ParseMediaType(v)
	if err != nil {
		return ErrDecodeContentType
	}
	d, ok := h.decoders[media]
	if !ok {
		return ErrDecodeContentType
	}
	err = d.Decode(req, form)
	if err != nil {
		return ErrDecodeRequestData
	}
	err = form.Validate()
	if err != nil {
		return ValidationError{err: err}
	}
	return nil
}

type jsonDecoder struct{}

// Decode implements the Decoder interface.
func (*jsonDecoder) Decode(req *http.Request, form Form) error {
	defer req.Body.Close()
	return json.NewDecoder(req.Body).Decode(form)
}
