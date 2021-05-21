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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
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
	mediaModule "github.com/superseriousbusiness/gotosocial/internal/api/client/media"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/status"
	"github.com/superseriousbusiness/gotosocial/internal/api/security"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gotosocial"
)

// Run creates and starts a gotosocial testrig server
var Run action.GTSAction = func(ctx context.Context, _ *config.Config, log *logrus.Logger) error {
	c := NewTestConfig()
	dbService := NewTestDB()
	federatingDB := NewTestFederatingDB(dbService)
	router := NewTestRouter()
	storageBackend := NewTestStorage()

	typeConverter := NewTestTypeConverter(dbService)
	transportController := NewTestTransportController(NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		r := ioutil.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}))
	federator := federation.NewFederator(dbService, federatingDB, transportController, c, log, typeConverter)
	processor := NewTestProcessor(dbService, storageBackend, federator)
	if err := processor.Start(); err != nil {
		return fmt.Errorf("error starting processor: %s", err)
	}

	StandardDBSetup(dbService)
	StandardStorageSetup(storageBackend, "./testrig/media")

	// build client api modules
	authModule := auth.New(c, dbService, NewTestOauthServer(dbService), log)
	accountModule := account.New(c, processor, log)
	appsModule := app.New(c, processor, log)
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
		appsModule,
		mm,
		fileServerModule,
		adminModule,
		statusModule,
	}

	for _, m := range apis {
		if err := m.Route(router); err != nil {
			return fmt.Errorf("routing error: %s", err)
		}
	}

	gts, err := gotosocial.New(dbService, router, federator, c)
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
