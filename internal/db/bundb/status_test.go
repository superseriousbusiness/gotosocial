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
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type StatusTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *StatusTestSuite) TestGetStatusByID() {
	status, err := suite.db.GetStatusByID(suite.T().Context(), suite.testStatuses["local_account_1_status_1"].ID)
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

	statuses, err := suite.db.GetStatusesByIDs(suite.T().Context(), ids)
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
	status, err := suite.db.GetStatusByURI(suite.T().Context(), suite.testStatuses["local_account_2_status_3"].URI)
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
	status, err := suite.db.GetStatusByID(suite.T().Context(), suite.testStatuses["admin_account_status_1"].ID)
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
	status, err := suite.db.GetStatusByID(suite.T().Context(), suite.testStatuses["local_account_2_status_5"].ID)
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
	_, err := suite.db.GetStatusByURI(suite.T().Context(), suite.testStatuses["local_account_1_status_1"].URI)
	suite.NoError(err)
	after1 := time.Now()
	duration1 := after1.Sub(before1)
	fmt.Println(duration1.Microseconds())

	before2 := time.Now()
	_, err = suite.db.GetStatusByURI(suite.T().Context(), suite.testStatuses["local_account_1_status_1"].URI)
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
	children, err := suite.db.GetStatusReplies(suite.T().Context(), targetStatus.ID)
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
	children, err := suite.db.GetStatusChildren(suite.T().Context(), targetStatus.ID)
	suite.NoError(err)
	suite.Len(children, 3)
}

