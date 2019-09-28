package mux

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/text/language"
)

func TestRequestID(t *testing.T) {
	h := New()
	h.Add("/", func(w http.ResponseWriter, req *http.Request) error {
		id := RequestID(req)
		_, err := w.Write([]byte(id))
		return err
	}, WithMethod(http.MethodGet))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Equal(body, nil) {
		t.Fatalf("request id should be set")
	}
}

func TestLocale(t *testing.T) {
	want := "zh-TW"
	matcher := language.NewMatcher([]language.Tag{language.MustParse(want)})
	h := New(WithLocales(matcher))
	h.Add("/", func(w http.ResponseWriter, req *http.Request) error {
		tag := Locale(req).String()
		_, err := w.Write([]byte(tag))
		return err
	}, WithMethod(http.MethodGet))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req.Header.Set("Accept-Language", "zh-tw")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	have, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(have, []byte(want)) {
		t.Fatalf("locale\nhave %q\nwant %q", have, want)
	}
}

func TestLocaleDefault(t *testing.T) {
	h := New()
	h.Add("/", func(w http.ResponseWriter, req *http.Request) error {
		tag := Locale(req).String()
		_, err := w.Write([]byte(tag))
		return err
	}, WithMethod(http.MethodGet))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	have, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []byte("en")
	if !bytes.Equal(have, want) {
		t.Fatalf("locale\nhave %q\nwant %q", have, want)
	}
}

func TestMatch(t *testing.T) {
	h := New()
	h.Add("/", func(w http.ResponseWriter, req *http.Request) error {
		route := Match(req)
		_, err := w.Write([]byte(route.Name()))
		return err
	}, WithMethod(http.MethodGet), WithName("index"))
	server := httptest.NewServer(h)
	defer server.Close()
	client := server.Client()
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Equal(body, nil) {
		t.Fatalf("matched route should be set")
	}
}
