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

const Login = require("./components/login");
const ErrorFallback = require("./components/error");

const oauthLib = require("./lib/oauth");

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
	const { loggedIn } = Redux.useSelector((state) => state.oauth);

	// const [oauth, setOauth] = React.useState();
	// const [hasAuth, setAuth] = React.useState(false);
	// const [oauthState, _setOauthState] = React.useState(localStorage.getItem("oauth"));

	// React.useEffect(() => {
	// 	let state = localStorage.getItem("oauth");
	// 	if (state != undefined) {
	// 		state = JSON.parse(state);
	// 		let restoredOauth = oauthLib(state.config, state);
	// 		Promise.try(() => {
	// 			return restoredOauth.callback();
	// 		}).then(() => {
	// 			setAuth(true);
	// 		});
	// 		setOauth(restoredOauth);
	// 	}
	// }, [setAuth, setOauth]);

	// if (!hasAuth && oauth && oauth.isAuthorized()) {
	// 	setAuth(true);
	// }

	if (loggedIn) {
		return (
			<>
				<div className="sidebar">
					{sidebar}
					{/* <button className="logout" onClick={oauth.logout}>Log out</button> */}
				</div>
				<section className="with-sidebar">
					<Switch>
						{panelRouter}
					</Switch>
				</section>
			</>
		);
	} else {
		return (
			<Login />
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