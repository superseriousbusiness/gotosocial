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
// See https://webfinger.net/
type Link struct {
	Rel      string `json:"rel"`
	Type     string `json:"type,omitempty"`
	Href     string `json:"href,omitempty"`
	Template string `json:"template,omitempty"`
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
}

// NodeInfoServices represents inbound and outbound services that this node offers connections to.
type NodeInfoServices struct {
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

// NodeInfoUsage represents usage information about this server, such as number of users.
type NodeInfoUsage struct {
	Users NodeInfoUsers `json:"users"`
}

// NodeInfoUsers is a stub for usage information, currently empty.
type NodeInfoUsers struct{}
