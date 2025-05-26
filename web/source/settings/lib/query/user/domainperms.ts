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

import { gtsApi } from "../gts-api";

import type { DomainPerm } from "../../types/domain-permission";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		instanceDomainBlocks: build.query<DomainPerm[], void>({
			query: () => ({
				url: `/api/v1/instance/domain_blocks`
			}),
		}),

		instanceDomainAllows: build.query<DomainPerm[], void>({
			query: () => ({
				url: `/api/v1/instance/domain_allows`
			})
		}),
	}),
});

/**
 * Get user-level view of all explicitly blocked domains.
 */
const useInstanceDomainBlocksQuery = extended.useInstanceDomainBlocksQuery;

/**
 * Get user-level view of all explicitly allowed domains.
 */
const useInstanceDomainAllowsQuery = extended.useInstanceDomainAllowsQuery;

export {
	useInstanceDomainBlocksQuery,
	useInstanceDomainAllowsQuery,
};
