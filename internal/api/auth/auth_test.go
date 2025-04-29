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

package auth_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"

	"code.superseriousbusiness.org/gotosocial/internal/admin"
	"code.superseriousbusiness.org/gotosocial/internal/api/auth"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/email"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/media"
	"code.superseriousbusiness.org/gotosocial/internal/middleware"
	"code.superseriousbusiness.org/gotosocial/internal/oidc"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/storage"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
)

type AuthStandardTestSuite struct {
	suite.Suite
	db           db.DB
	storage      *storage.Driver
	state        state.State
	mediaManager *media.Manager
	federator    *federation.Federator
	processor    *processing.Processor
	emailSender  email.Sender
	idp          oidc.IDP

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account

	// module being tested
	authModule *auth.Module
}

const (
	sessionUserID   = "userid"
	sessionClientID = "client_id"
)

func (suite *AuthStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
}

func (suite *AuthStandardTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	suite.db = testrig.NewTestDB(&suite.state)
	suite.state.DB = suite.db
	suite.state.AdminActions = admin.New(suite.state.DB, &suite.state.Workers)
	suite.storage = testrig.NewInMemoryStorage()
	suite.state.Storage = suite.storage
	suite.mediaManager = testrig.NewTestMediaManager(&suite.state)
	suite.federator = testrig.NewTestFederator(&suite.state, testrig.NewTestTransportController(&suite.state, testrig.NewMockHTTPClient(nil, "../../../testrig/media")), suite.mediaManager)
	suite.emailSender = testrig.NewEmailSender("../../../web/template/", nil)
	suite.processor = testrig.NewTestProcessor(
		&suite.state,
		suite.federator,
		suite.emailSender,
		testrig.NewNoopWebPushSender(),
		suite.mediaManager,
	)
	suite.authModule = auth.New(&suite.state, suite.processor, suite.idp)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
	testrig.StartNoopWorkers(&suite.state)
}

func (suite *AuthStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
	testrig.StopWorkers(&suite.state)
}

func (suite *AuthStandardTestSuite) newContext(
	requestMethod string,
	requestPath string,
	requestBody []byte,
	bodyContentType string,
) (*gin.Context, *httptest.ResponseRecorder) {
	// Create the recorder and test context.
	recorder := httptest.NewRecorder()
	ctx, engine := testrig.CreateGinTestContext(recorder, nil)

	// Load templates into the engine.
	testrig.ConfigureTemplatesWithGin(engine, "../../../web/template")

	// Create the request itself.
	protocol := config.GetProtocol()
	host := config.GetHost()
	baseURI := fmt.Sprintf("%s://%s", protocol, host)
	requestURI := fmt.Sprintf("%s/%s", baseURI, requestPath)
	ctx.Request = httptest.NewRequest(
		requestMethod,
		requestURI,
		bytes.NewReader(requestBody),
	)

	// Transmit appropriate Content-Type.
	if bodyContentType != "" {
		ctx.Request.Header.Set("Content-Type", bodyContentType)
	}

	// Accept whatever, so we can use
	// this to test both HTML and JSON.
	ctx.Request.Header.Set("accept", "*/*")

	// Trigger the session middleware on the context.
	store := memstore.NewStore(make([]byte, 32), make([]byte, 32))
	store.Options(middleware.SessionOptions(apiutil.NewCookiePolicy()))
	sessionMiddleware := sessions.Sessions("gotosocial-localhost", store)
	sessionMiddleware(ctx)

	return ctx, recorder
}
