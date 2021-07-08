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

import "mime/multipart"

// Instance represents the software instance of Mastodon running on this domain. See https://docs.joinmastodon.org/entities/instance/
type Instance struct {
	// REQUIRED

	// The domain name of the instance.
	URI string `json:"uri,omitempty"`
	// The title of the website.
	Title string `json:"title,omitempty"`
	// Admin-defined description of the Mastodon site.
	Description string `json:"description"`
	// A shorter description defined by the admin.
	ShortDescription string `json:"short_description"`
	// An email that may be contacted for any inquiries.
	Email string `json:"email"`
	// The version of Mastodon installed on the instance.
	Version string `json:"version"`
	// Primary langauges of the website and its staff.
	Languages []string `json:"languages,omitempty"`
	// Whether registrations are enabled.
	Registrations bool `json:"registrations"`
	// Whether registrations require moderator approval.
	ApprovalRequired bool `json:"approval_required"`
	// Whether invites are enabled.
	InvitesEnabled bool `json:"invites_enabled"`
	// URLs of interest for clients apps.
	URLS *InstanceURLs `json:"urls,omitempty"`
	// Statistics about how much information the instance contains.
	Stats map[string]int `json:"stats,omitempty"`
	// Banner image for the website.
	Thumbnail string `json:"thumbnail"`
	// A user that can be contacted, as an alternative to email.
	ContactAccount *Account `json:"contact_account,omitempty"`
	// What's the maximum allowed length of a post on this instance?
	// This is provided for compatibility with Tusky.
	MaxTootChars uint `json:"max_toot_chars"`
}

// InstanceURLs represents URLs necessary for successfully connecting to the instance as a user. See https://docs.joinmastodon.org/entities/instance/
type InstanceURLs struct {
	// Websockets address for push streaming.
	StreamingAPI string `json:"streaming_api"`
}

// InstanceStats represents some public-facing stats about the instance. See https://docs.joinmastodon.org/entities/instance/
type InstanceStats struct {
	// Users registered on this instance.
	UserCount int `json:"user_count"`
	// Statuses authored by users on instance.
	StatusCount int `json:"status_count"`
	// Domains federated with this instance.
	DomainCount int `json:"domain_count"`
}

// InstanceSettingsUpdateRequest is the form to be parsed on a PATCH to /api/v1/instance
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
