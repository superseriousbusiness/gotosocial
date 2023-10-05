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

import isValidDomain from "is-valid-domain";
import { get } from "psl";

/**
 * Check the domain string to ensure it's a valid domain
 * string that doesn't include a wildcard ("*").
 * @param domain 
 * @returns 
 */
export function isValidDomainPermission(domain: string): boolean {
	return isValidDomain(domain, {
		wildcard: false,
		allowUnicode: true
	});
}

/**
 * Checks a domain against the Public Suffix List <https://publicsuffix.org/> to see if we
 * should suggest removing subdomain(s), since they're likely owned/ran by the same party.
 * Eg., "social.example.com" suggests "example.com".
 * @param domain 
 * @returns 
 */
export function hasBetterScope(domain: string): string | undefined {
	const lookup = get(domain);
	if (lookup && lookup != domain) {
		return lookup;
	}
}
