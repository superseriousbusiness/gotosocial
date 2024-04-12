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

import React from "react";
import { Switch, Route } from "wouter";

import AccountDetail from "./detail";
import { AccountSearchForm } from "./search";

export default function Accounts({ baseUrl }) {
	return (
		<Switch>
			<Route path={`${baseUrl}/:accountId`}>
				<AccountDetail />
			</Route>
			<AccountOverview />
		</Switch>
	);
}

function AccountOverview({ }) {
	return (
		<div className="accounts-view">
			<h1>Accounts Overview</h1>
			<span>
				You can perform actions on an account by clicking
				its name in a report, or by searching for the account
				using the form below and clicking on its name.
			</span>
			<AccountSearchForm />
		</div>
	);
}
