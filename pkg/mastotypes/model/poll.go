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

type Poll struct {
}

// PollRequest represents a mastodon-api poll attached to a status POST request, as defined here: https://docs.joinmastodon.org/methods/statuses/
// It should be used at the path https://mastodon.example/api/v1/statuses
type PollRequest struct {
	// Array of possible answers. If provided, media_ids cannot be used, and poll[expires_in] must be provided.
	Options []string `form:"options"`
	// Duration the poll should be open, in seconds. If provided, media_ids cannot be used, and poll[options] must be provided.
	ExpiresIn int `form:"expires_in"`
	// Allow multiple choices?
	Multiple bool `form:"multiple"`
	// Hide vote counts until the poll ends?
	HideTotals bool `form:"hide_totals"`
}
