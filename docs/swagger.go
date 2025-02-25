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
//	      read: grants read access to everything
//	      write: grants write access to everything
//	      push: grants read/write access to push
//	      profile: grants read access to verify_credentials
//	      read:accounts: grants read access to accounts
//	      write:accounts: grants write access to accounts
//	      read:blocks: grants read access to blocks
//	      write:blocks: grants write access to blocks
//	      read:bookmarks: grants read access to bookmarks
//	      write:bookmarks: grants write access to bookmarks
//	      write:conversations: grants write access to conversations
//	      read:favourites: grants read access to accounts
//	      write:favourites: grants write access to favourites
//	      read:filters: grants read access to filters
//	      write:filters: grants write access to filters
//	      read:follows: grants read access to follows
//	      write:follows: grants write access to follows
//	      read:lists: grants read access to lists
//	      write:lists: grants write access to lists
//	      write:media: grants write access to media
//	      read:mutes: grants read access to mutes
//	      write:mutes: grants write access to mutes
//	      read:notifications: grants read access to notifications
//	      write:notifications: grants write access to notifications
//	      write:reports: grants write access to reports
//	      read:search: grants read access to search
//	      read:statuses: grants read access to statuses
//	      write:statuses: grants write access to statuses
//	      admin: grants admin access to everything
//	      admin:read: grants admin read access to everything
//	      admin:write: grants admin write access to everything
//	      admin:read:accounts: grants admin read access to accounts
//	      admin:write:accounts: grants write read access to accounts
//	      admin:read:reports: grants admin read access to reports
//	      admin:write:reports: grants admin write access to reports
//	      admin:read:domain_allows: grants admin read access to domain_allows
//	      admin:write:domain_allows: grants admin write access to domain_allows
//	      admin:read:domain_blocks: grants admin read access to domain_blocks
//	      admin:write:domain_blocks: grants write read access to domain_blocks
//	  OAuth2 Application:
//	    type: oauth2
//	    flow: application
//	    tokenUrl: https://example.org/oauth/token
//	    scopes:
//	      write:accounts: grants write access to accounts
//
// swagger:meta
package docs
