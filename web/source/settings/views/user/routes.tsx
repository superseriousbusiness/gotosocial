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

import { MenuItem } from "../../lib/navigation/menu";
import React, { lazy, Suspense } from "react";
import { BaseUrlContext, useBaseUrl } from "../../lib/navigation/util";
import { Redirect, Route, Router, Switch } from "wouter";
import { ErrorBoundary } from "../../lib/navigation/error";
import Loading from "../../components/loading";

/**
 * - /settings/user/profile
 * - /settings/user/settings
 * - /settings/user/migration
 */
export function UserMenu() {	
	return (
		<MenuItem
			name="User"
			itemUrl="user"
			defaultChild="profile"
		>
			{/* Profile */}
			<MenuItem
				name="Profile"
				itemUrl="profile"
				icon="fa-user"
			/>
			{/* Settings */}
			<MenuItem
				name="Settings"
				itemUrl="settings"
				icon="fa-cogs"
			/>
			{/* Migration */}
			<MenuItem
				name="Migration"
				itemUrl="migration"
				icon="fa-exchange"
			/>
		</MenuItem>
	);
}

/**
 * - /settings/user/profile
 * - /settings/user/settings
 * - /settings/user/migration
 */
export function UserRouter() {
	const baseUrl = useBaseUrl();
	const thisBase = "/user";
	const absBase = baseUrl + thisBase;

	const UserProfile = lazy(() => import('./profile'));
	const UserSettings = lazy(() => import('./settings'));
	const UserMigration = lazy(() => import('./migration'));

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Suspense fallback={<Loading/>}>
						<Switch>
							<Route path="/profile" component={UserProfile} />
							<Route path="/settings" component={UserSettings} />
							<Route path="/migration" component={UserMigration} />
							<Route><Redirect to="/profile" /></Route>
						</Switch>
					</Suspense>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}
