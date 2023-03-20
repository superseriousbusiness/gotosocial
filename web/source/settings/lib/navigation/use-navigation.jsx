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

const { Link, Route, useRoute } = require("wouter");
const { ErrorBoundary } = require("react-error-boundary");

const { ErrorFallback } = require("../../components/error");

const RoleContext = React.createContext([]);

module.exports = function useNavigation(nav, { permissions }) {
	return {
		sidebar: <Sidebar nav={nav} permissions={permissions} />,
		routedViews: nav.views
			.filter((v) => {
				console.log(v.name, v.permissions, permissions);
				return checkPermission(v.permissions, permissions);
			})
			.map((view) => (
				<Route path={`${view.url}/:page*`} key={view.url}>
					<ErrorBoundary FallbackComponent={ErrorFallback} onReset={() => { }}>
						{/* FIXME: implement onReset */}
						{view.data}
					</ErrorBoundary>
				</Route>
			)),
		fallbackRoutes: Object.entries(nav.fallback).map(([key, val]) => (
			<Route path={key} key={key}>{val}</Route>
		))
	};
};

function useActive(href) {
	return useRoute(`${href}/:anything?`)[0];
}

const Types = {
	Category({ node, level = 0 }) {
		let active = useActive(node.url);

		if (!useHasPermission(node.permissions)) {
			return null;
		}

		return (
			<li className={["category", level > 0 ? `nested-${level}` : "top-level", active ? "active" : ""].join(" ")}>
				<Link href={node.data[0].url}>
					<a tabIndex={(level == 0 || active) ? "-1" : null}>
						{node.icon && <i className={`fa fa-fw ${node.icon}`} aria-hidden="true" />}
						{node.name}
					</a>
				</Link>
				{(level == 0 || active) &&
					<ul>
						{node.data.map((node) => {
							return React.createElement(
								Types[node.type],
								{
									key: `${node.name}-${node.url}`,
									node,
									level: level + 1
								});
						})}
					</ul>
				}
			</li>
		);
	},
	View({ node }) {
		let active = useActive(node.url);
		return (
			<li className={active ? "active" : ""}>
				<Link href={node.url}>
					<a>
						{node.icon && <i className={`fa fa-fw ${node.icon}`} aria-hidden="true" />}
						{node.name}
					</a>
				</Link>
			</li>
		);
	}
};

function Sidebar({ nav, permissions }) {
	return (
		<RoleContext.Provider value={permissions}>
			<nav>
				<ul>
					{nav.nodes.map((node) => {
						return React.createElement(Types[node.type], { node, key: `${node.name}-${node.url}` });
					})}
				</ul>
			</nav>
		</RoleContext.Provider>
	);
}

function useHasPermission(permissions) {
	const roles = React.useContext(RoleContext);
	return checkPermission(permissions, roles);
}

function checkPermission(required, user) {
	if (required === true) {
		return true;
	}

	return user.some((role) => required.includes(role));
}