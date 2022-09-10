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
const React = require("react");
const ReactDom = require("react-dom");
const Redux = require("react-redux");
const { Switch } = require("wouter");
const { Provider } = require("react-redux");
const { PersistGate } = require("redux-persist/integration/react");

const { store, persistor } = require("./redux");
const api = require("./lib/api");

const Login = require("./components/login");

require("./style.css");

// TODO: nested categories?
const nav = {
	"User": {
		Component: require("./user"),
		entries: {
			"Profile": require("./user/profile.js"),
			"Settings": require("./user/settings.js"),
			"Customization": require("./user/customization.js")
		}
	},
	"Admin": {
		Component: require("./admin"),
		entries: {
			"Instance Settings": require("./admin/settings.js"),
			"Federation": require("./admin/federation.js"),
			"Customization": require("./admin/customization.js")
		}
	}
};

// Generate component tree from `nav` object once, as it won't change
const { sidebar, panelRouter } = require("./lib/generate-views")(nav);

function App() {
	const dispatch = Redux.useDispatch();
	const { loginState } = Redux.useSelector((state) => state.oauth);
	const reduxTempStatus = Redux.useSelector((state) => state.temporary.status);
	const [ errorMsg, setErrorMsg ] = React.useState();
	const [ tokenChecked, setTokenChecked ] = React.useState(false);

	React.useEffect(() => {
		Promise.try(() => {
			// Process OAUTH authorization token from URL if available
			if (loginState == "callback") {
				let urlParams = new URLSearchParams(window.location.search);
				let code = urlParams.get("code");
	
				if (code == undefined) {
					setErrorMsg(new Error("Waiting for OAUTH callback but no ?code= provided. You can try logging in again:"));
				} else {
					return dispatch(api.oauth.fetchToken(code));
				}
			}
		}).then(() => {
			// Check currently stored auth token for validity if available
			if (loginState == "callback" || loginState == "login") {
				return dispatch(api.oauth.verify());
			}
		}).then(() => {
			setTokenChecked(true);
		}).catch((e) => {
			setErrorMsg(e);
			console.error(e.message);
		});
	}, []);

	let ErrorElement = null;
	if (errorMsg != undefined) {
		ErrorElement = (
			<div className="error">
				<b>{errorMsg.type}</b>
				<span>{errorMsg.message}</span>
			</div>
		);
	}

	const LogoutElement = (
		<button className="logout" onClick={() => {dispatch(api.oauth.logout());}}>
			Log out
		</button>
	);

	if (reduxTempStatus != undefined) {
		return (
			<section>
				{reduxTempStatus}
			</section>
		);
	} else if (tokenChecked && loginState == "login") {
		return (
			<>
				<div className="sidebar">
					{sidebar}
					{LogoutElement}
				</div>
				<section className="with-sidebar">
					{ErrorElement}
					<Switch>
						{panelRouter}
					</Switch>
				</section>
			</>
		);
	} else if (loginState == "none") {
		return (
			<Login error={ErrorElement}/>
		);
	} else {
		let status;
		
		if (loginState == "login") {
			status = "Verifying stored login...";
		} else if (loginState == "callback") {
			status = "Processing OAUTH callback...";
		}

		return (
			<section>
				<div>
					{status}
				</div>
				{ErrorElement}
				{LogoutElement}
			</section>
		);
	}

}

function Main() {
	return (
		<Provider store={store}>
			<PersistGate loading={"loading..."} persistor={persistor}>
				<App />
			</PersistGate>
		</Provider>
	);
}

ReactDom.render(<React.StrictMode><Main /></React.StrictMode>, document.getElementById("root"));