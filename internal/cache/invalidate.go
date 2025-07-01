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
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// Below are cache invalidation hooks between other caches,
// as an invalidation indicates a database INSERT / UPDATE / DELETE.
// NOTE THEY ARE ONLY CALLED WHEN THE ITEM IS IN THE CACHE, SO FOR
// HOOKS TO BE CALLED ON DELETE YOU MUST FIRST POPULATE IT IN THE CACHE.
//
// Also note that while Timelines are a part of the Caches{} object,
// they are generally not modified as part of side-effects here, as
// they often need specific IDs or more information that can only be
// fetched from the database. As such, they are generally handled as
// side-effects in the ./internal/processor/workers/ package.

func (c *Caches) OnInvalidateAccount(account *gtsmodel.Account) {
	// Invalidate as possible visibility target result.
	c.Visibility.Invalidate("ItemID", account.ID)

	// If account is local, invalidate as
	// possible visibility result requester,
	// also, invalidate any cached stats.
	if account.IsLocal() {
		c.DB.AccountStats.Invalidate("AccountID", account.ID)
		c.Visibility.Invalidate("RequesterID", account.ID)
	}

	// Invalidate this account's
	// following / follower lists.
	// (see FollowIDs() comment for details).
	c.DB.FollowIDs.Invalidate(
		">"+account.ID,
		"l>"+account.ID,
		"<"+account.ID,
		"l<"+account.ID,
	)

	// Invalidate this account's
	// follow requesting / request lists.
	// (see FollowRequestIDs() comment for details).
	c.DB.FollowRequestIDs.Invalidate(
		">"+account.ID,
		"<"+account.ID,
	)

	// Invalidate this account's block lists.
	c.DB.BlockIDs.Invalidate(account.ID)

	// Invalidate this account's Move(s).
	c.DB.Move.Invalidate("OriginURI", account.URI)
	c.DB.Move.Invalidate("TargetURI", account.URI)
}

func (c *Caches) OnInvalidateApplication(app *gtsmodel.Application) {
	// TODO: invalidate tokens?
}

func (c *Caches) OnInvalidateBlock(block *gtsmodel.Block) {
	// Invalidate both block origin and target as
	// possible lookup targets for visibility results.
	c.Visibility.InvalidateIDs("ItemID", []string{
		block.TargetAccountID,
		block.AccountID,
	})

	// Track which of block / target are local.
	localAccountIDs := make([]string, 0, 2)

	// If origin is local (or uncertain), also invalidate
	// results for them as mute / visibility result requester.
	if block.Account == nil || block.Account.IsLocal() {
		localAccountIDs = append(localAccountIDs, block.AccountID)
	}

	// If target is local (or uncertain), also invalidate
	// results for them as mute / visibility result requester.
	if block.TargetAccount == nil || block.TargetAccount.IsLocal() {
		localAccountIDs = append(localAccountIDs, block.TargetAccountID)
	}

	// Now perform local visibility result invalidations.
	c.Visibility.InvalidateIDs("RequesterID", localAccountIDs)

	// Invalidate source account's block lists.
	c.DB.BlockIDs.Invalidate(block.AccountID)
}

func (c *Caches) OnInvalidateConversation(conversation *gtsmodel.Conversation) {
	// Invalidate owning account's conversation list.
	c.DB.ConversationLastStatusIDs.Invalidate(conversation.AccountID)
}

func (c *Caches) OnInvalidateEmojiCategory(category *gtsmodel.EmojiCategory) {
	// Invalidate any emoji in this category.
	c.DB.Emoji.Invalidate("CategoryID", category.ID)
}

func (c *Caches) OnInvalidateFilter(filter *gtsmodel.Filter) {
	// Invalidate list of filters for account.
	c.DB.FilterIDs.Invalidate(filter.AccountID)

	// Invalidate all associated keywords and statuses.
	c.DB.FilterKeyword.InvalidateIDs("ID", filter.KeywordIDs)
	c.DB.FilterStatus.InvalidateIDs("ID", filter.StatusIDs)

	// Invalidate account's status filter cache.
	c.StatusFilter.Invalidate("RequesterID", filter.AccountID)
}

func (c *Caches) OnInvalidateFilterKeyword(filterKeyword *gtsmodel.FilterKeyword) {
	// Invalidate filter that keyword associated with.
	c.DB.Filter.Invalidate("ID", filterKeyword.FilterID)
}

