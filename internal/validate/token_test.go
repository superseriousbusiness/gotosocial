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
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

func happyToken() *gtsmodel.Token {
	return &gtsmodel.Token{
		ID:          "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		ClientID:    "01FEEDMF6C0QD589MRK7919Z0R",
		UserID:      "01FEK0BFJKYXB4Y51RBQ7P5P79",
		RedirectURI: "oauth2redirect://com.keylesspalace.tusky/",
		Scope:       "read write follow",
	}
}

type TokenValidateTestSuite struct {
	suite.Suite
}

func (suite *TokenValidateTestSuite) TestValidateTokenHappyPath() {
	// no problem here
	t := happyToken()
	err := validate.Struct(t)
	suite.NoError(err)
}

func (suite *TokenValidateTestSuite) TestValidateTokenBadID() {
	t := happyToken()

	t.ID = ""
	err := validate.Struct(t)
	suite.EqualError(err, "Key: 'Token.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	t.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(t)
	suite.EqualError(err, "Key: 'Token.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *TokenValidateTestSuite) TestValidateTokenNoCreatedAt() {
	t := happyToken()

	t.CreatedAt = time.Time{}
	err := validate.Struct(t)
	suite.NoError(err)
}

func (suite *TokenValidateTestSuite) TestValidateTokenRedirectURI() {
	t := happyToken()

	t.RedirectURI = "invalid-uri"
	err := validate.Struct(t)
	suite.EqualError(err, "Key: 'Token.RedirectURI' Error:Field validation for 'RedirectURI' failed on the 'uri' tag")

	t.RedirectURI = ""
	err = validate.Struct(t)
	suite.EqualError(err, "Key: 'Token.RedirectURI' Error:Field validation for 'RedirectURI' failed on the 'required' tag")

	t.RedirectURI = "urn:ietf:wg:oauth:2.0:oob"
	err = validate.Struct(t)
	suite.NoError(err)
}

func (suite *TokenValidateTestSuite) TestValidateTokenScope() {
	t := happyToken()

	t.Scope = ""
	err := validate.Struct(t)
	suite.EqualError(err, "Key: 'Token.Scope' Error:Field validation for 'Scope' failed on the 'required' tag")
}

func TestTokenValidateTestSuite(t *testing.T) {
	suite.Run(t, new(TokenValidateTestSuite))
}
