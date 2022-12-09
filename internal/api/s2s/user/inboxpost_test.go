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

package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/pub"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InboxPostTestSuite struct {
	UserStandardTestSuite
}

func (suite *InboxPostTestSuite) TestPostBlock() {
	blockingAccount := suite.testAccounts["remote_account_1"]
	blockedAccount := suite.testAccounts["local_account_1"]
	blockURI := testrig.URLMustParse("http://fossbros-anonymous.io/users/foss_satan/blocks/01FG9C441MCTW3R2W117V2PQK3")

	block := streams.NewActivityStreamsBlock()

	// set the actor property to the block-ing account's URI
	actorProp := streams.NewActivityStreamsActorProperty()
	actorIRI := testrig.URLMustParse(blockingAccount.URI)
	actorProp.AppendIRI(actorIRI)
	block.SetActivityStreamsActor(actorProp)

	// set the ID property to the blocks's URI
	idProp := streams.NewJSONLDIdProperty()
	idProp.Set(blockURI)
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

	targetURI := testrig.URLMustParse(blockedAccount.InboxURI)

	signature, digestHeader, dateHeader := testrig.GetSignatureForActivity(block, blockingAccount.PublicKeyURI, blockingAccount.PrivateKey, targetURI)
	bodyI, err := streams.Serialize(block)
	suite.NoError(err)

	bodyJson, err := json.Marshal(bodyI)
	suite.NoError(err)
	body := bytes.NewReader(bodyJson)

	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../../testrig/media"), suite.db, fedWorker)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage, suite.mediaManager, fedWorker)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender, suite.mediaManager, clientWorker, fedWorker)
	userModule := user.New(processor).(*user.Module)
	suite.NoError(processor.Start())

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodPost, targetURI.String(), body) // the endpoint we're hitting
	ctx.Request.Header.Set("Signature", signature)
	ctx.Request.Header.Set("Date", dateHeader)
	ctx.Request.Header.Set("Digest", digestHeader)
	ctx.Request.Header.Set("Content-Type", "application/activity+json")

	// we need to pass the context through signature check first to set appropriate values on it
	suite.securityModule.SignatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
			Value: blockedAccount.Username,
		},
	}

	// trigger the function being tested
	userModule.InboxPOSTHandler(ctx)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Empty(b)

	// there should be a block in the database now between the accounts
	dbBlock, err := suite.db.GetBlock(context.Background(), blockingAccount.ID, blockedAccount.ID)
	suite.NoError(err)
	suite.NotNil(dbBlock)
	suite.WithinDuration(time.Now(), dbBlock.CreatedAt, 30*time.Second)
	suite.WithinDuration(time.Now(), dbBlock.UpdatedAt, 30*time.Second)
	suite.Equal("http://fossbros-anonymous.io/users/foss_satan/blocks/01FG9C441MCTW3R2W117V2PQK3", dbBlock.URI)
}

// TestPostUnblock verifies that a remote account with a block targeting one of our instance users should be able to undo that block.
func (suite *InboxPostTestSuite) TestPostUnblock() {
	blockingAccount := suite.testAccounts["remote_account_1"]
	blockedAccount := suite.testAccounts["local_account_1"]

	// first put a block in the database so we have something to undo
	blockURI := "http://fossbros-anonymous.io/users/foss_satan/blocks/01FG9C441MCTW3R2W117V2PQK3"
	dbBlockID, err := id.NewRandomULID()
	suite.NoError(err)

	dbBlock := &gtsmodel.Block{
		ID:              dbBlockID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             blockURI,
		AccountID:       blockingAccount.ID,
		TargetAccountID: blockedAccount.ID,
	}

	err = suite.db.PutBlock(context.Background(), dbBlock)
	suite.NoError(err)

	asBlock, err := suite.tc.BlockToAS(context.Background(), dbBlock)
	suite.NoError(err)

	targetAccountURI := testrig.URLMustParse(blockedAccount.URI)

	// create an Undo and set the appropriate actor on it
	undo := streams.NewActivityStreamsUndo()
	undo.SetActivityStreamsActor(asBlock.GetActivityStreamsActor())

	// Set the block as the 'object' property.
	undoObject := streams.NewActivityStreamsObjectProperty()
	undoObject.AppendActivityStreamsBlock(asBlock)
	undo.SetActivityStreamsObject(undoObject)

	// Set the To of the undo as the target of the block
	undoTo := streams.NewActivityStreamsToProperty()
	undoTo.AppendIRI(targetAccountURI)
	undo.SetActivityStreamsTo(undoTo)

	undoID := streams.NewJSONLDIdProperty()
	undoID.SetIRI(testrig.URLMustParse("http://fossbros-anonymous.io/72cc96a3-f742-4daf-b9f5-3407667260c5"))
	undo.SetJSONLDId(undoID)

	targetURI := testrig.URLMustParse(blockedAccount.InboxURI)

	signature, digestHeader, dateHeader := testrig.GetSignatureForActivity(undo, blockingAccount.PublicKeyURI, blockingAccount.PrivateKey, targetURI)
	bodyI, err := streams.Serialize(undo)
	suite.NoError(err)

	bodyJson, err := json.Marshal(bodyI)
	suite.NoError(err)
	body := bytes.NewReader(bodyJson)

	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../../testrig/media"), suite.db, fedWorker)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage, suite.mediaManager, fedWorker)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender, suite.mediaManager, clientWorker, fedWorker)
	userModule := user.New(processor).(*user.Module)
	suite.NoError(processor.Start())

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodPost, targetURI.String(), body) // the endpoint we're hitting
	ctx.Request.Header.Set("Signature", signature)
	ctx.Request.Header.Set("Date", dateHeader)
	ctx.Request.Header.Set("Digest", digestHeader)
	ctx.Request.Header.Set("Content-Type", "application/activity+json")

	// we need to pass the context through signature check first to set appropriate values on it
	suite.securityModule.SignatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
			Value: blockedAccount.Username,
		},
	}

	// trigger the function being tested
	userModule.InboxPOSTHandler(ctx)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Empty(b)
	suite.Equal(http.StatusOK, result.StatusCode)

	// the block should be undone
	block, err := suite.db.GetBlock(context.Background(), blockingAccount.ID, blockedAccount.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(block)
}

