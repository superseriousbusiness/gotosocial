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

import typia from "typia";
import { PermType } from "./perm";
import { Links } from "parse-link-header";

export const validateDomainPerms = typia.createValidate<DomainPerm[]>();

/**
 * A single domain permission entry (block, allow, draft, ignore).
 */
export interface DomainPerm {
	id?: string;
	domain: string;
	obfuscate?: boolean;
	private_comment?: string;
	public_comment?: string;
	created_at?: string;
	created_by?: string;
	subscription_id?: string;

	// Keys that should be stripped before
	// sending the domain permission (if imported).

	permission_type?: PermType;
	key?: string;
	suggest?: string;
	valid?: boolean;
	checked?: boolean;
	commentType?: string;
	private_comment_behavior?: "append" | "replace";
	public_comment_behavior?: "append" | "replace";
}

/**
 * Domain permissions mapped to an Object where the Object
 * keys are the "domain" value of each DomainPerm.
 */
export interface MappedDomainPerms {
	[key: string]: DomainPerm;
}

const domainPermStripOnImport: Set<keyof DomainPerm> = new Set([
	"key",
	"permission_type",
	"suggest",
	"valid",
	"checked",
	"commentType",
	"private_comment_behavior",
	"public_comment_behavior",
]);

/**
 * Returns true if provided DomainPerm Object key is one
 * that should be stripped when importing a domain permission.
 * 
 * @param key 
 * @returns 
 */
export function stripOnImport(key: keyof DomainPerm) {
	return domainPermStripOnImport.has(key);
}

export interface ImportDomainPermsParams {
	domains: DomainPerm[];

	// Internal processing keys;
	// remove before serdes of form.
	obfuscate?: boolean;
	commentType?: string;
	permType: PermType;
}

/**
 * Model domain permissions bulk export params.
 */
export interface ExportDomainPermsParams {
	permType: PermType;
	action: "export" | "export-file";
	exportType: "json" | "csv" | "plain";
}

/**
 * Parameters for GET to /api/v1/admin/domain_permission_drafts.
 */
export interface DomainPermDraftSearchParams {
	/**
	 * Show only drafts created by the given subscription ID.
	 */
	subscription_id?: string;
	/**
	 * Return only drafts that target the given domain.
	 */
	domain?: string;
	/**
	 * Filter on "block" or "allow" type drafts.
	 */
	permission_type?: PermType;
	/**
	 * Return only items *OLDER* than the given max ID (for paging downwards).
	 * The item with the specified ID will not be included in the response.
	 */
	max_id?: string;
	/**
	 * Return only items *NEWER* than the given since ID.
	 * The item with the specified ID will not be included in the response.
	 */
	since_id?: string;
	/**
	 * Return only items immediately *NEWER* than the given min ID (for paging upwards).
	 * The item with the specified ID will not be included in the response.
	 */
	min_id?: string;
	/**
	 * Number of items to return.
	 */
	limit?: number;
}

export interface DomainPermDraftSearchResp {
	drafts: DomainPerm[];
	links: Links | null;
}

export interface DomainPermDraftCreateParams {
	/**
	 * Domain to create the permission draft for.
	 */
	domain: string;
	/**
	 * Create a draft "allow" or a draft "block".
	 */
	permission_type: PermType;
	/**
	 * Obfuscate the name of the domain when serving it publicly.
	 * Eg., `example.org` becomes something like `ex***e.org`.
	 */
	obfuscate?: boolean;
	/**
	 * Public comment about this domain permission. This will be displayed
	 * alongside the domain permission if you choose to share permissions.
	 */
	public_comment?: string;
	/**
	 * Private comment about this domain permission.
	 * Will only be shown to other admins, so this is a useful way of
	 * internally keeping track of why a certain domain ended up permissioned.
	 */
	private_comment?: string;
}

/**
 * Parameters for GET to /api/v1/admin/domain_permission_excludes.
 */
export interface DomainPermExcludeSearchParams {
	/**
	 * Return only excludes that target the given domain.
	 */
	domain?: string;
	/**
	 * Return only items *OLDER* than the given max ID (for paging downwards).
	 * The item with the specified ID will not be included in the response.
	 */
	max_id?: string;
	/**
	 * Return only items *NEWER* than the given since ID.
	 * The item with the specified ID will not be included in the response.
	 */
	since_id?: string;
	/**
	 * Return only items immediately *NEWER* than the given min ID (for paging upwards).
	 * The item with the specified ID will not be included in the response.
	 */
	min_id?: string;
	/**
	 * Number of items to return.
	 */
	limit?: number;
}

export interface DomainPermExcludeSearchResp {
	excludes: DomainPerm[];
	links: Links | null;
}

export interface DomainPermExcludeCreateParams {
	/**
	 * Domain to create the permission exclude for.
	 */
	domain: string;
	/**
	 * Private comment about this domain permission.
	 * Will only be shown to other admins, so this is a useful way of
	 * internally keeping track of why a certain domain ended up permissioned.
	 */
	private_comment?: string;
}
