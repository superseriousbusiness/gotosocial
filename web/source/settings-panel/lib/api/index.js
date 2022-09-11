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

const { APIError } = require("../errors");
const { setInstanceInfo } = require("../../redux/reducers/instances").actions;

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

module.exports = {
	instance: {
		fetch: fetchInstance
	},
	oauth: require("./oauth")({apiCall, getCurrentUrl})
};