func (suite *StatusTestSuite) TestDeleteStatus() {
	// Take a copy of the status.
	targetStatus := &gtsmodel.Status{}
	*targetStatus = *suite.testStatuses["admin_account_status_1"]

	err := suite.db.DeleteStatusByID(suite.T().Context(), targetStatus.ID)
	suite.NoError(err)

	_, err = suite.db.GetStatusByID(suite.T().Context(), targetStatus.ID)
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
// GTS_DB_TYPE=postgres GTS_DB_ADDRESS=localhost go test ./internal/db/bundb -run '^TestStatusTestSuite$' -testify.m '^(TestUpdateStatus)$' code.superseriousbusiness.org/gotosocial/internal/db/bundb
func (suite *StatusTestSuite) TestUpdateStatus() {
	// Take a copy of the status.
	targetStatus := &gtsmodel.Status{}
	*targetStatus = *suite.testStatuses["admin_account_status_1"]

	targetStatus.PinnedAt = time.Time{}

	err := suite.db.UpdateStatus(suite.T().Context(), targetStatus, "pinned_at")
	suite.NoError(err)

	updated, err := suite.db.GetStatusByID(suite.T().Context(), targetStatus.ID)
	suite.NoError(err)
	suite.True(updated.PinnedAt.IsZero())
}

func (suite *StatusTestSuite) TestPutPopulatedStatus() {
	ctx := suite.T().Context()

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

func (suite *StatusTestSuite) TestPutStatusThreadingBoostOfIDSet() {
	ctx := suite.T().Context()

	// Fake account details.
	accountID := id.NewULID()
	accountURI := "https://example.com/users/" + accountID

	var err error

	// Prepare new status.
	statusID := id.NewULID()
	statusURI := accountURI + "/statuses/" + statusID
	status := &gtsmodel.Status{
		ID:                  statusID,
		URI:                 statusURI,
		AccountID:           accountID,
		AccountURI:          accountURI,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	// Insert original status into database.
	err = suite.db.PutStatus(ctx, status)
	suite.NoError(err)
	suite.NotEmpty(status.ThreadID)

	// Prepare new boost.
	boostID := id.NewULID()
	boostURI := accountURI + "/statuses/" + boostID
	boost := &gtsmodel.Status{
		ID:                  boostID,
		URI:                 boostURI,
		AccountID:           accountID,
		AccountURI:          accountURI,
		BoostOfID:           statusID,
		BoostOfAccountID:    accountID,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	// Insert boost wrapper into database.
	err = suite.db.PutStatus(ctx, boost)
	suite.NoError(err)

	// Boost wrapper should have inherited thread.
	suite.Equal(status.ThreadID, boost.ThreadID)
}

func (suite *StatusTestSuite) TestPutStatusThreadingInReplyToIDSet() {
	ctx := suite.T().Context()

	// Fake account details.
	accountID := id.NewULID()
	accountURI := "https://example.com/users/" + accountID

	var err error

	// Prepare new status.
	statusID := id.NewULID()
	statusURI := accountURI + "/statuses/" + statusID
	status := &gtsmodel.Status{
		ID:                  statusID,
		URI:                 statusURI,
		AccountID:           accountID,
		AccountURI:          accountURI,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	// Insert original status into database.
	err = suite.db.PutStatus(ctx, status)
	suite.NoError(err)
	suite.NotEmpty(status.ThreadID)

	// Prepare new reply.
	replyID := id.NewULID()
	replyURI := accountURI + "/statuses/" + replyID
	reply := &gtsmodel.Status{
		ID:                  replyID,
		URI:                 replyURI,
		AccountID:           accountID,
		AccountURI:          accountURI,
		InReplyToID:         statusID,
		InReplyToURI:        statusURI,
		InReplyToAccountID:  accountID,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	// Insert status reply into database.
	err = suite.db.PutStatus(ctx, reply)
	suite.NoError(err)

	// Status reply should have inherited thread.
	suite.Equal(status.ThreadID, reply.ThreadID)
}

func (suite *StatusTestSuite) TestPutStatusThreadingSiblings() {
	ctx := suite.T().Context()

	// Fake account details.
	accountID := id.NewULID()
	accountURI := "https://example.com/users/" + accountID

	// Main parent status ID.
	statusID := id.NewULID()
	statusURI := accountURI + "/statuses/" + statusID
	status := &gtsmodel.Status{
		ID:                  statusID,
		URI:                 statusURI,
		AccountID:           accountID,
		AccountURI:          accountURI,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	const siblingCount = 10
	var statuses []*gtsmodel.Status
	for range siblingCount {
		id := id.NewULID()
		uri := accountURI + "/statuses/" + id

		// Note here that inReplyToID not being set,
		// so as they get inserted it's as if children
		// are being dereferenced ahead of stored parent.
		//
		// Which is where out-of-sync threads can occur.
		statuses = append(statuses, &gtsmodel.Status{
			ID:                  id,
			URI:                 uri,
			AccountID:           accountID,
			AccountURI:          accountURI,
			InReplyToURI:        statusURI,
			Local:               util.Ptr(false),
			Federated:           util.Ptr(true),
			ActivityStreamsType: ap.ObjectNote,
		})
	}

	var err error
	var threadID string

	// Insert all of the sibling children
	// into the database, they should all
	// still get correctly threaded together.
	for _, child := range statuses {
		err = suite.db.PutStatus(ctx, child)
		suite.NoError(err)
		suite.NotEmpty(child.ThreadID)
		if threadID == "" {
			threadID = child.ThreadID
		} else {
			suite.Equal(threadID, child.ThreadID)
		}
	}

	// Finally, insert the parent status.
	err = suite.db.PutStatus(ctx, status)
	suite.NoError(err)

	// Parent should have inherited thread.
	suite.Equal(threadID, status.ThreadID)
}

func (suite *StatusTestSuite) TestPutStatusThreadingReconcile() {
	ctx := suite.T().Context()

	// Fake account details.
	accountID := id.NewULID()
	accountURI := "https://example.com/users/" + accountID

	const threadLength = 10
	var statuses []*gtsmodel.Status
	var lastURI, lastID string

	// Generate front-half of thread.
	for range threadLength / 2 {
		id := id.NewULID()
		uri := accountURI + "/statuses/" + id
		statuses = append(statuses, &gtsmodel.Status{
			ID:                  id,
			URI:                 uri,
			AccountID:           accountID,
			AccountURI:          accountURI,
			InReplyToID:         lastID,
			InReplyToURI:        lastURI,
			Local:               util.Ptr(false),
			Federated:           util.Ptr(true),
			ActivityStreamsType: ap.ObjectNote,
		})
		lastURI = uri
		lastID = id
	}

	// Generate back-half of thread.
	//
	// Note here that inReplyToID not being set past
	// the first item, so as they get inserted it's
	// as if the children are dereferenced ahead of
	// the stored parent, i.e. an out-of-sync thread.
	for range threadLength / 2 {
		id := id.NewULID()
		uri := accountURI + "/statuses/" + id
		statuses = append(statuses, &gtsmodel.Status{
			ID:                  id,
			URI:                 uri,
			AccountID:           accountID,
			AccountURI:          accountURI,
			InReplyToID:         lastID,
			InReplyToURI:        lastURI,
			Local:               util.Ptr(false),
			Federated:           util.Ptr(true),
			ActivityStreamsType: ap.ObjectNote,
		})
		lastURI = uri
		lastID = ""
	}

	var err error

	// Thread IDs we expect to see for
	// head statuses as we add them, and
	// for tail statuses as we add them.
	var thread0, threadN string

	// Insert status thread from head and tail,
	// specifically stopping before the middle.
	// These should each get threaded separately.
	for i := range (threadLength / 2) - 1 {
		i0, iN := i, len(statuses)-1-i

		// Insert i'th status from the start.
		err = suite.db.PutStatus(ctx, statuses[i0])
		suite.NoError(err)
		suite.NotEmpty(statuses[i0].ThreadID)

		// Check i0 thread.
		if thread0 == "" {
			thread0 = statuses[i0].ThreadID
		} else {
			suite.Equal(thread0, statuses[i0].ThreadID)
		}

		// Insert i'th status from the end.
		err = suite.db.PutStatus(ctx, statuses[iN])
		suite.NoError(err)
		suite.NotEmpty(statuses[iN].ThreadID)

		// Check iN thread.
		if threadN == "" {
			threadN = statuses[iN].ThreadID
		} else {
			suite.Equal(threadN, statuses[iN].ThreadID)
		}
	}

	// Finally, insert remaining statuses,
	// at some point among these it should
	// trigger a status thread reconcile.
	for _, status := range statuses {

		if status.ThreadID != "" {
			// already inserted
			continue
		}

		// Insert remaining status into db.
		err = suite.db.PutStatus(ctx, status)
		suite.NoError(err)
	}

	// The reconcile should pick the older,
	// i.e. smaller of two ULID thread IDs.
	finalThreadID := min(thread0, threadN)
	for _, status := range statuses {

		// Get ID of status.
		id := status.ID

		// Fetch latest status the from database.
		status, err := suite.db.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			id,
		)
		suite.NoError(err)

		// Ensure after reconcile uses expected thread.
		suite.Equal(finalThreadID, status.ThreadID)
	}
}

func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
