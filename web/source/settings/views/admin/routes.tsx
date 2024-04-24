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
import { Route, Router, Switch } from "wouter";
import EmojiDetail from "./emoji/local/detail";
import { EmojiOverview } from "./emoji/local/overview";
import RemoteEmoji from "./emoji/remote";
import InstanceSettings from "./settings";
import { InstanceRuleDetail, InstanceRules } from "./settings/rules";
import Media from "./actions/media";
import Keys from "./actions/keys";

/*
	EXPORTED COMPONENTS
*/

/**
 * Admininistration menu. Admin actions,
 * emoji import, instance settings.
 */
export function AdminMenu() {
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
 * Admininistration router. Admin actions,
 * emoji import, instance settings.
 */
export function AdminRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/admin";
	const absBase = parentUrl + thisBase;
	
	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Route path="/instance-settings" component={InstanceSettings}/>
				<Route path="/instance-rules" component={InstanceRules} />
				<Route path="/instance-rules/:ruleId" component={InstanceRuleDetail} />
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

function AdminEmojisRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/emojis";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Switch>
					<Route path="/local/:emojiId" component={EmojiDetail} />
					<Route path="/local" component={EmojiOverview} />
					<Route path="/remote" component={RemoteEmoji} />
					<Route component={EmojiOverview}/>
				</Switch>
			</Router>
		</BaseUrlContext.Provider>
	);
}

function AdminActionsRouter() {
	const parentUrl = useBaseUrl();
	const thisBase = "/actions";
	const absBase = parentUrl + thisBase;

	return (
		<BaseUrlContext.Provider value={absBase}>
			<Router base={thisBase}>
				<Switch>
					<Route path="/media" component={Media} />
					<Route path="/keys" component={Keys} />
					<Route component={Media}/>
				</Switch>
			</Router>
		</BaseUrlContext.Provider>
	);
}
