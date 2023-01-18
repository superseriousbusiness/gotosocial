/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

"use strict";

const Promise = require("bluebird");

const base = require("./base");
const { unwrapRes } = require("./lib");
const oauth = require("../../redux/oauth").actions;

function getSettingsURL() {
	/* needed in case the settings interface isn't hosted at /settings but
		 some subpath like /gotosocial/settings. Other parts of the code don't
		 take this into account yet so mostly future-proofing.

		 Also drops anything past /settings/, because authorization urls that are too long
		 get rejected by GTS.
	*/
	let [pre, _past] = window.location.pathname.split("/settings");
	return `${window.location.origin}${pre}/settings`;
}

const SETTINGS_URL = getSettingsURL();

const endpoints = (build) => ({
	verifyCredentials: build.query({
		providesTags: (_res, error) =>
			error == undefined
				? ["Auth"]
				: [],
		queryFn: (_arg, api, _extraOpts, baseQuery) => {
			const state = api.getState();

			return Promise.try(() => {
				// Process callback code first, if available
				if (state.oauth.loginState == "callback") {
					let urlParams = new URLSearchParams(window.location.search);
					let code = urlParams.get("code");

					if (code == undefined) {
						throw {
							message: "Waiting for callback, but no ?code= provided in url."
						};
					} else {
						let app = state.oauth.registration;

						if (app == undefined || app.client_id == undefined) {
							throw {
								message: "No stored registration data, can't finish login flow."
							};
						}

						return baseQuery({
							method: "POST",
							url: "/oauth/token",
							body: {
								client_id: app.client_id,
								client_secret: app.client_secret,
								redirect_uri: SETTINGS_URL,
								grant_type: "authorization_code",
								code: code
							}
						}).then(unwrapRes).then((token) => {
							// remove ?code= from url
							window.history.replaceState({}, document.title, window.location.pathname);
							api.dispatch(oauth.setToken(token));
						});
					}
				}
			}).then(() => {
				return baseQuery({
					url: `/api/v1/accounts/verify_credentials`
				});
			}).catch((e) => {
				return { error: e };
			});
		}
	}),
	authorizeFlow: build.mutation({
		queryFn: (formData, api, _extraOpts, baseQuery) => {
			let instance;
			const state = api.getState();

			return Promise.try(() => {
				if (!formData.instance.startsWith("http")) {
					formData.instance = `https://${formData.instance}`;
				}
				instance = new URL(formData.instance).origin;

				const stored = state.oauth.instance;
				if (stored?.instance == instance && stored.registration) {
					return stored.registration;
				}

				return baseQuery({
					method: "POST",
					baseUrl: instance,
					url: "/api/v1/apps",
					body: {
						client_name: "GoToSocial Settings",
						scopes: formData.scopes,
						redirect_uris: SETTINGS_URL,
						website: SETTINGS_URL
					}
				}).then(unwrapRes).then((app) => {
					app.scopes = formData.scopes;

					api.dispatch(oauth.authorize({
						instance: instance,
						registration: app,
						loginState: "callback",
						expectingRedirect: true
					}));

					return app;
				});
			}).then((app) => {
				let url = new URL(instance);
				url.pathname = "/oauth/authorize";
				url.searchParams.set("client_id", app.client_id);
				url.searchParams.set("redirect_uri", SETTINGS_URL);
				url.searchParams.set("response_type", "code");
				url.searchParams.set("scope", app.scopes);

				let redirectURL = url.toString();
				window.location.assign(redirectURL);

				return { data: null };
			}).catch((e) => {
				return { error: e };
			});
		},
	}),
	logout: build.mutation({
		queryFn: (_arg, api) => {
			api.dispatch(oauth.remove());
			return { data: null };
		},
		invalidatesTags: ["Auth"]
	})
});

module.exports = base.injectEndpoints({ endpoints });