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

func happyEmailDomainBlock() *gtsmodel.EmailDomainBlock {
	return &gtsmodel.EmailDomainBlock{
		ID:                 "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Domain:             "baddudes.suck",
		CreatedByAccountID: "01FEED79PRMVWPRMFHFQM8MJQN",
	}
}

type EmailDomainBlockValidateTestSuite struct {
	suite.Suite
}

func (suite *EmailDomainBlockValidateTestSuite) TestValidateEmailDomainBlockHappyPath() {
	// no problem here
	e := happyEmailDomainBlock()
	err := validate.Struct(e)
	suite.NoError(err)
}

func (suite *EmailDomainBlockValidateTestSuite) TestValidateEmailDomainBlockBadID() {
	e := happyEmailDomainBlock()

	e.ID = ""
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'EmailDomainBlock.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	e.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'EmailDomainBlock.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *EmailDomainBlockValidateTestSuite) TestValidateEmailDomainBlockNoCreatedAt() {
	e := happyEmailDomainBlock()

	e.CreatedAt = time.Time{}
	err := validate.Struct(e)
	suite.NoError(err)
}

func (suite *EmailDomainBlockValidateTestSuite) TestValidateEmailDomainBlockBadDomain() {
	e := happyEmailDomainBlock()

	e.Domain = ""
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'EmailDomainBlock.Domain' Error:Field validation for 'Domain' failed on the 'required' tag")

	e.Domain = "this-is-not-a-valid-domain"
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'EmailDomainBlock.Domain' Error:Field validation for 'Domain' failed on the 'fqdn' tag")
}

func (suite *EmailDomainBlockValidateTestSuite) TestValidateEmailDomainBlockCreatedByAccountID() {
	e := happyEmailDomainBlock()

	e.CreatedByAccountID = ""
	err := validate.Struct(e)
	suite.EqualError(err, "Key: 'EmailDomainBlock.CreatedByAccountID' Error:Field validation for 'CreatedByAccountID' failed on the 'required' tag")

	e.CreatedByAccountID = "this-is-not-a-valid-ulid"
	err = validate.Struct(e)
	suite.EqualError(err, "Key: 'EmailDomainBlock.CreatedByAccountID' Error:Field validation for 'CreatedByAccountID' failed on the 'ulid' tag")
}

func TestEmailDomainBlockValidateTestSuite(t *testing.T) {
	suite.Run(t, new(EmailDomainBlockValidateTestSuite))
}
