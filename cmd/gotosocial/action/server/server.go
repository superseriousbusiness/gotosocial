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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/superseriousbusiness/gotosocial/cmd/gotosocial/action"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/app"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/auth"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/blocks"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/bookmarks"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/emoji"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/favourites"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/filter"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/followrequest"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/list"
	mediaModule "github.com/superseriousbusiness/gotosocial/internal/api/client/media"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/notification"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/search"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/streaming"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/timeline"
	userClient "github.com/superseriousbusiness/gotosocial/internal/api/client/user"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/nodeinfo"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/webfinger"
	"github.com/superseriousbusiness/gotosocial/internal/api/security"
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

	router, err := router.New(ctx, dbService)
	if err != nil {
		return fmt.Errorf("error creating router: %s", err)
	}

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

	// create and start the message processor using the other services we've created so far
	processor := processing.NewProcessor(typeConverter, federator, oauthServer, mediaManager, storage, dbService, emailSender, clientWorker, fedWorker)
	if err := processor.Start(); err != nil {
		return fmt.Errorf("error starting processor: %s", err)
	}

	idp, err := oidc.NewIDP(ctx)
	if err != nil {
		return fmt.Errorf("error creating oidc idp: %s", err)
	}

	// build web module
	webModule := web.New(processor)

	// build client api modules
	authModule := auth.New(dbService, idp, processor)
	accountModule := account.New(processor)
	instanceModule := instance.New(processor)
	appsModule := app.New(processor)
	followRequestsModule := followrequest.New(processor)
	webfingerModule := webfinger.New(processor)
	nodeInfoModule := nodeinfo.New(processor)
	usersModule := user.New(processor)
	timelineModule := timeline.New(processor)
	notificationModule := notification.New(processor)
	searchModule := search.New(processor)
	filtersModule := filter.New(processor)
	emojiModule := emoji.New(processor)
	listsModule := list.New(processor)
	mm := mediaModule.New(processor)
	fileServerModule := fileserver.New(processor)
	adminModule := admin.New(processor)
	statusModule := status.New(processor)
	bookmarksModule := bookmarks.New(processor)
	securityModule := security.New(dbService, oauthServer)
	streamingModule := streaming.New(processor)
	favouritesModule := favourites.New(processor)
	blocksModule := blocks.New(processor)
	userClientModule := userClient.New(processor)

	apis := []api.ClientModule{
		// modules with middleware go first
		securityModule,
		authModule,

		// now the web module
		webModule,

		// now everything else
		accountModule,
		instanceModule,
		appsModule,
		followRequestsModule,
		mm,
		fileServerModule,
		adminModule,
		statusModule,
		bookmarksModule,
		webfingerModule,
		nodeInfoModule,
		usersModule,
		timelineModule,
		notificationModule,
		searchModule,
		filtersModule,
		emojiModule,
		listsModule,
		streamingModule,
		favouritesModule,
		blocksModule,
		userClientModule,
	}

	for _, m := range apis {
		if err := m.Route(router); err != nil {
			return fmt.Errorf("routing error: %s", err)
		}
	}

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
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	sig := <-sigs
	log.Infof("received signal %s, shutting down", sig)

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	log.Info("done! exiting...")
	return nil
}
