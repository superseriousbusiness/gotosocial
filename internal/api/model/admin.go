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

// AdminAccountInfo models the admin view of an account's details.
type AdminAccountInfo struct {
	// The ID of the account in the database.
	ID string `json:"id"`
	// The username of the account.
	Username string `json:"username"`
	// The domain of the account.
	Domain string `json:"domain"`
	// When the account was first discovered. (ISO 8601 Datetime)
	CreatedAt string `json:"created_at"`
	// The email address associated with the account.
	Email string `json:"email"`
	// The IP address last used to login to this account.
	IP string `json:"ip"`
	// The locale of the account. (ISO 639 Part 1 two-letter language code)
	Locale string `json:"locale"`
	// Invite request text
	InviteRequest string `json:"invite_request"`
	// The current role of the account.
	Role string `json:"role"`
	// Whether the account has confirmed their email address.
	Confirmed bool `json:"confirmed"`
	// Whether the account is currently approved.
	Approved bool `json:"approved"`
	// Whether the account is currently disabled.
	Disabled bool `json:"disabled"`
	// Whether the account is currently silenced
	Silenced bool `json:"silenced"`
	// Whether the account is currently suspended.
	Suspended bool `json:"suspended"`
	// User-level information about the account.
	Account *Account `json:"account"`
	// The ID of the application that created this account.
	CreatedByApplicationID string `json:"created_by_application_id,omitempty"`
	// The ID of the account that invited this user
	InvitedByAccountID string `json:"invited_by_account_id"`
}

// AdminReportInfo models the admin view of a report.
type AdminReportInfo struct {
	// The ID of the report in the database.
	ID string `json:"id"`
	// The action taken to resolve this report.
	ActionTaken string `json:"action_taken"`
	// An optional reason for reporting.
	Comment string `json:"comment"`
	// The time the report was filed. (ISO 8601 Datetime)
	CreatedAt string `json:"created_at"`
	// The time of last action on this report. (ISO 8601 Datetime)
	UpdatedAt string `json:"updated_at"`
	// The account which filed the report.
	Account *Account `json:"account"`
	// The account being reported.
	TargetAccount *Account `json:"target_account"`
	// The account of the moderator assigned to this report.
	AssignedAccount *Account `json:"assigned_account"`
	// The action taken by the moderator who handled the report.
	ActionTakenByAccount string `json:"action_taken_by_account"`
	// Statuses attached to the report, for context.
	Statuses []Status `json:"statuses"`
}

// AdminEmoji models the admin view of a custom emoji.
//
// swagger:model adminEmoji
type AdminEmoji struct {
	Emoji
	// The ID of the emoji.
	// example: 01GEM7SFDZ7GZNRXFVZ3X4E4N1
	ID string `json:"id"`
	// True if this emoji has been disabled by an admin action.
	// example: false
	Disabled bool `json:"disabled"`
	// The domain from which the emoji originated. Only defined for remote domains, otherwise key will not be set.
	//
	// example: example.org
	Domain string `json:"domain,omitempty"`
	// Time when the emoji image was last updated.
	// example: 2022-10-05T09:21:26.419Z
	UpdatedAt string `json:"updated_at"`
	// The total file size taken up by the emoji in bytes, including static and animated versions.
	// example: 69420
	TotalFileSize int `json:"total_file_size"`
	// The MIME content type of the emoji.
	// example: image/png
	ContentType string `json:"content_type"`
	// The ActivityPub URI of the emoji.
	// example: https://example.org/emojis/016T5Q3SQKBT337DAKVSKNXXW1
	URI string `json:"uri"`
}

// AdminAccountActionRequest models the admin view of an account's details.
//
// swagger:ignore
type AdminAccountActionRequest struct {
	// Type of the account action. One of disable, silence, suspend.
	Type string `form:"type" json:"type" xml:"type"`
	// Text describing why an action was taken.
	Text string `form:"text" json:"text" xml:"text"`
	// ID of the account to be acted on.
	TargetAccountID string `form:"-" json:"-" xml:"-"`
}

// MediaCleanupRequest models admin media cleanup parameters
//
// swagger:parameters mediaCleanup
type MediaCleanupRequest struct {
	// Number of days of remote media to keep. Native values will be treated as 0.
	// If value is not specified, the value of media-remote-cache-days in the server config will be used.
	RemoteCacheDays *int `form:"remote_cache_days" json:"remote_cache_days" xml:"remote_cache_days"`
}
