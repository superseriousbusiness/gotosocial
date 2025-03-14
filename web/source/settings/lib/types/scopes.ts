/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

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

/* Sub-scopes / scope components */

const scopeAccounts      = "accounts";
const scopeApplications  = "applications";
const scopeBlocks        = "blocks";
const scopeBookmarks     = "bookmarks";
const scopeConversations = "conversations";
const scopeDomainAllows  = "domain_allows";
const scopeDomainBlocks  = "domain_blocks";
const scopeFavourites    = "favourites";
const scopeFilters       = "filters";
const scopeFollows       = "follows";
const scopeLists         = "lists";
const scopeMedia         = "media";
const scopeMutes         = "mutes";
const scopeNotifications = "notifications";
const scopeReports       = "reports";
const scopeSearch        = "search";
const scopeStatuses      = "statuses";

/* Top-level scopes */

export const ScopeProfile    = "profile";
export const ScopePush       = "push";
export const ScopeRead       = "read";
export const ScopeWrite      = "write";
export const ScopeAdmin      = "admin";
export const ScopeAdminRead  = ScopeAdmin + ":" + ScopeRead;
export const ScopeAdminWrite = ScopeAdmin + ":" + ScopeWrite;

/* Granular scopes */

export const ScopeReadAccounts           = ScopeRead + ":" + scopeAccounts;
export const ScopeWriteAccounts          = ScopeWrite + ":" + scopeAccounts;
export const ScopeReadApplications       = ScopeRead + ":" + scopeApplications;
export const ScopeWriteApplications      = ScopeWrite + ":" + scopeApplications;
export const ScopeReadBlocks             = ScopeRead + ":" + scopeBlocks;
export const ScopeWriteBlocks            = ScopeWrite + ":" + scopeBlocks;
export const ScopeReadBookmarks          = ScopeRead + ":" + scopeBookmarks;
export const ScopeWriteBookmarks         = ScopeWrite + ":" + scopeBookmarks;
export const ScopeWriteConversations     = ScopeWrite + ":" + scopeConversations;
export const ScopeReadFavourites         = ScopeRead + ":" + scopeFavourites;
export const ScopeWriteFavourites        = ScopeWrite + ":" + scopeFavourites;
export const ScopeReadFilters            = ScopeRead + ":" + scopeFilters;
export const ScopeWriteFilters           = ScopeWrite + ":" + scopeFilters;
export const ScopeReadFollows            = ScopeRead + ":" + scopeFollows;
export const ScopeWriteFollows           = ScopeWrite + ":" + scopeFollows;
export const ScopeReadLists              = ScopeRead + ":" + scopeLists;
export const ScopeWriteLists             = ScopeWrite + ":" + scopeLists;
export const ScopeWriteMedia             = ScopeWrite + ":" + scopeMedia;
export const ScopeReadMutes              = ScopeRead + ":" + scopeMutes;
export const ScopeWriteMutes             = ScopeWrite + ":" + scopeMutes;
export const ScopeReadNotifications      = ScopeRead + ":" + scopeNotifications;
export const ScopeWriteNotifications     = ScopeWrite + ":" + scopeNotifications;
export const ScopeWriteReports           = ScopeWrite + ":" + scopeReports;
export const ScopeReadSearch             = ScopeRead + ":" + scopeSearch;
export const ScopeReadStatuses           = ScopeRead + ":" + scopeStatuses;
export const ScopeWriteStatuses          = ScopeWrite + ":" + scopeStatuses;
export const ScopeAdminReadAccounts      = ScopeAdminRead + ":" + scopeAccounts;
export const ScopeAdminWriteAccounts     = ScopeAdminWrite + ":" + scopeAccounts;
export const ScopeAdminReadReports       = ScopeAdminRead + ":" + scopeReports;
export const ScopeAdminWriteReports      = ScopeAdminWrite + ":" + scopeReports;
export const ScopeAdminReadDomainAllows  = ScopeAdminRead + ":" + scopeDomainAllows;
export const ScopeAdminWriteDomainAllows = ScopeAdminWrite + ":" + scopeDomainAllows;
export const ScopeAdminReadDomainBlocks  = ScopeAdminRead + ":" + scopeDomainBlocks;
export const ScopeAdminWriteDomainBlocks = ScopeAdminWrite + ":" + scopeDomainBlocks;

export const ValidScopes = [
	ScopeProfile,
	ScopePush,
	ScopeRead,
	ScopeWrite,
	ScopeAdmin,
	ScopeAdminRead,
	ScopeAdminWrite,
	ScopeReadAccounts,
	ScopeWriteAccounts,
	ScopeReadApplications,
	ScopeWriteApplications,
	ScopeReadBlocks,
	ScopeWriteBlocks,
	ScopeReadBookmarks,
	ScopeWriteBookmarks,
	ScopeWriteConversations,
	ScopeReadFavourites,
	ScopeWriteFavourites,
	ScopeReadFilters,
	ScopeWriteFilters,
	ScopeReadFollows,
	ScopeWriteFollows,
	ScopeReadLists,
	ScopeWriteLists,
	ScopeWriteMedia,
	ScopeReadMutes,
	ScopeWriteMutes,
	ScopeReadNotifications,
	ScopeWriteNotifications,
	ScopeWriteReports,
	ScopeReadSearch,
	ScopeReadStatuses,
	ScopeWriteStatuses,
	ScopeAdminReadAccounts,
	ScopeAdminWriteAccounts,
	ScopeAdminReadReports,
	ScopeAdminWriteReports,
	ScopeAdminReadDomainAllows,
	ScopeAdminWriteDomainAllows,
	ScopeAdminReadDomainBlocks,
	ScopeAdminWriteDomainBlocks,
];

export const ValidTopLevelScopes = [
	ScopeProfile,
	ScopePush,
	ScopeRead,
	ScopeWrite,
	ScopeAdmin,
	ScopeAdminRead,
	ScopeAdminWrite,
];
