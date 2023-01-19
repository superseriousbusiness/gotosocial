/*
	GoToSocial
	Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
const Redux = require("react-redux");

const query = require("../../lib/query");

const Login = require("./login");
const Loading = require("../loading");
const { Error } = require("../error");

module.exports = function Authorization({ App }) {
	const { loginState, expectingRedirect } = Redux.useSelector((state) => state.oauth);

	const { isLoading, isSuccess, data: account, error } = query.useVerifyCredentialsQuery(undefined, {
		skip: loginState == "none" || loginState == "logout" || expectingRedirect
	});

	console.log("skip verify:", loginState, expectingRedirect);

	let showLogin = true;
	let content = null;

	if (isLoading) {
		showLogin = false;

		let loadingInfo;
		if (loginState == "callback") {
			loadingInfo = "Processing OAUTH callback.";
		} else if (loginState == "login") {
			loadingInfo = "Verifying stored login.";
		}

		content = (
			<div>
				<Loading /> {loadingInfo}
			</div>
		);
	} else if (error != undefined) {
		content = (
			<div>
				<Error error={error} />
				You can attempt logging in again below:
			</div>
		);
	}

	if (loginState == "login" && isSuccess) {
		return <App account={account} />;
	} else {
		return (
			<section className="oauth">
				<h1>GoToSocial Settings</h1>
				{content}
				{showLogin && <Login />}
			</section>
		);
	}
};