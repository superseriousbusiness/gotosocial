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

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net"
	"strconv"
)

// Account models a fediverse account.
//
// The modelled account can be either a remote account, or one on this instance.
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
	// Account has opted into discovery features.
	Discoverable bool `json:"discoverable"`
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
	// Description of this account's avatar, for alt text.
	// example: A cute drawing of a smiling sloth.
	AvatarDescription string `json:"avatar_description,omitempty"`
	// Database ID of the media attachment for this account's avatar image.
	// Omitted if no avatar uploaded for this account (ie., default avatar).
	// example: 01JAJ3XCD66K3T99JZESCR137W
	AvatarMediaID string `json:"avatar_media_id,omitempty"`
	// Web location of the account's header image.
	// example: https://example.org/media/some_user/header/original/header.jpeg
	Header string `json:"header"`
	// Web location of a static version of the account's header.
	// Only relevant when the account's main header is a video or a gif.
	// example: https://example.org/media/some_user/header/static/header.png
	HeaderStatic string `json:"header_static"`
	// Description of this account's header, for alt text.
	// example: A sunlit field with purple flowers.
	HeaderDescription string `json:"header_description,omitempty"`
	// Database ID of the media attachment for this account's header image.
	// Omitted if no header uploaded for this account (ie., default header).
	// example: 01JAJ3XCD66K3T99JZESCR137W
	HeaderMediaID string `json:"header_media_id,omitempty"`
	// Number of accounts following this account, according to our instance.
	FollowersCount int `json:"followers_count"`
	// Number of account's followed by this account, according to our instance.
	FollowingCount int `json:"following_count"`
	// Number of statuses posted by this account, according to our instance.
	StatusesCount int `json:"statuses_count"`
	// When the account's most recent status was posted (ISO 8601 Date).
	// example: 2021-07-30
	LastStatusAt *string `json:"last_status_at"`
	// Array of custom emojis used in this account's note or display name.
	// Empty for blocked accounts.
	Emojis []Emoji `json:"emojis"`
	// Additional metadata attached to this account's profile.
	// Empty for blocked accounts.
	Fields []Field `json:"fields"`
	// Account has been suspended by our instance.
	Suspended bool `json:"suspended,omitempty"`
	// Extra profile information. Shown only if the requester owns the account being requested.
	Source *Source `json:"source,omitempty"`
	// Filename of user-selected CSS theme to include when rendering this account's profile or statuses. Eg., `blurple-light.css`.
	Theme string `json:"theme,omitempty"`
	// CustomCSS to include when rendering this account's profile or statuses.
	CustomCSS string `json:"custom_css,omitempty"`
	// Account has enabled RSS feed.
	// Key/value omitted if false.
	EnableRSS bool `json:"enable_rss,omitempty"`
	// Account has opted to hide their followers/following collections.
	// Key/value omitted if false.
	HideCollections bool `json:"hide_collections,omitempty"`
	// Role of the account on this instance.
	// Only available through the `verify_credentials` API method.
	// Key/value omitted for remote accounts.
	Role *AccountRole `json:"role,omitempty"`
	// Roles lists the public roles of the account on this instance.
	// Unlike Role, this is always available, but never includes permissions details.
	// Key/value omitted for remote accounts.
	Roles []AccountDisplayRole `json:"roles,omitempty"`
	// If set, indicates that this account is currently inactive, and has migrated to the given account.
	// Key/value omitted for accounts that haven't moved, and for suspended accounts.
	Moved *Account `json:"moved,omitempty"`
	// Account identifies as a Group actor.
	Group bool `json:"group"`
}

// WebAccount is like Account, but with
// additional fields not exposed via JSON;
// used only internally for templating etc.
//
// swagger:ignore
type WebAccount struct {
	*Account

	// Proper attachment model for the avatar.
	//
	// Only set if this account had an avatar set
	// (and not just the default "blank" image.)
	AvatarAttachment *WebAttachment `json:"-"`

	// Proper attachment model for the header.
	//
	// Only set if this account had a header set
	// (and not just the default "blank" image.)
	HeaderAttachment *WebAttachment `json:"-"`
}

