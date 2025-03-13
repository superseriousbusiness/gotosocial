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

package util

import (
	"strings"
)

type Scope string

const (
	/* Sub-scopes / scope components */

	scopeAccounts      = "accounts"
	scopeApplications  = "applications"
	scopeBlocks        = "blocks"
	scopeBookmarks     = "bookmarks"
	scopeConversations = "conversations"
	scopeDomainAllows  = "domain_allows"
	scopeDomainBlocks  = "domain_blocks"
	scopeFavourites    = "favourites"
	scopeFilters       = "filters"
	scopeFollows       = "follows"
	scopeLists         = "lists"
	scopeMedia         = "media"
	scopeMutes         = "mutes"
	scopeNotifications = "notifications"
	scopeReports       = "reports"
	scopeSearch        = "search"
	scopeStatuses      = "statuses"

	/* Top-level scopes */

	ScopeProfile    Scope = "profile"
	ScopePush       Scope = "push"
	ScopeRead       Scope = "read"
	ScopeWrite      Scope = "write"
	ScopeAdmin      Scope = "admin"
	ScopeAdminRead  Scope = ScopeAdmin + ":" + ScopeRead
	ScopeAdminWrite Scope = ScopeAdmin + ":" + ScopeWrite

	/* Granular scopes */

	ScopeReadAccounts           Scope = ScopeRead + ":" + scopeAccounts
	ScopeWriteAccounts          Scope = ScopeWrite + ":" + scopeAccounts
	ScopeReadApplications       Scope = ScopeRead + ":" + scopeApplications
	ScopeWriteApplications      Scope = ScopeWrite + ":" + scopeApplications
	ScopeReadBlocks             Scope = ScopeRead + ":" + scopeBlocks
	ScopeWriteBlocks            Scope = ScopeWrite + ":" + scopeBlocks
	ScopeReadBookmarks          Scope = ScopeRead + ":" + scopeBookmarks
	ScopeWriteBookmarks         Scope = ScopeWrite + ":" + scopeBookmarks
	ScopeWriteConversations     Scope = ScopeWrite + ":" + scopeConversations
	ScopeReadFavourites         Scope = ScopeRead + ":" + scopeFavourites
	ScopeWriteFavourites        Scope = ScopeWrite + ":" + scopeFavourites
	ScopeReadFilters            Scope = ScopeRead + ":" + scopeFilters
	ScopeWriteFilters           Scope = ScopeWrite + ":" + scopeFilters
	ScopeReadFollows            Scope = ScopeRead + ":" + scopeFollows
	ScopeWriteFollows           Scope = ScopeWrite + ":" + scopeFollows
	ScopeReadLists              Scope = ScopeRead + ":" + scopeLists
	ScopeWriteLists             Scope = ScopeWrite + ":" + scopeLists
	ScopeWriteMedia             Scope = ScopeWrite + ":" + scopeMedia
	ScopeReadMutes              Scope = ScopeRead + ":" + scopeMutes
	ScopeWriteMutes             Scope = ScopeWrite + ":" + scopeMutes
	ScopeReadNotifications      Scope = ScopeRead + ":" + scopeNotifications
	ScopeWriteNotifications     Scope = ScopeWrite + ":" + scopeNotifications
	ScopeWriteReports           Scope = ScopeWrite + ":" + scopeReports
	ScopeReadSearch             Scope = ScopeRead + ":" + scopeSearch
	ScopeReadStatuses           Scope = ScopeRead + ":" + scopeStatuses
	ScopeWriteStatuses          Scope = ScopeWrite + ":" + scopeStatuses
	ScopeAdminReadAccounts      Scope = ScopeAdminRead + ":" + scopeAccounts
	ScopeAdminWriteAccounts     Scope = ScopeAdminWrite + ":" + scopeAccounts
	ScopeAdminReadReports       Scope = ScopeAdminRead + ":" + scopeReports
	ScopeAdminWriteReports      Scope = ScopeAdminWrite + ":" + scopeReports
	ScopeAdminReadDomainAllows  Scope = ScopeAdminRead + ":" + scopeDomainAllows
	ScopeAdminWriteDomainAllows Scope = ScopeAdminWrite + ":" + scopeDomainAllows
	ScopeAdminReadDomainBlocks  Scope = ScopeAdminRead + ":" + scopeDomainBlocks
	ScopeAdminWriteDomainBlocks Scope = ScopeAdminWrite + ":" + scopeDomainBlocks
)

// Permits returns true if the
// scope permits the wanted scope.
func (has Scope) Permits(wanted Scope) bool {
	if has == wanted {
		// Exact match on either a
		// top-level or granular scope.
		return true
	}

	// Ensure we have a
	// known top-level scope.
	switch has {

	case ScopeProfile,
		ScopePush,
		ScopeRead,
		ScopeWrite,
		ScopeAdmin,
		ScopeAdminRead,
		ScopeAdminWrite:
		// Check if top-level includes wanted,
		// eg., have "admin", want "admin:read".
		return strings.HasPrefix(string(wanted), string(has)+":")

	default:
		// Unknown top-level scope,
		// can't permit anything.
		return false
	}
}
