"use strict";

const Promise = require("bluebird");
const React = require("react");
const oauthLib = require("../../lib/oauth");

module.exports = function Auth({setOauth}) {
	const [ instance, setInstance ] = React.useState("");

	React.useEffect(() => {
		let isStillMounted = true;
		// check if current domain runs an instance
		let thisUrl = new URL(window.location.origin);
		thisUrl.pathname = "/api/v1/instance";
		Promise.try(() => {
			return fetch(thisUrl.href);
		}).then((res) => {
			if (res.status == 200) {
				return res.json();
			}
		}).then((json) => {
			if (json && json.uri && isStillMounted) {
				setInstance(json.uri);
			}
		}).catch((e) => {
			console.log("error checking instance response:", e);
		});

		return () => {
			// cleanup function
			isStillMounted = false;
		};
	}, []);

	function doAuth() {
		return Promise.try(() => {
			return new URL(instance);
		}).catch(TypeError, () => {
			return new URL(`https://${instance}`);
		}).then((parsedURL) => {
			let url = parsedURL.toString();
			let oauth = oauthLib({
				instance: url,
				client_name: "GoToSocial Admin Panel",
				scope: ["admin"],
				website: window.location.href
			});
			setOauth(oauth);
			setInstance(url);
			return oauth.register().then(() => {
				return oauth;
			});
		}).then((oauth) => {
			return oauth.authorize();
		}).catch((e) => {
			console.log("error authenticating:", e);
		});
	}

	function updateInstance(e) {
		if (e.key == "Enter") {
			doAuth();
		} else {
			setInstance(e.target.value);
		}
	}

	return (
		<section className="login">
			<h1>OAUTH Login:</h1>
			<form onSubmit={(e) => e.preventDefault()}>
				<label htmlFor="instance">Instance: </label>
				<input value={instance} onChange={updateInstance} id="instance"/>
				<button onClick={doAuth}>Authenticate</button>
			</form>
		</section>
	);
};