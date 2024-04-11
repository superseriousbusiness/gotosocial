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
import { useSearchAccountsQuery } from "../../../lib/query";
import { AccountList } from "../../../components/account-list";

export default function AccountsPending({ baseUrl }) {
	const searchRes = useSearchAccountsQuery({status: "pending"});

	return (
		<div className="accounts-view">
			<h1>Pending Accounts</h1>
			<AccountList
				isLoading={searchRes.isLoading}
				isSuccess={searchRes.isSuccess}
				data={searchRes.data}
				isError={searchRes.isError}
				error={searchRes.error}
				emptyMessage="No pending account sign-ups."
			/>
		</div>
	);
};
