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

const { store, persistor } = require("./redux");
const { createNavigation, Menu, Item } = require("./lib/navigation");

const AuthorizationGate = require("./components/authorization");
const Loading = require("./components/loading");
const UserLogoutCard = require("./components/user-logout-card");
const { RoleContext } = require("./lib/navigation/util");

require("./style.css");

const { Sidebar, ViewRouter } = createNavigation("/settings", [
	Menu("User", [
		Item("Profile", require("./user/profile"), { icon: "fa-user" }),
		Item("Settings", require("./user/settings"), { icon: "fa-cogs" }),
	]),
	Menu("Moderation", {
		url: "admin",
		permissions: ["admin"]
	}, [
		Item("Reports", require("./admin/reports"), { icon: "fa-flag" }),
		Item("Users", require("./admin/reports"), { icon: "fa-users" }),
		Menu("Federation", { icon: "fa-hubzilla" }, [
			Item("Federation", require("./admin/federation"), { icon: "fa-hubzilla", url: "" }),
			Item("Import/Export", require("./admin/federation/import-export"), { icon: "fa-floppy-o" }),
		])
	]),
	Menu("Administration", {
		url: "admin",
		defaultUrl: "/settings/admin/settings",
		permissions: ["admin"]
	}, [
		Item("Actions", require("./admin/actions"), { icon: "fa-bolt" }),
		Menu("Custom Emoji", { icon: "fa-smile-o" }, [
			Item("Local", require("./admin/emoji/local"), { icon: "fa-home" }),
			Item("Remote", require("./admin/emoji/remote"), { icon: "fa-cloud" })
		]),
		Item("Settings", require("./admin/settings"), { icon: "fa-sliders" })
	])
]);

function App({ account }) {
	const permissions = [account.role.name];

	return (
		<RoleContext.Provider value={permissions}>
			<div className="sidebar">
				<UserLogoutCard />
				<Sidebar />
			</div>
			<section className="with-sidebar">
				<ViewRouter />
			</section>
		</RoleContext.Provider>
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