func (suite *InboxPostTestSuite) TestPostUpdate() {
	updatedAccount := *suite.testAccounts["remote_account_1"]
	updatedAccount.DisplayName = "updated display name!"

	// ad an emoji to the account; because we're serializing this remote
	// account from our own instance, we need to cheat a bit to get the emoji
	// to work properly, just for this test
	testEmoji := &gtsmodel.Emoji{}
	*testEmoji = *testrig.NewTestEmojis()["yell"]
	testEmoji.ImageURL = testEmoji.ImageRemoteURL // <- here's the cheat
	updatedAccount.Emojis = []*gtsmodel.Emoji{testEmoji}

	asAccount, err := suite.tc.AccountToAS(context.Background(), &updatedAccount)
	suite.NoError(err)

	receivingAccount := suite.testAccounts["local_account_1"]

	// create an update
	update := streams.NewActivityStreamsUpdate()

	// set the appropriate actor on it
	updateActor := streams.NewActivityStreamsActorProperty()
	updateActor.AppendIRI(testrig.URLMustParse(updatedAccount.URI))
	update.SetActivityStreamsActor(updateActor)

	// Set the account as the 'object' property.
	updateObject := streams.NewActivityStreamsObjectProperty()
	updateObject.AppendActivityStreamsPerson(asAccount)
	update.SetActivityStreamsObject(updateObject)

	// Set the To of the update as public
	updateTo := streams.NewActivityStreamsToProperty()
	updateTo.AppendIRI(testrig.URLMustParse(pub.PublicActivityPubIRI))
	update.SetActivityStreamsTo(updateTo)

	// set the cc of the update to the receivingAccount
	updateCC := streams.NewActivityStreamsCcProperty()
	updateCC.AppendIRI(testrig.URLMustParse(receivingAccount.URI))
	update.SetActivityStreamsCc(updateCC)

	// set some random-ass ID for the activity
	undoID := streams.NewJSONLDIdProperty()
	undoID.SetIRI(testrig.URLMustParse("http://fossbros-anonymous.io/d360613a-dc8d-4563-8f0b-b6161caf0f2b"))
	update.SetJSONLDId(undoID)

	targetURI := testrig.URLMustParse(receivingAccount.InboxURI)

	signature, digestHeader, dateHeader := testrig.GetSignatureForActivity(update, updatedAccount.PublicKeyURI, updatedAccount.PrivateKey, targetURI)
	bodyI, err := streams.Serialize(update)
	suite.NoError(err)

	bodyJson, err := json.Marshal(bodyI)
	suite.NoError(err)
	body := bytes.NewReader(bodyJson)

	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../../testrig/media"), suite.db, fedWorker)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage, suite.mediaManager, fedWorker)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender, suite.mediaManager, clientWorker, fedWorker)
	userModule := user.New(processor).(*user.Module)
	suite.NoError(processor.Start())

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodPost, targetURI.String(), body) // the endpoint we're hitting
	ctx.Request.Header.Set("Signature", signature)
	ctx.Request.Header.Set("Date", dateHeader)
	ctx.Request.Header.Set("Digest", digestHeader)
	ctx.Request.Header.Set("Content-Type", "application/activity+json")

	// we need to pass the context through signature check first to set appropriate values on it
	suite.securityModule.SignatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
			Value: receivingAccount.Username,
		},
	}

	// trigger the function being tested
	userModule.InboxPOSTHandler(ctx)

	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Empty(b)
	suite.Equal(http.StatusOK, result.StatusCode)

	// account should be changed in the database now
	var dbUpdatedAccount *gtsmodel.Account

	if !testrig.WaitFor(func() bool {
		// displayName should be updated
		dbUpdatedAccount, _ = suite.db.GetAccountByID(context.Background(), updatedAccount.ID)
		return dbUpdatedAccount.DisplayName == "updated display name!"
	}) {
		suite.FailNow("timed out waiting for account update")
	}

	// emojis should be updated
	suite.Contains(dbUpdatedAccount.EmojiIDs, testEmoji.ID)

	// account should be freshly webfingered
	suite.WithinDuration(time.Now(), dbUpdatedAccount.LastWebfingeredAt, 10*time.Second)

	// everything else should be the same as it was before
	suite.EqualValues(updatedAccount.Username, dbUpdatedAccount.Username)
	suite.EqualValues(updatedAccount.Domain, dbUpdatedAccount.Domain)
	suite.EqualValues(updatedAccount.AvatarMediaAttachmentID, dbUpdatedAccount.AvatarMediaAttachmentID)
	suite.EqualValues(updatedAccount.AvatarMediaAttachment, dbUpdatedAccount.AvatarMediaAttachment)
	suite.EqualValues(updatedAccount.AvatarRemoteURL, dbUpdatedAccount.AvatarRemoteURL)
	suite.EqualValues(updatedAccount.HeaderMediaAttachmentID, dbUpdatedAccount.HeaderMediaAttachmentID)
	suite.EqualValues(updatedAccount.HeaderMediaAttachment, dbUpdatedAccount.HeaderMediaAttachment)
	suite.EqualValues(updatedAccount.HeaderRemoteURL, dbUpdatedAccount.HeaderRemoteURL)
	suite.EqualValues(updatedAccount.Note, dbUpdatedAccount.Note)
	suite.EqualValues(updatedAccount.Memorial, dbUpdatedAccount.Memorial)
	suite.EqualValues(updatedAccount.AlsoKnownAs, dbUpdatedAccount.AlsoKnownAs)
	suite.EqualValues(updatedAccount.MovedToAccountID, dbUpdatedAccount.MovedToAccountID)
	suite.EqualValues(updatedAccount.Bot, dbUpdatedAccount.Bot)
	suite.EqualValues(updatedAccount.Reason, dbUpdatedAccount.Reason)
	suite.EqualValues(updatedAccount.Locked, dbUpdatedAccount.Locked)
	suite.EqualValues(updatedAccount.Discoverable, dbUpdatedAccount.Discoverable)
	suite.EqualValues(updatedAccount.Privacy, dbUpdatedAccount.Privacy)
	suite.EqualValues(updatedAccount.Sensitive, dbUpdatedAccount.Sensitive)
	suite.EqualValues(updatedAccount.Language, dbUpdatedAccount.Language)
	suite.EqualValues(updatedAccount.URI, dbUpdatedAccount.URI)
	suite.EqualValues(updatedAccount.URL, dbUpdatedAccount.URL)
	suite.EqualValues(updatedAccount.InboxURI, dbUpdatedAccount.InboxURI)
	suite.EqualValues(updatedAccount.OutboxURI, dbUpdatedAccount.OutboxURI)
	suite.EqualValues(updatedAccount.FollowingURI, dbUpdatedAccount.FollowingURI)
	suite.EqualValues(updatedAccount.FollowersURI, dbUpdatedAccount.FollowersURI)
	suite.EqualValues(updatedAccount.FeaturedCollectionURI, dbUpdatedAccount.FeaturedCollectionURI)
	suite.EqualValues(updatedAccount.ActorType, dbUpdatedAccount.ActorType)
	suite.EqualValues(updatedAccount.PublicKey, dbUpdatedAccount.PublicKey)
	suite.EqualValues(updatedAccount.PublicKeyURI, dbUpdatedAccount.PublicKeyURI)
	suite.EqualValues(updatedAccount.SensitizedAt, dbUpdatedAccount.SensitizedAt)
	suite.EqualValues(updatedAccount.SilencedAt, dbUpdatedAccount.SilencedAt)
	suite.EqualValues(updatedAccount.SuspendedAt, dbUpdatedAccount.SuspendedAt)
	suite.EqualValues(updatedAccount.HideCollections, dbUpdatedAccount.HideCollections)
	suite.EqualValues(updatedAccount.SuspensionOrigin, dbUpdatedAccount.SuspensionOrigin)
}

