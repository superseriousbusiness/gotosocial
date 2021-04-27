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

package testrig

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
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/admin"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/app"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/auth"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/fileserver"
	mediaModule "github.com/superseriousbusiness/gotosocial/internal/apimodule/media"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/security"
	"github.com/superseriousbusiness/gotosocial/internal/apimodule/status"
	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
)

// Run creates and starts a gotosocial testrig server
var Run action.GTSAction = func(ctx context.Context, _ *config.Config, log *logrus.Logger) error {
	dbService := NewTestDB()
	router := NewTestRouter()
	storageBackend := NewTestStorage()
	mediaHandler := NewTestMediaHandler(dbService, storageBackend)
	oauthServer := NewTestOauthServer(dbService)
	distributor := NewTestDistributor()
	if err := distributor.Start(); err != nil {
		return fmt.Errorf("error starting distributor: %s", err)
	}
	mastoConverter := NewTestTypeConverter(dbService)

	c := NewTestConfig()

	StandardDBSetup(dbService)
	StandardStorageSetup(storageBackend, "./testrig/media")

	// build client api modules
	authModule := auth.New(oauthServer, dbService, log)
	accountModule := account.New(c, dbService, oauthServer, mediaHandler, mastoConverter, log)
	appsModule := app.New(oauthServer, dbService, mastoConverter, log)
	mm := mediaModule.New(dbService, mediaHandler, mastoConverter, c, log)
	fileServerModule := fileserver.New(c, dbService, storageBackend, log)
	adminModule := admin.New(c, dbService, mediaHandler, mastoConverter, log)
	statusModule := status.New(c, dbService, mediaHandler, mastoConverter, distributor, log)
	securityModule := security.New(c, log)

	apiModules := []apimodule.ClientAPIModule{
		// modules with middleware go first
		securityModule,
		authModule,

		// now everything else
		accountModule,
		appsModule,
		mm,
		fileServerModule,
		adminModule,
		statusModule,
	}

	for _, m := range apiModules {
		if err := m.Route(router); err != nil {
			return fmt.Errorf("routing error: %s", err)
		}
		if err := m.CreateTables(dbService); err != nil {
			return fmt.Errorf("table creation error: %s", err)
		}
	}

	// if err := dbService.CreateInstanceAccount(); err != nil {
	// 	return fmt.Errorf("error creating instance account: %s", err)
	// }

	gts, err := gotosocial.New(dbService, &cache.MockCache{}, router, federation.New(dbService, c, log), c)
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

	StandardDBTeardown(dbService)
	StandardStorageTeardown(storageBackend)

	// close down all running services in order
	if err := gts.Stop(ctx); err != nil {
		return fmt.Errorf("error closing gotosocial service: %s", err)
	}

	log.Info("done! exiting...")
	return nil
}
