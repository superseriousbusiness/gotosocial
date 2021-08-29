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

package gtsmodel_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type ValidateTestSuite struct {
	suite.Suite
}

func (suite *ValidateTestSuite) TestValidatePointer() {
	var nilUser *gtsmodel.User
	suite.PanicsWithValue(gtsmodel.PointerValidationPanic, func() {
		gtsmodel.ValidateStruct(nilUser)
	})
}

func (suite *ValidateTestSuite) TestValidateNil() {
	suite.PanicsWithValue(gtsmodel.InvalidValidationPanic, func() {
		gtsmodel.ValidateStruct(nil)
	})
}

func (suite *ValidateTestSuite) TestValidateWeirdULID() {
	type a struct {
		ID bool `validate:"required,ulid"`
	}

	err := gtsmodel.ValidateStruct(a{ID: true})
	suite.Error(err)
}

func (suite *ValidateTestSuite) TestValidateNotStruct() {
	type aaaaaaa string
	aaaaaa := aaaaaaa("aaaa")
	suite.Panics(func() {
		gtsmodel.ValidateStruct(aaaaaa)
	})
}

func TestValidateTestSuite(t *testing.T) {
	suite.Run(t, new(ValidateTestSuite))
}
