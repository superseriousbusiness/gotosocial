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

import "encoding/xml"

// WellKnownResponse represents the response to either a webfinger request for an 'acct' resource, or a request to nodeinfo.
// For example, it would be returned from https://example.org/.well-known/webfinger?resource=acct:some_username@example.org
//
// See https://webfinger.net/
//
// swagger:model wellKnownResponse
type WellKnownResponse struct {
	Subject string   `json:"subject,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
	Links   []Link   `json:"links,omitempty"`
}

// Link represents one 'link' in a slice of links returned from a lookup request.
//
// See https://webfinger.net/ and https://www.rfc-editor.org/rfc/rfc6415.html#section-3.1
type Link struct {
	Rel      string `json:"rel" xml:"rel,attr"`
	Type     string `json:"type,omitempty" xml:"type,attr,omitempty"`
	Href     string `json:"href,omitempty" xml:"href,attr,omitempty"`
	Template string `json:"template,omitempty" xml:"template,attr,omitempty"`
}

// Nodeinfo represents a version 2.1 or version 2.0 nodeinfo schema.
// See: https://nodeinfo.diaspora.software/schema.html
//
// swagger:model nodeinfo
type Nodeinfo struct {
	// The schema version
	// example: 2.0
	Version string `json:"version"`
	// Metadata about server software in use.
	Software NodeInfoSoftware `json:"software"`
	// The protocols supported on this server.
	Protocols []string `json:"protocols"`
	// The third party sites this server can connect to via their application API.
	Services NodeInfoServices `json:"services"`
	// Whether this server allows open self-registration.
	// example: false
	OpenRegistrations bool `json:"openRegistrations"`
	// Usage statistics for this server.
	Usage NodeInfoUsage `json:"usage"`
	// Free form key value pairs for software specific values. Clients should not rely on any specific key present.
	Metadata map[string]interface{} `json:"metadata"`
}

// NodeInfoSoftware represents the name and version number of the software of this node.
type NodeInfoSoftware struct {
	// example: gotosocial
	Name string `json:"name"`
	// example: 0.1.2 1234567
	Version string `json:"version"`
	// Repository for the software. Omitted in version 2.0.
	// example: https://codeberg.org/superseriousbusiness/gotosocial
	Repository string `json:"repository,omitempty"`
	// Homepage for the software. Omitted in version 2.0.
	// example: https://docs.gotosocial.org
	Homepage string `json:"homepage,omitempty"`
}

// NodeInfoServices represents inbound and outbound services that this node offers connections to.
type NodeInfoServices struct {
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

// NodeInfoUsage represents usage information about this server, such as number of users.
type NodeInfoUsage struct {
	Users         NodeInfoUsers `json:"users"`
	LocalPosts    int           `json:"localPosts,omitempty"`
	LocalComments int           `json:"localComments,omitempty"`
}

// NodeInfoUsers represents aggregate information about the users on the server.
type NodeInfoUsers struct {
	Total          int `json:"total"`
	ActiveHalfYear int `json:"activeHalfYear,omitempty"`
	ActiveMonth    int `json:"activeMonth,omitempty"`
}

// HostMeta represents a hostmeta document.
// See: https://www.rfc-editor.org/rfc/rfc6415.html#section-3
//
// swagger:model hostmeta
type HostMeta struct {
	XMLName xml.Name `xml:"XRD"`
	XMLNS   string   `xml:"xmlns,attr"`
	Link    []Link   `xml:"Link"`
}
