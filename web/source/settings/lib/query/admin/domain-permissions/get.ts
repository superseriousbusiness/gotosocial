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

import { gtsApi } from "../../gts-api";

import type { DomainPerm, MappedDomainPerms } from "../../../types/domain-permission";
import { listToKeyedObject } from "../../transforms";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		domainBlocks: build.query<MappedDomainPerms, void>({
			query: () => ({
				url: `/api/v1/admin/domain_blocks`
			}),
			transformResponse: listToKeyedObject<DomainPerm>("domain"),
		}),

		domainAllows: build.query<MappedDomainPerms, void>({
			query: () => ({
				url: `/api/v1/admin/domain_allows`
			}),
			transformResponse: listToKeyedObject<DomainPerm>("domain"),
		}),

		domainPermissionDrafts: build.query<any, void>({
			query: () => ({
				url: `/api/v1/admin/domain_permission_drafts`
			}),
		}),
	}),
});

/**
 * Get admin view of all explicitly blocked domains.
 */
const useDomainBlocksQuery = extended.useDomainBlocksQuery;

/**
 * Get admin view of all explicitly allowed domains.
 */
const useDomainAllowsQuery = extended.useDomainAllowsQuery;

export {
	useDomainBlocksQuery,
	useDomainAllowsQuery,
};
