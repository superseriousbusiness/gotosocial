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

package user_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/user"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type EmailChangeTestSuite struct {
	UserStandardTestSuite
}

func (suite *EmailChangeTestSuite) TestEmailChangePOST() {
	// Get a new processor for this test, as
	// we're expecting an email, and we don't
	// want the other tests interfering if
	// we're running them at the same time.
	state := new(state.State)
	state.DB = testrig.NewTestDB(&suite.state)
	storage := testrig.NewInMemoryStorage()
	sentEmails := make(map[string]string)
	emailSender := testrig.NewEmailSender("../../../../web/template/", sentEmails)
	webPushSender := testrig.NewNoopWebPushSender()
	processor := testrig.NewTestProcessor(state, suite.federator, emailSender, webPushSender, suite.mediaManager)
	testrig.StartWorkers(state, processor.Workers())
	userModule := user.New(processor)
	testrig.StandardDBSetup(state.DB, suite.testAccounts)
	testrig.StandardStorageSetup(storage, "../../../../testrig/media")

	defer func() {
		testrig.StandardDBTeardown(state.DB)
		testrig.StandardStorageTeardown(storage)
		testrig.StopWorkers(state)
	}()

	response, code := suite.POST(user.EmailChangePath, map[string][]string{
		"password":  {"password"},
		"new_email": {"someone@example.org"},
	}, userModule.EmailChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusAccepted, code)
	b, err := io.ReadAll(response.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	apiUser := new(apimodel.User)
	if err := json.Unmarshal(b, apiUser); err != nil {
		suite.FailNow(err.Error())
	}

	// Unconfirmed email should be set now.
	suite.Equal("someone@example.org", apiUser.UnconfirmedEmail)

	// Ensure unconfirmed address gets an email.
	if !testrig.WaitFor(func() bool {
		_, ok := sentEmails["someone@example.org"]
		return ok
	}) {
		suite.FailNow("no email received")
	}
}

func (suite *EmailChangeTestSuite) TestEmailChangePOSTAddressInUse() {
	response, code := suite.POST(user.EmailChangePath, map[string][]string{
		"password":  {"password"},
		"new_email": {"admin@example.org"},
	}, suite.userModule.EmailChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusConflict, code)
	b, err := io.ReadAll(response.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{"error":"Conflict: new email address is already in use on this instance"}`, string(b))
}

func (suite *EmailChangeTestSuite) TestEmailChangePOSTSameEmail() {
	response, code := suite.POST(user.EmailChangePath, map[string][]string{
		"password":  {"password"},
		"new_email": {"zork@example.org"},
	}, suite.userModule.EmailChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusBadRequest, code)
	b, err := io.ReadAll(response.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{"error":"Bad Request: new email address cannot be the same as current email address"}`, string(b))
}

func (suite *EmailChangeTestSuite) TestEmailChangePOSTBadPassword() {
	response, code := suite.POST(user.EmailChangePath, map[string][]string{
		"password":  {"notmypassword"},
		"new_email": {"someone@example.org"},
	}, suite.userModule.EmailChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusUnauthorized, code)
	b, err := io.ReadAll(response.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(`{"error":"Unauthorized: password was incorrect"}`, string(b))
}

func TestEmailChangeTestSuite(t *testing.T) {
	suite.Run(t, &EmailChangeTestSuite{})
}
