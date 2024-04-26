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
import { ErrorBoundary } from "../../lib/navigation/error";
import InstanceSettings from "./instance/settings";
import InstanceRules from "./instance/rules";
import InstanceRuleDetail from "./instance/ruledetail";
import Media from "./actions/media";
import Keys from "./actions/keys";
import EmojiOverview from "./emoji/local/overview";
import EmojiDetail from "./emoji/local/detail";
import RemoteEmoji from "./emoji/remote";

/*
	EXPORTED COMPONENTS
*/

/**
 * - /settings/instance/settings
 * - /settings/instance/rules
 * - /settings/instance/rules/:ruleId
 * - /settings/admin/emojis
 * - /settings/admin/emojis/local
 * - /settings/admin/emojis/local/:emojiId
 * - /settings/admin/emojis/remote
 * - /settings/admin/actions
 * - /settings/admin/actions/media
 * - /settings/admin/actions/keys
 */
export default function AdminRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/admin";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<AdminInstanceRouter />
				<AdminEmojisRouter />
				<AdminActionsRouter />
			</Router>
		</BaseUrlContext.Provider>
	);
}

/*
	INTERNAL COMPONENTS
*/

/**
 * - /settings/admin/emojis
 * - /settings/admin/emojis/local
 * - /settings/admin/emojis/local/:emojiId
 * - /settings/admin/emojis/remote
 */
function AdminEmojisRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/emojis";
	const absBase = parentUrl + thisBase;

	const permissions = ["admin"];
	const admin = useHasPermission(permissions);
	if (!admin) {
		return null;
	}

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Switch>
						<Route path="/local" component={EmojiOverview} />
						<Route path="/local/:emojiId" component={EmojiDetail} />
						<Route path="/remote" component={RemoteEmoji} />
						<Route><Redirect to="/local" /></Route>
					</Switch>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}

/**
 * - /settings/admin/actions
 * - /settings/admin/actions/media
 * - /settings/admin/actions/keys
 */
function AdminActionsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/actions";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Switch>
						<Route path="/media" component={Media} />
						<Route path="/keys" component={Keys} />
						<Route><Redirect to="/media" /></Route>
					</Switch>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}

/**
 * - /settings/instance/settings
 * - /settings/instance/rules
 * - /settings/instance/rules/:ruleId
 */
function AdminInstanceRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/instance";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Switch>
						<Route path="/settings" component={InstanceSettings}/>
						<Route path="/rules" component={InstanceRules} />
						<Route path="/rules/:ruleId" component={InstanceRuleDetail} />
						<Route><Redirect to="/settings" /></Route>
					</Switch>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}
