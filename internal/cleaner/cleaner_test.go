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
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/cleaner"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
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
	suite.state.Workers.Scheduler.Start()

	// Initialize test database.
	_ = testrig.NewTestDB(&suite.state)
	testrig.StandardDBSetup(suite.state.DB, nil)

	// Initialize test storage (in-memory).
	suite.state.Storage = testrig.NewInMemoryStorage()

	// Initialize test cleaner instance.
	testrig.StartNoopWorkers(&suite.state)
	suite.cleaner = cleaner.New(&suite.state)

	// Allocate new test model emojis.
	suite.emojis = testrig.NewTestEmojis()
}

func (suite *CleanerTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.state.DB)
	testrig.StopWorkers(&suite.state)
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
