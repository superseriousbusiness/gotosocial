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
const ReactDom = require("react-dom/client");
const Redux = require("react-redux");
const { Switch, Route, Redirect } = require("wouter");
const { Provider } = require("react-redux");
const { PersistGate } = require("redux-persist/integration/react");

const { store, persistor } = require("./redux");
const api = require("./lib/api");
const oauth = require("./redux/reducers/oauth").actions;
const { AuthenticationError } = require("./lib/errors");

const Login = require("./components/login");

require("./style.css");

// TODO: nested categories?
const nav = {
	"User": {
		"Profile": require("./user/profile.js"),
		"Settings": require("./user/settings.js"),
	},
	"Admin": {
		adminOnly: true,
		"Instance Settings": require("./admin/settings.js"),
		"Actions": require("./admin/actions"),
		"Federation": require("./admin/federation.js"),
		"Custom Emoji": require("./admin/emoji"),
	}
};

const { sidebar, panelRouter } = require("./lib/get-views")(nav);

function App() {
	const dispatch = Redux.useDispatch();

	const { loginState, isAdmin } = Redux.useSelector((state) => state.oauth);
	const reduxTempStatus = Redux.useSelector((state) => state.temporary.status);

	const [errorMsg, setErrorMsg] = React.useState();
	const [tokenChecked, setTokenChecked] = React.useState(false);

	React.useEffect(() => {
		if (loginState == "login" || loginState == "callback") {
			Promise.try(() => {
				// Process OAUTH authorization token from URL if available
				if (loginState == "callback") {
					let urlParams = new URLSearchParams(window.location.search);
					let code = urlParams.get("code");

					if (code == undefined) {
						setErrorMsg(new Error("Waiting for OAUTH callback but no ?code= provided. You can try logging in again:"));
					} else {
						return dispatch(api.oauth.tokenize(code));
					}
				}
			}).then(() => {
				// Fetch current instance info
				return dispatch(api.instance.fetch());
			}).then(() => {
				// Check currently stored auth token for validity if available
				return dispatch(api.user.fetchAccount());
			}).then(() => {
				setTokenChecked(true);

				return dispatch(api.oauth.checkIfAdmin());
			}).catch((e) => {
				if (e instanceof AuthenticationError) {
					dispatch(oauth.remove());
					e.message = "Stored OAUTH token no longer valid, please log in again.";
				}
				setErrorMsg(e);
				console.error(e);
			});
		}
	}, [loginState, dispatch]);

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
		<button className="logout" onClick={() => { dispatch(api.oauth.logout()); }}>
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
					{sidebar.all}
					{isAdmin && sidebar.admin}
					{LogoutElement}
				</div>
				<section className="with-sidebar">
					{ErrorElement}
					<Switch>
						{panelRouter.all}
						{isAdmin && panelRouter.admin}
						<Route> {/* default route */}
							<Redirect to="/settings/user" />
						</Route>
					</Switch>
				</section>
			</>
		);
	} else if (loginState == "none") {
		return (
			<Login error={ErrorElement} />
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

const root = ReactDom.createRoot(document.getElementById("root"));
root.render(<React.StrictMode><Main /></React.StrictMode>);