/*
	GoToSocial
	Copyright (C) GoToSocial Authors admin@gotosocial.org
	SPDX-License-Identifier: AGPL-3.0-or-later

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

const { store, persistor } = require("./redux");
const { createNavigation, useNavigation } = require("./lib/navigation");

const AuthorizationGate = require("./components/authorization");
const Loading = require("./components/loading");
const UserLogoutCard = require("./components/user-logout-card");

require("./style.css");

const navigation = createNavigation("/settings", ({ Category, View }) => {
	return [
		Category("User", [
			View("Profile", require("./user/profile"), { icon: "fa-user" }),
			View("Settings", require("./user/settings"), { icon: "fa-cogs" }),
		]),
		Category("Moderation", {
			url: "admin",
			permissions: ["admin"]
		}, [
			View("Reports", require("./admin/reports"), { icon: "fa-flag" }),
			View("Users", require("./admin/reports"), { icon: "fa-users" }),
			Category("Federation", { icon: "fa-hubzilla" }, [
				View("Federation", require("./admin/federation"), { icon: "fa-hubzilla", url: "" }),
				View("Bulk Import/Export", require("./admin/federation/import-export"), { icon: "fa-floppy-o" }),
			])
		]),
		Category("Administration", {
			url: "admin",
			defaultUrl: "/settings/admin/settings",
			permissions: ["admin"]
		}, [
			View("Actions", require("./admin/actions"), { icon: "fa-bolt" }),
			Category("Custom Emoji", { icon: "fa-smile-o" }, [
				View("Local", require("./admin/emoji/local"), { icon: "fa-home" }),
				View("Remote", require("./admin/emoji/remote"), { icon: "fa-cloud" })
			]),
			View("Settings", require("./admin/settings"), { icon: "fa-sliders" })
		])
	];
});

function App({ account }) {
	const { sidebar, routedViews, fallbackRoutes } = useNavigation(navigation, {
		permissions: [account.role.name]
	});

	return (
		<>
			<div className="sidebar">
				<UserLogoutCard />
				{sidebar}
				{/* <div className="nav-container">
					{sidebar.all}
					{isAdmin && sidebar.adminOnly}
				</div> */}
			</div>
			<section className="with-sidebar">
				<Switch>
					{/* {viewRouter.all}
					{isAdmin && viewRouter.adminOnly} */}
					{routedViews}
					{fallbackRoutes}
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