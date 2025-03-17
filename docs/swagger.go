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

// GoToSocial Swagger documentation.
//
// This document describes the GoToSocial HTTP API.
//
// For information on how to authenticate with the API using an OAuth access token, see the documentation here: https://docs.gotosocial.org/en/latest/api/authentication/.
//
// Available scopes are:
//
//   - admin: grants admin access to everything
//   - admin:read: grants admin read access to everything
//   - admin:read:accounts: grants admin read access to accounts
//   - admin:read:domain_allows: grants admin read access to domain_allows
//   - admin:read:domain_blocks: grants admin read access to domain_blocks
//   - admin:read:reports: grants admin read access to reports
//   - admin:write: grants admin write access to everything
//   - admin:write:accounts: grants write read access to accounts
//   - admin:write:domain_allows: grants admin write access to domain_allows
//   - admin:write:domain_blocks: grants write read access to domain_blocks
//   - admin:write:reports: grants admin write access to reports
//   - profile: grants read access to verify_credentials
//   - push: grants read/write access to push
//   - read: grants read access to everything
//   - read:accounts: grants read access to accounts
//   - read:applications: grants read access to user-managed applications
//   - read:blocks: grants read access to blocks
//   - read:bookmarks: grants read access to bookmarks
//   - read:favourites: grants read access to accounts
//   - read:filters: grants read access to filters
//   - read:follows: grants read access to follows
//   - read:lists: grants read access to lists
//   - read:mutes: grants read access to mutes
//   - read:notifications: grants read access to notifications
//   - read:search: grants read access to search
//   - read:statuses: grants read access to statuses
//   - write: grants write access to everything
//   - write:accounts: grants write access to accounts
//   - write:applications: grants write access to user-managed applications
//   - write:blocks: grants write access to blocks
//   - write:bookmarks: grants write access to bookmarks
//   - write:conversations: grants write access to conversations
//   - write:favourites: grants write access to favourites
//   - write:filters: grants write access to filters
//   - write:follows: grants write access to follows
//   - write:lists: grants write access to lists
//   - write:media: grants write access to media
//   - write:mutes: grants write access to mutes
//   - write:notifications: grants write access to notifications
//   - write:reports: grants write access to reports
//   - write:statuses: grants write access to statuses
//
// ---
//
//	Schemes: https, http
//	BasePath: /
//	Version: REPLACE_ME
//	Host: example.org
//	License: AGPL3 https://www.gnu.org/licenses/agpl-3.0.en.html
//	Contact: GoToSocial Authors <admin@gotosocial.org>
//
//	SecurityDefinitions:
//	  OAuth2 Bearer:
//	    type: oauth2
//	    flow: accessCode
//	    authorizationUrl: https://example.org/oauth/authorize
//	    tokenUrl: https://example.org/oauth/token
//	    scopes:
//	      admin: grants admin access to everything
//	      admin:read: grants admin read access to everything
//	      admin:read:accounts: grants admin read access to accounts
//	      admin:read:domain_allows: grants admin read access to domain_allows
//	      admin:read:domain_blocks: grants admin read access to domain_blocks
//	      admin:read:reports: grants admin read access to reports
//	      admin:write: grants admin write access to everything
//	      admin:write:accounts: grants write read access to accounts
//	      admin:write:domain_allows: grants admin write access to domain_allows
//	      admin:write:domain_blocks: grants write read access to domain_blocks
//	      admin:write:reports: grants admin write access to reports
//	      profile: grants read access to verify_credentials
//	      push: grants read/write access to push
//	      read: grants read access to everything
//	      read:accounts: grants read access to accounts
//	      read:applications: grants read access to user-managed applications
//	      read:blocks: grants read access to blocks
//	      read:bookmarks: grants read access to bookmarks
//	      read:favourites: grants read access to accounts
//	      read:filters: grants read access to filters
//	      read:follows: grants read access to follows
//	      read:lists: grants read access to lists
//	      read:mutes: grants read access to mutes
//	      read:notifications: grants read access to notifications
//	      read:search: grants read access to search
//	      read:statuses: grants read access to statuses
//	      write: grants write access to everything
//	      write:accounts: grants write access to accounts
//	      write:applications: grants write access to user-managed applications
//	      write:blocks: grants write access to blocks
//	      write:bookmarks: grants write access to bookmarks
//	      write:conversations: grants write access to conversations
//	      write:favourites: grants write access to favourites
//	      write:filters: grants write access to filters
//	      write:follows: grants write access to follows
//	      write:lists: grants write access to lists
//	      write:media: grants write access to media
//	      write:mutes: grants write access to mutes
//	      write:notifications: grants write access to notifications
//	      write:reports: grants write access to reports
//	      write:statuses: grants write access to statuses
//	  OAuth2 Application:
//	    type: oauth2
//	    flow: application
//	    tokenUrl: https://example.org/oauth/token
//	    scopes:
//	      write:accounts: grants write access to accounts
//
// swagger:meta
package docs
