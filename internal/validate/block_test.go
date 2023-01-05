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

func happyBlock() *gtsmodel.Block {
	return &gtsmodel.Block{
		ID:              "01FE91RJR88PSEEE30EV35QR8N",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             "https://example.org/accounts/someone/blocks/01FE91RJR88PSEEE30EV35QR8N",
		AccountID:       "01FEED79PRMVWPRMFHFQM8MJQN",
		Account:         nil,
		TargetAccountID: "01FEEDMF6C0QD589MRK7919Z0R",
		TargetAccount:   nil,
	}
}

type BlockValidateTestSuite struct {
	suite.Suite
}

func (suite *BlockValidateTestSuite) TestValidateBlockHappyPath() {
	// no problem here
	b := happyBlock()
	err := validate.Struct(b)
	suite.NoError(err)
}

func (suite *BlockValidateTestSuite) TestValidateBlockBadID() {
	b := happyBlock()

	b.ID = ""
	err := validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	b.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *BlockValidateTestSuite) TestValidateBlockNoCreatedAt() {
	b := happyBlock()

	b.CreatedAt = time.Time{}
	err := validate.Struct(b)
	suite.NoError(err)
}

func (suite *BlockValidateTestSuite) TestValidateBlockCreatedByAccountID() {
	b := happyBlock()

	b.AccountID = ""
	err := validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.AccountID' Error:Field validation for 'AccountID' failed on the 'required' tag")

	b.AccountID = "this-is-not-a-valid-ulid"
	err = validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.AccountID' Error:Field validation for 'AccountID' failed on the 'ulid' tag")
}

func (suite *BlockValidateTestSuite) TestValidateBlockTargetAccountID() {
	b := happyBlock()

	b.TargetAccountID = "invalid-ulid"
	err := validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.TargetAccountID' Error:Field validation for 'TargetAccountID' failed on the 'ulid' tag")

	b.TargetAccountID = "01FEEDHX4G7EGHF5GD9E82Y51Q"
	err = validate.Struct(b)
	suite.NoError(err)

	b.TargetAccountID = ""
	err = validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.TargetAccountID' Error:Field validation for 'TargetAccountID' failed on the 'required' tag")
}

func (suite *BlockValidateTestSuite) TestValidateBlockURI() {
	b := happyBlock()

	b.URI = "invalid-uri"
	err := validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.URI' Error:Field validation for 'URI' failed on the 'url' tag")

	b.URI = ""
	err = validate.Struct(b)
	suite.EqualError(err, "Key: 'Block.URI' Error:Field validation for 'URI' failed on the 'required' tag")
}

func TestBlockValidateTestSuite(t *testing.T) {
	suite.Run(t, new(BlockValidateTestSuite))
}
