/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package gtsmodel_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	d := happyBlock()
	err := gtsmodel.ValidateStruct(*d)
	suite.NoError(err)
}

func (suite *BlockValidateTestSuite) TestValidateBlockBadID() {
	d := happyBlock()

	d.ID = ""
	err := gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.ID' Error:Field validation for 'ID' failed on the 'required' tag")

	d.ID = "01FE96W293ZPRG9FQQP48HK8N001FE96W32AT24VYBGM12WN3GKB"
	err = gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.ID' Error:Field validation for 'ID' failed on the 'ulid' tag")
}

func (suite *BlockValidateTestSuite) TestValidateBlockNoCreatedAt() {
	d := happyBlock()

	d.CreatedAt = time.Time{}
	err := gtsmodel.ValidateStruct(*d)
	suite.NoError(err)
}

func (suite *BlockValidateTestSuite) TestValidateBlockCreatedByAccountID() {
	d := happyBlock()

	d.AccountID = ""
	err := gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.AccountID' Error:Field validation for 'AccountID' failed on the 'required' tag")

	d.AccountID = "this-is-not-a-valid-ulid"
	err = gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.AccountID' Error:Field validation for 'AccountID' failed on the 'ulid' tag")
}

func (suite *BlockValidateTestSuite) TestValidateBlockTargetAccountID() {
	d := happyBlock()

	d.TargetAccountID = "invalid-ulid"
	err := gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.TargetAccountID' Error:Field validation for 'TargetAccountID' failed on the 'ulid' tag")

	d.TargetAccountID = "01FEEDHX4G7EGHF5GD9E82Y51Q"
	err = gtsmodel.ValidateStruct(*d)
	suite.NoError(err)

	d.TargetAccountID = ""
	err = gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.TargetAccountID' Error:Field validation for 'TargetAccountID' failed on the 'required' tag")
}

func (suite *BlockValidateTestSuite) TestValidateBlockURI() {
	d := happyBlock()

	d.URI = "invalid-uri"
	err := gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.URI' Error:Field validation for 'URI' failed on the 'url' tag")

	d.URI = ""
	err = gtsmodel.ValidateStruct(*d)
	suite.EqualError(err, "Key: 'Block.URI' Error:Field validation for 'URI' failed on the 'required' tag")
}

func TestBlockValidateTestSuite(t *testing.T) {
	suite.Run(t, new(BlockValidateTestSuite))
}
