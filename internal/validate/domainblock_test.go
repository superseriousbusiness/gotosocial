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
	"github.com/superseriousbusiness/gotosocial/testrig"
)

func happyDomainBlock() *gtsmodel.DomainBlock {
	return &gtsmodel.DomainBlock{
		ID:                 "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Domain:             "baddudes.suck",
		CreatedByAccountID: "01FEED79PRMVWPRMFHFQM8MJQN",
		PrivateComment:     "we don't like em",
		PublicComment:      "poo poo dudes",
		Obfuscate:          testrig.FalseBool(),
		SubscriptionID:     "",
	}
}

type DomainBlockValidateTestSuite struct {
	suite.Suite
}

func (suite *DomainBlockValidateTestSuite) TestValidateDomainBlockHappyPath() {
	// no problem here
	d := happyDomainBlock()
	err := validate.Struct(d)
	suite.NoError(err)
}

func (suite *DomainBlockValidateTestSuite) TestValidateDomainBlockBadID() {
	d := happyDomainBlock()

	d.ID = ""
	err := validate.Struct(d)
	suite.EqualError(err, "Key: 'DomainBlock.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	d.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(d)
	suite.EqualError(err, "Key: 'DomainBlock.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *DomainBlockValidateTestSuite) TestValidateDomainBlockNoCreatedAt() {
	d := happyDomainBlock()

	d.CreatedAt = time.Time{}
	err := validate.Struct(d)
	suite.NoError(err)
}

func (suite *DomainBlockValidateTestSuite) TestValidateDomainBlockBadDomain() {
	d := happyDomainBlock()

	d.Domain = ""
	err := validate.Struct(d)
	suite.EqualError(err, "Key: 'DomainBlock.Domain' Error:Field validation for 'Domain' failed on the 'required' tag")

	d.Domain = "this-is-not-a-valid-domain"
	err = validate.Struct(d)
	suite.EqualError(err, "Key: 'DomainBlock.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")
}

func (suite *DomainBlockValidateTestSuite) TestValidateDomainBlockCreatedByAccountID() {
	d := happyDomainBlock()

	d.CreatedByAccountID = ""
	err := validate.Struct(d)
	suite.EqualError(err, "Key: 'DomainBlock.CreatedByAccountID' Error:Field validation for 'CreatedByAccountID' failed on the 'required' tag")

	d.CreatedByAccountID = "this-is-not-a-valid-ulid"
	err = validate.Struct(d)
	suite.EqualError(err, "Key: 'DomainBlock.CreatedByAccountID' Error:Field validation for 'CreatedByAccountID' failed on the 'ulid' tag")
}

func (suite *DomainBlockValidateTestSuite) TestValidateDomainBlockComments() {
	d := happyDomainBlock()

	d.PrivateComment = ""
	d.PublicComment = ""
	err := validate.Struct(d)
	suite.NoError(err)
}

func (suite *DomainBlockValidateTestSuite) TestValidateDomainSubscriptionID() {
	d := happyDomainBlock()

	d.SubscriptionID = "invalid-ulid"
	err := validate.Struct(d)
	suite.EqualError(err, "Key: 'DomainBlock.SubscriptionID' Error:Field validation for 'SubscriptionID' failed on the 'ulid' tag")

	d.SubscriptionID = "01FEEDHX4G7EGHF5GD9E82Y51Q"
	err = validate.Struct(d)
	suite.NoError(err)
}

func TestDomainBlockValidateTestSuite(t *testing.T) {
	suite.Run(t, new(DomainBlockValidateTestSuite))
}
