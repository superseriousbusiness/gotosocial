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

package dereferencing

import (
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// getEmojiByShortcodeDomain searches input slice
// for emoji with given shortcode and domain.
func getEmojiByShortcodeDomain(
	emojis []*gtsmodel.Emoji,
	shortcode string,
	domain string,
) (
	*gtsmodel.Emoji,
	bool,
) {
	for _, emoji := range emojis {
		if emoji.Shortcode == shortcode &&
			emoji.Domain == domain {
			return emoji, true
		}
	}
	return nil, false
}

// emojiChanged returns whether an emoji has changed in a way
// that indicates that it should be refetched and refreshed.
func emojiChanged(existing, latest *gtsmodel.Emoji) bool {
	return existing.URI != latest.URI ||
		existing.ImageRemoteURL != latest.ImageRemoteURL ||
		existing.ImageStaticRemoteURL != latest.ImageStaticRemoteURL
}

// pollChanged returns whether a poll has changed in way that
// indicates that this should be an entirely new poll. i.e. if
// the available options have changed, or the expiry has increased.
func pollChanged(existing, latest *gtsmodel.Poll) bool {
	return !slices.Equal(existing.Options, latest.Options) ||
		!existing.ExpiresAt.Equal(latest.ExpiresAt)
}

// pollUpdated returns whether a poll has updated, i.e. if the
// vote counts have changed, or if it has expired / been closed.
func pollUpdated(existing, latest *gtsmodel.Poll) bool {
	return *existing.Voters != *latest.Voters ||
		!slices.Equal(existing.Votes, latest.Votes) ||
		!existing.ClosedAt.Equal(latest.ClosedAt)
}

// pollJustClosed returns whether a poll has *just* closed.
func pollJustClosed(existing, latest *gtsmodel.Poll) bool {
	return existing.ClosedAt.IsZero() && latest.Closed()
}
