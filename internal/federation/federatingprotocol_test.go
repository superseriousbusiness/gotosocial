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

package federation_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-fed/httpsig"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type FederatingProtocolTestSuite struct {
	FederatorStandardTestSuite
}

func (suite *FederatingProtocolTestSuite) postInboxRequestBodyHook(
	ctx context.Context,
	receivingAccount *gtsmodel.Account,
	activity testrig.ActivityWithSignature,
) context.Context {
	raw, err := ap.Serialize(activity.Activity)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.Marshal(raw)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NoError(err)
	request := httptest.NewRequest(http.MethodPost, receivingAccount.InboxURI, bytes.NewBuffer(b))
	request.Header.Set("Signature", activity.SignatureHeader)
	request.Header.Set("Date", activity.DateHeader)
	request.Header.Set("Digest", activity.DigestHeader)

	newContext, err := suite.federator.PostInboxRequestBodyHook(ctx, request, activity.Activity)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return newContext
}

func (suite *FederatingProtocolTestSuite) authenticatePostInbox(
	ctx context.Context,
	receivingAccount *gtsmodel.Account,
	activity testrig.ActivityWithSignature,
) (context.Context, bool, []byte, int) {
	raw, err := ap.Serialize(activity.Activity)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.Marshal(raw)
	if err != nil {
		suite.FailNow(err.Error())
	}

	request := httptest.NewRequest(http.MethodPost, receivingAccount.InboxURI, bytes.NewBuffer(b))
	request.Header.Set("Signature", activity.SignatureHeader)
	request.Header.Set("Date", activity.DateHeader)
	request.Header.Set("Digest", activity.DigestHeader)

	verifier, err := httpsig.NewVerifier(request)
	if err != nil {
		suite.FailNow(err.Error())
	}

	ctx = gtscontext.SetReceivingAccount(ctx, receivingAccount)
	ctx = gtscontext.SetHTTPSignatureVerifier(ctx, verifier)
	ctx = gtscontext.SetHTTPSignature(ctx, activity.SignatureHeader)
	ctx = gtscontext.SetHTTPSignaturePubKeyID(ctx, testrig.URLMustParse(verifier.KeyId()))

	recorder := httptest.NewRecorder()
	newContext, authed, err := suite.federator.AuthenticatePostInbox(ctx, recorder, request)
	if err != nil {
		suite.FailNow(err.Error())
	}

	res := recorder.Result()
	defer res.Body.Close()

	b, err = io.ReadAll(res.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return newContext, authed, b, res.StatusCode
}

func (suite *FederatingProtocolTestSuite) blocked(
	ctx context.Context,
	receivingAccount *gtsmodel.Account,
	requestingAccount *gtsmodel.Account,
	otherIRIs []*url.URL,
	actorIRIs []*url.URL,
) bool {
	ctx = gtscontext.SetReceivingAccount(ctx, receivingAccount)
	ctx = gtscontext.SetRequestingAccount(ctx, requestingAccount)
	ctx = gtscontext.SetOtherIRIs(ctx, otherIRIs)

	blocked, err := suite.federator.Blocked(ctx, actorIRIs)
	if err != nil {
		suite.FailNow(err.Error())
	}

	return blocked
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHookDM() {
	var (
		receivingAccount = suite.testAccounts["local_account_1"]
		activity         = suite.testActivities["dm_for_zork"]
	)

	ctx := suite.postInboxRequestBodyHook(
		context.Background(),
		receivingAccount,
		activity,
	)

	involvedIRIs := gtscontext.OtherIRIs(ctx)
	suite.Equal([]*url.URL{
		testrig.URLMustParse("http://localhost:8080/users/the_mighty_zork"),
	}, involvedIRIs)
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHookReply() {
	var (
		receivingAccount = suite.testAccounts["local_account_1"]
		activity         = suite.testActivities["reply_to_turtle_for_zork"]
	)

	ctx := suite.postInboxRequestBodyHook(
		context.Background(),
		receivingAccount,
		activity,
	)

	involvedIRIs := gtscontext.OtherIRIs(ctx)
	suite.Equal([]*url.URL{
		testrig.URLMustParse("http://localhost:8080/users/1happyturtle"),
		testrig.URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers"),
	}, involvedIRIs)
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHookReplyToReply() {
	var (
		receivingAccount = suite.testAccounts["local_account_2"]
		activity         = suite.testActivities["reply_to_turtle_for_turtle"]
	)

	ctx := suite.postInboxRequestBodyHook(
		context.Background(),
		receivingAccount,
		activity,
	)

	involvedIRIs := gtscontext.OtherIRIs(ctx)
	suite.Equal([]*url.URL{
		testrig.URLMustParse("http://localhost:8080/users/1happyturtle"),
		testrig.URLMustParse("http://fossbros-anonymous.io/users/foss_satan/followers"),
	}, involvedIRIs)
}

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostInbox() {
	var (
		activity         = suite.testActivities["dm_for_zork"]
		receivingAccount = suite.testAccounts["local_account_1"]
	)

	ctx, authed, resp, code := suite.authenticatePostInbox(
		context.Background(),
		receivingAccount,
		activity,
	)

	suite.NotNil(gtscontext.RequestingAccount(ctx))
	suite.True(authed)
	suite.Equal([]byte{}, resp)
	suite.Equal(http.StatusOK, code)
}

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostGoneWithTombstone() {
	var (
		activity         = suite.testActivities["delete_https://somewhere.mysterious/users/rest_in_piss#main-key"]
		receivingAccount = suite.testAccounts["local_account_1"]
	)

	ctx, authed, resp, code := suite.authenticatePostInbox(
		context.Background(),
		receivingAccount,
		activity,
	)

	// Tombstone exists for this account, should simply return accepted.
	suite.Nil(gtscontext.RequestingAccount(ctx))
	suite.False(authed)
	suite.Equal([]byte{}, resp)
	suite.Equal(http.StatusAccepted, code)
}

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostGoneNoTombstone() {
	var (
		activity         = suite.testActivities["delete_https://somewhere.mysterious/users/rest_in_piss#main-key"]
		receivingAccount = suite.testAccounts["local_account_1"]
		testTombstone    = suite.testTombstones["https://somewhere.mysterious/users/rest_in_piss#main-key"]
	)

	// Delete the tombstone; it'll have to be created again.
	if err := suite.state.DB.DeleteTombstone(context.Background(), testTombstone.ID); err != nil {
		suite.FailNow(err.Error())
	}

	ctx, authed, resp, code := suite.authenticatePostInbox(
		context.Background(),
		receivingAccount,
		activity,
	)

	suite.Nil(gtscontext.RequestingAccount(ctx))
	suite.False(authed)
	suite.Equal([]byte{}, resp)
	suite.Equal(http.StatusAccepted, code)

	// Tombstone should be back, baby!
	exists, err := suite.state.DB.TombstoneExistsWithURI(
		context.Background(),
		"https://somewhere.mysterious/users/rest_in_piss#main-key",
	)
	suite.NoError(err)
	suite.True(exists)
}

func (suite *FederatingProtocolTestSuite) TestBlockedNoProblem() {
	var (
		receivingAccount  = suite.testAccounts["local_account_1"]
		requestingAccount = suite.testAccounts["remote_account_1"]
		otherIRIs         = []*url.URL{}
		actorIRIs         = []*url.URL{
			testrig.URLMustParse(requestingAccount.URI),
		}
	)

	blocked := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.False(blocked)
}

func (suite *FederatingProtocolTestSuite) TestBlockedReceiverBlocksRequester() {
	var (
		receivingAccount  = suite.testAccounts["local_account_1"]
		requestingAccount = suite.testAccounts["remote_account_1"]
		otherIRIs         = []*url.URL{}
		actorIRIs         = []*url.URL{
			testrig.URLMustParse(requestingAccount.URI),
		}
	)

	// Insert a block from receivingAccount targeting requestingAccount.
	if err := suite.state.DB.PutBlock(context.Background(), &gtsmodel.Block{
		ID:              "01G3KBEMJD4VQ2D615MPV7KTRD",
		URI:             "whatever",
		AccountID:       receivingAccount.ID,
		TargetAccountID: requestingAccount.ID,
	}); err != nil {
		suite.Fail(err.Error())
	}

	blocked := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.True(blocked)
}

func (suite *FederatingProtocolTestSuite) TestBlockedCCd() {
	var (
		receivingAccount  = suite.testAccounts["local_account_1"]
		requestingAccount = suite.testAccounts["remote_account_1"]
		ccedAccount       = suite.testAccounts["remote_account_2"]
		otherIRIs         = []*url.URL{
			testrig.URLMustParse(ccedAccount.URI),
		}
		actorIRIs = []*url.URL{
			testrig.URLMustParse(requestingAccount.URI),
		}
	)

	// Insert a block from receivingAccount targeting ccedAccount.
	if err := suite.state.DB.PutBlock(context.Background(), &gtsmodel.Block{
		ID:              "01G3KBEMJD4VQ2D615MPV7KTRD",
		URI:             "whatever",
		AccountID:       receivingAccount.ID,
		TargetAccountID: ccedAccount.ID,
	}); err != nil {
		suite.Fail(err.Error())
	}

	blocked := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.True(blocked)
}

func (suite *FederatingProtocolTestSuite) TestBlockedRepliedStatus() {
	var (
		receivingAccount  = suite.testAccounts["local_account_1"]
		requestingAccount = suite.testAccounts["remote_account_1"]
		repliedStatus     = suite.testStatuses["local_account_2_status_1"]
		otherIRIs         = []*url.URL{
			// This status is involved because the
			// hypothetical activity replies to it.
			testrig.URLMustParse(repliedStatus.URI),
		}
		actorIRIs = []*url.URL{
			testrig.URLMustParse(requestingAccount.URI),
		}
	)

	blocked := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.True(blocked)
}

func TestFederatingProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(FederatingProtocolTestSuite))
}
