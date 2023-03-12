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

// Conversation represents a conversation with "direct message" visibility.
type Conversation struct {
	// REQUIRED

	// Local database ID of the conversation.
	ID string `json:"id"`
	// Participants in the conversation.
	Accounts []Account `json:"accounts"`
	// Is the conversation currently marked as unread?
	Unread bool `json:"unread"`

	// OPTIONAL

	// The last status in the conversation, to be used for optional display.
	LastStatus *Status `json:"last_status"`
}
