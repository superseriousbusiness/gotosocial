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
import { BaseUrlContext, useBaseUrl, useHasPermission } from "../../lib/navigation/util";
import { Redirect, Route, Router, Switch } from "wouter";
import ReportsSearch from "./reports/search";
import ReportDetail from "./reports/detail";
import { ErrorBoundary } from "../../lib/navigation/error";
import ImportExport from "./domain-permissions/import-export";
import DomainPermissionsOverview from "./domain-permissions/overview";
import DomainPermView from "./domain-permissions/detail";
import AccountsSearch from "./accounts";
import AccountsPending from "./accounts/pending";
import AccountDetail from "./accounts/detail";
import DomainPermissionDraftsSearch from "./domain-permissions/drafts";
import DomainPermissionDraftNew from "./domain-permissions/drafts/new";
import DomainPermissionDraftDetail from "./domain-permissions/drafts/detail";
import DomainPermissionExcludeDetail from "./domain-permissions/excludes/detail";
import DomainPermissionExcludesSearch from "./domain-permissions/excludes";
import DomainPermissionExcludeNew from "./domain-permissions/excludes/new";
import DomainPermissionSubscriptionsSearch from "./domain-permissions/subscriptions";
import DomainPermissionSubscriptionNew from "./domain-permissions/subscriptions/new";
import DomainPermissionSubscriptionDetail from "./domain-permissions/subscriptions/detail";
import DomainPermissionSubscriptionsPreview from "./domain-permissions/subscriptions/preview";

/*
	EXPORTED COMPONENTS
*/

/**
 * - /settings/moderation/reports/overview
 * - /settings/moderation/reports/:reportId
 * - /settings/moderation/accounts/search
 * - /settings/moderation/accounts/pending
 * - /settings/moderation/accounts/:accountID
 * - /settings/moderation/domain-permissions/:permType
 * - /settings/moderation/domain-permissions/:permType/:domain
 * - /settings/moderation/domain-permissions/import-export
 * - /settings/moderation/domain-permissions/process
 */
export default function ModerationRouter() {	
	const parentUrl = useBaseUrl();
	const thisBase = "/moderation";
	const absBase = parentUrl + thisBase;

	const permissions = ["moderator"];
	const moderator = useHasPermission(permissions);
	if (!moderator) {
		return null;
	}

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ModerationReportsRouter />
				<ModerationAccountsRouter />
				<ModerationDomainPermsRouter />
			</Router>
		</BaseUrlContext.Provider>
	);
}

/*
	INTERNAL COMPONENTS
*/

/**
 * - /settings/moderation/reports/overview
 * - /settings/moderation/reports/:reportId
 */
function ModerationReportsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/reports";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Switch>
						<Route path="/search" component={ReportsSearch}/>
						<Route path={"/:reportId"} component={ReportDetail} />
						<Route><Redirect to="/search"/></Route>
					</Switch>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}

/**
 * - /settings/moderation/accounts/search
 * - /settings/moderation/accounts/pending
 * - /settings/moderation/accounts/:accountID
 */
function ModerationAccountsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/accounts";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Switch>
						<Route path="/search" component={AccountsSearch}/>
						<Route path="/pending" component={AccountsPending}/>
						<Route path="/:accountID" component={AccountDetail}/>
						<Route><Redirect to="/search"/></Route>
					</Switch>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}

/**
 * - /settings/moderation/domain-permissions/:permType
 * - /settings/moderation/domain-permissions/:permType/:domain
 * - /settings/moderation/domain-permissions/import-export
 * - /settings/moderation/domain-permissions/process
 */
function ModerationDomainPermsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/domain-permissions";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Switch>
						<Route path="/import-export" component={ImportExport} />
						<Route path="/process" component={ImportExport} />
						<Route path="/drafts/search" component={DomainPermissionDraftsSearch} />
						<Route path="/drafts/new" component={DomainPermissionDraftNew} />
						<Route path="/drafts/:permDraftId" component={DomainPermissionDraftDetail} />
						<Route path="/excludes/search" component={DomainPermissionExcludesSearch} />
						<Route path="/excludes/new" component={DomainPermissionExcludeNew} />
						<Route path="/excludes/:excludeId" component={DomainPermissionExcludeDetail} />
						<Route path="/subscriptions/search" component={DomainPermissionSubscriptionsSearch} />
						<Route path="/subscriptions/new" component={DomainPermissionSubscriptionNew} />
						<Route path="/subscriptions/preview" component={DomainPermissionSubscriptionsPreview} />
						<Route path="/subscriptions/:permSubId" component={DomainPermissionSubscriptionDetail} />
						<Route path="/:permType" component={DomainPermissionsOverview} />
						<Route path="/:permType/:domain" component={DomainPermView} />
						<Route><Redirect to="/blocks"/></Route>
					</Switch>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}
