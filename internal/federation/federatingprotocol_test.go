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

package federation_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-fed/httpsig"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/worker"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatingProtocolTestSuite struct {
	FederatorStandardTestSuite
}

// make sure PostInboxRequestBodyHook properly sets the inbox username and activity on the context
func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHook() {
	// the activity we're gonna use
	activity := suite.testActivities["dm_for_zork"]

	fedWorker := worker.New[messages.FromFederator](-1, -1)

	// setup transport controller with a no-op client so we don't make external calls
	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, nil
	}), suite.db, fedWorker)
	// setup module being tested
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	// setup request
	ctx := context.Background()
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users/the_mighty_zork/inbox", nil) // the endpoint we're hitting
	request.Header.Set("Signature", activity.SignatureHeader)

	// trigger the function being tested, and return the new context it creates
	newContext, err := federator.PostInboxRequestBodyHook(ctx, request, activity.Activity)
	suite.NoError(err)
	suite.NotNil(newContext)

	// activity should be set on context now
	activityI := newContext.Value(ap.ContextActivity)
	suite.NotNil(activityI)
	returnedActivity, ok := activityI.(pub.Activity)
	suite.True(ok)
	suite.NotNil(returnedActivity)
	suite.EqualValues(activity.Activity, returnedActivity)
}

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostInbox() {
	// the activity we're gonna use
	activity := suite.testActivities["dm_for_zork"]
	sendingAccount := suite.testAccounts["remote_account_1"]
	inboxAccount := suite.testAccounts["local_account_1"]

	fedWorker := worker.New[messages.FromFederator](-1, -1)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db, fedWorker)
	// now setup module being tested, with the mock transport controller
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users/the_mighty_zork/inbox", nil)
	// we need these headers for the request to be validated
	request.Header.Set("Signature", activity.SignatureHeader)
	request.Header.Set("Date", activity.DateHeader)
	request.Header.Set("Digest", activity.DigestHeader)

	verifier, err := httpsig.NewVerifier(request)
	suite.NoError(err)

	ctx := context.Background()
	// by the time AuthenticatePostInbox is called, PostInboxRequestBodyHook should have already been called,
	// which should have set the account and username onto the request. We can replicate that behavior here:
	ctxWithAccount := context.WithValue(ctx, ap.ContextReceivingAccount, inboxAccount)
	ctxWithActivity := context.WithValue(ctxWithAccount, ap.ContextActivity, activity)
	ctxWithVerifier := context.WithValue(ctxWithActivity, ap.ContextRequestingPublicKeyVerifier, verifier)
	ctxWithSignature := context.WithValue(ctxWithVerifier, ap.ContextRequestingPublicKeySignature, activity.SignatureHeader)

	// we can pass this recorder as a writer and read it back after
	recorder := httptest.NewRecorder()

	// trigger the function being tested, and return the new context it creates
	newContext, authed, err := federator.AuthenticatePostInbox(ctxWithSignature, recorder, request)
	suite.NoError(err)
	suite.True(authed)

	// since we know this account already it should be set on the context
	requestingAccountI := newContext.Value(ap.ContextRequestingAccount)
	suite.NotNil(requestingAccountI)
	requestingAccount, ok := requestingAccountI.(*gtsmodel.Account)
	suite.True(ok)
	suite.Equal(sendingAccount.Username, requestingAccount.Username)
}

func TestFederatingProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(FederatingProtocolTestSuite))
}
