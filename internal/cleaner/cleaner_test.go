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

package cleaner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/cleaner"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type CleanerTestSuite struct {
	state   state.State
	cleaner *cleaner.Cleaner
	emojis  map[string]*gtsmodel.Emoji
	suite.Suite
}

func TestCleanerTestSuite(t *testing.T) {
	suite.Run(t, &CleanerTestSuite{})
}

func (suite *CleanerTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
}

func (suite *CleanerTestSuite) SetupTest() {
	// Initialize gts caches.
	suite.state.Caches.Init()

	// Ensure scheduler started (even if unused).
	suite.state.Workers.Scheduler.Start(nil)

	// Initialize test database.
	_ = testrig.NewTestDB(&suite.state)
	testrig.StandardDBSetup(suite.state.DB, nil)

	// Initialize test storage (in-memory).
	suite.state.Storage = testrig.NewInMemoryStorage()

	// Initialize test cleaner instance.
	suite.cleaner = cleaner.New(&suite.state)

	// Allocate new test model emojis.
	suite.emojis = testrig.NewTestEmojis()
}

func (suite *CleanerTestSuite) TearDownTest() {
	_ = suite.state.DB.Stop(context.Background())
	suite.state.DB = nil
}

// mapvals extracts a slice of values from the values contained within the map.
func mapvals[Key comparable, Val any](m map[Key]Val) []Val {
	var i int
	vals := make([]Val, len(m))
	for _, val := range m {
		vals[i] = val
		i++
	}
	return vals
}
