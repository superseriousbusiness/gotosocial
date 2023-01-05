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
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

type ValidateTestSuite struct {
	suite.Suite
}

func (suite *ValidateTestSuite) TestValidateNilPointer() {
	var nilUser *gtsmodel.User
	suite.Panics(func() {
		validate.Struct(nilUser)
	})
}

func (suite *ValidateTestSuite) TestValidatePointer() {
	user := &gtsmodel.User{}
	err := validate.Struct(user)
	suite.EqualError(err, "Key: 'User.ID' Error:Field validation for 'ID' failed on the 'required' tag\nKey: 'User.AccountID' Error:Field validation for 'AccountID' failed on the 'required' tag\nKey: 'User.EncryptedPassword' Error:Field validation for 'EncryptedPassword' failed on the 'required' tag\nKey: 'User.UnconfirmedEmail' Error:Field validation for 'UnconfirmedEmail' failed on the 'required_without' tag")
}

func (suite *ValidateTestSuite) TestValidateNil() {
	suite.Panics(func() {
		validate.Struct(nil)
	})
}

func (suite *ValidateTestSuite) TestValidateWeirdULID() {
	type a struct {
		ID bool `validate:"required,ulid"`
	}

	err := validate.Struct(a{ID: true})
	suite.Error(err)
}

func (suite *ValidateTestSuite) TestValidateNotStruct() {
	type aaaaaaa string
	aaaaaa := aaaaaaa("aaaa")
	suite.Panics(func() {
		validate.Struct(aaaaaa)
	})
}

func TestValidateTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateTestSuite))
}
