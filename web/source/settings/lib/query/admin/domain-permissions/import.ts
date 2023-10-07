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

import { replaceCacheOnMutation, domainListToObject } from "../../lib";
import { gtsApi } from "../../gts-api";
import { entryProcessor } from "./process";

import type { DomainPermsImportForm } from "../../../types/domain-permission";

function normalizePermsBody(formData: DomainPermsImportForm) {
	const { domains } = formData;

	// Add/replace comments, override obfuscate
	// if desired, remove internal keys.
	let process = entryProcessor(formData);
	return domains.map(process);
}

/**
 * POST domain blocks to /api/v1/admin/domain_blocks.
 */
const useImportDomainBlocksMutation = gtsApi.injectEndpoints({
	endpoints: (build) => ({		
		importDomainBlocks: build.mutation<any, DomainPermsImportForm>({
			query: (formData) => {
				const domains = normalizePermsBody(formData);

				return {
					method: "POST",
					url: `/api/v1/admin/domain_blocks?import=true`,
					asForm: true,
					discardEmpty: true,
					body: {
						domains: new Blob([JSON.stringify(domains)], { type: "application/json" })
					}
				};
			},
			transformResponse: domainListToObject,
			...replaceCacheOnMutation("instanceBlocks")
		})
	})
}).useImportDomainBlocksMutation;

/**
 * POST domain allows to /api/v1/admin/domain_allows.
 */
const useImportDomainAllowsMutation = gtsApi.injectEndpoints({
	endpoints: (build) => ({		
		importDomainAllows: build.mutation<any, DomainPermsImportForm>({
			query: (formData) => {
				const domains = normalizePermsBody(formData);

				return {
					method: "POST",
					url: `/api/v1/admin/domain_allows?import=true`,
					asForm: true,
					discardEmpty: true,
					body: {
						domains: new Blob([JSON.stringify(domains)], { type: "application/json" })
					}
				};
			},
			transformResponse: domainListToObject,
			...replaceCacheOnMutation("instanceAllows")
		})
	})
}).useImportDomainAllowsMutation;

export {
	useImportDomainBlocksMutation,
	useImportDomainAllowsMutation,
};
