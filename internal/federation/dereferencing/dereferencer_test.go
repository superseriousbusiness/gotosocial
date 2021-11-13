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
	"encoding/json"
	"io"
	"net/http"

	"codeberg.org/gruf/go-store/kv"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type DereferencerStandardTestSuite struct {
	suite.Suite
	config  *config.Config
	db      db.DB
	storage *kv.KVStore

	testRemoteStatuses    map[string]vocab.ActivityStreamsNote
	testRemotePeople      map[string]vocab.ActivityStreamsPerson
	testRemoteGroups      map[string]vocab.ActivityStreamsGroup
	testRemoteAttachments map[string]testrig.RemoteAttachmentFile
	testAccounts          map[string]*gtsmodel.Account

	dereferencer dereferencing.Dereferencer
}

func (suite *DereferencerStandardTestSuite) SetupSuite() {
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testRemoteStatuses = testrig.NewTestFediStatuses()
	suite.testRemotePeople = testrig.NewTestFediPeople()
	suite.testRemoteGroups = testrig.NewTestFediGroups()
	suite.testRemoteAttachments = testrig.NewTestFediAttachments("../../../testrig/media")
}

func (suite *DereferencerStandardTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	testrig.InitTestLog()
	suite.storage = testrig.NewTestStorage()
	suite.dereferencer = dereferencing.NewDereferencer(suite.config, suite.db, testrig.NewTestTypeConverter(suite.db), suite.mockTransportController(), testrig.NewTestMediaHandler(suite.db, suite.storage))
	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *DereferencerStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

// mockTransportController returns basically a miniature muxer, which returns a different
// value based on the request URL. It can be used to return remote statuses, profiles, etc,
// as though they were actually being dereferenced. If the URL doesn't correspond to any person
// or note or attachment that we have stored, then just a 200 code will be returned, with an empty body.
func (suite *DereferencerStandardTestSuite) mockTransportController() transport.Controller {
	do := func(req *http.Request) (*http.Response, error) {
		logrus.Debugf("received request for %s", req.URL)

		responseBytes := []byte{}
		responseType := ""
		responseLength := 0

		if note, ok := suite.testRemoteStatuses[req.URL.String()]; ok {
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
			responseType = "application/activity+json"
		}

		if person, ok := suite.testRemotePeople[req.URL.String()]; ok {
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
			responseType = "application/activity+json"
		}

		if group, ok := suite.testRemoteGroups[req.URL.String()]; ok {
			// the request is for a person that we have stored
			groupI, err := streams.Serialize(group)
			if err != nil {
				panic(err)
			}
			groupJson, err := json.Marshal(groupI)
			if err != nil {
				panic(err)
			}
			responseBytes = groupJson
			responseType = "application/activity+json"
		}

		if attachment, ok := suite.testRemoteAttachments[req.URL.String()]; ok {
			responseBytes = attachment.Data
			responseType = attachment.ContentType
		}

		if len(responseBytes) != 0 {
			// we found something, so print what we're going to return
			logrus.Debugf("returning response %s", string(responseBytes))
		}
		responseLength = len(responseBytes)

		reader := bytes.NewReader(responseBytes)
		readCloser := io.NopCloser(reader)
		response := &http.Response{
			StatusCode:    200,
			Body:          readCloser,
			ContentLength: int64(responseLength),
			Header: http.Header{
				"content-type": {responseType},
			},
		}

		return response, nil
	}
	mockClient := testrig.NewMockHTTPClient(do)
	return testrig.NewTestTransportController(mockClient, suite.db)
}
