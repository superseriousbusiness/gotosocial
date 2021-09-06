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

package trans_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	trans "github.com/superseriousbusiness/gotosocial/internal/trans/model"
)

type BlockTestSuite struct {
	ModelTestSuite
}

func (suite *AccountTestSuite) TestBlocksIdempotent() {
	// we should be able to get all blocks with the simple trans.Block struct
	blocks := []*trans.Block{}
	err := suite.db.GetAll(context.Background(), &blocks)
	suite.NoError(err)
	suite.NotEmpty(blocks)

	// we should be able to marshal the blocks to json with no problems
	b, err := json.Marshal(&blocks)
	suite.NoError(err)
	suite.NotNil(b)
	suite.T().Log(string(b))

	// the json should be idempotent
	mBlocks := []*trans.Block{}
	err = json.Unmarshal(b, &mBlocks)
	suite.NoError(err)
	suite.NotEmpty(mBlocks)
	suite.EqualValues(blocks, mBlocks)
}

func TestBlockTestSuite(t *testing.T) {
	suite.Run(t, &BlockTestSuite{})
}
