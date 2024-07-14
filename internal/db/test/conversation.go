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

package test

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/oklog/ulid"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type testSuite interface {
	FailNow(string, ...interface{}) bool
}

// ConversationFactory can be embedded or included by test suites that want to generate statuses and conversations.
type ConversationFactory struct {
	// Test suite, or at least the methods from it that we care about.
	suite testSuite
	// Test DB.
	db db.DB

	// TestStart is the timestamp used as a base for timestamps and ULIDs in any given test.
	TestStart time.Time
}

// SetupSuite should be called by the SetupSuite of test suites that use this mixin.
func (f *ConversationFactory) SetupSuite(suite testSuite) {
	f.suite = suite
}

// SetupTest should be called by the SetupTest of test suites that use this mixin.
func (f *ConversationFactory) SetupTest(db db.DB) {
	f.db = db
	f.TestStart = time.Now()
}

// NewULID is a version of id.NewULID that uses the test start time and an offset instead of the real time.
func (f *ConversationFactory) NewULID(offset time.Duration) string {
	ulid, err := ulid.New(
		ulid.Timestamp(f.TestStart.Add(offset)), rand.Reader,
	)
	if err != nil {
		panic(err)
	}
	return ulid.String()
}

func (f *ConversationFactory) NewTestStatus(localAccount *gtsmodel.Account, threadID string, nowOffset time.Duration, inReplyToStatus *gtsmodel.Status) *gtsmodel.Status {
	statusID := f.NewULID(nowOffset)
	createdAt := f.TestStart.Add(nowOffset)
	status := &gtsmodel.Status{
		ID:                  statusID,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		URI:                 "http://localhost:8080/users/" + localAccount.Username + "/statuses/" + statusID,
		AccountID:           localAccount.ID,
		AccountURI:          localAccount.URI,
		Local:               util.Ptr(true),
		ThreadID:            threadID,
		Visibility:          gtsmodel.VisibilityDirect,
		ActivityStreamsType: ap.ObjectNote,
		Federated:           util.Ptr(true),
	}
	if inReplyToStatus != nil {
		status.InReplyToID = inReplyToStatus.ID
		status.InReplyToURI = inReplyToStatus.URI
		status.InReplyToAccountID = inReplyToStatus.AccountID
	}
	if err := f.db.PutStatus(context.Background(), status); err != nil {
		f.suite.FailNow(err.Error())
	}
	return status
}

// NewTestConversation creates a new status and adds it to a new unread conversation, returning the conversation.
func (f *ConversationFactory) NewTestConversation(localAccount *gtsmodel.Account, nowOffset time.Duration) *gtsmodel.Conversation {
	threadID := f.NewULID(nowOffset)
	status := f.NewTestStatus(localAccount, threadID, nowOffset, nil)
	conversation := &gtsmodel.Conversation{
		ID:        f.NewULID(nowOffset),
		AccountID: localAccount.ID,
		ThreadID:  status.ThreadID,
		Read:      util.Ptr(false),
	}
	f.SetLastStatus(conversation, status)
	return conversation
}

// SetLastStatus sets an already stored status as the last status of a new or already stored conversation,
// and returns the updated conversation.
func (f *ConversationFactory) SetLastStatus(conversation *gtsmodel.Conversation, status *gtsmodel.Status) *gtsmodel.Conversation {
	conversation.LastStatusID = status.ID
	conversation.LastStatus = status
	if err := f.db.UpsertConversation(context.Background(), conversation, "last_status_id"); err != nil {
		f.suite.FailNow(err.Error())
	}
	if err := f.db.LinkConversationToStatus(context.Background(), conversation.ID, status.ID); err != nil {
		f.suite.FailNow(err.Error())
	}
	return conversation
}
