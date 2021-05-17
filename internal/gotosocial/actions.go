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

package gotosocial

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/action"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/account"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/admin"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/app"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/auth"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/fileserver"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/followrequest"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/instance"
	mediaModule "github.com/superseriousbusiness/gotosocial/internal/api/client/media"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/webfinger"
	"github.com/superseriousbusiness/gotosocial/internal/api/security"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db/pg"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/message"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
)

var models []interface{} = []interface{}{
	&gtsmodel.Account{},
	&gtsmodel.Application{},
	&gtsmodel.Block{},
	&gtsmodel.DomainBlock{},
	&gtsmodel.EmailDomainBlock{},
	&gtsmodel.Follow{},
	&gtsmodel.FollowRequest{},
	&gtsmodel.MediaAttachment{},
	&gtsmodel.Mention{},
	&gtsmodel.Status{},
	&gtsmodel.StatusFave{},
	&gtsmodel.StatusBookmark{},
	&gtsmodel.StatusMute{},
	&gtsmodel.Tag{},
	&gtsmodel.User{},
	&gtsmodel.Emoji{},
	&gtsmodel.Instance{},
	&oauth.Token{},
	&oauth.Client{},
}

// Run creates and starts a gotosocial server
var Run action.GTSAction = func(ctx context.Context, c *config.Config, log *logrus.Logger) error {
	dbService, err := pg.NewPostgresService(ctx, c, log)
	if err != nil {
		return fmt.Errorf("error creating dbservice: %s", err)
	}

	router, err := router.New(c, log)
	if err != nil {
		return fmt.Errorf("error creating router: %s", err)
	}

	storageBackend, err := storage.NewLocal(c, log)
	if err != nil {
		return fmt.Errorf("error creating storage backend: %s", err)
	}

	// build converters and util
	typeConverter := typeutils.NewConverter(c, dbService)

	// build backend handlers
	mediaHandler := media.New(c, dbService, storageBackend, log)
	oauthServer := oauth.New(dbService, log)
	transportController := transport.NewController(c, &federation.Clock{}, http.DefaultClient, log)
	federator := federation.NewFederator(dbService, transportController, c, log, typeConverter)
	processor := message.NewProcessor(c, typeConverter, federator, oauthServer, mediaHandler, storageBackend, dbService, log)
	if err := processor.Start(); err != nil {
		return fmt.Errorf("error starting processor: %s", err)
	}

	// build client api modules
	authModule := auth.New(c, dbService, oauthServer, log)
	accountModule := account.New(c, processor, log)
	instanceModule := instance.New(c, processor, log)
	appsModule := app.New(c, processor, log)
	followRequestsModule := followrequest.New(c, processor, log)
	webfingerModule := webfinger.New(c, processor, log)
	usersModule := user.New(c, processor, log)
	mm := mediaModule.New(c, processor, log)
	fileServerModule := fileserver.New(c, processor, log)
	adminModule := admin.New(c, processor, log)
	statusModule := status.New(c, processor, log)
	securityModule := security.New(c, log)

	apis := []api.ClientModule{
		// modules with middleware go first
		securityModule,
		authModule,

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
		usersModule,
	}

	for _, m := range apis {
		if err := m.Route(router); err != nil {
			return fmt.Errorf("routing error: %s", err)
		}
	}

	for _, m := range models {
		if err := dbService.CreateTable(m); err != nil {
			return fmt.Errorf("table creation error: %s", err)
		}
	}

	if err := dbService.CreateInstanceAccount(); err != nil {
		return fmt.Errorf("error creating instance account: %s", err)
	}

	if err := dbService.CreateInstanceInstance(); err != nil {
		return fmt.Errorf("error creating instance instance: %s", err)
	}

	gts, err := New(dbService, router, federator, c)
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

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	log.Info("done! exiting...")
	return nil
}
