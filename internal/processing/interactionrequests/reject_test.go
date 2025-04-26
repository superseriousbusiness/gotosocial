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

package interactionrequests_test

import (
	"context"
	"errors"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/processing/interactionrequests"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type RejectTestSuite struct {
	InteractionRequestsTestSuite
}

func (suite *RejectTestSuite) TestReject() {
	testStructs := testrig.SetupTestStructs(rMediaPath, rTemplatePath)
	defer testrig.TearDownTestStructs(testStructs)

	var (
		ctx    = context.Background()
		state  = testStructs.State
		acct   = suite.testAccounts["local_account_2"]
		intReq = suite.testInteractionRequests["admin_account_reply_turtle"]
	)

	// Create int reqs processor.
	p := interactionrequests.New(
		testStructs.Common,
		testStructs.State,
		testStructs.TypeConverter,
	)

	apiReq, errWithCode := p.Reject(ctx, acct, intReq.ID)
	if errWithCode != nil {
		suite.FailNow(errWithCode.Error())
	}

	// Get db interaction rejection.
	dbReq, err := state.DB.GetInteractionRequestByID(ctx, apiReq.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.True(dbReq.IsRejected())

	// Wait for interacting status to be deleted.
	testrig.WaitFor(func() bool {
		status, err := state.DB.GetStatusByURI(
			gtscontext.SetBarebones(ctx),
			dbReq.InteractionURI,
		)
		return status == nil && errors.Is(err, db.ErrNoEntries)
	})

	// Wait for a copy of the status
	// to be hurled into the sin bin.
	testrig.WaitFor(func() bool {
		sbStatus, err := state.DB.GetSinBinStatusByURI(
			gtscontext.SetBarebones(ctx),
			dbReq.InteractionURI,
		)
		return err == nil && sbStatus != nil
	})
}

func TestRejectTestSuite(t *testing.T) {
	suite.Run(t, new(RejectTestSuite))
}
