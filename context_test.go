package mux

import (
	"io"
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
	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "test"
	req.Header.Set("X-Request-ID", want)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	have, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(have) != want {
		t.Fatalf("request id\nhave %q\nwant %q", have, want)
	}
}

func TestLocale(t *testing.T) {
	want := "zh-TW"
	h := New(WithLocales([]language.Tag{language.MustParse(want)}))
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
	have, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(have) != want {
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
	have, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "en"
	if string(have) != want {
		t.Fatalf("locale\nhave %q\nwant %q", have, want)
	}
}

// https://github.com/golang/go/issues/24211
func TestLocaleDefaultAcceptRegion(t *testing.T) {
	h := New()
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
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	have, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "en"
	if string(have) != want {
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
	have, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "index"
	if string(have) != want {
		t.Fatalf("match\nhave %q\nwant %q", have, want)
	}
}
