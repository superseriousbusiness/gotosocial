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
import { BaseUrlContext, useBaseUrl } from "../../lib/navigation/util";
import { Redirect, Route, Router, Switch } from "wouter";
import { ErrorBoundary } from "../../lib/navigation/error";
import UserProfile from "./profile";
import UserMigration from "./migration";
import PostSettings from "./posts";
import EmailPassword from "./emailpassword";
import ExportImport from "./export-import";
import InteractionRequests from "./interactions";
import InteractionRequestDetail from "./interactions/detail";

/**
 * - /settings/user/profile
 * - /settings/user/posts
 * - /settings/user/emailpassword
 * - /settings/user/migration
 * - /settings/user/export-import
 * - /settings/users/interaction_requests
 */
export default function UserRouter() {
	const baseUrl = useBaseUrl();
	const thisBase = "/user";
	const absBase = baseUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Switch>
						<Route path="/profile" component={UserProfile} />
						<Route path="/posts" component={PostSettings} />
						<Route path="/emailpassword" component={EmailPassword} />
						<Route path="/migration" component={UserMigration} />
						<Route path="/export-import" component={ExportImport} />
						<InteractionRequestsRouter />
						<Route><Redirect to="/profile" /></Route>
					</Switch>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}

/**
 * - /settings/users/interaction_requests/search
 * - /settings/users/interaction_requests/{reqId}
 */
function InteractionRequestsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/interaction_requests";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Switch>
					<Route path="/search" component={InteractionRequests} />
					<Route path="/:reqId" component={InteractionRequestDetail} />
					<Route><Redirect to="/search"/></Route>
				</Switch>
			</Router>
		</BaseUrlContext.Provider>
	);
}
