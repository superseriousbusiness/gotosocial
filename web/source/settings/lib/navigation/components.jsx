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
const { Link, Route, Redirect, Switch, useLocation, useRouter } = require("wouter");
const syncpipe = require("syncpipe");

const {
	RoleContext,
	useHasPermission,
	checkPermission,
	BaseUrlContext
} = require("./util");

const ActiveRouteCtx = React.createContext();
function useActiveRoute() {
	return React.useContext(ActiveRouteCtx);
}

function Sidebar(menuTree, routing) {
	const components = menuTree.map((m) => m.MenuEntry);

	return function SidebarComponent() {
		const router = useRouter();
		const [location] = useLocation();

		let activeRoute = routing.find((l) => {
			let [match] = router.matcher(l.routingUrl, location);
			return match;
		})?.routingUrl;

		return (
			<nav className="menu-tree">
				<ul className="top-level">
					<ActiveRouteCtx.Provider value={activeRoute}>
						{components}
					</ActiveRouteCtx.Provider>
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
				(_) => _.filter((route) => checkPermission(route.permissions, permissions)),
				(_) => _.map((route) => {
					return (
						<Route path={route.routingUrl} key={route.key}>
							<ErrorBoundary>
								{/* FIXME: implement reset */}
								<BaseUrlContext.Provider value={route.url}>
									{route.view}
								</BaseUrlContext.Provider>
							</ErrorBoundary>
						</Route>
					);
				})
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

function MenuComponent({ type, name, url, icon, permissions, links, level, children }) {
	const activeRoute = useActiveRoute();

	if (!useHasPermission(permissions)) {
		return null;
	}

	const classes = [type];

	if (level == 0) {
		classes.push("top-level");
	} else if (level == 1) {
		classes.push("expanding");
	} else {
		classes.push("nested");
	}

	const isActive = links.includes(activeRoute);
	if (isActive) {
		classes.push("active");
	}

	const className = classes.join(" ");

	return (
		<li className={className}>
			<Link href={url}>
				<a tabIndex={level == 0 ? "-1" : null} className="title">
					{icon && <i className={`icon fa fa-fw ${icon}`} aria-hidden="true" />}
					{name}
				</a>
			</Link>
			{(type == "category" && (level == 0 || isActive) && children?.length > 0) &&
				<ul>
					{children}
				</ul>
			}
		</li>
	);
}

class ErrorBoundary extends React.Component {

	constructor() {
		super();
		this.state = {};

		this.resetErrorBoundary = () => {
			this.setState({});
		};
	}

	static getDerivedStateFromError(error) {
		return { hadError: true, error };
	}

	componentDidCatch(_e, info) {
		this.setState({
			...this.state,
			componentStack: info.componentStack
		});
	}

	render() {
		if (this.state.hadError) {
			return (
				<ErrorFallback
					error={this.state.error}
					componentStack={this.state.componentStack}
					resetErrorBoundary={this.resetErrorBoundary}
				/>
			);
		} else {
			return this.props.children;
		}
	}
}

function ErrorFallback({ error, componentStack, resetErrorBoundary }) {
	return (
		<div className="error">
			<p>
				{"An error occured, please report this on the "}
				<a href="https://github.com/superseriousbusiness/gotosocial/issues">GoToSocial issue tracker</a>
				{" or "}
				<a href="https://matrix.to/#/#gotosocial-help:superseriousbusiness.org">Matrix support room</a>.
				<br />Include the details below:
			</p>
			<div className="details">
				<pre>
					{error.name}: {error.message}

					{componentStack && [
						"\n\nComponent trace:",
						componentStack
					]}
					{["\n\nError trace: ", error.stack]}
				</pre>
			</div>
			<p>
				<button onClick={resetErrorBoundary}>Try again</button> or <a href="">refresh the page</a>
			</p>
		</div>
	);
}

module.exports = {
	Sidebar,
	ViewRouter,
	MenuComponent
};