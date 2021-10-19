package mux

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		checks HealthCheck
		status int
	}{
		{
			HealthCheck(nil),
			http.StatusOK,
		},
		{
			HealthCheck([]HealthChecker{}),
			http.StatusOK,
		},
		{
			HealthCheck([]HealthChecker{HealthCheckerFunc(func() error { return nil })}),
			http.StatusOK,
		},
		{
			HealthCheck([]HealthChecker{HealthCheckerFunc(func() error { return errors.New("test") })}),
			http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		h := New()
		h.Handle("/", tt.checks, WithMethod(http.MethodGet))
		server := httptest.NewServer(h)
		defer server.Close()
		client := server.Client()
		req, err := http.NewRequest(http.MethodGet, server.URL, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resp.Body.Close()
		assertStatus(t, resp, tt.status)
	}
}
