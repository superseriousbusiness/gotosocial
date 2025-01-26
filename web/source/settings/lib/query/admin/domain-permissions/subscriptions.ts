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
	DomainPermSub,
	DomainPermSubCreateUpdateParams,
	DomainPermSubSearchParams,
	DomainPermSubSearchResp,
} from "../../../types/domain-permission";
import parse from "parse-link-header";
import { PermType } from "../../../types/perm";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		searchDomainPermissionSubscriptions: build.query<DomainPermSubSearchResp, DomainPermSubSearchParams>({
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
					url: `/api/v1/admin/domain_permission_subscriptions${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: DomainPermSub[], meta) => {
				const subs = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { subs, links };
			},
			// Only provide TRANSFORMED tag id since this model is not the same
			// as getDomainPermissionSubscription model (due to transformResponse).
			providesTags: [{ type: "DomainPermissionSubscription", id: "TRANSFORMED" }]
		}),

		getDomainPermissionSubscriptionsPreview: build.query<DomainPermSub[], PermType>({
			query: (permType) => ({
				url: `/api/v1/admin/domain_permission_subscriptions/preview?permission_type=${permType}`
			}),
			providesTags: (_result, _error, permType) =>
				// Cache by permission type.
				[{ type: "DomainPermissionSubscription", id: `${permType}sByPriority` }]
		}),

		getDomainPermissionSubscription: build.query<DomainPermSub, string>({
			query: (id) => ({
				url: `/api/v1/admin/domain_permission_subscriptions/${id}`
			}),
			providesTags: (_result, _error, id) => [
				{ type: 'DomainPermissionSubscription', id }
			],
		}),

		createDomainPermissionSubscription: build.mutation<DomainPermSub, DomainPermSubCreateUpdateParams>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/domain_permission_subscriptions`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			invalidatesTags: (_res, _error, formData) =>
				[
					// Invalidate transformed list of all perm subs.
					{ type: "DomainPermissionSubscription", id: "TRANSFORMED" },
					// Invalidate perm subs of this type sorted by priority.
					{ type: "DomainPermissionSubscription", id: `${formData.permission_type}sByPriority` }
				]
		}),

		updateDomainPermissionSubscription: build.mutation<DomainPermSub, { id: string, permType: PermType, formData: DomainPermSubCreateUpdateParams }>({
			query: ({ id, formData }) => ({
				method: "PATCH",
				url: `/api/v1/admin/domain_permission_subscriptions/${id}`,
				asForm: true,
				body: formData,
			}),
			invalidatesTags: (_res, _error, { id, permType }) =>
				[
					// Invalidate this perm sub.
					{ type: "DomainPermissionSubscription", id: id },
					// Invalidate transformed list of all perms subs.
					{ type: "DomainPermissionSubscription", id: "TRANSFORMED" },
					// Invalidate perm subs of this type sorted by priority.
					{ type: "DomainPermissionSubscription", id: `${permType}sByPriority` }
				],
		}),

		removeDomainPermissionSubscription: build.mutation<DomainPermSub, { id: string, remove_children: boolean }>({
			query: ({ id, remove_children }) => ({
				method: "POST",
				url: `/api/v1/admin/domain_permission_subscriptions/${id}/remove`,
				asForm: true,
				body: { remove_children: remove_children },
			}),
		}),

		testDomainPermissionSubscription: build.mutation<{ error: string } | DomainPerm[], string>({
			query: (id) => ({
				method: "POST",
				url: `/api/v1/admin/domain_permission_subscriptions/${id}/test`,
			}),
		})
	}),
});

/**
 * View domain permission subscriptions.
 */
const useLazySearchDomainPermissionSubscriptionsQuery = extended.useLazySearchDomainPermissionSubscriptionsQuery;

/**
 * Get domain permission subscription with the given ID.
 */
const useGetDomainPermissionSubscriptionQuery = extended.useGetDomainPermissionSubscriptionQuery;

/**
 * Create a domain permission subscription with the given parameters.
 */
const useCreateDomainPermissionSubscriptionMutation = extended.useCreateDomainPermissionSubscriptionMutation;

/**
 * View domain permission subscriptions of selected perm type, sorted by priority descending.
 */
const useGetDomainPermissionSubscriptionsPreviewQuery = extended.useGetDomainPermissionSubscriptionsPreviewQuery;

/**
 * Update domain permission subscription.
 */
const useUpdateDomainPermissionSubscriptionMutation = extended.useUpdateDomainPermissionSubscriptionMutation;

/**
 * Remove a domain permission subscription and optionally its children (harsh).
 */
const useRemoveDomainPermissionSubscriptionMutation = extended.useRemoveDomainPermissionSubscriptionMutation;

/**
 * Test a domain permission subscription to see if data can be fetched + parsed.
 */
const useTestDomainPermissionSubscriptionMutation = extended.useTestDomainPermissionSubscriptionMutation;

export {
	useLazySearchDomainPermissionSubscriptionsQuery,
	useGetDomainPermissionSubscriptionQuery,
	useCreateDomainPermissionSubscriptionMutation,
	useGetDomainPermissionSubscriptionsPreviewQuery,
	useUpdateDomainPermissionSubscriptionMutation,
	useRemoveDomainPermissionSubscriptionMutation,
	useTestDomainPermissionSubscriptionMutation,
};
