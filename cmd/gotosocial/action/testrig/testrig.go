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
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/web"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

// Start creates and starts a gotosocial testrig server
var Start action.GTSAction = func(ctx context.Context) error {
	var state state.State

	testrig.InitTestConfig()
	testrig.InitTestLog()

	// Initialize caches
	state.Caches.Init()
	state.Caches.Start()
	defer state.Caches.Stop()

	state.DB = testrig.NewTestDB(&state)
	testrig.StandardDBSetup(state.DB, nil)

	if os.Getenv("GTS_STORAGE_BACKEND") == "s3" {
		state.Storage, _ = storage.NewS3Storage()
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

	processor := testrig.NewTestProcessor(&state, federator, emailSender, mediaManager)
	if err := processor.Start(); err != nil {
		return fmt.Errorf("error starting processor: %s", err)
	}

	/*
		HTTP router initialization
	*/

	router := testrig.NewTestRouter(state.DB)

	// attach global middlewares which are used for every request
	router.AttachGlobalMiddleware(
		middleware.Logger(),
		middleware.UserAgent(),
		middleware.CORS(),
		middleware.ExtraHeaders(),
	)

	// attach global no route / 404 handler to the router
	router.AttachNoRouteHandler(func(c *gin.Context) {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound))), processor.InstanceGetV1)
	})

	// build router modules
	var idp oidc.IDP
	var err error
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
