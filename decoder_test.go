package mux

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

func TestDecode(t *testing.T) {
	var form testData
	h, req := testDecodeRequest(t, testData{N: 1})
	err := h.Decode(req, &form)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestErrDecodeContentType(t *testing.T) {
	var form testData
	h, req := testDecodeRequest(t, testData{N: 1})
	req.Header.Del("Content-Type")
	err := h.Decode(req, &form)
	if err != ErrDecodeContentType {
		t.Fatalf("unexpected error: %v", err)
	}
	req.Header.Set("Content-Type", "invalid")
	err = h.Decode(req, &form)
	if err != ErrDecodeContentType {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestErrDecodeRequestData(t *testing.T) {
	var form testData
	h, req := testDecodeRequest(t, nil)
	err := h.Decode(req, &form)
	if err != ErrDecodeRequestData {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidationError(t *testing.T) {
	var form testData
	h, req := testDecodeRequest(t, testData{N: 0})
	err := h.Decode(req, &form)
	_, ok := err.(ValidationError)
	if !ok {
		t.Fatalf("unexpected error: %v", err)
	}
}

func testDecodeRequest(t *testing.T, v interface{}) (*Handler, *http.Request) {
	h := New()
	body := bytes.NewBuffer(make([]byte, 0))
	if v != nil {
		err := json.NewEncoder(body).Encode(v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	req := newTestRequest(http.MethodGet, "/", body)
	return h, req
}
