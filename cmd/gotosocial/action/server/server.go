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

package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/middleware"

	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/bundb"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/federation/federatingdb"
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	gtsstorage "github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/web"
)

// Start creates and starts a gotosocial server
var Start action.GTSAction = func(ctx context.Context) error {
	var state state.State

	// Initialize caches
	state.Caches.Init()

	// Open connection to the database
	dbService, err := bundb.NewBunDBService(ctx, &state)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	// Set the state DB connection
	state.DB = dbService

	if err := dbService.CreateInstanceAccount(ctx); err != nil {
		return fmt.Errorf("error creating instance account: %s", err)
	}

	if err := dbService.CreateInstanceInstance(ctx); err != nil {
		return fmt.Errorf("error creating instance instance: %s", err)
	}

	// Create the client API and federator worker pools
	// NOTE: these MUST NOT be used until they are passed to the
	// processor and it is started. The reason being that the processor
	// sets the Worker process functions and start the underlying pools
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	federatingDB := federatingdb.New(dbService, fedWorker)

	// build converters and util
	typeConverter := typeutils.NewConverter(dbService)

	// Open the storage backend
	storage, err := gtsstorage.AutoConfig()
	if err != nil {
		return fmt.Errorf("error creating storage backend: %w", err)
	}

	// Build HTTP client (TODO: add configurables here)
	client := httpclient.New(httpclient.Config{})

	// build backend handlers
	mediaManager, err := media.NewManager(dbService, storage)
	if err != nil {
		return fmt.Errorf("error creating media manager: %s", err)
	}
	oauthServer := oauth.New(ctx, dbService)
	transportController := transport.NewController(dbService, federatingDB, &federation.Clock{}, client)
	federator := federation.NewFederator(dbService, federatingDB, transportController, typeConverter, mediaManager)

	// decide whether to create a noop email sender (won't send emails) or a real one
	var emailSender email.Sender
	if smtpHost := config.GetSMTPHost(); smtpHost != "" {
		// host is defined so create a proper sender
		emailSender, err = email.NewSender()
		if err != nil {
			return fmt.Errorf("error creating email sender: %s", err)
		}
	} else {
		// no host is defined so create a noop sender
		emailSender, err = email.NewNoopSender(nil)
		if err != nil {
			return fmt.Errorf("error creating noop email sender: %s", err)
		}
	}

	// create the message processor using the other services we've created so far
	processor := processing.NewProcessor(typeConverter, federator, oauthServer, mediaManager, storage, dbService, emailSender, clientWorker, fedWorker)
	if err := processor.Start(); err != nil {
		return fmt.Errorf("error creating processor: %s", err)
	}

	/*
		HTTP router initialization
	*/

	router, err := router.New(ctx)
	if err != nil {
		return fmt.Errorf("error creating router: %s", err)
	}

	// attach global middlewares which are used for every request
	router.AttachGlobalMiddleware(
		middleware.Logger(),
		middleware.UserAgent(),
		middleware.CORS(),
		middleware.ExtraHeaders(),
	)

	// attach global no route / 404 handler to the router
	router.AttachNoRouteHandler(func(c *gin.Context) {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(errors.New(http.StatusText(http.StatusNotFound))), processor.InstanceGet)
	})

	// build router modules
	var idp oidc.IDP
	if config.GetOIDCEnabled() {
		idp, err = oidc.NewIDP(ctx)
		if err != nil {
			return fmt.Errorf("error creating oidc idp: %w", err)
		}
	}

	routerSession, err := dbService.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving router session for session middleware: %w", err)
	}

	sessionName, err := middleware.SessionName()
	if err != nil {
		return fmt.Errorf("error generating session name for session middleware: %w", err)
	}

	var (
		authModule        = api.NewAuth(dbService, processor, idp, routerSession, sessionName) // auth/oauth paths
		clientModule      = api.NewClient(dbService, processor)                                // api client endpoints
		fileserverModule  = api.NewFileserver(processor)                                       // fileserver endpoints
		wellKnownModule   = api.NewWellKnown(processor)                                        // .well-known endpoints
		nodeInfoModule    = api.NewNodeInfo(processor)                                         // nodeinfo endpoint
		activityPubModule = api.NewActivityPub(dbService, processor)                           // ActivityPub endpoints
		webModule         = web.New(processor)                                                 // web pages + user profiles + settings panels etc
	)

	// create required middleware
	limit := config.GetAdvancedRateLimitRequests()
	gzip := middleware.Gzip()               // all except fileserver
	clLimit := middleware.RateLimit(limit)  // client api
	s2sLimit := middleware.RateLimit(limit) // server-to-server (AP)
	fsLimit := middleware.RateLimit(limit)  // fileserver / web templates

	// these should be routed in order
	authModule.Route(router, clLimit, gzip)
	clientModule.Route(router, clLimit, gzip)
	fileserverModule.Route(router, fsLimit)
	wellKnownModule.Route(router, gzip, s2sLimit)
	nodeInfoModule.Route(router, s2sLimit, gzip)
	activityPubModule.Route(router, s2sLimit, gzip)
	webModule.Route(router, fsLimit, gzip)

	gts, err := gotosocial.NewServer(dbService, router, federator, mediaManager)
	if err != nil {
		return fmt.Errorf("error creating gotosocial service: %s", err)
	}

	if err := gts.Start(ctx); err != nil {
		return fmt.Errorf("error starting gotosocial service: %s", err)
	}

	// perform initial media prune in case value of MediaRemoteCacheDays changed
	if err := processor.AdminMediaPrune(ctx, config.GetMediaRemoteCacheDays()); err != nil {
		return fmt.Errorf("error during initial media prune: %s", err)
	}

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs // block until signal received
	log.Infof("received signal %s, shutting down", sig)

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	log.Info("done! exiting...")
	return nil
}
