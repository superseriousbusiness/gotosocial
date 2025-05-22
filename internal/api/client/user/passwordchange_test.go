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
	"io"
	"net/http"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/api/client/user"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"codeberg.org/gruf/go-byteutil"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type PasswordChangeTestSuite struct {
	UserStandardTestSuite
}

func (suite *PasswordChangeTestSuite) TestPasswordChangePOST() {
	response, code := suite.POST(user.PasswordChangePath, map[string][]string{
		"old_password": {"password"},
		"new_password": {"peepeepoopoopassword"},
	}, suite.userModule.PasswordChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusOK, code)

	dbUser := &gtsmodel.User{}
	err := suite.db.GetByID(suite.T().Context(), suite.testUsers["local_account_1"].ID, dbUser)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// new password should pass
	err = bcrypt.CompareHashAndPassword(
		byteutil.S2B(dbUser.EncryptedPassword),
		byteutil.S2B("peepeepoopoopassword"),
	)
	suite.NoError(err)

	// old password should fail
	err = bcrypt.CompareHashAndPassword(
		byteutil.S2B(dbUser.EncryptedPassword),
		byteutil.S2B("password"),
	)
	suite.EqualError(err, "crypto/bcrypt: hashedPassword is not the hash of the given password")
}

func (suite *PasswordChangeTestSuite) TestPasswordMissingOldPassword() {
	response, code := suite.POST(user.PasswordChangePath, map[string][]string{
		"new_password": {"peepeepoopoopassword"},
	}, suite.userModule.PasswordChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusBadRequest, code)
	b, err := io.ReadAll(response.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(`{"error":"Bad Request: password change request missing field old_password"}`, string(b))
}

func (suite *PasswordChangeTestSuite) TestPasswordIncorrectOldPassword() {
	response, code := suite.POST(user.PasswordChangePath, map[string][]string{
		"old_password": {"notright"},
		"new_password": {"peepeepoopoopassword"},
	}, suite.userModule.PasswordChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusUnauthorized, code)
	b, err := io.ReadAll(response.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(`{"error":"Unauthorized: old password was incorrect"}`, string(b))
}

func (suite *PasswordChangeTestSuite) TestPasswordWeakNewPassword() {
	response, code := suite.POST(user.PasswordChangePath, map[string][]string{
		"old_password": {"password"},
		"new_password": {"peepeepoopoo"},
	}, suite.userModule.PasswordChangePOSTHandler)
	defer response.Body.Close()

	// Check response
	suite.EqualValues(http.StatusBadRequest, code)
	b, err := io.ReadAll(response.Body)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Equal(`{"error":"Bad Request: password is only 94% strength, try including more special characters, using uppercase letters, using numbers or using a longer password"}`, string(b))
}

func TestPasswordChangeTestSuite(t *testing.T) {
	suite.Run(t, &PasswordChangeTestSuite{})
}
