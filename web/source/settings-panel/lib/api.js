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
const { setRegistration } = require("../redux/reducers/oauth").actions;
const { setInstanceInfo } = require("../redux/reducers/instances").actions;

function apiCall(base, method, route, {payload, headers={}}) {
	return Promise.try(() => {
		let url = new URL(base);
		url.pathname = route;
		let body;

		if (payload != undefined) {
			body = JSON.stringify(payload);
		}

		let fetchHeaders = {
			"Content-Type": "application/json",
			...headers
		};

		return fetch(url.toString(), {
			method: method,
			headers: fetchHeaders,
			body: body
		});
	}).then((res) => {
		if (res.status == 200) {
			return res.json();
		} else {
			throw res;
		}
	});
}

function getCurrentUrl() {
	return `${window.location.origin}${window.location.pathname}`;
}

function updateInstance(domain) {
	return function(dispatch, getState) {
		/* check if domain is valid instance, then register client if needed  */

		return Promise.try(() => {
			return apiCall(domain, "GET", "/api/v1/instance", {
				headers: {
					"Content-Type": "text/plain"
				}
			});
		}).then((json) => {
			if (json && json.uri) { // TODO: validate instance json more?
				dispatch(setInstanceInfo(json.uri, json));
				return json;
			}
		});
	};
}

function updateRegistration() {
	return function(dispatch, getState) {
		let base = getState().oauth.instance;
		return Promise.try(() => {
			return apiCall(base, "POST", "/api/v1/apps", {
				client_name: "GoToSocial Settings",
				scopes: "write admin",
				redirect_uris: getCurrentUrl(),
				website: getCurrentUrl()
			});
		}).then((json) => {
			console.log(json);
			dispatch(setRegistration(base, json));
		});
	};
}

module.exports = { updateInstance, updateRegistration };