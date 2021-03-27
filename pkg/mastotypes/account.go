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

// Account represents a mastodon-api Account object, as described here: https://docs.joinmastodon.org/entities/account/
type Account struct {
	// The account id
	ID string `json:"id"`
	// The username of the account, not including domain.
	Username string `json:"username"`
	// The Webfinger account URI. Equal to username for local users, or username@domain for remote users.
	Acct string `json:"acct"`
	// The profile's display name.
	DisplayName string `json:"display_name"`
	// Whether the account manually approves follow requests.
	Locked bool `json:"locked"`
	// Whether the account has opted into discovery features such as the profile directory.
	Discoverable bool `json:"discoverable,omitempty"`
	// A presentational flag. Indicates that the account may perform automated actions, may not be monitored, or identifies as a robot.
	Bot bool `json:"bot"`
	// When the account was created. (ISO 8601 Datetime)
	CreatedAt string `json:"created_at"`
	// The profile's bio / description.
	Note string `json:"note"`
	// The location of the user's profile page.
	URL string `json:"url"`
	// An image icon that is shown next to statuses and in the profile.
	Avatar string `json:"avatar"`
	// A static version of the avatar. Equal to avatar if its value is a static image; different if avatar is an animated GIF.
	AvatarStatic string `json:"avatar_static"`
	// An image banner that is shown above the profile and in profile cards.
	Header string `json:"header"`
	// A static version of the header. Equal to header if its value is a static image; different if header is an animated GIF.
	HeaderStatic string `json:"header_static"`
	//  The reported followers of this profile.
	FollowersCount int `json:"followers_count"`
	// The reported follows of this profile.
	FollowingCount int `json:"following_count"`
	// How many statuses are attached to this account.
	StatusesCount int `json:"statuses_count"`
	// When the most recent status was posted. (ISO 8601 Datetime)
	LastStatusAt string `json:"last_status_at"`
	// Custom emoji entities to be used when rendering the profile. If none, an empty array will be returned.
	Emojis []Emoji `json:"emojis"`
	// Additional metadata attached to a profile as name-value pairs.
	Fields []Field `json:"fields"`
	// An extra entity returned when an account is suspended.
	Suspended bool `json:"suspended,omitempty"`
	// When a timed mute will expire, if applicable. (ISO 8601 Datetime)
	MuteExpiresAt string `json:"mute_expires_at,omitempty"`
	// An extra entity to be used with API methods to verify credentials and update credentials.
	Source *Source `json:"source"`
}

// AccountCreateRequest represents the form submitted during a POST request to /api/v1/accounts.
// See https://docs.joinmastodon.org/methods/accounts/
type AccountCreateRequest struct {
	// Text that will be reviewed by moderators if registrations require manual approval.
	Reason string `form:"reason"`
	// The desired username for the account
	Username string `form:"username" binding:"required"`
	// The email address to be used for login
	Email string `form:"email" binding:"required"`
	// The password to be used for login
	Password string `form:"password" binding:"required"`
	// Whether the user agrees to the local rules, terms, and policies.
	// These should be presented to the user in order to allow them to consent before setting this parameter to TRUE.
	Agreement bool `form:"agreement" binding:"required"`
	// The language of the confirmation email that will be sent
	Locale string `form:"locale" binding:"required"`
}
