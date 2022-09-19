/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package router

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"codeberg.org/gruf/go-debug"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"golang.org/x/crypto/acme/autocert"
)

const (
	readTimeout       = 60 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 30 * time.Second
	readHeaderTimeout = 30 * time.Second
	shutdownTimeout   = 30 * time.Second
)

// Router provides the REST interface for gotosocial, using gin.
type Router interface {
	// Attach a gin handler to the router with the given method and path
	AttachHandler(method string, path string, f gin.HandlerFunc)
	// Attach a gin middleware to the router that will be used globally
	AttachMiddleware(handler gin.HandlerFunc)
	// Attach 404 NoRoute handler
	AttachNoRouteHandler(handler gin.HandlerFunc)
	// Attach a router group, and receive that group back.
	// More middlewares and handlers can then be attached on
	// the group by the caller.
	AttachGroup(path string, handlers ...gin.HandlerFunc) *gin.RouterGroup
	// Start the router
	Start()
	// Stop the router
	Stop(ctx context.Context) error
}

// router fulfils the Router interface using gin and logrus
type router struct {
	engine      *gin.Engine
	srv         *http.Server
	certManager *autocert.Manager
}

// Start starts the router nicely. It will serve two handlers if letsencrypt is enabled, and only the web/API handler if letsencrypt is not enabled.
func (r *router) Start() {
	// listen is the server start function, by
	// default pointing to regular HTTP listener,
	// but updated to TLS if LetsEncrypt is enabled.
	listen := r.srv.ListenAndServe

	if config.GetLetsEncryptEnabled() {
		// LetsEncrypt support is enabled

		// Prepare an HTTPS-redirect handler for LetsEncrypt fallback
		redirect := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				target += "?" + r.URL.RawQuery
			}
			http.Redirect(rw, r, target, http.StatusTemporaryRedirect)
		})

		go func() {
			// Take our own copy of HTTP server
			// with updated autocert manager endpoint
			srv := (*r.srv) //nolint
			srv.Handler = r.certManager.HTTPHandler(redirect)
			srv.Addr = fmt.Sprintf("%s:%d",
				config.GetBindAddress(),
				config.GetLetsEncryptPort(),
			)

			// Start the LetsEncrypt autocert manager HTTP server.
			log.Infof("letsencrypt listening on %s", srv.Addr)
			if err := srv.ListenAndServe(); err != nil &&
				err != http.ErrServerClosed {
				log.Fatalf("letsencrypt: listen: %s", err)
			}
		}()

		// TLS is enabled, update the listen function
		listen = func() error { return r.srv.ListenAndServeTLS("", "") }
	}

	// Pass the server handler through a debug pprof middleware handler.
	// For standard production builds this will be a no-op, but when the
	// "debug" or "debugenv" build-tag is set pprof stats will be served
	// at the standard "/debug/pprof" URL.
	r.srv.Handler = debug.WithPprof(r.srv.Handler)
	if debug.DEBUG() {
		// Profiling requires timeouts longer than 30s, so reset these.
		log.Warn("resetting http.Server{} timeout to support profiling")
		r.srv.ReadTimeout = 0
		r.srv.WriteTimeout = 0
	}

	// Start the main listener.
	go func() {
		log.Infof("listening on %s", r.srv.Addr)
		if err := listen(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s", err)
		}
	}()
}

// Stop shuts down the router nicely
func (r *router) Stop(ctx context.Context) error {
	log.Infof("shutting down http router with %s grace period", shutdownTimeout)
	timeout, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	if err := r.srv.Shutdown(timeout); err != nil {
		return fmt.Errorf("error shutting down http router: %s", err)
	}

	log.Info("http router closed connections and shut down gracefully")
	return nil
}

// New returns a new Router with the specified configuration.
//
// The given DB is only used in the New function for parsing config values, and is not otherwise
// pinned to the router.
func New(ctx context.Context, db db.DB) (Router, error) {
	gin.SetMode(gin.ReleaseMode)

	// create the actual engine here -- this is the core request routing handler for gts
	engine := gin.New()
	engine.Use(loggingMiddleware)

	// 8 MiB
	engine.MaxMultipartMemory = 8 << 20

	// set up IP forwarding via x-forward-* headers.
	trustedProxies := config.GetTrustedProxies()
	if err := engine.SetTrustedProxies(trustedProxies); err != nil {
		return nil, err
	}

	// enable cors on the engine
	if err := useCors(engine); err != nil {
		return nil, err
	}

	// enable gzip compression on the engine
	if err := useGzip(engine); err != nil {
		return nil, err
	}

	// enable session store middleware on the engine
	if err := useSession(ctx, db, engine); err != nil {
		return nil, err
	}

	// set template functions
	LoadTemplateFunctions(engine)

	// load templates onto the engine
	if err := LoadTemplates(engine); err != nil {
		return nil, err
	}

	// use the passed-in command context as the base context for the server,
	// since we'll never want the server to live past the command anyway
	baseCtx := func(_ net.Listener) context.Context {
		return ctx
	}

	bindAddress := config.GetBindAddress()
	port := config.GetPort()
	addr := fmt.Sprintf("%s:%d", bindAddress, port)

	s := &http.Server{
		Addr:              addr,
		Handler:           engine, // use gin engine as handler
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		BaseContext:       baseCtx,
	}

	// We need to spawn the underlying server slightly differently depending on whether lets encrypt is enabled or not.
	// In either case, the gin engine will still be used for routing requests.
	leEnabled := config.GetLetsEncryptEnabled()

	var m *autocert.Manager
	if leEnabled {
		// le IS enabled, so roll up an autocert manager for handling letsencrypt requests
		host := config.GetHost()
		leCertDir := config.GetLetsEncryptCertDir()
		leEmailAddress := config.GetLetsEncryptEmailAddress()
		m = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(host),
			Cache:      autocert.DirCache(leCertDir),
			Email:      leEmailAddress,
		}
		s.TLSConfig = m.TLSConfig()
	}

	return &router{
		engine:      engine,
		srv:         s,
		certManager: m,
	}, nil
}
