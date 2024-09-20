// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package router

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-debug"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"golang.org/x/crypto/acme/autocert"
)

const (
	readTimeout        = 60 * time.Second
	writeTimeout       = 30 * time.Second
	idleTimeout        = 30 * time.Second
	readHeaderTimeout  = 30 * time.Second
	shutdownTimeout    = 30 * time.Second
	maxMultipartMemory = int64(8 * bytesize.MiB)
)

// Router provides the HTTP REST
// interface for GoToSocial, using gin.
type Router struct {
	engine *gin.Engine
	srv    *http.Server
}

// New returns a new Router, which wraps
// an http server and gin handler engine.
//
// The router's Attach functions should be
// used *before* the router is Started.
//
// When the router's work is finished, Stop
// should be called on it to close connections
// gracefully.
//
// The provided context will be used as the base
// context for all requests passing through the
// underlying http.Server, so this should be a
// long-running context.
func New(ctx context.Context) (*Router, error) {
	// TODO: make this configurable?
	gin.SetMode(gin.ReleaseMode)

	// Create the engine here -- this is the core
	// request routing handler for GoToSocial.
	engine := gin.New()
	engine.MaxMultipartMemory = maxMultipartMemory
	engine.HandleMethodNotAllowed = true

	// Set up client IP forwarding via
	// trusted x-forwarded-* headers.
	trustedProxies := config.GetTrustedProxies()
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		return nil, err
	}

	// Attach functions used by HTML templating,
	// and load HTML templates into the engine.
	if err := LoadTemplates(engine); err != nil {
		return nil, err
	}

	// Use the passed-in cmd context as the base context for the
	// server, since we'll never want the server to live past the
	// `server start` command anyway.
	baseCtx := func(_ net.Listener) context.Context { return ctx }

	addr := fmt.Sprintf("%s:%d",
		config.GetBindAddress(),
		config.GetPort(),
	)

	// Wrap the gin engine handler in our
	// own timeout handler, to ensure we
	// don't keep very slow requests around.
	handler := timeoutHandler{engine}

	s := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		BaseContext:       baseCtx,
	}

	return &Router{
		engine: engine,
		srv:    s,
	}, nil
}

// Start starts the router nicely.
//
// It will serve two handlers if letsencrypt is enabled,
// and only the web/API handler if letsencrypt is not enabled.
func (r *Router) Start() error {
	var (
		// listen is the server start function.
		// By default this points to a regular
		// HTTP listener, but will be changed to
		// TLS if custom certs or LE are enabled.
		listen func() error
		err    error

		certFile  = config.GetTLSCertificateChain()
		keyFile   = config.GetTLSCertificateKey()
		leEnabled = config.GetLetsEncryptEnabled()
	)

	switch {
	// TLS with custom certs.
	case certFile != "":
		// During config validation we already checked
		// that either both or neither of Chain and Key
		// are set, so we can forego checking again here.
		listen, err = r.customTLS(certFile, keyFile)
		if err != nil {
			return err
		}

	// TLS with letsencrypt.
	case leEnabled:
		listen, err = r.letsEncryptTLS()
		if err != nil {
			return err
		}

	// Default listen. TLS must
	// be handled by reverse proxy.
	default:
		listen = r.srv.ListenAndServe
	}

	// Pass the server handler through a debug pprof middleware handler.
	// For standard production builds this will be a no-op, but when the
	// "debug" or "debugenv" build-tag is set pprof stats will be served
	// at the standard "/debug/pprof" URL.
	r.srv.Handler = debug.WithPprof(r.srv.Handler)
	if debug.DEBUG {
		// Profiling requires timeouts longer than 30s, so reset these.
		log.Warn(nil, "resetting http.Server{} timeout to support profiling")
		r.srv.ReadTimeout = 0
		r.srv.WriteTimeout = 0
	}

	// Start the main listener.
	go func() {
		log.Infof(nil, "listening on %s", r.srv.Addr)
		if err := listen(); err != nil && err != http.ErrServerClosed {
			log.Panicf(nil, "listen: %v", err)
		}
	}()

	return nil
}

// Stop shuts down the router nicely.
func (r *Router) Stop() error {
	log.Infof(nil, "shutting down http router with %s grace period", shutdownTimeout)
	timeout, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := r.srv.Shutdown(timeout); err != nil {
		return fmt.Errorf("error shutting down http router: %s", err)
	}

	log.Info(nil, "http router closed connections and shut down gracefully")
	return nil
}

// customTLS modifies the router's underlying
// http server to use custom TLS cert/key pair.
func (r *Router) customTLS(
	certFile string,
	keyFile string,
) (func() error, error) {
	// Load certificates from disk.
	cer, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		err = gtserror.Newf(
			"failed to load keypair from %s and %s, ensure they are "+
				"PEM-encoded and can be read by this process: %w",
			certFile, keyFile, err,
		)
		return nil, err
	}

	// Override server's TLSConfig.
	r.srv.TLSConfig = &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cer},
	}

	// Update listen function to use custom TLS.
	listen := func() error { return r.srv.ListenAndServeTLS("", "") }
	return listen, nil
}

// letsEncryptTLS modifies the router's underlying http
// server to use LetsEncrypt via an ACME Autocert manager.
//
// It also starts a listener on the configured LetsEncrypt
// port to validate LE requests.
func (r *Router) letsEncryptTLS() (func() error, error) {
	acm := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(config.GetHost()),
		Cache:      autocert.DirCache(config.GetLetsEncryptCertDir()),
		Email:      config.GetLetsEncryptEmailAddress(),
	}

	// Override server's TLSConfig.
	r.srv.TLSConfig = acm.TLSConfig()

	// Prepare a fallback handler for LetsEncrypt.
	//
	// This will redirect all non-LetsEncrypt http
	// reqs to https, preserving path and query params.
	var fallback http.HandlerFunc = func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		// Rewrite target to https.
		target := "https://" + r.Host + r.URL.Path
		if len(r.URL.RawQuery) > 0 {
			target += "?" + r.URL.RawQuery
		}

		http.Redirect(w, r, target, http.StatusTemporaryRedirect)
	}

	// Take our own copy of the HTTP server,
	// and update it to serve LetsEncrypt
	// requests via the autocert manager.
	leSrv := (*r.srv) //nolint:govet
	leSrv.Handler = acm.HTTPHandler(fallback)
	leSrv.Addr = fmt.Sprintf("%s:%d",
		config.GetBindAddress(),
		config.GetLetsEncryptPort(),
	)

	go func() {
		// Start the LetsEncrypt autocert manager HTTP server.
		log.Infof(nil, "letsencrypt listening on %s", leSrv.Addr)
		if err := leSrv.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Panicf(nil, "letsencrypt: listen: %v", err)
		}
	}()

	// Update listen function to use LetsEncrypt TLS.
	listen := func() error { return r.srv.ListenAndServeTLS("", "") }
	return listen, nil
}
