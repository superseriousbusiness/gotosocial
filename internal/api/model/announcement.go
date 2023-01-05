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

// Announcement models an admin announcement for the instance.
//
// swagger:model announcement
type Announcement struct {
	// The ID of the announcement.
	// example: 01FC30T7X4TNCZK0TH90QYF3M4
	ID string `json:"id"`
	// The body of the announcement.
	// Should be HTML formatted.
	// example: <p>This is an announcement. No malarky.</p>
	Content string `json:"content"`
	// When the announcement should begin to be displayed (ISO 8601 Datetime).
	// If the announcement has no start time, this will be omitted or empty.
	// example: 2021-07-30T09:20:25+00:00
	StartsAt string `json:"starts_at"`
	// When the announcement should stop being displayed (ISO 8601 Datetime).
	// If the announcement has no end time, this will be omitted or empty.
	// example: 2021-07-30T09:20:25+00:00
	EndsAt string `json:"ends_at"`
	// Announcement doesn't have begin time and end time, but begin day and end day.
	AllDay bool `json:"all_day"`
	// When the announcement was first published (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	PublishedAt string `json:"published_at"`
	// When the announcement was last updated (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	UpdatedAt string `json:"updated_at"`
	// Announcement is 'published', ie., visible to users.
	// Announcements that are not published should be shown only to admins.
	Published bool `json:"published"`
	// Requesting account has seen this announcement.
	Read bool `json:"read"`
	// Mentions this announcement contains.
	Mentions []Mention `json:"mentions"`
	// Statuses contained in this announcement.
	Statuses []Status `json:"statuses"`
	// Tags used in this announcement.
	Tags []Tag `json:"tags"`
	// Emojis used in this announcement.
	Emojis []Emoji `json:"emoji"`
	// Reactions to this announcement.
	Reactions []AnnouncementReaction `json:"reactions"`
}
