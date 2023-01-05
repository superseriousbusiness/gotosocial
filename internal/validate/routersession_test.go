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

func happyRouterSession() *gtsmodel.RouterSession {
	return &gtsmodel.RouterSession{
		ID:    "01FE91RJR88PSEEE30EV35QR8N",
		Auth:  []byte("12345678901234567890123456789012"),
		Crypt: []byte("12345678901234567890123456789012"),
	}
}

type RouterSessionValidateTestSuite struct {
	suite.Suite
}

func (suite *RouterSessionValidateTestSuite) TestValidateRouterSessionHappyPath() {
	// no problem here
	r := happyRouterSession()
	err := validate.Struct(r)
	suite.NoError(err)
}

func (suite *RouterSessionValidateTestSuite) TestValidateRouterSessionAuth() {
	r := happyRouterSession()

	// remove auth struct
	r.Auth = nil
	err := validate.Struct(r)
	suite.EqualError(err, "Key: 'RouterSession.Auth' Error:Field validation for 'Auth' failed on the 'required' tag")

	// auth bytes too long
	r.Auth = []byte("1234567890123456789012345678901234567890")
	err = validate.Struct(r)
	suite.EqualError(err, "Key: 'RouterSession.Auth' Error:Field validation for 'Auth' failed on the 'len' tag")

	// auth bytes too short
	r.Auth = []byte("12345678901")
	err = validate.Struct(r)
	suite.EqualError(err, "Key: 'RouterSession.Auth' Error:Field validation for 'Auth' failed on the 'len' tag")
}

func (suite *RouterSessionValidateTestSuite) TestValidateRouterSessionCrypt() {
	r := happyRouterSession()

	// remove crypt struct
	r.Crypt = nil
	err := validate.Struct(r)
	suite.EqualError(err, "Key: 'RouterSession.Crypt' Error:Field validation for 'Crypt' failed on the 'required' tag")

	// crypt bytes too long
	r.Crypt = []byte("1234567890123456789012345678901234567890")
	err = validate.Struct(r)
	suite.EqualError(err, "Key: 'RouterSession.Crypt' Error:Field validation for 'Crypt' failed on the 'len' tag")

	// crypt bytes too short
	r.Crypt = []byte("12345678901")
	err = validate.Struct(r)
	suite.EqualError(err, "Key: 'RouterSession.Crypt' Error:Field validation for 'Crypt' failed on the 'len' tag")
}

func TestRouterSessionValidateTestSuite(t *testing.T) {
	suite.Run(t, new(RouterSessionValidateTestSuite))
}
