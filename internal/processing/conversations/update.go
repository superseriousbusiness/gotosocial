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

package conversations

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	statusfilter "github.com/superseriousbusiness/gotosocial/internal/filter/status"
	"github.com/superseriousbusiness/gotosocial/internal/filter/usermute"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// UpdateConversationsForStatus updates all conversations related to a status,
// and returns a map from local account IDs to conversation notifications that should be sent to them.
func (p *Processor) UpdateConversationsForStatus(ctx context.Context, status *gtsmodel.Status) (map[string]*apimodel.Conversation, error) {
	notifications := map[string]*apimodel.Conversation{}

	// We need accounts to be populated for this.
	if err := p.state.DB.PopulateStatus(ctx, status); err != nil {
		return nil, err
	}

	if status.Visibility != gtsmodel.VisibilityDirect {
		// Only DMs are considered part of conversations.
		return nil, nil
	}
	if status.BoostOfID != "" {
		// Boosts can't be part of conversations.
		// FUTURE: This may change if we ever implement quote posts.
		return nil, nil
	}
	if status.ThreadID == "" {
		// If the status doesn't have a thread ID, it didn't mention a local account,
		// and thus can't be part of a conversation.
		return nil, nil
	}

	// The account which authored the status plus all mentioned accounts.
	allParticipantsSet := make(map[string]*gtsmodel.Account, 1+len(status.Mentions))
	allParticipantsSet[status.AccountID] = status.Account
	for _, mention := range status.Mentions {
		allParticipantsSet[mention.TargetAccountID] = mention.TargetAccount
	}

	// Create or update conversations for and send notifications to each local participant.
	for _, participant := range allParticipantsSet {
		if participant.IsRemote() {
			continue
		}
		localAccount := participant

		// If the status is not visible to this account, skip processing it for this account.
		visible, err := p.filter.StatusVisible(ctx, localAccount, status)
		if err != nil {
			log.Errorf(
				ctx,
				"error checking status %s visibility for account %s: %v",
				status.ID,
				localAccount.ID,
				err,
			)
			continue
		} else if !visible {
			continue
		}

		// TODO: (Vyr) find a prettier way to do this
		// Is the status filtered or muted for this user?
		filters, err := p.state.DB.GetFiltersForAccountID(ctx, localAccount.ID)
		if err != nil {
			log.Errorf(ctx, "error retrieving filters for account %s: %v", localAccount.ID, err)
			continue
		}
		mutes, err := p.state.DB.GetAccountMutes(gtscontext.SetBarebones(ctx), localAccount.ID, nil)
		if err != nil {
			log.Errorf(ctx, "error retrieving mutes for account %s: %v", localAccount.ID, err)
			continue
		}
		compiledMutes := usermute.NewCompiledUserMuteList(mutes)
		// Converting the status to an API status runs the filter/mute checks.
		_, err = p.converter.StatusToAPIStatus(
			ctx,
			status,
			localAccount,
			statusfilter.FilterContextNotifications,
			filters,
			compiledMutes,
		)
		if err != nil {
			// If the status matched a hide filter, skip processing it for this account.
			// If there was another kind of error, log that and skip it anyway.
			if !errors.Is(err, statusfilter.ErrHideStatus) {
				log.Errorf(
					ctx,
					"error checking status %s filtering/muting for account %s: %v",
					status.ID,
					localAccount.ID,
					err,
				)
			}
			continue
		}

		// Collect other accounts participating in the conversation.
		otherAccounts := make([]*gtsmodel.Account, 0, len(allParticipantsSet)-1)
		otherAccountIDs := make([]string, 0, len(allParticipantsSet)-1)
		for accountID, account := range allParticipantsSet {
			if accountID != localAccount.ID {
				otherAccounts = append(otherAccounts, account)
				otherAccountIDs = append(otherAccountIDs, accountID)
			}
		}

		// Check for a previously existing conversation, if there is one.
		conversation, err := p.state.DB.GetConversationByThreadAndAccountIDs(
			ctx,
			status.ThreadID,
			localAccount.ID,
			otherAccountIDs,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf(
				ctx,
				"error trying to find a previous conversation for status %s and account %s: %v",
				status.ID,
				localAccount.ID,
				err,
			)
			continue
		}

		if conversation == nil {
			// Create a new conversation.
			conversation = &gtsmodel.Conversation{
				ID:               id.NewULID(),
				AccountID:        localAccount.ID,
				OtherAccountIDs:  otherAccountIDs,
				OtherAccounts:    otherAccounts,
				OtherAccountsKey: gtsmodel.ConversationOtherAccountsKey(otherAccountIDs),
				ThreadID:         status.ThreadID,
				Read:             util.Ptr(true),
			}
		}

		// Assume that if the conversation owner posted the status, they've already read it.
		statusAuthoredByConversationOwner := status.AccountID == conversation.AccountID

		// Update the conversation.
		// If there is no previous last status or this one is more recently created, set it as the last status.
		if conversation.LastStatus == nil || conversation.LastStatus.CreatedAt.Before(status.CreatedAt) {
			conversation.LastStatusID = status.ID
			conversation.LastStatus = status
		}
		// If the conversation is unread, leave it marked as unread.
		// If the conversation is read but this status might not have been, mark the conversation as unread.
		if !statusAuthoredByConversationOwner {
			conversation.Read = util.Ptr(false)
		}

		// Create or update the conversation.
		err = p.state.DB.UpsertConversation(ctx, conversation)
		if err != nil {
			log.Errorf(
				ctx,
				"error creating or updating conversation %s for status %s and account %s: %v",
				conversation.ID,
				status.ID,
				localAccount.ID,
				err,
			)
			continue
		}

		// Link the conversation to the status.
		if err := p.state.DB.LinkConversationToStatus(ctx, conversation.ID, status.ID); err != nil {
			log.Errorf(
				ctx,
				"error linking conversation %s to status %s: %v",
				conversation.ID,
				status.ID,
				err,
			)
			continue
		}

		// Convert the conversation to API representation.
		apiConversation, err := p.converter.ConversationToAPIConversation(
			ctx,
			conversation,
			localAccount,
			filters,
			compiledMutes,
		)
		if err != nil {
			// If the conversation's last status matched a hide filter, skip it.
			// If there was another kind of error, log that and skip it anyway.
			if !errors.Is(err, statusfilter.ErrHideStatus) {
				log.Errorf(
					ctx,
					"error converting conversation %s to API representation for account %s: %v",
					status.ID,
					localAccount.ID,
					err,
				)
			}
			continue
		}

		// Generate a notification,
		// unless the status was authored by the user who would be notified,
		// in which case they already know.
		if status.AccountID != localAccount.ID {
			notifications[localAccount.ID] = apiConversation
		}
	}

	return notifications, nil
}