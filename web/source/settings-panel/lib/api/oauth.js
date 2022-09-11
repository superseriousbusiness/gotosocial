/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

const { OAUTHError } = require("../errors");

const oauth = require("../../redux/reducers/oauth").actions;
const temporary = require("../../redux/reducers/temporary").actions;

module.exports = function oauthAPI({apiCall, getCurrentUrl}) {
	return {

		register: function register(scopes = []) {
			return function (dispatch, getState) {
				return Promise.try(() => {
					return apiCall(getState(), "POST", "/api/v1/apps", {
						client_name: "GoToSocial Settings",
						scopes: scopes.join(" "),
						redirect_uris: getCurrentUrl(),
						website: getCurrentUrl()
					});
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
	
					return apiCall(getState(), "POST", "/oauth/token", {
						client_id: reg.client_id,
						client_secret: reg.client_secret,
						redirect_uri: getCurrentUrl(),
						grant_type: "authorization_code",
						code: code
					});
				}).then((json) => {
					console.log(json);
					window.history.replaceState({}, document.title, window.location.pathname);
					return dispatch(oauth.login(json));
				});
			};
		},
	
		verify: function verify() {
			return function (dispatch, getState) {
				console.log(getState());
				return Promise.try(() => {
					return apiCall(getState(), "GET", "/api/v1/accounts/verify_credentials");
				}).then((account) => {
					console.log(account);
				}).catch((e) => {
					dispatch(oauth.remove());
					throw e;
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