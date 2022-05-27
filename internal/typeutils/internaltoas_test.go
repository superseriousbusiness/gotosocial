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

package typeutils_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
)

type InternalToASTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *InternalToASTestSuite) TestAccountToAS() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	suite.NoError(err)

	ser, err := streams.Serialize(asPerson)
	suite.NoError(err)

	bytes, err := json.Marshal(ser)
	suite.NoError(err)

	// trim off everything up to 'discoverable';
	// this is necessary because the order of multiple 'context' entries is not determinate
	trimmed := strings.Split(string(bytes), "\"discoverable\"")[1]

	suite.Equal(`:true,"featured":"http://localhost:8080/users/the_mighty_zork/collections/featured","followers":"http://localhost:8080/users/the_mighty_zork/followers","following":"http://localhost:8080/users/the_mighty_zork/following","icon":{"mediaType":"image/jpeg","type":"Image","url":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/avatar/original/01F8MH58A357CV5K7R7TJMSH6S.jpeg"},"id":"http://localhost:8080/users/the_mighty_zork","image":{"mediaType":"image/jpeg","type":"Image","url":"http://localhost:8080/fileserver/01F8MH1H7YV1Z7D2C8K2730QBF/header/original/01PFPMWK2FF0D9WMHEJHR07C3Q.jpeg"},"inbox":"http://localhost:8080/users/the_mighty_zork/inbox","manuallyApprovesFollowers":false,"name":"original zork (he/they)","outbox":"http://localhost:8080/users/the_mighty_zork/outbox","preferredUsername":"the_mighty_zork","publicKey":{"id":"http://localhost:8080/users/the_mighty_zork/main-key","owner":"http://localhost:8080/users/the_mighty_zork","publicKeyPem":"-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwXTcOAvM1Jiw5Ffpk0qn\nr0cwbNvFe/5zQ+Tp7tumK/ZnT37o7X0FUEXrxNi+dkhmeJ0gsaiN+JQGNUewvpSk\nPIAXKvi908aSfCGjs7bGlJCJCuDuL5d6m7hZnP9rt9fJc70GElPpG0jc9fXwlz7T\nlsPb2ecatmG05Y4jPwdC+oN4MNCv9yQzEvCVMzl76EJaM602kIHC1CISn0rDFmYd\n9rSN7XPlNJw1F6PbpJ/BWQ+pXHKw3OEwNTETAUNYiVGnZU+B7a7bZC9f6/aPbJuV\nt8Qmg+UnDvW1Y8gmfHnxaWG2f5TDBvCHmcYtucIZPLQD4trAozC4ryqlmCWQNKbt\n0wIDAQAB\n-----END PUBLIC KEY-----\n"},"summary":"\u003cp\u003ehey yo this is my profile!\u003c/p\u003e","type":"Person","url":"http://localhost:8080/@the_mighty_zork"}`, trimmed)
}

func (suite *InternalToASTestSuite) TestOutboxToASCollection() {
	testAccount := suite.testAccounts["admin_account"]
	ctx := context.Background()

	collection, err := suite.typeconverter.OutboxToASCollection(ctx, testAccount.OutboxURI)
	suite.NoError(err)

	ser, err := streams.Serialize(collection)
	assert.NoError(suite.T(), err)

	bytes, err := json.Marshal(ser)
	suite.NoError(err)

	/*
		we want this:
		{
			"@context": "https://www.w3.org/ns/activitystreams",
			"first": "http://localhost:8080/users/admin/outbox?page=true",
			"id": "http://localhost:8080/users/admin/outbox",
			"type": "OrderedCollection"
		}
	*/

	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","first":"http://localhost:8080/users/admin/outbox?page=true","id":"http://localhost:8080/users/admin/outbox","type":"OrderedCollection"}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToAS() {
	testStatus := suite.testStatuses["local_account_1_status_1"]
	ctx := context.Background()

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := streams.Serialize(asStatus)
	assert.NoError(suite.T(), err)

	bytes, err := json.Marshal(ser)
	suite.NoError(err)

	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","attachment":[],"attributedTo":"http://localhost:8080/users/the_mighty_zork","cc":"http://localhost:8080/users/the_mighty_zork/followers","content":"hello everyone!","id":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY","published":"2021-10-20T12:40:37+02:00","replies":{"first":{"id":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?page=true","next":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies?only_other_accounts=false\u0026page=true","partOf":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies","type":"CollectionPage"},"id":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY/replies","type":"Collection"},"sensitive":true,"summary":"introduction post","tag":[],"to":"https://www.w3.org/ns/activitystreams#Public","type":"Note","url":"http://localhost:8080/@the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY"}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASWithMentions() {
	testStatusID := suite.testStatuses["admin_account_status_3"].ID
	ctx := context.Background()

	testStatus, err := suite.db.GetStatusByID(ctx, testStatusID)
	suite.NoError(err)

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := streams.Serialize(asStatus)
	assert.NoError(suite.T(), err)

	bytes, err := json.Marshal(ser)
	suite.NoError(err)

	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","attachment":[],"attributedTo":"http://localhost:8080/users/admin","cc":["http://localhost:8080/users/admin/followers","http://localhost:8080/users/the_mighty_zork"],"content":"hi @the_mighty_zork welcome to the instance!","id":"http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0","inReplyTo":"http://localhost:8080/users/the_mighty_zork/statuses/01F8MHAMCHF6Y650WCRSCP4WMY","published":"2021-11-20T13:32:16Z","replies":{"first":{"id":"http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies?page=true","next":"http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies?only_other_accounts=false\u0026page=true","partOf":"http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies","type":"CollectionPage"},"id":"http://localhost:8080/users/admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0/replies","type":"Collection"},"sensitive":false,"summary":"","tag":{"href":"http://localhost:8080/users/the_mighty_zork","name":"@the_mighty_zork@localhost:8080","type":"Mention"},"to":"https://www.w3.org/ns/activitystreams#Public","type":"Note","url":"http://localhost:8080/@admin/statuses/01FF25D5Q0DH7CHD57CTRS6WK0"}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusToASNotSensitive() {
	testStatus := suite.testStatuses["admin_account_status_1"]

	ctx := context.Background()

	asStatus, err := suite.typeconverter.StatusToAS(ctx, testStatus)
	suite.NoError(err)

	ser, err := streams.Serialize(asStatus)
	assert.NoError(suite.T(), err)

	bytes, err := json.Marshal(ser)
	suite.NoError(err)

	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","attachment":[],"attributedTo":"http://localhost:8080/users/admin","cc":"http://localhost:8080/users/admin/followers","content":"hello world! #welcome ! first post on the instance :rainbow: !","id":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R","published":"2021-10-20T11:36:45Z","replies":{"first":{"id":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?page=true","next":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies?only_other_accounts=false\u0026page=true","partOf":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies","type":"CollectionPage"},"id":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/replies","type":"Collection"},"sensitive":false,"summary":"","tag":[],"to":"https://www.w3.org/ns/activitystreams#Public","type":"Note","url":"http://localhost:8080/@admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R"}`, string(bytes))
}

