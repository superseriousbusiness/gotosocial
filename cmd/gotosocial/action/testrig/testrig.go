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

package testrig

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	tlprocessor "github.com/superseriousbusiness/gotosocial/internal/processing/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/tracing"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/web"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

// Start creates and starts a gotosocial testrig server
var Start action.GTSAction = func(ctx context.Context) error {
	var state state.State

	testrig.InitTestConfig()
	testrig.InitTestLog()

	if err := tracing.Initialize(); err != nil {
		return fmt.Errorf("error initializing tracing: %w", err)
	}

	// Initialize caches and database
	state.DB = testrig.NewTestDB(&state)

	// New test db inits caches so we don't need to do
	// that twice, we can just start the initialized caches.
	state.Caches.Start()
	defer state.Caches.Stop()

	testrig.StandardDBSetup(state.DB, nil)

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

	// Initialize workers.
	state.Workers.Start()
	defer state.Workers.Stop()

	// build backend handlers
	transportController := testrig.NewTestTransportController(&state, testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		r := io.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}, ""))
	mediaManager := testrig.NewTestMediaManager(&state)
	federator := testrig.NewTestFederator(&state, transportController, mediaManager)

	emailSender := testrig.NewEmailSender("./web/template/", nil)
	typeConverter := typeutils.NewConverter(&state)
	filter := visibility.NewFilter(&state)

	// Initialize timelines.
	state.Timelines.Home = timeline.NewManager(
		tlprocessor.HomeTimelineGrab(&state),
		tlprocessor.HomeTimelineFilter(&state, filter),
		tlprocessor.HomeTimelineStatusPrepare(&state, typeConverter),
		tlprocessor.SkipInsert(),
	)
	if err := state.Timelines.Home.Start(); err != nil {
		return fmt.Errorf("error starting home timeline: %s", err)
	}

	state.Timelines.List = timeline.NewManager(
		tlprocessor.ListTimelineGrab(&state),
		tlprocessor.ListTimelineFilter(&state, filter),
		tlprocessor.ListTimelineStatusPrepare(&state, typeConverter),
		tlprocessor.SkipInsert(),
	)
	if err := state.Timelines.List.Start(); err != nil {
		return fmt.Errorf("error starting list timeline: %s", err)
	}

	processor := testrig.NewTestProcessor(&state, federator, emailSender, mediaManager)

	/*
		HTTP router initialization
	*/

	router := testrig.NewTestRouter(state.DB)
	middlewares := []gin.HandlerFunc{
		middleware.AddRequestID(config.GetRequestIDHeader()), // requestID middleware must run before tracing
	}
	if config.GetTracingEnabled() {
		middlewares = append(middlewares, tracing.InstrumentGin())
	}
	middlewares = append(middlewares, []gin.HandlerFunc{
		middleware.Logger(config.GetLogClientIP()),
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
	router.AttachGlobalMiddleware(middlewares...)

	// attach global no route / 404 handler to the router
	router.AttachNoRouteHandler(func(c *gin.Context) {
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

	var (
		authModule        = api.NewAuth(state.DB, processor, idp, routerSession, sessionName) // auth/oauth paths
		clientModule      = api.NewClient(state.DB, processor)                                // api client endpoints
		fileserverModule  = api.NewFileserver(processor)                                      // fileserver endpoints
		wellKnownModule   = api.NewWellKnown(processor)                                       // .well-known endpoints
		nodeInfoModule    = api.NewNodeInfo(processor)                                        // nodeinfo endpoint
		activityPubModule = api.NewActivityPub(state.DB, processor)                           // ActivityPub endpoints
		webModule         = web.New(state.DB, processor)                                      // web pages + user profiles + settings panels etc
	)

	// these should be routed in order
	authModule.Route(router)
	clientModule.Route(router)
	fileserverModule.Route(router)
	wellKnownModule.Route(router)
	nodeInfoModule.Route(router)
	activityPubModule.Route(router)
	activityPubModule.RoutePublicKey(router)
	webModule.Route(router)

	gts, err := gotosocial.NewServer(state.DB, router, federator, mediaManager)
	if err != nil {
		return fmt.Errorf("error creating gotosocial service: %s", err)
	}

	if err := gts.Start(ctx); err != nil {
		return fmt.Errorf("error starting gotosocial service: %s", err)
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	sig := <-sigs
	log.Infof(ctx, "received signal %s, shutting down", sig)

	testrig.StandardDBTeardown(state.DB)
	testrig.StandardStorageTeardown(state.Storage)

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	log.Info(ctx, "done! exiting...")
	return nil
}
