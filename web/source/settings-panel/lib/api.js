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

const { APIError, OAUTHError } = require("./errors");
const oauth = require("../redux/reducers/oauth").actions;
const temporary = require("../redux/reducers/temporary").actions;
const { setInstanceInfo } = require("../redux/reducers/instances").actions;

function apiCall(state, method, route, payload) {
	let base = state.oauth.instance;
	let auth = state.oauth.token;
	console.log(method, base, route, auth);

	return Promise.try(() => {
		let url = new URL(base);
		url.pathname = route;
		let body;

		if (payload != undefined) {
			body = JSON.stringify(payload);
		}

		let headers = {
			"Accept": "application/json",
			"Content-Type": "application/json"
		};

		if (auth != undefined) {
			headers["Authorization"] = auth;
		}

		return fetch(url.toString(), {
			method,
			headers,
			body
		});
	}).then((res) => {
		let ok = res.ok;

		// try parse json even with error
		let json = res.json().catch((e) => {
			throw new APIError(`JSON parsing error: ${e.message}`);
		});

		return Promise.all([ok, json]);
	}).then(([ok, json]) => {
		if (!ok) {
			throw new APIError(json.error, {json});
		} else {
			return json;
		}
	});
}

function getCurrentUrl() {
	return `${window.location.origin}${window.location.pathname}`;
}

function fetchInstance(domain) {
	return function(dispatch, getState) {
		return Promise.try(() => {
			let lookup = getState().instances.info[domain];
			if (lookup != undefined) {
				return lookup;
			}

			// apiCall expects to pull the domain from state,
			// but we don't want to store it there yet
			// so we mock the API here with our function argument
			let fakeState = {
				oauth: {instance: domain}
			};

			return apiCall(fakeState, "GET", "/api/v1/instance");
		}).then((json) => {
			if (json && json.uri) { // TODO: validate instance json more?
				dispatch(setInstanceInfo([json.uri, json]));
				return json;
			}
		});
	};
}

function fetchRegistration(scopes=[]) {
	return function(dispatch, getState) {
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
}

function startAuthorize() {
	return function(dispatch, getState) {
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
}

function fetchToken(code) {
	return function(dispatch, getState) {
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
}

function verifyAuth() {
	return function(dispatch, getState) {
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
}

function oauthLogout() {
	return function(dispatch, _getState) {
		// TODO: GoToSocial does not have a logout API route yet

		return dispatch(oauth.remove());
	};
}

module.exports = {
	instance: {
		fetch: fetchInstance
	},
	oauth: {
		register: fetchRegistration,
		authorize: startAuthorize,
		fetchToken,
		verify: verifyAuth,
		logout: oauthLogout
	}
};