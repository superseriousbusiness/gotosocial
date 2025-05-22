// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bundb_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type InstanceTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *InstanceTestSuite) TestCountInstanceUsers() {
	count, err := suite.db.CountInstanceUsers(suite.T().Context(), config.GetHost())
	suite.NoError(err)
	suite.Equal(5, count)
}

func (suite *InstanceTestSuite) TestCountInstanceUsersRemote() {
	count, err := suite.db.CountInstanceUsers(suite.T().Context(), "fossbros-anonymous.io")
	suite.NoError(err)
	suite.Equal(1, count)
}

func (suite *InstanceTestSuite) TestCountInstanceStatuses() {
	count, err := suite.db.CountInstanceStatuses(suite.T().Context(), config.GetHost())
	suite.NoError(err)
	suite.Equal(23, count)
}

func (suite *InstanceTestSuite) TestCountInstanceStatusesRemote() {
	count, err := suite.db.CountInstanceStatuses(suite.T().Context(), "fossbros-anonymous.io")
	suite.NoError(err)
	suite.Equal(4, count)
}

func (suite *InstanceTestSuite) TestCountInstanceDomains() {
	count, err := suite.db.CountInstanceDomains(suite.T().Context(), config.GetHost())
	suite.NoError(err)
	suite.Equal(2, count)
}

func (suite *InstanceTestSuite) TestGetInstanceOK() {
	instance, err := suite.db.GetInstance(suite.T().Context(), "localhost:8080")
	suite.NoError(err)
	suite.NotNil(instance)
}

func (suite *InstanceTestSuite) TestGetInstanceNonexistent() {
	instance, err := suite.db.GetInstance(suite.T().Context(), "doesnt.exist.com")
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Nil(instance)
}

func (suite *InstanceTestSuite) TestGetInstancePeers() {
	peers, err := suite.db.GetInstancePeers(suite.T().Context(), false)
	suite.NoError(err)
	suite.Len(peers, 2)
}

func (suite *InstanceTestSuite) TestGetInstancePeersIncludeSuspended() {
	peers, err := suite.db.GetInstancePeers(suite.T().Context(), true)
	suite.NoError(err)
	suite.Len(peers, 2)
}

func (suite *InstanceTestSuite) TestGetInstanceAccounts() {
	accounts, err := suite.db.GetInstanceAccounts(suite.T().Context(), "fossbros-anonymous.io", "", 10)
	suite.NoError(err)
	suite.Len(accounts, 1)
}

func (suite *InstanceTestSuite) TestGetInstanceModeratorAddressesOK() {
	// We have one admin user by default.
	addresses, err := suite.db.GetInstanceModeratorAddresses(suite.T().Context())
	suite.NoError(err)
	suite.EqualValues([]string{"admin@example.org"}, addresses)
}

func (suite *InstanceTestSuite) TestGetInstanceModeratorAddressesZorkAsModerator() {
	// Promote zork to moderator role.
	testUser := &gtsmodel.User{}
	*testUser = *suite.testUsers["local_account_1"]
	testUser.Moderator = util.Ptr(true)
	if err := suite.db.UpdateUser(suite.T().Context(), testUser, "moderator"); err != nil {
		suite.FailNow(err.Error())
	}

	addresses, err := suite.db.GetInstanceModeratorAddresses(suite.T().Context())
	suite.NoError(err)
	suite.EqualValues([]string{"admin@example.org", "zork@example.org"}, addresses)
}

func (suite *InstanceTestSuite) TestGetInstanceModeratorAddressesNoAdmin() {
	// Demote admin from admin + moderator roles.
	testUser := &gtsmodel.User{}
	*testUser = *suite.testUsers["admin_account"]
	testUser.Admin = util.Ptr(false)
	testUser.Moderator = util.Ptr(false)
	if err := suite.db.UpdateUser(suite.T().Context(), testUser, "admin", "moderator"); err != nil {
		suite.FailNow(err.Error())
	}

	addresses, err := suite.db.GetInstanceModeratorAddresses(suite.T().Context())
	suite.ErrorIs(err, db.ErrNoEntries)
	suite.Empty(addresses)
}

func TestInstanceTestSuite(t *testing.T) {
	suite.Run(t, new(InstanceTestSuite))
}