// MutedAccount extends Account with a field used only by the muted user list.
//
// swagger:model mutedAccount
type MutedAccount struct {
	Account
	// If this account has been muted, when will the mute expire (ISO 8601 Datetime).
	// If the mute is indefinite, this will be null.
	// example: 2021-07-30T09:20:25+00:00
	MuteExpiresAt *string `json:"mute_expires_at"`
}

// AccountCreateRequest models account creation parameters.
//
// swagger:parameters accountCreate
type AccountCreateRequest struct {
	// Text that will be reviewed by moderators if registrations require manual approval.
	Reason string `form:"reason" json:"reason" xml:"reason"`
	// The desired username for the account.
	// swagger:parameters
	// pattern: [a-z0-9_]{2,64}
	// example: a_valid_username
	// required: true
	Username string `form:"username" json:"username" xml:"username" binding:"required"`
	// The email address to be used for login.
	// swagger:parameters
	// example: someone@wherever.com
	// required: true
	Email string `form:"email" json:"email" xml:"email" binding:"required"`
	// The password to be used for login. This will be hashed before storage.
	// swagger:parameters
	// example: some_really_really_really_strong_password
	// required: true
	Password string `form:"password" json:"password" xml:"password" binding:"required"`
	// The user agrees to the terms, conditions, and policies of the instance.
	// swagger:parameters
	// required: true
	Agreement bool `form:"agreement"  json:"agreement" xml:"agreement" binding:"required"`
	// The language of the confirmation email that will be sent.
	// swagger:parameters
	// example: en
	// Required: true
	Locale string `form:"locale" json:"locale" xml:"locale" binding:"required"`
	// The IP of the sign up request, will not be parsed from the form.
	// swagger:parameters
	// swagger:ignore
	IP net.IP `form:"-"`
}

// UpdateCredentialsRequest models an update to an account, by the account owner.
//
// swagger:ignore
type UpdateCredentialsRequest struct {
	// Account should be made discoverable and shown in the profile directory (if enabled).
	Discoverable *bool `form:"discoverable" json:"discoverable"`
	// Account is flagged as a bot.
	Bot *bool `form:"bot" json:"bot"`
	// The display name to use for the account.
	DisplayName *string `form:"display_name" json:"display_name"`
	// Bio/description of this account.
	Note *string `form:"note" json:"note"`
	// Avatar image encoded using multipart/form-data.
	Avatar *multipart.FileHeader `form:"avatar" json:"-"`
	// Description of the avatar image, for alt-text.
	AvatarDescription *string `form:"avatar_description" json:"avatar_description"`
	// Header image encoded using multipart/form-data
	Header *multipart.FileHeader `form:"header" json:"-"`
	// Description of the header image, for alt-text.
	HeaderDescription *string `form:"header_description" json:"header_description"`
	// Require manual approval of follow requests.
	Locked *bool `form:"locked" json:"locked"`
	// New Source values for this account.
	Source *UpdateSource `form:"source" json:"source"`
	// Profile metadata names and values.
	FieldsAttributes *[]UpdateField `form:"fields_attributes" json:"-"`
	// Profile metadata names and values, parsed from JSON.
	JSONFieldsAttributes *map[string]UpdateField `form:"-" json:"fields_attributes"`
	// Theme file name to be used when rendering this account's profile or statuses.
	// Use empty string to unset.
	Theme *string `form:"theme" json:"theme"`
	// Custom CSS to be included when rendering this account's profile or statuses.
	// Use empty string to unset.
	CustomCSS *string `form:"custom_css" json:"custom_css"`
	// Enable RSS feed of public toots for this account at /@[username]/feed.rss
	EnableRSS *bool `form:"enable_rss" json:"enable_rss"`
	// Hide this account's following/followers collections.
	HideCollections *bool `form:"hide_collections" json:"hide_collections"`
	// Visibility of statuses to show via the web view.
	// "none", "public" (default), or "unlisted" (which includes public as well).
	WebVisibility *string `form:"web_visibility" json:"web_visibility"`
}

