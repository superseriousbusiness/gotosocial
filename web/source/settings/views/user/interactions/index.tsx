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

import { Redirect, Route, Router, Switch } from "wouter";
import { BaseUrlContext, useBaseUrl } from "../../../lib/navigation/util";
import InteractionRequestDetail from "./detail";
import InteractionRequestsSearchForm from "./search";

function InteractionRequests() {
	return (
		<div className="interaction-requests-view">
			<div className="form-section-docs">
				<h1>Interaction Requests</h1>
				<p>
					On this page you can search through interaction requests
					targeting your statuses, and approve or reject them.
				</p>
			</div>
			<InteractionRequestsSearchForm />
		</div>
	);
}

/**
 * - /settings/users/interaction-requests/search
 * - /settings/users/interaction-requests/{reqId}
 */
export default function InteractionRequestsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/interaction-requests";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Switch>
					<Route path="/search" component={InteractionRequests} />
					<Route path="/:reqId" component={InteractionRequestDetail} />
					<Redirect to="/search" />
				</Switch>
			</Router>
		</BaseUrlContext.Provider>
	);
}
