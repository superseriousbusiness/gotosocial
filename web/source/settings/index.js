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

const React = require("react");
const ReactDom = require("react-dom/client");
const { Provider } = require("react-redux");
const { PersistGate } = require("redux-persist/integration/react");

const { store, persistor } = require("./redux/store");
const { createNavigation, Menu, Item } = require("./lib/navigation");

const { Authorization } = require("./components/authorization");
const Loading = require("./components/loading");
const UserLogoutCard = require("./components/user-logout-card");
const { RoleContext } = require("./lib/navigation/util");

const UserProfile = require("./user/profile").default;
const UserSettings = require("./user/settings").default;
const UserMigration = require("./user/migration").default;

const Reports = require("./admin/reports").default;

const Accounts = require("./admin/accounts").default;
const AccountsPending = require("./admin/accounts/pending").default;

const DomainPerms = require("./admin/domain-permissions").default;
const DomainPermsImportExport = require("./admin/domain-permissions/import-export").default;

const AdminMedia = require("./admin/actions/media").default;
const AdminKeys =  require("./admin/actions/keys").default;

const LocalEmoji = require("./admin/emoji/local").default;
const RemoteEmoji = require("./admin/emoji/remote").default;

const InstanceSettings = require("./admin/settings").default;
const InstanceRules = require("./admin/settings/rules").default;

require("./style.css");

const { Sidebar, ViewRouter } = createNavigation("/settings", [
	Menu("User", [
		Item("Profile", { icon: "fa-user" }, UserProfile),
		Item("Settings", { icon: "fa-cogs" }, UserSettings),
		Item("Migration", { icon: "fa-exchange" }, UserMigration),
	]),
	Menu("Moderation", {
		url: "admin",
		permissions: ["admin"]
	}, [
		Item("Reports", { icon: "fa-flag", wildcard: true }, Reports),
		Item("Accounts", { icon: "fa-users", wildcard: true }, [
			Item("Overview", { icon: "fa-list", url: "", wildcard: true }, Accounts),
			Item("Pending", { icon: "fa-question", url: "pending", wildcard: true }, AccountsPending),
		]),
		Menu("Domain Permissions", { icon: "fa-hubzilla" }, [
			Item("Blocks", { icon: "fa-close", url: "block", wildcard: true }, DomainPerms),
			Item("Allows", { icon: "fa-check", url: "allow", wildcard: true }, DomainPerms),
			Item("Import/Export", { icon: "fa-floppy-o", url: "import-export", wildcard: true }, DomainPermsImportExport),
		]),
	]),
	Menu("Administration", {
		url: "admin",
		defaultUrl: "/settings/admin/settings",
		permissions: ["admin"]
	}, [
		Menu("Actions", { icon: "fa-bolt" }, [
			Item("Media", { icon: "fa-photo" }, AdminMedia),
			Item("Keys", { icon: "fa-key-modern" }, AdminKeys),
		]),
		Menu("Custom Emoji", { icon: "fa-smile-o" }, [
			Item("Local", { icon: "fa-home", wildcard: true }, LocalEmoji),
			Item("Remote", { icon: "fa-cloud" }, RemoteEmoji),
		]),
		Menu("Settings", { icon: "fa-sliders" }, [
			Item("Settings", { icon: "fa-sliders", url: "" }, InstanceSettings),
			Item("Rules", { icon: "fa-dot-circle-o", wildcard: true }, InstanceRules),
		]),
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
				<Authorization App={App} />
			</PersistGate>
		</Provider>
	);
}

const root = ReactDom.createRoot(document.getElementById("root"));
root.render(<React.StrictMode><Main /></React.StrictMode>);