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

import { replaceCacheOnMutation } from "../query-modifiers";
import { gtsApi } from "../gts-api";
import type {
	MoveAccountFormData,
	UpdateAliasesFormData
} from "../../types/migration";
import type { Theme } from "../../types/theme";
import { User } from "../../types/user";
import { DefaultInteractionPolicies, UpdateDefaultInteractionPolicies } from "../../types/interaction";
import { Account } from "../../types/account";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		updateCredentials: build.mutation<Account, any>({
			query: (formData) => ({
				method: "PATCH",
				url: `/api/v1/accounts/update_credentials`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			...replaceCacheOnMutation("verifyCredentials")
		}),

		deleteHeader: build.mutation<Account, void>({
			query: (_) => ({
				method: "DELETE",
				url: `/api/v1/profile/header`,
			}),
			...replaceCacheOnMutation("verifyCredentials")
		}),

		deleteAvatar: build.mutation<Account, void>({
			query: (_) => ({
				method: "DELETE",
				url: `/api/v1/profile/avatar`,
			}),
			...replaceCacheOnMutation("verifyCredentials")
		}),
		
		user: build.query<User, void>({
			query: () => ({url: `/api/v1/user`}),
			providesTags: ["User"],
		}),
		
		passwordChange: build.mutation({
			query: (data) => ({
				method: "POST",
				url: `/api/v1/user/password_change`,
				body: data
			})
		}),
		
		emailChange: build.mutation<User, { password: string, new_email: string }>({
			query: (data) => ({
				method: "POST",
				url: `/api/v1/user/email_change`,
				body: data
			}),
			...replaceCacheOnMutation("user")
		}),
		
		aliasAccount: build.mutation<any, UpdateAliasesFormData>({
			async queryFn(formData, _api, _extraOpts, fetchWithBQ) {
				// Pull entries out from the hooked form.
				const entries: String[] = [];
				formData.also_known_as_uris.forEach(entry => {
					if (entry) {
						entries.push(entry);
					}
				});

				return fetchWithBQ({
					method: "POST",
					url: `/api/v1/accounts/alias`,
					body: { also_known_as_uris: entries },
				});
			}
		}),
		
		moveAccount: build.mutation<any, MoveAccountFormData>({
			query: (data) => ({
				method: "POST",
				url: `/api/v1/accounts/move`,
				body: data
			})
		}),

		accountThemes: build.query<Theme[], void>({
			query: () => ({
				url: `/api/v1/accounts/themes`
			})
		}),

		defaultInteractionPolicies: build.query<DefaultInteractionPolicies, void>({
			query: () => ({
				url: `/api/v1/interaction_policies/defaults`
			}),
			providesTags: ["DefaultInteractionPolicies"]
		}),

		updateDefaultInteractionPolicies: build.mutation<DefaultInteractionPolicies, UpdateDefaultInteractionPolicies>({
			query: (data) => ({
				method: "PATCH",
				url: `/api/v1/interaction_policies/defaults`,
				body: data,
			}),
			...replaceCacheOnMutation("defaultInteractionPolicies")
		}),

		resetDefaultInteractionPolicies: build.mutation<DefaultInteractionPolicies, void>({
			query: () => ({
				method: "PATCH",
				url: `/api/v1/interaction_policies/defaults`,
				body: {},
			}),
			invalidatesTags: ["DefaultInteractionPolicies"]
		}),
	})
});

export const {
	useUpdateCredentialsMutation,
	useDeleteHeaderMutation,
	useDeleteAvatarMutation,
	useUserQuery,
	usePasswordChangeMutation,
	useEmailChangeMutation,
	useAliasAccountMutation,
	useMoveAccountMutation,
	useAccountThemesQuery,
	useDefaultInteractionPoliciesQuery,
	useUpdateDefaultInteractionPoliciesMutation,
	useResetDefaultInteractionPoliciesMutation,
} = extended;
