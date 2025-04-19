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

import { RootState } from "../../../redux/store";
import {
	SearchAppParams,
	SearchAppResp,
	App,
	AppCreateParams,
} from "../../types/application";
import { OAuthAccessToken, OAuthAccessTokenRequestBody } from "../../types/oauth";
import { gtsApi } from "../gts-api";
import parse from "parse-link-header";

const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		searchApp: build.query<SearchAppResp, SearchAppParams>({
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
					url: `/api/v1/apps${query}`
				};
			},
			// Headers required for paging.
			transformResponse: (apiResp: App[], meta) => {
				const apps = apiResp;
				const linksStr = meta?.response?.headers.get("Link");
				const links = parse(linksStr);
				return { apps, links };
			},
			providesTags: [{ type: "Application", id: "TRANSFORMED" }]
		}),

		getApp: build.query<App, string>({
			query: (id) => ({
				method: "GET",
				url: `/api/v1/apps/${id}`,
			}),
			providesTags: (_result, _error, id) => [
				{ type: 'Application', id }
			],
		}),

		createApp: build.mutation<App, AppCreateParams>({
			query: (formData) => ({
				method: "POST",
				url: `/api/v1/apps`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
			invalidatesTags: [{ type: "Application", id: "TRANSFORMED" }],
		}),

		deleteApp: build.mutation<App, string>({
			query: (id) => ({
				method: "DELETE",
				url: `/api/v1/apps/${id}`
			}),
			invalidatesTags: (_result, _error, id) => [
				{ type: 'Application', id },
				{ type: "Application", id: "TRANSFORMED" },
				{ type: "TokenInfo", id: "TRANSFORMED" },
			],
		}),

		getOOBAuthCode: build.mutation<null, { app: App, scope: string, redirectURI: string }>({
			async queryFn({ app, scope, redirectURI }, api, _extraOpts, _fetchWithBQ) {
				// Fetch the instance URL string from
				// oauth state, eg., https://example.org.
				const state = api.getState() as RootState;
				if (!state.login.instanceUrl) {
					return {
						error: {
							status: 'CUSTOM_ERROR',
							error: "oauthState.instanceUrl undefined",
						}
					};
				}
				const instanceUrl = state.login.instanceUrl;

				// Parse instance URL + set params on it.
				//
				// Note that any space-separated scopes are
				// replaced by '+'-separated, to fit the API.
				const url = new URL(instanceUrl);
				url.pathname = "/oauth/authorize";
				url.searchParams.set("client_id", app.client_id);
				url.searchParams.set("redirect_uri", redirectURI);
				url.searchParams.set("response_type", "code");
				url.searchParams.set("scope", scope.replace(" ", "+"));

				// Set the app ID in state so we know which
				// app to get out of our store after redirect.
				url.searchParams.set("state", app.id);

				// Whisk the user away to the authorize page.
				window.location.assign(url.toString());
				return { data: null };
			}
		}),

		getAccessTokenForApp: build.mutation<OAuthAccessToken, OAuthAccessTokenRequestBody>({
			query: (formData) => ({
				method: "POST",
				url: `/oauth/token`,
				asForm: true,
				body: formData,
				discardEmpty: true
			}),
		}),
	})
});

export const {
	useLazySearchAppQuery,
	useCreateAppMutation,
	useGetAppQuery,
	useGetOOBAuthCodeMutation,
	useGetAccessTokenForAppMutation,
	useDeleteAppMutation,
} = extended;
