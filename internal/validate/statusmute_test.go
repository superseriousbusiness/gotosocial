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

func happyStatusMute() *gtsmodel.StatusMute {
	return &gtsmodel.StatusMute{
		ID:              "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:       time.Now(),
		AccountID:       "01FE96MAE58MXCE5C4SSMEMCEK",
		Account:         nil,
		TargetAccountID: "01FE96MXRHWZHKC0WH5FT82H1A",
		TargetAccount:   nil,
		StatusID:        "01FE96NBPNJNY26730FT6GZTFE",
		Status:          nil,
	}
}

type StatusMuteValidateTestSuite struct {
	suite.Suite
}

func (suite *StatusMuteValidateTestSuite) TestValidateStatusMuteHappyPath() {
	// no problem here
	m := happyStatusMute()
	err := validate.Struct(m)
	suite.NoError(err)
}

func (suite *StatusMuteValidateTestSuite) TestValidateStatusMuteBadID() {
	m := happyStatusMute()

	m.ID = ""
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'StatusMute.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	m.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'StatusMute.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *StatusMuteValidateTestSuite) TestValidateStatusMuteDodgyStatusID() {
	m := happyStatusMute()

	m.StatusID = "9HZJ76B6VXSKF"
	err := validate.Struct(m)
	suite.EqualError(err, "Key: 'StatusMute.StatusID' Error:Field validation for 'StatusID' failed on the 'ulid' tag")

	m.StatusID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!!!!!!!!!!!!"
	err = validate.Struct(m)
	suite.EqualError(err, "Key: 'StatusMute.StatusID' Error:Field validation for 'StatusID' failed on the 'ulid' tag")
}

func (suite *StatusMuteValidateTestSuite) TestValidateStatusMuteNoCreatedAt() {
	m := happyStatusMute()

	m.CreatedAt = time.Time{}
	err := validate.Struct(m)
	suite.NoError(err)
}

func TestStatusMuteValidateTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMuteValidateTestSuite))
}
