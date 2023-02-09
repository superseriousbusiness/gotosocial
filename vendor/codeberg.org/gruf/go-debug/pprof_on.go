//go:build debug || debugenv
// +build debug debugenv

package debug

import (
	"net/http"
	"net/http/pprof"
	"strings"
)

// ServePprof will start an HTTP server serving /debug/pprof only if debug enabled.
func ServePprof(addr string) error {
	if !DEBUG {
		// debug disabled in env
		return nil
	}
	handler := WithPprof(nil)
	return http.ListenAndServe(addr, handler)
}

// WithPprof will add /debug/pprof handling (provided by "net/http/pprof") only if debug enabled.
func WithPprof(handler http.Handler) http.Handler {
	if !DEBUG {
		// debug disabled in env
		return handler
	}

	// Default serve mux is setup with pprof
	pprofmux := http.DefaultServeMux

	if pprofmux == nil {
		// Someone nil'ed the default mux
		pprofmux = &http.ServeMux{}
		pprofmux.HandleFunc("/debug/pprof/", pprof.Index)
		pprofmux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pprofmux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pprofmux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		pprofmux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	if handler == nil {
		// Ensure handler is non-nil
		handler = http.NotFoundHandler()
	}

	// Debug enabled, return wrapped handler func
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		const prefix = "/debug/pprof"

		// /debug/pprof(/.*)? -> pass to pprofmux
		if strings.HasPrefix(r.URL.Path, prefix) {
			path := r.URL.Path[len(prefix):]
			if path == "" || path[0] == '/' {
				pprofmux.ServeHTTP(rw, r)
				return
			}
		}

		// .* -> pass to handler
		handler.ServeHTTP(rw, r)
	})
}
