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

const Auth = require("./components/auth");
const oauthLib = require("./lib/oauth");

require("./style.css");

const nav = {
	"User": [
		["Profile", require("./user/profile.js")],
		["Settings", require("./user/settings.js")],
		["Customization", require("./user/customization.js")]
	],
	"Admin": [
		["Instance Settings", require("./admin/settings.js")],
		["Federation", require("./admin/federation.js")],
		["Customization", require("./admin/customization.js")]
	]
};

function urlSafe(str) {
	return str.toLowerCase().replace(/\s+/g, "-");
}

// TODO: nested categories?
const sidebar = [];
const panelRouter = [];

Object.entries(nav).forEach(([category, entries]) => {
	let base = `/settings/${urlSafe(category)}`;

	// Category header goes to first page in category
	panelRouter.push(
		<Route key={base} path={base}>
			<Redirect to={`${base}/${urlSafe(entries[0][0])}`}/>
		</Route>
	);

	let links = entries.map(([name, component]) => {
		let url = `${base}/${urlSafe(name)}`;

		panelRouter.push(
			<Route key={url} path={url} component={component}/>
		);

		return <NavButton key={url} href={url} name={name} />;
	});

	sidebar.push(
		<React.Fragment key={category}>
			<Link href={`${base}/${urlSafe(entries[0][0])}`}>
				<a>
					<h2>{category}</h2>
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