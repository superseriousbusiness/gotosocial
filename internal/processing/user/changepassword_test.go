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

package user_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"golang.org/x/crypto/bcrypt"
)

type ChangePasswordTestSuite struct {
	UserStandardTestSuite
}

func (suite *ChangePasswordTestSuite) TestChangePasswordOK() {
	user := suite.testUsers["local_account_1"]

	errWithCode := suite.user.ChangePassword(context.Background(), user, "password", "verygoodnewpassword")
	suite.NoError(errWithCode)

	err := bcrypt.CompareHashAndPassword([]byte(user.EncryptedPassword), []byte("verygoodnewpassword"))
	suite.NoError(err)

	// get user from the db again
	dbUser := &gtsmodel.User{}
	err = suite.db.GetByID(context.Background(), user.ID, dbUser)
	suite.NoError(err)

	// check the password has changed
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.EncryptedPassword), []byte("verygoodnewpassword"))
	suite.NoError(err)
}

func (suite *ChangePasswordTestSuite) TestChangePasswordIncorrectOld() {
	user := suite.testUsers["local_account_1"]

	errWithCode := suite.user.ChangePassword(context.Background(), user, "ooooopsydoooopsy", "verygoodnewpassword")
	suite.EqualError(errWithCode, "crypto/bcrypt: hashedPassword is not the hash of the given password")
	suite.Equal(http.StatusUnauthorized, errWithCode.Code())
	suite.Equal("Unauthorized: old password was incorrect", errWithCode.Safe())

	// get user from the db again
	dbUser := &gtsmodel.User{}
	err := suite.db.GetByID(context.Background(), user.ID, dbUser)
	suite.NoError(err)

	// check the password has not changed
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.EncryptedPassword), []byte("password"))
	suite.NoError(err)
}

func (suite *ChangePasswordTestSuite) TestChangePasswordWeakNew() {
	user := suite.testUsers["local_account_1"]

	errWithCode := suite.user.ChangePassword(context.Background(), user, "password", "1234")
	suite.EqualError(errWithCode, "password is only 11% strength, try including more special characters, using lowercase letters, using uppercase letters or using a longer password")
	suite.Equal(http.StatusBadRequest, errWithCode.Code())
	suite.Equal("Bad Request: password is only 11% strength, try including more special characters, using lowercase letters, using uppercase letters or using a longer password", errWithCode.Safe())

	// get user from the db again
	dbUser := &gtsmodel.User{}
	err := suite.db.GetByID(context.Background(), user.ID, dbUser)
	suite.NoError(err)

	// check the password has not changed
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.EncryptedPassword), []byte("password"))
	suite.NoError(err)
}

func TestChangePasswordTestSuite(t *testing.T) {
	suite.Run(t, &ChangePasswordTestSuite{})
}
