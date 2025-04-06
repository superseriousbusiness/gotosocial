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

import {
	replaceCacheOnMutation,
	removeFromCacheOnMutation,
	updateCacheOnMutation,
} from "../../query-modifiers";
import { listToKeyedObject } from "../../transforms";
import type {
	DomainPerm,
	MappedDomainPerms
} from "../../../types/domain-permission";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		addDomainBlock: build.mutation<MappedDomainPerms, any>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/domain_blocks`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			transformResponse: listToKeyedObject<DomainPerm>("domain"),
			...replaceCacheOnMutation("domainBlocks"),
		}),

		addDomainAllow: build.mutation<MappedDomainPerms, any>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/domain_allows`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			transformResponse: listToKeyedObject<DomainPerm>("domain"),
			...replaceCacheOnMutation("domainAllows")
		}),

		updateDomainBlock: build.mutation<DomainPerm, any>({
			query: ({ id, ...formData}) => ({
				method: "PUT",
				url: `/api/v1/admin/domain_blocks/${id}`,
				asForm: true,
				body: formData,
				discardEmpty: false
			}),
			...updateCacheOnMutation("domainBlocks", {
				key: (_draft, newData) => {
					return newData.domain;
				}
			})
		}),

		updateDomainAllow: build.mutation<DomainPerm, any>({
			query: ({ id, ...formData}) => ({
				method: "PUT",
				url: `/api/v1/admin/domain_allows/${id}`,
				asForm: true,
				body: formData,
				discardEmpty: false
			}),
			...updateCacheOnMutation("domainAllows", {
				key: (_draft, newData) => {
					return newData.domain;
				}
			})
		}),

		removeDomainBlock: build.mutation<DomainPerm, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/domain_blocks/${id}`,
			}),
			...removeFromCacheOnMutation("domainBlocks", {
				key: (_draft, newData) => {
					return newData.domain;
				}
			})
		}),

		removeDomainAllow: build.mutation<DomainPerm, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/domain_allows/${id}`,
			}),
			...removeFromCacheOnMutation("domainAllows", {
				key: (_draft, newData) => {
					return newData.domain;
				}
			})
		}),
	}),
});

/**
 * Add a single domain permission (block) by POSTing to `/api/v1/admin/domain_blocks`.
 */
const useAddDomainBlockMutation = extended.useAddDomainBlockMutation;

/**
 * Add a single domain permission (allow) by POSTing to `/api/v1/admin/domain_allows`.
 */
const useAddDomainAllowMutation = extended.useAddDomainAllowMutation;

/**
 * Update a single domain permission (block) by PUTing to `/api/v1/admin/domain_blocks/{id}`.
 */
const useUpdateDomainBlockMutation = extended.useUpdateDomainBlockMutation;

/**
 * Update a single domain permission (allow) by PUTing to `/api/v1/admin/domain_allows/{id}`.
 */
const useUpdateDomainAllowMutation = extended.useUpdateDomainAllowMutation;

/**
 * Remove a single domain permission (block) by DELETEing to `/api/v1/admin/domain_blocks/{id}`.
 */
const useRemoveDomainBlockMutation = extended.useRemoveDomainBlockMutation;

/**
 * Remove a single domain permission (allow) by DELETEing to `/api/v1/admin/domain_allows/{id}`.
 */
const useRemoveDomainAllowMutation = extended.useRemoveDomainAllowMutation;

export {
	useAddDomainBlockMutation,
	useAddDomainAllowMutation,
	useUpdateDomainBlockMutation,
	useUpdateDomainAllowMutation,
	useRemoveDomainBlockMutation,
	useRemoveDomainAllowMutation
};
