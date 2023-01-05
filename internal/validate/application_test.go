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

func happyApplication() *gtsmodel.Application {
	return &gtsmodel.Application{
		ID:           "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Name:         "Tusky",
		Website:      "https://tusky.app",
		RedirectURI:  "oauth2redirect://com.keylesspalace.tusky/",
		ClientID:     "01FEEDMF6C0QD589MRK7919Z0R",
		ClientSecret: "bd740cf1-024a-4e4d-8c39-866538f52fe6",
		Scopes:       "read write follow",
	}
}

type ApplicationValidateTestSuite struct {
	suite.Suite
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationHappyPath() {
	// no problem here
	a := happyApplication()
	err := validate.Struct(a)
	suite.NoError(err)
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationBadID() {
	a := happyApplication()

	a.ID = ""
	err := validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	a.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationNoCreatedAt() {
	a := happyApplication()

	a.CreatedAt = time.Time{}
	err := validate.Struct(a)
	suite.NoError(err)
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationName() {
	a := happyApplication()

	a.Name = ""
	err := validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.Name' Error:Field validation for 'Name' failed on the 'required' tag")
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationWebsite() {
	a := happyApplication()

	a.Website = "invalid-website"
	err := validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.Website' Error:Field validation for 'Website' failed on the 'url' tag")

	a.Website = ""
	err = validate.Struct(a)
	suite.NoError(err)
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationRedirectURI() {
	a := happyApplication()

	a.RedirectURI = "invalid-uri"
	err := validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.RedirectURI' Error:Field validation for 'RedirectURI' failed on the 'uri' tag")

	a.RedirectURI = ""
	err = validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.RedirectURI' Error:Field validation for 'RedirectURI' failed on the 'required' tag")

	a.RedirectURI = "urn:ietf:wg:oauth:2.0:oob"
	err = validate.Struct(a)
	suite.NoError(err)
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationClientSecret() {
	a := happyApplication()

	a.ClientSecret = "invalid-uuid"
	err := validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.ClientSecret' Error:Field validation for 'ClientSecret' failed on the 'uuid' tag")

	a.ClientSecret = ""
	err = validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.ClientSecret' Error:Field validation for 'ClientSecret' failed on the 'required' tag")
}

func (suite *ApplicationValidateTestSuite) TestValidateApplicationScopes() {
	a := happyApplication()

	a.Scopes = ""
	err := validate.Struct(a)
	suite.EqualError(err, "Key: 'Application.Scopes' Error:Field validation for 'Scopes' failed on the 'required' tag")
}

func TestApplicationValidateTestSuite(t *testing.T) {
	suite.Run(t, new(ApplicationValidateTestSuite))
}
