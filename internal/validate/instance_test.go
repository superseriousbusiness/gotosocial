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

func happyInstance() *gtsmodel.Instance {
	return &gtsmodel.Instance{
		ID:                     "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		Domain:                 "example.org",
		Title:                  "Example Instance",
		URI:                    "https://example.org",
		SuspendedAt:            time.Time{},
		DomainBlockID:          "",
		DomainBlock:            nil,
		ShortDescription:       "This is a description for the example/testing instance.",
		Description:            "This is a way longer description for the example/testing instance!",
		Terms:                  "Don't be a knobhead.",
		ContactEmail:           "admin@example.org",
		ContactAccountUsername: "admin",
		ContactAccountID:       "01FEE20H5QWHJDEXAEE9G96PR0",
		ContactAccount:         nil,
		Reputation:             420,
		Version:                "gotosocial 0.1.0",
	}
}

type InstanceValidateTestSuite struct {
	suite.Suite
}

func (suite *InstanceValidateTestSuite) TestValidateInstanceHappyPath() {
	// no problem here
	m := happyInstance()
	err := validate.Struct(*m)
	suite.NoError(err)
}

func (suite *InstanceValidateTestSuite) TestValidateInstanceBadID() {
	m := happyInstance()

	m.ID = ""
	err := validate.Struct(*m)
	suite.EqualError(err, "Key: 'Instance.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	m.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(*m)
	suite.EqualError(err, "Key: 'Instance.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *InstanceValidateTestSuite) TestValidateInstanceAccountURI() {
	i := happyInstance()

	i.URI = ""
	err := validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.URI' Error:Field validation for 'URI' failed on the 'required' tag")

	i.URI = "---------------------------"
	err = validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.URI' Error:Field validation for 'URI' failed on the 'url' tag")
}

func (suite *InstanceValidateTestSuite) TestValidateInstanceDodgyAccountID() {
	i := happyInstance()

	i.ContactAccountID = "9HZJ76B6VXSKF"
	err := validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.ContactAccountID' Error:Field validation for 'ContactAccountID' failed on the 'ulid' tag")

	i.ContactAccountID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!!!!!!!!!!!!"
	err = validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.ContactAccountID' Error:Field validation for 'ContactAccountID' failed on the 'ulid' tag")

	i.ContactAccountID = ""
	err = validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.ContactAccountID' Error:Field validation for 'ContactAccountID' failed on the 'required_with' tag")

	i.ContactAccountUsername = ""
	err = validate.Struct(i)
	suite.NoError(err)
}

func (suite *InstanceValidateTestSuite) TestValidateInstanceDomain() {
	i := happyInstance()

	i.Domain = "poopoo"
	err := validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")

	i.Domain = ""
	err = validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.Domain' Error:Field validation for 'Domain' failed on the 'required' tag")

	i.Domain = "https://aaaaaaaaaaaaah.org"
	err = validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")
}

func (suite *InstanceValidateTestSuite) TestValidateInstanceContactEmail() {
	i := happyInstance()

	i.ContactEmail = "poopoo"
	err := validate.Struct(i)
	suite.EqualError(err, "Key: 'Instance.ContactEmail' Error:Field validation for 'ContactEmail' failed on the 'email' tag")

	i.ContactEmail = ""
	err = validate.Struct(i)
	suite.NoError(err)
}

func (suite *InstanceValidateTestSuite) TestValidateInstanceNoCreatedAt() {
	i := happyInstance()

	i.CreatedAt = time.Time{}
	err := validate.Struct(i)
	suite.NoError(err)
}

func TestInstanceValidateTestSuite(t *testing.T) {
	suite.Run(t, new(InstanceValidateTestSuite))
}
