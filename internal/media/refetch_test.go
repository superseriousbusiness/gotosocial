/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

package media_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RefetchTestSuite struct {
	MediaStandardTestSuite
}

func (suite *RefetchTestSuite) TestRefetchEmojisNothingToDo() {
	ctx := context.Background()

	adminAccount := suite.testAccounts["admin_account"]
	transport, err := suite.transportController.NewTransportForUsername(ctx, adminAccount.Username)
	if err != nil {
		suite.FailNow(err.Error())
	}

	refetched, err := suite.manager.RefetchEmojis(ctx, "", transport.DereferenceMedia)
	suite.NoError(err)
	suite.Equal(0, refetched)
}

func (suite *RefetchTestSuite) TestRefetchEmojis() {
	ctx := context.Background()

	if err := suite.storage.Delete(ctx, suite.testEmojis["yell"].ImagePath); err != nil {
		suite.FailNow(err.Error())
	}

	adminAccount := suite.testAccounts["admin_account"]
	transport, err := suite.transportController.NewTransportForUsername(ctx, adminAccount.Username)
	if err != nil {
		suite.FailNow(err.Error())
	}

	refetched, err := suite.manager.RefetchEmojis(ctx, "", transport.DereferenceMedia)
	suite.NoError(err)
	suite.Equal(1, refetched)
}

func (suite *RefetchTestSuite) TestRefetchEmojisLocal() {
	ctx := context.Background()

	// delete the image for a LOCAL emoji
	if err := suite.storage.Delete(ctx, suite.testEmojis["rainbow"].ImagePath); err != nil {
		suite.FailNow(err.Error())
	}

	adminAccount := suite.testAccounts["admin_account"]
	transport, err := suite.transportController.NewTransportForUsername(ctx, adminAccount.Username)
	if err != nil {
		suite.FailNow(err.Error())
	}

	refetched, err := suite.manager.RefetchEmojis(ctx, "", transport.DereferenceMedia)
	suite.NoError(err)
	suite.Equal(0, refetched) // shouldn't refetch anything because local
}

func TestRefetchTestSuite(t *testing.T) {
	suite.Run(t, &RefetchTestSuite{})
}
