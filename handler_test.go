package mux

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type testData struct {
	N int `json:"n"`
}

func (f testData) Validate() error {
	if f.N != 1 {
		return errors.New("n must be 1")
	}
	return nil
}

func testLogger(req *http.Request, err error) {}

func testHandler(w http.ResponseWriter, req *http.Request) error {
	_, err := io.WriteString(w, req.RequestURI)
	return err
}

func newTestRequest(method, path string, body io.Reader) *http.Request {
	rc := &requestContext{id: "test"}
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	return setContext(req, rc)
}

func TestAutomaticHEAD(t *testing.T) {
	h := New()
	h.Add("/", testHandler, WithMethod(http.MethodGet))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	resp, err := client.Head(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
}

func TestAutomaticOPTIONS(t *testing.T) {
	h := New()
	h.Add("/", testHandler, WithMethod(http.MethodGet, http.MethodDelete))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	req, err := http.NewRequest(http.MethodOptions, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	assertStatus(t, resp, http.StatusOK)
	assertHeader(t, resp, "Allow", "DELETE, GET, HEAD, OPTIONS")
}

func TestMethodNotAllowed(t *testing.T) {
	h := New()
	h.Add("/", testHandler, WithMethod(http.MethodGet, http.MethodDelete))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	resp, err := client.Post(server.URL, "text/plain", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resp.Body.Close()
	assertStatus(t, resp, http.StatusMethodNotAllowed)
	assertHeader(t, resp, "Allow", "DELETE, GET, HEAD, OPTIONS")
}

func TestPanic(t *testing.T) {
	req := newTestRequest(http.MethodGet, "https://example.com/", nil)
	testPanic(t, req)
}

func TestPanicWithoutHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	testPanic(t, req)
}

func testPanic(t *testing.T, req *http.Request) {
	h := New(WithLogger(testLogger))
	h.Add("/", func(w http.ResponseWriter, req *http.Request) error {
		panic(errors.New("test"))
		return nil
	}, WithMethod(http.MethodGet))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	assertStatus(t, resp, http.StatusInternalServerError)
	var view defaultErrorView
	err = json.NewDecoder(resp.Body).Decode(&view)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.RequestID == "" {
		t.Fatalf("request id should be set")
	}
}

func TestFileServer(t *testing.T) {
	h := New()
	h.FileServer("/public/*", http.Dir("testdata"))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	resp, err := client.Get(server.URL + "/public/test_file_server")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertString(t, "file", strings.TrimSpace(string(b)), "OK")
}

func assertInt(t *testing.T, what string, have, want int) {
	if have != want {
		t.Fatalf("%s\nhave %d\nwant %d", what, have, want)
	}
}

func assertString(t *testing.T, what, have, want string) {
	if have != want {
		t.Fatalf("%s\nhave '%s'\nwant '%s'", what, have, want)
	}
}

func assertDeepEqual(t *testing.T, what string, have, want interface{}) {
	if !reflect.DeepEqual(have, want) {
		t.Fatalf("%s\nhave %#v\nwant %#v", what, have, want)
	}
}

func assertStatus(t *testing.T, resp *http.Response, want int) {
	assertInt(t, "status code", resp.StatusCode, want)
}

func assertHeader(t *testing.T, resp *http.Response, key, want string) {
	have := resp.Header.Get(key)
	assertString(t, "headers", have, want)
}
