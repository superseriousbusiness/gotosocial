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
	DomainPermDraftCreateParams,
	DomainPermDraftSearchParams,
	DomainPermDraftSearchResp,
} from "../../../types/domain-permission";
import parse from "parse-link-header";
import { PermType } from "../../../types/perm";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		searchDomainPermissionDrafts: build.query<DomainPermDraftSearchResp, DomainPermDraftSearchParams>({
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
					url: `/api/v1/admin/domain_permission_drafts${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: DomainPerm[], meta) => {
				const drafts = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { drafts, links };
			},
			// Only provide TRANSFORMED tag id since this model is not the same
			// as getDomainPermissionDraft model (due to transformResponse).
			providesTags: [{ type: "DomainPermissionDraft", id: "TRANSFORMED" }]
		}),

		getDomainPermissionDraft: build.query<DomainPerm, string>({
			query: (id) => ({
				url: `/api/v1/admin/domain_permission_drafts/${id}`
			}),
			providesTags: (_result, _error, id) => [
				{ type: 'DomainPermissionDraft', id }
			],
		}),

		createDomainPermissionDraft: build.mutation<DomainPerm, DomainPermDraftCreateParams>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/domain_permission_drafts`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			invalidatesTags: [{ type: "DomainPermissionDraft", id: "TRANSFORMED" }],
		}),

		acceptDomainPermissionDraft: build.mutation<DomainPerm, { id: string, overwrite?: boolean, permType: PermType }>({
			query: ({ id, overwrite }) => ({
				method: "POST",
				url: `/api/v1/admin/domain_permission_drafts/${id}/accept`,
				asForm: true,
				body: {
					overwrite: overwrite,
				},
				discardEmpty: true
			}),
			invalidatesTags: (res, _error, { id, permType }) => {
				const invalidated: any[] = [];
				
				// If error, nothing to invalidate.
				if (!res) {
					return invalidated;
				}
				
				// Invalidate this draft by ID, and
				// the transformed list of all drafts.
				invalidated.push(
					{ type: 'DomainPermissionDraft', id: id },
					{ type: "DomainPermissionDraft", id: "TRANSFORMED" },
				);

				// Invalidate cached blocks/allows depending
				// on the permType of the accepted draft.
				if (permType === "allow") {
					invalidated.push("domainAllows");
				} else {
					invalidated.push("domainBlocks");
				}

				return invalidated;
			}
		}),

		removeDomainPermissionDraft: build.mutation<DomainPerm, { id: string, exclude_target?: boolean }>({
			query: ({ id, exclude_target }) => ({
				method: "POST",
				url: `/api/v1/admin/domain_permission_drafts/${id}/remove`,
				asForm: true,
				body: {
					exclude_target: exclude_target,
				},
				discardEmpty: true
			}),
			invalidatesTags: (res, _error, { id }) =>
				res
					? [
						{ type: "DomainPermissionDraft", id },
						{ type: "DomainPermissionDraft", id: "TRANSFORMED" },
					]
					: [],
		})

	}),
});

/**
 * View domain permission drafts.
 */
const useLazySearchDomainPermissionDraftsQuery = extended.useLazySearchDomainPermissionDraftsQuery;

/**
 * Get domain permission draft with the given ID.
 */
const useGetDomainPermissionDraftQuery = extended.useGetDomainPermissionDraftQuery;

/**
 * Create a domain permission draft with the given parameters.
 */
const useCreateDomainPermissionDraftMutation = extended.useCreateDomainPermissionDraftMutation;

/**
 * Accept a domain permission draft, turning it into an enforced domain permission.
 */
const useAcceptDomainPermissionDraftMutation = extended.useAcceptDomainPermissionDraftMutation;

/**
 * Remove a domain permission draft, optionally ignoring all future drafts targeting the given domain.
 */
const useRemoveDomainPermissionDraftMutation = extended.useRemoveDomainPermissionDraftMutation;

export {
	useLazySearchDomainPermissionDraftsQuery,
	useGetDomainPermissionDraftQuery,
	useCreateDomainPermissionDraftMutation,
	useAcceptDomainPermissionDraftMutation,
	useRemoveDomainPermissionDraftMutation,
};
