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
	"context"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type ScheduledStatusTestSuite struct {
	StatusStandardTestSuite
}

func (suite *ScheduledStatusTestSuite) TestUpdate() {
	ctx := suite.T().Context()

	account1 := suite.testAccounts["local_account_1"]
	scheduledStatus1 := suite.testScheduledStatuses["scheduled_status_1"]
	newScheduledAt := testrig.TimeMustParse("2080-07-02T21:37:00+02:00")

	suite.state.Workers.Scheduler.AddOnce(scheduledStatus1.ID, scheduledStatus1.ScheduledAt, func(ctx context.Context, t time.Time) {})

	// update scheduled status publication date
	scheduledStatus2, err := suite.status.ScheduledStatusesUpdate(ctx, account1, scheduledStatus1.ID, util.Ptr(newScheduledAt))
	suite.NoError(err)
	suite.NotNil(scheduledStatus2)
	suite.Equal(scheduledStatus2.ScheduledAt, util.FormatISO8601(newScheduledAt))
	// should be rescheduled
	suite.Equal(suite.state.Workers.Scheduler.Cancel(scheduledStatus1.ID), true)
}

func (suite *ScheduledStatusTestSuite) TestDelete() {
	ctx := suite.T().Context()

	account1 := suite.testAccounts["local_account_1"]
	scheduledStatus1 := suite.testScheduledStatuses["scheduled_status_1"]

	suite.state.Workers.Scheduler.AddOnce(scheduledStatus1.ID, scheduledStatus1.ScheduledAt, func(ctx context.Context, t time.Time) {})

	// delete scheduled status
	err := suite.status.ScheduledStatusesDelete(ctx, account1, scheduledStatus1.ID)
	suite.NoError(err)
	// should be already cancelled
	suite.Equal(suite.state.Workers.Scheduler.Cancel(scheduledStatus1.ID), false)
}

func TestScheduledStatusTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduledStatusTestSuite))
}
