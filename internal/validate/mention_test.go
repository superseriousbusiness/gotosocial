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

func happyMention() *gtsmodel.Mention {
	return &gtsmodel.Mention{
		ID:               "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		OriginAccountID:  "01FE96MAE58MXCE5C4SSMEMCEK",
		OriginAccountURI: "https://some-instance/accounts/bleepbloop",
		OriginAccount:    nil,
		TargetAccountID:  "01FE96MXRHWZHKC0WH5FT82H1A",
		TargetAccount:    nil,
		StatusID:         "01FE96NBPNJNY26730FT6GZTFE",
		Status:           nil,
	}
}

type MentionValidateTestSuite struct {
	suite.Suite
}

func (suite *MentionValidateTestSuite) TestValidateMentionHappyPath() {
	// no problem here
	m := happyMention()
	err := validate.Struct(m)
	suite.NoError(err)
}

func (suite *MentionValidateTestSuite) TestValidateMentionBadID() {
	m := happyMention()

	m.ID = ""
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'Mention.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	m.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'Mention.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *MentionValidateTestSuite) TestValidateMentionAccountURI() {
	m := happyMention()

	m.OriginAccountURI = ""
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'Mention.OriginAccountURI' Error:Field validation for 'OriginAccountURI' failed on the 'url' tag")

	m.OriginAccountURI = "---------------------------"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'Mention.OriginAccountURI' Error:Field validation for 'OriginAccountURI' failed on the 'url' tag")
}

func (suite *MentionValidateTestSuite) TestValidateMentionDodgyStatusID() {
	m := happyMention()

	m.StatusID = "9HZJ76B6VXSKF"
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'Mention.StatusID' Error:Field validation for 'StatusID' failed on the 'ulid' tag")

	m.StatusID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!!!!!!!!!!!!"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'Mention.StatusID' Error:Field validation for 'StatusID' failed on the 'ulid' tag")
}

func (suite *MentionValidateTestSuite) TestValidateMentionNoCreatedAt() {
	m := happyMention()

	m.CreatedAt = time.Time{}
	err := validate.Struct(m)
	suite.NoError(err)
}

func TestMentionValidateTestSuite(t *testing.T) {
	suite.Run(t, new(MentionValidateTestSuite))
}
