package mux

import "net/http"

// HealthChecker represents the ability to check the health of a resource.
type HealthChecker interface {
	// Check checks the status of a resource.
	// Return a non-nil error to fail the healthcheck.
	Check() error
}

// HealthCheckFunc is an adapter to allow the use of
// ordinary functions as HealthChecks.
type HealthCheckerFunc func() error

// Check implements the HealthChecker interface.
func (fn HealthCheckerFunc) Check() error {
	return fn()
}

// HealthCheck is a basic healthcheck handler.
//
// The handler will respond with a plain text HTTP 200 OK if and only if
// all checks return non-nil. If any check fails, the handler will respond
// with a HTTP 500 Internal Server Error.
type HealthCheck []HealthChecker

// ServeHTTP implements the http.Handler interface.
func (h HealthCheck) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, checker := range h {
		err := checker.Check()
		if err != nil {
			abort(w, http.StatusInternalServerError)
			return
		}
	}
	abort(w, http.StatusOK)
}
