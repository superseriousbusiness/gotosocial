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

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		updateCredentials: build.mutation({
			query: (formData) => ({
				method: "PATCH",
				url: `/api/v1/accounts/update_credentials`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			...replaceCacheOnMutation("verifyCredentials")
		}),
		passwordChange: build.mutation({
			query: (data) => ({
				method: "POST",
				url: `/api/v1/user/password_change`,
				body: data
			})
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
		})
	})
});

export const {
	useUpdateCredentialsMutation,
	usePasswordChangeMutation,
	useAliasAccountMutation,
	useMoveAccountMutation,
	useAccountThemesQuery,
} = extended;
