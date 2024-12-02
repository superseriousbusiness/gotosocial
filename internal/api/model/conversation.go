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

package model

// Conversation represents a conversation
// with "direct message" visibility.
//
// swagger:model conversation
type Conversation struct {
	// Local database ID of the conversation.
	ID string `json:"id"`
	// Is the conversation currently marked as unread?
	Unread bool `json:"unread"`
	// Participants in the conversation.
	//
	// If this is a conversation between no accounts (ie., a self-directed DM),
	// this will include only the requesting account itself. Otherwise, it will
	// include every other account in the conversation *except* the requester.
	Accounts []Account `json:"accounts"`
	// The last status in the conversation. May be `null`.
	LastStatus *Status `json:"last_status"`
}
