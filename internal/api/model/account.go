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

package model

import (
	"mime/multipart"
	"net"
)

// Account represents a fediverse account of some kind, either a remote one or one on this instance.
//
// swagger:model account
type Account struct {
	// The account id.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	ID string `json:"id"`
	// The username of the account, not including domain.
	// example: some_user
	Username string `json:"username"`
	// The account URI as discovered via webfinger.
	// Equal to username for local users, or username@domain for remote users.
	// example: some_user@example.org
	Acct string `json:"acct"`
	// The account's display name.
	// example: big jeff (he/him)
	DisplayName string `json:"display_name"`
	// Account manually approves follow requests.
	Locked bool `json:"locked"`
	// Account has opted into discovery features such as the profile directory.
	Discoverable bool `json:"discoverable,omitempty"`
	// Account identifies as a bot.
	Bot bool `json:"bot"`
	// When the account was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`
	// Bio/description of this account.
	Note string `json:"note"`
	// Web location of the account's profile page.
	// example: https://example.org/@some_user
	URL string `json:"url"`
	// Web location of the account's avatar.
	// example: https://example.org/media/some_user/avatar/original/avatar.jpeg
	Avatar string `json:"avatar"`
	// Web location of a static version of the account's avatar.
	// Only relevant when the account's main avatar is a video or a gif.
	// example: https://example.org/media/some_user/avatar/static/avatar.png
	AvatarStatic string `json:"avatar_static"`
	// Web location of the account's header image.
	// example: https://example.org/media/some_user/header/original/header.jpeg
	Header string `json:"header"`
	// Web location of a static version of the account's header.
	// Only relevant when the account's main header is a video or a gif.
	// example: https://example.org/media/some_user/header/static/header.png
	HeaderStatic string `json:"header_static"`
	// Number of accounts following this account, according to our instance.
	FollowersCount int `json:"followers_count"`
	// Number of account's followed by this account, according to our instance.
	FollowingCount int `json:"following_count"`
	// Number of statuses posted by this account, according to our instance.
	StatusesCount int `json:"statuses_count"`
	// When the account's most recent status was posted (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	LastStatusAt string `json:"last_status_at"`
	// Array of custom emojis used in this account's note or display name.
	Emojis []Emoji `json:"emojis"`
	// Additional metadata attached to this account's profile.
	Fields []Field `json:"fields"`
	// Account has been suspended by our instance.
	Suspended bool `json:"suspended,omitempty"`
	// If this account has been muted, when will the mute expire (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	MuteExpiresAt string `json:"mute_expires_at,omitempty"`
	// Extra profile information. Shown only if the requester owns the account being requested.
	Source *Source `json:"source,omitempty"`
}

// AccountCreateRequest represents the form submitted during a POST request to /api/v1/accounts.
//
// swagger:model accountCreateRequest
type AccountCreateRequest struct {
	// Text that will be reviewed by moderators if registrations require manual approval.
	Reason string `form:"reason" json:"reason" xml:"reason"`
	// The desired username for the account
	Username string `form:"username" json:"username" xml:"username" binding:"required"`
	// The email address to be used for login
	Email string `form:"email" json:"email" xml:"email" binding:"required"`
	// The password to be used for login
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
	// Whether the user agrees to the local rules, terms, and policies.
	// These should be presented to the user in order to allow them to consent before setting this parameter to TRUE.
	Agreement bool `form:"agreement"  json:"agreement" xml:"agreement" binding:"required"`
	// The language of the confirmation email that will be sent
	Locale string `form:"locale" json:"locale" xml:"locale" binding:"required"`
	// The IP of the sign up request, will not be parsed from the form but must be added manually
	IP net.IP `form:"-"`
}

// UpdateCredentialsRequest represents the form submitted during a PATCH request to /api/v1/accounts/update_credentials.
//
// swagger:model accountUpdateRequest
type UpdateCredentialsRequest struct {
	// Whether the account should be shown in the profile directory.
	Discoverable *bool `form:"discoverable" json:"discoverable" xml:"discoverable"`
	// Whether the account has a bot flag.
	Bot *bool `form:"bot" json:"bot" xml:"bot"`
	// The display name to use for the profile.
	DisplayName *string `form:"display_name" json:"display_name" xml:"display_name"`
	// The account bio.
	Note *string `form:"note" json:"note" xml:"note"`
	// Avatar image encoded using multipart/form-data
	Avatar *multipart.FileHeader `form:"avatar" json:"avatar" xml:"avatar"`
	// Header image encoded using multipart/form-data
	Header *multipart.FileHeader `form:"header" json:"header" xml:"header"`
	// Whether manual approval of follow requests is required.
	Locked *bool `form:"locked" json:"locked" xml:"locked"`
	// New Source values for this account
	Source *UpdateSource `form:"source" json:"source" xml:"source"`
	// Profile metadata name and value
	FieldsAttributes *[]UpdateField `form:"fields_attributes" json:"fields_attributes" xml:"fields_attributes"`
}

// UpdateSource is to be used specifically in an UpdateCredentialsRequest.
type UpdateSource struct {
	// Default post privacy for authored statuses.
	Privacy *string `form:"privacy" json:"privacy" xml:"privacy"`
	// Whether to mark authored statuses as sensitive by default.
	Sensitive *bool `form:"sensitive" json:"sensitive" xml:"sensitive"`
	// Default language to use for authored statuses. (ISO 6391)
	Language *string `form:"language" json:"language" xml:"language"`
}

// UpdateField is to be used specifically in an UpdateCredentialsRequest.
// By default, max 4 fields and 255 characters per property/value.
type UpdateField struct {
	// Name of the field
	Name *string `form:"name" json:"name" xml:"name"`
	// Value of the field
	Value *string `form:"value" json:"value" xml:"value"`
}

// AccountFollowRequest is for parsing requests at /api/v1/accounts/:id/follow
type AccountFollowRequest struct {
	// ID of the account to follow request
	// This should be a URL parameter not a form field
	TargetAccountID string `form:"-"`
	// Show reblogs for this account?
	Reblogs *bool `form:"reblogs" json:"reblogs" xml:"reblogs"`
	// Notify when this account posts?
	Notify *bool `form:"notify" json:"notify" xml:"notify"`
}
