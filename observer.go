package mux

import (
	"net/http"
	"time"
)

// Observer represents the ability to observe a request.
type Observer interface {
	// Abort is called after an error is resolved to a view.
	Abort(req *http.Request)

	// Begin is called immediately after the request context is
	// initialized and before the route is dispatched.
	Begin(req *http.Request)

	// Commit is called at the end of the request. The start time of
	// the request is passed for the ability to observe latency.
	Commit(req *http.Request, t time.Time)
}

type discardObserver struct{}

func (r *discardObserver) Abort(req *http.Request)               {}
func (r *discardObserver) Begin(req *http.Request)               {}
func (r *discardObserver) Commit(req *http.Request, t time.Time) {}
