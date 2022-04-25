//go:build !debug && !debugenv
// +build !debug,!debugenv

package debug

import "net/http"

// ServePprof will start an HTTP server serving /debug/pprof only if debug enabled.
func ServePprof(addr string) error {
	return nil
}

// WithPprof will add /debug/pprof handling (provided by "net/http/pprof") only if debug enabled.
func WithPprof(handler http.Handler) http.Handler {
	return handler
}