func (suite *InboxPostTestSuite) TestPostDelete() {
	deletedAccount := *suite.testAccounts["remote_account_1"]
	receivingAccount := suite.testAccounts["local_account_1"]

	// create a delete
	delete := streams.NewActivityStreamsDelete()

	// set the appropriate actor on it
	deleteActor := streams.NewActivityStreamsActorProperty()
	deleteActor.AppendIRI(testrig.URLMustParse(deletedAccount.URI))
	delete.SetActivityStreamsActor(deleteActor)

	// Set the account iri as the 'object' property.
	deleteObject := streams.NewActivityStreamsObjectProperty()
	deleteObject.AppendIRI(testrig.URLMustParse(deletedAccount.URI))
	delete.SetActivityStreamsObject(deleteObject)

	// Set the To of the delete as public
	deleteTo := streams.NewActivityStreamsToProperty()
	deleteTo.AppendIRI(testrig.URLMustParse(pub.PublicActivityPubIRI))
	delete.SetActivityStreamsTo(deleteTo)

	// set some random-ass ID for the activity
	deleteID := streams.NewJSONLDIdProperty()
	deleteID.SetIRI(testrig.URLMustParse("http://fossbros-anonymous.io/d360613a-dc8d-4563-8f0b-b6161caf0f2b"))
	delete.SetJSONLDId(deleteID)

	targetURI := testrig.URLMustParse(receivingAccount.InboxURI)

	signature, digestHeader, dateHeader := testrig.GetSignatureForActivity(delete, deletedAccount.PublicKeyURI, deletedAccount.PrivateKey, targetURI)
	bodyI, err := streams.Serialize(delete)
	suite.NoError(err)

	bodyJson, err := json.Marshal(bodyI)
	suite.NoError(err)
	body := bytes.NewReader(bodyJson)

	clientWorker := concurrency.NewWorkerPool[messages.FromClientAPI](-1, -1)
	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../../../testrig/media"), suite.db, fedWorker)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage, suite.mediaManager, fedWorker)
	emailSender := testrig.NewEmailSender("../../../../web/template/", nil)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator, emailSender, suite.mediaManager, clientWorker, fedWorker)
	suite.NoError(processor.Start())
	userModule := user.New(processor).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := testrig.CreateGinTestContext(recorder, nil)
	ctx.Request = httptest.NewRequest(http.MethodPost, targetURI.String(), body) // the endpoint we're hitting
	ctx.Request.Header.Set("Signature", signature)
	ctx.Request.Header.Set("Date", dateHeader)
	ctx.Request.Header.Set("Digest", digestHeader)
	ctx.Request.Header.Set("Content-Type", "application/activity+json")

	// we need to pass the context through signature check first to set appropriate values on it
	suite.securityModule.SignatureCheck(ctx)

	// normally the router would populate these params from the path values,
	// but because we're calling the function directly, we need to set them manually.
	ctx.Params = gin.Params{
		gin.Param{
			Key:   user.UsernameKey,
			Value: receivingAccount.Username,
		},
	}

	// trigger the function being tested
	userModule.InboxPOSTHandler(ctx)
	result := recorder.Result()
	defer result.Body.Close()
	b, err := ioutil.ReadAll(result.Body)
	suite.NoError(err)
	suite.Empty(b)
	suite.Equal(http.StatusOK, result.StatusCode)

	if !testrig.WaitFor(func() bool {
		// local account 2 blocked foss_satan, that block should be gone now
		testBlock := suite.testBlocks["local_account_2_block_remote_account_1"]
		dbBlock := &gtsmodel.Block{}
		err = suite.db.GetByID(ctx, testBlock.ID, dbBlock)
		return suite.ErrorIs(err, db.ErrNoEntries)
	}) {
		suite.FailNow("timed out waiting for block to be removed")
	}

	// no statuses from foss satan should be left in the database
	dbStatuses, err := suite.db.GetAccountStatuses(ctx, deletedAccount.ID, 0, false, false, "", "", false, false, false)
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Empty(dbStatuses)

	dbAccount, err := suite.db.GetAccountByID(ctx, deletedAccount.ID)
	suite.NoError(err)

	suite.Empty(dbAccount.Note)
	suite.Empty(dbAccount.DisplayName)
	suite.Empty(dbAccount.AvatarMediaAttachmentID)
	suite.Empty(dbAccount.AvatarRemoteURL)
	suite.Empty(dbAccount.HeaderMediaAttachmentID)
	suite.Empty(dbAccount.HeaderRemoteURL)
	suite.Empty(dbAccount.Reason)
	suite.Empty(dbAccount.Fields)
	suite.True(*dbAccount.HideCollections)
	suite.False(*dbAccount.Discoverable)
	suite.WithinDuration(time.Now(), dbAccount.SuspendedAt, 30*time.Second)
	suite.Equal(dbAccount.ID, dbAccount.SuspensionOrigin)
}

func TestInboxPostTestSuite(t *testing.T) {
	suite.Run(t, &InboxPostTestSuite{})
}
