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
		url: rootUrl,
		links: [],
	};

	const routing = [];

	const menuTree = menus.map((creatorFunc) =>
		creatorFunc(root, routing)
	);

	return {
		Sidebar: Sidebar(menuTree, routing),
		ViewRouter: ViewRouter(routing, root.redirectUrl)
	};
}

function MenuEntry(name, opts, contents) {
	if (contents == undefined) { // opts argument is optional
		contents = opts;
		opts = {};
	}

	return function createMenuEntry(root, routing) {
		const type = Array.isArray(contents) ? "category" : "view";

		let urlParts = [root.url];
		if (opts.url != "") {
			urlParts.push(opts.url ?? urlSafe(name));
		}

		const url = urlParts.join("/");
		let routingUrl = url;

		if (opts.wildcard) {
			routingUrl += "/:wildcard*";
		}

		const entry = {
			name, type,
			url, routingUrl,
			key: nanoid(),
			permissions: opts.permissions ?? false,
			icon: opts.icon,
			links: [routingUrl],
			level: (root.level ?? -1) + 1,
			redirectUrl: opts.defaultUrl
		};

		if (type == "category") {
			let entries = contents.map((creatorFunc) => creatorFunc(entry, routing));
			let routes = [];

			entries.forEach((e) => {
				// move empty wildcard routes to end of category, to prevent overlap
				if (e.url == entry.url) {
					routes.unshift(e);
				} else {
					routes.push(e);
				}
			});
			routes.reverse();

			routing.push(...routes);

			if (opts.redirectUrl != entry.url) {
				routing.push({
					key: entry.key,
					url: entry.url,
					permissions: entry.permissions,
					routingUrl: entry.redirectUrl + "/:fallback*",
					view: React.createElement(Redirect, { to: entry.redirectUrl })
				});
				entry.url = entry.redirectUrl;
			}

			root.links.push(...entry.links);

			entry.MenuEntry = React.createElement(
				MenuComponent,
				entry,
				entries.map((e) => e.MenuEntry)
			);
		} else {
			entry.links.push(routingUrl);
			root.links.push(routingUrl);

			entry.view = React.createElement(contents, { baseUrl: url });
			entry.MenuEntry = React.createElement(MenuComponent, entry);
		}

		if (root.redirectUrl == undefined) {
			root.redirectUrl = entry.url;
		}

		return entry;
	};
}

module.exports = {
	createNavigation,
	Menu: MenuEntry,
	Item: MenuEntry
};