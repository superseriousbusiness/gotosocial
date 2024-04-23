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
import React from "react";
import { BaseUrlContext, useBaseUrl } from "../../lib/navigation/util";
import { Redirect, Route, Router, Switch } from "wouter";
import AccountsOverview from "./accounts";
import AccountsPending from "./accounts/pending";
import AccountDetail from "./accounts/detail";
import { ReportOverview } from "./reports/overview";
import DomainPermissionsOverview from "./domain-permissions/overview";
import DomainPermDetail from "./domain-permissions/detail";
import ImportExport from "./domain-permissions/import-export";
import ReportDetail from "./reports/detail";

/*
	EXPORTED COMPONENTS
*/

/**
 * Moderation menu. Reports, accounts,
 * domain permissions import + export.
 */
export function ModerationMenu() {
	return (
		<MenuItem
			name="Moderation"
			itemUrl="moderation"
			defaultChild="reports"
			permissions={["moderator"]}
		>
			<ModerationReportsMenu />
			<ModerationAccountsMenu />
			<ModerationDomainPermsMenu />
		</MenuItem>
	);
}

/**
 * Moderation router. Reports, accounts,
 * domain permissions import + export.
 */
export function ModerationRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/moderation";
	const absBase = parentUrl + thisBase;
	
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

/*
	MENUS
*/

function ModerationReportsMenu() {
	return (
		<MenuItem
			name="Reports"
			itemUrl="reports"
			icon="fa-flag"
		/>
	);
}

function ModerationAccountsMenu() {
	return (
		<MenuItem
			name="Accounts"
			itemUrl="accounts"
			defaultChild="overview"
			icon="fa-users"
		>
			<MenuItem
				name="Overview"
				itemUrl="overview"
				icon="fa-list"
			/>
			<MenuItem
				name="Pending"
				itemUrl="pending"
				icon="fa-question"
			/>
		</MenuItem>
	);
}

function ModerationDomainPermsMenu() {
	return (
		<MenuItem
			name="Domain Permissions"
			itemUrl="domain-permissions"
			defaultChild="blocks"
			icon="fa-hubzilla"
		>
			<MenuItem
				name="Blocks"
				itemUrl="blocks"
				icon="fa-close"
			/>
			<MenuItem
				name="Allows"
				itemUrl="allows"
				icon="fa-check"
			/>
			<MenuItem
				name="Import/Export"
				itemUrl="import-export"
				icon="fa-floppy-o"
			/>
		</MenuItem>
	);
}

/*
	ROUTERS
*/

function ModerationReportsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/reports";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Switch>
					<Route path={"/:reportId"} component={ReportDetail} />
					<Route component={ReportOverview}/>
				</Switch>
			</Router>
		</BaseUrlContext.Provider>
	);
}

function ModerationAccountsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/accounts";
	const absBase = parentUrl + thisBase;
	
	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Switch>
					<Route path="/overview" component={AccountsOverview}/>
					<Route path="/pending" component={AccountsPending}/>
					<Route path="/:accountID" component={AccountDetail}/>
					<Route><Redirect to="/overview"/></Route>
				</Switch>
			</Router>
		</BaseUrlContext.Provider>
	);
}

function ModerationDomainPermsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/domain-permissions";
	const absBase = parentUrl + thisBase;
	
	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Switch>
					<Route path="/import-export" component={ImportExport} />
					<Route path="/process" component={ImportExport} />
					<Route path="/:permType/:domain" component={DomainPermDetail} />
					<Route path="/:permType" component={DomainPermissionsOverview} />
					<Route><Redirect to="/blocks"/></Route>
				</Switch>
			</Router>
		</BaseUrlContext.Provider>
	);
}
