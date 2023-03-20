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
const { Redirect } = require("wouter");

function createNavigation(rootUrl, creatorFunc) {
	const Types = {
		Category: parseType("Category"),
		View: parseType("View")
	};

	return recurseNodes(creatorFunc(Types), rootUrl);
}

function recurseNodes(nodes, url, { views = [], links = [], fallback = {}, permissions } = {}) {
	nodes.forEach((node) => {
		node.url = [url, node.url].join("/");

		if (permissions && permissions !== true) {
			if (Array.isArray(node.permissions)) {
				node.permissions = [...node.permissions, ...permissions];
			} else {
				node.permissions = [...permissions];
			}
		}

		if (node.type == "View") {
			links.forEach((link) => {
				link.push(node.url);
			});
			node.data = React.createElement(node.data, { baseUrl: node.url });
			views.push(node);
		} else if (node.type == "Category") {
			node.links = [];

			recurseNodes(node.data, node.url, {
				views,
				links: [node.links, ...links],
				fallback,
				permissions: node.permissions
			});

			fallback[node.url] = (
				<Redirect to={node.defaultUrl ?? node.data[0].url} />
			);
		}
	});
	return { nodes, views, fallback };
}

function parseType(type) {
	return (name, data, cfg) => {
		if (type == "Category") {
			if (cfg != undefined) {
				// swap arguments if optional cfg is present
				let _cfg = cfg;
				cfg = data;
				data = _cfg;
			}
		}

		return {
			type,
			name,
			url: [cfg?.url ?? urlSafe(name)],
			permissions: cfg?.permissions ?? true,
			defaultUrl: cfg?.defaultUrl,
			icon: cfg?.icon,
			data
		};
	};
}

function urlSafe(str) {
	return str.toLowerCase().replace(/[\s/]+/g, "-");
}

module.exports = {
	createNavigation,
	useNavigation: require("./use-navigation.jsx")
};