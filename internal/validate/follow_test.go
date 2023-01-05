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

func happyFollow() *gtsmodel.Follow {
	return &gtsmodel.Follow{
		ID:              "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		AccountID:       "01FE96MAE58MXCE5C4SSMEMCEK",
		Account:         nil,
		TargetAccountID: "01FE96MXRHWZHKC0WH5FT82H1A",
		TargetAccount:   nil,
		URI:             "https://example.org/users/user1/activity/follow/01FE91RJR88PSEEE30EV35QR8N",
	}
}

type FollowValidateTestSuite struct {
	suite.Suite
}

func (suite *FollowValidateTestSuite) TestValidateFollowHappyPath() {
	// no problem here
	f := happyFollow()
	err := validate.Struct(f)
	suite.NoError(err)
}

func (suite *FollowValidateTestSuite) TestValidateFollowBadID() {
	f := happyFollow()

	f.ID = ""
	err := validate.Struct(f)
	suite.EqualError(err, "Key: 'Follow.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	f.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(f)
	suite.EqualError(err, "Key: 'Follow.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *FollowValidateTestSuite) TestValidateFollowNoCreatedAt() {
	f := happyFollow()

	f.CreatedAt = time.Time{}
	err := validate.Struct(f)
	suite.NoError(err)
}

func (suite *FollowValidateTestSuite) TestValidateFollowNoURI() {
	f := happyFollow()

	f.URI = ""
	err := validate.Struct(f)
	suite.EqualError(err, "Key: 'Follow.URI' Error:Field validation for 'URI' failed on the 'required' tag")

	f.URI = "this-is-not-a-valid-url"
	err = validate.Struct(f)
	suite.EqualError(err, "Key: 'Follow.URI' Error:Field validation for 'URI' failed on the 'url' tag")
}

func TestFollowValidateTestSuite(t *testing.T) {
	suite.Run(t, new(FollowValidateTestSuite))
}
