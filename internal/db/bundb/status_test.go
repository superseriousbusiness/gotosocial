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
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type StatusTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *StatusTestSuite) TestGetStatusByID() {
	status, err := suite.db.GetStatusByID(context.Background(), suite.testStatuses["local_account_1_status_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.Nil(status.BoostOf)
	suite.Nil(status.BoostOfAccount)
	suite.Nil(status.InReplyTo)
	suite.Nil(status.InReplyToAccount)
	suite.True(*status.Federated)
}

func (suite *StatusTestSuite) TestGetStatusesByIDs() {
	ids := []string{
		suite.testStatuses["local_account_1_status_1"].ID,
		suite.testStatuses["local_account_2_status_3"].ID,
	}

	statuses, err := suite.db.GetStatusesByIDs(context.Background(), ids)
	if err != nil {
		suite.FailNow(err.Error())
	}

	if len(statuses) != 2 {
		suite.FailNow("expected 2 statuses in slice")
	}

	status1 := statuses[0]
	suite.NotNil(status1)
	suite.NotNil(status1.Account)
	suite.NotNil(status1.CreatedWithApplication)
	suite.Nil(status1.BoostOf)
	suite.Nil(status1.BoostOfAccount)
	suite.Nil(status1.InReplyTo)
	suite.Nil(status1.InReplyToAccount)
	suite.True(*status1.Federated)

	status2 := statuses[1]
	suite.NotNil(status2)
	suite.NotNil(status2.Account)
	suite.NotNil(status2.CreatedWithApplication)
	suite.Nil(status2.BoostOf)
	suite.Nil(status2.BoostOfAccount)
	suite.Nil(status2.InReplyTo)
	suite.Nil(status2.InReplyToAccount)
	suite.True(*status2.Federated)
}

func (suite *StatusTestSuite) TestGetStatusByURI() {
	status, err := suite.db.GetStatusByURI(context.Background(), suite.testStatuses["local_account_2_status_3"].URI)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.Nil(status.BoostOf)
	suite.Nil(status.BoostOfAccount)
	suite.Nil(status.InReplyTo)
	suite.Nil(status.InReplyToAccount)
	suite.True(*status.Federated)
}

func (suite *StatusTestSuite) TestGetStatusWithExtras() {
	status, err := suite.db.GetStatusByID(context.Background(), suite.testStatuses["admin_account_status_1"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.NotEmpty(status.Tags)
	suite.NotEmpty(status.Attachments)
	suite.NotEmpty(status.Emojis)
	suite.True(*status.Federated)
}

func (suite *StatusTestSuite) TestGetStatusWithMention() {
	status, err := suite.db.GetStatusByID(context.Background(), suite.testStatuses["local_account_2_status_5"].ID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotNil(status)
	suite.NotNil(status.Account)
	suite.NotNil(status.CreatedWithApplication)
	suite.NotEmpty(status.MentionIDs)
	suite.NotEmpty(status.InReplyToID)
	suite.NotEmpty(status.InReplyToAccountID)
	suite.True(*status.Federated)
}

// The below test was originally used to ensure that a second
// fetch is faster than a first fetch of a status, because in
// the second fetch it should be cached. However because we
// always run in-memory tests anyway, sometimes the first fetch
// is actually faster, which causes CI/CD to fail unpredictably.
// Since we know by now (Feb 2024) that the cache works fine,
// the test is commented out.
/*
func (suite *StatusTestSuite) TestGetStatusTwice() {
	before1 := time.Now()
	_, err := suite.db.GetStatusByURI(context.Background(), suite.testStatuses["local_account_1_status_1"].URI)
	suite.NoError(err)
	after1 := time.Now()
	duration1 := after1.Sub(before1)
	fmt.Println(duration1.Microseconds())

	before2 := time.Now()
	_, err = suite.db.GetStatusByURI(context.Background(), suite.testStatuses["local_account_1_status_1"].URI)
	suite.NoError(err)
	after2 := time.Now()
	duration2 := after2.Sub(before2)
	fmt.Println(duration2.Microseconds())

	// second retrieval should be several orders faster since it will be cached now
	suite.Less(duration2, duration1)
}
*/

func (suite *StatusTestSuite) TestGetStatusReplies() {
	targetStatus := suite.testStatuses["local_account_1_status_1"]
	children, err := suite.db.GetStatusReplies(context.Background(), targetStatus.ID)
	suite.NoError(err)
	suite.Len(children, 2)
	for _, c := range children {
		suite.Equal(targetStatus.URI, c.InReplyToURI)
		suite.Equal(targetStatus.AccountID, c.InReplyToAccountID)
		suite.Equal(targetStatus.ID, c.InReplyToID)
	}
}

func (suite *StatusTestSuite) TestGetStatusChildren() {
	targetStatus := suite.testStatuses["local_account_1_status_1"]
	children, err := suite.db.GetStatusChildren(context.Background(), targetStatus.ID)
	suite.NoError(err)
	suite.Len(children, 3)
}

func (suite *StatusTestSuite) TestDeleteStatus() {
	// Take a copy of the status.
	targetStatus := &gtsmodel.Status{}
	*targetStatus = *suite.testStatuses["admin_account_status_1"]

	err := suite.db.DeleteStatusByID(context.Background(), targetStatus.ID)
	suite.NoError(err)

	_, err = suite.db.GetStatusByID(context.Background(), targetStatus.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

// This test was added specifically to ensure that Postgres wasn't getting upset
// about trying to use a transaction in which an error has already occurred, which
// was previously leading to errors like 'current transaction is aborted, commands
// ignored until end of transaction block' when updating a status that already had
// emojis or tags set on it.
//
// To run this test for postgres specifically, start a postgres container on localhost
// and then run:
//
// GTS_DB_TYPE=postgres GTS_DB_ADDRESS=localhost go test ./internal/db/bundb -run '^TestStatusTestSuite$' -testify.m '^(TestUpdateStatus)$' github.com/superseriousbusiness/gotosocial/internal/db/bundb
func (suite *StatusTestSuite) TestUpdateStatus() {
	// Take a copy of the status.
	targetStatus := &gtsmodel.Status{}
	*targetStatus = *suite.testStatuses["admin_account_status_1"]

	targetStatus.PinnedAt = time.Time{}

	err := suite.db.UpdateStatus(context.Background(), targetStatus, "pinned_at")
	suite.NoError(err)

	updated, err := suite.db.GetStatusByID(context.Background(), targetStatus.ID)
	suite.NoError(err)
	suite.True(updated.PinnedAt.IsZero())
}

func (suite *StatusTestSuite) TestPutPopulatedStatus() {
	ctx := context.Background()

	targetStatus := &gtsmodel.Status{}
	*targetStatus = *suite.testStatuses["admin_account_status_1"]

	// Populate fields on the target status.
	if err := suite.db.PopulateStatus(ctx, targetStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// Delete it from the database.
	if err := suite.db.DeleteStatusByID(ctx, targetStatus.ID); err != nil {
		suite.FailNow(err.Error())
	}

	// Reinsert the populated version
	// so that it becomes cached.
	if err := suite.db.PutStatus(ctx, targetStatus); err != nil {
		suite.FailNow(err.Error())
	}

	// Update the status owner's
	// account with a new bio.
	account := &gtsmodel.Account{}
	*account = *targetStatus.Account
	account.Note = "new note for this test"
	if err := suite.db.UpdateAccount(ctx, account, "note"); err != nil {
		suite.FailNow(err.Error())
	}

	dbStatus, err := suite.db.GetStatusByID(ctx, targetStatus.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// Account note should be updated,
	// even though we stored this
	// status with the old note.
	suite.Equal(
		"new note for this test",
		dbStatus.Account.Note,
	)
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
