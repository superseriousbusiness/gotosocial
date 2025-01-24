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
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type WebPushTestSuite struct {
	BunDBStandardTestSuite
}

// Get the text fixture VAPID key pair.
func (suite *WebPushTestSuite) TestGetVAPIDKeyPair() {
	ctx := context.Background()

	vapidKeyPair, err := suite.db.GetVAPIDKeyPair(ctx)
	suite.NoError(err)
	if !suite.NotNil(vapidKeyPair) {
		suite.FailNow("Got a nil VAPID key pair, can't continue")
	}
	suite.NotEmpty(vapidKeyPair.Private)
	suite.NotEmpty(vapidKeyPair.Public)

	// Get it again. It should be the same one.
	vapidKeyPair2, err := suite.db.GetVAPIDKeyPair(ctx)
	suite.NoError(err)
	if suite.NotNil(vapidKeyPair2) {
		suite.Equal(vapidKeyPair.Private, vapidKeyPair2.Private)
		suite.Equal(vapidKeyPair.Public, vapidKeyPair2.Public)
	}
}

// Generate a VAPID key pair when there isn't one.
func (suite *WebPushTestSuite) TestGenerateVAPIDKeyPair() {
	ctx := context.Background()

	// Delete the text fixture VAPID key pair.
	if err := suite.db.DeleteVAPIDKeyPair(ctx); !suite.NoError(err) {
		suite.FailNow("Test setup failed: DB error deleting fixture VAPID key pair: %v", err)
	}

	// Get a new one.
	vapidKeyPair, err := suite.db.GetVAPIDKeyPair(ctx)
	suite.NoError(err)
	if !suite.NotNil(vapidKeyPair) {
		suite.FailNow("Got a nil VAPID key pair, can't continue")
	}
	suite.NotEmpty(vapidKeyPair.Private)
	suite.NotEmpty(vapidKeyPair.Public)

	// Get it again. It should be the same one.
	vapidKeyPair2, err := suite.db.GetVAPIDKeyPair(ctx)
	suite.NoError(err)
	if suite.NotNil(vapidKeyPair2) {
		suite.Equal(vapidKeyPair.Private, vapidKeyPair2.Private)
		suite.Equal(vapidKeyPair.Public, vapidKeyPair2.Public)
	}
}

func TestWebPushTestSuite(t *testing.T) {
	suite.Run(t, new(WebPushTestSuite))
}
