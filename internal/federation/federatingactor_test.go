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
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/testrig"
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
	)
	testActivity := testrig.WrapAPNoteInCreate(testrig.URLMustParse("http://localhost:8080/whatever_some_create"), testrig.URLMustParse(testAccount.URI), time.Now(), testNote)

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	// setup transport controller with a no-op client so we don't make external calls
	sentMessages := []*url.URL{}
	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		sentMessages = append(sentMessages, req.URL)
		r := ioutil.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}), suite.db, fedWorker)
	// setup module being tested
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	activity, err := federator.FederatingActor().Send(ctx, testrig.URLMustParse(testAccount.OutboxURI), testActivity)
	suite.NoError(err)
	suite.NotNil(activity)

	// because zork has no remote followers, sent messages should be empty (no messages sent to own instance)
	suite.Empty(sentMessages)
}

func (suite *FederatingActorTestSuite) TestSendRemoteFollower() {
	ctx := context.Background()
	testAccount := suite.testAccounts["local_account_1"]
	testRemoteAccount := suite.testAccounts["remote_account_1"]

	err := suite.db.Put(ctx, &gtsmodel.Follow{
		ID:              "01G1TRWV4AYCDBX5HRWT2EVBCV",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		AccountID:       testRemoteAccount.ID,
		TargetAccountID: testAccount.ID,
		ShowReblogs:     true,
		URI:             "http://fossbros-anonymous.io/users/foss_satan/follows/01G1TRWV4AYCDBX5HRWT2EVBCV",
		Notify:          false,
	})
	suite.NoError(err)

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
	)
	testActivity := testrig.WrapAPNoteInCreate(testrig.URLMustParse("http://localhost:8080/whatever_some_create"), testrig.URLMustParse(testAccount.URI), time.Now(), testNote)

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)

	// setup transport controller with a no-op client so we don't make external calls
	sentMessages := []*url.URL{}
	tc := testrig.NewTestTransportController(testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		sentMessages = append(sentMessages, req.URL)
		r := ioutil.NopCloser(bytes.NewReader([]byte{}))
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}), suite.db, fedWorker)
	// setup module being tested
	federator := federation.NewFederator(suite.db, testrig.NewTestFederatingDB(suite.db, fedWorker), tc, suite.tc, testrig.NewTestMediaManager(suite.db, suite.storage))

	activity, err := federator.FederatingActor().Send(ctx, testrig.URLMustParse(testAccount.OutboxURI), testActivity)
	suite.NoError(err)
	suite.NotNil(activity)

	// because we added 1 remote follower for zork, there should be a url in sentMessage
	suite.Len(sentMessages, 1)
	suite.Equal(testRemoteAccount.InboxURI, sentMessages[0].String())
}

func TestFederatingActorTestSuite(t *testing.T) {
	suite.Run(t, new(FederatingActorTestSuite))
}
