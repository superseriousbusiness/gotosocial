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
import { useCallback } from "react";
import { ValidScopes, ValidTopLevelScopes } from "../types/scopes";

/**
 * Validate the "domain" field of a form.
 * @param domain 
 * @returns 
 */
export function formDomainValidator(domain: string): string {
	if (domain.length === 0) {
		return "";
	}

	// Allow localhost for testing.
	if (domain === "localhost") {
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

export function urlValidator(urlStr: string): string {
	if (urlStr.length === 0) {
		return "";
	}

	let url: URL;
	try {
		url = new URL(urlStr);
	} catch (e) {
		return e.message;
	}

	if (url.protocol !== "http:" && url.protocol !== "https:") {
		return `invalid protocol, must be http or https`;
	}

	return formDomainValidator(url.hostname);
}

export function useScopesValidator(): (_scopes: string[]) => string {
	return useCallback((scopes) => {
		return scopes.
			map((scope) => validateScope(scope)).
			flatMap((msg) => msg || []).
			join(", ");
	}, []);
}

export function useScopeValidator(): (_scope: string) => string {
	return useCallback((scope) => validateScope(scope), []);
}

const validateScope = (scope: string) => {
	if (!ValidScopes.includes(scope)) {
		return scope + " is not a recognized scope";
	} 
	return "";
};

export function useScopesPermittedBy(): (_hasScopes: string[], _wantScopes: string[]) => string {
	return useCallback((hasScopes, wantsScopes) => {
		return wantsScopes.
			map((wanted) => scopePermittedByScopes(hasScopes, wanted)).
			flatMap((msg) => msg || []).
			join(", ");
	}, []);
}

const scopePermittedByScopes = (hasScopes: string[], wanted: string) => {
	if (hasScopes.some((hasScope) => scopePermittedByScope(hasScope, wanted) === "")) {
		return "";
	}
	return `scopes [${hasScopes}] do not permit ${wanted}`;
};

const scopePermittedByScope = (has: string, wanted: string) => {
	if (has === wanted) {
		// Exact match on either a
		// top-level or granular scope.
		return "";
	}

	// Ensure we have a
	// known top-level scope.
	switch (true) {
		case (ValidTopLevelScopes.includes(has)):
			// Check if top-level includes wanted,
			// eg., have "admin", want "admin:read".
			if (wanted.startsWith(has + ":")) {
				return "";
			} else {
				return `scope ${has} does not permit ${wanted}`;
			}
	
		default:
			// Unknown top-level scope,
			// can't permit anything.
			return `unrecognized scope ${has}`;
	}
};
