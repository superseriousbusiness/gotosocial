/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package model

// ScheduledStatus represents a status that will be published at a future scheduled date.
type ScheduledStatus struct {
	ID               string        `json:"id"`
	ScheduledAt      string        `json:"scheduled_at"`
	Params           *StatusParams `json:"params"`
	MediaAttachments []Attachment  `json:"media_attachments"`
}

// StatusParams represents parameters for a scheduled status.
type StatusParams struct {
	Text          string   `json:"text"`
	InReplyToID   string   `json:"in_reply_to_id,omitempty"`
	MediaIDs      []string `json:"media_ids,omitempty"`
	Sensitive     bool     `json:"sensitive,omitempty"`
	SpoilerText   string   `json:"spoiler_text,omitempty"`
	Visibility    string   `json:"visibility"`
	ScheduledAt   string   `json:"scheduled_at,omitempty"`
	ApplicationID string   `json:"application_id"`
}
