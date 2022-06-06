"use strict";

const Promise = require("bluebird");

function getCurrentUrl() {
	return window.location.origin + window.location.pathname; // strips ?query=string and #hash
}

module.exports = function oauthClient(config, initState) {
	/* config: 
		instance: instance domain (https://testingtesting123.xyz)
		client_name: "GoToSocial Admin Panel"
		scope: []
		website: 
	*/

	let state = initState;
	if (initState == undefined) {
		state = localStorage.getItem("oauth");
		if (state == undefined) {
			state = {
				config
			};
			storeState();
		} else {
			state = JSON.parse(state);
		}
	}

	function storeState() {
		localStorage.setItem("oauth", JSON.stringify(state));
	}

	/* register app
		/api/v1/apps
	*/
	function register() {
		if (state.client_id != undefined) {
			return true; // we already have a registration
		}
		let url = new URL(config.instance);
		url.pathname = "/api/v1/apps";

		return fetch(url.href, {
			method: "POST",
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				client_name: config.client_name,
				redirect_uris: getCurrentUrl(),
				scopes: config.scope.join(" "),
				website: getCurrentUrl()
			})
		}).then((res) => {
			if (res.status != 200) {
				throw res;
			}
			return res.json();
		}).then((json) => {
			state.client_id = json.client_id;
			state.client_secret = json.client_secret;
			storeState();
		});
	}
	
	/* authorize:
		/oauth/authorize
			?client_id=CLIENT_ID
			&redirect_uri=window.location.href
			&response_type=code
			&scope=admin
	*/
	function authorize() {
		let url = new URL(config.instance);
		url.pathname = "/oauth/authorize";
		url.searchParams.set("client_id", state.client_id);
		url.searchParams.set("redirect_uri", getCurrentUrl());
		url.searchParams.set("response_type", "code");
		url.searchParams.set("scope", config.scope.join(" "));

		window.location.assign(url.href);
	}
	
	function callback() {
		if (state.access_token != undefined) {
			return; // we're already done :)
		}
		let params = (new URL(window.location)).searchParams;
	
		let token = params.get("code");
		if (token != null) {
			console.log("got token callback:", token);
		}

		return authorizeToken(token)
			.catch((e) => {
				console.log("Error processing oauth callback:", e);
				logout(); // just to be sure
			});
	}

	function authorizeToken(token) {
		let url = new URL(config.instance);
		url.pathname = "/oauth/token";
		return fetch(url.href, {
			method: "POST",
			headers: {
				"Content-Type": "application/json"
			},
			body: JSON.stringify({
				client_id: state.client_id,
				client_secret: state.client_secret,
				redirect_uri: getCurrentUrl(),
				grant_type: "authorization_code",
				code: token
			})
		}).then((res) => {
			if (res.status != 200) {
				throw res;
			}
			return res.json();
		}).then((json) => {
			state.access_token = json.access_token;
			storeState();
			window.location = getCurrentUrl(); // clear ?token=
		});
	}

	function isAuthorized() {
		return (state.access_token != undefined);
	}

	function apiRequest(path, method, data, type="json") {
		if (!isAuthorized()) {
			throw new Error("Not Authenticated");
		}
		let url = new URL(config.instance);
		let [p, s] = path.split("?");
		url.pathname = p;
		url.search = s;
		let headers = {
			"Authorization": `Bearer ${state.access_token}`
		};
		let body = data;
		if (type == "json" && body != undefined) {
			headers["Content-Type"] = "application/json";
			body = JSON.stringify(data);
		}
		return fetch(url.href, {
			method,
			headers,
			body
		}).then((res) => {
			return Promise.all([res.json(), res]);
		}).then(([json, res]) => {
			if (res.status != 200) {
				if (json.error) {
					throw new Error(json.error);
				} else {
					throw new Error(`${res.status}: ${res.statusText}`);
				}
			} else {
				return json;
			}
		});
	}

	function logout() {
		let url = new URL(config.instance);
		url.pathname = "/oauth/revoke";
		return fetch(url.href, {
			method: "POST",
			headers: {
				"Content-Type": "application/json"
			},
			body: JSON.stringify({
				client_id: state.client_id,
				client_secret: state.client_secret,
				token: state.access_token,
			})
		}).then((res) => {
			if (res.status != 200) {
				// GoToSocial doesn't actually implement this route yet,
				// so error is to be expected
				return;
			}
			return res.json();
		}).catch(() => {
			// see above
		}).then(() => {
			localStorage.removeItem("oauth");
			window.location = getCurrentUrl();
		});
	}

	return {
		register, authorize, callback, isAuthorized, apiRequest, logout
	};
};
