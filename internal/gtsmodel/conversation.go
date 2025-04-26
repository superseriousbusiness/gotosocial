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

package gtsmodel

import (
	"slices"
	"strings"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
)

// Conversation represents direct messages between the owner account and a set of other accounts.
type Conversation struct {
	// ID of this item in the database.
	ID string `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`

	// When was this item created?
	CreatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// When was this item last updated?
	UpdatedAt time.Time `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`

	// Account that owns the conversation.
	AccountID string   `bun:"type:CHAR(26),nullzero,notnull,unique:conversations_thread_id_account_id_other_accounts_key_uniq,unique:conversations_account_id_last_status_id_uniq"`
	Account   *Account `bun:"-"`

	// Other accounts participating in the conversation.
	// Doesn't include the owner. May be empty in the case of a DM to yourself.
	OtherAccountIDs []string   `bun:"other_account_ids,array"`
	OtherAccounts   []*Account `bun:"-"`

	// Denormalized lookup key derived from unique OtherAccountIDs, sorted and concatenated with commas.
	// May be empty in the case of a DM to yourself.
	OtherAccountsKey string `bun:",notnull,unique:conversations_thread_id_account_id_other_accounts_key_uniq"`

	// Thread that the conversation is part of.
	ThreadID string `bun:"type:CHAR(26),nullzero,notnull,unique:conversations_thread_id_account_id_other_accounts_key_uniq"`

	// ID of the last status in this conversation.
	LastStatusID string  `bun:"type:CHAR(26),nullzero,notnull,unique:conversations_account_id_last_status_id_uniq"`
	LastStatus   *Status `bun:"-"`

	// Has the owner read all statuses in this conversation?
	Read *bool `bun:",default:false"`
}

// ConversationOtherAccountsKey creates an OtherAccountsKey from a list of OtherAccountIDs.
func ConversationOtherAccountsKey(otherAccountIDs []string) string {
	otherAccountIDs = xslices.Deduplicate(otherAccountIDs)
	slices.Sort(otherAccountIDs)
	return strings.Join(otherAccountIDs, ",")
}

// ConversationToStatus is an intermediate struct to facilitate the many2many relationship between a conversation and its statuses,
// including but not limited to the last status. These are used only when deleting a status from a conversation.
type ConversationToStatus struct {
	ConversationID string        `bun:"type:CHAR(26),unique:conversation_to_statuses_conversation_id_status_id_uniq,nullzero,notnull"`
	Conversation   *Conversation `bun:"rel:belongs-to"`
	StatusID       string        `bun:"type:CHAR(26),unique:conversation_to_statuses_conversation_id_status_id_uniq,nullzero,notnull"`
	Status         *Status       `bun:"rel:belongs-to"`
}
