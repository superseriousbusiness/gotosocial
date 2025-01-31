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
	"mime/multipart"
	"time"
)

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
	// Custom CSS for the instance.
	CustomCSS *string `form:"custom_css" json:"custom_css,omitempty" xml:"custom_css"`
	// Terms and conditions of the instance, max 5,000 chars. HTML formatting accepted.
	Terms *string `form:"terms" json:"terms" xml:"terms"`
	// Image to use as the instance thumbnail.
	Avatar *multipart.FileHeader `form:"thumbnail" json:"thumbnail" xml:"thumbnail"`
	// Image description for the instance avatar.
	AvatarDescription *string `form:"thumbnail_description" json:"thumbnail_description" xml:"thumbnail_description"`
	// Image to use as the instance header.
	Header *multipart.FileHeader `form:"header" json:"header" xml:"header"`
}

// InstanceConfigurationAccounts models instance account config parameters.
//
// swagger:model instanceConfigurationAccounts
type InstanceConfigurationAccounts struct {
	// Whether or not accounts on this instance are allowed to upload custom CSS for profiles and statuses.
	//
	// example: false
	AllowCustomCSS bool `json:"allow_custom_css"`
	// The maximum number of featured tags allowed for each account.
	// Currently not implemented, so this is hardcoded to 10.
	MaxFeaturedTags int `json:"max_featured_tags"`
	// The maximum number of profile fields allowed for each account.
	// Currently not configurable, so this is hardcoded to 6. (https://github.com/superseriousbusiness/gotosocial/issues/1876)
	MaxProfileFields int `json:"max_profile_fields"`
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
	// List of mime types that it's possible to use for statuses on this instance.
	//
	// example: ["text/plain","text/markdown"]
	SupportedMimeTypes []string `json:"supported_mime_types,omitempty"`
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

// InstanceConfigurationEmojis models instance emoji config parameters.
type InstanceConfigurationEmojis struct {
	// Max allowed emoji image size in bytes.
	//
	// example: 51200
	EmojiSizeLimit int `json:"emoji_size_limit"`
}

// swagger:ignore
type RandomStats struct {
	Statuses           int64
	TotalUsers         int64
	MonthlyActiveUsers int64
	Generated          time.Time
}
