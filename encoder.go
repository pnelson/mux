package mux

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

// Viewable represents a view. To provide an expressive API, this
// type is an alias for interface{} that is named for documentation.
type Viewable interface{}

// Encoder represents the ability to encode HTTP responses.
type Encoder interface {
	Encode(w io.Writer, view Viewable) error
	Headers() http.Header
}

// EncoderFunc represents the ability to negotiate an Encoder
// from an incoming HTTP request. Return the error ErrEncodeMatch
// to respond with a 406 Not Acceptable error.
type EncoderFunc func(req *http.Request) (Encoder, error)

// ErrEncodeMatch indicates that request failed to negotiate an Encoder.
var ErrEncodeMatch = errors.New("mux: no encoder matched request")

// Encode encodes the view and responds to the request.
func (h *Handler) Encode(w http.ResponseWriter, req *http.Request, view Viewable, code int) error {
	e, err := h.encoder(req)
	if err != nil {
		return err
	}
	b := h.pool.Get()
	defer h.pool.Put(b)
	err = e.Encode(b, view)
	if err != nil {
		return err
	}
	headers := w.Header()
	for k, vs := range e.Headers() {
		for _, v := range vs {
			headers.Add(k, v)
		}
	}
	w.WriteHeader(code)
	_, err = b.WriteTo(w)
	return err
}

// NewAcceptEncoder returns an EncoderFunc that returns the
// first negotiated Encoder based on the request Accept header.
func NewAcceptEncoder(e Encoder, mediaTypes []string) EncoderFunc {
	m := make(map[string]struct{})
	for _, t := range mediaTypes {
		m[t] = struct{}{}
	}
	fn := func(req *http.Request) (Encoder, error) {
		accept := req.Header.Get("Accept")
		// mime.ParseMediaType returns an unexported error for
		// the empty string, so we short-circuit it here.
		// An exact match is great, too.
		_, ok := m[accept]
		if ok {
			return e, nil
		}
		for _, t := range strings.Split(accept, ",") {
			mediaType, _, err := mime.ParseMediaType(t)
			if err != nil {
				return nil, err
			}
			_, ok := m[mediaType]
			if ok {
				return e, nil
			}
		}
		return nil, ErrEncodeMatch
	}
	return fn
}

type jsonEncoder struct{}

func (*jsonEncoder) Encode(w io.Writer, view Viewable) error {
	return json.NewEncoder(w).Encode(view)
}

func (*jsonEncoder) Headers() http.Header {
	return http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
}
