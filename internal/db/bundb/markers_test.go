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
// along with this program.  If not, see <http:www.gnu.org/licenses/>.

package bundb_test

import (
	"context"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

// MarkersTestSuite uses home timelines for Get tests
// and notifications timelines for Update tests
// so that multiple tests running at once can't step on each other.
type MarkersTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *MarkersTestSuite) TestGetExisting() {
	ctx := context.Background()

	// This account has home and notifications markers set.
	localAccount1 := suite.testAccounts["local_account_1"]
	marker, err := suite.db.GetMarker(ctx, localAccount1.ID, gtsmodel.MarkerNameHome)
	suite.NoError(err)
	if err != nil {
		suite.FailNow(err.Error())
	}
	// Should match our fixture.
	suite.Equal("01F8MH82FYRXD2RC6108DAJ5HB", marker.LastReadID)
}

func (suite *MarkersTestSuite) TestGetUnset() {
	ctx := context.Background()

	// This account has no markers set.
	localAccount2 := suite.testAccounts["local_account_2"]
	marker, err := suite.db.GetMarker(ctx, localAccount2.ID, gtsmodel.MarkerNameHome)
	// Should not return anything.
	suite.Nil(marker)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func (suite *MarkersTestSuite) TestUpdateExisting() {
	ctx := context.Background()

	now := time.Now()
	// This account has home and notifications markers set.
	localAccount1 := suite.testAccounts["local_account_1"]
	prevMarker := suite.testMarkers["local_account_1_notification_marker"]
	marker := &gtsmodel.Marker{
		AccountID:  localAccount1.ID,
		Name:       gtsmodel.MarkerNameNotifications,
		LastReadID: "01H57YZECGJ2ZW39H8TJWAH0KY",
	}
	err := suite.db.UpdateMarker(ctx, marker)
	suite.NoError(err)
	if err != nil {
		suite.FailNow(err.Error())
	}
	// Modifies the update and version fields of the marker as an intentional side effect.
	suite.GreaterOrEqual(marker.UpdatedAt, now)
	suite.Greater(marker.Version, prevMarker.Version)

	// Re-fetch it from the DB and confirm that we got the updated version.
	marker2, err := suite.db.GetMarker(ctx, localAccount1.ID, gtsmodel.MarkerNameNotifications)
	suite.NoError(err)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.GreaterOrEqual(marker2.UpdatedAt, now)
	suite.GreaterOrEqual(marker2.Version, prevMarker.Version)
	suite.Equal("01H57YZECGJ2ZW39H8TJWAH0KY", marker2.LastReadID)
}

func (suite *MarkersTestSuite) TestUpdateUnset() {
	ctx := context.Background()

	now := time.Now()
	// This account has no markers set.
	localAccount2 := suite.testAccounts["local_account_2"]
	marker := &gtsmodel.Marker{
		AccountID:  localAccount2.ID,
		Name:       gtsmodel.MarkerNameNotifications,
		LastReadID: "01H57ZVGMD348ZJD5WENDZDH9Z",
	}
	err := suite.db.UpdateMarker(ctx, marker)
	suite.NoError(err)
	if err != nil {
		suite.FailNow(err.Error())
	}
	// Modifies the update and version fields of the marker as an intentional side effect.
	suite.GreaterOrEqual(marker.UpdatedAt, now)
	suite.GreaterOrEqual(marker.Version, 0)

	// Re-fetch it from the DB and confirm that we got the updated version.
	marker2, err := suite.db.GetMarker(ctx, localAccount2.ID, gtsmodel.MarkerNameNotifications)
	suite.NoError(err)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.GreaterOrEqual(marker2.UpdatedAt, now)
	suite.GreaterOrEqual(marker2.Version, 0)
	suite.Equal("01H57ZVGMD348ZJD5WENDZDH9Z", marker2.LastReadID)
}

func TestMarkersTestSuite(t *testing.T) {
	suite.Run(t, new(MarkersTestSuite))
}
