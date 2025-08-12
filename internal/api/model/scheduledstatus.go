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

import "time"

// ScheduledStatus represents a status that will be published at a future scheduled date.
//
// swagger:model scheduledStatus
type ScheduledStatus struct {
	ID               string                 `json:"id"`
	ScheduledAt      string                 `json:"scheduled_at"`
	Params           *ScheduledStatusParams `json:"params"`
	MediaAttachments []*Attachment          `json:"media_attachments"`
}

// StatusParams represents parameters for a scheduled status.
type ScheduledStatusParams struct {
	Text              string                     `json:"text"`
	MediaIDs          []string                   `json:"media_ids,omitempty"`
	Sensitive         bool                       `json:"sensitive,omitempty"`
	Poll              *ScheduledStatusParamsPoll `json:"poll,omitempty"`
	SpoilerText       string                     `json:"spoiler_text,omitempty"`
	Visibility        Visibility                 `json:"visibility"`
	InReplyToID       string                     `json:"in_reply_to_id,omitempty"`
	Language          string                     `json:"language"`
	ApplicationID     string                     `json:"application_id"`
	LocalOnly         bool                       `json:"local_only,omitempty"`
	ContentType       StatusContentType          `json:"content_type,omitempty"`
	InteractionPolicy *InteractionPolicy         `json:"interaction_policy,omitempty"`
	ScheduledAt       *string                    `json:"scheduled_at"`
}

type ScheduledStatusParamsPoll struct {
	Options    []string `json:"options"`
	ExpiresIn  int      `json:"expires_in"`
	Multiple   bool     `json:"multiple"`
	HideTotals bool     `json:"hide_totals"`
}

// ScheduledStatusUpdateRequest models a request to update the scheduled status publication date.
//
// swagger:ignore
type ScheduledStatusUpdateRequest struct {
	// ISO 8601 Datetime at which to schedule a status.
	ScheduledAt *time.Time `form:"scheduled_at" json:"scheduled_at"`
}
