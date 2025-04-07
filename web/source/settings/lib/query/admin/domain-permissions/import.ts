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

import { replaceCacheOnMutation } from "../../query-modifiers";
import { gtsApi } from "../../gts-api";

import {
	type DomainPerm,
	type ImportDomainPermsParams,
	type MappedDomainPerms,
	stripOnImport,
} from "../../../types/domain-permission";
import { listToKeyedObject } from "../../transforms";

/**
 * Builds up a map function that can be applied to a
 * list of DomainPermission entries in order to normalize
 * them before submission to the API.
 * @param formData 
 * @returns 
 */
function importEntriesProcessor(formData: ImportDomainPermsParams): (_entry: DomainPerm) => DomainPerm {
	let processingFuncs: { (_entry: DomainPerm): void; }[] = [];

	// Override each obfuscate entry if necessary.
	if (formData.obfuscate !== undefined) {
		processingFuncs.push((entry: DomainPerm) => {
			entry.obfuscate = formData.obfuscate;
		});
	}

	// Check whether we need to replace
	// private_comment and/or public_comment.
	["private_comment","public_comment"].forEach((commentType) => {
		if (formData[`replace_${commentType}`]) {
			const text = formData[commentType]?.trim();
			processingFuncs.push((entry: DomainPerm) => {
				entry[commentType] = text;
			});
		}
	});

	return function process(entry) {
		// Call all the assembled processing functions.
		processingFuncs.forEach((f) => f(entry));

		// Unset all internal processing keys
		// and any undefined keys on this entry.
		Object.entries(entry).forEach(([key, val]: [keyof DomainPerm, any]) => {			
			if (val == undefined || stripOnImport(key)) {
				delete entry[key];
			}
		});

		return entry;
	};
}

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({		
		importDomainPerms: build.mutation<MappedDomainPerms, ImportDomainPermsParams>({
			query: (formData) => {
				// Add/replace comments, remove internal keys.
				const process = importEntriesProcessor(formData);
				const domains = formData.domains.map(process);

				return {
					method: "POST",
					url: `/api/v1/admin/domain_${formData.permType}s?import=true`,
					asForm: true,
					discardEmpty: true,
					body: {
						import: true,
						domains: new Blob(
							[JSON.stringify(domains)],
							{ type: "application/json" },
						),
					}
				};
			},
			transformResponse: listToKeyedObject<DomainPerm>("domain"),
			...replaceCacheOnMutation((formData: ImportDomainPermsParams) => {
				// Query names for blocks and allows are like
				// `domainBlocks` and `domainAllows`, so we need
				// to convert `block` -> `Block` or `allow` -> `Allow`
				// to do proper cache invalidation.
				const permType =
					formData.permType.charAt(0).toUpperCase() +
					formData.permType.slice(1); 
				return `domain${permType}s`;
			}),
		})
	})
});

/**
 * POST domain permissions to /api/v1/admin/domain_{permType}s.
 * Returns the newly created permissions.
 */
const useImportDomainPermsMutation = extended.useImportDomainPermsMutation;

export {
	useImportDomainPermsMutation,
};
