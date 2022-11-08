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
const d = require("dotty");

const { APIError, AuthenticationError } = require("../errors");
const { setInstanceInfo, setNamedInstanceInfo } = require("../../redux/reducers/instances").actions;

function apiCall(method, route, payload, type = "json") {
	return function (dispatch, getState) {
		const state = getState();
		let base = state.oauth.instance;
		let auth = state.oauth.token;

		return Promise.try(() => {
			let url = new URL(base);
			let [path, query] = route.split("?");
			url.pathname = path;
			if (query != undefined) {
				url.search = query;
			}
			let body;

			let headers = {
				"Accept": "application/json",
			};

			if (payload != undefined) {
				if (type == "json") {
					headers["Content-Type"] = "application/json";
					body = JSON.stringify(payload);
				} else if (type == "form") {
					body = convertToForm(payload);
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
				if (auth != undefined && (res.status == 401 || res.status == 403)) {
					// stored access token is invalid
					throw new AuthenticationError("401: Authentication error", {json, status: res.status});
				} else {
					throw new APIError(json.error, { json });
				}
			} else {
				return json;
			}
		});
	};
}

/*
	Takes an object with (nested) keys, and transforms it into
	a FormData object to be sent over the API
*/
function convertToForm(payload) {
	const formData = new FormData();
	Object.entries(payload).forEach(([key, val]) => {
		if (isPlainObject(val)) {
			Object.entries(val).forEach(([key2, val2]) => {
				if (val2 != undefined) {
					formData.set(`${key}[${key2}]`, val2);
				}
			});
		} else {
			if (val != undefined) {
				formData.set(key, val);
			}
		}
	});
	return formData;
}

function getChanges(state, keys) {
	const { formKeys = [], fileKeys = [], renamedKeys = {} } = keys;
	const update = {};

	formKeys.forEach((key) => {
		let value = d.get(state, key);
		if (value == undefined) {
			return;
		}
		if (renamedKeys[key]) {
			key = renamedKeys[key];
		}
		d.put(update, key, value);
	});

	fileKeys.forEach((key) => {
		let file = d.get(state, `${key}File`);
		if (file != undefined) {
			if (renamedKeys[key]) {
				key = renamedKeys[key];
			}
			d.put(update, key, file);
		}
	});

	return update;
}

function getCurrentUrl() {
	let [pre, _past] = window.location.pathname.split("/settings");
	return `${window.location.origin}${pre}/settings`;
}

function fetchInstanceWithoutStore(domain) {
	return function (dispatch, getState) {
		return Promise.try(() => {
			let lookup = getState().instances.info[domain];
			if (lookup != undefined) {
				return lookup;
			}

			// apiCall expects to pull the domain from state,
			// but we don't want to store it there yet
			// so we mock the API here with our function argument
			let fakeState = {
				oauth: { instance: domain }
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
	return function (dispatch, _getState) {
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

let submoduleArgs = { apiCall, getCurrentUrl, getChanges };

module.exports = {
	instance: {
		fetchWithoutStore: fetchInstanceWithoutStore,
		fetch: fetchInstance
	},
	oauth: require("./oauth")(submoduleArgs),
	user: require("./user")(submoduleArgs),
	admin: require("./admin")(submoduleArgs),
	apiCall,
	convertToForm,
	getChanges
};