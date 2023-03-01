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

// InstanceV1 models information about this instance.
//
// swagger:model instanceV1
type InstanceV1 struct {
	// The URI of the instance.
	// example: https://gts.example.org
	URI string `json:"uri,omitempty"`
	// The domain of accounts on this instance.
	// This will not necessarily be the same as
	// simply the Host part of the URI.
	// example: example.org
	AccountDomain string `json:"account_domain,omitempty"`
	// The title of the instance.
	// example: GoToSocial Example Instance
	Title string `json:"title,omitempty"`
	// Description of the instance.
	//
	// Should be HTML formatted, but might be plaintext.
	//
	// This should be displayed on the 'about' page for an instance.
	Description string `json:"description"`
	// A shorter description of the instance.
	//
	// Should be HTML formatted, but might be plaintext.
	//
	// This should be displayed on the instance splash/landing page.
	ShortDescription string `json:"short_description"`
	// An email address that may be used for inquiries.
	// example: admin@example.org
	Email string `json:"email"`
	// The version of GoToSocial installed on the instance.
	//
	// This will contain at least a semantic version number.
	//
	// It may also contain, after a space, the short git commit ID of the running software.
	//
	// example: 0.1.1 cb85f65
	Version string `json:"version"`
	// Primary language of the instance.
	// example: en
	Languages []string `json:"languages,omitempty"`
	// New account registrations are enabled on this instance.
	Registrations bool `json:"registrations"`
	// New account registrations require admin approval.
	ApprovalRequired bool `json:"approval_required"`
	// Invites are enabled on this instance.
	InvitesEnabled bool `json:"invites_enabled"`
	// Configuration object containing values about status limits etc.
	// This key/value will be omitted for remote instances.
	Configuration InstanceV1Configuration `json:"configuration,omitempty"`
	// URLs of interest for client applications.
	URLs InstanceV1URLs `json:"urls,omitempty"`
	// Statistics about the instance: number of posts, accounts, etc.
	Stats map[string]int `json:"stats,omitempty"`
	// URL of the instance avatar/banner image.
	// example: https://example.org/files/instance/thumbnail.jpeg
	Thumbnail string `json:"thumbnail"`
	// MIME type of the instance thumbnail.
	// example: image/png
	ThumbnailType string `json:"thumbnail_type,omitempty"`
	// Description of the instance thumbnail.
	// example: picture of a cute lil' friendly sloth
	ThumbnailDescription string `json:"thumbnail_description,omitempty"`
	// Contact account for the instance.
	ContactAccount *Account `json:"contact_account,omitempty"`
	// Maximum allowed length of a post on this instance, in characters.
	//
	// This is provided for compatibility with Tusky and other apps.
	//
	// example: 5000
	MaxTootChars uint `json:"max_toot_chars"`
}

// InstanceV1URLs models instance-relevant URLs for client application consumption.
//
// swagger:model instanceV1URLs
type InstanceV1URLs struct {
	// Websockets address for status and notification streaming.
	// example: wss://example.org
	StreamingAPI string `json:"streaming_api"`
}

// InstanceV1Configuration models instance configuration parameters.
//
// swagger:model instanceV1Configuration
type InstanceV1Configuration struct {
	// Instance configuration pertaining to status limits.
	Statuses InstanceConfigurationStatuses `json:"statuses"`
	// Instance configuration pertaining to media attachment types + size limits.
	MediaAttachments InstanceConfigurationMediaAttachments `json:"media_attachments"`
	// Instance configuration pertaining to poll limits.
	Polls InstanceConfigurationPolls `json:"polls"`
	// Instance configuration pertaining to accounts.
	Accounts InstanceConfigurationAccounts `json:"accounts"`
	// Instance configuration pertaining to emojis.
	Emojis InstanceConfigurationEmojis `json:"emojis"`
}
