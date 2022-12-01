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

package testrig

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
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
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/web"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

// Start creates and starts a gotosocial testrig server
var Start action.GTSAction = func(ctx context.Context) error {
	testrig.InitTestConfig()
	testrig.InitTestLog()

	dbService := testrig.NewTestDB()
	testrig.StandardDBSetup(dbService, nil)
	router := testrig.NewTestRouter(dbService)
	var storageBackend *storage.Driver
	if os.Getenv("GTS_STORAGE_BACKEND") == "s3" {
		storageBackend, _ = storage.NewS3Storage()
	} else {
		storageBackend = testrig.NewInMemoryStorage()
	}
	testrig.StandardStorageSetup(storageBackend, "./testrig/media")

	// Create client API and federator worker pools
	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	// build backend handlers
	oauthServer := testrig.NewTestOauthServer(dbService)
	transportController := testrig.NewTestTransportController(testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		r := io.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}, ""), dbService, fedWorker)
	mediaManager := testrig.NewTestMediaManager(dbService, storageBackend)
	federator := testrig.NewTestFederator(dbService, transportController, storageBackend, mediaManager, fedWorker)

	emailSender := testrig.NewEmailSender("./web/template/", nil)

	processor := testrig.NewTestProcessor(dbService, storageBackend, federator, emailSender, mediaManager, clientWorker, fedWorker)
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

	// catch shutdown signals from the operating system
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	sig := <-sigs
	log.Infof("received signal %s, shutting down", sig)

	testrig.StandardDBTeardown(dbService)
	testrig.StandardStorageTeardown(storageBackend)

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	log.Info("done! exiting...")
	return nil
}
