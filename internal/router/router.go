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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	engine      *gin.Engine
	srv         *http.Server
	certManager *autocert.Manager
}

// Add Gin StaticFile handler
func (r *router) AttachStaticFS(relativePath string, fs http.FileSystem) {
	r.engine.StaticFS(relativePath, fs)
}

// Start starts the router nicely. It will serve two handlers if letsencrypt is enabled, and only the web/API handler if letsencrypt is not enabled.
func (r *router) Start() {
	keys := config.Keys
	leEnabled := viper.GetBool(keys.LetsEncryptEnabled)

	if leEnabled {
		bindAddress := viper.GetString(keys.BindAddress)
		lePort := viper.GetInt(keys.LetsEncryptPort)

		// serve the http handler on the selected letsencrypt port, for receiving letsencrypt requests and solving their devious riddles
		go func() {
			listen := fmt.Sprintf("%s:%d", bindAddress, lePort)
			logrus.Infof("letsencrypt listening on %s", listen)
			if err := http.ListenAndServe(listen, r.certManager.HTTPHandler(http.HandlerFunc(httpsRedirect))); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("letsencrypt: listen: %s", err)
			}
		}()

		// and serve the actual TLS handler
		go func() {
			logrus.Infof("listening on %s", r.srv.Addr)
			if err := r.srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("listen: %s", err)
			}
		}()
	} else {
		// no tls required
		go func() {
			logrus.Infof("listening on %s", r.srv.Addr)
			if err := r.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("listen: %s", err)
			}
		}()
	}
}

// Stop shuts down the router nicely
func (r *router) Stop(ctx context.Context) error {
	return r.srv.Shutdown(ctx)
}

// New returns a new Router with the specified configuration.
//
// The given DB is only used in the New function for parsing config values, and is not otherwise
// pinned to the router.
func New(ctx context.Context, db db.DB) (Router, error) {
	keys := config.Keys

	gin.SetMode(gin.ReleaseMode)

	// create the actual engine here -- this is the core request routing handler for gts
	engine := gin.New()

	engine.Use(gin.RecoveryWithWriter(logrus.StandardLogger().Writer()))
	engine.Use(loggingMiddleware())

	// 8 MiB
	engine.MaxMultipartMemory = 8 << 20

	// set up IP forwarding via x-forward-* headers.
	trustedProxies := viper.GetStringSlice(keys.TrustedProxies)
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
	if err := loadTemplates(engine); err != nil {
		return nil, err
	}

	// create the http server here, passing the gin engine as handler
	bindAddress := viper.GetString(keys.BindAddress)
	port := viper.GetInt(keys.Port)
	listen := fmt.Sprintf("%s:%d", bindAddress, port)
	s := &http.Server{
		Addr:              listen,
		Handler:           engine,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	// We need to spawn the underlying server slightly differently depending on whether lets encrypt is enabled or not.
	// In either case, the gin engine will still be used for routing requests.
	leEnabled := viper.GetBool(keys.LetsEncryptEnabled)

	var m *autocert.Manager
	if leEnabled {
		// le IS enabled, so roll up an autocert manager for handling letsencrypt requests
		host := viper.GetString(keys.Host)
		leCertDir := viper.GetString(keys.LetsEncryptCertDir)
		leEmailAddress := viper.GetString(keys.LetsEncryptEmailAddress)
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

func httpsRedirect(w http.ResponseWriter, req *http.Request) {
	target := "https://" + req.Host + req.URL.Path

	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}

	http.Redirect(w, req, target, http.StatusTemporaryRedirect)
}
