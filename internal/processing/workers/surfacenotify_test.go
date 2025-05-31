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

package workers_test

import (
	"sync"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/filter/mutes"
	"code.superseriousbusiness.org/gotosocial/internal/filter/visibility"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/processing/workers"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type SurfaceNotifyTestSuite struct {
	WorkersTestSuite
}

func (suite *SurfaceNotifyTestSuite) TestSpamNotifs() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	surface := &workers.Surface{
		State:         testStructs.State,
		Converter:     testStructs.TypeConverter,
		Stream:        testStructs.Processor.Stream(),
		VisFilter:     visibility.NewFilter(testStructs.State),
		MuteFilter:    mutes.NewFilter(testStructs.State),
		EmailSender:   testStructs.EmailSender,
		WebPushSender: testStructs.WebPushSender,
		Conversations: testStructs.Processor.Conversations(),
	}

	var (
		ctx              = suite.T().Context()
		notificationType = gtsmodel.NotificationFollow
		targetAccount    = suite.testAccounts["local_account_1"]
		originAccount    = suite.testAccounts["local_account_2"]
	)

	// Set up a bunch of goroutines to surface
	// a notification at exactly the same time.
	wg := sync.WaitGroup{}
	wg.Add(20)
	startAt := time.Now().Add(2 * time.Second)

	for i := 0; i < 20; i++ {
		go func() {
			defer wg.Done()

			// Wait for it....
			untilTick := time.Until(startAt)
			<-time.Tick(untilTick)

			// ...Go!
			if err := surface.Notify(ctx,
				notificationType,
				targetAccount,
				originAccount,
				nil,
				nil,
			); err != nil {
				suite.FailNow(err.Error())
			}
		}()
	}

	// Wait for all notif creation
	// attempts to complete.
	wg.Wait()

	// Get all notifs for this account.
	notifs, err := testStructs.State.DB.GetAccountNotifications(
		gtscontext.SetBarebones(ctx),
		targetAccount.ID,
		nil, nil, nil,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	var gotOne bool
	for _, notif := range notifs {
		if notif.NotificationType == notificationType &&
			notif.TargetAccountID == targetAccount.ID &&
			notif.OriginAccountID == originAccount.ID {
			// This is the notif...
			if gotOne {
				// We already had
				// the notif, d'oh!
				suite.FailNow("already had notif")
			} else {
				gotOne = true
			}
		}
	}
}

func TestSurfaceNotifyTestSuite(t *testing.T) {
	suite.Run(t, new(SurfaceNotifyTestSuite))
}
