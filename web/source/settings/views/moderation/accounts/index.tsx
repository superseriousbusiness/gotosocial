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
import { AccountSearchForm } from "./search";

export default function AccountsSearch({ }) {
	return (
		<div className="accounts-view">
			<div className="form-section-docs">
				<h1>Accounts Search</h1>
				<p>
					You can perform actions on an account by clicking
					its name in a report, or by searching for the account
					using the form below and clicking on its name.
					<br/>
					All fields in the below form are optional.
				</p>
			</div>
			<AccountSearchForm />
		</div>
	);
}
