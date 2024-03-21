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

package cache

import (
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Below are cache invalidation hooks between other caches,
// as an invalidation indicates a database INSERT / UPDATE / DELETE.
// NOTE THEY ARE ONLY CALLED WHEN THE ITEM IS IN THE CACHE, SO FOR
// HOOKS TO BE CALLED ON DELETE YOU MUST FIRST POPULATE IT IN THE CACHE.

func (c *Caches) OnInvalidateAccount(account *gtsmodel.Account) {
	// Invalidate status counts for this account.
	c.GTS.AccountCounts.Invalidate(account.ID)

	// Invalidate account ID cached visibility.
	c.Visibility.Invalidate("ItemID", account.ID)
	c.Visibility.Invalidate("RequesterID", account.ID)

	// Invalidate this account's
	// following / follower lists.
	// (see FollowIDs() comment for details).
	c.GTS.FollowIDs.InvalidateAll(
		">"+account.ID,
		"l>"+account.ID,
		"<"+account.ID,
		"l<"+account.ID,
	)

	// Invalidate this account's
	// follow requesting / request lists.
	// (see FollowRequestIDs() comment for details).
	c.GTS.FollowRequestIDs.InvalidateAll(
		">"+account.ID,
		"<"+account.ID,
	)

	// Invalidate this account's block lists.
	c.GTS.BlockIDs.Invalidate(account.ID)

	// Invalidate this account's Move(s).
	c.GTS.Move.Invalidate("OriginURI", account.URI)
	c.GTS.Move.Invalidate("TargetURI", account.URI)
}

func (c *Caches) OnInvalidateBlock(block *gtsmodel.Block) {
	// Invalidate block origin account ID cached visibility.
	c.Visibility.Invalidate("ItemID", block.AccountID)
	c.Visibility.Invalidate("RequesterID", block.AccountID)

	// Invalidate block target account ID cached visibility.
	c.Visibility.Invalidate("ItemID", block.TargetAccountID)
	c.Visibility.Invalidate("RequesterID", block.TargetAccountID)

	// Invalidate source account's block lists.
	c.GTS.BlockIDs.Invalidate(block.AccountID)
}

func (c *Caches) OnInvalidateEmojiCategory(category *gtsmodel.EmojiCategory) {
	// Invalidate any emoji in this category.
	c.GTS.Emoji.Invalidate("CategoryID", category.ID)
}

func (c *Caches) OnInvalidateFollow(follow *gtsmodel.Follow) {
	// Invalidate follow request with this same ID.
	c.GTS.FollowRequest.Invalidate("ID", follow.ID)

	// Invalidate any related list entries.
	c.GTS.ListEntry.Invalidate("FollowID", follow.ID)

	// Invalidate follow origin account ID cached visibility.
	c.Visibility.Invalidate("ItemID", follow.AccountID)
	c.Visibility.Invalidate("RequesterID", follow.AccountID)

	// Invalidate follow target account ID cached visibility.
	c.Visibility.Invalidate("ItemID", follow.TargetAccountID)
	c.Visibility.Invalidate("RequesterID", follow.TargetAccountID)

	// Invalidate source account's following
	// lists, and destination's follwer lists.
	// (see FollowIDs() comment for details).
	c.GTS.FollowIDs.InvalidateAll(
		">"+follow.AccountID,
		"l>"+follow.AccountID,
		"<"+follow.AccountID,
		"l<"+follow.AccountID,
		"<"+follow.TargetAccountID,
		"l<"+follow.TargetAccountID,
		">"+follow.TargetAccountID,
		"l>"+follow.TargetAccountID,
	)
}

func (c *Caches) OnInvalidateFollowRequest(followReq *gtsmodel.FollowRequest) {
	// Invalidate follow with this same ID.
	c.GTS.Follow.Invalidate("ID", followReq.ID)

	// Invalidate source account's followreq
	// lists, and destinations follow req lists.
	// (see FollowRequestIDs() comment for details).
	c.GTS.FollowRequestIDs.InvalidateAll(
		">"+followReq.AccountID,
		"<"+followReq.AccountID,
		">"+followReq.TargetAccountID,
		"<"+followReq.TargetAccountID,
	)
}

func (c *Caches) OnInvalidateList(list *gtsmodel.List) {
	// Invalidate all cached entries of this list.
	c.GTS.ListEntry.Invalidate("ListID", list.ID)
}

func (c *Caches) OnInvalidateMedia(media *gtsmodel.MediaAttachment) {
	if (media.Avatar != nil && *media.Avatar) ||
		(media.Header != nil && *media.Header) {
		// Invalidate cache of attaching account.
		c.GTS.Account.Invalidate("ID", media.AccountID)
	}

	if media.StatusID != "" {
		// Invalidate cache of attaching status.
		c.GTS.Status.Invalidate("ID", media.StatusID)
	}
}

func (c *Caches) OnInvalidatePoll(poll *gtsmodel.Poll) {
	// Invalidate all cached votes of this poll.
	c.GTS.PollVote.Invalidate("PollID", poll.ID)

	// Invalidate cache of poll vote IDs.
	c.GTS.PollVoteIDs.Invalidate(poll.ID)
}

func (c *Caches) OnInvalidatePollVote(vote *gtsmodel.PollVote) {
	// Invalidate cached poll (contains no. votes).
	c.GTS.Poll.Invalidate("ID", vote.PollID)

	// Invalidate cache of poll vote IDs.
	c.GTS.PollVoteIDs.Invalidate(vote.PollID)
}

func (c *Caches) OnInvalidateStatus(status *gtsmodel.Status) {
	// Invalidate status counts for this account.
	c.GTS.AccountCounts.Invalidate(status.AccountID)

	// Invalidate status ID cached visibility.
	c.Visibility.Invalidate("ItemID", status.ID)

	// Invalidate each media by the IDs we're aware of.
	// This must be done as the status table is aware of
	// the media IDs in use before the media table is
	// aware of the status ID they are linked to.
	//
	// c.GTS.Media().Invalidate("StatusID") will not work.
	c.GTS.Media.InvalidateIDs("ID", status.AttachmentIDs)

	if status.BoostOfID != "" {
		// Invalidate boost ID list of the original status.
		c.GTS.BoostOfIDs.Invalidate(status.BoostOfID)
	}

	if status.InReplyToID != "" {
		// Invalidate in reply to ID list of original status.
		c.GTS.InReplyToIDs.Invalidate(status.InReplyToID)
	}

	if status.PollID != "" {
		// Invalidate cache of attached poll ID.
		c.GTS.Poll.Invalidate("ID", status.PollID)
	}
}

func (c *Caches) OnInvalidateStatusFave(fave *gtsmodel.StatusFave) {
	// Invalidate status fave ID list for this status.
	c.GTS.StatusFaveIDs.Invalidate(fave.StatusID)
}

func (c *Caches) OnInvalidateUser(user *gtsmodel.User) {
	// Invalidate local account ID cached visibility.
	c.Visibility.Invalidate("ItemID", user.AccountID)
	c.Visibility.Invalidate("RequesterID", user.AccountID)
}
