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

import type {
	DomainPerm,
	DomainPermExcludeCreateParams,
	DomainPermExcludeSearchParams,
	DomainPermExcludeSearchResp,
} from "../../../types/domain-permission";
import parse from "parse-link-header";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		searchDomainPermissionExcludes: build.query<DomainPermExcludeSearchResp, DomainPermExcludeSearchParams>({
			query: (form) => {
				const params = new(URLSearchParams);
				Object.entries(form).forEach(([k, v]) => {
					if (v !== undefined) {
						params.append(k, v);
					}
				});

				let query = "";
				if (params.size !== 0) {
					query = `?${params.toString()}`;
				}

				return {
					url: `/api/v1/admin/domain_permission_excludes${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: DomainPerm[], meta) => {
				const excludes = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { excludes, links };
			},
			// Only provide TRANSFORMED tag id since this model is not the same
			// as getDomainPermissionExclude model (due to transformResponse).
			providesTags: [{ type: "DomainPermissionExclude", id: "TRANSFORMED" }]
		}),

		getDomainPermissionExclude: build.query<DomainPerm, string>({
			query: (id) => ({
				url: `/api/v1/admin/domain_permission_excludes/${id}`
			}),
			providesTags: (_result, _error, id) => [
				{ type: 'DomainPermissionExclude', id }
			],
		}),

		createDomainPermissionExclude: build.mutation<DomainPerm, DomainPermExcludeCreateParams>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/domain_permission_excludes`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			invalidatesTags: [{ type: "DomainPermissionExclude", id: "TRANSFORMED" }],
		}),

		deleteDomainPermissionExclude: build.mutation<DomainPerm, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/domain_permission_excludes/${id}`,
			}),
			invalidatesTags: (res, _error, id) =>
				res
					? [
						{ type: "DomainPermissionExclude", id },
						{ type: "DomainPermissionExclude", id: "TRANSFORMED" },
					]
					: [],
		})

	}),
});

/**
 * View domain permission excludes.
 */
const useLazySearchDomainPermissionExcludesQuery = extended.useLazySearchDomainPermissionExcludesQuery;

/**
 * Get domain permission exclude with the given ID.
 */
const useGetDomainPermissionExcludeQuery = extended.useGetDomainPermissionExcludeQuery;

/**
 * Create a domain permission exclude with the given parameters.
 */
const useCreateDomainPermissionExcludeMutation = extended.useCreateDomainPermissionExcludeMutation;

/**
 * Delete a domain permission exclude.
 */
const useDeleteDomainPermissionExcludeMutation = extended.useDeleteDomainPermissionExcludeMutation;

export {
	useLazySearchDomainPermissionExcludesQuery,
	useGetDomainPermissionExcludeQuery,
	useCreateDomainPermissionExcludeMutation,
	useDeleteDomainPermissionExcludeMutation,
};
