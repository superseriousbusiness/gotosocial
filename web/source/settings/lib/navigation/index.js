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
const { nanoid } = require("nanoid");
const { Redirect } = require("wouter");

const { urlSafe } = require("./util");

const {
	Sidebar,
	ViewRouter,
	MenuComponent
} = require("./components");

function createNavigation(rootUrl, menus) {
	const root = {
		url: rootUrl
	};

	const routing = {
		view: {},
		fallback: {}
	};

	const menuTree = menus.map((creatorFunc) =>
		creatorFunc(root, routing)
	);

	return {
		Sidebar: Sidebar(menuTree),
		ViewRouter: ViewRouter(routing, root.redirectUrl)
	};
}

function Menu(name, opts, items) {
	if (items == undefined) { // opts argument is optional
		items = opts;
		opts = {};
	}

	const menu = {
		name,
		key: nanoid(),
		permissions: opts.permissions ?? true,
		url: opts.url ?? urlSafe(name),
		icon: opts.icon
	};

	return function _menu(root, routing) {
		if (menu.url != "") {
			menu.url = [root.url, menu.url].join("/");
		} else {
			menu.url = root.url;
		}

		if (root?.permissions && root.permissions !== true) {
			menu.permissions = root.permissions;
		}

		menu.links = [];
		menu.redirectUrl = menu.defaultUrl;
		menu.level = (root.level ?? -1) + 1;

		const contents = items.map((creatorFunc) => {
			return creatorFunc(menu, routing);
		});

		if (root.links != undefined) {
			root.links.push(...menu.links);
		}

		if (menu.redirectUrl != menu.url) {
			routing.fallback[menu.url] = {
				permissions: menu.permissions,
				view: (
					<Redirect to={menu.redirectUrl} />
				)
			};
			menu.url = menu.redirectUrl;
		}

		if (root.redirectUrl == undefined) {
			// first component in (sub)tree
			root.redirectUrl = menu.url;
		}

		return React.createElement(MenuComponent, menu, contents);
	};
}

function Item(name, view, opts) {
	const item = {
		name,
		key: nanoid(),
		permissions: opts.permissions ?? true,
		url: opts.url ?? urlSafe(name),
		icon: opts.icon
	};

	return function _Item(root, routing) {
		if (item.url == "") {
			item.url = root.url;
		} else {
			item.url = [root.url, item.url].join("/");
		}

		if (root?.permissions && root.permissions !== true) {
			item.permissions = root.permissions;
		}

		if (root.redirectUrl == undefined) {
			// first component in (sub)tree
			root.redirectUrl = item.url;
		}

		root.links.push(item.url);
		routing.view[item.url] = {
			permissions: item.permissions,
			view: React.createElement(view, { baseUrl: item.url })
		};

		return React.createElement(MenuComponent, item);
	};
}

module.exports = {
	createNavigation,
	Menu,
	Item
};