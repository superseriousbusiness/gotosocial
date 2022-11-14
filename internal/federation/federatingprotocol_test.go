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
	"net/url"
	"testing"

	"github.com/go-fed/httpsig"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatingProtocolTestSuite struct {
	FederatorStandardTestSuite
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHook1() {
	// the activity we're gonna use
	activity := suite.testActivities["dm_for_zork"]

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)
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

	involvedIRIsI := newContext.Value(ap.ContextOtherInvolvedIRIs)
	involvedIRIs, ok := involvedIRIsI.([]*url.URL)
	if !ok {
		suite.FailNow("couldn't get involved IRIs from context")
	}

	suite.Len(involvedIRIs, 1)
	suite.Contains(involvedIRIs, testrig.URLMustParse("http://localhost:8080/users/the_mighty_zork"))
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHook2() {
	// the activity we're gonna use
	activity := suite.testActivities["reply_to_turtle_for_zork"]

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)

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

	involvedIRIsI := newContext.Value(ap.ContextOtherInvolvedIRIs)
	involvedIRIs, ok := involvedIRIsI.([]*url.URL)
	if !ok {
		suite.FailNow("couldn't get involved IRIs from context")
	}

	suite.Len(involvedIRIs, 2)
	suite.Contains(involvedIRIs, testrig.URLMustParse("http://localhost:8080/users/1happyturtle"))
	suite.Contains(involvedIRIs, testrig.URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers"))
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHook3() {
	// the activity we're gonna use
	activity := suite.testActivities["reply_to_turtle_for_turtle"]

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)

	// setup module being tested
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	// setup request
	ctx := context.Background()
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users/1happyturtle/inbox", nil) // the endpoint we're hitting
	request.Header.Set("Signature", activity.SignatureHeader)

	// trigger the function being tested, and return the new context it creates
	newContext, err := federator.PostInboxRequestBodyHook(ctx, request, activity.Activity)
	suite.NoError(err)
	suite.NotNil(newContext)

	involvedIRIsI := newContext.Value(ap.ContextOtherInvolvedIRIs)
	involvedIRIs, ok := involvedIRIsI.([]*url.URL)
	if !ok {
		suite.FailNow("couldn't get involved IRIs from context")
	}

	suite.Len(involvedIRIs, 2)
	suite.Contains(involvedIRIs, testrig.URLMustParse("http://localhost:8080/users/1happyturtle"))
	suite.Contains(involvedIRIs, testrig.URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers"))
}

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostInbox() {
	// the activity we're gonna use
	activity := suite.testActivities["dm_for_zork"]
	sendingAccount := suite.testAccounts["remote_account_1"]
	inboxAccount := suite.testAccounts["local_account_1"]

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)

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
	ctxWithVerifier := context.WithValue(ctxWithAccount, ap.ContextRequestingPublicKeyVerifier, verifier)
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

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostGone() {
	// the activity we're gonna use
	activity := suite.testActivities["delete_https://somewhere.mysterious/users/rest_in_piss#main-key"]
	inboxAccount := suite.testAccounts["local_account_1"]

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)

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
	ctxWithVerifier := context.WithValue(ctxWithAccount, ap.ContextRequestingPublicKeyVerifier, verifier)
	ctxWithSignature := context.WithValue(ctxWithVerifier, ap.ContextRequestingPublicKeySignature, activity.SignatureHeader)

	// we can pass this recorder as a writer and read it back after
	recorder := httptest.NewRecorder()

	// trigger the function being tested, and return the new context it creates
	_, authed, err := federator.AuthenticatePostInbox(ctxWithSignature, recorder, request)
	suite.NoError(err)
	suite.False(authed)
	suite.Equal(http.StatusAccepted, recorder.Code)
}

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostGoneNoTombstoneYet() {
	// delete the relevant tombstone
	if err := suite.db.DeleteTombstone(context.Background(), suite.testTombstones["https://somewhere.mysterious/users/rest_in_piss#main-key"].ID); err != nil {
		suite.FailNow(err.Error())
	}

	// the activity we're gonna use
	activity := suite.testActivities["delete_https://somewhere.mysterious/users/rest_in_piss#main-key"]
	inboxAccount := suite.testAccounts["local_account_1"]

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)

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
	ctxWithVerifier := context.WithValue(ctxWithAccount, ap.ContextRequestingPublicKeyVerifier, verifier)
	ctxWithSignature := context.WithValue(ctxWithVerifier, ap.ContextRequestingPublicKeySignature, activity.SignatureHeader)

	// we can pass this recorder as a writer and read it back after
	recorder := httptest.NewRecorder()

	// trigger the function being tested, and return the new context it creates
	_, authed, err := federator.AuthenticatePostInbox(ctxWithSignature, recorder, request)
	suite.NoError(err)
	suite.False(authed)
	suite.Equal(http.StatusAccepted, recorder.Code)

	// there should be a tombstone in the db now for this account
	exists, err := suite.db.TombstoneExistsWithURI(ctx, "https://somewhere.mysterious/users/rest_in_piss#main-key")
	suite.NoError(err)
	suite.True(exists)
}

