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

func happyTag() *gtsmodel.Tag {
	return &gtsmodel.Tag{
		ID:                     "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
		URL:                    "https://example.org/tags/some_tag",
		Name:                   "some_tag",
		FirstSeenFromAccountID: "01FE91SR5P2GW06K3AJ98P72MT",
		Useable:                testrig.TrueBool(),
		Listable:               testrig.TrueBool(),
		LastStatusAt:           time.Now(),
	}
}

type TagValidateTestSuite struct {
	suite.Suite
}

func (suite *TagValidateTestSuite) TestValidateTagHappyPath() {
	// no problem here
	t := happyTag()
	err := validate.Struct(t)
	suite.NoError(err)
}

func (suite *TagValidateTestSuite) TestValidateTagNoName() {
	t := happyTag()
	t.Name = ""

	err := validate.Struct(t)
	suite.EqualError(err, "Key: 'Tag.Name' Error:Field validation for 'Name' failed on the 'required' tag")
}

func (suite *TagValidateTestSuite) TestValidateTagBadURL() {
	t := happyTag()

	t.URL = ""
	err := validate.Struct(t)
	suite.EqualError(err, "Key: 'Tag.URL' Error:Field validation for 'URL' failed on the 'required' tag")

	t.URL = "no-schema.com"
	err = validate.Struct(t)
	suite.EqualError(err, "Key: 'Tag.URL' Error:Field validation for 'URL' failed on the 'url' tag")

	t.URL = "justastring"
	err = validate.Struct(t)
	suite.EqualError(err, "Key: 'Tag.URL' Error:Field validation for 'URL' failed on the 'url' tag")

	t.URL = "https://aaa\n\n\naaaaaaaa"
	err = validate.Struct(t)
	suite.EqualError(err, "Key: 'Tag.URL' Error:Field validation for 'URL' failed on the 'url' tag")
}

func (suite *TagValidateTestSuite) TestValidateTagNoFirstSeenFromAccountID() {
	t := happyTag()
	t.FirstSeenFromAccountID = ""

	err := validate.Struct(t)
	suite.NoError(err)
}

func TestTagValidateTestSuite(t *testing.T) {
	suite.Run(t, new(TagValidateTestSuite))
}