// UpdateSource is to be used specifically in an UpdateCredentialsRequest.
//
// swagger:ignore
type UpdateSource struct {
	// Default post privacy for authored statuses.
	Privacy *string `form:"privacy" json:"privacy"`
	// Mark authored statuses as sensitive by default.
	Sensitive *bool `form:"sensitive" json:"sensitive"`
	// Default language to use for authored statuses. (ISO 6391)
	Language *string `form:"language" json:"language"`
	// Default format for authored statuses (text/plain or text/markdown).
	StatusContentType *string `form:"status_content_type" json:"status_content_type"`
}

// UpdateField is to be used specifically in an UpdateCredentialsRequest.
// By default, max 6 fields and 255 characters per property/value.
//
// swagger:ignore
type UpdateField struct {
	// Key this form field was submitted with;
	// only set if it was submitted as JSON.
	Key int `form:"-" json:"-"`
	// Name of the field
	Name *string `form:"name" json:"name"`
	// Value of the field
	Value *string `form:"value" json:"value"`
}

// AccountFollowRequest models a request to follow an account.
//
// swagger:ignore
type AccountFollowRequest struct {
	// The id of the account to follow.
	ID string `form:"-" json:"-" xml:"-"`
	// Show reblogs from this account.
	Reblogs *bool `form:"reblogs" json:"reblogs" xml:"reblogs"`
	// Notify when this account posts.
	Notify *bool `form:"notify" json:"notify" xml:"notify"`
}

// AccountDeleteRequest models a request to delete an account.
//
// swagger:ignore
type AccountDeleteRequest struct {
	// Password of the account's user, for confirmation.
	Password string `form:"password" json:"password" xml:"password"`
}

// AccountMoveRequest models a request to Move an account.
//
// swagger:ignore
type AccountMoveRequest struct {
	// Password of the account's user, for confirmation.
	Password string `form:"password" json:"password" xml:"password"`
	// ActivityPub URI of the account that's being moved to.
	MovedToURI string `form:"moved_to_uri" json:"moved_to_uri" xml:"moved_to_uri"`
}

// AccountAliasRequest models a request
// to set an account's alsoKnownAs URIs.
type AccountAliasRequest struct {
	// ActivityPub URIs of any accounts that this one is being aliased to.
	AlsoKnownAsURIs []string `form:"also_known_as_uris" json:"also_known_as_uris" xml:"also_known_as_uris"`
}

// AccountDisplayRole models a public, displayable role of an account.
// This is a subset of AccountRole.
//
// swagger:model accountDisplayRole
type AccountDisplayRole struct {
	// ID of the role.
	// Not used by GotoSocial, but we set it to the role name, just in case a client expects a unique ID.
	ID string `json:"id"`

	// Name of the role.
	Name AccountRoleName `json:"name"`

	// Color is a 6-digit CSS-style hex color code with leading `#`, or an empty string if this role has no color.
	// Since GotoSocial doesn't use role colors, we leave this empty.
	Color string `json:"color"`
}

// AccountRole models the role of an account.
//
// swagger:model accountRole
type AccountRole struct {
	AccountDisplayRole

	// Permissions is a bitmap serialized as a numeric string, indicating which admin/moderation actions a user can perform.
	Permissions AccountRolePermissions `json:"permissions"`

	// Highlighted indicates whether the role is publicly visible on the user profile.
	// This is always true for GotoSocial's built-in admin and moderator roles, and false otherwise.
	Highlighted bool `json:"highlighted"`
}

// AccountRoleName represent the name of the role of an account.
//
// swagger:type string
type AccountRoleName string

const (
	AccountRoleUser      AccountRoleName = "user"      // Standard user
	AccountRoleModerator AccountRoleName = "moderator" // Moderator privileges
	AccountRoleAdmin     AccountRoleName = "admin"     // Instance admin
	AccountRoleUnknown   AccountRoleName = ""          // We don't know / remote account
)

// AccountRolePermissions is a bitmap representing a set of user permissions.
// It's used for Mastodon API compatibility: internally, GotoSocial only tracks admins and moderators.
//
// swagger:type string
type AccountRolePermissions int

// MarshalJSON serializes an AccountRolePermissions as a numeric string.
func (a *AccountRolePermissions) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.Itoa(int(*a)))
}

