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

const Promise = require("bluebird");
const React = require("react");
const ReactDom = require("react-dom");
const { Link, Route, Switch, useRoute, Redirect } = require("wouter");
const { ErrorBoundary } = require("react-error-boundary");

const Auth = require("./components/auth");
const ErrorFallback = require("./components/error");

const oauthLib = require("./lib/oauth");

require("./style.css");

const UserPanel = require("./user");
const AdminPanel = require("./admin");

const nav = {
	"User": {
		Component: require("./user"),
		entries: {
			"Profile": require("./user/profile.js"),
			"Settings": require("./user/settings.js"),
			"Customization": require("./user/customization.js")
		}
	},
	"Admin": {
		Component: require("./admin"),
		entries: {
			"Instance Settings": require("./admin/settings.js"),
			"Federation": require("./admin/federation.js"),
			"Customization": require("./admin/customization.js")
		}
	}
};

function urlSafe(str) {
	return str.toLowerCase().replace(/\s+/g, "-");
}

// TODO: nested categories?
const sidebar = [];
const panelRouter = [];

// Generate component tree from `nav` object once, as it won't change
Object.entries(nav).forEach(([name, {Component, entries}]) => {
	let base = `/settings/${urlSafe(name)}`;

	let links = [];
	let routes = [];

	let firstRoute;

	Object.entries(entries).forEach(([name, component]) => {
		let url = `${base}/${urlSafe(name)}`;

		if (firstRoute == undefined) {
			firstRoute = `${base}/${urlSafe(name)}`;
		}

		routes.push([url, component]);

		links.push(
			<NavButton key={url} href={url} name={name} />
		);
	});

	panelRouter.push(
		<Route key={base} path={base}>
			<Redirect to={firstRoute}/>
		</Route>
	);

	let childrenPath = `${base}/:section`;
	panelRouter.push(
		<Route key={childrenPath} path={childrenPath}>
			<ErrorBoundary FallbackComponent={ErrorFallback} onReset={() => {}}>
				{/* FIXME: implement onReset */}
				<Component routes={routes}/>
			</ErrorBoundary>
		</Route>
	);

	sidebar.push(
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

function NavButton({href, name}) {
	const [isActive] = useRoute(href);
	return (
		<Link href={href}>
			<a className={isActive ? "active" : ""} data-content={name}>
				{name}
			</a>
		</Link>
	);
}

function App() {
	const [oauth, setOauth] = React.useState();
	const [hasAuth, setAuth] = React.useState(false);
	const [oauthState, _setOauthState] = React.useState(localStorage.getItem("oauth"));

	React.useEffect(() => {
		let state = localStorage.getItem("oauth");
		if (state != undefined) {
			state = JSON.parse(state);
			let restoredOauth = oauthLib(state.config, state);
			Promise.try(() => {
				return restoredOauth.callback();
			}).then(() => {
				setAuth(true);
			});
			setOauth(restoredOauth);
		}
	}, [setAuth, setOauth]);

	if (!hasAuth && oauth && oauth.isAuthorized()) {
		setAuth(true);
	}

	if (oauth && oauth.isAuthorized()) {
		return (
			<>
				<div className="sidebar">
					{sidebar}
					<button className="logout" onClick={oauth.logout}>Log out</button>
				</div>
				<section>
					<Switch>
						{panelRouter}
					</Switch>
				</section>
			</>
		);
	} else if (oauthState != undefined) {
		return (
			<section>
				processing oauth...
			</section>
		);
	} else {
		return <Auth setOauth={setOauth}/>;
	}
}

ReactDom.render(<React.StrictMode><App/></React.StrictMode>, document.getElementById("root"));