func (c *Caches) OnInvalidateFilterStatus(filterStatus *gtsmodel.FilterStatus) {
	// Invalidate filter that status associated with.
	c.DB.Filter.Invalidate("ID", filterStatus.FilterID)
}

func (c *Caches) OnInvalidateFollow(follow *gtsmodel.Follow) {
	// Invalidate follow request with this same ID.
	c.DB.FollowRequest.Invalidate("ID", follow.ID)

	// Invalidate both follow origin and target as
	// possible lookup targets for visibility results.
	c.Visibility.InvalidateIDs("ItemID", []string{
		follow.TargetAccountID,
		follow.AccountID,
	})

	// Track which of follow / target are local.
	localAccountIDs := make([]string, 0, 2)

	// If origin is local (or uncertain), also invalidate
	// results for them as mute / visibility result requester.
	if follow.Account == nil || follow.Account.IsLocal() {
		localAccountIDs = append(localAccountIDs, follow.AccountID)
	}

	// If target is local (or uncertain), also invalidate
	// results for them as mute / visibility result requester.
	if follow.TargetAccount == nil || follow.TargetAccount.IsLocal() {
		localAccountIDs = append(localAccountIDs, follow.TargetAccountID)
	}

	// Now perform local visibility result invalidations.
	c.Visibility.InvalidateIDs("RequesterID", localAccountIDs)

	// Invalidate ID slice cache.
	c.DB.FollowIDs.Invalidate(

		// Invalidate follow ID lists
		// TARGETTING origin account
		// (including local-only follows).
		">"+follow.AccountID,
		"l>"+follow.AccountID,

		// Invalidate follow ID lists
		// FROM the origin account
		// (including local-only follows).
		"<"+follow.AccountID,
		"l<"+follow.AccountID,

		// Invalidate follow ID lists
		// TARGETTING the target account
		// (including local-only follows).
		">"+follow.TargetAccountID,
		"l>"+follow.TargetAccountID,

		// Invalidate follow ID lists
		// FROM the target account
		// (including local-only follows).
		"<"+follow.TargetAccountID,
		"l<"+follow.TargetAccountID,
	)

	// Invalidate ID slice cache.
	c.DB.ListIDs.Invalidate(

		// Invalidate source
		// account's owned lists.
		"a"+follow.AccountID,

		// Invalidate target account's.
		"a"+follow.TargetAccountID,

		// Invalidate lists containing
		// list entries for follow.
		"f"+follow.ID,
	)
}

func (c *Caches) OnInvalidateFollowRequest(followReq *gtsmodel.FollowRequest) {
	// Invalidate follow with this same ID.
	c.DB.Follow.Invalidate("ID", followReq.ID)

	// Invalidate ID slice cache.
	c.DB.FollowRequestIDs.Invalidate(

		// Invalidate follow request ID
		// lists TARGETTING origin account
		// (including local-only follows).
		">"+followReq.AccountID,

		// Invalidate follow request ID
		// lists FROM the origin account
		// (including local-only follows).
		"<"+followReq.AccountID,

		// Invalidate follow request ID
		// lists TARGETTING target account
		// (including local-only follows).
		">"+followReq.TargetAccountID,

		// Invalidate follow request ID
		// lists FROM the target account
		// (including local-only follows).
		"<"+followReq.TargetAccountID,
	)
}

func (c *Caches) OnInvalidateInstance(instance *gtsmodel.Instance) {
	// Invalidate the local domains count.
	c.DB.LocalInstance.Domains.Store(nil)
}

func (c *Caches) OnInvalidateList(list *gtsmodel.List) {
	// Invalidate list IDs cache.
	c.DB.ListIDs.Invalidate(
		"a" + list.AccountID,
	)

	// Invalidate ID slice cache.
	c.DB.ListedIDs.Invalidate(

		// Invalidate list of
		// account IDs in list.
		"a"+list.ID,

		// Invalidate list of
		// follow IDs in list.
		"f"+list.ID,
	)
}

func (c *Caches) OnInvalidateMedia(media *gtsmodel.MediaAttachment) {
	if (media.Avatar != nil && *media.Avatar) ||
		(media.Header != nil && *media.Header) {
		// Invalidate cache of attaching account.
		c.DB.Account.Invalidate("ID", media.AccountID)
	}

	if media.StatusID != "" {
		// Invalidate cache of attaching status.
		c.DB.Status.Invalidate("ID", media.StatusID)
	}
}

