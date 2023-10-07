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

export interface DomainPerm {
	domain: string;
	obfuscate?: boolean;
	private_comment?: string;
	public_comment?: string;

	// Internal keys, remove before
	// submission of domain perm.
	key?: string;
	suggest?: string;
	valid?: boolean;
	checked?: boolean;
	commentType?: string;
	private_comment_behavior?: "append" | "replace";
	public_comment_behavior?: "append" | "replace";
}

export const DomainPermInternalKeys = new Set([
	"key",
	"suggest",
	"valid",
	"checked",
	"commentType",
	"private_comment_behavior",
	"public_comment_behavior",
]);

export interface DomainPermsImportForm {
	domains: DomainPerm[];

	// Internal keys.
	obfuscate?: boolean;
	commentType?: string;
	perm_type: "block" | "allow";
}

export interface MappedDomainPerms {
	[key: string]: DomainPerm;
}
