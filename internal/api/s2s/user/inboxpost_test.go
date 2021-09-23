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

package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-fed/activity/streams"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/s2s/user"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
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
	fmt.Println(string(bodyJson))
	body := bytes.NewReader(bodyJson)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator)
	userModule := user.New(suite.config, processor, suite.log).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
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

func (suite *InboxPostTestSuite) TestPostUnblock() {
   blockingAccount := suite.testAccounts["remote_account_1"]
	blockedAccount := suite.testAccounts["local_account_1"]

	// first put a block in the database so we have something to undo
	blockURI := "http://fossbros-anonymous.io/users/foss_satan/blocks/01FG9C441MCTW3R2W117V2PQK3"
	undoURI := "http://fossbros-anonymous.io/72cc96a3-f742-4daf-b9f5-3407667260c5"
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

	err = suite.db.Put(context.Background(), dbBlock)
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
   undoID.SetIRI(testrig.URLMustParse(undoURI))
   undo.SetJSONLDId(undoID)

	targetURI := testrig.URLMustParse(blockedAccount.InboxURI)

	signature, digestHeader, dateHeader := testrig.GetSignatureForActivity(undo, blockingAccount.PublicKeyURI, blockingAccount.PrivateKey, targetURI)
	bodyI, err := streams.Serialize(undo)
	suite.NoError(err)

	bodyJson, err := json.Marshal(bodyI)
	suite.NoError(err)
	fmt.Println(string(bodyJson))
	body := bytes.NewReader(bodyJson)

	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil), suite.db)
	federator := testrig.NewTestFederator(suite.db, tc, suite.storage)
	processor := testrig.NewTestProcessor(suite.db, suite.storage, federator)
	userModule := user.New(suite.config, processor, suite.log).(*user.Module)

	// setup request
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
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


}

func TestInboxPostTestSuite(t *testing.T) {
	suite.Run(t, &InboxPostTestSuite{})
}
