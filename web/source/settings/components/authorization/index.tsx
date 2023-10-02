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

import { useVerifyCredentialsQuery } from "../../lib/query/oauth";
import { store } from "../../redux/store";

import React from "react";

import Login from "./login";
import Loading from "../loading";
import { Error } from "../error";

export function Authorization({ App }) {
	const { loginState, expectingRedirect } = store.getState().oauth;
	const skip = (loginState == "none" || loginState == "logout" || expectingRedirect);

	const {
		isLoading,
		isSuccess,
		data: account,
		error,
	} = useVerifyCredentialsQuery(null, { skip: skip });

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
