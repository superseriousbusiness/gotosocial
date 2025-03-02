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

	errorsv2 "codeberg.org/gruf/go-errors/v2"
	"codeberg.org/superseriousbusiness/httpsig"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
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
	if withCode := errorsv2.AsV2[gtserror.WithCode](err); // nocollapse
	(withCode != nil && withCode.Code() >= 500) || (err != nil && withCode == nil) {
		// NOTE: the behaviour here is a little strange as we have
		// the competing code styles of the go-fed interface expecting
		// that any err is a no-go, but authed bool is intended to be
		// the main passer of whether failed auth occurred, but we in
		// the gts codebase use errors to pass-back non-200 status codes,
		// so we specifically have to check for an internal error code.
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

	otherIRIs := gtscontext.OtherIRIs(ctx)
	otherIRIStrs := make([]string, 0, len(otherIRIs))
	for _, i := range otherIRIs {
		otherIRIStrs = append(otherIRIStrs, i.String())
	}

	suite.Equal([]string{
		"http://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6/activity",
		"http://localhost:8080/users/the_mighty_zork",
		"http://fossbros-anonymous.io/users/foss_satan/statuses/5424b153-4553-4f30-9358-7b92f7cd42f6",
	}, otherIRIStrs)
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

	otherIRIs := gtscontext.OtherIRIs(ctx)
	otherIRIStrs := make([]string, 0, len(otherIRIs))
	for _, i := range otherIRIs {
		otherIRIStrs = append(otherIRIStrs, i.String())
	}

	suite.Equal([]string{
		"http://fossbros-anonymous.io/users/foss_satan/statuses/2f1195a6-5cb0-4475-adf5-92ab9a0147fe",
		"http://fossbros-anonymous.io/users/foss_satan/followers",
		"http://localhost:8080/users/1happyturtle",
	}, otherIRIStrs)
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

	otherIRIs := gtscontext.OtherIRIs(ctx)
	otherIRIStrs := make([]string, 0, len(otherIRIs))
	for _, i := range otherIRIs {
		otherIRIStrs = append(otherIRIStrs, i.String())
	}

	suite.Equal([]string{
		"http://fossbros-anonymous.io/users/foss_satan/statuses/2f1195a6-5cb0-4475-adf5-92ab9a0147fe",
		"http://fossbros-anonymous.io/users/foss_satan/followers",
		"http://localhost:8080/users/1happyturtle",
	}, otherIRIStrs)
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHookAnnounceForwardedToTurtle() {
	var (
		receivingAccount = suite.testAccounts["local_account_2"]
		activity         = suite.testActivities["announce_forwarded_1_turtle"]
	)

	ctx := suite.postInboxRequestBodyHook(
		context.Background(),
		receivingAccount,
		activity,
	)

	otherIRIs := gtscontext.OtherIRIs(ctx)
	otherIRIStrs := make([]string, 0, len(otherIRIs))
	for _, i := range otherIRIs {
		otherIRIStrs = append(otherIRIStrs, i.String())
	}

	suite.Equal([]string{
		"http://fossbros-anonymous.io/users/foss_satan/first_announce",
		"http://example.org/users/Some_User",
		"http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1",
	}, otherIRIStrs)
}

func (suite *FederatingProtocolTestSuite) TestPostInboxRequestBodyHookAnnounceForwardedToZork() {
	var (
		receivingAccount = suite.testAccounts["local_account_1"]
		activity         = suite.testActivities["announce_forwarded_2_zork"]
	)

	ctx := suite.postInboxRequestBodyHook(
		context.Background(),
		receivingAccount,
		activity,
	)

	otherIRIs := gtscontext.OtherIRIs(ctx)
	otherIRIStrs := make([]string, 0, len(otherIRIs))
	for _, i := range otherIRIs {
		otherIRIStrs = append(otherIRIStrs, i.String())
	}

	suite.Equal([]string{
		"http://fossbros-anonymous.io/users/foss_satan/second_announce",
		"http://example.org/users/Some_User",
		"http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1",
	}, otherIRIStrs)
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

func (suite *FederatingProtocolTestSuite) TestAuthenticatePostInboxKeyExpired() {
	var (
		ctx              = context.Background()
		activity         = suite.testActivities["dm_for_zork"]
		receivingAccount = suite.testAccounts["local_account_1"]
	)

	// Update remote account to mark key as expired.
	remoteAcct := &gtsmodel.Account{}
	*remoteAcct = *suite.testAccounts["remote_account_1"]
	remoteAcct.PublicKeyExpiresAt = testrig.TimeMustParse("2022-06-10T15:22:08Z")
	if err := suite.state.DB.UpdateAccount(ctx, remoteAcct, "public_key_expires_at"); err != nil {
		suite.FailNow(err.Error())
	}

	ctx, authed, resp, code := suite.authenticatePostInbox(
		ctx,
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

func (suite *FederatingProtocolTestSuite) blocked(
	ctx context.Context,
	receivingAccount *gtsmodel.Account,
	requestingAccount *gtsmodel.Account,
	otherIRIs []*url.URL,
	actorIRIs []*url.URL,
) (bool, error) {
	ctx = gtscontext.SetReceivingAccount(ctx, receivingAccount)
	ctx = gtscontext.SetRequestingAccount(ctx, requestingAccount)
	ctx = gtscontext.SetOtherIRIs(ctx, otherIRIs)
	return suite.federator.Blocked(ctx, actorIRIs)
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

	blocked, err := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.NoError(err)
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

	blocked, err := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.NoError(err)
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

	blocked, err := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.EqualError(err, "block exists between http://localhost:8080/users/the_mighty_zork and one or more of [http://example.org/users/Some_User]")
	suite.False(blocked)
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

	blocked, err := suite.blocked(
		context.Background(),
		receivingAccount,
		requestingAccount,
		otherIRIs,
		actorIRIs,
	)

	suite.EqualError(err, "block exists between http://fossbros-anonymous.io/users/foss_satan and one or more of [http://localhost:8080/users/1happyturtle/statuses/01F8MHBQCBTDKN6X5VHGMMN4MA]")
	suite.False(blocked)
}

func TestFederatingProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(FederatingProtocolTestSuite))
}
