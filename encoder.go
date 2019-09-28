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

// ErrEncodeAccept indicates that there was no mapping
// of Accept header media type to an Encoder.
var ErrEncodeAccept = errors.New("mux: no encoder matched request")

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

func (h *Handler) encoder(req *http.Request) (Encoder, error) {
	accept := req.Header.Get("Accept")
	// mime.ParseMediaType returns an unexported error for
	// the empty string, so we short-circuit it here.
	// An exact match is great, too.
	e, ok := h.encoders[accept]
	if ok {
		return e, nil
	}
	for _, header := range strings.Split(accept, ",") {
		media, _, err := mime.ParseMediaType(header)
		if err != nil {
			return nil, err
		}
		e, ok := h.encoders[media]
		if ok {
			return e, nil
		}
	}
	return nil, ErrEncodeAccept
}

type jsonEncoder struct{}

func (*jsonEncoder) Encode(w io.Writer, view Viewable) error {
	return json.NewEncoder(w).Encode(view)
}

func (*jsonEncoder) Headers() http.Header {
	return http.Header{"Content-Type": []string{"application/json; charset=utf-8"}}
}
