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

// DecoderFunc represents the ability to negotiate a Decoder
// from an incoming HTTP request. Return the error ErrDecodeContentType
// to respond with a 415 Unsupported Media Type error.
type DecoderFunc func(req *http.Request) (Decoder, error)

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
	d, err := h.decoder(req)
	if err != nil {
		return err
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

// NewContentTypeDecoder returns a DecoderFunc that returns the
// first negotiated Decoder from the request Content-Type header.
func NewContentTypeDecoder(decoders map[string]Decoder) DecoderFunc {
	fn := func(req *http.Request) (Decoder, error) {
		v := req.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(v)
		if err != nil {
			return nil, ErrDecodeContentType
		}
		d, ok := decoders[mediaType]
		if !ok {
			return nil, ErrDecodeContentType
		}
		return d, nil
	}
	return fn
}

type jsonDecoder struct{}

// Decode implements the Decoder interface.
func (*jsonDecoder) Decode(req *http.Request, form Form) error {
	defer req.Body.Close()
	return json.NewDecoder(req.Body).Decode(form)
}
