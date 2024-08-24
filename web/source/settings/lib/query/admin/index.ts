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

import { replaceCacheOnMutation, removeFromCacheOnMutation } from "../query-modifiers";
import { gtsApi } from "../gts-api";
import { listToKeyedObject } from "../transforms";
import { ActionAccountParams, AdminAccount, HandleSignupParams, SearchAccountParams, SearchAccountResp } from "../../types/account";
import { InstanceRule, MappedRules } from "../../types/rules";
import parse from "parse-link-header";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		updateInstance: build.mutation({
			query: (formData) => ({
				method: "PATCH",
				url: `/api/v1/instance`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			...replaceCacheOnMutation("instanceV1"),
		}),

		getAccount: build.query<AdminAccount, string>({
			query: (id) => ({
				url: `/api/v1/admin/accounts/${id}`
			}),
			providesTags: (_result, _error, id) => [
				{ type: 'Account', id }
			],
		}),

		searchAccounts: build.query<SearchAccountResp, SearchAccountParams>({
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
					url: `/api/v2/admin/accounts${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: AdminAccount[], meta) => {
				const accounts = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { accounts, links };
			},
			// Only provide LIST tag id since this model is not the
			// same as getAccount model (due to transformResponse).
			providesTags: [{ type: "Account", id: "TRANSFORMED" }]
		}),

		actionAccount: build.mutation<string, ActionAccountParams>({
			query: ({ id, action, reason }) => ({
				method: "POST",
				url: `/api/v1/admin/accounts/${id}/action`,
				asForm: true,
				body: {
					type: action,
					text: reason
				}
			}),
			// Do an optimistic update on this account to mark
			// it according to whatever action was submitted.
			async onQueryStarted({ id, action }, { dispatch, queryFulfilled }) {
				const patchResult = dispatch(
					extended.util.updateQueryData("getAccount", id, (draft) => {
						if (action === "suspend") {
							draft.suspended = true;
							draft.account.suspended = true;
						}
					})
				);

				// Revert optimistic
				// update if query fails.
				try {
					await queryFulfilled;
				} catch {
					patchResult.undo();
				}
			}
		}),

		handleSignup: build.mutation<AdminAccount, HandleSignupParams>({
			query: ({id, approve_or_reject, ...formData}) => {
				return {
					method: "POST",
					url: `/api/v1/admin/accounts/${id}/${approve_or_reject}`,
					asForm: true,
					body: approve_or_reject === "reject" && formData,
				};
			},
			// Do an optimistic update on this account to mark it approved
			// if approved was true, else just invalidate getAccount.
			async onQueryStarted({ id, approve_or_reject }, { dispatch, queryFulfilled }) {
				if (approve_or_reject === "reject") {
					// Just invalidate this ID and getAccounts.
					dispatch(extended.util.invalidateTags([
						{ type: "Account", id: id },
						{ type: "Account", id: "TRANSFORMED" }
					]));
					return;
				}
				
				const patchResult = dispatch(
					extended.util.updateQueryData("getAccount", id, (draft) => {
						draft.approved = true;
					})
				);

				// Revert optimistic
				// update if query fails.
				try {
					await queryFulfilled;
				} catch {
					patchResult.undo();
				}
			}
		}),

		instanceRules: build.query<MappedRules, void>({
			query: () => ({
				url: `/api/v1/admin/instance/rules`
			}),
			transformResponse: listToKeyedObject<InstanceRule>("id")
		}),

		addInstanceRule: build.mutation<MappedRules, any>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/admin/instance/rules`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			transformResponse: listToKeyedObject<InstanceRule>("id"),
			...replaceCacheOnMutation("instanceRules"),
		}),

		updateInstanceRule: build.mutation({
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
			...replaceCacheOnMutation("instanceRules"),
		}),

		deleteInstanceRule: build.mutation({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/admin/instance/rules/${id}`
			}),
			...removeFromCacheOnMutation("instanceRules", {
				key: (_draft, rule) => rule.id,
			})
		})
	})
});

export const {
	useUpdateInstanceMutation,
	useGetAccountQuery,
	useLazyGetAccountQuery,
	useActionAccountMutation,
	useSearchAccountsQuery,
	useLazySearchAccountsQuery,
	useHandleSignupMutation,
	useInstanceRulesQuery,
	useAddInstanceRuleMutation,
	useUpdateInstanceRuleMutation,
	useDeleteInstanceRuleMutation,
} = extended;
