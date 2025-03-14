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

import { useLogoutMutation, useVerifyCredentialsQuery } from "../../lib/query/login";
import { store } from "../../redux/store";
import React, { ReactNode } from "react";

import Login from "./login";
import Loading from "../loading";
import { Error } from "../error";
import { NoArg } from "../../lib/types/query";

export function Authorization({ App }) {
	const { current: loginState, expectingRedirect } = store.getState().login;
	const skip = (loginState == "none" || loginState == "loggedout" || expectingRedirect);
	const [ logoutQuery ] = useLogoutMutation();

	const {
		isLoading,
		isFetching,
		isSuccess,
		data: account,
		error,
	} = useVerifyCredentialsQuery(NoArg, { skip: skip });

	let showLogin = true;
	let content: ReactNode;

	if (isLoading || isFetching) {
		showLogin = false;

		let loadingInfo = "";
		if (loginState == "awaitingcallback") {
			loadingInfo = "Processing OAUTH callback.";
		} else if (loginState == "loggedin") {
			loadingInfo = "Verifying stored login.";
		}

		content = (
			<div>
				<Loading /> {loadingInfo}
			</div>
		);
	} else if (error !== undefined) {
		// Something went wrong,
		// log the user out.
		logoutQuery(NoArg);

		content = (
			<div>
				<Error error={error} />
				You can attempt logging in again below:
			</div>
		);
	}

	if (loginState == "loggedin" && isSuccess) {
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
}
