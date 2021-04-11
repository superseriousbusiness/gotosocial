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
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/action"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/account"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/app"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/auth"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/fileserver"
	mediaModule "github.com/superseriousbusiness/gotosocial/internal/apimodule/media"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/mastotypes"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// Run creates and starts a gotosocial server
var Run action.GTSAction = func(ctx context.Context, c *config.Config, log *logrus.Logger) error {
	dbService, err := db.New(ctx, c, log)
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

	// build backend handlers
	mediaHandler := media.New(c, dbService, storageBackend, log)
	oauthServer := oauth.New(dbService, log)

	// build converters and util
	mastoConverter := mastotypes.New(c, dbService)

	// build client api modules
	authModule := auth.New(oauthServer, dbService, log)
	accountModule := account.New(c, dbService, oauthServer, mediaHandler, mastoConverter, log)
	appsModule := app.New(oauthServer, dbService, mastoConverter, log)
	mm := mediaModule.New(dbService, mediaHandler, mastoConverter, c, log)
	fileServerModule := fileserver.New(c, dbService, storageBackend, log)

	apiModules := []apimodule.ClientAPIModule{
		authModule, // this one has to go first so the other modules use its middleware
		accountModule,
		appsModule,
		mm,
		fileServerModule,
	}

	for _, m := range apiModules {
		if err := m.Route(router); err != nil {
			return fmt.Errorf("routing error: %s", err)
		}
		if err := m.CreateTables(dbService); err != nil {
			return fmt.Errorf("table creation error: %s", err)
		}
	}

	gts, err := New(dbService, &cache.MockCache{}, router, federation.New(dbService, log), c)
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
