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

package v1_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/admin"
	filtersV1 "github.com/superseriousbusiness/gotosocial/internal/api/client/filters/v1"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/email"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/filter/visibility"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/media"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/storage"
	"github.com/superseriousbusiness/gotosocial/internal/stream"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FiltersTestSuite struct {
	suite.Suite
	db           db.DB
	storage      *storage.Driver
	mediaManager *media.Manager
	federator    *federation.Federator
	processor    *processing.Processor
	emailSender  email.Sender
	sentEmails   map[string]string
	state        state.State

	// standard suite models
	testTokens         map[string]*gtsmodel.Token
	testApplications   map[string]*gtsmodel.Application
	testUsers          map[string]*gtsmodel.User
	testAccounts       map[string]*gtsmodel.Account
	testStatuses       map[string]*gtsmodel.Status
	testFilters        map[string]*gtsmodel.Filter
	testFilterKeywords map[string]*gtsmodel.FilterKeyword
	testFilterStatuses map[string]*gtsmodel.FilterStatus

	// module being tested
	filtersModule *filtersV1.Module
}

func (suite *FiltersTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testFilters = testrig.NewTestFilters()
	suite.testFilterKeywords = testrig.NewTestFilterKeywords()
	suite.testFilterStatuses = testrig.NewTestFilterStatuses()
}

func (suite *FiltersTestSuite) SetupTest() {
	suite.state.Caches.Init()
	testrig.StartNoopWorkers(&suite.state)

	testrig.InitTestConfig()
	config.Config(func(cfg *config.Configuration) {
		cfg.WebAssetBaseDir = "../../../../../web/assets/"
		cfg.WebTemplateBaseDir = "../../../../../web/templates/"
	})
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage

	testrig.StartTimelines(
		&suite.state,
		visibility.NewFilter(&suite.state),
		typeutils.NewConverter(&suite.state),
	)

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../../../testrig/media")), suite.mediaManager)
	suite.sentEmails = make(map[string]string)
	suite.emailSender = testrig.NewEmailSender("../../../../../web/template/", suite.sentEmails)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)
	suite.filtersModule = filtersV1.New(suite.processor)

	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../../testrig/media")
}

func (suite *FiltersTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}

func (suite *FiltersTestSuite) openHomeStream(account *gtsmodel.Account) *stream.Stream {
	stream, err := suite.processor.Stream().Open(context.Background(), account, stream.TimelineHome)
	if err != nil {
		suite.FailNow(err.Error())
	}
	return stream
}

func (suite *FiltersTestSuite) checkStreamed(
	str *stream.Stream,
	expectMessage bool,
	expectPayload string,
	expectEventType string,
) {
	// Set a 5s timeout on context.
	ctx := context.Background()
	ctx, cncl := context.WithTimeout(ctx, time.Second*5)
	defer cncl()

	msg, ok := str.Recv(ctx)

	if expectMessage && !ok {
		suite.FailNow("expected a message but message was not received")
	}

	if !expectMessage && ok {
		suite.FailNow("expected no message but message was received")
	}

	if expectPayload != "" && msg.Payload != expectPayload {
		suite.FailNow("", "expected payload %s but payload was: %s", expectPayload, msg.Payload)
	}

	if expectEventType != "" && msg.Event != expectEventType {
		suite.FailNow("", "expected event type %s but event type was: %s", expectEventType, msg.Event)
	}
}

func TestFiltersTestSuite(t *testing.T) {
	suite.Run(t, new(FiltersTestSuite))
}