func (suite *FederatingProtocolTestSuite) TestBlocked1() {
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	sendingAccount := suite.testAccounts["remote_account_1"]
	inboxAccount := suite.testAccounts["local_account_1"]
	otherInvolvedIRIs := []*url.URL{}
	actorIRIs := []*url.URL{
		testrig.URLMustParse(sendingAccount.URI),
	}

	ctx := context.Background()
	ctxWithReceivingAccount := context.WithValue(ctx, ap.ContextReceivingAccount, inboxAccount)
	ctxWithRequestingAccount := context.WithValue(ctxWithReceivingAccount, ap.ContextRequestingAccount, sendingAccount)
	ctxWithOtherInvolvedIRIs := context.WithValue(ctxWithRequestingAccount, ap.ContextOtherInvolvedIRIs, otherInvolvedIRIs)

	blocked, err := federator.Blocked(ctxWithOtherInvolvedIRIs, actorIRIs)
	suite.NoError(err)
	suite.False(blocked)
}

func (suite *FederatingProtocolTestSuite) TestBlocked2() {
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	sendingAccount := suite.testAccounts["remote_account_1"]
	inboxAccount := suite.testAccounts["local_account_1"]
	otherInvolvedIRIs := []*url.URL{}
	actorIRIs := []*url.URL{
		testrig.URLMustParse(sendingAccount.URI),
	}

	ctx := context.Background()
	ctxWithReceivingAccount := context.WithValue(ctx, ap.ContextReceivingAccount, inboxAccount)
	ctxWithRequestingAccount := context.WithValue(ctxWithReceivingAccount, ap.ContextRequestingAccount, sendingAccount)
	ctxWithOtherInvolvedIRIs := context.WithValue(ctxWithRequestingAccount, ap.ContextOtherInvolvedIRIs, otherInvolvedIRIs)

	// insert a block from inboxAccount targeting sendingAccount
	if err := suite.db.Put(context.Background(), &gtsmodel.Block{
		ID:              "01G3KBEMJD4VQ2D615MPV7KTRD",
		URI:             "whatever",
		AccountID:       inboxAccount.ID,
		TargetAccountID: sendingAccount.ID,
	}); err != nil {
		suite.Fail(err.Error())
	}

	// request should be blocked now
	blocked, err := federator.Blocked(ctxWithOtherInvolvedIRIs, actorIRIs)
	suite.NoError(err)
	suite.True(blocked)
}

func (suite *FederatingProtocolTestSuite) TestBlocked3() {
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	sendingAccount := suite.testAccounts["remote_account_1"]
	inboxAccount := suite.testAccounts["local_account_1"]
	ccedAccount := suite.testAccounts["remote_account_2"]

	otherInvolvedIRIs := []*url.URL{
		testrig.URLMustParse(ccedAccount.URI),
	}
	actorIRIs := []*url.URL{
		testrig.URLMustParse(sendingAccount.URI),
	}

	ctx := context.Background()
	ctxWithReceivingAccount := context.WithValue(ctx, ap.ContextReceivingAccount, inboxAccount)
	ctxWithRequestingAccount := context.WithValue(ctxWithReceivingAccount, ap.ContextRequestingAccount, sendingAccount)
	ctxWithOtherInvolvedIRIs := context.WithValue(ctxWithRequestingAccount, ap.ContextOtherInvolvedIRIs, otherInvolvedIRIs)

	// insert a block from inboxAccount targeting CCed account
	if err := suite.db.Put(context.Background(), &gtsmodel.Block{
		ID:              "01G3KBEMJD4VQ2D615MPV7KTRD",
		URI:             "whatever",
		AccountID:       inboxAccount.ID,
		TargetAccountID: ccedAccount.ID,
	}); err != nil {
		suite.Fail(err.Error())
	}

	blocked, err := federator.Blocked(ctxWithOtherInvolvedIRIs, actorIRIs)
	suite.NoError(err)
	suite.True(blocked)
}

func (suite *FederatingProtocolTestSuite) TestBlocked4() {
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(httpClient, suite.db, fedWorker)
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	sendingAccount := suite.testAccounts["remote_account_1"]
	inboxAccount := suite.testAccounts["local_account_1"]
	repliedStatus := suite.testStatuses["local_account_2_status_1"]

	otherInvolvedIRIs := []*url.URL{
		testrig.URLMustParse(repliedStatus.URI), // this status is involved because the hypothetical activity is a reply to this status
	}
	actorIRIs := []*url.URL{
		testrig.URLMustParse(sendingAccount.URI),
	}

	ctx := context.Background()
	ctxWithReceivingAccount := context.WithValue(ctx, ap.ContextReceivingAccount, inboxAccount)
	ctxWithRequestingAccount := context.WithValue(ctxWithReceivingAccount, ap.ContextRequestingAccount, sendingAccount)
	ctxWithOtherInvolvedIRIs := context.WithValue(ctxWithRequestingAccount, ap.ContextOtherInvolvedIRIs, otherInvolvedIRIs)

	// local account 2 (replied status account) blocks sending account already so we don't need to add a block here

	blocked, err := federator.Blocked(ctxWithOtherInvolvedIRIs, actorIRIs)
	suite.NoError(err)
	suite.True(blocked)
}

func TestFederatingProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(FederatingProtocolTestSuite))
}
