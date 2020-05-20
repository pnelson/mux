package mux

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAbort(t *testing.T) {
	var tests = map[int]error{
		http.StatusBadRequest:           ErrDecodeRequestData, // 400
		http.StatusNotFound:             ErrNotFound,          // 404
		http.StatusNotAcceptable:        ErrEncodeMatch,       // 406
		http.StatusUnsupportedMediaType: ErrDecodeContentType, // 415
		http.StatusInternalServerError:  errors.New("test"),   // 500
	}
	h := New(WithLogger(testLogger))
	for want, err := range tests {
		w := httptest.NewRecorder()
		req := newTestRequest(http.MethodGet, "/", nil)
		h.Abort(w, req, err)
		resp := w.Result()
		resp.Body.Close()
		assertStatus(t, resp, want)
	}
}

func TestAbortErrMethodNotAllowed(t *testing.T) {
	h := New()
	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodPost, "/", nil)
	h.Abort(w, req, ErrMethodNotAllowed([]string{"DELETE", "GET", "HEAD", "OPTIONS"}))
	resp := w.Result()
	resp.Body.Close()
	assertStatus(t, resp, http.StatusMethodNotAllowed)
	assertHeader(t, resp, "Allow", "DELETE, GET, HEAD, OPTIONS")
}

func TestAbortValidationError(t *testing.T) {
	h := New()
	w := httptest.NewRecorder()
	body := bytes.NewBuffer(make([]byte, 0))
	err := json.NewEncoder(body).Encode(testData{N: 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req := newTestRequest(http.MethodGet, "/", body)
	var form testData
	err = h.Decode(req, &form)
	_, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("unexpected error: %v", err)
	}
	h.Abort(w, req, err)
	resp := w.Result()
	resp.Body.Close()
	assertStatus(t, resp, http.StatusUnprocessableEntity)
}
