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

export const isDomainPerms = typia.createIs<DomainPerm[]>();

export type PermType = "block" | "allow";

/**
 * A single domain permission entry (block or allow).
 */
export interface DomainPerm {
	id?: string;
	domain: string;
	obfuscate?: boolean;
	private_comment?: string;
	public_comment?: string;

	// Internal processing keys; remove
	// before serdes of domain perm.
	key?: string;
	permType?: PermType;
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

const domainPermInternalKeys: Set<keyof DomainPerm> = new Set([
	"key",
	"permType",
	"suggest",
	"valid",
	"checked",
	"commentType",
	"private_comment_behavior",
	"public_comment_behavior",
]);

/**
 * Returns true if provided DomainPerm Object key is
 * "internal"; ie., it's just for our use, and it shouldn't
 * be serialized to or deserialized from the GtS API.
 * 
 * @param key 
 * @returns 
 */
export function isDomainPermInternalKey(key: keyof DomainPerm) {
	return domainPermInternalKeys.has(key);
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
