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

// Instance represents the software instance of Mastodon running on this domain. See https://docs.joinmastodon.org/entities/instance/
type Instance struct {
	// REQUIRED

	// The domain name of the instance.
	URI string `json:"uri"`
	// The title of the website.
	Title string `json:"title"`
	// Admin-defined description of the Mastodon site.
	Description string `json:"description"`
	// A shorter description defined by the admin.
	ShortDescription string `json:"short_description"`
	// An email that may be contacted for any inquiries.
	Email string `json:"email"`
	// The version of Mastodon installed on the instance.
	Version string `json:"version"`
	// Primary langauges of the website and its staff.
	Languages []string `json:"languages"`
	// Whether registrations are enabled.
	Registrations bool `json:"registrations"`
	// Whether registrations require moderator approval.
	ApprovalRequired bool `json:"approval_required"`
	// Whether invites are enabled.
	InvitesEnabled bool `json:"invites_enabled"`
	// URLs of interest for clients apps.
	URLS *InstanceURLs `json:"urls"`
	// Statistics about how much information the instance contains.
	Stats *InstanceStats `json:"stats"`

	// OPTIONAL

	// Banner image for the website.
	Thumbnail string `json:"thumbnail,omitempty"`
	// A user that can be contacted, as an alternative to email.
	ContactAccount *Account `json:"contact_account,omitempty"`
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