// UnmarshalJSON deserializes an AccountRolePermissions from a number or numeric string.
func (a *AccountRolePermissions) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}

	i := 0
	if err := json.Unmarshal(b, &i); err != nil {
		s := ""
		if err := json.Unmarshal(b, &s); err != nil {
			return errors.New("not a number or string")
		}

		i, err = strconv.Atoi(s)
		if err != nil {
			return errors.New("not a numeric string")
		}
	}

	*a = AccountRolePermissions(i)
	return nil
}

const (
	// AccountRolePermissionsNone represents an empty set of permissions.
	AccountRolePermissionsNone AccountRolePermissions = 0
	// AccountRolePermissionsAdministrator ignores all permission checks.
	AccountRolePermissionsAdministrator AccountRolePermissions = 1 << (iota - 1)
	// AccountRolePermissionsDevops is not used by GotoSocial.
	AccountRolePermissionsDevops
	// AccountRolePermissionsViewAuditLog is not used by GotoSocial.
	AccountRolePermissionsViewAuditLog
	// AccountRolePermissionsViewDashboard is not used by GotoSocial.
	AccountRolePermissionsViewDashboard
	// AccountRolePermissionsManageReports indicates that the user can view and resolve reports.
	AccountRolePermissionsManageReports
	// AccountRolePermissionsManageFederation indicates that the user can edit federation allows and blocks.
	AccountRolePermissionsManageFederation
	// AccountRolePermissionsManageSettings indicates that the user can edit instance metadata.
	AccountRolePermissionsManageSettings
	// AccountRolePermissionsManageBlocks indicates that the user can manage non-federation blocks, currently including HTTP header blocks.
	AccountRolePermissionsManageBlocks
	// AccountRolePermissionsManageTaxonomies is not used by GotoSocial.
	AccountRolePermissionsManageTaxonomies
	// AccountRolePermissionsManageAppeals is not used by GotoSocial.
	AccountRolePermissionsManageAppeals
	// AccountRolePermissionsManageUsers indicates that the user can view user details and perform user moderation actions.
	AccountRolePermissionsManageUsers
	// AccountRolePermissionsManageInvites is not used by GotoSocial.
	AccountRolePermissionsManageInvites
	// AccountRolePermissionsManageRules indicates that the user can edit instance rules.
	AccountRolePermissionsManageRules
	// AccountRolePermissionsManageAnnouncements is not used by GotoSocial.
	AccountRolePermissionsManageAnnouncements
	// AccountRolePermissionsManageCustomEmojis indicates that the user can edit custom emoji.
	AccountRolePermissionsManageCustomEmojis
	// AccountRolePermissionsManageWebhooks is not used by GotoSocial.
	AccountRolePermissionsManageWebhooks
	// AccountRolePermissionsInviteUsers is not used by GotoSocial.
	AccountRolePermissionsInviteUsers
	// AccountRolePermissionsManageRoles is not used by GotoSocial.
	AccountRolePermissionsManageRoles
	// AccountRolePermissionsManageUserAccess is not used by GotoSocial.
	AccountRolePermissionsManageUserAccess
	// AccountRolePermissionsDeleteUserData indicates that the user can permanently delete user data.
	AccountRolePermissionsDeleteUserData

	// FUTURE: If we decide to assign our own permissions bits,
	// they should start at the other end of the int to avoid conflicts with future Mastodon permissions.

	// AccountRolePermissionsForAdminRole includes all of the permissions assigned to GotoSocial's built-in administrator role.
	AccountRolePermissionsForAdminRole = AccountRolePermissionsAdministrator |
		AccountRolePermissionsManageReports |
		AccountRolePermissionsManageFederation |
		AccountRolePermissionsManageSettings |
		AccountRolePermissionsManageBlocks |
		AccountRolePermissionsManageUsers |
		AccountRolePermissionsManageRules |
		AccountRolePermissionsManageCustomEmojis |
		AccountRolePermissionsDeleteUserData

	// AccountRolePermissionsForModeratorRole includes all of the permissions assigned to GotoSocial's built-in moderator role.
	// (Currently, there aren't any.)
	AccountRolePermissionsForModeratorRole = AccountRolePermissionsNone
)

// AccountNoteRequest models a request to update the private note for an account.
//
// swagger:ignore
type AccountNoteRequest struct {
	// Comment to use for the note text.
	Comment string `form:"comment" json:"comment" xml:"comment"`
}