func (c *Caches) OnInvalidatePoll(poll *gtsmodel.Poll) {
	// Invalidate all cached votes of this poll.
	c.DB.PollVote.Invalidate("PollID", poll.ID)

	// Invalidate cache of poll vote IDs.
	c.DB.PollVoteIDs.Invalidate(poll.ID)
}

func (c *Caches) OnInvalidatePollVote(vote *gtsmodel.PollVote) {
	// Invalidate cached poll (contains no. votes).
	c.DB.Poll.Invalidate("ID", vote.PollID)

	// Invalidate cache of poll vote IDs.
	c.DB.PollVoteIDs.Invalidate(vote.PollID)
}

func (c *Caches) OnInvalidateStatus(status *gtsmodel.Status) {
	// Invalidate cached stats objects for this account.
	c.DB.AccountStats.Invalidate("AccountID", status.AccountID)

	// Invalidate filter results targeting status.
	c.StatusFilter.Invalidate("StatusID", status.ID)

	// Invalidate status ID cached visibility.
	c.Visibility.Invalidate("ItemID", status.ID)

	// Invalidate mute results involving status.
	c.Mutes.Invalidate("StatusID", status.ID)
	c.Mutes.Invalidate("ThreadID", status.ThreadID)

	// Invalidate each media by the IDs we're aware of.
	// This must be done as the status table is aware of
	// the media IDs in use before the media table is
	// aware of the status ID they are linked to.
	//
	// c.DB.Media.Invalidate("StatusID") will not work.
	c.DB.Media.InvalidateIDs("ID", status.AttachmentIDs)

	if status.BoostOfID != "" {
		// Invalidate boost ID list of the original status.
		c.DB.BoostOfIDs.Invalidate(status.BoostOfID)
	}

	if status.InReplyToID != "" {
		// Invalidate in reply to ID list of original status.
		c.DB.InReplyToIDs.Invalidate(status.InReplyToID)
	}

	if status.PollID != "" {
		// Invalidate cache of attached poll ID.
		c.DB.Poll.Invalidate("ID", status.PollID)
	}

	if util.PtrOrZero(status.Local) {
		// Invalidate the local statuses count.
		c.DB.LocalInstance.Statuses.Store(nil)
	}
}

func (c *Caches) OnInvalidateStatusBookmark(bookmark *gtsmodel.StatusBookmark) {
	// Invalidate status bookmark ID list for this status.
	c.DB.StatusBookmarkIDs.Invalidate(bookmark.StatusID)
}

func (c *Caches) OnInvalidateStatusEdit(edit *gtsmodel.StatusEdit) {
	// Invalidate cache of related status model.
	c.DB.Status.Invalidate("ID", edit.StatusID)
}

func (c *Caches) OnInvalidateStatusFave(fave *gtsmodel.StatusFave) {
	// Invalidate status fave ID list for this status.
	c.DB.StatusFaveIDs.Invalidate(fave.StatusID)
}

func (c *Caches) OnInvalidateThreadMute(mute *gtsmodel.ThreadMute) {
	// Invalidate cached mute ressults encapsulating this thread and account.
	c.Mutes.Invalidate("RequesterID,ThreadID", mute.AccountID, mute.ThreadID)
}

func (c *Caches) OnInvalidateToken(token *gtsmodel.Token) {
	// Invalidate token's push subscription.
	c.DB.WebPushSubscription.Invalidate("ID", token.ID)
}

func (c *Caches) OnInvalidateUser(user *gtsmodel.User) {
	// Invalidate local account ID cached visibility.
	c.Visibility.Invalidate("ItemID", user.AccountID)
	c.Visibility.Invalidate("RequesterID", user.AccountID)

	// Invalidate the local users count.
	c.DB.LocalInstance.Users.Store(nil)
}

func (c *Caches) OnInvalidateUserMute(mute *gtsmodel.UserMute) {
	// Invalidate source account's user mute lists.
	c.DB.UserMuteIDs.Invalidate(mute.AccountID)

	// Invalidate source account's cached mute results.
	c.Mutes.Invalidate("RequesterID", mute.AccountID)
}

func (c *Caches) OnInvalidateWebPushSubscription(subscription *gtsmodel.WebPushSubscription) {
	// Invalidate source account's Web Push subscription list.
	c.DB.WebPushSubscriptionIDs.Invalidate(subscription.AccountID)
}
