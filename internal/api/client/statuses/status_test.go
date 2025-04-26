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

package statuses_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api/client/statuses"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type StatusStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	tc           *typeutils.Converter
	mediaManager *media.Manager
	federator    *federation.Federator
	emailSender  email.Sender
	processor    *processing.Processor
	storage      *storage.Driver
	state        state.State

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testFollows      map[string]*gtsmodel.Follow

	// module being tested
	statusModule *statuses.Module
}

// Normalizes a status response to a determinate
// form, and pretty-prints it to JSON.
func (suite *StatusStandardTestSuite) parseStatusResponse(
	recorder *httptest.ResponseRecorder,
) (string, *httptest.ResponseRecorder) {

	result := recorder.Result()
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	rawMap := make(map[string]any)
	if err := json.Unmarshal(data, &rawMap); err != nil {
		suite.FailNow(err.Error())
	}

	// Make status fields determinate.
	suite.determinateStatus(rawMap)

	// For readability, don't
	// escape HTML, and indent json.
	out := new(bytes.Buffer)
	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(&rawMap); err != nil {
		suite.FailNow(err.Error())
	}

	return strings.TrimSpace(out.String()), recorder
}

func (suite *StatusStandardTestSuite) determinateStatus(rawMap map[string]any) {
	// Replace any fields from the raw map that
	// aren't determinate (date, id, url, etc).
	if _, ok := rawMap["id"]; ok {
		rawMap["id"] = id.Highest
	}

	if _, ok := rawMap["uri"]; ok {
		rawMap["uri"] = "http://localhost:8080/some/determinate/url"
	}

	if _, ok := rawMap["url"]; ok {
		rawMap["url"] = "http://localhost:8080/some/determinate/url"
	}

	if _, ok := rawMap["created_at"]; ok {
		rawMap["created_at"] = "right the hell just now babyee"
	}

	// Make ID of any mentions determinate.
	if menchiesRaw, ok := rawMap["mentions"]; ok {
		menchies, ok := menchiesRaw.([]any)
		if !ok {
			suite.FailNow("couldn't coerce menchies")
		}

		for _, menchieRaw := range menchies {
			menchie, ok := menchieRaw.(map[string]any)
			if !ok {
				suite.FailNow("couldn't coerce menchie")
			}

			if _, ok := menchie["id"]; ok {
				menchie["id"] = id.Highest
			}
		}
	}

	// Make fields of any poll determinate.
	if pollRaw, ok := rawMap["poll"]; ok && pollRaw != nil {
		poll, ok := pollRaw.(map[string]any)
		if !ok {
			suite.FailNow("couldn't coerce poll")
		}

		if _, ok := poll["id"]; ok {
			poll["id"] = id.Highest
		}

		if _, ok := poll["expires_at"]; ok {
			poll["expires_at"] = "ah like you know whatever dude it's chill"
		}
	}

	// Replace account since that's not really
	// what we care about for these tests.
	if _, ok := rawMap["account"]; ok {
		rawMap["account"] = "yeah this is my account, what about it punk"
	}

	// If status contains an embedded
	// reblog do the same thing for that.
	if reblogRaw, ok := rawMap["reblog"]; ok && reblogRaw != nil {
		reblog, ok := reblogRaw.(map[string]any)
		if !ok {
			suite.FailNow("couldn't coerce reblog")
		}
		suite.determinateStatus(reblog)
	}
}

func (suite *StatusStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testFollows = testrig.NewTestFollows()
}

func (suite *StatusStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage

	suite.tc = typeutils.NewConverter(&suite.state)

	testrig.StandardDBSetup(suite.db, nil)
	testrig.StandardStorageSetup(suite.storage, "../../../../testrig/media")

	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../../testrig/media")), suite.mediaManager)
	suite.emailSender = testrig.NewEmailSender("../../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)
	suite.statusModule = statuses.New(suite.processor)

	testrig.StartWorkers(&suite.state, suite.processor.Workers())
}

func (suite *StatusStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StandardStorageTeardown(suite.storage)
	testrig.StopWorkers(&suite.state)
}
