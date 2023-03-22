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
const { Link, Route, Redirect, Switch, useLocation } = require("wouter");
const { ErrorBoundary } = require("react-error-boundary");
const syncpipe = require("syncpipe");

const { ErrorFallback } = require("../../components/error");

const {
	RoleContext,
	useHasPermission,
	checkPermission,
	BaseUrlContext
} = require("./util");

function Sidebar(menuTree) {
	return function SidebarComponent() {
		return (
			<nav>
				<ul>
					{menuTree}
				</ul>
			</nav>
		);
	};
}

function ViewRouter(routing, defaultRoute) {
	return function ViewRouterComponent() {
		const permissions = React.useContext(RoleContext);

		const filteredRoutes = React.useMemo(() => {
			return syncpipe(routing, [
				(_) => Object.values(_),
				(_) => _.map((v) => Object.entries(v)),
				(_) => _.flat(1),
				(_) => _.filter(([_url, v]) => checkPermission(v.permissions, permissions)),
				(_) => _.map(([url, item]) => (
					<Route path={`${url}/:page*`} key={url}>
						{/* <Route path={url} key={url}> */}
						<ErrorBoundary FallbackComponent={ErrorFallback} onReset={() => { }}>
							{/* FIXME: implement onReset */}
							<BaseUrlContext.Provider value={url}>
								{item.view}
							</BaseUrlContext.Provider>
						</ErrorBoundary>
					</Route>
				))
			]);
		}, [permissions]);

		return (
			<Switch>
				{filteredRoutes}
				<Redirect to={defaultRoute} />
			</Switch>
		);
	};
}

function MenuComponent({ name, url, icon, permissions, links, level = 0, children }) {
	let [location] = useLocation();
	// FIXME: doesn't match quite as well for wildcard routes
	let isActive = url == location || links?.includes(location) || location.startsWith(url);

	if (!useHasPermission(permissions)) {
		return null;
	}

	const className = [
		children?.length > 0
			? "category"
			: "",
		level > 0
			? `nested-${level}`
			: "top-level",
		isActive
			? "active"
			: ""
	].join(" ");

	return (
		<li className={className}>
			<Link href={url}>
				<a tabIndex={(level == 0 || isActive) ? "-1" : null}>
					{icon && <i className={`fa fa-fw ${icon}`} aria-hidden="true" />}
					{name}
				</a>
			</Link>
			{((level == 0 || isActive) && children?.length > 0) &&
				<ul>
					{children}
				</ul>
			}
		</li>
	);
}

module.exports = {
	Sidebar,
	ViewRouter,
	MenuComponent
};