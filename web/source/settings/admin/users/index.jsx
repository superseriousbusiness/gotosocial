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
const { Switch, Route, Link } = require("wouter");

const query = require("../../lib/query");
const { useTextInput } = require("../../lib/form");

const UserDetail = require("./detail");
const { useBaseUrl } = require("../../lib/navigation/util");
const { Error } = require("../../components/error");

module.exports = function Users({ baseUrl }) {
	return (
		<div className="users">
			<Switch>
				<Route path={`${baseUrl}/:userId`}>
					<UserDetail />
				</Route>
				<UserOverview />
			</Switch>
		</div>
	);
};

function UserOverview({ }) {
	return (
		<>
			<h1>Users</h1>
			<div>
				Pending <a href="https://github.com/superseriousbusiness/gotosocial/issues/581">#581</a>,
				there is currently no way to list user accounts.<br />
				You can perform actions on reported users by clicking their name in the report, or searching for a username below.
			</div>

			<UserSearchForm />
		</>
	);
}

function UserSearchForm() {
	const [searchUser, result] = query.useSearchUserMutation();

	const [onAccountChange, _resetAccount, { account }] = useTextInput("account");

	function submitSearch(e) {
		e.preventDefault();
		if (account.trim().length != 0) {
			searchUser(account);
		}
	}

	return (
		<div className="account-search">
			<form onSubmit={submitSearch}>
				<div className="form-field text">
					<label htmlFor="url">
						User:
					</label>
					<div className="row">
						<input
							type="text"
							id="account"
							name="account"
							onChange={onAccountChange}
							value={account}
						/>
						<button disabled={result.isLoading}>
							<i className={[
								"fa fa-fw",
								(result.isLoading
									? "fa-refresh fa-spin"
									: "fa-search")
							].join(" ")} aria-hidden="true" title="Search" />
							<span className="sr-only">Search</span>
						</button>
					</div>
				</div>
			</form>
			<AccountList
				isSuccess={result.isSuccess}
				data={result.data}
				isError={result.isError}
				error={result.error}
			/>
		</div>
	);
}

function AccountList({ isSuccess, data, isError, error }) {
	const baseUrl = useBaseUrl();

	if (!(isSuccess || isError)) {
		return null;
	}

	if (error) {
		return <Error error={error} />;
	}

	if (data.length == 0) {
		return <b>No accounts found that match your query</b>;
	}

	return (
		<>
			<h2>Results:</h2>
			<div className="list">
				{data.map((acc) => (
					<Link key={acc.acct} className="account entry" to={`${baseUrl}/${acc.id}`}>
						{acc.display_name?.length > 0
							? acc.display_name
							: acc.username
						}
						<span id="username">(@{acc.acct})</span>
					</Link>
				))}
			</div>
		</>
	);
}