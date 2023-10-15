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

import DomainPermissionsOverview from "./overview";
import { PermType } from "../../lib/types/domain-permission";
import DomainPermDetail from "./detail";

export default function DomainPermissions({ baseUrl }: { baseUrl: string }) {
	return (
		<Switch>
			<Route path="/settings/admin/domain-permissions/:permType/:domain">
				{params => (
					<DomainPermDetail
						permType={params.permType as PermType}
						baseUrl={baseUrl}
						domain={params.domain}
					/>
				)}
			</Route>
			<Route path="/settings/admin/domain-permissions/:permType">
				{params => (
					<DomainPermissionsOverview
						permType={params.permType as PermType}
						baseUrl={baseUrl}
					/>
				)}
			</Route>
		</Switch>
	);
}
