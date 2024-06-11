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

	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Conversation represents direct messages between the owner account and a set of other accounts.
type Conversation struct {
	ID               string     `bun:"type:CHAR(26),pk,nullzero,notnull,unique"`                                                        // id of this item in the database
	CreatedAt        time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                     // when was item created
	UpdatedAt        time.Time  `bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"`                                     // when was item last updated
	AccountID        string     `bun:"type:CHAR(26),nullzero,notnull,unique:conversation_thread_id_account_id_other_accounts_key_uniq"` // Account that owns the conversation
	Account          *Account   `bun:"-"`                                                                                               //
	OtherAccountIDs  []string   `bun:",array"`                                                                                          // Other accounts participating in the conversation (doesn't include the owner, may be empty in the case of a DM to yourself)
	OtherAccounts    []*Account `bun:"m2m:conversation_to_accounts"`                                                                    //
	OtherAccountsKey string     `bun:",notnull,unique:conversation_thread_id_account_id_other_accounts_key_uniq"`                       // Denormalized lookup key derived from unique OtherAccountIDs, sorted and concatenated with commas, may be empty in the case of a DM to yourself
	ThreadID         string     `bun:"type:CHAR(26),nullzero,notnull,unique:conversation_thread_id_account_id_other_accounts_key_uniq"` // Thread that the conversation is part of
	LastStatusID     string     `bun:"type:CHAR(26),nullzero,notnull"`                                                                  // id of the last status in this conversation
	LastStatus       *Status    `bun:"-"`                                                                                               //
	Read             *bool      `bun:",default:false"`                                                                                  // Has the owner read all statuses in this conversation?
}

// ConversationOtherAccountsKey creates an OtherAccountsKey from a list of OtherAccountIDs.
func ConversationOtherAccountsKey(otherAccountIDs []string) string {
	otherAccountIDs = util.UniqueStrings(otherAccountIDs)
	slices.Sort(otherAccountIDs)
	return strings.Join(otherAccountIDs, ",")
}

// ConversationToAccount is an intermediate struct to facilitate the many2many relationship between a conversation and one or more non-owner accounts participating in it.
type ConversationToAccount struct {
	ConversationID string        `bun:"type:CHAR(26),unique:conversation_to_accounts_conversation_id_account_id_uniq,nullzero,notnull"`
	Conversation   *Conversation `bun:"rel:belongs-to"`
	AccountID      string        `bun:"type:CHAR(26),unique:conversation_to_accounts_conversation_id_account_id_uniq,nullzero,notnull"`
	Account        *Account      `bun:"rel:belongs-to"`
}
