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
import React, { Suspense, lazy } from "react";
import { BaseUrlContext, useBaseUrl, useHasPermission } from "../../lib/navigation/util";
import { Redirect, Route, Router, Switch } from "wouter";
import Loading from "../../components/loading";
import { ErrorBoundary } from "../../lib/navigation/error";

/*
	EXPORTED COMPONENTS
*/

/**
 * - /settings/admin/instance-settings
 * - /settings/admin/instance-rules
 * - /settings/admin/instance-rules/:ruleId
 * - /settings/admin/emojis
 * - /settings/admin/emojis/local
 * - /settings/admin/emojis/local/:emojiId
 * - /settings/admin/emojis/remote
 * - /settings/admin/actions
 * - /settings/admin/actions/media
 * - /settings/admin/actions/keys
 */
export function AdminMenu() {
	// Don't route if logged-in user
	// doesn't have permissions to access.
	if (!useHasPermission(["admin"])) {
		return null;
	}
	
	return (
		<MenuItem
			name="Administration"
			itemUrl="admin"
			defaultChild="actions"
			permissions={["admin"]}
		>
			<MenuItem
				name="Instance Settings"
				itemUrl="instance-settings"
				icon="fa-sliders"
			/>
			<MenuItem
				name="Instance Rules"
				itemUrl="instance-rules"
				icon="fa-dot-circle-o"
			/>
			<AdminEmojisMenu />
			<AdminActionsMenu />
		</MenuItem>
	);
}

/**
 * - /settings/admin/instance-settings
 * - /settings/admin/instance-rules
 * - /settings/admin/instance-rules/:ruleId
 * - /settings/admin/emojis
 * - /settings/admin/emojis/local
 * - /settings/admin/emojis/local/:emojiId
 * - /settings/admin/emojis/remote
 * - /settings/admin/actions
 * - /settings/admin/actions/media
 * - /settings/admin/actions/keys
 */
export function AdminRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/admin";
	const absBase = parentUrl + thisBase;
	
	const InstanceSettings = lazy(() => import('./settings/settings'));
	const InstanceRules = lazy(() => import("./settings/rules"));
	const InstanceRuleDetail = lazy(() => import('./settings/ruledetail'));

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Suspense fallback={<Loading/>}>
						<Switch>
							<Route path="/instance-settings" component={InstanceSettings}/>
							<Route path="/instance-rules" component={InstanceRules} />
							<Route path="/instance-rules/:ruleId" component={InstanceRuleDetail} />
							<Route><Redirect to="/instance-settings" /></Route>
						</Switch>
					</Suspense>
				</ErrorBoundary>
				<AdminEmojisRouter />
				<AdminActionsRouter />
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

function AdminActionsMenu() {
	return (
		<MenuItem
			name="Actions"
			itemUrl="actions"
			defaultChild="media"
			icon="fa-bolt"
		>
			<MenuItem
				name="Media"
				itemUrl="media"
				icon="fa-photo"
			/>
			<MenuItem
				name="Keys"
				itemUrl="keys"
				icon="fa-key-modern"
			/>
		</MenuItem>
	);
}

function AdminEmojisMenu() {
	return (
		<MenuItem
			name="Custom Emoji"
			itemUrl="emojis"
			defaultChild="local"
			icon="fa-smile-o"
		>
			<MenuItem
				name="Local"
				itemUrl="local"
				icon="fa-home"
			/>
			<MenuItem
				name="Remote"
				itemUrl="remote"
				icon="fa-cloud"
			/>
		</MenuItem>
	);
}

/*
	ROUTERS
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

	const EmojiOverview = lazy(() => import('./emoji/local/overview'));
	const EmojiDetail = lazy(() => import('./emoji/local/detail'));
	const RemoteEmoji = lazy(() => import('./emoji/remote'));

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Suspense fallback={<Loading/>}>
						<Switch>
							<Route path="/local" component={EmojiOverview} />
							<Route path="/local/:emojiId" component={EmojiDetail} />
							<Route path="/remote" component={RemoteEmoji} />
							<Route><Redirect to="/local" /></Route>
						</Switch>
					</Suspense>
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

	const Media = lazy(() => import('./actions/media'));
	const Keys = lazy(() => import('./actions/keys'));

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<ErrorBoundary>
					<Suspense fallback={<Loading/>}>
						<Switch>
							<Route path="/media" component={Media} />
							<Route path="/keys" component={Keys} />
							<Route><Redirect to="/media" /></Route>
						</Switch>
					</Suspense>
				</ErrorBoundary>
			</Router>
		</BaseUrlContext.Provider>
	);
}
