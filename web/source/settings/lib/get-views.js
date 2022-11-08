/*
	GoToSocial
	Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
const { Link, Route, Redirect } = require("wouter");
const { ErrorBoundary } = require("react-error-boundary");

const ErrorFallback = require("../components/error");
const NavButton = require("../components/nav-button");

function urlSafe(str) {
	return str.toLowerCase().replace(/\s+/g, "-");
}

module.exports = function getViews(struct) {
	const sidebar = {
		all: [],
		admin: [],
	};

	const panelRouter = {
		all: [],
		admin: [],
	};

	Object.entries(struct).forEach(([name, entries]) => {
		let sidebarEl = sidebar.all;
		let panelRouterEl = panelRouter.all;

		if (entries.adminOnly) {
			sidebarEl = sidebar.admin;
			panelRouterEl = panelRouter.admin;
			delete entries.adminOnly;
		}

		let base = `/settings/${urlSafe(name)}`;

		let links = [];

		let firstRoute;

		Object.entries(entries).forEach(([name, ViewComponent]) => {
			let url = `${base}/${urlSafe(name)}`;

			if (firstRoute == undefined) {
				firstRoute = url;
			}

			panelRouterEl.push((
				<Route path={`${url}/:page?`} key={url}>
					<ErrorBoundary FallbackComponent={ErrorFallback} onReset={() => { }}>
						{/* FIXME: implement onReset */}
						<ViewComponent />
					</ErrorBoundary>
				</Route>
			));

			links.push(
				<NavButton key={url} href={url} name={name} />
			);
		});

		panelRouterEl.push(
			<Route key={base} path={base}>
				<Redirect to={firstRoute} />
			</Route>
		);

		sidebarEl.push(
			<React.Fragment key={name}>
				<Link href={firstRoute}>
					<a>
						<h2>{name}</h2>
					</a>
				</Link>
				<nav>
					{links}
				</nav>
			</React.Fragment>
		);
	});

	return { sidebar, panelRouter };
};