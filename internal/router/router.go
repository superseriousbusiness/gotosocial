/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"golang.org/x/crypto/acme/autocert"
)

var (
	readTimeout       = 60 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 30 * time.Second
	readHeaderTimeout = 30 * time.Second
)

// Router provides the REST interface for gotosocial, using gin.
type Router interface {
	// Attach a gin handler to the router with the given method and path
	AttachHandler(method string, path string, f gin.HandlerFunc)
	// Attach a gin middleware to the router that will be used globally
	AttachMiddleware(handler gin.HandlerFunc)
	// Attach 404 NoRoute handler
	AttachNoRouteHandler(handler gin.HandlerFunc)
	// Add Gin StaticFile handler 
	AttachStaticFS(relativePath string, fs http.FileSystem)
	// Start the router
	Start()
	// Stop the router
	Stop(ctx context.Context) error
}

// router fulfils the Router interface using gin and logrus
type router struct {
	logger      *logrus.Logger
	engine      *gin.Engine
	srv         *http.Server
	config      *config.Config
	certManager *autocert.Manager
}

// Add Gin StaticFile handler 
func (r *router) AttachStaticFS(relativePath string, fs http.FileSystem) {
	r.engine.StaticFS(relativePath, fs);
}

// Start starts the router nicely.
//
// Different ports and handlers will be served depending on whether letsencrypt is enabled or not.
// If it is enabled, then port 80 will be used for handling LE requests, and port 443 will be used
// for serving actual requests.
//
// If letsencrypt is not being used, then port 8080 only will be used for serving requests.
func (r *router) Start() {
	if r.config.LetsEncryptConfig.Enabled {
		// serve the http handler on port 80 for receiving letsencrypt requests and solving their devious riddles
		go func() {
			if err := http.ListenAndServe(":http", r.certManager.HTTPHandler(http.HandlerFunc(httpsRedirect))); err != nil && err != http.ErrServerClosed {
				r.logger.Fatalf("listen: %s", err)
			}
		}()

		// and serve the actual TLS handler on port 443
		go func() {
			if err := r.srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				r.logger.Fatalf("listen: %s", err)
			}
		}()
	} else {
		// no tls required so just serve on port 8080
		go func() {
			if err := r.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				r.logger.Fatalf("listen: %s", err)
			}
		}()
	}
}

// Stop shuts down the router nicely
func (r *router) Stop(ctx context.Context) error {
	return r.srv.Shutdown(ctx)
}

// New returns a new Router with the specified configuration, using the given logrus logger.
//
// The given DB is only used in the New function for parsing config values, and is not otherwise
// pinned to the router.
func New(cfg *config.Config, db db.DB, logger *logrus.Logger) (Router, error) {

	// gin has different log modes; for convenience, we match the gin log mode to
	// whatever log mode has been set for logrus
	lvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse log level %s to set router level: %s", cfg.LogLevel, err)
	}
	switch lvl {
	case logrus.TraceLevel, logrus.DebugLevel:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	// create the actual engine here -- this is the core request routing handler for gts
	engine := gin.Default()
	engine.MaxMultipartMemory = 8 << 20 // 8 MiB

	// enable cors on the engine
	if err := useCors(cfg, engine); err != nil {
		return nil, err
	}

	// set template functions
	loadTemplateFunctions(engine)

	// load templates onto the engine
	if err := loadTemplates(cfg, engine); err != nil {
		return nil, err
	}

	// enable session store middleware on the engine
	if err := useSession(cfg, db, engine); err != nil {
		return nil, err
	}

	// create the http server here, passing the gin engine as handler
	s := &http.Server{
		Handler:           engine,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// We need to spawn the underlying server slightly differently depending on whether lets encrypt is enabled or not.
	// In either case, the gin engine will still be used for routing requests.

	var m *autocert.Manager
	if cfg.LetsEncryptConfig.Enabled {
		// le IS enabled, so roll up an autocert manager for handling letsencrypt requests
		m = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cfg.Host),
			Cache:      autocert.DirCache(cfg.LetsEncryptConfig.CertDir),
			Email:      cfg.LetsEncryptConfig.EmailAddress,
		}
		// and create an HTTPS server
		s.Addr = ":https"
		s.TLSConfig = m.TLSConfig()
	} else {
		// le is NOT enabled, so just serve bare requests on port 8080
		s.Addr = ":8080"
	}

	return &router{
		logger:      logger,
		engine:      engine,
		srv:         s,
		config:      cfg,
		certManager: m,
	}, nil
}

func httpsRedirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
