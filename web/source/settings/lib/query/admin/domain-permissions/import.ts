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

import { replaceCacheOnMutation } from "../../lib";
import { gtsApi } from "../../gts-api";
import { entryProcessor } from "./process";

import type { DomainPermsImportForm } from "../../../types/domain-permission";
import { domainPermsToObject } from "./transforms";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({		
		importDomainPerms: build.mutation<any, DomainPermsImportForm>({
			query: (formData) => {
				// Add/replace comments, remove internal keys.
				const process = entryProcessor(formData);
				const domains = formData.domains.map(process);

				return {
					method: "POST",
					url: `/api/v1/admin/domain_${formData.perm_type}s`,
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
			transformResponse: domainPermsToObject,
			...replaceCacheOnMutation("instanceBlocks")
		})
	})
});

/**
 * POST domain permissions to /api/v1/admin/domain_{perm_type}s.
 */
const useImportDomainPermsMutation = extended.useImportDomainPermsMutation;

export {
	useImportDomainPermsMutation,
};
