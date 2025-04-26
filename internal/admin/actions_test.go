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

package admin_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

const (
	rMediaPath    = "../../testrig/media"
	rTemplatePath = "../../web/template"
)

type ActionsTestSuite struct {
	suite.Suite
}

func (suite *ActionsTestSuite) SetupSuite() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
}

func (suite *ActionsTestSuite) TestActionOverlap() {
	var (
		testStructs = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		ctx         = context.Background()
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Suspend account.
	action1 := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryAccount,
		TargetID:       "01H90S1CXQ97J9625C5YBXZWGT",
		Type:           gtsmodel.AdminActionSuspend,
		AccountID:      "01H90S1ZZXP4N74H4A9RVW1MRP",
	}
	key1 := action1.Key()
	suite.Equal("account/01H90S1CXQ97J9625C5YBXZWGT", key1)

	// Unsuspend account.
	action2 := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryAccount,
		TargetID:       "01H90S1CXQ97J9625C5YBXZWGT",
		Type:           gtsmodel.AdminActionUnsuspend,
		AccountID:      "01H90S1ZZXP4N74H4A9RVW1MRP",
	}
	key2 := action2.Key()
	suite.Equal("account/01H90S1CXQ97J9625C5YBXZWGT", key2)

	errWithCode := testStructs.State.AdminActions.Run(
		ctx,
		action1,
		func(ctx context.Context) gtserror.MultiError {
			// Noop, just sleep (mood).
			time.Sleep(3 * time.Second)
			return nil
		},
	)
	suite.NoError(errWithCode)

	// While first action is sleeping, try to
	// process another with the same key.
	errWithCode = testStructs.State.AdminActions.Run(
		ctx,
		action2,
		func(ctx context.Context) gtserror.MultiError {
			return nil
		},
	)
	if errWithCode == nil {
		suite.FailNow("expected error with code, but error was nil")
	}

	// Code should be 409.
	suite.Equal(http.StatusConflict, errWithCode.Code())

	// Wait for action to finish.
	if !testrig.WaitFor(func() bool {
		return testStructs.State.AdminActions.TotalRunning() == 0
	}) {
		suite.FailNow("timed out waiting for admin action(s) to finish")
	}

	// Try again.
	errWithCode = testStructs.State.AdminActions.Run(
		ctx,
		action2,
		func(ctx context.Context) gtserror.MultiError {
			return nil
		},
	)
	suite.NoError(errWithCode)

	// Wait for action to finish.
	if !testrig.WaitFor(func() bool {
		return testStructs.State.AdminActions.TotalRunning() == 0
	}) {
		suite.FailNow("timed out waiting for admin action(s) to finish")
	}
}

func (suite *ActionsTestSuite) TestActionWithErrors() {
	var (
		testStructs = testrig.SetupTestStructs(rMediaPath, rTemplatePath)
		ctx         = context.Background()
	)
	defer testrig.TearDownTestStructs(testStructs)

	// Suspend a domain.
	action := &gtsmodel.AdminAction{
		ID:             id.NewULID(),
		TargetCategory: gtsmodel.AdminActionCategoryDomain,
		TargetID:       "example.org",
		Type:           gtsmodel.AdminActionSuspend,
		AccountID:      "01H90S1ZZXP4N74H4A9RVW1MRP",
	}

	errWithCode := testStructs.State.AdminActions.Run(
		ctx,
		action,
		func(ctx context.Context) gtserror.MultiError {
			// Noop, just return some errs.
			return gtserror.MultiError{
				db.ErrNoEntries,
				errors.New("fucky wucky"),
			}
		},
	)
	suite.NoError(errWithCode)

	// Wait for action to finish.
	if !testrig.WaitFor(func() bool {
		return testStructs.State.AdminActions.TotalRunning() == 0
	}) {
		suite.FailNow("timed out waiting for admin action(s) to finish")
	}

	// Get action from the db.
	dbAction, err := testStructs.State.DB.GetAdminAction(ctx, action.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.EqualValues([]string{
		"sql: no rows in result set",
		"fucky wucky",
	}, dbAction.Errors)
}

func TestActionsTestSuite(t *testing.T) {
	suite.Run(t, new(ActionsTestSuite))
}
