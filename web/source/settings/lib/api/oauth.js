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

const { OAUTHError, AuthenticationError } = require("../errors");

const oauth = require("../../redux/reducers/oauth").actions;
const temporary = require("../../redux/reducers/temporary").actions;
const admin = require("../../redux/reducers/admin").actions;

module.exports = function oauthAPI({ apiCall, getCurrentUrl }) {
	return {

		register: function register(scopes = []) {
			return function (dispatch, _getState) {
				return Promise.try(() => {
					return dispatch(apiCall("POST", "/api/v1/apps", {
						client_name: "GoToSocial Settings",
						scopes: scopes.join(" "),
						redirect_uris: getCurrentUrl(),
						website: getCurrentUrl()
					}));
				}).then((json) => {
					json.scopes = scopes;
					dispatch(oauth.setRegistration(json));
				});
			};
		},

		authorize: function authorize() {
			return function (dispatch, getState) {
				let state = getState();
				let reg = state.oauth.registration;
				let base = new URL(state.oauth.instance);

				base.pathname = "/oauth/authorize";
				base.searchParams.set("client_id", reg.client_id);
				base.searchParams.set("redirect_uri", getCurrentUrl());
				base.searchParams.set("response_type", "code");
				base.searchParams.set("scope", reg.scopes.join(" "));

				dispatch(oauth.setLoginState("callback"));
				dispatch(temporary.setStatus("Redirecting to instance login..."));

				// send user to instance's login flow
				window.location.assign(base.href);
			};
		},

		tokenize: function tokenize(code) {
			return function (dispatch, getState) {
				let reg = getState().oauth.registration;

				return Promise.try(() => {
					if (reg == undefined || reg.client_id == undefined) {
						throw new OAUTHError("Callback code present, but no client registration is available from localStorage. \nNote: localStorage is unavailable in Private Browsing.");
					}

					return dispatch(apiCall("POST", "/oauth/token", {
						client_id: reg.client_id,
						client_secret: reg.client_secret,
						redirect_uri: getCurrentUrl(),
						grant_type: "authorization_code",
						code: code
					}));
				}).then((json) => {
					window.history.replaceState({}, document.title, window.location.pathname);
					return dispatch(oauth.login(json));
				});
			};
		},

		checkIfAdmin: function checkIfAdmin() {
			return function (dispatch, getState) {
				const state = getState();
				let stored = state.oauth.isAdmin;
				if (stored != undefined) {
					return stored;
				}

				// newer GoToSocial version will include a `role` in the Account data, check that first
				if (state.user.profile.role == "admin") {
					dispatch(oauth.setAdmin(true));
					return true;
				}

				// no role info, try fetching an admin-only route and see if we get an error
				return Promise.try(() => {
					return dispatch(apiCall("GET", "/api/v1/admin/domain_blocks"));
				}).then((data) => {
					return Promise.all([
						dispatch(oauth.setAdmin(true)),
						dispatch(admin.setBlockedInstances(data))
					]);
				}).catch(AuthenticationError, () => {
					return dispatch(oauth.setAdmin(false));
				});
			};
		},

		logout: function logout() {
			return function (dispatch, _getState) {
				// TODO: GoToSocial does not have a logout API route yet

				return dispatch(oauth.remove());
			};
		}
	};
};