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

package typeutils_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/go-fed/activity/streams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type InternalToASTestSuite struct {
	TypeUtilsTestSuite
}

func (suite *InternalToASTestSuite) TestAccountToAS() {
	testAccount := suite.testAccounts["local_account_1"] // take zork for this test

	asPerson, err := suite.typeconverter.AccountToAS(context.Background(), testAccount)
	assert.NoError(suite.T(), err)

	ser, err := streams.Serialize(asPerson)
	assert.NoError(suite.T(), err)

	bytes, err := json.Marshal(ser)
	assert.NoError(suite.T(), err)

	fmt.Println(string(bytes))
	// TODO: write assertions here, rn we're just eyeballing the output
}

func TestInternalToASTestSuite(t *testing.T) {
	suite.Run(t, new(InternalToASTestSuite))
}
