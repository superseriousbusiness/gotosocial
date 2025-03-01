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
import TokensSearchForm from "./search";

export default function Tokens() {
	return (
		<div className="tokens-view">
			<div className="form-section-docs">
				<h1>App Tokens</h1>
				<p>
					On this page you can search through access tokens owned by applications that you have authorized to access your account and/or perform actions on your behalf.
					<br/>You can invalidate a token by clicking on the invalidate button under a token. This will remove the token from the database.
					<br/>The application that was authorized to access your account with that token will then no longer be authorized to do so, and you will need to log out and log in again with that application.
					<br/>In cases where you've logged into an application multiple times, or logged in with multiple devices or browsers, you may see multiple tokens for one application. This is normal!
					<br/>That said, feel free to invalidate old tokens that are never used, it's good security practice and it's fun to click the big red button.
				</p>
			</div>
			<TokensSearchForm />
		</div>
	);
}
