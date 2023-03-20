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
const { Link, Route, Redirect } = require("wouter");
const { ErrorBoundary } = require("react-error-boundary");

const { ErrorFallback } = require("../components/error");
const NavButton = require("../components/nav-button");

function urlSafe(str) {
	return str.toLowerCase().replace(/\s+/g, "-");
}

const RoleContext = React.createContext([]);

const rootPath = "/settings";

// function routingTree(tree, rootPath = ["/settings"]) {
// 	return syncpipe(tree, [
// 		(_) => Object.entries(_),
// 		(_) => _.map(([name, data]) => {
// 			if (name.startsWith("_")) {
// 				return [name, data];
// 			}

// 			let path = [...rootPath, data._url ?? urlSafe(name)];

// 			return [
// 				path.join("/"),
// 				(typeof data == "function")
// 					? {
// 						name,
// 						component: data
// 					}
// 					: routingTree(data, path)
// 			];
// 		}),
// 		(_) => Object.fromEntries(_)
// 	]);
// }

module.exports = function createUseNavigation(categories) {
	const sidebar = [];
	const viewRouter = [];

	Object.entries(categories).forEach(([categoryName, contents]) => {
		const {
			_permissions: permissions = true,
			_url: url = urlSafe(categoryName),
			...subCategories
		} = contents;

		let path = [rootPath, url];

		sidebarCategories.push(<SidebarCategory
			name={categoryName}
			key={path.join("/")}
			path={path}
			permissions={_permissions}
			entries={entries}
		/>);



	});

	return function useNavigation({ permissions }) {
		return {
			sidebar: (
				<RoleContext.provider value={permissions}>
					<nav>
						<ul>
							{sidebarCategories}
						</ul>
					</nav>
				</RoleContext.provider>
			),
			viewRouter
		};
	};
};

function useHasPermission(permissions) {
	const roles = React.useContext(RoleContext);

	if (permissions === true) {
		return true;
	}

	return roles.some((role) => permissions.includes(role));
}

function SidebarCategory({ name, path, permissions, entries }) {
	if (!useHasPermission(permissions)) {
		return null;
	}

	return (
		<li className="nav-category">
			<a className="nav-category-name">{name}</a>
			<ul className="entries">
				{entries.forEach(([entryName, entry]) => {
					let entryPath = path.push(entryName);
					return (
						<SidebarEntry
							key={entryPath.join("/")}
							path={entryPath}
							entry={entry}
						/>
					);
				})}
			</ul>
		</li>
	);
}

function SidebarEntry({ }) {

}

function generateNavigation(struct) {
	const sidebar = {
		all: [],
		admin: [],
	};

	const panelRouter = {
		all: [],
		admin: [],
	};

	Object.entries(struct).forEach(([categoryName, entries]) => {
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
				<Route path={`${url}/:page*`} key={url}>
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
			<div className="nav-category" key={name}>
				<Link href={firstRoute}>
					<a className="nav-category-name">
						{name}
					</a>
				</Link>
				<nav>
					{links}
				</nav>
			</div>
		);
	});

	return { sidebar, panelRouter };
};