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

func happyClient() *gtsmodel.Client {
	return &gtsmodel.Client{
		ID:        "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Secret:    "bd740cf1-024a-4e4d-8c39-866538f52fe6",
		Domain:    "oauth2redirect://com.keylesspalace.tusky/",
		UserID:    "01FEEDMF6C0QD589MRK7919Z0R",
	}
}

type ClientValidateTestSuite struct {
	suite.Suite
}

func (suite *ClientValidateTestSuite) TestValidateClientHappyPath() {
	// no problem here
	c := happyClient()
	err := validate.Struct(c)
	suite.NoError(err)
}

func (suite *ClientValidateTestSuite) TestValidateClientBadID() {
	c := happyClient()

	c.ID = ""
	err := validate.Struct(c)
	suite.EqualError(err, "Key: 'Client.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	c.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(c)
	suite.EqualError(err, "Key: 'Client.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *ClientValidateTestSuite) TestValidateClientNoCreatedAt() {
	c := happyClient()

	c.CreatedAt = time.Time{}
	err := validate.Struct(c)
	suite.NoError(err)
}

func (suite *ClientValidateTestSuite) TestValidateClientDomain() {
	c := happyClient()

	c.Domain = "invalid-uri"
	err := validate.Struct(c)
	suite.EqualError(err, "Key: 'Client.Domain' Error:Field validation for 'Domain' failed on the 'uri' tag")

	c.Domain = ""
	err = validate.Struct(c)
	suite.EqualError(err, "Key: 'Client.Domain' Error:Field validation for 'Domain' failed on the 'required' tag")

	c.Domain = "urn:ietf:wg:oauth:2.0:oob"
	err = validate.Struct(c)
	suite.NoError(err)
}

func (suite *ClientValidateTestSuite) TestValidateSecret() {
	c := happyClient()

	c.Secret = "invalid-uuid"
	err := validate.Struct(c)
	suite.EqualError(err, "Key: 'Client.Secret' Error:Field validation for 'Secret' failed on the 'uuid' tag")

	c.Secret = ""
	err = validate.Struct(c)
	suite.EqualError(err, "Key: 'Client.Secret' Error:Field validation for 'Secret' failed on the 'required' tag")
}

func TestClientValidateTestSuite(t *testing.T) {
	suite.Run(t, new(ClientValidateTestSuite))
}
