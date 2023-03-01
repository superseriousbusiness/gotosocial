/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

const React = require("react");
const ReactDom = require("react-dom/client");
const { Provider } = require("react-redux");
const { PersistGate } = require("redux-persist/integration/react");
const { Switch, Route, Redirect } = require("wouter");

const query = require("./lib/query");

const { store, persistor } = require("./redux");
const AuthorizationGate = require("./components/authorization");
const Loading = require("./components/loading");

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
		"Federation": require("./admin/federation"),
		"Reports": require("./admin/reports")
	},
	"Custom Emoji": {
		adminOnly: true,
		"Local": require("./admin/emoji/local"),
		"Remote": require("./admin/emoji/remote"),
	}
};

const { sidebar, panelRouter } = require("./lib/get-views")(nav);

function App({ account }) {
	const isAdmin = account.role.name == "admin";
	const [logoutQuery] = query.useLogoutMutation();

	return (
		<>
			<div className="sidebar">
				{sidebar.all}
				{isAdmin && sidebar.admin}
				<button className="logout" onClick={logoutQuery}>
					Log out
				</button>
			</div>
			<section className="with-sidebar">
				<Switch>
					{panelRouter.all}
					{isAdmin && panelRouter.admin}
					<Route>
						<Redirect to="/settings/user" />
					</Route>
				</Switch>
			</section>
		</>
	);
}

function Main() {
	return (
		<Provider store={store}>
			<PersistGate loading={<section><Loading /></section>} persistor={persistor}>
				<AuthorizationGate App={App} />
			</PersistGate>
		</Provider>
	);
}

const root = ReactDom.createRoot(document.getElementById("root"));
root.render(<React.StrictMode><Main /></React.StrictMode>);