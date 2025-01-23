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

package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/api/activitypub/users"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InboxPostTestSuite struct {
	UserStandardTestSuite
}

func (suite *InboxPostTestSuite) inboxPost(
	activity pub.Activity,
	requestingAccount *gtsmodel.Account,
	targetAccount *gtsmodel.Account,
	expectedHTTPStatus int,
	expectedBody string,
	middlewares ...func(*gin.Context),
) {
	var (
		recorder = httptest.NewRecorder()
		ctx, _   = testrig.CreateGinTestContext(recorder, nil)
	)

	// Prepare the requst body bytes.
	bodyI, err := ap.Serialize(activity)
	if err != nil {
		suite.FailNow(err.Error())
	}

	b, err := json.MarshalIndent(bodyI, "", "  ")
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.T().Logf("prepared POST body:\n%s", string(b))

	// Prepare signature headers for this Activity.
	signature, digestHeader, dateHeader := testrig.GetSignatureForActivity(
		activity,
		requestingAccount.PublicKeyURI,
		requestingAccount.PrivateKey,
		testrig.URLMustParse(targetAccount.InboxURI),
	)

	// Put the request together.
	ctx.AddParam(users.UsernameKey, targetAccount.Username)
	ctx.Request = httptest.NewRequest(http.MethodPost, targetAccount.InboxURI, bytes.NewReader(b))
	ctx.Request.Header.Set("Signature", signature)
	ctx.Request.Header.Set("Date", dateHeader)
	ctx.Request.Header.Set("Digest", digestHeader)
	ctx.Request.Header.Set("Content-Type", "application/activity+json")

	// Pass the context through provided middlewares.
	for _, middleware := range middlewares {
		middleware(ctx)
	}

	// Trigger the function being tested.
	suite.userModule.InboxPOSTHandler(ctx)

	// Read the result.
	result := recorder.Result()
	defer result.Body.Close()

	b, err = io.ReadAll(result.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	errs := gtserror.NewMultiError(2)

	// Check expected code + body.
	if resultCode := recorder.Code; expectedHTTPStatus != resultCode {
		errs.Appendf("expected %d got %d", expectedHTTPStatus, resultCode)
	}

	// If we got an expected body, return early.
	if expectedBody != "" && string(b) != expectedBody {
		errs.Appendf("expected %s got %s", expectedBody, string(b))
	}

	if err := errs.Combine(); err != nil {
		suite.FailNow("", "%v (body %s)", err, string(b))
	}
}

func (suite *InboxPostTestSuite) newBlock(blockID string, blockingAccount *gtsmodel.Account, blockedAccount *gtsmodel.Account) vocab.ActivityStreamsBlock {
	block := streams.NewActivityStreamsBlock()

	// set the actor property to the block-ing account's URI
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI := testrig.URLMustParse(blockingAccount.URI)
	actorProp.AppendIRI(actorIRI)
	block.SetActivityStreamsActor(actorProp)

	// set the ID property to the blocks's URI
	idProp := streams.NewJSONLDIdProperty()
	idProp.Set(testrig.URLMustParse(blockID))
	block.SetJSONLDId(idProp)

	// set the object property to the target account's URI
	objectProp := streams.NewActivityStreamsObjectProperty()
	targetIRI := testrig.URLMustParse(blockedAccount.URI)
	objectProp.AppendIRI(targetIRI)
	block.SetActivityStreamsObject(objectProp)

	// set the TO property to the target account's IRI
	toProp := streams.NewActivityStreamsToProperty()
	toIRI := testrig.URLMustParse(blockedAccount.URI)
	toProp.AppendIRI(toIRI)
	block.SetActivityStreamsTo(toProp)

	return block
}

func (suite *InboxPostTestSuite) newUndo(
	originalActivity pub.Activity,
	objectF func() vocab.ActivityStreamsObjectProperty,
	to string,
	undoIRI string,
) vocab.ActivityStreamsUndo {
	undo := streams.NewActivityStreamsUndo()

	// Set the appropriate actor.
	undo.SetActivityStreamsActor(originalActivity.GetActivityStreamsActor())

	// Set the original activity uri as the 'object' property.
	undo.SetActivityStreamsObject(objectF())

	// Set the To of the undo as the target of the activity.
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(testrig.URLMustParse(to))
	undo.SetActivityStreamsTo(undoTo)

	// Set the ID property to the undo's URI.
	undoID := streams.NewJSONLDIdProperty()
	undoID.SetIRI(testrig.URLMustParse(undoIRI))
	undo.SetJSONLDId(undoID)

	return undo
}

func (suite *InboxPostTestSuite) newDelete(actorIRI string, objectIRI string, deleteIRI string) vocab.ActivityStreamsDelete {
	// create a delete
	delete := streams.NewActivityStreamsDelete()

	// set the appropriate actor on it
	deleteActor := streams.NewActivityStreamsActorProperty()
	deleteActor.AppendIRI(testrig.URLMustParse(actorIRI))
	delete.SetActivityStreamsActor(deleteActor)

	// Set 'object' property.
	deleteObject := streams.NewActivityStreamsObjectProperty()
	deleteObject.AppendIRI(testrig.URLMustParse(objectIRI))
	delete.SetActivityStreamsObject(deleteObject)

	// Set the To of the delete as public
	deleteTo := streams.NewActivityStreamsToProperty()
	deleteTo.AppendIRI(ap.PublicURI())
	delete.SetActivityStreamsTo(deleteTo)

	// set some random-ass ID for the activity
	deleteID := streams.NewJSONLDIdProperty()
	deleteID.SetIRI(testrig.URLMustParse(deleteIRI))
	delete.SetJSONLDId(deleteID)

	return delete
}

// TestPostBlock verifies that a remote account can block one of
// our instance users.
func (suite *InboxPostTestSuite) TestPostBlock() {
	var (
		requestingAccount = suite.testAccounts["remote_account_1"]
		targetAccount     = suite.testAccounts["local_account_1"]
		activityID        = requestingAccount.URI + "/some-new-activity/01FG9C441MCTW3R2W117V2PQK3"
	)

	block := suite.newBlock(activityID, requestingAccount, targetAccount)

	// Block.
	suite.inboxPost(
		block,
		requestingAccount,
		targetAccount,
		http.StatusAccepted,
		`{"status":"Accepted"}`,
		suite.signatureCheck,
	)

	// Ensure block created in the database.
	var (
		dbBlock *gtsmodel.Block
		err     error
	)

	if !testrig.WaitFor(func() bool {
		dbBlock, err = suite.db.GetBlock(context.Background(), requestingAccount.ID, targetAccount.ID)
		return err == nil && dbBlock != nil
	}) {
		suite.FailNow("timed out waiting for block to be created")
	}
}

// TestPostUnblock verifies that a remote account who blocks
// one of our instance users should be able to undo that block.
func (suite *InboxPostTestSuite) TestPostUnblock() {
	var (
		ctx               = context.Background()
		requestingAccount = suite.testAccounts["remote_account_1"]
		targetAccount     = suite.testAccounts["local_account_1"]
		blockID           = "http://fossbros-anonymous.io/blocks/01H1462TPRTVG2RTQCTSQ7N6Q0"
		undoID            = "http://fossbros-anonymous.io/some-activity/01H1463RDQNG5H98F29BXYHW6B"
	)

	// Put a block in the database so we have something to undo.
	block := &gtsmodel.Block{
		ID:              id.NewULID(),
		URI:             blockID,
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
	}
	if err := suite.db.PutBlock(ctx, block); err != nil {
		suite.FailNow(err.Error())
	}

	// Create the undo from the AS model block.
	asBlock, err := suite.tc.BlockToAS(ctx, block)
	if err != nil {
		suite.FailNow(err.Error())
	}

	undo := suite.newUndo(asBlock, func() vocab.ActivityStreamsObjectProperty {
		// Append the whole block as Object.
		op := streams.NewActivityStreamsObjectProperty()
		op.AppendActivityStreamsBlock(asBlock)
		return op
	}, targetAccount.URI, undoID)

	// Undo.
	suite.inboxPost(
		undo,
		requestingAccount,
		targetAccount,
		http.StatusAccepted,
		`{"status":"Accepted"}`,
		suite.signatureCheck,
	)

	// Ensure block removed from the database.
	if !testrig.WaitFor(func() bool {
		_, err := suite.db.GetBlockByID(ctx, block.ID)
		return errors.Is(err, db.ErrNoEntries)
	}) {
		suite.FailNow("timed out waiting for block to be removed")
	}
}

func (suite *InboxPostTestSuite) TestPostUpdate() {
	var (
		requestingAccount  = new(gtsmodel.Account)
		targetAccount      = suite.testAccounts["local_account_1"]
		updatedDisplayName = "updated display name!"
	)

	// Copy the requesting account, since we'll be changing it.
	*requestingAccount = *suite.testAccounts["remote_account_1"]

	// Update the account's display name.
	requestingAccount.DisplayName = updatedDisplayName

	// Add an emoji to the account; because we're serializing this
	// remote account from our own instance, we need to cheat a bit
	// to get the emoji to work properly, just for this test.
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *testrig.NewTestEmojis()["yell"]
	testEmoji.ImageURL = testEmoji.ImageRemoteURL // <- here's the cheat
	requestingAccount.Emojis = []*gtsmodel.Emoji{testEmoji}

	// Create an update from the account.
	accountable, err := suite.tc.AccountToAS(context.Background(), requestingAccount)
	if err != nil {
		suite.FailNow(err.Error())
	}
	update, err := suite.tc.WrapAccountableInUpdate(accountable)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Set the ID to something from fossbros anonymous.
	idProp := streams.NewJSONLDIdProperty()
	idProp.SetIRI(testrig.URLMustParse("https://fossbros-anonymous.io/updates/waaaaaaaaaaaaaaaaa"))
	update.SetJSONLDId(idProp)

	// Update.
	suite.inboxPost(
		update,
		requestingAccount,
		targetAccount,
		http.StatusAccepted,
		`{"status":"Accepted"}`,
		suite.signatureCheck,
	)

	// account should be changed in the database now
	var dbUpdatedAccount *gtsmodel.Account

	if !testrig.WaitFor(func() bool {
		// displayName should be updated
		dbUpdatedAccount, _ = suite.db.GetAccountByID(context.Background(), requestingAccount.ID)
		return dbUpdatedAccount.DisplayName == updatedDisplayName
	}) {
		suite.FailNow("timed out waiting for account update")
	}

	// emojis should be updated
	var haveUpdatedEmoji bool
	for _, emoji := range dbUpdatedAccount.Emojis {
		if emoji.Shortcode == testEmoji.Shortcode &&
			emoji.Domain == testEmoji.Domain &&
			emoji.ImageRemoteURL == emoji.ImageRemoteURL &&
			emoji.ImageStaticRemoteURL == emoji.ImageStaticRemoteURL {
			haveUpdatedEmoji = true
			break
		}
	}
	suite.True(haveUpdatedEmoji)

	// account should be freshly fetched
	suite.WithinDuration(time.Now(), dbUpdatedAccount.FetchedAt, 10*time.Second)

	// everything else should be the same as it was before
	suite.EqualValues(requestingAccount.Username, dbUpdatedAccount.Username)
	suite.EqualValues(requestingAccount.Domain, dbUpdatedAccount.Domain)
	suite.EqualValues(requestingAccount.AvatarMediaAttachmentID, dbUpdatedAccount.AvatarMediaAttachmentID)
	suite.EqualValues(requestingAccount.AvatarMediaAttachment, dbUpdatedAccount.AvatarMediaAttachment)
	suite.EqualValues(requestingAccount.AvatarRemoteURL, dbUpdatedAccount.AvatarRemoteURL)
	suite.EqualValues(requestingAccount.HeaderMediaAttachmentID, dbUpdatedAccount.HeaderMediaAttachmentID)
	suite.EqualValues(requestingAccount.HeaderMediaAttachment, dbUpdatedAccount.HeaderMediaAttachment)
	suite.EqualValues(requestingAccount.HeaderRemoteURL, dbUpdatedAccount.HeaderRemoteURL)
	suite.EqualValues(requestingAccount.Note, dbUpdatedAccount.Note)
	suite.EqualValues(requestingAccount.Memorial, dbUpdatedAccount.Memorial)
	suite.EqualValues(requestingAccount.AlsoKnownAsURIs, dbUpdatedAccount.AlsoKnownAsURIs)
	suite.EqualValues(requestingAccount.MovedToURI, dbUpdatedAccount.MovedToURI)
	suite.EqualValues(requestingAccount.Bot, dbUpdatedAccount.Bot)
	suite.EqualValues(requestingAccount.Locked, dbUpdatedAccount.Locked)
	suite.EqualValues(requestingAccount.Discoverable, dbUpdatedAccount.Discoverable)
	suite.EqualValues(requestingAccount.URI, dbUpdatedAccount.URI)
	suite.EqualValues(requestingAccount.URL, dbUpdatedAccount.URL)
	suite.EqualValues(requestingAccount.InboxURI, dbUpdatedAccount.InboxURI)
	suite.EqualValues(requestingAccount.OutboxURI, dbUpdatedAccount.OutboxURI)
	suite.EqualValues(requestingAccount.FollowingURI, dbUpdatedAccount.FollowingURI)
	suite.EqualValues(requestingAccount.FollowersURI, dbUpdatedAccount.FollowersURI)
	suite.EqualValues(requestingAccount.FeaturedCollectionURI, dbUpdatedAccount.FeaturedCollectionURI)
	suite.EqualValues(requestingAccount.ActorType, dbUpdatedAccount.ActorType)
	suite.EqualValues(requestingAccount.PublicKey, dbUpdatedAccount.PublicKey)
	suite.EqualValues(requestingAccount.PublicKeyURI, dbUpdatedAccount.PublicKeyURI)
	suite.EqualValues(requestingAccount.SensitizedAt, dbUpdatedAccount.SensitizedAt)
	suite.EqualValues(requestingAccount.SilencedAt, dbUpdatedAccount.SilencedAt)
	suite.EqualValues(requestingAccount.SuspendedAt, dbUpdatedAccount.SuspendedAt)
	suite.EqualValues(requestingAccount.SuspensionOrigin, dbUpdatedAccount.SuspensionOrigin)
}

func (suite *InboxPostTestSuite) TestPostDelete() {
	var (
		ctx               = context.Background()
		requestingAccount = suite.testAccounts["remote_account_1"]
		targetAccount     = suite.testAccounts["local_account_1"]
		activityID        = requestingAccount.URI + "/some-new-activity/01FG9C441MCTW3R2W117V2PQK3"
	)

	delete := suite.newDelete(requestingAccount.URI, requestingAccount.URI, activityID)

	// Delete.
	suite.inboxPost(
		delete,
		requestingAccount,
		targetAccount,
		http.StatusAccepted,
		`{"status":"Accepted"}`,
		suite.signatureCheck,
	)

	if !testrig.WaitFor(func() bool {
		// local account 2 blocked foss_satan, that block should be gone now
		testBlock := suite.testBlocks["local_account_2_block_remote_account_1"]
		_, err := suite.db.GetBlockByID(ctx, testBlock.ID)
		return suite.ErrorIs(err, db.ErrNoEntries)
	}) {
		suite.FailNow("timed out waiting for block to be removed")
	}

	if !testrig.WaitFor(func() bool {
		// no statuses from foss satan should be left in the database
		dbStatuses, err := suite.db.GetAccountStatuses(ctx, requestingAccount.ID, 0, false, false, "", "", false, false)
		return len(dbStatuses) == 0 && errors.Is(err, db.ErrNoEntries)
	}) {
		suite.FailNow("timed out waiting for statuses to be removed")
	}

	// Account should be stubbified.
	dbAccount, err := suite.db.GetAccountByID(ctx, requestingAccount.ID)
	suite.NoError(err)
	suite.Empty(dbAccount.Note)
	suite.Empty(dbAccount.DisplayName)
	suite.Empty(dbAccount.AvatarMediaAttachmentID)
	suite.Empty(dbAccount.AvatarRemoteURL)
	suite.Empty(dbAccount.HeaderMediaAttachmentID)
	suite.Empty(dbAccount.HeaderRemoteURL)
	suite.Empty(dbAccount.Fields)
	suite.False(*dbAccount.Discoverable)
	suite.WithinDuration(time.Now(), dbAccount.SuspendedAt, 30*time.Second)
	suite.Equal(dbAccount.ID, dbAccount.SuspensionOrigin)
}

func (suite *InboxPostTestSuite) TestPostEmptyCreate() {
	var (
		requestingAccount = suite.testAccounts["remote_account_1"]
		targetAccount     = suite.testAccounts["local_account_1"]
	)

	// Post a create with no object, this
	// should get accepted and silently dropped
	// as the lack of ID marks it as transient.
	create := streams.NewActivityStreamsCreate()

	suite.inboxPost(
		create,
		requestingAccount,
		targetAccount,
		http.StatusAccepted,
		`{"status":"Accepted"}`,
		suite.signatureCheck,
	)
}

func (suite *InboxPostTestSuite) TestPostCreateMalformedBlock() {
	var (
		blockingAcc = suite.testAccounts["remote_account_1"]
		blockedAcc  = suite.testAccounts["local_account_1"]
		activityID  = blockingAcc.URI + "/some-new-activity/01FG9C441MCTW3R2W117V2PQK3"
	)

	block := streams.NewActivityStreamsBlock()

	// set the actor property to the block-ing account's URI
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI := testrig.URLMustParse(blockingAcc.URI)
	actorProp.AppendIRI(actorIRI)
	block.SetActivityStreamsActor(actorProp)

	// set the ID property to the blocks's URI
	idProp := streams.NewJSONLDIdProperty()
	idProp.Set(testrig.URLMustParse(activityID))
	block.SetJSONLDId(idProp)

	// set the object property with MISSING block-ed URI.
	objectProp := streams.NewActivityStreamsObjectProperty()
	block.SetActivityStreamsObject(objectProp)

	// set the TO property to the target account's IRI
	toProp := streams.NewActivityStreamsToProperty()
	toIRI := testrig.URLMustParse(blockedAcc.URI)
	toProp.AppendIRI(toIRI)
	block.SetActivityStreamsTo(toProp)

	suite.inboxPost(
		block,
		blockingAcc,
		blockedAcc,
		http.StatusBadRequest,
		`{"error":"Bad Request: malformed incoming activity"}`,
		suite.signatureCheck,
	)
}

func (suite *InboxPostTestSuite) TestPostFromBlockedAccount() {
	var (
		requestingAccount = suite.testAccounts["remote_account_1"]
		targetAccount     = suite.testAccounts["local_account_2"]
	)

	// Create an update from the account.
	accountable, err := suite.tc.AccountToAS(context.Background(), requestingAccount)
	if err != nil {
		suite.FailNow(err.Error())
	}
	update, err := suite.tc.WrapAccountableInUpdate(accountable)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Post an update from foss satan
	// to turtle, who blocks him.
	suite.inboxPost(
		update,
		requestingAccount,
		targetAccount,
		http.StatusForbidden,
		`{"error":"Forbidden: blocked"}`,
		suite.signatureCheck,
	)
}

func (suite *InboxPostTestSuite) TestPostFromBlockedAccountToOtherAccount() {
	var (
		requestingAccount = suite.testAccounts["remote_account_1"]
		targetAccount     = suite.testAccounts["local_account_1"]
		activity          = suite.testActivities["reply_to_turtle_for_turtle"]
		statusURI         = "http://fossbros-anonymous.io/users/foss_satan/statuses/2f1195a6-5cb0-4475-adf5-92ab9a0147fe"
	)

	// Post an reply to turtle to ZORK from remote account.
	// Turtle blocks the remote account but is only tangentially
	// related to this POST request. The response will indicate
	// accepted but the post won't actually be processed.
	suite.inboxPost(
		activity.Activity,
		requestingAccount,
		targetAccount,
		http.StatusAccepted,
		`{"status":"Accepted"}`,
		suite.signatureCheck,
	)

	_, err := suite.state.DB.GetStatusByURI(context.Background(), statusURI)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *InboxPostTestSuite) TestPostUnauthorized() {
	var (
		requestingAccount = suite.testAccounts["remote_account_1"]
		targetAccount     = suite.testAccounts["local_account_1"]
	)

	// Post an empty create.
	create := streams.NewActivityStreamsCreate()

	suite.inboxPost(
		create,
		requestingAccount,
		targetAccount,
		http.StatusUnauthorized,
		`{"error":"Unauthorized: http request wasn't signed or http signature was invalid: (verifier)"}`,
		// Omit signature check middleware.
	)
}

func TestInboxPostTestSuite(t *testing.T) {
	suite.Run(t, &InboxPostTestSuite{})
}
