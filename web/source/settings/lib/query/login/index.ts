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

import type { FetchBaseQueryError } from '@reduxjs/toolkit/query';

import { gtsApi } from "../gts-api";
import {
	setToken as oauthSetToken,
	remove as oauthRemove,
	authorize as oauthAuthorize,
} from "../../../redux/login";
import { RootState } from '../../../redux/store';
import { Account } from '../../types/account';
import { OAuthAccessTokenRequestBody } from '../../types/oauth';
import { App } from '../../types/application';

function getSettingsURL() {
	/*
		needed in case the settings interface isn't hosted at /settings but
		some subpath like /gotosocial/settings. Other parts of the code don't
		take this into account yet so mostly future-proofing.

		 Also drops anything past /settings/, because authorization urls that are too long
		 get rejected by GTS.
	*/
	const [pre, _past] = window.location.pathname.split("/settings");
	return `${window.location.origin}${pre}/settings`;
}

const SETTINGS_URL = (getSettingsURL());

// Couple auth functions here require multiple requests as
// part of an OAuth token 'flow'. To keep things simple for
// callers of these query functions, the multiple requests
// are chained within one query.
//
// https://redux-toolkit.js.org/rtk-query/usage/customizing-queries#performing-multiple-requests-with-a-single-query
const extended = gtsApi.injectEndpoints({
	endpoints: (build) => ({
		verifyCredentials: build.query<Account, void>({
			providesTags: (_res, error) =>
				error == undefined ? ["Auth"] : [],
			async queryFn(_arg, api, _extraOpts, fetchWithBQ) {
				const state = api.getState() as RootState;
				const loginState = state.login;

				// If we're not in the middle of an auth/callback,
				// we may already have an auth token, so just
				// return a standard verify_credentials query.
				if (loginState.current != 'awaitingcallback') {
					return fetchWithBQ({
						url: `/api/v1/accounts/verify_credentials`
					});
				}

				// We're in the middle of an auth/callback flow.
				// Try to retrieve callback code from URL query.
				const urlParams = new URLSearchParams(window.location.search);
				const code = urlParams.get("code");
				if (code == undefined) {
					return {
						error: {
							status: 400,
							statusText: 'Bad Request',
							data: {"error":"Waiting for callback, but no ?code= provided in url."},
						},
					};
				}
				
				// Retrieve app with which the
				// callback code was generated.
				const app = loginState.app;
				if (app == undefined || app.client_id == undefined) {
					return {
						error: {
							status: 400,
							statusText: 'Bad Request',
							data: {"error":"No stored app registration data, can't finish login flow."},
						},
					};
				}
				
				// Use the provided code and app
				// secret to request an auth token.
				const tokenReqBody: OAuthAccessTokenRequestBody = {
					client_id: app.client_id,
					client_secret: app.client_secret,
					redirect_uri: SETTINGS_URL,
					grant_type: "authorization_code",
					code: code
				};

				const tokenResult = await fetchWithBQ({
					method: "POST",
					url: "/oauth/token",
					body: tokenReqBody,
				});
				if (tokenResult.error) {
					return { error: tokenResult.error as FetchBaseQueryError };
				}
				
				// Remove ?code= query param from
				// url, we don't want it anymore.
				window.history.replaceState({}, document.title, window.location.pathname);
				
				// Store returned token in redux.
				api.dispatch(oauthSetToken(tokenResult.data));
				
				// We're now authed! So return
				// standard verify_credentials query.
				return fetchWithBQ({
					url: `/api/v1/accounts/verify_credentials`
				});
			}
		}),

		authorizeFlow: build.mutation<any, { instance: string, scopes: string }>({
			async queryFn(formData, api, _extraOpts, fetchWithBQ) {
				const state = api.getState() as RootState;
				const loginState = state.login;

				let instanceUrl: string;
				if (!formData.instance.startsWith("http")) {
					formData.instance = `https://${formData.instance}`;
				}

				instanceUrl = new URL(formData.instance).origin;
				if (loginState?.instanceUrl == instanceUrl && loginState.app) {
					return { data: loginState.app };
				}
				
				const appResult = await fetchWithBQ({
					method: "POST",
					baseUrl: instanceUrl,
					url: "/api/v1/apps",
					body: {
						client_name: "GoToSocial Settings",
						scopes: formData.scopes,
						redirect_uris: SETTINGS_URL,
						website: SETTINGS_URL
					}
				});
				if (appResult.error) {
					return { error: appResult.error as FetchBaseQueryError };
				}

				const app = appResult.data as App;
				api.dispatch(oauthAuthorize({
					instanceUrl: instanceUrl,
					app: {
						client_id: app.client_id,
						client_secret: app.client_secret,
					},
					current: "awaitingcallback",
					expectingRedirect: true
				}));

				// Parse instance URL + set params on it.
				//
				// Note that scopes are '+'-separated to fit the API.
				const url = new URL(instanceUrl);
				url.pathname = "/oauth/authorize";
				url.searchParams.set("client_id", app.client_id);
				url.searchParams.set("redirect_uri", SETTINGS_URL);
				url.searchParams.set("response_type", "code");
				url.searchParams.set("scope", app.scopes.join("+"));
				
				const redirectURL = url.toString();
				window.location.assign(redirectURL);
				return { data: null };
			},
		}),
		logout: build.mutation({
			async queryFn(_arg, api, _extraOpts, fetchWithBQ) {
				const state = api.getState() as RootState;
				const loginState = state.login;
				
				// Try to log out politely by revoking
				// our access token. First fetch app,
				// then token, then post to /oauth/revoke.

				const app = loginState.app;
				if (app === undefined) {
					// This should never happen.
					throw "trying to log out with undefined app";
				}
				
				let token = loginState.token;
				if (token === undefined) {
					// This should never happen.
					throw "trying to log out with undefined token";
				}

				// Trim "Bearer " from stored token
				// to get just the access token part.
				token = token.substring(7);

				// Try to revoke the token. If we fail, just
				// log the error and clear our state anyway.
				const invalidateResult = await fetchWithBQ({
					method: "POST",
					url: "/oauth/revoke",
					body: {
						token: token,
						client_id: app.client_id,
						client_secret: app.client_secret,
					},
					asForm: true,
				});
				if (invalidateResult.error) {
					// eslint-disable-next-line no-console
					console.error("error logging out: ", invalidateResult.error);
				}

				// Clear our state.
				api.dispatch(oauthRemove());
				return { data: null };
			},
			invalidatesTags: ["Auth"]
		})
	})
});

export const {
	useVerifyCredentialsQuery,
	useAuthorizeFlowMutation,
	useLogoutMutation,
} = extended;