func (suite *InternalToASTestSuite) TestStatusesToASOutboxPage() {
	testAccount := suite.testAccounts["admin_account"]
	ctx := context.Background()

	// get public statuses from testaccount
	statuses, err := suite.db.GetAccountStatuses(ctx, testAccount.ID, 30, true, true, "", "", false, false, true)
	suite.NoError(err)

	page, err := suite.typeconverter.StatusesToASOutboxPage(ctx, testAccount.OutboxURI, "", "", statuses)
	suite.NoError(err)

	ser, err := streams.Serialize(page)
	assert.NoError(suite.T(), err)

	bytes, err := json.Marshal(ser)
	suite.NoError(err)

	/*

		we want this:

		{
			"@context": "https://www.w3.org/ns/activitystreams",
			"id": "http://localhost:8080/users/admin/outbox?page=true",
			"next": "http://localhost:8080/users/admin/outbox?page=true&max_id=01F8MH75CBF9JFX4ZAD54N0W0R",
			"orderedItems": [
				{
					"actor": "http://localhost:8080/users/admin",
					"cc": "http://localhost:8080/users/admin/followers",
					"id": "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37/activity",
					"object": "http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37",
					"published": "2021-10-20T12:36:45Z",
					"to": "https://www.w3.org/ns/activitystreams#Public",
					"type": "Create"
				},
				{
					"actor": "http://localhost:8080/users/admin",
					"cc": "http://localhost:8080/users/admin/followers",
					"id": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/activity",
					"object": "http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R",
					"published": "2021-10-20T11:36:45Z",
					"to": "https://www.w3.org/ns/activitystreams#Public",
					"type": "Create"
				}
			],
			"partOf": "http://localhost:8080/users/admin/outbox",
			"prev": "http://localhost:8080/users/admin/outbox?page=true&min_id=01F8MHAAY43M6RJ473VQFCVH37",
			"type": "OrderedCollectionPage"
		}
	*/

	suite.Equal(`{"@context":"https://www.w3.org/ns/activitystreams","id":"http://localhost:8080/users/admin/outbox?page=true","next":"http://localhost:8080/users/admin/outbox?page=true\u0026max_id=01F8MH75CBF9JFX4ZAD54N0W0R","orderedItems":[{"actor":"http://localhost:8080/users/admin","cc":"http://localhost:8080/users/admin/followers","id":"http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37/activity","object":"http://localhost:8080/users/admin/statuses/01F8MHAAY43M6RJ473VQFCVH37","published":"2021-10-20T12:36:45Z","to":"https://www.w3.org/ns/activitystreams#Public","type":"Create"},{"actor":"http://localhost:8080/users/admin","cc":"http://localhost:8080/users/admin/followers","id":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R/activity","object":"http://localhost:8080/users/admin/statuses/01F8MH75CBF9JFX4ZAD54N0W0R","published":"2021-10-20T11:36:45Z","to":"https://www.w3.org/ns/activitystreams#Public","type":"Create"}],"partOf":"http://localhost:8080/users/admin/outbox","prev":"http://localhost:8080/users/admin/outbox?page=true\u0026min_id=01F8MHAAY43M6RJ473VQFCVH37","type":"OrderedCollectionPage"}`, string(bytes))
}

func TestInternalToASTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToASTestSuite))
}
