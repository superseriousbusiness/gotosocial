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
	"net/url"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/filter/interaction"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type FederatingActorTestSuite struct {
	FederatorStandardTestSuite
}

func (suite *FederatingActorTestSuite) TestSendNoRemoteFollowers() {
	ctx := context.Background()
	testAccount := suite.testAccounts["local_account_1"]
	testNote := testrig.NewAPNote(
		testrig.URLMustParse("http://localhost:8080/users/the_mighty_zork/statuses/01G1TR6BADACCZWQMNF9X21TV5"),
		testrig.URLMustParse("http://localhost:8080/@the_mighty_zork/statuses/01G1TR6BADACCZWQMNF9X21TV5"),
		time.Now(),
		"boobies",
		"",
		testrig.URLMustParse(testAccount.URI),
		[]*url.URL{testrig.URLMustParse(testAccount.FollowersURI)},
		nil,
		false,
		nil,
		nil,
		nil,
	)
	testActivity := testrig.WrapAPNoteInCreate(testrig.URLMustParse("http://localhost:8080/whatever_some_create"), testrig.URLMustParse(testAccount.URI), time.Now(), testNote)

	// setup transport controller with a no-op client so we don't make external calls
	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(&suite.state, httpClient)

	// setup module being tested
	federator := federation.NewFederator(
		&suite.state,
		testrig.NewTestFederatingDB(&suite.state),
		tc,
		suite.typeconverter,
		visibility.NewFilter(&suite.state),
		interaction.NewFilter(&suite.state),
		testrig.NewTestMediaManager(&suite.state),
	)

	activity, err := federator.FederatingActor().Send(ctx, testrig.URLMustParse(testAccount.OutboxURI), testActivity)
	suite.NoError(err)
	suite.NotNil(activity)

	// because zork has no remote followers, sent messages should be empty (no messages sent to own instance)
	suite.Empty(&httpClient.SentMessages)
}

func (suite *FederatingActorTestSuite) TestSendRemoteFollower() {
	ctx := context.Background()
	testAccount := suite.testAccounts["local_account_1"]
	testRemoteAccount := suite.testAccounts["remote_account_1"]

	err := suite.state.DB.Put(ctx, &gtsmodel.Follow{
		ID:              "01G1TRWV4AYCDBX5HRWT2EVBCV",
		CreatedAt:       testrig.TimeMustParse("2022-06-02T12:22:21+02:00"),
		UpdatedAt:       testrig.TimeMustParse("2022-06-02T12:22:21+02:00"),
		AccountID:       testRemoteAccount.ID,
		TargetAccountID: testAccount.ID,
		ShowReblogs:     util.Ptr(true),
		URI:             "http://fossbros-anonymous.io/users/foss_satan/follows/01G1TRWV4AYCDBX5HRWT2EVBCV",
		Notify:          util.Ptr(false),
	})
	suite.NoError(err)

	testNote := testrig.NewAPNote(
		testrig.URLMustParse("http://localhost:8080/users/the_mighty_zork/statuses/01G1TR6BADACCZWQMNF9X21TV5"),
		testrig.URLMustParse("http://localhost:8080/@the_mighty_zork/statuses/01G1TR6BADACCZWQMNF9X21TV5"),
		testrig.TimeMustParse("2022-06-02T12:22:21+02:00"),
		"boobies",
		"",
		testrig.URLMustParse(testAccount.URI),
		[]*url.URL{testrig.URLMustParse(testAccount.FollowersURI)},
		nil,
		false,
		nil,
		nil,
		nil,
	)
	testActivity := testrig.WrapAPNoteInCreate(testrig.URLMustParse("http://localhost:8080/whatever_some_create"), testrig.URLMustParse(testAccount.URI), testrig.TimeMustParse("2022-06-02T12:22:21+02:00"), testNote)

	httpClient := testrig.NewMockHTTPClient(nil, "../../testrig/media")
	tc := testrig.NewTestTransportController(&suite.state, httpClient)

	// setup module being tested
	federator := federation.NewFederator(
		&suite.state,
		testrig.NewTestFederatingDB(&suite.state),
		tc,
		suite.typeconverter,
		visibility.NewFilter(&suite.state),
		interaction.NewFilter(&suite.state),
		testrig.NewTestMediaManager(&suite.state),
	)

	activity, err := federator.FederatingActor().Send(ctx, testrig.URLMustParse(testAccount.OutboxURI), testActivity)
	suite.NoError(err)
	suite.NotNil(activity)

	// because we added 1 remote follower for zork, there should be a url in sentMessage
	var sent []byte
	if !testrig.WaitFor(func() bool {
		delivery, ok := suite.state.Workers.Delivery.Queue.Pop()
		if !ok {
			return false
		}
		if !testrig.EqualRequestURIs(delivery.Request.URL, *testRemoteAccount.SharedInboxURI) {
			panic("differing request uris")
		}
		sent, err = io.ReadAll(delivery.Request.Body)
		if err != nil {
			panic("error reading body: " + err.Error())
		}
		return true

	}) {
		suite.FailNow("timed out waiting for message")
	}

	dst := new(bytes.Buffer)
	err = json.Indent(dst, sent, "", "  ")
	suite.NoError(err)
	suite.Equal(`{
  "@context": "https://www.w3.org/ns/activitystreams",
  "actor": "http://localhost:8080/users/the_mighty_zork",
  "id": "http://localhost:8080/whatever_some_create",
  "object": {
    "attributedTo": "http://localhost:8080/users/the_mighty_zork",
    "content": "boobies",
    "id": "http://localhost:8080/users/the_mighty_zork/statuses/01G1TR6BADACCZWQMNF9X21TV5",
    "published": "2022-06-02T12:22:21+02:00",
    "tag": [],
    "to": "http://localhost:8080/users/the_mighty_zork/followers",
    "type": "Note",
    "url": "http://localhost:8080/@the_mighty_zork/statuses/01G1TR6BADACCZWQMNF9X21TV5"
  },
  "published": "2022-06-02T12:22:21+02:00",
  "to": "http://localhost:8080/users/the_mighty_zork/followers",
  "type": "Create"
}`, dst.String())
}

func TestFederatingActorTestSuite(t *testing.T) {
	suite.Run(t, new(FederatingActorTestSuite))
}
