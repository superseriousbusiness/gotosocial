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

package federation_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-fed/activity/pub"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type ProtocolTestSuite struct {
	suite.Suite
	config     *config.Config
	db         db.DB
	log        *logrus.Logger
	federator  *federation.Federator
	tc         transport.Controller
	activities map[string]pub.Activity
}

// SetupSuite sets some variables on the suite that we can use as consts (more or less) throughout
func (suite *ProtocolTestSuite) SetupSuite() {
	// setup standard items
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.tc = testrig.NewTestTransportController(suite.db, testrig.NewMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, nil
	}))
	suite.activities = testrig.NewTestActivities()

	// setup module being tested
	suite.federator = federation.NewFederator(suite.db, suite.log, suite.config, suite.tc).(*federation.Federator)
}

func (suite *ProtocolTestSuite) SetupTest() {
	testrig.StandardDBSetup(suite.db)
}

// TearDownTest drops tables to make sure there's no data in the db
func (suite *ProtocolTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

// make sure PostInboxRequestBodyHook properly sets the inbox username and activity on the context
func (suite *ProtocolTestSuite) TestPostInboxRequestBodyHook() {

	activity := suite.activities["dm_for_zork"]

	// setup
	ctx := context.Background()
	request := httptest.NewRequest(http.MethodPost, "http://localhost:8080/users/the_mighty_zork/inbox", nil) // the endpoint we're hitting

	newContext, err := suite.federator.PostInboxRequestBodyHook(ctx, request, activity)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), newContext)

	usernameI := newContext.Value(util.APUsernameKey)
	assert.NotNil(suite.T(), usernameI)
	username, ok := usernameI.(string)
	assert.True(suite.T(), ok)
	assert.NotEmpty(suite.T(), username)
	assert.Equal(suite.T(), "the_mighty_zork", username)

	activityI := newContext.Value(util.APActivityKey)
	assert.NotNil(suite.T(), activityI)
	returnedActivity, ok := activityI.(pub.Activity)
	assert.True(suite.T(), ok)
	assert.NotNil(suite.T(), returnedActivity)
	assert.EqualValues(suite.T(), activity, returnedActivity)

	r, err := returnedActivity.Serialize()
	assert.NoError(suite.T(), err)

	b, err := json.Marshal(r)
	assert.NoError(suite.T(), err)

	fmt.Println(string(b))

}

func TestProtocolTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolTestSuite))
}
