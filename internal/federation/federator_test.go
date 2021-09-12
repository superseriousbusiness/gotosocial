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

package federation_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"git.iim.gay/grufwub/go-store/kv"
	"github.com/go-fed/activity/pub"
	"github.com/go-fed/httpsig"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ProtocolTestSuite struct {
	suite.Suite
	config        *config.Config
	db            db.DB
	log           *logrus.Logger
	storage       *kv.KVStore
	typeConverter typeutils.TypeConverter
	accounts      map[string]*gtsmodel.Account
	activities    map[string]testrig.ActivityWithSignature
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *ProtocolTestSuite) SetupSuite() {
	// setup standard items
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.storage = testrig.NewTestStorage()
	suite.typeConverter = testrig.NewTestTypeConverter(suite.db)
	suite.accounts = testrig.NewTestAccounts()
	suite.activities = testrig.NewTestActivities(suite.accounts)
}

func (suite *ProtocolTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db, suite.accounts)
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *ProtocolTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

// make sure PostInboxRequestBodyHook properly sets the inbox username and activity on the context
func (suite *ProtocolTestSuite) TestPostInboxRequestBodyHook() {
	// the activity we're gonna use
	activity := suite.activities["dm_for_zork"]

	// setup transport controller with a no-op client so we don't make external calls
	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, nil
	}), suite.db)
	// setup module being tested
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db), tc, suite.config, suite.log, suite.typeConverter, testrig.NewTestMediaHandler(suite.db, suite.storage))

	// setup request
	ctx := context.Background()
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users/the_mighty_zork/inbox", nil) // the endpoint we're hitting
	request.Header.Set("Signature", activity.SignatureHeader)

	// trigger the function being tested, and return the new context it creates
	newContext, err := federator.PostInboxRequestBodyHook(ctx, request, activity.Activity)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), newContext)

	// activity should be set on context now
	activityI := newContext.Value(util.APActivity)
	assert.NotNil(suite.T(), activityI)
	returnedActivity, ok := activityI.(pub.Activity)
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), returnedActivity)
	assert.EqualValues(suite.T(), activity.Activity, returnedActivity)
}

func (suite *ProtocolTestSuite) TestAuthenticatePostInbox() {
	// the activity we're gonna use
	activity := suite.activities["dm_for_zork"]
	sendingAccount := suite.accounts["remote_account_1"]
	inboxAccount := suite.accounts["local_account_1"]

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	// now setup module being tested, with the mock transport controller
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db), tc, suite.config, suite.log, suite.typeConverter, testrig.NewTestMediaHandler(suite.db, suite.storage))

	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users/the_mighty_zork/inbox", nil)
	// we need these headers for the request to be validated
	request.Header.Set("Signature", activity.SignatureHeader)
	request.Header.Set("Date", activity.DateHeader)
	request.Header.Set("Digest", activity.DigestHeader)

	verifier, err := httpsig.NewVerifier(request)
	assert.NoError(suite.T(), err)

	ctx := context.Background()
	// by the time AuthenticatePostInbox is called, PostInboxRequestBodyHook should have already been called,
	// which should have set the account and username onto the request. We can replicate that behavior here:
	ctxWithAccount := context.WithValue(ctx, util.APAccount, inboxAccount)
	ctxWithActivity := context.WithValue(ctxWithAccount, util.APActivity, activity)
	ctxWithVerifier := context.WithValue(ctxWithActivity, util.APRequestingPublicKeyVerifier, verifier)

	// we can pass this recorder as a writer and read it back after
	recorder := httptest.NewRecorder()

	// trigger the function being tested, and return the new context it creates
	newContext, authed, err := federator.AuthenticatePostInbox(ctxWithVerifier, recorder, request)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), authed)

	// since we know this account already it should be set on the context
	requestingAccountI := newContext.Value(util.APRequestingAccount)
	assert.NotNil(suite.T(), requestingAccountI)
	requestingAccount, ok := requestingAccountI.(*gtsmodel.Account)
	assert.True(suite.T(), ok)
	assert.Equal(suite.T(), sendingAccount.Username, requestingAccount.Username)
}

func TestProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolTestSuite))
}
