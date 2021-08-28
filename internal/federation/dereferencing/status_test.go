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

package dereferencing_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/go-fed/activity/streams"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type StatusTestSuite struct {
	DereferencerStandardTestSuite
}

// mockTransportController returns basically a miniature muxer, which returns a different
// value based on the request URL. It can be used to return remote statuses, profiles, etc,
// as though they were actually being dereferenced. If the URL doesn't correspond to any person
// or note or attachment that we have stored, then just a 200 code will be returned, with an empty body.
func (suite *StatusTestSuite) mockTransportController() transport.Controller {
	do := func(req *http.Request) (*http.Response, error) {
		suite.log.Debugf("received request for %s", req.URL)

		responseBytes := []byte{}

		note, ok := suite.testRemoteStatuses[req.URL.String()]
		if ok {
			// the request is for a note that we have stored
			noteI, err := streams.Serialize(note)
			if err != nil {
				panic(err)
			}
			noteJson, err := json.Marshal(noteI)
			if err != nil {
				panic(err)
			}
			responseBytes = noteJson
		}

		person, ok := suite.testRemoteAccounts[req.URL.String()]
		if ok {
			// the request is for a person that we have stored
			personI, err := streams.Serialize(person)
			if err != nil {
				panic(err)
			}
			personJson, err := json.Marshal(personI)
			if err != nil {
				panic(err)
			}
			responseBytes = personJson
		}

		if len(responseBytes) != 0 {
			// we found something, so print what we're going to return
			suite.log.Debugf("returning response %s", string(responseBytes))
		}

		reader := bytes.NewReader(responseBytes)
		readCloser := io.NopCloser(reader)
		response := &http.Response{
			StatusCode: 200,
			Body:       readCloser,
		}

		return response, nil
	}
	mockClient := testrig.NewMockHTTPClient(do)
	return testrig.NewTestTransportController(mockClient, suite.db)
}

func (suite *StatusTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testRemoteStatuses = testrig.NewTestFediStatuses()
	suite.testRemoteAccounts = testrig.NewTestFediPeople()
}

func (suite *StatusTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.dereferencer = dereferencing.NewDereferencer(suite.config,
		suite.db,
		testrig.NewTestTypeConverter(suite.db),
		suite.mockTransportController(),
		testrig.NewTestMediaHandler(suite.db, testrig.NewTestStorage()),
		suite.log)
	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *StatusTestSuite) TestDereferenceSimpleStatus() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839")
	status, statusable, new, err := suite.dereferencer.GetRemoteStatus(context.Background(), fetchingAccount.Username, statusURL, false)
	suite.NoError(err)
	suite.NotNil(status)
	suite.NotNil(statusable)
	suite.True(new)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01FE4NTHKWW7THT67EF10EB839", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01FE4NTHKWW7THT67EF10EB839", status.URL)
	suite.Equal("Hello world!", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(gtsmodel.ActivityStreamsNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)
}

func (suite *StatusTestSuite) TestDereferenceStatusWithMention() {
	fetchingAccount := suite.testAccounts["local_account_1"]

	statusURL := testrig.URLMustParse("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV")
	status, statusable, new, err := suite.dereferencer.GetRemoteStatus(context.Background(), fetchingAccount.Username, statusURL, false)
	suite.NoError(err)
	suite.NotNil(status)
	suite.NotNil(statusable)
	suite.True(new)

	// status values should be set
	suite.Equal("https://unknown-instance.com/users/brand_new_person/statuses/01FE5Y30E3W4P7TRE0R98KAYQV", status.URI)
	suite.Equal("https://unknown-instance.com/users/@brand_new_person/01FE5Y30E3W4P7TRE0R98KAYQV", status.URL)
	suite.Equal("Hey @the_mighty_zork@localhost:8080 how's it going?", status.Content)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", status.AccountURI)
	suite.False(status.Local)
	suite.Empty(status.ContentWarning)
	suite.Equal(gtsmodel.VisibilityPublic, status.Visibility)
	suite.Equal(gtsmodel.ActivityStreamsNote, status.ActivityStreamsType)

	// status should be in the database
	dbStatus, err := suite.db.GetStatusByURI(context.Background(), status.URI)
	suite.NoError(err)
	suite.Equal(status.ID, dbStatus.ID)

	// account should be in the database now too
	account, err := suite.db.GetAccountByURI(context.Background(), status.AccountURI)
	suite.NoError(err)
	suite.NotNil(account)
	suite.True(account.Discoverable)
	suite.Equal("https://unknown-instance.com/users/brand_new_person", account.URI)
	suite.Equal("hey I'm a new person, your instance hasn't seen me yet uwu", account.Note)
	suite.Equal("Geoff Brando New Personson", account.DisplayName)
	suite.Equal("brand_new_person", account.Username)
	suite.NotNil(account.PublicKey)
	suite.Nil(account.PrivateKey)

	// we should have a mention in the database
	m := &gtsmodel.Mention{}
	err = suite.db.GetWhere(context.Background(), []db.Where{{Key: "status_id", Value: status.ID}}, m)
	suite.NoError(err)
	suite.NotNil(m)
	suite.Equal(status.ID, m.StatusID)
	suite.Equal(account.ID, m.OriginAccountID)
	suite.Equal(fetchingAccount.ID, m.TargetAccountID)
	suite.Equal(account.URI, m.OriginAccountURI)
	suite.WithinDuration(time.Now(), m.CreatedAt, 5*time.Minute)
	suite.WithinDuration(time.Now(), m.UpdatedAt, 5*time.Minute)
	suite.False(m.Silent)
}

func (suite *StatusTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
