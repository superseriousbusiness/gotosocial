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

//go:build debug || debugenv

package testrig

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"code.superseriousbusiness.org/gotosocial/cmd/gotosocial/action"
	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/language"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/observability"
	"code.superseriousbusiness.org/gotosocial/internal/oidc"
	"code.superseriousbusiness.org/gotosocial/internal/router"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/subscriptions"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/web"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-gonic/gin"
)

// Start creates and starts a gotosocial testrig server.
// This is only enabled in debug builds, else is nil.
var Start action.GTSAction = func(ctx context.Context) error {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	var (
		// Define necessary core variables
		// before anything so we can prepare
		// defer function for safe shutdown
		// depending on what services were
		// managed to be started.

		state = new(state.State)
		route *router.Router
	)

	defer func() {
		// Stop caches with
		// background tasks.
		state.Caches.Stop()

		if route != nil {
			// We reached a point where the API router
			// was created + setup. Ensure it gets stopped
			// first to stop processing new information.
			if err := route.Stop(); err != nil {
				log.Errorf(ctx, "error stopping router: %v", err)
			}
		}

		// Stop any currently running
		// worker processes / scheduled
		// tasks from being executed.
		testrig.StopWorkers(state)

		if state.Storage != nil {
			// If storage was created, ensure torn down.
			testrig.StandardStorageTeardown(state.Storage)
		}

		if state.DB != nil {
			// Lastly, if database service was started,
			// ensure it gets closed now all else stopped.
			testrig.StandardDBTeardown(state.DB)
			if err := state.DB.Close(); err != nil {
				log.Errorf(ctx, "error stopping database: %v", err)
			}
		}

		// Finally reached end of shutdown.
		log.Info(ctx, "done! exiting...")
	}()

	parsedLangs, err := language.InitLangs(config.GetInstanceLanguages().TagStrs())
	if err != nil {
		return fmt.Errorf("error initializing languages: %w", err)
	}
	config.SetInstanceLanguages(parsedLangs)

	if err := observability.InitializeTracing(ctx); err != nil {
		return fmt.Errorf("error initializing tracing: %w", err)
	}

	// Initialize caches and database
	state.DB = testrig.NewTestDB(state)

	// Set Actions on state, providing workers to
	// Actions as well for triggering side effects.
	state.AdminActions = admin.New(state.DB, &state.Workers)

	// New test db inits caches so we don't need to do
	// that twice, we can just start the initialized caches.
	state.Caches.Start()

	testrig.StandardDBSetup(state.DB, nil)

	// Get the instance account (we'll need this later).
	instanceAccount, err := state.DB.GetInstanceAccount(ctx, "")
	if err != nil {
		return fmt.Errorf("error retrieving instance account: %w", err)
	}

	if os.Getenv("GTS_STORAGE_BACKEND") == "s3" {
		var err error
		state.Storage, err = storage.NewS3Storage()
		if err != nil {
			return fmt.Errorf("error initializing storage: %w", err)
		}
	} else {
		state.Storage = testrig.NewInMemoryStorage()
	}
	testrig.StandardStorageSetup(state.Storage, "./testrig/media")

	// build backend handlers
	httpClient := testrig.NewMockHTTPClient(nil, "./testrig/media")
	transportController := testrig.NewTestTransportController(state, httpClient)
	mediaManager := testrig.NewTestMediaManager(state)
	federator := testrig.NewTestFederator(state, transportController, mediaManager)

	emailSender := testrig.NewEmailSender("./web/template/", nil)
	webPushSender := testrig.NewWebPushMockSender()
	typeConverter := typeutils.NewConverter(state)

	processor := testrig.NewTestProcessor(state, federator, emailSender, webPushSender, mediaManager)

	// Initialize workers.
	testrig.StartWorkers(state, processor.Workers())
	defer testrig.StopWorkers(state)

	// Initialize metrics.
	if err := observability.InitializeMetrics(ctx, state.DB); err != nil {
		return fmt.Errorf("error initializing metrics: %w", err)
	}

	// Run advanced migrations.
	if err := processor.AdvancedMigrations().Migrate(ctx); err != nil {
		return err
	}

	/*
		HTTP router initialization
	*/

	route = testrig.NewTestRouter(state.DB)
	middlewares := []gin.HandlerFunc{
		middleware.AddRequestID(config.GetRequestIDHeader()), // requestID middleware must run before tracing
	}
	if config.GetTracingEnabled() {
		middlewares = append(middlewares, observability.TracingMiddleware())
	}

	if config.GetMetricsEnabled() {
		middlewares = append(middlewares, observability.MetricsMiddleware())
	}

	middlewares = append(middlewares, []gin.HandlerFunc{
		middleware.Logger(config.GetLogClientIP()),
		middleware.HeaderFilter(state),
		middleware.UserAgent(),
		middleware.CORS(),
		middleware.ExtraHeaders(),
	}...)

	// Instantiate Content-Security-Policy
	// middleware, with extra URIs.
	cspExtraURIs := make([]string, 0)

	// Probe storage to check if extra URI is needed in CSP.
	// Error here means something is wrong with storage.
	storageCSPUri, err := state.Storage.ProbeCSPUri(ctx)
	if err != nil {
		return fmt.Errorf("error deriving Content-Security-Policy uri from storage: %w", err)
	}

	// storageCSPUri may be empty string if
	// not S3-backed storage; check for this.
	if storageCSPUri != "" {
		cspExtraURIs = append(cspExtraURIs, storageCSPUri)
	}

	// Add any extra CSP URIs from config.
	cspExtraURIs = append(cspExtraURIs, config.GetAdvancedCSPExtraURIs()...)

	// Add CSP to middlewares.
	middlewares = append(middlewares, middleware.ContentSecurityPolicy(cspExtraURIs...))

	// attach global middlewares which are used for every request
	route.AttachGlobalMiddleware(middlewares...)

	// attach global no route / 404 handler to the router
	route.AttachNoRouteHandler(func(c *gin.Context) {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound))), processor.InstanceGetV1)
	})

	// build router modules
	var idp oidc.IDP
	if config.GetOIDCEnabled() {
		idp, err = oidc.NewIDP(ctx)
		if err != nil {
			return fmt.Errorf("error creating oidc idp: %w", err)
		}
	}

	routerSession, err := state.DB.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving router session for session middleware: %w", err)
	}

	sessionName, err := middleware.SessionName()
	if err != nil {
		return fmt.Errorf("error generating session name for session middleware: %w", err)
	}

	// Configure our instance cookie policy.
	cookiePolicy := apiutil.NewCookiePolicy()

	var (
		authModule        = api.NewAuth(state, processor, idp, routerSession, sessionName, cookiePolicy) // auth/oauth paths
		clientModule      = api.NewClient(state, processor)                                              // api client endpoints
		healthModule      = api.NewHealth(state.DB.Ready)                                                // Health check endpoints
		fileserverModule  = api.NewFileserver(processor)                                                 // fileserver endpoints
		robotsModule      = api.NewRobots()                                                              // robots.txt endpoint
		wellKnownModule   = api.NewWellKnown(processor)                                                  // .well-known endpoints
		nodeInfoModule    = api.NewNodeInfo(processor)                                                   // nodeinfo endpoint
		activityPubModule = api.NewActivityPub(state.DB, processor)                                      // ActivityPub endpoints
		webModule         = web.New(state.DB, processor, cookiePolicy)                                   // web pages + user profiles + settings panels etc
	)

	// these should be routed in order
	authModule.Route(route)
	clientModule.Route(route)
	healthModule.Route(route)
	fileserverModule.Route(route)
	fileserverModule.RouteEmojis(route, instanceAccount.ID)
	robotsModule.Route(route)
	wellKnownModule.Route(route)
	nodeInfoModule.Route(route)
	activityPubModule.Route(route)
	activityPubModule.RoutePublicKey(route)
	webModule.Route(route)

	// Create background cleaner.
	cleaner := cleaner.New(state)

	// Schedule background cleaning tasks.
	if err := cleaner.ScheduleJobs(); err != nil {
		return fmt.Errorf("error scheduling cleaner jobs: %w", err)
	}

	// Create subscriptions fetcher.
	subscriptions := subscriptions.New(
		state,
		transportController,
		typeConverter,
	)

	// Schedule background subscriptions updating.
	if err := subscriptions.ScheduleJobs(); err != nil {
		return fmt.Errorf("error scheduling subscriptions jobs: %w", err)
	}

	// Finally start the main http server!
	if err := route.Start(); err != nil {
		return fmt.Errorf("error starting router: %w", err)
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	sig := <-sigs
	log.Infof(ctx, "received signal %s, shutting down", sig)

	return nil
}
