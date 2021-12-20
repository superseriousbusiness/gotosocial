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
	// example: https://example.org
	URI string `json:"uri,omitempty"`
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
	// URLs of interest for client applications.
	URLS *InstanceURLs `json:"urls,omitempty"`
	// Statistics about the instance: number of posts, accounts, etc.
	Stats map[string]int `json:"stats,omitempty"`
	// URL of the instance avatar/banner image.
	// example: https://example.org/files/instance/thumbnail.jpeg
	Thumbnail string `json:"thumbnail"`
	// Contact account for the instance.
	ContactAccount *Account `json:"contact_account,omitempty"`
	// Maximum allowed length of a post on this instance, in characters.
	//
	// This is provided for compatibility with Tusky and other apps.
	//
	// example: 5000
	MaxTootChars uint `json:"max_toot_chars"`
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
	Avatar *multipart.FileHeader `form:"avatar" json:"avatar" xml:"avatar"`
	// Image to use as the instance header.
	Header *multipart.FileHeader `form:"header" json:"header" xml:"header"`
}
