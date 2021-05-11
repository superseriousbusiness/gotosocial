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
	"crypto/rand"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"golang.org/x/crypto/acme/autocert"
)

// Router provides the REST interface for gotosocial, using gin.
type Router interface {
	// Attach a gin handler to the router with the given method and path
	AttachHandler(method string, path string, f gin.HandlerFunc)
	// Attach a gin middleware to the router that will be used globally
	AttachMiddleware(handler gin.HandlerFunc)
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

// AttachHandler attaches the given gin.HandlerFunc to the router with the specified method and path.
// If the path is set to ANY, then the handlerfunc will be used for ALL methods at its given path.
func (r *router) AttachHandler(method string, path string, handler gin.HandlerFunc) {
	if method == "ANY" {
		r.engine.Any(path, handler)
	} else {
		r.engine.Handle(method, path, handler)
	}
}

// AttachMiddleware attaches a gin middleware to the router that will be used globally
func (r *router) AttachMiddleware(middleware gin.HandlerFunc) {
	r.engine.Use(middleware)
}

// New returns a new Router with the specified configuration, using the given logrus logger.
func New(config *config.Config, logger *logrus.Logger) (Router, error) {
	lvl, err := logrus.ParseLevel(config.LogLevel)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse log level %s to set router level: %s", config.LogLevel, err)
	}
	switch lvl {
	case logrus.TraceLevel, logrus.DebugLevel:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	// create the actual engine here -- this is the core request routing handler for gts
	engine := gin.Default()
	engine.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	engine.MaxMultipartMemory = 8 << 20 // 8 MiB

	// create a new session store middleware
	store, err := sessionStore()
	if err != nil {
		return nil, fmt.Errorf("error creating session store: %s", err)
	}
	engine.Use(sessions.Sessions("gotosocial-session", store))

	// load html templates for use by the router
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %s", err)
	}
	tmPath := filepath.Join(cwd, fmt.Sprintf("%s*", config.TemplateConfig.BaseDir))
	logger.Debugf("loading templates from %s", tmPath)
	engine.LoadHTMLGlob(tmPath)

	// create the actual http server here
	s := &http.Server{
		Handler:           engine,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 30 * time.Second,
	}
	var m *autocert.Manager

	// We need to spawn the underlying server slightly differently depending on whether lets encrypt is enabled or not.
	// In either case, the gin engine will still be used for routing requests.
	if config.LetsEncryptConfig.Enabled {
		// le IS enabled, so roll up an autocert manager for handling letsencrypt requests
		m = &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.Host),
			Cache:      autocert.DirCache(config.LetsEncryptConfig.CertDir),
			Email:      config.LetsEncryptConfig.EmailAddress,
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
		config:      config,
		certManager: m,
	}, nil
}

// sessionStore returns a new session store with a random auth and encryption key.
// This means that cookies using the store will be reset if gotosocial is restarted!
func sessionStore() (memstore.Store, error) {
	auth := make([]byte, 32)
	crypt := make([]byte, 32)

	if _, err := rand.Read(auth); err != nil {
		return nil, err
	}
	if _, err := rand.Read(crypt); err != nil {
		return nil, err
	}

	return memstore.NewStore(auth, crypt), nil
}

func httpsRedirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
