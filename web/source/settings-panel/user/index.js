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
const { Route, Switch } = require("wouter");

module.exports = function UserPanel({oauth, routes}) {
	// const [account, setAccount] = React.useState({});
	// const [errorMsg, setError] = React.useState("");
	// const [statusMsg, setStatus] = React.useState("Fetching user info");

	// React.useEffect(() => {
	// 	Promise.try(() => {
	// 		return oauth.apiRequest("/api/v1/accounts/verify_credentials", "GET");
	// 	}).then((json) => {
	// 		setAccount(json);
	// 	}).catch((e) => {
	// 		setError(e.message);
	// 		setStatus("");
	// 	});
	// }, [oauth, setAccount, setError, setStatus]);

	// throw new Error("test");

	return (
		<Switch>
			{routes.map(([path, component]) => {
				console.log(component);
				return <Route key={path} path={path} component={component}/>;
			})}
		</Switch>
	);
};