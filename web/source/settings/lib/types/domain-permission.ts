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
import { PermSubContentType } from "./permsubcontenttype";

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
	replace_private_comment?: boolean;
	replace_public_comment?: boolean;
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
	"replace_private_comment",
	"replace_public_comment",
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

/**
 * API model of one domain permission susbcription.
 */
export interface DomainPermSub {
	/**
	 * The ID of the domain permission subscription.
	 */
	id: string;
	/**
	 * The priority of the domain permission subscription.
	 */
	priority: number;
	/**
	 *  Time at which the subscription was created (ISO 8601 Datetime).
	 */
	created_at: string;
	/**
	 * Title of this subscription, as set by admin who created or updated it.
	 */
	title: string;
	/**
	 * The type of domain permission subscription (allow, block).
	 */
	permission_type: PermType;
	/**
	 * If true, domain permissions arising from this subscription will be created as drafts that must be approved by a moderator to take effect.
	 * If false, domain permissions from this subscription will come into force immediately.
	 */
	as_draft: boolean;
	/**
	 * If true, this domain permission subscription will "adopt" domain permissions
	 * which already exist on the instance, and which meet the following conditions:
	 * 1) they have no subscription ID (ie., they're "orphaned") and 2) they are present
	 * in the subscribed list. Such orphaned domain permissions will be given this
	 * subscription's subscription ID value and be managed by this subscription.
	 */
	adopt_orphans: boolean;
	/**
	 * ID of the account that created this subscription.
	 */
	created_by: string;
	/**
	 * URI to call in order to fetch the permissions list.
	 */
	uri: string;
	/**
	 * MIME content type to use when parsing the permissions list.
	 */
	content_type: PermSubContentType;
	/**
	 * (Optional) username to set for basic auth when doing a fetch of URI.
	 */
	fetch_username?: string;
	/**
	 * (Optional) password to set for basic auth when doing a fetch of URI.
	 */
	fetch_password?: string;
	/**
	 * Time at which the most recent fetch was attempted (ISO 8601 Datetime).
	 */
	fetched_at?: string;
	/**
	 *  Time of the most recent successful fetch (ISO 8601 Datetime).
	 */
	successfully_fetched_at?: string;
	/**
	 * If most recent fetch attempt failed, this field will contain an error message related to the fetch attempt.
	 */
	error?: string;
	/**
	 * Count of domain permission entries discovered at URI on last (successful) fetch.
	 */
	count: number;
}

/**
 * Parameters for GET to /api/v1/admin/domain_permission_subscriptions.
 */
export interface DomainPermSubSearchParams {
	/**
	 * Return only block or allow subscriptions.
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

export interface DomainPermSubCreateUpdateParams {
	/**
	 * The priority of the domain permission subscription.
	 */
	priority?: number;
	/**
	 * Title of this subscription, as set by admin who created or updated it.
	 */
	title?: string;
	/**
	 * URI to call in order to fetch the permissions list.
	 */
	uri: string;
	/**
	 * MIME content type to use when parsing the permissions list.
	 */
	content_type: PermSubContentType;
	/**
	 * If true, domain permissions arising from this subscription will be created as drafts that must be approved by a moderator to take effect.
	 * If false, domain permissions from this subscription will come into force immediately.
	 */
	as_draft?: boolean;
	/**
	 * If true, this domain permission subscription will "adopt" domain permissions
	 * which already exist on the instance, and which meet the following conditions:
	 * 1) they have no subscription ID (ie., they're "orphaned") and 2) they are present
	 * in the subscribed list. Such orphaned domain permissions will be given this
	 * subscription's subscription ID value and be managed by this subscription.
	 */
	adopt_orphans?: boolean;
	/**
	 * (Optional) username to set for basic auth when doing a fetch of URI.
	 */
	fetch_username?: string;
	/**
	 * (Optional) password to set for basic auth when doing a fetch of URI.
	 */
	fetch_password?: string;
	/**
	 * The type of domain permission subscription to create or update (allow, block).
	 */
	permission_type: PermType;
}

export interface DomainPermSubSearchResp {
	subs: DomainPermSub[];
	links: Links | null;
}
