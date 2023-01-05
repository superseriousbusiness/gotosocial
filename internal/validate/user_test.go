/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package validate_test

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func happyUser() *gtsmodel.User {
	return &gtsmodel.User{
		ID:                     "01FE8TTK9F34BR0KG7639AJQTX",
		Email:                  "whatever@example.org",
		AccountID:              "01FE8TWA7CN8J7237K5DFS1RY5",
		Account:                nil,
		EncryptedPassword:      "$2y$10$tkRapNGW.RWkEuCMWdgArunABFvsPGRvFQY3OibfSJo0RDL3z8WfC",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		SignUpIP:               net.ParseIP("128.64.32.16"),
		CurrentSignInAt:        time.Now(),
		CurrentSignInIP:        net.ParseIP("128.64.32.16"),
		LastSignInAt:           time.Now(),
		LastSignInIP:           net.ParseIP("128.64.32.16"),
		SignInCount:            0,
		InviteID:               "",
		ChosenLanguages:        []string{},
		FilteredLanguages:      []string{},
		Locale:                 "en",
		CreatedByApplicationID: "01FE8Y5EHMWCA1MHMTNHRVZ1X4",
		CreatedByApplication:   nil,
		LastEmailedAt:          time.Now(),
		ConfirmationToken:      "",
		ConfirmedAt:            time.Now(),
		ConfirmationSentAt:     time.Time{},
		UnconfirmedEmail:       "",
		Moderator:              testrig.FalseBool(),
		Admin:                  testrig.FalseBool(),
		Disabled:               testrig.FalseBool(),
		Approved:               testrig.TrueBool(),
	}
}

type UserValidateTestSuite struct {
	suite.Suite
}

func (suite *UserValidateTestSuite) TestValidateUserHappyPath() {
	// no problem here
	u := happyUser()
	err := validate.Struct(u)
	suite.NoError(err)
}

func (suite *UserValidateTestSuite) TestValidateUserNoID() {
	// user has no id set
	u := happyUser()
	u.ID = ""

	err := validate.Struct(u)
	suite.EqualError(err, "Key: 'User.ID' Error:Field validation for 'ID' failed on the 'required' tag")
}

func (suite *UserValidateTestSuite) TestValidateUserNoEmail() {
	// user has no email or unconfirmed email set
	u := happyUser()
	u.Email = ""

	err := validate.Struct(u)
	suite.EqualError(err, "Key: 'User.Email' Error:Field validation for 'Email' failed on the 'required_with' tag\nKey: 'User.UnconfirmedEmail' Error:Field validation for 'UnconfirmedEmail' failed on the 'required_without' tag")
}

func (suite *UserValidateTestSuite) TestValidateUserOnlyUnconfirmedEmail() {
	// user has only UnconfirmedEmail but ConfirmedAt is set
	u := happyUser()
	u.Email = ""
	u.UnconfirmedEmail = "whatever@example.org"

	err := validate.Struct(u)
	suite.EqualError(err, "Key: 'User.Email' Error:Field validation for 'Email' failed on the 'required_with' tag")
}

func (suite *UserValidateTestSuite) TestValidateUserOnlyUnconfirmedEmailOK() {
	// user has only UnconfirmedEmail and ConfirmedAt is not set
	u := happyUser()
	u.Email = ""
	u.UnconfirmedEmail = "whatever@example.org"
	u.ConfirmedAt = time.Time{}

	err := validate.Struct(u)
	suite.NoError(err)
}

func (suite *UserValidateTestSuite) TestValidateUserNoConfirmedAt() {
	// user has Email but no ConfirmedAt
	u := happyUser()
	u.ConfirmedAt = time.Time{}

	err := validate.Struct(u)
	suite.EqualError(err, "Key: 'User.ConfirmedAt' Error:Field validation for 'ConfirmedAt' failed on the 'required_with' tag")
}

func (suite *UserValidateTestSuite) TestValidateUserUnlikelySignInCount() {
	// user has Email but no ConfirmedAt
	u := happyUser()
	u.SignInCount = -69

	err := validate.Struct(u)
	suite.EqualError(err, "Key: 'User.SignInCount' Error:Field validation for 'SignInCount' failed on the 'min' tag")
}

func TestUserValidateTestSuite(t *testing.T) {
	suite.Run(t, new(UserValidateTestSuite))
}
