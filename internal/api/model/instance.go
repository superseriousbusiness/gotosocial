/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

import "mime/multipart"

// Instance models information about this or another instance.
//
// swagger:model instance
type Instance struct {
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
	Configuration *InstanceConfiguration `json:"configuration,omitempty"`
	// URLs of interest for client applications.
	URLS *InstanceURLs `json:"urls,omitempty"`
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

// InstanceConfiguration models instance configuration parameters.
//
// swagger:model instanceConfiguration
type InstanceConfiguration struct {
	// Instance configuration pertaining to status limits.
	Statuses *InstanceConfigurationStatuses `json:"statuses"`
	// Instance configuration pertaining to media attachment types + size limits.
	MediaAttachments *InstanceConfigurationMediaAttachments `json:"media_attachments"`
	// Instance configuration pertaining to poll limits.
	Polls *InstanceConfigurationPolls `json:"polls"`
	// Instance configuration pertaining to accounts.
	Accounts *InstanceConfigurationAccounts `json:"accounts"`
	// Instance configuration pertaining to emojis.
	Emojis *InstanceConfigurationEmojis `json:"emojis"`
}

// InstanceConfigurationStatuses models instance status config parameters.
//
// swagger:model instanceConfigurationStatuses
type InstanceConfigurationStatuses struct {
	// Maximum allowed length of a post on this instance, in characters.
	//
	// example: 5000
	MaxCharacters int `json:"max_characters"`
	// Max number of attachments allowed on a status.
	//
	// example: 4
	MaxMediaAttachments int `json:"max_media_attachments"`
	// Amount of characters clients should assume a url takes up.
	//
	// example: 25
	CharactersReservedPerURL int `json:"characters_reserved_per_url"`
}

// InstanceConfigurationMediaAttachments models instance media attachment config parameters.
//
// swagger:model instanceConfigurationMediaAttachments
type InstanceConfigurationMediaAttachments struct {
	// List of mime types that it's possible to upload to this instance.
	//
	// example: ["image/jpeg","image/gif"]
	SupportedMimeTypes []string `json:"supported_mime_types"`
	// Max allowed image size in bytes
	//
	// example: 2097152
	ImageSizeLimit int `json:"image_size_limit"`
	// Max allowed image size in pixels as height*width.
	//
	// GtS doesn't set a limit on this, but for compatibility
	// we give Mastodon's 4096x4096px value here.
	//
	// example: 16777216
	ImageMatrixLimit int `json:"image_matrix_limit"`
	// Max allowed video size in bytes
	//
	// example: 10485760
	VideoSizeLimit int `json:"video_size_limit"`
	// Max allowed video frame rate.
	//
	// example: 60
	VideoFrameRateLimit int `json:"video_frame_rate_limit"`
	// Max allowed video size in pixels as height*width.
	//
	// GtS doesn't set a limit on this, but for compatibility
	// we give Mastodon's 4096x4096px value here.
	//
	// example: 16777216
	VideoMatrixLimit int `json:"video_matrix_limit"`
}

// InstanceConfigurationPolls models instance poll config parameters.
//
// swagger:model instanceConfigurationPolls
type InstanceConfigurationPolls struct {
	// Number of options permitted in a poll on this instance.
	//
	// example: 4
	MaxOptions int `json:"max_options"`
	// Number of characters allowed per option in the poll.
	//
	// example: 50
	MaxCharactersPerOption int `json:"max_characters_per_option"`
	// Minimum expiration time of the poll in seconds.
	//
	// example: 300
	MinExpiration int `json:"min_expiration"`
	// Maximum expiration time of the poll in seconds.
	//
	// example: 2629746
	MaxExpiration int `json:"max_expiration"`
}

// InstanceConfigurationAccounts models instance account config parameters.
type InstanceConfigurationAccounts struct {
	// Whether or not accounts on this instance are allowed to upload custom CSS for profiles and statuses.
	//
	// example: false
	AllowCustomCSS bool `json:"allow_custom_css"`
}

// InstanceConfigurationEmojis models instance emoji config parameters.
type InstanceConfigurationEmojis struct {
	// Max allowed emoji image size in bytes.
	//
	// example: 51200
	EmojiSizeLimit int `json:"emoji_size_limit"`
}

// InstanceURLs models instance-relevant URLs for client application consumption.
//
// swagger:model instanceURLs
type InstanceURLs struct {
	// Websockets address for status and notification streaming.
	// example: wss://example.org
	StreamingAPI string `json:"streaming_api"`
}

// InstanceSettingsUpdateRequest models an instance update request.
//
// swagger:ignore
type InstanceSettingsUpdateRequest struct {
	// Title to use for the instance. Max 40 characters.
	Title *string `form:"title" json:"title" xml:"title"`
	// Username for the instance contact account. Must be the username of an existing admin.
	ContactUsername *string `form:"contact_username" json:"contact_username" xml:"contact_username"`
	// Email for reaching the instance administrator(s).
	ContactEmail *string `form:"contact_email" json:"contact_email" xml:"contact_email"`
	// Short description of the instance, max 500 chars. HTML formatting accepted.
	ShortDescription *string `form:"short_description" json:"short_description" xml:"short_description"`
	// Longer description of the instance, max 5,000 chars. HTML formatting accepted.
	Description *string `form:"description" json:"description" xml:"description"`
	// Terms and conditions of the instance, max 5,000 chars. HTML formatting accepted.
	Terms *string `form:"terms" json:"terms" xml:"terms"`
	// Image to use as the instance thumbnail.
	Avatar *multipart.FileHeader `form:"thumbnail" json:"thumbnail" xml:"thumbnail"`
	// Image description for the instance avatar.
	AvatarDescription *string `form:"thumbnail_description" json:"thumbnail_description" xml:"thumbnail_description"`
	// Image to use as the instance header.
	Header *multipart.FileHeader `form:"header" json:"header" xml:"header"`
}
