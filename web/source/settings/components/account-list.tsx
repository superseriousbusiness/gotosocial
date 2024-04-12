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
import { Link } from "wouter";
import { Error } from "./error";
import { AdminAccount } from "../lib/types/account";
import { SerializedError } from "@reduxjs/toolkit";
import { FetchBaseQueryError } from "@reduxjs/toolkit/query";

export interface AccountListProps {
	isSuccess: boolean,
	data: AdminAccount[] | undefined,
	isLoading: boolean,
	isError: boolean,
	error: FetchBaseQueryError | SerializedError | undefined,
	emptyMessage: string,
}

export function AccountList({
	isLoading,
	isSuccess,
	data,
	isError,
	error,
	emptyMessage,
}: AccountListProps) {
	if (!(isSuccess || isError)) {
		// Hasn't been called yet.
		return null;
	}

	if (isLoading) {
		return <i
			className="fa fa-fw fa-refresh fa-spin"
			aria-hidden="true"
			title="Loading..."
		/>;
	}

	if (error) {
		return <Error error={error} />;
	}

	if (data == undefined || data.length == 0) {
		return <b>{emptyMessage}</b>;
	}

	return (
		<div className="list">
			{data.map(({ account: acc }) => (		
				<Link
					key={acc.acct}
					className="account entry"
					href={`/settings/admin/accounts/${acc.id}`}
				>
					{acc.display_name?.length > 0
						? acc.display_name
						: acc.username
					}
					<span id="username">(@{acc.acct})</span>
				</Link>
			))}
		</div>
	);
}