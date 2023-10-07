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

import { replaceCacheOnMutation, removeFromCacheOnMutation, idListToObject } from "../lib";
import { gtsApi } from "../gts-api";

const extended = gtsApi.injectEndpoints({
	endpoints: (builder) => ({
		updateInstance: builder.mutation({
			query: (formData) => ({
				method: "PATCH",
				url: `/api/v1/instance`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			...replaceCacheOnMutation("instance")
		}),

		mediaCleanup: builder.mutation({
			query: (days) => ({
				method: "POST",
				url: `/api/v1/admin/media_cleanup`,
				params: {
					remote_cache_days: days
				}
			})
		}),

		instanceKeysExpire: builder.mutation({
			query: (domain) => ({
				method: "POST",
				url: `/api/v1/admin/domain_keys_expire`,
				params: {
					domain: domain
				}
			})
		}),

		addInstanceBlock: builder.mutation({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/domain_blocks`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			transformResponse: (data) => {
				return {
					[data.domain]: data
				};
			},
			...replaceCacheOnMutation("instanceBlocks")
		}),

		removeInstanceBlock: builder.mutation({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/domain_blocks/${id}`,
			}),
			...removeFromCacheOnMutation("instanceBlocks", {
				findKey: (_draft, newData) => {
					return newData.domain;
				},
				key: undefined,
				arg: undefined,
			})
		}),

		getAccount: builder.query({
			query: (id) => ({
				url: `/api/v1/accounts/${id}`
			}),
			providesTags: (_, __, id) => [{ type: "Account", id }]
		}),

		actionAccount: builder.mutation({
			query: ({ id, action, reason }) => ({
				method: "POST",
				url: `/api/v1/admin/accounts/${id}/action`,
				asForm: true,
				body: {
					type: action,
					text: reason
				}
			}),
			invalidatesTags: (_, __, { id }) => [{ type: "Account", id }]
		}),

		searchAccount: builder.mutation({
			query: (username) => ({
				url: `/api/v2/search?q=${encodeURIComponent(username)}&resolve=true`
			}),
			transformResponse: (res) => {
				return res.accounts ?? [];
			}
		}),

		instanceRules: builder.query({
			query: () => ({
				url: `/api/v1/admin/instance/rules`
			}),
			transformResponse: idListToObject
		}),

		addInstanceRule: builder.mutation({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/instance/rules`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			transformResponse: (data) => {
				return {
					[data.id]: data
				};
			},
			...replaceCacheOnMutation("instanceRules")
		}),

		updateInstanceRule: builder.mutation({
			query: ({ id, ...edit }) => ({
				method: "PATCH",
				url: `/api/v1/admin/instance/rules/${id}`,
				asForm: true,
				body: edit,
				discardEmpty: true
			}),
			transformResponse: (data) => {
				return {
					[data.id]: data
				};
			},
			...replaceCacheOnMutation("instanceRules")
		}),

		deleteInstanceRule: builder.mutation({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/instance/rules/${id}`
			}),
			...removeFromCacheOnMutation("instanceRules", {
				findKey: (_draft, rule) => rule.id,
				key: undefined,
				arg: undefined,
			})
		})
	})
});

export const {
	useUpdateInstanceMutation,
	useMediaCleanupMutation,
	useInstanceKeysExpireMutation,
	useAddInstanceBlockMutation,
	useRemoveInstanceBlockMutation,
	useGetAccountQuery,
	useActionAccountMutation,
	useSearchAccountMutation,
	useInstanceRulesQuery,
	useAddInstanceRuleMutation,
	useUpdateInstanceRuleMutation,
	useDeleteInstanceRuleMutation,
} = extended;
