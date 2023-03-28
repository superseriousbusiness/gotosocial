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
const syncpipe = require("syncpipe");

function createNavigation(rootUrl, menus) {
	const root = {
		url: [rootUrl],
		routes: []
	};

	const routing = [];

	const menuTree = menus.map((creatorFunc) =>
		creatorFunc(root, routing)
	);

	console.log(routing);

	return {
		Sidebar: Sidebar(menuTree),
		ViewRouter: ViewRouter(sortRoutes(routing), root.redirectUrl)
	};
}

function sortRoutes(routing) {
	return syncpipe(routing, [
		(_) => Object.values(_),
		(_) => _.map((v) => Object.entries(v)),
		(_) => _.flat(1),
		(_) => _.sort(([_urlA, optsA], [_urlB, optsB]) => {
			// sort with most specific routes first (most path segments)
			let comp = optsB.path.length - optsA.path.length;
			if (comp == 0) {
				if (optsA.wildcard) {
					comp = 1;
				} else if (optsB.wildcard) {
					comp = -1;
				}
			}
			return comp;
		}),
		(_) => {
			_.forEach(([url, opts]) => console.log(url + (opts.wildcard ? "/*" : "")));
			return _;
		}
	]);
}

function MenuEntry(name, opts, contents) {
	if (contents == undefined) { // opts argument is optional
		contents = opts;
		opts = {};
	}

	return function createMenuEntry(root, routing) {
		const type = Array.isArray(contents) ? "category" : "page";

		const path = opts.url ?? urlSafe(name);
		const url = [...root.url, path];

		const entry = {
			name, path, url, type,
			key: nanoid(),
			permissions: opts.permissions ?? true,
			icon: opts.icon,
			links: []
		};

		if (type == "category") {
			let entries = contents.map((creatorFunc) => creatorFunc(entry, routing));
			let routes = [];

			entries.forEach((e) => {
				if (e.path == "") {
					routes.unshift(e);
				} else {
					routes.push(e);
				}
			});
			routes.reverse();

			entry.routes = routes;
			routing.push(...routes);

			console.log("name:", name, type, "routes:", routes);
			console.log("  contents:", routes);
		} else {
			if (opts.wildcard) {
				url.push(":wildcard*");
			}
		}

		return entry;
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
		icon: opts.icon,
		links: [],
		wildcardLinks: []
	};

	return function _menu(root, routing) {
		menu.path = [...root.path];
		if (menu.url != "") {
			menu.path.push(menu.url);
		}
		menu.url = menu.path.join("/");

		if (root?.permissions && root.permissions !== true) {
			menu.permissions = root.permissions;
		}

		menu.redirectUrl = menu.defaultUrl;
		menu.level = (root.level ?? -1) + 1;

		const contents = items.map((creatorFunc) => {
			return creatorFunc(menu, routing);
		});

		if (root.links != undefined) {
			root.links.push(...menu.links);
		}

		if (root.wildcardLinks != undefined) {
			root.wildcardLinks.push(...menu.wildcardLinks);
		}

		if (menu.redirectUrl != menu.url) {
			routing.fallback[menu.url] = {
				permissions: menu.permissions,
				path: menu.path,
				routeUrl: `${menu.url}/:page*`,
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
		item.path = [...root.path];
		if (item.url != "") {
			item.path.push(item.url);
		}

		item.url = item.path.join("/");

		if (root?.permissions && root.permissions !== true) {
			item.permissions = root.permissions;
		}

		if (root.redirectUrl == undefined) {
			// first component in (sub)tree
			root.redirectUrl = item.url;
		}

		item.level = (root.level ?? -1) + 1;

		root.links.push(item.url);
		if (opts.wildcard) {
			item.wildcardLinks = [item.url];
			root.wildcardLinks.push(item.url);
		}

		item.routeUrl = opts.wildcard ? `${item.url}/:page*` : item.url;

		routing.view[item.url] = {
			path: item.path,
			permissions: item.permissions,
			routeUrl: item.routeUrl,
			view: React.createElement(view, { baseUrl: item.url })
		};

		return React.createElement(MenuComponent, item);
	};
}

module.exports = {
	createNavigation,
	Menu: MenuEntry,
	Item: MenuEntry
};