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
const { isPlainObject } = require("is-plain-object");

const { APIError } = require("../errors");
const { setInstanceInfo, setNamedInstanceInfo } = require("../../redux/reducers/instances").actions;
const oauth = require("../../redux/reducers/oauth").actions;

function apiCall(method, route, payload, type="json") {
	return function (dispatch, getState) {
		const state = getState();
		let base = state.oauth.instance;
		let auth = state.oauth.token;
		console.log(method, base, route, "auth:", auth != undefined);
	
		return Promise.try(() => {
			let url = new URL(base);
			url.pathname = route;
			let body;
	
			let headers = {
				"Accept": "application/json",
			};

			if (payload != undefined) {
				if (type == "json") {
					headers["Content-Type"] = "application/json";
					body = JSON.stringify(payload);
				} else if (type == "form") {
					const formData = new FormData();
					Object.entries(payload).forEach(([key, val]) => {
						if (isPlainObject(val)) {
							Object.entries(val).forEach(([key2, val2]) => {
								formData.set(`${key}[${key2}]`, val2);
							});
						} else {
							formData.set(key, val);
						}
					});
					body = formData;
				}
			}
	
			if (auth != undefined) {
				headers["Authorization"] = auth;
			}
	
			return fetch(url.toString(), {
				method,
				headers,
				body
			});
		}).then((res) => {
			// try parse json even with error
			let json = res.json().catch((e) => {
				throw new APIError(`JSON parsing error: ${e.message}`);
			});
	
			return Promise.all([res, json]);
		}).then(([res, json]) => {
			if (!res.ok) {
				if (auth != undefined && res.status == 401) {
					// stored access token is invalid
					dispatch(oauth.remove());
					throw new APIError("Stored OAUTH login was no longer valid, please log in again.");
				}
				throw new APIError(json.error, {json});
			} else {
				return json;
			}
		});
	};
}

function getCurrentUrl() {
	return `${window.location.origin}${window.location.pathname}`;
}

function fetchInstanceWithoutStore(domain) {
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

			return apiCall("GET", "/api/v1/instance")(dispatch, () => fakeState);
		}).then((json) => {
			if (json && json.uri) { // TODO: validate instance json more?
				dispatch(setNamedInstanceInfo([domain, json]));
				return json;
			}
		});
	};
}

function fetchInstance() {
	return function(dispatch, _getState) {
		return Promise.try(() => {
			return dispatch(apiCall("GET", "/api/v1/instance"));
		}).then((json) => {
			if (json && json.uri) {
				dispatch(setInstanceInfo(json));
				return json;
			}
		});
	};
}

module.exports = {
	instance: {
		fetchWithoutStore: fetchInstanceWithoutStore,
		fetch: fetchInstance
	},
	oauth: require("./oauth")({apiCall, getCurrentUrl}),
	user: require("./user")({apiCall}),
	apiCall
};