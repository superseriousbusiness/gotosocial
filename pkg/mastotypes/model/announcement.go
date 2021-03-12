/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package mastotypes

// Announcement represents an admin/moderator announcement for local users. See here: https://docs.joinmastodon.org/entities/announcement/
type Announcement struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	StartsAt    string                 `json:"starts_at"`
	EndsAt      string                 `json:"ends_at"`
	AllDay      bool                   `json:"all_day"`
	PublishedAt string                 `json:"published_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Published   bool                   `json:"published"`
	Read        bool                   `json:"read"`
	Mentions    []Mention              `json:"mentions"`
	Statuses    []Status               `json:"statuses"`
	Tags        []Tag                  `json:"tags"`
	Emojis      []Emoji                `json:"emoji"`
	Reactions   []AnnouncementReaction `json:"reactions"`
}
