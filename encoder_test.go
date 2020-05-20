package mux

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncode(t *testing.T) {
	h := New()
	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/", nil)
	err := h.Encode(w, req, testData{N: 1}, http.StatusTeapot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp := w.Result()
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusTeapot)
	assertHeader(t, resp, "Content-Type", "application/json; charset=utf-8")
	var data testData
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = data.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestErrEncodeMatch(t *testing.T) {
	h := New()
	w := httptest.NewRecorder()
	req := newTestRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept", "text/html")
	err := h.Encode(w, req, testData{N: 1}, http.StatusTeapot)
	if err != ErrEncodeMatch {
		t.Fatalf("unexpected error: %v", err)
	}
}
