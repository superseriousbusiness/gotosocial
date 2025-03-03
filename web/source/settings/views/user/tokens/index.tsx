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
				<h1>Access Tokens</h1>
				<p>
					On this page you can search through access tokens owned by applications that you have authorized to
					access your account and/or perform actions on your behalf. You can invalidate a token by clicking on
					the invalidate button under a token. This will remove the token from the database.
					<br/><br/>
					<strong>
						If you see any tokens from applications that you do not recognize, or do not remember authorizing to access
						your account, then you should invalidate them, and consider changing your password as soon as possible.
					</strong>
				</p>
				<a
					href="https://docs.gotosocial.org/en/latest/user_guide/settings/#access-tokens"
					target="_blank"
					className="docslink"
					rel="noreferrer"
				>
					Learn more about managing your access tokens (opens in a new tab)
				</a>
			</div>
			<TokensSearchForm />
		</div>
	);
}
