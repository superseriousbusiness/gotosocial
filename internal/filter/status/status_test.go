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

package status_test

import (
	"testing"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type StatusFilterTestSuite struct {
	suite.Suite

	state     state.State
	filter    *status.Filter
	converter *typeutils.Converter

	testAccounts map[string]*gtsmodel.Account
	testFilters  map[string]*gtsmodel.Filter
	testStatuses map[string]*gtsmodel.Status
}

func (suite *StatusFilterTestSuite) SetupTest() {
	suite.state.Caches.Init()

	testrig.InitTestConfig()
	testrig.InitTestLog()

	db := testrig.NewTestDB(&suite.state)
	suite.state.DB = db

	suite.filter = status.NewFilter(&suite.state)

	suite.converter = typeutils.NewConverter(&suite.state)

	suite.testAccounts = testrig.NewTestAccounts()
	suite.testFilters = testrig.NewTestFilters()
	suite.testStatuses = testrig.NewTestStatuses()

	testrig.StandardDBSetup(suite.state.DB, nil)
}

func (suite *StatusFilterTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.state.DB)
	testrig.StopWorkers(&suite.state)
}

func (suite *StatusFilterTestSuite) TestHideFilteredStatus() {
	filtered, hide, err := suite.testFilterStatus(gtsmodel.FilterActionHide, false)
	suite.NoError(err)
	suite.True(hide)
	suite.Empty(filtered)
}

func (suite *StatusFilterTestSuite) TestWarnFilteredStatus() {
	filtered, hide, err := suite.testFilterStatus(gtsmodel.FilterActionWarn, false)
	suite.NoError(err)
	suite.False(hide)
	suite.NotEmpty(filtered)
}

func (suite *StatusFilterTestSuite) TestHideFilteredBoost() {
	filtered, hide, err := suite.testFilterStatus(gtsmodel.FilterActionHide, true)
	suite.NoError(err)
	suite.True(hide)
	suite.Empty(filtered)
}

func (suite *StatusFilterTestSuite) TestWarnFilteredBoost() {
	filtered, hide, err := suite.testFilterStatus(gtsmodel.FilterActionWarn, true)
	suite.NoError(err)
	suite.False(hide)
	suite.NotEmpty(filtered)
}

func (suite *StatusFilterTestSuite) TestHashtagWholewordStatusFiltered() {
	suite.testFilteredStatusWithHashtag(true, false)
}

func (suite *StatusFilterTestSuite) TestHashtagWholewordBoostFiltered() {
	suite.testFilteredStatusWithHashtag(true, true)
}

func (suite *StatusFilterTestSuite) TestHashtagAnywhereStatusFiltered() {
	suite.testFilteredStatusWithHashtag(false, false)
}

func (suite *StatusFilterTestSuite) TestHashtagAnywhereBoostFiltered() {
	suite.testFilteredStatusWithHashtag(false, true)
}

func (suite *StatusFilterTestSuite) testFilterStatus(action gtsmodel.FilterAction, boost bool) ([]apimodel.FilterResult, bool, error) {
	ctx := suite.T().Context()

	status := suite.testStatuses["admin_account_status_1"]
	status.Content += " fnord"
	status.Text += " fnord"

	if boost {
		// Modify a fixture boost into a boost of the above status.
		boost := suite.testStatuses["admin_account_status_4"]
		boost.BoostOf = status
		boost.BoostOfID = status.ID
		status = boost
	}

	requester := suite.testAccounts["local_account_1"]

	filter := suite.testFilters["local_account_1_filter_1"]
	filter.Action = action

	err := suite.state.DB.UpdateFilter(ctx, filter, "action")
	suite.NoError(err)

	return suite.filter.StatusFilterResultsInContext(ctx,
		requester,
		status,
		gtsmodel.FilterContextHome,
	)
}

func (suite *StatusFilterTestSuite) testFilteredStatusWithHashtag(wholeword, boost bool) {
	ctx := suite.T().Context()

	status := new(gtsmodel.Status)
	*status = *suite.testStatuses["admin_account_status_1"]
	status.Content = `<p>doggo doggin' it</p><p><a href="https://example.test/tags/dogsofmastodon" class="mention hashtag" rel="tag nofollow noreferrer noopener" target="_blank">#<span>dogsofmastodon</span></a></p>`

	if boost {
		boost, err := suite.converter.StatusToBoost(
			suite.T().Context(),
			status,
			suite.testAccounts["admin_account"],
			"",
		)
		suite.NoError(err)
		status = boost
	}

	var err error

	requester := suite.testAccounts["local_account_1"]

	filter := &gtsmodel.Filter{
		ID:        id.NewULID(),
		Title:     id.NewULID(),
		AccountID: requester.ID,
		Action:    gtsmodel.FilterActionWarn,
		Contexts:  gtsmodel.FilterContexts(gtsmodel.FilterContextHome),
	}

	filterKeyword := &gtsmodel.FilterKeyword{
		ID:        id.NewULID(),
		FilterID:  filter.ID,
		Keyword:   "#dogsofmastodon",
		WholeWord: &wholeword,
	}

	filter.KeywordIDs = []string{filterKeyword.ID}

	err = suite.state.DB.PutFilterKeyword(ctx, filterKeyword)
	suite.NoError(err)

	err = suite.state.DB.PutFilter(ctx, filter)
	suite.NoError(err)

	filtered, hide, err := suite.filter.StatusFilterResultsInContext(ctx,
		requester,
		status,
		gtsmodel.FilterContextHome,
	)
	suite.NoError(err)
	suite.False(hide)
	suite.NotEmpty(filtered)
}

func TestStatusFilterTestSuite(t *testing.T) {
	suite.Run(t, new(StatusFilterTestSuite))
}
