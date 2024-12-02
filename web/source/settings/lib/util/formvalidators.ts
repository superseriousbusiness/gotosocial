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

/**
 * Validate the "domain" field of a form.
 * @param domain 
 * @returns 
 */
export function formDomainValidator(domain: string): string {
	if (domain.length === 0) {
		return "";
	}

	if (domain[domain.length-1] === ".") {
		return "invalid domain";
	}

	const valid = isValidDomain(domain, {
		subdomain: true,
		wildcard: false,
		allowUnicode: true,
		topLevel: false,
	});

	if (valid) {
		return "";
	}

	return "invalid domain";
}
