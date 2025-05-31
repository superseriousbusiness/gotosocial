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

package mutes_test

import (
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/stretchr/testify/suite"
)

type StatusMuteTestSuite struct {
	FilterStandardTestSuite
}

func (suite *StatusMuteTestSuite) TestMutedStatusAuthor() {
	ctx := suite.T().Context()

	status := suite.testStatuses["admin_account_status_1"]
	requester := suite.testAccounts["local_account_1"]
	replyer := suite.testAccounts["local_account_2"]

	// Generate a new reply
	// to the above status.
	replyID := id.NewULID()
	reply := &gtsmodel.Status{
		ID:                  replyID,
		URI:                 replyer.URI + "/statuses/" + replyID,
		ThreadID:            status.ThreadID,
		AccountID:           replyer.ID,
		AccountURI:          replyer.URI,
		InReplyToID:         status.ID,
		InReplyToURI:        status.URI,
		InReplyToAccountID:  status.AccountID,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	// And insert reply into the database.
	err := suite.db.PutStatus(ctx, reply)
	suite.NoError(err)

	// Ensure that neither status nor reply are muted to requester.
	muted1, err1 := suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 := suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Ensure notifications for neither status nor reply are muted to requester.
	muted1, err = suite.filter.StatusNotificationsMuted(ctx, requester, status)
	muted2, err = suite.filter.StatusNotificationsMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Insert new user mute targetting first status author.
	err = suite.state.DB.PutMute(ctx, &gtsmodel.UserMute{
		ID:              id.NewULID(),
		AccountID:       requester.ID,
		TargetAccountID: status.AccountID,
		Notifications:   util.Ptr(false),
	})
	suite.NoError(err)

	// Now ensure that both status and reply are muted to requester.
	muted1, err1 = suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 = suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.True(muted1)
	suite.True(muted2)

	// Though neither status nor reply should have notifications muted to requester.
	muted1, err = suite.filter.StatusNotificationsMuted(ctx, requester, status)
	muted2, err = suite.filter.StatusNotificationsMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Now delete account mutes to / from requesting account.
	err = suite.state.DB.DeleteAccountMutes(ctx, requester.ID)
	suite.NoError(err)

	// Now ensure that both status and reply are unmuted again.
	muted1, err1 = suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 = suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)
}

func (suite *StatusMuteTestSuite) TestMutedStatusMentionee() {
	ctx := suite.T().Context()

	status := suite.testStatuses["admin_account_status_5"]
	requester := suite.testAccounts["local_account_1"]
	mentionee := suite.testAccounts["local_account_2"]
	replyer := suite.testAccounts["local_account_3"]

	// Generate a new reply
	// to the above status.
	replyID := id.NewULID()
	reply := &gtsmodel.Status{
		ID:                  replyID,
		URI:                 replyer.URI + "/statuses/" + replyID,
		ThreadID:            status.ThreadID,
		AccountID:           replyer.ID,
		AccountURI:          replyer.URI,
		InReplyToID:         status.ID,
		InReplyToURI:        status.URI,
		InReplyToAccountID:  status.AccountID,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	// And insert reply into the database.
	err := suite.db.PutStatus(ctx, reply)
	suite.NoError(err)

	// Ensure that neither status nor reply are muted to requester.
	muted1, err1 := suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 := suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Ensure notifications for neither status nor reply are muted to requester.
	muted1, err = suite.filter.StatusNotificationsMuted(ctx, requester, status)
	muted2, err = suite.filter.StatusNotificationsMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Insert user visibility mute targetting status author.
	err = suite.state.DB.PutMute(ctx, &gtsmodel.UserMute{
		ID:              id.NewULID(),
		AccountID:       requester.ID,
		TargetAccountID: mentionee.ID,
		Notifications:   util.Ptr(false),
	})
	suite.NoError(err)

	// Now ensure that both status and reply are muted to requester.
	muted1, err1 = suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 = suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.True(muted1)
	suite.True(muted2)

	// Though neither status nor reply should have notifications muted to requester.
	muted1, err = suite.filter.StatusNotificationsMuted(ctx, requester, status)
	muted2, err = suite.filter.StatusNotificationsMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Now delete account mutes to / from requesting account.
	err = suite.state.DB.DeleteAccountMutes(ctx, requester.ID)
	suite.NoError(err)

	// Now ensure that both status and reply are unmuted again.
	muted1, err1 = suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 = suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)
}

func (suite *StatusMuteTestSuite) TestMutedStatusThread() {
	ctx := suite.T().Context()

	status := suite.testStatuses["admin_account_status_1"]
	requester := suite.testAccounts["local_account_1"]
	replyer := suite.testAccounts["local_account_2"]

	// Generate a new reply
	// to the above status.
	replyID := id.NewULID()
	reply := &gtsmodel.Status{
		ID:                  replyID,
		URI:                 replyer.URI + "/statuses/" + replyID,
		ThreadID:            status.ThreadID,
		AccountID:           replyer.ID,
		AccountURI:          replyer.URI,
		InReplyToID:         status.ID,
		InReplyToURI:        status.URI,
		InReplyToAccountID:  status.AccountID,
		Local:               util.Ptr(false),
		Federated:           util.Ptr(true),
		ActivityStreamsType: ap.ObjectNote,
	}

	// And insert reply into the database.
	err := suite.db.PutStatus(ctx, reply)
	suite.NoError(err)

	// Ensure that neither status nor reply are muted to requester.
	muted1, err1 := suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 := suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Ensure notifications for neither status nor reply are muted to requester.
	muted1, err = suite.filter.StatusNotificationsMuted(ctx, requester, status)
	muted2, err = suite.filter.StatusNotificationsMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	threadMuteID := id.NewULID()

	// Insert new notification mute targetting status thread.
	err = suite.db.PutThreadMute(ctx, &gtsmodel.ThreadMute{
		ID:        threadMuteID,
		AccountID: requester.ID,
		ThreadID:  status.ThreadID,
	})

	// Ensure status and reply are still not muted to requester.
	muted1, err1 = suite.filter.StatusMuted(ctx, requester, status)
	muted2, err2 = suite.filter.StatusMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)

	// Though now ensure notifications for both ARE muted to requester.
	muted1, err = suite.filter.StatusNotificationsMuted(ctx, requester, status)
	muted2, err = suite.filter.StatusNotificationsMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.True(muted1)
	suite.True(muted2)

	// Now delete the mute from requester targetting thread.
	err = suite.state.DB.DeleteThreadMute(ctx, threadMuteID)
	suite.NoError(err)

	// Andf ensure notifications for both are unmuted to the requester again.
	muted1, err = suite.filter.StatusNotificationsMuted(ctx, requester, status)
	muted2, err = suite.filter.StatusNotificationsMuted(ctx, requester, reply)
	suite.NoError(err1)
	suite.NoError(err2)
	suite.False(muted1)
	suite.False(muted2)
}

func TestStatusMuteTestSuite(t *testing.T) {
	suite.Run(t, new(StatusMuteTestSuite))
}
